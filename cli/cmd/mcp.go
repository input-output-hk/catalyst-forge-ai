package cmd

import "github.com/spf13/cobra"

// mcpCmd represents the MCP command group
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Model Context Protocol (MCP) server commands",
	Long:  `Commands for running the MCP server used by agents to interact with Forge AI.`,
	// No Run: this is a command group
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
