package flags

import (
	"fmt"

	"github.com/launchdarkly/ld-find-code-refs/v2/internal/helpers"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/log"
	"github.com/launchdarkly/ld-find-code-refs/v2/options"
)

const (
	minFlagKeyLen = 3 // Minimum flag key length helps reduce the number of false positives
)

func GetFlagKeys(opts options.Options, repoParams ld.RepoParams) map[string][]string {
	isDryRun := opts.DryRun
	ldApi := ld.InitApiClient(ld.ApiOptions{ApiKey: opts.AccessToken, BaseUri: opts.BaseUri, UserAgent: helpers.GetUserAgent(opts.UserAgent)})
	ignoreServiceErrors := opts.IgnoreServiceErrors

	if !isDryRun {
		err := ldApi.MaybeUpsertCodeReferenceRepository(repoParams)
		if err != nil {
			helpers.FatalServiceError(err, ignoreServiceErrors)
		}
	}

	flagKeys := make(map[string][]string)
	for _, proj := range opts.Projects {
		flags, err := getFlags(ldApi, proj.Key, opts.SkipArchivedFlags)
		if err != nil {
			helpers.FatalServiceError(fmt.Errorf("could not retrieve flag keys from LaunchDarkly for project `%s`: %w", proj.Key, err), ignoreServiceErrors)
		}
		addFlagKeys(flagKeys, flags, proj.Key)
	}
	return flagKeys
}

// Very short flag keys lead to many false positives when searching in code,
// so we filter them out.
func filterShortFlagKeys(flags []string) (filtered []string, omitted []string) {
	filteredFlags := []string{}
	omittedFlags := []string{}
	for _, flag := range flags {
		if len(flag) >= minFlagKeyLen {
			filteredFlags = append(filteredFlags, flag)
		} else {
			omittedFlags = append(omittedFlags, flag)
		}
	}
	return filteredFlags, omittedFlags
}

func addFlagKeys(flagKeys map[string][]string, flags []string, projKey string) {
	filteredFlags, omittedFlags := filterShortFlagKeys(flags)
	if len(filteredFlags) == 0 {
		log.Warning.Printf("no flag keys longer than the minimum flag key length (%v) were found for project: %s. Skipping project",
			minFlagKeyLen, projKey)
		return
	} else if len(omittedFlags) > 0 {
		log.Warning.Printf("omitting %d flags with keys less than minimum (%d) for project: %s", len(omittedFlags), minFlagKeyLen, projKey)
	}
	flagKeys[projKey] = filteredFlags
}

func getFlags(ldApi ld.ApiClient, projKey string, skipArchivedFlags bool) ([]string, error) {
	flags, err := ldApi.GetFlagKeyList(projKey, skipArchivedFlags)
	if err != nil {
		return nil, err
	}
	return flags, nil
}
