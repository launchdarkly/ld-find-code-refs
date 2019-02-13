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
	out, err := cmd.Output()
	if err != nil {
		return "", err
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
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	ret := strings.TrimSpace(string(out))
	log.Debug.Printf("identified sha: %s", ret)
	return ret, nil
}

func (c Client) SearchForFlags(flags []string, ctxLines int, delimiters []string) ([][]string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("ag --nogroup --case-sensitive"))
	if ctxLines > 0 {
		sb.WriteString(fmt.Sprintf(" -C%d", ctxLines))
	}

	searchPattern := generateSearchPattern(flags, delimiters, runtime.GOOS == windows)
	fmt.Println(searchPattern)
	var command *exec.Cmd
	if runtime.GOOS == windows {
		args := strings.Split(sb.String(), " ")
		/* #nosec */
		command = exec.Command(args[0], args[1:]...)
		command.Args = append(command.Args, searchPattern, c.Workspace)
	} else {
		sb.WriteString(` "` + searchPattern + `" ` + c.Workspace)
		/* #nosec */
		command = exec.Command("sh", "-c", sb.String())
	}
	out, err := command.Output()
	if err != nil {
		if err.Error() == "exit status 1" {
			return [][]string{}, nil
		}
		return nil, err
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

func generateFlagRegex(flags []string) string {
	flagRegexes := []string{}
	for _, v := range flags {
		escapedFlag := regexp.QuoteMeta(v)
		flagRegexes = append(flagRegexes, escapedFlag)
	}
	return strings.Join(flagRegexes, "|")
}

func generateDelimiterRegex(delimiters []string) (lookBehind, lookAhead string) {
	var escapedDelims = make([]string, 0, len(delimiters))
	for i, v := range delimiters {
		escapedDelims = append(escapedDelims, regexp.QuoteMeta(v))
		// escaping for command line ag
		switch escapedDelims[i] {
		case `"`:
			escapedDelims[i] = "\\\""
		case "`":
			escapedDelims[i] = "\\`"
		}
	}
	delims := strings.Join(escapedDelims, "")
	lookBehind = fmt.Sprintf("(?<=[%s])", delims)
	lookAhead = fmt.Sprintf("(?=[%s])", delims)
	return lookBehind, lookAhead
}

func generateSearchPattern(flags, delimiters []string, padPattern bool) string {
	flagRegex := generateFlagRegex(flags)
	lookBehind, lookAhead := generateDelimiterRegex(delimiters)
	if padPattern {
		// Padding the left-most and right-most search terms with the "?!" regular expression, which never matches anything. This is done to work-around strange behavior causing the left-most and right-most items to be ignored by ag on windows
		// example: (?<=[\"'\`])(?!|flag1|flag2|flag3|?1(?=[\"'\`])"
		return lookBehind + "('?!|" + flagRegex + "|?!')" + lookAhead
	}
	// example: (?<=[\"'\`])(flag1|flag2|flag3(?=[\"'\`])"
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
