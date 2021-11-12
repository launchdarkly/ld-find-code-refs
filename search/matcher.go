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

	elementsByPatternIndex [][]string
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
	flagMatcher := NewElementMatcher(opts.ProjKey, delimiters, flagKeys, aliasesByFlagKey)

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

func NewElementMatcher(projKey string, delimiters string, elements []string, aliasesByElement map[string][]string) ElementMatcher {
	matcherBuilder := ahocorasick.NewAhoCorasickBuilder(ahocorasick.Opts{DFA: true, MatchKind: ahocorasick.StandardMatch})

	allFlagPatternsAndAliases := make([]string, 0)
	elementsByPatternIndex := make([][]string, 0)
	patternIndex := make(map[string]int)

	recordPatternsForElement := func(element string, patterns []string) {
		for _, p := range patterns {
			index, exists := patternIndex[p]
			if !exists {
				allFlagPatternsAndAliases = append(allFlagPatternsAndAliases, p)
				index = len(elementsByPatternIndex)
				elementsByPatternIndex = append(elementsByPatternIndex, []string{})
			}
			patternIndex[p] = index
			elementsByPatternIndex[index] = append(elementsByPatternIndex[index], element)
		}
	}

	patternsByElement := buildElementPatterns(elements, delimiters)
	flagMatcherByKey := make(map[string]ahocorasick.AhoCorasick, len(patternsByElement))
	for element, patterns := range patternsByElement {
		flagMatcherByKey[element] = matcherBuilder.Build(patterns)
		recordPatternsForElement(element, patterns)
	}

	aliasMatcherByElement := make(map[string]ahocorasick.AhoCorasick, len(aliasesByElement))
	for element, elementAliases := range aliasesByElement {
		aliasMatcherByElement[element] = matcherBuilder.Build(elementAliases)
		recordPatternsForElement(element, elementAliases)
	}

	return ElementMatcher{
		ProjKey:  projKey,
		Elements: elements,

		matcherByElement:            flagMatcherByKey,
		aliasMatcherByElement:       aliasMatcherByElement,
		allElementAndAliasesMatcher: matcherBuilder.Build(allFlagPatternsAndAliases),

		elementsByPatternIndex: elementsByPatternIndex,
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

func (m Matcher) FindAliases(line, element string) []string {
	matches := make([]string, 0)
	for _, em := range m.Elements {
		matches = append(matches, em.FindAliases(line, element)...)
	}
	return helpers.Dedupe(matches)
}

func (m ElementMatcher) FindMatches(line string) []string {
	elements := make([]string, 0)
	iter := m.allElementAndAliasesMatcher.IterOverlapping(line)
	for match := iter.Next(); match != nil; match = iter.Next() {
		elements = append(elements, m.elementsByPatternIndex[match.Pattern()]...)
	}
	return helpers.Dedupe(elements)
}

func (m ElementMatcher) FindAliases(line, element string) []string {
	aliasMatches := make([]string, 0)
	if aliasMatcher, exists := m.aliasMatcherByElement[element]; exists {
		iter := aliasMatcher.IterOverlapping(line)
		for match := iter.Next(); match != nil; match = iter.Next() {
			aliasMatches = append(aliasMatches, line[match.Start():match.End()])
		}
	}
	return aliasMatches
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
