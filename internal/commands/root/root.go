package root

import (
	"github.com/urfave/cli/v3"
)

const (
	Name      = "gh-secrets-sync"
	usage     = "Github CLI extension that syncs GitHub secrets across different repositories"
	usageText = "gh secrets-sync <command> [options]\n\nManage a local secrets config file and push those secrets to one or more\nGitHub repositories via the GitHub API.\n\nOn first run a config file is created automatically. Edit it directly with\n'gh secrets-sync config', or use the add / attach / edit / delete sub-commands.\nOnce your secrets are configured, run 'gh secrets-sync sync' to push them."

	// FlagConfig is the name of the global --config flag.
	FlagConfig = "config"
)

// options holds the configuration for the root command.
type options struct {
	version  string
	commands []*cli.Command
	onInitFn cli.BeforeFunc
}

// Option is a functional option for configuring the root command.
type Option func(*options)

// WithVersion sets the version string shown by --version.
func WithVersion(v string) Option {
	return func(o *options) { o.version = v }
}

// WithCommand appends a single sub-command to the root command.
// Call it multiple times to register multiple sub-commands.
func WithCommand(cmd *cli.Command) Option {
	return func(o *options) { o.commands = append(o.commands, cmd) }
}

// WithOnInit sets the Before hook called before every command invocation.
func WithOnInit(fn cli.BeforeFunc) Option {
	return func(o *options) { o.onInitFn = fn }
}

// New returns the root *cli.Command with the supplied options applied.
func New(opts ...Option) *cli.Command {
	o := &options{
		version: "0.0.0-dev",
	}
	for _, opt := range opts {
		opt(o)
	}

	return &cli.Command{
		Name:                  Name,
		Version:               o.version,
		Usage:                 usage,
		UsageText:             usageText,
		EnableShellCompletion: true,
		Before:                o.onInitFn,
		Commands:              o.commands,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     FlagConfig,
				Usage:    "Path to the config file (default: ~/.config/gh-secrets-sync/secrets.yaml)",
				Category: "Configuration",
			},
		},
	}
}
