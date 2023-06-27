package git

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	object "github.com/go-git/go-git/v5/plumbing/object"

	"github.com/launchdarkly/ld-find-code-refs/v2/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/v2/options"
	"github.com/launchdarkly/ld-find-code-refs/v2/search"

	"github.com/launchdarkly/ld-find-code-refs/v2/internal/log"
)

type Client struct {
	workspace    string
	GitBranch    string
	GitSha       string
	GitTimestamp int64
}

func NewClient(path string, branch string, allowTags bool) (*Client, error) {
	if !filepath.IsAbs(path) {
		log.Error.Fatalf("expected an absolute path but received a relative path: %s", path)
	}

	client := Client{workspace: path}

	_, err := exec.LookPath("git")
	if err != nil {
		return &client, errors.New("git is a required dependency, but was not found in the system PATH")
	}

	currBranch, refType, err := client.getRef(branch, allowTags)
	if err != nil {
		return &client, err
	}

	log.Info.Printf("git %s: %s", refType, currBranch)
	client.GitBranch = currBranch

	head, err := client.headSha()
	if err != nil {
		return &client, fmt.Errorf("error parsing current commit sha: %s", err)
	}
	client.GitSha = head

	timeStamp, err := client.commitTime()
	if err != nil {
		return &client, fmt.Errorf("error parsing current commit timestamp: %s", err)
	}
	client.GitTimestamp = timeStamp
	return &client, nil
}

func (c *Client) getRef(branch string, allowTags bool) (name string, refType string, err error) {
	if branch != "" {
		return branch, "branch", nil
	}

	name, err = c.branchName()
	if err != nil {
		return "", "", fmt.Errorf("error parsing git branch name: %s", err)
	}

	if name != "" {
		return name, "branch", nil
	}

	if !allowTags {
		return "", "", fmt.Errorf("error parsing git branch name: git repo at %s must be checked out to a valid branch or --branch option must be set", c.workspace)
	}

	name, err = c.tagName()
	if err != nil {
		return "", "", fmt.Errorf("error parsing git tag name: %s", err)
	}

	if name != "" {
		return name, "tag", nil
	}

	return "", "", fmt.Errorf("error parsing git tag name: git repo at %s must be checked out to a valid branch or tag, or --branch option must be set", c.workspace)
}

func (c *Client) branchName() (name string, err error) {
	repo, err := git.PlainOpen(c.workspace)
	if err != nil {
		return name, err
	}
	ref, err := repo.Head()
	if err != nil {
		return name, err
	}
	name = ref.Name().Short()
	log.Debug.Printf("identified branch name: %s", name)
	if name == "HEAD" {
		return "", nil
	}
	return name, nil
}

func (c *Client) tagName() (name string, err error) {
	repo, err := git.PlainOpen(c.workspace)
	if err != nil {
		return name, err
	}
	head, err := repo.Head()
	if err != nil {
		return name, err
	}
	iter, err := repo.Tags()
	if err != nil {
		return name, err
	}

	if err := iter.ForEach(func(ref *plumbing.Reference) error {
		// Check if ref is an annotated tag
		obj, err := repo.TagObject(ref.Hash())
		if err != nil {
			if errors.Is(err, plumbing.ErrObjectNotFound) {
				// Ref is lightweight tag
				if head.Hash() == ref.Hash() {
					name = ref.Name().Short()
					iter.Close()
				}
				return nil
			}
			return err
		}
		// Annotated tag target should be commit
		if obj.TargetType != plumbing.CommitObject {
			return nil
		}
		if head.Hash() == obj.Target {
			name = obj.Name
			iter.Close()
		}
		return nil
	}); err != nil {
		return name, err
	}
	if name == "" {
		return name, nil
	}
	log.Debug.Printf("identified tag name: %s", name)
	return name, err
}

func (c *Client) headSha() (sha string, err error) {
	repo, err := git.PlainOpen(c.workspace)
	if err != nil {
		return sha, err
	}
	ref, err := repo.Head()
	if err != nil {
		return sha, err
	}
	sha = ref.Hash().String()
	log.Debug.Printf("identified head sha: %s", sha)
	return sha, nil
}

