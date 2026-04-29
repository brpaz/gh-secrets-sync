package attach

import (
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/AlecAivazis/survey/v2"
	"github.com/urfave/cli/v3"

	"github.com/brpaz/gh-secrets-sync/internal/cmdutil"
	"github.com/brpaz/gh-secrets-sync/internal/config"
	"github.com/brpaz/gh-secrets-sync/internal/gh"
)

const (
	name      = "attach"
	usage     = "Attach existing secrets to the current repository"
	usageText = "gh secrets-sync attach\n\nInteractively selects existing secrets from the local config file, adds the\ncurrent GitHub repository to each selected secret, saves the config file, and\nthen syncs those secrets to that repository.\n\nExample:\n   gh secrets-sync attach"
)

type GitHubClient interface {
	CurrentRepository(ctx context.Context) (string, error)
	UpsertRepoSecret(ctx context.Context, req gh.UpsertSecretRequest) error
}

// New returns the CLI subcommand for attaching secrets to the current repository.
func New(client GitHubClient) *cli.Command {
	return &cli.Command{
		Name:      name,
		Usage:     usage,
		UsageText: usageText,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run(ctx, cmd, client)
		},
	}
}

func run(ctx context.Context, cmd *cli.Command, client GitHubClient) error {
	if client == nil {
		return fmt.Errorf("github client is required")
	}

	path, err := cmdutil.ConfigPath(cmd)
	if err != nil {
		return err
	}

	cfg, err := config.Load(path)
	if err != nil {
		return err
	}

	if len(cfg.Secrets) == 0 {
		return fmt.Errorf("no secrets configured – run 'gh secrets-sync add' first")
	}

	currentRepo, err := client.CurrentRepository(ctx)
	if err != nil {
		return err
	}

	selectedNames, err := pickSecrets(cfg, currentRepo)
	if err != nil {
		return err
	}

	return attachSelectedSecrets(ctx, cmd.Writer, path, cfg, selectedNames, currentRepo, client)
}

func attachSelectedSecrets(ctx context.Context, w io.Writer, path string, cfg *config.Config, selectedNames []string, currentRepo string, client GitHubClient) error {
	if len(selectedNames) == 0 {
		return fmt.Errorf("no secrets selected")
	}

	selectedSecrets := make([]config.Secret, 0, len(selectedNames))
	for i, secret := range cfg.Secrets {
		if !slices.Contains(selectedNames, secret.Name) {
			continue
		}

		cfg.Secrets[i].Repositories = addUniqueRepo(secret.Repositories, currentRepo)
		selectedSecrets = append(selectedSecrets, cfg.Secrets[i])
	}

	if err := cfg.Save(path); err != nil {
		return err
	}

	failed := 0
	for _, secret := range selectedSecrets {
		if err := client.UpsertRepoSecret(ctx, gh.UpsertSecretRequest{
			Repo:  currentRepo,
			Name:  secret.Name,
			Value: secret.Value,
		}); err != nil {
			fmt.Fprintf(w, "  ✗ %s → %s  [%v]\n", secret.Name, currentRepo, err)
			failed++
			continue
		}

		fmt.Fprintf(w, "  ✓ %s → %s\n", secret.Name, currentRepo)
	}

	if failed > 0 {
		return fmt.Errorf("%d attach sync operation(s) failed", failed)
	}

	return nil
}

func pickSecrets(cfg *config.Config, currentRepo string) ([]string, error) {
	options := make([]string, len(cfg.Secrets))
	defaults := make([]string, 0)

	for i, secret := range cfg.Secrets {
		options[i] = secret.Name
		if slices.Contains(secret.Repositories, currentRepo) {
			defaults = append(defaults, secret.Name)
		}
	}

	var selected []string
	if err := survey.AskOne(&survey.MultiSelect{
		Message: fmt.Sprintf("Select secrets to attach to %s:", currentRepo),
		Options: options,
		Default: defaults,
	}, &selected); err != nil {
		return nil, err
	}

	return selected, nil
}

func addUniqueRepo(repos []string, repo string) []string {
	if slices.Contains(repos, repo) {
		return repos
	}

	return append(repos, repo)
}
