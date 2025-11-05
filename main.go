// Package main provides the entry point for the bkp CLI application.
//
// It simply invokes the root command defined in the cmd package and
// handles any errors returned during execution by exiting with a
// non-zero status code.
package main

import (
	"os"

	"github.com/marcelhaindl/bkp/cmd"
)

// main executes the root CLI command.
//
// If cmd.Execute returns an error, the program exits with status code 1.
func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
