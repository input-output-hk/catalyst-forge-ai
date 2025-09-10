# Forge AI System Design

## Overview

Forge AI is a structured workflow system that guides AI agents through software development tasks. It enforces a consistent three-phase process while allowing flexibility within each phase to adapt to different task requirements. The system uses Model Context Protocol (MCP) to create an "AI API" that manages all interactions between LLM agents and the Forge system.

## Core Architecture

### Three Foundational Components

1. **CLI Tool (`forge-ai`)** - State machine API that manages task lifecycle, phase transitions, and serves as MCP server
2. **MCP Integration** - STDIO-based protocol enabling any MCP-compatible agent to interact with Forge
3. **AI Functions** - Focused prompts delivered via MCP that guide agents through specific activities

### Key Design Principles

- **Universal Structure**: Every task follows the same three phases
- **Plan as Contract**: Planning phase produces a binding plan document
- **Human Gates**: Phase transitions require explicit human approval
- **Agent Autonomy**: Agents work independently within a phase via MCP tools
- **State Protection**: CLI exclusively manages all state transitions - LLMs never directly modify files
- **Programmatic Validation**: CLI enforces all structural rules and business logic in code
 - **CUE Schema + Programmatic Validation**: CLI validates all state against CUE schemas and enforces business logic in code
- **Atomic Operations**: All state changes are validated before execution by the CLI
- **State Persistence**: All work tracked in YAML files managed by CLI
- **Memory System**: Context preserved across sessions
- **Context Injection**: CLI dynamically injects relevant context into AI Functions

## MCP Integration Architecture

### Transport and Protocol
- **Transport**: STDIO (standard input/output) for local agent communication
- **Protocol**: JSON-RPC 2.0 message format
- **Execution**: CLI exposes MCP tools that agents (Claude Code, Cursor, etc.) can invoke
- **Output**: Tools return composed AI Function prompts with injected context

### Architectural Advantage
**Key Innovation**: The CLI is the single authority for state management and validation. All state changes are validated against CUE schemas before writes and applied atomically by the CLI.

**Traditional Approach** (direct file manipulation):
- LLM reads/writes YAML files directly
- Validation happens after modifications (detective control)
- Higher risk of state corruption between validation checks

**Forge MCP Architecture** (schema-validated, CLI-gated):
- CLI exclusively manages all state transitions
- LLM requests changes via MCP tools (no direct state file writes)
- CLI validates proposed changes using the CUE Go API against `.forge/ai/schemas/*.cue`
- Atomic, guaranteed-valid operations

## The Three-Phase Model

Every task progresses through exactly three phases:

### 1. Planning Phase
**Purpose**: Define what needs to be done and how
**AI Functions** (delivered via MCP):
- `DISCOVER.ai.md` - Understand requirements and gather context
- `PLAN.ai.md` - Create structured plan with steps
- `ASSESS.ai.md` - Review step sizing and completeness

**Outputs**:
- `plan.yaml` - Contract defining deliverables, steps, and success criteria
- Discovery notes and design documents
- Memory entries for key decisions

### 2. Implementation Phase
**Purpose**: Execute the plan without deviation
**AI Functions** (delivered via MCP):
- `EXECUTE.ai.md` - Work through plan steps sequentially
- `CHECKPOINT.ai.md` - Save progress and context periodically

**Outputs**:
- All deliverables specified in the plan
- Progress updates in task.yaml (via MCP tools)
- Memory entries for issues and decisions

**Constraint**: Cannot modify plan.yaml during this phase (enforced by CLI)

### 3. Validation Phase
**Purpose**: Verify implementation satisfies the plan
**AI Functions** (delivered via MCP):
- `VALIDATE.ai.md` - Check all success criteria
- `REPORT.ai.md` - Generate comprehensive validation report

**Outputs**:
- Validation report confirming all criteria met
- Final task status update

## MCP Tool Categories

