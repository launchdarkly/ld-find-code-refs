package search

import (
	"path/filepath"

	"github.com/bucketeer-io/code-refs/flags"
	"github.com/bucketeer-io/code-refs/internal/bucketeer"
	"github.com/bucketeer-io/code-refs/internal/log"
	"github.com/bucketeer-io/code-refs/options"
)

// Scan checks the configured directory for flags based on the options configured for Code References.
func Scan(opts options.Options, dir string) (Matcher, []bucketeer.ReferenceHunksRep) {
	flagKeys := flags.GetFlagKeys(opts)
	matcher := NewEnvironmentMatcher(opts, dir, flagKeys)

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
