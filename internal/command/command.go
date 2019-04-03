package command

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/launchdarkly/ld-find-code-refs/internal/log"
)

type Client struct {
	Workspace string
	GitBranch string
	GitSha    string
	SearchTool
}

func NewClient(path string, searchTool SearchTool) (Client, error) {
	client := Client{SearchTool: searchTool}

	absPath, err := normalizeAndValidatePath(path)
	if err != nil {
		return client, fmt.Errorf("could not validate directory option: %s", err)
	}
	client.Workspace = absPath

	_, err = exec.LookPath("git")
	if err != nil {
		return client, errors.New("git is a required dependency, but was not found in the system PATH")
	}

	_, err = exec.LookPath(string(searchTool))
	if err != nil {
		switch searchTool {
		case SilverSearcher:
			return client, errors.New("ag (The Silver Searcher) is a required dependency, but was not found in the system PATH")
		case Ripgrep:
			return client, errors.New("rg (ripgrep) is a required dependency, but was not found in the system PATH")
		default:
			return client, fmt.Errorf("%s is not a valid search tool", searchTool)
		}

	}

	currBranch, err := client.branchName()
	if err != nil {
		return client, fmt.Errorf("error parsing git branch name: %s", err)
	} else if currBranch == "" {
		return client, fmt.Errorf("error parsing git branch name: git repo at %s must be checked out to a valid branch", client.Workspace)
	}
	client.GitBranch = currBranch

	headSha, err := client.revParse(currBranch)
	if err != nil {
		return client, fmt.Errorf("error parsing current commit sha: %s", err)
	}
	client.GitSha = headSha

	return client, nil
}

func (c Client) branchName() (string, error) {
	/* #nosec */
	cmd := exec.Command("git", "-C", c.Workspace, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New(string(out))
	}
	ret := strings.TrimSpace(string(out))
	log.Debug.Printf("identified branch name: %s", ret)
	if ret == "HEAD" {
		return "", nil
	}
	return ret, nil
}

func (c Client) revParse(branch string) (string, error) {
	/* #nosec */
	cmd := exec.Command("git", "-C", c.Workspace, "rev-parse", branch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New(string(out))
	}
	ret := strings.TrimSpace(string(out))
	log.Debug.Printf("identified sha: %s", ret)
	return ret, nil
}

type SearchTool string

const (
	SilverSearcher = "ag"
	Ripgrep        = "rg"
)

// globalSearchArgs returns a standardized set of arguments to be run with each search tool
func (c Client) globalSearchArgs() []string {
	argsForClient := map[SearchTool][]string{
		SilverSearcher: {"--nogroup", "--case-sensitive"},
		Ripgrep:        {"--no-heading", "--case-sensitive", "--line-number", "--pcre2"},
	}
	return argsForClient[c.SearchTool]
}

// ignoreFileArg returns the argument to respect the .ldignore file in searches
func (c Client) ignoreFileArg(ignoreFile string) string {
	ignoreArgForClient := map[SearchTool]string{
		SilverSearcher: fmt.Sprintf("--path-to-ignore=%s", ignoreFile),
		Ripgrep:        fmt.Sprintf("--ignore-file=%s", ignoreFile),
	}
	return ignoreArgForClient[c.SearchTool]
}

// contextLinesArg returns the argument to include context lines in the search result:w
func (c Client) contextLineArg(contextLines int) string {
	return fmt.Sprintf("-C%d", contextLines)
}

/*
	SearchResults is an abstraction for the [][]string returned by a regex
	on the results of a code reference search.
	Each reference has it's results split into and array of the following shape:
	Index 0: Full match
	Index 1: File path.
	Index 2: Line number.
	Index 3: Separator. A colon indicates a match, a hyphen indicates a context lines
	Index 4: Line contents.
*/
type SearchResults [][]string

func (c Client) SearchForFlags(flags []string, ctxLines int, delimiters []rune) (SearchResults, error) {
	args := c.globalSearchArgs()
	ignoreFileName := ".ldignore"
	pathToIgnore := filepath.Join(c.Workspace, ignoreFileName)
	if fileExists(pathToIgnore) {
		log.Debug.Printf("excluding files matched in %s", ignoreFileName)
		args = append(args, c.ignoreFileArg(pathToIgnore))
	}
	if ctxLines > 0 {
		args = append(args, c.contextLineArg(ctxLines))
	}

	searchPattern := generateSearchPattern(flags, delimiters, runtime.GOOS == windows)
	/* #nosec */
	cmd := exec.Command(string(c.SearchTool), args...)
	cmd.Args = append(cmd.Args, searchPattern, c.Workspace)

	out, err := cmd.CombinedOutput()
	if err != nil {
		if err.Error() == "exit status 1" {
			return [][]string{}, nil
		}
		return nil, errors.New(string(out))
	}

	/*
	   searchRegex splits resulting grep lines into groups
	   Group 1: File path.
	   Group 2: Line number.
	   Group 3: Separator. A colon indicates a match, a hyphen indicates a context lines
	   Group 4: Line contents.

	   Example output from line search tool: /path/to/file/flags.go:2:ldClient.BoolVariation("my-flag-key", user, false)
	   Example result: ["/path/to/file/flags.go:2:ldClient.BoolVariation(\"my-flag-key\", user, false)", "/path/to/file/flags.go", 2, ":", "ldClient.BoolVariation(\"my-flag-key\", user, false)"]
	*/
	searchRegex := regexp.MustCompile("(.*)[:-]([0-9]+)(:|-)(.*)")

	searchRegexWithFilteredPath, err := regexp.Compile("(?:" + regexp.QuoteMeta(c.Workspace) + "[\\/]?)" + searchRegex.String())
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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
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

func normalizeAndValidatePath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid directory: %s", err)
	}
	log.Info.Printf("absolute directory path: %s", absPath)

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
