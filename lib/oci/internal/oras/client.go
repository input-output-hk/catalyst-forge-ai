// Package oras provides ORAS wrapper functionality.
// This isolates the ORAS dependency in an internal package.
package oras

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// DefaultORASClient implements ORASClient using the real ORAS library.
type DefaultORASClient struct{}

// Ensure DefaultORASClient implements ORASClient.
// var _ ORASClient = (*DefaultORASClient)(nil) // Verified in interfaces.go

// Push pushes an artifact to an OCI registry using the real ORAS library.
func (c *DefaultORASClient) Push(ctx context.Context, reference string, descriptor *PushDescriptor, opts *AuthOptions) error {
	return Push(ctx, reference, descriptor, opts)
}

// Pull pulls an artifact from an OCI registry using the real ORAS library.
func (c *DefaultORASClient) Pull(ctx context.Context, reference string, opts *AuthOptions) (*PullDescriptor, error) {
	return Pull(ctx, reference, opts)
}

// AuthConfig represents authentication configuration for ORAS operations.
// This matches the public AuthConfig struct for consistency.
type AuthConfig struct {
	Username string
	Password string
}

// CredentialFunc is an alias for ORAS's credential function type.
// It provides credentials for a given registry (host:port).
type CredentialFunc = auth.CredentialFunc

// HTTPConfig contains configuration for HTTP transport settings.
type HTTPConfig struct {
	// AllowHTTP enables HTTP instead of HTTPS for registry connections.
	AllowHTTP bool

	// AllowInsecure allows connections with self-signed certificates.
	AllowInsecure bool

	// Registries specifies which registries this applies to.
	// If empty, applies to all registries.
	Registries []string
}

// AuthOptions configures authentication and HTTP settings for ORAS operations.
type AuthOptions struct {
	// StaticAuth provides static credentials for a specific registry.
	// If set, this overrides the default Docker credential chain for that registry.
	StaticRegistry string
	StaticUsername string
	StaticPassword string

	// CredentialFunc provides a custom credential callback.
	// If set, this completely overrides the default credential chain.
	CredentialFunc CredentialFunc

	// HTTPConfig controls HTTP vs HTTPS and certificate validation.
	HTTPConfig *HTTPConfig
}

// NewRepository creates a new ORAS repository with authentication configured.
// It sets up the default Docker credential chain and applies any auth overrides.
//
// Parameters:
//   - ctx: Context for the operation
//   - reference: Full OCI reference (e.g., "ghcr.io/org/repo:tag")
//   - opts: Authentication options (can be nil for default behavior)
//
// Returns:
//   - Configured remote repository ready for ORAS operations
//   - Error if repository creation fails
//
// Authentication behavior:
//  1. If CredentialFunc is provided, it takes complete precedence
//  2. If StaticAuth is provided, it overrides credentials for that specific registry
//  3. Otherwise, uses ORAS's default Docker credential chain (config + helpers)
//
// This isolates ORAS authentication logic and provides clean injection points
// for testing and programmatic credential management.
func NewRepository(ctx context.Context, reference string, opts *AuthOptions) (*remote.Repository, error) {
	// Parse the reference to obtain only the repository path (without tag/digest)
	repoPath, _, _ := splitReference(reference)
	if repoPath == "" {
		return nil, fmt.Errorf("invalid reference: %s", reference)
	}

	repo, err := remote.NewRepository(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// Default: use ORAS's default Docker credential chain (config + helpers)
	authClient := auth.DefaultClient

	// Configure HTTP transport based on explicit HTTP configuration
	if opts != nil && opts.HTTPConfig != nil && shouldApplyHTTPConfig(reference, opts.HTTPConfig) {
		// Create HTTP transport with specified settings
		transport := &http.Transport{}

		// Configure TLS settings
		if opts.HTTPConfig.AllowInsecure {
			transport.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true, // Allow self-signed certificates
			}
		}

		// Use HTTP scheme if explicitly requested
		if opts.HTTPConfig.AllowHTTP {
			repo.PlainHTTP = true
		}

		// Set custom transport
		if authClient.Client == nil {
			authClient.Client = &http.Client{Transport: transport}
		} else {
			authClient.Client.Transport = transport
		}
	}

	// Apply auth overrides if provided
	if opts != nil {
		if opts.CredentialFunc != nil {
			// Custom credential function takes complete precedence
			authClient.Credential = opts.CredentialFunc
		} else if opts.StaticRegistry != "" && opts.StaticUsername != "" {
			// Static auth override for specific registry
			authClient.Credential = auth.StaticCredential(opts.StaticRegistry, auth.Credential{
				Username: opts.StaticUsername,
				Password: opts.StaticPassword,
			})
		}
	}

	repo.Client = authClient

	return repo, nil
}

