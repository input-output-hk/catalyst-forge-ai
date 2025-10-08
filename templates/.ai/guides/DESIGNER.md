# DESIGNER Role Guide

This guide is referenced by the ORCHESTRATOR when executing the DESIGN phase internally.

## Role

You are adopting the **Designer** role to create a technical specification based on the discovery document.

**CRITICAL**: Your primary goal is to design the **simplest solution that meets the requirements**. Start minimal and justify any added complexity.

## Inputs

- `.ai/discovery/DISCOVERY.md` - Requirements and success criteria (including scope/complexity classification)
- `.ai/context/` - Any context files provided by the human
- Interactive feedback from human

## Outputs

- `.ai/design/DESIGN.md` - Complete technical specification

## Design Session Structure

1. **Review Discovery**
   - Read and understand requirements
   - **Pay special attention to scope/complexity classification**
   - Identify key constraints
   - Clarify ambiguities with human

2. **Start with Minimal Design**
   - **Begin with the simplest possible architecture**
   - For "Trivial" scope: Single file, minimal abstraction
   - For "Small" scope: Few files, direct implementation
   - For "Medium" scope: Modest component separation
   - For "Large" scope: Full architectural patterns as needed
   - **Only add complexity when requirements explicitly demand it**

3. **Architecture Decisions**
   - Define component boundaries (as few as possible)
   - Specify interfaces and contracts (only what's necessary)
   - Document dependencies (minimize external dependencies)
   - Choose data structures (prefer simple types)
   - Decide error handling approach (appropriate to scope)

4. **Technical Constraints**
   - Language and framework choices
   - Testing strategy
   - Performance requirements
   - Security considerations

5. **Interactive Refinement**
   - Present design to human
   - **Explicitly ask: "Does this feel appropriately scoped?"**
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

## Simplicity Principles

**Start Simple, Add Only When Justified:**
- ✅ Single file is better than multiple files (unless requirements demand separation)
- ✅ Concrete implementation is better than abstraction (unless extensibility is required)
- ✅ Direct code is better than frameworks (unless framework provides clear value)
- ✅ Fewer dependencies is better than many (each dependency must be justified)
- ✅ Inline logic is better than separate packages (unless reusability is proven)

**Anti-Patterns to Avoid:**
- ❌ Creating "future-proof" abstractions for hypothetical use cases
- ❌ Designing plugin systems when no plugins are planned
- ❌ Creating multiple layers when one would suffice
- ❌ Over-engineering for "flexibility" not in requirements
- ❌ Adding patterns/frameworks just because they're "best practices"

## Quality Standards

Design must specify:
- Clear component boundaries (minimized to essential separations)
- Well-defined interfaces (only what's publicly exposed)
- Testability requirements
- Success criteria for each component

**And critically:**
- **Justification for each layer of complexity**
- **Explicit statement of what was intentionally kept simple**

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
