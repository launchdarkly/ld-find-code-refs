package ld

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	h "github.com/hashicorp/go-retryablehttp"

	ldapi "github.com/launchdarkly/api-client-go"

	"github.com/launchdarkly/git-flag-parser/internal/log"
	jsonpatch "github.com/launchdarkly/json-patch"
)

type ApiClient struct {
	ldClient   *ldapi.APIClient
	httpClient *h.Client
	Options    ApiOptions
}

type ApiOptions struct {
	ApiKey   string
	ProjKey  string
	BaseUri  string
	RetryMax *int
}

const (
	v2ApiPath = "/api/v2"
	reposPath = v2ApiPath + "/code-refs/repositories"
)

var (
	RepositoryPostErr     = fmt.Errorf("error creating repository")
	RepositoryPatchErr    = fmt.Errorf("error updating repository")
	RepositoryGetErr      = fmt.Errorf("error retrieving repository")
	RepositoryNotFoundErr = fmt.Errorf("repository not found")
	RepositoryDisabledErr = fmt.Errorf("repository is disabled")
	BranchPutErr          = fmt.Errorf("error updating branch")
)

func InitApiClient(options ApiOptions) ApiClient {
	if options.BaseUri == "" {
		options.BaseUri = "https://app.launchdarkly.com"
	}
	client := h.NewClient()
	if options.RetryMax != nil && *options.RetryMax >= 0 {
		client.RetryMax = *options.RetryMax
	}
	return ApiClient{
		ldClient: ldapi.NewAPIClient(&ldapi.Configuration{
			BasePath:  options.BaseUri + v2ApiPath,
			UserAgent: "github-actor",
		}),
		httpClient: client,
		Options:    options,
	}
}

func (c ApiClient) GetFlagKeyList() ([]string, error) {
	ctx := context.WithValue(context.Background(), ldapi.ContextAPIKey, ldapi.APIKey{Key: c.Options.ApiKey})
	flags, _, err := c.ldClient.FeatureFlagsApi.GetFeatureFlags(ctx, c.Options.ProjKey, nil)
	if err != nil {
		return nil, err
	}
	flagKeys := make([]string, 0, len(flags.Items))
	for _, flag := range flags.Items {
		flagKeys = append(flagKeys, flag.Key)
	}
	return flagKeys, nil
}

func (c ApiClient) repoUrl() string {
	return fmt.Sprintf("%s%s", c.Options.BaseUri, reposPath)
}

func (c ApiClient) patchCodeReferenceRepository(currentRepo, repo RepoParams) error {
	originalBytes, err := json.Marshal(currentRepo)
	if err != nil {
		return err
	}
	newBytes, err := json.Marshal(repo)
	if err != nil {
		return err
	}

	patch, err := jsonpatch.CreateMergePatch(originalBytes, newBytes)
	if err != nil {
		return err
	}

	req, err := h.NewRequest("PATCH", fmt.Sprintf("%s/%s", c.repoUrl(), repo.Name), bytes.NewBuffer(patch))
	if err != nil {
		return err
	}
	res, err := c.do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return RepositoryPatchErr
	}
	return nil
}

func (c ApiClient) getCodeReferenceRepository(name string) (*RepoRep, error) {
	req, err := h.NewRequest("GET", fmt.Sprintf("%s/%s", c.repoUrl(), name), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusOK {
		resBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		var repo RepoRep
		err = json.Unmarshal(resBytes, &repo)
		if err != nil {
			return nil, err
		}
		return &repo, err
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, RepositoryNotFoundErr
	}
	return nil, RepositoryGetErr
}

func (c ApiClient) postCodeReferenceRepository(repo RepoParams) error {
	repoBytes, err := json.Marshal(repo)
	if err != nil {
		return err
	}

	req, err := h.NewRequest("POST", c.repoUrl(), bytes.NewBuffer(repoBytes))
	if err != nil {
		return err
	}
	res, err := c.do(req)
	if err != nil {
		return RepositoryPostErr
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusConflict {
		return RepositoryPostErr
	}

	return nil
}

func (c ApiClient) MaybeUpsertCodeReferenceRepository(repo RepoParams) error {
	currentRepo, err := c.getCodeReferenceRepository(repo.Name)
	if err != nil && err != RepositoryNotFoundErr {
		return err
	}
	if currentRepo != nil {
		if !currentRepo.Enabled {
			return RepositoryDisabledErr
		}
		currentRepoParams := RepoParams{
			Name:              currentRepo.Name,
			Type:              currentRepo.Type,
			Url:               currentRepo.Url,
			CommitUrlTemplate: currentRepo.CommitUrlTemplate,
			HunkUrlTemplate:   currentRepo.HunkUrlTemplate,
		}
		if !reflect.DeepEqual(currentRepoParams, repo) {
			return c.patchCodeReferenceRepository(currentRepoParams, repo)
		}
		return nil
	}

	return c.postCodeReferenceRepository(repo)
}

func (c ApiClient) PutCodeReferenceBranch(branch BranchRep, repoName string) error {
	branchBytes, err := json.Marshal(branch)
	if err != nil {
		return err
	}
	putUrl := fmt.Sprintf("%s%s/%s/branches/%s", c.Options.BaseUri, reposPath, repoName, url.PathEscape(branch.Name))
	log.Debug("Sending code references", log.Field("url", putUrl))
	req, err := h.NewRequest("PUT", putUrl, bytes.NewBuffer(branchBytes))
	if err != nil {
		return err
	}
	res, err := c.do(req)
	if err != nil {
		return BranchPutErr
	}

	log.Debug("LaunchDarkly PUT branches endpoint responded with status "+res.Status, log.Field("url", putUrl))

	if res.StatusCode != 200 {
		return BranchPutErr
	}

	return nil
}

func (c ApiClient) do(req *h.Request) (*http.Response, error) {
	req.Header.Add("Authorization", c.Options.ApiKey)
	req.Header.Add("Content-Type", "application/json")
	return c.httpClient.Do(req)
}

type RepoParams struct {
	Type              string `json:"type"`
	Name              string `json:"name"`
	Url               string `json:"sourceLink"`
	CommitUrlTemplate string `json:"commitUrlTemplate"`
	HunkUrlTemplate   string `json:"hunkUrlTemplate"`
}

type RepoRep struct {
	Type              string `json:"type"`
	Name              string `json:"name"`
	Url               string `json:"sourceLink"`
	CommitUrlTemplate string `json:"commitUrlTemplate"`
	HunkUrlTemplate   string `json:"hunkUrlTemplate"`
	Enabled           bool   `json:"enabled,omitempty"`
}
type BranchRep struct {
	Name             string              `json:"name"`
	Head             string              `json:"head"`
	UpdateSequenceId *int64              `json:"updateSequenceId,omitempty"`
	SyncTime         int64               `json:"syncTime"`
	IsDefault        bool                `json:"isDefault"`
	References       []ReferenceHunksRep `json:"references,omitempty"`
}

type ReferenceHunksRep struct {
	Path  string    `json:"path"`
	Hunks []HunkRep `json:"hunks"`
}

type HunkRep struct {
	StartingLineNumber int    `json:"startingLineNumber"`
	Lines              string `json:"lines,omitempty"`
	ProjKey            string `json:"projKey"`
	FlagKey            string `json:"flagKey"`
}
