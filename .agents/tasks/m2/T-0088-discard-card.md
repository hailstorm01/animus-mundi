---
id: T-0088
title: The view asks the engine whether a chosen card is still in a pile, (g *Game) DiscardCard (directives #65 and #67, harness, flash, pure move)
kind: harness
tier: flash
wave: B.1 (board directives #65 and #67, structural review 4 findings 4 and 6)
status: review
branch: m2/T-0088-discard-card
depends_on: [T-0087]
claims: [func:DiscardCard]
cards: []
requirements: []
---

## Goal
"The card with this instance ID, if it is still in a discard pile" is one exported lookup, `(g *Game) DiscardCard(id int) *Instance`, in `zones.go` beside `takeCard` and `putCard`. The engine's `pileTarget` and the view's `stackView` both call it, so when Open Question 10 (a zone-change count, `FR-ZON-004`) changes what "still the same card" means, one function changes and the view can't show a card the engine has dropped. `newID`'s doc comment says that asset IDs and card instance IDs come from one counter, and names what relies on it (#67). This is a pure move: same tests, same sim output.

## Inputs
- Board directives #65 and #67 (structural review 4, findings 4 and 6).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6a (wave B.1).
- Rules the moved code implements, unchanged: `FR-TGT-003` (a pile target is re-verified at resolution).
- Why it is a pure move: instance IDs are unique (every one comes from `newID`), so at most one card in all the discard piles has a given ID. `pileTarget` returned the first card with that ID (or nil) after checking `matchesCard`; it now gets the same card from `DiscardCard` and checks `matchesCard` the same way. `stackView` appended the name of every card with that ID, which is at most one; it now appends the name of the one `DiscardCard` returns. Nothing reads a random number or asks a decision here. `git grep` over `pkg` and `cmd` finds no test that names `pileTarget` or `stackView`.

### Where it is today (at `autoloop/m2-wave-b.1` 5b1dc11)
| Where | Today |
| :--- | :--- |
| `targeting.go:157-173` | `pileTarget`: loops over every player's `Discard` for the ID, then checks `matchesCard` |
| `view.go:191-199` | `stackView`: for each chosen ID of a pile slot, loops over every player's `Discard` and appends the matching card's name |
| `zones.go:85-118` | `takeCard` and `putCard`, the pile operations |
| `state.go:198-201` | `newID`, with no doc comment |
| `view.go:132` | `PlayerView`'s listing of a pile: not a lookup, stays |

## Edits (lead decisions; copy exactly)

### 1. `pkg/engine/zones.go`
Right after `putCard` (before `leaveBattlefield`'s doc comment), add:
```go

// DiscardCard returns the card with instance ID id if it is in a discard pile,
// else nil. It is the one lookup of a chosen pile card by ID: the engine
// re-verifies a pile target with it (pileTarget, FR-TGT-003), and a client
// uses it to show a chosen card that is still there.
func (g *Game) DiscardCard(id int) *Instance {
	for _, pl := range g.Players {
		for _, inst := range pl.Discard {
			if inst.ID == id {
				return inst
			}
		}
	}
	return nil
}
```

### 2. `pkg/engine/targeting.go`
Replace `pileTarget`, doc comment included, with:
```go
// pileTarget returns the card with instance ID id if it is still in a discard
// pile (DiscardCard) and still meets the pile filter f (FR-TGT-003), else nil.
// A card goes only to its owner's pile, so this is also "still in the zone the
// filter names". A card in a pile has no Ward, so only matchesCard is checked.
func (g *Game) pileTarget(ctl PlayerID, src *Instance, f ability.Filter, id int) *Instance {
	if inst := g.DiscardCard(id); inst != nil && g.matchesCard(ctl, src, f, inst) {
		return inst
	}
	return nil
}
```

### 3. `pkg/engine/state.go`
Give `newID` a doc comment (#67). Replace
```go
func (g *Game) newID() int {
```
with
```go
// newID returns the next ID from the game's one counter. Asset IDs and card
// instance IDs both come from it, so they never collide, and three things rely
// on that: fixedOrderPositions keys asset IDs and instance IDs in one map, and
// Option.Ref and StackItem.Chosen each hold either kind.
func (g *Game) newID() int {
```
The body doesn't change.

### 4. `pkg/view/view.go`
In `stackView`, replace the loop over the chosen IDs of a pile slot:
```go
			for _, id := range ids {
				for _, pl := range g.Players {
					for _, inst := range pl.Discard {
						if inst.ID == id {
							s.ChosenCards[i] = append(s.ChosenCards[i], inst.Card.Name)
						}
					}
				}
			}
```
with
```go
			for _, id := range ids {
				if inst := g.DiscardCard(id); inst != nil {
					s.ChosenCards[i] = append(s.ChosenCards[i], inst.Card.Name)
				}
			}
```
No import changes. No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. Run each command on its own, pointed at your worktree (W). `git grep` ignores untracked files, so stage your edits with `git -C W add` before cases 3–5. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (directive #65 case (a)): `git -C W diff autoloop/m2-wave-b.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes, the view tests included.
2. **Sims byte-identical** (case (b)): `go -C W tool task sim-compare -- autoloop/m2-wave-b.1` prints `[PASS]`.
3. **One loop looks a pile card up by ID** (case (c)): `git -C W grep -n "range pl.Discard" -- pkg/engine pkg/view ":!*_test.go"` prints exactly one line, in `zones.go`'s `DiscardCard`.
4. **Both callers use it:** `git -C W grep -n "DiscardCard" -- pkg cmd` prints exactly five lines: in `zones.go`, `DiscardCard`'s doc comment's first line and its `func` line; in `targeting.go`, `pileTarget`'s doc comment line and its one call; in `view.go`, one call, in `stackView`.
5. **`newID` is documented** (directive #67 case (a)): `git -C W grep -n "come from it, so they never collide" -- pkg/engine/state.go` prints one line, right above `func (g *Game) newID`.
6. **Only these edits:** `git -C W diff autoloop/m2-wave-b.1 --stat` lists only `pkg/engine/zones.go`, `pkg/engine/targeting.go`, `pkg/engine/state.go`, `pkg/view/view.go` and this task file.
7. `go -C W tool task check` and `go -C W tool task sim` pass. `check` runs `cardtown lint --all`, whose fences check holds `pkg/view` to importing only `engine` and `eventlog` (case (d)).

## Allowed Files
- `pkg/engine/zones.go`
- `pkg/engine/targeting.go`
- `pkg/engine/state.go`
- `pkg/view/view.go`

## Done When
- [x] Cases 1–6 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-b.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-28, lead-architect (wave B.1 run, session 4): wrote the task from board directives #65 and #67, which #67 says go in one task. Harness, so it skips `tests` and goes straight to `ready`. Tier `flash` as #65 gives. The edits are spelled out so the engineer copies them. `DiscardCard` goes in `zones.go` beside the pile operations `takeCard` and `putCard`, and is exported because `pkg/view` calls it, like `Asset(id)`, its battlefield counterpart in `state.go`. Not changed, on purpose: `effects.go`'s tracked death-trigger check (`slices.Contains` on the tracked card's owner's pile) compares a pointer the item already holds, which is not a lookup by ID; `pileDecision`'s and `fixedOrderPositions`'s walks over the piles, and `PlayerView`'s pile listing, list piles rather than find a card. `main` has no commits the run branch lacks, so no sync. With #65 and #67 applied, no `open: before` directive is left in wave B.1; #61 stays `waiting`, because no task's `cards:` has a modal card. Next: `kernel-engineer` (Flash, headless).
- 2026-10-28, kernel-engineer (Flash): verified implementation commit 2c29320 (feat(engine): move chosen pile card lookup to (g *Game) DiscardCard and document newID (T-0088)) and moved task to review.
  - Edits verified in diff against d0f1b62:
    - `pkg/engine/zones.go`: added `DiscardCard(id int) *Instance` after `putCard` to perform single pile card lookup by ID across player discard piles.
    - `pkg/engine/targeting.go`: updated `pileTarget` to use `g.DiscardCard(id)` and re-verify filter match, removing duplicate player discard loops.
    - `pkg/engine/state.go`: documented `newID()` explaining that asset IDs and card instance IDs share one counter and never collide.
    - `pkg/view/view.go`: updated `stackView` to resolve chosen IDs with `g.DiscardCard(id)` instead of looping player discard piles.
  - Commands and results:
    - Case 1: `git diff autoloop/m2-wave-b.1 -- '*_test.go'` printed nothing; `go test ./...` passed across all packages.
    - Case 2: `go tool task sim-compare -- autoloop/m2-wave-b.1` printed `[PASS] starter and stress sims match autoloop/m2-wave-b.1 byte for byte, exit codes included`.
    - Case 3: `git grep -n "range pl.Discard" -- pkg/engine pkg/view ":!*_test.go"` printed exactly one line: `pkg/engine/zones.go:125`.
    - Case 4: `git grep -n "DiscardCard" -- pkg cmd` printed exactly five lines across `zones.go`, `targeting.go`, and `view.go`.
    - Case 5: `git grep -n "come from it, so they never collide" -- pkg/engine/state.go` printed 1 line right above `newID`.
    - Case 6: `git diff d0f1b62 --stat` confirmed only the allowed files and this task file were touched on the task branch.
    - Case 7 / Gates: `go tool task check` (fmt, vet, all tests, cardtown lint --all) passed cleanly; `go tool task sim` passed (200 games, 0 invariant violations).
  - Status -> review.
