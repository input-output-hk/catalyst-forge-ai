# Catalyst Forge AI Workspace

This directory contains the AI-assisted development workspace for your project.

## Directory Structure

- **`state.yml`**: Current state of the build process (phase, tasks, blockers)
- **`guides/`**: Agent role definitions (orchestrator, planner, coder, reviewer)
- **`templates/`**: Document templates for discovery, design, tasks, and reviews
- **`context/`**: Optional directory for context files you want to provide to agents
- **`discovery/`**: Created during DISCOVERY phase
- **`design/`**: Created during DESIGN phase
- **`planning/`**: Created during PLANNING phase with roadmap and task specs
- **`implementation/`**: Created during IMPLEMENTATION phase with task notes and reviews
- **`completion/`**: Created during FINAL_REVIEW phase

## Getting Started

1. **(Optional)** Add any context files to `context/` that would help the agents understand your domain
2. Start the orchestrator:
   ```bash
   catalyst-ai start
   ```

The orchestrator will guide you through the phases interactively.

## Workflow Phases

1. **DISCOVERY** - Interactive session to understand what to build
2. **DESIGN** - Create technical specification and architecture
3. **PLANNING** - Break design into executable tasks
4. **IMPLEMENTATION** - Build each task with review loops
5. **INTEGRATION_TESTING** - Validate everything works together
6. **FINAL_REVIEW** - Complete audit before production

## State Management

The `state.yml` file is updated by the orchestrator as work progresses. You can inspect it at any time to see current status.

## Resuming Work

If you need to pause and resume later, simply run `catalyst-ai start` again. The orchestrator will read `state.yml` and continue from where you left off.

## Human Approval Checkpoints

The system requires human approval at several points:
- After DISCOVERY phase
- After DESIGN phase
- After PLANNING phase
- **After each task during IMPLEMENTATION** (MVP requirement)
- After INTEGRATION_TESTING
- After FINAL_REVIEW

This ensures you maintain control and can provide feedback throughout the process.
