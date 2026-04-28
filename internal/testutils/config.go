// Package testutils provides shared test helpers used across command packages.
package testutils

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/brpaz/gh-secrets-sync/internal/config"
)

// SetupConfig creates a temporary config file pre-populated with cfg.
// It sets GH_SECRETS_SYNC_CONFIG_FILE and registers a t.Cleanup to unset it.
// Returns the path to the config file.
func SetupConfig(t *testing.T, cfg *config.Config) string {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "secrets.yaml")
	require.NoError(t, cfg.Save(cfgPath))
	t.Setenv(config.EnvConfigFile, cfgPath)
	return cfgPath
}
