# IMPLEMENTATION Phase Guide

## Overview

You are executing the IMPLEMENTATION phase. Coordinate CODER and REVIEWER agents to implement each task sequentially.

## Process for Each Task

### 1. Prepare
- Read task spec from `.ai/planning/tasks/<task-id>.md`
- Create directory: `.ai/implementation/tasks/<task-id>/`
- Set iteration counter to 1

### 2. Implement (via CODER)

```bash
catalyst-ai run coder task_id=<task-id> project=.
```

Example for task 001:
```bash
catalyst-ai run coder task_id=001 project=.
```

The coder will:
- Read task spec from `.ai/planning/tasks/<task-id>.md`
- Implement the code
- Create implementation notes at `.ai/implementation/tasks/<task-id>/notes.md`

**Note**: This automatically uses `cursor-agent` which is optimized for code generation.

### 3. Review (via REVIEWER)

```bash
catalyst-ai run reviewer task_id=<task-id> project=.
```

Example for task 001:
```bash
catalyst-ai run reviewer task_id=001 project=.
```

The reviewer will:
- Check task specification
- Review implementation
- Run automated checks (compilation, tests, linting)
- Output review to `.ai/implementation/tasks/<task-id>/iteration-N/review.md`

### 4. Handle Review Results

**If APPROVED**:
- Update state: `phases.IMPLEMENTATION.tasks.<task-id>.status = COMPLETE`
- **PAUSE and wait for human approval**
- After approval, proceed to next task

**If NEEDS_REVISION**:
- Check iteration count
- If < 5: Go back to step 2 with feedback
- If >= 5: Escalate to human, add blocker to state.yml

**If BLOCKED**:
- Add blocker to `state.yml` blockers array
- Escalate to human immediately
- Wait for resolution

### 5. Update State
```yaml
phases.IMPLEMENTATION.tasks.<task-id>:
  status: COMPLETE | IN_REVIEW | NEEDS_REVISION | BLOCKED
  iterations: N
  completed_at: "timestamp"
```

### 6. Move to Next Task
- Set `phases.IMPLEMENTATION.current_task = <next-task-id>`
- Repeat process

## Completion

When all tasks complete:
- Set `phases.IMPLEMENTATION.status = COMPLETE`
- Update `current_phase = INTEGRATION_TESTING`

## Critical Rules

1. **Sequential execution only** - One task at a time
2. **Human approval required after each task** - Non-negotiable in MVP
3. **Max 5 iterations per task** - Then escalate
4. **Update state after every change** - Keep state.yml current
