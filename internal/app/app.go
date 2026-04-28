package app

import (
	"context"
	"fmt"

	cli "github.com/urfave/cli/v3"

	addcmd "github.com/brpaz/gh-secrets-sync/internal/commands/add"
	configcmd "github.com/brpaz/gh-secrets-sync/internal/commands/configeditor"
	deletecmd "github.com/brpaz/gh-secrets-sync/internal/commands/delete"
	listcmd "github.com/brpaz/gh-secrets-sync/internal/commands/list"
	rootcmd "github.com/brpaz/gh-secrets-sync/internal/commands/root"
	synccmd "github.com/brpaz/gh-secrets-sync/internal/commands/sync"
	updatecmd "github.com/brpaz/gh-secrets-sync/internal/commands/update"
	"github.com/brpaz/gh-secrets-sync/internal/config"
	"github.com/brpaz/gh-secrets-sync/internal/gh"
)

// App is the composition root for the gh-secrets-sync CLI. It holds all
// top-level dependencies and exposes a single Run entry-point.
type App struct {
	Info         VersionInfo
	GitHubClient *gh.Client
}

// Option is a functional option for configuring an App.
type Option func(*App)

// WithVersionInfo sets the build-time version metadata.
func WithVersionInfo(info VersionInfo) Option {
	return func(a *App) { a.Info = info }
}

// New constructs an App with the provided options
func New(opts ...Option) (*App, error) {
	appInstance := &App{
		Info: VersionInfo{
			Version:   "0.0.0-dev",
			Commit:    "n/a",
			BuildDate: "n/a",
		},
	}

	for _, opt := range opts {
		opt(appInstance)
	}

	if err := appInstance.setup(); err != nil {
		return nil, fmt.Errorf("app setup failed: %w", err)
	}

	return appInstance, nil
}

// setup initializes any application dependencies.
func (app *App) setup() error {
	client, err := gh.NewClient()
	if err != nil {
		return fmt.Errorf("failed to initialize GitHub client: %w", err)
	}
	app.GitHubClient = client

	return nil
}

// Run builds the root command and executes it with the provided arguments.
func (app *App) Run(ctx context.Context, args []string) error {
	root := rootcmd.New(
		rootcmd.WithVersion(app.Info.String()),
		rootcmd.WithOnInit(onInit),
		rootcmd.WithCommand(addcmd.New()),
		rootcmd.WithCommand(configcmd.New()),
		rootcmd.WithCommand(deletecmd.New()),
		rootcmd.WithCommand(listcmd.New()),
		rootcmd.WithCommand(synccmd.New(app.GitHubClient)),
		rootcmd.WithCommand(updatecmd.New()),
	)

	return root.Run(ctx, args)
}

// onInit creates the config file on the very first run and prints a
// getting-started hint so the user knows where to configure their secrets.
func onInit(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	path, err := config.DefaultConfigPath()
	if err != nil {
		return ctx, err
	}

	created, err := config.EnsureConfigExists(path)
	if err != nil {
		return ctx, err
	}

	if created {
		fmt.Fprintf(cmd.Writer, "✓ Config file created: %s\n", path)
		fmt.Fprintln(cmd.Writer, "  Edit the file to add your secrets, then run 'gh secrets-sync sync' to push them.")
		fmt.Fprintln(cmd.Writer)
	}

	return ctx, nil
}
