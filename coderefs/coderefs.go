package coderefs

import (
	"github.com/launchdarkly/ld-find-code-refs/internal/helpers"
	"github.com/launchdarkly/ld-find-code-refs/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/internal/version"
	"github.com/launchdarkly/ld-find-code-refs/options"
)

func Prune(opts options.Options, branches []string) {
	ldApi := ld.InitApiClient(ld.ApiOptions{ApiKey: opts.AccessToken, BaseUri: opts.BaseUri, ProjKey: opts.ProjKey, UserAgent: "LDFindCodeRefs/" + version.Version})
	err := ldApi.PostDeleteBranchesTask(opts.RepoName, branches)
	if err != nil {
		helpers.FatalServiceError(err, opts.IgnoreServiceErrors)
	}
}
