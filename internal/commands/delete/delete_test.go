package delete_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	deletecmd "github.com/brpaz/gh-secrets-sync/internal/commands/delete"
	"github.com/brpaz/gh-secrets-sync/internal/config"
	"github.com/brpaz/gh-secrets-sync/internal/testutils"
)

func TestNew(t *testing.T) {
	cmd := deletecmd.New()
	assert.IsType(t, cmd, &cli.Command{})
}

func TestDeleteCommand(t *testing.T) {
	t.Run("deletes secret and skips confirmation with --yes", func(t *testing.T) {
		cfgPath := testutils.SetupConfig(t, &config.Config{
			Secrets: []config.Secret{
				{Name: "MY_TOKEN", Value: "abc", Repositories: []string{"owner/repo"}},
			},
		})

		var out strings.Builder
		cmd := deletecmd.New()
		cmd.Writer = &out

		err := cmd.Run(t.Context(), []string{"delete", "--name", "MY_TOKEN", "--yes"})
		require.NoError(t, err)

		assert.Contains(t, out.String(), "MY_TOKEN")
		assert.Contains(t, out.String(), "NOT removed from any GitHub repositories")

		loaded, err := config.Load(cfgPath)
		require.NoError(t, err)
		assert.Empty(t, loaded.Secrets)
	})

	t.Run("errors when secret not found", func(t *testing.T) {
		testutils.SetupConfig(t, &config.Config{})

		cmd := deletecmd.New()
		err := cmd.Run(t.Context(), []string{"delete", "--name", "MISSING", "--yes"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "MISSING")
	})
}
