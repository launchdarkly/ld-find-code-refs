package search

import (
	"strings"

	"github.com/bucketeer-io/code-refs/aliases"
	"github.com/bucketeer-io/code-refs/internal/log"
	"github.com/bucketeer-io/code-refs/options"
)

type Matcher struct {
	Element  ElementMatcher
	ctxLines int
}

func NewEnvironmentMatcher(opts options.Options, dir string, flagKeys []string) Matcher {
	delimiters := strings.Join(GetDelimiters(opts), "")

	aliasesByFlagKey, err := aliases.GenerateAliases(flagKeys, opts.Aliases, dir)
	if err != nil {
		log.Error.Fatalf("failed to generate aliases: %s", err)
	}

	element := NewElementMatcher("", opts.Subdirectory, delimiters, flagKeys, aliasesByFlagKey)

	return Matcher{
		ctxLines: opts.ContextLines,
		Element:  element,
	}
}

func (m Matcher) MatchElement(line, element string) bool {
	if e, exists := m.Element.matcherByElement[element]; exists {
		if e.Iter(line).Next() != nil {
			return true
		}
	}
	return false
}

func (m Matcher) GetElementMatcher() *ElementMatcher {
	return &m.Element
}

func (m Matcher) FindAliases(line, element string) []string {
	return m.Element.FindAliases(line, element)
}

func (m Matcher) GetElements() (elements [][]string) {
	return [][]string{m.Element.Elements}
}

func buildElementPatterns(flags []string, delimiters string) map[string][]string {
	patternsByFlag := make(map[string][]string, len(flags))
	for _, flag := range flags {
		var patterns []string
		if delimiters != "" {
			patterns = make([]string, 0, len(delimiters)*len(delimiters))
			for _, left := range delimiters {
				for _, right := range delimiters {
					var sb strings.Builder
					sb.Grow(len(flag) + 2) //nolint:mnd
					sb.WriteRune(left)
					sb.WriteString(flag)
					sb.WriteRune(right)
					patterns = append(patterns, sb.String())
				}
			}
		} else {
			patterns = []string{flag}
		}
		patternsByFlag[flag] = patterns
	}
	return patternsByFlag
}
