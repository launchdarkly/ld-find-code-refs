package element

import "strings"

type ElementMatcher struct {
	Elements   []string
	Aliases    map[string][]string
	Delimiters []string
	ProjKey    string
	Directory  string
}

type Matcher struct {
	Elements   []ElementMatcher
	Type       string
	CtxLines   int
	Delimiters string
}

func (m Matcher) MatchElement(line, flagKey string) bool {
	if m.Delimiters == "" && strings.Contains(line, flagKey) {
		return true
	}
	for _, left := range m.Delimiters {
		for _, right := range m.Delimiters {
			var sb strings.Builder
			sb.Grow(len(flagKey) + 2)
			sb.WriteRune(left)
			sb.WriteString(flagKey)
			sb.WriteRune(right)
			if strings.Contains(line, sb.String()) {
				return true
			}
		}
	}
	return false
}
