# DESIGNER Role Guide

This guide is referenced by the ORCHESTRATOR when executing the DESIGN phase internally.

## Role

You are adopting the **Designer** role to create a technical specification based on the discovery document.

## Inputs

- `.ai/discovery/DISCOVERY.md` - Requirements and success criteria
- `.ai/context/` - Any context files provided by the human
- Interactive feedback from human

## Outputs

- `.ai/design/DESIGN.md` - Complete technical specification

## Design Session Structure

1. **Review Discovery**
   - Read and understand requirements
   - Identify key constraints
   - Clarify ambiguities with human

2. **Architecture Decisions**
   - Define component boundaries
   - Specify interfaces and contracts
   - Document dependencies
   - Choose data structures
   - Decide error handling approach

3. **Technical Constraints**
   - Language and framework choices
   - Testing strategy
   - Performance requirements
   - Security considerations

4. **Interactive Refinement**
   - Present design to human
   - Incorporate feedback
   - Iterate until approved

## Design Document Template

Use `.ai/templates/DESIGN.md.tmpl` as the structure.

## Artifact-Specific Guidance

### Library
Focus on:
- Public API surface
- Dependencies (internal and external)
- Test coverage requirements
- Example usage

### Service
Focus on:
- API contracts (REST, gRPC, etc.)
- Infrastructure dependencies
- Configuration approach
- Deployment considerations

### XRD (Crossplane Resource Definition)
Focus on:
- Resource specification
- Composition logic
- Dependencies on other XRDs
- Validation rules

### Script
Focus on:
- Inputs and outputs
- Error handling
- Logging approach
- Usage examples

## Quality Standards

Design must specify:
- Clear component boundaries
- Well-defined interfaces
- Testability requirements
- Success criteria for each component

## Human Interaction

This is an interactive phase. Ask questions to clarify:
- Unclear requirements
- Trade-offs between approaches
- Prioritization of features
- Acceptance criteria

## Completion Criteria

Design complete when:
- All components defined
- Interfaces documented
- Dependencies identified
- Human approves specification
