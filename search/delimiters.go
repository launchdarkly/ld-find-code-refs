package search

import (
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/helpers"
	"github.com/launchdarkly/ld-find-code-refs/v2/options"
)

// Get a list of delimiters to use for flag key matching
// If defaults are disabled, only additional configured delimiters will be used
func GetDelimiters(opts options.Options) []string {
	delims := []string{`"`, `'`, "`"}
	if opts.Delimiters.DisableDefaults {
		delims = []string{}
	}

	delims = append(delims, opts.Delimiters.Additional...)

	return helpers.Dedupe(delims)
}
