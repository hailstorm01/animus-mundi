---
id: T-0035
title: Split cmd/cardtown/main.go into one file per command (pure move)
kind: harness
tier: flash
wave: none (board directive #26, cleanup 4 of 4, before cardtown play)
status: in-progress
branch: m2/T-0035-split-cardtown-main
depends_on: [T-0034]
claims: [file:cmd/cardtown/sim.go, file:cmd/cardtown/replay.go, file:cmd/cardtown/lint.go, file:cmd/cardtown/tools.go]
cards: []
requirements: []
---

## Goal
`cmd/cardtown/main.go` (about 860 lines) holds two groups of commands in one file:
- the toolchain commands: hook, audit, tdd, whatsup and lint;
- the game commands: sim and replay.

This task moves them into one file per command, all still in `package main`, so `cardtown play` (design note Deliverable 5) starts in a file of its own. Nothing is renamed, and no flag, message or output changes.

## Inputs
- Board directive #26 (the user's audit). Run `grep -n "^func " cmd/cardtown/main.go` for today's functions.
- T-0033 has already changed `runSim` and `runReplay` to use `sim.DecklistName`, `sim.StressName` and `sim.ResolveDeck`.
- Rulings that apply: none.

### Target layout (lead decision)
| File | Holds |
| :--- | :--- |
| `main.go` | `main`, `printUsage`, `findRepoRoot`, `normalizeArgs` |
| `sim.go` (new) | `runSim`, `printCoverage`, `eventCollector` and its methods |
| `replay.go` (new) | `runReplay` |
| `lint.go` (new) | `runLint`, `isRulesPath`, `isFencesPath` |
| `tools.go` (new) | `runHook`, `runAudit`, `runTdd`, `runWhatsUp` |

- Comments travel with their declarations.
- Each new file starts with a one-line comment saying which commands it holds.
- Imports are fixed up per file, and `gofmt` is clean.
- Any helper not listed stays in `main.go`.
- **Not in this task:** any behavior change, new flags, or the `play` command.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests. These are checks recorded in the Log, re-run by the reviewer, each against a baseline CLI built from `rework/executable-rules` before any edit.
1. **Same functions:** the sorted list of top-level declarations in `cmd/cardtown/*.go` is identical before and after, apart from file names.
2. **Only moves:** `git diff --color-moved=zebra` shows only moved blocks, header comments and import blocks.
3. **Outputs unchanged.** These are byte-identical to the baseline:
   - `cardtown` with no arguments (the usage text);
   - `cardtown sim -h` and `cardtown replay -h`;
   - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH`;
   - bare `cardtown sim`;
   - `cardtown lint --all`.

   `cardtown replay` on a fresh 5-game `-events` file at game 3 gives the same output from both binaries.
4. `go tool task check` passes, and the pre-commit hook still works: `cardtown hook install` is unchanged.

## Allowed Files
- `cmd/cardtown/main.go`
- new: `cmd/cardtown/sim.go`, `replay.go`, `lint.go`, `tools.go`

## Done When
- [ ] Acceptance cases 1 to 4 hold, with their commands and results in the Log
- [ ] `go tool task check` passes

## Log
- 2026-10-21 lead-architect: Wrote the task from board directive #26, with status `queued` behind T-0034 (harness, flash). It is a pure move, with five files including the four new ones. The size rule's purpose (one review's scope) is met by the moved-blocks check.
- 2026-10-21 lead-architect: T-0034 is approved (a69b191). Set to `ready`. The function list in `main.go` matches the Target layout table (checked with `grep -n "^func \|^type " cmd/cardtown/main.go`). The lead has checked out the task branch for the engineer (see T-0034's process slip).
  - Next: **the user runs `@kernel-engineer T-0035` in the Antigravity agent (Gemini Flash)**.
