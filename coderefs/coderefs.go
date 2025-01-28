package coderefs

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bucketeer-io/code-refs/internal/bucketeer"
	"github.com/bucketeer-io/code-refs/internal/git"
	"github.com/bucketeer-io/code-refs/internal/helpers"
	"github.com/bucketeer-io/code-refs/internal/log"
	"github.com/bucketeer-io/code-refs/internal/validation"
	"github.com/bucketeer-io/code-refs/options"
	"github.com/bucketeer-io/code-refs/search"
)

func Run(opts options.Options, output bool) {
	absPath, err := validation.NormalizeAndValidatePath(opts.Dir)
	if err != nil {
		log.Error.Fatalf("could not validate directory option: %s", err)
	}

	log.Info.Printf("absolute directory path: %s", absPath)
	bucketeerApi := bucketeer.InitApiClient(bucketeer.ApiOptions{
		ApiKey:    opts.ApiKey,
		BaseUri:   opts.BaseUri,
		UserAgent: helpers.GetUserAgent(opts.UserAgent),
	})

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

	repoType := strings.ToUpper(opts.RepoType)
	if repoType != "GITHUB" && repoType != "GITLAB" && repoType != "BITBUCKET" {
		repoType = "CUSTOM"
	}

	matcher, refs := search.Scan(opts, absPath)

	if output {
		generateOutput(opts, matcher, refs, bucketeerApi)
	}

	if !opts.DryRun {
		for _, ref := range refs {
			for _, hunk := range ref.Hunks {
				codeRef := bucketeer.CodeReference{
					FeatureID:        hunk.FlagKey,
					FilePath:         ref.Path,
					LineNumber:       hunk.StartingLineNumber,
					CodeSnippet:      hunk.Lines,
					ContentHash:      hunk.ContentHash,
					Aliases:          hunk.Aliases,
					RepositoryName:   opts.RepoName,
					RepositoryType:   repoType,
					RepositoryBranch: strings.TrimPrefix(branchName, "refs/heads/"),
					CommitHash:       revision,
					EnvironmentID:    opts.EnvironmentID,
				}

				err := bucketeerApi.CreateCodeReference(codeRef)
				if err != nil {
					helpers.FatalServiceError(fmt.Errorf("error sending code reference to Bucketeer: %w", err), opts.IgnoreServiceErrors)
				}
			}
		}
	}
}

func generateOutput(opts options.Options, matcher search.Matcher, refs []bucketeer.ReferenceHunksRep, bucketeerApi bucketeer.ApiClient) {
	outDir := opts.OutDir
	if outDir != "" {
		outPath, err := writeToCSV(outDir, opts.EnvironmentID, opts.RepoName, opts.Revision, refs)
		if err != nil {
			log.Error.Fatalf("error writing code references to csv: %s", err)
		}
		log.Info.Printf("wrote code references to %s", outPath)
	}

	if opts.Debug {
		printReferenceCountTable(refs)
	}

	if opts.DryRun {
		totalHunks := 0
		for _, ref := range refs {
			totalHunks += len(ref.Hunks)
		}
		log.Info.Printf(
			"dry run found %d code references across %d flags and %d files",
			totalHunks,
			len(matcher.Element.Elements),
			len(refs),
		)
		return
	}

	log.Info.Printf(
		"sending %d code references across %d flags and %d files to Bucketeer for environment: %s",
		getTotalHunkCount(refs),
		len(matcher.Element.Elements),
		len(refs),
		opts.EnvironmentID,
	)
}

func getTotalHunkCount(refs []bucketeer.ReferenceHunksRep) int {
	total := 0
	for _, ref := range refs {
		total += len(ref.Hunks)
	}
	return total
}

func writeToCSV(outDir, environmentID, repoName, revision string, refs []bucketeer.ReferenceHunksRep) (string, error) {
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("code-references-%s-%s-%s-%d.csv", environmentID, repoName, revision, timestamp)
	outPath := filepath.Join(outDir, filename)

	file, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	err = writer.Write([]string{"Flag Key", "File Path", "Line Number", "Code Snippet"})
	if err != nil {
		return "", err
	}

	// Write data
	for _, ref := range refs {
		for _, hunk := range ref.Hunks {
			err := writer.Write([]string{
				hunk.FlagKey,
				ref.Path,
				fmt.Sprintf("%d", hunk.StartingLineNumber),
				hunk.Lines,
			})
			if err != nil {
				return "", err
			}
		}
	}

	return outPath, nil
}

func printReferenceCountTable(refs []bucketeer.ReferenceHunksRep) {
	flagCounts := make(map[string]int)
	for _, ref := range refs {
		for _, hunk := range ref.Hunks {
			flagCounts[hunk.FlagKey]++
		}
	}

	log.Info.Printf("Flag Reference Counts:")
	for flag, count := range flagCounts {
		log.Info.Printf("  %s: %d", flag, count)
	}
}
