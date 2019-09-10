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
				if nextPageSize > 0 {
					return simplePaginatedSearch(cmd, flags, pageSize, ctxLines, delims)
				}
				return nil, NoSearchPatternErr
			}
			return nil, err
		}
		results = append(results, result...)
	}
	return results, nil
}

// paginatedSearch attempts to come up with a maximal set of flag keys to search for at a time dependent on the value set by safePaginationCharCount
func paginatedSearch(cmd command.Client, flags []string, maxSumFlagKeyLength, ctxLines int, delims []rune) ([][]string, error) {
	var results [][]string
	nextSearchKeys := []string{}
	totalKeyLength := len(delims) * 2
	from := 0
	for to, key := range flags {
		totalKeyLength += len(key) + strings.Count(key, ".")
		if totalKeyLength > maxSumFlagKeyLength {
			log.Debug.Printf("searching for flags in group: [%d, %d]", from, to)
			result, err := cmd.SearchForFlags(nextSearchKeys, ctxLines, delims)
			err = command.SearchTooLargeErr
			if err != nil {
				if err == command.SearchTooLargeErr {
					// this shouldn't be possible with a valid maxSumFlagKeyLength
					if len(nextSearchKeys) <= 0 {
						return nil, FatalPaginatedSearchErr
					}

					// if this fails unexpectedly, fallback to the simple search algorithm for the remaining pages
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
			nextSearchKeys = []string{}
			totalKeyLength = len(delims) * 2
			from = to + 1
		}
		nextSearchKeys = append(nextSearchKeys, key)
	}
	return results, nil
}
