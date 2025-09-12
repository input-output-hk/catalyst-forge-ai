package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCommand_Validation(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError bool
		contains  string
	}{
		{name: "missing args", args: []string{"init"}, wantError: true, contains: "accepts 1 arg(s)"},
		{name: "missing template flag", args: []string{"init", "myproj"}, wantError: true, contains: "required flag"},
		{name: "invalid oci ref", args: []string{"init", "myproj", "--template=invalid"}, wantError: true, contains: "invalid registry format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "forge-ai"}
			cmd.AddCommand(initCmd)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.wantError {
				assert.Error(t, err)
				if tt.contains != "" {
					assert.Contains(t, err.Error(), tt.contains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInitCommand_ProjectStructureAndProjectYAML(t *testing.T) {
	tmp := t.TempDir()

	// Change into temp dir so the command writes here
	cwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(cwd) }()
	require.NoError(t, os.Chdir(tmp))

	// Inject fake OCI client
	originalFactory := newOCIClient
	newOCIClient = func() (ociClient, error) {
		return &fakeOCIClient{pullCalled: false}, nil
	}
	defer func() { newOCIClient = originalFactory }()

	args := []string{"init", "sample-project", "--template=ghcr.io/forge/template:v1.0.0"}

	root := &cobra.Command{Use: "forge-ai"}
	root.AddCommand(initCmd)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.Execute()
	require.NoError(t, err)

	// Assertions that structure and files exist
	_, err = os.Stat(filepath.Join(".forge", "ai"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(".forge", "project.yaml"))
	assert.NoError(t, err)
}

// fakeOCIClient is a minimal fake implementing Pull for tests
type fakeOCIClient struct{ pullCalled bool }

func (f *fakeOCIClient) Pull(_ context.Context, _ string, targetDir string) error {
	// Ensure target dir exists to simulate extraction
	_ = os.MkdirAll(targetDir, 0o755)
	f.pullCalled = true
	return nil
}
