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
	o "github.com/launchdarkly/ld-find-code-refs/internal/options"
)

/*
grepRegex splits resulting grep lines into groups
Group 1: File path.
Group 2: Separator. A colon indicates a match, a hyphen indicates a context lines
Group 3: Line number
Group 4: Line contents
*/
var grepRegex = regexp.MustCompile("([^:]+)(:|-)([0-9]+)[:-](.*)")

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

type Client interface {
	SearchForFlags(flags []string, ctxLines int, delimiters []rune) ([][]string, error)
	RemoteBranches() (map[string]bool, error)
}

type gitClient struct {
	Workspace string
	GitBranch string
	GitSha    string
}

func newGitClient(path string) (gitClient, error) {
	client := gitClient{}

	absPath, err := normalizeAndValidatePath(path)
	if err != nil {
		return client, fmt.Errorf("could not validate directory option: %s", err)
	}
	client.Workspace = absPath

	_, err = exec.LookPath("git")
	if err != nil {
		return client, errors.New("git is a required dependency, but was not found in the system PATH")
	}

	currBranch, err := client.branchName()
	if err != nil {
		return client, fmt.Errorf("error parsing git branch name: %s", err)
	} else if currBranch == "" {
		return client, fmt.Errorf("error parsing git branch name: git repo at %s must be checked out to a valid branch or --branch option must be set", client.Workspace)
	}
	log.Info.Printf("git branch: %s", currBranch)
	client.GitBranch = currBranch

	head, err := client.headSha()
	if err != nil {
		return client, fmt.Errorf("error parsing current commit sha: %s", err)
	}
	client.GitSha = head

	return client, nil
}

func (c gitClient) branchName() (string, error) {
	// Some CI systems leave the repository in a detached HEAD state. To support those, this logic allows
	// users to pass the branch name in by hand as an option.
	if o.Branch.Value() != "" {
		return o.Branch.Value(), nil
	}

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

func (c gitClient) headSha() (string, error) {
	/* #nosec */
	cmd := exec.Command("git", "-C", c.Workspace, "rev-parse", "HEAD")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New(string(out))
	}
	ret := strings.TrimSpace(string(out))
	log.Debug.Printf("identified head sha: %s", ret)
	return ret, nil
}

func (c gitClient) RemoteBranches() (map[string]bool, error) {
	/* #nosec */
	cmd := exec.Command("git", "-C", c.Workspace, "ls-remote", "--quiet", "--heads")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.New(string(out))
	}
	rgx := regexp.MustCompile("refs/heads/(.*)")
	results := rgx.FindAllStringSubmatch(string(out), -1)
	log.Debug.Printf("found %d branches on remote", len(results))
	ret := map[string]bool{}
	for _, r := range results {
		ret[r[1]] = true
	}
	// the current branch should be in the list of remote branches
	ret[c.GitBranch] = true
	return ret, nil
}

type AgClient struct {
	gitClient
}

func NewAgClient(path string) (*AgClient, error) {
	gitClient, err := newGitClient(path)
	if err != nil {
		return nil, err
	}

	_, err = exec.LookPath("ag")
	if err != nil {
		return nil, errors.New("ag (The Silver Searcher) is a required dependency, but was not found in the system PATH")
	}

	return &AgClient{gitClient}, nil
}

func (c *AgClient) SearchForFlags(flags []string, ctxLines int, delimiters []rune) ([][]string, error) {
	args := []string{"--nogroup", "--case-sensitive"}
	ignoreFileName := ".ldignore"
	pathToIgnore := filepath.Join(c.Workspace, ignoreFileName)
	if fileExists(pathToIgnore) {
		log.Debug.Printf("excluding files matched in %s", ignoreFileName)
		args = append(args, fmt.Sprintf("--path-to-ignore=%s", pathToIgnore))
	}
	if ctxLines > 0 {
		args = append(args, fmt.Sprintf("-C%d", ctxLines))
	}

	searchPattern := generateSearchPattern(flags, delimiters, runtime.GOOS == windows)
	/* #nosec */
	cmd := exec.Command("ag", args...)
	cmd.Args = append(cmd.Args, searchPattern, c.Workspace)
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

	grepRegexWithFilteredPath, err := regexp.Compile("(?:" + regexp.QuoteMeta(c.Workspace) + "/)" + grepRegex.String())
	if err != nil {
		return nil, err
	}

	output := string(out)
	if runtime.GOOS == windows {
		output = fromWindows1252(output)
	}

	ret := grepRegexWithFilteredPath.FindAllStringSubmatch(output, -1)
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
