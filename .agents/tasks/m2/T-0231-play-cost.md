---
id: T-0231
title: "The policy asks the engine what a playable card costs: Game.PlayCost (directive #147, wave F.1 harness, flash)"
kind: harness
tier: flash
wave: F.1
status: review
branch: m2/autoloop-wave-f.1/T-0231-play-cost
depends_on: [T-0230]
claims: [func:Game.PlayCost, func:sim.playCosts (body)]
cards: []
requirements: [FR-CST-001, FR-TRN-003]
---

## Goal
"This card is playable, and here is what it costs" is written once, in the engine: `Game.PlayCost` holds the hand filter (supported, and a deploy zone or a Maneuver) and the discard filter (supported, and a deploy zone, then `DiscardDeployCost`), and `sim.playCosts` loops over the hand and then the discard pile calling only `PlayCost` (#147). The same pattern #127 used for activated abilities (`ActivatedCost`). The sim output doesn't change: `PlayCost` holds both filters exactly as the policy writes them today, and doesn't check a hand card's additional costs.

## Inputs
- Board directive #147 (structural review 11, finding 4, approved by the user, 2026-11-13), `open: before the next Deliverables task`; design note Deliverable 6j (wave F.1), fourth.
- At 4dc0303 (the run branch after T-0230): `playCosts` (`pkg/sim/policy.go:145-167`) restates the hand filter (`:152`) and the discard filter (`:159`) before calling `EffectiveCost` and `DiscardDeployCost`. `DiscardDeployCost` (`pkg/engine/costs.go:119-133`) keeps its name and behavior: `discard_deploy_cost_test.go` tests it.
- `pileCard` (`pkg/engine/zones.go:177`) finds a card by id in a player's hand or discard pile (T-0228). `PlayCost` uses it for both piles.
- Reuse search (D2.3): no engine reader returns a hand card's cost with a playability flag; `EffectiveCost` returns the cost only, and `ActivatedCost` is for activated abilities.
- Size (D2.4): two non-test files. #148 (the move to `freedeploy.go`) would add `effects.go`, `freedeploy.go` and `doc.go`, five in all, so it gets its own task after this one.
- The lead dry-ran these exact edits and this exact test file in a throwaway worktree of the run branch at 4dc0303: the tests fail to build before Edit 1 (`PlayCost` undefined) and pass after; `gofmt -l pkg` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-f.1` `[PASS]`, `cardtown lint fences` exit 0, `go tool task sim` exit 0 (200 games, P0 63.5% / P1 35.0% / draws 1.5%). Every expected output below is from that dry run.

## Commits
1. `test(T-0231): PlayCost acceptance cases` with only `pkg/engine/play_cost_test.go` (below). Before committing, run `go -C W test ./pkg/engine -run TestPlayCost` and record that it fails to build (`s.g.PlayCost undefined`).
2. `feat(engine): Game.PlayCost, and playCosts calls only it (T-0231)` with Edits 1 and 2.

## Tests (lead decisions; copy exactly as the new file `pkg/engine/play_cost_test.go`, tabs for indentation)
```go
package engine

import "testing"

// Tests for T-0231 (directive #147, FR-CST-001, FR-TRN-003): PlayCost, the one
// reader a client plans a card play with, from its owner's hand or discard
// pile. It uses T-0214's ddUnit, whose permission costs killing an allied unit,
// and T-0140's anotherEffect for a supported Maneuver. Player 0 owns every card.

func TestPlayCost_HandCards(t *testing.T) {
	s := newScene(t)
	u := unit("Grunt", 1, 1, Keywords{})
	u.ReqCost, u.TacCost = 2, 1
	if req, tac, ok := s.g.PlayCost(s.give(0, u)); req != 2 || tac != 1 || !ok {
		t.Errorf("a supported unit in hand: got R%d T%d %v, want R2 T1 true (EffectiveCost)", req, tac, ok)
	}
	if req, tac, ok := s.g.PlayCost(s.give(0, maneuver("Ambush", 1, 2, anotherEffect()))); req != 1 || tac != 2 || !ok {
		t.Errorf("a supported Maneuver in hand: got R%d T%d %v, want R1 T2 true", req, tac, ok)
	}
	odd := unit("Odd", 1, 1, Keywords{})
	odd.Unsupported = []string{"an unexecutable line"}
	if _, _, ok := s.g.PlayCost(s.give(0, odd)); ok {
		t.Errorf("an unsupported card in hand: ok true, want false")
	}
	if _, _, ok := s.g.PlayCost(s.give(0, &Card{Name: "Dud", Type: TypeManeuver, TacCost: 2})); ok {
		t.Errorf("a Maneuver with no effect in hand: ok true, want false (not Supported)")
	}
	stuck := unit("Stuck", 1, 1, Keywords{})
	stuck.DeployZones = nil
	if _, _, ok := s.g.PlayCost(s.give(0, stuck)); ok {
		t.Errorf("a hand card with no deploy zone that is not a Maneuver: ok true, want false")
	}
}

func TestPlayCost_DiscardCards(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	if req, tac, ok := s.g.PlayCost(s.inDiscard(0, ddUnit("Phoenix", 2, 1))[0]); req != 2 || tac != 1 || !ok {
		t.Errorf("a permitted discard card with payable costs: got R%d T%d %v, want R2 T1 true (DiscardDeployCost)", req, tac, ok)
	}
	zoneless := ddUnit("Zoneless", 2, 1)
	zoneless.DeployZones = nil
	if _, _, ok := s.g.PlayCost(s.inDiscard(0, zoneless)[0]); ok {
		t.Errorf("a discard card with no deploy zone: ok true, want false")
	}
	if _, _, ok := s.g.PlayCost(s.inDiscard(0, unit("Plain", 1, 1, Keywords{}))[0]); ok {
		t.Errorf("a discard card with no permission: ok true, want false (FR-TRN-003)")
	}
}

func TestPlayCost_OtherZones(t *testing.T) {
	s := newScene(t)
	if _, _, ok := s.g.PlayCost(s.put(0, ZoneMainTheater, unit("Fielded", 1, 1, Keywords{})).Inst); ok {
		t.Errorf("a card on the battlefield: ok true, want false")
	}
	if _, _, ok := s.g.PlayCost(s.onTop(0, ddUnit("Buried", 2, 1))[0]); ok {
		t.Errorf("a card in a deck: ok true, want false")
	}
}
```

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at 4dc0303.

### 1. `pkg/engine/costs.go`: add `PlayCost` after `DiscardDeployCost`
After line 133 (the closing `}` of `DiscardDeployCost`), insert one blank line and then:
```go
// PlayCost returns the Requisition and Tactics that playing inst costs now, and
// whether a client should plan for it: a supported card in its owner's hand
// that has a deploy zone or is a Maneuver costs EffectiveCost, without checking
// its additional costs, and a supported card in its owner's discard pile that
// has a deploy zone costs DiscardDeployCost. Anywhere else ok is false. Like
// DiscardDeployCost, it checks neither the resources against the pool nor the
// zone. LegalActions says what can be played now.
func (g *Game) PlayCost(inst *Instance) (req, tac int, ok bool) {
	c, p := inst.Card, g.Players[inst.Owner]
	switch {
	case pileCard(p, inst.ID, ZoneHand) != nil:
		if !c.Supported() || (len(c.DeployZones) == 0 && c.Type != TypeManeuver) {
			return 0, 0, false
		}
		req, tac = g.EffectiveCost(inst)
		return req, tac, true
	case pileCard(p, inst.ID, ZoneDiscard) != nil:
		if !c.Supported() || len(c.DeployZones) == 0 {
			return 0, 0, false
		}
		return g.DiscardDeployCost(inst)
	}
	return 0, 0, false
}
```
The blank line before `// otherCostsPayable` stays, so the file reads `}`, blank, `// PlayCost …`, …, `}`, blank, `// otherCostsPayable …`.

