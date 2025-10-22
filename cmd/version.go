package main

import (
	"fmt"
	"runtime"

	"github.com/1Money-Co/1money-go-sdk"
)

// Build information.
// These variables are set via -ldflags during build time.
var (
	// version is the semantic version, defaults to SDK version.
	// Can be overridden via: -ldflags "-X main.version=x.x.x"
	version = onemoney.Version

	// gitCommit is the git commit hash.
	// Set via: -ldflags "-X main.gitCommit=$(git rev-parse --short HEAD)"
	gitCommit = "unknown"

	// buildTime is the build timestamp.
	// Set via: -ldflags "-X main.buildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')"
	buildTime = "unknown"

	// goVersion is the Go version used to build the binary.
	goVersion = runtime.Version()
)

// VersionInfo returns a formatted version string with all build information.
func VersionInfo() string {
	return fmt.Sprintf(
		"Version:    %s\nGit Commit: %s\nBuild Time: %s\nGo Version: %s",
		version,
		gitCommit,
		buildTime,
		goVersion,
	)
}

// ShortVersion returns just the version number.
func ShortVersion() string {
	if gitCommit != "unknown" && gitCommit != "" {
		return fmt.Sprintf("%s (commit: %s)", version, gitCommit)
	}
	return version
}
