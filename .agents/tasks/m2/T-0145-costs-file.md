---
id: T-0145
title: Additional costs and a card's cost get their own file, costs.go (directive #94, harness, flash, pure move)
kind: harness
tier: flash
wave: C.1 (board directive #94, structural review 7, finding 3)
status: in-progress
branch: m2/autoloop-wave-c.1/T-0145-costs-file
depends_on: [T-0144]
claims: [file:pkg/engine/costs.go]
cards: []
requirements: []
---

## Goal
A new `pkg/engine/costs.go` holds `EffectiveCost`, `payCost`, `costDecision`, `chooseCostDiscards`, `costsPayable` and `payCosts`, moved unchanged from `actions.go` and `placement.go`, and `doc.go`'s Files list names it (#94). A pure move: same declarations, same test names, same sim output.

## Inputs
- Board directive #94 (structural review 7, approved by the user, 2026-11-03).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6f (wave C.1).
- Split from T-0144 under the D2 size rule: #92, #93 and #94 together touch five non-test files. This task touches four: `actions.go`, `placement.go`, `costs.go` and `doc.go`.
- Rules the moved code implements, unchanged: `FR-CST-001`, `FR-TRN-003`, `FR-TRN-003a`, `SVY-R2`, `SVY-R4`, `SVY-R15`, `SVY-R29`.
- Not in this task: `declare`, `maneuverPlacement` and the `placement` struct stay in `placement.go`: they place an effect, and a placement's `costs` and `discards` fields are what the moved functions read. The imports of `actions.go` and `placement.go` don't change, because both files still use every package they import.
- Why it is a pure move: the moved text is byte-identical, and a Go package's files share one scope, so no reference changes. The lead dry-ran these exact edits in a throwaway worktree of the run branch at 233023f: `gofmt -l` clean, `go vet ./pkg/engine/` clean, `go tool task check` passes, `go tool task sim` runs clean, and `go tool task sim-compare -- autoloop/m2-wave-c.1` printed `[PASS]`.

### Where it is today (at `autoloop/m2-wave-c.1`, engine code as at 233023f)
| Where | Today |
| :--- | :--- |
| `pkg/engine/actions.go:214-226` | `EffectiveCost` and `payCost`, with their doc comments; line 227 is the blank line after them |
| `pkg/engine/placement.go:56-141` | `costDecision`, `chooseCostDiscards`, `costsPayable` and `payCosts`, with their doc comments; line 142 is the blank line after them |
| `pkg/engine/doc.go:59-61` | the Files entry for `trigger.go`, `placement.go`, `targeting.go` and `effects.go` |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/engine/costs.go` (new)
Create it with exactly this text. It is a new first comment and import block, then `actions.go:214-227`, then `placement.go:56-141`, every moved line unchanged. It ends with one newline after the closing `}`.
```go
// Costs: reading and paying a card's cost, and checking and paying a placement's additional costs (FR-CST-001, FR-TRN-003a).

package engine

import (
	"fmt"

	"animus-mundi/pkg/engine/ability"
)

// EffectiveCost returns what deploying or enacting inst costs now, in
// Requisition and Tactics, the one read of a card's cost (FR-CST-001). No card
// changes a cost yet, so it is the printed cost.
func (g *Game) EffectiveCost(inst *Instance) (req, tac int) {
	return inst.Card.ReqCost, inst.Card.TacCost
}

// payCost takes inst's cost (EffectiveCost) from p.
func (g *Game) payCost(p *Player, inst *Instance) {
	req, tac := g.EffectiveCost(inst)
	p.Req -= req
	p.Tac -= tac
}

// costDecision builds the pick of the allied asset that pays cost c, with the
// assets it lists in option order. It is an untargeted selection (DecidePick),
// so Ward doesn't apply, a timeout takes the first option in fixed order
// (FR-MOD-001 §6) and a single option is taken unasked (§1). Legality
// (costsPayable) and payment (payCosts) both build their decision with it. It
// is never called for a CostDiscard, which takes no asset.
func (g *Game) costDecision(pl placement, c ability.Cost) (Decision, []*Asset) {
	verb := "kill"
	if c.Kind == ability.CostExhaust {
		verb = "exhaust"
	}
	prompt := fmt.Sprintf("%s: choose an asset to %s (additional cost)", pl.label, verb)
	return g.assetDecision(pl.controller, DecidePick, prompt, 0, func(a *Asset) bool {
		if c.Kind == ability.CostExhaust && a.Exhausted {
			return false
		}
		return g.matchesFilter(pl.controller, pl.source, c.Filter, a)
	})
}

// chooseCostDiscards has pl's controller pick the hand cards its discard cost
// discards, before the modes and targets are chosen, so the number discarded is
// known when the targets are (FR-TRN-003a, SVY-R4). It moves nothing: payCosts
// discards them after the push. Without a discard cost it asks nothing and
// returns nil. A placement has at most one discard cost.
func (g *Game) chooseCostDiscards(pl placement) []*Instance {
	var picked []*Instance
	found := false
	for _, c := range pl.costs {
		if c.Kind != ability.CostDiscard {
			continue
		}
		if found {
			panic("engine: a placement has two discard costs")
		}
		found = true
		picked = g.chooseDiscards(pl.controller, c.UpTo, true)
	}
	return picked
}

