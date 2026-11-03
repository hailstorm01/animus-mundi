---
id: T-0144
title: One declare sequence, (g *Game) declare, and one read of a card's cost, (g *Game) EffectiveCost (directives #92 and #93, harness, flash, pure move)
kind: harness
tier: flash
wave: C.1 (board directives #92 and #93, structural review 7, findings 1 and 2)
status: in-progress
branch: m2/autoloop-wave-c.1/T-0144-declare-and-effective-cost
depends_on: []
claims: [func:declare, func:EffectiveCost, func:payCost signature]
cards: []
requirements: []
---

## Goal
The declare sequence lives in one function, `(g *Game) declare(pl placement, e ability.Effect, item *StackItem)` in `placement.go`, so wave D's activated abilities and wave F's *Veil Render* call it instead of copying `ActEnact` (#92). A card's cost is read in one place, `(g *Game) EffectiveCost(inst *Instance) (req, tac int)` in `actions.go`, so wave E's *Boneyard Amalgam* changes one function (#93). Both are pure moves: same tests, same sim output.

## Inputs
- Board directives #92 and #93 (structural review 7, approved by the user, 2026-11-03). The design note says #92, #93 and #94 may share a task, but together they touch five non-test files (`actions.go`, `placement.go`, `costs.go`, `doc.go`, `pkg/sim/policy.go`). So #94 is T-0145, `queued`, under the D2 size rule. It moves `EffectiveCost` and `payCost` to `costs.go` afterwards.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6f (wave C.1).
- Rules the moved code implements, unchanged: `FR-TRN-003`, `FR-TRN-003a`, `SVY-R2`, `SVY-R4` (the declare order), `FR-CST-001` (the cost is a computed value; no card changes it yet).
- Not in this task: `ActDeploy` keeps its own push (#92: unchanged until *Veil Render*'s task; no Unit or Structure has a `Def.ExtraCosts`). `view.CardView` and the play screen keep showing the printed cost (#93).
- Why it is a pure move: `declare` holds `ActEnact`'s four steps in their old order, and `item.Effect` and `item.Chosen` are filled before the push, as the struct literal filled them. `EffectiveCost` returns the fields `handPlay`, `payCost` and `ProgressPolicy` read before. The lead dry-ran these exact edits in a throwaway worktree of the run branch at e64ca5f: `gofmt -l` and `go vet` clean, `go test ./pkg/...` passes, and `go tool task sim-compare -- autoloop/m2-wave-c.1` printed `[PASS]`.

### Where it is today (at `autoloop/m2-wave-c.1`, engine code as at e64ca5f)
| Where | Today |
| :--- | :--- |
| `pkg/engine/actions.go:78-91` | the `ActDeploy` and `ActEnact` cases of `Apply` |
| `pkg/engine/actions.go:206-213` | `handPlay`'s cost check |
| `pkg/engine/actions.go:216-220` | `payCost` |
| `pkg/engine/placement.go:35` | the doc comment of `maneuverPlacement` |
| `pkg/sim/policy.go:38-45` | `ProgressPolicy`'s shortfall loop |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/engine/actions.go`, `ActDeploy` (#93)
Line 80, `payCost(p, inst.Card)`, becomes:
```go
		g.payCost(p, inst)
```

### 2. `pkg/engine/actions.go`, `ActEnact` (#92, #93)
Replace lines 84-91 (from `inst := handCard(p, a.Card)` through `g.payCosts(pl)`) with:
```go
		inst := handCard(p, a.Card)
		g.payCost(p, inst)
		g.takeCard(inst)
		g.declare(maneuverPlacement(a.Player, inst), inst.Card.Ability.Effect,
			&StackItem{Kind: StackManeuver, Controller: a.Player, Inst: inst})
```

### 3. `pkg/engine/actions.go`, `handPlay` (#93)
Replace lines 210-212 (the `if p.Req < c.ReqCost …` block) with:
```go
	req, tac := g.EffectiveCost(inst)
	if p.Req < req || p.Tac < tac {
		return nil, illegal("cannot pay for %s (needs R%d T%d)", c.Name, req, tac)
	}
```

### 4. `pkg/engine/actions.go`, `payCost` (#93)
Replace lines 216-220 (the doc comment and `payCost`) with:
```go
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
```

### 5. `pkg/engine/placement.go` (#92)
Before line 35 (`// maneuverPlacement is the placement of the Maneuver inst …`), insert this and one blank line:
```go
// declare runs the declare sequence for item, an effect declared by a manual
// action, in rules order: the discard cost's cards are picked, e's modes and
// targets are chosen and filled into item, item goes on the stack, and the
// additional costs are paid (FR-TRN-003, FR-TRN-003a).
func (g *Game) declare(pl placement, e ability.Effect, item *StackItem) {
	pl.discards = g.chooseCostDiscards(pl)
	item.Effect, item.Chosen = g.chooseOnPlacement(pl, e)
	g.push(item)
	g.payCosts(pl)
}
```

### 6. `pkg/sim/policy.go` (#93)
Replace lines 43-44 (`needReq = …` and `needTac = …`) with:
```go
			req, tac := g.EffectiveCost(inst)
			needReq = needReq || req > p.Req
			needTac = needTac || tac > p.Tac
```

No other file changes. `doc.go` needs nothing: no file is added.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#92, #93): `git -C W diff autoloop/m2-wave-c.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#92, #93): `go -C W tool task sim-compare -- autoloop/m2-wave-c.1` prints `[PASS]`.
3. **`actions.go` no longer runs the sequence itself** (#92): `git -C W grep -n -e chooseCostDiscards -e chooseOnPlacement -e payCosts -- pkg/engine/actions.go` prints nothing (git grep exits 1 when nothing matches; that is the expected result).
4. **One read of a card's cost** (#93): `git -C W grep -n -e ReqCost -e TacCost -- pkg/engine/actions.go pkg/sim/policy.go` prints exactly one line, in `pkg/engine/actions.go`: the `return` in `EffectiveCost`.
5. **The functions exist where the board puts them:** `git -C W grep -n -e "func (g \*Game) declare(" -e "func (g \*Game) EffectiveCost(" -e "func (g \*Game) payCost(" -- pkg/engine` prints exactly three lines: one in `pkg/engine/placement.go` and two in `pkg/engine/actions.go`.
6. **Moved, not changed:** `git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg` prints exactly these three lines: `19	13	pkg/engine/actions.go`, `11	0	pkg/engine/placement.go` and `3	2	pkg/sim/policy.go`.
7. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
8. **Only these edits:** `git -C W diff autoloop/m2-wave-c.1 --stat` lists only the three files in case 6 and this task file.
9. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/actions.go`
- `pkg/engine/placement.go`
- `pkg/sim/policy.go`

## Done When
- [ ] Cases 1–9 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-c.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-03, lead-architect (wave C.1 run, session 1): created the run file and wrote this task from board directives #92 and #93, first in board order, against the run branch at e64ca5f. No `main` commits to sync. Not `WAVE-COMPLETE`: #92, #93 and #95–#100 are `open: before the next Deliverables task`. #92–#94 in one task would be five non-test files, so #94 is queued as T-0145 (D2.4), which moves what this task adds to `costs.go`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$` or unquoted parentheses. Searched for existing equivalents: `payCost` is kept and now reads `EffectiveCost`; nothing else computes a cost or runs the declare steps. The only other reads of the printed cost, `view.go:136` and `play_screen.go:72`, stay as #93 says. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2): no such card). Next: `kernel-engineer` on Flash.
