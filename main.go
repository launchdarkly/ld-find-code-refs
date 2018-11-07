package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	ld "github.com/launchdarkly/git-flag-parser/ld"
)

type option struct {
	name         string
	defaultValue string
	usage        string
}

func init() {
	options := []option{
		option{"apiKey", "", "LaunchDarkly personal access token with write-level access."},
		option{"projKey", "", "LaunchDarkly project key."},
		option{"baseUri", "", "LaunchDarkly base URI."},
		option{"exclude", "", "Exclude any files with code references that match this regex pattern"},
		option{"contextLines", "-1", "The number of context lines. If < 0, no source code will be sent to LaunchDarkly. If 0, only the lines containing flag references will be sent. If > 0, will send that number of context lines above and below the flag reference."},
		option{"repoName", "", "Git repo name. Will be displayed in LaunchDarkly."},
		option{"repoOwner", "", "Git repo owner/org."},
		option{"repoPath", "", "Path to git repo."},
		// option{"repoCloneUrl", "", "If provided, will clone the repo from this url to repoPath."},
		// option{"repoCloneAuth", "", "Git credentials"},
		option{"repoHead", "master", "The HEAD or ref to retrieve code references from."},
		option{"defaultBranch", "master", "The git default branch. The LaunchDarkly UI will default to this branch."},
		option{"pushTime", "", ""},
	}
	for _, o := range options {
		flag.String(o.name, o.defaultValue, o.usage)
	}
	flag.Parse()
}

func getArg(name string) string {
	return flag.Lookup(name).Value.String()
}

func getFlags() []string {
	ldApi := ld.InitApiClient(ld.ApiOptions{ApiKey: getArg("apiKey"), BaseUri: getArg("baseUri")})
	flags, err := ldApi.FlagKeyList(context.Background(), getArg("projKey"))
	if err != nil {
		fmt.Println(err)
		return []string{}
	}
	return flags
}

func contextLineCount() int {
	contextLines := getArg("contextLines")
	if i, err := strconv.Atoi(contextLines); err == nil && i > 0 {
		return i
	}
	return -1
}

func generateGrep(flags []string) *exec.Cmd {
	sh := exec.Command("git")
	sh.Args = make([]string, 0, 2*len(flags)+4)
	sh.Args = append(sh.Args, "git", "grep", "-nF")
	ctxLines := contextLineCount()

	if ctxLines > 0 {
		sh.Args = append(sh.Args, fmt.Sprintf("-C%d", ctxLines))
	}

	for _, f := range flags {
		sh.Args = append(sh.Args, "-e", f)
	}

	sh.Dir = getArg("repoPath")
	return sh
}

func findRefs(flags []string) (string, error) {
	checkout := exec.Command("git", "checkout", getArg("repoHead"))
	checkout.Dir = getArg("repoPath")
	err := checkout.Run()
	if err != nil {
		return "", err
	}
	cmd := generateGrep(flags)
	out, err := cmd.Output()
	return string(out), err
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

var refRegex, _ = regexp.Compile("(.+)(:|-)([0-9]+)[:-](.+)")

func getHeadSha() (string, error) {
	cmd := exec.Command("git", "rev-parse", getArg("repoHead"))
	cmd.Dir = getArg("repoPath")
	out, err := cmd.Output()
	return string(out), err
}

type branch struct {
	Ref        string      `json:"ref"`
	Head       string      `json:"head"`
	IsDefault  bool        `json:"isDefault"`
	PushTime   int64       `json:"pushTime"`
	SyncTime   int64       `json:"syncTime"`
	References []reference `json:"references"`
}

type reference struct {
	Path     string   `json:"path"`
	Line     string   `json:"line"`
	Text     string   `json:"text"`
	FlagKeys []string `json:"flagKeys"`
}

func main() {
	exclude := getArg("exclude")
	var excludeRegex *regexp.Regexp
	var err error
	if exclude != "" {
		excludeRegex, err = regexp.Compile(exclude)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	currBranch := getArg("repoHead")
	defaultBranch := getArg("defaultBranch")
	headSha, err := getHeadSha()
	if err != nil {
		fmt.Println(err)
		return
	}
	pushTime, err := strconv.ParseInt(getArg("pushTime"), 10, 64)
	if err != nil {
		fmt.Println(err)
		return
	}

	flags := getFlags()
	out, err := findRefs(flags)
	if err != nil {
		fmt.Println(out, err)
		return
	}
	references := []reference{}
	refs := refRegex.FindAllStringSubmatch(out, -1)
	ctxLines := contextLineCount()

FindRefs:
	for _, r := range refs {
		if excludeRegex != nil && excludeRegex.MatchString(r[1]) {
			continue FindRefs
		}

		ref := reference{Path: r[1], Line: r[3]}
		if r[2] == ":" {
			ref.FlagKeys = findReferencedFlags(r[4], flags)
		}
		if ctxLines >= 0 {
			ref.Text = r[4]
		}
		references = append(references, ref)
	}

	b := branch{
		Ref:        currBranch,
		Head:       headSha,
		IsDefault:  currBranch == defaultBranch,
		PushTime:   pushTime,
		SyncTime:   makeTimestamp(),
		References: references,
	}
	j, _ := json.Marshal(b)
	fmt.Println(string(j))
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
