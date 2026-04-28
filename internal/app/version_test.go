package app_test

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/gh-secrets-sync/internal/app"
)

func TestVersionInfo_String(t *testing.T) {
	t.Parallel()
	t.Run("formats version info correctly", func(t *testing.T) {
		t.Parallel()
		v := app.VersionInfo{
			Version:   "1.2.3",
			Commit:    "abc123",
			BuildDate: "2024-06-01T12:00:00Z",
		}
		expected := "1.2.3 (commit: abc123, built: 2024-06-01T12:00:00Z, " + runtime.Version() + ")"
		assert.Equal(t, expected, v.String())
	})
}
