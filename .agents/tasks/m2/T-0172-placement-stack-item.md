---
id: T-0172
title: The placement builds its stack item, placement.stackItem, and declare takes a kind (directive #109, harness, flash, pure move)
kind: harness
tier: flash
wave: D.1 (board directive #109, structural review 8, finding 6)
status: review
branch: m2/autoloop-wave-d.1/T-0172-placement-stack-item
depends_on: [T-0171]
claims: [helper:placement.stackItem, signature:Game.declare]
cards: []
requirements: []
---

## Goal
A placed stack item's controller, source and instance are written in one place: `(placement).stackItem(kind StackKind) *StackItem` in `pkg/engine/placement.go`. `declare` takes the kind instead of an item, and `flushTriggers` uses the method and sets `Tracked` and `Keyword` itself (#109). A pure move: same behavior, same test names, same sim output.

## Inputs
- Board directive #109 (structural review 8, approved by the user, 2026-11-05).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6g (wave D.1), and Open Question 10 (the zone-change count that would otherwise be added to all three copies).
- Not in this task: `replaceDeath`'s item at `pkg/engine/replace.go:112` stays as it is, because it never goes on the stack (#109). The Deploy, March, Attack and Scratch items aren't placements: nothing is chosen for them, so they stay too.
- Why it is a pure move:
  - **Maneuver:** `maneuverPlacement` sets `controller: pid` and `inst: inst` and no `source`, so `stackItem(StackManeuver)` builds the item the old literal did, with `Source` 0.
  - **Activated ability:** `activatedPlacement` sets `controller: pid`, `source: a.ID` and `inst: a.Inst`, which are the old literal's `a.Player`, `as.ID` and `as.Inst`.
  - **Trigger:** `pendingTrigger` embeds `placement`, so `p.stackItem` reads the same `p.controller`, `p.source` and `p.inst` the old literal read.
  - `stackItem` has no side effects, so building the item before `chooseOnPlacement` instead of after changes nothing. The choice still comes before the push in all three.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at f301c42: `gofmt -l` clean, `go tool task check` passes, `go tool task sim` runs clean, and `go tool task sim-compare -- autoloop/m2-wave-d.1` printed `[PASS]`. Every expected output in the cases below is from that dry run.

### Where it is today (at `autoloop/m2-wave-d.1`, engine code as at f301c42)
| Where | Today |
| :--- | :--- |
| `pkg/engine/actions.go:93-94` | Enact builds its `StackManeuver` item for `declare` |
| `pkg/engine/actions.go:98-99` | Activate builds its `StackAbility` item for `declare` |
| `pkg/engine/trigger.go:180-183` | `flushTriggers` builds its `StackAbility` item in full |
| `pkg/engine/placement.go:36-45` | `declare` takes an `item *StackItem` |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/engine/placement.go`, the new method
Between the closing brace of `type placement struct` and the comment on `chooseOnPlacement`, add this, with one blank line on each side:
```go
// stackItem builds the item pl goes on the stack as: its kind, and the
// controller, source and instance pl already holds.
func (pl placement) stackItem(kind StackKind) *StackItem {
	return &StackItem{Kind: kind, Controller: pl.controller, Source: pl.source, Inst: pl.inst}
}
```

### 2. `pkg/engine/placement.go`, `declare`
Replace lines 36-43, from `// declare runs the declare sequence` through `g.push(item)`, with:
```go
// declare runs the declare sequence for an effect declared by a manual action,
// in rules order: the discard cost's cards are picked, e's modes and targets are
// chosen and filled into a stack item of this kind, the item goes on the stack,
// and the additional costs are paid (FR-TRN-003, FR-TRN-003a).
func (g *Game) declare(pl placement, e ability.Effect, kind StackKind) {
	pl.discards = g.chooseCostDiscards(pl)
	item := pl.stackItem(kind)
	item.Effect, item.Chosen = g.chooseOnPlacement(pl, e)
	g.push(item)
```
The `g.payCosts(pl)` line and the closing brace stay.

### 3. `pkg/engine/actions.go`, `Apply`, `case ActEnact:`
Replace lines 93-94, the `g.declare` call and its `&StackItem` line, with this line:
```go
		g.declare(maneuverPlacement(a.Player, inst), inst.Card.Ability.Effect, StackManeuver)
```

### 4. `pkg/engine/actions.go`, `Apply`, `case ActActivate:`
Replace lines 98-99, the `g.declare` call and its `&StackItem` line, with this line:
```go
		g.declare(activatedPlacement(a.Player, as, act), act.Effect, StackAbility)
```

### 5. `pkg/engine/trigger.go`, `flushTriggers`
Replace lines 180-183, from `eff, chosen := g.chooseOnPlacement` through `Chosen: chosen})`, with:
```go
			item := p.stackItem(StackAbility)
			item.Effect, item.Chosen = g.chooseOnPlacement(p.placement, p.trig.Effect)
			item.Tracked, item.Keyword = p.trig.Hook == ability.HookKilled, p.trig.Keyword
			g.push(item)
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move, so no `test(T-0172)` commit. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#109): `git -C W diff autoloop/m2-wave-d.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#109): `go -C W tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`.
3. **One method, called twice:** `git -C W grep -n -e 'stackItem(' -- pkg/engine` prints exactly three lines: `pkg/engine/placement.go:27` (the declaration), `pkg/engine/placement.go:48` and `pkg/engine/trigger.go:180`.
4. **`declare` takes a kind:** `git -C W grep -n -e 'g.declare(' -e 'func (g [*]Game) declare(' -- pkg/engine` prints exactly three lines: `pkg/engine/actions.go:93`, ending `StackManeuver)`; `pkg/engine/actions.go:97`, ending `StackAbility)`; and `pkg/engine/placement.go:46`, ending `kind StackKind) {`.
5. **The copies are gone:** `git -C W grep -n -e 'StackItem{' -- pkg/engine/actions.go pkg/engine/placement.go pkg/engine/trigger.go pkg/engine/replace.go` prints exactly five lines: `pkg/engine/actions.go:90` (`StackDeploy`), `pkg/engine/actions.go:101` (`StackMarch`), `pkg/engine/actions.go:106` (`StackScratch`), `pkg/engine/placement.go:28` (the method) and `pkg/engine/replace.go:112` (`replaceDeath`, unchanged).
6. **The trigger sets its own fields:** `git -C W grep -n -e 'item.Tracked, item.Keyword = ' -- pkg/engine` prints exactly one line, `pkg/engine/trigger.go:182`.
7. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.1 --numstat -- pkg cmd` prints exactly these three lines: `2	4	pkg/engine/actions.go`, `12	5	pkg/engine/placement.go` and `4	4	pkg/engine/trigger.go`.
8. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
9. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.1 --stat` lists only `pkg/engine/actions.go`, `pkg/engine/placement.go`, `pkg/engine/trigger.go` and this task file.
10. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/actions.go`
- `pkg/engine/placement.go`
- `pkg/engine/trigger.go`

## Done When
- [x] Cases 1–10 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-d.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-05, lead-architect (wave D.1 run, session 3): wrote this task from board directive #109, next in board order, against the run branch at f301c42. No `main` commits to sync. Not `WAVE-COMPLETE`: #109–#119 and #121 are `open: before the next Deliverables task`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Searched for existing equivalents: `placement` has no methods, and nothing else builds a `StackItem` from one. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120: no such card). Next: `kernel-engineer` on Flash.
- 2026-11-05, kernel-engineer (T-0172): added (pl placement).stackItem(kind StackKind) *StackItem method and updated declare signature to take kind StackKind in pkg/engine/placement.go; updated ActEnact and ActActivate calls to declare in pkg/engine/actions.go; updated flushTriggers in pkg/engine/trigger.go to use p.stackItem(StackAbility) and set Tracked and Keyword directly. Reused placement fields (controller, source, inst) and existing chooseOnPlacement / push. Acceptance cases 1–10 executed and verified:
  - Case 1: git -C W diff autoloop/m2-wave-d.1 -- '*_test.go' printed nothing; go -C W test ./... passed (exit 0).
  - Case 2: go -C W tool task sim-compare -- autoloop/m2-wave-d.1 passed ([PASS] starter and stress sims match autoloop/m2-wave-d.1 byte for byte, exit codes included).
  - Case 3: git -C W grep -n -e 'stackItem(' -- pkg/engine returned pkg/engine/placement.go:27, pkg/engine/placement.go:48, pkg/engine/trigger.go:180.
  - Case 4: git -C W grep -n -e 'g.declare(' -e 'func (g [*]Game) declare(' -- pkg/engine returned pkg/engine/actions.go:93, pkg/engine/actions.go:97, pkg/engine/placement.go:46.
  - Case 5: git -C W grep -n -e 'StackItem{' -- pkg/engine/actions.go pkg/engine/placement.go pkg/engine/trigger.go pkg/engine/replace.go returned pkg/engine/actions.go:90, pkg/engine/actions.go:101, pkg/engine/actions.go:106, pkg/engine/placement.go:28, pkg/engine/replace.go:112.
  - Case 6: git -C W grep -n -e 'item.Tracked, item.Keyword = ' -- pkg/engine returned pkg/engine/trigger.go:182.
  - Case 7: git -C W diff autoloop/m2-wave-d.1 --numstat -- pkg cmd returned 2 4 pkg/engine/actions.go, 12 5 pkg/engine/placement.go, 4 4 pkg/engine/trigger.go.
  - Case 8: go -C W run ./cmd/cardtown lint fences passed (exit 0).
  - Case 9: git -C W diff autoloop/m2-wave-d.1 --stat lists only pkg/engine/actions.go, pkg/engine/placement.go, pkg/engine/trigger.go, and this task file.
  - Case 10: go -C W tool task check and go -C W tool task sim passed.
  Reviewer note: placement.stackItem constructs the StackItem with pl.controller, pl.source, and pl.inst; replaceDeath in replace.go remains untouched as it never enters the stack. Sims remain byte-identical to autoloop/m2-wave-d.1.
