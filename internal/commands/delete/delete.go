package delete

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/urfave/cli/v3"

	"github.com/brpaz/gh-secrets-sync/internal/cmdutil"
	"github.com/brpaz/gh-secrets-sync/internal/config"
)

const (
	name      = "delete"
	usage     = "Delete a secret from the config"
	usageText = "gh secrets-sync delete [--name <name>] [--yes]\n\nRemoves a secret entry from the local config file. A confirmation prompt is\nshown before the deletion unless --yes is passed.\n\nNote: this command only modifies the local config file. It does NOT remove\nthe secret from any GitHub repository. To revoke a secret from GitHub you\nmust do so via the GitHub UI or API separately.\n\nExample:\n   gh secrets-sync delete --name MY_TOKEN\n   gh secrets-sync delete --name MY_TOKEN --yes"
)

// New returns the CLI subcommand for deleting a secret from the config.
func New() *cli.Command {
	return &cli.Command{
		Name:      name,
		Usage:     usage,
		UsageText: usageText,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "Secret name to delete",
			},
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "Skip confirmation prompt",
			},
		},
		Action: run,
	}
}

func run(_ context.Context, cmd *cli.Command) error {
	name := cmd.String("name")
	if name == "" {
		if err := survey.AskOne(&survey.Input{Message: "Secret name:"}, &name, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
	}

	if !cmd.Bool("yes") {
		var confirmed bool
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Delete secret %q from config?", name),
			Default: false,
		}
		if err := survey.AskOne(prompt, &confirmed); err != nil {
			return err
		}
		if !confirmed {
			fmt.Fprintln(cmd.Writer, "Aborted.")
			return nil
		}
	}

	path, err := cmdutil.ConfigPath(cmd)
	if err != nil {
		return err
	}

	cfg, err := config.Load(path)
	if err != nil {
		return err
	}

	if err := cfg.DeleteSecret(name); err != nil {
		return err
	}

	if err := cfg.Save(path); err != nil {
		return err
	}

	fmt.Fprintf(cmd.Writer, "✓ Secret %q deleted from config.\n", name)
	fmt.Fprintln(cmd.Writer, "  Note: the secret was NOT removed from any GitHub repositories.")
	return nil
}
