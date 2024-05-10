package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_IsSubDirValid(t *testing.T) {
	tempDir := os.TempDir()
	testDir := "testDir"
	exampleDir := filepath.Join(os.TempDir(), testDir)

	defer os.Remove(exampleDir)

	require.NoError(t, os.MkdirAll(exampleDir, 0700))
	require.NoError(t, IsSubDirValid(tempDir, testDir))

	require.Error(t, IsSubDirValid(tempDir, "doesNotExist"))
	require.Error(t, IsSubDirValid(tempDir, "./"+testDir))
	require.Error(t, IsSubDirValid(tempDir, "/"+testDir))
	require.Error(t, IsSubDirValid(tempDir, "\\"+testDir))
}
