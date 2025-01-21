package ld

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	h "github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"

	"github.com/bucketeer-io/code-refs/internal/log"
)

func TestMain(m *testing.M) {
	log.Init(true)
	os.Exit(m.Run())
}

func TestPostCodeReferenceRepository(t *testing.T) {
	specs := []struct {
		name           string
		responseStatus int
		expectedErr    error
	}{
		{"succeeds", 200, nil},
		{"succeeds on conflict", 409, ConflictErr},
	}
	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(tt.responseStatus)
			}))
			defer testServer.Close()

			retryMax := 0
			client := InitApiClient(ApiOptions{ApiKey: "api-x", ProjKey: "default", BaseUri: testServer.URL, RetryMax: &retryMax})
			err := client.postCodeReferenceRepository(RepoParams{Type: "custom", Name: "test"})
			require.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestGetCodeReferenceRepository(t *testing.T) {
	specs := []struct {
		name           string
		responseStatus int
		responseBody   string
		expectedErr    error
	}{
		{"succeeds", 200, `{"name":"test","type":"custom","sourceLink":"https://example.org"}`, nil},
		{"fails on not found", 404, ``, NotFoundErr},
	}
	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(tt.responseStatus)
				_, err := res.Write([]byte(tt.responseBody))
				require.NoError(t, err)
			}))
			defer testServer.Close()

			retryMax := 0
			client := InitApiClient(ApiOptions{ApiKey: "api-x", ProjKey: "default", BaseUri: testServer.URL, RetryMax: &retryMax})
			_, err := client.getCodeReferenceRepository("test")
			require.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestPatchCodeReferenceRepository(t *testing.T) {
	specs := []struct {
		name           string
		oldRepo        RepoParams
		newRepo        RepoParams
		responseStatus int
		expectedErr    error
	}{
		{"succeeds", RepoParams{Url: "github.com"}, RepoParams{Url: "bitbucket.com"}, 200, nil},
		{"fails on 404", RepoParams{Url: "github.com"}, RepoParams{Url: "bitbucket.com"}, 404, NotFoundErr},
	}
	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(tt.responseStatus)
			}))
			defer testServer.Close()

			retryMax := 0
			client := InitApiClient(ApiOptions{ApiKey: "api-x", ProjKey: "default", BaseUri: testServer.URL, RetryMax: &retryMax})
			err := client.patchCodeReferenceRepository(tt.oldRepo, tt.newRepo)
			require.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestPutCodeReferenceBranch(t *testing.T) {
	specs := []struct {
		name           string
		responseStatus int
		expectedErr    error
	}{
		{"succeeds", 200, nil},
	}

	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(tt.responseStatus)
			}))
			defer testServer.Close()

			retryMax := 0
			client := InitApiClient(ApiOptions{ApiKey: "api-x", ProjKey: "default", BaseUri: testServer.URL, RetryMax: &retryMax})
			err := client.PutCodeReferenceBranch(BranchRep{}, "test")
			require.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestPostDeleteBranchesTask(t *testing.T) {
	specs := []struct {
		name           string
		responseStatus int
		expectedErr    error
	}{
		{"succeeds", 200, nil},
	}

	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(tt.responseStatus)
			}))
			defer testServer.Close()

			retryMax := 0
			client := InitApiClient(ApiOptions{ApiKey: "api-x", ProjKey: "default", BaseUri: testServer.URL, RetryMax: &retryMax})
			err := client.PostDeleteBranchesTask("test", []string{"main"})
			require.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestGetCodeReferenceRepositoryBranches(t *testing.T) {
	specs := []struct {
		name           string
		responseStatus int
		responseBody   string
		expectedErr    error
	}{
		{"succeeds", 200, `{"items":[{"name":"main"}]}`, nil},
		{"fails on not found", 404, ``, NotFoundErr},
	}
	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(tt.responseStatus)
				_, err := res.Write([]byte(tt.responseBody))
				require.NoError(t, err)
			}))
			defer testServer.Close()

			retryMax := 0
			client := InitApiClient(ApiOptions{ApiKey: "api-x", ProjKey: "default", BaseUri: testServer.URL, RetryMax: &retryMax})
			_, err := client.GetCodeReferenceRepositoryBranches("test")
			require.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestCountAll(t *testing.T) {
	flagKey := "testFlag"

	h := HunkRep{
		StartingLineNumber: 1,
		Lines:              "testtest",
		ProjKey:            "example",
		FlagKey:            flagKey,
		Aliases:            []string{},
	}
	b := BranchRep{
		Name:             "",
		Head:             "",
		UpdateSequenceId: nil,
		SyncTime:         0,
		References: []ReferenceHunksRep{{
			Hunks: []HunkRep{h},
		}},
	}
	count := b.CountAll()
	want := make(map[string]int64)
	want[flagKey] = 1
	require.Equal(t, count, want)

}

