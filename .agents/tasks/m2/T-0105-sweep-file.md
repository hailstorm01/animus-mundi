---
id: T-0105
title: The death sweep gets its own file, sweep.go (directive #79, harness, flash, pure move)
kind: harness
tier: flash
wave: G.1 (board directive #79, structural review 5, finding 4)
status: done
branch: m2/T-0105-sweep-file
depends_on: [T-0104]
claims: [file:pkg/engine/sweep.go]
cards: []
requirements: []
---

## Goal
`damage.go` has one job again, the damage pipeline (`FR-REP-004`). The death sweep, destruction and the match-loss checks (`destroyAll`, `deathSweep`, `newlyLethal`, `destroyDying`, `sweepDeaths`, `end`; `FR-LAY-002`, `FR-WIN-001`) move unchanged to a new file, `pkg/engine/sweep.go`. `doc.go`'s file map names `sweep.go`, and `placement.go`, which it left out (directive #80). This is a pure move: same tests, same sim output.

## Inputs
- Board directive #79 (structural review 5, approved by the user, 2026-10-29), and #80's `doc.go` line.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6d (wave G.1).
- Rules the moved code implements, unchanged: `FR-LAY-002` (the state-based death sweep), `FR-WIN-001` (match loss), `FR-ZON-004` (last-known state at departure), `SVY-R20` to `SVY-R22`.
- #79's third bullet (an `OPEN(SVY-R24)` comment in `sweepDeaths`) is dropped: survey R24 is ruled (FRD 01 v3.18), and directive #81 changes that line after this task. This task moves the code as it is, comments included.
- Why it is a pure move: the six functions keep their bodies and doc comments byte for byte. They stay in package `engine` and use no imports (`damage.go` has none either), so nothing else changes.

### Where it is today (at `autoloop/m2-wave-g.1` aa0772b)
| Where | Today |
| :--- | :--- |
| `pkg/engine/damage.go:1-63` | `package engine`, `DamageResult`, `Taken`, `dealDamage`: these stay |
| `pkg/engine/damage.go:64` | a blank line |
| `pkg/engine/damage.go:65-171` | `destroyAll` (doc comment from line 65), `deathSweep`, `newlyLethal`, `destroyDying`, `sweepDeaths`, `end` (line 158 to the file's last line, 171): these move |
| `pkg/engine/doc.go:52-64` | the `Files:` paragraph of the package comment |

## Edits (lead decisions; copy exactly)

### 1. New file `pkg/engine/sweep.go`
Its first four lines are:
```go
// Death sweep: destroying lethal assets, replaced deaths first, and the match-loss checks (FR-LAY-002, FR-WIN-001).

package engine

```
Then, unchanged and in the same order, `damage.go` lines 65-171: from the line `// destroyAll destroys the assets as one simultaneous event. Every asset's` through the closing `}` of `end`. No import block. The file ends with that `}` and one newline.

### 2. `pkg/engine/damage.go`
Delete lines 64-171: the blank line after `dealDamage`'s closing `}`, and everything after it. The file then ends with `dealDamage`'s closing `}` (line 63) and one newline. Nothing else in `damage.go` changes.

### 3. `pkg/engine/doc.go`: the file map
Replace this line:
```go
//   - trigger.go: the pending-trigger buffer and firing. targeting.go: target
```
with these two lines:
```go
//   - trigger.go: the pending-trigger buffer and firing. placement.go: placing
//     an effect, its modes and then its targets. targeting.go: target
```
and replace this line:
```go
//     damage.go: the damage pipeline. replace.go: destruction replacement. layers.go: derived power, health and keywords.
```
with these two lines:
```go
//     damage.go: the damage pipeline. sweep.go: the death sweep, destruction
//     and match loss. replace.go: destruction replacement. layers.go: derived power, health and keywords.
```
Nothing else in `doc.go` changes.

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. Run each command on its own, pointed at your worktree (W). Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#79 case (a)): `git -C W diff autoloop/m2-wave-g.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#79 case (b)): `go -C W tool task sim-compare -- autoloop/m2-wave-g.1` prints `[PASS]`.
3. **`damage.go` keeps only the pipeline** (#79 case (c)): `git -C W grep -n -e "^func " -e "^type " -e "^var " -e "^const " -- pkg/engine/damage.go` prints exactly three lines: `type DamageResult struct {`, `func (r DamageResult) Taken() bool …` and `func (g *Game) dealDamage(…`.
4. **`sweep.go` holds the six:** `git -C W grep -n -e "^func " -e "^type " -e "^var " -e "^const " -- pkg/engine/sweep.go` prints exactly six lines, in this order: `destroyAll`, `deathSweep`, `newlyLethal`, `destroyDying`, `sweepDeaths`, `end`.
5. **Moved, not changed:** `git -C W diff autoloop/m2-wave-g.1 --numstat -- pkg/engine/damage.go pkg/engine/sweep.go` prints `0	108	pkg/engine/damage.go` and `111	0	pkg/engine/sweep.go` (107 moved lines, plus the deleted blank line and the new file's four header lines).
6. **The file map is complete** (#79 case (d)): `git -C W ls-files pkg/engine` lists the files; every one directly in `pkg/engine` (not in `ability/` or `cards/`) that is not `doc.go` and does not end in `_test.go` appears in `doc.go`'s `Files:` paragraph as `<name>.go:`. In particular, `git -C W grep -n -e "sweep.go:" -e "placement.go:" -- pkg/engine/doc.go` prints two lines.
7. **Fences** (#79 case (e)): `go -C W run ./cmd/cardtown lint fences` exits 0.
8. **Only these edits:** `git -C W diff autoloop/m2-wave-g.1 --stat` lists only `pkg/engine/damage.go`, `pkg/engine/doc.go`, `pkg/engine/sweep.go` and this task file.
9. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/damage.go`
- `pkg/engine/doc.go`
- `pkg/engine/sweep.go` (new)

## Done When
- [x] Cases 1–9 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-g.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-29, lead-architect (wave G.1 run, session 4): wrote the task from board directive #79, the last of #76 to #79 in board order now that #78 is applied (T-0104). Harness, so it skips `tests` and goes straight to `ready`; tier `flash` as #79 gives; every command here is on the headless allow-list, with no `|`, parentheses or `$` (case 3, 4 and 6 use `-e` patterns). #79's `OPEN(SVY-R24)` bullet is dropped as #81 says, since R24 is ruled (FRD 01 v3.18). `doc.go`'s line for `placement.go` is #80's. `main` (556811f) has no commits the run branch lacks, so no sync. Waiting directives checked: none has triggered (#8: no `deck.go`; #15, #20 (b), #58, #59 (2), #61, #71: no such hook, shape, claim or card here; #6, #7: mode is `serial`). #81 stays `open: after #79`, so it is the next task once this one is `done`. Next: `kernel-engineer` on Flash.
- 2026-10-29, kernel-engineer (session 1, flash): implemented T-0105 per directive #79 and task specification.
  - Step 1: Loaded task in scratch worktree and set status: in-progress (commit 65b56bc).
  - Step 3 (Reuse): Pure move of existing death sweep functions (`destroyAll`, `deathSweep`, `newlyLethal`, `destroyDying`, `sweepDeaths`, `end`). No new shapes added.
  - Step 5 (Implementation): Created `pkg/engine/sweep.go` with 4-line header followed by `damage.go` lines 65-171 byte-for-byte; trimmed `pkg/engine/damage.go` to line 63 ending after `dealDamage`; updated `pkg/engine/doc.go` file map with `placement.go` and `sweep.go`. Committed as `feat(engine): move death sweep to sweep.go (T-0105)` (commit 3590cdb).
  - Acceptance Cases:
    1. Tests unchanged: `git -C W diff autoloop/m2-wave-g.1 -- "*_test.go"` printed nothing. `go -C W test ./...` passed (33s).
    2. Sims byte-identical: `go -C W tool task sim-compare -- autoloop/m2-wave-g.1` passed: `[PASS] starter and stress sims match autoloop/m2-wave-g.1 byte for byte, exit codes included`.
    3. Pipeline only in damage.go: `git -C W grep -n -e "^func " -e "^type " -e "^var " -e "^const " -- pkg/engine/damage.go` printed exactly 3 lines (`type DamageResult struct {`, `func (r DamageResult) Taken() bool`, `func (g *Game) dealDamage(...) DamageResult`).
    4. Six functions in sweep.go: `git -C W grep -n -e "^func " -e "^type " -e "^var " -e "^const " -- pkg/engine/sweep.go` printed exactly 6 lines in order (`destroyAll`, `deathSweep`, `newlyLethal`, `destroyDying`, `sweepDeaths`, `end`).
    5. Moved not changed: `git -C W diff autoloop/m2-wave-g.1 --numstat -- pkg/engine/damage.go pkg/engine/sweep.go` printed `0 108 pkg/engine/damage.go` and `111 0 pkg/engine/sweep.go`.
    6. Complete file map: `git -C W grep -n -e "sweep.go:" -e "placement.go:" -- pkg/engine/doc.go` printed 2 lines. Every non-test, non-doc .go file directly in `pkg/engine` is listed in `doc.go`.
    7. Fences: `go -C W run ./cmd/cardtown lint fences` passed (`[PASS] Package code is free of card names, invalid markers and disallowed imports.`).
    8. Only these edits: `git -C W diff autoloop/m2-wave-g.1 --stat` listed only `pkg/engine/damage.go`, `pkg/engine/doc.go`, `pkg/engine/sweep.go`, and the task file.
    9. Standard gates: `go -C W tool task check` (gofmt, vet, test, all linters) and `go -C W tool task sim` (starter decks 200 matches) passed cleanly.
  - Reviewer notes: Diff is a pure file split and documentation map update; no logic or tests were altered. Ready for lead-architect review.
- 2026-10-29, lead-architect (wave G.1 run, session 4, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-g.1` as dd82c77. Flash run `4a25adb4-6a34-4a59-8382-190615af9eae`: exit 0, `status: SUCCESS`, 312 s, no refusals (the JSON has no `denied_actions` entry, `cli.log` has no `soft-denying` line and the transcript no `permission check failed`). Review of `git diff autoloop/m2-wave-g.1...m2/T-0105-sweep-file` (81bf1bb):
  - Scope: only `damage.go`, `doc.go`, `sweep.go` and this file changed. No test file changed.
  - Pure move, checked byte for byte: `sweep.go` after its four header lines equals the run branch's `damage.go` lines 65-171, and the new `damage.go` equals the old one's lines 1-63. `doc.go` changes only the two map lines in Edit 3.
  - Fences and placement: no card name, no new rule; `sweep.go` stays in package `engine`, which the Boundaries table gives the rules. Every non-test file directly in `pkg/engine` other than `doc.go` is named in the file map (checked against `git ls-files`).
  - Gates re-run in a scratch worktree of the task branch, each with exit 0: case 1 (no `_test.go` diff); cases 3 and 4 (three and six declarations, as listed); case 5 (`0 108` and `111 0`); case 7 `cardtown lint fences`; `go tool task sim-compare -- autoloop/m2-wave-g.1` `[PASS]` (starter and stress byte-identical); `go tool task check` passes; `go tool task sim` clean, 200 games, P0 67.0% / P1 32.5% / draws 0.5% (unchanged), 32/50 Animus and 25/50 Path main cards executable.
  - The merge's tree equals the task tip's, so the gates hold on the run branch. Directive #79 is applied, and with it #76 to #79. #81 is now `open: before the next Deliverables task`.
