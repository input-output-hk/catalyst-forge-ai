package cmd

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"testing"

	vfs "github.com/input-output-hk/catalyst-forge-libs/fs"
	"github.com/input-output-hk/catalyst-forge-libs/fs/billy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ifs "github.com/input-output-hk/catalyst-forge-ai/cli/internal/fs"
)

func TestGenerateTaskID(t *testing.T) {
	tests := []struct {
		name         string
		title        string
		existingIDs  []string
		expectedID   string
		expectedPath string
	}{
		{
			name:         "first task",
			title:        "Implement User Authentication",
			existingIDs:  []string{},
			expectedID:   "001-implement-user-authentication",
			expectedPath: "tasks/001-implement-user-authentication",
		},
		{
			name:         "second task",
			title:        "Add Database Schema",
			existingIDs:  []string{"001-implement-user-authentication"},
			expectedID:   "002-add-database-schema",
			expectedPath: "tasks/002-add-database-schema",
		},
		{
			name:         "task with special characters",
			title:        "Fix API/Response Handling!",
			existingIDs:  []string{"001-implement-user-authentication", "002-add-database-schema"},
			expectedID:   "003-fix-api-response-handling",
			expectedPath: "tasks/003-fix-api-response-handling",
		},
		{
			name:         "large number",
			title:        "Some Task",
			existingIDs:  []string{"099-last-task"},
			expectedID:   "100-some-task",
			expectedPath: "tasks/100-some-task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use synthetic repo root entirely in memory
			repoRoot := "/repo"

			// Generate task ID (explicit FS)
			mem := billy.NewInMemoryFS()
			// Create tasks dir to simulate environment
			_ = mem.MkdirAll(filepath.Join(repoRoot, "tasks"), 0o755)
			// Mirror existing IDs in in-memory FS
			for _, id := range tt.existingIDs {
				_ = mem.MkdirAll(filepath.Join(repoRoot, "tasks", id), 0o755)
				_ = mem.WriteFile(filepath.Join(repoRoot, "tasks", id, "task.yaml"), []byte("id: \""+id+"\"\n"), 0o644)
			}
			id, err := generateTaskID(mem, repoRoot, tt.title)

			// Assert the result
			require.NoError(t, err)
			assert.Equal(t, tt.expectedID, id)

			// Assert the task directory path
			expectedPath := filepath.Join(repoRoot, tt.expectedPath)
			assert.Equal(t, expectedPath, filepath.Join(repoRoot, "tasks", id))
		})
	}
}

