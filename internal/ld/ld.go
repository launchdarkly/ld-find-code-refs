package ld

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	h "github.com/hashicorp/go-retryablehttp"
	"github.com/olekukonko/tablewriter"

	ldapi "github.com/launchdarkly/api-client-go/v15"
	jsonpatch "github.com/launchdarkly/json-patch"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/log"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/validation"
)

type ApiClient struct {
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
	apiVersion       = "20220603"
	apiVersionHeader = "LD-API-Version"
	v2ApiPath        = "/api/v2"
	reposPath        = "/code-refs/repositories"
)

type ConfigurationError struct {
	error
}

func newConfigurationError(e string) ConfigurationError {
	return ConfigurationError{errors.New((e))}
}

var (
	NotFoundErr                       = errors.New("not found")
	ConflictErr                       = errors.New("conflict")
	RateLimitExceededErr              = errors.New("rate limit exceeded")
	InternalServiceErr                = errors.New("internal service error")
	ServiceUnavailableErr             = errors.New("service unavailable")
	BranchUpdateSequenceIdConflictErr = errors.New("updateSequenceId conflict")
	RepositoryDisabledErr             = newConfigurationError("repository is disabled")
	UnauthorizedErr                   = newConfigurationError("unauthorized, check your LaunchDarkly access token")
	EntityTooLargeErr                 = newConfigurationError("entity too large")
)

// IsTransient returns true if the error returned by the LaunchDarkly API is either unexpected, or unable to be resolved by the user.
func IsTransient(err error) bool {
	var e ConfigurationError
	return !errors.As(err, &e)
}

// LaunchDarkly API uses the X-Ratelimit-Reset header to communicate when to retry after a 429
// Fallback to default backoff if header can't be parsed
// https://apidocs.launchdarkly.com/#section/Overview/Rate-limiting
// Method is curried in order to avoid stubbing the time package and fallback Backoff in unit tests
func RateLimitBackoff(now func() time.Time, fallbackBackoff h.Backoff) func(time.Duration, time.Duration, int, *http.Response) time.Duration {
	return func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
		if resp != nil {
			if resp.StatusCode == http.StatusTooManyRequests {
				if s, ok := resp.Header["X-Ratelimit-Reset"]; ok {
					if sleepUntil, err := strconv.ParseInt(s[0], 10, 64); err == nil {
						sleep := sleepUntil - now().UnixMilli()
						log.Info.Printf("rate limit for %s %s hit, sleeping for %d ms", resp.Request.Method, resp.Request.URL, sleep)

						if sleep > 0 {
							return time.Millisecond * time.Duration(sleep)
						} else {
							return time.Duration(0)
						}
					}
				}
			}
		}

		return fallbackBackoff(min, max, attemptNum, resp)
	}
}

func InitApiClient(options ApiOptions) ApiClient {
	if options.BaseUri == "" {
		options.BaseUri = "https://app.launchdarkly.com"
	}
	client := h.NewClient()
	client.Logger = log.Debug
	if options.RetryMax != nil && *options.RetryMax >= 0 {
		client.RetryMax = *options.RetryMax
	}
	client.Backoff = RateLimitBackoff(time.Now, h.LinearJitterBackoff)

	return ApiClient{
		httpClient: client,
		Options:    options,
	}
}

// path should have leading slash
func (c ApiClient) getPath(path string) string {
	return fmt.Sprintf("%s%s%s", c.Options.BaseUri, v2ApiPath, path)
}

func (c ApiClient) GetFlagKeyList(projKey string) ([]string, error) {
	env, err := c.getProjectEnvironment(projKey)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if env != nil {
		params.Add("env", env.Key)
	}
	activeFlags, err := c.getFlags(projKey, params)
	if err != nil {
		return nil, err
	}

	params.Add("filter", "state:archived")
	archivedFlags, err := c.getFlags(projKey, params)
	if err != nil {
		return nil, err
	}
	flags := make([]ldapi.FeatureFlag, 0, len(activeFlags)+len(archivedFlags))
	flags = append(flags, activeFlags...)
	flags = append(flags, archivedFlags...)

	flagKeys := make([]string, 0, len(flags))
	for _, flag := range flags {
		flagKeys = append(flagKeys, flag.Key)
	}

	return flagKeys, nil
}

