package mcpserver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewServer_CreatesInstance(t *testing.T) {
	s := NewServer()
	require.NotNil(t, s)
}

// For Task 3.2 we will add tests for the next tool registration and behavior.
// Here we only assert that the server can run stdio with a cancellable context
// without panicking when immediately cancelled.
func TestRunStdio_CancelImmediately(t *testing.T) {
	// Running STDIO in tests will block, so ensure we cancel immediately.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// Should return promptly without error; underlying SDK may return nil or context error
	_ = RunStdio(ctx, nil)
}
