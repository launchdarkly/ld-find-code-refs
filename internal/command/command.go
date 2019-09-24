package command

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	"github.com/launchdarkly/ld-find-code-refs/internal/validation"
)

/*
searchRegex splits search result lines into groups
Group 1: File path.
Group 2: Separator. A colon indicates a match, a hyphen indicates a context lines
Group 3: Line number
Group 4: Line contents
*/
var searchRegex = regexp.MustCompile("([^:]+)(:|-)([0-9]+)[:-](.*)")

var SearchTooLargeErr = errors.New("regular expression is too large")

// SafePaginationCharCount determines the maximum sum of flag key lengths to be used in a single smart paginated search.
// Safely bounded under the 2^16 limit of pcre_compile() with the parameters set by our underlying search tool (ag)
// https://github.com/vmg/pcre/blob/master/pcre_internal.h#L436
func SafePaginationCharCount() int {
	if runtime.GOOS == windows {
		// workaround win32 limitation on maximum command length
		// https://support.microsoft.com/en-us/help/830473/command-prompt-cmd-exe-command-line-string-limitation
		return 30000
	}

	return 60000
}

func FlagKeyCost(key string) int {
	// periods need to be escaped, so they count as 2 characters
	return len(key) + strings.Count(key, ".")
}

func DelimCost(delims []rune) int {
	return len(delims) * 2
}

type Searcher interface {
	SearchForFlags(flags []string, ctxLines int, delimiters []rune) ([][]string, error)
}

type AgClient struct {
	workspace string
}

func NewAgClient(path string) (*AgClient, error) {
	if !filepath.IsAbs(path) {
		log.Fatal.Fatalf("expected an absolute path but received a relative path: %s", path)
	}
	_, err := exec.LookPath("ag")
	if err != nil {
		return nil, errors.New("ag (The Silver Searcher) is a required dependency, but was not found in the system PATH")
	}

	return &AgClient{workspace: path}, nil
}

func (c *AgClient) SearchForFlags(flags []string, ctxLines int, delimiters []rune) ([][]string, error) {
	args := []string{"--nogroup", "--case-sensitive"}
	ignoreFileName := ".ldignore"
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

func generateFlagRegex(flags []string) string {
	flagRegexes := []string{}
	for _, v := range flags {
		escapedFlag := regexp.QuoteMeta(v)
		flagRegexes = append(flagRegexes, escapedFlag)
	}
	return strings.Join(flagRegexes, "|")
}

func generateDelimiterRegex(delimiters []rune) (lookBehind, lookAhead string) {
	delims := string(delimiters)
	lookBehind = fmt.Sprintf("(?<=[%s])", delims)
	lookAhead = fmt.Sprintf("(?=[%s])", delims)
	return lookBehind, lookAhead
}

func generateSearchPattern(flags []string, delimiters []rune, padPattern bool) string {
	flagRegex := generateFlagRegex(flags)
	lookBehind, lookAhead := generateDelimiterRegex(delimiters)
	if padPattern {
		// Padding the left-most and right-most search terms with the "a^" regular expression, which never matches anything. This is done to work-around strange behavior causing the left-most and right-most items to be ignored by ag on windows
		// example: (?<=[\"'\`])(a^|flag1|flag2|flag3|a^)(?=[\"'\`])"
		return lookBehind + "(a^|" + flagRegex + "|a^)" + lookAhead
	}
	// example: (?<=[\"'\`])(flag1|flag2|flag3)(?=[\"'\`])"
	return lookBehind + "(" + flagRegex + ")" + lookAhead
}
