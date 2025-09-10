# AI Functions Specification

## Overview

AI Functions are the structured guidance system that directs LLM agents through specific activities within the Forge AI workflow. They serve as focused prompts delivered through MCP tools, ensuring consistent execution while maintaining the flexibility of natural language processing.

This document details the design, structure, and implementation of AI Functions within the Forge AI system.

## Core Concepts

### What Are AI Functions?

AI Functions are plain text instructions that guide LLM agents through specific tasks. Unlike traditional software functions with typed parameters and return values, AI Functions leverage the natural language capabilities of LLMs:

- **Input**: Described in plain text as context requirements
- **Body**: Natural language instructions for the task
- **Output**: Plain text description of expected deliverables

This approach harnesses LLM strengths rather than forcing rigid structures that would limit their capabilities.

### Delivery Mechanism

AI Functions are delivered through the Forge CLI's MCP (Model Context Protocol) server:

1. Agent calls an MCP tool (e.g., `forge_task_work`)
2. CLI determines current context and appropriate function
3. CLI injects context into the function template
4. CLI returns the composed prompt via JSON-RPC
5. Agent executes the prompt immediately

## AI Function Anatomy

### Standard Structure

Every AI Function follows this template:

```markdown
# [FUNCTION_NAME].ai.md

## AI Function Protocol
You are executing an AI Function within the Forge AI system. AI Functions guide specific
activities within a structured workflow. This function has access to certain inputs, must
perform specific tasks, and produce defined outputs.

When executing this function:
- Read all files listed in "Context Required" before starting
- Follow the "Instructions" as your primary directive
- Produce outputs matching "Expected Output" specifications
- Handle ambiguity using the escalation rules below

### Escalation Rules
- **Continue with assumptions**: When details can be inferred or won't affect core
  deliverables (document in memory)
- **Request clarification**: When requirements conflict or success cannot be defined
- **Create memory entries**: For all decisions, assumptions, and workarounds

## Context Required
[List of state files, memories, and data the function needs access to]
- task.yaml (current task state)
- plan.yaml section for current step
- Relevant memory entries

## Instructions
[Core guidance for the task, including decision frameworks and thinking processes]

## Expected Output
[Specific deliverables and state changes the function should produce]
```

### Component Details

#### AI Function Protocol (Preamble)
A reusable section that teaches the LLM what an AI Function is and how to execute it. This section is consistent across all functions and includes:
- Explanation of the AI Function concept
- General execution guidelines
- Escalation rules for handling ambiguity
- Memory creation guidelines

#### Context Required
Specifies what information the CLI should inject:
- State files (task.yaml, plan.yaml, project.yaml)
- Memory entries (filtered by tags or relevance)
- Source code files
- Previous function outputs
- Project-level configuration

The CLI reads these files and injects their contents into the prompt before delivery.

#### Instructions
The core logic of the function, written in natural language. Should include:
- Step-by-step thinking processes
- Decision frameworks
- Examples where helpful
- Specific techniques or patterns to follow
- Quality criteria

#### Expected Output
Clear specification of what the function produces:
- State changes (via MCP tool calls)
- Documents to create
- Memory entries to add
- Tool calls to make

## Function Categories by Phase

### Planning Phase Functions

#### DISCOVER.ai.md
**Purpose**: Understand requirements and gather context
**Context Required**:
- Task description from task.yaml
- Project-level memories tagged ["architecture", "patterns", "decisions"]
- Any existing documentation

**Instructions**:
- Interview the user about requirements
- Identify constraints and non-functional requirements
- Explore edge cases and error conditions
- Document assumptions that need validation

**Expected Output**:
- Discovery notes document (via `forge_artifact_create`)
- Memory entries for key requirements (via `forge_memory_add`)
- List of clarification questions if needed

#### PLAN.ai.md
**Purpose**: Create structured plan with sequential steps
**Context Required**:
- Discovery notes from previous step
- Task description
- Project patterns and architecture decisions

**Instructions**:
```
Given the discovered requirements, create a plan that:
1. Identifies concrete deliverables with file paths
2. Breaks work into sequential steps (2-6 hours each)
3. Writes binary success criteria for each step

Use this thinking process:
- First, identify the key deliverables
- Work backwards to identify required steps
- Ensure each step has 2-4 specific success criteria
- Verify steps can be completed independently
```

