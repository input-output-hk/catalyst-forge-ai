# OCI Bundle Distribution Module - Implementation Guide

## Module Overview

A focused Go module for distributing file bundles as OCI artifacts. The module provides a simple API for pushing directories to and pulling archives from OCI registries using ORAS.

**Module Path**: `github.com/yourdomain/ocibundle`

## Architecture

### Core Design Principles

1. **Single Responsibility**: Push and pull file bundles via OCI registries
2. **Secure by Default**: Prevent common vulnerabilities without configuration
3. **Extensible Archiving**: Start with tar.gz, allow future formats via interfaces
4. **Flexible Authentication**: Support multiple auth mechanisms
5. **Clear Boundaries**: No business logic, just OCI transport

### Package Structure

```
ocibundle/
├── client.go          # Main client interface and implementation
├── archive.go         # Archive interface and tar.gz implementation
├── auth.go            # Authentication providers
├── security.go        # Security validators and constraints
├── options.go         # Functional options for configuration
├── errors.go          # Domain-specific error types
└── internal/
    ├── oras/          # ORAS wrapper (isolate dependency)
    └── validate/      # Path and content validators
```

## Core Interfaces

### Archive Strategy (Port/Adapter Pattern)

```go
// Archiver handles compression/decompression of file bundles
type Archiver interface {
    // Archive creates an archive from a directory
    Archive(ctx context.Context, sourceDir string, output io.Writer) error

    // Extract expands an archive to a directory
    Extract(ctx context.Context, input io.Reader, targetDir string, opts ExtractOptions) error

    // MediaType returns the OCI media type for this archive format
    MediaType() string
}

// ExtractOptions controls extraction behavior
type ExtractOptions struct {
    MaxFiles      int    // Maximum number of files (0 = unlimited)
    MaxSize       int64  // Maximum total uncompressed size
    MaxFileSize   int64  // Maximum individual file size
    StripPrefix   string // Remove prefix from paths during extraction
    PreservePerms bool   // Preserve file permissions
}
```

### Client Interface

```go
// Client provides OCI bundle operations
type Client struct {
    registry  RegistryClient  // Internal ORAS wrapper
    archiver  Archiver       // Archive strategy
    auth      Authenticator  // Auth provider
    validator Validator      // Security validator
}

// Push uploads a directory to an OCI registry
func (c *Client) Push(ctx context.Context, sourceDir, reference string, opts ...PushOption) error

// Pull downloads and extracts an OCI artifact
func (c *Client) Pull(ctx context.Context, reference, targetDir string, opts ...PullOption) error
```

### Authentication

```go
// Authenticator provides registry credentials
type Authenticator interface {
    // Credentials returns auth for a given registry
    Credentials(ctx context.Context, registry string) (AuthConfig, error)
}

// Built-in implementations:
// - DockerConfigAuthenticator (reads ~/.docker/config.json)
// - StaticAuthenticator (programmatic credentials)
// - ChainAuthenticator (try multiple in sequence)
```

## Public API Design

### Client Creation

```go
// New creates a client with sensible defaults
client, err := ocibundle.New()

// NewWithOptions allows customization
client, err := ocibundle.NewWithOptions(
    ocibundle.WithArchiver(ocibundle.NewTarGzArchiver()),
    ocibundle.WithAuth(ocibundle.DockerConfigAuth()),
    ocibundle.WithArtifactType("application/vnd.myapp.bundle"),
)
```

### Push Operation

```go
// Simple push with defaults
err := client.Push(ctx, "./my-files", "ghcr.io/org/bundle:v1.0.0")

// Push with options
err := client.Push(ctx, "./my-files", "ghcr.io/org/bundle:v1.0.0",
    ocibundle.WithAnnotations(map[string]string{
        "version": "1.0.0",
        "author": "team",
    }),
    ocibundle.WithPlatform("linux/amd64"),
)
```

### Pull Operation

