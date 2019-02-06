package coderefs

import (
	"container/list"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/launchdarkly/ld-find-code-refs/internal/command"
	"github.com/launchdarkly/ld-find-code-refs/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	o "github.com/launchdarkly/ld-find-code-refs/internal/options"
)

// These are defensive limits intended to prevent corner cases stemming from
// large repos, false positives, etc. The goal is a) to prevent the program
// from taking a very long time to run and b) to prevent the program from
// PUTing a massive json payload. These limits will likely be tweaked over
// time. The LaunchDarkly backend will also apply limits.
const (
	minFlagKeyLen                     = 3
	maxFileCount                      = 5000
	maxLineCharCount                  = 500
	maxHunkCount                      = 5000
	maxHunksPerFileCount              = 1000
	maxHunkedLinesPerFileAndFlagCount = 500
	maxProjKeyLength                  = 20
)

type grepResultLine struct {
	Path     string
	LineNum  int
	LineText string
	FlagKeys []string
}

type grepResultLines []grepResultLine

// map of flag keys to slices of lines those flags occur on
type flagReferenceMap map[string][]*list.Element

// this struct contains a linked list of all the grep result lines
// for a single file, and a map of flag keys to slices of lines where
// those flags occur.
type fileGrepResults struct {
	path                string
	fileGrepResultLines *list.List
	flagReferenceMap
}

type branch struct {
	Name             string
	Head             string
	UpdateSequenceId *int64
	SyncTime         int64
	GrepResults      grepResultLines
}

