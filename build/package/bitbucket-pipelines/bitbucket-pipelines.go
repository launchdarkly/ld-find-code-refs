package main

import (
	"os"
	"strconv"

	"github.com/launchdarkly/ld-find-code-refs/v2/coderefs"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/log"
	o "github.com/launchdarkly/ld-find-code-refs/v2/options"
)

func main() {
	log.Init(false)
	dir := os.Getenv("BITBUCKET_CLONE_DIR")
	opts, err := o.GetWrapperOptions(dir, mergeBitbucketOptions)
	if err != nil {
		log.Error.Fatal(err)
	}
	log.Init(opts.Debug)
	coderefs.Run(opts, true)
}

func mergeBitbucketOptions(opts o.Options) (o.Options, error) {
	log.Info.Printf("Setting Bitbucket Pipelines env vars")
	if opts.RepoName == "" {
		opts.RepoName = os.Getenv("BITBUCKET_REPO_SLUG")
	}
	opts.RepoType = "bitbucket"
	opts.RepoUrl = os.Getenv("BITBUCKET_GIT_HTTP_ORIGIN")
	updateSequenceId, err := strconv.Atoi(os.Getenv("BITBUCKET_BUILD_NUMBER"))
	if err != nil {
		updateSequenceId = -1
	}
	opts.UpdateSequenceId = updateSequenceId
	opts.UserAgent = "bitbucket-pipeline"
	return opts, opts.Validate()
}
