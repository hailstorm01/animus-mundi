---
id: T-0240
title: "`doc.go` says what waves E to F.2 built, and `deployedTo` becomes `zoneMatches` (directive #154 (c), (d), (e), wave F.3 harness, flash)"
kind: harness
tier: flash
wave: F.3
status: done
branch: m2/autoloop-wave-f.3/T-0240-engine-docs-and-zone-matches
depends_on: [T-0239]
claims: [func:zoneMatches, doc:pkg/engine/doc.go]
cards: []
requirements: [FR-ZON-005]
---

## Goal
Package `engine`'s doc comment reads as the code does after waves E to F.2 and T-0239: "What runs now" names `Decision.Source` and `couldDoNothing`, and the Files list names `StackLabel`, `shownCard`, `SourceCard`, `couldDoNothing`, the `askFor` scope, `Modifier` and `Duration`. The zone check `deployedTo` is renamed `zoneMatches`, because filters read it too and it is no longer about deploying. Comments and one pure rename: sims are byte-identical. When this merges, #154 is `applied`.

## Inputs
- Board directive #154 (structural review 12, finding 2, approved by user 2026-11-14), `open: before the next Deliverables task`. Wave F.3, second and last (design note Deliverable 6l).
- **The lead's part is already on the run branch**, in the commit that wrote this task: the design note's (a) `Static` bullet (`StaticDeployPermission` is built), (b) Stack bullet (`Tracked` covers a pile trigger; `Locked`), (c) Choices bullet (`Decision.Source`, `couldDoNothing`), and (f) Open Question 10, re-checked 2026-11-14 and still holding. This task does (c)'s `doc.go` half, (d) and (e). Don't touch `docs/`.
- **One deviation from #154 (e):** it names the new function `inZone`, but package `engine` already has a test helper `inZone(list []*Instance, inst *Instance) bool` ([self_tracking_test.go:22](../../../../../pkg/engine/self_tracking_test.go)), called on 45 lines of earlier tasks' protected tests, so `inZone` doesn't build (`inZone redeclared in this block`, seen in the lead's dry run). Renaming that helper would edit 45 protected test lines for a name. The function is named `zoneMatches` instead, which meets the finding's point (it is no longer about deploying); `git grep -n zoneMatches` found nothing before. Every other part of (e) holds.
- **Scope of "`git grep -n deployedTo` finds nothing":** the board, the directive archive and done task files record the old name as history, and #154's own row and the design note's Deliverable 6l line name it, so the check is scoped to code: `git grep -n deployedTo -- pkg cmd` finds nothing.
- The three uses today (at 65d05bc): the declaration [zones.go:256-258](../../../../../pkg/engine/zones.go), [filter.go:152](../../../../../pkg/engine/filter.go) and [trigger.go:146](../../../../../pkg/engine/trigger.go). No test names it.
- Size (D2.4): four non-test files, no tests (comments and a rename; the gates and `sim-compare` are the check).
- **Expected output difference:** none. `go tool task sim-compare` passes.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 65d05bc: `gofmt -l pkg cmd` clean; `go tool task check` exit 0; `go tool task sim-compare -- autoloop/m2-wave-f.3` `[PASS]`; `cardtown lint fences` exit 0; `cardtown lint frd` `[PASS]`; `go tool task sim` exit 0. Blobs: `doc.go` b190a84, `filter.go` 8f449f3, `trigger.go` 6b782a0, `zones.go` 4e44d3f.

## Commits
1. `refactor(engine): doc.go says what waves E to F.2 built; deployedTo becomes zoneMatches (T-0240)` with Edits 1-4. There is no test commit: a harness task with no new behavior, whose check is the gates.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at 65d05bc. Every line in `doc.go` begins with `//` and a space, as shown.

### 1. `pkg/engine/zones.go`: the rename
Lines 256-258, which read
```go
// deployedTo reports whether an asset in zone at is in zone z, which for a deploy
// trigger is the zone its deploy or move put it in. ZoneAny always holds.
func deployedTo(at Zone, z ability.Zone) bool {
```
become
```go
// zoneMatches reports whether an asset in zone at is in zone z, the zone a shape
// names: a filter's zone, or for a deploy or move trigger the zone its deploy
// or move put it in. ZoneAny always holds.
func zoneMatches(at Zone, z ability.Zone) bool {
```

### 2. `pkg/engine/filter.go`, line 152
`	if f.InPile() || !deployedTo(a.Zone, f.Zone) {` becomes `	if f.InPile() || !zoneMatches(a.Zone, f.Zone) {`

