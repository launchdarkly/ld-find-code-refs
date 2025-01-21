package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/bucketeer-io/code-refs/coderefs"
	"github.com/bucketeer-io/code-refs/internal/log"
	o "github.com/bucketeer-io/code-refs/options"
)

const (
	millisecondsInSecond = 1000 // Descriptive constant for milliseconds conversion
)

func main() {
	log.Init(false)
	dir := os.Getenv("GITHUB_WORKSPACE")
	opts, err := o.GetWrapperOptions(dir, mergeGithubOptions)
	if err != nil {
		log.Error.Fatal(err)
	}
	log.Init(opts.Debug)
	coderefs.Run(opts, true)
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
	ghBranch, err := parseBranch(os.Getenv("GITHUB_REF"), event, opts.AllowTags)
	if err != nil {
		log.Error.Fatalf("error detecting git branch: %s", err)
	}

	repoUrl := ""
	defaultBranch := ""
	updateSequenceId := -1
	if event != nil {
		repoUrl = event.Repo.Url
		defaultBranch = event.Repo.DefaultBranch
		updateSequenceId = int(time.Now().Unix() * millisecondsInSecond) // seconds to ms
	}

	opts.RepoType = "github"
	opts.RepoName = repoName
	opts.RepoUrl = repoUrl
	opts.DefaultBranch = defaultBranch
	opts.Branch = ghBranch
	opts.UpdateSequenceId = updateSequenceId
	opts.UserAgent = "github-actions"

	return opts, opts.Validate()
}

type Event struct {
	Repo     `json:"repository"`
	*Pull    `json:"pull_request,omitempty"`
	*Release `json:"release"`
	Sender   `json:"sender"`
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

type Release struct {
	TagName string `json:"tag_name"`
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

	eventJsonBytes, err := io.ReadAll(eventJsonFile)
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

func parseBranch(ref string, event *Event, allowTags bool) (string, error) {
	name_index := 1
	re := regexp.MustCompile(`^refs/heads/(.+)$`)
	if allowTags {
		name_index = 2
		re = regexp.MustCompile(`^refs/(heads|tags)/(.+)$`)
	}
	results := re.FindStringSubmatch(ref)

	if results != nil {
		return results[name_index], nil
	}

	// The GITHUB_REF wasn't valid, so check if it's a pull request and use the pull request ref instead
	if event != nil && event.Pull != nil {
		return event.Pull.Head.Ref, nil
	}

	// If it's not a pull request, check if it's a Release
	if allowTags && event != nil && event.Release != nil {
		return event.Release.TagName, nil
	}

	addendum := ""

	if allowTags {
		addendum = " or refs/tags/"
	}

	return "", fmt.Errorf("expected ref name starting with refs/heads/%s, got: %s", addendum, ref)
}
