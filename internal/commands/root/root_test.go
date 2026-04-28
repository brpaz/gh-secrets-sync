package root_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"github.com/brpaz/gh-secrets-sync/internal/commands/root"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("returns a cli.Command", func(t *testing.T) {
		t.Parallel()
		cmd := root.New()
		require.NotNil(t, cmd)
		assert.IsType(t, &cli.Command{}, cmd)
	})

	t.Run("defaults", func(t *testing.T) {
		t.Parallel()
		cmd := root.New()
		assert.Equal(t, root.Name, cmd.Name)
		assert.Equal(t, "0.0.0-dev", cmd.Version)
		assert.True(t, cmd.EnableShellCompletion)
		assert.Nil(t, cmd.Before)
		assert.Empty(t, cmd.Commands)
	})

	t.Run("WithVersion sets the version", func(t *testing.T) {
		t.Parallel()
		cmd := root.New(root.WithVersion("1.2.3"))
		assert.Equal(t, "1.2.3", cmd.Version)
	})

	t.Run("WithCommand registers sub-commands in order", func(t *testing.T) {
		t.Parallel()
		sub1 := &cli.Command{Name: "sub1"}
		sub2 := &cli.Command{Name: "sub2"}
		cmd := root.New(root.WithCommand(sub1), root.WithCommand(sub2))
		require.Len(t, cmd.Commands, 2)
		assert.Equal(t, "sub1", cmd.Commands[0].Name)
		assert.Equal(t, "sub2", cmd.Commands[1].Name)
	})

	t.Run("WithOnInit sets Before hook and it is called", func(t *testing.T) {
		t.Parallel()
		called := false
		fn := func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			called = true
			return ctx, nil
		}
		cmd := root.New(root.WithOnInit(fn))
		require.NotNil(t, cmd.Before)
		_, err := cmd.Before(t.Context(), cmd)
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("all options applied together", func(t *testing.T) {
		t.Parallel()
		initCalled := false
		sub := &cli.Command{Name: "mysub"}
		cmd := root.New(
			root.WithVersion("2.0.0"),
			root.WithCommand(sub),
			root.WithOnInit(func(ctx context.Context, c *cli.Command) (context.Context, error) {
				initCalled = true
				return ctx, nil
			}),
		)
		assert.Equal(t, "2.0.0", cmd.Version)
		require.Len(t, cmd.Commands, 1)
		assert.Equal(t, "mysub", cmd.Commands[0].Name)
		require.NotNil(t, cmd.Before)
		_, err := cmd.Before(t.Context(), cmd)
		require.NoError(t, err)
		assert.True(t, initCalled)
	})
}
