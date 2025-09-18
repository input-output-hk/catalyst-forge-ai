package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	vfs "github.com/input-output-hk/catalyst-forge-libs/fs"
	"github.com/input-output-hk/catalyst-forge-libs/fs/billy"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	ifs "github.com/input-output-hk/catalyst-forge-ai/cli/internal/fs"
	"github.com/input-output-hk/catalyst-forge-ai/cli/internal/state"
)

func TestTaskPhaseNext(t *testing.T) {
	tests := []struct {
		name           string
		currentPhase   string
		stepsCompleted bool
		activeTask     string
		wantError      bool
		errorContains  string
		expectedPhase  string
	}{
		{
			name:           "successful planning to implementation transition",
			currentPhase:   "planning",
			stepsCompleted: true,
			activeTask:     "001-test-task",
			wantError:      false,
			expectedPhase:  "implementation",
		},
		{
			name:           "successful implementation to validation transition",
			currentPhase:   "implementation",
			stepsCompleted: true,
			activeTask:     "001-test-task",
			wantError:      false,
			expectedPhase:  "validation",
		},
		{
			name:           "cannot transition from planning with incomplete steps",
			currentPhase:   "planning",
			stepsCompleted: false,
			activeTask:     "001-test-task",
			wantError:      true,
			errorContains:  "incomplete steps present",
		},
		{
			name:           "cannot transition from implementation with incomplete steps",
			currentPhase:   "implementation",
			stepsCompleted: false,
			activeTask:     "001-test-task",
			wantError:      true,
			errorContains:  "incomplete steps present",
		},
		{
			name:          "cannot transition from validation phase",
			currentPhase:  "validation",
			activeTask:    "001-test-task",
			wantError:     true,
			errorContains: "already in final phase",
		},
		{
			name:          "no active task",
			activeTask:    "",
			wantError:     true,
			errorContains: "no active task set",
		},
		{
			name:          "active task directory does not exist",
			activeTask:    "999-nonexistent-task",
			wantError:     true,
			errorContains: "active task directory does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Synthetic in-mem repo root
			repoRoot := "/repo"
			taskID := "001-test-task"

			// Create a temporary directory and change to it
			tempDir, err := os.MkdirTemp("", "forge-test")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// Change to temp directory so FindRoot will work
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				if chdirErr := os.Chdir(oldWd); chdirErr != nil {
					t.Logf("Warning: failed to restore working directory: %v", chdirErr)
				}
			}()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			// Inject FS into context
			mem := billy.NewInMemoryFS()
			ctx := ifs.With(context.Background(), mem)

			// Setup in-memory project structure
			setupTestProject(t, mem, repoRoot, taskID, tt.currentPhase, tt.stepsCompleted)

			// Override findRoot to return our synthetic repo root
			origFindRoot := findRoot
			defer func() { findRoot = origFindRoot }()
			findRoot = func(fs vfs.Filesystem) (string, error) {
				return repoRoot, nil
			}

			// Build command tree with proper context
			cmd := &cobra.Command{Use: "forge-ai"}
			cmd.AddCommand(taskCmd)
			cmd.SetContext(ctx)
			taskCmd.SetContext(ctx)
			taskPhaseCmd.SetContext(ctx)
			taskPhaseNextCmd.SetContext(ctx)

			// Override findActiveTask for test
			origFindActiveTask := findActiveTask
			defer func() { findActiveTask = origFindActiveTask }()
			switch tt.activeTask {
			case "":
				findActiveTask = func(_ context.Context, _ vfs.Filesystem, _ string) (string, error) {
					return "", fmt.Errorf("no active task set in project")
				}
			case "999-nonexistent-task":
				findActiveTask = func(_ context.Context, _ vfs.Filesystem, _ string) (string, error) {
					return "", fmt.Errorf("active task directory does not exist: %s", tt.activeTask)
				}
			default:
				findActiveTask = func(_ context.Context, _ vfs.Filesystem, _ string) (string, error) {
					return tt.activeTask, nil
				}
			}

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs([]string{"task", "phase", "next"})

			err = cmd.Execute()

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)

				// Verify task.yaml was updated
				taskYAMLPath := filepath.Join(repoRoot, ".forge", "ai", "tasks", taskID, "task.yaml")
				content, err := mem.ReadFile(taskYAMLPath)
				require.NoError(t, err)

				var task map[string]interface{}
				err = yaml.Unmarshal(content, &task)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedPhase, task["current_phase"])

				// Verify project.yaml was updated
				projectYAMLPath := filepath.Join(repoRoot, ".forge", "ai", "project.yaml")
				content, err = mem.ReadFile(projectYAMLPath)
				require.NoError(t, err)

				var project map[string]interface{}
				err = yaml.Unmarshal(content, &project)
				require.NoError(t, err)

				tasks := project["tasks"].(map[string]interface{})
				taskEntry := tasks[taskID].(map[string]interface{})
				assert.Equal(t, tt.expectedPhase, taskEntry["phase"])
			}
		})
	}
}

