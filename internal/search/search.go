package search

import (
	"container/list"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/launchdarkly/ld-find-code-refs/internal/helpers"
	"github.com/launchdarkly/ld-find-code-refs/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/internal/log"
)

const (
	ignoreFileName = ".ldignore"

	// These are defensive limits intended to prevent corner cases stemming from
	// large repos, false positives, etc. The goal is a) to prevent the program
	// from taking a very long time to run and b) to prevent the program from
	// PUTing a massive json payload. These limits will likely be tweaked over
	// time. The LaunchDarkly backend will also apply limits.
	maxFileCount                      = 10000 // Maximum number of files containing code references
	maxHunkCount                      = 25000 // Maximum number of total code references
	maxLineCharCount                  = 500   // Maximum number of characters per line
	maxHunkedLinesPerFileAndFlagCount = 500   // Maximum number of lines per flag in a file
)

// map of flag keys to slices of lines those flags occur on
type flagReferenceMap map[string][]*list.Element

// this struct contains a linked list of all the search result lines
// for a single file, and a map of flag keys to slices of lines where
// those flags occur.
type FileSearchResults struct {
	path                  string
	fileSearchResultLines *list.List
	flagReferenceMap
}

/*
searchRegex splits search result lines into groups
Group 1: File path.
Group 2: Separator. A colon indicates a match, a hyphen indicates a context lines
Group 3: Line number
Group 4: Line contents
*/
var searchRegex = regexp.MustCompile("([^:]+)(:|-)([0-9]+)[:-](.*)")

var SearchTooLargeErr = errors.New("regular expression is too large")
var NoSearchPatternErr = errors.New("failed to generate a valid search pattern")

// SafePaginationCharCount determines the maximum sum of flag key lengths to be used in a single smart paginated search.
// Safely bounded under the 2^16 limit of pcre_compile() with the parameters set by our underlying search tool (ag)
// https://github.com/vmg/pcre/blob/master/pcre_internal.h#L436
func SafePaginationCharCount() int {
	if runtime.GOOS == windows {
		// workaround win32 limitation on maximum command length
		// https://support.microsoft.com/en-us/help/830473/command-prompt-cmd-exe-command-line-string-limitation
		return 30000
	}

	return 60000
}

func FlagKeyCost(key string) int {
	// periods need to be escaped, so they count as 2 characters
	return len(key) + strings.Count(key, ".")
}

func DelimCost(delims []byte) int {
	return len(delims) * 2
}

type Searcher interface {
	FindReferences(flags []string, aliases map[string][]string, ctxLines int, delimiters string) (SearchResultLines, error)
}

type SearchResultLine struct {
	Path     string
	LineNum  int
	LineText string
	FlagKeys map[string][]string
}

type SearchResultLines []SearchResultLine

func (lines SearchResultLines) Len() int {
	return len(lines)
}

func (lines SearchResultLines) Less(i, j int) bool {
	if lines[i].Path < lines[j].Path {
		return true
	}
	if lines[i].Path > lines[j].Path {
		return false
	}
	return lines[i].LineNum < lines[j].LineNum
}

func (lines SearchResultLines) Swap(i, j int) {
	lines[i], lines[j] = lines[j], lines[i]
}

func (g SearchResultLines) MakeReferenceHunksReps(projKey string, ctxLines int) []ld.ReferenceHunksRep {
	reps := []ld.ReferenceHunksRep{}

	aggregatedSearchResults := g.aggregateByPath()

	if len(aggregatedSearchResults) > maxFileCount {
		log.Warning.Printf("found %d files with code references, which exceeded the limit of %d", len(aggregatedSearchResults), maxFileCount)
		aggregatedSearchResults = aggregatedSearchResults[0:maxFileCount]
	}

	numHunks := 0

	shouldSuppressUnexpectedError := false
	for _, fileSearchResults := range aggregatedSearchResults {
		if numHunks > maxHunkCount {
			log.Warning.Printf("found %d code references across all files, which exceeeded the limit of %d. halting code reference search", numHunks, maxHunkCount)
			break
		}

		hunks := fileSearchResults.makeHunkReps(projKey, ctxLines)

		if len(hunks) == 0 && !shouldSuppressUnexpectedError {
			log.Error.Printf("expected code references but found none in '%s'", fileSearchResults.path)
			log.Debug.Printf("%+v", fileSearchResults)
			// if this error occurred, it's likely to occur for many other files, and create a lot of noise. So, suppress the message for all other occurrences
			shouldSuppressUnexpectedError = true
			continue
		}

		numHunks += len(hunks)

		reps = append(reps, ld.ReferenceHunksRep{Path: fileSearchResults.path, Hunks: hunks})
	}
	return reps
}

// Assumes invariant: SearchResultLines will already be sorted by path.
func (g SearchResultLines) aggregateByPath() []FileSearchResults {
	allFileResults := []FileSearchResults{}

	if len(g) == 0 {
		return allFileResults
	}

	// initialize first file
	currentFileResults := FileSearchResults{
		path:                  g[0].Path,
		flagReferenceMap:      flagReferenceMap{},
		fileSearchResultLines: list.New(),
	}

	for _, searchResult := range g {
		// If we reach a search result with a new path, append the old one to our list and start a new one
		if searchResult.Path != currentFileResults.path {
			allFileResults = append(allFileResults, currentFileResults)

			currentFileResults = FileSearchResults{
				path:                  searchResult.Path,
				flagReferenceMap:      flagReferenceMap{},
				fileSearchResultLines: list.New(),
			}
		}

		elem := currentFileResults.addSearchResult(searchResult)

		if len(searchResult.FlagKeys) > 0 {
			for flagKey := range searchResult.FlagKeys {
				currentFileResults.addFlagReference(flagKey, elem)
			}
		}
	}

	// append last file
	allFileResults = append(allFileResults, currentFileResults)

	return allFileResults
}

