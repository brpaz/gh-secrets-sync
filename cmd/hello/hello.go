package hello

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:    "hello",
		Aliases: []string{"h"},
		Usage:   "Say hello",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "name to greet",
				Value:   "World",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := cmd.String("name")
			fmt.Printf("Hello, %s!\n", name)
			return nil
		},
		ShellComplete: func(ctx context.Context, cmd *cli.Command) {
			// Custom completion suggestions
			if cmd.NArg() == 0 {
				fmt.Println("--name")
				fmt.Println("-n")
			}
		},
	}
}
