package coderefs

import (
	"errors"
	"strings"

	"github.com/launchdarkly/ld-find-code-refs/internal/command"
	"github.com/launchdarkly/ld-find-code-refs/internal/log"
)

var (
	NoSearchPatternErr      = errors.New("failed to generate a valid search pattern")
	FatalPaginatedSearchErr = errors.New("encountered a fatal error attempting to paginate")
)

// simplePaginatedSearch attempts to search by reducing pageSize until a valid search is generated
func simplePaginatedSearch(cmd command.Client, flags []string, pageSize, ctxLines int, delims []rune) ([][]string, error) {
	var results [][]string
	// base case: cannot search for 0 flags at a time
	if pageSize <= 0 {
		return nil, NoSearchPatternErr
	}

	for from := 0; from < len(flags); from += pageSize {
		to := from + pageSize
		if to > len(flags) {
			to = len(flags)
		}

		log.Debug.Printf("searching for flags in group: [%d, %d]", from, to)
		result, err := cmd.SearchForFlags(flags[from:to], ctxLines, delims)
		if err != nil {
			if err == command.SearchTooLargeErr {
				nextPageSize := pageSize / 2
				log.Debug.Printf("encountered an error paginating at pagesize %d, attempting with a pagesize of %d", pageSize, nextPageSize)
				return simplePaginatedSearch(cmd, flags, nextPageSize, ctxLines, delims)
			}
			return nil, err
		}
		results = append(results, result...)
	}
	return results, nil
}

// paginatedSearch uses approximations to decide the number of flags to scan for at once using maxSumFlagKeyLength as an upper bound
func paginatedSearch(cmd command.Client, flags []string, maxSumFlagKeyLength, ctxLines int, delims []rune) ([][]string, error) {
	var results [][]string
	nextSearchKeys := []string{}

	// The approximate length of delimeters when added to the beginning and end of a search pattern
	totalKeyLength := len(delims) * 2
	from := 0
	for to, key := range flags {
		// periods need to be escaped, so they count as 2 characters
		totalKeyLength += len(key) + strings.Count(key, ".")
		if totalKeyLength > maxSumFlagKeyLength {
			log.Debug.Printf("searching for flags in group: [%d, %d]", from, to)
			result, err := cmd.SearchForFlags(nextSearchKeys, ctxLines, delims)
			if err != nil {
				if err == command.SearchTooLargeErr {
					// this shouldn't be possible with a valid maxSumFlagKeyLength
					if len(nextSearchKeys) == 0 {
						return nil, FatalPaginatedSearchErr
					}

					// we expect all search implementations to complete successfully
					// if pagination fails unexpectedly, fallback to a simple search algorithm for the remaining pages
					log.Debug.Printf("encountered an error paginating, falling back to simple paginated search")
					remainder, err := simplePaginatedSearch(cmd, flags[from:], len(nextSearchKeys)-1, ctxLines, delims)
					if err != nil {
						return nil, err
					}
					return append(results, remainder...), nil
				}
				return nil, err
			}
			results = append(results, result...)
			// capacity of the next search is likely similar to the previous search
			nextSearchKeys = make([]string, 0, len(nextSearchKeys))
			totalKeyLength = len(delims) * 2
			from = to + 1
		}
		nextSearchKeys = append(nextSearchKeys, key)
	}
	return results, nil
}
