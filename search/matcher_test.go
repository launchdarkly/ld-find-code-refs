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

// func Test_MatchElement(t *testing.T) {
// 	const FLAG_KEY = "testflag"

// 	testFlagDelimitedFlags := map[string][]string{"testflag": {"\"testflag\"", "\"testflag'", "\"testflag`", "'testflag\"", "'testflag'", "'testflag`", "`testflag\"", "`testflag'", "`testflag`"}}

// 	differentFlagDelimitedFlags := map[string][]string{"different-flag": {"\"different-flag\"", "\"different-flag'", "\"different-flag`", "'different-flag\"", "'different-flag'", "'different-flag`", "`different-flag\"", "`different-flag'", "`different-flag`"}}

// 	specs := []struct {
// 		name     string
// 		expected bool
// 		line     string
// 		matcher  Matcher
// 	}{
// 		{
// 			name:     "match found - no delimiters",
// 			expected: true,
// 			line:     "var flagKey = 'testflag'",
// 			matcher:  Matcher{Delimiters: ""},
// 		},
// 		{
// 			name:     "match found - with delimters",
// 			expected: true,
// 			line:     "var flagKey = 'testflag'",
// 			matcher:  Matcher{Delimiters: defaultDelims, Elements: []ElementMatcher{{DelimitedFlags: differentFlagDelimitedFlags}, {DelimitedFlags: testFlagDelimitedFlags}}},
// 		},
// 		{
// 			name:     "no match found - no delimiters",
// 			expected: false,
// 			line:     "var flagKey = 'another-flag'",
// 			matcher:  Matcher{Delimiters: ""},
// 		},
// 		{
// 			name:     "no match found - with delimiters",
// 			expected: false,
// 			line:     "var flagKey = 'another-flag'",
// 			matcher:  Matcher{Delimiters: defaultDelims, Elements: []ElementMatcher{{DelimitedFlags: differentFlagDelimitedFlags}, {DelimitedFlags: testFlagDelimitedFlags}}},
// 		},
// 	}

// 	for _, tt := range specs {
// 		tt := tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			require.Equal(t, tt.expected, tt.matcher.MatchElement(tt.line, FLAG_KEY))
// 		})
// 	}
// }
