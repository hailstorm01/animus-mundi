---
id: T-0035
title: Split cmd/cardtown/main.go into one file per command (pure move)
kind: harness
tier: flash
wave: none (board directive #26, cleanup 4 of 4, before cardtown play)
status: done
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
- [x] Acceptance cases 1 to 4 hold, with their commands and results in the Log
- [x] `go tool task check` passes

## Log
- 2026-10-21 lead-architect: Wrote the task from board directive #26, with status `queued` behind T-0034 (harness, flash). It is a pure move, with five files including the four new ones. The size rule's purpose (one review's scope) is met by the moved-blocks check.
- 2026-10-21 lead-architect: T-0034 is approved (a69b191). Set to `ready`. The function list in `main.go` matches the Target layout table (checked with `grep -n "^func \|^type " cmd/cardtown/main.go`). The lead has checked out the task branch for the engineer (see T-0034's process slip).
  - Next: **the user runs `@kernel-engineer T-0035` in the Antigravity agent (Gemini Flash)**.
- 2026-10-21 kernel-engineer: Built baseline binary `bin/cardtown-baseline.exe` prior to file split. Split `cmd/cardtown/main.go` into `sim.go`, `replay.go`, `lint.go`, `tools.go` with 1-line command header comments and imports per target layout, leaving shared helpers (`main`, `printUsage`, `findRepoRoot`, `normalizeArgs`) in `main.go`. Ran `go tool task fmt`.
  - Case 1 (Same functions): `git grep --untracked -h -E "^(func|type) " cmd/cardtown/ | sort` confirmed exact match of all 17 top-level declarations with baseline.
  - Case 2 (Only moves): `git diff --cached --color-moved=zebra` shows only moved declaration blocks, file headers, and imports.
  - Case 3 (Outputs unchanged): Binary `fc /b` verified byte-identical output against baseline for:
    - Usage text (`cardtown` no args)
    - Help texts (`cardtown sim -h`, `cardtown replay -h`)
    - Playouts (`cardtown sim -deck starter -p0 ANIMUS -p1 PATH`, bare `cardtown sim`)
    - Linters (`cardtown lint --all`)
    - Deterministic replay (`cardtown replay -events <file> -game 3` on fresh 5-game event log)
  - Case 4 (Check & hook): `cardtown hook install` ran cleanly; `go tool task check` passed completely (fmt, vet, test, build, lint). Set status to `review`.
- 2026-10-21 lead-architect: Review: **approved**, status `done`.
  - Process: the commits landed on the task branch, which the lead checked out before dispatch. d30b931 touches only `cmd/cardtown/`.
  - A multiset comparison of removed against added lines shows every removed line reappears. The only added lines are the 4 headers, 4 `package main` lines and the import blocks. The layout matches the table, and `main.go` is down to 150 lines.
  - Gates re-run by the reviewer:
    - `go tool task check` exits 0.
    - These are byte-identical, with the same exit codes, to a baseline CLI built from 81a01f5: `cardtown` (usage), `sim -h`, `replay -h`, `sim -deck starter -p0 ANIMUS -p1 PATH`, bare `sim` and `lint --all`.
    - `replay -game 3 -at 12` exits 0 on a fresh 5-game starter file.
    - A cross-binary replay is refused by design, because the build revisions differ, so case 3's "both binaries" part is covered by each binary replaying its own records.
  - Board directive #26 is complete.
