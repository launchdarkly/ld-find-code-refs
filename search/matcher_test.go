package search

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_buildDelimiterList(t *testing.T) {
	testFlagKey := "testflag"
	delimitedFlag := buildDelimiterList([]string{testFlagKey}, defaultDelims)
	want := map[string][]string{"testflag": []string{"\"testflag\"", "\"testflag'", "\"testflag`", "'testflag\"", "'testflag'", "'testflag`", "`testflag\"", "`testflag'", "`testflag`"}}
	require.Equal(t, want, delimitedFlag)
}

func Test_MatchElement(t *testing.T) {
	matcher := Matcher{
		Delimiters: defaultDelims,
		Elements: []ElementMatcher{
			{
				Elements:       []string{"testflag"},
				Aliases:        map[string][]string{"testflag": {}},
				Delimiters:     []string{"\"", "'", "`"},
				DelimitedFlags: map[string][]string{"testflag": {"\"testflag\"", "\"testflag'", "\"testflag`", "'testflag\"", "'testflag'", "'testflag`", "`testflag\"", "`testflag'", "`testflag`"}},
			},
		},
	}

	require.Equal(t, true, matcher.MatchElement("var featureFlag = 'testflag';", "testflag"))
}