```go
// Simple pull
err := client.Pull(ctx, "ghcr.io/org/bundle:v1.0.0", "./output")

// Pull with constraints
err := client.Pull(ctx, "ghcr.io/org/bundle:v1.0.0", "./output",
    ocibundle.WithMaxSize(100 * 1024 * 1024),  // 100MB max
    ocibundle.WithMaxFiles(1000),
)
```

### Programmatic Authentication

```go
// Static credentials
client, _ := ocibundle.NewWithOptions(
    ocibundle.WithAuth(ocibundle.StaticAuth(
        "ghcr.io", "username", "password",
    )),
)

// Multiple auth sources
client, _ := ocibundle.NewWithOptions(
    ocibundle.WithAuth(ocibundle.ChainAuth(
        ocibundle.EnvAuth(),           // Check environment vars
        ocibundle.DockerConfigAuth(),  // Fall back to Docker config
    )),
)
```

## Security Implementation

### Validator Chain

```go
// Validator checks for security issues
type Validator interface {
    ValidatePath(path string) error
    ValidateFile(info FileInfo) error
    ValidateArchive(stats ArchiveStats) error
}

// Default validator chain includes:
// - PathTraversalValidator (no .., absolute paths, symlinks outside root)
// - SizeValidator (configurable limits)
// - FileCountValidator (prevent zip bombs)
// - PermissionSanitizer (remove setuid/setgid bits)
```

### Security Defaults

```go
// DefaultExtractOptions provides safe defaults
var DefaultExtractOptions = ExtractOptions{
    MaxFiles:      10000,              // Prevent file count attacks
    MaxSize:       1 * 1024 * 1024 * 1024,  // 1GB total
    MaxFileSize:   100 * 1024 * 1024,       // 100MB per file
    PreservePerms: false,              // Sanitize permissions
}
```

### Path Security

The module should reject:
- Paths containing `..`
- Absolute paths
- Symlinks pointing outside the archive root
- Hidden files starting with `.` (configurable)
- Files with problematic names (NUL bytes, control characters)

## Error Handling

```go
// Typed errors for different failure modes
var (
    ErrAuthenticationFailed = errors.New("authentication failed")
    ErrRegistryUnreachable  = errors.New("registry unreachable")
    ErrInvalidReference     = errors.New("invalid OCI reference")
    ErrSecurityViolation    = errors.New("security constraint violated")
    ErrArchiveCorrupted     = errors.New("archive corrupted or invalid")
)

// Detailed error with context
type BundleError struct {
    Op        string  // Operation that failed
    Reference string  // OCI reference
    Err       error   // Underlying error
}
```

## Extension Points

### Custom Archive Formats

To add new archive formats:

1. Implement the `Archiver` interface
2. Register media type mapping
3. Pass to client via `WithArchiver()`

```go
// Example: Adding zip support
type ZipArchiver struct{}

func (z *ZipArchiver) MediaType() string {
    return "application/zip"
}

// Use it
client, _ := ocibundle.NewWithOptions(
    ocibundle.WithArchiver(&ZipArchiver{}),
)
```

### Custom Validators

```go
// Add business-specific validation
type LicenseValidator struct{}

func (v *LicenseValidator) ValidateFile(info FileInfo) error {
    if info.Name == "LICENSE" {
        return nil
    }
    // Require LICENSE file
    return errors.New("missing LICENSE file")
}
```

## Implementation Notes

### ORAS Integration

- Wrap ORAS in an internal package to isolate the dependency
- Use ORAS v2 Go library (not CLI)
- Handle ORAS-specific errors and convert to domain errors

### Streaming vs Memory

- Use streaming for large files (don't load entire archive into memory)
- Implement chunked reading/writing for push/pull operations
- Consider using temporary files for very large operations

### Registry Compatibility

- Test against major registries (Docker Hub, GHCR, GCR, ECR, ACR)
- Handle registry-specific quirks in the internal ORAS wrapper
- Support both Docker Registry HTTP API V2 and OCI Distribution Spec

### Concurrency

- Client should be safe for concurrent use
- Use context for cancellation
- No global state

### Testing Strategy

- Unit tests with mock registry
- Integration tests against local registry (using testcontainers)
- Security tests with malicious archives
- Benchmark tests for large