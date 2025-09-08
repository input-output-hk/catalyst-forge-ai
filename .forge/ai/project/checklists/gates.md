# Minimal Phase Gates

## Brainstorm → Spec
- [ ] Problem stated (1–3 paragraphs)
- [ ] 2–3 candidate approaches listed
- [ ] Risks/assumptions captured
- [ ] Decision recorded (pursue/defer/drop)

## Spec → Plan
- [ ] Out-of-scope list (3–7 bullets)
- [ ] High-level interfaces outlined
- [ ] Draft acceptance criteria written

## Plan → Tasks
- [ ] Solution approach chosen and justified
- [ ] 1–3 tasks identified with short scopes (≤1 day)
- [ ] Sequencing and dependencies noted

## Tasks → Implement
- [ ] `/tasks/<nnn>-<slug>/task.md` created
- [ ] Acceptance criteria checkboxes included
- [ ] Links to `project/spec.md` and `project/plan.md`

## Implement → Merge
- [ ] Failing tests proposed first and approved
- [ ] Implementation proposed to make tests pass
- [ ] Docs updated if public surface changed

## Merge → Post-merge
- [ ] PR description prepared using `project/pr-template.md`
- [ ] Evidence attached (tests passing, screenshots/logs)
- [ ] Linked tasks referenced

## Post-merge → Complete
- [ ] Entry appended to `memory/project/log.md`
- [ ] Follow-ups captured
- [ ] Task(s) marked done
EOF

cat > /Users/josh/work/catalyst-forge-ai/.forge/ai/project/spec.md <<\"EOF\"
# Spec

## Problem

## Out of Scope
- 

## Interfaces (High-level)
- 

## Acceptance Criteria
- [ ] 
EOF

cat > /Users/josh/work/catalyst-forge-ai/.forge/ai/project/plan.md <<\"EOF\"
# Plan

## Approach

## Tasks (1–3)
- [ ] 001-sample: 

## Sequencing & Risks
- 
EOF

cat > /Users/josh/work/catalyst-forge-ai/.forge/ai/project/pr-template.md <<\"EOF\"
# PR Summary

## Linked Task(s)
- 

## Acceptance Evidence
- Tests passing: 
- Screenshots/Logs: 

## Notes
- 
EOF

# Ensure memory log exists
: > /Users/josh/work/catalyst-forge-ai/.forge/ai/memory/project/log.md

# 9) Ensure brainstorms dir has a README placeholder to guide usage
cat > /Users/josh/work/catalyst-forge-ai/.forge/ai/brainstorms/README.md <<\"EOF\"
Brainstorm notes live here. Use files like `YYYY-MM-DD-slug.md` with:
- Problem
- 2–3 approaches
- Risks
- Decision
EOF

# 10) Ensure root tasks dir exists (already created)

printf "Done.\n"
