package options

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/launchdarkly/ld-find-code-refs/internal/validation"
	"gopkg.in/yaml.v2"
)

type AliasType string

func (a AliasType) IsValid() error {
	switch a {
	case Literal, CamelCase, PascalCase, SnakeCase, UpperSnakeCase, KebabCase, DotCase, FilePattern, Command:
		return nil
	}
	return fmt.Errorf("%s is not a valid alias type", a)
}

func (a Alias) Generate(flag string) ([]string, error) {
	ret := []string{}
	switch a.Type {
	case Literal:
		ret = a.Flags[flag]
	case CamelCase:
		ret = []string{strcase.ToLowerCamel(flag)}
	case PascalCase:
		ret = []string{strcase.ToCamel(flag)}
	case SnakeCase:
		ret = []string{strcase.ToSnake(flag)}
	case UpperSnakeCase:
		ret = []string{strcase.ToScreamingSnake(flag)}
	case KebabCase:
		ret = []string{strcase.ToKebab(flag)}
	case DotCase:
		ret = []string{strcase.ToDelimited(flag, '.')}
	case FilePattern:
		pattern := regexp.MustCompile(strings.ReplaceAll(*a.Pattern, "FLAG_KEY", flag))
		results := pattern.FindAllStringSubmatch(string(a.FileContents), -1)
		for _, res := range results {
			if len(res) > 1 {
				ret = append(ret, res[1:]...)
			}
		}
	case Command:
		ctx := context.Background()
		if a.Timeout != nil && *a.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Second*time.Duration(*a.Timeout)))
			defer cancel()
		}
		tokens := strings.Split(*a.Command, " ")
		name := tokens[0]
		args := []string{}
		if len(tokens) > 1 {
			args = tokens[1:]
		}
		/* #nosec */
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Stdin = strings.NewReader(flag)
		cmd.Dir = Dir.Value()
		stdout, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to execute alias command: %w", err)
		}
		err = json.Unmarshal(stdout, &ret)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal json output of alias command: %w", err)
		}
	}

	return ret, nil
}

const (
	Literal AliasType = "literal"

	CamelCase      AliasType = "camelCase"
	PascalCase     AliasType = "pascalCase"
	SnakeCase      AliasType = "snakeCase"
	UpperSnakeCase AliasType = "upperSnakeCase"
	KebabCase      AliasType = "kebabCase"
	DotCase        AliasType = "dotCase"

	FilePattern AliasType = "filePattern"

	Command AliasType = "command"
)

type Alias struct {
	Type    AliasType           `yaml:"type"`
	Flags   map[string][]string `yaml:"flags,omitempty"`
	Path    *string             `yaml:"path,omitempty"`
	Pattern *string             `yaml:"pattern,omitempty"`
	Command *string             `yaml:"command,omitempty"`
	Timeout *int64              `yaml:"timeout,omitempty"`
	// data for pattern matching
	FileContents []byte `yaml:"-"`
}

func (a *Alias) IsValid() error {
	err := a.Type.IsValid()
	if err != nil {
		return err
	}

	// TODO: DRY this up
	switch a.Type {
	case Literal:
		if a.Flags == nil {
			return errors.New("literal aliases must provide an 'flags'")
		}
		if a.Command != nil {
			return errors.New("unexpected field for literal alias: 'command'")
		}
		if a.Path != nil {
			return errors.New("unexpected field for literal alias: 'path'")
		}
		if a.Pattern != nil {
			return errors.New("unexpected field for literal alias: 'pattern'")
		}
		if a.Timeout != nil {
			return errors.New("unexpected field for literal alias: 'timeout'")
		}
	case CamelCase, PascalCase, SnakeCase, UpperSnakeCase, KebabCase, DotCase:
		if a.Flags != nil {
			return errors.New("unexpected field for case alias: 'flags'")
		}
		if a.Command != nil {
			return errors.New("unexpected field for case alias: 'command'")
		}
		if a.Path != nil {
			return errors.New("unexpected field for case alias: 'path'")
		}
		if a.Pattern != nil {
			return errors.New("unexpected field for case alias: 'pattern'")
		}
		if a.Timeout != nil {
			return errors.New("unexpected field for literal alias: 'timeout'")
		}
	case FilePattern:
		if a.Path == nil {
			return errors.New("filePattern aliases must provide a 'path'")
		}
		if a.Pattern == nil {
			return errors.New("filePattern aliases must provide a 'pattern'")
		}

		if !strings.Contains(*a.Pattern, "FLAG_KEY") {
			return errors.New("pattern must contain 'FLAG_KEY' for templating")
		}
		_, err := regexp.Compile(*a.Pattern)
		if err != nil {
			return fmt.Errorf("could not validate regex pattern: %v", err)
		}

		if a.Flags != nil {
			return errors.New("unexpected field for filePattern alias: 'flags'")
		}
		if a.Command != nil {
			return errors.New("unexpected field for case alias: 'command'")
		}
		if a.Timeout != nil {
			return errors.New("unexpected field for literal alias: 'timeout'")
		}
	case Command:
		if a.Command == nil {
			return errors.New("command aliases must provide a 'command'")
		}
		if a.Path != nil {
			return errors.New("unexpected field for command alias: 'path'")
		}
		if a.Pattern != nil {
			return errors.New("unexpected field for command alias: 'pattern'")
		}
		if a.Flags != nil {
			return errors.New("unexpected field for command alias: 'flags'")
		}
		if a.Timeout != nil && *a.Timeout < 0 {
			return errors.New("field 'timeout' must be >= 0")
		}
	}

	return nil
}

type YamlOptions struct {
	Aliases []Alias `yaml:"aliases"`
}

func (o *YamlOptions) IsValid() error {
	for _, a := range o.Aliases {
		err := a.IsValid()
		if err != nil {
			return err
		}
	}
	return nil
}

func Yaml() (*YamlOptions, error) {
	pathToYaml := filepath.Join(Dir.Value(), ".launchdarkly/config.yaml")
	if !validation.FileExists(pathToYaml) {
		return nil, nil
	}

	/* #nosec */
	data, err := ioutil.ReadFile(pathToYaml)
	if err != nil {
		return nil, err
	}

	o := YamlOptions{}
	err = yaml.Unmarshal(data, &o)
	if err != nil {
		return nil, err
	}
	err = o.IsValid()
	if err != nil {
		return nil, err
	}
	for i, a := range o.Aliases {
		if a.Path != nil {
			path := filepath.Join(Dir.Value(), filepath.Dir(*a.Path))
			absPath, err := validation.NormalizeAndValidatePath(path)
			if err != nil {
				return nil, err
			}

			absFile := filepath.Join(absPath, filepath.Base(*a.Path))
			if !validation.FileExists(absFile) {
				return nil, fmt.Errorf("could not find file at path '%s'", absFile)
			}
			/* #nosec */
			data, err := ioutil.ReadFile(absFile)
			if err != nil {
				return nil, fmt.Errorf("could not process file at path '%s': %v", absFile, err)
			}
			a.FileContents = data
			o.Aliases[i] = a
		}
	}
	return &o, nil
}
