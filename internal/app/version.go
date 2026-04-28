package app

import (
	"fmt"
	"runtime"
)

// VersionInfo holds build-time metadata for the application.
type VersionInfo struct {
	Version   string
	Commit    string
	BuildDate string
}

// String returns a formatted version string suitable for CLI output.
func (v VersionInfo) String() string {
	return fmt.Sprintf("%s (commit: %s, built: %s, %s)", v.Version, v.Commit, v.BuildDate, runtime.Version())
}
