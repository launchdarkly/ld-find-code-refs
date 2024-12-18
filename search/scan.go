package search

import (
	"path/filepath"

	"github.com/launchdarkly/ld-find-code-refs/v2/flags"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/log"
	"github.com/launchdarkly/ld-find-code-refs/v2/options"
)

// Scan checks the configured directory for flags based on the options configured for Code References.
func Scan(opts options.Options, repoParams ld.RepoParams, dir string) (Matcher, []ld.ReferenceHunksRep) {
	flagKeys := flags.GetFlagKeys(opts, repoParams)
	matcher := NewMultiProjectMatcher(opts, dir, flagKeys)

	searchDir := dir
	if opts.Subdirectory != "" {
		searchDir = filepath.Join(dir, opts.Subdirectory)
	}

	refs, err := SearchForRefs(searchDir, matcher)
	if err != nil {
		log.Error.Fatalf("error searching for flag key references: %s", err)
	}

	return matcher, refs
}
