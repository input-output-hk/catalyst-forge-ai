# INTEGRATION_TESTING Phase Guide

## Overview

You are executing the INTEGRATION_TESTING phase. Validate that all components work together correctly.

## Process

1. **Run Full Test Suite**
   - Execute all unit tests
   - Execute all integration tests
   - Check test coverage

2. **Validate Integration Points**
   - Verify components integrate correctly
   - Check data flow between components
   - Validate error propagation

3. **Document Results**
   - Create test results summary
   - Note any issues or concerns
   - Capture performance metrics if applicable

4. **Update State**
   - Set `phases.INTEGRATION_TESTING.status = IN_PROGRESS` at start
   - Set `phases.INTEGRATION_TESTING.status = COMPLETE` when done

5. **Get Human Approval**
   - Present test results
   - Discuss any concerns
   - Wait for explicit approval
   - Set `phases.INTEGRATION_TESTING.human_approved = true`
   - Update `current_phase = FINAL_REVIEW`

## What to Check

- All tests pass
- Test coverage meets threshold (>80%)
- No compilation warnings
- Linting passes
- Integration tests demonstrate end-to-end functionality

## If Tests Fail

- Document failures clearly
- Identify which task(s) need revision
- Add blocker to state.yml
- Escalate to human for decision on how to proceed