### Workflow Tools (Return AI Functions)
These tools return composed AI Function prompts with injected context:
- `forge_task_work` - Returns EXECUTE.ai.md for current step
- `forge_task_checkpoint` - Returns CHECKPOINT.ai.md with progress
- `forge_task_validate` - Returns VALIDATE.ai.md for validation phase
- `forge_task_plan` - Returns appropriate planning phase function
- `forge_task_discover` - Returns DISCOVER.ai.md for requirements gathering
- `forge_task_assess` - Returns ASSESS.ai.md for plan review

### State Management Tools (Atomic Operations)
These tools modify state with built-in validation logic:
- `forge_step_complete` - Mark step as complete with evidence
- `forge_step_start` - Begin working on a step
- `forge_step_abandon` - Abandon step with documented reason
- `forge_progress_update` - Update progress notes
- `forge_plan_update` - Modify plan (only in planning phase)
- `forge_deliverable_register` - Register completed deliverable

### Memory Tools
- `forge_memory_add` - Create memory entry with tags and context
- `forge_memory_search` - Query memories by tags/keywords
- `forge_memory_relevant` - Get context-appropriate memories for current state

### Navigation Tools
- `forge_next_step` - Get next incomplete step details
- `forge_phase_status` - Check phase and transition readiness
- `forge_task_status` - Get comprehensive current task state
- `forge_plan_view` - View current plan structure
- `forge_task_list` - List all tasks in project

## AI Function Structure

AI Functions are templates that the CLI composes with context:

```markdown
# [FUNCTION_NAME].ai.md

## AI Function Protocol
[Reusable preamble explaining AI Functions and escalation rules]
- You are executing an AI Function within the Forge AI system
- Use MCP tools to interact with the system
- Never attempt to directly read or write YAML files
- Request clarification when requirements are ambiguous

## Context Required
[Specification of what context will be injected by CLI]
- Current phase and step information
- Relevant plan sections
- Task progress
- Memory entries
- File contents as needed

## Instructions
[Function-specific guidance in plain text]

## Expected Output
[What the function should produce via MCP tools]
```

## MCP Communication Protocol

### Tool Discovery
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/list",
  "params": {}
}
```

### Tool Invocation Example
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "forge_task_work",
    "arguments": {}
  }
}
```

### AI Function Delivery Response
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "content": [{
      "type": "text",
      "text": "[Complete AI Function with injected context]"
    }]
  }
}
```

### State Update Example
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "forge_step_complete",
    "arguments": {
      "step_id": "setup-project",
      "evidence": ["go.mod created", "Makefile functional"]
    }
  }
}
```

## State Management

The CLI enforces the structure, validity, and business rules of all state files. All state is validated against CUE schemas (stored in `.forge/ai/schemas/*.cue`) using the CUE Go API before writes. At build time, Go structs are generated from these CUE definitions to ensure code and schemas remain synchronized.

```bash
# Build-time synchronization
cue get go ./.forge/ai/schemas/...   # Generate Go structs from CUE schemas
go generate ./...                 # Additional code generation as needed
go build                          # Compile with generated types
```

### Project State (`.forge/ai/project.yaml`)
CLI-managed with programmatic validation:

```yaml
projectName: "Example Project"
template:
  source: "ghcr.io/catalyst-forge/scaffold"
  version: "v1.0.0"
status: "active"  # active, maintenance, archived
activeTask: "001-cli-design"  # Current CLI context for convenience

tasks:
  "001-cli-design":
    path: "tasks/001-cli-design"
    phase: "implementation"
    status: "active"
    plan_version: "v1"

  "002-fix-bug":
    path: "tasks/002-fix-bug"
    phase: "planning"
    status: "active"
    plan_version: "v1"
```

### Task Plan (`tasks/<task-id>/plan.yaml`)
Created via MCP tools with CLI validation, immutable during implementation:

