package sync_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	synccmd "github.com/brpaz/gh-secrets-sync/internal/commands/sync"
	"github.com/brpaz/gh-secrets-sync/internal/config"
	"github.com/brpaz/gh-secrets-sync/internal/gh"
	"github.com/brpaz/gh-secrets-sync/internal/testutils"
)

// MockGitHubClient is a testify/mock implementation of synccmd.GitHubClient.
type MockGitHubClient struct {
	mock.Mock
}

func (m *MockGitHubClient) UpsertRepoSecret(ctx context.Context, req gh.UpsertSecretRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func TestNew(t *testing.T) {
	cmd := synccmd.New(&MockGitHubClient{})
	assert.IsType(t, cmd, &cli.Command{})
}

func TestSyncCommand(t *testing.T) {
	t.Run("dry-run prints planned operations without calling syncer", func(t *testing.T) {
		testutils.SetupConfig(t, &config.Config{
			Secrets: []config.Secret{
				{Name: "TOKEN_A", Value: "aaa", Repositories: []string{"owner/repo1", "owner/repo2"}},
			},
		})

		var out strings.Builder
		cmd := synccmd.New(nil) // nil client – dry-run must never call it
		cmd.Writer = &out

		err := cmd.Run(t.Context(), []string{"sync", "--dry-run"})
		require.NoError(t, err)

		output := out.String()
		assert.Contains(t, output, "[DRY RUN]")
		assert.Contains(t, output, "TOKEN_A")
		assert.Contains(t, output, "owner/repo1")
		assert.Contains(t, output, "owner/repo2")
		assert.Contains(t, output, "2 synced, 0 failed")
	})

	t.Run("errors when config has no secrets", func(t *testing.T) {
		testutils.SetupConfig(t, &config.Config{})

		cmd := synccmd.New(nil)
		err := cmd.Run(t.Context(), []string{"sync", "--dry-run"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no secrets configured")
	})

	t.Run("errors when requested secret does not exist", func(t *testing.T) {
		testutils.SetupConfig(t, &config.Config{
			Secrets: []config.Secret{
				{Name: "TOKEN_A", Value: "aaa", Repositories: []string{"owner/repo1"}},
			},
		})

		cmd := synccmd.New(nil)
		err := cmd.Run(t.Context(), []string{"sync", "--dry-run", "--secret", "MISSING"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "MISSING")
	})

	t.Run("dry-run syncs only the named secret", func(t *testing.T) {
		testutils.SetupConfig(t, &config.Config{
			Secrets: []config.Secret{
				{Name: "TOKEN_A", Value: "aaa", Repositories: []string{"owner/repo1"}},
				{Name: "TOKEN_B", Value: "bbb", Repositories: []string{"owner/repo2"}},
			},
		})

		var out strings.Builder
		cmd := synccmd.New(nil)
		cmd.Writer = &out

		err := cmd.Run(t.Context(), []string{"sync", "--dry-run", "--secret", "TOKEN_A"})
		require.NoError(t, err)

		output := out.String()
		assert.Contains(t, output, "TOKEN_A")
		assert.NotContains(t, output, "TOKEN_B")
		assert.Contains(t, output, "1 synced, 0 failed")
	})

	t.Run("dry-run skips secrets with no repositories", func(t *testing.T) {
		testutils.SetupConfig(t, &config.Config{
			Secrets: []config.Secret{
				{Name: "ORPHAN", Value: "val", Repositories: []string{}},
				{Name: "TOKEN_A", Value: "aaa", Repositories: []string{"owner/repo1"}},
			},
		})

		var out strings.Builder
		cmd := synccmd.New(nil)
		cmd.Writer = &out

		err := cmd.Run(t.Context(), []string{"sync", "--dry-run"})
		require.NoError(t, err)

		output := out.String()
		assert.Contains(t, output, "ORPHAN")
		assert.Contains(t, output, "skipped")
		assert.Contains(t, output, "1 synced, 0 failed")
	})

	t.Run("pushes all secrets via client", func(t *testing.T) {
		testutils.SetupConfig(t, &config.Config{
			Secrets: []config.Secret{
				{Name: "TOKEN_A", Value: "aaa", Repositories: []string{"owner/repo1", "owner/repo2"}},
			},
		})

		client := &MockGitHubClient{}
		client.On("UpsertRepoSecret", mock.Anything, gh.UpsertSecretRequest{
			Repo: "owner/repo1", Name: "TOKEN_A", Value: "aaa",
		}).Return(nil)
		client.On("UpsertRepoSecret", mock.Anything, gh.UpsertSecretRequest{
			Repo: "owner/repo2", Name: "TOKEN_A", Value: "aaa",
		}).Return(nil)

		var out strings.Builder
		cmd := synccmd.New(client)
		cmd.Writer = &out

		err := cmd.Run(t.Context(), []string{"sync"})
		require.NoError(t, err)

		client.AssertExpectations(t)
		assert.Contains(t, out.String(), "2 synced, 0 failed")
	})

	t.Run("reports partial failure and returns error", func(t *testing.T) {
		testutils.SetupConfig(t, &config.Config{
			Secrets: []config.Secret{
				{Name: "TOKEN_A", Value: "aaa", Repositories: []string{"owner/repo1", "owner/repo2"}},
			},
		})

		client := &MockGitHubClient{}
		client.On("UpsertRepoSecret", mock.Anything, gh.UpsertSecretRequest{
			Repo: "owner/repo1", Name: "TOKEN_A", Value: "aaa",
		}).Return(nil)
		client.On("UpsertRepoSecret", mock.Anything, gh.UpsertSecretRequest{
			Repo: "owner/repo2", Name: "TOKEN_A", Value: "aaa",
		}).Return(fmt.Errorf("403 Forbidden"))

		var out strings.Builder
		cmd := synccmd.New(client)
		cmd.Writer = &out

		err := cmd.Run(t.Context(), []string{"sync"})
		require.Error(t, err)

		client.AssertExpectations(t)
		output := out.String()
		assert.Contains(t, output, "✓ TOKEN_A → owner/repo1")
		assert.Contains(t, output, "✗ TOKEN_A → owner/repo2")
		assert.Contains(t, output, "1 synced, 1 failed")
	})
}
