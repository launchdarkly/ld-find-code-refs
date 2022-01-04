package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_IsSubDirValid(t *testing.T) {
	baseDir, err := os.Getwd()
	testDir := "testDir"
	exampleDir := filepath.Join(baseDir, testDir)
	defer os.Remove(exampleDir)

	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(exampleDir, 0700))
	require.NoError(t, IsSubDirValid(baseDir, testDir))

	require.Error(t, IsSubDirValid(baseDir, "doesNotExist"))
	require.Error(t, IsSubDirValid(baseDir, "./"+testDir))
	require.Error(t, IsSubDirValid(baseDir, "/"+testDir))
	require.Error(t, IsSubDirValid(baseDir, "\\"+testDir))
}