// costsPayable reports whether every additional cost of pl has at least one
// option now (FR-TRN-003a, SVY-R29). A discard cost is always payable, because
// discarding none pays it.
func (g *Game) costsPayable(pl placement) bool {
	for _, c := range pl.costs {
		if c.Kind == ability.CostDiscard {
			continue // discarding none pays it
		}
		if _, opts := g.costDecision(pl, c); len(opts) == 0 {
			return false
		}
	}
	return true
}

// payCosts pays each additional cost of pl in printed order: its controller
// picks the asset (costDecision, pickFrom) and it is killed through kill
// (CostKill) or exhausted through exhaust (CostExhaust), while a discard cost
// moves the cards picked earlier (CostDiscard). A replaced kill (not the discard
// cost) still counts as paid (SVY-R15), and the triggers either fires wait in
// the pending buffer to go above the declared card (FR-TRN-003a). It runs after
// the Maneuver is pushed, so the card is already on the stack when the cost's
// triggers are queued (FR-TRN-003), and after its targets were chosen, so they
// were judged against the state before the cost moved anything (SVY-R2). A
// discard cost's cards were picked before the modes and targets
// (chooseCostDiscards) and are moved here, through discardCards.
func (g *Game) payCosts(pl placement) {
	for _, c := range pl.costs {
		if c.Kind == ability.CostDiscard {
			g.discardCards(pl.controller, pl.discards)
			continue
		}
		d, opts := g.costDecision(pl, c)
		for _, a := range pickFrom(g, d, opts) {
			switch c.Kind {
			case ability.CostKill:
				g.kill([]*Asset{a})
			case ability.CostExhaust:
				g.exhaust(a)
			default:
				panic(fmt.Sprintf("engine: unknown cost kind %d", c.Kind))
			}
		}
	}
}
```

### 2. `pkg/engine/actions.go`
Delete lines 214-227: from `// EffectiveCost returns what deploying or enacting inst costs now, in` through the blank line after `payCost`'s closing `}`. `handPlay`'s closing `}`, one blank line and `func (g *Game) readyAsset(a Action) (*Asset, error) {` are then adjacent.

### 3. `pkg/engine/placement.go`
Delete lines 56-142: from `// costDecision builds the pick of the allied asset that pays cost c, with the` through the blank line after `payCosts`' closing `}`. `maneuverPlacement`'s closing `}`, one blank line and `// slotsLegal reports whether the exactly-one slots of ts can all be filled at` are then adjacent.

### 4. `pkg/engine/doc.go`
Replace lines 60-61:
```go
//     an effect, its modes and then its targets. targeting.go: target
//     choice and filters. effects.go: running a resolved ability's steps.
```
with:
```go
//     an effect, its modes and then its targets. costs.go: a card's cost and
//     additional costs. targeting.go: target choice and filters. effects.go:
//     running a resolved ability's steps.
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. W is your worktree. Run each command on its own, pointed at W, after `git -C W add pkg/engine/costs.go` (a `git diff` doesn't list an untracked file). Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#94): `git -C W diff autoloop/m2-wave-c.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#94): `go -C W tool task sim-compare -- autoloop/m2-wave-c.1` prints `[PASS]`.
3. **All six functions sit in `costs.go`** (#94): `git -C W grep -n -e 'func (g \*Game) EffectiveCost(' -e 'func (g \*Game) payCost(' -e 'func (g \*Game) costDecision(' -e 'func (g \*Game) chooseCostDiscards(' -e 'func (g \*Game) costsPayable(' -e 'func (g \*Game) payCosts(' -- pkg/engine` prints exactly six lines, all in `pkg/engine/costs.go`, at lines 14, 19, 31, 50, 69 and 92.
4. **`placement.go` and `actions.go` keep only calls:** `git -C W grep -n -e 'ability.Cost' -e 'ReqCost' -e 'TacCost' -- pkg/engine/actions.go pkg/engine/placement.go` prints exactly one line, the `costs` field of `placement` in `pkg/engine/placement.go` (`costs      []ability.Cost`).
5. **`doc.go` names it** (#94): `git -C W grep -n 'costs.go' -- pkg/engine/doc.go` prints exactly one line.
6. **Moved, not changed:** `git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg` prints exactly these four lines: `0	14	pkg/engine/actions.go`, `110	0	pkg/engine/costs.go`, `3	2	pkg/engine/doc.go` and `0	87	pkg/engine/placement.go`.
7. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
8. **Only these edits:** `git -C W diff autoloop/m2-wave-c.1 --stat` lists only the four files in case 6 and this task file.
9. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/costs.go` (new)
- `pkg/engine/actions.go`
- `pkg/engine/placement.go`
- `pkg/engine/doc.go`

## Done When
- [ ] Cases 1–9 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-c.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-03, lead-architect (wave C.1 run, session 1): queued with T-0144 (D2.4).
- 2026-11-03, lead-architect (wave C.1 run, session 2): wrote this task from directive #94 against the run branch at 233023f, after syncing `main` (merge 2b0e1e2, which changes nothing under `pkg/`). Directive #103 (remove the TDD toolchain) arrived with that sync; it is `open`, not blocking, and sits after #94 in board order, so it goes into a later task. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$` or unquoted parentheses. Searched for existing equivalents: `pkg/engine` has no `costs.go`, and nothing else reads a card's cost or pays an additional cost. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2): no such card). Next: `kernel-engineer` on Flash.