// setupTestProject creates a minimal test project with the specified task state
func setupTestProject(t *testing.T, mem vfs.Filesystem, repoRoot, taskID, currentPhase string, stepsCompleted bool) {
	// Write CUE schemas for validation
	writeSchemas(t, mem)

	// Create basic directory structure
	require.NoError(t, mem.MkdirAll(filepath.Join(repoRoot, ".git"), 0o755))
	require.NoError(t, mem.MkdirAll(filepath.Join(repoRoot, ".forge", "ai"), 0o755))
	require.NoError(t, mem.MkdirAll(filepath.Join(repoRoot, ".forge", "ai", "tasks", taskID), 0o755))
	require.NoError(t, mem.MkdirAll(filepath.Join(repoRoot, "tasks", taskID), 0o755))

	// Create .git directory at repoRoot to satisfy findRoot
	require.NoError(t, mem.MkdirAll(filepath.Join(repoRoot, ".git"), 0o755))

	// Create minimal project.yaml
	projectYAML := map[string]interface{}{
		"projectName": "Test Project",
		"template": map[string]interface{}{
			"source":  "test-template",
			"version": "v1.0.0",
		},
		"activeTask": taskID,
		"tasks": map[string]interface{}{
			taskID: map[string]interface{}{
				"path":             fmt.Sprintf("tasks/%s", taskID),
				"phase":            currentPhase,
				"status":           "active",
				"template_version": "v1",
				"created_at":       time.Now().Format(time.RFC3339),
			},
		},
	}

	projectContent, err := yaml.Marshal(projectYAML)
	require.NoError(t, err)
	require.NoError(t, mem.WriteFile(filepath.Join(repoRoot, ".forge", "ai", "project.yaml"), projectContent, 0o644))

	// Create task.yaml with appropriate phase and steps
	stepStatus := "completed"
	if !stepsCompleted {
		stepStatus = "pending"
	}

	taskYAML := map[string]interface{}{
		"id":            taskID,
		"title":         "Test Task",
		"current_phase": currentPhase,
		"status":        "active",
		"phases": map[string]interface{}{
			"planning": map[string]interface{}{
				"status":     "completed",
				"modifiable": false,
				"steps": []map[string]interface{}{
					{
						"id":               "step-1",
						"description":      "Test planning step",
						"ai_function":      "DISCOVER",
						"status":           stepStatus,
						"success_criteria": []string{"Criteria 1"},
					},
				},
			},
			"implementation": map[string]interface{}{
				"status":     "pending",
				"modifiable": true,
				"steps": []map[string]interface{}{
					{
						"id":               "step-2",
						"description":      "Test implementation step",
						"ai_function":      "EXECUTE",
						"status":           stepStatus,
						"success_criteria": []string{"Criteria 2"},
					},
				},
			},
			"validation": map[string]interface{}{
				"status":     "pending",
				"modifiable": false,
				"steps": []map[string]interface{}{
					{
						"id":               "step-3",
						"description":      "Test validation step",
						"ai_function":      "VALIDATE",
						"status":           stepStatus,
						"success_criteria": []string{"Criteria 3"},
					},
				},
			},
		},
	}

	taskContent, err := yaml.Marshal(taskYAML)
	require.NoError(t, err)
	require.NoError(
		t,
		mem.WriteFile(filepath.Join(repoRoot, ".forge", "ai", "tasks", taskID, "task.yaml"), taskContent, 0o644),
	)
}

