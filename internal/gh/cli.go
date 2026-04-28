package gh

import (
	"bytes"
	"context"
	"fmt"

	ghCLI "github.com/cli/go-gh/v2"
)

// Executor abstracts the gh CLI subprocess calls so the Client can be unit-tested.
type Executor interface {
	// Path returns the path to the gh binary, or an error if not found.
	Path() (string, error)
	// ExecContext invokes a gh subcommand and returns its stdout and stderr.
	ExecContext(ctx context.Context, args ...string) (stdout, stderr bytes.Buffer, err error)
}

// defaultExecutor delegates to the real go-gh library.
type defaultExecutor struct{}

func (defaultExecutor) Path() (string, error) {
	return ghCLI.Path()
}

func (defaultExecutor) ExecContext(ctx context.Context, args ...string) (stdout, stderr bytes.Buffer, err error) {
	return ghCLI.ExecContext(ctx, args...)
}

// Client executes secret operations via the gh CLI.
// Executor is embedded so methods like ExecContext are promoted directly onto Client.
type Client struct {
	Executor
}

// Option configures a Client.
type Option func(*Client)

// WithExecutor replaces the default gh CLI executor. Primarily for testing.
func WithExecutor(exec Executor) Option {
	return func(c *Client) {
		c.Executor = exec
	}
}

// NewClient returns a Client, verifying that the gh CLI is available.
// Pass functional options to override defaults (e.g. WithExecutor for tests).
func NewClient(opts ...Option) (*Client, error) {
	c := &Client{Executor: defaultExecutor{}}
	for _, o := range opts {
		o(c)
	}
	if _, err := c.Path(); err != nil {
		return nil, fmt.Errorf("gh CLI not found: %w", err)
	}
	return c, nil
}
