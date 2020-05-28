package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_readFiles(t *testing.T) {
	files := make(chan file, 8)
	err := readFiles(files, "testdata")
	require.NoError(t, err)
	got := []file{}
	for file := range files {
		got = append(got, file)
		switch file.path {
		case "fileWithNoRefs":
			assert.Equal(t, []string{"fileWithNoRefs"}, file.lines)
		case "fileWithRefs":
			assert.Equal(t, testFile.lines, file.lines)
		case "ignoredFiles/included":
			assert.Equal(t, []string{"IGNORED BUT INCLUDED"}, file.lines)
		default:
			assert.Fail(t, "Read unexpected file", file)
		}
	}
	assert.Len(t, got, 3, "Expected 3 valid files to have been found")
}
