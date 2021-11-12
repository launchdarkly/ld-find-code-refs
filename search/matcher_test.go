package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_buildFlagPatterns(t *testing.T) {
	testFlagKey := "testflag"
	patterns := buildElementPatterns([]string{testFlagKey}, defaultDelims)
	want := map[string][]string{"testflag": []string{"\"testflag\"", "\"testflag'", "\"testflag`", "'testflag\"", "'testflag'", "'testflag`", "`testflag\"", "`testflag'", "`testflag`"}}
	require.Equal(t, want, patterns)
}

func TestElementMatcher_FindAliases(t *testing.T) {
	t.Run("overlapping aliases are reported separately", func(t *testing.T) {
		matcher := NewElementMatcher("project", "", nil, map[string][]string{"flag": {"alias", "alias1"}})
		assert.ElementsMatch(t, []string{"alias", "alias1"}, matcher.FindAliases("alias1", "flag"))
	})
}

func TestElementMatcher_FindMatches(t *testing.T) {
	t.Run("overlapping flags are reported separately", func(t *testing.T) {
		matcher := NewElementMatcher("project", "", []string{"flag", "flag1"}, nil)
		assert.ElementsMatch(t, []string{"flag", "flag1"}, matcher.FindMatches("flag1"))
	})
}
