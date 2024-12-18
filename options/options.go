package options

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/launchdarkly/ld-find-code-refs/v2/internal/validation"
)

const (
	maxProjKeyLength = 20 // Maximum project key length
)

type RepoType string

func (repoType RepoType) isValid() error {
	switch repoType {
	case GITHUB, GITLAB, BITBUCKET, CUSTOM:
		return nil
	default:
		return fmt.Errorf(`invalid value %q for "repoType": must be %s, %s, %s, or %s`, repoType, GITHUB, GITLAB, BITBUCKET, CUSTOM)
	}
}

const (
	GITHUB    RepoType = "github"
	GITLAB    RepoType = "gitlab"
	BITBUCKET RepoType = "bitbucket"
	CUSTOM    RepoType = "custom"
)

type Project struct {
	Key     string  `mapstructure:"key"`
	Dir     string  `mapstructure:"dir"`
	Aliases []Alias `mapstructure:"aliases"`
}
type Options struct {
	AccessToken         string `mapstructure:"accessToken"`
	BaseUri             string `mapstructure:"baseUri"`
	Branch              string `mapstructure:"branch"`
	CommitUrlTemplate   string `mapstructure:"commitUrlTemplate"`
	DefaultBranch       string `mapstructure:"defaultBranch"`
	Dir                 string `mapstructure:"dir" yaml:"-"`
	HunkUrlTemplate     string `mapstructure:"hunkUrlTemplate"`
	OutDir              string `mapstructure:"outDir"`
	ProjKey             string `mapstructure:"projkey"`
	RepoName            string `mapstructure:"repoName"`
	RepoType            string `mapstructure:"repoType"`
	RepoUrl             string `mapstructure:"repoUrl"`
	Revision            string `mapstructure:"revision"`
	Subdirectory        string `mapstructure:"subdirectory"`
	UserAgent           string `mapstructure:"userAgent"`
	ContextLines        int    `mapstructure:"contextLines"`
	Lookback            int    `mapstructure:"lookback"`
	UpdateSequenceId    int    `mapstructure:"updateSequenceId"`
	AllowTags           bool   `mapstructure:"allowTags"`
	Debug               bool   `mapstructure:"debug"`
	DryRun              bool   `mapstructure:"dryRun"`
	IgnoreServiceErrors bool   `mapstructure:"ignoreServiceErrors"`
	Prune               bool   `mapstructure:"prune"`

	// The following options can only be configured via YAML configuration

	Aliases    []Alias    `mapstructure:"aliases"`
	Delimiters Delimiters `mapstructure:"delimiters"`
	Projects   []Project  `mapstructure:"projects"`
}

type Delimiters struct {
	// If set to `true`, the default delimiters (single-quote, double-qoute, and backtick) will not be used unless provided as `additional` delimiters
	DisableDefaults bool     `mapstructure:"disableDefaults"`
	Additional      []string `mapstructure:"additional"`
}

func Init(flagSet *pflag.FlagSet) error {
	for _, f := range flags {
		usage := strings.ReplaceAll(f.usage, "\n", " ")
		switch value := f.defaultValue.(type) {
		case string:
			flagSet.StringP(f.name, f.short, value, usage)
		case int:
			flagSet.IntP(f.name, f.short, value, usage)
		case bool:
			flagSet.BoolP(f.name, f.short, value, usage)
		}
	}

	flagSet.VisitAll(func(f *pflag.Flag) {
		viper.BindEnv(f.Name, "LD_"+strcase.ToScreamingSnake(f.Name))
	})

	return viper.BindPFlags(flagSet)
}

func InitYAML() error {
	err := validateYAMLPreconditions()
	if err != nil {
		return err
	}
	absPath, err := validation.NormalizeAndValidatePath(viper.GetString("dir"))
	if err != nil {
		return err
	}
	subdirectoryPath := viper.GetString("subdirectory")
	viper.SetConfigName("coderefs")
	viper.SetConfigType("yaml")
	configPath := filepath.Join(absPath, subdirectoryPath, ".launchdarkly")
	viper.AddConfigPath(configPath)
	err = viper.ReadInConfig()
	if err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) {
		return err
	}
	return nil
}

// validatePreconditions ensures required flags have been set
func validateYAMLPreconditions() error {
	token := viper.GetString("accessToken")
	dir := viper.GetString("dir")
	missingRequiredOptions := []string{}
	if token == "" {
		missingRequiredOptions = append(missingRequiredOptions, "accessToken")
	}
	if dir == "" {
		missingRequiredOptions = append(missingRequiredOptions, "dir")
	}
	if len(missingRequiredOptions) > 0 {
		return fmt.Errorf("missing required option(s): %v", missingRequiredOptions)
	}
	return nil
}

