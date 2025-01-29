package flags

import (
	"fmt"

	"github.com/bucketeer-io/code-refs/internal/bucketeer"
	"github.com/bucketeer-io/code-refs/internal/helpers"
	"github.com/bucketeer-io/code-refs/internal/log"
	"github.com/bucketeer-io/code-refs/options"
)

const (
	minFlagKeyLen = 3 // Minimum flag key length helps reduce the number of false positives
)

func GetFlagKeys(opts options.Options) []string {
	bucketeerApi := bucketeer.InitApiClient(bucketeer.ApiOptions{
		ApiKey:    opts.ApiKey,
		BaseUri:   opts.BaseUri,
		UserAgent: helpers.GetUserAgent(opts.UserAgent),
	})
	ignoreServiceErrors := opts.IgnoreServiceErrors

	flags, err := getFlags(bucketeerApi, opts)
	if err != nil {
		helpers.FatalServiceError(fmt.Errorf("could not retrieve flag keys from Bucketeer: %w", err), ignoreServiceErrors)
	}

	filteredFlags, omittedFlags := filterShortFlagKeys(flags)
	if len(filteredFlags) == 0 {
		log.Warning.Printf("no flag keys longer than the minimum flag key length (%v) were found. Skipping",
			minFlagKeyLen)
		return nil
	} else if len(omittedFlags) > 0 {
		log.Warning.Printf("omitting %d flags with keys less than minimum (%d)", len(omittedFlags), minFlagKeyLen)
	}
	if opts.Debug {
		log.Debug.Printf("filtered flags: %v", filteredFlags)
	}
	return filteredFlags
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

func getFlags(bucketeerApi bucketeer.ApiClient, opts options.Options) ([]string, error) {
	flags, err := bucketeerApi.GetFlagKeyList(opts)
	if err != nil {
		return nil, err
	}
	return flags, nil
}
