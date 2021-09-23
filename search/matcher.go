package search

import (
	"strings"

	"github.com/launchdarkly/ld-find-code-refs/flags"
	"github.com/launchdarkly/ld-find-code-refs/internal/helpers"
	"github.com/launchdarkly/ld-find-code-refs/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	"github.com/launchdarkly/ld-find-code-refs/options"
)

type ElementMatcher struct {
	Elements       []string
	Aliases        map[string][]string
	Delimiters     []string
	ProjKey        string
	Directory      string
	DelimitedFlags map[string][]string
}

type Matcher struct {
	Elements   []ElementMatcher
	CtxLines   int
	Delimiters string
}

// Scan checks the configured directory for flags base on the options configured for Code References.
func Scan(opts options.Options, repoParams ld.RepoParams) (Matcher, []ld.ReferenceHunksRep) {
	flagMatcher := ElementMatcher{
		Directory: opts.Dir,
	}
	flagMatcher.Elements, flagMatcher.Aliases = flags.GenerateSearchElements(opts, repoParams)

	matcher := Matcher{
		CtxLines: opts.ContextLines,
	}

	// Configure delimiters
	delims := getDelimiters(opts)
	matcher.Delimiters = strings.Join(helpers.Dedupe(delims), "")
	flagMatcher.DelimitedFlags = buildDelimiterList(flagMatcher.Elements, matcher.Delimiters)
	// Begin search for elements.
	matcher.Elements = []ElementMatcher{flagMatcher}

	refs, err := SearchForRefs(matcher)
	if err != nil {
		log.Error.Fatalf("error searching for flag key references: %s", err)
	}

	return matcher, refs
}

func getDelimiters(opts options.Options) []string {
	delims := []string{`"`, `'`, "`"}
	if opts.Delimiters.DisableDefaults {
		delims = []string{}
	}

	delims = append(delims, opts.Delimiters.Additional...)

	return delims
}

func (m Matcher) MatchElement(line, flagKey string) bool {
	if m.Delimiters == "" && strings.Contains(line, flagKey) {
		return true
	}

	for _, element := range m.Elements {
		delimitedFlags := element.DelimitedFlags[flagKey]
		for _, delimitedflagKey := range delimitedFlags {
			if strings.Contains(line, delimitedflagKey) {
				return true
			}
		}
	}

	return false
}

func buildDelimiterList(flags []string, delimiters string) map[string][]string {
	delimiterMap := make(map[string][]string)
	if delimiters == "" {
		return delimiterMap
	}
	for _, flag := range flags {
		tempFlags := []string{}
		for _, left := range delimiters {
			for _, right := range delimiters {
				var sb strings.Builder
				sb.Grow(len(flag) + 2)
				sb.WriteRune(left)
				sb.WriteString(flag)
				sb.WriteRune(right)
				tempFlags = append(tempFlags, sb.String())
			}
		}
		delimiterMap[flag] = tempFlags
	}
	return delimiterMap
}
