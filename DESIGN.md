# Catalyst Forge AI - Design Document

## Executive Summary

`catalyst-forge-ai` is a multi-agent system for building components of the Catalyst Forge platform through LLM-assisted development. The system uses a structured process with distinct phases (Discovery, Design, Planning, Implementation, Integration Testing, Final Review) coordinated by an orchestrator agent that delegates work to specialized agents.

The system is designed to be generic enough to build any code artifact (libraries, services, XRDs, scripts) while maintaining strict quality standards through automated review loops and human checkpoints.

## Goals

1. Accelerate platform development through LLM assistance
2. Maintain production-quality code standards
3. Enable rapid iteration with early feedback
4. Provide clear audit trail of all decisions and implementations
5. Support resume capability if process is interrupted

## Non-Goals

1. Deployment and infrastructure management (human responsibility)
2. Generic "build anything" framework (specialized for Catalyst Forge platform)
3. Fully autonomous operation (human approval required at checkpoints)

## System Architecture

### Process Phases

```
DISCOVERY → DESIGN → PLANNING → IMPLEMENTATION → INTEGRATION_TESTING → FINAL_REVIEW
    ↓         ↓          ↓              ↓                 ↓                  ↓
  Human    Human      Human       Human (per task)     Human             Human
  Dialog   Review     Review         Approval          Review           Approval
```

**DISCOVERY**:
- Orchestrator works interactively with human to understand problem
- Human provides context files as needed
- Output: Discovery document defining what to build and success criteria
- Determines artifact type (library, service, xrd, script)

**DESIGN**:
- Orchestrator adopts designer role
- Interactive session with human to create technical specification
- Output: Design document with architecture decisions and constraints
- Human approves before proceeding

**PLANNING**:
- Orchestrator delegates to Planner agent
- Breaks design into executable tasks with dependencies
- Output: Roadmap and individual task specifications
- Human approves task breakdown

**IMPLEMENTATION**:
- For each task sequentially:
  1. Orchestrator delegates to Coder agent
  2. Coder produces implementation + notes
  3. Orchestrator delegates to Reviewer agent
  4. Reviewer validates against design and style guide
  5. **Human approves task before proceeding to next** (MVP requirement)
  6. If revisions needed: max 5 iterations before human intervention
- Human approval required after each task completes

**INTEGRATION_TESTING**:
- Run full integration test suite
- Validate all components work together
- Human reviews results

**FINAL_REVIEW**:
- Complete audit of all code and artifacts
- Human final approval

### Agent Responsibilities

**ORCHESTRATOR** (Primary Agent):
- Manages state.yml
- Executes DISCOVERY phase directly (interactive with human)
- Executes DESIGN phase directly (adopts DESIGNER role internally)
- Delegates PLANNING to Planner agent via `claude` CLI
- Delegates IMPLEMENTATION tasks to Coder agent via `cursor-agent` CLI
- Delegates REVIEW to Reviewer agent via `claude` CLI
- Coordinates feedback loops
- Enforces human approval checkpoints
- Handles error conditions and blockers

**PLANNER** (Spawned by Orchestrator):
- Reads design document
- Creates task breakdown with dependencies
- Produces roadmap and task specifications
- Invoked via: `claude --system "$(cat .ai/guides/PLANNER.md)" "Create plan"`

**CODER** (Spawned by Orchestrator per task):
- Implements individual task based on specification
- Produces code + implementation notes
- Invoked via: `cursor-agent --system "$(cat .ai/guides/CODER.md)" "Implement task 001"`

**REVIEWER** (Spawned by Orchestrator per task):
- Validates implementation against design and style guide
- Produces structured feedback (YAML format)
- Determines APPROVED | NEEDS_REVISION | BLOCKED
- Invoked via: `claude --system "$(cat .ai/guides/REVIEWER.md)" "Review task 001"`

## CLI Tool

### Commands

**`catalyst-ai init`**

Purpose: One-time setup to cache catalyst-forge-ai repository locally

Behavior:
- Clones `github.com/input-output-hk/catalyst-forge-ai` to `~/.cache/catalyst-forge-ai`
- Stores commit SHA for version tracking
- Idempotent (safe to run multiple times)

Example:
```bash
$ catalyst-ai init

Checking for catalyst-forge-ai repository...
✓ Cloned to ~/.cache/catalyst-forge-ai
✓ Using commit: abc123def

You can now use 'catalyst-ai new' to start a project.
```

**`catalyst-ai new`**

Purpose: Initialize AI workspace in current directory

Behavior:
- Checks if `.ai/` exists (error if present)
- Copies templates from cache to `.ai/`
- Creates initial state.yml
- Prints next steps

Copied files:
- `.ai/state.yml` (initialized to DISCOVERY phase)
- `.ai/guides/` (all agent role definitions)
- `.ai/templates/` (document templates)
- `.ai/context/.gitkeep` (empty directory for context files)
- `.ai/README.md` (usage instructions)

