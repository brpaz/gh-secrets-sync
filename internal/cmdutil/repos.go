// Package cmdutil provides small helpers shared across CLI command packages.
package cmdutil

import "strings"

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
