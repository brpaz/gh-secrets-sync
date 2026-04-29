package gh_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/gh-secrets-sync/internal/gh"
)

func TestUpsertRepoSecret(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		exec := &mockExecutor{}
		exec.On("Path").Return("/usr/local/bin/gh", nil)
		exec.On("ExecContext", mock.Anything,
			[]string{"secret", "set", "MY_TOKEN", "--repo", "myorg/myrepo", "--body", "supersecret"},
		).Return("", "", nil)

		client, err := gh.NewClient(gh.WithExecutor(exec))
		require.NoError(t, err)

		err = client.UpsertRepoSecret(context.Background(), gh.UpsertSecretRequest{
			Repo: "myorg/myrepo", Name: "MY_TOKEN", Value: "supersecret",
		})
		require.NoError(t, err)
		exec.AssertExpectations(t)
	})

	t.Run("invalid repo format rejected before exec", func(t *testing.T) {
		t.Parallel()
		exec := &mockExecutor{}
		exec.On("Path").Return("/usr/local/bin/gh", nil)

		client, err := gh.NewClient(gh.WithExecutor(exec))
		require.NoError(t, err)

		err = client.UpsertRepoSecret(context.Background(), gh.UpsertSecretRequest{
			Repo: "not-a-valid-repo", Name: "TOKEN", Value: "val",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid repository format")
		exec.AssertNotCalled(t, "ExecContext")
	})

	t.Run("exec error uses stderr", func(t *testing.T) {
		t.Parallel()
		exec := &mockExecutor{}
		exec.On("Path").Return("/usr/local/bin/gh", nil)
		exec.On("ExecContext", mock.Anything,
			[]string{"secret", "set", "MY_TOKEN", "--repo", "myorg/myrepo", "--body", "val"},
		).Return("", "HTTP 403: Resource not accessible by personal access token", errors.New("exit status 1"))

		client, err := gh.NewClient(gh.WithExecutor(exec))
		require.NoError(t, err)

		err = client.UpsertRepoSecret(context.Background(), gh.UpsertSecretRequest{
			Repo: "myorg/myrepo", Name: "MY_TOKEN", Value: "val",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "myorg/myrepo")
		assert.Contains(t, err.Error(), "HTTP 403")
		exec.AssertExpectations(t)
	})

	t.Run("exec error falls back to err message when stderr is empty", func(t *testing.T) {
		t.Parallel()
		exec := &mockExecutor{}
		exec.On("Path").Return("/usr/local/bin/gh", nil)
		exec.On("ExecContext", mock.Anything,
			[]string{"secret", "set", "TOKEN", "--repo", "myorg/myrepo", "--body", "val"},
		).Return("", "", errors.New("connection refused"))

		client, err := gh.NewClient(gh.WithExecutor(exec))
		require.NoError(t, err)

		err = client.UpsertRepoSecret(context.Background(), gh.UpsertSecretRequest{
			Repo: "myorg/myrepo", Name: "TOKEN", Value: "val",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connection refused")
		exec.AssertExpectations(t)
	})
}

func TestCurrentRepository(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		exec := &mockExecutor{}
		exec.On("Path").Return("/usr/local/bin/gh", nil)
		exec.On("ExecContext", mock.Anything,
			[]string{"repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner"},
		).Return("myorg/myrepo\n", "", nil)

		client, err := gh.NewClient(gh.WithExecutor(exec))
		require.NoError(t, err)

		repo, err := client.CurrentRepository(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "myorg/myrepo", repo)
		exec.AssertExpectations(t)
	})

	t.Run("exec error uses stderr", func(t *testing.T) {
		t.Parallel()

		exec := &mockExecutor{}
		exec.On("Path").Return("/usr/local/bin/gh", nil)
		exec.On("ExecContext", mock.Anything,
			[]string{"repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner"},
		).Return("", "not a git repository", errors.New("exit status 1"))

		client, err := gh.NewClient(gh.WithExecutor(exec))
		require.NoError(t, err)

		_, err = client.CurrentRepository(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to determine current repository")
		assert.Contains(t, err.Error(), "not a git repository")
		exec.AssertExpectations(t)
	})

	t.Run("empty output errors", func(t *testing.T) {
		t.Parallel()

		exec := &mockExecutor{}
		exec.On("Path").Return("/usr/local/bin/gh", nil)
		exec.On("ExecContext", mock.Anything,
			[]string{"repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner"},
		).Return("", "", nil)

		client, err := gh.NewClient(gh.WithExecutor(exec))
		require.NoError(t, err)

		_, err = client.CurrentRepository(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty gh output")
		exec.AssertExpectations(t)
	})
}
