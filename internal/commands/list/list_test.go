package list_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	listcmd "github.com/brpaz/gh-secrets-sync/internal/commands/list"
	"github.com/brpaz/gh-secrets-sync/internal/config"
	"github.com/brpaz/gh-secrets-sync/internal/testutils"
)

func TestNew(t *testing.T) {
	cmd := listcmd.New()
	assert.IsType(t, cmd, &cli.Command{})
}

func TestListCommand(t *testing.T) {
	t.Run("prints empty state message when no secrets configured", func(t *testing.T) {
		testutils.SetupConfig(t, &config.Config{})

		var out strings.Builder
		cmd := listcmd.New()
		cmd.Writer = &out

		err := cmd.Run(t.Context(), []string{"list"})
		require.NoError(t, err)
		assert.Contains(t, out.String(), "No secrets configured")
		assert.Contains(t, out.String(), "gh secrets-sync add")
	})

	t.Run("masks secret values by default", func(t *testing.T) {
		testutils.SetupConfig(t, &config.Config{
			Secrets: []config.Secret{
				{Name: "MY_TOKEN", Value: "supersecret", Repositories: []string{"owner/repo1"}},
			},
		})

		var out strings.Builder
		cmd := listcmd.New()
		cmd.Writer = &out

		err := cmd.Run(t.Context(), []string{"list"})
		require.NoError(t, err)

		output := out.String()
		assert.Contains(t, output, "MY_TOKEN")
		assert.Contains(t, output, "****")
		assert.NotContains(t, output, "supersecret")
		assert.Contains(t, output, "owner/repo1")
	})

	t.Run("reveals values with --reveal --yes", func(t *testing.T) {
		testutils.SetupConfig(t, &config.Config{
			Secrets: []config.Secret{
				{Name: "MY_TOKEN", Value: "supersecret", Repositories: []string{"owner/repo1"}},
			},
		})

		var out strings.Builder
		cmd := listcmd.New()
		cmd.Writer = &out

		err := cmd.Run(t.Context(), []string{"list", "--reveal", "--yes"})
		require.NoError(t, err)

		output := out.String()
		assert.Contains(t, output, "MY_TOKEN")
		assert.Contains(t, output, "supersecret")
		assert.NotContains(t, output, "****")
	})

	t.Run("shows dash for secrets with no repositories", func(t *testing.T) {
		testutils.SetupConfig(t, &config.Config{
			Secrets: []config.Secret{
				{Name: "ORPHAN", Value: "val", Repositories: []string{}},
			},
		})

		var out strings.Builder
		cmd := listcmd.New()
		cmd.Writer = &out

		err := cmd.Run(t.Context(), []string{"list"})
		require.NoError(t, err)

		output := out.String()
		assert.Contains(t, output, "ORPHAN")
		assert.Contains(t, output, "—")
	})

	t.Run("lists all secrets with their repositories", func(t *testing.T) {
		testutils.SetupConfig(t, &config.Config{
			Secrets: []config.Secret{
				{Name: "TOKEN_A", Value: "aaa", Repositories: []string{"owner/repo1", "owner/repo2"}},
				{Name: "TOKEN_B", Value: "bbb", Repositories: []string{"owner/repo3"}},
			},
		})

		var out strings.Builder
		cmd := listcmd.New()
		cmd.Writer = &out

		err := cmd.Run(t.Context(), []string{"list"})
		require.NoError(t, err)

		output := out.String()
		assert.Contains(t, output, "TOKEN_A")
		assert.Contains(t, output, "TOKEN_B")
		assert.Contains(t, output, "owner/repo1")
		assert.Contains(t, output, "owner/repo2")
		assert.Contains(t, output, "owner/repo3")
	})
}
