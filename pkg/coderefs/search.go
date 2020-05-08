package coderefs

import (
	"errors"
	"strconv"

	"github.com/launchdarkly/ld-find-code-refs/internal/command"
	"github.com/launchdarkly/ld-find-code-refs/internal/log"
)

var NoSearchPatternErr = errors.New("failed to generate a valid search pattern")

type searchResultLine struct {
	Path     string
	LineNum  int
	LineText string
	FlagKeys map[string][]string
}

type searchResultLines []searchResultLine

func (lines searchResultLines) Len() int {
	return len(lines)
}

func (lines searchResultLines) Less(i, j int) bool {
	if lines[i].Path < lines[j].Path {
		return true
	}
	if lines[i].Path > lines[j].Path {
		return false
	}
	return lines[i].LineNum < lines[j].LineNum
}

func (lines searchResultLines) Swap(i, j int) {
	lines[i], lines[j] = lines[j], lines[i]
}

// paginatedSearch uses approximations to decide the number of flags to scan for at once using maxSumFlagKeyLength as an upper bound
func paginatedSearch(cmd command.Searcher, flags []string, maxSumFlagKeyLength, ctxLines int, delims []byte) ([][]string, error) {
	searchType := "flags"
	if delims == nil {
		searchType = "aliases"
	}

	if maxSumFlagKeyLength == 0 {
		return nil, NoSearchPatternErr
	}

	var results [][]string
	nextSearchKeys := []string{}

	totalKeyLength := command.DelimCost(delims)
	from := 0
	for to, key := range flags {
		totalKeyLength += command.FlagKeyCost(key)
		nextSearchKeys = append(nextSearchKeys, key)

		// if we've reached the end of the loop, or the current page has reached maximum length
		if to == len(flags)-1 || totalKeyLength+command.FlagKeyCost(flags[to+1]) > maxSumFlagKeyLength {

			log.Debug.Printf("searching for %s in group: [%d, %d]", searchType, from, to)
			result, err := cmd.SearchForFlags(nextSearchKeys, ctxLines, delims)
			if err != nil {
				if err == command.SearchTooLargeErr {
					// we expect all search implementations to complete successfully
					// if pagination fails unexpectedly, repeat the search with a smaller page size
					log.Debug.Printf("encountered an error paginating group [%d, %d], trying again with a lower page size", from, to)
					remainder, err := paginatedSearch(cmd, flags[from:], maxSumFlagKeyLength/2, ctxLines, delims)
					if err != nil {
						return nil, err
					}
					return append(results, remainder...), nil
				}
				return nil, err
			}

			results = append(results, result...)

			// loop bookkeeping
			nextSearchKeys = make([]string, 0, len(nextSearchKeys))
			totalKeyLength = command.DelimCost(delims)
			from = to + 1
		}
	}
	return results, nil
}

func findReferences(cmd command.Searcher, flags []string, aliases map[string][]string, ctxLines int, delimiters string) (searchResultLines, error) {
	log.Info.Printf("finding code references with delimiters: %s", delimiters)
	paginationCharCount := command.SafePaginationCharCount()
	results, err := paginatedSearch(cmd, flags, paginationCharCount, ctxLines, []byte(delimiters))
	if err != nil {
		return searchResultLines{}, err
	}
	flattenedAliases := make([]string, 0, len(flags))
	for _, flagAliases := range aliases {
		flattenedAliases = append(flattenedAliases, flagAliases...)
	}
	aliasResults, err := paginatedSearch(cmd, flattenedAliases, paginationCharCount, ctxLines, nil)
	if err != nil {
		return searchResultLines{}, err
	}
	results = append(results, aliasResults...)
	return generateReferences(aliases, results, ctxLines, delimiters), nil
}

func generateReferences(aliases map[string][]string, searchResult [][]string, ctxLines int, delims string) []searchResultLine {
	references := []searchResultLine{}

	for _, r := range searchResult {
		path := r[1]
		contextContainsFlagKey := r[2] == ":"
		lineNumber := r[3]
		lineText := r[4]
		lineNum, err := strconv.Atoi(lineNumber)
		if err != nil {
			log.Error.Fatalf("encountered an unexpected error generating flag references: %s", err)
		}
		ref := searchResultLine{Path: path, LineNum: lineNum}
		if contextContainsFlagKey {
			ref.FlagKeys = findReferencedFlags(lineText, aliases, delims)
		}
		if ctxLines >= 0 {
			ref.LineText = lineText
		}
		references = append(references, ref)
	}

	return references
}
