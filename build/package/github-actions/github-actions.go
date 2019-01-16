package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/launchdarkly/git-flag-parser/internal/log"
	o "github.com/launchdarkly/git-flag-parser/internal/options"
	"github.com/launchdarkly/git-flag-parser/pkg/parse"
)

func main() {
	log.Info("Setting GitHub action env vars", nil)
	ghRepo := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
	if len(ghRepo) < 2 {
		log.Fatal("Invalid GitHub repository name", fmt.Errorf("unable to find GitHub repository name"))
	}
	event, err := parseEvent(os.Getenv("GITHUB_EVENT_PATH"))
	if err != nil {
		log.Fatal("Error parsing GitHub event payload", fmt.Errorf("unable to parse GitHub event payload: %+v at '%s'\n", err, os.Getenv("GITHUB_EVENT_PATH")))
	}

	options := map[string]string{
		"repoType":         "github",
		"repoName":         ghRepo[1],
		"repoHead":         os.Getenv("GITHUB_REF"),
		"dir":              os.Getenv("GITHUB_WORKSPACE"),
		"updateSequenceId": strconv.FormatInt(event.Repo.PushedAt*1000, 10), // seconds to milliseconds
		"defaultBranch":    event.Repo.DefaultBranch,
		"repoUrl":          event.Repo.Url,
	}
	ldOptions, err := o.GetLDOptionsFromEnv()
	if err != nil {
		log.Fatal("Error settings options", err)
	}
	for k, v := range ldOptions {
		options[k] = v
	}

	o.Populate()
	for k, v := range options {
		err := flag.Set(k, v)
		if err != nil {
			log.Fatal(fmt.Sprintf("Error setting option %s", k), err)
		}
	}
	// Don't log ld access token
	optionsForLog := options
	optionsForLog["accessToken"] = ""
	log.Info(fmt.Sprintf("Starting repo parsing program with options:\n %+v\n", options), nil)
	parse.Parse()
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