func GetOptions() (Options, error) {
	var opts Options
	err := viper.Unmarshal(&opts)
	return opts, err
}

func GetWrapperOptions(dir string, merge func(Options) (Options, error)) (Options, error) {
	flags := pflag.CommandLine

	err := Init(flags)
	if err != nil {
		return Options{}, err
	}

	// Set precondition flags
	err = flags.Set("accessToken", os.Getenv("LD_ACCESS_TOKEN"))
	if err != nil {
		return Options{}, err
	}
	err = flags.Set("dir", dir)
	if err != nil {
		return Options{}, err
	}

	err = InitYAML()
	if err != nil {
		return Options{}, err
	}

	opts, err := GetOptions()
	if err != nil {
		return opts, err
	}

	return merge(opts)
}

func (o Options) ValidateRequired() error {
	missingRequiredOptions := []string{}
	if o.AccessToken == "" {
		missingRequiredOptions = append(missingRequiredOptions, "accessToken")
	}
	if o.Dir == "" {
		missingRequiredOptions = append(missingRequiredOptions, "dir")
	}
	if o.ProjKey == "" && len(o.Projects) == 0 {
		missingRequiredOptions = append(missingRequiredOptions, "projKey/projects")
	}
	if o.RepoName == "" {
		missingRequiredOptions = append(missingRequiredOptions, "repoName")
	}
	if len(missingRequiredOptions) > 0 {
		return fmt.Errorf("missing required option(s): %v", missingRequiredOptions)
	}

	if len(o.ProjKey) > 0 && len(o.Projects) > 0 {
		return fmt.Errorf("`--projKey` cannot be combined with `projects` in configuration")
	}

	if len(o.ProjKey) > maxProjKeyLength {
		return projKeyValidation(o.ProjKey)
	}

	if len(o.Projects) > 0 {
		for _, project := range o.Projects {
			if len(project.Key) > maxProjKeyLength {
				err := projKeyValidation(project.Key)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Validate ensures all options have been set to a valid value
func (o Options) Validate() error {
	if err := o.ValidateRequired(); err != nil {
		return err
	}

	maxContextLines := 5
	if o.ContextLines > maxContextLines {
		return fmt.Errorf(`invalid value %q for "contextLines": must be <= %d`, o.ContextLines, maxContextLines)
	}

	repoType := RepoType(strings.ToLower(o.RepoType))
	if err := repoType.isValid(); err != nil {
		return err
	}

	if o.RepoUrl != "" {
		if _, err := url.ParseRequestURI(o.RepoUrl); err != nil {
			return fmt.Errorf(`invalid value %q for "repoUrl": %+v`, o.RepoUrl, err)
		}
	}

	// match all non-control ASCII characters
	validDelims := regexp.MustCompile("^[\x20-\x7E]$")
	for i, d := range o.Delimiters.Additional {
		if !validDelims.MatchString(d) {
			return fmt.Errorf(`invalid value %q for "delimiters.additional[%d]": each delimiter must be a valid non-control ASCII character`, d, i)
		}
	}

	if _, err := validation.NormalizeAndValidatePath(o.Dir); err != nil {
		return fmt.Errorf(`invalid value for "dir": %+v`, err)
	}

	if o.OutDir != "" {
		if _, err := validation.NormalizeAndValidatePath(o.OutDir); err != nil {
			return fmt.Errorf(`invalid valid for "outDir": %+v`, err)
		}
	}

	for _, a := range o.Aliases {
		if err := a.IsValid(); err != nil {
			return err
		}
	}

	if o.Revision != "" && o.Branch == "" {
		return fmt.Errorf(`"branch" option is required when "revision" option is set`)
	}

	if len(o.Projects) > 0 {
		for _, project := range o.Projects {
			if project.Dir == "" {
				return nil
			}
			if err := validation.IsSubDirValid(o.Dir, project.Dir); err != nil {
				return err
			}
		}
	}

	return nil
}

func projKeyValidation(projKey string) error {
	if strings.HasPrefix(projKey, "sdk-") {
		return fmt.Errorf("provided project key (%s) appears to be a LaunchDarkly SDK key", "sdk-xxxx")
	} else if strings.HasPrefix(projKey, "api-") {
		return fmt.Errorf("provided project key (%s) appears to be a LaunchDarkly API access token", "api-xxxx")
	}

	return nil
}

func (o Options) GetProjectKeys() (projects []string) {
	for _, project := range o.Projects {
		projects = append(projects, project.Key)
	}
	return projects
}
