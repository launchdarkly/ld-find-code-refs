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

	"github.com/launchdarkly/ld-find-code-refs/internal/validation"
)

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
	ContextLines        int    `mapstructure:"contextLines"`
	UpdateSequenceId    int    `mapstructure:"updateSequenceId"`
	Debug               bool   `mapstructure:"debug"`
	DryRun              bool   `mapstructure:"dryRun"`
	IgnoreServiceErrors bool   `mapstructure:"ignoreServiceErrors"`

	// The following options can only be configured via YAML configuration

	Aliases    []Alias    `mapstructure:"aliases"`
	Delimiters Delimiters `mapstructure:"delimiters"`
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
	viper.SetConfigName("coderefs")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Join(absPath, ".launchdarkly"))
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
	if o.ProjKey == "" {
		missingRequiredOptions = append(missingRequiredOptions, "projKey")
	}
	if o.RepoName == "" {
		missingRequiredOptions = append(missingRequiredOptions, "repoName")
	}
	if len(missingRequiredOptions) > 0 {
		return fmt.Errorf("missing required option(s): %v", missingRequiredOptions)
	}
	return nil
}

// Validate ensures all options have been set to a valid value
func (o Options) Validate() error {
	err := o.ValidateRequired()
	if err != nil {
		return err
	}

	maxContextLines := 5
	if o.ContextLines > maxContextLines {
		return fmt.Errorf(`invalid value %q for "contextLines": must be <= %d`, o.ContextLines, maxContextLines)
	}

	repoType := strings.ToLower(o.RepoType)
	if repoType != "custom" && repoType != "github" && repoType != "bitbucket" {
		return fmt.Errorf(`invalid value %q for "repoType": must be "custom", "bitbucket", or "github"`, o.RepoType)
	}

	if o.RepoUrl != "" {
		_, err := url.ParseRequestURI(o.RepoUrl)
		if err != nil {
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

	_, err = validation.NormalizeAndValidatePath(o.Dir)
	if err != nil {
		return fmt.Errorf(`invalid value for "dir": %+v`, err)
	}

	if o.OutDir != "" {
		_, err = validation.NormalizeAndValidatePath(o.OutDir)
		if err != nil {
			return fmt.Errorf(`invalid valid for "outDir": %+v`, err)
		}
	}

	for _, a := range o.Aliases {
		err := a.IsValid()
		if err != nil {
			return err
		}
	}

	if o.Revision != "" && o.Branch == "" {
		return fmt.Errorf(`"branch" option is required when "revision" option is set`)
	}

	return nil
}
