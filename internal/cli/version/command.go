// Package version implements the `version` subcommand for the bkp CLI.
//
// It displays build metadata such as the version number, commit hash,
// and build date. These values are typically injected at build time
// using -ldflags, for example:
//
//	go build -ldflags "-X 'github.com/marcelhaindl/bkp/internal/cli/version.Version=v1.2.3' \
//	                   -X 'github.com/marcelhaindl/bkp/internal/cli/version.Commit=abc123' \
//	                   -X 'github.com/marcelhaindl/bkp/internal/cli/version.BuildDate=2025-11-03'"
//
// This enables CI/CD pipelines to produce builds with precise,
// traceable version information.
package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version metadata, typically set at build time via -ldflags.
// Default values are used for local development builds.
var (
	Version   = "dev"     // Semantic version (e.g., "v1.2.3")
	Commit    = "none"    // Git commit hash
	BuildDate = "unknown" // Build timestamp
)

// NewCommand returns a Cobra command that displays version information.
//
// Example:
//
//	bkp version
//
// Returns:
//   - *cobra.Command: the configured version subcommand
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show the current version of bkp",
		Long: `Display detailed version information for the bkp CLI,
including the version number, commit hash, and build date.`,
		RunE: runE,
	}
}

// runE executes the version command, printing version information
// to the command's configured standard output.
//
// Parameters:
//   - cmd:  the Cobra command instance executing this function
//   - args: command-line arguments (ignored for this command)
//
// Returns:
//   - error: any error encountered while writing output
//
// Output format:
//
//	bkp version <Version> (commit <Commit>, built <BuildDate>)
func runE(cmd *cobra.Command, args []string) error {
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "bkp version %s (commit %s, built %s)\n", Version, Commit, BuildDate)
	return err
}
