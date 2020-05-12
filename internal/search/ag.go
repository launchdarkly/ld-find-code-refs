package search

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	"github.com/launchdarkly/ld-find-code-refs/internal/validation"
)

type AgClient struct {
	workspace string
}

func NewAgClient(path string) (*AgClient, error) {
	if !filepath.IsAbs(path) {
		log.Error.Fatalf("expected an absolute path but received a relative path: %s", path)
	}
	_, err := exec.LookPath("ag")
	if err != nil {
		return nil, errors.New("ag (The Silver Searcher) is a required dependency, but was not found in the system PATH")
	}

	return &AgClient{workspace: path}, nil
}

func (c *AgClient) FindReferences(flags []string, aliases map[string][]string, ctxLines int, delimiters string) (SearchResultLines, error) {
	log.Info.Printf("finding code references with delimiters: %s", delimiters)
	paginationCharCount := SafePaginationCharCount()
	results, err := c.paginatedSearch(flags, paginationCharCount, ctxLines, []byte(delimiters))
	if err != nil {
		return SearchResultLines{}, err
	}
	flattenedAliases := make([]string, 0, len(flags))
	for _, flagAliases := range aliases {
		flattenedAliases = append(flattenedAliases, flagAliases...)
	}
	aliasResults, err := c.paginatedSearch(flattenedAliases, paginationCharCount, ctxLines, nil)
	if err != nil {
		return SearchResultLines{}, err
	}
	results = append(results, aliasResults...)
	return c.generateReferences(aliases, results, ctxLines, delimiters), nil
}

// paginatedSearch uses approximations to decide the number of flags to scan for at once using maxSumFlagKeyLength as an upper bound
func (c *AgClient) paginatedSearch(flags []string, maxSumFlagKeyLength, ctxLines int, delims []byte) ([][]string, error) {
	searchType := "flags"
	if delims == nil {
		searchType = "aliases"
	}

	if maxSumFlagKeyLength == 0 {
		return nil, NoSearchPatternErr
	}

	var results [][]string
	nextSearchKeys := []string{}

	totalKeyLength := DelimCost(delims)
	from := 0
	for to, key := range flags {
		totalKeyLength += FlagKeyCost(key)
		nextSearchKeys = append(nextSearchKeys, key)

		// if we've reached the end of the loop, or the current page has reached maximum length
		if to == len(flags)-1 || totalKeyLength+FlagKeyCost(flags[to+1]) > maxSumFlagKeyLength {

			log.Debug.Printf("searching for %s in group: [%d, %d]", searchType, from, to)
			result, err := c.searchForFlags(nextSearchKeys, ctxLines, delims)
			if err != nil {
				if err == SearchTooLargeErr {
					// we expect all search implementations to complete successfully
					// if pagination fails unexpectedly, repeat the search with a smaller page size
					log.Debug.Printf("encountered an error paginating group [%d, %d], trying again with a lower page size", from, to)
					remainder, err := c.paginatedSearch(flags[from:], maxSumFlagKeyLength/2, ctxLines, delims)
					if err != nil {
						return nil, err
					}
					return append(results, remainder...), nil
				}
				return nil, err
			}

			results = append(results, result...)

			// loop bookkeeping
			nextSearchKeys = make([]string, 0, len(nextSearchKeys))
			totalKeyLength = DelimCost(delims)
			from = to + 1
		}
	}
	return results, nil
}

func (c *AgClient) searchForFlags(flags []string, ctxLines int, delimiters []byte) ([][]string, error) {
	args := []string{"--nogroup", "--case-sensitive"}
	pathToIgnore := filepath.Join(c.workspace, ignoreFileName)
	if validation.FileExists(pathToIgnore) {
		log.Debug.Printf("excluding files matched in %s", ignoreFileName)
		args = append(args, fmt.Sprintf("--path-to-ignore=%s", pathToIgnore))
	}
	if ctxLines > 0 {
		args = append(args, fmt.Sprintf("-C%d", ctxLines))
	}

	searchPattern := generateSearchPattern(flags, delimiters, runtime.GOOS == windows)
	/* #nosec */
	cmd := exec.Command("ag", args...)
	cmd.Args = append(cmd.Args, searchPattern, c.workspace)
	out, err := cmd.CombinedOutput()
	res := string(out)
	if err != nil {
		if err.Error() == "exit status 1" {
			return nil, nil
		} else if strings.Contains(res, SearchTooLargeErr.Error()) ||
			(runtime.GOOS == windows && strings.Contains(err.Error(), windowsSearchTooLargeErr.Error())) {
			return nil, SearchTooLargeErr
		}
		return nil, errors.New(res)
	}

	searchRegexWithFilteredPath, err := regexp.Compile("(?:" + regexp.QuoteMeta(c.workspace) + "/)" + searchRegex.String())
	if err != nil {
		return nil, err
	}

	output := string(out)
	if runtime.GOOS == windows {
		output = fromWindows1252(output)
	}

	ret := searchRegexWithFilteredPath.FindAllStringSubmatch(output, -1)
	return ret, err
}

func (c *AgClient) generateReferences(aliases map[string][]string, searchResult [][]string, ctxLines int, delims string) []SearchResultLine {
	references := []SearchResultLine{}

	for _, r := range searchResult {
		path := r[1]
		contextContainsFlagKey := r[2] == ":"
		lineNumber := r[3]
		lineText := r[4]
		lineNum, err := strconv.Atoi(lineNumber)
		if err != nil {
			log.Error.Fatalf("encountered an unexpected error generating flag references: %s", err)
		}
		ref := SearchResultLine{Path: path, LineNum: lineNum}
		if contextContainsFlagKey {
			ref.FlagKeys = c.findReferencedFlags(lineText, aliases, delims)
		}
		if ctxLines >= 0 {
			ref.LineText = lineText
		}
		references = append(references, ref)
	}

	return references
}

func (c *AgClient) findReferencedFlags(ref string, aliases map[string][]string, delims string) map[string][]string {
	ret := make(map[string][]string, len(aliases))
	for key, flagAliases := range aliases {
		matcher := regexp.MustCompile(regexp.QuoteMeta(key))
		if len(delims) > 0 {
			matcher = regexp.MustCompile(fmt.Sprintf("[%s]%s[%s]", delims, regexp.QuoteMeta(key), delims))
		}
		if matcher.MatchString(ref) {
			ret[key] = make([]string, 0, len(flagAliases))
		}
		for _, alias := range flagAliases {
			aliasMatcher := regexp.MustCompile(regexp.QuoteMeta(alias))
			if aliasMatcher.MatchString(ref) {
				if ret[key] == nil {
					ret[key] = make([]string, 0, len(flagAliases))
				}
				ret[key] = append(ret[key], alias)
			}
		}
	}
	return ret
}
