package state

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/input-output-hk/catalyst-forge-libs/fs/billy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupProjectState(t *testing.T) (*ProjectState, *billy.FS) {
	mem := billy.NewInMemoryFS()
	writeSchemas(t, mem)
	// minimal valid project
	require.NoError(t, mem.MkdirAll(filepath.Join(".", ".forge", "ai"), 0o755))
	projectYAML := []byte("" +
		"projectName: \"Test\"\n" +
		"template:\n" +
		"  source: \"ghcr.io/example/template\"\n" +
		"  version: \"v1\"\n" +
		"status: \"active\"\n",
	)
	require.NoError(t, mem.WriteFile(filepath.Join(".", ".forge", "ai", "project.yaml"), projectYAML, 0o644))
	ps, err := NewProjectState(context.Background(), mem, ".")
	require.NoError(t, err)
	return ps, mem
}

func TestProjectState_SetActiveTask(t *testing.T) {
	ps, _ := setupProjectState(t)
	require.NoError(t, ps.SetActiveTask(context.Background(), "001-task"))
	assert.Equal(t, "001-task", ps.Value().ActiveTask)
}

func TestProjectState_RegisterAndUpdateTask(t *testing.T) {
	ps, _ := setupProjectState(t)
	in := ProjectTaskEntryInput{
		Path:            "tasks/001-task",
		Phase:           "planning",
		Status:          "active",
		TemplateVersion: "v1",
		CreatedAt:       "2025-01-01T00:00:00Z",
	}
	require.NoError(t, ps.RegisterTask(context.Background(), "001-task", &in))
	assert.Contains(t, ps.Value().Tasks, "001-task")

	require.NoError(t, ps.UpdateTaskPhase(context.Background(), "001-task", "implementation"))
	assert.Equal(t, "implementation", ps.Value().Tasks["001-task"].Phase)

	require.NoError(t, ps.UpdateTaskStatus(context.Background(), "001-task", "blocked"))
	assert.Equal(t, "blocked", ps.Value().Tasks["001-task"].Status)
}
