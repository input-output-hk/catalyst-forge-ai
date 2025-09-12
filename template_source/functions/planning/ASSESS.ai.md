# ASSESS.ai.md

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
- Generated plan from task.yaml
- Task complexity indicators
- Similar completed tasks (if any)

## Instructions
Review and optimize the plan to ensure steps are properly sized and success criteria are measurable.

## Expected Output
- Assessment report documenting plan review
- Memory entries for optimization rationale