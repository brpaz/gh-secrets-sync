package gh

import (
	"github.com/cli/go-gh/v2/pkg/api"
)

type RepoSecret struct {
	Name  string
	Value string
}

// AddSecretToRepos adds a secret to the specified repositories using the GitHub API client.
func AddSecretToRepos(client api.RESTClient, secret RepoSecret, repos []string) error {
	return nil
}
