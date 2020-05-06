package coderefs

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/launchdarkly/ld-find-code-refs/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/internal/log"
)

// Since our hunking algorithm uses some maps, resulting slice orders are not deterministic
// We use these sorters to make sure the results are always in a deterministic order.
type byStartingLineNumber []ld.HunkRep

func (h byStartingLineNumber) Len() int      { return len(h) }
func (h byStartingLineNumber) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h byStartingLineNumber) Less(i, j int) bool {
	return h[i].StartingLineNumber < h[j].StartingLineNumber
}

func init() {
	log.Init(true)
}

func delimit(s string, delim string) string {
	return delim + s + delim
}

const (
	testFlagKey     = "someFlag"
	testFlagKey2    = "anotherFlag"
	testFlagAlias   = "some-flag"
	testFlagAlias2  = "some.flag"
	testFlag2Alias  = "another-flag"
	testFlag2Alias2 = "another.flag"
)

func TestMain(m *testing.M) {
	log.Init(true)
	os.Exit(m.Run())
}

var (
	firstFlag            = map[string][]string{testFlagKey: {}}
	firstFlagWithAlias   = map[string][]string{testFlagKey: {testFlagAlias}}
	firstFlagWithAliases = map[string][]string{testFlagKey: {testFlagAlias, testFlagAlias2}}
	secondFlag           = map[string][]string{testFlagKey2: {}}
	twoFlags             = map[string][]string{testFlagKey: {}, testFlagKey2: {}}
	twoFlagsWithAliases  = map[string][]string{testFlagKey: {testFlagAlias, testFlagAlias2}, testFlagKey2: {testFlag2Alias, testFlag2Alias2}}
	noFlags              = map[string][]string{}
)

