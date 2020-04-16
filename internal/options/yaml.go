package options

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/launchdarkly/ld-find-code-refs/internal/validation"
	"gopkg.in/yaml.v2"
)

type AliasType string

func (a AliasType) IsValid() error {
	switch a {
	case Literal, CamelCase, PascalCase, SnakeCase, UpperSnakeCase, KebabCase, DotCase, FilePattern, JavaScript:
		return nil
	}
	return fmt.Errorf("%s is not a valid alias type", a)
}

func (a Alias) Generate(flag string) ([]string, error) {
	ret := []string{}
	switch a.Type {
	case Literal:
		ret = a.AliasMap[flag]
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
		// TODO
	case JavaScript:
		// TODO
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

	JavaScript AliasType = "js"
)

type Alias struct {
	Type     AliasType           `yaml:"type"`
	AliasMap map[string][]string `yaml:"aliasMap,omitempty"`
	Path     *string             `yaml:"path,omitempty"`
	Pattern  *string             `yaml:"pattern,omitempty"`
}

func (a Alias) IsValid() error {
	err := a.Type.IsValid()
	if err != nil {
		return err
	}

	// TODO: DRY this up
	switch a.Type {
	case Literal:
		if a.AliasMap == nil {
			return errors.New("literal aliases must provide an 'aliasMap'")
		}
	case CamelCase, PascalCase, SnakeCase, UpperSnakeCase, KebabCase, DotCase:
		if a.AliasMap != nil {
			return errors.New("unexpected field for case alias: 'aliasMap'")
		}
		if a.Path != nil {
			return errors.New("unexpected field for case alias: 'path'")
		}
		if a.Pattern != nil {
			return errors.New("unexpected field for case alias: 'pattern'")
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

		if a.AliasMap != nil {
			return errors.New("unexpected field for filePattern alias: 'aliasMap'")
		}

	case JavaScript:
		if a.Path == nil {
			return errors.New("js aliases must provide a 'path'")
		}
		if a.Pattern != nil {
			return errors.New("unexpected field for js alias: 'pattern'")
		}
		if a.AliasMap != nil {
			return errors.New("unexpected field for js alias: 'aliasMap'")
		}
	}

	if a.Path != nil {
		path := filepath.Join(Dir.Value(), filepath.Dir(*a.Path))
		absPath, err := validation.NormalizeAndValidatePath(path)
		if err != nil {
			return err
		}
		absFile := filepath.Join(absPath, filepath.Base(*a.Path))
		if !validation.FileExists(absFile) {
			return fmt.Errorf("could not find file at path '%s'", absFile)
		}
	}

	return nil
}

type YamlOptions struct {
	Aliases []Alias `yaml:"aliases"`
}

func (o YamlOptions) IsValid() error {
	for _, a := range o.Aliases {
		err := a.IsValid()
		return err
	}
	return nil
}

func Yaml() (*YamlOptions, error) {
	pathToYaml := filepath.Join(Dir.Value(), ".launchdarkly/config.yaml")
	if !validation.FileExists(pathToYaml) {
		return nil, nil
	}

	data, err := ioutil.ReadFile(pathToYaml)
	if err != nil {
		return nil, err
	}

	o := YamlOptions{}
	err = yaml.Unmarshal(data, &o)
	if err != nil {
		return nil, err
	}

	return &o, nil
}
