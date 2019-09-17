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
