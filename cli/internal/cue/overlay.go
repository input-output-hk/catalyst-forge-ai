// Package cue provides helpers for working with CUE in this project.
package cue

import (
	"fmt"
	"os"
	"path/filepath"

	"cuelang.org/go/cue/load"
	"github.com/input-output-hk/catalyst-forge-libs/fs"
)

// BuildOverlay walks the provided rootDir within the supplied filesystem and
// constructs a CUE overlay suitable for use with load.Config. All .cue files
// beneath rootDir are included, mapped to an absolute synthetic path under
// /overlay.
func BuildOverlay(filesystem fs.Filesystem, rootDir string) (map[string]load.Source, string, error) {
	overlay := map[string]load.Source{}

	// Ensure root exists (ignore errors here; Walk will return if missing)
	_ = filesystem.MkdirAll(rootDir, 0o755)

	walkErr := filesystem.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".cue" {
			return nil
		}
		b, readErr := filesystem.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("failed to read CUE file %s: %w", path, readErr)
		}
		overlay[filepath.Join("/overlay", path)] = load.FromBytes(b)
		return nil
	})
	if walkErr != nil {
		return nil, "", fmt.Errorf("failed to walk directory %s: %w", rootDir, walkErr)
	}

	dir := filepath.Join("/overlay", rootDir)
	return overlay, dir, nil
}
