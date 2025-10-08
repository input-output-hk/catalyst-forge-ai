# Catalyst Forge AI

Multi-agent system for building Catalyst Forge platform components through LLM-assisted development.

## Overview

`catalyst-forge-ai` provides a structured multi-phase process (Discovery → Design → Planning → Implementation → Integration Testing → Final Review) coordinated by an orchestrator agent that delegates work to specialized agents.

See [DESIGN.md](DESIGN.md) for the complete design document.

## Prerequisites

- Go 1.21 or later
- `git` - For repository operations
- `claude` - For orchestrator and review agents ([claude.ai](https://claude.ai))
- `cursor-agent` - For code implementation

## Installation

```bash
go install github.com/input-output-hk/catalyst-forge-ai/cmd/catalyst-ai@latest
```

Or build from source:

```bash
git clone https://github.com/input-output-hk/catalyst-forge-ai.git
cd catalyst-forge-ai
go build -o catalyst-ai ./cmd/catalyst-ai
# Move to your PATH
mv catalyst-ai /usr/local/bin/
```

## Quick Start

### 1. Initialize the system

This clones the catalyst-forge-ai repository to `~/.cache/catalyst-forge-ai` for template storage:

```bash
catalyst-ai init
```

### 2. Create a new project

Navigate to your project directory and initialize the AI workspace:

```bash
cd my-project
catalyst-ai new
```

This creates a `.ai/` directory with:
- Agent role definitions
- Document templates
- State tracking file
- Context directory for optional reference materials

### 3. Start building

Launch the orchestrator agent:

```bash
catalyst-ai start
```

The orchestrator will guide you through the phases interactively, starting with Discovery.

## Workflow Phases

1. **DISCOVERY** - Interactive session to understand what to build
2. **DESIGN** - Create technical specification and architecture
3. **PLANNING** - Break design into executable tasks with dependencies
4. **IMPLEMENTATION** - Build each task with automated review loops
5. **INTEGRATION_TESTING** - Validate all components work together
6. **FINAL_REVIEW** - Complete audit before production

## Human Approval Checkpoints

The system requires human approval:
- After each phase completes
- **After each task during IMPLEMENTATION** (MVP requirement)

This ensures you maintain control and can provide feedback throughout.

## Commands

| Command | Description |
|---------|-------------|
| `catalyst-ai init` | Clone and cache catalyst-forge-ai repository |
| `catalyst-ai new` | Initialize AI workspace in current directory |
| `catalyst-ai start` | Launch orchestrator agent |
| `catalyst-ai help` | Show help message |

## Directory Structure

After running `catalyst-ai new`, your project will have:

```
<project-repo>/
├── .ai/                      # gitignored
│   ├── state.yml             # Current state tracking
│   ├── README.md             # Workspace documentation
│   ├── guides/               # Agent role definitions
│   ├── templates/            # Document templates
│   ├── context/              # Optional context files
│   ├── discovery/            # Created during DISCOVERY
│   ├── design/               # Created during DESIGN
│   ├── planning/             # Created during PLANNING
│   ├── implementation/       # Created during IMPLEMENTATION
│   └── completion/           # Created during FINAL_REVIEW
└── (your code)
```

## Optional: Adding Context

Before starting, you can add reference materials to `.ai/context/`:
- Architecture documentation
- Code examples
- API specifications
- Style guides

The agents will have access to these files during their work.

## Resuming Work

If you need to pause, simply run `catalyst-ai start` again. The orchestrator reads `.ai/state.yml` and continues where you left off.

## License

See [LICENSE](LICENSE) for details.