func Scan() {
	cmd, err := command.NewClient(o.Dir.Value())
	if err != nil {
		log.Error.Fatalf("%s", err)
	}

	projKey := o.ProjKey.Value()

	// Check for potential sdk keys or access tokens provided as the project key
	if len(projKey) > maxProjKeyLength {
		if strings.HasPrefix(projKey, "sdk-") {
			log.Warning.Printf("provided projKey (%s) appears to be a LaunchDarkly SDK key", "sdk-xxxx")
		} else if strings.HasPrefix(projKey, "api-") {
			log.Warning.Printf("provided projKey (%s) appears to be a LaunchDarkly API access token", "api-xxxx")
		}
	}

	ldApi := ld.InitApiClient(ld.ApiOptions{ApiKey: o.AccessToken.Value(), BaseUri: o.BaseUri.Value(), ProjKey: projKey})
	repoParams := ld.RepoParams{
		Type:              o.RepoType.Value(),
		Name:              o.RepoName.Value(),
		Url:               o.RepoUrl.Value(),
		CommitUrlTemplate: o.CommitUrlTemplate.Value(),
		HunkUrlTemplate:   o.HunkUrlTemplate.Value(),
		DefaultBranch:     o.DefaultBranch.Value(),
	}

	err = ldApi.MaybeUpsertCodeReferenceRepository(repoParams)
	if err != nil {
		log.Error.Fatalf(err.Error())
	}

	flags, err := getFlags(ldApi)
	if err != nil {
		log.Error.Fatalf("could not retrieve flag keys from LaunchDarkly: %s", err)
	}
	if len(flags) == 0 {
		log.Info.Printf("no flag keys found for project: %s, exiting early", projKey)
		os.Exit(0)
	}

	filteredFlags, omittedFlags := filterShortFlagKeys(flags)
	if len(filteredFlags) == 0 {
		log.Info.Printf("no flag keys longer than the minimum flag key length (%v) were found for project: %s, exiting early",
			minFlagKeyLen, projKey)
		os.Exit(0)
	} else if len(omittedFlags) > 0 {
		log.Warning.Printf("omitting %d flags with keys less than minimum (%d)", len(omittedFlags), minFlagKeyLen)
	}

	ctxLines := o.ContextLines.Value()
	var updateId *int64
	if o.UpdateSequenceId.Value() >= 0 {
		updateIdOption := o.UpdateSequenceId.Value()
		updateId = &updateIdOption
	}
	b := &branch{
		Name:             cmd.GitBranch,
		UpdateSequenceId: updateId,
		SyncTime:         makeTimestamp(),
		Head:             cmd.GitSha,
	}

	// exclude option has already been validated as regex
	exclude, _ := regexp.Compile(o.Exclude.Value())
	refs, err := b.findReferences(cmd, filteredFlags, ctxLines, exclude)
	if err != nil {
		log.Error.Fatalf("error searching for flag key references: %s", err)
	}
	b.GrepResults = refs

	branchRep := b.makeBranchRep(projKey, ctxLines)
	log.Info.Printf("sending %d code references across %d flags and %d files to LaunchDarkly for project: %s", branchRep.TotalHunkCount(), len(filteredFlags), len(branchRep.References), projKey)

	if o.Debug.Value() {
		branchRep.PrintReferenceCountTable()
	}

	err = ldApi.PutCodeReferenceBranch(branchRep, repoParams.Name)
	if err != nil {
		if err == ld.BranchUpdateSequenceIdConflictErr && b.UpdateSequenceId != nil {
			log.Warning.Printf("updateSequenceId (%d) must be greater than previously submitted updateSequenceId", *b.UpdateSequenceId)
		} else {
			log.Error.Fatalf("error sending code references to LaunchDarkly: %s", err)
		}
	}
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

func getFlags(ldApi ld.ApiClient) ([]string, error) {
	flags, err := ldApi.GetFlagKeyList()
	if err != nil {
		return nil, err
	}
	return flags, nil
}

func (b *branch) findReferences(cmd command.Client, flags []string, ctxLines int, exclude *regexp.Regexp) (grepResultLines, error) {
	grepResult, err := cmd.SearchForFlags(flags, ctxLines)
	if err != nil {
		return grepResultLines{}, err
	}

	return generateReferencesFromGrep(flags, grepResult, ctxLines, exclude), nil
}

func generateReferencesFromGrep(flags []string, grepResult [][]string, ctxLines int, exclude *regexp.Regexp) []grepResultLine {
	references := []grepResultLine{}

	for _, r := range grepResult {
		path := r[1]
		if exclude != nil && exclude.String() != "" && exclude.MatchString(path) {
			continue
		}
		contextContainsFlagKey := r[2] == ":"
		lineNumber := r[3]
		lineText := r[4]
		lineNum, err := strconv.Atoi(lineNumber)
		if err != nil {
			log.Error.Fatalf("encountered an unexpected error generating flag references: %s", err)
		}
		ref := grepResultLine{Path: path, LineNum: lineNum}
		if contextContainsFlagKey {
			ref.FlagKeys = findReferencedFlags(lineText, flags)
		}
		if ctxLines >= 0 {
			ref.LineText = lineText
		}
		references = append(references, ref)
	}

	return references
}

func findReferencedFlags(ref string, flags []string) []string {
	ret := []string{}
	for _, flag := range flags {
		if strings.Contains(ref, flag) {
			ret = append(ret, flag)
		}
	}
	return ret
}

func (b *branch) makeBranchRep(projKey string, ctxLines int) ld.BranchRep {
	return ld.BranchRep{
		Name:             strings.TrimPrefix(b.Name, "refs/heads/"),
		Head:             b.Head,
		UpdateSequenceId: b.UpdateSequenceId,
		SyncTime:         b.SyncTime,
		References:       b.GrepResults.makeReferenceHunksReps(projKey, ctxLines),
	}
}

func (g grepResultLines) makeReferenceHunksReps(projKey string, ctxLines int) []ld.ReferenceHunksRep {
	reps := []ld.ReferenceHunksRep{}

	aggregatedGrepResults := g.aggregateByPath()

	if len(aggregatedGrepResults) > maxFileCount {
		log.Warning.Printf("found %d files with code references, which exceeded the limit of %d", len(aggregatedGrepResults), maxFileCount)
		aggregatedGrepResults = aggregatedGrepResults[0:maxFileCount]
	}

	numHunks := 0

	for _, fileGrepResults := range aggregatedGrepResults {
		if numHunks > maxHunkCount {
			log.Warning.Printf("found %d code references across all files, which exceeeded the limit of %d. halting code reference search", numHunks, maxHunkCount)
			break
		}

		hunks := fileGrepResults.makeHunkReps(projKey, ctxLines)

		if len(hunks) > maxHunksPerFileCount {
			log.Warning.Printf("found %d code references in %s, which exceeded the limit of %d, truncating file hunks", len(hunks), fileGrepResults.path, maxHunksPerFileCount)
			hunks = hunks[0:maxHunksPerFileCount]
		}

		numHunks += len(hunks)

		reps = append(reps, ld.ReferenceHunksRep{Path: fileGrepResults.path, Hunks: hunks})
	}
	return reps
}

// Assumes invariant: grepResultLines will already be sorted by path.
func (g grepResultLines) aggregateByPath() []fileGrepResults {
	allFileResults := []fileGrepResults{}

	if len(g) == 0 {
		return allFileResults
	}

	// initialize first file
	currentFileResults := fileGrepResults{
		path:                g[0].Path,
		flagReferenceMap:    flagReferenceMap{},
		fileGrepResultLines: list.New(),
	}

	for _, grepResult := range g {
		// If we reach a grep result with a new path, append the old one to our list and start a new one
		if grepResult.Path != currentFileResults.path {
			allFileResults = append(allFileResults, currentFileResults)

			currentFileResults = fileGrepResults{
				path:                grepResult.Path,
				flagReferenceMap:    flagReferenceMap{},
				fileGrepResultLines: list.New(),
			}
		}

		elem := currentFileResults.addGrepResult(grepResult)

		if len(grepResult.FlagKeys) > 0 {
			for _, flagKey := range grepResult.FlagKeys {
				currentFileResults.addFlagReference(flagKey, elem)
			}
		}
	}

	// append last file
	allFileResults = append(allFileResults, currentFileResults)

	return allFileResults
}

func (fgr *fileGrepResults) addGrepResult(grepResult grepResultLine) *list.Element {
	prev := fgr.fileGrepResultLines.Back()
	if prev != nil && prev.Value.(grepResultLine).LineNum > grepResult.LineNum {
		// This should never happen, as `ag` (and any other grep program we might use
		// should always return search results sorted by line number. We sanity check
		// that lines are sorted _just in case_ since the downstream hunking algorithm
		// only works on sorted lines.
		log.Error.Fatalf("grep results returned out of order")
	}

	return fgr.fileGrepResultLines.PushBack(grepResult)
}

func (fgr *fileGrepResults) addFlagReference(key string, ref *list.Element) {
	_, ok := fgr.flagReferenceMap[key]

	if ok {
		fgr.flagReferenceMap[key] = append(fgr.flagReferenceMap[key], ref)
	} else {
		fgr.flagReferenceMap[key] = []*list.Element{ref}
	}
}

func (fgr fileGrepResults) makeHunkReps(projKey string, ctxLines int) []ld.HunkRep {
	hunks := []ld.HunkRep{}

	for flagKey, flagReferences := range fgr.flagReferenceMap {
		flagHunks := buildHunksForFlag(projKey, flagKey, fgr.path, flagReferences, fgr.fileGrepResultLines, ctxLines)
		hunks = append(hunks, flagHunks...)
	}

	return hunks
}

func buildHunksForFlag(projKey, flag, path string, flagReferences []*list.Element, fileLines *list.List, ctxLines int) []ld.HunkRep {
	hunks := []ld.HunkRep{}

	var previousHunk *ld.HunkRep
	var currentHunk ld.HunkRep

	lastSeenLineNum := -1

	var hunkStringBuilder strings.Builder

	appendToPreviousHunk := false

	numHunkedLines := 0

	for _, ref := range flagReferences {
		// Each ref is either the start of a new hunk or a continuation of the previous hunk.
		// NOTE: its possible that this flag reference is totally contained in the previous hunk
		ptr := ref

		numCtxLinesBeforeFlagRef := 0

		// Attempt to seek to the start of the new hunk.
		for i := 0; i < ctxLines; i++ {
			// If we seek to a nil pointer, we're at the start of the file and can go no further.
			if ptr.Prev() != nil {
				ptr = ptr.Prev()
				numCtxLinesBeforeFlagRef++
			}

			// If we seek earlier than the end of the last hunk, this reference overlaps at least
			// partially with the last hunk and we should (possibly) expand the previous hunk rather than
			// starting a new hunk.
			if ptr.Value.(grepResultLine).LineNum <= lastSeenLineNum {
				appendToPreviousHunk = true
			}
		}

		// If we are starting a new hunk, initialize it
		if !appendToPreviousHunk {
			currentHunk = initHunk(projKey, flag)
			currentHunk.StartingLineNumber = ptr.Value.(grepResultLine).LineNum
			hunkStringBuilder.Reset()
		}

		// From the current position (at the theoretical start of the hunk) seek forward line by line X times,
		// where X = (numCtxLinesBeforeFlagRef + 1 + ctxLines). Note that if the flag reference occurs close to the
		// start of the file, numCtxLines may be smaller than ctxLines.
		//   For each line, check if we have seeked past the end of the last hunk
		//     If so: write that line to the hunkStringBuilder
		//     Record that line as the last seen line.
		for i := 0; i < numCtxLinesBeforeFlagRef+1+ctxLines; i++ {
			ptrLineNum := ptr.Value.(grepResultLine).LineNum
			if ptrLineNum > lastSeenLineNum {
				lineText := truncateLine(ptr.Value.(grepResultLine).LineText)
				hunkStringBuilder.WriteString(lineText + "\n")
				lastSeenLineNum = ptrLineNum
				numHunkedLines += 1
			}

			if ptr.Next() != nil {
				ptr = ptr.Next()
			}
		}

		if appendToPreviousHunk {
			previousHunk.Lines = hunkStringBuilder.String()
			appendToPreviousHunk = false
		} else {
			currentHunk.Lines = hunkStringBuilder.String()
			hunks = append(hunks, currentHunk)
			previousHunk = &hunks[len(hunks)-1]
		}

		// If we have written more than the max. allowed number of lines for this file and flag, finish this hunk and exit early.
		// This guards against a situation where the user has very long files with many false positive matches.
		if numHunkedLines > maxHunkedLinesPerFileAndFlagCount {
			log.Warning.Printf("Found %d code reference lines in %s for the flag %s, which exceeded the limit of %d. Truncating code references for this path and flag.",
				numHunkedLines, path, flag, maxHunkedLinesPerFileAndFlagCount)
			return hunks
		}
	}

	return hunks
}

func initHunk(projKey, flagKey string) ld.HunkRep {
	return ld.HunkRep{
		ProjKey: projKey,
		FlagKey: flagKey,
	}
}

func makeTimestamp() int64 {
	return int64(time.Now().UnixNano()) / int64(time.Millisecond)
}

// Truncate lines to prevent sending over massive hunks, e.g. a minified file.
// NOTE: We may end up truncating a valid flag key reference. We accept this risk
//       and will handle hunks missing flag key references on the frontend.
func truncateLine(line string) string {
	// len(line) returns number of bytes, not num. characters, but it's a close enough
	// approximation for our purposes
	if len(line) > maxLineCharCount {
		// convert to rune slice so that we don't truncate multibyte unicode characters
		runes := []rune(line)
		return string(runes[0:maxLineCharCount]) + "…"
	} else {
		return line
	}
}
