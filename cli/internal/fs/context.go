// Package fs provides filesystem context utilities for managing filesystem instances
// in Go contexts, allowing dependency injection of different filesystem implementations.
package fs

import (
	"context"

	"github.com/input-output-hk/catalyst-forge-libs/fs"
)

type ctxKey string

const key ctxKey = "forge.fs"

// With returns a new context containing the provided filesystem.
func With(ctx context.Context, fs fs.Filesystem) context.Context {
	return context.WithValue(ctx, key, fs)
}

// From returns the filesystem from context, or nil if none is set.
// Note: Returns interface to allow different filesystem implementations.
//
//nolint:ireturn // intentionally returns interface for dependency injection
func From(ctx context.Context) fs.Filesystem {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(key); v != nil {
		if fs, ok := v.(fs.Filesystem); ok && fs != nil {
			return fs
		}
	}

	return nil
}
