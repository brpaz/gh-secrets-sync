package configeditor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/brpaz/gh-secrets-sync/internal/config"
)

// resolveEditor returns the value of $EDITOR, or the platform default when the
// variable is unset or empty.
func resolveEditor() string {
	if v := os.Getenv("EDITOR"); v != "" {
		return v
	}
	if runtime.GOOS == "windows" {
		return "notepad"
	}
	return "vi"
}

// editorRunner splits the editor string into a binary and optional arguments
// (supporting values like "code --wait"), verifies the binary exists in PATH,
// then runs the editor interactively.
func editorRunner(ctx context.Context, editor, path string) error {
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return fmt.Errorf("editor command is empty – set the $EDITOR environment variable to your preferred editor")
	}

	bin := parts[0]
	if _, err := exec.LookPath(bin); err != nil {
		return fmt.Errorf("editor %q not found – set the $EDITOR environment variable to your preferred editor", bin)
	}

	args := append(parts[1:], path)
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// A non-zero exit code from the editor is normal (e.g. Vim `:q!`).
		// Only surface genuine launch or IO errors.
		if _, ok := err.(*exec.ExitError); !ok {
			return fmt.Errorf("editor exited with an unexpected error: %w", err)
		}
	}
	return nil
}

const (
	name      = "config"
	usage     = "Open the config file in your editor"
	usageText = "gh secrets-sync config\n\nOpens the secrets config file in your preferred editor. The editor is\ndetermined by the $EDITOR environment variable. When $EDITOR is not set,\nthe default is 'vi' on Unix-like systems and 'notepad' on Windows.\n\nEditors that require extra flags (e.g. VS Code) are supported:\n   EDITOR=\"code --wait\" gh secrets-sync config\n\nThe config file is created automatically on first run if it does not exist."
)

// New returns the CLI subcommand for opening the config file in an editor.
func New() *cli.Command {
	return &cli.Command{
		Name:      name,
		Usage:     usage,
		UsageText: usageText,
		Action: func(ctx context.Context, _ *cli.Command) error {
			path, err := config.DefaultConfigPath()
			if err != nil {
				return err
			}
			return editorRunner(ctx, resolveEditor(), path)
		},
	}
}
