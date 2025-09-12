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