```yaml
# Produced during planning phase, immutable during implementation
version: "v1"
created_at: "2024-01-15T10:00:00Z"

objectives:
  - "Create CLI tool for managing forge-ai projects"
  - "Support project initialization and task management"

deliverables:
  - id: "cli-binary"
    description: "Compiled Go CLI application"
    path: "dist/forge-ai"
  - id: "user-docs"
    description: "User documentation"
    path: "docs/user-guide.md"

steps:  # Sequential execution in v1
  - id: "setup-project"
    description: "Initialize Go module and project structure"
    success_criteria:
      - "go.mod exists with correct module name"
      - "Basic directory structure created"
      - "Makefile with build targets"

  - id: "implement-init"
    description: "Create init command for new projects"
    success_criteria:
      - "Init command creates valid project structure"
      - "Template files copied correctly"
      - "Validation passes on created files"

  - id: "write-tests"
    description: "Add comprehensive test coverage"
    success_criteria:
      - "Unit tests for all commands"
      - "Coverage exceeds 80%"

success_criteria:  # Overall task success
  - "All deliverables present"
  - "All step criteria satisfied"
  - "Manual testing confirms functionality"
```

### Task State (`tasks/<task-id>/task.yaml`)
Updated exclusively through MCP with CLI validation:

```yaml
# Tracks progress only, references plan.yaml for requirements
id: "001-cli-design"
title: "Design the forge-ai CLI"
current_phase: "implementation"
status: "active"  # active, blocked, completed

phases:
  planning:
    status: "completed"
    completed_at: "2024-01-15T10:00:00Z"

  implementation:
    status: "active"
    started_at: "2024-01-15T14:00:00Z"

  validation:
    status: "pending"

progress:
  steps:
    "setup-project":
      status: "completed"
      completed_at: "2024-01-15T14:30:00Z"
      notes: "Used cobra for command structure"

    "implement-init":
      status: "in-progress"
      started_at: "2024-01-15T15:00:00Z"
      notes: "Working on template copying logic"

    "write-tests":
      status: "pending"
```

## Memory System

### Structure
Memory is managed through MCP tools with CLI-enforced structure:

```
.forge/memory/
└── project/
    ├── index.yaml
    ├── 001-go-over-rust.md
    └── 002-auth-strategy.md

tasks/<task-id>/memory/
├── index.yaml
├── 001-template-issue.md
└── 002-validation-approach.md
```

### Project Memory Index
```yaml
# .forge/ai/memory/project/index.yaml
memories:
  - id: "001-go-over-rust"
    date: "2024-01-10"
    title: "Chose Go over Rust for CLI"
    tags: ["decision", "technology", "cli"]
    summary: "Selected Go for team expertise and compilation speed"

  - id: "002-auth-strategy"
    date: "2024-01-12"
    title: "JWT authentication approach"
    tags: ["decision", "security", "api"]
    summary: "Using JWT with refresh tokens for stateless auth"
```

### Task Memory Index
```yaml
# tasks/001-cli-design/memory/index.yaml
memories:
  - id: "001-template-issue"
    step_id: "implement-init"  # Optional: ties to specific step
    date: "2024-01-15"
    title: "Template copying symlink issue"
    tags: ["issue", "workaround", "cli"]
    summary: "Symlinks in templates cause issues, skipping them"

  - id: "002-validation-approach"
    step_id: "implement-validation"
    date: "2024-01-16"
    title: "Non-blocking validation by default"
    tags: ["decision", "ux", "cli"]
    summary: "Validation warns but doesn't block unless --strict"
```

### Memory Document Format
```markdown
<!-- tasks/001-cli-design/memory/001-template-issue.md -->
# Template Copying Symlink Issue

**Date**: 2024-01-15
**Step**: implement-init

## Context
While implementing template copying for the init command, discovered that symlinks in template directories cause issues when copying across filesystems.

## Problem
- Symlinks may point outside template directory
- Cross-filesystem copies fail with symlinks
- Go's embed doesn't preserve symlinks anyway

## Solution
Skip symlinks during template copying and document this limitation.

## Code Reference
See `internal/template/copy.go:45`
```

