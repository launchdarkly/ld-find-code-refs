package search

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/launchdarkly/ld-find-code-refs/internal/validation"
	"github.com/monochromegane/go-gitignore"
	"golang.org/x/tools/godoc/util"
	"golang.org/x/tools/godoc/vfs"
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

func (f file) linesIfMatch(aliases map[string][]string, aliasToFlag map[string]string, matchLineNum, ctxLines int, match, key, delimiters string) SearchResultLines {
	if !strings.Contains(match, key) {
		return nil
	}
	startingLineNum := matchLineNum - ctxLines
	context := f.lines[startingLineNum : matchLineNum+ctxLines+1]
	linesForMatch := make([]SearchResultLine, 0, len(context))
	flagKey := key
	_, isFlagKey := aliases[key]
	if isFlagKey && !matchHasDelimiters(match, flagKey, delimiters) {
		return nil
	} else if !isFlagKey {
		flagKey = aliasToFlag[key]
	}

	for i, line := range context {
		srl := SearchResultLine{
			Path:     f.path,
			LineNum:  startingLineNum + i,
			LineText: strings.TrimSuffix(line, "\n"),
		}
		if i == ctxLines {
			if srl.FlagKeys[flagKey] == nil {
				srl.FlagKeys = map[string][]string{flagKey: {}}
			}
			if !isFlagKey {
				srl.FlagKeys[flagKey] = append(srl.FlagKeys[flagKey], key)
			}
		}
		linesForMatch = append(linesForMatch, srl)
	}
	return linesForMatch
}

func matchHasDelimiters(match string, flagKey string, delimiters string) bool {
	for _, left := range delimiters {
		for _, right := range delimiters {
			if strings.Contains(match, string(left)+flagKey+string(right)) {
				return true
			}
		}
	}
	return false
}

func (f file) toSearchResultLines(projKey string, aliases map[string][]string, ctxLines int, delimiters string) SearchResultLines {
	ret := SearchResultLines{}

	ctxLinesString := ""
	for i := 0; i <= ctxLines; i++ {
		ctxLinesString = ctxLinesString + ".*\\n?"
	}
	flagKeys := []string{}
	flattenedAliases := []string{}
	aliasToFlag := map[string]string{}
	for flagKey, flagAliases := range aliases {
		flagKeys = append(flagKeys, regexp.QuoteMeta(flagKey))
		for _, alias := range flagAliases {
			flattenedAliases = append(flattenedAliases, regexp.QuoteMeta(alias))
			aliasToFlag[alias] = flagKey
		}
	}
	for _, key := range append(flagKeys, flattenedAliases...) {
		for i, line := range f.lines {
			match := f.linesIfMatch(aliases, aliasToFlag, i, ctxLines, line, key, delimiters)
			if match != nil {
				ret = append(ret, match...)
			}
		}
	}

	if len(ret) == 0 {
		return nil
	}
	return ret
}

type opener struct{}

func (o opener) Open(name string) (vfs.ReadSeekCloser, error) {
	return os.Open(name)
}

func SearchForRefs(projKey, workspace string, searchTerms []string, aliases map[string][]string, ctxLines int, delimiters []byte) (SearchResultLines, error) {
	ignoreFiles := []string{".gitignore", ".ignore", ".ldignore"}
	allIgnores := newIgnore(workspace, ignoreFiles)
	// maxConcurrency := runtime.NumCPU()

	fullStartTime := time.Now()
	files := make(chan file, 100000)
	references := make(chan SearchResultLines)
	// Start a workers to process files asynchronously
	go func() {
		w := new(sync.WaitGroup)
		for file := range files {
			file := file
			w.Add(1)
			go func() {
				reference := file.toSearchResultLines(projKey, aliases, ctxLines, string(delimiters))
				if reference != nil {
					references <- reference
				}
				w.Done()
			}()
		}
		w.Wait()
		delta := time.Now().Sub(fullStartTime).Milliseconds()
		fmt.Println("done with refs", delta)
		close(references)
	}()

	fileWg := sync.WaitGroup{}
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

		fileWg.Add(1)
		go func() error {
			defer fileWg.Done()
			lines, err := readFileLines(path)
			if err != nil {
				return err
			}

			if !util.IsText([]byte(strings.Join(lines, "\n"))) {
				return nil
			}

			files <- file{path: strings.TrimPrefix(path, workspace+"/"), lines: lines}
			return nil
		}()
		return nil
	}

	err := filepath.Walk(workspace, readFile)
	if err != nil {
		return nil, err
	}
	fileWg.Wait()
	close(files)

	delta := time.Now().Sub(fullStartTime).Milliseconds()
	fmt.Println("done reading files", delta)

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
