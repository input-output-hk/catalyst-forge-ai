# PLANNING Phase Guide

## Overview

You are executing the PLANNING phase. Delegate to the PLANNER agent to break down the design into executable tasks.

## Process

1. **Invoke PLANNER Agent**

   ```bash
   catalyst-ai run planner project=.
   ```

   The planner will:
   - Read the design document from `.ai/design/DESIGN.md`
   - Create roadmap at `.ai/planning/ROADMAP.md`
   - Generate task specifications in `.ai/planning/tasks/`

2. **Review Output**
   - Check `.ai/planning/ROADMAP.md` for task overview
   - Review task files in `.ai/planning/tasks/`
   - Verify dependencies are clear
   - Ensure task scope is appropriate
   - **Verify all components/interfaces/types referenced exist in DESIGN.md**
   - **Check for fabricated details not in design document**

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

- Tasks that are too large (>100 lines of specification)
- Vague acceptance criteria
- Missing dependencies
- Tasks with code examples (should be high-level guidance only)
- **Components/interfaces/types not mentioned in DESIGN.md**
- **Implementation details that seem invented vs. from design**
- **Task guidance that doesn't reference design sections**
- **Assumptions about architecture not in design document**