### 3. `pkg/engine/trigger.go`, line 146
`		if (hook == ability.HookDeploy || hook == ability.HookMoves) && !deployedTo(src.zone, t.ToZone) {` becomes `		if (hook == ability.HookDeploy || hook == ability.HookMoves) && !zoneMatches(src.zone, t.ToZone) {`

### 4. `pkg/engine/doc.go`: five places
(a) Line 29, which reads
```go
//     FR-MOD-001 §5).
```
becomes
```go
//     FR-MOD-001 §5). An optional trigger that could do nothing if accepted
//     (every step acts on a target slot, and no slot has a legal entity) is
//     neither asked nor placed (couldDoNothing, SVY-R48).
```
(b) After line 31 (the long line that ends `and a look at the top cards of your own deck (DecideLook, SVY-R41–SVY-R43).`) add these four lines:
```go
//     Each decision names the card whose Deploy, Maneuver or ability is being
//     placed or resolved as it is asked, as both seats see it (Decision.Source,
//     FR-ZON-005), from the scope askFor opens, and its match record carries
//     that card's name (DecisionRecord.Source).
```
(c) Lines 85-89, which read
```go
//   - turn.go: beginTurn and endTurn. zones.go: moving cards and assets, deploying a card, and the zone a shape names.
//   - actions.go: Action, Apply, checks and LegalActions. combat.go: attack
//     legality and combat resolution. stack.go: the Resolution Stack.
//   - trigger.go: the pending-trigger buffer and firing. placement.go: placing
//     an effect, its modes and then its targets. costs.go: a card's cost and
```
become
```go
//   - turn.go: beginTurn and endTurn. zones.go: moving cards and assets, deploying a card, and the zone a shape names (zoneMatches).
//   - actions.go: Action, Apply, checks and LegalActions. combat.go: attack
//     legality and combat resolution. stack.go: the Resolution Stack, and
//     StackLabel, the one wording of what a stack item is.
//   - trigger.go: the pending-trigger buffer and firing. placement.go: placing
//     an effect, its modes and then its targets, the card it shows as its
//     source (shownCard, SourceCard), and whether an optional trigger could do
//     nothing (couldDoNothing). costs.go: a card's cost and
```
(d) Line 97, which reads
```go
//   - events.go: Event, the Observer, match records and emitting. choice.go: decisions.
```
becomes
```go
//   - events.go: Event, the Observer, match records and emitting. choice.go: decisions,
//     and the askFor scope that names their source card.
```
(e) Line 101, which reads
```go
//     the Cornered check and EffectiveState.
```
becomes
```go
//     the Cornered check and EffectiveState, and an asset's Modifier list with
//     each modifier's Duration.
```
No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The old name is gone from code:** `git -C W grep -n deployedTo -- pkg cmd` prints nothing (exit 1).
2. **The new name is in exactly five places:** `git -C W grep -n zoneMatches -- pkg cmd` prints exactly five lines: `pkg/engine/doc.go:91`, `pkg/engine/filter.go:152`, `pkg/engine/trigger.go:146`, `pkg/engine/zones.go:256` (the comment) and `pkg/engine/zones.go:259` (the declaration).
3. **doc.go names what the code has:** `git -C W grep -n -e couldDoNothing -e Decision.Source -e StackLabel -e SourceCard -e askFor -e "Modifier list" -- pkg/engine/doc.go` prints exactly eight lines, numbered 31, 35, 36, 94, 97, 98, 107 and 111.
4. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-f.3` prints `[PASS] starter and stress sims match autoloop/m2-wave-f.3 byte for byte, exit codes included`.
5. **Only the listed edits:** `git -C W diff autoloop/m2-wave-f.3 --numstat -- pkg cmd docs` prints exactly four lines: `17	6	pkg/engine/doc.go`, `1	1	pkg/engine/filter.go`, `1	1	pkg/engine/trigger.go` and `4	3	pkg/engine/zones.go`.
6. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
7. **Nothing else changed:** `git -C W diff autoloop/m2-wave-f.3 --stat` lists only the four files of case 5 and this task file.
8. `gofmt -l W/cmd W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/doc.go`
- `pkg/engine/zones.go`
- `pkg/engine/filter.go`
- `pkg/engine/trigger.go`

## Done When
- [x] Cases 1-8 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-f.3` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-14, lead-architect (wave F.3 run, session 2): written in full against 65d05bc for #154. Nothing on `main` to sync (`main` at d4b38f2 is the run branch's base); the scratch folder was empty. Not `WAVE-COMPLETE` before the task (#154 `open: before the next Deliverables task`). The lead did #154 (a), (b), the design note's half of (c), and (f) itself in this commit (`docs/architecture/` is the lead's to write; precedent #141 (4)). (f): Open Question 10 still holds; the reasoning is in the question. Harness, so it skips `tests` and goes straight to `ready`. Tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list: one per line, no pipes, no `$`, no unquoted parentheses. Waiting directives checked: #6, #7 (mode `serial`), #15 (no turn-boundary hook), #71, #91 (2), #106, #120, #141 (3), #142: not fired. Next: `kernel-engineer` on Flash.
- 2026-11-14, kernel-engineer (wave F.3 run, session 3): implemented edits 1-4 per task spec in scratch worktree.
  - Diff shows:
    - `pkg/engine/zones.go`: renamed `deployedTo` to `zoneMatches` and updated doc comment.
    - `pkg/engine/filter.go`: updated call to `zoneMatches`.
    - `pkg/engine/trigger.go`: updated call to `zoneMatches`.
    - `pkg/engine/doc.go`: updated package doc comment across 5 sections to document `couldDoNothing`, `Decision.Source` / `DecisionRecord.Source` / `askFor`, `zoneMatches`, `StackLabel`, `shownCard`, `SourceCard`, and `Modifier` / `Duration`.
  - Reuse: pure rename and comment docs update; no new shapes.
  - Acceptance cases verified:
    - Case 1: `git -C W grep -n deployedTo -- pkg cmd` printed nothing (exit 1).
    - Case 2: `git -C W grep -n zoneMatches -- pkg cmd` printed exactly five lines across `pkg/engine/doc.go:91`, `pkg/engine/filter.go:152`, `pkg/engine/trigger.go:146`, `pkg/engine/zones.go:256`, `pkg/engine/zones.go:259`.
    - Case 3: `git -C W grep -n -e couldDoNothing -e Decision.Source -e StackLabel -e SourceCard -e askFor -e "Modifier list" -- pkg/engine/doc.go` printed exactly eight lines (31, 35, 36, 94, 97, 98, 107, 111).
    - Case 4: `go -C W tool task sim-compare -- autoloop/m2-wave-f.3` reported `[PASS] starter and stress sims match autoloop/m2-wave-f.3 byte for byte, exit codes included`.
    - Case 5: `git -C W diff autoloop/m2-wave-f.3 --numstat -- pkg cmd docs` reported `17 6 pkg/engine/doc.go`, `1 1 pkg/engine/filter.go`, `1 1 pkg/engine/trigger.go`, `4 3 pkg/engine/zones.go`.
    - Case 6: `go -C W run ./cmd/cardtown lint fences` passed (exit 0).
    - Case 7: `git -C W diff autoloop/m2-wave-f.3 --stat` listed only the four code files and the task file.
    - Case 8: `gofmt -l W/cmd W/pkg` clean; `go -C W tool task check` passed; `go -C W tool task sim` passed (200/200 games clean, all invariants holding).
  - Status set to `review`. Ready for lead review.
