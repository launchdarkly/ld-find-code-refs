package main

import (
	"flag"
	"strconv"
	"strings"
	"time"
	"os"
	"regexp"
	"encoding/json"

	"github.com/launchdarkly/git-flag-parser/internal/ld"
	"github.com/launchdarkly/git-flag-parser/internal/git"

	log "github.com/sirupsen/logrus"
)

func init() {
	initOptions()
	initLogging()
}

func main() {
	workspace := getOption("workspace")
	if workspace == "" {
		log.WithFields(log.Fields{"workspace": workspace}).Fatal("workspace option not set")
		return
	}
	currBranch := getOption("repoHead")
	if currBranch == "" {
		log.WithFields(log.Fields{"currBranch": currBranch}).Fatal("currBranch option not set")
		return
	}
	pushTime, err := strconv.ParseInt(getOption("pushTime"), 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"pushTime": pushTime, "error": err}).Fatal("Error parsing pushTime option")
		return
	}

	cmd := git.Commander{Workspace: workspace, Head: currBranch}
	endpoint := getOption("cloneEndpoint")
	if endpoint != "" {
		err := cmd.Clone(endpoint)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Fatal("Unable to clone repo")
			return
		}
	}
	headSha, err := cmd.RevParse()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("Unable to parse current commit sha")
		return
	}


	baseUri := getOption("baseUri")
	ldApi := ld.InitApiClient(ld.ApiOptions{ApiKey: getOption("apiKey"), BaseUri: baseUri})

	b := &branch{Branch: currBranch, IsDefault: getOption("defaultBranch") == currBranch, PushTime: pushTime, Head: headSha}
	b, err = b.findReferences(cmd, ldApi)
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error()}).Fatal("Error searching for flag key references")
		return
	}

	branchBytes, err := json.Marshal(b)
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error()}).Fatal("Error marshalling branch to json")
	}
	err = ldApi.PutCodeReferenceBranch(branchBytes)
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error()}).Fatal("Error sending code references to LaunchDarkly")
	}
}

func getFlags(ldApi ld.ApiClient) ([]string, error) {
	projKey := getOption("projKey")
	log.WithFields(log.Fields{"projKey": projKey}).Debug("Requesting flag list from LaunchDarkly")
	flags, err := ldApi.GetFlagKeyList(projKey)
	if err != nil {
		log.WithFields(log.Fields{"projKey": projKey, "error": err.Error()}).Error("Error retrieving flag list from LaunchDarkly")
		return nil, err
	}
	return flags, nil
}

func contextLineCount() int {
	contextLines := getOption("contextLines")
	if i, err := strconv.Atoi(contextLines); err == nil && i > 0 {
		return i
	}
	return -1
}

type branch struct {
	Branch     string      `json:"branch"`
	Head       string      `json:"head"`
	IsDefault  bool        `json:"isDefault"`
	PushTime   int64       `json:"pushTime"`
	SyncTime   int64       `json:"syncTime"`
	References []reference `json:"references,omitempty"`
}

type reference struct {
	Path     string   `json:"path"`
	Line     string   `json:"line"`
	Context  string   `json:"context,omitempty"`
	FlagKeys []string `json:"flagKeys,omitempty"`
}

func (b *branch) findReferences(cmd git.Commander, ldApi ld.ApiClient) (*branch, error) {
	flags, err := getFlags(ldApi)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Unable to retrieve flag keys")
		return b, err
	}

	exclude := getOption("exclude")
	var excludeRegex *regexp.Regexp
	if exclude != "" {
		excludeRegex, err = regexp.Compile(exclude)
		if err != nil {
			log.WithFields(log.Fields{"exclude": exclude, "error": err}).Error("Invalid exclude option")
			return b, err
		}
	}

	err = cmd.Checkout()
	if err != nil {
		return b, err
	}

	grepResult, err := cmd.Grep(flags, contextLineCount())
	if err != nil {
		return b, err
	}

	b.SyncTime = makeTimestamp()
	b.References = generateReferencesFromGrep(flags, grepResult, excludeRegex)
	return b, nil
}

func generateReferencesFromGrep(flags []string, grepResult [][]string, exclude *regexp.Regexp) []reference {
	references := []reference{}
	ctxLines := contextLineCount()

FindRefs:
	for _, r := range grepResult {
		path := r[1]
		isReference := r[2] == ":"
		lineNumber := r[3]
		context := r[4]

		if exclude != nil && exclude.MatchString(r[1]) {
			continue FindRefs
		}

		ref := reference{Path: path, Line: lineNumber}
		if isReference {
			ref.FlagKeys = findReferencedFlags(context, flags)
		}
		if ctxLines >= 0 {
			ref.Context = context
		}
		references = append(references, ref)
	}

	return references
}

func findReferencedFlags(ref string, flags []string) []string {
	ret := []string{}
	for _, flag := range flags {
		if strings.Contains(ref, flag) {
			ret = append(ret, flag)
		}
	}
	return ret
}

func initLogging() {
  log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	logLevel := log.WarnLevel
	switch strings.ToUpper(getOption("logLevel")) {
	case "TRACE":
		logLevel = log.TraceLevel
	case "DEBUG":
		logLevel = log.DebugLevel
	case "INFO":
		logLevel = log.InfoLevel
	case "WARN":
		logLevel = log.WarnLevel
	case "ERROR":
		logLevel = log.ErrorLevel
	case "FATAL":
		logLevel = log.FatalLevel
	case "PANIC":
		logLevel = log.PanicLevel
	}
  log.SetLevel(logLevel)
}

type option struct {
	name         string
	defaultValue string
	usage        string
}

func initOptions() {
	options := []option{
		option{"apiKey", "", "LaunchDarkly personal access token with write-level access."},
		option{"projKey", "", "LaunchDarkly project key."},
		option{"baseUri", "", "LaunchDarkly base URI."},
		option{"exclude", "", "Exclude any files with code references that match this regex pattern"},
		option{"contextLines", "-1", "The number of context lines. If < 0, no source code will be sent to LaunchDarkly. If 0, only the lines containing flag references will be sent. If > 0, will send that number of context lines above and below the flag reference."},
		option{"repoName", "", "Git repo name. Will be displayed in LaunchDarkly."},
		option{"repoOwner", "", "Git repo owner/org."},
		option{"workspace", "", "Path to git repo."},
		option{"cloneEndpoint", "", "If provided, will clone the repo from this endpoint to the provided workspace. If authentication is required, this endpoint should be authenticated. Example: https://username:password@github.com/username/repository.git"},
		option{"repoHead", "master", "The HEAD or ref to retrieve code references from."},
		option{"defaultBranch", "master", "The git default branch. The LaunchDarkly UI will default to this branch."},
		option{"pushTime", "", ""},
		option{"logLevel", "WARN", ""},
	}
	for _, o := range options {
		flag.String(o.name, o.defaultValue, o.usage)
	}
	flag.Parse()
}

func getOption(name string) string {
	return flag.Lookup(name).Value.String()
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
