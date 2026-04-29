package gh

import (
	"context"
	"fmt"
	"strings"
)

// CurrentRepository returns the current GitHub repository in owner/repo format.
func (c *Client) CurrentRepository(ctx context.Context) (string, error) {
	stdout, stderr, err := c.ExecContext(ctx, "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner")
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}

		return "", fmt.Errorf("failed to determine current repository: %s", msg)
	}

	repo := strings.TrimSpace(stdout.String())
	if repo == "" {
		return "", fmt.Errorf("failed to determine current repository: empty gh output")
	}

	return repo, nil
}