func Test_generateReferences(t *testing.T) {
	testResult := []string{"", "flags.txt", ":", "12", delimit(testFlagKey, `"`)}
	testWant := searchResultLine{Path: "flags.txt", LineNum: 12, LineText: delimit(testFlagKey, `"`), FlagKeys: firstFlag}

	tests := []struct {
		name         string
		flags        map[string][]string
		searchResult [][]string
		ctxLines     int
		want         []searchResultLine
	}{
		{
			name:         "succeeds",
			flags:        twoFlags,
			searchResult: [][]string{testResult},
			ctxLines:     0,
			want:         []searchResultLine{testWant},
		},
		{
			name:         "succeeds with no LineText lines",
			flags:        twoFlags,
			searchResult: [][]string{testResult},
			ctxLines:     -1,
			want: []searchResultLine{
				{Path: "flags.txt", LineNum: 12, FlagKeys: firstFlag},
			},
		},
		{
			name:  "succeeds with multiple references",
			flags: twoFlags,
			searchResult: [][]string{
				testResult,
				{"", "path/flags.txt", ":", "12", `"someFlag" "anotherFlag"`},
			},
			ctxLines: 0,
			want: []searchResultLine{
				testWant,
				{Path: "path/flags.txt", LineNum: 12, LineText: `"someFlag" "anotherFlag"`, FlagKeys: twoFlags},
			},
		},
		{
			name:  "succeeds with aliases",
			flags: firstFlagWithAliases,
			searchResult: [][]string{
				{"", "path/flags.txt", ":", "12", testFlagAlias},
			},
			ctxLines: 0,
			want: []searchResultLine{
				{Path: "path/flags.txt", LineNum: 12, LineText: testFlagAlias, FlagKeys: firstFlagWithAlias},
			},
		},
		{
			name:  "succeeds with alias and flag key",
			flags: firstFlagWithAliases,
			searchResult: [][]string{
				{"", "path/flags.txt", ":", "12", delimit(testFlagKey, "'") + " " + testFlagAlias},
			},
			ctxLines: 0,
			want: []searchResultLine{
				{Path: "path/flags.txt", LineNum: 12, LineText: delimit(testFlagKey, "'") + " " + testFlagAlias, FlagKeys: firstFlagWithAlias},
			},
		},
		{
			name:  "succeeds with extra LineText lines",
			flags: twoFlags,
			searchResult: [][]string{
				{"", "flags.txt", "-", "11", "not a flag key line"},
				testResult,
				{"", "flags.txt", "-", "13", "not a flag key line"},
			},
			ctxLines: 1,
			want: []searchResultLine{
				{Path: "flags.txt", LineNum: 11, LineText: "not a flag key line"},
				testWant,
				{Path: "flags.txt", LineNum: 13, LineText: "not a flag key line"},
			},
		},
		{
			name:  "succeeds with extra LineText lines and multiple flags",
			flags: twoFlags,
			searchResult: [][]string{
				{"", "flags.txt", "-", "11", "not a flag key line"},
				testResult,
				{"", "flags.txt", "-", "13", "not a flag key line"},
				{"", "flags.txt", ":", "14", delimit(testFlagKey2, `"`)},
				{"", "flags.txt", "-", "15", "not a flag key line"},
			},
			ctxLines: 1,
			want: []searchResultLine{
				{Path: "flags.txt", LineNum: 11, LineText: "not a flag key line"},
				testWant,
				{Path: "flags.txt", LineNum: 13, LineText: "not a flag key line"},
				{Path: "flags.txt", LineNum: 14, LineText: delimit(testFlagKey2, `"`), FlagKeys: secondFlag},
				{Path: "flags.txt", LineNum: 15, LineText: "not a flag key line"},
			},
		},
		{
			name:         "does not match substring flag key",
			flags:        map[string][]string{testFlagKey: {}, testFlagKey2[:4]: {}},
			searchResult: [][]string{testResult},
			ctxLines:     0,
			want:         []searchResultLine{testWant},
		},
		{
			// delimeters don't have to match on both sides
			name:  "succeeds with multiple delimiters",
			flags: map[string][]string{testFlagKey: {}, "some": {}},
			searchResult: [][]string{
				{"", "flags.txt", ":", "12", `"` + testFlagKey + "'"},
			},
			ctxLines: 0,
			want: []searchResultLine{
				{Path: "flags.txt", LineNum: 12, LineText: `"` + testFlagKey + "'", FlagKeys: firstFlag},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, err)
			got := generateReferences(tt.flags, tt.searchResult, tt.ctxLines, `"'`)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_findReferencedFlags(t *testing.T) {
	tests := []struct {
		name string
		ref  string
		want map[string][]string
	}{
		{
			name: "finds a flag",
			ref:  "line contains " + delimit(testFlagKey, `"`),
			want: firstFlag,
		},
		{
			name: "finds multiple flags",
			ref:  "line contains " + delimit(testFlagKey, `"`) + " " + delimit(testFlagKey2, `"`),
			want: twoFlags,
		},
		{
			name: "finds no flags",
			ref:  "line contains no flags",
			want: noFlags,
		},
		{
			name: "finds one alias",
			ref:  fmt.Sprintf("line contains %s", testFlagAlias),
			want: firstFlagWithAlias,
		},
		{
			name: "finds all aliases",
			ref:  fmt.Sprintf("line contains %s %s %s %s %s %s", delimit(testFlagKey, `"`), delimit(testFlagKey2, `"`), testFlagAlias, testFlagAlias2, testFlag2Alias, testFlag2Alias2),
			want: twoFlagsWithAliases,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findReferencedFlags(tt.ref, twoFlagsWithAliases, `"`)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_makeReferenceHunksReps(t *testing.T) {
	projKey := "test"

	tests := []struct {
		name string
		refs searchResultLines
		want []ld.ReferenceHunksRep
	}{
		{
			name: "no references",
			refs: searchResultLines{},
			want: []ld.ReferenceHunksRep{},
		},
		{
			name: "single path, single reference with context lines",
			refs: searchResultLines{
				searchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				searchResultLine{
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
			refs: searchResultLines{
				searchResultLine{
					Path:     "a/b",
					LineNum:  1,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				searchResultLine{
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
			refs: searchResultLines{
				searchResultLine{
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
			refs: searchResultLines{
				searchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: testFlagAlias,
					FlagKeys: firstFlagWithAliases,
				},
				searchResultLine{
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
			got := tt.refs.makeReferenceHunksReps(projKey, 1)

			require.Equal(t, tt.want, got)
		})
	}
}

func Test_makeHunkReps(t *testing.T) {
	projKey := "test"

	tests := []struct {
		name     string
		ctxLines int
		refs     searchResultLines
		want     []ld.HunkRep
	}{
		{
			name:     "single reference with context lines",
			ctxLines: 1,
			refs: searchResultLines{
				searchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				searchResultLine{
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
			refs: searchResultLines{
				searchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context inner",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  8,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				searchResultLine{
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
			refs: searchResultLines{
				searchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context inner",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  8,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				searchResultLine{
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
			refs: searchResultLines{
				searchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "a context -1",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: "a " + testFlagKey,
					FlagKeys: firstFlag,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "a context +1",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  9,
					LineText: "b context -1",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  10,
					LineText: "b " + testFlagKey,
					FlagKeys: firstFlag,
				},
				searchResultLine{
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
			refs: searchResultLines{
				searchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context inner",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  8,
					LineText: testFlagKey2,
					FlagKeys: secondFlag,
				},
				searchResultLine{
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
			refs: searchResultLines{
				searchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "a context -1",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: "a " + testFlagKey,
					FlagKeys: firstFlag,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "a context +1",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  8,
					LineText: "b context -1",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  9,
					LineText: "b " + testFlagKey2,
					FlagKeys: secondFlag,
				},
				searchResultLine{
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
			refs: searchResultLines{
				searchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context inner",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  8,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				searchResultLine{
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
			refs: searchResultLines{
				searchResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context inner",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  8,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				searchResultLine{
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
			refs: searchResultLines{
				searchResultLine{
					Path:     "a/b",
					LineNum:  1,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  2,
					LineText: "context+1",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  3,
					LineText: "context+3",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  10,
					LineText: "context-1",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  11,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				searchResultLine{
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
			refs: searchResultLines{
				searchResultLine{
					Path:     "a/b",
					LineNum:  1,
					LineText: "context-1",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  2,
					LineText: testFlagKey,
					FlagKeys: firstFlag,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  3,
					LineText: "context+1",
					FlagKeys: noFlags,
				},
				searchResultLine{
					Path:     "a/b",
					LineNum:  4,
					LineText: "context+2",
					FlagKeys: noFlags,
				},
				searchResultLine{
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

func Test_groupIntoPathMap(t *testing.T) {
	searchResultPathALine1 := searchResultLine{
		Path:     "a",
		LineNum:  1,
		LineText: testFlagKey,
		FlagKeys: firstFlag,
	}

	searchResultPathALine2 := searchResultLine{
		Path:     "a",
		LineNum:  2,
		LineText: testFlagKey2,
		FlagKeys: secondFlag,
	}

	searchResultPathBLine1 := searchResultLine{
		Path:     "b",
		LineNum:  1,
		LineText: "flag-3",
		FlagKeys: map[string][]string{"flag-3": {}},
	}
	searchResultPathBLine2 := searchResultLine{
		Path:     "b",
		LineNum:  2,
		LineText: testFlagKey2,
		FlagKeys: map[string][]string{"flag-4": {}},
	}

	lines := searchResultLines{
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

func Test_filterShortFlags(t *testing.T) {
	// Note: these specs assume minFlagKeyLen is 3
	tests := []struct {
		name  string
		flags []string
		want  []string
	}{
		{
			name:  "Empty input/output",
			flags: []string{},
			want:  []string{},
		},
		{
			name:  "all flags are too short",
			flags: []string{"a", "b"},
			want:  []string{},
		},
		{
			name:  "some flags are too short",
			flags: []string{"abcdefg", "b", "ab", "abc"},
			want:  []string{"abcdefg", "abc"},
		},
		{
			name:  "no flags are too short",
			flags: []string{"catsarecool", "dogsarecool"},
			want:  []string{"catsarecool", "dogsarecool"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := filterShortFlagKeys(tt.flags)
			require.Equal(t, tt.want, got)
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

func Test_calculateStaleBranches(t *testing.T) {
	specs := []struct {
		name           string
		branches       []string
		remoteBranches []string
		expected       []string
	}{
		{
			name:           "stale branch",
			branches:       []string{"master", "another-branch"},
			remoteBranches: []string{"master"},
			expected:       []string{"another-branch"},
		},
		{
			name:           "no stale branches",
			branches:       []string{"master"},
			remoteBranches: []string{"master"},
			expected:       []string{},
		},
	}

	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			// transform test args into the format expected by calculateStaleBranches
			branchReps := make([]ld.BranchRep, 0, len(tt.branches))
			for _, b := range tt.branches {
				branchReps = append(branchReps, ld.BranchRep{Name: b})
			}
			remoteBranchMap := map[string]bool{}
			for _, b := range tt.remoteBranches {
				remoteBranchMap[b] = true
			}

			assert.ElementsMatch(t, tt.expected, calculateStaleBranches(branchReps, remoteBranchMap))
		})
	}
}
