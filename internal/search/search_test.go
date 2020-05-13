package search

import (
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/launchdarkly/ld-find-code-refs/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	log.Init(true)
}
func TestMain(m *testing.M) {
	log.Init(true)
	os.Exit(m.Run())
}

const (
	testFlagKey     = "someFlag"
	testFlagKey2    = "anotherFlag"
	testFlagAlias   = "some-flag"
	testFlagAlias2  = "some.flag"
	testFlag2Alias  = "another-flag"
	testFlag2Alias2 = "another.flag"
)

var (
	firstFlag            = map[string][]string{testFlagKey: {}}
	firstFlagWithAlias   = map[string][]string{testFlagKey: {testFlagAlias}}
	firstFlagWithAliases = map[string][]string{testFlagKey: {testFlagAlias, testFlagAlias2}}
	secondFlag           = map[string][]string{testFlagKey2: {}}
	twoFlags             = map[string][]string{testFlagKey: {}, testFlagKey2: {}}
	twoFlagsWithAliases  = map[string][]string{testFlagKey: {testFlagAlias, testFlagAlias2}, testFlagKey2: {testFlag2Alias, testFlag2Alias2}}
	noFlags              = map[string][]string{}
)

func delimit(s string, delim string) string {
	return delim + s + delim
}

// Since our hunking algorithm uses some maps, resulting slice orders are not deterministic
// We use these sorters to make sure the results are always in a deterministic order.
type byStartingLineNumber []ld.HunkRep

func (h byStartingLineNumber) Len() int      { return len(h) }
func (h byStartingLineNumber) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h byStartingLineNumber) Less(i, j int) bool {
	return h[i].StartingLineNumber < h[j].StartingLineNumber
}

var defaultDelims = []byte{'"', '\'', '`'}

type MockClient struct {
	results SearchResultLines
	err     error
	pages   [][]string
}

func (c *MockClient) searchForRefs(searchTerms []string, aliases map[string][]string, ctxLines int, delimiters []byte) (SearchResultLines, error) {
	c.pages = append(c.pages, searchTerms)
	return c.results, c.err
}

func Test_paginatedSearch(t *testing.T) {
	specs := []struct {
		name                string
		maxSumFlagKeyLength int
		mockResults         SearchResultLines
		mockErr             error
		expectedResults     SearchResultLines
		expectedPages       [][]string
		expectedErr         error
	}{
		{
			name:                "returns results with 1 page",
			mockResults:         SearchResultLines{{LineText: "hello"}},
			expectedResults:     SearchResultLines{{LineText: "hello"}},
			expectedPages:       [][]string{{"flag1", "flag2"}},
			maxSumFlagKeyLength: 12,
		},
		{
			name:                "combines results with multiple pages",
			mockResults:         SearchResultLines{{LineText: "hello"}},
			expectedResults:     SearchResultLines{{LineText: "hello"}, {LineText: "hello"}},
			expectedPages:       [][]string{{"flag1"}, {"flag2"}},
			maxSumFlagKeyLength: 7,
		},
		{
			name:    "pagination fails when client fails to generate a search pattern",
			mockErr: searchTooLargeErr,
			// should try to recursively page 3 times and fail every time
			expectedPages:       [][]string{{"flag1"}, {"flag1"}, {"flag1"}},
			expectedErr:         noSearchPatternErr,
			maxSumFlagKeyLength: 7,
		},
		{
			// this case should be impossible outside of tests
			name:                "pagination fails when maxSumFlagKeyLength is too low",
			mockErr:             searchTooLargeErr,
			expectedErr:         noSearchPatternErr,
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
				"flags",
				[]string{"flag1", "flag2"},
				// TODO: test with aliases
				map[string][]string{},
				tt.maxSumFlagKeyLength,
				0,
				[]byte{'"'},
			)
			assert.Equal(t, tt.expectedPages, client.pages)
			assert.Equal(t, tt.expectedResults, res)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func Test_sortSearchResults(t *testing.T) {
	cats1 := SearchResultLine{Path: "/dev/null/cats", LineNum: 1, LineText: "", FlagKeys: map[string][]string{"src/meow/yes/pls": {}}}
	cats2 := SearchResultLine{Path: "/dev/null/cats", LineNum: 2, LineText: "", FlagKeys: map[string][]string{"src/meow/feed/me": {}}}
	dogs5 := SearchResultLine{Path: "/dev/null/dogs", LineNum: 5, LineText: "", FlagKeys: map[string][]string{"src/woof/oh/fine": {}}}
	dogs15 := SearchResultLine{Path: "/dev/null/dogs", LineNum: 15, LineText: "", FlagKeys: map[string][]string{"src/woof/walk/me": {}}}

	linesToSort := SearchResultLines{dogs15, cats2, dogs5, cats1}
	expectedResults := SearchResultLines{cats1, cats2, dogs5, dogs15}

	sort.Sort(linesToSort)

	assert.Exactly(t, linesToSort, expectedResults, "search order for searchResultLines not as expected")
}

func Test_generateFlagRegex(t *testing.T) {
	specs := []struct {
		name     string
		flags    []string
		expected string
	}{
		{
			name:     "succeeds for single flag",
			flags:    []string{"flag"},
			expected: "flag",
		},
		{
			name:     "succeeds for single flag with escape char",
			flags:    []string{"^flag"},
			expected: `\^flag`,
		},
		{
			name:     "succeeds for multiple flags",
			flags:    []string{"flag1", "flag2"},
			expected: "flag1|flag2",
		},
		{
			name:     "succeeds for multiple flags with escape characters",
			flags:    []string{"^flag1", ".flag2", "*flag3"},
			expected: `\^flag1|\.flag2|\*flag3`,
		},
	}
	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, generateFlagRegex(tt.flags))
		})
	}
}

