package main

import (
	"github.com/urfave/cli/v3"

	"github.com/brpaz/gh-secrets-sync/cmd/hello"
)

// RootCmd returns the root command for the gh-secrets-sync CLI application.
func RootCmd() *cli.Command {
	return &cli.Command{
		Name:                  "gh-secrets-sync",
		Version:               Version,
		Usage:                 "Github CLI extension that syncs GitHub secrets across different repositories",
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			hello.Command(),
		},
	}
}
