# PLANNING Phase Guide

## Overview

You are executing the PLANNING phase. Delegate to the PLANNER agent to break down the design into executable tasks.

## Process

1. **Invoke PLANNER Agent**

   Use your internal **Task tool** with the `general-purpose` subagent:

   ```
   Task tool invocation:
   - subagent_type: "general-purpose"
   - description: "Create implementation plan"
   - prompt: "You are the PLANNER agent. Read .ai/guides/PLANNER.md for your role definition, then create an implementation plan based on .ai/design/DESIGN.md and .ai/discovery/DISCOVERY.md. Output the roadmap and task files as specified in the guide."
   ```

2. **Review Output**
   - Check `.ai/planning/ROADMAP.md` for task overview
   - Review task files in `.ai/planning/tasks/`
   - Verify dependencies are clear
   - Ensure task scope is appropriate

3. **Update State**
   - Set `phases.PLANNING.status = IN_PROGRESS` at start
   - Count tasks and set `phases.PLANNING.task_count = N`
   - Set `phases.PLANNING.status = COMPLETE` when done

4. **Get Human Approval**
   - Present the roadmap
   - Summarize task breakdown
   - Wait for explicit approval
   - Set `phases.PLANNING.human_approved = true`
   - Update `current_phase = IMPLEMENTATION`

## What to Look For

- **Task size**: Each task should be completable in one focused session (1-3 files)
- **Dependencies**: Clear ordering, no circular dependencies
- **Coverage**: All design components have corresponding tasks
- **Clarity**: Each task has clear acceptance criteria

## Red Flags

- Tasks that are too large (>500 lines of guidance)
- Vague acceptance criteria
- Missing dependencies
- Tasks with code examples (should be high-level guidance only)