func Test_generateDelimiterRegex(t *testing.T) {
	specs := []struct {
		name               string
		delimiters         []byte
		expectedLookBehind string
		expectedLookAhead  string
	}{
		{
			name:               "succeeds for default delimiters",
			delimiters:         defaultDelims,
			expectedLookBehind: "(?<=[\"'`])",
			expectedLookAhead:  "(?=[\"'`])",
		},
		{
			name:               "succeeds for extra delimiter",
			delimiters:         append(defaultDelims, 'a'),
			expectedLookBehind: "(?<=[\"'`a])",
			expectedLookAhead:  "(?=[\"'`a])",
		},
	}
	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			lb, la := generateDelimiterRegex(tt.delimiters)
			assert.Equal(t, tt.expectedLookBehind, lb)
			assert.Equal(t, tt.expectedLookAhead, la)
		})
	}
}

func Test_generateSearchPattern(t *testing.T) {
	specs := []struct {
		name       string
		padPattern bool
		expected   string
	}{
		{
			name:       "correctly pads flag pattern",
			padPattern: true,
			expected:   "(?<=[\"'`])(a^|flag|a^)(?=[\"'`])",
		},
		{
			name:       "correctly doesn't pad pattern",
			padPattern: false,
			expected:   "(?<=[\"'`])(flag)(?=[\"'`])",
		},
	}
	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, generateSearchPattern([]string{"flag"}, defaultDelims, tt.padPattern))
		})
	}
}

