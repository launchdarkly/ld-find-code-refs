package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type Git struct {
	Workspace string
	// Head      string
}

/*
grepRegex splits resulting grep lines into groups
Group 1: File path.
Group 2: Seperator. A colon indicates a match, a hyphen indicates a context lines
Group 3: Line number
Group 4: Line contents
*/
var grepRegex, _ = regexp.Compile("([^:]+)(:|-)([0-9]+)[:-](.*)")

func (g Git) BranchName() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = g.Workspace
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (g Git) RevParse(branch string) (string, error) {
	cmd := exec.Command("git", "rev-parse", branch)
	cmd.Dir = g.Workspace
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
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
	sh.Dir = g.Workspace
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
