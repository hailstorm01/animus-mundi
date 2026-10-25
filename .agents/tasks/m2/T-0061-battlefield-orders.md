---
id: T-0061
title: One place for the battlefield orders, in state.go (directive #54, harness, flash, pure move)
kind: harness
tier: flash
wave: none (board directive #54, structural review 3 finding 6, before T-0055)
status: ready
branch: m2/T-0061-battlefield-orders
depends_on: [T-0060]
claims: [helper:engine.assetsByID (move), doc:engine battlefield orders]
cards: []
requirements: []
---

## Goal
The three orders the battlefield is listed in sit together in `pkg/engine/state.go`, under one comment that names each order and the rule that uses it. Untargeted picks (waves B and C) then have one place to choose from. `fixedOrderPositions` reuses `TheaterOf` instead of repeating it. The task also carries two T-0056 nits: rewrap the `chooseTargets` comment, and use `slices` instead of `sort` in `choice.go`. Nothing else changes: same tests, same sim output.

## Inputs
- Board directive #54 (structural review 3, finding 6).
- Rulings named in the comment: `SVY-R7` (`FR-TRG-003`, trigger source order), `SVY-R11` (the design note lists target options oldest first under it), `SVY-R16` (`FR-MOD-001` §6, timeout fixed order).
- Pattern to follow: T-0058 and T-0060 (pure moves).

### The three orders (at `main` c0fc883)
| Function | Where | Order | Used by |
| :--- | :--- | :--- | :--- |
| `Assets` | `pkg/engine/state.go:199-210` | Headquarters, then every Reserves, then the Main Theater, each in arrival order | the death sweep and trigger collection, plus order-free uses |
| `assetsByID` | `pkg/engine/targeting.go:96-101` | lowest ID (oldest) first | decision option lists (`chooseTargets`), "all" steps (`stepAssets`) |
| `fixedOrderPositions` | `pkg/engine/state.go:213-243` | the choosing player's assets, then the opponent's; per player Headquarters, Reserves, then Main Theater by arrival | timeout fallbacks (`fixedOrderPicks`) |

## Edits (lead decisions; copy exactly)

### 1. `pkg/engine/state.go`: the comment block
Insert this block, followed by one blank line, directly above `// Assets returns every asset on the battlefield, HQs first.`:
```go
// Battlefield orders. The battlefield is listed in exactly three orders, and
// each rule that walks it uses one of them. New code picks one of these three
// rather than adding a fourth.
//
// Assets: the Headquarters, then every player's Reserves, then the Main
// Theater, each zone in arrival order. The death sweep and trigger collection
// use it, so simultaneous triggers are collected in source order (SVY-R7,
// FR-TRG-003). Uses where order doesn't matter (lookups, readying, invariants,
// legal actions) also use it.
//
// assetsByID: oldest (lowest ID) first. Decision option lists and "all" steps
// use it (SVY-R11).
//
// fixedOrderPositions: the choosing player's assets before the opponent's;
// within a player the Headquarters, then Reserves, then the Main Theater, each
// in arrival order. Timeout fallbacks use it (SVY-R16, FR-MOD-001 §6).
```

### 2. Move `assetsByID` from `targeting.go` to `state.go`
Cut the function and its one-line doc comment (`targeting.go:96-101`) and paste them unchanged in `state.go`, directly after the closing brace of `Assets` and before `// fixedOrderPositions maps …`, with one blank line on each side. Then fix the imports:
- `state.go` adds `"slices"` to its import block (which becomes `"fmt"`, `"math/rand/v2"`, `"slices"`, in that order).
- `targeting.go` drops `"slices"`, because `assetsByID` is its only use.

### 3. `fixedOrderPositions` uses `TheaterOf`
In `fixedOrderPositions`, the loop
```go
		for _, a := range g.Theater {
			if a.Controller == id {
				add(a)
			}
		}
```
becomes
```go
		for _, a := range g.TheaterOf(id) {
			add(a)
		}
```
`TheaterOf` filters `g.Theater` by controller in the same order, so the positions are the same.

### 4. Rewrap the `chooseTargets` comment (`targeting.go:12-18`)
Same words, new line breaks. Replace the whole comment with:
```go
// chooseTargets asks p's controller to choose each of its Effect.Targets, in
// slot order, when the trigger is placed on the stack (FR-TGT-001). Only legal
// entities are offered. A single-entity slot is not asked about with one
// option, and a timeout picks the first legal one in fixed order (SVY-R16).
// An "up to N" slot is always asked (Min 0 != Max), its picks are distinct
// and a timeout picks none (SVY-R11). A slot with no legal entity stays empty
// and asks nothing: deploying never requires a target (FR-TGT-003).
```

### 5. `choice.go`: `slices` instead of `sort`
In `fixedOrderPicks`,
```go
	sort.SliceStable(idx, func(a, b int) bool { return rank(idx[a]) < rank(idx[b]) })
```
becomes
```go
	slices.SortStableFunc(idx, func(a, b int) int { return rank(a) - rank(b) })
```
Note the change of argument meaning: `sort.SliceStable`'s function gets positions in `idx`, so it reads `idx[a]`; `slices.SortStableFunc`'s gets the elements themselves, which are already option indexes. Both sorts are stable, so the result is the same. In the import block, `"sort"` becomes `"slices"` (`sort` has no other use in `choice.go`).

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. Run each command on its own, pointed at your worktree. Record each command and its result in the Log; the reviewer re-runs them.
1. **Existing tests unchanged:** `git -C <worktree> diff main -- '*_test.go'` prints nothing, and `go -C <worktree> test ./...` passes.
2. **Sims byte-identical:** `go -C <worktree> tool task sim-compare` prints `[PASS]`.
3. **Three orders, one file:** `git -C <worktree> grep -nE 'func \(g \*Game\) (Assets|assetsByID|fixedOrderPositions)\(' -- pkg/engine` prints exactly three lines, all in `pkg/engine/state.go`.
4. **TheaterOf reused:** `git -C <worktree> grep -n 'range g.Theater {' -- pkg/engine/state.go` prints only the line inside `TheaterOf` (`Assets` appends `g.Theater...` and doesn't loop over it).
5. **No `sort` left in choice.go:** `git -C <worktree> grep -n 'sort' -- pkg/engine/choice.go` prints nothing.
6. **Only these edits:** `git -C <worktree> diff main --stat` lists only `pkg/engine/state.go`, `pkg/engine/targeting.go`, `pkg/engine/choice.go` and this task file.
7. `go -C <worktree> tool task check` and `go -C <worktree> tool task sim` pass.

## Allowed Files
- `pkg/engine/state.go` (the comment block, `assetsByID` moved in, `fixedOrderPositions`, imports)
- `pkg/engine/targeting.go` (`assetsByID` moved out, the `chooseTargets` comment, imports)
- `pkg/engine/choice.go` (the sort in `fixedOrderPicks`, imports)

## Done When
- [ ] Cases 1–6 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare` passes (starter and stress sims byte-identical to the integration branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-25 lead-architect: Wrote the task from board directive #54, the last of the four directives the user put before T-0055 (#49, #52 and #53 are applied). Harness, tier flash, pure move, so it skips `tests` and is `ready` on branch `m2/T-0061-battlefield-orders` (cut from `main` at this commit). Checked the directive's line references against `main` c0fc883: `Assets` is at `state.go:199` (the directive said 200), `fixedOrderPositions` at 217, `TheaterOf` at 256 (directive: 255), `assetsByID` at `targeting.go:96`. All the same code. Every command under Acceptance Cases and Done When fits the headless allowlist (D2, flash commands). Waiting triggers checked: #6 and #7 (mode is `serial`), #8 (no `deck.go` behavior change), #15 (no turn hook), #20 (no wave C shape task), #50 (no task claims the exhaust hook) and #51 (no "if you do" task). None has fired.
  - Next: dispatch `kernel-engineer` on Gemini Flash through `agy` (D3).
