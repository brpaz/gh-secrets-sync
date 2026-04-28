package gh_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/gh-secrets-sync/internal/gh"
)

// mockExecutor is a testify/mock implementation of gh.Executor.
type mockExecutor struct {
	mock.Mock
}

func (m *mockExecutor) Path() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *mockExecutor) ExecContext(ctx context.Context, execArgs ...string) (stdout, stderr bytes.Buffer, err error) {
	called := m.Called(ctx, execArgs)
	stdout.WriteString(called.String(0))
	stderr.WriteString(called.String(1))
	err = called.Error(2)
	return
}

func TestNewClient(t *testing.T) {
	t.Run("default executor requires real gh binary", func(t *testing.T) {
		t.Parallel()
		client, err := gh.NewClient()
		if err != nil {
			t.Skipf("gh CLI not available: %v", err)
		}
		assert.NotNil(t, client)
	})

	t.Run("gh found", func(t *testing.T) {
		t.Parallel()
		exec := &mockExecutor{}
		exec.On("Path").Return("/usr/local/bin/gh", nil)

		client, err := gh.NewClient(gh.WithExecutor(exec))
		require.NoError(t, err)
		assert.NotNil(t, client)
		exec.AssertExpectations(t)
	})

	t.Run("gh not found", func(t *testing.T) {
		t.Parallel()
		exec := &mockExecutor{}
		exec.On("Path").Return("", errors.New("not in PATH"))

		_, err := gh.NewClient(gh.WithExecutor(exec))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "gh CLI not found")
		exec.AssertExpectations(t)
	})
}
