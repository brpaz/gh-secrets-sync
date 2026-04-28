package gh

import (
	"context"
	"fmt"
	"strings"
)

// UpsertSecretRequest holds the data needed to upsert one secret to one repo.
// Repo must be in "owner/repo" format.
type UpsertSecretRequest struct {
	Repo  string
	Name  string
	Value string
}

// Validate checks that the request has all required fields.
func (r UpsertSecretRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("secret name is required")
	}
	if r.Repo == "" {
		return fmt.Errorf("secret repo is required")
	}
	if r.Value == "" {
		return fmt.Errorf("secret value is required")
	}

	parts := strings.SplitN(r.Repo, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid repository format %q: expected owner/repo", r.Repo)
	}

	return nil
}

// UpsertRepoSecret sets a secret on a repository by delegating to the gh CLI.
func (c *Client) UpsertRepoSecret(ctx context.Context, req UpsertSecretRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid request received: %w", err)
	}

	_, stderr, err := c.ExecContext(ctx, "secret", "set", req.Name, "--repo", req.Repo, "--body", req.Value)
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("%s: %s", req.Repo, msg)
	}
	return nil
}
