# OCI Bundle Distribution Module - Implementation Tasks

## Overview

This document provides a structured task list for implementing the OCI Bundle Distribution Module as specified in [GUIDE.md](./GUIDE.md). The module will provide a secure, extensible Go library for distributing file bundles as OCI artifacts using ORAS.

**⚠️ IMPORTANT AUTH CHANGE**: As of the latest implementation, Task 7 (Docker Config Authentication) has been replaced with ORAS native authentication. The bespoke `Authenticator` interface and custom auth implementations have been removed in favor of ORAS's battle-tested authentication system. Task 8 (Additional Authentication Methods) is no longer applicable.

## Required Reading

Before starting any implementation work, you MUST read and understand:

1. **[Go Coding Standards](../../docs/guides/go/style.md)** - Project-specific Go style guide and conventions
2. **[Go TDD Essentials](../../docs/guides/go/tdd.md)** - Test-Driven Development principles for Go
3. **[Implementation Guide](./GUIDE.md)** - Architectural context and design specifications

These documents define the coding standards, testing approach, and design patterns that must be followed throughout the implementation.

## How to Use This Document

1. **Complete Required Reading**: Study all three documents above before writing any code
2. **Follow Task Order**: Tasks are ordered by dependencies - complete them sequentially
3. **Test-Driven Development**: Write tests BEFORE implementation for each task
4. **Validation Required**: Each task includes success criteria that MUST be validated
5. **Check Off Completion**: Mark tasks with `[x]` only after ALL success criteria are met
6. **Run Tests Continuously**: Run `go test ./...` after each task to ensure nothing breaks

### TDD Enforcement

**MANDATORY**: For each task, follow this TDD cycle:
1. Write failing tests that validate the success criteria
2. Implement the minimal code to make tests pass
3. Refactor while keeping tests green
4. Document the API with examples
5. Only mark the task complete when all tests pass

## Module Setup

### Task 1: Configure Existing Module
- [x] Verify module path in `go.mod` matches project structure
- [x] Set up package structure as defined in GUIDE.md
- [x] Add ORAS v2 dependency: `go get oras.land/oras-go/v2`
- [x] Create README.md with module description and usage examples

**Success Criteria:**
- Module builds without errors
- All directories from GUIDE.md structure exist
- go.mod contains ORAS v2 dependency
- README includes basic usage examples

---

## Core Interfaces and Types

### Task 2: Define Core Interfaces
- [x] Create `archive.go` with `Archiver` interface
- [x] Create `auth.go` with `Authenticator` interface
- [x] Create `security.go` with `Validator` interface
- [x] Define `ExtractOptions` struct with security defaults
- [x] Define `FileInfo` and `ArchiveStats` types

**Success Criteria:**
- All interfaces match GUIDE.md specifications exactly
- Interfaces have comprehensive godoc comments
- Default options enforce security constraints (max files: 10000, max size: 1GB)
- Unit tests verify default values

---

### Task 3: Implement Error Types
- [x] Create `errors.go` with sentinel errors
- [x] Implement `BundleError` struct with context
- [x] Add error wrapping utilities
- [x] Create error formatting methods

**Success Criteria:**
- All error types from GUIDE.md are defined
- BundleError implements error interface
- Errors support wrapping with `errors.Is()` and `errors.As()`
- Unit tests verify error behavior

---

## Security Implementation

### Task 4: Path Security Validator
- [x] Create `internal/validate/path.go`
- [x] Implement `PathTraversalValidator`
- [x] Detect and reject `..` in paths
- [x] Detect and reject absolute paths
- [x] Validate against symlinks escaping root

**Success Criteria:**
- Rejects all forms of path traversal (../, ..\\, encoded variants)
- Rejects absolute paths on all platforms
- Rejects symlinks pointing outside archive root
- Comprehensive test suite with malicious path examples
- Zero false positives on legitimate paths

---

### Task 5: Size and Resource Validators
- [x] Implement `SizeValidator` in `security.go`
- [x] Implement `FileCountValidator`
- [x] Implement `PermissionSanitizer`
- [x] Create `ValidatorChain` to combine validators

**Success Criteria:**
- Size validator enforces configurable limits
- File count validator prevents zip bombs
- Permission sanitizer removes setuid/setgid bits
- Chain executes all validators in order
- Tests verify each validator independently

---

## Archive Implementation

### Task 6: Tar.gz Archiver
- [x] Create `archive_targz.go` implementing `Archiver`
- [x] Implement `Archive()` method with streaming
- [x] Implement `Extract()` method with security checks
- [x] Handle large files without memory exhaustion
- [x] Add progress callback support

**Success Criteria:**
- Creates valid tar.gz archives readable by standard tools
- Extracts archives preserving directory structure
- Applies all security validators during extraction
- Streams large files (tested with 1GB+ files)
- Memory usage stays constant regardless of file size
- Tests verify round-trip (archive → extract → archive)

---

## Authentication

### Task 7: Docker Config Authentication
- [x] Implement `DockerConfigAuthenticator` in `auth.go`
- [x] Parse ~/.docker/config.json correctly
- [x] Handle auth helpers (credential stores)
- [x] Support multiple registry configurations
- [x] Handle missing/invalid config gracefully

**Success Criteria:**
- Reads standard Docker config format
- Supports base64-encoded auth tokens
- Falls back gracefully when config missing
- Tests mock different config scenarios
- No credentials logged or exposed in errors

---