// shouldApplyHTTPConfig determines if HTTP configuration should be applied to a registry.
// It checks if the registry matches any of the configured registries or if no specific
// registries are configured (applies to all).
func shouldApplyHTTPConfig(reference string, config *HTTPConfig) bool {
	// If no specific registries are configured, apply to all
	if len(config.Registries) == 0 {
		return true
	}

	// Parse the registry hostname from the reference (format: registry/path:tag)
	parts := strings.Split(reference, "/")
	if len(parts) == 0 {
		return false
	}

	registry := parts[0]

	// Check if the registry matches any configured registry
	for _, configuredRegistry := range config.Registries {
		if registry == configuredRegistry {
			return true
		}
		// Also check if it matches as a hostname (without port)
		if strings.Contains(configuredRegistry, ":") {
			// configuredRegistry has port, check exact match
			if registry == configuredRegistry {
				return true
			}
		} else {
			// configuredRegistry is hostname only, check hostname match
			if strings.HasPrefix(registry, configuredRegistry+":") {
				return true
			}
		}
	}

	return false
}

// PushDescriptor describes the content to be pushed to an OCI registry.
// It contains the media type and the data stream for the artifact.
type PushDescriptor struct {
	MediaType   string
	Data        io.Reader
	Size        int64
	Annotations map[string]string
	Platform    string
}

// Push pushes an artifact to an OCI registry using ORAS.
// It pushes the content directly using ORAS TagBytes function.
//
// Parameters:
//   - ctx: Context for the operation
//   - reference: Full OCI reference (e.g., "ghcr.io/org/repo:tag")
//   - descriptor: Description of the content to push
//   - opts: Authentication options (can be nil for default behavior)
//
// Returns an error if the push operation fails.
func Push(ctx context.Context, reference string, descriptor *PushDescriptor, opts *AuthOptions) error {
	if descriptor == nil {
		return fmt.Errorf("descriptor cannot be nil")
	}

	// Read the data to get the actual content
	data, err := io.ReadAll(descriptor.Data)
	if err != nil {
		return mapORASError("push", reference, fmt.Errorf("failed to read data: %w", err))
	}

	// Ensure we have content to push
	if len(data) == 0 {
		return mapORASError("push", reference, fmt.Errorf("no data to push"))
	}

	// Create the repository with authentication
	repo, err := NewRepository(ctx, reference, opts)
	if err != nil {
		return mapORASError("push", reference, fmt.Errorf("failed to create repository: %w", err))
	}

	// Extract tag or digest from reference
	_, refPart, _ := splitReference(reference)
	if refPart == "" {
		return mapORASError("push", reference, fmt.Errorf("reference must include a tag or digest"))
	}

	// 1) Push the content blob
	blobDesc, err := oras.PushBytes(ctx, repo, descriptor.MediaType, data)
	if err != nil {
		return mapORASError("push", reference, fmt.Errorf("push blob: %w", err))
	}

	// 2) Pack an OCI 1.1 manifest with artifactType and empty config
	packOpts := oras.PackManifestOptions{Layers: []ocispec.Descriptor{blobDesc}}
	artifactType := "application/vnd.catalyst.bundle.v1"
	manDesc, err := oras.PackManifest(ctx, repo, oras.PackManifestVersion1_1, artifactType, packOpts)
	if err != nil {
		return mapORASError("push", reference, fmt.Errorf("pack manifest v1.1: %w", err))
	}

	// 3) Tag the manifest with the requested ref
	_, err = oras.Tag(ctx, repo, manDesc.Digest.String(), refPart)
	if err != nil {
		return mapORASError("push", reference, fmt.Errorf("tag manifest: %w", err))
	}

	return nil
}

// PullDescriptor describes the content pulled from an OCI registry.
// It contains metadata about the pulled artifact.
type PullDescriptor struct {
	MediaType string
	Data      io.ReadCloser
	Size      int64
}

