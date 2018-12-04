package ld

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	ldapi "github.com/launchdarkly/api-client-go"

	"github.com/launchdarkly/git-flag-parser/parse/internal/log"
)

type ApiClient struct {
	client  *ldapi.APIClient
	Options ApiOptions
}

type ApiOptions struct {
	ApiKey  string
	ProjKey string
	BaseUri string
}

func InitApiClient(options ApiOptions) ApiClient {
	if options.BaseUri == "" {
		options.BaseUri = "https://app.launchdarkly.com"
	}
	return ApiClient{
		client: ldapi.NewAPIClient(&ldapi.Configuration{
			BasePath:  options.BaseUri + "/api/v2",
			UserAgent: "github-actor",
		}),
		Options: options,
	}
}

func (c ApiClient) GetFlagKeyList() ([]string, error) {
	ctx := context.WithValue(context.Background(), ldapi.ContextAPIKey, ldapi.APIKey{Key: c.Options.ApiKey})
	flags, _, err := c.client.FeatureFlagsApi.GetFeatureFlags(ctx, c.Options.ProjKey, nil)
	if err != nil {
		return nil, err
	}
	flagKeys := make([]string, 0, len(flags.Items))
	for _, flag := range flags.Items {
		flagKeys = append(flagKeys, flag.Key)
	}
	return flagKeys, nil
}

func (c ApiClient) PostCodeReferenceRepository(repo RepoParams) error {
	// Custom repos don't allow owners, so swallow the owner if configured for a custom repo.
	if repo.Type == "custom" {
		repo.Owner = ""
	}
	repoBytes, err := json.Marshal(repo)
	if err != nil {
		return err
	}
	postUrl := fmt.Sprintf("%s/api/v2/code-refs/repositories", c.Options.BaseUri)
	log.Debug("Attempting to create code reference repository", log.Field("url", postUrl))
	req, err := http.NewRequest("POST", postUrl, bytes.NewBuffer(repoBytes))
	if err != nil {
		return err
	}
	res, err := c.do(req)
	log.Debug("LaunchDarkly POST repository endpoint responded with status "+res.Status, log.Field("url", postUrl))
	return err
}

func (c ApiClient) PutCodeReferenceBranch(branch BranchRep, repo RepoParams) error {
	branchBytes, err := json.Marshal(branch)
	if err != nil {
		return err
	}
	putUrl := fmt.Sprintf("%s/api/v2/code-refs/repositories/%s/%s/%s/branches/%s", c.Options.BaseUri, repo.Type, repo.Owner, repo.Name, url.PathEscape(branch.Name))
	if repo.Type == "custom" {
		putUrl = fmt.Sprintf("%s/api/v2/code-refs/repositories/custom/%s/branches/%s", c.Options.BaseUri, repo.Name, url.PathEscape(branch.Name))
	}
	log.Debug("Sending code references", log.Field("url", putUrl))
	req, err := http.NewRequest("PUT", putUrl, bytes.NewBuffer(branchBytes))
	if err != nil {
		return err
	}
	res, err := c.do(req)
	log.Debug("LaunchDarkly PUT branches endpoint responded with status "+res.Status, log.Field("url", putUrl))
	return err
}

func (c ApiClient) do(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", c.Options.ApiKey)
	req.Header.Add("Content-Type", "application/json")
	// TODO: Use a client with timeouts and retries
	client := http.Client{}
	return client.Do(req)
}

type RepoParams struct {
	Type  string `json:"type"`
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

type BranchRep struct {
	Name       string         `json:"name"`
	Head       string         `json:"head"`
	PushTime   int64          `json:"pushTime"`
	SyncTime   int64          `json:"syncTime"`
	IsDefault  bool           `json:"isDefault"`
	References []ReferenceRep `json:"references,omitempty"`
}

type ReferenceRep struct {
	Path  string    `json:"path"`
	Hunks []HunkRep `json:"hunks"`
}

type HunkRep struct {
	Offset  int    `json:"offset"`
	Lines   string `json:"lines,omitempty"`
	ProjKey string `json:"projKey"`
	FlagKey string `json:"flagKey"`
}
