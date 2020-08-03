package git

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/launchdarkly/ld-find-code-refs/internal/ld"
)

// command line flag to auto-update testdata files
var update = flag.Bool("update", false, "update testdata files")

// TestFindExtinctions is an integration test against a real Git repository stored under the testdata directory.
func TestFindExtinctions(t *testing.T) {
	c := Client{workspace: "testdata/repo"}
	extinctions, err := c.FindExtinctions("default", []string{"flag1", "flag2"}, `"`, 10)
	require.NoError(t, err)

	testResultsPath := "testdata/extinction_results_rep.json"

	var expected []ld.ExtinctionRep
	// This is a golden file test. If `go test` is run with the `-update` flag, the results of this test will be updated. This is useful when manipulating the commit history of testdata/repo
	if *update {
		t.Logf("updating %s", testResultsPath)
		bytes, _ := json.Marshal(extinctions)
		err := ioutil.WriteFile(testResultsPath, bytes, 0644)
		require.NoErrorf(t, err, "failed to update %s: %s", testResultsPath, err)
	} else {
		jsonFile, err := os.Open(testResultsPath)
		require.NoError(t, err)
		defer jsonFile.Close()
		bytes, err := ioutil.ReadAll(jsonFile)
		json.Unmarshal(bytes, &expected)
		require.NoError(t, err)
		require.Equal(t, expected, extinctions)
	}

}