func Test_truncateLine(t *testing.T) {
	longLine := strings.Repeat("a", maxLineCharCount)

	veryLongLine := strings.Repeat("a", maxLineCharCount+1)

	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "empty line",
			line: "",
			want: "",
		},
		{
			name: "line shorter than max length",
			line: "abc efg",
			want: "abc efg",
		},
		{
			name: "long line",
			line: longLine,
			want: longLine,
		},
		{
			name: "very long line",
			line: veryLongLine,
			want: veryLongLine[0:maxLineCharCount] + "…",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateLine(tt.line)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_groupIntoPathMap(t *testing.T) {
	searchResultPathALine1 := SearchResultLine{
		Path:     "a",
		LineNum:  1,
		LineText: testFlagKey,
		FlagKeys: firstFlag,
	}

	searchResultPathALine2 := SearchResultLine{
		Path:     "a",
		LineNum:  2,
		LineText: testFlagKey2,
		FlagKeys: secondFlag,
	}

	searchResultPathBLine1 := SearchResultLine{
		Path:     "b",
		LineNum:  1,
		LineText: "flag-3",
		FlagKeys: map[string][]string{"flag-3": {}},
	}
	searchResultPathBLine2 := SearchResultLine{
		Path:     "b",
		LineNum:  2,
		LineText: testFlagKey2,
		FlagKeys: map[string][]string{"flag-4": {}},
	}

	lines := SearchResultLines{
		searchResultPathALine1,
		searchResultPathALine2,
		searchResultPathBLine1,
		searchResultPathBLine2,
	}

	linesByPath := lines.aggregateByPath()

	aRefs := linesByPath[0]
	require.Equal(t, aRefs.path, "a")

	aRefMap := aRefs.flagReferenceMap
	require.Equal(t, len(aRefMap), 2)

	require.Contains(t, aRefMap, testFlagKey)
	require.Contains(t, aRefMap, testFlagKey2)

	aLines := aRefs.fileSearchResultLines
	require.Equal(t, aLines.Len(), 2)
	require.Equal(t, aLines.Front().Value, searchResultPathALine1)
	require.Equal(t, aLines.Back().Value, searchResultPathALine2)

	bRefs := linesByPath[1]
	require.Equal(t, bRefs.path, "b")

	bRefMap := bRefs.flagReferenceMap
	require.Equal(t, len(aRefMap), 2)

	require.Contains(t, bRefMap, "flag-3")
	require.Contains(t, bRefMap, "flag-4")

	bLines := bRefs.fileSearchResultLines
	require.Equal(t, bLines.Len(), 2)
	require.Equal(t, bLines.Front().Value, searchResultPathBLine1)
	require.Equal(t, bLines.Back().Value, searchResultPathBLine2)
}

func Test_makeHunkReps(t *testing.T) {
	projKey := "test"

	tests := []struct {
		name     string
		ctxLines int
		refs     SearchResultLines
		want     []ld.HunkRep
	}{
		{
			name:     "single reference with context lines",
			ctxLines: 1,
			refs: SearchResultLines{
				SearchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context +1",
					FlagKeys: noFlags,
				},
			},
			want: []ld.HunkRep{
				{
					StartingLineNumber: 5,
					Lines:              "context -1\n" + testFlagKey + "\ncontext +1\n",
					ProjKey:            projKey,
					FlagKey:            testFlagKey,
					Aliases:            []string{},
				},
			},
		},
		{
			name:     "multiple references, single flag, one hunk",
			ctxLines: 1,
			refs: SearchResultLines{
				SearchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context inner",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  8,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  9,
					LineText: "context +1",
					FlagKeys: noFlags,
				},
			},
			want: []ld.HunkRep{
				{
					StartingLineNumber: 5,
					Lines:              "context -1\n" + testFlagKey + "\ncontext inner\n" + testFlagKey + "\ncontext +1\n",
					ProjKey:            projKey,
					FlagKey:            testFlagKey,
					Aliases:            []string{},
				},
			},
		},
		{
			name:     "multiple references, multiple context lines, single flag, one hunk",
			ctxLines: 2,
			refs: SearchResultLines{
				SearchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context inner",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  8,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  9,
					LineText: "context +1",
					FlagKeys: noFlags,
				},
			},
			want: []ld.HunkRep{
				{
					StartingLineNumber: 5,
					Lines:              "context -1\n" + testFlagKey + "\ncontext inner\n" + testFlagKey + "\ncontext +1\n",
					ProjKey:            projKey,
					FlagKey:            testFlagKey,
					Aliases:            []string{},
				},
			},
		},
		{
			name:     "multiple references, single flag, multiple hunks",
			ctxLines: 1,
			refs: SearchResultLines{
				SearchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "a context -1",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: "a " + testFlagKey,
					FlagKeys: firstFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "a context +1",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  9,
					LineText: "b context -1",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  10,
					LineText: "b " + testFlagKey,
					FlagKeys: firstFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  11,
					LineText: "b context +1",
					FlagKeys: noFlags,
				},
			},
			want: []ld.HunkRep{
				{
					StartingLineNumber: 5,
					Lines:              "a context -1\na " + testFlagKey + "\na context +1\n",
					ProjKey:            projKey,
					FlagKey:            testFlagKey,
					Aliases:            []string{},
				},
				{
					StartingLineNumber: 9,
					Lines:              "b context -1\nb " + testFlagKey + "\nb context +1\n",
					ProjKey:            projKey,
					FlagKey:            testFlagKey,
					Aliases:            []string{},
				},
			},
		},
		{
			name:     "multiple consecutive references, multiple flags, multiple hunks",
			ctxLines: 1,
			refs: SearchResultLines{
				SearchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context inner",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  8,
					LineText: testFlagKey2,
					FlagKeys: secondFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  9,
					LineText: "context +1",
					FlagKeys: noFlags,
				},
			},
			want: []ld.HunkRep{
				{
					StartingLineNumber: 5,
					Lines:              "context -1\n" + testFlagKey + "\ncontext inner\n",
					ProjKey:            projKey,
					FlagKey:            testFlagKey,
					Aliases:            []string{},
				},
				{
					StartingLineNumber: 7,
					Lines:              "context inner\n" + testFlagKey2 + "\ncontext +1\n",
					ProjKey:            projKey,
					FlagKey:            testFlagKey2,
					Aliases:            []string{},
				},
			},
		},
		{
			name:     "multiple consecutive (non overlapping) references, multiple flags, multiple hunks",
			ctxLines: 1,
			refs: SearchResultLines{
				SearchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "a context -1",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: "a " + testFlagKey,
					FlagKeys: firstFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "a context +1",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  8,
					LineText: "b context -1",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  9,
					LineText: "b " + testFlagKey2,
					FlagKeys: secondFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  10,
					LineText: "b context +1",
					FlagKeys: noFlags,
				},
			},
			want: []ld.HunkRep{
				{
					StartingLineNumber: 5,
					Lines:              "a context -1\na " + testFlagKey + "\na context +1\n",
					ProjKey:            projKey,
					FlagKey:            testFlagKey,
					Aliases:            []string{},
				},
				{
					StartingLineNumber: 8,
					Lines:              "b context -1\nb " + testFlagKey2 + "\nb context +1\n",
					ProjKey:            projKey,
					FlagKey:            testFlagKey2,
					Aliases:            []string{},
				},
			},
		},
		{
			name:     "multiple references, single flag, 0 context, multiple hunks",
			ctxLines: 0,
			refs: SearchResultLines{
				SearchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context inner",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  8,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  9,
					LineText: "context +1",
					FlagKeys: noFlags,
				},
			},
			want: []ld.HunkRep{
				{
					StartingLineNumber: 6,
					Lines:              testFlagKey + "\n",
					ProjKey:            projKey,
					FlagKey:            testFlagKey,
					Aliases:            []string{},
				},
				{
					StartingLineNumber: 8,
					Lines:              testFlagKey + "\n",
					ProjKey:            projKey,
					FlagKey:            testFlagKey,
					Aliases:            []string{},
				},
			},
		},
		{
			name:     "multiple references, single flag, negative context, multiple hunks",
			ctxLines: -1,
			refs: SearchResultLines{
				SearchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context inner",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  8,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  9,
					LineText: "context +1",
					FlagKeys: noFlags,
				},
			},
			want: []ld.HunkRep{
				{
					StartingLineNumber: 6,
					Lines:              "",
					ProjKey:            projKey,
					FlagKey:            testFlagKey,
					Aliases:            []string{},
				},
				{
					StartingLineNumber: 8,
					Lines:              "",
					ProjKey:            projKey,
					FlagKey:            testFlagKey,
					Aliases:            []string{},
				},
			},
		},
		{
			// This case verifies an edge case we found where the hunking algorithm
			// would walk past the end of the context for a given flag under certain circumstances.
			// This would happen when the actual flag reference happened at the beginning of the file (so
			// the algorithm couldn't walk ctxLines times backwards. it would naively walk ctxLines*2+1
			// times forwards, which would walk past the correct end of the hunk and into the next hunk
			name:     "multiple references, first reference at start of file",
			ctxLines: 1,
			refs: SearchResultLines{
				SearchResultLine{
					Path:     "a/b",
					LineNum:  1,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  2,
					LineText: "context+1",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  3,
					LineText: "context+3",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  10,
					LineText: "context-1",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  11,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  12,
					LineText: "context+1",
					FlagKeys: noFlags,
				},
			},
			want: []ld.HunkRep{
				{
					StartingLineNumber: 1,
					Lines:              testFlagKey + "\ncontext+1\n",
					ProjKey:            projKey,
					FlagKey:            testFlagKey,
					Aliases:            []string{},
				},
				{
					StartingLineNumber: 10,
					Lines:              "context-1\n" + testFlagKey + "\ncontext+1\n",
					ProjKey:            projKey,
					FlagKey:            testFlagKey,
					Aliases:            []string{},
				},
			},
		},
		{
			// This is another test case guarding against the bug described in the
			// previous test case
			name:     "multiple references, first reference at start of file",
			ctxLines: 2,
			refs: SearchResultLines{
				SearchResultLine{
					Path:     "a/b",
					LineNum:  1,
					LineText: "context-1",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  2,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  3,
					LineText: "context+1",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  4,
					LineText: "context+2",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  10,
					LineText: "context+alot+shouldn'tbeinhunk",
					FlagKeys: noFlags,
				},
			},
			want: []ld.HunkRep{
				{
					StartingLineNumber: 1,
					Lines:              "context-1\n" + testFlagKey + "\ncontext+1\ncontext+2\n",
					ProjKey:            projKey,
					FlagKey:            testFlagKey,
					Aliases:            []string{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupedResults := tt.refs.aggregateByPath()

			require.Equal(t, len(groupedResults), 1)

			fileSearchResults := groupedResults[0]

			got := fileSearchResults.makeHunkReps(projKey, tt.ctxLines)

			sort.Sort(byStartingLineNumber(got))

			require.Equal(t, tt.want, got)
		})
	}
}

