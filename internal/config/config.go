package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// AppName is the directory name used under the OS config dir.
	AppName = "gh-secrets-sync"
	// ConfigFileName is the name of the config file.
	ConfigFileName = "secrets.yaml"

	// dirPerm is the permission mode for the config directory.
	dirPerm = 0o700
	// filePerm is the permission mode for the config file.
	filePerm = 0o600
)

// skeletonYAML is written into a newly created config file so users have a
// commented example to start from.
const skeletonYAML = `# gh-secrets-sync configuration
#
# Add your secrets below. Each entry requires:
#   name         – the GitHub Actions secret name (e.g. NPM_TOKEN)
#   value        – the secret value (treated as plaintext; protect this file)
#   repositories – list of "owner/repo" targets to sync the secret to
#
# Example:
#
# secrets:
#   - name: "NPM_TOKEN"
#     value: "npm_xxxxxxxxxxxx"
#     repositories:
#       - "my-org/repo-a"
#       - "my-org/repo-b"

secrets: []
`

// Secret represents a single secret entry in the config file.
type Secret struct {
	Name         string   `yaml:"name"`
	Value        string   `yaml:"value"`
	Repositories []string `yaml:"repositories"`
}

// Config is the top-level structure of secrets.yaml.
type Config struct {
	Secrets []Secret `yaml:"secrets"`
}

// DefaultConfigPath returns the OS-appropriate path for the secrets.yaml
// config file, using os.UserConfigDir() for cross-platform compatibility.
//
// Linux/macOS: ~/.config/gh-secrets-sync/secrets.yaml
// Windows:     %APPDATA%\gh-secrets-sync\secrets.yaml
func DefaultConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not determine user config directory: %w", err)
	}

	return filepath.Join(dir, AppName, ConfigFileName), nil
}

// EnsureConfigExists checks whether the config file at path exists and creates
// it (along with any missing parent directories) if it does not.
//
// The directory is created with 0700 permissions and the file with 0600
// permissions. It returns true when the file was newly created and false when
// it already existed.
func EnsureConfigExists(path string) (created bool, err error) {
	// Check if the file already exists.
	if _, statErr := os.Stat(path); statErr == nil {
		return false, nil
	} else if !os.IsNotExist(statErr) {
		return false, fmt.Errorf("could not stat config file %s: %w", path, statErr)
	}

	// Create the directory with restricted permissions.
	dir := filepath.Dir(path)
	if mkdirErr := os.MkdirAll(dir, dirPerm); mkdirErr != nil {
		return false, fmt.Errorf("could not create config directory %s: %w", dir, mkdirErr)
	}

	// Write the skeleton config file with restricted permissions.
	if writeErr := os.WriteFile(path, []byte(skeletonYAML), filePerm); writeErr != nil {
		return false, fmt.Errorf("could not create config file %s: %w", path, writeErr)
	}

	return true, nil
}

// AddSecret adds s to cfg. It returns an error if a secret with the same name
// already exists and force is false. When force is true the existing entry is
// overwritten in place (preserving its position in the slice).
func (cfg *Config) AddSecret(s Secret, force bool) error {
	for i, existing := range cfg.Secrets {
		if existing.Name == s.Name {
			if !force {
				return fmt.Errorf("secret %q already exists – use --force to overwrite", s.Name)
			}
			cfg.Secrets[i] = s
			return nil
		}
	}
	cfg.Secrets = append(cfg.Secrets, s)
	return nil
}

// UpdateSecret updates an existing secret by name. Only non-zero fields in
// patch are applied: if patch.Value is non-empty it replaces the current value;
// if patch.Repositories is non-nil/non-empty it replaces the current repo list.
// Returns an error if no secret with the given name exists.
func (cfg *Config) UpdateSecret(name string, patch Secret) error {
	for i, existing := range cfg.Secrets {
		if existing.Name == name {
			if patch.Value != "" {
				cfg.Secrets[i].Value = patch.Value
			}
			if len(patch.Repositories) > 0 {
				cfg.Secrets[i].Repositories = patch.Repositories
			}
			return nil
		}
	}
	return fmt.Errorf("secret %q not found", name)
}

// DeleteSecret removes the secret with the given name from cfg.
// It returns an error if no secret with that name exists.
func (cfg *Config) DeleteSecret(name string) error {
	for i, s := range cfg.Secrets {
		if s.Name == name {
			cfg.Secrets = append(cfg.Secrets[:i], cfg.Secrets[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("secret %q not found", name)
}

// Save marshals cfg to YAML and writes it to path with 0600 permissions.
func (cfg *Config) Save(path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, filePerm); err != nil {
		return fmt.Errorf("could not write config file %s: %w", path, err)
	}
	return nil
}

// Load reads and parses the YAML config file at path.
// It returns a descriptive error (including the file path) when the file cannot
// be read or contains invalid YAML.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid YAML in config file %s: %w", path, err)
	}

	return &cfg, nil
}
