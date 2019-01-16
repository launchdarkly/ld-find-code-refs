package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/launchdarkly/git-flag-parser/internal/log"
	o "github.com/launchdarkly/git-flag-parser/internal/options"
	"github.com/launchdarkly/git-flag-parser/pkg/parse"
)

func main() {
	log.Info("Setting Bitbucket action env vars", nil)

	options := map[string]string{
		"repoType":         "bitbucket",
		"repoName":         os.Getenv("BITBUCKET_REPO_SLUG"),
		"repoHead":         os.Getenv("BITBUCKET_BRANCH"),
		"dir":              os.Getenv("BITBUCKET_CLONE_DIR"),
		"repoUrl":          os.Getenv("BITBUCKET_GIT_HTTP_ORIGIN"),
		"updateSequenceId": os.Getenv("BITBUCKET_BUILD_NUMBER"),
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
	log.Info(fmt.Sprintf("Starting repo parsing program with options:\n %+v\n", options), nil)
	parse.Parse()
}
