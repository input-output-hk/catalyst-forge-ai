# DISCOVERY Phase Guide

## Overview

You are executing the DISCOVERY phase. Your goal is to understand what to build through interactive dialog with the human.

## Process

1. **Understand the Problem**
   - Ask questions to clarify requirements
   - Identify constraints and limitations
   - Understand success criteria

2. **Determine Artifact Type**
   - **library**: Reusable code package with public API
   - **service**: Standalone application/microservice
   - **xrd**: Crossplane Resource Definition
   - **script**: Utility script or automation tool

3. **Assess Scope and Complexity**
   - Determine the expected size/complexity of the artifact
   - Classify as: **Trivial** / **Small** / **Medium** / **Large**
   - Define explicit scope constraints to guide design
   - Identify what should be simple vs what requires complexity

4. **Create Discovery Document**
   - Use template: `.ai/templates/DISCOVERY.md.tmpl`
   - Output location: `.ai/discovery/DISCOVERY.md`
   - Include: problem statement, requirements, success criteria, constraints

5. **Update State**
   - Set `phases.DISCOVERY.status = IN_PROGRESS` at start
   - Set `phases.DISCOVERY.status = COMPLETE` when done
   - Update `artifact.type`, `artifact.name`, `artifact.description`

6. **Get Human Approval**
   - Present the discovery document
   - Wait for explicit approval
   - Set `phases.DISCOVERY.human_approved = true`
   - Update `current_phase = DESIGN`

## Key Questions to Ask

- What problem does this solve?
- Who will use it and how?
- What are the inputs and outputs?
- Are there any existing implementations to reference?
- What are the performance/security requirements?
- What should it NOT do? (out of scope)
- **How complex should this be?** (single file? multi-package? microservice?)
- **What's the expected size?** (lines of code, number of components)
- **What level of abstraction is appropriate?** (concrete implementation vs flexible framework)

## Tips

- Be thorough but focused
- Capture the "why" not just the "what"
- Ask clarifying questions when requirements are vague
- Document assumptions explicitly
