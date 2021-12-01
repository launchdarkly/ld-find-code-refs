package search

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/launchdarkly/ld-find-code-refs/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/internal/log"
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

	defaultDelims = `"` + "'`"
)

var (
	testAliases = map[string][]string{
		testFlagKey:  {testFlagAlias, testFlagAlias2},
		testFlagKey2: {testFlag2Alias, testFlag2Alias2},
	}

	// Go definition of testdata/fileWithRefs
	testFile = file{
		path:  "fileWithRefs",
		lines: []string{testFlagKey, testFlagKey2, testFlagKey + testFlagKey2, testFlagAlias, testFlag2Alias},
	}
	testResultHunks = []ld.HunkRep{
		makeHunk(1, testFlagKey),
		*withAliases(makeHunkPtr(3, testFlagKey+testFlagKey2, testFlagAlias), testFlagAlias), //combined
		*withFlagKey(makeHunkPtr(2, testFlagKey2, testFlagKey+testFlagKey2), testFlagKey2),   //combined
		*withFlagKey(withAliases(makeHunkPtr(5, testFlag2Alias), testFlag2Alias), testFlagKey2),
	}

	delimitedTestFlagKey = delimit(testFlagKey, `"`)
)

func Test_hunkForLine(t *testing.T) {
	tests := []struct {
		name    string
		lineNum int
		lines   []string
		flagKey string
		matcher Matcher
		want    *ld.HunkRep
	}{
		{
			name: "does not match flag flag key without delimiters",
			matcher: Matcher{
				ctxLines: 0,
				Elements: []ElementMatcher{
					NewElementMatcher("my-project", ``, `"`, []string{testFlagKey}, nil),
				},
			},
			lineNum: 0,
			flagKey: testFlagKey,
			lines:   []string{testFlagKey},
			want:    nil,
		},
		{
			name: "matches flag key with delimiters",
			matcher: Matcher{
				ctxLines: 0,
				Elements: []ElementMatcher{
					NewElementMatcher("my-project", ``, `"`, []string{testFlagKey}, nil),
				},
			},
			lineNum: 0,
			flagKey: testFlagKey,
			lines:   []string{delimitedTestFlagKey},
			want:    makeHunkPtr(1, delimitedTestFlagKey),
		},
		{
			name: "matches no context lines without delimiters",
			matcher: Matcher{
				ctxLines: -1,
				Elements: []ElementMatcher{
					NewElementMatcher("my-project", ``, ``, []string{testFlagKey}, nil),
				},
			},
			lineNum: 0,
			flagKey: testFlagKey,
			lines:   []string{testFlagKey},
			want:    makeHunkPtr(1),
		},
		{
			name: "matches with alias",
			matcher: Matcher{
				ctxLines: -1,
				Elements: []ElementMatcher{
					NewElementMatcher("my-project", ``, ``, nil, testAliases),
				},
			},
			lineNum: 0,
			flagKey: testFlagKey,
			lines:   []string{testFlagAlias},
			want:    withAliases(makeHunkPtr(1), testFlagAlias),
		},
		{
			name: "matches with aliases",
			matcher: Matcher{
				ctxLines: -1,
				Elements: []ElementMatcher{
					NewElementMatcher("my-project", ``, ``, nil, testAliases),
				},
			},
			lineNum: 0,
			flagKey: testFlagKey,
			lines:   []string{testFlagAlias + " " + testFlagAlias2},
			want:    withAliases(makeHunkPtr(1), testFlagAlias, testFlagAlias2),
		},
		{
			name: "matches with line",
			matcher: Matcher{
				ctxLines: 0,
				Elements: []ElementMatcher{
					NewElementMatcher("my-project", ``, ``, []string{testFlagKey}, nil),
				},
			},
			lineNum: 1,
			flagKey: testFlagKey,
			lines:   []string{"", testFlagKey, ""},
			want:    makeHunkPtr(2, testFlagKey),
		},
		{
			name: "matches with context lines",
			matcher: Matcher{
				ctxLines: 1,
				Elements: []ElementMatcher{
					NewElementMatcher("my-project", ``, ``, []string{testFlagKey}, nil),
				},
			},
			lineNum: 1,
			flagKey: testFlagKey,
			lines:   []string{"", testFlagKey, ""},
			want:    makeHunkPtr(1, "", testFlagKey, ""),
		},
		{
			name: "truncates long line",
			matcher: Matcher{
				ctxLines: 0,
				Elements: []ElementMatcher{
					NewElementMatcher("my-project", ``, ``, []string{testFlagKey}, nil),
				},
			},
			lineNum: 0,
			flagKey: testFlagKey,
			lines:   []string{testFlagKey + strings.Repeat("a", maxLineCharCount)},
			want:    makeHunkPtr(1, testFlagKey+strings.Repeat("a", maxLineCharCount-len(testFlagKey))+"…"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := file{lines: tt.lines}
			got := f.hunkForLine("default", tt.flagKey, tt.lineNum, tt.matcher)
			require.Equal(t, tt.want, got)
		})
	}

}