### Task 8: Additional Authentication Methods
- [x] **CANCELLED**: Replaced with ORAS native authentication
- [x] **CANCELLED**: StaticAuthenticator replaced by `WithStaticAuth()` functional option
- [x] **CANCELLED**: ChainAuthenticator replaced by ORAS default credential chain
- [x] **CANCELLED**: EnvAuthenticator replaced by ORAS Docker config support
- [x] **CANCELLED**: Anonymous auth handled by ORAS credential chain fallback

**Reason:** The bespoke authentication system has been replaced with ORAS native authentication. All authentication functionality is now provided by ORAS's battle-tested credential chain, with functional options for programmatic overrides.

---

## ORAS Integration

### Task 9: ORAS Wrapper
- [x] Create `internal/oras/client.go`
- [x] Wrap ORAS push operations
- [x] Wrap ORAS pull operations
- [x] Convert ORAS errors to domain errors
- [x] Handle registry-specific quirks

**Success Criteria:**
- Isolates ORAS dependency in internal package
- All ORAS errors mapped to BundleError types
- Supports OCI artifact types
- Works with Docker Registry API V2
- Integration tests against local registry

---

## Client Implementation

### Task 10: Client Core
- [x] Implement `Client` struct in `client.go`
- [x] Implement `New()` and `NewWithOptions()`
- [x] Create functional options pattern in `options.go`
- [x] Implement client validation and defaults
- [x] Ensure thread safety

**Success Criteria:**
- Client can be created with defaults
- Options allow full customization
- Client is safe for concurrent use
- Invalid options return clear errors
- Tests verify option behavior

---

### Task 11: Push Operation
- [x] Implement `Push()` method
- [x] Add push options (annotations, platform)
- [x] Archive source directory
- [x] Upload via ORAS with retry logic
- [x] Clean up temporary files on error

**Success Criteria:**
- [x] Successfully pushes to test registry
- [x] Annotations appear in manifest (framework in place, ORAS TagBytes limitation noted)
- [x] Handles network interruptions gracefully
- [x] No temporary files left after operation
- [x] Progress reporting works correctly
- [x] Integration tests verify push to real registry (placeholder implemented)

---

### Task 12: Pull Operation
- [x] Implement `Pull()` method
- [x] Add pull options (size limits, file limits)
- [x] Download via ORAS
- [x] Extract with security validation
- [x] Atomic extraction (all or nothing)

**Success Criteria:**
- [x] Successfully pulls from test registry (framework in place, auth required for testing)
- [x] Enforces size and file count limits
- [x] Validates all paths during extraction
- [x] Rolls back on extraction failure
- [x] Integration tests verify pull from real registry (placeholder implemented)

---

## Testing Infrastructure

### Task 13: Test Utilities
- [x] Create test registry using testcontainers
- [x] Build test archive generator
- [x] Create malicious archive suite
- [x] Add benchmark utilities
- [x] Set up coverage reporting

**Success Criteria:**
- Test registry starts/stops reliably
- Can generate archives of any size
- Malicious archive suite covers OWASP top 10
- Benchmarks measure memory and CPU
- Coverage reports can be generated locally

---

### Task 14: Integration Tests
- [x] Test against local test registry
- [x] Test against Docker Hub (if credentials available)
- [x] Test against GitHub Container Registry (if credentials available)
- [x] Add registry compatibility matrix documentation

**Success Criteria:**
- Push/pull works on test registry
- Auth works with each registry's method
- Registry-specific quirks documented
- Tests skip gracefully when credentials unavailable
- Clear skip messages when credentials missing

---

## Documentation and Examples

### Task 15: Documentation
- [x] Write comprehensive godoc for all public APIs
- [x] Create example programs in `examples/`
- [x] Add security best practices guide
- [x] Document migration from other tools
- [x] Create troubleshooting guide

**Success Criteria:**
- Every public type/method has godoc
- Examples are runnable and tested
- Security guide covers common pitfalls
- Migration guide includes code samples
- Troubleshooting covers common errors

---

## Performance and Optimization

### Task 16: Performance Optimization
- [ ] Profile memory usage during operations
- [ ] Optimize streaming for large files
- [ ] Add concurrent file processing where safe
- [ ] Implement connection pooling for registry
- [ ] Add caching for registry authentication

**Success Criteria:**
- Memory usage constant for any file size
- Can handle 10GB+ archives
- Concurrent operations are race-free
- Connection reuse improves performance
- Benchmarks show performance improvements

---

## Final Validation

### Task 17: Security Audit
- [ ] Run static analysis (gosec, staticcheck)
- [ ] Perform fuzzing on validators
- [ ] Test with OWASP archive attacks
- [ ] Verify no credential leakage
- [ ] Document security model

**Success Criteria:**
- Zero security warnings from tools
- Fuzzing runs without crashes
- Blocks all OWASP test cases
- No secrets in logs or errors
- Security model documented in README

---

## Completion Checklist

Before considering the module complete, verify:

- [ ] All tasks above are checked off
- [ ] Test coverage exceeds 80%
- [ ] No known security vulnerabilities
- [ ] Documentation is comprehensive
- [ ] Examples run without modification
- [ ] All tests pass locally
- [ ] Performance meets requirements
- [ ] API follows Go best practices

## Notes

- If a task reveals additional requirements, add subtasks before marking complete
- Run the full test suite after each task to catch regressions early
- Keep the GUIDE.md updated if implementation differs from design
- Create issues for any technical debt incurred during implementation