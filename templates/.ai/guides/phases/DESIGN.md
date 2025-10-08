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
   - Present architecture proposals to human
   - Discuss trade-offs and alternatives
   - Refine based on feedback
   - Document key decisions

4. **Create Design Document**
   - Use template: `.ai/templates/DESIGN.md.tmpl`
   - Output location: `.ai/design/DESIGN.md`
   - Include: architecture, components, interfaces, data structures, dependencies

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

- **Clarity**: Simple, understandable design
- **Modularity**: Well-defined component boundaries
- **Testability**: Design for easy testing
- **Maintainability**: Code that's easy to change
- **Appropriate abstraction**: Not too abstract, not too concrete
