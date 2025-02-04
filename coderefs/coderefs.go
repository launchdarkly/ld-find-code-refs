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
		generateOutput(opts, matcher, refs)
	}

	if !opts.DryRun {
		// First, list all existing code references for each feature flag
		existingRefs := make(map[string]bucketeer.CodeReference)

		// Get unique feature flags from the references
		flagCounts := aggregateFeatureFlags(refs)

		// List code references for each feature flag
		for flag := range flagCounts {
			codeRefs, _, _, err := bucketeerApi.ListCodeReferences(opts, flag, 1000)
			if err != nil {
				helpers.FatalServiceError(fmt.Errorf("error getting existing code references from Bucketeer for flag %s: %w", flag, err), opts.IgnoreServiceErrors)
			}

			for _, ref := range codeRefs {
				existingRefs[ref.ContentHash] = ref
			}
		}

		// Now process new references
		for _, ref := range refs {
			for _, hunk := range ref.Hunks {
				codeRef := bucketeer.CodeReference{
					FeatureID:        hunk.FlagKey,
					FilePath:         ref.Path,
					FileExtension:    hunk.FileExt,
					LineNumber:       hunk.StartingLineNumber,
					CodeSnippet:      hunk.Lines,
					ContentHash:      hunk.ContentHash,
					Aliases:          hunk.Aliases,
					RepositoryName:   opts.RepoName,
					RepositoryOwner:  opts.RepoOwner,
					RepositoryType:   repoType,
					RepositoryBranch: strings.TrimPrefix(branchName, "refs/heads/"),
					CommitHash:       revision,
					EnvironmentID:    opts.EnvironmentID,
				}

				if existing, exists := existingRefs[hunk.ContentHash]; exists {
					// Update the reference to ensure metadata is current
					log.Info.Printf("updating code reference in Bucketeer: id: %s, content hash: %s", existing.ID, codeRef.ContentHash)
					err := bucketeerApi.UpdateCodeReference(opts, existing.ID, codeRef)
					if err != nil {
						helpers.FatalServiceError(fmt.Errorf("error updating code reference in Bucketeer: %w", err), opts.IgnoreServiceErrors)
					}
					delete(existingRefs, hunk.ContentHash)
				} else {
					// Create new reference if content hash doesn't exist
					err := bucketeerApi.CreateCodeReference(opts, codeRef)
					if err != nil {
						helpers.FatalServiceError(fmt.Errorf("error sending code reference to Bucketeer: %w", err), opts.IgnoreServiceErrors)
					}
				}
			}
		}

		// Delete references that no longer exist in the codebase
		for _, ref := range existingRefs {
			// Only delete references from the same repository
			if ref.RepositoryOwner == opts.RepoOwner && ref.RepositoryName == opts.RepoName {
				err := bucketeerApi.DeleteCodeReference(opts, ref.ID)
				if err != nil {
					helpers.FatalServiceError(fmt.Errorf("error deleting code reference from Bucketeer: %w", err), opts.IgnoreServiceErrors)
				}
				log.Info.Printf("deleted code reference from Bucketeer: %+v", ref)
			}
		}
	}
}

func generateOutput(opts options.Options, matcher search.Matcher, refs []bucketeer.ReferenceHunksRep) {
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
	err = writer.Write([]string{
		"Flag Key",
		"File Path",
		"File Extension",
		"Line Number",
		"Code Snippet",
		"Content Hash",
		"Aliases",
	})
	if err != nil {
		return "", err
	}

	// Write data
	for _, ref := range refs {
		for _, hunk := range ref.Hunks {
			err := writer.Write([]string{
				hunk.FlagKey,
				ref.Path,
				hunk.FileExt,
				fmt.Sprintf("%d", hunk.StartingLineNumber),
				hunk.Lines,
				hunk.ContentHash,
				strings.Join(hunk.Aliases, "|"),
			})
			if err != nil {
				return "", err
			}
		}
	}

	return outPath, nil
}

// aggregateFeatureFlags returns a map of feature flags and their counts from the references
func aggregateFeatureFlags(refs []bucketeer.ReferenceHunksRep) map[string]int {
	flagCounts := make(map[string]int)
	for _, ref := range refs {
		for _, hunk := range ref.Hunks {
			flagCounts[hunk.FlagKey]++
		}
	}
	return flagCounts
}

func printReferenceCountTable(refs []bucketeer.ReferenceHunksRep) {
	flagCounts := aggregateFeatureFlags(refs)
	log.Info.Printf("Flag Reference Counts:")
	for flag, count := range flagCounts {
		log.Info.Printf("  %s: %d", flag, count)
	}
}
