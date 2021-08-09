package matcher

import (
	"fmt"
	"strings"

	"github.com/launchdarkly/ld-find-code-refs/element"
	"github.com/launchdarkly/ld-find-code-refs/flags"
	"github.com/launchdarkly/ld-find-code-refs/internal/git"
	"github.com/launchdarkly/ld-find-code-refs/internal/helpers"
	"github.com/launchdarkly/ld-find-code-refs/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	"github.com/launchdarkly/ld-find-code-refs/internal/validation"
	"github.com/launchdarkly/ld-find-code-refs/internal/version"
	"github.com/launchdarkly/ld-find-code-refs/options"
	"github.com/launchdarkly/ld-find-code-refs/search"
)

func Run(opts options.Options) {
	dir := opts.Dir
	absPath, err := validation.NormalizeAndValidatePath(dir)
	if err != nil {
		log.Error.Fatalf("could not validate directory option: %s", err)
	}

	log.Info.Printf("absolute directory path: %s", absPath)
	ldApi := ld.InitApiClient(ld.ApiOptions{ApiKey: opts.AccessToken, BaseUri: opts.BaseUri, ProjKey: opts.ProjKey, UserAgent: "LDFindCodeRefs/" + version.Version})

	branchName := opts.Branch
	revision := opts.Revision
	var gitClient *git.Client
	if revision == "" {
		gitClient, err = git.NewClient(absPath, branchName)
		if err != nil {
			log.Error.Fatalf("%s", err)
		}
		branchName = gitClient.GitBranch
		revision = gitClient.GitSha
	}

	repoParams := ld.RepoParams{
		Type:              opts.RepoType,
		Name:              opts.RepoName,
		Url:               opts.RepoUrl,
		CommitUrlTemplate: opts.CommitUrlTemplate,
		HunkUrlTemplate:   opts.HunkUrlTemplate,
		DefaultBranch:     opts.DefaultBranch,
	}

	matcher, refs := Scan(opts, repoParams)

	var updateId *int
	if opts.UpdateSequenceId >= 0 {
		updateIdOption := opts.UpdateSequenceId
		updateId = &updateIdOption
	}

	branch := ld.BranchRep{
		Name:             strings.TrimPrefix(branchName, "refs/heads/"),
		Head:             revision,
		UpdateSequenceId: updateId,
		SyncTime:         helpers.MakeTimestamp(),
		References:       refs,
	}

	handleOutput(opts, matcher, branch, repoParams, ldApi)

	if gitClient != nil {
		runExtinctions(opts, matcher, branch, repoParams, gitClient, ldApi)
	}

}

// Scan checks the configured directory for flags base on the options configured for Code References.
func Scan(opts options.Options, repoParams ld.RepoParams) (element.Matcher, []ld.ReferenceHunksRep) {
	flagMatcher := flags.GenerateSearchElements(opts, repoParams)

	matcher := element.Matcher{
		Elements: []element.ElementMatcher{flagMatcher},
		CtxLines: opts.ContextLines,
	}

	// Configure delimiters
	delims := getDelimiters(opts)
	matcher.Delimiters = strings.Join(helpers.Dedupe(delims), "")

	// Begin search for elements.
	refs, err := search.SearchForRefs(matcher)
	if err != nil {
		log.Error.Fatalf("error searching for flag key references: %s", err)
	}

	return matcher, refs
}

func getDelimiters(opts options.Options) []string {
	delims := []string{`"`, `'`, "`"}
	if opts.Delimiters.DisableDefaults {
		delims = []string{}
	}

	delims = append(delims, opts.Delimiters.Additional...)

	return delims
}

func deleteStaleBranches(ldApi ld.ApiClient, repoName string, remoteBranches map[string]bool) error {
	branches, err := ldApi.GetCodeReferenceRepositoryBranches(repoName)
	if err != nil {
		return err
	}

	staleBranches := calculateStaleBranches(branches, remoteBranches)
	if len(staleBranches) > 0 {
		log.Debug.Printf("marking stale branches for code reference pruning: %v", staleBranches)
		err = ldApi.PostDeleteBranchesTask(repoName, staleBranches)
		if err != nil {
			return err
		}
	}

	return nil
}

