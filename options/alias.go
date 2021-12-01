package options

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type AliasType string

func (a AliasType) IsValid() error {
	switch a.Canonical() {
	case Literal, CamelCase, PascalCase, SnakeCase, UpperSnakeCase, KebabCase, DotCase, FilePattern, Command:
		return nil
	}
	return fmt.Errorf("'%s' is not a valid alias type", a)
}

func (a AliasType) String() string {
	return strings.ToLower(string(a))
}

func (a AliasType) Canonical() AliasType {
	return AliasType(a.String())
}

func (a AliasType) unexpectedFieldErr(field string) error {
	return fmt.Errorf("unexpected field for %s alias: '%s'", a, field)
}

const (
	Literal AliasType = "literal"

	CamelCase      AliasType = "camelcase"
	PascalCase     AliasType = "pascalcase"
	SnakeCase      AliasType = "snakecase"
	UpperSnakeCase AliasType = "uppersnakecase"
	KebabCase      AliasType = "kebabcase"
	DotCase        AliasType = "dotcase"

	FilePattern AliasType = "filepattern"

	Command AliasType = "command"
)

// Alias is a catch-all type for alias configurations
type Alias struct {
	Type AliasType `mapstructure:"type"`
	Name string    `mapstructure:"name"`

	// Literal
	Flags map[string][]string `mapstructure:"flags,omitempty"`

	// FilePattern
	Paths    []string `mapstructure:"paths,omitempty"`
	Patterns []string `mapstructure:"patterns,omitempty"`

	// Command
	Command *string `mapstructure:"command,omitempty"`
	Timeout *int64  `mapstructure:"timeout,omitempty"`
}

func (a *Alias) IsValid() error {
	err := a.Type.IsValid()
	if err != nil {
		return err
	}
	// Validate expected fields
	switch a.Type {
	case Literal:
		if a.Flags == nil {
			return errors.New("literal aliases must provide an 'flags'")
		}
	case FilePattern:
		if len(a.Paths) == 0 {
			return errors.New("filepattern aliases must provide at least one path in 'paths'")
		}
		if len(a.Patterns) == 0 {
			return errors.New("filepattern aliases must provide at least one pattern in 'patterns'")
		}
		for _, pattern := range a.Patterns {
			if !strings.Contains(pattern, "FLAG_KEY") {
				return fmt.Errorf("filepattern regex '%s' must contain 'FLAG_KEY' for templating", pattern)
			}
			_, err := regexp.Compile(pattern)
			if err != nil {
				return fmt.Errorf("could not validate regex pattern: %v", err)
			}
		}
	case Command:
		if a.Command == nil {
			return errors.New("command aliases must provide a 'command'")
		}
		if a.Timeout != nil && *a.Timeout < 0 {
			return errors.New("field 'timeout' must be >= 0")
		}
	}

	// Validate unexpected fields
	var unexpectedField string
	switch {
	case a.Type != Literal:
		if a.Flags != nil {
			unexpectedField = "flags"
		}
	case a.Type != FilePattern:
		if len(a.Paths) > 0 {
			unexpectedField = "paths"
		}
		if len(a.Patterns) > 0 {
			unexpectedField = "patterns"
		}
	case a.Type != Command:
		if a.Command != nil {
			unexpectedField = "command"
		}
		if a.Timeout != nil {
			unexpectedField = "timeout"
		}
	}
	if unexpectedField != "" {
		return a.Type.unexpectedFieldErr(unexpectedField)
	}

	return nil
}
