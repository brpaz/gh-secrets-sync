// Package testutils provides shared test helpers used across command packages.
package testutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/brpaz/gh-secrets-sync/internal/config"
)

// SetupConfig creates a temporary XDG_CONFIG_HOME directory containing a
// gh-secrets-sync/secrets.yaml file pre-populated with cfg. It registers a
// t.Setenv("XDG_CONFIG_HOME", …) cleanup automatically and returns the path
// to the config file.
func SetupConfig(t *testing.T, cfg *config.Config) string {
	t.Helper()
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "gh-secrets-sync")
	require.NoError(t, os.MkdirAll(cfgDir, 0o700))
	cfgPath := filepath.Join(cfgDir, "secrets.yaml")
	require.NoError(t, cfg.Save(cfgPath))
	t.Setenv("XDG_CONFIG_HOME", dir)
	return cfgPath
}
