package search

import (
	"strings"

	ahocorasick "github.com/petar-dambovaliev/aho-corasick"

	"github.com/launchdarkly/ld-find-code-refs/aliases"
	"github.com/launchdarkly/ld-find-code-refs/flags"
	"github.com/launchdarkly/ld-find-code-refs/internal/helpers"
	"github.com/launchdarkly/ld-find-code-refs/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	"github.com/launchdarkly/ld-find-code-refs/options"
)

type ElementMatcher struct {
	ProjKey  string
	Elements []string

	allElementAndAliasesMatcher ahocorasick.AhoCorasick
	matcherByElement            map[string]ahocorasick.AhoCorasick
	aliasMatcherByElement       map[string]ahocorasick.AhoCorasick
}

type Matcher struct {
	Elements []ElementMatcher
	ctxLines int
}

// Scan checks the configured directory for flags base on the options configured for Code References.
func Scan(opts options.Options, repoParams ld.RepoParams) (Matcher, []ld.ReferenceHunksRep) {
	flagKeys := flags.GetFlagKeys(opts, repoParams)
	aliasesByFlagKey, err := aliases.GenerateAliases(flagKeys, opts.Aliases, opts.Dir)
	if err != nil {
		log.Error.Fatalf("failed to generate aliases: %s", err)
	}
	delimiters := strings.Join(helpers.Dedupe(getDelimiters(opts)), "")
	flagMatcher := NewFlagMatcher(opts.ProjKey, delimiters, flagKeys, aliasesByFlagKey)

	matcher := Matcher{
		ctxLines: opts.ContextLines,
		Elements: []ElementMatcher{flagMatcher},
	}

	refs, err := SearchForRefs(opts.Dir, matcher)
	if err != nil {
		log.Error.Fatalf("error searching for flag key references: %s", err)
	}

	return matcher, refs
}

func NewFlagMatcher(projKey string, delimiters string, flagKeys []string, aliasesByFlagKey map[string][]string) ElementMatcher {
	matcherBuilder := ahocorasick.NewAhoCorasickBuilder(ahocorasick.Opts{DFA: true})

	var allFlagPatternsAndAliases []string

	patternsByFlag := buildFlagPatterns(flagKeys, delimiters)
	flagMatcherByKey := make(map[string]ahocorasick.AhoCorasick, len(patternsByFlag))
	for flagKey, patterns := range patternsByFlag {
		flagMatcherByKey[flagKey] = matcherBuilder.Build(patterns)
		allFlagPatternsAndAliases = append(allFlagPatternsAndAliases, patterns...)
	}

	aliasMatcherByFlagKey := make(map[string]ahocorasick.AhoCorasick, len(aliasesByFlagKey))
	for key, aliasesForFlag := range aliasesByFlagKey {
		aliasMatcherByFlagKey[key] = matcherBuilder.Build(aliasesForFlag)
		allFlagPatternsAndAliases = append(allFlagPatternsAndAliases, aliasesForFlag...)
	}

	return ElementMatcher{
		ProjKey:  projKey,
		Elements: flagKeys,

		matcherByElement:            flagMatcherByKey,
		aliasMatcherByElement:       aliasMatcherByFlagKey,
		allElementAndAliasesMatcher: matcherBuilder.Build(allFlagPatternsAndAliases),
	}
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
	for _, element := range m.Elements {
		if e, exists := element.matcherByElement[flagKey]; exists {
			if e.Iter(line).Next() != nil {
				return true
			}
		}
	}

	return false
}

func buildFlagPatterns(flags []string, delimiters string) map[string][]string {
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
