package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	o "github.com/launchdarkly/ld-find-code-refs/internal/options"
	"github.com/launchdarkly/ld-find-code-refs/pkg/coderefs"
)

func main() {
	log.Init(false)
	dir := os.Getenv("GITHUB_WORKSPACE")
	opts, err := o.GetWrapperOptions(dir, mergeGithubOptions)
	if err != nil {
		log.Error.Fatal(err)
	}
	log.Init(opts.Debug)
	coderefs.Scan(opts)
}

// mergeGithubOptions sets inferred options from the github actions environment, when available
func mergeGithubOptions(opts o.Options) (o.Options, error) {
	log.Info.Printf("Setting GitHub action env vars")
	ghRepo := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
	repoName := ""
	if len(ghRepo) > 1 {
		repoName = ghRepo[1]
	} else {
		log.Error.Printf("unable to validate GitHub repository name: %v", ghRepo)
	}

	ghBranch, err := parseBranch(os.Getenv("GITHUB_REF"))
	if err != nil {
		log.Error.Printf("error parsing GITHUB_REF: %v", err)
	}

	event, err := parseEvent(os.Getenv("GITHUB_EVENT_PATH"))
	if err != nil {
		log.Error.Printf("error parsing GitHub event payload at %q: %v", os.Getenv("GITHUB_EVENT_PATH"), err)
	}

	repoUrl := ""
	defaultBranch := ""
	updateSequenceId := -1
	if event != nil {
		repoUrl = event.Repo.Url
		defaultBranch = event.Repo.DefaultBranch
		updateSequenceId = int(event.Repo.PushedAt * 1000) // seconds to ms
	}

	opts.RepoType = "github"
	opts.RepoName = repoName
	opts.RepoUrl = repoUrl
	opts.DefaultBranch = defaultBranch
	opts.Branch = ghBranch
	opts.UpdateSequenceId = updateSequenceId

	return opts, opts.Validate()
}

type Event struct {
	Repo   `json:"repository"`
	Sender `json:"sender"`
}

type Repo struct {
	Url           string `json:"html_url"`
	DefaultBranch string `json:"default_branch"`
	PushedAt      int64  `json:"pushed_at"`
}

type Sender struct {
	Username string `json:"login"`
}

func parseEvent(path string) (*Event, error) {
	/* #nosec */
	eventJsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	eventJsonBytes, err := ioutil.ReadAll(eventJsonFile)
	if err != nil {
		return nil, err
	}
	var evt Event
	err = json.Unmarshal(eventJsonBytes, &evt)
	if err != nil {
		return nil, err
	}
	return &evt, err
}

func parseBranch(ref string) (string, error) {
	re := regexp.MustCompile(`^refs/heads/(.+)$`)
	results := re.FindStringSubmatch(ref)

	if results == nil {
		return "", fmt.Errorf("expected branch name starting with refs/heads/, got: %q", ref)
	}

	return results[1], nil
}
