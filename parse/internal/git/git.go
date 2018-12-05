package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	log "github.com/launchdarkly/git-flag-parser/parse/internal/log"
)

type Git struct {
	Workspace string
	Head      string
	RepoName  string
}

/*
grepRegex splits resulting grep lines into groups
Group 1: File path.
Group 2: Seperator. A colon indicates a match, a hyphen indicates a context lines
Group 3: Line number
Group 4: Line contents
*/
// TODO: add a test that fails if last capture group is (.+)
var grepRegex, _ = regexp.Compile("(.+)(:|-)([0-9]+)[:-](.*)")

func (g Git) RevParse() (string, error) {
	g.logDebug("Parsing latest commit", nil)
	cmd := exec.Command("git", "rev-parse", g.Head)
	cmd.Dir = g.Workspace
	out, err := cmd.Output()
	if err != nil {
		g.logError("Failed to parse latest commit", err, nil)
	}
	return strings.TrimSpace(string(out)), nil
}

func (g Git) Clone(endpoint string) error {
	g.logDebug("Cloning repo", nil)
	cmd := exec.Command("git", "clone", "--depth", "1", "--single-branch", "--branch", g.Head, endpoint, g.Workspace)
	err := cmd.Run()
	if err != nil {
		g.logError("Failed to clone repo", err, nil)
	}
	return err
}

func (g Git) Checkout() error {
	g.logDebug("Checking out to selected branch", nil)
	checkout := exec.Command("git", "checkout", g.Head)
	checkout.Dir = g.Workspace
	err := checkout.Run()
	if err != nil {
		g.logError("Failed to checkout selected branch", err, nil)
	}
	return err
}

func (g Git) Grep(flags []string, ctxLines int) ([][]string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("ag --nogroup"))
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
	g.logDebug("Grepping for flag keys", map[string]interface{}{"numFlags": len(escapedFlags), "contextLines": ctxLines, "cmd": cmd})
	sh.Dir = g.Workspace
	out, err := sh.Output()
	if err != nil {
		if err.Error() == "exit status 1" {
			return [][]string{}, nil
		}
		g.logError("Error grepping for flag keys", err, map[string]interface{}{"numFlags": len(escapedFlags), "contextLines": ctxLines})
		return nil, err
	}
	grepRegexWithFilteredPath, err := regexp.Compile("(?:" + regexp.QuoteMeta(g.Workspace) + "/)" + grepRegex.String())
	if err != nil {
		return nil, err
	}
	ret := grepRegexWithFilteredPath.FindAllStringSubmatch(string(out), -1)
	return ret, err
}

func (g Git) logDebug(msg string, fields map[string]interface{}) {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	fields["dir"] = g.Workspace
	fields["branch"] = g.Head
	log.Debug(msg, fields)
}

func (g Git) logError(msg string, err error, fields map[string]interface{}) {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	fields["dir"] = g.Workspace
	fields["branch"] = g.Head
	log.Error(msg, err, fields)
}
