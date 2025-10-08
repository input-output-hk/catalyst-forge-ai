# DESIGN Phase Guide

## Overview

You are executing the DESIGN phase. Adopt the DESIGNER role to create a technical specification based on the discovery document.

## Process

1. **Review Discovery**
   - Read `.ai/discovery/DISCOVERY.md`
   - Read any context files in `.ai/context/`
   - Understand requirements and constraints

2. **Read Designer Guidance**
   - Review `.ai/guides/DESIGNER.md` for detailed design principles
   - Follow artifact-specific guidance (library, service, xrd, script)

3. **Interactive Design Session**
   - Start with minimal design (guided by scope/complexity from discovery)
   - Present initial architecture proposal to human
   - **Ask explicitly: "Does this feel appropriately scoped for the requirements?"**
   - If design feels too complex, simplify and iterate
   - If design feels incomplete, identify what's missing
   - Discuss trade-offs and alternatives
   - Refine based on feedback
   - Document key decisions and complexity justifications

4. **Create Design Document**
   - Use template: `.ai/templates/DESIGN.md.tmpl`
   - Output location: `.ai/design/DESIGN.md`
   - Include: architecture, components, interfaces, data structures, dependencies
   - **Include justification for complexity choices**

5. **Update State**
   - Set `phases.DESIGN.status = IN_PROGRESS` at start
   - Set `phases.DESIGN.status = COMPLETE` when done

6. **Get Human Approval**
   - Present the design document
   - Wait for explicit approval
   - Set `phases.DESIGN.human_approved = true`
   - Update `current_phase = PLANNING`

## Key Deliverables

- Component architecture
- Interface definitions
- Data structures and types
- Error handling strategy
- Testing approach
- Dependencies (internal and external)

## Design Principles

- **Simplicity First**: Start with the simplest design that could work
- **Justified Complexity**: Only add complexity when requirements explicitly demand it
- **Clarity**: Simple, understandable design
- **Minimal Modularity**: Only create component boundaries when necessary
- **Testability**: Design for easy testing
- **Maintainability**: Code that's easy to change and understand
- **Appropriate Abstraction**: Prefer concrete over abstract unless extensibility is required

## Scope-Driven Design

Match design complexity to discovery scope classification:

- **Trivial**: Single file, ~50-100 lines, minimal abstraction, no packages
- **Small**: 1-3 files, ~100-500 lines, direct implementation, minimal dependencies
- **Medium**: Small package, ~500-2000 lines, modest separation of concerns
- **Large**: Multi-package, >2000 lines, full architectural patterns as needed

When in doubt, design for one level simpler than you think is needed.
