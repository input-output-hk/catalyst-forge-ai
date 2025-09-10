# State Management Specification

## Overview

State management in Forge AI represents the persistent data layer that tracks project structure, task progress, and organizational memory. All state is managed through YAML files validated by CUE schemas, creating a robust system that prevents corruption while maintaining human readability.

This document details the structure, transitions, and management of state files within the Forge AI system.

## State Management Principles

### Core Design Decisions

1. **CLI as Single Authority**: Only the CLI can modify state files, enforcing all business rules
2. **CUE Schema Validation**: All state files validated against CUE schemas before writes
3. **Phase-Based Access Control**: Different workflow phases have different modification permissions
4. **Separation of Contract and Progress**: Plans (contracts) are separate from execution state
5. **Atomic Transactions**: All state changes are validated before application
6. **Human Readability**: YAML format ensures state is inspectable and debuggable
7. **Version Control Friendly**: All state changes produce meaningful diffs

### State Hierarchy

```
Project Level (Singleton)
├── project.yaml          # Project registry and metadata
├── memory/index.yaml     # Project-wide knowledge base
│
Task Level (Multiple)
├── task.yaml            # Execution progress
├── plan.yaml            # Binding contract
└── memory/index.yaml    # Task-specific knowledge
```

### Schema Organization

All state file schemas are defined in CUE format:

```
.forge/
└── schemas/
    ├── project.cue       # Validates project.yaml
    ├── task.cue          # Validates task.yaml
    ├── plan.cue          # Validates plan.yaml
    └── memory_index.cue  # Validates memory index.yaml files
```

The CLI uses the CUE Go API to validate state files before writes. At build time, Go structs are generated from these CUE definitions using `cue get go`, ensuring the schemas and Go code remain synchronized.

## State File Specifications

### Project State: `project.yaml`

**Purpose**: Central registry maintaining project-wide configuration and task inventory.

**Location**: `.forge/project.yaml`

**Schema**: `schemas/project.cue`

```cue
package schemas

import "time"

#Project: {
    // Project metadata
    projectName: string & =~"^.+$"  // Non-empty string
    template: {
        source: string & =~"^(ghcr|docker|quay)\\.io/[a-z0-9-/]+$"  // Valid OCI registry URL
        version: string & =~"^v\\d+\\.\\d+\\.\\d+$"  // Semantic version (e.g., v1.0.0)
    }
    status: "active" | "maintenance" | "archived"
    activeTask?: string  // Must reference existing task ID when validated

    // Task registry
    tasks: [ID=string]: {
        path: string & =~"^tasks/[a-z0-9-]+$"  // Relative path to task directory
        phase: "planning" | "implementation" | "validation"
        status: "active" | "blocked" | "completed" | "abandoned"
        plan_version: string & =~"^v\\d+$"  // e.g., "v1", "v2"
        created_at: time.Time
        completed_at?: time.Time

        // Business rule: completed_at required if status is completed/abandoned
        if status == "completed" || status == "abandoned" {
            completed_at: time.Time
        }
    }

    // Ensure activeTask references an existing task
    if activeTask != _|_ {
        activeTask: or([ for id, _ in tasks {id}])
    }
}
```

**State Transitions**:
```
Project Status:
active → maintenance → archived
       ↖─────────────↙
```

### Task Plan: `plan.yaml`

**Purpose**: Immutable contract defining what needs to be done, created during planning phase.

**Location**: `tasks/<task-id>/plan.yaml`

**Schema**: `schemas/plan.cue`

