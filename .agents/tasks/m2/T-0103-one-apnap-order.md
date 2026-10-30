---
id: T-0103
title: One APNAP order, (g *Game) apnap, and replacedQueue takes "oldest first" from assetsByID (directive #77, harness, flash, pure move)
kind: harness
tier: flash
wave: G.1 (board directive #77, structural review 5, finding 2)
status: in-progress
branch: m2/T-0103-one-apnap-order
depends_on: [T-0102]
claims: [func:apnap, func:replacedQueue]
cards: []
requirements: []
---

## Goal
"The active player's, then the other player's" is written once, in `(g *Game) apnap() []PlayerID`, and "oldest first" is written once, in `assetsByID`. `flushTriggers` and `replacedQueue` both call `apnap`, and `replacedQueue` takes its order from `assetsByID` with no comparator of its own. The battlefield-orders comment names both replacement walks under `assetsByID`. This is a pure move: same tests, same sim output.

## Inputs
- Board directive #77 (structural review 5, approved by the user, 2026-10-29).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6d (wave G.1).
- Rules the moved code implements, unchanged: `FR-TRG-003` (APNAP placement), `SVY-R20` (oldest first), `SVY-R22` (the active player's replaced deaths first), `SVY-R11` (option lists and "all" steps by ID).
- Why it is a pure move:
  - `apnap` returns the same two-element slice both loops build inline today.
  - `replacedQueue` today keeps the assets of `dying` that a replacement applies to, splits them by controller, and sorts each player's by ID. Walking `assetsByID()` and keeping the assets that are in `dying` (by pointer) and that a replacement applies to gives each player's list already in ID order, so the split in `apnap` order is the same queue. `replacementsFor` reads state only, so calling it in a different order changes nothing.
  - Two inputs could differ, and neither changes the result. (1) An asset of `dying` that is not on the battlefield: today it is queued, but `replaceDeaths` skips it (`g.Asset(a.ID) != a`) and leaves it out of the rest; now it is never queued. (2) The same asset twice in `dying`: both callers build `dying` without repeats (`newlyLethal` walks `g.Assets()` once; `stepKill` takes one step's assets, from one target slot, `SubjectSelf`, `SubjectReplaced`, a `SubjectAll` walk or a pick).
  - IDs are assigned when an asset enters the battlefield (`zones.go:165`, a new `Asset` per entry) and never change when it moves between zones, so "lowest ID" is "entered the battlefield first".
  - `go tool task sim-compare` is the check.

### Where it is today (at `autoloop/m2-wave-g.1` 3741e27)
| Where | Today |
| :--- | :--- |
| `pkg/engine/state.go:202-205` | `newID` |
| `pkg/engine/state.go:207-226` | the battlefield-orders comment; its `assetsByID` paragraph is lines 219-220 |
| `pkg/engine/state.go:242-247` | `assetsByID`, the one ID comparator that stays |
| `pkg/engine/trigger.go:166` | `flushTriggers`: `for _, pid := range []PlayerID{g.Active, g.Active.Opponent()} {` |
| `pkg/engine/replace.go:40-58` | `replacedQueue`: its own APNAP slice (`:52`) and its own ID sort (`:54`) |

## Edits (lead decisions; copy exactly)

### 1. `pkg/engine/state.go`: `apnap`
Right after `newID`'s closing `}` and its blank line, before the `// Battlefield orders.` comment, add:
```go
// apnap returns the players in APNAP order: the active player, then the
// non-active player. Simultaneous triggers are placed in this order
// (FR-TRG-003), and replaced deaths are taken up in it (SVY-R22).
func (g *Game) apnap() []PlayerID {
	return []PlayerID{g.Active, g.Active.Opponent()}
}

```

### 2. `pkg/engine/state.go`: the battlefield-orders comment
Replace its `assetsByID` paragraph, these two lines:
```go
// assetsByID: oldest (lowest ID) first. Decision option lists and "all" steps
// use it (SVY-R11).
```
with:
```go
// assetsByID: oldest (lowest ID) first. "Oldest" means the asset that entered
// the battlefield first, not the one that reached its current zone first: an
// asset gets its ID when it enters the battlefield and keeps it. Decision
// option lists and "all" steps use it (SVY-R11), and so do replacementsFor
// (oldest source first) and replacedQueue (each player's oldest replaced asset
// first, the players in apnap order) (SVY-R20, SVY-R22).
```
Nothing else in the comment changes.

### 3. `pkg/engine/trigger.go`: `flushTriggers`
Replace line 166:
```go
	for _, pid := range []PlayerID{g.Active, g.Active.Opponent()} {
```
with:
```go
	for _, pid := range g.apnap() {
```
Nothing else in `trigger.go` changes. `slices` is still used there (`DeleteFunc`, `SortStableFunc`).

### 4. `pkg/engine/replace.go`: `replacedQueue`
Replace the whole function, doc comment included (lines 40-58), with:
```go
// replacedQueue returns the assets of dying that a replacement applies to, in
// the order their deaths are replaced: the players in apnap order, so the active
// player's first (SVY-R22), each player's oldest asset first (assetsByID,
// SVY-R20). No player orders them.
func (g *Game) replacedQueue(dying []*Asset) []*Asset {
	var queued [2][]*Asset
	for _, a := range g.assetsByID() {
		if slices.Contains(dying, a) && len(g.replacementsFor(a)) > 0 {
			queued[a.Controller] = append(queued[a.Controller], a)
		}
	}
	var out []*Asset
	for _, pid := range g.apnap() {
		out = append(out, queued[pid]...)
	}
	return out
}
```
`replace.go` keeps its `slices` import (`slices.Contains`).

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. Run each command on its own, pointed at your worktree (W). Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#77 case (a)): `git -C W diff autoloop/m2-wave-g.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#77 case (b)): `go -C W tool task sim-compare -- autoloop/m2-wave-g.1` prints `[PASS]`.
3. **One APNAP slice** (#77 case (c)): `git -C W grep -n "g.Active, g.Active.Opponent()" -- pkg` prints exactly one line, in `pkg/engine/state.go` (the body of `apnap`).
4. **One ID comparator** (#77 case (c)): `git -C W grep -n "x.ID - y.ID" -- pkg` prints exactly one line, `pkg/engine/state.go`, inside `assetsByID`.
5. **Both walks call `apnap`:** `git -C W grep -n "g.apnap()" -- pkg/engine` prints exactly two lines, one in `pkg/engine/trigger.go` and one in `pkg/engine/replace.go`.
6. **The comment names the walks:** `git -C W grep -n "replacedQueue (each player's oldest" -- pkg/engine/state.go` prints one line.
7. **Only these edits:** `git -C W diff autoloop/m2-wave-g.1 --stat` lists only `pkg/engine/replace.go`, `pkg/engine/state.go`, `pkg/engine/trigger.go` and this task file.
8. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/replace.go`
- `pkg/engine/state.go`
- `pkg/engine/trigger.go`

## Done When
- [ ] Cases 1–8 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-g.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-29, lead-architect (wave G.1 run, session 2): wrote the task from board directive #77, the next of #76 to #79 in board order now that #76 is applied (T-0102). Harness, so it skips `tests` and goes straight to `ready`; tier `flash` as #77 gives; every command here is on the headless allow-list. The two `{p, p.Opponent()}` loops in `state.go:277` (`fixedOrderPositions`) and `targeting.go:98` start from the choosing player, not the active one, so they are a different order (`SVY-R16`) and stay. `main` has no commits the run branch lacks, so no sync. Waiting directives checked: none has triggered (#8: no `deck.go`; #15, #20 (b), #58, #59 (2), #61, #71: no such hook, shape, claim or card here; #6, #7: mode is `serial`). #78, #79 stay `open: before the next Deliverables task`; #81 stays `open: after #79`. Next: `kernel-engineer` on Flash.