## CLI Commands

The CLI serves dual purposes: direct human interaction and MCP server for agents.

### Human Interface Commands

```bash
# Project & Task Management
forge-ai init <project>                    # Initialize new project
forge-ai task new --title=<title>         # Create new task
forge-ai task checkout <task-id>          # Set active task context
forge-ai task list                        # Show all tasks

# Phase Control
forge-ai task phase                       # Show current phase
forge-ai task phase next                  # Request transition to next phase
forge-ai task phase prev --reason=<why>   # Move back to previous phase

# Status & Validation
forge-ai status                           # Project overview
forge-ai task status                      # Current task details
forge-ai validate                         # Run internal consistency checks

# Memory Management
forge-ai memory search --tags=<tags>      # Search memories by tags
forge-ai memory relevant                  # Show memories for current context

# Task Completion
forge-ai task complete                    # Mark task as done
forge-ai task abandon --reason=<why>      # Abandon task
```

### MCP Server Mode

```bash
# Start MCP server for agent integration
forge-ai mcp serve                        # Starts STDIO JSON-RPC server
```

## Phase Transitions

All phase transitions require human approval via CLI commands (not MCP tools):

### Planning → Implementation
- Plan document (plan.yaml) exists and is complete
- All steps have success criteria
- Deliverables clearly specified
- Human reviews and approves via CLI

### Implementation → Validation
- All plan steps marked complete or abandoned
- No work in progress
- Human approves transition via CLI

### Validation → Complete
- All success criteria validated
- Validation report complete
- Human approves completion via CLI

### Backward Transitions
Moving backward is allowed with documented reasons:

#### Implementation → Planning
Triggers when:
- Critical gaps in plan discovered
- Requirements changed
- Step success criteria cannot be met

Process:
1. Human approves return to planning with reason (via CLI)
2. Agent updates plan.yaml through MCP tools (version increments)
3. ASSESS.ai.md reviews which steps are affected
4. All step statuses reset to pending
5. Implementation resumes from first incomplete step

## Context Injection System

The CLI dynamically injects context when delivering AI Functions:

### Injected Context Categories
- **Phase Context**: Current phase, available transitions
- **Step Context**: Current step details, success criteria, dependencies
- **Plan Context**: Relevant plan sections, deliverables
- **Progress Context**: Task status, completed steps, blockers
- **Memory Context**: Filtered relevant memories based on tags and recency
- **Project Context**: Project-level configuration and patterns
- **File Context**: Contents of files specified in AI Function's "Context Required"

### Context Filtering Rules
- Only inject memories tagged with current step or phase
- Limit to most recent N memories unless specifically requested
- Include project memories marked as "always-relevant"
- Exclude sensitive information based on security rules

## Error Handling

### MCP Error Responses
When operations fail, tools return structured errors:

```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "error": {
    "code": -32001,
    "message": "Invalid state transition",
    "data": {
      "current_phase": "implementation",
      "reason": "Cannot modify plan.yaml during implementation phase",
      "suggestion": "Request phase transition to planning first"
    }
  }
}
```

### Error Categories
- **Validation Errors**: Business rule violations (enforced by CLI logic)
- **State Errors**: Invalid state transitions or operations
- **Permission Errors**: Attempting forbidden operations in current phase
- **Context Errors**: Missing required context for operation
- **Concurrency Errors**: Conflicting simultaneous operations

### Escalation Protocol
AI Functions include escalation rules:
1. Attempt operation via appropriate MCP tool
2. If error, analyze error message and adjust approach
3. If unclear, call `forge_request_clarification` tool
4. If blocked, create memory entry and notify human

## Workflow Example with MCP

