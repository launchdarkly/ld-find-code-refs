package main

import (
	"os"
	"testing"

	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	o "github.com/launchdarkly/ld-find-code-refs/options"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	log.Init(true)
	os.Exit(m.Run())
}

func TestParseBranch(t *testing.T) {
	specs := []struct {
		name        string
		in          string
		allowTags   bool
		event       *Event
		expectedOut string
		expectError bool
	}{
		{
			name:        "succeeds for well formed branch input",
			in:          "refs/heads/a",
			allowTags:   false,
			expectedOut: "a",
			expectError: false,
		},
		{
			name:        "succeeds for well formed branch input when tags are enabled",
			in:          "refs/heads/a",
			allowTags:   true,
			expectedOut: "a",
			expectError: false,
		},
		{
			name:        "succeeds for well formed tag input when tags are enabled",
			in:          "refs/tags/a",
			allowTags:   true,
			expectedOut: "a",
			expectError: false,
		},
		{
			name:        "works for branches with slashes",
			in:          "refs/heads/a/b",
			allowTags:   false,
			expectedOut: "a/b",
			expectError: false,
		},
		{
			name:        "works for branches with different character types",
			in:          "refs/heads/a-b.1+*",
			allowTags:   false,
			expectedOut: "a-b.1+*",
			expectError: false,
		},
		{
			name:        "returns an error for poorly formed input",
			in:          "notaref",
			allowTags:   false,
			expectedOut: "",
			expectError: true,
		},
		{
			name:        "returns an error for an empty branch name",
			in:          "refs/heads/",
			allowTags:   false,
			expectedOut: "",
			expectError: true,
		},
		{
			name:        "returns the event branch name for an invalid GITHUB_REF",
			in:          "refs/pull/1",
			allowTags:   false,
			expectedOut: "master",
			expectError: false,
			event:       &Event{Pull: &Pull{Head: Head{Ref: "master"}}},
		},
		{
			name:        "returns the event tag name for an invalid GITHUB_REF",
			in:          "random",
			allowTags:   true,
			expectedOut: "v1.0.0",
			expectError: false,
			event:       &Event{Release: &Release{TagName: "v1.0.0"}},
		},
		{
			name:        "returns an err for tags when tags are not allowed",
			in:          "refs/tags/a",
			allowTags:   false,
			expectedOut: "",
			expectError: true,
		},
		{
			name:        "returns an err for tags when tags are not allowed, even when event has tag name",
			in:          "random",
			allowTags:   false,
			expectedOut: "",
			expectError: true,
			event:       &Event{Release: &Release{TagName: "v1.0.0"}},
		},
	}
	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			out, err := parseBranch(tt.in, tt.event, tt.allowTags)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOut, out)
			}
		})
	}
}

func TestMergeGithubOptions_withCliRepoName(t *testing.T) {
	os.Setenv("GITHUB_REF", "refs/heads/test")
	var options o.Options = o.Options{
		AccessToken: "deaf-beef",
		ProjKey:     "project-x",
		RepoName:    "myapp-react",
	}
	result, _ := mergeGithubOptions(options)
	assert.Equal(t, "myapp-react", result.RepoName)
}

func TestMergeGithubOptions_withGithubRepoName(t *testing.T) {
	os.Setenv("GITHUB_REPOSITORY", "yusinto/myapp-golang")
	os.Setenv("GITHUB_REF", "refs/heads/test")
	var options o.Options = o.Options{
		AccessToken: "deaf-beef",
		ProjKey:     "project-x",
	}
	result, _ := mergeGithubOptions(options)
	assert.Equal(t, "myapp-golang", result.RepoName)
}
