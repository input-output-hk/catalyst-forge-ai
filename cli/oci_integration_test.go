package main

import (
	"context"
	"testing"

	oci "github.com/input-output-hk/catalyst-forge-ai/lib/oci"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOCIClientInstantiation(t *testing.T) {
	// Test that we can instantiate the OCI client from lib/oci
	client, err := oci.New()
	require.NoError(t, err, "Failed to create OCI client")
	assert.NotNil(t, client, "OCI client should not be nil")
}

func TestOCIClientPushMethod(t *testing.T) {
	// Test that the Push method exists and has the expected signature
	client, err := oci.New()
	require.NoError(t, err)

	// Verify Push method exists by attempting to call it with nil context
	// This will fail, but it proves the method exists with the expected signature
	ctx := context.Background()
	err = client.Push(ctx, "", "")
	assert.Error(t, err, "Push with empty params should fail")
}

func TestOCIClientPullMethod(t *testing.T) {
	// Test that the Pull method exists and has the expected signature
	client, err := oci.New()
	require.NoError(t, err)

	// Verify Pull method exists by attempting to call it with nil context
	// This will fail, but it proves the method exists with the expected signature
	ctx := context.Background()
	err = client.Pull(ctx, "", "")
	assert.Error(t, err, "Pull with empty params should fail")
}
