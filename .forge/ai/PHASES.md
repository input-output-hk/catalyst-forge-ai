# Forge AI Phase Flow System

## Overview

The Forge AI system enforces a strict phase-gated workflow to ensure consistent, high-quality software development. Each phase has specific objectives, constraints, and transition gates that must be satisfied before proceeding.

## Phase Progression

```
Brainstorm → Spec → Plan → Tasks → Implement → Merge → Post-merge
```

No phase can be skipped. Each transition requires satisfying specific gate criteria documented in the checklists.

## Phase Details

### 1. Brainstorm Phase

**Purpose**: Explore ideas and quickly determine viability.

**Location**: `.forge/ai/brainstorms/` (gitignored, local-only)

**Activities**:
- Capture initial ideas and assumptions
- Explore 3-5 candidate approaches
- Identify risks and constraints
- Make pursue/defer/drop decision

**Access**:
- Read: `brainstorms/**` only
- Write: `brainstorms/**` only
- Memory: Disabled

**Outputs**:
- Decision record (pursue/defer/drop)
- Initial problem statement draft
- Candidate approaches with risks

**Exit Gate**: [brainstorm-to-spec.md](project/checklists/brainstorm-to-spec.md)

---

### 2. Spec Phase

**Purpose**: Define the problem and solution boundaries clearly.

**Location**: `project/architecture/`, `project/guidelines/`

**Activities**:
- Formalize problem statement
- Define out-of-scope items
- Outline high-level interfaces
- Draft acceptance criteria

**Access**:
- Read: `project/architecture/**`, `project/guidelines/**`, `schemas/**`
- Write: `project/architecture/**`
- Memory: Disabled

**Outputs**:
- Problem statement (1-3 paragraphs)
- Out-of-scope list (3-7 bullets)
- High-level interface design
- Draft acceptance criteria

**Exit Gate**: [spec-to-plan.md](project/checklists/spec-to-plan.md)

---

### 3. Plan Phase

**Purpose**: Create detailed solution approach and work breakdown.

**Location**: `tasks/*/prd.md`, `project/`

**Activities**:
- Design solution architecture
- Identify technical approach
- Break down into tasks
- Create PRDs for features

**Access**:
- Read: `project/**`, `tasks/*/**`, `schemas/**`
- Write: `tasks/*/task.cue`, `tasks/*/acceptance.cue`, `tasks/*/prd.md`
- Memory: Task-scoped enabled

**Outputs**:
- Solution plan with milestones
- Task breakdown (≤1 day each)
- PRDs for feature work
- Risk mitigation strategies

**Exit Gate**: [plan-to-tasks.md](project/checklists/plan-to-tasks.md)

---

### 4. Tasks Phase

**Purpose**: Finalize work items with typed acceptance criteria.

**Location**: `tasks/NNN-slug/`

**Activities**:
- Create task manifests (`task.cue`)
- Define acceptance criteria (`acceptance.cue`)
- Establish traceability
- Prepare for implementation

**Access**:
- Read: `tasks/*/**`, `project/guidelines/**`, `schemas/**`
- Write: `tasks/*/**`
- Memory: Task-scoped enabled

**Outputs**:
- `task.cue` with type, status, links
- `acceptance.cue` with testable criteria
- Traceability plan (PR/branch strategy)

**Exit Gate**: [tasks-to-implement.md](project/checklists/tasks-to-implement.md)

---

### 5. Implement Phase

**Purpose**: Execute the work following test-first development.

**Location**: Full repository (constrained writes)

**Activities**:
- Write failing tests first
- Implement to make tests pass
- Update documentation
- Follow Constitution rules

**Access**:
- Read: `tasks/*/**`, `project/guidelines/**`, full codebase
- Write: Codebase except `brainstorms/**`, `.github/**`, `constitution/**`, `memory/project/**`
- Memory: Task-scoped enabled

**Outputs**:
- Working code with tests
- Updated documentation
- Diff proposals for review

**Exit Gate**: [implement-to-merge.md](project/checklists/implement-to-merge.md)

---

### 6. Merge Phase

**Purpose**: Integrate changes with proper documentation.

**Location**: Repository-wide documentation updates

**Activities**:
- Finalize PR with links
- Update ADRs if architecture changed
- Ensure all tests pass
- Update roadmap if needed

**Access**:
- Read: Full repository
- Write: `tasks/*/task.cue`, `adr/**`, `project/roadmap/**`, `docs/**`
- Memory: Full access (project + task)

**Outputs**:
- Merged PR with traceability
- Updated ADRs
- Documentation updates
- CI/CD validation passed

**Exit Gate**: PR merged successfully

---

### 7. Post-merge Phase

**Purpose**: Clean up and capture lessons learned.

**Location**: Memory system and task status updates

**Activities**:
- Mark task as "done"
- Capture memory entries
- Create follow-up tasks
- Update observability

**Access**:
- Read: Full repository
- Write: `memory/**`, `tasks/*/task.cue`, `project/roadmap/**`
- Memory: Full access (project + task)

**Outputs**:
- Task status updated to "done"
- Memory entries published
- Follow-up tasks created
- Metrics/monitoring updated

**Exit Gate**: [post-merge.md](project/checklists/post-merge.md)

---

## Gate Transitions

Each phase transition requires satisfying specific criteria documented in the gate checklists:

| From | To | Gate Checklist |
|------|----|----|
| Brainstorm | Spec | [brainstorm-to-spec.md](project/checklists/brainstorm-to-spec.md) |
| Spec | Plan | [spec-to-plan.md](project/checklists/spec-to-plan.md) |
| Plan | Tasks | [plan-to-tasks.md](project/checklists/plan-to-tasks.md) |
| Tasks | Implement | [tasks-to-implement.md](project/checklists/tasks-to-implement.md) |
| Implement | Merge | [implement-to-merge.md](project/checklists/implement-to-merge.md) |
| Post-merge | Complete | [post-merge.md](project/checklists/post-merge.md) |

## Key Principles

1. **No Phase Skipping**: Every phase must be completed in order
2. **Gate Enforcement**: Transitions only allowed when gate criteria met
3. **Context Isolation**: Each phase has specific read/write boundaries
4. **Traceability**: All work links back to tasks, specs, and ADRs
5. **Test-First**: Implementation always starts with failing tests
6. **Diff-Only**: Changes proposed as diffs for human approval

## Agent Behavior

Agents operating within this system must:

1. Declare current phase at session start
2. Respect phase-specific file access constraints
3. Follow the Constitution rules without exception
4. Stop and ask when hitting stop conditions
5. Maintain traceability throughout

## Enforcement

The phase system is enforced through:

- **Technical**: Context access controls in `context/policy.cue`
- **Process**: Gate checklists in `project/checklists/`
- **Cultural**: Constitution in `constitution/CONSTITUTION.md`
- **Automated**: CI validation of structure and schemas

## Getting Started

1. Start with a brainstorm in `brainstorms/`
2. Use the appropriate gate checklist to transition phases
3. Follow the Constitution rules throughout
4. Maintain traceability from task to PR
5. Capture learnings in the memory system