package ld

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"

	h "github.com/hashicorp/go-retryablehttp"
	"github.com/olekukonko/tablewriter"

	ldapi "github.com/launchdarkly/api-client-go"
	jsonpatch "github.com/launchdarkly/json-patch"
	"github.com/launchdarkly/ld-find-code-refs/internal/log"
)

type ApiClient struct {
	ldClient   *ldapi.APIClient
	httpClient *h.Client
	Options    ApiOptions
}

type ApiOptions struct {
	ApiKey    string
	ProjKey   string
	BaseUri   string
	UserAgent string
	RetryMax  *int
}

const (
	v2ApiPath = "/api/v2"
	reposPath = v2ApiPath + "/code-refs/repositories"
)

var (
	NotFoundErr                       = errors.New("not found")
	ConflictErr                       = errors.New("conflict")
	EntityTooLargeErr                 = errors.New("entity too large")
	RateLimitExceededErr              = errors.New("rate limit exceeded")
	InternalServiceErr                = errors.New("internal service error")
	ServiceUnavailableErr             = errors.New("service unavailable")
	UnauthorizedErr                   = errors.New("unauthorized, check your LaunchDarkly access token")
	UnknownErr                        = errors.New("an unknown error occured")
	RepositoryDisabledErr             = errors.New("repository is disabled")
	BranchUpdateSequenceIdConflictErr = errors.New("updateSequenceId conflict")
)