Example:
```bash
$ cd my-project
$ catalyst-ai new

✓ Created .ai/ workspace
✓ Copied templates from ~/.cache/catalyst-forge-ai

Next steps:
1. (Optional) Add context files to .ai/context/
2. Start the orchestrator:

   catalyst-ai start
```

**`catalyst-ai start`**

Purpose: Launch orchestrator agent with correct prompts

Behavior:
- Reads `.ai/state.yml` to determine current phase
- Invokes `claude` with orchestrator prompt and resume instructions
- Quality of life wrapper around manual `claude` invocation

Example:
```bash
$ catalyst-ai start

Launching orchestrator...
Current phase: DISCOVERY

[Claude session starts with orchestrator prompt loaded]
```

Implementation:
```bash
claude --system "$(cat .ai/guides/ORCHESTRATOR.md)" \
       "Resume work based on current state in .ai/state.yml"
```

## Directory Structure

### catalyst-forge-ai Repository

```
catalyst-forge-ai/
├── cmd/catalyst-ai/
│   └── main.go               # CLI implementation
├── templates/
│   └── .ai/
│       ├── state.yml.tmpl
│       ├── README.md         # Instructions for using the system
│       ├── guides/
│       │   ├── ORCHESTRATOR.md
│       │   ├── DESIGNER.md   # Referenced by orchestrator
│       │   ├── PLANNER.md
│       │   ├── CODER.md
│       │   └── REVIEWER.md
│       ├── templates/
│       │   ├── DISCOVERY.md.tmpl
│       │   ├── DESIGN.md.tmpl
│       │   ├── TASK.md.tmpl
│       │   └── REVIEW.md.tmpl
│       └── context/
│           └── .gitkeep
└── README.md                 # How to install and use catalyst-ai CLI
```

### Target Repository After Init

```
<project-repo>/
├── .ai/                      # gitignored
│   ├── state.yml
│   ├── README.md
│   ├── guides/               # Agent role definitions
│   │   ├── ORCHESTRATOR.md
│   │   ├── DESIGNER.md
│   │   ├── PLANNER.md
│   │   ├── CODER.md
│   │   └── REVIEWER.md
│   ├── templates/            # Document templates
│   │   ├── DISCOVERY.md.tmpl
│   │   ├── DESIGN.md.tmpl
│   │   ├── TASK.md.tmpl
│   │   └── REVIEW.md.tmpl
│   ├── context/              # User-provided context files (optional)
│   │   └── (architecture docs, examples, etc.)
│   ├── discovery/            # Created during DISCOVERY
│   │   └── DISCOVERY.md
│   ├── design/               # Created during DESIGN
│   │   └── DESIGN.md
│   ├── planning/             # Created during PLANNING
│   │   ├── ROADMAP.md
│   │   └── tasks/
│   │       ├── 001-interfaces.md
│   │       ├── 002-types.md
│   │       └── ...
│   ├── implementation/       # Created during IMPLEMENTATION
│   │   └── tasks/
│   │       ├── 001/
│   │       │   ├── notes.md
│   │       │   ├── iteration-1/
│   │       │   │   └── review.md
│   │       │   └── iteration-2/
│   │       │       └── review.md
│   │       └── 002/
│   │           └── ...
│   └── completion/           # Created during FINAL_REVIEW
│       └── report.md
└── (actual code being built)
    ├── walker.go
    ├── walker_test.go
    └── ...
```

## State Management

### state.yml Schema

```yaml
# Metadata
version: "1.0"
created_at: "2025-01-15T10:30:00Z"
cache_source: "abc123def"  # commit SHA from catalyst-forge-ai

# What we're building
artifact:
  type: ""           # Set during DISCOVERY: library | service | xrd | script
  name: ""           # e.g., "discovery" or "platform-api"
  description: ""    # One-liner from DISCOVERY

# Current progress
current_phase: DISCOVERY
phases:
  DISCOVERY:
    status: PENDING  # PENDING | IN_PROGRESS | COMPLETE | BLOCKED
    started_at: null
    completed_at: null
    human_approved: false

  DESIGN:
    status: PENDING
    started_at: null
    completed_at: null
    human_approved: false

  PLANNING:
    status: PENDING
    started_at: null
    completed_at: null
    human_approved: false
    task_count: 0

  IMPLEMENTATION:
    status: PENDING
    started_at: null
    current_task: null
    tasks: {}
    # Example populated during implementation:
    # tasks:
    #   "001-interfaces":
    #     status: COMPLETE
    #     iterations: 2
    #     completed_at: "2025-01-15T12:00:00Z"
    #   "002-types":
    #     status: IN_REVIEW
    #     iterations: 1

  INTEGRATION_TESTING:
    status: PENDING
    started_at: null
    completed_at: null
    human_approved: false

  FINAL_REVIEW:
    status: PENDING
    started_at: null
    completed_at: null
    human_approved: false

# Issues blocking progress
blockers: []
# Example:
# blockers:
#   - phase: IMPLEMENTATION
#     task: "003-parser"
#     description: "Dependency on upstream library not ready"
#     created_at: "2025-01-15T14:00:00Z"
```

