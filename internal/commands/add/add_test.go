package add_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"github.com/brpaz/gh-secrets-sync/internal/commands/add"
	"github.com/brpaz/gh-secrets-sync/internal/config"
	"github.com/brpaz/gh-secrets-sync/internal/testutils"
)

func TestNew(t *testing.T) {
	cmd := add.New()
	assert.IsType(t, cmd, &cli.Command{})
}

func TestAddCommand(t *testing.T) {
	t.Run("adds secret with comma-separated repos", func(t *testing.T) {
		cfgPath := testutils.SetupConfig(t, &config.Config{})

		var out strings.Builder
		cmd := add.New()
		cmd.Writer = &out

		err := cmd.Run(t.Context(), []string{"add", "--name", "MY_TOKEN", "--value", "abc123", "--repos", "owner/repo1,owner/repo2"})
		require.NoError(t, err)

		assert.Contains(t, out.String(), "MY_TOKEN")
		assert.Contains(t, out.String(), "owner/repo1")
		assert.Contains(t, out.String(), "owner/repo2")

		loaded, err := config.Load(cfgPath)
		require.NoError(t, err)
		require.Len(t, loaded.Secrets, 1)
		assert.Equal(t, "MY_TOKEN", loaded.Secrets[0].Name)
		assert.Equal(t, "abc123", loaded.Secrets[0].Value)
		assert.Equal(t, []string{"owner/repo1", "owner/repo2"}, loaded.Secrets[0].Repositories)
	})

	t.Run("adds secret with repeated --repos flag", func(t *testing.T) {
		cfgPath := testutils.SetupConfig(t, &config.Config{})

		cmd := add.New()
		err := cmd.Run(t.Context(), []string{"add", "--name", "TOKEN", "--value", "val", "--repos", "owner/repo1", "--repos", "owner/repo2"})
		require.NoError(t, err)

		loaded, err := config.Load(cfgPath)
		require.NoError(t, err)
		require.Len(t, loaded.Secrets, 1)
		assert.Equal(t, []string{"owner/repo1", "owner/repo2"}, loaded.Secrets[0].Repositories)
	})

	t.Run("errors on duplicate without --force", func(t *testing.T) {
		testutils.SetupConfig(t, &config.Config{
			Secrets: []config.Secret{{Name: "MY_TOKEN", Value: "old", Repositories: []string{"owner/r"}}},
		})

		cmd := add.New()
		err := cmd.Run(t.Context(), []string{"add", "--name", "MY_TOKEN", "--value", "new", "--repos", "owner/r"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "MY_TOKEN")
	})

	t.Run("overwrites existing secret with --force", func(t *testing.T) {
		cfgPath := testutils.SetupConfig(t, &config.Config{
			Secrets: []config.Secret{{Name: "MY_TOKEN", Value: "old", Repositories: []string{"owner/r"}}},
		})

		cmd := add.New()
		err := cmd.Run(t.Context(), []string{"add", "--name", "MY_TOKEN", "--value", "new", "--repos", "owner/r2", "--force"})
		require.NoError(t, err)

		loaded, err := config.Load(cfgPath)
		require.NoError(t, err)
		require.Len(t, loaded.Secrets, 1)
		assert.Equal(t, "new", loaded.Secrets[0].Value)
		assert.Equal(t, []string{"owner/r2"}, loaded.Secrets[0].Repositories)
	})
}
