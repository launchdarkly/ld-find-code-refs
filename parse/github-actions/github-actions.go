package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/launchdarkly/git-flag-parser/parse"
	"github.com/launchdarkly/git-flag-parser/parse/github-actions/internal/gh"
	o "github.com/launchdarkly/git-flag-parser/parse/internal/options"
)

func main() {
	fmt.Println("Setting GitHub action env vars")
	ghRepo := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
	if len(ghRepo) < 2 {
		fmt.Println("No gh repo set")
		return
	}
	event := gh.ParseEvent(os.Getenv("GITHUB_EVENT_PATH"))
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
		"pushTime":      strconv.FormatFloat(event["repository_pushed_at"].(float64), 'f', 0, 64),
		"defaultBranch": event["repository_default_branch"].(string),
	}
	o.Populate()
	for k, v := range options {
		flag.Set(k, v)
	}
	fmt.Printf("Starting repo parsing program with options:\n %+v\n", options)
	parse.Parse()
}
