// Package cmdutil provides small helpers shared across CLI command packages.
package cmdutil

import (
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/brpaz/gh-secrets-sync/internal/config"
)

// SplitRepos flattens a slice of possibly comma-separated repository strings
// into individual trimmed, non-empty values. It handles both repeated flag
// usage ([]string{"owner/repo1", "owner/repo2"}) and a single comma-separated
// string ([]string{"owner/repo1,owner/repo2"}).
func SplitRepos(raw []string) []string {
	var repos []string
	for _, entry := range raw {
		for r := range strings.SplitSeq(entry, ",") {
			r = strings.TrimSpace(r)
			if r != "" {
				repos = append(repos, r)
			}
		}
	}
	return repos
}

// ConfigPath returns the config path from the CLI command's --config flag,
// or falls back to config.DefaultConfigPath.
func ConfigPath(cmd *cli.Command) (string, error) {
	if path := cmd.String("config"); path != "" {
		return path, nil
	}
	return config.DefaultConfigPath()
}
