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
}

var grepRegex, _ = regexp.Compile("(.+)(:|-)([0-9]+)[:-](.+)")

func (c Commander) RevParse() (string, error) {
	c.logDebug("Parsing latest commit", nil)
	cmd := exec.Command("git", "rev-parse", c.Head)
	cmd.Dir = c.Workspace
	out, err := cmd.Output()
	if err != nil {
		c.logError("Failed to parse latest commit", err, nil)
	}
	return strings.Replace(string(out), "\n", "", -1), err
}

func (c Commander) Clone(endpoint string) error {
	c.logDebug("Cloning repo", nil)
	cmd := exec.Command("git", "clone", "--single-branch", "--branch", c.Head, endpoint, c.Workspace)
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
	c.logDebug("Grepping for flag keys", &log.Fields{"numFlags": len(flags), "contestLines": ctxLines})
	sh := exec.Command("git")
	sh.Args = make([]string, 0, 2*len(flags)+4)
	sh.Args = append(sh.Args, "git", "grep", "-nF")

	if ctxLines > 0 {
		sh.Args = append(sh.Args, fmt.Sprintf("-C%d", ctxLines))
	}

	for _, f := range flags {
		sh.Args = append(sh.Args, "-e", f)
	}

	sh.Dir = c.Workspace
	out, err := sh.Output()
	if err != nil {
		c.logError("Error grepping for flag keys", err, &log.Fields{"numFlags": len(flags), "contextLines": ctxLines})
	}
	ret := grepRegex.FindAllStringSubmatch(string(out), -1)
	return ret, err
}

func (c Commander) logDebug(msg string, fields *log.Fields) {
	if fields == nil {
		fields = &log.Fields{}
	}
	(*fields)["workspace"] = c.Workspace
	(*fields)["branch"] = c.Head
	log.WithFields(log.Fields{"workspace": c.Workspace, "branch": c.Head}).Debug(msg)
}

func (c Commander) logError(msg string, err error, fields *log.Fields) {
	if fields == nil {
		fields = &log.Fields{}
	}
	(*fields)["workspace"] = c.Workspace
	(*fields)["branch"] = c.Head
	(*fields)["error"] = err.Error()
	log.WithFields(*fields).Error(msg)
}