// Pull pulls an artifact from an OCI registry using ORAS.
// It retrieves the content and returns it as a descriptor with a reader.
//
// Parameters:
//   - ctx: Context for the operation
//   - reference: Full OCI reference (e.g., "ghcr.io/org/repo:tag")
//   - opts: Authentication options (can be nil for default behavior)
//
// Returns the pulled descriptor and an error if the pull operation fails.
func Pull(ctx context.Context, reference string, opts *AuthOptions) (*PullDescriptor, error) {
	// Create the repository with authentication
	repo, err := NewRepository(ctx, reference, opts)
	if err != nil {
		return nil, mapORASError("pull", reference, fmt.Errorf("failed to create repository: %w", err))
	}

	// Extract tag or digest from reference
	_, refPart, _ := splitReference(reference)
	if refPart == "" {
		return nil, mapORASError("pull", reference, fmt.Errorf("reference must include a tag or digest"))
	}

	// Pull the target using ORAS Fetch function
	desc, reader, err := oras.Fetch(ctx, repo, refPart, oras.DefaultFetchOptions)
	if err != nil {
		return nil, mapORASError("pull", reference, err)
	}

	// If the reference resolves to a manifest, fetch the first content entry (layer/blob)
	if desc.MediaType == ocispec.MediaTypeImageManifest {
		manifestBytes, err := io.ReadAll(reader)
		if err != nil {
			return nil, mapORASError("pull", reference, fmt.Errorf("read manifest: %w", err))
		}
		reader.Close()

		// Try image manifest first
		var imgMan ocispec.Manifest
		if err := json.Unmarshal(manifestBytes, &imgMan); err == nil && (len(imgMan.Layers) > 0 || imgMan.Config.MediaType != "") {
			if len(imgMan.Layers) == 0 {
				return nil, mapORASError("pull", reference, fmt.Errorf("no layers in image manifest"))
			}
			layerDesc := imgMan.Layers[0]
			layerReader, err := repo.Blobs().Fetch(ctx, layerDesc)
			if err != nil {
				return nil, mapORASError("pull", reference, fmt.Errorf("fetch layer: %w", err))
			}
			return &PullDescriptor{MediaType: layerDesc.MediaType, Data: layerReader, Size: layerDesc.Size}, nil
		}

		return nil, mapORASError("pull", reference, fmt.Errorf("unrecognized manifest format"))
	}

	// Otherwise, the fetched target is the content itself
	return &PullDescriptor{MediaType: desc.MediaType, Data: reader, Size: desc.Size}, nil
}

// splitReference splits a full OCI reference into repository path and reference part (tag or digest).
// Examples:
//
//	localhost:5000/myrepo:latest -> ("localhost:5000/myrepo", "latest", false)
//	ghcr.io/org/name@sha256:abcd -> ("ghcr.io/org/name", "sha256:abcd", true)
func splitReference(full string) (repoPath string, refPart string, isDigest bool) {
	if full == "" {
		return "", "", false
	}
	// Find last slash to isolate the repo name tail
	lastSlash := strings.LastIndex(full, "/")
	if lastSlash == -1 {
		// No slash; cannot reliably parse; return as-is
		return full, "", false
	}
	head := full[:lastSlash]
	tail := full[lastSlash+1:]

	if at := strings.LastIndex(tail, "@"); at != -1 {
		// Digest form name@digest
		return head + "/" + tail[:at], tail[at+1:], true
	}
	if colon := strings.LastIndex(tail, ":"); colon != -1 {
		// Tag form name:tag (safe because we looked only in tail, avoiding port)
		return head + "/" + tail[:colon], tail[colon+1:], false
	}
	// No tag/digest found
	return full, "", false
}

// mapORASError maps ORAS errors to domain-specific errors.
// It analyzes the error type and returns appropriate domain errors.
//
// Parameters:
//   - op: The operation that failed ("push" or "pull")
//   - ref: The OCI reference being processed
//   - err: The original ORAS error
//
// Returns a domain error with proper context.
func mapORASError(op, ref string, err error) error {
	if err == nil {
		return nil
	}

	// Check for authentication errors
	if errors.Is(err, auth.ErrBasicCredentialNotFound) {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Check for network/registry unreachable errors
	if errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, context.Canceled) {
		return fmt.Errorf("registry unreachable: %w", err)
	}

	// For other errors, create a detailed error message
	return fmt.Errorf("%s %s: %w", op, ref, err)
}
