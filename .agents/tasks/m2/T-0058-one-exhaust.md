---
id: T-0058
title: One place exhausts an asset, (g *Game) exhaust (directive #49, harness, flash, pure move)
kind: harness
tier: flash
wave: none (board directive #49, structural review 3 finding 1, before T-0055)
status: review
branch: m2/T-0058-one-exhaust
depends_on: [T-0057]
claims: [helper:engine.exhaust]
cards: []
requirements: []
---

## Goal
A readied asset becomes Exhausted in exactly one function, `(g *Game) exhaust(a *Asset)`. Generate, March, Scratch and Attack call it in place of their own `Exhausted = true` writes. Animoceps Shambler's exhaust hook (`SVY-R17`: every change from Readied to Exhausted, for any reason) is added inside it later, by the Shambler shape task, and fires on all four. Nothing else changes: same tests, same sim output.

## Inputs
- Board directive #49 (structural review 3, finding 1).
- Ruling: `SVY-R17` (FRD 01 v3.11, `FR-CAP-001`). It is why one place is needed. This task does **not** add the hook (directive #50 covers the hook task).
- Pattern to follow: T-0048 and T-0041 (pure moves, checked by byte-identical outputs).

### The four writes (at `main` 9a38051)
| Action | Where | Today |
| :--- | :--- | :--- |
| Generate | `pkg/engine/actions.go:65`, in `Apply` | `as.Exhausted = true` |
| March | `pkg/engine/actions.go:86`, in `Apply` | `as.Exhausted = true` |
| Scratch | `pkg/engine/actions.go:91`, in `Apply` | `p.HQ.Exhausted = true` (the Headquarters is an `*Asset`) |
| Attack | `pkg/engine/combat.go:10`, in `declareAttack` | `att.Exhausted = true` |

Each one becomes `g.exhaust(<same asset>)`, on the same line position, so the order of everything around it is unchanged.

### Left alone (lead decision)
- Entering the battlefield Exhausted, `pkg/engine/stack.go:105` (`resolveDeploy`) and `pkg/engine/zones.go:41` (`scratch`). `SVY-R17` says entering Exhausted is not a change from Readied, so these stay plain assignments from `DeploysExhausted`.
- Readying, `pkg/engine/turn.go:11`.
- `pkg/engine/snapshot.go` (copies the field) and every read of `Exhausted`.

### The helper (lead decision; copy exactly)
In `pkg/engine/actions.go`, directly after the closing brace of `Apply`:
```go
// exhaust is the one place a Readied asset becomes Exhausted, for any reason
// (SVY-R17). Entering the battlefield Exhausted is not an exhaust.
func (g *Game) exhaust(a *Asset) {
	a.Exhausted = true
}
```
The receiver is unused for now; the Shambler task fires the hook from it. Do not add a readied check, an event or a log line.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. Before any edit, build a baseline CLI from `main` and save its sim outputs. Record each check's commands and results in the Log; the reviewer re-runs them.
1. **Existing tests unchanged:** `git diff main -- '*_test.go'` is empty, and `go test ./...` passes.
2. **Starter sim byte-identical:** `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` from the baseline and the new binary: stdout+stderr and exit code match (`cmp`).
3. **Stress sim byte-identical:** the same for `cardtown sim` with no `-deck`.
4. **One write:** `git grep -n '\.Exhausted = true' -- 'pkg/engine/*.go' ':!pkg/engine/*_test.go'` prints exactly one line, inside `exhaust` in `actions.go`. (The escaped dot keeps out `kw.DeploysExhausted = true` in `cardload.go`, which the directive's pattern would also match.)
5. **Four calls:** `git grep -n "g.exhaust(" -- 'pkg/engine/*.go' ':!pkg/engine/*_test.go'` prints exactly four lines: three in `Apply` (Generate, March, Scratch) and one in `declareAttack`.
6. **Only these edits:** `git diff main --stat` lists only `pkg/engine/actions.go`, `pkg/engine/combat.go` and this task file.
7. `go tool task check` and `go tool task sim` pass.

## Allowed Files
- `pkg/engine/actions.go` (the helper, and the three writes in `Apply`)
- `pkg/engine/combat.go` (the write in `declareAttack`)

## Done When
- [x] Cases 1–6 hold, with their commands and results in the Log
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-24 lead-architect: Wrote the task from board directive #49, the first of the four directives the user put before T-0055 (#49, #52, #53, #54, in that order). Harness, tier flash, pure move, so it skips `tests` and is `ready` on branch `m2/T-0058-one-exhaust` (cut from `main` at this commit). Checked the directive's line references against `main` 9a38051; they still hold. Waiting triggers checked: #6 and #7 (mode is `serial`), #8 (no `deck.go` behavior change), #15 (no turn hook), #20 (no wave C shape task), #50 (no task claims the exhaust hook; this one claims only the helper), #51 (no "if you do" task) and #55 (#53's task not written). None has fired. Starter-deck cards not yet executable: 28, plus both Headquarters.
  - Next: `agy` is not installed in the lead's environment, so **the user runs `@kernel-engineer T-0058` in the Antigravity agent (Gemini Flash)** (D3 fallback).
- 2026-10-24 kernel-engineer: Pure move done. Baseline CLI `cardtown-baseline.exe` built from `main` (4a132c8) before edits and baseline starter/stress outputs captured. Status set to `in-progress` in 94484b5.
  - Reused / checked: Searched `pkg/engine` for `exhaust(` before implementing; none existed. Confirmed only four writes existed for `.Exhausted = true` (Generate, March, Scratch, Attack).
  - Implemented: Added `(g *Game) exhaust(a *Asset)` in `pkg/engine/actions.go` immediately after `Apply`. Replaced `as.Exhausted = true` in Generate and March with `g.exhaust(as)`, replaced `p.HQ.Exhausted = true` in Scratch with `g.exhaust(p.HQ)`, and replaced `att.Exhausted = true` in `declareAttack` (`pkg/engine/combat.go`) with `g.exhaust(att)`.
  - Case 1 (Existing tests unchanged): `git diff main -- '*_test.go'` is empty; `go test ./...` passed (exit code 0).
  - Case 2 (Starter sim byte-identical): `fc.exe /b` between baseline and new binary `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` reported `FC: no differences encountered` (exit code 0).
  - Case 3 (Stress sim byte-identical): `fc.exe /b` between baseline and new binary `cardtown sim` reported `FC: no differences encountered` (exit code 0).
  - Case 4 (One write): `git grep -n '\.Exhausted = true' -- 'pkg/engine/*.go' ':!pkg/engine/*_test.go'` returned exactly 1 line: `pkg/engine/actions.go:106:	a.Exhausted = true`.
  - Case 5 (Four calls): `git grep -n "g.exhaust(" -- 'pkg/engine/*.go' ':!pkg/engine/*_test.go'` returned exactly 4 lines (`actions.go:65`, `actions.go:86`, `actions.go:91`, `combat.go:10`).
  - Case 6 (Only these edits): `git diff main --stat` lists only `pkg/engine/actions.go`, `pkg/engine/combat.go` and this task file.
  - Case 7 (Gates): `go tool task check` (fmt-check, vet, all tests, cardtown lint --all) passed cleanly; `go tool task sim` passed cleanly on starter decks.
  - Status set to `review`.