```cue
package schemas

import (
    "time"
    "list"
)

#Plan: {
    version: string & =~"^v\\d+$"  // Plan version (v1, v2, etc.)
    created_at: time.Time
    updated_at: time.Time

    // What we're trying to achieve (at least one required)
    objectives: [...string] & list.MinItems(1)
    objectives: [...=~"^.+$"]  // Each must be non-empty

    // What we'll deliver (at least one required)
    deliverables: [...#Deliverable] & list.MinItems(1)
    #Deliverable: {
        id: string & =~"^[a-z][a-z0-9-]*$"  // Valid identifier
        description: string & =~"^.+$"
        path: string  // Where it will be created
        type?: "code" | "document" | "config" | "data"
        required: bool | *true  // Default to true if not specified
    }

    // How we'll do it (sequential steps, at least one required)
    steps: [...#Step] & list.MinItems(1)
    #Step: {
        id: string & =~"^[a-z][a-z0-9-]*$"  // Valid identifier
        description: string & =~"^.+$"
        estimated_hours?: number & >=0.5 & <=40
        dependencies?: [...string]  // Step IDs that must complete first (future)

        // At least 2 success criteria per step
        success_criteria: [...string] & list.MinItems(2)
        success_criteria: [...=~"^.+$"]

        deliverables?: [...string]  // References to deliverable IDs
    }

    // Overall success definition
    success_criteria: [...string]
    success_criteria: [...=~"^.+$"]

    // Planning metadata
    planning_notes?: string
    assumptions?: [...string]
    risks?: [...string]

    // Ensure unique IDs
    _uniqueDeliverableIDs: {
        for i, d in deliverables {
            for j, d2 in deliverables if i < j {
                d.id != d2.id
            }
        }
    }

    _uniqueStepIDs: {
        for i, s in steps {
            for j, s2 in steps if i < j {
                s.id != s2.id
            }
        }
    }

    // Ensure step deliverables reference valid deliverable IDs
    for _, step in steps if step.deliverables != _|_ {
        for _, ref in step.deliverables {
            ref: or([ for d in deliverables {d.id}])
        }
    }
}
```

**Version Management**:
- Version increments when returning from implementation to planning
- Previous versions archived as `plan.v1.yaml`, `plan.v2.yaml`, etc.
- Version tracking enables plan evolution analysis

### Task State: `task.yaml`

**Purpose**: Tracks execution progress, referencing plan.yaml for requirements.

**Location**: `tasks/<task-id>/task.yaml`

**Schema**: `schemas/task.cue`

```cue
package schemas

import (
    "time"
    "list"
)

#Task: {
    // Task identity
    id: string & =~"^[a-z0-9][a-z0-9-]*$"  // Matches directory name
    title: string & =~"^.+$"
    description?: string
    created_at: time.Time

    // Current status
    current_phase: "planning" | "implementation" | "validation"
    status: "active" | "blocked" | "completed" | "abandoned"
    blocked_reason?: string

    // If blocked, reason is required
    if status == "blocked" {
        blocked_reason: string & =~"^.+$"
    }

    // Phase tracking
    phases: {
        planning: #PhaseStatus
        implementation: #PhaseStatus
        validation: #PhaseStatus
    }

    #PhaseStatus: {
        status: "pending" | "active" | "completed" | "skipped"
        started_at?: time.Time
        completed_at?: time.Time
        notes?: string

        // If completed, must have completed_at
        if status == "completed" {
            completed_at: time.Time
        }
        // If active or completed, must have started_at
        if status == "active" || status == "completed" {
            started_at: time.Time
        }
    }

    // Execution progress (references plan.yaml for requirements)
    progress: {
        steps: [ID=string]: #StepProgress
        deliverables_completed: [...string]  // Deliverable IDs from plan
    }

    #StepProgress: {
        status: "pending" | "in-progress" | "completed" | "abandoned" | "blocked"
        started_at?: time.Time
        completed_at?: time.Time
        abandoned_at?: time.Time
        blocked_at?: time.Time
        blocked_reason?: string
        notes?: string
        evidence?: [...string]  // How success criteria were met

        // Status-specific field requirements
        if status == "in-progress" {
            started_at: time.Time
        }
        if status == "completed" {
            started_at: time.Time
            completed_at: time.Time
            evidence: [...string] & list.MinItems(1)
        }
        if status == "abandoned" {
            abandoned_at: time.Time
            notes: string  // Reason for abandonment
        }
        if status == "blocked" {
            blocked_at: time.Time
            blocked_reason: string & =~"^.+$"
        }
    }

    // Work tracking
    work_log: [...#WorkLogEntry]
    #WorkLogEntry: {
        timestamp: time.Time
        phase: "planning" | "implementation" | "validation"
        step_id?: string
        action: string  // Step started, completed, checkpoint, etc.
        notes?: string
    }

    // Current phase must match phase history
    _phaseConsistency: {
        if current_phase == "planning" {
            phases.planning.status: "active"
        }
        if current_phase == "implementation" {
            phases.planning.status: "completed"
            phases.implementation.status: "active"
        }
        if current_phase == "validation" {
            phases.implementation.status: "completed"
            phases.validation.status: "active"
        }
    }
}
```