### State Transitions

State is updated by orchestrator at key checkpoints:
- Phase start: Update `phases.<phase>.status = IN_PROGRESS`
- Phase completion: Update `phases.<phase>.status = COMPLETE`
- Human approval: Update `phases.<phase>.human_approved = true`
- Task completion: Update `phases.IMPLEMENTATION.tasks.<task>.status = COMPLETE`
- Blockers: Append to `blockers` array

## Critical Design Decisions

### 1. Human Approval After Each Task (MVP)

During IMPLEMENTATION phase, orchestrator must pause after each task completion and wait for explicit human approval before proceeding to next task. This provides:
- Early feedback on agent quality
- Opportunity to adjust prompts/approach
- Granular control during initial rollout

Future enhancement: Add `--auto-approve` flag for trusted operations.

### 2. Two-Tool Strategy

**`claude` CLI**: Planning, review, orchestration (tasks requiring reasoning)

**`cursor-agent` CLI**: Code implementation (tasks requiring code generation)

This separation optimizes for each tool's strengths.

### 3. Feedback Loop with Guardrails

Review → Revise cycle limited to 5 iterations before escalating to human. Prevents infinite loops while allowing reasonable iteration.

### 4. Ephemeral Agent Invocations

Each agent invocation is stateless. All context comes from:
- Role definition in `.ai/guides/`
- Current state in `.ai/state.yml`
- Artifacts in `.ai/` subdirectories

This enables parallelization in future and simplifies debugging.

### 5. Discovery Phase Generalization

Discovery phase adapts output based on artifact type:
- **Library**: Focus on interfaces, dependencies, test coverage
- **Service**: Focus on API contracts, infrastructure dependencies
- **XRD**: Focus on resource specs, composition logic
- **Script**: Focus on inputs, outputs, error handling

Template structure remains the same; content adapts.

## Validation Strategy

### Per-Task Validation (Automated)

Reviewer agent checks:
- Code compiles (language-specific)
- Tests pass
- Linting passes (golangci-lint for Go)
- Test coverage meets threshold (>80%)

### Integration Validation (Manual)

Human reviews:
- All components work together
- Integration tests pass
- Documentation complete
- Meets original success criteria from DISCOVERY

### Final Validation (Manual)

Human approves:
- Code quality meets standards
- Design implemented correctly
- Ready for production use

## Success Criteria

Artifact complete when:
- ✓ All tasks implemented and reviewed
- ✓ Compiles without errors
- ✓ Passes linting (golangci-lint for Go)
- ✓ Test coverage > 80%
- ✓ Integration tests pass
- ✓ Human approves final review

## Future Enhancements (Out of Scope for MVP)

1. **Auto-approve mode**: Skip human checkpoint after each task for trusted operations
2. **Parallel task execution**: Allow independent tasks to run concurrently
3. **CLI validation commands**: `catalyst-ai validate` to run checks without agent
4. **State management commands**: `catalyst-ai status`, `catalyst-ai reset-task`
5. **Fine-grained resume**: Resume from exact iteration within a task
6. **Multi-artifact projects**: Build multiple related artifacts in single workspace
7. **Dependency management**: Enforce build order for dependent artifacts
8. **Update mechanism**: `catalyst-ai init --update` to pull latest agent prompts

## Open Questions

1. How do we handle dependency conflicts if Task A requires output from Task B that isn't complete yet?
   - Initial answer: Planner must create dependency-aware task order
   - Fallback: Human intervention to reorder tasks

2. What happens if Coder agent produces code that can't be reviewed automatically (e.g., integration test requires external service)?
   - Initial answer: Flag as BLOCKED, require human review
   - Reviewer should detect and escalate

3. How do we version the agent prompts themselves?
   - Initial answer: Prompts copied into `.ai/guides/` at `new` time, frozen for project
   - Future: Support prompt updates via `catalyst-ai upgrade-prompts`

## Implementation Plan

### Phase 1: CLI Skeleton
- Implement `init`, `new`, `start` commands
- Template copying logic
- state.yml initialization

### Phase 2: Orchestrator Prompt
- Write ORCHESTRATOR.md
- Define state transitions
- Define delegation patterns

### Phase 3: Agent Prompts
- Write PLANNER.md
- Write CODER.md
- Write REVIEWER.md
- Write DESIGNER.md (referenced by orchestrator)

### Phase 4: Document Templates
- DISCOVERY.md.tmpl
- DESIGN.md.tmpl
- TASK.md.tmpl
- REVIEW.md.tmpl

### Phase 5: Integration Testing
- Build first library using the system
- Iterate on prompts based on failures
- Document learnings

---

**Last Updated:** 2025-01-15