func generateFlagRegex(flags []string) string {
	flagRegexes := []string{}
	for _, v := range flags {
		escapedFlag := regexp.QuoteMeta(v)
		flagRegexes = append(flagRegexes, escapedFlag)
	}
	return strings.Join(flagRegexes, "|")
}

func generateDelimiterRegex(delimiters []byte) (lookBehind, lookAhead string) {
	if len(delimiters) == 0 {
		return "", ""
	}
	delims := string(delimiters)
	lookBehind = fmt.Sprintf("(?<=[%s])", delims)
	lookAhead = fmt.Sprintf("(?=[%s])", delims)
	return lookBehind, lookAhead
}

func generateSearchPattern(flags []string, delimiters []byte, padPattern bool) string {
	flagRegex := generateFlagRegex(flags)
	lookBehind, lookAhead := generateDelimiterRegex(delimiters)
	if padPattern {
		// Padding the left-most and right-most search terms with the "a^" regular expression, which never matches anything. This is done to work-around strange behavior causing the left-most and right-most items to be ignored by ag on windows
		// example: (?<=[\"'\`])(a^|flag1|flag2|flag3|a^)(?=[\"'\`])"
		return lookBehind + "(a^|" + flagRegex + "|a^)" + lookAhead
	}
	// example: (?<=[\"'\`])(flag1|flag2|flag3)(?=[\"'\`])"
	return lookBehind + "(" + flagRegex + ")" + lookAhead
}

func (fsr *FileSearchResults) addSearchResult(searchResult SearchResultLine) *list.Element {
	prev := fsr.fileSearchResultLines.Back()
	if prev != nil && prev.Value.(SearchResultLine).LineNum > searchResult.LineNum {
		// This should never happen, as `ag` (and any other search program we might use
		// should always return search results sorted by line number. We sanity check
		// that lines are sorted _just in case_ since the downstream hunking algorithm
		// only works on sorted lines.
		log.Error.Fatalf("search results returned out of order")
	}

	return fsr.fileSearchResultLines.PushBack(searchResult)
}

func (fsr *FileSearchResults) addFlagReference(key string, ref *list.Element) {
	_, ok := fsr.flagReferenceMap[key]

	if ok {
		fsr.flagReferenceMap[key] = append(fsr.flagReferenceMap[key], ref)
	} else {
		fsr.flagReferenceMap[key] = []*list.Element{ref}
	}
}

func (fsr FileSearchResults) makeHunkReps(projKey string, ctxLines int) []ld.HunkRep {
	hunks := []ld.HunkRep{}

	for flagKey, flagReferences := range fsr.flagReferenceMap {
		flagHunks := buildHunksForFlag(projKey, flagKey, fsr.path, flagReferences, ctxLines)
		hunks = append(hunks, flagHunks...)
	}

	return hunks
}
func buildHunksForFlag(projKey, flag, path string, flagReferences []*list.Element, ctxLines int) []ld.HunkRep {
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
			if ptr.Value.(SearchResultLine).LineNum <= lastSeenLineNum {
				appendToPreviousHunk = true
			}
		}

		// If we are starting a new hunk, initialize it
		if !appendToPreviousHunk {
			currentHunk = initHunk(projKey, flag)
			currentHunk.StartingLineNumber = ptr.Value.(SearchResultLine).LineNum
			hunkStringBuilder.Reset()
		}

		// From the current position (at the theoretical start of the hunk) seek forward line by line X times,
		// where X = (numCtxLinesBeforeFlagRef + 1 + ctxLines). Note that if the flag reference occurs close to the
		// start of the file, numCtxLines may be smaller than ctxLines.
		//   For each line, check if we have seeked past the end of the last hunk
		//     If so: write that line to the hunkStringBuilder
		//     Record that line as the last seen line.
		for i := 0; i < numCtxLinesBeforeFlagRef+1+ctxLines; i++ {
			ptrLineNum := ptr.Value.(SearchResultLine).LineNum
			if ptrLineNum > lastSeenLineNum {
				lineText := truncateLine(ptr.Value.(SearchResultLine).LineText)
				hunkStringBuilder.WriteString(lineText + "\n")
				lastSeenLineNum = ptrLineNum
				numHunkedLines += 1
				currentHunk.Aliases = append(currentHunk.Aliases, ptr.Value.(SearchResultLine).FlagKeys[flag]...)
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
			currentHunk.Aliases = helpers.Dedupe(currentHunk.Aliases)
			hunks = append(hunks, currentHunk)
			previousHunk = &hunks[len(hunks)-1]
		}

		// If we have written more than the max. allowed number of lines for this file and flag, finish this hunk and exit early.
		// This guards against a situation where the user has very long files with many false positive matches.
		if numHunkedLines > maxHunkedLinesPerFileAndFlagCount {
			log.Warning.Printf("found %d code reference lines in %s for the flag %s, which exceeded the limit of %d. truncating code references for this path and flag.",
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
		Aliases: []string{},
	}
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
