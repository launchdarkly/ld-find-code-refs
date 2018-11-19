package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
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
	return strings.TrimSpace(string(out)), err
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

func (c Commander) Grep(flags []string, ctxLines int, exclude string) ([][]string, error) {
	var sb strings.Builder
	// not using git grep until we figure out why it takes so long when running on github actions containers
	// sb.WriteString(fmt.Sprintf("cd %s && git grep -nF", c.Workspace))
	// if ctxLines > 0 {
	// 	sb.WriteString(fmt.Sprintf(" -C%d", ctxLines))
	// }
	// for _, f := range flags {
	// 	sb.WriteString(fmt.Sprintf(" -e %s", f))
	// }

	sb.WriteString(fmt.Sprintf("ag --nogroup"))
	if ctxLines > 0 {
		sb.WriteString(fmt.Sprintf(" -C%d", ctxLines))
	}
	if exclude != "" {
		sb.WriteString(fmt.Sprintf(" --ignore %s", exclude))
	}
	escapedFlags := []string{}
	for _, v := range flags {
		escapedFlags = append(escapedFlags, regexp.QuoteMeta(v))
	}
	sb.WriteString(" '" + strings.Join(escapedFlags, "|") + "' " + c.Workspace)

	cmd := sb.String()
	sh := exec.Command("sh", "-c", cmd)
	log.WithFields(log.Fields{"numFlags": len(escapedFlags), "contextLines": ctxLines, "cmd": cmd}).Debug("Grepping for flag keys")
	sh.Dir = c.Workspace
	out, err := sh.Output()
	if err != nil {
		c.logError("Error grepping for flag keys", err, &log.Fields{"numFlags": len(escapedFlags), "contextLines": ctxLines})
	}
	grepRegexWithFilteredPath, err := regexp.Compile("(?:" + regexp.QuoteMeta(c.Workspace) + "/)" + grepRegex.String())
	if err != nil {
		// TODO: handle this error
		return nil, err
	}
	ret := grepRegexWithFilteredPath.FindAllStringSubmatch(string(out), -1)
	return ret, err
}

func (c Commander) logDebug(msg string, fields *log.Fields) {
	if fields == nil {
		fields = &log.Fields{}
	}
	(*fields)["dir"] = c.Workspace
	(*fields)["branch"] = c.Head
	log.WithFields(log.Fields{"workspace": c.Workspace, "branch": c.Head}).Debug(msg)
}

func (c Commander) logError(msg string, err error, fields *log.Fields) {
	if fields == nil {
		fields = &log.Fields{}
	}
	(*fields)["dir"] = c.Workspace
	(*fields)["branch"] = c.Head
	(*fields)["error"] = err.Error()
	log.WithFields(*fields).Error(msg)
}