// Get the first environment we can find for a project
func (c ApiClient) getProjectEnvironment(projKey string) (*ldapi.Environment, error) {
	urlStr := c.getPath(fmt.Sprintf("/projects/%s/environments", projKey))
	req, err := h.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("limit", "1")
	req.URL.RawQuery = params.Encode()

	res, err := c.do(req)
	if err != nil {
		return nil, err
	}

	resBytes, err := io.ReadAll(res.Body)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	var collection ldapi.Environments
	if err := json.Unmarshal(resBytes, &collection); err != nil {
		return nil, err
	}
	if len(collection.Items) == 0 {
		return nil, nil
	}

	env := collection.Items[0]

	return &env, nil
}

func (c ApiClient) getFlags(projKey string, params url.Values) ([]ldapi.FeatureFlag, error) {
	url := c.getPath(fmt.Sprintf("/flags/%s", projKey))
	req, err := h.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()

	res, err := c.do(req)
	if err != nil {
		return nil, err
	}

	resBytes, err := io.ReadAll(res.Body)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	var flags ldapi.FeatureFlags
	if err := json.Unmarshal(resBytes, &flags); err != nil {
		return nil, err
	}

	return flags.Items, nil
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

	req, err := h.NewRequest("PATCH", c.getPath(fmt.Sprintf("%s/%s", reposPath, repo.Name)), bytes.NewBuffer(patch))
	if err != nil {
		return err
	}

	if res, err := c.do(req); err != nil {
		return err
	} else if res != nil {
		defer res.Body.Close()
	}

	return nil
}

func (c ApiClient) getCodeReferenceRepository(name string) (*RepoRep, error) {
	req, err := h.NewRequest("GET", c.getPath(fmt.Sprintf("%s/%s", reposPath, name)), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.do(req)
	if err != nil {
		return nil, err
	}

	resBytes, err := io.ReadAll(res.Body)
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
	req, err := h.NewRequest("GET", c.getPath(fmt.Sprintf("%s/%s/branches", reposPath, repoName)), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.do(req)
	if err != nil {
		return nil, err
	}

	resBytes, err := io.ReadAll(res.Body)
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

	req, err := h.NewRequest("POST", c.getPath(reposPath), bytes.NewBuffer(repoBytes))
	if err != nil {
		return err
	}

	if res, err := c.do(req); err != nil {
		return err
	} else if res != nil {
		defer res.Body.Close()
	}

	return nil
}

func (c ApiClient) MaybeUpsertCodeReferenceRepository(repo RepoParams) error {
	currentRepo, err := c.getCodeReferenceRepository(repo.Name)
	if err != nil && err != NotFoundErr {
		return fmt.Errorf("error retrieving repository: %w", err)
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
				return fmt.Errorf("error updating repository: %w", err)
			}
		}
		return nil
	}

	err = c.postCodeReferenceRepository(repo)
	if err != nil {
		return fmt.Errorf("error creating repository: %w", err)
	}

	return nil
}

func (c ApiClient) PutCodeReferenceBranch(branch BranchRep, repoName string) error {
	branchBytes, err := json.Marshal(branch)
	if err != nil {
		return err
	}
	putUrl := c.getPath(fmt.Sprintf("%s/%s/branches/%s", reposPath, repoName, url.PathEscape(branch.Name)))
	req, err := h.NewRequest("PUT", putUrl, bytes.NewBuffer(branchBytes))
	if err != nil {
		return err
	}

	if res, err := c.do(req); err != nil {
		return err
	} else if res != nil {
		defer res.Body.Close()
	}

	return nil
}

func (c ApiClient) PostExtinctionEvents(extinctions []ExtinctionRep, repoName, branchName string) error {
	data, err := json.Marshal(extinctions)
	if err != nil {
		return err
	}
	url := c.getPath(fmt.Sprintf("%s/%s/branches/%s/extinction-events", reposPath, repoName, url.PathEscape(branchName)))
	req, err := h.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	if res, err := c.do(req); err != nil {
		return err
	} else if res != nil {
		defer res.Body.Close()
	}

	return nil
}

func (c ApiClient) PostDeleteBranchesTask(repoName string, branches []string) error {
	body, err := json.Marshal(branches)
	if err != nil {
		return err
	}
	url := c.getPath(fmt.Sprintf("%s/%s/branch-delete-tasks", reposPath, repoName))
	req, err := h.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	if res, err := c.do(req); err != nil {
		return err
	} else if res != nil {
		defer res.Body.Close()
	}

	return nil
}

type ldErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (c ApiClient) do(req *h.Request) (*http.Response, error) {
	req.Header.Set("Authorization", c.Options.ApiKey)
	req.Header.Set(apiVersionHeader, apiVersion)
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
		resBytes, err := io.ReadAll(res.Body)
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
			case "invalid_request":
				return res, errors.New(ldErr.Message)
			case "updateSequenceId_conflict":
				return res, BranchUpdateSequenceIdConflictErr
			case "not_found":
				return res, NotFoundErr
			case "request_entity_too_large":
				return res, EntityTooLargeErr
			case "":
				// do nothing
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
		return UnauthorizedErr
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
	UpdateSequenceId *int                `json:"updateSequenceId,omitempty"`
	SyncTime         int64               `json:"syncTime"`
	References       []ReferenceHunksRep `json:"references,omitempty"`
	CommitTime       int64               `json:"commitTime,omitempty"`
}

func (b BranchRep) TotalHunkCount() int {
	count := 0
	for _, r := range b.References {
		count += len(r.Hunks)
	}
	return count
}

func (b BranchRep) WriteToCSV(outDir, projKey, repo, sha string) (path string, err error) {
	// Try to create a filename with a shortened sha, but if the sha is too short for some unexpected reason, use the branch name instead
	var tag string
	if len(sha) >= 7 {
		tag = sha[:7]
	} else {
		tag = b.Name
	}

	absPath, err := validation.NormalizeAndValidatePath(outDir)
	if err != nil {
		return "", fmt.Errorf("invalid outDir '%s': %w", outDir, err)
	}
	path = filepath.Join(absPath, fmt.Sprintf("coderefs_%s_%s_%s.csv", projKey, repo, tag))

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	records := make([][]string, 0, len(b.References)+1)
	for _, ref := range b.References {
		records = append(records, ref.toRecords()...)
	}

	// sort csv by flag key
	sort.Slice(records, func(i, j int) bool {
		// sort by flagKey -> path -> startingLineNumber
		for k := 0; k < 3; k++ {
			if records[i][k] != records[j][k] {
				return records[i][k] < records[j][k]
			}
		}
		// above loop should always return since startingLineNumber is guaranteed to be unique
		return false
	})

	records = append([][]string{{"flagKey", "projKey", "path", "startingLineNumber", "lines", "aliases", "contentHash"}}, records...)
	return path, w.WriteAll(records)
}

type ReferenceHunksRep struct {
	Path  string    `json:"path"`
	Hunks []HunkRep `json:"hunks"`
}

func (r ReferenceHunksRep) toRecords() [][]string {
	ret := make([][]string, 0, len(r.Hunks))
	for _, hunk := range r.Hunks {
		ret = append(ret, []string{hunk.FlagKey, hunk.ProjKey, r.Path, strconv.FormatInt(int64(hunk.StartingLineNumber), 10), hunk.Lines, strings.Join(hunk.Aliases, " "), hunk.ContentHash})
	}
	return ret
}

type HunkRep struct {
	StartingLineNumber int      `json:"startingLineNumber"`
	Lines              string   `json:"lines,omitempty"`
	ProjKey            string   `json:"projKey"`
	FlagKey            string   `json:"flagKey"`
	Aliases            []string `json:"aliases,omitempty"`
	ContentHash        string   `json:"contentHash,omitempty"`
}

// Returns the number of lines overlapping between the receiver (h) and the parameter (hr) hunkreps
// The return value will be negative if the hunks do not overlap
func (h HunkRep) Overlap(hr HunkRep) int {
	return h.StartingLineNumber + h.NumLines() - hr.StartingLineNumber
}

func (h HunkRep) NumLines() int {
	return strings.Count(h.Lines, "\n") + 1
}

type ExtinctionRep struct {
	Revision string `json:"revision"`
	Message  string `json:"message"`
	Time     int64  `json:"time"`
	ProjKey  string `json:"projKey"`
	FlagKey  string `json:"flagKey"`
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

func (b BranchRep) CountAll() map[string]int64 {
	refCount := map[string]int64{}
	for _, ref := range b.References {
		for _, hunk := range ref.Hunks {
			refCount[hunk.FlagKey]++
		}
	}
	return refCount
}

func (b BranchRep) CountByProjectAndFlag(matcher [][]string, projects []string) map[string]map[string]int64 {
	refCountByFlag := map[string]map[string]int64{}
	for i, project := range projects {
		refCountByFlag[project] = map[string]int64{}
		for _, flag := range matcher[i] {
			refCountByFlag[project][flag] = 0
		}
		for _, ref := range b.References {
			for _, hunk := range ref.Hunks {
				if hunk.ProjKey == project {
					refCountByFlag[project][hunk.FlagKey]++
				}
			}
		}
	}
	return refCountByFlag
}

func (b BranchRep) PrintReferenceCountTable() {
	data := tableData{}

	for k, v := range b.CountAll() {
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