func calculateStaleBranches(branches []ld.BranchRep, remoteBranches map[string]bool) []string {
	staleBranches := []string{}
	for _, branch := range branches {
		if !remoteBranches[branch.Name] {
			staleBranches = append(staleBranches, branch.Name)
		}
	}
	log.Info.Printf("found %d stale branches to be marked for code reference pruning", len(staleBranches))
	return staleBranches
}

func handleOutput(opts options.Options, matcher element.Matcher, branch ld.BranchRep, repoParams ld.RepoParams, ldApi ld.ApiClient) {
	outDir := opts.OutDir
	if outDir != "" {
		outPath, err := branch.WriteToCSV(outDir, opts.ProjKey, repoParams.Name, opts.Revision)
		if err != nil {
			log.Error.Fatalf("error writing code references to csv: %s", err)
		}
		log.Info.Printf("wrote code references to %s", outPath)
	}

	if opts.Debug {
		branch.PrintReferenceCountTable()
	}

	if opts.DryRun {
		log.Info.Printf(
			"dry run found %d code references across %d flags and %d files",
			branch.TotalHunkCount(),
			len(matcher.Elements[0].Elements),
			len(branch.References),
		)
		return
	}

	log.Info.Printf(
		"sending %d code references across %d flags and %d files to LaunchDarkly for project: %s",
		branch.TotalHunkCount(),
		len(matcher.Elements[0].Elements),
		len(branch.References),
		opts.ProjKey,
	)
	err := ldApi.PutCodeReferenceBranch(branch, repoParams.Name)
	switch {
	case err == ld.BranchUpdateSequenceIdConflictErr:
		if branch.UpdateSequenceId != nil {
			log.Warning.Printf("updateSequenceId (%d) must be greater than previously submitted updateSequenceId", *branch.UpdateSequenceId)
		}
	case err == ld.EntityTooLargeErr:
		log.Error.Fatalf("code reference payload too large for LaunchDarkly API - consider excluding more files with .ldignore")
	case err != nil:
		helpers.FatalServiceError(fmt.Errorf("error sending code references to LaunchDarkly: %w", err), opts.IgnoreServiceErrors)
	}
}

func runExtinctions(opts options.Options, matcher element.Matcher, branch ld.BranchRep, repoParams ld.RepoParams, gitClient *git.Client, ldApi ld.ApiClient) {
	lookback := opts.Lookback
	if lookback > 0 {
		missingFlags := []string{}
		for flag, count := range branch.CountByFlag(matcher.Elements[0].Elements) {
			if count == 0 {
				missingFlags = append(missingFlags, flag)
			}

		}
		log.Info.Printf("checking if %d flags without references were removed in the last %d commits", len(missingFlags), opts.Lookback)
		removedFlags, err := gitClient.FindExtinctions(opts.ProjKey, missingFlags, matcher.Delimiters, lookback+1)
		if err != nil {
			log.Warning.Printf("unable to generate flag extinctions: %s", err)
		} else {
			log.Info.Printf("found %d removed flags", len(removedFlags))
		}
		if len(removedFlags) > 0 {
			err = ldApi.PostExtinctionEvents(removedFlags, repoParams.Name, branch.Name)
			if err != nil {
				log.Error.Printf("error sending extinction events to LaunchDarkly: %s", err)
			}
		}
	}
	log.Info.Printf("attempting to prune old code reference data from LaunchDarkly")
	remoteBranches, err := gitClient.RemoteBranches()
	if err != nil {
		log.Warning.Printf("unable to retrieve branch list from remote, skipping code reference pruning: %s", err)
	} else {
		err = deleteStaleBranches(ldApi, repoParams.Name, remoteBranches)
		if err != nil {
			helpers.FatalServiceError(fmt.Errorf("failed to mark old branches for code reference pruning: %w", err), opts.IgnoreServiceErrors)
		}
	}
}
