package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublishCommand(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "missing source flag",
			args:      []string{"publish", "--registry=localhost:5000/test/repo:v1.0.0"},
			wantError: true,
			errorMsg:  "required flag",
		},
		{
			name:      "missing registry flag",
			args:      []string{"publish", "--source=./template"},
			wantError: true,
			errorMsg:  "source directory does not exist",
		},
		{
			name:      "invalid registry format",
			args:      []string{"publish", "--source=./template", "--registry=invalid-registry"},
			wantError: true,
			errorMsg:  "invalid registry format",
		},
		{
			name:      "source directory doesn't exist",
			args:      []string{"publish", "--source=./nonexistent", "--registry=localhost:5000/test/repo:v1.0.0"},
			wantError: true,
			errorMsg:  "source directory does not exist",
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
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPublishCommandFlags(t *testing.T) {
	// Test that source flag exists
	sourceFlag := publishCmd.Flags().Lookup("source")
	require.NotNil(t, sourceFlag, "source flag should exist")
	assert.Equal(t, "source", sourceFlag.Name)
	assert.Equal(t, "Path to template source directory", sourceFlag.Usage)

	// Test that registry flag exists
	registryFlag := publishCmd.Flags().Lookup("registry")
	require.NotNil(t, registryFlag, "registry flag should exist")
	assert.Equal(t, "registry", registryFlag.Name)
	assert.Equal(t, "OCI registry reference (e.g., ghcr.io/org/template:v1.0.0)", registryFlag.Usage)
}

func TestValidatePublishArgs(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		registry  string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "empty source",
			source:    "",
			registry:  "ghcr.io/test/repo:v1.0.0",
			wantError: true,
			errorMsg:  "source directory is required",
		},
		{
			name:      "empty registry",
			source:    "./template",
			registry:  "",
			wantError: true,
			errorMsg:  "registry reference is required",
		},
		{
			name:      "invalid registry format",
			source:    "./template",
			registry:  "not-a-valid-registry",
			wantError: true,
			errorMsg:  "invalid registry format",
		},
		{
			name:      "valid inputs",
			source:    "./template",
			registry:  "ghcr.io/org/template:v1.0.0",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePublishArgs(tt.source, tt.registry)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRegistryFormat(t *testing.T) {
	tests := []struct {
		registry string
		valid    bool
	}{
		{"ghcr.io/org/repo:v1.0.0", true},
		{"docker.io/user/image:latest", true},
		{"quay.io/company/app:v2.1.0", true},
		{"localhost:5000/test:v1", true},
		{"invalid", false},
		{"http://registry.com/repo", false},
		{"registry.com", false},
		{"registry.com/", false},
	}

	for _, tt := range tests {
		t.Run(tt.registry, func(t *testing.T) {
			valid := isValidRegistryFormat(tt.registry)
			assert.Equal(t, tt.valid, valid, "Registry format validation failed for: %s", tt.registry)
		})
	}
}

func TestSourceDirectoryExists(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, tmpDir string)
		path     string
		expected bool
	}{
		{
			name:     "non-existent directory",
			setup:    func(t *testing.T, tmpDir string) {},
			path:     "non-existent-dir",
			expected: false,
		},
		{
			name: "existing directory",
			setup: func(t *testing.T, tmpDir string) {
				err := os.MkdirAll(tmpDir+"/existing-dir", 0o755)
				require.NoError(t, err)
			},
			path:     "existing-dir",
			expected: true,
		},
		{
			name: "existing file (not directory)",
			setup: func(t *testing.T, tmpDir string) {
				err := os.WriteFile(tmpDir+"/test-file.txt", []byte("test"), 0o644)
				require.NoError(t, err)
			},
			path:     "test-file.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir := t.TempDir()

			// Set up the test scenario
			tt.setup(t, tmpDir)

			// Change to temp directory for relative path testing
			oldWd, err := os.Getwd()
			require.NoError(t, err)

			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			// Test the function
			exists := sourceDirectoryExists(tt.path)
			assert.Equal(t, tt.expected, exists)

			// Change back to original directory
			err = os.Chdir(oldWd)
			require.NoError(t, err)
		})
	}
}
