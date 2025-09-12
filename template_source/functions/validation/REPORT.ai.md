# REPORT.ai.md

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
- Validation results
- Task history
- Key decisions from memory

## Instructions
Generate a comprehensive validation report summarizing what was delivered.

## Expected Output
- Final validation report via artifact_save
- Task completion status
- Lessons learned memories