package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/launchdarkly/git-flag-parser/parse"
	"github.com/launchdarkly/git-flag-parser/parse/internal/log"
	o "github.com/launchdarkly/git-flag-parser/parse/internal/options"
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
		"repoType":      "github",
		"repoName":      ghRepo[1],
		"repoHead":      os.Getenv("GITHUB_REF"),
		"dir":           os.Getenv("GITHUB_WORKSPACE"),
		"accessToken":   os.Getenv("LD_ACCESS_TOKEN"),
		"projKey":       os.Getenv("LD_PROJ_KEY"),
		"exclude":       os.Getenv("LD_EXCLUDE"),
		"contextLines":  os.Getenv("LD_CONTEXT_LINES"),
		"baseUri":       os.Getenv("LD_BASE_URI"),
		"pushTime":      strconv.FormatInt(event.Repo.PushedAt*1000, 10), // seconds to milliseconds
		"defaultBranch": event.Repo.DefaultBranch,
	}
	ldOptions, err := getLDOptionsFromEnv()
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
	log.Info(fmt.Sprintf("Starting repo parsing program with options:\n %+v\n", options), nil)
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

func getLDOptionsFromEnv() (map[string]string, error) {
	ldOptions := map[string]string{
		"accessToken":  os.Getenv("LD_ACCESS_TOKEN"),
		"projKey":      os.Getenv("LD_PROJ_KEY"),
		"exclude":      os.Getenv("LD_EXCLUDE"),
		"contextLines": os.Getenv("LD_CONTEXT_LINES"),
		"baseUri":      os.Getenv("LD_BASE_URI"),
	}
	_, err := regexp.Compile(ldOptions["exclude"])
	if err != nil {
		return ldOptions, fmt.Errorf("couldn't parse LD_EXCLUDE as regex: %+v", err)
	}

	if ldOptions["contextLines"] == "" {
		ldOptions["contextLines"] = "-1"
	}
	_, err = strconv.ParseInt(ldOptions["contextLines"], 10, 32)
	if err != nil {
		return ldOptions, fmt.Errorf("coudln't parse LD_CONTEXT_LINES as an integer: %+v", err)
	}

	return ldOptions, nil
}