**Expected Output**:
- plan.yaml structure (via `forge_plan_create`)
- Memory entries for planning decisions

#### ASSESS.ai.md
**Purpose**: Review and optimize the plan
**Context Required**:
- Generated plan.yaml
- Task complexity indicators
- Similar completed tasks (if any)

**Instructions**:
- Verify step sizes (not too large or small)
- Check success criteria are measurable
- Identify missing steps or deliverables
- Suggest optimizations

**Expected Output**:
- Plan revisions (via `forge_plan_update`)
- Assessment report
- Memory entries for optimization rationale

### Implementation Phase Functions

#### EXECUTE.ai.md
**Purpose**: Work through plan steps sequentially
**Context Required**:
- Current step from plan.yaml
- Step success criteria
- Previous step outputs
- Relevant memories

**Instructions**:
```
Execute the current step following these guidelines:
1. Review success criteria carefully
2. Implement solution incrementally
3. Test against each criterion
4. Document key decisions

If you encounter blockers:
- First, check memories for similar issues
- Try alternative approaches
- Document workarounds in memory
- Only escalate if success criteria cannot be met
```

**Expected Output**:
- Step deliverables (code, documents, etc.)
- Progress update (via `forge_step_complete`)
- Memory entries for decisions/issues

#### CHECKPOINT.ai.md
**Purpose**: Save progress during long-running steps
**Context Required**:
- Current step status
- Work in progress
- Time elapsed

**Instructions**:
- Document current progress
- Save partial work
- Create memory entry for context
- Estimate remaining work

**Expected Output**:
- Progress notes (via `forge_progress_update`)
- Memory checkpoint entry
- Partial deliverables saved

### Validation Phase Functions

#### VALIDATE.ai.md
**Purpose**: Verify all success criteria are met
**Context Required**:
- Complete plan.yaml
- All deliverables
- Step completion status
- Test results

**Instructions**:
```
Systematically validate the implementation:
1. Check each deliverable exists at specified path
2. Verify each step's success criteria
3. Run any automated tests
4. Document evidence for each criterion

For any failures:
- Identify root cause
- Determine if remediation needed
- Document gaps clearly
```

**Expected Output**:
- Validation checklist with pass/fail status
- Evidence documentation
- Remediation recommendations

#### REPORT.ai.md
**Purpose**: Generate comprehensive validation report
**Context Required**:
- Validation results
- Task history
- Key decisions from memory

**Instructions**:
- Summarize what was delivered
- Highlight any deviations from plan
- Document lessons learned
- Provide recommendations

**Expected Output**:
- Final validation report (via `forge_artifact_create`)
- Task completion status (via `forge_task_complete`)
- Lessons learned memories

## MCP Tool Integration

### How Functions Are Invoked

AI Functions are delivered through MCP tools exposed by the CLI:

```json
// Agent calls tool
{
  "method": "tools/call",
  "params": {
    "name": "forge_task_work",
    "arguments": {}
  }
}

// CLI returns composed function
{
  "result": {
    "content": [{
      "type": "text",
      "text": "[Complete AI Function with context]"
    }]
  }
}
```

### Available MCP Tools

#### Workflow Tools (Return AI Functions)
- `forge_task_work` - Returns appropriate function for current phase/step
- `forge_task_plan` - Returns planning phase function
- `forge_task_validate` - Returns validation function
- `forge_task_checkpoint` - Returns checkpoint function

#### State Management Tools
- `forge_step_complete` - Mark step as complete
- `forge_step_start` - Begin working on step
- `forge_progress_update` - Update progress notes
- `forge_plan_create` - Create plan.yaml
- `forge_plan_update` - Modify plan.yaml

#### Memory Tools
- `forge_memory_add` - Add memory entry with tags
- `forge_memory_search` - Search memories by tags
- `forge_memory_relevant` - Get context-relevant memories

#### Artifact Tools
- `forge_artifact_create` - Create document/deliverable
- `forge_artifact_update` - Update existing artifact

## Context Injection

### What Gets Injected

The CLI automatically injects relevant context based on the function being executed:

1. **Current State**: Phase, active step, progress status
2. **Plan Context**: Relevant plan sections, success criteria
3. **Memory Context**: Filtered memories based on tags and relevance
4. **Project Context**: Architecture decisions, patterns, configurations
5. **File Contents**: Any files specified in "Context Required"

