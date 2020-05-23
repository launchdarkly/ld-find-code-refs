package search

import (
	"bufio"
	"errors"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/launchdarkly/ld-find-code-refs/internal/validation"
	"github.com/monochromegane/go-gitignore"
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

type file struct {
	path  string
	lines []string
}

func (f file) toSearchResultLines(aliases map[string][]string, ctxLines int, delimiters []byte) SearchResultLines {
	path := f.path
	lines := f.lines
	ret := SearchResultLines{}
	for lineNum, line := range f.lines {
		matchFlagKeys := map[string][]string{}
		for flagKey, flagAliases := range aliases {
			rgxp := regexp.MustCompile("[" + string(delimiters) + "]" + regexp.QuoteMeta(flagKey) + "[" + string(delimiters) + "]")
			if rgxp.MatchString(line) {
				matchFlagKeys[flagKey] = []string{}
			}
			for _, alias := range flagAliases {
				rgxp := regexp.MustCompile(regexp.QuoteMeta(alias))
				if rgxp.MatchString(line) {
					_, ok := matchFlagKeys[flagKey]
					if !ok {
						matchFlagKeys[flagKey] = []string{}
					}
					matchFlagKeys[flagKey] = append(matchFlagKeys[flagKey], alias)
				}
			}
		}

		if len(matchFlagKeys) > 0 {
			startingContextLine := int(math.Max(float64(lineNum-ctxLines), 0))
			endingContextLine := int(math.Min(float64(len(lines)), float64(lineNum+ctxLines))) + 1
			matchLines := lines[startingContextLine:endingContextLine]
			linesForMatch := make([]SearchResultLine, 0, len(matchLines))
			for i, line := range matchLines {
				srl := SearchResultLine{
					Path:     path,
					LineNum:  startingContextLine + i + 1,
					LineText: strings.TrimSuffix(line, "\n"),
				}
				if startingContextLine+i == lineNum {
					srl.FlagKeys = matchFlagKeys
				}
				linesForMatch = append(linesForMatch, srl)
			}
			ret = append(ret, linesForMatch...)
		}
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}

func SearchForRefs(workspace string, searchTerms []string, aliases map[string][]string, ctxLines int, delimiters []byte) (SearchResultLines, error) {
	ignoreFiles := []string{".gitignore", ".ignore", ".ldignore"}
	allIgnores := newIgnore(workspace, ignoreFiles)
	maxConcurrency := runtime.NumCPU()

	files := make(chan file, maxConcurrency)
	references := make(chan SearchResultLines, maxConcurrency)

	// Start a workers to process files asynchronously
	go func() {
		w := new(sync.WaitGroup)
		for file := range files {
			file := file
			w.Add(1)
			go func() {
				reference := file.toSearchResultLines(aliases, ctxLines, delimiters)
				if reference != nil {
					references <- reference
				}
				w.Done()
			}()
		}
		w.Wait()
		close(references)
	}()

	readFile := func(path string, info os.FileInfo, err error) error {
		isDir := info.IsDir()
		if strings.HasPrefix(info.Name(), ".") || allIgnores.Match(path, isDir) {
			if isDir {
				return filepath.SkipDir
			}
			return nil
		} else if isDir {
			return nil
		}

		lines, err := readFileLines(path)
		if err != nil {
			return err
		}
		files <- file{path: strings.TrimPrefix(path, workspace+"/"), lines: lines}
		return nil
	}

	err := filepath.Walk(workspace, readFile)
	close(files)
	if err != nil {
		return nil, err
	}

	ret := SearchResultLines{}
	for reference := range references {
		ret = append(ret, reference...)
	}
	return ret, nil
}

func readFileLines(path string) ([]string, error) {
	if !validation.FileExists(path) {
		return nil, errors.New("file does not exist")
	}

	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtlines []string

	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}

	return txtlines, nil
}
