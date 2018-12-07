package ld

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/facebookgo/httpcontrol"

	ldapi "github.com/launchdarkly/api-client-go"

	"github.com/launchdarkly/git-flag-parser/parse/internal/log"
)

type ApiClient struct {
	ldClient   *ldapi.APIClient
	httpClient *http.Client
	Options    ApiOptions
}

type ApiOptions struct {
	ApiKey  string
	ProjKey string
	BaseUri string
}

const (
	v2ApiPath = "/api/v2"
	reposPath = v2ApiPath + "/code-refs/repositories"
)

func InitApiClient(options ApiOptions) ApiClient {
	transport := httpcontrol.Transport{
		RequestTimeout: 3 * time.Second,
		DialTimeout:    3 * time.Second,
		DialKeepAlive:  1 * time.Minute,
		MaxTries:       3,
	}

	if options.BaseUri == "" {
		options.BaseUri = "https://app.launchdarkly.com"
	}

	return ApiClient{
		ldClient: ldapi.NewAPIClient(&ldapi.Configuration{
			BasePath:  options.BaseUri + v2ApiPath,
			UserAgent: "github-actor",
		}),
		httpClient: &http.Client{Transport: &transport},
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

func (c ApiClient) PostCodeReferenceRepository(repo RepoParams) error {
	// Custom repos don't allow owners, so swallow the owner if configured for a custom repo.
	if repo.Type == "custom" && repo.Owner != "" {
		log.Info("Ignoring repoOwner because repoType is 'custom'", nil)
		repo.Owner = ""
	}
	repoBytes, err := json.Marshal(repo)
	if err != nil {
		return err
	}
	postUrl := fmt.Sprintf("%s%s", c.Options.BaseUri, reposPath)
	log.Debug("Attempting to create code reference repository", log.Field("url", postUrl))
	req, err := http.NewRequest("POST", postUrl, bytes.NewBuffer(repoBytes))
	if err != nil {
		return err
	}
	res, err := c.do(req)
	if err != nil {
		return err
	}

	log.Debug("LaunchDarkly POST repository endpoint responded with status "+res.Status, log.Field("url", postUrl))
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusConflict {
		return fmt.Errorf("error creating repository")
	}

	return err
}

func (c ApiClient) PutCodeReferenceBranch(branch BranchRep, repo RepoParams) error {
	branchBytes, err := json.Marshal(branch)
	if err != nil {
		return err
	}
	putUrl := fmt.Sprintf("%s%s/%s/%s/%s/branches/%s", c.Options.BaseUri, reposPath, repo.Type, repo.Owner, repo.Name, url.PathEscape(branch.Name))
	if repo.Type == "custom" {
		putUrl = fmt.Sprintf("%s%s/custom/%s/branches/%s", c.Options.BaseUri, reposPath, repo.Name, url.PathEscape(branch.Name))
	}
	log.Debug("Sending code references", log.Field("url", putUrl))
	req, err := http.NewRequest("PUT", putUrl, bytes.NewBuffer(branchBytes))
	if err != nil {
		return err
	}
	res, err := c.do(req)
	if err != nil {
		return err
	}

	log.Debug("LaunchDarkly PUT branches endpoint responded with status "+res.Status, log.Field("url", putUrl))

	if res.StatusCode != 200 {
		return fmt.Errorf("error creating branches")
	}

	return nil
}

func (c ApiClient) do(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", c.Options.ApiKey)
	req.Header.Add("Content-Type", "application/json")
	return c.httpClient.Do(req)
}

type RepoParams struct {
	Type  string `json:"type"`
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

type BranchRep struct {
	Name       string              `json:"name"`
	Head       string              `json:"head"`
	PushTime   int64               `json:"pushTime"`
	SyncTime   int64               `json:"syncTime"`
	IsDefault  bool                `json:"isDefault"`
	References []ReferenceHunksRep `json:"references,omitempty"`
}

type ReferenceHunksRep struct {
	Path  string    `json:"path"`
	Hunks []HunkRep `json:"hunks"`
}

type HunkRep struct {
	Offset  int    `json:"offset"`
	Lines   string `json:"lines,omitempty"`
	ProjKey string `json:"projKey"`
	FlagKey string `json:"flagKey"`
}
