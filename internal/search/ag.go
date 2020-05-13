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

// Code reference search using ag (the silver searcher): https://github.com/ggreer/the_silver_searcher
// We configure ag to run with the following flags:
// --case-sensitive  as feature flags and aliases are case sensitive
// --nogroup         for cleanlier parsing of output
// --path-to-ignore  for ignore files in .ldignore
// -C                for providing context lines

/* example
$ ag --nogroup --case-sensitive -C1 ${search.go#generateSearchPattern} ${dir}
> test/flags.go:5-        AnotherFlag = "another-flag"
> test/flags.go:6:        MyFlag = "my-flag"
> test/flags.go:7-        YetAnotherFlag = "yet-another-flag"
>
> test/index.ts:2-// export my flag
> test/index.ts:3:export const isMyFlagEnabled = flagPredicate('my-flag');
> test/index.ts:4-console.log('exported my flag');
*/

/*
agSearchRegex splits search result lines into groups. See the above example to compare these groups to real output
Group 1: File path.
Group 2: Separator. A colon indicates a match, a hyphen indicates a context line
Group 3: Line number
Group 4: Line contents
*/
var agSearchRegex = regexp.MustCompile("([^:]+)(:|-)([0-9]+)[:-](.*)")

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

func (c *AgClient) searchForRefs(searchTerms []string, aliases map[string][]string, ctxLines int, delimiters []byte) (SearchResultLines, error) {
	args := []string{"--nogroup", "--case-sensitive"}
	pathToIgnore := filepath.Join(c.workspace, ignoreFileName)
	if validation.FileExists(pathToIgnore) {
		log.Debug.Printf("excluding files matched in %s", ignoreFileName)
		args = append(args, fmt.Sprintf("--path-to-ignore=%s", pathToIgnore))
	}
	if ctxLines > 0 {
		args = append(args, fmt.Sprintf("-C%d", ctxLines))
	}

	searchPattern := generateSearchPattern(searchTerms, delimiters, runtime.GOOS == windows)
	/* #nosec */
	cmd := exec.Command("ag", args...)
	cmd.Args = append(cmd.Args, searchPattern, c.workspace)
	out, err := cmd.CombinedOutput()
	res := string(out)
	if err != nil {
		if err.Error() == "exit status 1" {
			return nil, nil
		} else if strings.Contains(res, searchTooLargeErr.Error()) ||
			(runtime.GOOS == windows && strings.Contains(err.Error(), windowsSearchTooLargeErr.Error())) {
			return nil, searchTooLargeErr
		}
		return nil, errors.New(res)
	}

	searchRegexWithFilteredPath, err := regexp.Compile("(?:" + regexp.QuoteMeta(c.workspace) + "/)" + agSearchRegex.String())
	if err != nil {
		return nil, err
	}

	output := string(out)
	if runtime.GOOS == windows {
		output = fromWindows1252(output)
	}

	ret := searchRegexWithFilteredPath.FindAllStringSubmatch(output, -1)

	return c.generateReferences(aliases, ret, ctxLines, string(delimiters)), err
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
