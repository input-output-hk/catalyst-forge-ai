package cmd

import (
	"bytes"
	"testing"

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
			args:      []string{"publish", "--registry=ghcr.io/test/repo:v1.0.0"},
			wantError: true,
			errorMsg:  "source",
		},
		{
			name:      "missing registry flag",
			args:      []string{"publish", "--source=./template"},
			wantError: true,
			errorMsg:  "registry",
		},
		{
			name:      "both flags provided",
			args:      []string{"publish", "--source=./template_source", "--registry=ghcr.io/test/repo:v1.0.0"},
			wantError: true, // Will fail because directory doesn't exist in test, but validates flags are parsed
			errorMsg:  "",   // The actual error will be about the source not existing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset command for each test
			rootCmd = newRootCmd()

			// Capture output
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()

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
	cmd := newPublishCmd()

	// Test that source flag exists
	sourceFlag := cmd.Flags().Lookup("source")
	require.NotNil(t, sourceFlag, "source flag should exist")
	assert.Equal(t, "source", sourceFlag.Name)
	assert.Equal(t, "Path to template source directory", sourceFlag.Usage)

	// Test that registry flag exists
	registryFlag := cmd.Flags().Lookup("registry")
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
	// Test with non-existent directory
	exists := sourceDirectoryExists("/this/does/not/exist")
	assert.False(t, exists)

	// Test with current directory (should exist)
	exists = sourceDirectoryExists(".")
	assert.True(t, exists)
}
