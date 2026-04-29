package edit

import (
	"context"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/urfave/cli/v3"

	"github.com/brpaz/gh-secrets-sync/internal/cmdutil"
	"github.com/brpaz/gh-secrets-sync/internal/config"
)

const (
	name      = "edit"
	usage     = "Edit an existing secret in the config"
	usageText = "gh secrets-sync edit [--name <name>] [--value <value>] [--repos <owner/repo>,...]\n\nEdits an existing secret entry in the local config file. All flags are\noptional – if --name is omitted you will be prompted to pick from a list of\nconfigured secrets.\n\nWhen --value or --repos are omitted, the current settings are shown in the\nprompts. Secret values stay masked by default, while repositories are shown as\na pre-filled comma-separated list. Repositories may be left empty so the secret\nis stored without sync targets.\n\nExample:\n   gh secrets-sync edit --name MY_TOKEN --value newvalue\n   gh secrets-sync edit --name MY_TOKEN --repos myorg/api,myorg/web"
)

// New returns the CLI subcommand for updating an existing secret in the config.
func New() *cli.Command {
	return &cli.Command{
		Name:      name,
		Usage:     usage,
		UsageText: usageText,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "Secret name",
			},
			&cli.StringFlag{
				Name:    "value",
				Aliases: []string{"v"},
				Usage:   "New secret value (optional – only updated when provided)",
			},
			&cli.StringSliceFlag{
				Name:    "repos",
				Aliases: []string{"r"},
				Usage:   "New target repositories (owner/repo); can be repeated or comma-separated",
			},
		},
		Action: run,
	}
}

func run(_ context.Context, cmd *cli.Command) error {
	path, err := cmdutil.ConfigPath(cmd)
	if err != nil {
		return err
	}

	cfg, err := config.Load(path)
	if err != nil {
		return err
	}

	// Resolve secret name first – prompt if not provided via flag.
	name := cmd.String("name")
	if name == "" {
		name, err = pickSecret(cfg)
		if err != nil {
			return err
		}
	}

	secret, err := findSecret(cfg, name)
	if err != nil {
		return err
	}

	value := cmd.String("value")
	if !cmd.IsSet("value") {
		if err := survey.AskOne(&survey.Password{
			Message: "Secret value (current: ****, leave blank to keep):",
		}, &value); err != nil {
			return err
		}
	}

	var repos []string
	if cmd.IsSet("repos") {
		rawRepos := cmd.StringSlice("repos")
		repos = cmdutil.SplitRepos(rawRepos)
		if len(repos) == 0 {
			repos = []string{}
		}
	} else {
		var raw string
		if err := survey.AskOne(&survey.Input{
			Message: "Repositories (comma-separated, optional):",
			Default: strings.Join(secret.Repositories, ","),
		}, &raw); err != nil {
			return err
		}
		repos = cmdutil.SplitRepos([]string{raw})
		if raw == "" {
			repos = []string{}
		}
	}

	patch := config.Secret{
		Value:        value,
		Repositories: repos,
	}

	if err := cfg.UpdateSecret(name, patch); err != nil {
		return err
	}

	if err := cfg.Save(path); err != nil {
		return err
	}

	fmt.Fprintf(cmd.Writer, "✓ Secret %q updated.\n", name)
	return nil
}

// pickSecret prompts the user to select from existing secrets via a survey list.
func pickSecret(cfg *config.Config) (string, error) {
	if len(cfg.Secrets) == 0 {
		return "", fmt.Errorf("no secrets configured – run 'gh secrets-sync add' first")
	}

	names := make([]string, len(cfg.Secrets))
	for i, s := range cfg.Secrets {
		names[i] = s.Name
	}

	var selected string
	if err := survey.AskOne(&survey.Select{
		Message: "Select a secret to edit:",
		Options: names,
	}, &selected); err != nil {
		return "", err
	}

	return selected, nil
}

func findSecret(cfg *config.Config, name string) (config.Secret, error) {
	for _, secret := range cfg.Secrets {
		if secret.Name == name {
			return secret, nil
		}
	}

	return config.Secret{}, fmt.Errorf("secret %q not found", name)
}