func InitApiClient(options ApiOptions) ApiClient {
	if options.BaseUri == "" {
		options.BaseUri = "https://app.launchdarkly.com"
	}
	client := h.NewClient()
	client.Logger = log.Debug
	if options.RetryMax != nil && *options.RetryMax >= 0 {
		client.RetryMax = *options.RetryMax
	}
	return ApiClient{
		ldClient: ldapi.NewAPIClient(&ldapi.Configuration{
			BasePath:  options.BaseUri + v2ApiPath,
			UserAgent: options.UserAgent,
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

	_, err = c.do(req)
	if err != nil {
		return err
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

	resBytes, err := ioutil.ReadAll(res.Body)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	var repo RepoRep
	err = json.Unmarshal(resBytes, &repo)
	if err != nil {
		return nil, err
	}
	return &repo, err
}

func (c ApiClient) GetCodeReferenceRepositoryBranches(repoName string) ([]BranchRep, error) {
	req, err := h.NewRequest("GET", fmt.Sprintf("%s/%s/branches", c.repoUrl(), repoName), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.do(req)
	if err != nil {
		return nil, err
	}

	resBytes, err := ioutil.ReadAll(res.Body)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	var branches BranchCollection
	err = json.Unmarshal(resBytes, &branches)
	if err != nil {
		return nil, err
	}
	return branches.Items, err
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

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

func (c ApiClient) MaybeUpsertCodeReferenceRepository(repo RepoParams) error {
	currentRepo, err := c.getCodeReferenceRepository(repo.Name)
	if err != nil && err != NotFoundErr {
		return fmt.Errorf("error retrieving repository: %s", err)
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
			DefaultBranch:     currentRepo.DefaultBranch,
		}

		// Don't patch templates if command line arguments are not provided.
		// This is done because the LaunchDarkly API may return autogenerated url templates for non-custom connections.
		if currentRepo.Type != "custom" {
			if repo.CommitUrlTemplate == "" {
				currentRepoParams.CommitUrlTemplate = ""
			}
			if repo.HunkUrlTemplate == "" {
				currentRepoParams.HunkUrlTemplate = ""
			}
		}

		// If defaultBranch is absent and repo already exists, do nothing
		if currentRepoParams.DefaultBranch == "" {
			currentRepoParams.DefaultBranch = repo.DefaultBranch
		}

		if !reflect.DeepEqual(currentRepoParams, repo) {
			err = c.patchCodeReferenceRepository(currentRepoParams, repo)
			if err != nil {
				return fmt.Errorf("error updating repository: %s", err)
			}
		}
		return nil
	}

	err = c.postCodeReferenceRepository(repo)
	if err != nil {
		return fmt.Errorf("error creating repository: %s", err)
	}

	return nil
}

func (c ApiClient) PutCodeReferenceBranch(branch BranchRep, repoName string) error {
	branchBytes, err := json.Marshal(branch)
	if err != nil {
		return err
	}
	putUrl := fmt.Sprintf("%s%s/%s/branches/%s", c.Options.BaseUri, reposPath, repoName, url.PathEscape(branch.Name))
	req, err := h.NewRequest("PUT", putUrl, bytes.NewBuffer(branchBytes))
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

func (c ApiClient) PostDeleteBranchesTask(repoName string, branches []string) error {
	body, err := json.Marshal(branches)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s%s/%s/branch-delete-tasks", c.Options.BaseUri, reposPath, repoName)
	req, err := h.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

type ldErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (c ApiClient) do(req *h.Request) (*http.Response, error) {
	req.Header.Set("Authorization", c.Options.ApiKey)
	req.Header.Set("User-Agent", c.Options.UserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.FormatInt(req.ContentLength, 10))
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Check for all general status codes returned by the code references API, attempting to deconstruct LD error messages, if possible.
	switch res.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		return res, nil
	default:
		resBytes, err := ioutil.ReadAll(res.Body)
		if res != nil {
			defer res.Body.Close()
		}
		if err != nil {
			return nil, err
		}
		var ldErr ldErrorResponse
		err = json.Unmarshal(resBytes, &ldErr)

		if err == nil {
			switch ldErr.Code {
			case "updateSequenceId_conflict":
				return res, BranchUpdateSequenceIdConflictErr
			case "not_found":
				return res, NotFoundErr
			case "":
				// do nothing
			default:
				return res, fmt.Errorf("%s, %s", ldErr.Code, ldErr.Message)
			}
		}
		// The LaunchDarkly API should guarantee that we never have to fallback to these generic error messages, but we have them as a safeguard
		return res, fallbackErrorForStatus(res.StatusCode)
	}
}

func fallbackErrorForStatus(code int) error {
	switch code {
	case http.StatusBadRequest:
		return errors.New("bad request")
	case http.StatusUnauthorized:
		return errors.New("unauthorized, check your LaunchDarkly access token")
	case http.StatusNotFound:
		return NotFoundErr
	case http.StatusConflict:
		return ConflictErr
	case http.StatusRequestEntityTooLarge:
		return EntityTooLargeErr
	case http.StatusTooManyRequests:
		return RateLimitExceededErr
	case http.StatusInternalServerError:
		return InternalServiceErr
	case http.StatusServiceUnavailable:
		return ServiceUnavailableErr
	default:
		return fmt.Errorf("LaunchDarkly API responded with status code %d", code)
	}
}

type RepoParams struct {
	Type              string `json:"type"`
	Name              string `json:"name"`
	Url               string `json:"sourceLink"`
	CommitUrlTemplate string `json:"commitUrlTemplate"`
	HunkUrlTemplate   string `json:"hunkUrlTemplate"`
	DefaultBranch     string `json:"defaultBranch"`
}

type RepoRep struct {
	Type              string `json:"type"`
	Name              string `json:"name"`
	Url               string `json:"sourceLink"`
	CommitUrlTemplate string `json:"commitUrlTemplate"`
	HunkUrlTemplate   string `json:"hunkUrlTemplate"`
	DefaultBranch     string `json:"defaultBranch"`
	Enabled           bool   `json:"enabled,omitempty"`
}

type BranchCollection struct {
	Items []BranchRep `json:"items"`
}

type BranchRep struct {
	Name             string              `json:"name"`
	Head             string              `json:"head"`
	UpdateSequenceId *int64              `json:"updateSequenceId,omitempty"`
	SyncTime         int64               `json:"syncTime"`
	References       []ReferenceHunksRep `json:"references,omitempty"`
}

func (b BranchRep) TotalHunkCount() int {
	count := 0
	for _, r := range b.References {
		count += len(r.Hunks)
	}
	return count
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

type tableData [][]string

func (t tableData) Len() int {
	return len(t)
}

func (t tableData) Less(i, j int) bool {
	first, _ := strconv.ParseInt(t[i][1], 10, 32)
	second, _ := strconv.ParseInt(t[j][1], 10, 32)
	return first > second
}

func (t tableData) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

const maxFlagKeysDisplayed = 50

func (b BranchRep) PrintReferenceCountTable() {
	data := tableData{}
	refCountByFlag := map[string]int64{}
	for _, ref := range b.References {
		for _, hunk := range ref.Hunks {
			refCountByFlag[hunk.FlagKey]++
		}
	}
	for k, v := range refCountByFlag {
		data = append(data, []string{k, strconv.FormatInt(v, 10)})
	}
	sort.Sort(data)

	truncatedData := data
	var additionalRefCount int64 = 0
	if len(truncatedData) > maxFlagKeysDisplayed {
		truncatedData = data[0:maxFlagKeysDisplayed]

		for _, v := range data[maxFlagKeysDisplayed:] {
			i, _ := strconv.ParseInt(v[1], 10, 64)
			additionalRefCount += i
		}
	}
	truncatedData = append(truncatedData, []string{"Other flags", strconv.FormatInt(additionalRefCount, 10)})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Flag", "# References"})
	table.SetBorder(false)
	table.AppendBulk(truncatedData)
	table.Render()
}
