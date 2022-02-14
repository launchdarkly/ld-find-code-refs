package ld

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/launchdarkly/ld-find-code-refs/internal/log"
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
			err := client.PostDeleteBranchesTask("test", []string{"master"})
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
		{"succeeds", 200, `{"items":[{"name":"master"}]}`, nil},
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
