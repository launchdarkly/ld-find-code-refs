package git

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/launchdarkly/ld-find-code-refs/internal/log"
)

/*
grepRegex splits resulting grep lines into groups
Group 1: File path.
Group 2: Seperator. A colon indicates a match, a hyphen indicates a context lines
Group 3: Line number
Group 4: Line contents
*/
var grepRegex, _ = regexp.Compile("([^:]+)(:|-)([0-9]+)[:-](.*)")

type Git struct {
	Workspace string
}

func NewClient(path string) (Git, error) {
	client := Git{path}
	_, err := exec.LookPath("git")
	if err != nil {
		return client, errors.New("git is a required dependency, but was not found in the system PATH")
	}
	_, err = exec.LookPath("ag")
	if err != nil {
		return client, errors.New("ag (The Silver Searcher) is a required dependency, but was not found in the system PATH")
	}

	return client, nil
}

func (g Git) BranchName() (string, error) {
	cmd := exec.Command("git", "-C", g.Workspace, "rev-parse", "--abbrev-ref", "HEAD")
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

func (g Git) RevParse(branch string) (string, error) {
	cmd := exec.Command("git", "-C", g.Workspace, "rev-parse", branch)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	ret := strings.TrimSpace(string(out))
	log.Debug.Printf("identified sha: %s", ret)
	return ret, nil
}

func (g Git) SearchForFlags(flags []string, ctxLines int) ([][]string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("ag --nogroup --case-sensitive"))
	if ctxLines > 0 {
		sb.WriteString(fmt.Sprintf(" -C%d", ctxLines))
	}

	escapedFlags := []string{}
	for _, v := range flags {
		escapedFlags = append(escapedFlags, regexp.QuoteMeta(v))
	}
	sb.WriteString(" '" + strings.Join(escapedFlags, "|") + "' " + g.Workspace)

	cmd := sb.String()
	sh := exec.Command("sh", "-c", cmd)
	out, err := sh.Output()
	if err != nil {
		if err.Error() == "exit status 1" {
			return [][]string{}, nil
		}
		return nil, err
	}
	grepRegexWithFilteredPath, err := regexp.Compile("(?:" + regexp.QuoteMeta(g.Workspace) + "/)" + grepRegex.String())
	if err != nil {
		return nil, err
	}
	ret := grepRegexWithFilteredPath.FindAllStringSubmatch(string(out), -1)
	return ret, err
}
