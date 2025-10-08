# FINAL_REVIEW Phase Guide

## Overview

You are executing the FINAL_REVIEW phase. Conduct a complete audit of all artifacts before completion.

## Process

1. **Review All Artifacts**
   - Discovery document
   - Design document
   - All task implementations
   - Test results
   - Documentation

2. **Validate Completeness**
   - All acceptance criteria met from discovery
   - All design components implemented
   - All tasks completed successfully
   - Tests passing with adequate coverage
   - No open blockers

3. **Quality Audit**
   - Code follows style guide
   - Documentation is complete
   - Error handling is appropriate
   - Security considerations addressed
   - Performance requirements met

4. **Create Completion Report**
   - Summary of what was built
   - Key decisions made
   - Test coverage achieved
   - Known limitations (if any)
   - Next steps or recommendations
   - Save to: `.ai/completion/report.md`

5. **Update State**
   - Set `phases.FINAL_REVIEW.status = IN_PROGRESS` at start
   - Set `phases.FINAL_REVIEW.status = COMPLETE` when done

6. **Get Human Approval**
   - Present completion report
   - Highlight key achievements
   - Discuss any concerns or limitations
   - Wait for explicit approval
   - Set `phases.FINAL_REVIEW.human_approved = true`

## Success Criteria

Project is complete when:
- ✓ All phases marked COMPLETE in state.yml
- ✓ All tasks approved
- ✓ Integration tests pass
- ✓ Human approves final review
- ✓ No blockers remaining
- ✓ Code meets quality standards

## Completion Report Template

```markdown
# Project Completion Report

## Summary
[What was built and why]

## Artifacts Delivered
- [List of deliverables]

## Test Coverage
- Unit tests: X passing
- Integration tests: Y passing
- Coverage: Z%

## Key Decisions
- [Important architectural or design decisions]

## Known Limitations
- [Any limitations or caveats]

## Recommendations
- [Suggestions for future improvements]
```