```bash
# 1. Human creates task
forge-ai task new --title="Add authentication to API"

# 2. MCP client launches forge-ai via configured command (STDIO session)
#    (Client executes the command from its MCP config and connects over STDIO)

# 3. Session established; agent begins planning phase
# Agent calls: forge_task_discover
# - CLI returns DISCOVER.ai.md with context
# - Agent gathers requirements

# Agent calls: forge_task_plan
# - CLI returns PLAN.ai.md with discovered context
# - Agent creates plan via forge_plan_update tool
# - CLI validates plan structure programmatically

# Agent calls: forge_task_assess
# - CLI returns ASSESS.ai.md with plan context
# - Agent reviews and optimizes plan

# 4. Human reviews plan
forge-ai task status        # Review the plan document
forge-ai task phase next    # Approve transition (CLI validates readiness)

# 5. Agent begins implementation phase
# Agent calls: forge_task_work
# - CLI returns EXECUTE.ai.md for first incomplete step
# - Agent works on step

# Agent calls: forge_step_complete
# - CLI validates completion criteria and marks step complete

# Agent calls: forge_task_checkpoint
# - CLI returns CHECKPOINT.ai.md
# - Agent saves progress

# 6. Agent discovers gap in plan
# Agent calls: forge_memory_add (documents issue)
# Human intervenes: forge-ai task phase prev --reason="Missing error handling specs"

# 7. Agent updates plan
# Agent calls: forge_task_plan
# - Updates plan with missing specs
# - CLI validates and increments version
# Agent calls: forge_task_assess
# - Reviews affected steps

# 8. Human approves updated plan
forge-ai task phase next

# 9. Agent continues implementation
# Agent calls: forge_task_work (resumes from first incomplete step)

# 10. Human moves to validation
forge-ai task phase next

# 11. Agent performs validation
# Agent calls: forge_task_validate
# Agent calls: forge_task_report (generates report)

# 12. Human completes task
forge-ai task complete
```

## Filesystem Structure

```
.forge/
└── ai/
    ├── manifest.yaml
    ├── project.yaml                 # Project registry and metadata
    ├── schemas/                     # CUE schemas for validation
    │   ├── project.cue
    │   ├── task.cue
    │   ├── plan.cue
    │   └── memory_index.cue
    ├── functions/                   # AI Function templates (template-provided)
    │   ├── planning/
    │   │   ├── DISCOVER.ai.md
    │   │   ├── PLAN.ai.md
    │   │   └── ASSESS.ai.md
    │   ├── implementation/
    │   │   ├── EXECUTE.ai.md
    │   │   └── CHECKPOINT.ai.md
    │   └── validation/
    │       ├── VALIDATE.ai.md
    │       └── REPORT.ai.md
    ├── memory/                      # Project-level memories
    │   └── project/
    │       ├── index.yaml
    │       ├── 001-architecture-decision.md
    │       └── 002-technology-choice.md
    ├── backups/                     # Automatic backups (managed by CLI)
    ├── logs/                        # Transaction and migration logs
    │   ├── transactions.log
    │   └── migrations.log
    └── tasks/                       # All task data (per-task directories)
        └── 001-cli-design/
            ├── task.yaml            # Task progress tracking
            ├── plan.yaml            # Task plan (contract)
            ├── memory/              # Task-specific memories
            │   ├── index.yaml
            │   ├── 001-template-issue.md
            │   └── 002-validation-approach.md
            └── artifacts/           # Task deliverables
                ├── prd.md
                ├── technical-design.md
                └── validation-report.md
```

### Directory Purposes

All paths below are relative to `.forge/ai/`:

**`project.yaml`**: Central registry of all tasks and current project status (CLI-managed)

**`schemas/`**: CUE schema definitions used to validate state files (CLI-enforced)

**`functions/`**: AI Function templates that CLI composes with context for MCP delivery

**`memory/project/`**: Project-wide decisions, patterns, and lessons that apply across tasks

