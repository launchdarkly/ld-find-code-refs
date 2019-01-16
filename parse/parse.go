package parse

import (
	"container/list"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/launchdarkly/git-flag-parser/parse/internal/git"
	"github.com/launchdarkly/git-flag-parser/parse/internal/ld"
	"github.com/launchdarkly/git-flag-parser/parse/internal/log"
	o "github.com/launchdarkly/git-flag-parser/parse/internal/options"
)

// These are defensive limits intended to prevent corner cases stemming from
// large repos, false positives, etc. The goal is a) to prevent the parser
// from taking a very long time to run and b) to prevent the parser from
// PUTing a massive json payload. These limits will likely be tweaked over
// time. The LaunchDarkly backend will also apply limits.
const minFlagKeyLen = 3
const maxFileCount = 5000
const maxLineCharCount = 500
const maxHunkCount = 5000
const maxHunksPerFileCount = 1000
const maxHunkedLinesPerFileAndFlagCount = 500

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
	IsDefault        bool
	UpdateSequenceId *int64
	SyncTime         int64
	GrepResults      grepResultLines
}

func Parse() {
	err, cb := o.Init()
	if err != nil {
		log.Error("Unable to validate command line options", err, nil)
		cb()
		os.Exit(1)
	}

	currBranch := o.RepoHead.Value()

	cmd := git.Git{Workspace: o.Dir.Value(), Head: currBranch, RepoName: o.RepoName.Value()}

	// TODO: Reintroduce this codepath if we decide the flag parser should be able to clone repos
	// endpoint := o.CloneEndpoint.Value()
	// if endpoint != "" {
	// 	dir, err := ioutil.TempDir("", cmd.RepoName)
	// 	if err != nil {
	// 		fatal("Failed to create temp directory for repo clone", err)
	// 	}
	// 	defer os.RemoveAll(dir)
	// 	cmd.Workspace = dir
	// 	err = cmd.Clone(endpoint)
	// 	if err != nil {
	// 		fatal("Unable to clone repo", err)
	// 	}
	// }

	headSha, err := cmd.RevParse()
	if err != nil {
		fatal("Unable to parse current commit sha", err)
	}
	projKey := o.ProjKey.Value()
	ldApi := ld.InitApiClient(ld.ApiOptions{ApiKey: o.AccessToken.Value(), BaseUri: o.BaseUri.Value(), ProjKey: projKey})
	repoParams := ld.RepoParams{
		Type:              o.RepoType.Value(),
		Name:              o.RepoName.Value(),
		Url:               o.RepoUrl.Value(),
		CommitUrlTemplate: o.CommitUrlTemplate.Value(),
		HunkUrlTemplate:   o.HunkUrlTemplate.Value(),
	}

	err = ldApi.MaybeUpsertCodeReferenceRepository(repoParams)
	if err != nil {
		fatal("Unable to connect repository to LaunchDarkly", err)
	}

	flags, err := getFlags(ldApi)
	if err != nil {
		fatal("Unable to retrieve flag keys", err)
	}
	if len(flags) == 0 {
		log.Info("No flag keys found for selected project, exiting early", log.Field("projKey", projKey))
		os.Exit(0)
	}

	filteredFlags := filterShortFlagKeys(flags)
	if len(filteredFlags) == 0 {
		msg := fmt.Sprintf("No flag keys larger than the min. flag key length of %v were found, exiting early", minFlagKeyLen)
		log.Info(msg, log.Field("projKey", projKey))
		os.Exit(0)
	}

	ctxLines := o.ContextLines.Value()
	var updateId *int64
	if o.UpdateSequenceId.Value() >= 0 {
		updateIdOption := o.UpdateSequenceId.Value()
		updateId = &updateIdOption
	}
	b := &branch{
		Name:             currBranch,
		IsDefault:        o.DefaultBranch.Value() == currBranch,
		UpdateSequenceId: updateId,
		SyncTime:         makeTimestamp(),
		Head:             headSha,
	}

	// exclude option has already been validated as regex
	exclude, _ := regexp.Compile(o.Exclude.Value())
	refs, err := b.findReferences(cmd, filteredFlags, ctxLines, exclude)
	if err != nil {
		fatal("Error searching for flag key references", err)
	}
	b.GrepResults = refs

	err = ldApi.PutCodeReferenceBranch(b.makeBranchRep(projKey, ctxLines), repoParams.Name)

	if err != nil {
		fatal("Error sending code references to LaunchDarkly", err)
	}
}

// Very short flag keys lead to many false positives when searching in code,
// so we filter them out.
func filterShortFlagKeys(flags []string) []string {
	filteredFlags := []string{}

	for _, flag := range flags {
		if len(flag) >= minFlagKeyLen {
			filteredFlags = append(filteredFlags, flag)
		}

	}

	return filteredFlags
}

func getFlags(ldApi ld.ApiClient) ([]string, error) {
	log.Debug("Requesting flag list from LaunchDarkly", log.Field("projKey", ldApi.Options.ProjKey))
	flags, err := ldApi.GetFlagKeyList()
	if err != nil {
		log.Error("Error retrieving flag list from LaunchDarkly", err, log.Field("projKey", ldApi.Options.ProjKey))
		return nil, err
	}
	return flags, nil
}

func (b *branch) findReferences(cmd git.Git, flags []string, ctxLines int, exclude *regexp.Regexp) (grepResultLines, error) {
	err := cmd.Checkout()
	if err != nil {
		return grepResultLines{}, err
	}

	grepResult, err := cmd.Grep(flags, ctxLines)
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
			fatal("encountered an error generating flag references", err)
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
		IsDefault:        b.IsDefault,
		References:       b.GrepResults.makeReferenceHunksReps(projKey, ctxLines),
	}
}

func (g grepResultLines) makeReferenceHunksReps(projKey string, ctxLines int) []ld.ReferenceHunksRep {
	reps := []ld.ReferenceHunksRep{}

	aggregatedGrepResults := g.aggregateByPath()

	if len(aggregatedGrepResults) > maxFileCount {
		log.Info("number of files containing code references exceeded limit",
			map[string]interface{}{"number of matched files": len(aggregatedGrepResults), "file limit": maxFileCount})
		aggregatedGrepResults = aggregatedGrepResults[0:maxFileCount]
	}

	numHunks := 0

	for _, fileGrepResults := range aggregatedGrepResults {
		if numHunks > maxHunkCount {
			log.Info("Exceeded maximum hunk limit, halting code reference search.",
				map[string]interface{}{"hunk count": numHunks, "limit": maxHunkCount})
			break
		}

		hunks := fileGrepResults.makeHunkReps(projKey, ctxLines)

		if len(hunks) > maxHunksPerFileCount {
			log.Info("Exceded hunk limit for file, truncating file hunks",
				map[string]interface{}{"hunk count": len(hunks), "limit": maxHunksPerFileCount, "path": fileGrepResults.path})
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
		log.Fatal("grep results returned out of order", nil)
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
		flagHunks := buildHunksForFlag(projKey, flagKey, flagReferences, fgr.fileGrepResultLines, ctxLines)
		hunks = append(hunks, flagHunks...)
	}

	return hunks
}

func buildHunksForFlag(projKey, flag string, flagReferences []*list.Element, fileLines *list.List, ctxLines int) []ld.HunkRep {
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
			log.Info("Exceeded permitted number of flag reference lines + context lines for file",
				map[string]interface{}{"flag": flag, "limit": maxHunkedLinesPerFileAndFlagCount})
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

func fatal(msg string, err error) {
	log.Fatal(msg, err)
	os.Exit(1)
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
