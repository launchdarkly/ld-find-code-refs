package coderefs

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/launchdarkly/ld-find-code-refs/internal/ld"
)

// Since our hunking algorithm uses some maps, resulting slice orders are not deterministic
// We use these sorters to make sure the results are always in a deterministic order.
type byStartingLineNumber []ld.HunkRep

func (h byStartingLineNumber) Len() int      { return len(h) }
func (h byStartingLineNumber) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h byStartingLineNumber) Less(i, j int) bool {
	return h[i].StartingLineNumber < h[j].StartingLineNumber
}

func Test_generateReferencesFromGrep(t *testing.T) {
	tests := []struct {
		name       string
		flags      []string
		grepResult [][]string
		ctxLines   int
		want       []grepResultLine
		exclude    string
	}{
		{
			name:  "succeeds",
			flags: []string{"someFlag", "anotherFlag"},
			grepResult: [][]string{
				{"", "flags.txt", ":", "12", "someFlag"},
			},
			ctxLines: 0,
			want: []grepResultLine{
				{Path: "flags.txt", LineNum: 12, LineText: "someFlag", FlagKeys: []string{"someFlag"}},
			},
		},
		{
			name:  "succeeds with exclude",
			flags: []string{"someFlag", "anotherFlag"},
			grepResult: [][]string{
				{"", "flags.txt", ":", "12", "someFlag"},
			},
			ctxLines: 0,
			want:     []grepResultLine{},
			exclude:  ".*",
		},
		{
			name:  "succeeds with no LineText lines",
			flags: []string{"someFlag", "anotherFlag"},
			grepResult: [][]string{
				{"", "flags.txt", ":", "12", "someFlag"},
			},
			ctxLines: -1,
			want: []grepResultLine{
				{Path: "flags.txt", LineNum: 12, FlagKeys: []string{"someFlag"}},
			},
		},
		{
			name:  "succeeds with multiple references",
			flags: []string{"someFlag", "anotherFlag"},
			grepResult: [][]string{
				{"", "flags.txt", ":", "12", "someFlag"},
				{"", "path/flags.txt", ":", "12", "someFlag anotherFlag"},
			},
			ctxLines: 0,
			want: []grepResultLine{
				{Path: "flags.txt", LineNum: 12, LineText: "someFlag", FlagKeys: []string{"someFlag"}},
				{Path: "path/flags.txt", LineNum: 12, LineText: "someFlag anotherFlag", FlagKeys: []string{"someFlag", "anotherFlag"}},
			},
		},
		{
			name:  "succeeds with extra LineText lines",
			flags: []string{"someFlag", "anotherFlag"},
			grepResult: [][]string{
				{"", "flags.txt", "-", "11", "not a flag key line"},
				{"", "flags.txt", ":", "12", "someFlag"},
				{"", "flags.txt", "-", "13", "not a flag key line"},
			},
			ctxLines: 1,
			want: []grepResultLine{
				{Path: "flags.txt", LineNum: 11, LineText: "not a flag key line"},
				{Path: "flags.txt", LineNum: 12, LineText: "someFlag", FlagKeys: []string{"someFlag"}},
				{Path: "flags.txt", LineNum: 13, LineText: "not a flag key line"},
			},
		},
		{
			name:  "succeeds with extra LineText lines and multiple flags",
			flags: []string{"someFlag", "anotherFlag"},
			grepResult: [][]string{
				{"", "flags.txt", "-", "11", "not a flag key line"},
				{"", "flags.txt", ":", "12", "someFlag"},
				{"", "flags.txt", "-", "13", "not a flag key line"},
				{"", "flags.txt", ":", "14", "anotherFlag"},
				{"", "flags.txt", "-", "15", "not a flag key line"},
			},
			ctxLines: 1,
			want: []grepResultLine{
				{Path: "flags.txt", LineNum: 11, LineText: "not a flag key line"},
				{Path: "flags.txt", LineNum: 12, LineText: "someFlag", FlagKeys: []string{"someFlag"}},
				{Path: "flags.txt", LineNum: 13, LineText: "not a flag key line"},
				{Path: "flags.txt", LineNum: 14, LineText: "anotherFlag", FlagKeys: []string{"anotherFlag"}},
				{Path: "flags.txt", LineNum: 15, LineText: "not a flag key line"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ex, err := regexp.Compile(tt.exclude)
			require.NoError(t, err)
			got := generateReferencesFromGrep(tt.flags, tt.grepResult, tt.ctxLines, ex)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_findReferencedFlags(t *testing.T) {
	tests := []struct {
		name string
		ref  string
		want []string
	}{
		{
			name: "finds a flag",
			ref:  "line contains someFlag",
			want: []string{"someFlag"},
		},
		{
			name: "finds multiple flags",
			ref:  "line contains someFlag and anotherFlag",
			want: []string{"someFlag", "anotherFlag"},
		},
		{
			name: "finds no flags",
			ref:  "line contains no flags",
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findReferencedFlags(tt.ref, []string{"someFlag", "anotherFlag"})
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_makeReferenceHunksReps(t *testing.T) {
	projKey := "test"

	tests := []struct {
		name string
		refs grepResultLines
		want []ld.ReferenceHunksRep
	}{
		{
			name: "no references",
			refs: grepResultLines{},
			want: []ld.ReferenceHunksRep{},
		},
		{
			name: "single path, single reference with context lines",
			refs: grepResultLines{
				grepResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: "flag-1",
					FlagKeys: []string{"flag-1"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context +1",
					FlagKeys: []string{},
				},
			},
			want: []ld.ReferenceHunksRep{
				ld.ReferenceHunksRep{
					Path: "a/b",
					Hunks: []ld.HunkRep{
						ld.HunkRep{
							StartingLineNumber: 5,
							Lines:              "context -1\nflag-1\ncontext +1\n",
							ProjKey:            projKey,
							FlagKey:            "flag-1",
						},
					},
				},
			},
		},
		{
			name: "multiple paths, single reference with context lines",
			refs: grepResultLines{
				grepResultLine{
					Path:     "a/b",
					LineNum:  1,
					LineText: "flag-1",
					FlagKeys: []string{"flag-1"},
				},
				grepResultLine{
					Path:     "a/c/d",
					LineNum:  10,
					LineText: "flag-2",
					FlagKeys: []string{"flag-2"},
				},
			},
			want: []ld.ReferenceHunksRep{
				ld.ReferenceHunksRep{
					Path: "a/b",
					Hunks: []ld.HunkRep{
						ld.HunkRep{
							StartingLineNumber: 1,
							Lines:              "flag-1\n",
							ProjKey:            projKey,
							FlagKey:            "flag-1",
						},
					},
				},
				ld.ReferenceHunksRep{
					Path: "a/c/d",
					Hunks: []ld.HunkRep{
						ld.HunkRep{
							StartingLineNumber: 10,
							Lines:              "flag-2\n",
							ProjKey:            projKey,
							FlagKey:            "flag-2",
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
		refs     grepResultLines
		want     []ld.HunkRep
	}{
		{
			name:     "single reference with context lines",
			ctxLines: 1,
			refs: grepResultLines{
				grepResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: "flag-1",
					FlagKeys: []string{"flag-1"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context +1",
					FlagKeys: []string{},
				},
			},
			want: []ld.HunkRep{
				ld.HunkRep{
					StartingLineNumber: 5,
					Lines:              "context -1\nflag-1\ncontext +1\n",
					ProjKey:            projKey,
					FlagKey:            "flag-1",
				},
			},
		},
		{
			name:     "multiple references, single flag, one hunk",
			ctxLines: 1,
			refs: grepResultLines{
				grepResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: "flag-1",
					FlagKeys: []string{"flag-1"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context inner",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  8,
					LineText: "flag-1",
					FlagKeys: []string{"flag-1"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  9,
					LineText: "context +1",
					FlagKeys: []string{},
				},
			},
			want: []ld.HunkRep{
				ld.HunkRep{
					StartingLineNumber: 5,
					Lines:              "context -1\nflag-1\ncontext inner\nflag-1\ncontext +1\n",
					ProjKey:            projKey,
					FlagKey:            "flag-1",
				},
			},
		},
		{
			name:     "multiple references, multiple context lines, single flag, one hunk",
			ctxLines: 2,
			refs: grepResultLines{
				grepResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: "flag-1",
					FlagKeys: []string{"flag-1"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context inner",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  8,
					LineText: "flag-1",
					FlagKeys: []string{"flag-1"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  9,
					LineText: "context +1",
					FlagKeys: []string{},
				},
			},
			want: []ld.HunkRep{
				ld.HunkRep{
					StartingLineNumber: 5,
					Lines:              "context -1\nflag-1\ncontext inner\nflag-1\ncontext +1\n",
					ProjKey:            projKey,
					FlagKey:            "flag-1",
				},
			},
		},
		{
			name:     "multiple references, single flag, multiple hunks",
			ctxLines: 1,
			refs: grepResultLines{
				grepResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "a context -1",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: "a flag-1",
					FlagKeys: []string{"flag-1"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "a context +1",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  9,
					LineText: "b context -1",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  10,
					LineText: "b flag-1",
					FlagKeys: []string{"flag-1"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  11,
					LineText: "b context +1",
					FlagKeys: []string{},
				},
			},
			want: []ld.HunkRep{
				ld.HunkRep{
					StartingLineNumber: 5,
					Lines:              "a context -1\na flag-1\na context +1\n",
					ProjKey:            projKey,
					FlagKey:            "flag-1",
				},
				ld.HunkRep{
					StartingLineNumber: 9,
					Lines:              "b context -1\nb flag-1\nb context +1\n",
					ProjKey:            projKey,
					FlagKey:            "flag-1",
				},
			},
		},
		{
			name:     "multiple consecutive references, multiple flags, multiple hunks",
			ctxLines: 1,
			refs: grepResultLines{
				grepResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: "flag-1",
					FlagKeys: []string{"flag-1"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context inner",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  8,
					LineText: "flag-2",
					FlagKeys: []string{"flag-2"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  9,
					LineText: "context +1",
					FlagKeys: []string{},
				},
			},
			want: []ld.HunkRep{
				ld.HunkRep{
					StartingLineNumber: 5,
					Lines:              "context -1\nflag-1\ncontext inner\n",
					ProjKey:            projKey,
					FlagKey:            "flag-1",
				},
				ld.HunkRep{
					StartingLineNumber: 7,
					Lines:              "context inner\nflag-2\ncontext +1\n",
					ProjKey:            projKey,
					FlagKey:            "flag-2",
				},
			},
		},
		{
			name:     "multiple consecutive (non overlapping) references, multiple flags, multiple hunks",
			ctxLines: 1,
			refs: grepResultLines{
				grepResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "a context -1",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: "a flag-1",
					FlagKeys: []string{"flag-1"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "a context +1",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  8,
					LineText: "b context -1",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  9,
					LineText: "b flag-2",
					FlagKeys: []string{"flag-2"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  10,
					LineText: "b context +1",
					FlagKeys: []string{},
				},
			},
			want: []ld.HunkRep{
				ld.HunkRep{
					StartingLineNumber: 5,
					Lines:              "a context -1\na flag-1\na context +1\n",
					ProjKey:            projKey,
					FlagKey:            "flag-1",
				},
				ld.HunkRep{
					StartingLineNumber: 8,
					Lines:              "b context -1\nb flag-2\nb context +1\n",
					ProjKey:            projKey,
					FlagKey:            "flag-2",
				},
			},
		},
		{
			name:     "multiple references, single flag, 0 context, multiple hunks",
			ctxLines: 0,
			refs: grepResultLines{
				grepResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: "flag-1",
					FlagKeys: []string{"flag-1"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context inner",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  8,
					LineText: "flag-1",
					FlagKeys: []string{"flag-1"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  9,
					LineText: "context +1",
					FlagKeys: []string{},
				},
			},
			want: []ld.HunkRep{
				ld.HunkRep{
					StartingLineNumber: 6,
					Lines:              "flag-1\n",
					ProjKey:            projKey,
					FlagKey:            "flag-1",
				},
				ld.HunkRep{
					StartingLineNumber: 8,
					Lines:              "flag-1\n",
					ProjKey:            projKey,
					FlagKey:            "flag-1",
				},
			},
		},
		{
			name:     "multiple references, single flag, negative context, multiple hunks",
			ctxLines: -1,
			refs: grepResultLines{
				grepResultLine{
					Path:     "a/b",
					LineNum:  5,
					LineText: "context -1",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  6,
					LineText: "flag-1",
					FlagKeys: []string{"flag-1"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  7,
					LineText: "context inner",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  8,
					LineText: "flag-1",
					FlagKeys: []string{"flag-1"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  9,
					LineText: "context +1",
					FlagKeys: []string{},
				},
			},
			want: []ld.HunkRep{
				ld.HunkRep{
					StartingLineNumber: 6,
					Lines:              "",
					ProjKey:            projKey,
					FlagKey:            "flag-1",
				},
				ld.HunkRep{
					StartingLineNumber: 8,
					Lines:              "",
					ProjKey:            projKey,
					FlagKey:            "flag-1",
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
			refs: grepResultLines{
				grepResultLine{
					Path:     "a/b",
					LineNum:  1,
					LineText: "flag-1",
					FlagKeys: []string{"flag-1"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  2,
					LineText: "context+1",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  3,
					LineText: "context+3",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  10,
					LineText: "context-1",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  11,
					LineText: "flag-1",
					FlagKeys: []string{"flag-1"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  12,
					LineText: "context+1",
					FlagKeys: []string{},
				},
			},
			want: []ld.HunkRep{
				ld.HunkRep{
					StartingLineNumber: 1,
					Lines:              "flag-1\ncontext+1\n",
					ProjKey:            projKey,
					FlagKey:            "flag-1",
				},
				ld.HunkRep{
					StartingLineNumber: 10,
					Lines:              "context-1\nflag-1\ncontext+1\n",
					ProjKey:            projKey,
					FlagKey:            "flag-1",
				},
			},
		},
		{
			// This is another test case guarding against the bug described in the
			// previous test case
			name:     "multiple references, first reference at start of file",
			ctxLines: 2,
			refs: grepResultLines{
				grepResultLine{
					Path:     "a/b",
					LineNum:  1,
					LineText: "context-1",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  2,
					LineText: "flag-1",
					FlagKeys: []string{"flag-1"},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  3,
					LineText: "context+1",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  4,
					LineText: "context+2",
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  10,
					LineText: "context+alot+shouldn'tbeinhunk",
					FlagKeys: []string{},
				},
			},
			want: []ld.HunkRep{
				ld.HunkRep{
					StartingLineNumber: 1,
					Lines:              "context-1\nflag-1\ncontext+1\ncontext+2\n",
					ProjKey:            projKey,
					FlagKey:            "flag-1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupedResults := tt.refs.aggregateByPath()

			require.Equal(t, len(groupedResults), 1)

			fileGrepResults := groupedResults[0]

			got := fileGrepResults.makeHunkReps(projKey, tt.ctxLines)

			sort.Sort(byStartingLineNumber(got))

			require.Equal(t, tt.want, got)
		})
	}
}

func Test_groupIntoPathMap(t *testing.T) {
	grepResultPathALine1 := grepResultLine{
		Path:     "a",
		LineNum:  1,
		LineText: "flag-1",
		FlagKeys: []string{"flag-1"},
	}

	grepResultPathALine2 := grepResultLine{
		Path:     "a",
		LineNum:  2,
		LineText: "flag-2",
		FlagKeys: []string{"flag-2"},
	}

	grepResultPathBLine1 := grepResultLine{
		Path:     "b",
		LineNum:  1,
		LineText: "flag-3",
		FlagKeys: []string{"flag-3"},
	}
	grepResultPathBLine2 := grepResultLine{
		Path:     "b",
		LineNum:  2,
		LineText: "flag-2",
		FlagKeys: []string{"flag-4"},
	}

	lines := grepResultLines{
		grepResultPathALine1,
		grepResultPathALine2,
		grepResultPathBLine1,
		grepResultPathBLine2,
	}

	linesByPath := lines.aggregateByPath()

	aRefs := linesByPath[0]
	require.Equal(t, aRefs.path, "a")

	aRefMap := aRefs.flagReferenceMap
	require.Equal(t, len(aRefMap), 2)

	require.Contains(t, aRefMap, "flag-1")
	require.Contains(t, aRefMap, "flag-2")

	aLines := aRefs.fileGrepResultLines
	require.Equal(t, aLines.Len(), 2)
	require.Equal(t, aLines.Front().Value, grepResultPathALine1)
	require.Equal(t, aLines.Back().Value, grepResultPathALine2)

	bRefs := linesByPath[1]
	require.Equal(t, bRefs.path, "b")

	bRefMap := bRefs.flagReferenceMap
	require.Equal(t, len(aRefMap), 2)

	require.Contains(t, bRefMap, "flag-3")
	require.Contains(t, bRefMap, "flag-4")

	bLines := bRefs.fileGrepResultLines
	require.Equal(t, bLines.Len(), 2)
	require.Equal(t, bLines.Front().Value, grepResultPathBLine1)
	require.Equal(t, bLines.Back().Value, grepResultPathBLine2)
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
