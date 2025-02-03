package bucketeer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bucketeer-io/code-refs/internal/log"
	"github.com/bucketeer-io/code-refs/options"
	"github.com/cenkalti/backoff/v4"
)

type ReferenceHunksRep struct {
	Path  string
	Hunks []HunkRep
}

type HunkRep struct {
	FlagKey            string
	StartingLineNumber int
	Lines              string
	ContentHash        string
	Aliases            []string
	FileExt            string
}

// NumLines returns the number of lines in the hunk
func (h HunkRep) NumLines() int {
	return strings.Count(h.Lines, "\n") + 1
}

// Overlap returns the number of overlapping lines between the receiver (h) and the parameter (hr) hunkreps
// The return value will be negative if the hunks do not overlap
func (h HunkRep) Overlap(hr HunkRep) int {
	aLines := strings.Split(h.Lines, "\n")
	bLines := strings.Split(hr.Lines, "\n")

	aStart := h.StartingLineNumber
	aEnd := aStart + len(aLines)
	bStart := hr.StartingLineNumber
	bEnd := bStart + len(bLines)

	if bStart > aEnd || aStart > bEnd {
		return -1
	}

	if bStart >= aStart {
		return len(aLines) - (bStart - aStart)
	}
	return len(bLines) - (aStart - bStart)
}

type ApiClient interface {
	GetFlagKeyList(opts options.Options) ([]string, error)
	CreateCodeReference(opts options.Options, ref CodeReference) error
	UpdateCodeReference(opts options.Options, id string, ref CodeReference) error
	DeleteCodeReference(opts options.Options, id string) error
}

type ApiOptions struct {
	ApiKey    string
	BaseUri   string
	UserAgent string
	RetryMax  *int
}

type RepoParams struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Url  string `json:"url,omitempty"`
}

type CodeReference struct {
	ID               string   `json:"id,omitempty"`
	FeatureID        string   `json:"featureId"`
	FilePath         string   `json:"filePath"`
	FileExtension    string   `json:"fileExtension"`
	LineNumber       int      `json:"lineNumber"`
	CodeSnippet      string   `json:"codeSnippet"`
	ContentHash      string   `json:"contentHash"`
	Aliases          []string `json:"aliases"`
	RepositoryName   string   `json:"repositoryName"`
	RepositoryOwner  string   `json:"repositoryOwner"`
	RepositoryType   string   `json:"repositoryType"`
	RepositoryBranch string   `json:"repositoryBranch"`
	CommitHash       string   `json:"commitHash"`
	EnvironmentID    string   `json:"environmentId"`
	CreatedAt        int64    `json:"createdAt,omitempty"`
	UpdatedAt        int64    `json:"updatedAt,omitempty"`
}

type apiClient struct {
	apiKey    string
	baseUri   string
	userAgent string
	retryMax  int
	client    *http.Client
}

func InitApiClient(opts ApiOptions) ApiClient {
	retryMax := 3
	if opts.RetryMax != nil {
		retryMax = *opts.RetryMax
	}
	return &apiClient{
		apiKey:    opts.ApiKey,
		baseUri:   opts.BaseUri,
		userAgent: opts.UserAgent,
		retryMax:  retryMax,
		client:    &http.Client{},
	}
}

func (c *apiClient) do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	op := func() error {
		var err error
		resp, err = c.client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 500 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return fmt.Errorf("server error: %s", body)
		}
		return nil
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = time.Duration(c.retryMax) * time.Second

	err := backoff.Retry(op, b)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *apiClient) GetFlagKeyList(opts options.Options) ([]string, error) {
	url := fmt.Sprintf("%s/v1/features?pageSize=1000", c.baseUri)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", c.apiKey)
	req.Header.Add("User-Agent", c.userAgent)

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if opts.Debug {
		body, _ := io.ReadAll(resp.Body)
		log.Debug.Printf("[GetFlagKeyList] Response Status: %d, Body: %s", resp.StatusCode, string(body))
		// Create a new reader with the body content for the subsequent json.Decode
		resp.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	var response struct {
		Features []struct {
			ID string `json:"id"`
		} `json:"features"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	var flags []string
	for _, feature := range response.Features {
		flags = append(flags, feature.ID)
	}
	return flags, nil
}

func (c *apiClient) CreateCodeReference(opts options.Options, ref CodeReference) error {
	url := fmt.Sprintf("%s/v1/code_references", c.baseUri)
	body, err := json.Marshal(ref)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", c.apiKey)
	req.Header.Add("User-Agent", c.userAgent)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if opts.Debug {
		body, _ := io.ReadAll(resp.Body)
		log.Debug.Printf("[CreateCodeReference] Response Status: %d, Body: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *apiClient) UpdateCodeReference(opts options.Options, id string, ref CodeReference) error {
	url := fmt.Sprintf("%s/v1/code_references/%s", c.baseUri, id)
	body, err := json.Marshal(ref)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", c.apiKey)
	req.Header.Add("User-Agent", c.userAgent)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if opts.Debug {
		body, _ := io.ReadAll(resp.Body)
		log.Debug.Printf("[UpdateCodeReference] Response Status: %d, Body: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *apiClient) DeleteCodeReference(opts options.Options, id string) error {
	url := fmt.Sprintf("%s/v1/code_references/%s", c.baseUri, id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", c.apiKey)
	req.Header.Add("User-Agent", c.userAgent)

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if opts.Debug {
		body, _ := io.ReadAll(resp.Body)
		log.Debug.Printf("[DeleteCodeReference] Response Status: %d, Body: %s", resp.StatusCode, string(body))
	}
	return nil
}
