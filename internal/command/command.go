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

/*
grepRegex splits resulting grep lines into groups
Group 1: File path.
Group 2: Separator. A colon indicates a match, a hyphen indicates a context lines
Group 3: Line number
Group 4: Line contents
*/
var grepRegex = regexp.MustCompile("([^:]+)(:|-)([0-9]+)[:-](.*)")

type Client struct {
	Workspace string
	GitBranch string
	GitSha    string
}

func NewClient(path string) (Client, error) {
	client := Client{}

	absPath, err := normalizeAndValidatePath(path)
	if err != nil {
		return client, fmt.Errorf("could not validate directory option: %s", err)
	}
	client.Workspace = absPath

	_, err = exec.LookPath("git")
	if err != nil {
		return client, errors.New("git is a required dependency, but was not found in the system PATH")
	}
	_, err = exec.LookPath("ag")
	if err != nil {
		return client, errors.New("ag (The Silver Searcher) is a required dependency, but was not found in the system PATH")
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

func (c Client) SearchForFlags(flags []string, ctxLines int, delimiters []rune) ([][]string, error) {
	args := []string{"--nogroup", "--case-sensitive"}
	pathToIgnore := filepath.Join(c.Workspace, ".ldignore")
	if fileExists(pathToIgnore) {
		log.Debug.Printf("excluding files matched in .ldignore")
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
	if err != nil {
		if err.Error() == "exit status 1" {
			return [][]string{}, nil
		}
		return nil, errors.New(string(out))
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