**State Transition Rules**:

Phase Transitions:
```
planning → implementation → validation → [complete]
    ↑←←←←←←←←←↙         ↑←←←←←←←←←←↙
   (with reason)      (with reason)
```

Step Status Transitions:
```
pending → in-progress → completed
            ↓      ↓
         blocked  abandoned
            ↓
      (unblocked) → in-progress
```

### Memory Index: `index.yaml`

**Purpose**: Catalogs memory entries for efficient retrieval and search.

**Locations**:
- `.forge/memory/project/index.yaml` (project-level)
- `tasks/<task-id>/memory/index.yaml` (task-level)

**Schema**: `schemas/memory_index.cue`

```cue
package schemas

import "time"

#MemoryIndex: {
    memories: [...#Memory]

    #Memory: {
        // Core fields
        id: string & =~"^\\d{3}-[a-z][a-z0-9-]*$"  // e.g., "001-auth-decision"
        date: string & =~"^\\d{4}-\\d{2}-\\d{2}$"  // YYYY-MM-DD format
        timestamp: time.Time
        title: string & =~"^.+$"
        tags: [...string]
        tags: [...=~"^[a-z][a-z0-9-]*$"]  // Lowercase, hyphenated
        summary: string & =~"^.{1,100}$"  // Max 100 characters
        file: string & =~"^\\d{3}-[a-z][a-z0-9-]*\\.md$"  // Markdown file

        // Task-specific fields (only in task memory)
        step_id?: string & =~"^[a-z][a-z0-9-]*$"
        phase?: "planning" | "implementation" | "validation"

        // Categorization
        type?: "decision" | "issue" | "learning" | "assumption" | "risk"
        priority?: "low" | "medium" | "high" | "critical"

        // Cross-references
        references?: {
            tasks?: [...string]
            memories?: [...string]
            deliverables?: [...string]
        }
    }

    // Ensure unique memory IDs
    _uniqueMemoryIDs: {
        for i, m in memories {
            for j, m2 in memories if i < j {
                m.id != m2.id
            }
        }
    }

    // Ensure sequential IDs (001, 002, 003...)
    _sequentialIDs: {
        for i, m in memories {
            m.id: =~"^\\d{3}-.*$"
        }
    }
}
```

## State Transition Management

### Phase Transition Requirements

#### Planning → Implementation

**Prerequisites**:
- plan.yaml exists and passes CUE Go API validation
- All steps have at least 2 success criteria (enforced by CUE)
- All deliverables are defined (enforced by CUE)
- No CUE Go API validation errors

**CLI Validation Process**:
1. Use CUE Go API to validate plan.yaml against schemas/plan.cue
2. Verify plan file exists at expected path
3. Check any cross-file dependencies

If validation fails, CUE Go API provides specific error messages indicating which constraints were violated.

**Human Gate**: `forge-ai task phase next`

#### Implementation → Validation

**Prerequisites**:
- All required steps completed or explicitly abandoned
- No steps in "in-progress" or "blocked" status
- All required deliverables marked complete

**CLI Validation Process**:
1. Verify task.yaml passes CUE validation
2. Check every step in plan has a progress entry
3. Verify no steps are in "in-progress" or "blocked" status
4. Confirm all required deliverables are marked complete
5. Check deliverable files exist at specified paths

**Human Gate**: `forge-ai task phase next`

#### Validation → Complete

**Prerequisites**:
- Validation report exists
- All success criteria checked
- Human reviews and approves

### Backward Transitions

#### Implementation → Planning

**Trigger Conditions**:
- Critical gap in plan discovered
- Requirements fundamentally changed
- Success criteria cannot be met

**Process**:
1. Human initiates: `forge-ai task phase prev --reason="..."`
2. CLI validates reason is provided
3. Plan version increments (v1 → v2)
4. All step progress reset to pending
5. Planning phase reactivated