func TestTaskNewCommand_Validation(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantError  bool
		contains   string
		setupGit   bool
		setupForge bool
	}{
		{
			name:      "missing title flag",
			args:      []string{"task", "new"},
			wantError: true,
			contains:  "title cannot be empty",
		},
		{
			name:      "empty title",
			args:      []string{"task", "new", "--title="},
			wantError: true,
			contains:  "title cannot be empty",
		},
		{
			name:       "no git repository",
			args:       []string{"task", "new", "--title=Test Task"},
			wantError:  true,
			contains:   "no git repository found",
			setupGit:   false,
			setupForge: false,
		},
		{
			name:       "not a forge project",
			args:       []string{"task", "new", "--title=Test Task"},
			wantError:  true,
			contains:   "not a Forge AI project",
			setupGit:   true,
			setupForge: false,
		},
		{
			name:       "valid title",
			args:       []string{"task", "new", "--title=Implement User Auth"},
			wantError:  true, // This will fail because we haven't set up the full project structure yet
			contains:   "",
			setupGit:   true,
			setupForge: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Synthetic in-mem repo root
			repoRoot := "/repo"

			// Build command tree
			cmd := &cobra.Command{Use: "forge-ai"}
			cmd.AddCommand(taskCmd)

			// Inject FS into context
			mem := billy.NewInMemoryFS()
			ctx := ifs.With(context.Background(), mem)
			cmd.SetContext(ctx)
			taskCmd.SetContext(ctx)
			taskNewCmd.SetContext(ctx)

			// Override findRoot for test
			origFindRoot := findRoot
			defer func() { findRoot = origFindRoot }()

			switch tt.name {
			case "no git repository":
				findRoot = func(_ vfs.Filesystem) (string, error) {
					return "", fmt.Errorf("no git repository found")
				}
			default:
				findRoot = func(_ vfs.Filesystem) (string, error) {
					return repoRoot, nil
				}
			}

			// Setup in-memory project depending on scenario
			if tt.setupGit {
				_ = mem.MkdirAll(filepath.Join(repoRoot, ".git"), 0o755)
			}
			if tt.setupForge {
				_ = mem.MkdirAll(filepath.Join(repoRoot, ".forge", "ai"), 0o755)
			}

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			// Ensure title is provided via viper for cases expecting downstream errors
			switch tt.name {
			case "no git repository", "not a forge project", "valid title":
				// Match provided args' titles for clarity
				if tt.name == "valid title" {
					viper.Set("task.new.title", "Implement User Auth")
				} else {
					viper.Set("task.new.title", "Test Task")
				}
				t.Cleanup(func() { viper.Set("task.new.title", "") })
			}

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

func TestTaskNewCommand_TaskCreation(t *testing.T) {
	// Synthetic repo root
	repoRoot := "/repo"

	// Execute the command
	cmd := &cobra.Command{Use: "forge-ai"}
	cmd.AddCommand(taskCmd)

	// Inject FS into context
	mem := billy.NewInMemoryFS()
	// Mirror project structure in the in-memory FS
	_ = mem.MkdirAll(filepath.Join(repoRoot, ".git"), 0o755)
	_ = mem.MkdirAll(filepath.Join(repoRoot, ".forge", "ai"), 0o755)
	// Create default template file in mem FS
	_ = mem.MkdirAll(filepath.Join(repoRoot, ".forge", "ai", "templates", "tasks"), 0o755)
	_ = mem.WriteFile(
		filepath.Join(repoRoot, ".forge", "ai", "templates", "tasks", "default.yaml"),
		[]byte("id: \"\"\ntitle: \"\"\ncurrent_phase: \"planning\"\nstatus: \"active\"\n"),
		0o644,
	)
	// Create project.yaml in mem FS
	projectYAML := `projectName: test-project
status: active
template:
  source: ghcr.io/forge/template
  version: v1.0.0
`
	_ = mem.MkdirAll(filepath.Join(repoRoot, ".forge"), 0o755)
	_ = mem.WriteFile(filepath.Join(repoRoot, ".forge", "project.yaml"), []byte(projectYAML), 0o644)

	// Override findRoot to use synthetic repo root
	origFindRoot := findRoot
	findRoot = func(_ vfs.Filesystem) (string, error) {
		return repoRoot, nil
	}
	defer func() { findRoot = origFindRoot }()

	ctx := ifs.With(context.Background(), mem)
	cmd.SetContext(ctx)
	taskCmd.SetContext(ctx)
	taskNewCmd.SetContext(ctx)

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Provide title via viper instead of flag for clarity
	viper.Set("task.new.title", "Implement User Authentication")
	t.Cleanup(func() { viper.Set("task.new.title", "") })

	cmd.SetArgs([]string{"task", "new"})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify task directory was created in the in-memory FS
	entries, err := mem.ReadDir(filepath.Join(repoRoot, "tasks"))
	require.NoError(t, err)
	require.Len(t, entries, 1)

	taskID := entries[0].Name()
	taskDir := filepath.Join(repoRoot, "tasks", taskID)

	// Verify task.yaml exists and has correct content (via mem FS)
	taskYAMLPath := filepath.Join(taskDir, "task.yaml")
	_, statErr := mem.Stat(taskYAMLPath)
	require.NoError(t, statErr)

	taskYAMLContent, err := mem.ReadFile(taskYAMLPath)
	require.NoError(t, err)

	// Check that the task.yaml contains expected fields
	taskYAMLStr := string(taskYAMLContent)
	assert.Contains(t, taskYAMLStr, fmt.Sprintf("id: \"%s\"", taskID))
	assert.Contains(t, taskYAMLStr, "title: \"Implement User Authentication\"")
	assert.Contains(t, taskYAMLStr, "current_phase: \"planning\"")
	assert.Contains(t, taskYAMLStr, "status: \"active\"")

	// Verify project.yaml was updated in mem FS
	updatedProjectYAML, err := mem.ReadFile(filepath.Join(repoRoot, ".forge", "project.yaml"))
	require.NoError(t, err)

	projectYAMLStr := string(updatedProjectYAML)
	assert.Contains(t, projectYAMLStr, taskID)
	assert.Contains(t, projectYAMLStr, "phase: planning")
	assert.Contains(t, projectYAMLStr, "status: active")
	assert.Contains(t, projectYAMLStr, fmt.Sprintf("path: tasks/%s", taskID))
}
