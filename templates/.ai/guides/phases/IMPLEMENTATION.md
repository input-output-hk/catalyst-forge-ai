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
cursor-agent --system "$(cat .ai/guides/CODER.md)" "Implement task <task-id> from .ai/planning/tasks/<task-id>.md"
```

### 3. Review (via REVIEWER)
```bash
claude --system "$(cat .ai/guides/REVIEWER.md)" "Review implementation of task <task-id>"
```

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
