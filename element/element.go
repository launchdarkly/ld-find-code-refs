package element

import "strings"

type ElementMatcher struct {
	Elements   []string
	Aliases    map[string][]string
	Delimiters []string
	ProjKey    string
	Directory  string
}

type ElementsMatcher struct {
	Elements   []ElementMatcher
	Type       string
	CtxLines   int
	Delimiters string
}

func (e ElementsMatcher) MatchElement(line, flagKey string) bool {
	if e.Delimiters == "" && strings.Contains(line, flagKey) {
		return true
	}
	for _, left := range e.Delimiters {
		for _, right := range e.Delimiters {
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
