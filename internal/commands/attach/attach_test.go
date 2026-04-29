package attach

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/urfave/cli/v3"

	"github.com/brpaz/gh-secrets-sync/internal/gh"
)

type mockGitHubClient struct {
	mock.Mock
}

func (m *mockGitHubClient) CurrentRepository(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *mockGitHubClient) UpsertRepoSecret(ctx context.Context, req gh.UpsertSecretRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func TestNew(t *testing.T) {
	cmd := New(&mockGitHubClient{})
	assert.IsType(t, cmd, &cli.Command{})
}
