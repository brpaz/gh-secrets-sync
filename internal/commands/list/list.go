package list

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/AlecAivazis/survey/v2"
	"github.com/brpaz/gh-secrets-sync/internal/cmdutil"
	"github.com/brpaz/gh-secrets-sync/internal/config"

	"github.com/urfave/cli/v3"
)

const (
	name      = "list"
	usage     = "List all configured secrets"
	usageText = "gh secrets-sync list [--reveal] [--yes]\n\nPrints a table of all secrets defined in the local config file, showing\nthe secret name, its value (masked by default), and the target repositories.\n\nUse --reveal to display plain-text values. A confirmation prompt is shown\nbefore revealing values unless --yes is also passed.\n\nExample:\n   gh secrets-sync list\n   gh secrets-sync list --reveal\n   gh secrets-sync list --reveal --yes"
)

// New returns the CLI subcommand for listing all configured secrets.
func New() *cli.Command {
	return &cli.Command{
		Name:      name,
		Aliases:   []string{"ls"},
		Usage:     usage,
		UsageText: usageText,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "reveal",
				Usage: "Display actual secret values instead of masked output",
			},
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "Skip confirmation prompt when using --reveal",
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

	if len(cfg.Secrets) == 0 {
		fmt.Fprintln(cmd.Writer, "No secrets configured. Run 'gh secrets-sync add' to get started.")
		return nil
	}

	reveal := cmd.Bool("reveal")
	if reveal && !cmd.Bool("yes") {
		var confirmed bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "This will display secret values in plain text. Continue?",
			Default: false,
		}, &confirmed); err != nil {
			return err
		}
		if !confirmed {
			fmt.Fprintln(cmd.Writer, "Aborted.")
			return nil
		}
	}

	w := tabwriter.NewWriter(cmd.Writer, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SECRET NAME\tVALUE\tREPOSITORIES")
	fmt.Fprintln(w, "-----------\t-----\t------------")

	for _, s := range cfg.Secrets {
		value := "****"
		if reveal {
			value = s.Value
		}

		repos := "—"
		if len(s.Repositories) > 0 {
			repos = strings.Join(s.Repositories, ", ")
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n", s.Name, value, repos)
	}

	return w.Flush()
}