**State Preservation**:
- Previous plan archived
- Memory entries retained
- Work products preserved in artifacts/

## Validation Framework

### CUE-Based Validation

The CLI uses CUE schemas to validate all state files before writes:

**Validation Process**:
1. CLI prepares state changes in memory
2. Writes to temporary file (e.g., `task.yaml.tmp`)
3. Uses CUE Go API to validate temporary file against schemas/task.cue
4. If validation passes, atomically renames temp file to actual file
5. If validation fails, discards temp file and returns error with CUE's detailed message

**Schema Benefits**:
- **Type Safety**: CUE enforces correct data types
- **Constraint Validation**: Regex patterns, ranges, and enums enforced
- **Cross-Field Dependencies**: CUE can validate relationships between fields
- **Code Generation**: `cue get go` generates Go structs from schemas at build time
- **Single Source of Truth**: Schema defines both structure and validation rules

### Business Rule Validation

Beyond CUE's structural validation, the CLI enforces business rules:

**Phase Transition Rules**:
- Phases must progress sequentially (no skipping)
- Planning → Implementation requires valid plan.yaml passing CUE validation
- Implementation → Validation requires all required steps completed or abandoned
- Validation → Complete requires validation report and human approval
- Backward transitions require documented reasons

**Plan Immutability**:
- During implementation phase, plan.yaml cannot be modified
- Changes require returning to planning phase with reason
- Plan version increments when modified after implementation has begun

**Cross-File Validation**:
While CUE validates individual files, the CLI validates relationships between files:
- Step IDs in task.yaml must exist in plan.yaml
- Deliverable IDs in task.yaml must exist in plan.yaml
- Task IDs in project.yaml must have corresponding directories
- Memory references must point to existing entities

### Schema-CLI Synchronization

The build process ensures schemas and CLI code remain synchronized:

```bash
# At build time
cue get go ./schemas/...  # Generate Go structs from CUE
go generate ./...         # Additional code generation
go build                  # Compile with generated code
```

At runtime, the CLI uses the CUE Go API for validation, which guarantees that the Go structs used by the CLI exactly match the CUE validation rules.

## Concurrency and Locking

### Lockfile-Based Exclusive Access

The CLI enforces single-agent access per task using a simple lockfile mechanism:

**Lock Creation**:
- When MCP server starts (STDIO initialization): Creates `.forge/tasks/<task-id>/.lockfile`
- Lockfile contains session metadata:
  ```yaml
  session_id: string      # Unique session identifier
  pid: number            # Process ID of MCP server
  started_at: datetime   # When lock was acquired
  agent: string          # Agent identifier (e.g., "claude-code", "cursor")
  ```

**Lock Enforcement**:
- Every state-modifying operation MUST check for lockfile presence
- If lockfile exists and session_id doesn't match current session: **REFUSE operation**
- If lockfile missing or session_id matches: Allow operation
- Error message when locked: "Task locked by another session since [timestamp]"

**Lock Removal**:
- Automatic: When MCP server shuts down cleanly (STDIO tear-down)
- Manual override: `forge-ai task unlock --force` (for stuck locks)
- Stale lock detection: CLI can detect and clean locks from dead processes

**Benefits**:
- Simple and robust - no complex version tracking needed
- Clear ownership - exactly one agent can modify a task at a time
- Easy debugging - lockfile is human-readable
- Crash recovery - stale locks can be detected and cleaned

### Operation Atomicity

Despite the lockfile preventing concurrent access, operations are still atomic for safety:

**Atomic Write Process**:
1. Validate the complete operation before any writes
2. Write to temporary file first (e.g., `task.yaml.tmp`)
3. Atomically rename temporary file to actual file
4. This ensures no partial writes even if process crashes mid-operation

## Recovery Mechanisms

### State Corruption Detection

The CLI performs integrity checks to detect state corruption:

**Integrity Checks**:
- Verify all required files exist (.forge/project.yaml, task files)
- Validate files against embedded CUE schemas using the CUE Go API
- Check that all cross-references between files are valid
- Ensure required fields are present (enforced by CUE Go API)
- Verify enum values are within allowed sets (enforced by CUE Go API)

