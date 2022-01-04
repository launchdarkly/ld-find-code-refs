package validation

import (
	"fmt"
	"os"
	"path/filepath"
)

func NormalizeAndValidatePath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid directory: %s", err)
	}

	exists, err := dirExists(absPath)
	if err != nil {
		return "", fmt.Errorf("invalid directory: %s", err)
	}

	if !exists {
		return "", fmt.Errorf("directory does not exist: %s", absPath)
	}

	return absPath, nil
}

func dirExists(path string) (bool, error) {
	fileInfo, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return fileInfo.Mode().IsDir(), nil
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func IsSubDirValid(base, subdir string) error {
	if prefixExists(subdir) {
		return fmt.Errorf(`project subdirectory should not start with prefix: %s`, string(subdir[0]))
	}
	path := filepath.Join(base, subdir)
	pathInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !pathInfo.IsDir() {
		return fmt.Errorf(`path: %s is not a directory`, path)
	}

	return nil
}

func prefixExists(path string) bool {
	return path[0] == '\\' || path[0] == '/' || path[0] == '.'
}
