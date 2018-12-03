package parse

import (
	"io/ioutil"
	"os"
	"regexp"
	"sort"
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

	err = ldApi.PutCodeReferenceBranch(b.makeBranchRep(), ld.RepoParams{Type: o.RepoType.Value(), Name: o.RepoName.Value(), Owner: o.RepoOwner.Value()})
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
	References references
}

func (b *branch) makeBranchRep() ld.BranchRep {
	return ld.BranchRep{Name: strings.TrimPrefix(b.Name, "refs/heads/"), Head: b.Head, PushTime: b.PushTime, SyncTime: b.SyncTime, IsDefault: b.IsDefault, References: b.References.makeReferenceReps()}
}

type reference struct {
	Path     string
	LineNum  int
	LineText string
	FlagKeys []string
}

type references []reference
type referencePathMap map[string]references

func (r references) Len() int {
	return len(r)
}

func (r references) Less(i, j int) bool {
	return r[i].LineNum < r[j].LineNum
}

func (r references) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r references) makeReferenceReps() []ld.ReferenceHunksRep {
	pathMap := referencePathMap{}
	for _, ref := range r {
		if pathMap[ref.Path] == nil {
			pathMap[ref.Path] = []reference{}
		}
		pathMap[ref.Path] = append(pathMap[ref.Path], ref)
	}

	reps := []ld.ReferenceHunksRep{}
	for k, refs := range pathMap {
		if len(refs) > 0 {
			reps = append(reps, ld.ReferenceHunksRep{Path: k, Hunks: refs.makeHunkReps()})
		}
	}
	return reps
}

// MakeHunkReps coallesces single-line references into hunks per flag-key
func (r references) makeHunkReps() []ld.HunkRep {
	sort.Sort(r)
	hunks := []ld.HunkRep{}
	var nextHunkBuilder strings.Builder
	currLine := r[0]
	for i, v := range r {
		if i != 0 {
			if r[i-1].LineNum != r[i].LineNum-1 {
				lineNum := currLine.LineNum
				for _, flagKey := range v.FlagKeys {
					hunks = append(hunks, ld.HunkRep{Offset: lineNum, Lines: nextHunkBuilder.String(), ProjKey: o.ProjKey.Value(), FlagKey: flagKey})
				}
				if len(v.FlagKeys) == 0 {
					hunks = append(hunks, ld.HunkRep{Offset: lineNum, Lines: nextHunkBuilder.String()})
				}
				nextHunkBuilder.Reset()
				currLine = v
			} else {
				nextHunkBuilder.WriteString("\n")
			}
		}
		nextHunkBuilder.WriteString(v.LineText)
	}

	for _, flagKey := range currLine.FlagKeys {
		hunks = append(hunks, ld.HunkRep{Offset: currLine.LineNum, Lines: nextHunkBuilder.String(), ProjKey: o.ProjKey.Value(), FlagKey: flagKey})
	}
	if len(currLine.FlagKeys) == 0 {
		hunks = append(hunks, ld.HunkRep{Offset: currLine.LineNum, Lines: nextHunkBuilder.String()})
	}
	return hunks
}

func (b *branch) findReferences(cmd git.Git, flags []string, ctxLines int, exclude *regexp.Regexp) (references, error) {
	err := cmd.Checkout()
	if err != nil {
		return references{}, err
	}

	grepResult, err := cmd.Grep(flags, ctxLines)
	if err != nil {
		return references{}, err
	}

	return generateReferencesFromGrep(flags, grepResult, ctxLines, exclude), nil
}

func generateReferencesFromGrep(flags []string, grepResult [][]string, ctxLines int, exclude *regexp.Regexp) []reference {
	references := []reference{}

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
		ref := reference{Path: path, LineNum: lineNum}
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
