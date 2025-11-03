// Package cmd defines the root command for the bkp CLI application.
//
// It serves as the entry point for the command-line interface,
// initializing the base command and integrating all available
// subcommands (e.g., `bkp version`).
//
// The package uses the Cobra library to handle command parsing,
// argument validation, and flag management. Additional subcommands
// should be added to the root command in their respective init()
// functions (typically in other files within the cmd package).
//
// Example usage:
//
//	bkp --help
//	bkp version
//	bkp ./mydata ./backup
package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcelhaindl/bkp/internal/cli/version"
	"github.com/spf13/cobra"
)

// rootCmd is the base command for the bkp CLI.
//
// When invoked with two arguments, it performs a backup operation
// (copying files or directories). When invoked without arguments,
// it shows help or delegates to a subcommand such as `version`.
var rootCmd = &cobra.Command{
	Use:   "bkp [source] [destination]",
	Short: "A simple CLI tool to back up files and directories",
	Long: `bkp is a simple command-line tool to back up files and directories.

Easily create backups of important data to ensure your files are safe and secure.
The tool is lightweight, user-friendly, and supports various backup options to fit
your needs — from single files to entire directories.`,
	Args:         cobra.MaximumNArgs(2),
	RunE:         runE,
	SilenceUsage: true,
}

// Execute runs the root command and acts as the main entry point
// for the CLI application. It should be called from main().
//
// Example:
//
//	func main() {
//	    cmd.Execute()
//	}
func Execute() error {
	return rootCmd.Execute()
}

// init configures the CLI by adding subcommands and initializing
// flags or other global settings.
func init() {
	rootCmd.AddCommand(version.NewCommand())
}

// runE executes the backup operation when two positional arguments
// (source and destination) are provided. If fewer arguments are
// supplied, it displays the help text.
//
// Parameters:
//   - cmd:  the Cobra command instance executing this function
//   - args: the command-line arguments (expected: [source, destination])
//
// Returns:
//   - error: any error encountered during the backup operation
func runE(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return cmd.Help()
	}

	src, dst := args[0], args[1]

	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("unable to access source: %w", err)
	}

	if srcInfo.IsDir() {
		return backupDirectory(src, dst)
	}

	return backupFile(src, dst)
}

// backupDirectory recursively copies all files and subdirectories
// from src to dst, preserving directory structure and file permissions.
//
// Parameters:
//   - src: source directory path
//   - dst: destination directory path
//
// Returns:
//   - error: any error encountered during the directory traversal or copy
func backupDirectory(src, dst string) error {
	if err := ensureDstOutsideSrc(src, dst); err != nil {
		return err
	}

	return filepath.WalkDir(src, func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if dir.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		return backupFile(path, targetPath)
	})
}

// backupFile copies a single file from src to dst.
//
// If dst is an existing directory, the source file is copied into that
// directory, preserving its filename. If dst is a file path, it is overwritten.
//
// Parameters:
//   - src: source file path
//   - dst: destination file path or directory path
//
// Returns:
//   - error: any error encountered while reading or writing files
func backupFile(src, dst string) error {
	if err := ensureDstOutsideSrc(src, dst); err != nil {
		return err
	}

	dstInfo, err := os.Stat(dst)
	if err == nil && dstInfo.IsDir() {
		dst = filepath.Join(dst, filepath.Base(src))
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to access destination: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directories: %w", err)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return dstFile.Sync()
}

// ensureDstOutsideSrc validates that the destination path is not inside
// the source path.
//
// Parameters:
//   - src: source file or directory path
//   - dst: destination file or directory path
//
// Returns:
//   - error: if the destination is inside the source directory or if there
//     was an error computing absolute or relative paths
func ensureDstOutsideSrc(src, dst string) error {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of source: %w", err)
	}

	absDst, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of destination: %w", err)
	}

	rel, err := filepath.Rel(absSrc, absDst)
	if err != nil {
		return fmt.Errorf("failed to compute relative path: %w", err)
	}

	if rel == "." || !strings.HasPrefix(rel, "..") {
		return fmt.Errorf("destination cannot be inside source directory")
	}

	return nil
}
