package edit_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	editcmd "github.com/brpaz/gh-secrets-sync/internal/commands/edit"
	"github.com/brpaz/gh-secrets-sync/internal/config"
	"github.com/brpaz/gh-secrets-sync/internal/testutils"
)

func TestNew(t *testing.T) {
	cmd := editcmd.New()
	assert.IsType(t, cmd, &cli.Command{})
}

func TestEditCommand(t *testing.T) {
	t.Run("updates value", func(t *testing.T) {
		cfgPath := testutils.SetupConfig(t, &config.Config{
			Secrets: []config.Secret{
				{Name: "MY_TOKEN", Value: "old", Repositories: []string{"owner/repo"}},
			},
		})

		var out strings.Builder
		cmd := editcmd.New()
		cmd.Writer = &out

		err := cmd.Run(t.Context(), []string{"edit", "--name", "MY_TOKEN", "--value", "newval", "--repos", "owner/repo"})
		require.NoError(t, err)
		assert.Contains(t, out.String(), "MY_TOKEN")

		loaded, err := config.Load(cfgPath)
		require.NoError(t, err)
		require.Len(t, loaded.Secrets, 1)
		assert.Equal(t, "newval", loaded.Secrets[0].Value)
		assert.Equal(t, []string{"owner/repo"}, loaded.Secrets[0].Repositories)
	})

	t.Run("updates repos while keeping existing value", func(t *testing.T) {
		cfgPath := testutils.SetupConfig(t, &config.Config{
			Secrets: []config.Secret{
				{Name: "MY_TOKEN", Value: "secret", Repositories: []string{"owner/repo1"}},
			},
		})

		cmd := editcmd.New()
		// Pass --value with the current value so the survey is not triggered.
		err := cmd.Run(t.Context(), []string{"edit", "--name", "MY_TOKEN", "--value", "secret", "--repos", "owner/repo2,owner/repo3"})
		require.NoError(t, err)

		loaded, err := config.Load(cfgPath)
		require.NoError(t, err)
		assert.Equal(t, "secret", loaded.Secrets[0].Value)
		assert.Equal(t, []string{"owner/repo2", "owner/repo3"}, loaded.Secrets[0].Repositories)
	})

	t.Run("allows clearing repos with explicit empty flag", func(t *testing.T) {
		cfgPath := testutils.SetupConfig(t, &config.Config{
			Secrets: []config.Secret{
				{Name: "MY_TOKEN", Value: "secret", Repositories: []string{"owner/repo1"}},
			},
		})

		cmd := editcmd.New()
		err := cmd.Run(t.Context(), []string{"edit", "--name", "MY_TOKEN", "--value", "secret", "--repos", ""})
		require.NoError(t, err)

		loaded, err := config.Load(cfgPath)
		require.NoError(t, err)
		assert.Equal(t, "secret", loaded.Secrets[0].Value)
		assert.Empty(t, loaded.Secrets[0].Repositories)
	})

	t.Run("errors when secret not found", func(t *testing.T) {
		testutils.SetupConfig(t, &config.Config{
			Secrets: []config.Secret{
				{Name: "OTHER", Value: "val", Repositories: []string{"owner/repo"}},
			},
		})

		cmd := editcmd.New()
		err := cmd.Run(t.Context(), []string{"edit", "--name", "MISSING", "--value", "x", "--repos", "owner/repo"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "MISSING")
	})
}
