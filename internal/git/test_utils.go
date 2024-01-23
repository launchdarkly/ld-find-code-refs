package git

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	REPO_DIR = "testdata/repo"
)

func copyFile(t *testing.T, src, dst string) {
	sourceFileStat, err := os.Stat(src)
	require.NoError(t, err)
	require.Truef(t, sourceFileStat.Mode().IsRegular(), "%s is not a regular file", src)

	source, err := os.Open(src)
	require.NoError(t, err)
	defer source.Close()

	destination, err := os.Create(dst)
	require.NoError(t, err)

	defer destination.Close()
	_, err = io.Copy(destination, source)
	require.NoError(t, err)
}

func createRepoFile(t *testing.T, path string, content *string) {
	flagFile, err := os.Create(repoPath(path))
	require.NoError(t, err)
	if content != nil {
		_, err = flagFile.WriteString(*content)
		require.NoError(t, err)
	}
	require.NoError(t, flagFile.Close())
}

func removeRepoFile(t *testing.T, path string) {
	require.NoError(t, os.Remove(filepath.Join(REPO_DIR, path)))
}

func repoPath(path string) string {
	return filepath.Join(REPO_DIR, path)
}
