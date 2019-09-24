package command

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	o "github.com/launchdarkly/ld-find-code-refs/internal/options"
)

type GitClient struct {
	workspace string
	GitBranch string
	GitSha    string
}

func NewGitClient(path string) (GitClient, error) {
	if !filepath.IsAbs(path) {
		log.Fatal.Fatalf("expected an absolute path but received a relative path: %s", path)
	}

	client := GitClient{workspace: path}

	_, err := exec.LookPath("git")
	if err != nil {
		return client, errors.New("git is a required dependency, but was not found in the system PATH")
	}

	currBranch, err := client.branchName()
	if err != nil {
		return client, fmt.Errorf("error parsing git branch name: %s", err)
	} else if currBranch == "" {
		return client, fmt.Errorf("error parsing git branch name: git repo at %s must be checked out to a valid branch or --branch option must be set", client.workspace)
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

func (c GitClient) branchName() (string, error) {
	// Some CI systems leave the repository in a detached HEAD state. To support those, this logic allows
	// users to pass the branch name in by hand as an option.
	if o.Branch.Value() != "" {
		return o.Branch.Value(), nil
	}

	/* #nosec */
	cmd := exec.Command("git", "-C", c.workspace, "rev-parse", "--abbrev-ref", "HEAD")
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

func (c GitClient) headSha() (string, error) {
	/* #nosec */
	cmd := exec.Command("git", "-C", c.workspace, "rev-parse", "HEAD")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New(string(out))
	}
	ret := strings.TrimSpace(string(out))
	log.Debug.Printf("identified head sha: %s", ret)
	return ret, nil
}

func (c GitClient) RemoteBranches() (map[string]bool, error) {
	/* #nosec */
	cmd := exec.Command("git", "-C", c.workspace, "ls-remote", "--quiet", "--heads")
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
