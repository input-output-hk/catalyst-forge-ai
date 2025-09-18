package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	ifs "github.com/input-output-hk/catalyst-forge-ai/cli/internal/fs"
	"github.com/input-output-hk/catalyst-forge-ai/cli/internal/mcpserver"
)

// mcpServeCmd starts the MCP server over STDIO.
var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server over STDIO",
	Long:  `Starts a JSON-RPC MCP server that communicates over stdin/stdout. Useful for integration with MCP-compatible agents.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Context with cancellation on SIGINT/SIGTERM
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		// Inject primary filesystem into context
		fs := ifs.From(cmd.Context())
		ctx = ifs.With(ctx, fs)

		logger := slog.Default()
		if err := mcpserver.RunStdio(ctx, logger); err != nil {
			// The server should return cleanly on disconnect; other errors are reported.
			return fmt.Errorf("mcp server failed: %w", err)
		}
		return nil
	},
}

func init() {
	// Ensure parent exists and then register serve
	mcpCmd.AddCommand(mcpServeCmd)
	// No flags for MVP; logging handled via slog default
	// Keep stderr available for diagnostics
	_ = os.Stderr
}
