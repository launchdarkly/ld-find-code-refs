package coderefs

import (
	"errors"

	"github.com/launchdarkly/ld-find-code-refs/internal/command"
	"github.com/launchdarkly/ld-find-code-refs/internal/log"
)

var NoSearchPatternErr = errors.New("failed to generate a valid search pattern")

// paginatedSearch uses approximations to decide the number of flags to scan for at once using maxSumFlagKeyLength as an upper bound
func paginatedSearch(cmd command.Searcher, flags []string, maxSumFlagKeyLength, ctxLines int, delims []rune) ([][]string, error) {
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
			log.Debug.Printf("searching for flags in group: [%d, %d]", from, to)
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
