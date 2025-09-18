package state

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/input-output-hk/catalyst-forge-libs/fs/billy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTaskState(t *testing.T) (*TaskState, *billy.FS) {
	mem := billy.NewInMemoryFS()
	writeSchemas(t, mem)

	// Minimal valid task
	id := "001-sample"
	base := filepath.Join(".", ".forge", "ai", "tasks", id)
	require.NoError(t, mem.MkdirAll(base, 0o755))
	content := []byte("" +
		"id: \"001-sample\"\n" +
		"title: \"Task\"\n" +
		"current_phase: \"planning\"\n" +
		"status: \"active\"\n" +
		"phases:\n" +
		"  planning:\n" +
		"    status: \"pending\"\n" +
		"    modifiable: false\n" +
		"    steps: []\n" +
		"  implementation:\n" +
		"    status: \"pending\"\n" +
		"    modifiable: true\n" +
		"    steps: []\n" +
		"  validation:\n" +
		"    status: \"pending\"\n" +
		"    steps: []\n",
	)
	require.NoError(t, mem.WriteFile(filepath.Join(base, "task.yaml"), content, 0o644))

	ts, err := NewTaskState(context.Background(), mem, ".", id)
	require.NoError(t, err)
	return ts, mem
}

func TestTaskState_AddStep_Start_Complete_Transition(t *testing.T) {
	ts, _ := setupTaskState(t)

	// Add a step to implementation during planning
	require.NoError(t, ts.AddStep(context.Background(), "implementation", StepInput{
		ID:          "s1",
		Description: "Do work",
		AIFunction:  "EXECUTE",
	}))

	// Transition planning -> implementation (planning has no steps, should be allowed)
	require.NoError(t, ts.TransitionPhase(context.Background(), "implementation"))
	assert.Equal(t, "implementation", ts.Value().Current_phase)

	// Start it in implementation
	require.NoError(t, ts.StepStart(context.Background(), "s1"))

	// Complete it
	require.NoError(t, ts.StepComplete(context.Background(), "s1"))

	// Transition implementation -> validation now that steps are completed
	require.NoError(t, ts.TransitionPhase(context.Background(), "validation"))
	assert.Equal(t, "validation", ts.Value().Current_phase)
}
