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
