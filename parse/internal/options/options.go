package options

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Can't wait for contracts
type Option interface {
	name() string
}

type StringOption string
type IntOption string
type Int64Option string

func (o StringOption) name() string {
	return string(o)
}
func (o IntOption) name() string {
	return string(o)
}
func (o Int64Option) name() string {
	return string(o)
}

func (o StringOption) Value() string {
	return flag.Lookup(string(o)).Value.String()
}

func (o IntOption) Value() int {
	return flag.Lookup(string(o)).Value.(flag.Getter).Get().(int)
}

func (o IntOption) maximumError(max int) error {
	if o.Value() > max {
		return fmt.Errorf("%s option must be <= %d", string(o), max)
	}
	return nil
}

func (o Int64Option) Value() int64 {
	return flag.Lookup(string(o)).Value.(flag.Getter).Get().(int64)
}

const (
	AccessToken       = StringOption("accessToken")
	BaseUri           = StringOption("baseUri")
	CloneEndpoint     = StringOption("cloneEndpoint")
	ContextLines      = IntOption("contextLines")
	DefaultBranch     = StringOption("defaultBranch")
	Dir               = StringOption("dir")
	Exclude           = StringOption("exclude")
	ProjKey           = StringOption("projKey")
	UpdateSequenceId  = Int64Option("updateSequenceId")
	RepoHead          = StringOption("repoHead")
	RepoName          = StringOption("repoName")
	RepoType          = StringOption("repoType")
	RepoUrl           = StringOption("repoUrl")
	CommitUrlTemplate = StringOption("commitUrlTemplate")
	HunkUrlTemplate   = StringOption("hunkUrlTemplate")
)

type option struct {
	defaultValue interface{}
	usage        string
	required     bool
}

type optionMap map[Option]option

func (m optionMap) find(name string) *option {
	for n, o := range m {
		if n.name() == name {
			return &o
		}
	}
	return nil
}

const (
	noUpdateSequenceId = int64(-1)
	noContextLines     = -1
)

var options = optionMap{
	AccessToken:       option{"", "LaunchDarkly personal access token with write-level access.", true},
	BaseUri:           option{"https://app.launchdarkly.com", "LaunchDarkly base URI.", false},
	CloneEndpoint:     option{"", "If provided, will clone the repo from this endpoint. If authentication is required, this endpoint should be authenticated. Supports the https protocol for git cloning. Example: https://username:password@github.com/username/repository.git", false},
	ContextLines:      option{noContextLines, "The number of context lines to send to LaunchDarkly. If < 0, no source code will be sent to LaunchDarkly. If 0, only the lines containing flag references will be sent. If > 0, will send that number of context lines above and below the flag reference. A maximum of 5 context lines may be provided.", false},
	DefaultBranch:     option{"master", "The git default branch. The LaunchDarkly UI will default to this branch.", false},
	Dir:               option{"", "Path to existing checkout of the git repo. If a cloneEndpoint is provided, this option is not required.", false},
	Exclude:           option{"", `A regular expression (PCRE) defining the files and directories which the flag parser should exclude. Partial matches are allowed. Examples: "vendor/", "vendor/*`, false},
	ProjKey:           option{"", "LaunchDarkly project key.", true},
	UpdateSequenceId:  option{noUpdateSequenceId, `An integer representing the order number of code reference updates. Used to version updates across concurrent executions of the flag parser. If not provided, data will always be updated. If provided, data will only be updated if the existing "updateSequenceId" is less than the new "updateSequenceId". Examples: the time a "git push" was initiated, CI build number, the current unix timestamp.`, false},
	RepoHead:          option{"master", "he branch to scan for code references. Should be provided if the `git push` was initiated on a non-master branch.", false},
	RepoName:          option{"", `Git repo name. Will be displayed in LaunchDarkly. Case insensitive. Both a repo name and the repo name with an organization identifier are valid. Examples: "linux", "torvalds/linux."`, true},
	RepoType:          option{"custom", "The repo service provider. Used to correctly categorize repositories in the LaunchDarkly UI. Aceptable values: github|bitbucket|custom.", false},
	RepoUrl:           option{"", "The display url for the repository. If provided for a github or bitbucket repository, LaunchDarkly will attempt to automatically generate source code links.", false},
	CommitUrlTemplate: option{"", "If provided, LaunchDarkly will attempt to generate links to your Git service provider per commit. Example: `https://github.com/launchdarkly/git-flag-parser/commit/${sha}`. Allowed template variables: `branchName`, `sha`. If `commitUrlTemplate` is not provided, but `repoUrl` is provided and `repoType` is not custom, LaunchDarkly will automatically generate links to the repository for each commit.", false},
	HunkUrlTemplate:   option{"", "If provided, LaunchDarkly will attempt to generate links to your Git service provider per code reference. Example: `https://github.com/launchdarkly/git-flag-parser/blob/${sha}/${filePath}#L${lineNumber}`. Allowed template variables: `sha`, `filePath`, `lineNumber`. If `hunkUrlTemplate` is not provided, but repoUrl is provided and `repoType` is not custom, LaunchDarkly will automatically generate links to the repository for each code reference.", false},
}

