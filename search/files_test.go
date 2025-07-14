package search

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_readFiles(t *testing.T) {
	t.Run("don't ignore .github by default", func(t *testing.T) {
		files := make(chan file, 8)
		err := readFiles(context.Background(), files, "testdata/include-github-files", "")
		require.NoError(t, err)
		got := []file{}
		for file := range files {
			got = append(got, file)
			switch file.path {
			case "fileWithNoRefs":
				assert.Equal(t, []string{"fileWithNoRefs"}, file.lines)
			case "fileWithRefs", ".github/workflows/workflow.yml":
				assert.Equal(t, testFile.lines, file.lines)
			case "ignoredFiles/included":
				assert.Equal(t, []string{"IGNORED BUT INCLUDED"}, file.lines)
			case "symlink":
				assert.Fail(t, "Should not read symlink contents")
			default:
				assert.Fail(t, "Read unexpected file", file)
			}
		}
		assert.Len(t, got, 4, "Expected 4 valid files to have been found")
	})

	t.Run("explicitly ignore .github files", func(t *testing.T) {
		files := make(chan file, 8)
		err := readFiles(context.Background(), files, "testdata/exclude-github-files", "")
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
			case "symlink":
				assert.Fail(t, "Should not read symlink contents")
			default:
				assert.Fail(t, "Read unexpected file", file)
			}
		}
		assert.Len(t, got, 3, "Expected 3 valid files to have been found")
	})
}