func TestMakeReferenceHunksReps(t *testing.T) {
	projKey := "test"

	tests := []struct {
		name string
		refs SearchResultLines
		want []ld.ReferenceHunksRep
	}{
		{
			name: "no references",
			refs: SearchResultLines{},
			want: []ld.ReferenceHunksRep{},
		},
		{
			name: "single path, single reference with context lines",
			refs: SearchResultLines{
				SearchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: noFlags,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context +1",
					FlagKeys: noFlags,
				},
			},
			want: []ld.ReferenceHunksRep{
				{
					Path: "a/b",
					Hunks: []ld.HunkRep{
						{
							StartingLineNumber: 5,
							Lines:              "context -1\n" + testFlagKey + "\ncontext +1\n",
							ProjKey:            projKey,
							FlagKey:            testFlagKey,
							Aliases:            []string{},
						},
					},
				},
			},
		},
		{
			name: "multiple paths, single reference with context lines",
			refs: SearchResultLines{
				SearchResultLine{
					Path:     "a/b",
					LineNum:  1,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				SearchResultLine{
					Path:     "a/c/d",
					LineNum:  10,
					LineText: testFlagKey2,
					FlagKeys: secondFlag,
				},
			},
			want: []ld.ReferenceHunksRep{
				{
					Path: "a/b",
					Hunks: []ld.HunkRep{
						{
							StartingLineNumber: 1,
							Lines:              testFlagKey + "\n",
							ProjKey:            projKey,
							FlagKey:            testFlagKey,
							Aliases:            []string{},
						},
					},
				},
				{
					Path: "a/c/d",
					Hunks: []ld.HunkRep{
						{
							StartingLineNumber: 10,
							Lines:              testFlagKey2 + "\n",
							ProjKey:            projKey,
							FlagKey:            testFlagKey2,
							Aliases:            []string{},
						},
					},
				},
			},
		},
		{
			name: "alias",
			refs: SearchResultLines{
				SearchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: testFlagAlias,
					FlagKeys: firstFlagWithAlias,
				},
			},
			want: []ld.ReferenceHunksRep{
				{
					Path: "a/b",
					Hunks: []ld.HunkRep{
						{
							StartingLineNumber: 6,
							Lines:              testFlagAlias + "\n",
							ProjKey:            projKey,
							FlagKey:            testFlagKey,
							Aliases:            []string{testFlagAlias},
						},
					},
				},
			},
		},
		{
			name: "aliases",
			refs: SearchResultLines{
				SearchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: testFlagAlias,
					FlagKeys: firstFlagWithAliases,
				},
				SearchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: testFlagAlias2,
					FlagKeys: firstFlagWithAliases,
				},
			},
			want: []ld.ReferenceHunksRep{
				{
					Path: "a/b",
					Hunks: []ld.HunkRep{
						{
							StartingLineNumber: 5,
							Lines:              testFlagAlias + "\n" + testFlagAlias2 + "\n",
							ProjKey:            projKey,
							FlagKey:            testFlagKey,
							Aliases:            []string{testFlagAlias, testFlagAlias2},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.refs.MakeReferenceHunksReps(projKey, 1)

			require.Equal(t, tt.want, got)
		})
	}
}