When corruption is detected, the CLI provides clear error messages from CUE Go API validation indicating exactly which constraints were violated.

### Recovery Strategies

**1. Automatic Backup System**

Location: `.forge/backups/`

The CLI creates automatic backups:
- Before any state modification operation
- Timestamped copies: `project.yaml.20250109.143022.backup`
- Retained for 7 days by default
- Maximum 10 backups per file (oldest removed)

**2. Transaction Log**

Location: `.forge/logs/transactions.log`

Every state change is logged with:
- Timestamp of operation
- Type of change (phase transition, step completion, etc.)
- Before and after state snapshots
- User who initiated the change
- MCP tool that triggered the change

**3. Recovery Process**

When corruption is detected, the CLI can:

**Automatic Recovery**:
- Attempt to restore from most recent backup
- Validate restored state with CUE schemas
- If valid, continue with warning message
- If invalid, try next older backup

**Manual Recovery**:
```
forge-ai repair --check           # Diagnose issues using CUE Go API validation
forge-ai repair --auto            # Attempt automatic fixes
forge-ai repair --from-backup     # Restore from backup
forge-ai repair --rebuild         # Rebuild state from git history
```

**Common Repairs**:
- Remove references to deleted tasks
- Reset invalid enum values to defaults (guided by CUE)
- Reconstruct missing index files from existing memories
- Rebuild project.yaml from task directories

## Query Patterns

### MCP Tools for State Queries

The CLI exposes query capabilities through MCP tools that agents can invoke:

**Task Query Tools**:
- `forge_task_list` - List all tasks with filtering options
- `forge_task_status` - Get comprehensive status of current task
- `forge_task_find_blocked` - Find all blocked tasks in project
- `forge_task_by_phase` - Get tasks in specific phase

**Step Query Tools**:
- `forge_next_step` - Get next incomplete step to work on
- `forge_step_status` - Get detailed status of specific step
- `forge_steps_remaining` - List all incomplete steps

**Memory Query Tools**:
- `forge_memory_search` - Search by tags, keywords, or date range
- `forge_memory_relevant` - Get memories relevant to current context
- `forge_memory_by_step` - Find memories related to specific step
- `forge_memory_recent` - Get N most recent memory entries

### Query Optimization

The CLI implements efficient query strategies:

**Indexed Access**:
- Task registry in project.yaml serves as primary index
- Memory index files enable fast tag-based searches
- Step IDs enable direct lookup without scanning

**Filtered Loading**:
- Load only requested state files (lazy loading)
- Filter memories by relevance before returning
- Limit result sets to prevent overwhelming context

**Query Examples**:

When an agent calls `forge_next_step`:
1. CLI loads current task's plan and progress
2. Finds first step with status "pending" or "in-progress"
3. Returns step details with success criteria
4. Includes any memories tagged with that step ID

When an agent calls `forge_memory_relevant`:
1. CLI determines current phase and step
2. Filters memories by matching tags
3. Prioritizes recent memories
4. Returns limited set (e.g., 5 most relevant)

## Migration Patterns

### Manifest Version Upgrades

When the Forge template manifest is upgraded (e.g., from v1.0.0 to v2.0.0), existing state files may need migration to match new structures or requirements.

**Migration Triggers**:
- User updates project template: `forge-ai upgrade --template=v2.0.0`
- New required fields added to CUE schemas
- Field semantics change (e.g., "status" values renamed)
- Structural reorganization (e.g., flat to nested structure)

### Migration Process

**1. Detection Phase**:
The CLI compares the current manifest version with state file versions:
- Each state file contains a `state_version` field
- Manifest declares compatible state versions
- CLI identifies which files need migration

**2. Migration Execution**:
For each outdated state file:
- Create backup with version suffix (e.g., `task.yaml.v1.backup`)
- Apply sequential migrations (v1→v2, v2→v3, etc.)
- Validate migrated state against new CUE schemas
- Update `state_version` field

**3. Common Migration Types**:

**Field Addition**:
- New required fields get sensible defaults
- Optional fields added as null/empty
- Example: Adding `work_log` field to track all state changes

