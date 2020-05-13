package coderefs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/launchdarkly/ld-find-code-refs/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/internal/log"
)

func init() {
	log.Init(true)
}
func TestMain(m *testing.M) {
	log.Init(true)
	os.Exit(m.Run())
}

func Test_filterShortFlags(t *testing.T) {
	// Note: these specs assume minFlagKeyLen is 3
	tests := []struct {
		name  string
		flags []string
		want  []string
	}{
		{
			name:  "Empty input/output",
			flags: []string{},
			want:  []string{},
		},
		{
			name:  "all flags are too short",
			flags: []string{"a", "b"},
			want:  []string{},
		},
		{
			name:  "some flags are too short",
			flags: []string{"abcdefg", "b", "ab", "abc"},
			want:  []string{"abcdefg", "abc"},
		},
		{
			name:  "no flags are too short",
			flags: []string{"catsarecool", "dogsarecool"},
			want:  []string{"catsarecool", "dogsarecool"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := filterShortFlagKeys(tt.flags)
			require.Equal(t, tt.want, got)
		})
	}
}
func Test_calculateStaleBranches(t *testing.T) {
	specs := []struct {
		name           string
		branches       []string
		remoteBranches []string
		expected       []string
	}{
		{
			name:           "stale branch",
			branches:       []string{"master", "another-branch"},
			remoteBranches: []string{"master"},
			expected:       []string{"another-branch"},
		},
		{
			name:           "no stale branches",
			branches:       []string{"master"},
			remoteBranches: []string{"master"},
			expected:       []string{},
		},
	}

	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			// transform test args into the format expected by calculateStaleBranches
			branchReps := make([]ld.BranchRep, 0, len(tt.branches))
			for _, b := range tt.branches {
				branchReps = append(branchReps, ld.BranchRep{Name: b})
			}
			remoteBranchMap := map[string]bool{}
			for _, b := range tt.remoteBranches {
				remoteBranchMap[b] = true
			}

			assert.ElementsMatch(t, tt.expected, calculateStaleBranches(branchReps, remoteBranchMap))
		})
	}
}
