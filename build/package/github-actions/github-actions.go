package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/launchdarkly/ld-find-code-refs/coderefs"
	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	o "github.com/launchdarkly/ld-find-code-refs/options"
)

func main() {
	log.Init(false)
	dir := os.Getenv("GITHUB_WORKSPACE")
	opts, err := o.GetWrapperOptions(dir, mergeGithubOptions)
	if err != nil {
		log.Error.Fatal(err)
	}
	log.Init(opts.Debug)
	coderefs.Run(opts)
}

// mergeGithubOptions sets inferred options from the github actions environment, when available
func mergeGithubOptions(opts o.Options) (o.Options, error) {
	log.Info.Printf("Setting GitHub action env vars")
	ghRepo := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
	repoName := ""

	if opts.RepoName != "" {
		repoName = opts.RepoName
	} else {
		if len(ghRepo) > 1 {
			repoName = ghRepo[1]
		} else {
			log.Error.Printf("unable to validate GitHub repository name: %v", ghRepo)
		}
	}
	event, err := parseEvent(os.Getenv("GITHUB_EVENT_PATH"))
	if err != nil {
		log.Error.Printf("error parsing GitHub event payload at %q: %v", os.Getenv("GITHUB_EVENT_PATH"), err)
	}
	ghBranch, err := parseBranch(os.Getenv("GITHUB_REF"), event)
	if err != nil {
		log.Error.Fatalf("error detecting git branch: %s", err)
	}

	repoUrl := ""
	defaultBranch := ""
	updateSequenceId := -1
	if event != nil {
		repoUrl = event.Repo.Url
		defaultBranch = event.Repo.DefaultBranch
		updateSequenceId = int(time.Now().Unix() * 1000) // seconds to ms
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
	*Pull  `json:"pull_request,omitempty"`
	Sender `json:"sender"`
}

type Repo struct {
	Url           string `json:"html_url"`
	DefaultBranch string `json:"default_branch"`
}

type Pull struct {
	Head `json:"head"`
}

type Head struct {
	Ref string `json:"ref"`
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

func parseBranch(ref string, event *Event) (string, error) {
	re := regexp.MustCompile(`^refs/heads/(.+)$`)
	results := re.FindStringSubmatch(ref)

	if results == nil {
		// The GITHUB_REF wasn't valid, so check if it's a pull request and use the pull request ref instead
		if event != nil && event.Pull != nil {
			return event.Pull.Head.Ref, nil
		} else {
			return "", fmt.Errorf("expected branch name starting with refs/heads/, got: %s", ref)
		}
	}
	return results[1], nil
}
