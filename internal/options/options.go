package options

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/launchdarkly/ld-find-code-refs/internal/validation"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Init(cmd *cobra.Command) error {
	flags := cmd.PersistentFlags()

	for _, o := range options {
		usage := strings.ReplaceAll(o.usage, "\n", " ")
		switch value := o.v.(type) {
		case *string:
			flags.StringVarP(value, o.name, o.short, *value, usage)
		case *int:
			flags.IntVarP(value, o.name, o.short, *value, usage)
		case *bool:
			flags.BoolVarP(value, o.name, o.short, *value, usage)
		case *[]string:
			flags.StringArrayVarP(value, o.name, o.short, *value, usage)
		}
		if o.required {
			cmd.MarkPersistentFlagRequired(o.name)
		}
		if o.directory {
			cmd.MarkPersistentFlagDirname(o.name)
		}
	}

	return viper.BindPFlags(cmd.PersistentFlags())
}

func ValidateOptions() error {
	maxContextLines := 5
	if ContextLines > maxContextLines {
		return fmt.Errorf(`invalid value %q for "contextLines": must be <= %d`, ContextLines, maxContextLines)
	}

	repoType := strings.ToLower(RepoType)
	if repoType != "custom" && repoType != "github" && repoType != "bitbucket" {
		return fmt.Errorf(`invalid value %q for "repoType": must be "custom", "bitbucket", or "github"`, RepoType)
	}

	if RepoUrl != "" {
		_, err := url.ParseRequestURI(RepoUrl)
		if err != nil {
			return fmt.Errorf(`invalid value %q for "repoUrl": %+v`, RepoUrl, err)
		}
	}

	// match all non-control ASCII characters
	validDelims := regexp.MustCompile("[\x20-\x7E]")
	for _, d := range Delimiters {
		for _, dd := range d {
			if !validDelims.MatchString(string(dd)) {
				return fmt.Errorf(`invalid value %q for "delimiters": must be a valid non-control ASCII character`, d)
			}
		}
	}

	_, err := validation.NormalizeAndValidatePath(Dir)
	if err != nil {
		return fmt.Errorf(`invalid value for "dir": %+v`, err)
	}

	if OutDir != "" {
		_, err = validation.NormalizeAndValidatePath(OutDir)
		if err != nil {
			return fmt.Errorf(`invalid valid for "outDir": %+v`, err)
		}
	}

	return nil
}
