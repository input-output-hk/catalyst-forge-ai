package cue

import (
	"fmt"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/load"
	"github.com/input-output-hk/catalyst-forge-libs/fs"
)

// BuildSchemaValue composes the CUE overlay from the given filesystem and root
// directory and returns a compiled cue.Value using the provided context.
// The returned value represents the built CUE instance containing the schemas.
func BuildSchemaValue(ctx *cue.Context, filesystem fs.Filesystem, rootDir string) (cue.Value, error) {
	overlay, dir, err := BuildOverlay(filesystem, rootDir)
	if err != nil {
		return cue.Value{}, fmt.Errorf("build overlay: %w", err)
	}

	cfg := &load.Config{Dir: dir, Overlay: overlay}
	insts := load.Instances([]string{"."}, cfg)
	if len(insts) == 0 {
		return cue.Value{}, fmt.Errorf("no CUE instances found for %s", filepath.Join("/overlay", rootDir))
	}

	v := ctx.BuildInstance(insts[0])
	if v.Err() != nil {
		return cue.Value{}, fmt.Errorf("build instance: %w", v.Err())
	}
	return v, nil
}
