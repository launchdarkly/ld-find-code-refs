package coderefs

import (
	"fmt"
	"strings"

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
	ldApi := ld.InitApiClient(ld.ApiOptions{ApiKey: opts.AccessToken, BaseUri: opts.BaseUri, UserAgent: "LDFindCodeRefs/" + version.Version})

	branchName := opts.Branch
	revision := opts.Revision
	var gitClient *git.Client
	if revision == "" {
		gitClient, err = git.NewClient(absPath, branchName, opts.AllowTags)
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

	matcher, refs := search.Scan(opts, repoParams)

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

func Prune(opts options.Options, branches []string) {
	ldApi := ld.InitApiClient(ld.ApiOptions{ApiKey: opts.AccessToken, BaseUri: opts.BaseUri, ProjKey: opts.ProjKey, UserAgent: "LDFindCodeRefs/" + version.Version})
	err := ldApi.PostDeleteBranchesTask(opts.RepoName, branches)
	if err != nil {
		helpers.FatalServiceError(err, opts.IgnoreServiceErrors)
	}
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

func handleOutput(opts options.Options, matcher search.Matcher, branch ld.BranchRep, repoParams ld.RepoParams, ldApi ld.ApiClient) {
	outDir := opts.OutDir
	var projects []string
	if len(opts.Projects) > 0 {
		for _, proj := range opts.Projects {
			projects = append(projects, proj.ProjectKey)
		}
	} else {
		projects = append(projects, opts.AccessToken)
	}
	if outDir != "" {
		outPath, err := branch.WriteToCSV(outDir, repoParams.Name, opts.Revision, projects)
		if err != nil {
			log.Error.Fatalf("error writing code references to csv: %s", err)
		}
		log.Info.Printf("wrote code references to %s", outPath)
	}

	if opts.Debug {
		branch.PrintReferenceCountTable()
	}

	if opts.DryRun {
		totalFlags := 0
		for _, searchElems := range matcher.Elements {
			totalFlags += len(searchElems.Elements)
		}
		log.Info.Printf(
			"dry run found %d code references across %d flags and %d files",
			branch.TotalHunkCount(),
			totalFlags,
			len(branch.References),
		)
		return
	}

	log.Info.Printf(
		"sending %d code references across %d flags and %d files to LaunchDarkly for project(s): %s",
		branch.TotalHunkCount(),
		len(matcher.Elements[0].Elements),
		len(branch.References),
		projects,
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

func runExtinctions(opts options.Options, matcher search.Matcher, branch ld.BranchRep, repoParams ld.RepoParams, gitClient *git.Client, ldApi ld.ApiClient) {
	lookback := opts.Lookback
	dryRun := opts.DryRun
	if lookback > 0 {
		var removedFlags []ld.ExtinctionRep
		for i, project := range opts.Projects {
			missingFlags := []string{}
			for flag, count := range branch.CountByFlag(matcher.Elements[i].Elements, project.ProjectKey) {
				if count == 0 {
					missingFlags = append(missingFlags, flag)
				}

			}
			log.Info.Printf("checking if %d flags without references were removed in the last %d commits for project: %s", len(missingFlags), opts.Lookback, project.ProjectKey)
			removedFlagsByProject, err := gitClient.FindExtinctions(project, missingFlags, matcher, lookback+1)
			if err != nil {
				log.Warning.Printf("unable to generate flag extinctions: %s", err)
			} else {
				log.Info.Printf("found %d removed flags", len(removedFlagsByProject))
			}
			removedFlags = append(removedFlags, removedFlagsByProject...)
		}
		if len(removedFlags) > 0 && !dryRun {
			err := ldApi.PostExtinctionEvents(removedFlags, repoParams.Name, branch.Name)
			if err != nil {
				log.Error.Printf("error sending extinction events to LaunchDarkly: %s", err)
			}
		}

	}
	if !dryRun {
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
}
