---
id: T-0065
title: One place builds an asset entering the battlefield, (g *Game) newAsset (directive #56, harness, flash, pure move)
kind: harness
tier: flash
wave: none (board directive #56, FRD 01 review finding 4, before the next Deliverables task)
status: done
branch: m2/T-0065-new-asset
depends_on: []
claims: [helper:engine.newAsset]
cards: []
requirements: []
---

## Goal
Every asset that enters the battlefield is built by one function, `(g *Game) newAsset`, which takes a new ID and sets the entry state from the printed keywords. Today three places build an asset, and they have drifted apart: `resolveDeploy` sets `Exhausted` and `Armor`, `scratch` sets only `Exhausted`, and Headquarters setup sets neither. Directive #57 will add an arrival number to entering assets, and it needs one place to set it. Nothing else changes: same tests, same sim output.

## Inputs
- Board directive #56 (FRD 01 review, finding 4). The directive names two places; the lead found a third, Headquarters setup, which builds its asset the same way.
- Rulings in the new comment: `FR-KWD-002l` (Deploys Exhausted), `FR-ZON-005` (the Scratch Company face prints it), `FR-LAY-002` and `FR-KWD-002k` (Armor N fills the Armor pool on entry). These are the IDs the current comments cite.
- Pattern to follow: T-0058 and T-0061 (pure moves).

### Why the output can't change (at `main` bd7ed10)
- The Scratch Company face (`scratchFace`, `pkg/engine/card.go:82-91`) prints no Armor, so `scratch` setting `Armor` sets 0.
- No Headquarters card prints Armor or Deploys Exhausted (`data/cards/units-and-structures-HQ.csv`), and the test helper `hq` (`scenario_test.go:14`) sets no keywords.
- Headquarters setup gives the asset its ID before its instance gets one. Edit 4 keeps that order, so every ID in the sims stays the same.

## Edits (lead decisions; copy exactly)

### 1. `pkg/engine/zones.go`: add `newAsset`
Insert this function, followed by one blank line, directly above `// placeOnBattlefield puts a into zone under its controller.`:
```go
// newAsset builds the asset inst becomes as it enters the battlefield under
// controller, with a new ID. Every way onto the battlefield builds its asset
// here. The entry state comes from the printed keywords: Deploys Exhausted
// replaces entering Readied (FR-KWD-002l; the Scratch Company face prints it,
// FR-ZON-005), and Armor N fills the Armor pool (FR-LAY-002, FR-KWD-002k).
func (g *Game) newAsset(inst *Instance, controller PlayerID, faceDown bool) *Asset {
	a := &Asset{ID: g.newID(), Inst: inst, Controller: controller, FaceDown: faceDown}
	kw := a.Printed().Keywords
	a.Exhausted = kw.DeploysExhausted
	a.Armor = kw.Armor
	return a
}
```

### 2. `pkg/engine/zones.go`: `scratch` uses it
In `scratch`, these three lines
```go
		a := &Asset{ID: g.newID(), Inst: inst, Controller: pid, FaceDown: true}
		// The face prints Deploys Exhausted, so it enters Exhausted (FR-ZON-005, FR-KWD-002l).
		a.Exhausted = a.Printed().Keywords.DeploysExhausted
```
become
```go
		a := g.newAsset(inst, pid, true)
```

### 3. `pkg/engine/stack.go`: `resolveDeploy` uses it
In `resolveDeploy`, these five lines
```go
	a := &Asset{ID: g.newID(), Inst: it.Inst, Controller: it.Controller}
	// Deploys Exhausted replaces entering Readied (FR-KWD-002l).
	a.Exhausted = c.Keywords.DeploysExhausted
	// Armor N initializes the Armor pool on entry (FR-LAY-002, FR-KWD-002k).
	a.Armor = c.Keywords.Armor
```
become
```go
	a := g.newAsset(it.Inst, it.Controller, false)
```
Keep `c := it.Inst.Card`: the log event still uses `c.Name`.

### 4. `pkg/engine/setup.go`: Headquarters setup uses it
In `NewGame`, the line
```go
		hq := &Asset{ID: g.newID(), Inst: &Instance{ID: g.newID(), Card: deck.HQ, Owner: pid}, Controller: pid}
```
becomes
```go
		// The asset takes its ID before its instance does, so every ID stays as it was.
		inst := &Instance{Card: deck.HQ, Owner: pid}
		hq := g.newAsset(inst, pid, false)
		inst.ID = g.newID()
```
Do not write `g.newAsset(&Instance{ID: g.newID(), …}, …)`: Go evaluates the argument first, which would swap the two IDs and change the sim output.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. Run each command on its own, pointed at your worktree. Record each command and its result in the Log; the reviewer re-runs them.
1. **Existing tests unchanged:** `git -C <worktree> diff main -- '*_test.go'` prints nothing, and `go -C <worktree> test ./...` passes.
2. **Sims byte-identical:** `go -C <worktree> tool task sim-compare` prints `[PASS]`.
3. **One place builds an asset:** `git -C <worktree> grep -n "&Asset{" -- pkg/engine ":!*_test.go"` prints exactly one line, in `pkg/engine/zones.go`, inside `newAsset`.
4. **Three callers:** `git -C <worktree> grep -n "g.newAsset(" -- pkg/engine` prints exactly three lines: one each in `pkg/engine/setup.go`, `pkg/engine/stack.go` and `pkg/engine/zones.go`.
5. **Only these edits:** `git -C <worktree> diff main --stat` lists only `pkg/engine/setup.go`, `pkg/engine/stack.go`, `pkg/engine/zones.go` and this task file.
6. `go -C <worktree> tool task check` and `go -C <worktree> tool task sim` pass.

## Allowed Files
- `pkg/engine/zones.go` (`newAsset` added, `scratch` uses it)
- `pkg/engine/stack.go` (`resolveDeploy` uses it)
- `pkg/engine/setup.go` (Headquarters setup uses it)

## Done When
- [x] Cases 1–5 hold, with their commands and results in the Log
- [x] `go tool task sim-compare` passes (starter and stress sims byte-identical to the integration branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-26 lead-architect: written from directive #56, with Headquarters setup added as a third caller (see Inputs). Harness pure move, so it goes straight to `ready`. Dispatching to `kernel-engineer` (flash) through `agy`. Next on the board: directive #57 (open), which builds on `newAsset`.
- 2026-10-26 kernel-engineer: implemented directive #56 / finding 4 pure move.
  - Added `(g *Game) newAsset(inst *Instance, controller PlayerID, faceDown bool) *Asset` in `pkg/engine/zones.go`.
  - Replaced inline asset constructions in `scratch` (`pkg/engine/zones.go`), `resolveDeploy` (`pkg/engine/stack.go`), and `NewGame` HQ setup (`pkg/engine/setup.go`) with calls to `newAsset`.
  - In `setup.go`, `inst` is allocated without ID, `newAsset` allocates asset ID, then `inst.ID = g.newID()` preserves original ID allocation sequence.
  - Acceptance cases:
    1. Tests unchanged: `git diff main -- '*_test.go'` printed nothing; `go test ./...` passed.
    2. Sims byte-identical: `go tool task sim-compare` passed with `[PASS] starter and stress sims match main byte for byte, exit codes included`.
    3. One asset construction: `git grep -n "&Asset{" -- pkg/engine ":!*_test.go"` returned exactly one line in `pkg/engine/zones.go` inside `newAsset`.
    4. Three callers: `git grep -n "g.newAsset(" -- pkg/engine` returned exactly three lines (setup.go, stack.go, zones.go).
    5. Stat check: `git diff main --stat` lists only setup.go, stack.go, zones.go, and this task file.
    6. Gates: `go tool task check` and `go tool task sim` passed cleanly.
- 2026-10-26 lead-architect (review): **Approved.** Verified against `git diff main...m2/T-0065-new-asset`, not the Log. Scope: only `setup.go`, `stack.go`, `zones.go` and this file changed, and no test changed (`git diff main -- '*_test.go'` is empty). The four edits match the spec exactly, including the Headquarters ID order in `setup.go`. Fences: no card names, nothing new outside `engine`. Reuse: `&Asset{` now appears once in non-test code (`zones.go:118`). Gates re-run in a review worktree: `go tool task sim-compare` [PASS] (byte-identical, exit codes included), `go tool task check` PASS, `go tool task sim` clean. Merged with `--no-ff`; directive #56 applied. Next: directive #57 (trigger fallback by source position), which builds on `newAsset`.
