package search

import (
	"strings"

	"github.com/launchdarkly/ld-find-code-refs/v2/internal/helpers"
)

type Matcher struct {
	Elements []ElementMatcher
	ctxLines int
}

func (m Matcher) MatchElement(line, element string) bool {
	for _, em := range m.Elements {
		if e, exists := em.matcherByElement[element]; exists {
			if e.Iter(line).Next() != nil {
				return true
			}
		}
	}

	return false
}

func (m Matcher) GetProjectElementMatcher(projectKey string) *ElementMatcher {
	var elementMatcher ElementMatcher
	for _, element := range m.Elements {
		if element.ProjKey == projectKey {
			elementMatcher = element
			break
		}
	}
	return &elementMatcher
}

func (m Matcher) FindAliases(line, element string) []string {
	matches := make([]string, 0)
	for _, em := range m.Elements {
		matches = append(matches, em.FindAliases(line, element)...)
	}
	return helpers.Dedupe(matches)
}

func (m Matcher) GetElements() (elements [][]string) {
	for _, element := range m.Elements {
		elements = append(elements, element.Elements)
	}
	return elements
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
					sb.Grow(len(flag) + 2)
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