- 2026-11-14, lead-architect (wave F.3 run, session 2, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-f.3` as abee8e2. Flash run `bdf9b330-3a91-4410-8117-ca4e352b5ef9`: exit 0, `status: SUCCESS`, no `denied_actions` key, and no step with `"status":"ERROR"` in its transcript; `.err` holds only agy's "root agent idle" notice. Review of `git diff autoloop/m2-wave-f.3...m2/autoloop-wave-f.3/T-0240-engine-docs-and-zone-matches` (9ab8a1f):
  - Scope: the four Allowed Files and this file, in one implementation commit (eb71703), whose blobs are the lead's dry run's (`doc.go` b190a84, `filter.go` 8f449f3, `trigger.go` 6b782a0, `zones.go` 4e44d3f). No test changed.
  - Fences and placement: comments and one unexported rename in package `engine`; no card name. Reuse: nothing added.
  - Re-run by the reviewer in a scratch worktree of the task branch: case 1 `git grep -n deployedTo -- pkg cmd` exit 1; case 2 five lines as listed; case 3 lines 31, 35, 36, 94, 97, 98, 107, 111; `gofmt -l cmd pkg` clean; `cardtown lint fences` exit 0; `cardtown lint frd` `[PASS]`; `go tool task check` exit 0; `go tool task sim-compare -- autoloop/m2-wave-f.3` `[PASS]`; `go tool task sim` exit 0, 200 games, P0 58.5% / P1 39.5% / draws 2.0%, P0 50/50 and P1 50/50 main cards executable.
  - The merge's tree matches the task tip's (`git diff` between them is empty), so the gates hold on the run branch. #154 is `applied` (the lead's commit 3fd8068 and T-0240) and moves to the directive archive.
