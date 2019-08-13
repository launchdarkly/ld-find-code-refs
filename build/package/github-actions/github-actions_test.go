package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBranch(t *testing.T) {
	specs := []struct {
		name        string
		in          string
		expectedOut string
		expectError bool
	}{
		{
			name:        "succeeds for well formed input",
			in:          "refs/heads/a",
			expectedOut: "a",
			expectError: false,
		},
		{
			name:        "works for branches with slashes",
			in:          "refs/heads/a/b",
			expectedOut: "a/b",
			expectError: false,
		},
		{
			name:        "works for branches with different character types",
			in:          "refs/heads/a-b.1+*",
			expectedOut: "a-b.1+*",
			expectError: false,
		},
		{
			name:        "returns an error for poorly formed input",
			in:          "notaref",
			expectedOut: "",
			expectError: true,
		},
		{
			name:        "returns an error for an empty branch name",
			in:          "refs/heads/",
			expectedOut: "",
			expectError: true,
		},
	}
	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			out, err := parseBranch(tt.in)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOut, out)
			}
		})
	}
}
