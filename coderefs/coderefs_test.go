package coderefs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/launchdarkly/ld-find-code-refs/v2/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/v2/internal/log"
	"github.com/launchdarkly/ld-find-code-refs/v2/options"
)

func init() {
	log.Init(true)
}
func TestMain(m *testing.M) {
	log.Init(true)
	os.Exit(m.Run())
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
			branches:       []string{"main", "another-branch"},
			remoteBranches: []string{"main"},
			expected:       []string{"another-branch"},
		},
		{
			name:           "no stale branches",
			branches:       []string{"main"},
			remoteBranches: []string{"main"},
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

func Test_GenerateAliases(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "alias-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create .launchdarkly directory
	ldDir := filepath.Join(tmpDir, ".launchdarkly")
	err = os.MkdirAll(ldDir, 0755)
	require.NoError(t, err)

	// Create a simple coderefs.yaml file with alias configuration
	configContent := `aliases:
  - type: camelcase
  - type: snakecase
`
	configPath := filepath.Join(ldDir, "coderefs.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	t.Run("local mode with flag key", func(t *testing.T) {
		opts := options.AliasOptions{
			Dir:     tmpDir,
			FlagKey: "test-flag",
		}

		// This should not return an error for local mode
		err := GenerateAliases(opts)
		require.NoError(t, err)
	})

	t.Run("validation error for invalid dir", func(t *testing.T) {
		opts := options.AliasOptions{
			Dir:     "/nonexistent/directory/path",
			FlagKey: "test-flag",
		}

		err := GenerateAliases(opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not validate directory option")
	})
}
