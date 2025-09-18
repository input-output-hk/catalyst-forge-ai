package cmd

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/input-output-hk/catalyst-forge-libs/fs/billy"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ifs "github.com/input-output-hk/catalyst-forge-ai/cli/internal/fs"
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
		{
			name:      "invalid oci ref",
			args:      []string{"init", "myproj", "--template=invalid"},
			wantError: true,
			contains:  "invalid registry format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "forge-ai"}
			cmd.AddCommand(initCmd)

			// Inject in-memory filesystem into context
			ctx := ifs.With(context.Background(), billy.NewInMemoryFS())
			cmd.SetContext(ctx)

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
	// Inject fake OCI client
	originalFactory := newOCIClient
	newOCIClient = func(templateRef string) (ociClient, error) {
		return &fakeOCIClient{pullCalled: false}, nil
	}
	defer func() { newOCIClient = originalFactory }()

	args := []string{"init", "sample-project", "--template=ghcr.io/forge/template:v1.0.0"}

	root := &cobra.Command{Use: "forge-ai"}
	root.AddCommand(initCmd)

	// Inject in-memory filesystem into context
	ctx := ifs.With(context.Background(), billy.NewInMemoryFS())
	root.SetContext(ctx)
	// Ensure subcommand sees the same context
	initCmd.SetContext(ctx)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	execErr := root.Execute()
	require.NoError(t, execErr)

	// Assertions that structure and files exist in injected FS
	fs := ifs.From(root.Context())
	// Generated paths are relative to current working dir in the in-memory FS
	_, statErr := fs.Stat(filepath.Join(".forge", "ai"))
	assert.NoError(t, statErr)
	_, statErr = fs.Stat(filepath.Join(".forge", "project.yaml"))
	assert.NoError(t, statErr)
}

// fakeOCIClient is a minimal fake implementing Pull for tests
type fakeOCIClient struct{ pullCalled bool }

func (f *fakeOCIClient) Pull(_ context.Context, _, targetDir string) error {
	// This fake does nothing; the real code writes project.yaml separately
	f.pullCalled = true
	return nil
}