// Init reads specified options and exits if options of invalid types or unspecified options were provided.
// Returns an error if a required option has not been set, or if an option is invalid.
func Init() (err error, errCb func()) {
	if !populated {
		Populate()
	}

	flag.Parse()

	opt := ""
	flag.VisitAll(func(f *flag.Flag) {
		o := options.find(f.Name)
		if o != nil && o.required {
			val := f.Value.(flag.Getter).Get()
			switch v := val.(type) {
			case int64:
				if v == 0 {
					opt = f.Name
				}
			case string:
				if v == "" {
					opt = f.Name
				}
			}
		}
	})

	if opt != "" {
		return fmt.Errorf("required option %s not set", opt), flag.PrintDefaults
	}
	err = ContextLines.maximumError(5)
	if err != nil {
		return err, flag.PrintDefaults
	}
	repoType := strings.ToLower(RepoType.Value())
	if repoType != "custom" && repoType != "github" && repoType != "bitbucket" {
		return fmt.Errorf("repo type must be \"custom\", \"bitbucket\", or \"github\""), flag.PrintDefaults
	}
	_, err = regexp.Compile(Exclude.Value())
	if err != nil {
		return fmt.Errorf("exclude must be a valid regular expression: %+v", err), flag.PrintDefaults
	}
	_, err = url.Parse(RepoUrl.Value())
	if err != nil {
		return fmt.Errorf("error parsing repo url: %+v", err), flag.PrintDefaults
	}
	return nil, flag.PrintDefaults
}

var populated = false

func Populate() {
	populated = true
	for n, o := range options {
		name := n.name()
		switch v := o.defaultValue.(type) {
		case int64:
			flag.Int64(name, v, o.usage)
		case int:
			flag.Int(name, v, o.usage)
		case string:
			flag.String(name, v, o.usage)
		}
	}
}

func GetLDOptionsFromEnv() (map[string]string, error) {
	ldOptions := map[string]string{
		"accessToken":  os.Getenv("LD_ACCESS_TOKEN"),
		"projKey":      os.Getenv("LD_PROJ_KEY"),
		"exclude":      os.Getenv("LD_EXCLUDE"),
		"contextLines": os.Getenv("LD_CONTEXT_LINES"),
		"baseUri":      os.Getenv("LD_BASE_URI"),
	}
	_, err := regexp.Compile(ldOptions["exclude"])
	if err != nil {
		return ldOptions, fmt.Errorf("couldn't parse LD_EXCLUDE as regex: %+v", err)
	}

	if ldOptions["contextLines"] == "" {
		ldOptions["contextLines"] = "-1"
	}
	_, err = strconv.ParseInt(ldOptions["contextLines"], 10, 32)
	if err != nil {
		return ldOptions, fmt.Errorf("coudln't parse LD_CONTEXT_LINES as an integer: %+v", err)
	}

	return ldOptions, nil
}
