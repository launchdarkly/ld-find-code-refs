package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	log "github.com/launchdarkly/git-flag-parser/parse/internal/log"
)

type Commander struct {
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
var grepRegex, _ = regexp.Compile("(.+)(:|-)([0-9]+)[:-](.+)")

func (c Commander) RevParse() (string, error) {
	c.logDebug("Parsing latest commit", nil)
	cmd := exec.Command("git", "rev-parse", c.Head)
	cmd.Dir = c.Workspace
	out, err := cmd.Output()
	if err != nil {
		c.logError("Failed to parse latest commit", err, nil)
	}
	return strings.TrimSpace(string(out)), nil
}

func (c Commander) Clone(endpoint string) error {
	c.logDebug("Cloning repo", nil)
	cmd := exec.Command("git", "clone", "--depth", "1", "--single-branch", "--branch", c.Head, endpoint, c.Workspace)
	err := cmd.Run()
	if err != nil {
		c.logError("Failed to clone repo", err, nil)
	}
	return err
}

func (c Commander) Checkout() error {
	c.logDebug("Checking out to selected branch", nil)
	checkout := exec.Command("git", "checkout", c.Head)
	checkout.Dir = c.Workspace
	err := checkout.Run()
	if err != nil {
		c.logError("Failed to checkout selected branch", err, nil)
	}
	return err
}

func (c Commander) Grep(flags []string, ctxLines int) ([][]string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("ag --nogroup"))
	if ctxLines > 0 {
		sb.WriteString(fmt.Sprintf(" -C%d", ctxLines))
	}

	escapedFlags := []string{}
	for _, v := range flags {
		escapedFlags = append(escapedFlags, regexp.QuoteMeta(v))
	}
	sb.WriteString(" '" + strings.Join(escapedFlags, "|") + "' " + c.Workspace)

	cmd := sb.String()
	sh := exec.Command("sh", "-c", cmd)
	c.logDebug("Grepping for flag keys", map[string]interface{}{"numFlags": len(escapedFlags), "contextLines": ctxLines, "cmd": cmd})
	sh.Dir = c.Workspace
	out, err := sh.Output()
	if err != nil {
		if err.Error() == "exit status 1" {
			return [][]string{}, nil
		}
		c.logError("Error grepping for flag keys", err, map[string]interface{}{"numFlags": len(escapedFlags), "contextLines": ctxLines})
		return nil, err
	}
	grepRegexWithFilteredPath, err := regexp.Compile("(?:" + regexp.QuoteMeta(c.Workspace) + "/)" + grepRegex.String())
	if err != nil {
		return nil, err
	}
	ret := grepRegexWithFilteredPath.FindAllStringSubmatch(string(out), -1)
	return ret, err
}

func (c Commander) logDebug(msg string, fields map[string]interface{}) {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	fields["dir"] = c.Workspace
	fields["branch"] = c.Head
	log.Debug(msg, fields)
}

func (c Commander) logError(msg string, err error, fields map[string]interface{}) {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	fields["dir"] = c.Workspace
	fields["branch"] = c.Head
	log.Error(msg, err, fields)
}