func TestCountByProjectAndFlag(t *testing.T) {
	flagKey := "testFlag"
	notFoundKey := "notFoundFlag"
	notFoundKey2 := "notFoundFlag2"

	projectKey := "exampleProject"
	h := HunkRep{
		StartingLineNumber: 1,
		Lines:              "testtest",
		ProjKey:            projectKey,
		FlagKey:            flagKey,
		Aliases:            []string{},
	}
	notFound := HunkRep{
		StartingLineNumber: 1,
		Lines:              "testtest",
		ProjKey:            "notfound",
		FlagKey:            flagKey,
		Aliases:            []string{},
	}
	b := BranchRep{
		Name:             "",
		Head:             "",
		UpdateSequenceId: nil,
		SyncTime:         0,
		References: []ReferenceHunksRep{{
			Hunks: []HunkRep{h, notFound},
		}},
	}
	projects := []string{"exampleProject"}
	elements := [][]string{{flagKey, notFoundKey, notFoundKey2}}
	count := b.CountByProjectAndFlag(elements, projects)
	want := make(map[string]map[string]int64)
	want[projectKey] = make(map[string]int64)
	want[projectKey][flagKey] = 1
	want[projectKey][notFoundKey] = 0
	want[projectKey][notFoundKey2] = 0
	require.Equal(t, count, want)

}

func TestRateLimitBackoff(t *testing.T) {
	// Backoff instance where the time is always 0
	backoff := RateLimitBackoff(func() time.Time { return time.Unix(0, 0) }, h.DefaultBackoff)

	defaultBackoff := time.Second * time.Duration(1)

	invalidRateLimitReset := "abc"
	validRateLimitReset := "2000"
	pastRateLimitReset := "-1000"
	specs := []struct {
		name           string
		status         int
		rateLimitReset *string
		expected       time.Duration
	}{
		{"falls back to default backoff due to status", http.StatusBadGateway, nil, defaultBackoff},
		{"falls back to default backoff due to missing header", http.StatusTooManyRequests, nil, defaultBackoff},
		{"falls back to default backoff due to invalid header", http.StatusTooManyRequests, &invalidRateLimitReset, defaultBackoff},
		{"returns difference between reset and current time", http.StatusTooManyRequests, &validRateLimitReset, time.Second * time.Duration(2)},
		{"returns 0 because reset is in past", http.StatusTooManyRequests, &pastRateLimitReset, time.Duration(0)},
	}
	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			testUrl, _ := url.Parse("http://test.example/")
			req := &http.Request{
				Method: "POST",
				URL:    testUrl,
			}
			resp := &http.Response{
				StatusCode: tt.status,
				Header:     make(http.Header),
				Request:    req,
			}

			if tt.rateLimitReset != nil {
				resp.Header.Set("X-Ratelimit-Reset", *tt.rateLimitReset)
			}

			actual := backoff(defaultBackoff, time.Second*time.Duration(10), 0, resp)
			require.Equal(t, tt.expected, actual)
		})
	}
}
