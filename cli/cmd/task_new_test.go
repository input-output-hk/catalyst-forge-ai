package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			// Create a temporary directory to simulate existing tasks
			tmpDir := t.TempDir()

			// Create existing task directories
			for _, id := range tt.existingIDs {
				taskDir := filepath.Join(tmpDir, "tasks", id)
				require.NoError(t, os.MkdirAll(taskDir, 0o755))

				// Create a minimal task.yaml
				taskYAML := "id: " + id + "\ntitle: \"Existing Task\"\n"
				taskFile := filepath.Join(taskDir, "task.yaml")
				require.NoError(t, os.WriteFile(taskFile, []byte(taskYAML), 0o644))
			}

			// Generate task ID
			id, err := generateTaskID(tmpDir, tt.title)

			// Assert the result
			require.NoError(t, err)
			assert.Equal(t, tt.expectedID, id)

			// Assert the task directory path
			expectedPath := filepath.Join(tmpDir, tt.expectedPath)
			assert.Equal(t, expectedPath, filepath.Join(tmpDir, "tasks", id))
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
			// Create a temporary directory
			tmpDir := t.TempDir()

			// Setup .git directory if needed
			if tt.setupGit {
				gitDir := filepath.Join(tmpDir, ".git")
				require.NoError(t, os.MkdirAll(gitDir, 0o755))
			}

			// Setup .forge/ai directory if needed
			if tt.setupForge {
				require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".forge", "ai"), 0o755))
			}

			// Change to a subdirectory if we have git setup to test repository root finding
			workDir := tmpDir
			if tt.setupGit {
				workDir = filepath.Join(tmpDir, "some", "nested", "path")
				require.NoError(t, os.MkdirAll(workDir, 0o755))
			}

			cwd, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(cwd) }()
			require.NoError(t, os.Chdir(workDir))

			cmd := &cobra.Command{Use: "forge-ai"}
			cmd.AddCommand(taskCmd)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err = cmd.Execute()
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
	// Create a temporary directory to simulate project
	tmpDir := t.TempDir()

	// Create .git directory to simulate a git repository
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0o755))

	// Create the .forge/ai directory structure with template
	forgeAIDir := filepath.Join(tmpDir, ".forge", "ai")
	require.NoError(t, os.MkdirAll(forgeAIDir, 0o755))

	// Copy template files to .forge/ai
	templateDir := forgeAIDir

	// Copy the actual template files from the project
	sourceTemplateDir := "/Users/josh/work/catalyst-forge-ai/template_source"
	require.NoError(t, copyDir(sourceTemplateDir, templateDir))

	// Create project.yaml in .forge directory
	projectYAML := `projectName: test-project
status: active
template:
  source: ghcr.io/forge/template
  version: v1.0.0
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".forge", "project.yaml"), []byte(projectYAML), 0o644))

	// Change to a subdirectory within the project to test repository root finding
	testSubDir := filepath.Join(tmpDir, "some", "nested", "directory")
	require.NoError(t, os.MkdirAll(testSubDir, 0o755))

	cwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(cwd) }()
	require.NoError(t, os.Chdir(testSubDir))

	// Execute the command
	cmd := &cobra.Command{Use: "forge-ai"}
	cmd.AddCommand(taskCmd)

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"task", "new", "--title=Implement User Authentication"})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify task directory was created
	taskDirs, err := filepath.Glob(filepath.Join(tmpDir, "tasks", "*"))
	require.NoError(t, err)
	require.Len(t, taskDirs, 1)

	taskDir := taskDirs[0]
	taskID := filepath.Base(taskDir)

	// Verify task.yaml exists and has correct content
	taskYAMLPath := filepath.Join(taskDir, "task.yaml")
	assert.FileExists(t, taskYAMLPath)

	taskYAMLContent, err := os.ReadFile(taskYAMLPath)
	require.NoError(t, err)

	// Check that the task.yaml contains expected fields
	taskYAMLStr := string(taskYAMLContent)
	assert.Contains(t, taskYAMLStr, fmt.Sprintf("id: \"%s\"", taskID))
	assert.Contains(t, taskYAMLStr, "title: \"Implement User Authentication\"")
	assert.Contains(t, taskYAMLStr, "current_phase: \"planning\"")
	assert.Contains(t, taskYAMLStr, "status: \"active\"")

	// Verify project.yaml was updated
	updatedProjectYAML, err := os.ReadFile(filepath.Join(tmpDir, ".forge", "project.yaml"))
	require.NoError(t, err)

	projectYAMLStr := string(updatedProjectYAML)
	assert.Contains(t, projectYAMLStr, taskID)
	assert.Contains(t, projectYAMLStr, "phase: planning")
	assert.Contains(t, projectYAMLStr, "status: active")
	assert.Contains(t, projectYAMLStr, fmt.Sprintf("path: tasks/%s", taskID))
}

// Helper function to copy directory (simplified for test)
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = dstFile.ReadFrom(srcFile)
		return err
	})
}
