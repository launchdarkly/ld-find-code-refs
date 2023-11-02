package search

import (
	"strings"

	"github.com/launchdarkly/ld-find-code-refs/v2/aliases"
	"github.com/launchdarkly/ld-find-code-refs/v2/flags"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/helpers"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/log"
	"github.com/launchdarkly/ld-find-code-refs/v2/options"
)

// ScanForFlags checks the configured directory for flags based on the options configured for Code References.
// flagKeys is a map of flag keys per-project
func ScanForFlags(opts options.Options, flagKeys map[string][]string, dir string) (Matcher, []ld.ReferenceHunksRep) {
	elements := []ElementMatcher{}

	for _, project := range opts.Projects {
		projectFlags := flagKeys[project.Key]
		projectAliases := opts.Aliases
		projectAliases = append(projectAliases, project.Aliases...)
		aliasesByFlagKey, err := aliases.GenerateAliases(projectFlags, projectAliases, dir)
		if err != nil {
			log.Error.Fatalf("failed to generate aliases: %s for project: %s", err, project.Key)
		}

		delimiters := strings.Join(helpers.Dedupe(getDelimiters(opts)), "")
		elements = append(elements, NewElementMatcher(project.Key, project.Dir, delimiters, projectFlags, aliasesByFlagKey))
	}
	matcher := Matcher{
		ctxLines: opts.ContextLines,
		Elements: elements,
	}

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
