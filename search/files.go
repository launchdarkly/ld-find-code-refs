package search

import (
	"bufio"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/monochromegane/go-gitignore"
	"golang.org/x/tools/godoc/util"

	"github.com/launchdarkly/ld-find-code-refs/v2/internal/validation"
)

type ignore struct {
	path    string
	ignores []gitignore.IgnoreMatcher
}

func newIgnore(path string, ignoreFiles []string) ignore {
	ignores := make([]gitignore.IgnoreMatcher, 0, len(ignoreFiles))
	for _, ignoreFile := range ignoreFiles {
		i, err := gitignore.NewGitIgnore(filepath.Join(path, ignoreFile))
		if err != nil {
			continue
		}
		ignores = append(ignores, i)
	}
	return ignore{path: path, ignores: ignores}
}

func (m ignore) Match(path string, isDir bool) bool {
	for _, i := range m.ignores {
		if i.Match(path, isDir) {
			return true
		}
	}

	return false
}

func readFileLines(path string) ([]string, error) {
	if !validation.FileExists(path) {
		return nil, errors.New("file does not exist")
	}

	/* #nosec */
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, nil
}

func readFiles(ctx context.Context, files chan<- file, workspace string) error {
	defer close(files)
	ignoreFiles := []string{".gitignore", ".ignore", ".ldignore"}
	allIgnores := newIgnore(workspace, ignoreFiles)
	workspace = filepath.ToSlash(workspace)

	readFile := func(path string, info os.FileInfo, err error) error {
		if err != nil || ctx.Err() != nil {
			// global context cancelled, don't read any more files
			return nil
		}

		isDir := info.IsDir()
		path = filepath.ToSlash(path)

		// Skip directories, hidden files, and ignored files
		if strings.HasPrefix(info.Name(), ".") || allIgnores.Match(path, isDir) {
			if isDir {
				return filepath.SkipDir
			}
			return nil
		} else if !info.Mode().IsRegular() {
			return nil
		}

		lines, err := readFileLines(path)
		if err != nil {
			return err
		}

		// only read text files
		if !util.IsText([]byte(strings.Join(lines, "\n"))) {
			return nil
		}

		files <- file{path: strings.TrimPrefix(path, workspace+"/"), lines: lines}
		return nil
	}

	return filepath.Walk(workspace, readFile)
}
