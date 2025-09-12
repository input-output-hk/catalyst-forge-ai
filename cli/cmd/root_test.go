package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError bool
	}{
		{
			name:      "help flag",
			args:      []string{"--help"},
			wantError: false,
		},
		{
			name:      "no args shows help",
			args:      []string{},
			wantError: false,
		},
		{
			name:      "invalid command",
			args:      []string{"invalid-command"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh command instance for each test
			cmd := &cobra.Command{
				Use:   "forge-ai",
				Short: "Forge AI CLI for managing AI-assisted development workflows",
				Long: `Forge AI is a structured workflow system that guides AI agents through
software development tasks using the Model Context Protocol (MCP).`,
			}

			// Add the config flag
			cmd.PersistentFlags().StringVar(&cfgFile, "config", "",
				"config file (default is $HOME/.forge-ai.yaml)")

			// Add subcommands
			cmd.AddCommand(publishCmd)

			// Reset viper config for each test
			cfgFile = ""

			// Capture output
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRootCommandStructure(t *testing.T) {
	// Test that root command has correct properties
	assert.Equal(t, "forge-ai", rootCmd.Use)
	assert.Contains(t, rootCmd.Short, "Forge AI CLI")
	assert.Contains(t, rootCmd.Long, "Model Context Protocol")

	// Test that root command has no Run function (it's a command group)
	assert.Nil(t, rootCmd.Run)
	assert.Nil(t, rootCmd.RunE)
}

func TestRootCommandFlags(t *testing.T) {
	// Test that config flag exists and is properly bound
	configFlag := rootCmd.PersistentFlags().Lookup("config")
	require.NotNil(t, configFlag, "config flag should exist")
	assert.Equal(t, "config", configFlag.Name)
	assert.Contains(t, configFlag.Usage, "config file")
	assert.Contains(t, configFlag.Usage, ".forge-ai.yaml")
}

func TestExecuteFunction(t *testing.T) {
	// Test that Execute function exists and can be called
	// This is mainly a compilation test, but we can test basic functionality
	assert.NotNil(t, Execute)
}

func TestInitConfig(t *testing.T) {
	tests := []struct {
		name     string
		cfgFile  string
		expected string
	}{
		{
			name:     "explicit config file",
			cfgFile:  "/tmp/test-config.yaml",
			expected: "/tmp/test-config.yaml",
		},
		{
			name:     "no config file specified",
			cfgFile:  "",
			expected: "", // Will use default search paths
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgFile = tt.cfgFile
			initConfig()

			if tt.cfgFile != "" {
				// Should use explicit config file
				assert.Equal(t, tt.expected, cfgFile)
			}
		})
	}
}
