---
id: T-0170
title: One read of "can pay" and one payment, (*Player).canPay and (*Player).pay (directive #107, harness, flash, pure move)
kind: harness
tier: flash
wave: D.1 (board directive #107, structural review 8, finding 4)
status: ready
branch: m2/autoloop-wave-d.1/T-0170-can-pay-and-pay
depends_on: []
claims: [helper:Player.canPay, helper:Player.pay]
cards: []
requirements: []
---

## Goal
Whether a player holds some Requisition and Tactics, and taking them, each live in one place: `(*Player).canPay(req, tac int) bool` and `(*Player).pay(req, tac int)` in `pkg/engine/costs.go`. `handPlay`, `payCost`, `costsPayable`, `payCosts` and Convert all call them (#107). A pure move: same behavior, same test names, same sim output.

## Inputs
- Board directive #107 (structural review 8, approved by the user, 2026-11-05).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6g (wave D.1).
- Placement: both methods go in `costs.go`, beside `payCost`, for now. #114 moves them to `resources.go`.
- One site beyond #107's four, by D2.3: the search for existing equivalents found Convert's own "can pay" and payment, `p.Req < 2` and `p.Req -= 2` (`actions.go:160`, `actions.go:86`). Leaving them would keep a copy of the scheme this task removes, and folding them in is two lines in a file the task already touches, so this task removes them. `check`'s message for it stays, and so does `p.Tac++`, which is what Convert gives. #114 still extracts `convert` from `Apply`.
- Not in this task: Generate's `p.Req += yield` and `p.Tac += yield` gain resources, they don't pay. `invariants.go:18`'s `p.Req < 0 || p.Tac < 0` checks an invariant, not a cost. `ProgressPolicy`'s `req > p.Req` and `tac > p.Tac` (`pkg/sim/policy.go`) ask which resource falls short, one at a time, from another package, so `canPay` can't answer them.
- Why it is a pure move:
  - `canPay`'s `p.Req >= req && p.Tac >= tac` is the negation of `p.Req < req || p.Tac < tac`, on the same ints.
  - `pay` subtracts in the order the old lines did.
  - `p.Req < 2` is `!p.canPay(2, 0)` because `p.Tac >= 0` always holds (`invariants.go:18`), and `p.pay(2, 0)` leaves `p.Tac` as it was.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 1062867: `gofmt -l` clean, `go tool task check` passes, `go tool task sim` runs clean, and `go tool task sim-compare -- autoloop/m2-wave-d.1` printed `[PASS]`. Every expected output in the cases below is from that dry run.

### Where it is today (at `autoloop/m2-wave-d.1`, engine code as at 1062867)
| Where | Today |
| :--- | :--- |
| `pkg/engine/actions.go:86` | Convert's payment in `Apply` |
| `pkg/engine/actions.go:160` | Convert's check in `check` |
| `pkg/engine/actions.go:236` | `handPlay`'s check |
| `pkg/engine/costs.go:18-23` | `payCost` |
| `pkg/engine/costs.go:78-83` | the `CostResources` case of `costsPayable` |
| `pkg/engine/costs.go:117-121` | the `CostResources` case of `payCosts` |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/engine/costs.go`, the new methods and `payCost`
Replace lines 18-23 (the doc comment of `payCost` through its closing brace) with:
```go
// canPay reports whether p holds req Requisition and tac Tactics, the one read
// of "can pay".
func (p *Player) canPay(req, tac int) bool {
	return p.Req >= req && p.Tac >= tac
}

// pay takes req Requisition and tac Tactics from p, the one payment.
func (p *Player) pay(req, tac int) {
	p.Req -= req
	p.Tac -= tac
}

// payCost takes inst's cost (EffectiveCost) from p.
func (g *Game) payCost(p *Player, inst *Instance) {
	req, tac := g.EffectiveCost(inst)
	p.pay(req, tac)
}
```

### 2. `pkg/engine/costs.go`, `costsPayable`
In `case ability.CostResources:`, replace the two lines `p := g.Players[pl.controller]` and `if p.Req < c.Req || p.Tac < c.Tac {` with this line:
```go
			if !g.Players[pl.controller].canPay(c.Req, c.Tac) {
```
The `return false`, the closing brace and the `continue` stay.

### 3. `pkg/engine/costs.go`, `payCosts`
In `case ability.CostResources:`, replace the three lines `p := g.Players[pl.controller]`, `p.Req -= c.Req` and `p.Tac -= c.Tac` with this line:
```go
			g.Players[pl.controller].pay(c.Req, c.Tac)
```
The `continue` stays.

### 4. `pkg/engine/actions.go`, `Apply`, `case ActConvert:`
Replace line 86, `p.Req -= 2`, with this line:
```go
		p.pay(2, 0)
```
`p.Tac++` stays.

### 5. `pkg/engine/actions.go`, `check`, `case ActConvert:`
Replace line 160, `if p.Req < 2 {`, with this line:
```go
		if !p.canPay(2, 0) {
```

### 6. `pkg/engine/actions.go`, `handPlay`
Replace line 236, `if p.Req < req || p.Tac < tac {`, with this line:
```go
	if !p.canPay(req, tac) {
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move, so no `test(T-0170)` commit. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#107): `git -C W diff autoloop/m2-wave-d.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#107): `go -C W tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`.
3. **One declaration of each:** `git -C W grep -n -e 'func (p \*Player) canPay(' -e 'func (p \*Player) pay(' -- pkg/engine` prints exactly two lines, at `pkg/engine/costs.go:20` and `pkg/engine/costs.go:25`.
4. **Every site calls them:** `git -C W grep -n -e 'canPay(' -e '\.pay(' -- pkg/engine` prints exactly seven lines: `pkg/engine/actions.go:86`, `pkg/engine/actions.go:160`, `pkg/engine/actions.go:236`, `pkg/engine/costs.go:20` (the declaration), `pkg/engine/costs.go:33`, `pkg/engine/costs.go:90` and `pkg/engine/costs.go:128`.
5. **The written-out forms are gone:** `git -C W grep -n -e 'p.Req <' -e 'p.Tac <' -e 'p.Req -=' -e 'p.Tac -=' -- pkg/engine/costs.go pkg/engine/actions.go` prints exactly two lines, both inside `pay`: `pkg/engine/costs.go:26` and `pkg/engine/costs.go:27`.
6. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.1 --numstat -- pkg cmd` prints exactly these two lines: `3	3	pkg/engine/actions.go` and `15	7	pkg/engine/costs.go`.
7. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
8. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.1 --stat` lists only `pkg/engine/actions.go`, `pkg/engine/costs.go` and this task file.
9. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/actions.go`
- `pkg/engine/costs.go`

## Done When
- [ ] Cases 1–9 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-05, lead-architect (wave D.1 run, session 1): created the run file and wrote this task from board directive #107, first in board order, against the run branch at 1062867. No `main` commits to sync. Not `WAVE-COMPLETE`: #107–#119 and #121 are `open: before the next Deliverables task`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Searched for existing equivalents: `pkg/engine` has no `canPay` or `pay` and no `Player` methods; the only other "can pay" and payment for resources is Convert's, which this task folds in (D2.3, see Inputs). Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120: no such card). Next: `kernel-engineer` on Flash.