### Injection Example

Original function template:
```markdown
## Context Required
- Current step from plan.yaml
- Previous step results
```

After CLI injection:
```markdown
## Context Required
- Current step from plan.yaml:
  ```yaml
  id: "implement-init"
  description: "Create init command for new projects"
  success_criteria:
    - "Init command creates valid project structure"
    - "Template files copied correctly"
  ```
- Previous step results:
  - setup-project: Completed successfully
  - Created go.mod with module name
  - Basic directory structure in place
```

## Error Handling and Escalation

### Escalation Decision Framework

AI Functions include built-in guidance for handling ambiguity:

```markdown
### When to Continue vs. Escalate

**Continue and document your assumption when:**
- Missing details can be reasonably inferred from context
- Multiple valid approaches exist (pick one, note it in memory)
- Ambiguity doesn't affect the core deliverable

Example: "User wants a CLI but didn't specify the framework"
→ Choose based on project context, document decision in memory

**Escalate to user when:**
- Core requirements are contradictory
- Success criteria cannot be defined
- Critical technical decisions lack context

Example: "User wants real-time sync but also offline-first"
→ These conflict fundamentally, need clarification
```

### Memory as Error Log

Instead of throwing exceptions, AI Functions use memory entries to track:
- Assumptions made
- Workarounds implemented
- Alternative approaches considered
- Issues encountered

This creates an audit trail without interrupting workflow.

## Function Evolution

### Versioning Strategy

AI Functions evolve through practical experience:

1. **Initial Version**: Start with minimal structure
2. **Refinement**: Add specific guidance based on usage
3. **Optimization**: Incorporate learned patterns
4. **Specialization**: Create variants for specific domains

### Template Customization

Projects can override base functions:

```
.forge/ai/functions/
├── planning/
│   ├── DISCOVER.ai.md      # Base version
│   └── DISCOVER.custom.ai.md # Project override
```

The CLI loads custom versions when present, allowing domain-specific adaptations while maintaining the core workflow.

## Best Practices

### Writing Effective Functions

1. **Clear Thinking Processes**: Include step-by-step reasoning
2. **Concrete Examples**: Provide examples for complex decisions
3. **Measurable Outputs**: Define success in binary terms
4. **Progressive Refinement**: Start simple, add complexity as needed
5. **Memory Integration**: Encourage documentation of decisions

### Prompt Engineering Patterns

Effective patterns for AI Functions:

```markdown
## Instructions
Use this thinking process:
1. First, [initial analysis step]
2. Then, [decision point with criteria]
3. Finally, [execution with verification]

Consider these factors:
- [Factor 1 with weight/importance]
- [Factor 2 with trade-offs]
- [Factor 3 with examples]

If you encounter [specific situation]:
- Try [approach A] when [condition]
- Use [approach B] if [alternative condition]
- Document why you chose your approach
```

### Context Management

Keep context focused and relevant:
- Filter memories by tags and recency
- Include only necessary file contents
- Summarize long documents
- Preserve critical details

## Implementation Roadmap

### Phase 1: Core Functions
- Implement basic planning functions
- Create execution function
- Add simple validation

### Phase 2: MCP Integration
- Build STDIO server in CLI
- Expose workflow tools
- Add state management tools

### Phase 3: Context System
- Implement memory filtering
- Add context injection
- Create relevance scoring

### Phase 4: Optimization
- Refine function templates
- Add domain specializations
- Implement function metrics

## Summary

AI Functions transform the Forge AI system from a rigid workflow engine into an intelligent development assistant. By combining natural language guidance with structured state management through MCP tools, the system achieves both flexibility and reliability.

The key insights:
- Plain text descriptions leverage LLM strengths
- MCP tools provide clean state management
- Artifact chains preserve work products with intelligent loading
- Context injection ensures relevance without overwhelming the context window
- Memory system captures organizational knowledge
- Escalation rules handle ambiguity gracefully

The artifact system is particularly powerful - by maintaining an indexed catalog of all work products with descriptions, agents can maintain awareness of the complete task context while selectively loading only what they need. This creates an efficient, traceable workflow where each function builds naturally on the previous one's outputs.

This design allows AI Functions to evolve with usage while maintaining the consistency and predictability required for software development workflows.