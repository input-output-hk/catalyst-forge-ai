package mcpserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewServer constructs a basic MCP server instance with no tools registered.
// The server responds to core protocol methods like tools/list.
func NewServer() *mcp.Server {
	impl := &mcp.Implementation{Name: "forge-ai-mcp", Version: "v0.1.0"}
	s := mcp.NewServer(impl, nil)
	// Register built-in tools for MVP
	registerNextTool(s)
	registerPlanningTools(s)
	registerImplementationTools(s)
	return s
}

// RunStdio starts the MCP server over STDIO and blocks until the client disconnects.
// Basic lifecycle logs are emitted to help with debugging.
func RunStdio(ctx context.Context, logger *slog.Logger) error {
	if logger == nil {
		logger = slog.Default()
	}

	server := NewServer()
	logger.Info("starting MCP server (stdio)")
	defer logger.Info("MCP server stopped")

	// The stdio transport uses os.Stdin/stdout under the hood.
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		return fmt.Errorf("failed to run MCP server: %w", err)
	}
	return nil
}
