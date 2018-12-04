package parse

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/launchdarkly/git-flag-parser/parse/internal/ld"
)

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
							Offset:  5,
							Lines:   "context -1\nflag-1\ncontext +1",
							ProjKey: projKey,
							FlagKey: "flag-1",
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
					Path: "a/c/d",
					Hunks: []ld.HunkRep{
						ld.HunkRep{
							Offset:  10,
							Lines:   "flag-2",
							ProjKey: projKey,
							FlagKey: "flag-2",
						},
					},
				},
				ld.ReferenceHunksRep{
					Path: "a/b",
					Hunks: []ld.HunkRep{
						ld.HunkRep{
							Offset:  1,
							Lines:   "flag-1",
							ProjKey: projKey,
							FlagKey: "flag-1",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.refs.makeReferenceHunksReps(projKey)

			require.Equal(t, tt.want, got)
		})
	}
}

func Test_makeHunkReps(t *testing.T) {
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
			name: "single reference with context lines",
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
							Offset:  5,
							Lines:   "context -1\nflag-1\ncontext +1",
							ProjKey: projKey,
							FlagKey: "flag-1",
						},
					},
				},
			},
		},
		{
			name: "multiple references, single flag, one hunk",
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
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  9,
					LineText: "context +1",
					FlagKeys: []string{},
				},
			},
			want: []ld.ReferenceHunksRep{
				ld.ReferenceHunksRep{
					Path: "a/b",
					Hunks: []ld.HunkRep{
						ld.HunkRep{
							Offset:  5,
							Lines:   "context -1\nflag-1\ncontext inner\nflag-1\ncontext +1",
							ProjKey: projKey,
							FlagKey: "flag-1",
						},
					},
				},
			},
		},
		{
			name: "multiple references, single flag, multiple hunks",
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
					FlagKeys: []string{},
				},
				grepResultLine{
					Path:     "a/b",
					LineNum:  11,
					LineText: "b context +1",
					FlagKeys: []string{},
				},
			},
			want: []ld.ReferenceHunksRep{
				ld.ReferenceHunksRep{
					Path: "a/b",
					Hunks: []ld.HunkRep{
						ld.HunkRep{
							Offset:  5,
							Lines:   "a context -1\na flag-1\na context +1",
							ProjKey: projKey,
							FlagKey: "flag-1",
						},
						ld.HunkRep{
							Offset:  9,
							Lines:   "b context -1\nb flag-1\nb context +1",
							ProjKey: projKey,
							FlagKey: "flag-1",
						},
					},
				},
			},
		},
		{
			name: "multiple consecutive references, multiple flags, multiple hunks",
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
			want: []ld.ReferenceHunksRep{
				ld.ReferenceHunksRep{
					Path: "a/b",
					Hunks: []ld.HunkRep{
						ld.HunkRep{
							Offset:  5,
							Lines:   "context -1\nflag-1\ncontext inner",
							ProjKey: projKey,
							FlagKey: "flag-1",
						},
						ld.HunkRep{
							Offset:  7,
							Lines:   "flag-1\nflag-2\ncontext +1",
							ProjKey: projKey,
							FlagKey: "flag-2",
						},
					},
				},
			},
		},
		{
			name: "multiple consecutive (non overlapping) references, multiple flags, multiple hunks",
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
			want: []ld.ReferenceHunksRep{
				ld.ReferenceHunksRep{
					Path: "a/b",
					Hunks: []ld.HunkRep{
						ld.HunkRep{
							Offset:  5,
							Lines:   "a context -1\na flag-1\na context +1",
							ProjKey: projKey,
							FlagKey: "flag-1",
						},
						ld.HunkRep{
							Offset:  7,
							Lines:   "b context -1\nb flag-1\nb context +1",
							ProjKey: projKey,
							FlagKey: "flag-2",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.refs.makeReferenceHunksReps(projKey)

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
	grepResultPathALine3 := grepResultLine{
		Path:     "a",
		LineNum:  3,
		LineText: "context",
		FlagKeys: []string{},
	}
	grepResultPathBLine1 := grepResultLine{
		Path:     "b",
		LineNum:  1,
		LineText: "flag-1",
		FlagKeys: []string{"flag-1"},
	}
	grepResultPathBLine2 := grepResultLine{
		Path:     "b",
		LineNum:  2,
		LineText: "flag-2",
		FlagKeys: []string{"flag-2"},
	}
	grepResultPathBLine3 := grepResultLine{
		Path:     "b",
		LineNum:  3,
		LineText: "context",
		FlagKeys: []string{},
	}
	tests := []struct {
		name        string
		grepResults grepResultLines
		want        grepResultPathMap
	}{
		{
			name:        "no grep results",
			grepResults: grepResultLines{},
			want:        grepResultPathMap{},
		},
		{
			name: "one file, one flag reference, no context",
			grepResults: grepResultLines{
				grepResultPathALine1,
			},
			want: grepResultPathMap{
				"a": &fileGrepResults{
					flagReferenceMap: flagReferenceMap{
						"flag-1": []*grepResultLine{&grepResultPathALine1},
					},
					fileGrepResultLines: []grepResultLine{
						grepResultPathALine1,
					},
				},
			},
		},
		{
			name: "one file, multiple flags",
			grepResults: grepResultLines{
				grepResultPathALine1,
				grepResultPathALine2,
			},
			want: grepResultPathMap{
				"a": &fileGrepResults{
					flagReferenceMap: flagReferenceMap{
						"flag-1": []*grepResultLine{&grepResultPathALine1},
						"flag-2": []*grepResultLine{&grepResultPathALine2},
					},
					fileGrepResultLines: []grepResultLine{
						grepResultPathALine1,
						grepResultPathALine2,
					},
				},
			},
		},
		{
			name: "one file, multiple flags and context",
			grepResults: grepResultLines{
				grepResultPathALine1,
				grepResultPathALine2,
				grepResultPathALine3,
			},
			want: grepResultPathMap{
				"a": &fileGrepResults{
					flagReferenceMap: flagReferenceMap{
						"flag-1": []*grepResultLine{&grepResultPathALine1},
						"flag-2": []*grepResultLine{&grepResultPathALine2},
					},
					fileGrepResultLines: []grepResultLine{
						grepResultPathALine1,
						grepResultPathALine2,
						grepResultPathALine3,
					},
				},
			},
		},
		{
			name: "multiple files, multiple flags and context",
			grepResults: grepResultLines{
				grepResultPathALine1,
				grepResultPathALine2,
				grepResultPathALine3,
				grepResultPathBLine1,
				grepResultPathBLine2,
				grepResultPathBLine3,
			},
			want: grepResultPathMap{
				"a": &fileGrepResults{
					flagReferenceMap: flagReferenceMap{
						"flag-1": []*grepResultLine{&grepResultPathALine1},
						"flag-2": []*grepResultLine{&grepResultPathALine2},
					},
					fileGrepResultLines: []grepResultLine{
						grepResultPathALine1,
						grepResultPathALine2,
						grepResultPathALine3,
					},
				},
				"b": &fileGrepResults{
					flagReferenceMap: flagReferenceMap{
						"flag-1": []*grepResultLine{&grepResultPathBLine1},
						"flag-2": []*grepResultLine{&grepResultPathBLine2},
					},
					fileGrepResultLines: []grepResultLine{
						grepResultPathBLine1,
						grepResultPathBLine2,
						grepResultPathBLine3,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.grepResults.groupIntoPathMap()

			require.Equal(t, tt.want, got)
		})

	}

}