func TestFindActiveTaskFromProject(t *testing.T) {
	tests := []struct {
		name          string
		activeTask    string
		taskExists    bool
		wantError     bool
		errorContains string
	}{
		{
			name:       "valid active task",
			activeTask: "001-test-task",
			taskExists: true,
			wantError:  false,
		},
		{
			name:          "no active task set",
			activeTask:    "",
			wantError:     true,
			errorContains: "no active task set",
		},
		{
			name:          "task directory does not exist",
			activeTask:    "999-nonexistent-task",
			taskExists:    false,
			wantError:     true,
			errorContains: "active task directory does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoRoot := "/repo"

			mem := billy.NewInMemoryFS()
			ctx := ifs.With(context.Background(), mem)

			// Write CUE schemas for validation
			writeSchemas(t, mem)

			// Setup in-memory project structure
			require.NoError(t, mem.MkdirAll(filepath.Join(repoRoot, ".forge", "ai"), 0o755))

			// Create project.yaml
			projectYAML := map[string]interface{}{
				"projectName": "Test Project",
				"template": map[string]interface{}{
					"source":  "test-template",
					"version": "v1.0.0",
				},
				"activeTask": tt.activeTask,
			}

			if tt.activeTask != "" {
				projectYAML["tasks"] = map[string]interface{}{
					tt.activeTask: map[string]interface{}{
						"path":             fmt.Sprintf("tasks/%s", tt.activeTask),
						"phase":            "planning",
						"status":           "active",
						"template_version": "v1",
						"created_at":       time.Now().Format(time.RFC3339),
					},
				}
			}

			projectContent, err := yaml.Marshal(projectYAML)
			require.NoError(t, err)
			require.NoError(
				t,
				mem.WriteFile(filepath.Join(repoRoot, ".forge", "ai", "project.yaml"), projectContent, 0o644),
			)

			// Create task directory if needed
			if tt.taskExists && tt.activeTask != "" {
				require.NoError(t, mem.MkdirAll(filepath.Join(repoRoot, "tasks", tt.activeTask), 0o755))
			}

			// Test the ProjectState.GetActiveTask method directly
			projectState, err := state.NewProjectState(ctx, mem, repoRoot)
			require.NoError(t, err)

			taskID, err := projectState.GetActiveTask(ctx)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.activeTask, taskID)
			}
		})
	}
}

