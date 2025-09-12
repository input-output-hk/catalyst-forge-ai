// Package main is the entry point for the Forge AI CLI application.
package main

import (
	"os"

	"github.com/input-output-hk/catalyst-forge-ai/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
