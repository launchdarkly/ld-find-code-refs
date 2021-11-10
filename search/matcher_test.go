package search

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_buildFlagPatterns(t *testing.T) {
	testFlagKey := "testflag"
	patterns := buildFlagPatterns([]string{testFlagKey}, defaultDelims)
	want := map[string][]string{"testflag": []string{"\"testflag\"", "\"testflag'", "\"testflag`", "'testflag\"", "'testflag'", "'testflag`", "`testflag\"", "`testflag'", "`testflag`"}}
	require.Equal(t, want, patterns)
}
