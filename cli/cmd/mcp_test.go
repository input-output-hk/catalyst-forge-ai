package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestMCPCmd_Structure(t *testing.T) {
	c := &cobra.Command{Use: "forge-ai"}
	c.AddCommand(mcpCmd)

	// Root should contain the mcp command
	found := false
	for _, sc := range c.Commands() {
		if sc.Use == mcpCmd.Use {
			found = true
			break
		}
	}
	assert.True(t, found, "root should include mcp command")

	// mcp is a parent grouping command (no Run/RunE)
	assert.Nil(t, mcpCmd.RunE)
	assert.Nil(t, mcpCmd.Run)

	// Help output should include Short description
	buf := new(bytes.Buffer)
	mcpCmd.SetOut(buf)
	mcpCmd.SetErr(buf)
	_ = mcpCmd.Help()
	// Help will print the Long description; assert on either Short or Long to avoid brittleness
	out := buf.String()
	foundDesc := (bytes.Contains([]byte(out), []byte(mcpCmd.Short)) || bytes.Contains([]byte(out), []byte(mcpCmd.Long)))
	assert.True(t, foundDesc, "help output should include command description")
}