func (c *Client) commitTime() (commitTime int64, err error) {
	repo, err := git.PlainOpen(c.workspace)
	if err != nil {
		return commitTime, err
	}
	head, err := repo.Head()
	if err != nil {
		return commitTime, err
	}
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return commitTime, err
	}
	commitTime = commit.Author.When.UnixMilli()
	return commitTime, nil
}

func (c *Client) RemoteBranches() (branches map[string]bool, err error) {
	branches = map[string]bool{}
	repo, err := git.PlainOpen(c.workspace)
	if err != nil {
		return branches, err
	}

	remotes, err := repo.Remotes()
	if err != nil {
		return branches, err
	}

	for _, r := range remotes {
		refList, err := r.List(&git.ListOptions{})
		if err != nil {
			return branches, err
		}
		refPrefix := "refs/heads/"
		for _, ref := range refList {
			refName := ref.Name().String()
			if !strings.HasPrefix(refName, refPrefix) {
				continue
			}
			branchName := refName[len(refPrefix):]
			log.Debug.Printf("found remote branch: %s/%s", r.Config().Name, branchName)
			branches[branchName] = true
		}
	}

	// the current branch should be in the list of remote branches
	branches[c.GitBranch] = true
	return branches, nil
}

type CommitData struct {
	commit *object.Commit
	tree   *object.Tree
}

// FindExtinctions searches commit history for flags that had references removed recently
func (c Client) FindExtinctions(project options.Project, flags []string, matcher search.Matcher, lookback int) ([]ld.ExtinctionRep, error) {
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
		log.Debug.Printf("Examining commit: %s", c.commit.Hash)
		changes, err := commits[i+1].tree.Diff(c.tree)
		if err != nil {
			return nil, err
		}
		patch, err := changes.PatchContext(context.Background())
		if err != nil {
			return nil, err
		}

		// get matcher for project
		elementMatcher := matcher.GetProjectElementMatcher(project.Key)
		if elementMatcher == nil {
			// This is actually a huge issue if it happens
			panic(fmt.Sprintf("Matcher for project (%s) not found", project.Key))
		}

		nextFlags := make([]string, 0, len(flags))
		flagMap := make(map[string]int, len(flags))
		for _, flag := range flags {
			flagMap[flag] = 0
		}

		for _, filePatch := range patch.FilePatches() {
			if !shouldScanFilePatch(project.Dir, filePatch) {
				continue
			}

			for _, chunk := range filePatch.Chunks() {
				delta := 0
				switch chunk.Type() {
				case diff.Delete:
					delta = 1
				case diff.Add:
					delta = -1
				}
				if delta != 0 {
					for _, line := range strings.Split(chunk.Content(), "\n") {
						for _, el := range elementMatcher.FindMatches(line) {
							if _, ok := flagMap[el]; ok {
								flagMap[el] += delta
							}
						}
					}
				}
			}
		}

		for flag, removalCount := range flagMap {
			if removalCount > 0 {
				ret = append(ret, ld.ExtinctionRep{
					Revision: c.commit.Hash.String(),
					Message:  c.commit.Message,
					Time:     c.commit.Author.When.Unix() * 1000,
					ProjKey:  project.Key,
					FlagKey:  flag,
				})
				log.Debug.Printf("Found extinct flag: %s in project: %s", flag, project.Key)
			} else {
				// this flag was not removed in the current commit, so check for it again in the next commit
				nextFlags = append(nextFlags, flag)
			}
		}
		flags = nextFlags
	}

	return ret, err
}

// Determine if changed file should be scanned
func shouldScanFilePatch(projectDir string, filePatch diff.FilePatch) bool {
	fromFile, toFile := filePatch.Files()
	printDebugStatement(fromFile, toFile)

	if projectDir == "" {
		return true
	}

	// Ignore files outside of the project directory

	if toFile != nil && strings.HasPrefix(toFile.Path(), projectDir) {
		return true
	}

	if fromFile != nil && strings.HasPrefix(fromFile.Path(), projectDir) {
		return true
	}

	return false
}

func printDebugStatement(fromFile, toFile diff.File) {
	fromPath, toPath := "FROM_PATH", "TO_PATH"
	if fromFile != nil {
		fromPath = fromFile.Path()
	}
	if toFile != nil {
		toPath = toFile.Path()
	}
	log.Debug.Printf("Scanning from file: %s and to file: %s", fromPath, toPath)
}
