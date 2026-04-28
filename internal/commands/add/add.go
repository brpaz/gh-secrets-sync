package add

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
	name      = "add"
	usage     = "Add a new secret to the config"
	usageText = "gh secrets-sync add [--name <name>] [--value <value>] [--repos <owner/repo>,...] [--force]\n\nAdds a new secret entry to the local config file. All flags are optional –\nany value not provided via a flag will be prompted for interactively.\n\nRepositories can be supplied as multiple --repos flags or as a single\ncomma-separated list (e.g. --repos owner/repo1,owner/repo2).\n\nUse --force to overwrite a secret that already exists in the config.\n\nExample:\n   gh secrets-sync add --name MY_TOKEN --value s3cr3t --repos myorg/api,myorg/web\n   gh secrets-sync add --name MY_TOKEN --force"
)

// New returns the CLI subcommand for adding a new secret to the config.
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
				Usage:   "Secret value",
			},
			&cli.StringSliceFlag{
				Name:    "repos",
				Aliases: []string{"r"},
				Usage:   "Target repositories (owner/repo); can be repeated or comma-separated",
			},
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Overwrite existing secret if name already exists",
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

	value := cmd.String("value")
	if value == "" {
		if err := survey.AskOne(&survey.Password{Message: "Secret value:"}, &value, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
	}

	repos := cmdutil.SplitRepos(cmd.StringSlice("repos"))
	if len(repos) == 0 {
		var raw string
		if err := survey.AskOne(&survey.Input{Message: "Repositories (comma-separated):"}, &raw, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		repos = cmdutil.SplitRepos([]string{raw})
	}

	force := cmd.Bool("force")

	path, err := cmdutil.ConfigPath(cmd)
	if err != nil {
		return err
	}

	cfg, err := config.Load(path)
	if err != nil {
		return err
	}

	secret := config.Secret{
		Name:         name,
		Value:        value,
		Repositories: repos,
	}

	if err := cfg.AddSecret(secret, force); err != nil {
		return err
	}

	if err := cfg.Save(path); err != nil {
		return err
	}

	fmt.Fprintf(cmd.Writer, "✓ Secret %q added for repos: %s\n", name, strings.Join(repos, ", "))
	return nil
}
