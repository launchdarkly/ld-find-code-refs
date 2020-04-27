package options

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/launchdarkly/ld-find-code-refs/internal/validation"
	"github.com/launchdarkly/ld-find-code-refs/internal/version"
)

type Option interface {
	name() string
}

type stringOption string
type intOption string
type int64Option string
type boolOption string
type runeSet string

func (o stringOption) name() string {
	return string(o)
}
func (o intOption) name() string {
	return string(o)
}
func (o int64Option) name() string {
	return string(o)
}
func (o boolOption) name() string {
	return string(o)
}
func (o runeSet) name() string {
	return string(o)
}

func (o stringOption) Value() string {
	return flag.Lookup(string(o)).Value.String()
}

func (o intOption) Value() int {
	return flag.Lookup(string(o)).Value.(flag.Getter).Get().(int)
}

func (o intOption) maximumError(max int) error {
	if o.Value() > max {
		return fmt.Errorf("%s option must be <= %d", string(o), max)
	}
	return nil
}

func (o int64Option) Value() int64 {
	return flag.Lookup(string(o)).Value.(flag.Getter).Get().(int64)
}

func (o boolOption) Value() bool {
	return flag.Lookup(string(o)).Value.(flag.Getter).Get().(bool)
}

func (o runeSet) Value() RuneSet {
	return flag.Lookup(string(o)).Value.(flag.Getter).Get().(RuneSet)
}

type RuneSet []rune

func (o *RuneSet) Set(value string) error {
	chars := value
	for _, v := range chars {
		if !o.contains(v) {
			*o = append(*o, v)
		}
	}
	return nil
}

func (o *RuneSet) String() string {
	return "[" + string(*o) + "]"
}

func (o *RuneSet) Get() interface{} {
	return *o
}

func (o *RuneSet) contains(c rune) bool {
	for _, v := range *o {
		if v == c {
			return true
		}
	}
	return false
}

const (
	AccessToken         = stringOption("accessToken")
	BaseUri             = stringOption("baseUri")
	Branch              = stringOption("branch")
	ContextLines        = intOption("contextLines")
	Debug               = boolOption("debug")
	DefaultBranch       = stringOption("defaultBranch")
	Dir                 = stringOption("dir")
	DryRun              = boolOption("dryRun")
	Exclude             = stringOption("exclude")
	IgnoreServiceErrors = boolOption("ignoreServiceErrors")
	OutDir              = stringOption("outDir")
	ProjKey             = stringOption("projKey")
	UpdateSequenceId    = int64Option("updateSequenceId")
	RepoName            = stringOption("repoName")
	RepoType            = stringOption("repoType")
	RepoUrl             = stringOption("repoUrl")
	CommitUrlTemplate   = stringOption("commitUrlTemplate")
	HunkUrlTemplate     = stringOption("hunkUrlTemplate")
	Version             = boolOption("version")
	Delimiters          = runeSet("delimiters")
	delimiterShort      = runeSet("D")
)

var Aliases []Alias

type option struct {
	defaultValue interface{}
	usage        string
	required     bool
}

type optionMap map[Option]option

func (m optionMap) find(name string) *option {
	for n, o := range m {
		opt := o
		if n.name() == name {
			return &opt
		}
	}
	return nil
}

const (
	noUpdateSequenceID  = int64(-1)
	defaultContextLines = 2
)

var (
	delimiters = RuneSet{'"', '\'', '`'}
)

var options = optionMap{
	AccessToken:         option{"", "LaunchDarkly personal access token with write-level access.", true},
	BaseUri:             option{"https://app.launchdarkly.com", "LaunchDarkly base URI.", false},
	Branch:              option{"", "The currently checked out git branch. If not provided, branch name will be auto-detected. Please provide when using CI systems that leave the repository in a detached HEAD state.", false},
	ContextLines:        option{defaultContextLines, "The number of context lines to send to LaunchDarkly. If < 0, no source code will be sent to LaunchDarkly. If 0, only the lines containing flag references will be sent. If > 0, will send that number of context lines above and below the flag reference. A maximum of 5 context lines may be provided.", false},
	DefaultBranch:       option{"", "The git default branch. The LaunchDarkly UI will default to this branch. If not provided, will fallback to `master`.", false},
	Dir:                 option{"", "Path to existing checkout of the git repo.", true},
	Debug:               option{false, "Enables verbose debug logging", false},
	DryRun:              option{false, "If enabled, the scanner will run without sending code references to LaunchDarkly. Combine with the `outDir` option to output code references to a CSV.", false},
	Exclude:             option{"", `A regular expression (PCRE) defining the files and directories which the flag finder should exclude. Partial matches are allowed. Examples: "vendor/", "\.css"`, false},
	OutDir:              option{"", "If provided, will output a csv file containing all code references for the project to this directory.", false},
	ProjKey:             option{"", "LaunchDarkly project key.", true},
	UpdateSequenceId:    option{noUpdateSequenceID, `An integer representing the order number of code reference updates. Used to version updates across concurrent executions of the flag finder. If not provided, data will always be updated. If provided, data will only be updated if the existing "updateSequenceId" is less than the new "updateSequenceId". Examples: the time a "git push" was initiated, CI build number, the current unix timestamp.`, false},
	RepoName:            option{"", `Git repo name. Will be displayed in LaunchDarkly. Case insensitive. Repo names must only contain letters, numbers, '.', '_' or '-'."`, true},
	RepoType:            option{"custom", "The repo service provider. Used to correctly categorize repositories in the LaunchDarkly UI. Aceptable values: github|bitbucket|custom.", false},
	RepoUrl:             option{"", "The display url for the repository. If provided for a github or bitbucket repository, LaunchDarkly will attempt to automatically generate source code links.", false},
	IgnoreServiceErrors: option{false, "If enabled, the scanner will terminate with exit code 0 when the LaunchDarkly API is unreachable or returns an unexpected response.", false},
	CommitUrlTemplate:   option{"", "If provided, LaunchDarkly will attempt to generate links to your Git service provider per commit. Example: `https://github.com/launchdarkly/ld-find-code-refs/commit/${sha}`. Allowed template variables: `branchName`, `sha`. If `commitUrlTemplate` is not provided, but `repoUrl` is provided and `repoType` is not custom, LaunchDarkly will automatically generate links to the repository for each commit.", false},
	HunkUrlTemplate:     option{"", "If provided, LaunchDarkly will attempt to generate links to your Git service provider per code reference. Example: `https://github.com/launchdarkly/ld-find-code-refs/blob/${sha}/${filePath}#L${lineNumber}`. Allowed template variables: `sha`, `filePath`, `lineNumber`. If `hunkUrlTemplate` is not provided, but repoUrl is provided and `repoType` is not custom, LaunchDarkly will automatically generate links to the repository for each code reference.", false},
	Version:             option{false, "If provided, the scanner will print the version number and exit early", false},
	Delimiters:          option{&delimiters, "Specifies additional delimiters used to match flag keys. Must be a non-control ASCII character. If more than one character is provided in `delimiters`, each character will be treated as a separate delimiter. Will only match flag keys with surrounded by any of the specified delimeters. This option may also be specified multiple times for multiple delimiters. By default, only flags delimited by single-quotes, double-quotes, and backticks will be matched.", false},
	delimiterShort:      option{&delimiters, "Same as -delimiters", false},
}

