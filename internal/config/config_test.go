package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/gh-secrets-sync/internal/config"
)

func TestDefaultConfigPath(t *testing.T) {
	t.Parallel()

	t.Run("returns path ending with app name and config file name", func(t *testing.T) {
		t.Parallel()

		path, err := config.DefaultConfigPath()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(config.AppName, config.ConfigFileName), filepath.Base(filepath.Dir(path))+string(filepath.Separator)+filepath.Base(path))
	})
}

func TestEnsureConfigExists(t *testing.T) {
	t.Parallel()

	t.Run("creates directory and file on first call", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "subdir", config.ConfigFileName)

		created, err := config.EnsureConfigExists(path)
		require.NoError(t, err)
		assert.True(t, created)

		info, err := os.Stat(path)
		require.NoError(t, err, "config file should exist")
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "file perm should be 0600")

		dirInfo, err := os.Stat(filepath.Dir(path))
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm(), "dir perm should be 0700")
	})

	t.Run("file contains commented skeleton YAML", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, config.ConfigFileName)

		_, err := config.EnsureConfigExists(path)
		require.NoError(t, err)

		data, err := os.ReadFile(path)
		require.NoError(t, err)

		content := string(data)
		for _, want := range []string{"secrets:", "name", "value", "repositories"} {
			assert.Contains(t, content, want, "skeleton YAML should contain %q", want)
		}
	})

	t.Run("does not overwrite existing file", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, config.ConfigFileName)

		_, err := config.EnsureConfigExists(path)
		require.NoError(t, err)

		customContent := "secrets:\n  - name: FOO\n    value: bar\n    repositories: []\n"
		require.NoError(t, os.WriteFile(path, []byte(customContent), 0o600))

		created, err := config.EnsureConfigExists(path)
		require.NoError(t, err)
		assert.False(t, created, "created should be false on second call")

		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, customContent, string(data), "existing file should not be modified")
	})
}

func TestLoad(t *testing.T) {
	t.Parallel()

	t.Run("parses valid YAML with secrets", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, config.ConfigFileName)
		content := `secrets:
  - name: "NPM_TOKEN"
    value: "npm_abc123"
    repositories:
      - "owner/repo1"
      - "owner/repo2"
`
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

		cfg, err := config.Load(path)
		require.NoError(t, err)
		require.Len(t, cfg.Secrets, 1)

		s := cfg.Secrets[0]
		assert.Equal(t, "NPM_TOKEN", s.Name)
		assert.Equal(t, "npm_abc123", s.Value)
		assert.Equal(t, []string{"owner/repo1", "owner/repo2"}, s.Repositories)
	})

	t.Run("parses empty secrets list", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, config.ConfigFileName)
		require.NoError(t, os.WriteFile(path, []byte("secrets: []\n"), 0o600))

		cfg, err := config.Load(path)
		require.NoError(t, err)
		assert.Empty(t, cfg.Secrets)
	})

	t.Run("returns error with file path for invalid YAML", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, config.ConfigFileName)
		require.NoError(t, os.WriteFile(path, []byte(":: invalid yaml ::\n"), 0o600))

		_, err := config.Load(path)
		require.Error(t, err)
		assert.Contains(t, err.Error(), path)
	})

	t.Run("returns error with file path for missing file", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "nonexistent.yaml")

		_, err := config.Load(path)
		require.Error(t, err)
		assert.Contains(t, err.Error(), path)
	})
}

func TestAddSecret(t *testing.T) {
	t.Parallel()

	t.Run("adds new secret", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{}
		s := config.Secret{Name: "FOO", Value: "bar", Repositories: []string{"owner/repo1"}}

		err := cfg.AddSecret(s, false)
		require.NoError(t, err)
		require.Len(t, cfg.Secrets, 1)
		assert.Equal(t, s, cfg.Secrets[0])
	})

	t.Run("errors on duplicate name without force", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			Secrets: []config.Secret{{Name: "FOO", Value: "old", Repositories: []string{"owner/repo1"}}},
		}

		err := cfg.AddSecret(config.Secret{Name: "FOO", Value: "new"}, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "FOO")
		assert.Contains(t, err.Error(), "--force")
		assert.Equal(t, "old", cfg.Secrets[0].Value, "original value must not change")
	})

	t.Run("overwrites in place with force", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			Secrets: []config.Secret{{Name: "FOO", Value: "old", Repositories: []string{"owner/repo1"}}},
		}
		updated := config.Secret{Name: "FOO", Value: "new", Repositories: []string{"owner/repo2"}}

		err := cfg.AddSecret(updated, true)
		require.NoError(t, err)
		require.Len(t, cfg.Secrets, 1, "no new entry should be appended")
		assert.Equal(t, updated, cfg.Secrets[0])
	})
}

