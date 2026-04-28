package configeditor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"

	configeditor "github.com/brpaz/gh-secrets-sync/internal/commands/configeditor"
)

func TestNew(t *testing.T) {
	cmd := configeditor.New()
	assert.IsType(t, cmd, &cli.Command{})
}
