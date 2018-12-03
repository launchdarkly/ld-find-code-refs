package parse

import (
	"flag"
	"fmt"
	"regexp"
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
	AccessToken   = StringOption("accessToken")
	BaseUri       = StringOption("baseUri")
	CloneEndpoint = StringOption("cloneEndpoint")
	ContextLines  = IntOption("contextLines")
	DefaultBranch = StringOption("defaultBranch")
	Dir           = StringOption("dir")
	Exclude       = StringOption("exclude")
	ProjKey       = StringOption("projKey")
	PushTime      = Int64Option("pushTime")
	RepoHead      = StringOption("repoHead")
	RepoName      = StringOption("repoName")
	RepoOwner     = StringOption("repoOwner")
	RepoType      = StringOption("repoType")
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

var options = optionMap{
	AccessToken:   option{"", "LaunchDarkly personal access token with write-level access.", true},
	BaseUri:       option{"https://app.launchdarkly.com", "LaunchDarkly base URI.", false},
	CloneEndpoint: option{"", "If provided, will clone the repo from this endpoint. If authentication is required, this endpoint should be authenticated. Supports the https protocol for git cloning. Example: https://username:password@github.com/username/repository.git", false},
	ContextLines:  option{-1, "The number of context lines to send to LaunchDarkly. If < 0, no source code will be sent to LaunchDarkly. If 0, only the lines containing flag references will be sent. If > 0, will send that number of context lines above and below the flag reference. A maximum of 5 context lines may be provided.", false},
	DefaultBranch: option{"master", "The git default branch. The LaunchDarkly UI will default to this branch.", false},
	Dir:           option{"", "Path to existing checkout of the git repo. If a cloneEndpoint is provided, this option is not required.", false},
	Exclude:       option{"", "Exclude any files or directories that match this regular expression pattern", false},
	ProjKey:       option{"", "LaunchDarkly project key.", true},
	PushTime:      option{int64(0), "The time the push was initiated formatted as a unix millis timestamp.", true},
	RepoHead:      option{"master", "The HEAD or ref to retrieve code references from.", false},
	RepoName:      option{"", "Git repo name. Will be displayed in LaunchDarkly.", true},
	RepoOwner:     option{"", "Git repo owner/org.", false},
	RepoType:      option{"custom", "github|bitbucket|custom", false},
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
		return fmt.Errorf("Required option %s not set", opt), flag.PrintDefaults
	}
	err = ContextLines.maximumError(5)
	if err != nil {
		return err, flag.PrintDefaults
	}
	repoType := strings.ToLower(RepoType.Value())
	if repoType != "custom" && repoType != "github" && repoType != "bitbucket" {
		return fmt.Errorf("Repo type must be \"custom\", \"bitbucket\", or \"github\""), flag.PrintDefaults
	}
	_, err = regexp.Compile(Exclude.Value())
	if err != nil {
		return fmt.Errorf("Exclude must be a valid regular expression: %+v", err), flag.PrintDefaults
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
