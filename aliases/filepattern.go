package aliases

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/log"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/validation"
	"github.com/launchdarkly/ld-find-code-refs/v2/options"
)

var filepathGlobCache = make(map[string][]string)
var absGlobContents = make(map[string][]byte)

func GenerateAliasesFromFilePattern(a options.Alias, flag, dir string, allFileContents FileContentsMap) ([]string, error) {
	ret := make([]string, 0)
	// Concatenate the contents of all files into a single byte array to be matched by specified patterns
	fileContents := []byte{}
	fmt.Printf("looking for aliases for flag %s\n", flag)
	for _, path := range a.Paths {
		fmt.Printf("finding aliases for path %s\n", path)
		absGlob := filepath.Join(dir, path)
		if contents, ok := absGlobContents[absGlob]; ok {
			fmt.Println("using found contents")
			fileContents = append(fileContents, contents...)
		} else {
			fmt.Println("searching for contents")
			matches, err := filepathGlob(dir, path)
			if err != nil {
				return nil, fmt.Errorf("filepattern '%s': could not process path glob '%s'", a.Name, path)
			}

			contents := []byte{}
			for _, match := range matches {
				if pathFileContents := allFileContents[match]; len(pathFileContents) > 0 {
					contents = append(contents, pathFileContents...)
				}
			}
			absGlobContents[absGlob] = contents
			fileContents = append(fileContents, contents...)
		}
	}

	for _, p := range a.Patterns {
		pattern := regexp.MustCompile(strings.ReplaceAll(p, "FLAG_KEY", flag))
		results := pattern.FindAllStringSubmatch(string(fileContents), -1)
		for _, res := range results {
			if len(res) > 1 {
				ret = append(ret, res[1:]...)
			}
		}
	}

	return ret, nil
}

// processFileContent reads and stores the content of files specified by filePattern alias matchers to be matched for aliases
func processFileContent(aliases []options.Alias, dir string) (FileContentsMap, error) {
	allFileContents := map[string][]byte{}
	for idx, a := range aliases {
		if a.Type.Canonical() != options.FilePattern {
			continue
		}

		aliasId := strconv.Itoa(idx)
		if a.Name != "" {
			aliasId = a.Name
		}

		paths := []string{}
		for _, glob := range a.Paths {
			matches, err := filepathGlob(dir, glob)
			if err != nil {
				return nil, fmt.Errorf("filepattern '%s': could not process path glob '%s'", aliasId, glob)
			}
			if matches == nil {
				log.Info.Printf("filepattern '%s': no matching files found for alias path glob '%s'", aliasId, glob)
			}
			paths = append(paths, matches...)
		}

		for _, path := range paths {
			_, pathAlreadyProcessed := allFileContents[path]
			if pathAlreadyProcessed {
				continue
			}

			if !validation.FileExists(path) {
				return nil, fmt.Errorf("filepattern '%s': could not find file at path '%s'", aliasId, path)
			}
			/* #nosec */
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("filepattern '%s': could not process file at path '%s': %v", aliasId, path, err)
			}
			allFileContents[path] = data
		}
	}
	return allFileContents, nil
}

func filepathGlob(dir, glob string) ([]string, error) {
	fmt.Printf("looking for glob %s\n", glob)
	absGlob := filepath.Join(dir, glob)
	if cachedFilePaths, ok := filepathGlobCache[absGlob]; ok {
		fmt.Println("found cache")
		return cachedFilePaths, nil
	}

	matches, err := doublestar.FilepathGlob(absGlob)
	filepathGlobCache[absGlob] = matches

	return matches, err
}
