# Minimal Full-Cycle Test Guide (Greet CLI)

This guide walks a human through one end-to-end cycle using the stripped-down Forge AI flow with a tiny Go CLI: `greet`.

Phases follow `.forge/ai/PHASES.md`. Use the consolidated gate checklist at `.forge/ai/project/checklists/gates.md`.

---

## 0) Prep
- Ensure you are on a working branch: `git checkout -b 001-greet-cli`
- Open the safety prompts for reference:
  - `.forge/ai/prompts/00-session-start.txt`
  - `.forge/ai/prompts/40-implement-tests-first.txt`
  - `.forge/ai/prompts/50-ask-before-apply.txt`

---

## 1) Brainstorm
- Create: `.forge/ai/brainstorms/YYYY-MM-DD-greet-cli.md`
- Include:
  - Problem (1–3 paragraphs)
  - 2–3 approaches (e.g., plain `flag`, pure function + thin `main`)
  - Risks/assumptions
  - Decision (pursue)
- In `project/checklists/gates.md` check the Brainstorm → Spec items.

---

## 2) Spec
- Edit: `.forge/ai/project/spec.md`
- Fill in:
  - Problem (tiny CLI to say "Hello, <name>!"; default: world)
  - Out of Scope (networking, external deps, complex parsing)
  - Interfaces (function `Greet(name string) (string, error)` and CLI flags `--name`, `--json`)
  - Acceptance Criteria (copy these):
    - [ ] `go run ./cmd/greet` prints "Hello, world!" (exit 0)
    - [ ] `go run ./cmd/greet --name Alice` prints "Hello, Alice!" (exit 0)
    - [ ] `--json` prints `{"greeting":"Hello, <name>!"}` (valid JSON)
    - [ ] Names with digits cause non-zero exit and an error message
    - [ ] `go test ./...` includes success, edge (empty→world), failure (digits) and passes
- Check Spec → Plan items in `gates.md`.

---

## 3) Plan
- Edit: `.forge/ai/project/plan.md`
- Fill in:
  - Approach (pure function + thin CLI)
  - Tasks (<=3):
    - [ ] 001-greet-core: add `greet/greet.go`, `greet/greet_test.go`
    - [ ] 002-cli-main: add `cmd/greet/main.go`
  - Sequencing & Risks (e.g., tests-first, input validation)
- Check Plan → Tasks items in `gates.md`.

---

## 4) Tasks
- Create directory: `tasks/001-greet-core/`
- Create file: `tasks/001-greet-core/task.md` with:
  - Title, Scope
  - Acceptance (copy relevant parts from Spec for core function)
  - Links to `project/spec.md` and `project/plan.md`
- Check Tasks → Implement items in `gates.md`.

---

## 5) Implement (Tests First)
- Use the prompts:
  - `.forge/ai/prompts/40-implement-tests-first.txt`
  - `.forge/ai/prompts/50-ask-before-apply.txt`
- Steps:
  1. Write failing tests first:
     - Add `greet/greet_test.go` covering: default world, custom name, invalid (digits)
  2. After tests are failing, implement to pass:
     - Add `greet/greet.go` implementing `Greet`
     - Add `cmd/greet/main.go` for CLI flags `--name`, `--json`
  3. Ensure `go.mod` exists and `go test ./...` passes.
  4. Update `project/spec.md` if public surface changed.
- Check Implement → Merge items in `gates.md`.

---

## 6) Merge
- Prepare PR using: `.forge/ai/project/pr-template.md`
- Include:
  - Linked task(s): `tasks/001-greet-core/task.md`
  - Acceptance evidence: test output, sample runs for default/name/json/error
- Check Merge → Post-merge items in `gates.md`.

---

## 7) Post-merge
- Append an entry to `.forge/ai/memory/project/log.md`:
  - Date, linked task(s)
  - What worked / what did not
  - Follow-ups (e.g., add `--uppercase`, packaging)
- Check Post-merge → Complete items in `gates.md`.

---

## Notes for Humans
- Keep changes small and focused per task.
- Always propose diffs and review before applying when collaborating with an LLM.
- The minimal structure intentionally avoids schemas/CI; we validate manually via the gates and tests.
