package state

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/input-output-hk/catalyst-forge-libs/fs/billy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskState_LoadValidateSave(t *testing.T) {
	mem := billy.NewInMemoryFS()
	writeSchemas(t, mem)

	// Create task file structure
	id := "001-sample"
	base := filepath.Join(".", ".forge", "ai", "tasks", id)
	require.NoError(t, mem.MkdirAll(base, 0o755))

	// Minimal valid task YAML
	taskYAML := []byte("" +
		"id: \"001-sample\"\n" +
		"title: \"Sample Task\"\n" +
		"current_phase: \"planning\"\n" +
		"status: \"active\"\n" +
		"phases:\n" +
		"  planning:\n" +
		"    status: \"pending\"\n" +
		"    modifiable: true\n" +
		"    steps: []\n" +
		"  implementation:\n" +
		"    status: \"pending\"\n" +
		"    steps: []\n" +
		"  validation:\n" +
		"    status: \"pending\"\n" +
		"    steps: []\n",
	)
	require.NoError(t, mem.WriteFile(filepath.Join(base, "task.yaml"), taskYAML, 0o644))

	ctx := context.Background()
	ts, err := NewTaskState(ctx, mem, ".", id)
	require.NoError(t, err)
	assert.Equal(t, "001-sample", ts.Value().Id)

	// Validate and save
	require.NoError(t, ts.Validate(ctx))
	require.NoError(t, ts.Save(ctx))

	// Expect a backup file to exist
	entries, err := mem.ReadDir(filepath.Join(".forge", "ai", "backups"))
	require.NoError(t, err)
	assert.NotEmpty(t, entries)
}