**Field Renaming**:
- Map old field names to new ones
- Preserve data while updating structure
- Example: `in_progress` → `active` status value

**Structure Changes**:
- Flatten or nest data structures
- Split or merge fields
- Example: Combining separate date/time into single timestamp

### Backward Compatibility

The CLI maintains compatibility with older state versions:

**Reading Strategy**:
1. Check `state_version` field in the state file
2. If older than current version, apply migrations sequentially in memory
3. Work with the migrated structure internally
4. Write back in new format on next save operation

**Migration Chain**:
- Migrations are applied in sequence (v1→v2→v3→current)
- Each migration transforms the state to the next version
- All migrations must succeed or the entire operation fails

**Compatibility Window**:
- Support reading 2 major versions back
- Automatic migration for 1 major version back
- Manual migration required for older versions

### Migration Safety

**Safeguards**:
- All migrations are reversible (backups retained)
- Dry-run mode: `forge-ai upgrade --dry-run`
- Validation after each migration step using CUE
- Rollback on any migration failure
- User confirmation required for destructive changes

**Migration Log**:
Location: `.forge/logs/migrations.log`

Records:
- Timestamp of migration
- Files affected
- Version transitions
- Any data transformations applied
- Success or failure status

## Error Messages and Recovery

### User-Friendly Error Reporting

The CLI provides clear, actionable error messages that guide users toward resolution:

**CUE Validation Error**:
```
Validation failed for task.yaml:
  status: conflicting values "active" and "blocked"
  blocked_reason: field is required when status is "blocked"
Validation performed using CUE Go API against schemas/task.cue
```

**Plan Immutability Error**:
```
Cannot modify plan during implementation phase.
To update the plan:
  1. Return to planning: forge-ai task phase prev --reason="..."
  2. Make your changes
  3. Return to implementation: forge-ai task phase next
```

**Incomplete Steps Error**:
```
Cannot proceed to validation with incomplete steps:
  - setup-project: blocked - waiting for dependencies
  - implement-auth: in-progress

Complete these steps or explicitly abandon them:
  forge-ai step abandon <step-id> --reason="..."
```

**Lock Conflict Error**:
```
Task locked by another session since 2025-01-09 14:30:22
Session: claude-code-3847
PID: 12345

Wait for session to complete or force unlock:
  forge-ai task unlock --force
```

**Version Mismatch Error**:
```
Task state version (v1) incompatible with manifest (v2)
Run migration: forge-ai upgrade --migrate
Or view changes: forge-ai upgrade --dry-run
```

### Error Categories

**CUE Go API Validation Errors**: Schema constraint violations
- Type mismatches
- Missing required fields
- Regex pattern failures
- Enum value violations
- Cross-field dependency failures

**Business Rule Violations**: Operations that break workflow rules
- Attempting to modify immutable plan
- Skipping required phases
- Completing task with incomplete steps

**Reference Errors**: Broken links between state files
- Step ID not found in plan
- Task ID not in registry
- Memory referencing deleted step

**Lock Errors**: Concurrency issues
- Task locked by another session
- Stale lock detected
- Lock file corrupted

## Summary

The state management system in Forge AI provides a robust, schema-validated, and human-readable persistence layer. Key achievements:

1. **CUE Schema Validation**: All state files validated against CUE schemas with Go struct generation
2. **Single-Agent Locking**: Simple lockfile mechanism prevents concurrent modifications
3. **Phase Isolation**: Clear boundaries on what can be modified when
4. **Atomic Operations**: Prevents partial updates and corruption
5. **Recovery Capabilities**: Automatic and manual recovery paths with backups
6. **Query Interface**: MCP tools provide efficient state access for agents
7. **Migration Support**: Smooth upgrades when manifest versions change
8. **Auditability**: All changes tracked with timestamps and reasons

The integration of CUE schemas ensures that validation rules are always in sync with the Go implementation through code generation. The separation of plan (contract) from progress (execution) ensures that commitments remain stable while allowing flexible tracking of actual work. The memory system provides a knowledge base that grows with the project, capturing decisions and lessons learned.

This design creates a reliable foundation for AI-assisted development, where agents can work confidently knowing that the CLI will prevent invalid state transitions while maintaining a clear audit trail of all changes.