func TestSave(t *testing.T) {
	t.Parallel()

	t.Run("writes file with 0600 permissions", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "secrets.yaml")
		cfg := &config.Config{
			Secrets: []config.Secret{
				{Name: "TOKEN", Value: "abc123", Repositories: []string{"owner/repo"}},
			},
		}

		require.NoError(t, cfg.Save(path))

		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})

	t.Run("saved data round-trips through Load", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "secrets.yaml")
		cfg := &config.Config{
			Secrets: []config.Secret{
				{Name: "TOKEN", Value: "abc123", Repositories: []string{"owner/repo"}},
			},
		}

		require.NoError(t, cfg.Save(path))

		loaded, err := config.Load(path)
		require.NoError(t, err)
		require.Len(t, loaded.Secrets, 1)
		assert.Equal(t, cfg.Secrets[0], loaded.Secrets[0])
	})
}

func TestDeleteSecret(t *testing.T) {
	t.Parallel()

	t.Run("removes the named secret", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			Secrets: []config.Secret{
				{Name: "FOO", Value: "bar", Repositories: []string{"owner/repo"}},
				{Name: "BAZ", Value: "qux", Repositories: []string{"owner/repo2"}},
			},
		}

		err := cfg.DeleteSecret("FOO")
		require.NoError(t, err)
		require.Len(t, cfg.Secrets, 1)
		assert.Equal(t, "BAZ", cfg.Secrets[0].Name)
	})

	t.Run("errors when secret not found", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{}
		err := cfg.DeleteSecret("MISSING")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "MISSING")
	})

	t.Run("preserves order of remaining secrets", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			Secrets: []config.Secret{
				{Name: "A"}, {Name: "B"}, {Name: "C"},
			},
		}

		require.NoError(t, cfg.DeleteSecret("B"))
		require.Len(t, cfg.Secrets, 2)
		assert.Equal(t, "A", cfg.Secrets[0].Name)
		assert.Equal(t, "C", cfg.Secrets[1].Name)
	})
}

func TestUpdateSecret(t *testing.T) {
	t.Parallel()

	t.Run("updates value only", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			Secrets: []config.Secret{{Name: "FOO", Value: "old", Repositories: []string{"owner/repo1"}}},
		}

		err := cfg.UpdateSecret("FOO", config.Secret{Value: "new"})
		require.NoError(t, err)
		assert.Equal(t, "new", cfg.Secrets[0].Value)
		assert.Equal(t, []string{"owner/repo1"}, cfg.Secrets[0].Repositories, "repos must be unchanged")
	})

	t.Run("updates repos only", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			Secrets: []config.Secret{{Name: "FOO", Value: "old", Repositories: []string{"owner/repo1"}}},
		}

		err := cfg.UpdateSecret("FOO", config.Secret{Repositories: []string{"owner/repo2"}})
		require.NoError(t, err)
		assert.Equal(t, "old", cfg.Secrets[0].Value, "value must be unchanged")
		assert.Equal(t, []string{"owner/repo2"}, cfg.Secrets[0].Repositories)
	})

	t.Run("allows clearing repos to empty", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			Secrets: []config.Secret{{Name: "FOO", Value: "old", Repositories: []string{"owner/repo1"}}},
		}

		err := cfg.UpdateSecret("FOO", config.Secret{Repositories: []string{}})
		require.NoError(t, err)
		assert.Equal(t, "old", cfg.Secrets[0].Value, "value must be unchanged")
		assert.Empty(t, cfg.Secrets[0].Repositories)
	})

	t.Run("updates both value and repos", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			Secrets: []config.Secret{{Name: "FOO", Value: "old", Repositories: []string{"owner/repo1"}}},
		}

		err := cfg.UpdateSecret("FOO", config.Secret{Value: "new", Repositories: []string{"owner/repo2"}})
		require.NoError(t, err)
		assert.Equal(t, "new", cfg.Secrets[0].Value)
		assert.Equal(t, []string{"owner/repo2"}, cfg.Secrets[0].Repositories)
	})

	t.Run("errors when secret not found", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{}

		err := cfg.UpdateSecret("NONEXISTENT", config.Secret{Value: "x"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "NONEXISTENT")
	})
}
