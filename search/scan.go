package search

import (
	"github.com/launchdarkly/ld-find-code-refs/v2/flags"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/log"
	"github.com/launchdarkly/ld-find-code-refs/v2/options"
)

// ScanForFlags checks the configured directory for flags based on the options configured for Code References.
// flagKeys is a map of flag keys per-project
func ScanForFlags(opts options.Options, flagKeys map[string][]string, dir string) (Matcher, []ld.ReferenceHunksRep) {
	matcher := NewMultiProjectMatcher(opts, flagKeys, dir)

	refs, err := SearchForRefs(dir, matcher)
	if err != nil {
		log.Error.Fatalf("error searching for flag key references: %s", err)
	}

	return matcher, refs
}

// Scan checks the configured directory for flags based on the options configured for Code References.
// @Deprecated: Use ScanForFlags instead
func Scan(opts options.Options, repoParams ld.RepoParams, dir string) (Matcher, []ld.ReferenceHunksRep) {
	flagKeys := flags.GetFlagKeys(opts, repoParams)
	return ScanForFlags(opts, flagKeys, dir)
}
