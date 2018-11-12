package parse

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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

	b := &branch{Branch: currBranch, IsDefault: o.DefaultBranch.Value() == currBranch, PushTime: o.PushTime.Value(), Head: headSha}
	b, err = b.findReferences(cmd, flags)
	if err != nil {
		fatal("Error searching for flag key references", err)
	}

	branchBytes, err := json.Marshal(b)
	if err != nil {
		fatal("Error marshalling branch to json", err)
	}
	err = ldApi.PutCodeReferenceBranch(branchBytes)
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

type branch struct {
	Branch     string      `json:"branch"`
	Head       string      `json:"head"`
	IsDefault  bool        `json:"isDefault"`
	PushTime   uint64      `json:"pushTime"`
	SyncTime   uint64      `json:"syncTime"`
	References []reference `json:"references,omitempty"`
}

type reference struct {
	Path     string   `json:"path"`
	Line     string   `json:"line"`
	Context  string   `json:"context,omitempty"`
	FlagKeys []string `json:"flagKeys,omitempty"`
}

func (b *branch) findReferences(cmd git.Commander, flags []string) (*branch, error) {
	err := cmd.Checkout()
	if err != nil {
		return b, err
	}

	grepResult, err := cmd.Grep(flags, o.ContextLines.Value(), o.Exclude.Value())
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

		ref := reference{Path: path, Line: lineNumber}
		if contextContainsFlagKey {
			ref.FlagKeys = findReferencedFlags(context, flags)
		}
		if ctxLines > 0 {
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

func makeTimestamp() uint64 {
	return uint64(time.Now().UnixNano()) / uint64(time.Millisecond)
}

func fatal(msg string, err error) {
	log.Fatal(msg, err)
	os.Exit(1)
}