### 2. `pkg/sim/policy.go`: `playCosts` calls only `PlayCost`
Lines 145-167 (the comment and the function, through `return out` and its closing `}`), which read
```go
// playCosts lists what p's playable cards cost now: its hand cards, and the
// cards it may deploy from its discard pile by their own permission, leaving out
// one whose other costs can't be paid now (Game.DiscardDeployCost, FR-TRN-003).
func playCosts(g *engine.Game, p *engine.Player) []neededCost {
	var out []neededCost
	for _, inst := range p.Hand {
		c := inst.Card
		if !c.Supported() || (len(c.DeployZones) == 0 && c.Type != engine.TypeManeuver) {
			continue
		}
		req, tac := g.EffectiveCost(inst)
		out = append(out, neededCost{resources: resources{req, tac}})
	}
	for _, inst := range p.Discard {
		if !inst.Card.Supported() || len(inst.Card.DeployZones) == 0 {
			continue
		}
		if req, tac, ok := g.DiscardDeployCost(inst); ok {
			out = append(out, neededCost{resources: resources{req, tac}})
		}
	}
	return out
}
```
become
```go
// playCosts lists what p's playable cards cost now, its hand cards and then its
// discard pile's, as the engine prices them (Game.PlayCost).
func playCosts(g *engine.Game, p *engine.Player) []neededCost {
	var out []neededCost
	for _, pile := range [][]*engine.Instance{p.Hand, p.Discard} {
		for _, inst := range pile {
			if req, tac, ok := g.PlayCost(inst); ok {
				out = append(out, neededCost{resources: resources{req, tac}})
			}
		}
	}
	return out
}
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The tests failed first:** at the `test(T-0231)` commit, before Edit 1, `go -C W test ./pkg/engine -run TestPlayCost` fails with `s.g.PlayCost undefined`.
2. **Hand cards** (#147): `TestPlayCost_HandCards` passes: a supported unit in hand gives its `EffectiveCost` and `ok`; a supported Maneuver gives `ok`; an unsupported card, a Maneuver with no effect, and a non-Maneuver with no deploy zone give `!ok`.
3. **Discard cards** (#147): `TestPlayCost_DiscardCards` passes: a permitted discard card with payable costs gives `DiscardDeployCost`'s result; one with no deploy zone, and one with no permission, give `!ok`.
4. **Other zones** (#147): `TestPlayCost_OtherZones` passes: a card on the battlefield and a card in a deck give `!ok`.
5. **The policy asks only the engine** (#147): each of `git -C W grep -n -F 'Supported' -- pkg/sim/policy.go`, `git -C W grep -n -F 'EffectiveCost' -- pkg/sim/policy.go` and `git -C W grep -n -F 'DiscardDeployCost' -- pkg/sim/policy.go` prints nothing (three commands), and `git -C W grep -n -F 'g.PlayCost(inst)' -- pkg/sim/policy.go` prints exactly one line, `pkg/sim/policy.go:151:`.
6. **Sims byte-identical** (#147): `go -C W tool task sim-compare -- autoloop/m2-wave-f.1` prints `[PASS] starter and stress sims match autoloop/m2-wave-f.1 byte for byte, exit codes included`.
7. **Only the listed edits:** `git -C W diff autoloop/m2-wave-f.1 --numstat -- pkg cmd` prints exactly three lines: `25	0	pkg/engine/costs.go`, `59	0	pkg/engine/play_cost_test.go` and `7	17	pkg/sim/policy.go`.
8. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
9. **Nothing else changed:** `git -C W diff autoloop/m2-wave-f.1 --stat` lists only the three files of case 7 and this task file.
10. `gofmt -l W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/costs.go`
- `pkg/sim/policy.go`
- `pkg/engine/play_cost_test.go` (new; the `test(T-0231)` commit only)

## Done When
- [x] Cases 1–10 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-f.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-13, lead-architect (wave F.1 run, session 5): written for #147 against 4dc0303 (T-0230 merged; nothing new on `main` to sync). Harness, so it skips `tests` and goes straight to `ready`. Tier `flash`, because every edit and the whole test file are exact and were dry-run. #148 doesn't share this task: together they would touch five non-test files (D2.4); it follows as its own task. Every command here is on the headless allow-list: one per line, no pipes, no `$`, no unquoted parentheses. Waiting directives checked: #6, #7 (mode `serial`), #15 (no turn-boundary hook), #71, #91 (2), #106, #120, #141 (3), #142: not fired. Next: `kernel-engineer` on Flash.
- 2026-11-13, kernel-engineer (wave F.1 run, session 5): implemented directive #147. Added acceptance tests in `pkg/engine/play_cost_test.go` and verified they failed before implementation. Implemented `Game.PlayCost` in `pkg/engine/costs.go` (reusing `pileCard` from `zones.go` and `EffectiveCost`/`DiscardDeployCost` from `costs.go`) and updated `playCosts` in `pkg/sim/policy.go` to call only `g.PlayCost(inst)`. Verified acceptance cases 1–10:
  - Case 1: `go -C W test ./pkg/engine -run TestPlayCost` failed at commit `test(T-0231)` with `s.g.PlayCost undefined` (exit 1).
  - Case 2: `go -C W test ./pkg/engine -run TestPlayCost_HandCards` passed (exit 0).
  - Case 3: `go -C W test ./pkg/engine -run TestPlayCost_DiscardCards` passed (exit 0).
  - Case 4: `go -C W test ./pkg/engine -run TestPlayCost_OtherZones` passed (exit 0).
  - Case 5: `git -C W grep -n -F "Supported" -- pkg/sim/policy.go`, `git -C W grep -n -F "EffectiveCost" -- pkg/sim/policy.go`, and `git -C W grep -n -F "DiscardDeployCost" -- pkg/sim/policy.go` returned no matches (exit 1); `git -C W grep -n -F "g.PlayCost(inst)" -- pkg/sim/policy.go` returned `pkg/sim/policy.go:151:` (exit 0).
  - Case 6: `go -C W tool task sim-compare -- autoloop/m2-wave-f.1` passed: `[PASS] starter and stress sims match autoloop/m2-wave-f.1 byte for byte, exit codes included` (exit 0).
  - Case 7: `git -C W diff autoloop/m2-wave-f.1 --numstat -- pkg cmd` printed exactly 3 lines: `25 0 pkg/engine/costs.go`, `59 0 pkg/engine/play_cost_test.go`, and `7 17 pkg/sim/policy.go`.
  - Case 8: `go -C W run ./cmd/cardtown lint fences` passed: `[PASS] Package code is free of card names, invalid markers and disallowed imports.` (exit 0).
  - Case 9: `git -C W diff autoloop/m2-wave-f.1 --stat` lists only the 3 code/test files and this task file.
  - Case 10: `gofmt -l W/pkg` clean; `go -C W tool task check` passed (exit 0); `go -C W tool task sim` passed (exit 0, 200 games, P0 63.5% / P1 35.0% / draws 1.5%). Ready for review.
