package parse

import (
	"io/ioutil"
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

func Parse() {
	err, cb := o.Init()
	if err != nil {
		log.Fatal("Unable to validate command line options", err)
		cb()
		os.Exit(1)
	}

	currBranch := o.RepoHead.Value()

	cmd := git.Git{Workspace: o.Dir.Value(), Head: currBranch, RepoName: o.RepoName.Value()}
	endpoint := o.CloneEndpoint.Value()
	if endpoint != "" {
		dir, err := ioutil.TempDir("", cmd.RepoName)
		if err != nil {
			fatal("Failed to create temp directory for repo clone", err)
		}
		defer os.RemoveAll(dir)
		cmd.Workspace = dir
		err = cmd.Clone(endpoint)
		if err != nil {
			fatal("Unable to clone repo", err)
		}
	}
	headSha, err := cmd.RevParse()
	if err != nil {
		fatal("Unable to parse current commit sha", err)
	}
	projKey := o.ProjKey.Value()
	ldApi := ld.InitApiClient(ld.ApiOptions{ApiKey: o.AccessToken.Value(), BaseUri: o.BaseUri.Value(), ProjKey: projKey})
	flags, err := getFlags(ldApi)
	if err != nil {
		fatal("Unable to retrieve flag keys", err)
	}
	if len(flags) == 0 {
		log.Info("No flag keys found for selected project, exiting early", log.Field("projKey", projKey))
		os.Exit(0)
	}
	ctxLines := o.ContextLines.Value()
	b := &branch{Name: currBranch, IsDefault: o.DefaultBranch.Value() == currBranch, PushTime: o.PushTime.Value(), SyncTime: makeTimestamp(), Head: headSha}

	// exclude option has already been validated as regex
	exclude, _ := regexp.Compile(o.Exclude.Value())
	refs, err := b.findReferences(cmd, flags, ctxLines, exclude)
	if err != nil {
		fatal("Error searching for flag key references", err)
	}
	b.References = refs

	err = ldApi.PutCodeReferenceBranch(b.makeBranchRep(projKey), ld.RepoParams{Type: o.RepoType.Value(), Name: o.RepoName.Value(), Owner: o.RepoOwner.Value()})
	if err != nil {
		fatal("Error sending code references to LaunchDarkly", err)
	}
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

// TODO: add links
type branch struct {
	Name       string
	Head       string
	IsDefault  bool
	PushTime   int64
	SyncTime   int64
	References grepResultLines
}

func (b *branch) makeBranchRep(projKey string) ld.BranchRep {
	return ld.BranchRep{
		Name:       strings.TrimPrefix(b.Name, "refs/heads/"),
		Head:       b.Head,
		PushTime:   b.PushTime,
		SyncTime:   b.SyncTime,
		IsDefault:  b.IsDefault,
		References: b.References.makeReferenceHunksReps(projKey),
	}
}

type grepResultLine struct {
	Path     string
	LineNum  int
	LineText string
	FlagKeys []string
}

type grepResultLines []grepResultLine

type flagReferenceMap map[string][]*grepResultLine

type fileGrepResults struct {
	flagReferenceMap    flagReferenceMap
	fileGrepResultLines []grepResultLine
}

type grepResultPathMap map[string]*fileGrepResults

func (fgr fileGrepResults) areEmpty() bool {
	return len(fgr.fileGrepResultLines) == 0
}

func (g grepResultLines) makeReferenceHunksReps(projKey string) []ld.ReferenceHunksRep {
	pathMap := g.groupIntoPathMap()

	reps := []ld.ReferenceHunksRep{}
	for path, grepResults := range pathMap {
		if !grepResults.areEmpty() {
			reps = append(reps, ld.ReferenceHunksRep{Path: path, Hunks: grepResults.makeHunkReps(projKey)})
		}
	}
	return reps
}

func (frm flagReferenceMap) addFlagReference(key string, ref *grepResultLine) {
	if frm[key] == nil {
		frm[key] = []*grepResultLine{ref}
	} else {
		frm[key] = append(frm[key], ref)
	}
}

func (pathMap grepResultPathMap) getOrInitResultsForPath(path string) *fileGrepResults {
	_, ok := pathMap[path]

	if !ok {
		pathMap[path] = &fileGrepResults{
			flagReferenceMap:    flagReferenceMap{},
			fileGrepResultLines: []grepResultLine{},
		}
	}

	return pathMap[path]
}

func (fgr *fileGrepResults) addGrepResult(grepResult grepResultLine) {
	fgr.fileGrepResultLines = append(fgr.fileGrepResultLines, grepResult)
}

func (fgr *fileGrepResults) addFlagReference(key string, ref *grepResultLine) {
	_, ok := fgr.flagReferenceMap[key]

	if ok {
		fgr.flagReferenceMap[key] = append(fgr.flagReferenceMap[key], ref)
	} else {
		fgr.flagReferenceMap[key] = []*grepResultLine{ref}
	}
}

func (g grepResultLines) groupIntoPathMap() grepResultPathMap {
	pathMap := grepResultPathMap{}

	for _, grepResult := range g {
		rescopedGrepResult := grepResult

		resultsForPath := pathMap.getOrInitResultsForPath(rescopedGrepResult.Path)

		resultsForPath.addGrepResult(rescopedGrepResult)

		if len(grepResult.FlagKeys) > 0 {
			for _, flagKey := range grepResult.FlagKeys {
				resultsForPath.addFlagReference(flagKey, &rescopedGrepResult)
			}
		}
	}

	return pathMap
}

// MakeHunkReps coallesces single-line references into hunks per flag-key
func (r fileGrepResults) makeHunkReps(projKey string) []ld.HunkRep {
	return []ld.HunkRep{}
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

func makeTimestamp() int64 {
	return int64(time.Now().UnixNano()) / int64(time.Millisecond)
}

func fatal(msg string, err error) {
	log.Fatal(msg, err)
	os.Exit(1)
}