// Init reads specified options and exits if options of invalid types or unspecified options were provided.
// Returns an error if a required option has not been set, or if an option is invalid.
func Init() (err error, errCb func()) {
	if !populated {
		err = Populate()
		if err != nil {
			return err, nil
		}
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

	fmt.Println("ld-find-code-refs version", version.Version)
	if Version.Value() {
		os.Exit(0)
	}

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

	// match all non-control ASCII characters
	validDelims := regexp.MustCompile("[\x20-\x7E]")
	for _, d := range delimiters {
		if !validDelims.MatchString(string(d)) {
			return fmt.Errorf("delimiter option must be a valid non-control ASCII character"), flag.PrintDefaults
		}
	}

	_, err = validation.NormalizeAndValidatePath(Dir.Value())
	if err != nil {
		return fmt.Errorf("invalid dir: %s", err), flag.PrintDefaults
	}

	if OutDir.Value() != "" {
		_, err = validation.NormalizeAndValidatePath(OutDir.Value())
		if err != nil {
			return fmt.Errorf("invalid outDir: %s", err), flag.PrintDefaults
		}
	}

	yamlOptions, err := Yaml()
	if err != nil {
		return err, nil
	}
	if yamlOptions != nil {
		Aliases = yamlOptions.Aliases
	}

	return nil, flag.PrintDefaults
}

var populated = false

func Populate() error {

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
		case bool:
			flag.Bool(name, v, o.usage)
		case *RuneSet:
			flag.Var(v, name, o.usage)
		}
	}

	yamlOptions, err := Yaml()
	if err != nil {
		return err
	}
	if yamlOptions != nil {
		Aliases = yamlOptions.Aliases
	}

	return nil
}

// GetLDOptionsFromEnv returns a map of all expected environment variables for ld-find-code-refs wrappers
func GetLDOptionsFromEnv() (map[string]string, error) {
	ldOptions := map[string]string{
		"accessToken":         os.Getenv("LD_ACCESS_TOKEN"),
		"projKey":             os.Getenv("LD_PROJ_KEY"),
		"exclude":             os.Getenv("LD_EXCLUDE"),
		"contextLines":        os.Getenv("LD_CONTEXT_LINES"),
		"baseUri":             os.Getenv("LD_BASE_URI"),
		"debug":               os.Getenv("LD_DEBUG"),
		"delimiters":          os.Getenv("LD_DELIMITERS"),
		"ignoreServiceErrors": os.Getenv("LD_IGNORE_SERVICE_ERRORS"),
	}

	if ldOptions["debug"] == "" {
		ldOptions["debug"] = "false"
	}

	if ldOptions["ignoreServiceErrors"] == "" {
		ldOptions["ignoreServiceErrors"] = "false"
	}

	_, err := regexp.Compile(ldOptions["exclude"])
	if err != nil {
		return ldOptions, fmt.Errorf("couldn't parse LD_EXCLUDE as regex: %+v", err)
	}

	if ldOptions["contextLines"] == "" {
		ldOptions["contextLines"] = "2"
	}
	_, err = strconv.ParseInt(ldOptions["contextLines"], 10, 32)
	if err != nil {
		return ldOptions, fmt.Errorf("couldn't parse LD_CONTEXT_LINES as an integer: %+v", err)
	}

	return ldOptions, nil
}

func GetDebugOptionFromEnv() (bool, error) {
	debug := os.Getenv("LD_DEBUG")
	if debug == "" {
		return false, nil
	}
	return strconv.ParseBool(debug)
}
