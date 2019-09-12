package coderefs

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/launchdarkly/ld-find-code-refs/internal/command"
)

type MockClient struct {
	results [][]string
	err     error
}

func (c MockClient) RemoteBranches() (map[string]bool, error) {
	return nil, errors.New("Mock error")
}

func (c MockClient) SearchForFlags(flags []string, ctxLines int, delimiters []rune) ([][]string, error) {
	return c.results, c.err
}

func Test_paginatedSearch(t *testing.T) {
	specs := []struct {
		name                string
		maxSumFlagKeyLength int
		mockResults         [][]string
		mockErr             error
		expectedResults     [][]string
		expectedErr         error
	}{
		{
			name:                "returns results with 1 page",
			mockResults:         [][]string{{"hello"}},
			expectedResults:     [][]string{{"hello"}},
			maxSumFlagKeyLength: 20,
		},
		{
			name:                "combines results with multiple pages",
			mockResults:         [][]string{{"hello"}},
			expectedResults:     [][]string{{"hello"}, {"hello"}},
			maxSumFlagKeyLength: 5,
		},
		{
			name:                "pagination fails when client fails to generate a search pattern",
			mockErr:             command.SearchTooLargeErr,
			expectedErr:         NoSearchPatternErr,
			maxSumFlagKeyLength: 10,
		},
		{
			// this case should be impossible outside of tests
			name:                "pagination fails when maxSumFlagKeyLength is too low",
			mockErr:             command.SearchTooLargeErr,
			expectedErr:         FatalPaginatedSearchErr,
			maxSumFlagKeyLength: 0,
		},
	}

	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			res, err := paginatedSearch(
				MockClient{
					results: tt.mockResults,
					err:     tt.mockErr,
				},
				[]string{testFlagKey, testFlagKey2},
				tt.maxSumFlagKeyLength,
				0,
				[]rune{'"'},
			)
			assert.Equal(t, tt.expectedResults, res)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func Test_simplePaginatedSearch(t *testing.T) {
	specs := []struct {
		name             string
		startingPageSize int
		mockResults      [][]string
		mockErr          error
		expectedResults  [][]string
		expectedErr      error
	}{
		{
			name:             "returns results with 1 page",
			startingPageSize: 2,
			mockResults:      [][]string{{"hello"}},
			expectedResults:  [][]string{{"hello"}},
		},
		{
			name:             "combines results with multiple pages",
			startingPageSize: 1,
			mockResults:      [][]string{{"hello"}},
			expectedResults:  [][]string{{"hello"}, {"hello"}},
		},
		{
			name:             "pagination fails when client fails to generate a search pattern",
			mockErr:          command.SearchTooLargeErr,
			expectedErr:      NoSearchPatternErr,
			startingPageSize: 2,
		},
		{
			name:             "pagination fails when pageSize is too low",
			expectedErr:      NoSearchPatternErr,
			startingPageSize: -1,
		},
	}

	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			res, err := simplePaginatedSearch(
				MockClient{
					results: tt.mockResults,
					err:     tt.mockErr,
				},
				[]string{"someFlag", "anotherFlag"},
				tt.startingPageSize,
				0,
				[]rune{'"'},
			)
			assert.Equal(t, tt.expectedResults, res)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
