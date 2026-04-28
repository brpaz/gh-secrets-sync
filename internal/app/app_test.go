package app_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/gh-secrets-sync/internal/app"
	rootcmd "github.com/brpaz/gh-secrets-sync/internal/commands/root"
)

var mockVersion = app.VersionInfo{
	Version:   "0.1.0",
	Commit:    "abc123",
	BuildDate: "2024-06-01",
}

func TestApp_New(t *testing.T) {
	t.Parallel()

	t.Run("constructs a new app instance with default options", func(t *testing.T) {
		appInstance, err := app.New()

		assert.NoError(t, err)
		assert.NotNil(t, appInstance)
		assert.Equal(t, "0.0.0-dev", appInstance.Info.Version)
		assert.Equal(t, "n/a", appInstance.Info.Commit)
		assert.Equal(t, "n/a", appInstance.Info.BuildDate)
	})

	t.Run("constructs a new app instance with version information", func(t *testing.T) {
		appInstance, err := app.New(app.WithVersionInfo(mockVersion))

		assert.NoError(t, err)
		assert.NotNil(t, appInstance)
		assert.Equal(t, mockVersion, appInstance.Info)
	})
}

func TestApp_Run(t *testing.T) {
	appInstance, err := app.New()
	require.NoError(t, err)

	t.Run("runs successfully", func(t *testing.T) {
		err := appInstance.Run(context.Background(), []string{rootcmd.Name, "--version"})
		assert.NoError(t, err)
	})
}
