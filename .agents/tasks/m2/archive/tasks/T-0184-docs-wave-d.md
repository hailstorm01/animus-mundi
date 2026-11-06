---
id: T-0184
title: The docs say what wave D built, doc.go, placement.go and ability.go comments (directive #121, harness, flash, comments only)
kind: harness
tier: flash
wave: D.1
status: done
branch: m2/autoloop-wave-d.1/T-0184-docs-wave-d
depends_on: [T-0183]
claims: [comments:pkg/engine/doc.go, comments:pkg/engine/placement.go costs, comments:pkg/engine/ability/ability.go CostKind and Subject]
cards: []
requirements: []
---

## Goal
Comments that wave D and wave D.1 left stale say what the code does now (#121 (a), (b), (c)). No code changes: same declarations, same tests, same sim output. The lead made #121 (d), the design note's edits, in the commit before this task's (18c12fe).

## Inputs
- Board directive #121 (structural review 8, finding 16, approved by the user, 2026-11-05). Its line numbers have moved; the table below has today's.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6g (wave D.1).
- The prompt at `pkg/engine/costs.go:36` stays, as #121 (a) says: no activated ability reaches it. `ExtraCosts`' comment (`ability.go:28`) and `doc.go`'s Maneuver bullet stay too, because "additional costs" is right for a Maneuver.
- Three lead decisions beyond #121's text, all comments:
  - **"Not built" keeps what is still unbuilt.** #117 built only a card's own Cornered keywords. *Brigand Highlander*'s "Cornered: This unit has +2 power." and auras are still not built, and both are on `cards.NotYetExecutable`, so the line names them instead of dropping the words outright.
  - **What runs now gains a bullet** for the Cornered keywords, so the shape #121 (c) takes off the "not built" list is listed as built.
  - **Two more Files entries were wrong after #113 and #114:** `targeting.go`'s says "and filters", which `filter.go` now holds, and `state.go`'s omits `exhaust`. Each entry now matches its file's header.
- `stepSearchDeck` reads the subject through `stepPlayers` (`pkg/engine/effects.go:111`), so both `Subject` comments name `OpSearchDeck` (#121 (b)).
- Why it is comments only: the lead dry-ran these exact edits in a throwaway worktree of the run branch at 18c12fe. `gofmt -l` is clean, and parsing the three files on both sides with comments dropped gives the same SHA-256 values. `go tool task check` and `go tool task sim` exit 0, and `go tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`. Every expected output below is from that dry run.

### Where it is today (at `autoloop/m2-wave-d.1`, 18c12fe)
| Where | Today |
| :--- | :--- |
| `pkg/engine/doc.go:50` | the bullet `//   - Ward on targets, ...`, which the new bullet goes before |
| `pkg/engine/doc.go:61` | `//   - Continuous abilities, continuous Cornered,` |
| `pkg/engine/doc.go:75-77` | the Files lines for `targeting.go`, `effects.go` and `state.go` |
| `pkg/engine/placement.go:21` | the `costs` field's trailing comment |
| `pkg/engine/placement.go:45` | the last line of `declare`'s comment |
| `pkg/engine/placement.go:111` | the second line of `placeable`'s comment |
| `pkg/engine/ability/ability.go:56` | `CostKind`'s comment |
| `pkg/engine/ability/ability.go:270`, `:272` | `SubjectController`'s and `SubjectOpponents`' trailing comments |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/engine/doc.go`, What runs now
Add these three lines right before the line `//   - Ward on targets, and last-known effective state for effects whose`:
```go
//   - A card's own Cornered keywords ("Cornered: Aggressive 3.", Def.Cornered):
//     g.Keywords adds them as a Layer 3 grant on a face-up asset while its controller
//     is Cornered (FR-LAY-001, FR-KWD-002e, SVY-R18).
```

### 2. `pkg/engine/doc.go`, "Not built yet"
Replace the line `//   - Continuous abilities, continuous Cornered,` with these two lines. The next line, `//     Regroup N, ...`, stays:
```go
//   - Continuous abilities other than a card's own Cornered keywords (auras, and
//     "Cornered: This unit has +2 power."),
```

### 3. `pkg/engine/doc.go`, the Files entries
Replace these three lines:
```go
//     a placement's costs. targeting.go: target choice and filters. effects.go:
//     running a resolved ability's steps.
//   - state.go: core state types and lookups. card.go: card types.
```
with these five:
```go
//     a placement's costs. resources.go: Requisition and Tactics, generating and
//     converting them, and paying with them. targeting.go: target choice and
//     untargeted picks. filter.go: checking assets and discard-pile cards against
//     an ability's filter. effects.go: running a resolved ability's steps.
//   - state.go: core state types and lookups, and exhausting an asset. card.go: card types.
```

### 4. `pkg/engine/placement.go`, line 21
Only the trailing comment changes. The line becomes:
```go
	costs      []ability.Cost // costs paid with this placement: a Maneuver's additional costs (FR-TRN-003a) or an activated ability's (FR-ACT-001); a trigger has none
```

### 5. `pkg/engine/placement.go`, line 45
Replace `// and the additional costs are paid (FR-TRN-003, FR-TRN-003a).` with:
```go
// and the costs are paid (FR-TRN-003, FR-TRN-003a, FR-ACT-001).
```

### 6. `pkg/engine/placement.go`, line 111
Replace `// (SVY-R29): every additional cost payable (costsPayable), then at least` with:
```go
// (SVY-R29): every cost payable (costsPayable), then at least
```

### 7. `pkg/engine/ability/ability.go`, line 56
Replace `// CostKind is what an additional cost asks the payer to do. It has no default.` with:
```go
// CostKind is what a cost asks the payer to do. It has no default.
```

### 8. `pkg/engine/ability/ability.go`, lines 270 and 272
Only the trailing comments change, and the code and the alignment stay:
- On line 270, replace `(the zero value; OpScratch, OpDraw, OpDiscard)` with `(the zero value; OpScratch, OpDraw, OpDiscard, OpSearchDeck)`.
- On line 272, replace `each opponent of the ability's controller (OpScratch, OpDraw, OpDiscard)` with `each opponent of the ability's controller (OpScratch, OpDraw, OpDiscard, OpSearchDeck)`.

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: comments only. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#121): `git -C W diff autoloop/m2-wave-d.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#121): `go -C W tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`.
3. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.1 --numstat -- pkg` prints exactly these three lines: `3	3	pkg/engine/ability/ability.go`, `10	4	pkg/engine/doc.go` and `3	3	pkg/engine/placement.go`.
4. **"Additional costs" stays only on Maneuvers** (#121 (a)): `git -C W grep -n -F 'additional cost' -- pkg/engine/placement.go pkg/engine/ability/ability.go` prints exactly two lines, at `ability.go:28` and `placement.go:21`.
5. **The subjects name the search** (#121 (b)): `git -C W grep -n -F 'OpDiscard, OpSearchDeck)' -- pkg/engine/ability/ability.go` prints exactly two lines, at `ability.go:270` and `:272`.
6. **`doc.go` names what wave D.1 built** (#121 (c)): `git -C W grep -n -e filter.go -e resources.go -e Def.Cornered -e 'exhausting an asset' -- pkg/engine/doc.go` prints exactly four lines, at `doc.go:50`, `:79`, `:81` and `:83`.
7. **`doc.go`'s stale words are gone** (#121 (c)): `git -C W grep -n -e 'continuous Cornered' -e 'choice and filters' -- pkg/engine/doc.go` prints nothing (git exits 1).
8. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
9. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.1 --stat` lists only the three files in case 3 and this task file.
10. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/doc.go`
- `pkg/engine/placement.go`
- `pkg/engine/ability/ability.go`

## Done When
- [x] Cases 1–10 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-d.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-05, lead-architect (wave D.1 run, session 15): wrote this task from directive #121 against the run branch at 18c12fe; `main` has no commits to sync. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$` or unquoted parentheses. Checked each claim against the code: `CorneredKeywords` is read in `g.Keywords`, `stepSearchDeck` goes through `stepPlayers`, and `filter.go`'s, `resources.go`'s, `targeting.go`'s and `state.go`'s headers say what the Files entries now say. #121 (d) landed in 18c12fe, which also lists `Cornered` on the design note's `Def` line. Next: `kernel-engineer` on Flash.
- 2026-11-05, kernel-engineer (T-0184): implemented comment updates for directive #121 ((a), (b), (c)) in `pkg/engine/doc.go`, `pkg/engine/placement.go`, and `pkg/engine/ability/ability.go`. Diff shows comments-only updates: doc.go adds Cornered keywords under "What runs now", refines continuous abilities under "Not built yet", and lists `resources.go`, `filter.go`, `targeting.go` and `state.go` accurately; placement.go updates costs comments for activated abilities and placements; ability.go updates CostKind and mentions OpSearchDeck for SubjectController and SubjectOpponents.
  Reused: existing documentation structures and wave D/D.1 shapes; zero code or behavior changes.
  Commands executed:
  - `git -C W diff autoloop/m2-wave-d.1 -- "*_test.go"`: PASS (no output, tests unchanged)
  - `go -C W test ./...`: PASS (all tests pass)
  - `go -C W tool task sim-compare -- autoloop/m2-wave-d.1`: PASS ([PASS] starter and stress sims match autoloop/m2-wave-d.1 byte for byte, exit codes included)
  - `git -C W diff autoloop/m2-wave-d.1 --numstat -- pkg`: PASS (3 3 pkg/engine/ability/ability.go, 10 4 pkg/engine/doc.go, 3 3 pkg/engine/placement.go)
  - `git -C W grep -n -F "additional cost" -- pkg/engine/placement.go pkg/engine/ability/ability.go`: PASS (lines ability.go:28 and placement.go:21)
  - `git -C W grep -n -F "OpDiscard, OpSearchDeck)" -- pkg/engine/ability/ability.go`: PASS (lines ability.go:270 and ability.go:272)
  - `git -C W grep -n -e filter.go -e resources.go -e Def.Cornered -e "exhausting an asset" -- pkg/engine/doc.go`: PASS (lines doc.go:50, :79, :81, :83)
  - `git -C W grep -n -e "continuous Cornered" -e "choice and filters" -- pkg/engine/doc.go`: PASS (exit code 1, 0 matches)
  - `go -C W run ./cmd/cardtown lint fences`: PASS ([PASS] Package code is free of card names, invalid markers and disallowed imports)
  - `git -C W diff autoloop/m2-wave-d.1 --stat`: PASS (lists only the 3 pkg files and the task file)
  - `go -C W tool task check`: PASS (gofmt, go vet, all tests, cardtown lint --all passed)
  - `go -C W tool task sim`: PASS (200 games, all invariants held)
  Moved task to `status: review`.
- 2026-11-05, lead-architect (wave D.1 run, session 15, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-d.1` as 587cc92. Flash run `df54bbea-2f7f-4cf8-b121-3ca9a3b1bb7f`: exit 0, 225 s, no refusals (no `denied_actions` key; the transcript has no step with `"status":"ERROR"` and no `permission check failed` match; the `.err` file holds only agy's "root agent idle" notice). Review of `git diff autoloop/m2-wave-d.1...m2/autoloop-wave-d.1/T-0184-docs-wave-d` (23a71c3):
  - Scope: only the three Allowed Files and this file changed, and this file's only other changes are its status, Log and ticked boxes. No test file changed.
  - Comments only, checked against the diff: all three files are byte-identical to the lead's dry run. Parsing each on both sides with comments dropped and hashing the printed code gives the same three SHA-256 values (a throwaway `go/parser` checker in the scratch folder, not committed).
  - Fences and placement: no card name in engine code, no new rule, no code.
  - Re-run by the reviewer in a scratch worktree of the task branch, all as expected: cases 1 and 3-9; `go test ./...` passes; `cardtown lint fences` `[PASS]`; `go tool task check` exit 0; `go tool task sim` clean, 200 games, P0 84.0% / P1 15.0% / draws 1.0%, 47/50 and 34/50 main cards executable; `go tool task sim-compare -- autoloop/m2-wave-d.1` `[PASS]`.
  - `go tool task check` and `go tool task sim` also pass on the merged run branch at 587cc92, and `cardtown lint frd` passes with the design note's edits (18c12fe, #121 (d)). Directive #121 is applied.
