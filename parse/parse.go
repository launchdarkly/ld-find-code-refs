package parse

import (
	"io/ioutil"
	"os"
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

	cmd := git.Commander{Workspace: o.Dir.Value(), Head: currBranch, RepoName: o.RepoName.Value()}
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

	ldApi := ld.InitApiClient(ld.ApiOptions{ApiKey: o.AccessToken.Value(), BaseUri: o.BaseUri.Value()})
	flags, err := getFlags(ldApi)
	if err != nil {
		fatal("Unable to retrieve flag keys", err)
	}
	if len(flags) == 0 {
		log.Info("No flag keys found for selected project, exiting early", log.Field("projKey", o.ProjKey.Value()))
		os.Exit(0)
	}
	b := &branch{Name: currBranch, IsDefault: o.DefaultBranch.Value() == currBranch, PushTime: o.PushTime.Value(), Head: headSha}
	b, err = b.findReferences(cmd, flags)
	if err != nil {
		fatal("Error searching for flag key references", err)
	}

	err = ldApi.PutCodeReferenceBranch(b.MakeBranchRep(), ld.RepoParams{Type: o.RepoType.Value(), Name: o.RepoName.Value(), Owner: o.RepoOwner.Value()})
	if err != nil {
		fatal("Error sending code references to LaunchDarkly", err)
	}
}

func getFlags(ldApi ld.ApiClient) ([]string, error) {
	projKey := o.ProjKey.Value()
	log.Debug("Requesting flag list from LaunchDarkly", log.Field("projKey", projKey))
	flags, err := ldApi.GetFlagKeyList(projKey)
	if err != nil {
		log.Error("Error retrieving flag list from LaunchDarkly", err, log.Field("projKey", projKey))
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

func (b *branch) MakeBranchRep() ld.BranchRep {
	return ld.BranchRep{Name: b.Name, Head: b.Head, PushTime: b.PushTime, SyncTime: b.SyncTime, IsDefault: b.IsDefault, References: b.References.MakeReferenceReps()}
}

type reference struct {
	Path     string
	LineNum  int
	Context  string
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

func (r references) MakeReferenceReps() []ld.ReferenceRep {
	pathMap := referencePathMap{}
	for _, ref := range r {
		if pathMap[ref.Path] == nil {
			pathMap[ref.Path] = []reference{}
		}
		pathMap[ref.Path] = append(pathMap[ref.Path], ref)
	}

	reps := []ld.ReferenceRep{}
	for k, refs := range pathMap {
		if len(refs) > 0 {
			reps = append(reps, ld.ReferenceRep{Path: k, Hunks: refs.MakeHunkReps()})
		}
	}

	return reps
}

// MakeHunkReps coallesces single-line references into hunks per flag-key
func (r references) MakeHunkReps() []ld.HunkRep {
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
		nextHunkBuilder.WriteString(v.Context)
	}

	for _, flagKey := range currLine.FlagKeys {
		hunks = append(hunks, ld.HunkRep{Offset: currLine.LineNum, Lines: nextHunkBuilder.String(), ProjKey: o.ProjKey.Value(), FlagKey: flagKey})
	}
	if len(currLine.FlagKeys) == 0 {
		hunks = append(hunks, ld.HunkRep{Offset: currLine.LineNum, Lines: nextHunkBuilder.String()})
	}
	return hunks
}

func (b *branch) findReferences(cmd git.Commander, flags []string) (*branch, error) {
	err := cmd.Checkout()
	if err != nil {
		return b, err
	}

	grepResult, err := cmd.Grep(flags, o.ContextLines.Value())
	if err != nil {
		return b, err
	}

	b.SyncTime = makeTimestamp()
	b.References = generateReferencesFromGrep(flags, grepResult)
	return b, nil
}

func generateReferencesFromGrep(flags []string, grepResult [][]string) []reference {
	references := []reference{}
	ctxLines := o.ContextLines.Value()

	for _, r := range grepResult {
		path := r[1]
		contextContainsFlagKey := r[2] == ":"
		lineNumber := r[3]
		context := r[4]
		lineNum, err := strconv.Atoi(lineNumber)
		if err != nil {
			fatal("encountered an error generating flag references", err)
		}
		ref := reference{Path: path, LineNum: lineNum}
		if contextContainsFlagKey {
			ref.FlagKeys = findReferencedFlags(context, flags)
		}
		if ctxLines >= 0 {
			ref.Context = context
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
