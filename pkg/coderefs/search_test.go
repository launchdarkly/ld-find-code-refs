package coderefs

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/launchdarkly/ld-find-code-refs/internal/command"
)

type MockClient struct {
	results [][]string
	err     error
	pages   [][]string
}

func (c *MockClient) SearchForFlags(flags []string, ctxLines int, delimiters []rune) ([][]string, error) {
	c.pages = append(c.pages, flags)
	return c.results, c.err
}

func Test_paginatedSearch(t *testing.T) {
	specs := []struct {
		name                string
		maxSumFlagKeyLength int
		mockResults         [][]string
		mockErr             error
		expectedResults     [][]string
		expectedPages       [][]string
		expectedErr         error
	}{
		{
			name:                "returns results with 1 page",
			mockResults:         [][]string{{"hello"}},
			expectedResults:     [][]string{{"hello"}},
			expectedPages:       [][]string{{"flag1", "flag2"}},
			maxSumFlagKeyLength: 12,
		},
		{
			name:                "combines results with multiple pages",
			mockResults:         [][]string{{"hello"}},
			expectedResults:     [][]string{{"hello"}, {"hello"}},
			expectedPages:       [][]string{{"flag1"}, {"flag2"}},
			maxSumFlagKeyLength: 7,
		},
		{
			name:    "pagination fails when client fails to generate a search pattern",
			mockErr: command.SearchTooLargeErr,
			// should try to recursively page 3 times and fail every time
			expectedPages:       [][]string{{"flag1"}, {"flag1"}, {"flag1"}},
			expectedErr:         NoSearchPatternErr,
			maxSumFlagKeyLength: 7,
		},
		{
			// this case should be impossible outside of tests
			name:                "pagination fails when maxSumFlagKeyLength is too low",
			mockErr:             command.SearchTooLargeErr,
			expectedErr:         NoSearchPatternErr,
			maxSumFlagKeyLength: 0,
		},
	}

	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			client := MockClient{
				results: tt.mockResults,
				err:     tt.mockErr,
			}

			res, err := paginatedSearch(
				&client,
				[]string{"flag1", "flag2"},
				tt.maxSumFlagKeyLength,
				0,
				[]rune{'"'},
			)
			assert.Equal(t, tt.expectedPages, client.pages)
			assert.Equal(t, tt.expectedResults, res)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func Test_sortSearchResults(t *testing.T) {
	cats1 := searchResultLine{Path: "/dev/null/cats", LineNum: 1, LineText: "", FlagKeys: map[string][]string{"src/meow/yes/pls": {}}}
	cats2 := searchResultLine{Path: "/dev/null/cats", LineNum: 2, LineText: "", FlagKeys: map[string][]string{"src/meow/feed/me": {}}}
	dogs5 := searchResultLine{Path: "/dev/null/dogs", LineNum: 5, LineText: "", FlagKeys: map[string][]string{"src/woof/oh/fine": {}}}
	dogs15 := searchResultLine{Path: "/dev/null/dogs", LineNum: 15, LineText: "", FlagKeys: map[string][]string{"src/woof/walk/me": {}}}

	linesToSort := searchResultLines{dogs15, cats2, dogs5, cats1}
	expectedResults := searchResultLines{cats1, cats2, dogs5, dogs15}

	sort.Sort(linesToSort)

	assert.Exactly(t, linesToSort, expectedResults, "search order for searchResultLines not as expected")
}
