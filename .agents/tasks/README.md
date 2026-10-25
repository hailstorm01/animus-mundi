# Build Tasks

One file per task, under `.agents/tasks/<milestone>/`. The task file is the only place a task's state lives. Agents hand work to each other through task files and git branches, never through chat, because the personas run in different tools on different models.

| Persona | Model | Runs in |
| :--- | :--- | :--- |
| `lead-architect` (Runbook D: queue owner and reviewer) | Opus, always | Claude Code, main session |
| `rules-test-author` | Sonnet | Claude Code (subagent `rules-test-author`) |
| `kernel-engineer`, tier `sonnet` | Sonnet | Claude Code (subagent `kernel-engineer`) |
| `kernel-engineer`, tier `flash` | Gemini 3.8 Flash (High) | Antigravity CLI, started by the lead: `agy -p "@kernel-engineer T-NNNN"` (Runbook D3) |

## Lifecycle

```text
queued → tests → tests-review → ready → in-progress → review → done
                     │                                  │
                     └──── tests (rework) ◄──┐          ├──► changes → in-progress
                                             │          └──► blocked (needs a shape, a ruling, or a split)
                                   lead rejects tests
```

| Status | Who acts next | What happens |
| :--- | :--- | :--- |
| `queued` | lead | Task written; waiting for its dependencies. |
| `tests` | `rules-test-author` | Writes the failing tests on the task branch. |
| `tests-review` | lead | Checks the tests against FRD 01 before any code is written. |
| `ready` | `kernel-engineer` | Tests approved; code may start. |
| `in-progress` | `kernel-engineer` | Making the tests pass. |
| `review` | lead | Reviews the diff against the checklist in `lead-architect` Runbook D. |
| `changes` | `kernel-engineer` | Fixes the numbered review notes. |
| `blocked` | lead | The engineer stopped: a missing shape, a rules question, or a task too big. |
| `done` | — | Merged into the integration branch. |

**Harness tasks** (tooling, linters, test helpers) skip `tests` and `tests-review`, because `rules-test-author` only tests game rules. The lead writes the task's `Acceptance Cases`. The engineer turns them into tests, commits those first as `test(T-NNNN): …`, then implements. The lead reviews tests and code together at `review`. A harness task goes `queued → ready → in-progress → review → done`.

**Test files hold only tests and setup.** Code that the tests exercise goes in non-test files. An edit to a test from an earlier task happens only when the task names it, and it lands in the `test(T-NNNN)` commit.

**One at a time (current mode):** at most one task is in `tests`, `in-progress` or `changes`. Parallel work, with one git worktree per task, starts only when the lead says so in the milestone's `BOARD.md`.

**Worktrees:** no persona switches the main checkout's branch. Task-branch work happens in a scratch worktree that is removed at handoff (the **Worktrees** guard rail in `.agents/rules/AGENTS.md`).

## Task File Template (`.agents/tasks/m2/T-NNNN-<slug>.md`)

```markdown
---
id: T-NNNN
title: <short imperative title>
kind: shape | card | harness
tier: flash | sonnet
wave: A1 | A2 | B | C | D | E | F | G
status: queued
branch: m2/T-NNNN-<slug>
depends_on: [T-NNNN]
claims: [<shape or file this task alone may create or change, e.g. primitive:SCRATCH, hook:deploy>]
cards: [<card names, for card tasks>]
requirements: [<FR-* IDs whose behavior this task makes executable>]
---

## Goal
One or two sentences: what runs when this is done.

## Inputs
- FRD 01 sections: <FR IDs and line ranges>
- Survey rows: docs/design/m2_starter_set_ability_survey.md (<cards>)
- Rulings that apply: <SVY-*/KIM-* IDs, or none>
- Design note: docs/architecture/m2-ability-model.md (<section>)

## Acceptance Cases (harness tasks only; written by the lead)
- <input or repository state> → <expected result, exit code or message>

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
- `TestFR_<DOM>_<NNN>_<Behavior>` in `<file>`: <one line>
- `TestCard_<CardSlug>_<Behavior>` in `<file>`: <one line>

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- `func (s *scene) <name>(<params>) <result>`: <one line, setup only>

## Allowed Files
- <non-test files the engineer may change>

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run '<regex>'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] <once the task touches gameplay> `go tool task sim` runs clean on the starter decks
- [ ] <pure moves> `go tool task sim-compare` passes (starter and stress sims byte-identical to the integration branch)
- [ ] <card tasks> the card loads with no unsupported lines

## Log
<!-- Each persona appends a dated entry when it hands off: what it did, what it ran, what is next. -->
```