**`tasks/`**: Individual task directories containing all task-specific data:
- `task.yaml`: Progress tracking and phase status (CLI-managed)
- `plan.yaml`: The binding contract produced during planning (CLI-managed)
- `memory/`: Task-specific context and decisions (CLI-managed)
- `artifacts/`: Documents and deliverables produced during the task

### File Ownership by Phase (via MCP Tools)

**Planning Phase** can modify (through MCP):
- `tasks/<id>/plan.yaml` (create/update via forge_plan_update)
- `tasks/<id>/artifacts/` (design docs, PRDs)
- `tasks/<id>/memory/` (decisions via forge_memory_add)

**Implementation Phase** can modify (through MCP):
- `tasks/<id>/task.yaml` (progress updates via forge_progress_update)
- `tasks/<id>/artifacts/` (deliverables)
- `tasks/<id>/memory/` (issues, workarounds via forge_memory_add)
- Project codebase (outside `.forge/ai/`)

**Validation Phase** can modify (through MCP):
- `tasks/<id>/task.yaml` (validation status)
- `tasks/<id>/artifacts/` (validation report)

### Version Control Considerations

For v1:

**Ignored**:
- `.forge/ai/` (all state, schemas, memory, tasks, functions, backups, logs)
  - Rationale: Avoid committing AI-managed state/templates until shared-state behavior for multi-developer environments is analyzed. Future versions may revise this policy.

**Committed**:
- Application/source code outside `.forge/ai/`

**Additional gitignore candidates** (optional):
- Temporary context injection files

## Template Distribution

Templates are packaged as OCI images containing:
- AI Function definitions (.ai.md files)
- MCP tool specifications
- Default project structure
- Template memory entries
- Manifest configuration

```bash
forge-ai publish --registry=ghcr.io/org/template
forge-ai init myproject --template=ghcr.io/org/template:v1.0.0
```

## Benefits of MCP Integration

1. **Contractual Interface**: CLI enforces the three-phase model and all business rules
2. **Guaranteed Validity**: CUE schema validation plus CLI business rules ensure correctness
3. **Atomic Operations**: State changes are validated and atomic
4. **No State Corruption**: LLM cannot create invalid state
5. **Clear Error Messages**: Invalid operations return actionable error messages
6. **Schema-Driven**: CUE schemas are the single source of truth with code generation
7. **Simplified AI Functions**: Functions focus on guidance, not file manipulation
8. **Universal Compatibility**: Works with any MCP-compatible agent
9. **Audit Trail**: All operations can be logged and traced
10. **Context Management**: CLI handles all context gathering and injection
11. **Consistency**: Same three-phase process for all tasks
12. **Flexibility**: Steps within phases adapt to task needs
13. **Contract-Driven**: Plan document ensures clear expectations
14. **Resumable**: Memory and progress tracking enable continuity
15. **Quality Gates**: Multiple AI functions per phase ensure thoroughness
16. **Traceability**: Clear path from plan to implementation to validation

## Summary

The MCP integration transforms Forge AI from a file-based workflow system into an API-driven state machine. By centralizing state changes behind the CLI and validating all state using CUE schemas with programmatic business rules, the architecture remains robust and maintainable. The CLI becomes the intelligent middleware that bridges structured workflows with natural language agents, ensuring consistency while maintaining flexibility. This creates a robust "AI API" where LLMs interact through well-defined MCP tools rather than direct file manipulation, with validation guaranteed by the combination of CUE schemas and CLI logic.

## Next Steps

1. Implement MCP server in CLI with full tool catalog
2. Build programmatic validation logic for all state transitions
3. Create AI Function templates with context injection points
4. Develop comprehensive error handling and recovery mechanisms
5. Create testing framework for MCP interactions
6. Develop agent-specific integration guides (Claude Code, Cursor, etc.)
7. Design monitoring and observability for MCP operations
8. Implement security and rate limiting for MCP endpoints
9. Build state migration tools for template upgrades (with backup/rollback and CUE validation)
10. Create debugging tools for MCP message inspection