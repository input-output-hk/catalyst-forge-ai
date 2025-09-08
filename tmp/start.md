Love it. Here’s a tight, **one-topic-at-a-time conversation plan** to design your personal, spec-inspired “system of work.” Each module ends with concrete outputs so we don’t drift.

# Conversation Plan (v1)

## 1) Outcomes & Constraints

**Objective:** Nail the purpose and guardrails before we invent process.
**Key questions:** What problems must this solve (consistency, speed, fewer regressions)? What must it avoid (agent chaos, over-ritual)? What are non-negotiables (single-agent, local-first, docs+scripts, repo-scoped memory)?
**Outputs:** 1-pager of principles + success metrics (“Definition of Done” for the system).

## 2) Canonical Artifacts & Repo Layout

**Objective:** Decide what documents exist and where.
**Key questions:** Minimal set for you: Brainstorm, PRD/ADR, Architecture doc, Roadmap, Playbooks (debug/perf), Checklist templates, CI blueprint. Typed vs prose (e.g., **CUE** schemas + markdown)? Standard paths (e.g., `.forge/`, `docs/architecture/`, `playbooks/`)?
**Outputs:** Repo tree + file naming conventions + template stubs.

## 3) Guardrails (“Constitution”) & Checklists

**Objective:** Encode rules the agent must obey.
**Key questions:** What lives in the constitution (coding standards, design tenets, security, perf budgets, commit/PR rules)? What gets enforced by lint/tests vs human review? What checklists gate movement between phases?
**Outputs:** `CONSTITUTION.md` (or `constitution.cue`) + phase checklists.

## 4) Agent Integration & Memory Model

**Objective:** Make single-agent tools actually behave.
**Key questions:** Which agents (Claude Code, Cursor) and how do they ingest context (files, RAG, summaries)? Repo-scoped memory store (local index) and **what** is cached (specs, ADRs, API surfaces, runbooks)? Allowed actions vs forbidden ones; prompt kit structure; fail-safes (dry-runs, diff-only, “ask-before-apply”).
**Outputs:** Prompt kit + context contract + memory policy (write/read/expire).

## 5) Initial Project Architecting Flow

**Objective:** Turn ideas → validated plan without chaos.
**Key questions:** Phase gates for Brainstorm → Spec → Architecture → Roadmap. What is “frozen” at each gate? Required artifacts (e.g., ADR-0001, CUE schema of core config). How to timebox/limit scope (walking skeleton first)?
**Outputs:** Step-by-step playbook + template commands (see Module 8).

## 6) Feature & Improvement Flow (Maintenance)

**Objective:** A tight loop for adding/changing behavior.
**Key questions:** Small spec per change? When to open an ADR vs amend the spec? Branch naming, task slicing, acceptance criteria shape, doc updates required.
**Outputs:** Feature RFC template, PR template, “small-batch” guideline.

## 7) Debugging & Performance Flow (Maintenance)

**Objective:** A reliable loop for incidents and speed work.
**Key questions:** Standard triage notes (signals, hypotheses, experiments), perf budget targets (latency, cost), reproducibility protocol (fixtures, traces). Where do runbooks live and how are they kept fresh?
**Outputs:** Debug playbook, Perf playbook, runbook template + retention rule.

## 8) Automation Harness (CLI + Scripts) and CI/CD Hooks

**Objective:** Replace human glue with repeatable commands.
**Key questions:** A tiny CLI (your **FORGE**) that scaffolds docs, enforces layout, runs checklists, generates ADR numbers, opens branches, kicks Earthly builds, and wires CI. What runs locally vs in CI? Pre-commit hooks and required checks?
**Outputs:** Command spec (e.g., `forge spec new`, `forge adr new`, `forge plan check`), make/Earthly targets, CI workflows (validate spec, docs-up-to-date, schema-lint).

## 9) Metrics, Review Cadence, and Evolution

**Objective:** Ensure the system improves over time.
**Key questions:** What will we measure (lead time, PR rework rate, “agent correction” count, perf SLOs)? When do we review constitution/templates? What triggers refactors of the system itself?
**Outputs:** Minimal metrics sheet, review checklist, change log for the system.

---

## How we’ll use this plan

* We tackle **one module at a time**, produce the stated outputs (usually a short 1-pager + a couple of templates), and only then move on.
* We keep a “**Parking Lot**” for ideas that pop up out of sequence (e.g., typed API contracts, golden tests, secrets policy).

If this looks good, say **“Start with 1”** and we’ll lock in your principles and success metrics first.
