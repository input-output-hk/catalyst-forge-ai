# Forge AI MVP Implementation Tasks

## Project Structure

**IMPORTANT**: The CLI implementation should be placed in a `/cli` directory from the repository root:
- Module path: `github.com/input-output-hk/catalyst-forge-ai/cli`
- All Go source files will be under `/cli`
- The existing `lib/oci` package will be imported from the parent module

## Required Reading Before Starting

**IMPORTANT**: Before beginning any task, you MUST read and understand these foundational documents:

### Internal Documentation
1. **[@docs/CONSTITUTION.md](docs/CONSTITUTION.md)** - The non-negotiable rules that govern all development. Keep this in context at all times.
2. **[@docs/guides/go/style.md](docs/guides/go/style.md)** - Strict Go coding standards that MUST be followed without exception.
3. **[@docs/guides/go/cli.md](docs/guides/go/cli.md)** - Required patterns for CLI development with Cobra and Viper.
4. **[@docs/guides/go/tdd.md](docs/guides/go/tdd.md)** - Test-Driven Development practices that MUST be followed for all code.

### External Documentation
5. **[MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)** - The official Go SDK for Model Context Protocol implementation.
6. **[MCP Go SDK Documentation](https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk)** - API documentation for the MCP Go SDK.
7. **[Cobra CLI Framework](https://cobra.dev/)** - Documentation for the Cobra CLI framework used throughout this project.

## Development Methodology: Test-Driven Development (TDD)

**MANDATORY**: All code MUST be developed using Test-Driven Development as outlined in [@docs/guides/go/tdd.md](docs/guides/go/tdd.md).

### TDD Process for Each Task:
1. **RED**: Write a failing test for the next small piece of behavior
2. **GREEN**: Write the simplest code to make the test pass
3. **REFACTOR**: Clean up the code while keeping tests green

### Key TDD Requirements:
- Write tests BEFORE implementation code
- Work in tiny Red-Green-Refactor cycles
- Test behavior, not implementation details
- Keep tests deterministic and hermetic (no real I/O in unit tests)
- Use table-driven tests for multiple scenarios
- Run with Go race detector for any concurrent code
- Maintain >80% test coverage for critical paths

## Key Dependencies

The following Go packages will be used throughout the implementation:

```go
// CLI Framework
github.com/spf13/cobra      // CLI framework
github.com/spf13/viper      // Configuration management

// MCP Implementation
github.com/modelcontextprotocol/go-sdk  // MCP server implementation

// OCI/Container Registry
oras.land/oras-go/v2        // OCI artifact management (embedded in lib/oci)

// Validation
cuelang.org/go/cue          // CUE schema validation

// Standard Libraries
gopkg.in/yaml.v3            // YAML parsing
```

## Task Completion Process

To mark a task complete:
1. Verify ALL tests were written BEFORE implementation (TDD)
2. Verify ALL success criteria are met
3. Run all lints and ensure they pass: `golangci-lint run ./...`
4. Run all tests and ensure they pass: `go test ./...`
5. Run with race detector: `go test -race ./...`
6. Check test coverage meets requirements: `go test -cover ./...`
7. Check the markdown checkbox: `[x]`
8. Create a git commit with descriptive message
9. Move to the next uncompleted task

## Section 1: Template Filesystem and Distribution

### Task 1.1: Create Minimal Template Source Directory
- [x] **Description**: Set up the initial template source directory structure with minimal viable files as specified in [@MVP.md](MVP.md#11-implement-template-filesystem-preparation).

**Implementation Details**:
- Create `template_source/` directory with the exact structure from MVP.md
- Implement `manifest.yml` with name and version fields (refer to [@docs/DISTRIBUTION.md](docs/DISTRIBUTION.md#manifest-specification))
- Create AI Function files (*.ai.md) with standard preamble and one-sentence instructions (refer to [@docs/FUNCTIONS.md](docs/FUNCTIONS.md#standard-structure))
- Implement minimal CUE schemas for `project.cue` and `task.cue` (refer to [@docs/STATE.md](docs/STATE.md#schema-organization))
- Create `default.yaml` task template with three phases as per [@docs/ARCHITECTURE.md](docs/ARCHITECTURE.md#the-three-phase-model)

**Success Criteria**:
- [x] Directory structure matches MVP specification exactly
- [x] All required files exist with minimal but valid content
- [x] CUE schemas validate successfully using `cue vet`
- [x] Task template contains planning, implementation, and validation phases

### Task 1.2: Implement OCI Module Integration
- [x] **Description**: Set up the Go module structure in the `/cli` directory and integrate the existing `lib/oci` package for OCI operations.

**Implementation Details**:
- Create `/cli` directory from repository root
- Initialize Go module: `cd cli && go mod init github.com/input-output-hk/catalyst-forge-ai/cli`
- Verify `lib/oci` package exists and contains `ocibundle` client
- Add required dependencies for ORAS Go library
- Create basic test to verify OCI client can be instantiated

**Success Criteria**:
- [x] `/cli` directory created with Go module initialized
- [x] Go module named `github.com/input-output-hk/catalyst-forge-ai/cli`
- [x] `lib/oci` package imports successfully
- [x] ORAS dependencies resolved in go.mod
- [x] Basic instantiation test passes

### Task 1.3: Implement the publish Command
- [x] **Description**: Create the `forge-ai publish` command following Cobra patterns from [@docs/guides/go/cli.md](docs/guides/go/cli.md).

**Implementation Details**:
- Create `cli/cmd/publish.go` following the exact patterns in the CLI guide
- Implement flags: `--source` and `--registry` using Viper binding
- Use `lib/oci` client's `Push(ctx, sourceDir, reference)` method (import from parent module)
- Follow error handling patterns from [@docs/guides/go/style.md](docs/guides/go/style.md#error-handling)
- Add validation for source directory existence and registry format

**Success Criteria**:
- [x] Command exists at `forge-ai publish`
- [x] Flags properly bound to Viper
- [x] Successfully pushes template to OCI registry when valid inputs provided
- [x] Appropriate error messages for invalid inputs
- [x] Follows all patterns from CLI guide (no logic in Run, uses RunE, etc.)

## Section 2: Project and Task Initialization

### Task 2.1: Implement Root Command Structure
- [x] **Description**: Set up the root command and global configuration following [@docs/guides/go/cli.md](docs/guides/go/cli.md#root-command-setup).

**Implementation Details**:
- Create `cli/cmd/root.go` with exported `Execute()` function
- Create minimal `cli/main.go` that only calls `cmd.Execute()`
- Set up `cobra.OnInitialize` for config loading
- Implement global flags (--config)
- Follow the exact patterns from the CLI guide

**Success Criteria**:
- [x] `main.go` contains only the minimal code shown in guide
- [x] Root command properly exports Execute()
- [x] Global configuration initialization in place
- [x] Command runs without errors: `go run . --help`

### Task 2.2: Implement the init Command
- [ ] **Description**: Create the `forge-ai init <project-name> --template=<oci-ref>` command as specified in [@MVP.md](MVP.md#21-implement-the-init-command).

**Implementation Details**:
- Create `cli/cmd/init.go` following Cobra patterns
- Use positional argument validation with `cobra.ExactArgs(1)`
- Implement `--template` flag with OCI reference validation
- Use `lib/oci` client's `Pull(ctx, reference, targetDir)` method
- Create `.forge/ai/` directory structure
- Generate initial `project.yaml` following schema from [@docs/STATE.md](docs/STATE.md#project-state-projectyaml)

**Success Criteria**:
- [ ] Command validates exactly one project name argument
- [ ] Successfully pulls template from OCI registry
- [ ] Creates correct directory structure under `.forge/ai/`
- [ ] Generates valid `project.yaml` that passes CUE validation
- [ ] Handles network and authentication errors gracefully

### Task 2.3: Implement the task new Command
- [ ] **Description**: Create the `forge-ai task new --title="<title>"` command as specified in [@MVP.md](MVP.md#22-implement-the-task-new-command).

**Implementation Details**:
- Create `cli/cmd/task.go` as parent command (no Run function)
- Create `cli/cmd/task_new.go` for the subcommand
- Generate unique task ID (format: `NNN-slug` where NNN is sequential)
- Create task directory: `tasks/<task-id>/`
- Copy `templates/tasks/default.yaml` to `tasks/<task-id>/task.yaml`
- Update task.yaml with ID, title, and set `current_phase: planning`
- Register task in project.yaml

**Success Criteria**:
- [ ] Command creates unique, sequential task IDs
- [ ] Task directory created with correct structure
- [ ] Task.yaml contains all required fields per schema
- [ ] Task registered in project.yaml
- [ ] Initial phase set to "planning"

## Section 3: MCP Server and Task Lifecycle

### Task 3.1: Implement Basic MCP Server
- [ ] **Description**: Create the `forge-ai mcp serve` command that starts a STDIO-based JSON-RPC server as specified in [@docs/ARCHITECTURE.md](docs/ARCHITECTURE.md#mcp-communication-protocol).

**Implementation Details**:
- Create `cli/cmd/mcp.go` as parent command
- Create `cli/cmd/mcp_serve.go` for the serve subcommand
- Use the [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) for server implementation
- Implement STDIO transport as per MCP SDK examples
- Handle `tools/list` method to return available tools
- Implement basic request/response logging for debugging
- Follow MCP protocol examples from ARCHITECTURE.md and SDK documentation

**Success Criteria**:
- [ ] Server starts and listens on STDIO
- [ ] Responds to `tools/list` with tool catalog
- [ ] Handles malformed JSON gracefully
- [ ] Implements proper JSON-RPC error responses
- [ ] Can be tested with manual JSON input

### Task 3.2: Implement the next Tool
- [ ] **Description**: Implement the core `next` tool that returns AI Functions with injected context as specified in [@MVP.md](MVP.md#31-implement-the-mcp-server-and-agent-interaction-loop).

**Implementation Details**:
- Add `next` tool to MCP server's tool catalog using MCP Go SDK tool registration
- Read active task's `task.yaml` state file
- Identify current phase and first incomplete step
- Load corresponding AI Function from `.forge/ai/functions/`
- Inject context (current state, relevant task.yaml sections)
- Return composed AI Function via MCP SDK's response format

**Success Criteria**:
- [ ] Tool appears in `tools/list` response
- [ ] Correctly identifies next incomplete step
- [ ] Loads and returns appropriate AI Function
- [ ] Context properly injected into function template
- [ ] Returns phase transition message when phase complete

### Task 3.3: Implement step_add Tool for Planning Phase
- [ ] **Description**: Implement the `step_add` tool that allows adding steps during planning as specified in [@docs/ARCHITECTURE.md](docs/ARCHITECTURE.md#planning-tools).

**Implementation Details**:
- Add `step_add` to MCP server's tool catalog using SDK tool registration
- Validate current phase is "planning"
- Validate target phase has `modifiable: true`
- Prevent adding steps to current phase
- Append new step to target phase's steps array
- Validate step structure against CUE schema
- Update task.yaml atomically
- Return success/error via MCP SDK response format

**Success Criteria**:
- [ ] Tool only works during planning phase
- [ ] Correctly validates modifiable phases
- [ ] Steps added have required fields (id, description, ai_function, success_criteria)
- [ ] Task.yaml updates pass CUE validation
- [ ] Appropriate errors for invalid operations

### Task 3.4: Implement Human Gate Commands
- [ ] **Description**: Implement `forge-ai task phase next` command for phase transitions as specified in [@MVP.md](MVP.md#32-implement-phase-1-planning).

**Implementation Details**:
- Create `cli/cmd/task_phase.go` for phase management
- Implement validation for phase transition prerequisites
- For planning→implementation: verify planning steps complete
- For implementation→validation: verify all steps complete or abandoned
- Update `current_phase` in task.yaml
- Log phase transitions

**Success Criteria**:
- [ ] Command validates phase transition prerequisites
- [ ] Updates current_phase correctly
- [ ] Prevents invalid transitions
- [ ] Provides clear error messages for blocked transitions
- [ ] Maintains phase consistency in task.yaml

### Task 3.5: Implement artifact_save and step_complete Tools
- [ ] **Description**: Implement tools for saving artifacts and marking steps complete as specified in [@MVP.md](MVP.md#33-implement-phase-2-implementation).

**Implementation Details**:
- Add both tools to MCP server catalog using SDK tool registration
- `artifact_save`: Record file path in task.yaml's step artifacts array
- `step_complete`:
  - Require `evidence` array parameter
  - Update step status to "completed"
  - Set completed_at timestamp
  - Validate evidence is non-empty
- Ensure atomic updates to task.yaml

**Success Criteria**:
- [ ] Both tools appear in tool catalog
- [ ] artifact_save records paths correctly
- [ ] step_complete requires and stores evidence
- [ ] Status transitions follow state machine rules
- [ ] Timestamps recorded accurately

### Task 3.6: Implement task complete Command
- [ ] **Description**: Implement the final `forge-ai task complete` command as specified in [@MVP.md](MVP.md#34-implement-phase-3-validation).

**Implementation Details**:
- Create command in `cli/cmd/task_complete.go`
- Validate all validation phase steps are complete
- Update task status to "completed" in task.yaml
- Update task status in project.yaml registry
- Set completed_at timestamp
- Provide completion summary

**Success Criteria**:
- [ ] Command validates validation phase completion
- [ ] Updates task status in both task.yaml and project.yaml
- [ ] Records completion timestamp
- [ ] Prevents completion with incomplete validation
- [ ] Shows meaningful completion message

## Section 4: State Management and Validation

### Task 4.1: Implement CUE Schema Validation
- [ ] **Description**: Set up CUE validation for all state files using the CUE Go API as specified in [@docs/STATE.md](docs/STATE.md#cue-based-validation).

**Implementation Details**:
- Add CUE Go library dependency
- Create validation functions for each schema type
- Implement temp file write → validate → atomic rename pattern
- Generate Go structs from CUE at build time with `cue get go`
- Add Makefile target for schema→struct generation

**Success Criteria**:
- [ ] CUE Go API integrated successfully
- [ ] Validation functions work for all state file types
- [ ] Atomic file updates prevent corruption
- [ ] Build process generates Go structs from CUE
- [ ] Validation errors include clear CUE constraint messages

### Task 4.2: Implement State File Readers and Writers
- [ ] **Description**: Create safe readers and writers for state files following patterns from [@docs/STATE.md](docs/STATE.md#validation-framework).

**Implementation Details**:
- Create `cli/internal/state` package
- Implement readers that unmarshal YAML to generated structs
- Implement writers that validate→write atomically
- Add file locking to prevent concurrent access
- Create backup before each write operation

**Success Criteria**:
- [ ] State package follows internal package patterns
- [ ] All writes validated against CUE schemas
- [ ] Atomic writes prevent partial updates
- [ ] Backups created before modifications
- [ ] File locking prevents race conditions

## Section 5: Integration Testing

### Task 5.1: Create End-to-End Test for MVP Flow
- [ ] **Description**: Implement integration test that exercises the complete MVP workflow.

**Test Flow**:
1. Publish template to local registry
2. Initialize project with template
3. Create new task
4. Start MCP server
5. Simulate agent calling `next` for each phase
6. Add steps via `step_add`
7. Complete steps with evidence
8. Transition through phases
9. Complete task

**Success Criteria**:
- [ ] Test covers all three phases
- [ ] Validates state at each step
- [ ] Checks file system changes
- [ ] Verifies tool responses
- [ ] Passes consistently

### Task 5.2: Add Unit Tests for Critical Components
- [ ] **Description**: Add comprehensive unit tests following Go testing patterns from [@docs/guides/go/style.md](docs/guides/go/style.md#testing-standards).

**Components to Test**:
- State validation functions
- MCP tool handlers
- Phase transition logic
- Task ID generation
- OCI operations (with mocks)

**Success Criteria**:
- [ ] Test coverage > 80% for critical paths
- [ ] Table-driven tests for multiple scenarios
- [ ] Mocks generated with moq where appropriate
- [ ] Tests follow naming convention from style guide
- [ ] All tests pass: `go test ./...`

## Section 6: Documentation and Polish

### Task 6.1: Add Command Help Documentation
- [ ] **Description**: Ensure all commands have comprehensive help text following Cobra best practices.

**Implementation Details**:
- Add Short and Long descriptions to all commands
- Document all flags with clear descriptions
- Add usage examples in Long description
- Ensure consistent formatting across commands

**Success Criteria**:
- [ ] Every command has Short and Long descriptions
- [ ] Flag help text is clear and complete
- [ ] Examples provided for complex commands
- [ ] `forge-ai --help` shows organized command tree

### Task 6.2: Create MVP Demo Script
- [ ] **Description**: Create a demonstration script that showcases the MVP functionality.

**Script Contents**:
- Publish a template
- Initialize a project
- Create a task
- Show MCP interaction for planning phase
- Demonstrate phase transitions
- Complete the task lifecycle

**Success Criteria**:
- [ ] Script runs without errors
- [ ] Demonstrates all MVP features
- [ ] Includes explanatory comments
- [ ] Can be used for testing/validation

## Completion Checklist

Before declaring MVP complete, verify:

- [ ] All tasks above are marked complete
- [ ] Code follows [@docs/CONSTITUTION.md](docs/CONSTITUTION.md) principles
- [ ] Go code adheres to [@docs/guides/go/style.md](docs/guides/go/style.md)
- [ ] CLI structure matches [@docs/guides/go/cli.md](docs/guides/go/cli.md)
- [ ] All tests pass: `go test ./...`
- [ ] No linting errors: `golangci-lint run`
- [ ] Demo script runs successfully end-to-end
- [ ] State files validate against CUE schemas
- [ ] MCP protocol implementation is complete