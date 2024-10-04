package search

import (
	"github.com/launchdarkly/ld-find-code-refs/v2/flags"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/log"
	"github.com/launchdarkly/ld-find-code-refs/v2/options"
)

// Scan checks the configured directory for flags based on the options configured for Code References.
func Scan(opts options.Options, repoParams ld.RepoParams, dir string) (Matcher, []ld.ReferenceHunksRep) {
	// declare a new variable called repo params and assign it the value of the repoParams parameter
	newRepoParams := repoParams
	newRepoParams.Url = "https://FirstAmCorp@dev.azure.com/FirstAmCorp/Eclipse/_git/AT_Eclipse"
	flagKeys := flags.GetFlagKeys(opts, newRepoParams)
	matcher := NewMultiProjectMatcher(opts, dir, flagKeys)

	refs, err := SearchForRefs(dir, matcher)
	if err != nil {
		log.Error.Fatalf("error searching for flag key references: %s", err)
	}

	return matcher, refs
}
