package search

import (
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

func Test_consolidateHunks(t *testing.T) {
	tests := []struct {
		name     string
		ctxLines int
		hunks    []ld.HunkRep
		want     []ld.HunkRep
	}{
		{
			name:     "does not combine any hunks when there are no overlaps",
			ctxLines: 1,
			hunks: []ld.HunkRep{
				makeHunk(1, "a", "b", "c"),
				makeHunk(5, "e", "f", "g"),
				makeHunk(9, "i", "j", "k"),
			},
			want: []ld.HunkRep{
				makeHunk(1, "a", "b", "c"),
				makeHunk(5, "e", "f", "g"),
				makeHunk(9, "i", "j", "k"),
			},
		},
		{
			name:     "combines adjacent hunks",
			ctxLines: 0,
			hunks: []ld.HunkRep{
				makeHunk(1, "a"),
				makeHunk(2, "b"),
				makeHunk(3, "c"),
			},
			want: []ld.HunkRep{
				makeHunk(1, "a", "b", "c"),
			},
		},
		{
			name:     "does not combine when there are no context lines",
			ctxLines: -1,
			hunks: []ld.HunkRep{
				makeHunk(1, "a"),
				makeHunk(2, "b"),
				makeHunk(3, "c"),
			},
			want: []ld.HunkRep{
				makeHunk(1, "a"),
				makeHunk(2, "b"),
				makeHunk(3, "c"),
			},
		},
		{
			name:     "combines overlapping hunks",
			ctxLines: 1,
			hunks: []ld.HunkRep{
				makeHunk(1, "a", "b", "c"),
				makeHunk(3, "c", "d", "e"),
				makeHunk(5, "e", "f", "g"),
			},
			want: []ld.HunkRep{
				makeHunk(1, "a", "b", "c", "d", "e", "f", "g"),
			},
		},
		{
			name:     "combines multiple types of overlaps",
			ctxLines: 1,
			hunks: []ld.HunkRep{
				makeHunk(1, "a", "b", "c"),
				makeHunk(4, "d", "e", "f"),
				makeHunk(6, "f", "g", "h"),
				makeHunk(10, "j", "k", "l"),
			},
			want: []ld.HunkRep{
				makeHunk(1, "a", "b", "c", "d", "e", "f", "g", "h"),
				makeHunk(10, "j", "k", "l"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := consolidateHunks(tt.hunks, tt.ctxLines)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_mergeHunks(t *testing.T) {
	tests := []struct {
		name     string
		ctxLines int
		hunk1    ld.HunkRep
		hunk2    ld.HunkRep
		want     []ld.HunkRep
	}{
		{
			name:     "combine adjacent hunks",
			ctxLines: 1,
			hunk1:    makeHunk(1, "a", "b", "c"),
			hunk2:    makeHunk(4, "d", "e", "f"),
			want:     []ld.HunkRep{makeHunk(1, "a", "b", "c", "d", "e", "f")},
		},
		{
			name:     "combine overlapping hunks",
			ctxLines: 1,
			hunk1:    makeHunk(1, "a", "b", "c"),
			hunk2:    makeHunk(3, "c", "d", "e"),
			want:     []ld.HunkRep{makeHunk(1, "a", "b", "c", "d", "e")},
		},
		{
			name:     "combine overlapping hunks provided in the wrong order",
			ctxLines: 1,
			hunk1:    makeHunk(3, "c", "d", "e"),
			hunk2:    makeHunk(1, "a", "b", "c"),
			want:     []ld.HunkRep{makeHunk(1, "a", "b", "c", "d", "e")},
		},
		{
			name:     "combine same hunk",
			ctxLines: 1,
			hunk1:    makeHunk(1, "a", "b", "c"),
			hunk2:    makeHunk(1, "a", "b", "c"),
			want:     []ld.HunkRep{makeHunk(1, "a", "b", "c")},
		},
		{
			name:     "combine subset hunk",
			ctxLines: 2,
			hunk1:    makeHunk(1, "a", "b", "c", "d", "e"),
			hunk2:    makeHunk(3, "c", "d"),
			want:     []ld.HunkRep{makeHunk(1, "a", "b", "c", "d", "e")},
		},
		{
			// if the hunks do not overlap and are not adjacent, expect just the first hunk to be returned
			name:     "do not combine disjoint hunks",
			ctxLines: 1,
			hunk1:    makeHunk(1, "a", "b", "c"),
			hunk2:    makeHunk(5, "e", "f", "g"),
			want:     []ld.HunkRep{makeHunk(1, "a", "b", "c"), makeHunk(5, "e", "f", "g")},
		},
		{
			// if the hunks are provided out of order, expect both hunks to be returned in the correct order
			name:     "do not combine hunks provided out of order",
			ctxLines: 1,
			hunk1:    makeHunk(5, "e", "f", "g"),
			hunk2:    makeHunk(1, "a", "b", "c"),
			want:     []ld.HunkRep{makeHunk(1, "a", "b", "c"), makeHunk(5, "e", "f", "g")},
		},
		{
			name:     "does not combine with no context lines",
			ctxLines: -1,
			hunk1:    makeHunk(1, "a"),
			hunk2:    makeHunk(2, "b"),
			want:     []ld.HunkRep{makeHunk(1, "a"), makeHunk(2, "b")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeHunks(tt.hunk1, tt.hunk2, tt.ctxLines)
			require.Equal(t, tt.want, got)
		})
	}
}

func makeHunk(startingLineNumber int, lines ...string) ld.HunkRep {
	return ld.HunkRep{StartingLineNumber: startingLineNumber, Lines: strings.Join(lines, "\n"), Aliases: []string{}}
}