func Test_aggregateHunksForFlag(t *testing.T) {
	tests := []struct {
		name    string
		matcher Matcher
		lines   []string
		aliases []string
		want    []ld.HunkRep
	}{
		{
			name: "does not set lines when context lines are disabled",
			matcher: Matcher{
				ctxLines: -1,
				Elements: []ElementMatcher{NewElementMatcher("default", ``, defaultDelims, []string{testFlagKey}, nil)},
			},
			lines: []string{delimitedTestFlagKey, delimitedTestFlagKey, delimitedTestFlagKey},
			want: []ld.HunkRep{
				makeHunk(1),
				makeHunk(2),
				makeHunk(3),
			},
		},
		{
			name: "combines adjacent hunks with no additional context lines",
			matcher: Matcher{
				ctxLines: 0,
				Elements: []ElementMatcher{NewElementMatcher("default", ``, defaultDelims, []string{testFlagKey}, nil)},
			}, lines: []string{delimitedTestFlagKey, delimitedTestFlagKey, delimitedTestFlagKey},
			want: []ld.HunkRep{
				makeHunk(1, delimitedTestFlagKey, delimitedTestFlagKey, delimitedTestFlagKey),
			},
		},
		{
			name: "combines adjacent hunks",
			matcher: Matcher{
				ctxLines: 1,
				Elements: []ElementMatcher{NewElementMatcher("default", ``, defaultDelims, []string{testFlagKey}, nil)},
			}, lines: []string{delimitedTestFlagKey, "", "", delimitedTestFlagKey, "", "", delimitedTestFlagKey},
			want: []ld.HunkRep{
				makeHunk(1, delimitedTestFlagKey, "", "", delimitedTestFlagKey, "", "", delimitedTestFlagKey),
			},
		},
		{
			name: "does not combine hunks with no overlap",
			matcher: Matcher{
				ctxLines: 1,
				Elements: []ElementMatcher{NewElementMatcher("default", ``, defaultDelims, []string{testFlagKey}, nil)},
			},
			lines: []string{delimitedTestFlagKey, "", "", "", delimitedTestFlagKey, "", "", "", delimitedTestFlagKey},
			want: []ld.HunkRep{
				makeHunk(1, delimitedTestFlagKey, ""),
				makeHunk(4, "", delimitedTestFlagKey, ""),
				makeHunk(8, "", delimitedTestFlagKey),
			},
		},
		{
			name: "combines overlapping hunks",
			matcher: Matcher{
				ctxLines: 1,
				Elements: []ElementMatcher{NewElementMatcher("default", ``, defaultDelims, []string{testFlagKey}, nil)},
			},
			lines: []string{delimitedTestFlagKey, "", delimitedTestFlagKey, "", delimitedTestFlagKey},
			want: []ld.HunkRep{
				makeHunk(1, delimitedTestFlagKey, "", delimitedTestFlagKey, "", delimitedTestFlagKey),
			},
		},
		{
			name: "combines multiple types of overlaps",
			matcher: Matcher{
				ctxLines: 1,
				Elements: []ElementMatcher{NewElementMatcher("default", ``, defaultDelims, []string{testFlagKey}, nil)},
			},
			lines: []string{delimitedTestFlagKey, "", delimitedTestFlagKey, "", delimitedTestFlagKey},
			want: []ld.HunkRep{
				makeHunk(1, delimitedTestFlagKey, "", delimitedTestFlagKey, "", delimitedTestFlagKey),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := file{lines: tt.lines}
			var lineNumbers []int
			for i := range tt.lines {
				lineNumbers = append(lineNumbers, i)
			}
			got := f.aggregateHunksForFlag("default", testFlagKey, tt.matcher, lineNumbers)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_mergeHunks(t *testing.T) {
	tests := []struct {
		name  string
		hunk1 ld.HunkRep
		hunk2 ld.HunkRep
		want  []ld.HunkRep
	}{
		{
			name:  "combine adjacent hunks",
			hunk1: makeHunk(1, "a", "b", "c"),
			hunk2: makeHunk(4, "d", "e", "f"),
			want:  []ld.HunkRep{makeHunk(1, "a", "b", "c", "d", "e", "f")},
		},
		{
			name:  "combine overlapping hunks",
			hunk1: makeHunk(1, "a", "b", "c"),
			hunk2: makeHunk(3, "c", "d", "e"),
			want:  []ld.HunkRep{makeHunk(1, "a", "b", "c", "d", "e")},
		},
		{
			name:  "combine overlapping hunks provided in the wrong order",
			hunk1: makeHunk(3, "c", "d", "e"),
			hunk2: makeHunk(1, "a", "b", "c"),
			want:  []ld.HunkRep{makeHunk(1, "a", "b", "c", "d", "e")},
		},
		{
			name:  "combine same hunk",
			hunk1: makeHunk(1, "a", "b", "c"),
			hunk2: makeHunk(1, "a", "b", "c"),
			want:  []ld.HunkRep{makeHunk(1, "a", "b", "c")},
		},
		{
			name:  "combine subset hunk",
			hunk1: makeHunk(1, "a", "b", "c", "d", "e"),
			hunk2: makeHunk(3, "c", "d"),
			want:  []ld.HunkRep{makeHunk(1, "a", "b", "c", "d", "e")},
		},
		{
			// if the hunks do not overlap and are not adjacent, expect just the first hunk to be returned
			name:  "do not combine disjoint hunks",
			hunk1: makeHunk(1, "a", "b", "c"),
			hunk2: makeHunk(5, "e", "f", "g"),
			want:  []ld.HunkRep{makeHunk(1, "a", "b", "c"), makeHunk(5, "e", "f", "g")},
		},
		{
			// if the hunks are provided out of order, expect both hunks to be returned in the correct order
			name:  "do not combine hunks provided out of order",
			hunk1: makeHunk(5, "e", "f", "g"),
			hunk2: makeHunk(1, "a", "b", "c"),
			want:  []ld.HunkRep{makeHunk(1, "a", "b", "c"), makeHunk(5, "e", "f", "g")},
		},
		{
			name:  "does not combine with no context lines",
			hunk1: makeHunk(1),
			hunk2: makeHunk(2),
			want:  []ld.HunkRep{makeHunk(1), makeHunk(2)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeHunks(tt.hunk1, tt.hunk2)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_toHunks(t *testing.T) {
	f := testFile
	matcher := Matcher{
		ctxLines: 0,
		Elements: []ElementMatcher{
			NewElementMatcher("default", "", "", []string{testFlagKey, testFlagKey2}, testAliases),
		},
	}
	got := f.toHunks(matcher)
	require.Equal(t, "fileWithRefs", got.Path)
	require.Equal(t, len(testResultHunks), len(got.Hunks))
	// no hunks should generate no references
	emptyMatcher := Matcher{
		ctxLines: 0,
		Elements: []ElementMatcher{
			NewElementMatcher("default", "", "", nil, nil),
		},
	}
	require.Nil(t, f.toHunks(emptyMatcher))
}

func Test_processFiles(t *testing.T) {
	f := testFile
	linesCopy := make([]string, len(f.lines))
	copy(linesCopy, f.lines)
	f2 := file{path: f.path + "2", lines: linesCopy}

	files := make(chan file, 3)
	references := make(chan ld.ReferenceHunksRep, 3)
	files <- f
	files <- f2
	files <- file{path: "no-refs"}
	close(files)
	matcher := Matcher{
		ctxLines: 0,
	}
	matcher.Elements = append(matcher.Elements,
		NewElementMatcher("default", "", "", []string{testFlagKey, testFlagKey2}, testAliases),
	)
	go processFiles(context.Background(), files, references, matcher)
	totalRefs := 0
	totalHunks := 0
	for reference := range references {
		totalRefs++
		totalHunks += len(reference.Hunks)
	}
	require.Equal(t, 2, totalRefs, "The file with no references should not have been added to refs")
	require.Equal(t, 8, totalHunks, "See Test_toHunks for a more comprehensive example of why this should be 4 per file (2 files with the same refs)")
}

func Test_SearchForRefs(t *testing.T) {
	os.Symlink("testdata/fileWithRefs", "testdata/symlink")
	want := []ld.ReferenceHunksRep{{Path: testFile.path}}
	matcher := Matcher{
		ctxLines: 0,
	}
	matcher.Elements = append(matcher.Elements,
		NewElementMatcher("default", "", "", []string{testFlagKey, testFlagKey2}, nil),
	)
	t.Cleanup(func() { os.Remove("testdata/symlink") })
	got, err := SearchForRefs("testdata", matcher)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, want[0].Path, got[0].Path)
}

func withAliases(hunk *ld.HunkRep, aliases ...string) *ld.HunkRep {
	hunk.Aliases = aliases
	return hunk
}

func withFlagKey(hunk *ld.HunkRep, flagKey string) *ld.HunkRep {
	hunk.FlagKey = flagKey
	return hunk
}

func makeHunkPtr(startingLineNumber int, lines ...string) *ld.HunkRep {
	hunk := makeHunk(startingLineNumber, lines...)
	return &hunk
}

func makeHunk(startingLineNumber int, lines ...string) ld.HunkRep {
	hunkLines := ""
	if len(lines) != 0 {
		hunkLines = strings.Join(lines, "\n")
	}
	return ld.HunkRep{
		ProjKey:            "default",
		FlagKey:            testFlagKey,
		StartingLineNumber: startingLineNumber,
		Lines:              hunkLines,
		Aliases:            []string{},
	}
}

func delimit(s string, delim string) string {
	return delim + s + delim
}
