package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/launchdarkly/git-flag-parser/parse"
	o "github.com/launchdarkly/git-flag-parser/parse/internal/options"
)

func main() {
	fmt.Println("Setting GitHub action env vars")
	ghRepo := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
	if len(ghRepo) < 2 {
		fmt.Printf("Invalid GitHub repository name set: '%s'\n", os.Getenv("GITHUB_REPOSITORY"))
		os.Exit(1)
	}
	event, err := parseEvent(os.Getenv("GITHUB_EVENT_PATH"))
	if err != nil {
		fmt.Printf("Error parsing GitHub event payload: %+v at '%s'\n", err, os.Getenv("GITHUB_EVENT_PATH"))
		os.Exit(1)
	}

	options := map[string]string{
		"repoType":      "github",
		"repoOwner":     ghRepo[0],
		"repoName":      ghRepo[1],
		"repoHead":      os.Getenv("GITHUB_REF"),
		"dir":           os.Getenv("GITHUB_WORKSPACE"),
		"accessToken":   os.Getenv("LD_ACCESS_TOKEN"),
		"projKey":       os.Getenv("LD_PROJ_KEY"),
		"exclude":       os.Getenv("LD_EXCLUDE"),
		"contextLines":  os.Getenv("LD_CONTEXT_LINES"),
		"baseUri":       os.Getenv("LD_BASE_URI"),
		"pushTime":      strconv.FormatInt(event.Repo.PushedAt, 10),
		"defaultBranch": event.Repo.DefaultBranch,
	}
	o.Populate()
	for k, v := range options {
		err := flag.Set(k, v)
		if err != nil {
			fmt.Printf("Error setting option: %s", k)
			os.Exit(1)
		}
	}
	fmt.Printf("Starting repo parsing program with options:\n %+v\n", options)
	parse.Parse()
}

type Event struct {
	Repo   `json:"repository"`
	Sender `json:"sender"`
}

type Repo struct {
	Url           string `json:"url"`
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