func TestTaskPhaseNextIntegration(t *testing.T) {
	// This is an integration test that exercises the full workflow
	repoRoot := "/repo"
	taskID := "001-integration-test"

	// Setup in-memory filesystem
	mem := billy.NewInMemoryFS()
	ctx := ifs.With(context.Background(), mem)

	// Write CUE schemas for validation
	writeSchemas(t, mem)

	// Setup complete project structure
	require.NoError(t, mem.MkdirAll(filepath.Join(repoRoot, ".git"), 0o755))
	require.NoError(t, mem.MkdirAll(filepath.Join(repoRoot, ".forge", "ai"), 0o755))
	require.NoError(t, mem.MkdirAll(filepath.Join(repoRoot, ".forge", "ai", "tasks", taskID), 0o755))
	require.NoError(t, mem.MkdirAll(filepath.Join(repoRoot, "tasks", taskID), 0o755))

	// Create initial project state
	projectYAML := map[string]interface{}{
		"projectName": "Integration Test Project",
		"template": map[string]interface{}{
			"source":  "test-template",
			"version": "v1.0.0",
		},
		"activeTask": taskID,
		"tasks": map[string]interface{}{
			taskID: map[string]interface{}{
				"path":             fmt.Sprintf("tasks/%s", taskID),
				"phase":            "planning",
				"status":           "active",
				"template_version": "v1",
				"created_at":       time.Now().Format(time.RFC3339),
			},
		},
	}

	projectContent, err := yaml.Marshal(projectYAML)
	require.NoError(t, err)
	require.NoError(t, mem.WriteFile(filepath.Join(repoRoot, ".forge", "ai", "project.yaml"), projectContent, 0o644))

	// Create initial task state with completed planning steps
	taskYAML := map[string]interface{}{
		"id":            taskID,
		"title":         "Integration Test Task",
		"current_phase": "planning",
		"status":        "active",
		"phases": map[string]interface{}{
			"planning": map[string]interface{}{
				"status":     "completed",
				"modifiable": false,
				"steps": []map[string]interface{}{
					{
						"id":               "discover-requirements",
						"description":      "Discover project requirements",
						"ai_function":      "DISCOVER",
						"status":           "completed",
						"success_criteria": []string{"Requirements documented"},
					},
					{
						"id":               "plan-implementation",
						"description":      "Plan implementation approach",
						"ai_function":      "PLAN",
						"status":           "completed",
						"success_criteria": []string{"Implementation plan created"},
					},
				},
			},
			"implementation": map[string]interface{}{
				"status":     "pending",
				"modifiable": true,
				"steps":      []map[string]interface{}{},
			},
			"validation": map[string]interface{}{
				"status":     "pending",
				"modifiable": false,
				"steps":      []map[string]interface{}{},
			},
		},
	}

	taskContent, err := yaml.Marshal(taskYAML)
	require.NoError(t, err)
	require.NoError(
		t,
		mem.WriteFile(filepath.Join(repoRoot, ".forge", "ai", "tasks", taskID, "task.yaml"), taskContent, 0o644),
	)

	// Test phase transition from planning to implementation
	taskState, err := state.NewTaskState(ctx, mem, repoRoot, taskID)
	require.NoError(t, err)

	// Verify initial state
	assert.Equal(t, "planning", taskState.Value().Current_phase)

	// Transition to implementation phase
	err = taskState.TransitionPhase(ctx, "implementation")
	require.NoError(t, err)

	// Save the updated state
	err = taskState.Save(ctx)
	require.NoError(t, err)

	// Verify task state was updated
	updatedTaskState, err := state.NewTaskState(ctx, mem, repoRoot, taskID)
	require.NoError(t, err)
	assert.Equal(t, "implementation", updatedTaskState.Value().Current_phase)

	// Verify project state was updated
	projectState, err := state.NewProjectState(ctx, mem, repoRoot)
	require.NoError(t, err)

	taskEntry := projectState.Value().Tasks[taskID]
	assert.Equal(t, "implementation", taskEntry.Phase)

	t.Log("✓ Integration test passed: successful phase transition from planning to implementation")
}

// writeSchemas writes minimal CUE schemas required for validation into the
// in-memory filesystem at internal/schemas/.
func writeSchemas(t *testing.T, fs vfs.Filesystem) {
	t.Helper()
	// Ensure schema directory exists
	require.NoError(t, fs.MkdirAll(filepath.Join("internal", "schemas"), 0o755))

	projectCue := []byte(`package schemas

// Exported alias for Go type generation
Project: #Project

#Project: {
	projectName: string | *""
	template: {
		source:  string & !=""
		version: string & !=""
	}
	status: "active" | "maintenance" | "archived" | *"active"
	activeTask?: string & !~"\\s"
	tasks?: [string]: {
		path:             string & !~"\\s"
		phase:            "planning" | "implementation" | "validation"
		status:           "active" | "blocked" | "completed"
		template_version: string & !=""
		created_at?:      string
		completed_at?:    string
	}
}
`)
	require.NoError(t, fs.WriteFile(filepath.Join("internal", "schemas", "project.cue"), projectCue, 0o644))

	taskCue := []byte(`package schemas

#Step: {
	id:          string & !~"\\s" & !=""
	description: string & !=""
	ai_function: "DISCOVER" | "PLAN" | "ASSESS" | "EXECUTE" | "VALIDATE" | "REPORT"
	status:      "pending" | "in-progress" | "completed" | *"pending"
	success_criteria: [...string & !=""] | *[]
}

#Phase: {
	status:      "active" | "pending" | "completed" | *"pending"
	modifiable?: bool | *false
	steps: [...#Step] | *[]
}

// Exported alias for Go type generation
Task: #Task

#Task: {
	id:            string & !=""
	title:         string & !=""
	created_at?:   string
	current_phase: "planning" | "implementation" | "validation"
	status:        "active" | "blocked" | "completed" | *"active"
	phases: {
		planning:       #Phase
		implementation: #Phase
		validation:     #Phase
	}
}
`)
	require.NoError(t, fs.WriteFile(filepath.Join("internal", "schemas", "task.cue"), taskCue, 0o644))
}
