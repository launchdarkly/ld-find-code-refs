package search

import (
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/helpers"
	ahocorasick "github.com/petar-dambovaliev/aho-corasick"
)

type ElementMatcher struct {
	ProjKey                     string
	Elements                    []string
	Dir                         string
	allElementAndAliasesMatcher ahocorasick.AhoCorasick
	matcherByElement            map[string]ahocorasick.AhoCorasick
	aliasMatcherByElement       map[string]ahocorasick.AhoCorasick

	elementsByPatternIndex [][]string
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

func NewElementMatcher(projKey, dir, delimiters string, elements []string, aliasesByElement map[string][]string) ElementMatcher {
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
		Elements:                    elements,
		ProjKey:                     projKey,
		Dir:                         dir,
		matcherByElement:            flagMatcherByKey,
		aliasMatcherByElement:       aliasMatcherByElement,
		allElementAndAliasesMatcher: matcherBuilder.Build(allFlagPatternsAndAliases),

		elementsByPatternIndex: elementsByPatternIndex,
	}
}
