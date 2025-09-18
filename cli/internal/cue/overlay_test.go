package cue

import (
	"path/filepath"
	"testing"

	"github.com/input-output-hk/catalyst-forge-libs/fs/billy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildOverlay_EmptyDir(t *testing.T) {
	mem := billy.NewInMemoryFS()

	overlay, dir, err := BuildOverlay(mem, filepath.Join("internal", "schemas"))
	require.NoError(t, err)
	assert.NotNil(t, overlay)
	assert.Equal(t, filepath.Join("/overlay", "internal", "schemas"), dir)
	assert.Len(t, overlay, 0)
}

func TestBuildOverlay_SingleCueFile(t *testing.T) {
	mem := billy.NewInMemoryFS()
	root := filepath.Join("internal", "schemas")
	require.NoError(t, mem.MkdirAll(root, 0o755))

	content := []byte("package schemas\nX: 1\n")
	require.NoError(t, mem.WriteFile(filepath.Join(root, "x.cue"), content, 0o644))

	overlay, dir, err := BuildOverlay(mem, root)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("/overlay", root), dir)
	assert.Len(t, overlay, 1)
	_, ok := overlay[filepath.Join("/overlay", root, "x.cue")]
	assert.True(t, ok)
}

func TestBuildOverlay_MultipleCueFiles_NonCueIgnored(t *testing.T) {
	mem := billy.NewInMemoryFS()
	root := filepath.Join("internal", "schemas")
	require.NoError(t, mem.MkdirAll(filepath.Join(root, "sub"), 0o755))

	// Two cue files and one non-cue file
	require.NoError(t, mem.WriteFile(filepath.Join(root, "a.cue"), []byte("package a\n"), 0o644))
	require.NoError(t, mem.WriteFile(filepath.Join(root, "b.txt"), []byte("ignore"), 0o644))
	require.NoError(t, mem.WriteFile(filepath.Join(root, "sub", "c.cue"), []byte("package sub\n"), 0o644))

	overlay, dir, err := BuildOverlay(mem, root)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("/overlay", root), dir)
	assert.Len(t, overlay, 2)
	assert.Contains(t, overlay, filepath.Join("/overlay", root, "a.cue"))
	assert.Contains(t, overlay, filepath.Join("/overlay", root, "sub", "c.cue"))
}
