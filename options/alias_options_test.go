package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_AliasOptions_ValidateAliasOptions(t *testing.T) {
	t.Run("local mode validation - only dir required", func(t *testing.T) {
		opts := &AliasOptions{
			Dir:     "/some/dir",
			FlagKey: "test-flag",
		}

		err := opts.ValidateAliasOptions()
		require.NoError(t, err)
	})

	t.Run("local mode validation - missing dir defaults to cwd", func(t *testing.T) {
		opts := &AliasOptions{
			FlagKey: "test-flag",
		}

		err := opts.ValidateAliasOptions()
		require.NoError(t, err)
		assert.NotEmpty(t, opts.Dir) // Dir should be set to current working directory
	})

	t.Run("API mode validation - all required fields present", func(t *testing.T) {
		opts := &AliasOptions{
			Dir:         "/some/dir",
			AccessToken: "token",
			ProjKey:     "project",
		}

		err := opts.ValidateAliasOptions()
		require.NoError(t, err)
	})

	t.Run("API mode validation - missing access token", func(t *testing.T) {
		opts := &AliasOptions{
			Dir:     "/some/dir",
			ProjKey: "project",
		}

		err := opts.ValidateAliasOptions()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "accessToken")
	})

	t.Run("API mode validation - missing project key", func(t *testing.T) {
		opts := &AliasOptions{
			Dir:         "/some/dir",
			AccessToken: "token",
		}

		err := opts.ValidateAliasOptions()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "projKey/projects")
	})

	t.Run("API mode validation - with projects instead of projKey", func(t *testing.T) {
		opts := &AliasOptions{
			Dir:         "/some/dir",
			AccessToken: "token",
			Projects:    []Project{{Key: "project1"}},
		}

		err := opts.ValidateAliasOptions()
		require.NoError(t, err)
	})
}