package git

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	git "github.com/go-git/go-git/v5"
	object "github.com/go-git/go-git/v5/plumbing/object"

	"github.com/launchdarkly/ld-find-code-refs/internal/ld"

	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	"github.com/launchdarkly/ld-find-code-refs/search"
)

type Client struct {
	workspace string
	GitBranch string
	GitSha    string
}

func NewClient(path string, branch string) (*Client, error) {
	if !filepath.IsAbs(path) {
		log.Error.Fatalf("expected an absolute path but received a relative path: %s", path)
	}

	client := Client{workspace: path}

	_, err := exec.LookPath("git")
	if err != nil {
		return &client, errors.New("git is a required dependency, but was not found in the system PATH")
	}

	var currBranch = branch
	if branch == "" {
		currBranch, err = client.branchName()
		if err != nil {
			return &client, fmt.Errorf("error parsing git branch name: %s", err)
		} else if currBranch == "" {
			return &client, fmt.Errorf("error parsing git branch name: git repo at %s must be checked out to a valid branch or --branch option must be set", client.workspace)
		}
	}
	log.Info.Printf("git branch: %s", currBranch)
	client.GitBranch = currBranch

	head, err := client.headSha()
	if err != nil {
		return &client, fmt.Errorf("error parsing current commit sha: %s", err)
	}
	client.GitSha = head

	return &client, nil
}

func (c *Client) branchName() (string, error) {
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

func (c *Client) headSha() (string, error) {
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

func (c *Client) RemoteBranches() (map[string]bool, error) {
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

type CommitData struct {
	commit *object.Commit
	tree   *object.Tree
}

// FindExtinctions searches commit history for flags that had references removed recently
func (c Client) FindExtinctions(projKey string, flags []string, delimiters string, lookback int) ([]ld.ExtinctionRep, error) {
	repo, err := git.PlainOpen(c.workspace)
	if err != nil {
		return nil, err
	}
	logResult, err := repo.Log(&git.LogOptions{})
	if err != nil {
		return nil, err
	}

	commits := []CommitData{}
	for i := 0; i < lookback; i++ {
		commit, err := logResult.Next()
		if err != nil {
			// reached end of commit tree
			if err == io.EOF {
				break
			}
			return nil, err
		}
		tree, err := commit.Tree()
		if err != nil {
			return nil, err
		}
		commits = append(commits, CommitData{commit, tree})
	}

	ret := []ld.ExtinctionRep{}
	for i, c := range commits[:len(commits)-1] {
		changes, err := commits[i+1].tree.Diff(c.tree)
		if err != nil {
			return nil, err
		}
		patch, err := changes.Patch()
		if err != nil {
			return nil, err
		}
		patchLines := strings.Split(patch.String(), "\n")
		nextFlags := make([]string, 0, len(flags))
		for _, flag := range flags {
			removalCount := 0
			for _, patchLine := range patchLines {
				delta := 0
				// Is a change line and not a metadata line
				if strings.HasPrefix(patchLine, "-") && !strings.HasPrefix(patchLine, "---") {
					delta = 1
				} else if strings.HasPrefix(patchLine, "+") && !strings.HasPrefix(patchLine, "+++") {
					delta = -1
				}

				if delta != 0 && search.MatchDelimiters(patchLine, flag, delimiters) {
					removalCount += delta
				}
			}
			if removalCount > 0 {
				ret = append(ret, ld.ExtinctionRep{
					Revision: c.commit.Hash.String(),
					Message:  c.commit.Message,
					Time:     c.commit.Author.When.Unix() * 1000,
					ProjKey:  projKey,
					FlagKey:  flag,
				})
			} else {
				// this flag was not removed in the current commit, so check for it again in the next commit
				nextFlags = append(nextFlags, flag)
			}
		}
		flags = nextFlags
	}

	return ret, err
}
