package sync

import (
	"context"
	"fmt"
	"io"

	"github.com/brpaz/gh-secrets-sync/internal/cmdutil"
	"github.com/brpaz/gh-secrets-sync/internal/config"
	"github.com/brpaz/gh-secrets-sync/internal/gh"

	"github.com/urfave/cli/v3"
)

// GitHubClient is the interface this command requires from the GitHub client.
// Defined here (caller side) following idiomatic Go interface placement.
type GitHubClient interface {
	UpsertRepoSecret(ctx context.Context, req gh.UpsertSecretRequest) error
}

const (
	name      = "sync"
	usage     = "Sync secrets to GitHub repositories"
	usageText = "gh secrets-sync sync [--secret <name>] [--dry-run]\n\nPushes all secrets from the local config file to their configured GitHub\nrepositories using the GitHub API (via the gh CLI authentication).\n\nUse --secret to sync only a single named secret instead of all secrets.\nUse --dry-run to preview what would be synced without making any API calls\n(works even when gh is not authenticated).\n\nExample:\n   gh secrets-sync sync\n   gh secrets-sync sync --secret MY_TOKEN\n   gh secrets-sync sync --dry-run"
)

// New returns the CLI subcommand for syncing secrets to GitHub repositories.
// client is the GitHub client injected from the caller; it may be nil only when
// --dry-run is used (the client is never called in that mode).
func New(client GitHubClient) *cli.Command {
	return &cli.Command{
		Name:      name,
		Usage:     usage,
		UsageText: usageText,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "secret",
				Aliases: []string{"s"},
				Usage:   "Sync only the named secret",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Print what would be synced without making API calls",
			},
		},
		Action: runAction(client),
	}
}

func runAction(client GitHubClient) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		path, err := cmdutil.ConfigPath(cmd)
		if err != nil {
			return err
		}

		cfg, err := config.Load(path)
		if err != nil {
			return err
		}

		if len(cfg.Secrets) == 0 {
			return fmt.Errorf("no secrets configured – run 'gh secrets-sync add' to get started")
		}

		secrets := cfg.Secrets
		if name := cmd.String("secret"); name != "" {
			secrets = nil
			for _, s := range cfg.Secrets {
				if s.Name == name {
					secrets = []config.Secret{s}
					break
				}
			}
			if secrets == nil {
				return fmt.Errorf("secret %q not found in config", name)
			}
		}

		dryRun := cmd.Bool("dry-run")
		if dryRun {
			fmt.Fprintln(cmd.Writer, "[DRY RUN] The following secrets would be synced:")
		}

		synced, failed := runSync(ctx, cmd.Writer, secrets, client, dryRun)

		fmt.Fprintln(cmd.Writer)
		fmt.Fprintf(cmd.Writer, "Summary: %d synced, %d failed\n", synced, failed)

		if failed > 0 {
			return fmt.Errorf("%d sync operation(s) failed", failed)
		}
		return nil
	}
}

// runSync processes all secrets and returns the counts of synced and failed operations.
// client may be nil when dryRun is true.
func runSync(ctx context.Context, w io.Writer, secrets []config.Secret, client GitHubClient, dryRun bool) (synced, failed int) {
	for _, secret := range secrets {
		if len(secret.Repositories) == 0 {
			fmt.Fprintf(w, "⚠ %s has no target repositories configured and was skipped.\n", secret.Name)
			continue
		}

		for _, ownerRepo := range secret.Repositories {
			if dryRun {
				fmt.Fprintf(w, "  → %s → %s\n", secret.Name, ownerRepo)
				synced++
				continue
			}

			if err := client.UpsertRepoSecret(ctx, gh.UpsertSecretRequest{
				Repo:  ownerRepo,
				Name:  secret.Name,
				Value: secret.Value,
			}); err != nil {
				fmt.Fprintf(w, "  ✗ %s → %s  [%v]\n", secret.Name, ownerRepo, err)
				failed++
				continue
			}

			fmt.Fprintf(w, "  ✓ %s → %s\n", secret.Name, ownerRepo)
			synced++
		}
	}
	return synced, failed
}
