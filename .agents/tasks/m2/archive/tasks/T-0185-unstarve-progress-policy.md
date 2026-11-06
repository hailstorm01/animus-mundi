---
id: T-0185
title: Unstarve ProgressPolicy without scenario seeding (directive #122, harness, flash)
kind: harness
tier: flash
wave: D.1
status: done
branch: m2/T-0185-unstarve-progress-policy
depends_on: [T-0184]
claims: [func:sim.ProgressPolicy, const:sim.marchPast, const:engine.ConvertReq, fence:pkg/sim imports pkg/engine/ability, test:TestProgress_UnreachableCardTapsNothing, test:TestProgress_ReachRespectsTheZoneLimit, test:TestProgress_ReachCountsConversion, test:TestProgress_NoTapBreaksReach, test:TestProgress_ActivationBudgetsAndActivates, test:TestProgress_ActivationIsNotEager, test:TestProgress_SavingHoldsMarchesOutOfReserves, test:TestProgress_StarterGamesFireActivations]
cards: []
requirements: []
---

## Goal
`ProgressPolicy` stops tapping every asset for a hand card it can't pay for this turn, so units attack and march, and activated abilities and the deck search fire in real starter games. No scenario seeding: `pkg/engine` and `MatchStart` gain no setup or deck-stacking fields. The rules are general and name no card.

## Inputs
- Board directive #122 and its [outcome](../../../../../docs/design/outcomes/2026-11-05-oq18-sim-policy-without-scenario-seeding.md), including the outcome's Revision, which replaces its body where they differ. This task takes the "soften the wall" route of its item 5. (a) is not taken as written: `ActActivate` stays in `acts`, and `TestProgress_ActivationIsAnAct` passes unedited.
- `pkg/sim/policy.go:47-91`: the policy this replaces.
- `pkg/engine/actions.go:247-250`: `LegalActions` offers Generate only where `canYield` allows it, so the policy reads the zone limit (FR-CAP-001) from `legal` and doesn't copy it.
- Reuse search (D2.3): nothing in `pkg/sim`, `pkg/view` or `pkg/engine` works out what a player can reach this turn. `canYield` and `canPay` are unexported, and `legal` and the plain comparison stand in for them. `convertReq` is the one copy of the conversion rate, so it is exported and not written a second time.

## Lead decisions
1. **What a closer is (#122 (b), Revision item 2).** A Generate or Convert keeps its priority over acts only if it adds a resource a hand card is short of and leaves that card's cost reachable this turn. Reach counts the current pool, every asset's legal yields and conversion, with each asset tapping for one resource, and uses reach ≥ cost. `TestProgress_Case2_Preferences` and `TestProgress_Case6_EnactsManeuvers` sit exactly at reach = cost and pass unedited. Requisition counts toward a Tactics shortfall only when no asset can tap for Tactics, because Case6's Case 3 wants the HQ's Tactics, not the Worker's Requisition.
2. **Budgeting an activation (#122 (a)'s budget, Revision item 5).** Revision item 5 makes (a)'s *eagerness* optional, and the acceptance line still asks that an asset "budgets and activates". So a Generate or Convert toward an activated ability's resource cost is added to `acts`, on equal terms with an attack or a march, and never to `closers`. An asset doesn't tap to fund its own ability. A *Sinuous Growth* therefore isn't made to kill itself any more eagerly than a march. `activationCosts` doesn't check targets, because the engine's placement check is unexported: a tap may fund a *Boneyard Scavenger* with no Maneuver in the discard pile. That costs one tap, and the Activate is then simply not legal.
3. **A rule #122 doesn't list: `marchPast` (measured).** With decisions 1 and 2 alone, units march out of Reserves, where they yield only Tactics, and expensive cards lose their Requisition. Over `-seed 1`: *Allshan* 33 → 3 deploys, *Arkha* 24 → 3, *Hammerhand Brigade* 64 → 13, *Boneyard Scavenger* 66 → 16, and 4 activations of *Boneyard Scavenger*, under the reopen trigger's line. The old wall was protecting these cards by accident. So while a hand card is short of Requisition it can't reach, a Reserves unit with a Requisition yield marches only `marchPast` of the time. It is not a hard rule, because a hard rule pins every Reserves unit for as long as a card the player may never afford stays in hand (*Boneyard Amalgam*, R10, once it is executable), which is the starvation #122 removes. Tried at 0 (hard), 0.25 and 0.5, on seeds 1 and 1001. 0.25 keeps every shape over the trigger and the expensive cards near their baseline.
4. **The fence widens by one entry.** `activationCosts` reads `ability.CostResources` and `ability.CostExhaustSelf`, and `allowedImports` lets `pkg/sim` import only `pkg/engine`. `pkg/sim` → `pkg/engine/ability` runs in the allowed direction, and `Card.Ability` already exposes that type. The alternative, an engine accessor for activation costs, would add to package `engine` something the rules don't need. The lead updates the design note's Boundaries at close.
5. **`sim-compare` is expected to differ** (Revision item 3), so it is not a gate here.

## Measured on the lead's dry run
The lead dry-ran these exact edits in a throwaway worktree of `main` at 0d3d5a8. `cardtown sim -deck starter -p0 ANIMUS -p1 PATH`, 200 games, counted over the `-events` file:

| | Baseline, `-seed 1` | This task, `-seed 1` | This task, `-seed 1001` |
| :--- | :---: | :---: | :---: |
| P0 / P1 / draws | 84.0 / 15.0 / 1.0% | 78.5 / 19.0 / 2.5% | 81.0 / 18.5 / 0.5% |
| Average turns | 30.6 | 24.7 | 24.5 |
| Attacks | 2173 | 2778 | 2749 |
| *Cliff of Portents*: deployed / tapped / activated | 80 / 338 / 1 | 71 / 273 / 20 | 79 / 341 / 22 |
| *Sinuous Growth*: activated | 96 | 232 | 242 |
| *Boneyard Scavenger*: deployed / activated | 66 / 2 | 83 / 14 | 87 / 18 |
| Deck searches | 1 | 20 | 22 |
| *Allshan* / *Arkha* / *Hammerhand Brigade* deploys | 33 / 24 / 64 | 49 / 19 / 59 | 32 / 27 / 76 |

In 1000 games at `-seed 1`, `TestProgress_StarterGamesFireActivations` logs *Cliff* 120, *Sinuous Growth* 1266, *Boneyard Scavenger* 121 and 120 deck searches. `gofmt -l` is clean, and `go tool task check` and `go tool task sim` exit 0. The stress sim runs clean.

**Each new test fails on the current policy** (all 8, including the 1000-game floors). **Mutations, each reverted, and the tests that caught them:** reach off by one → `UnreachableCardTapsNothing`. No conversion route → `ReachCountsConversion`. A tap not removed from the rest → `UnreachableCardTapsNothing`, `NoTapBreaksReach`. Activation funding as a closer → `ActivationIsNotEager`. Activate as a play → `ActivationIsAnAct`, `ActivationIsNotEager`. No hold → `SavingHoldsMarchesOutOfReserves`. `marchPast` 0 → the same. A source funding itself → `ActivationBudgetsAndActivates`. Printed yields ignoring the zone → `ReachRespectsTheZoneLimit` and three others. No activation funding → `ActivationBudgetsAndActivates`, `ActivationIsNotEager`. A hold on a unit with no Requisition yield → `SavingHoldsMarchesOutOfReserves`. A hold on any shortfall → the same.

## Acceptance Cases
Each is one test in Edit 1, by the name given. All fail before the implementation commit.
1. A 3R card against HQ R1 + Worker R1, one short of reach → no Generate or Convert on any of 32 seeds, and the units act on at least one (`TestProgress_UnreachableCardTapsNothing`).
2. 1R 2T against two Reserves units printing T2 → no Generate: they can't yield Tactics there (`TestProgress_ReachRespectsTheZoneLimit`).
3. A 1T Maneuver and two Reserves units, HQ exhausted → Generate, Generate, Convert, Enact on all 8 seeds (`TestProgress_ReachCountsConversion`).
4. 1R 1T against HQ (R1 or T2) and a Reserves Worker → Generate on all 32 seeds, and the HQ never taps for Requisition, which would put Tactics out of reach (`TestProgress_NoTapBreaksReach`).
5. A readied asset with "R1, exhaust: draw 1" and a readied HQ, pool 0 → it activates on at least one of 32 seeds, never taps itself and nothing taps for Tactics (`TestProgress_ActivationBudgetsAndActivates`).
6. An affordable activation beside a March, and a reachable one beside a March → both options come up in 64 seeds each, so activation is not a play or a closer (`TestProgress_ActivationIsNotEager`).
7. A 9R card and a Reserves unit, 64 seeds → some marches and at most half. With an empty hand, a unit with no Requisition yield, or a card short only of Tactics → more than half (`TestProgress_SavingHoldsMarchesOutOfReserves`).
8. 1000 starter games → *Cliff of Portents*, *Sinuous Growth* and *Boneyard Scavenger* each activate at least 25 times, and at least 25 deck searches (`TestProgress_StarterGamesFireActivations`).

Existing tests pass unedited, in particular `TestProgress_Case2_Preferences`, `TestProgress_Case6_EnactsManeuvers`, `TestProgress_Case4_Coverage` and `TestProgress_ActivationIsAnAct`.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/sim/policy_test.go`, in the `test(T-0185)` commit
Add this at the end of the file, after one blank line. It needs no new imports.

```go
// Tests for T-0185 (directive #122): each is the acceptance case it names.

// costlyUnit is a vanilla unit costing req Requisition.
func costlyUnit(req int) *engine.Card {
	c := unit("Costly", 1, 1, engine.Keywords{})
	c.ReqCost = req
	return c
}

// emptyPool clears player 0's hand and pool, and exhausts its Headquarters unless
// keepHQ.
func emptyPool(s *scene, keepHQ bool) {
	s.g.Players[0].Hand = nil
	s.g.Players[0].Req, s.g.Players[0].Tac = 0, 0
	if !keepHQ {
		s.g.Players[0].HQ.Exhausted = true
	}
}

// activator is a unit with the activated ability "req R, tac T, exhaust: draw 1".
func activator(req, tac int) *engine.Card {
	c := unit("Engine", 1, 3, engine.Keywords{})
	c.Ability = &ability.Def{Activated: []ability.Activated{{
		Costs: []ability.Cost{
			{Kind: ability.CostResources, Req: req, Tac: tac},
			{Kind: ability.CostExhaustSelf},
		},
		Effect: ability.Effect{Steps: []ability.Step{{Op: ability.OpDraw, Amount: ability.Value{N: 1}}}},
	}}}
	return c
}

// Case 1: a card the board can't pay for this turn taps nothing, and the units
// act instead. 3R against HQ R1 + Worker R1 is one short of reach
// (TestProgress_Case2_Preferences' second scene is the same board at reach = cost).
func TestProgress_UnreachableCardTapsNothing(t *testing.T) {
	acted := 0
	for seed := uint64(1); seed <= 32; seed++ {
		s := newScene(t)
		emptyPool(s, true)
		s.give(0, costlyUnit(3))
		s.put(0, engine.ZoneReserves, unit("Worker", 1, 1, engine.Keywords{}))
		s.put(0, engine.ZoneMainTheater, unit("Fighter", 1, 1, engine.Keywords{}))
		// The Fighter yields Tactics only in the Main Theater, so it adds no reach.
		got := ProgressPolicy(seed)(s.g, s.g.LegalActions())
		switch got.Kind {
		case engine.ActGenerate, engine.ActConvert:
			t.Errorf("seed %d: 3R is out of reach, got %+v", seed, got)
		case engine.ActAttack, engine.ActMarch, engine.ActScratch:
			acted++
		}
	}
	if acted == 0 {
		t.Errorf("over 32 seeds the units never acted")
	}
}

// Case 2: reach respects the zone limit on unit yields (FR-CAP-001). A unit in
// Reserves prints T2 but can't yield Tactics there. Counted by their printed
// yields, two of them would pay 1R 2T; by their zone they can't, so neither taps
// for the 1R.
func TestProgress_ReachRespectsTheZoneLimit(t *testing.T) {
	for seed := uint64(1); seed <= 16; seed++ {
		s := newScene(t)
		emptyPool(s, false)
		s.give(0, drawManeuver("Costly", 1, 2))
		for range 2 {
			worker := unit("Worker", 1, 1, engine.Keywords{})
			worker.TacYield = 2
			s.put(0, engine.ZoneReserves, worker)
		}
		if got := ProgressPolicy(seed)(s.g, s.g.LegalActions()); got.Kind == engine.ActGenerate {
			t.Errorf("seed %d: Tactics is out of reach from Reserves, got %+v", seed, got)
		}
	}
}

// Case 3: Requisition converts toward a Tactics cost when no asset can tap for
// Tactics (FR-RES-003): two Reserves units fund a 1T Maneuver, tap by tap.
func TestProgress_ReachCountsConversion(t *testing.T) {
	for seed := uint64(1); seed <= 8; seed++ {
		s := newScene(t)
		emptyPool(s, false)
		s.give(0, drawManeuver("Study", 0, 1))
		s.put(0, engine.ZoneReserves, unit("Worker", 1, 1, engine.Keywords{}))
		s.put(0, engine.ZoneReserves, unit("Worker", 1, 1, engine.Keywords{}))
		policy := ProgressPolicy(seed)
		for _, want := range []engine.ActionKind{engine.ActGenerate, engine.ActGenerate, engine.ActConvert, engine.ActEnact} {
			got := policy(s.g, s.g.LegalActions())
			if got.Kind != want {
				t.Fatalf("seed %d: got %+v, want %s", seed, got, want)
			}
			if err := s.g.Apply(got); err != nil {
				t.Fatalf("seed %d: %v", seed, err)
			}
		}
	}
}

// Case 4: no tap that would put the cost out of reach. 1R 1T against HQ (R1 or
// T2) and a Reserves Worker (R1): the HQ must tap for Tactics, since tapping it
// for Requisition leaves nothing that yields Tactics.
func TestProgress_NoTapBreaksReach(t *testing.T) {
	for seed := uint64(1); seed <= 32; seed++ {
		s := newScene(t)
		emptyPool(s, true)
		c := unit("Costly", 1, 1, engine.Keywords{})
		c.ReqCost, c.TacCost = 1, 1
		s.give(0, c)
		s.put(0, engine.ZoneReserves, unit("Worker", 1, 1, engine.Keywords{}))
		got := ProgressPolicy(seed)(s.g, s.g.LegalActions())
		if got.Kind != engine.ActGenerate {
			t.Errorf("seed %d: 1R 1T is reachable, got %s, want Generate", seed, got.Kind)
		}
		if got.Asset == s.g.Players[0].HQ.ID && got.Resource == engine.Requisition {
			t.Errorf("seed %d: the HQ tapped for Requisition, which puts Tactics out of reach", seed)
		}
	}
}

// Case 5: a readied asset with an R/T activated ability budgets and activates:
// another asset taps toward its cost and it fires. It never taps itself for its
// own cost, and the HQ doesn't tap for a resource the ability doesn't cost.
func TestProgress_ActivationBudgetsAndActivates(t *testing.T) {
	activated := 0
	for seed := uint64(1); seed <= 32; seed++ {
		s := newScene(t)
		emptyPool(s, true)
		engineAsset := s.put(0, engine.ZoneReserves, activator(1, 0))
		policy := ProgressPolicy(seed)
		for range 8 {
			got := policy(s.g, s.g.LegalActions())
			if got.Kind == engine.ActEndTurn {
				break
			}
			if got.Kind == engine.ActGenerate && got.Asset == engineAsset.ID {
				t.Errorf("seed %d: the asset tapped itself for its own activation", seed)
			}
			if got.Kind == engine.ActGenerate && got.Resource == engine.Tactics {
				t.Errorf("seed %d: tapped for Tactics toward an R1 cost", seed)
			}
			if got.Kind == engine.ActActivate {
				activated++
			}
			if err := s.g.Apply(got); err != nil {
				t.Fatalf("seed %d: %v", seed, err)
			}
		}
	}
	if activated == 0 {
		t.Errorf("over 32 seeds the asset never activated")
	}
}

// Case 6: an activation stays an act, not a play: it doesn't crowd out the other
// acts when it is affordable (a "kill this unit" cost would kill the unit every
// time), and a Generate toward it doesn't outrank them when it isn't.
func TestProgress_ActivationIsNotEager(t *testing.T) {
	picks := map[engine.ActionKind]int{}
	for seed := uint64(1); seed <= 64; seed++ {
		s := newScene(t)
		emptyPool(s, false)
		s.g.Players[0].Req = 1
		s.put(0, engine.ZoneReserves, activator(1, 0))
		picks[ProgressPolicy(seed)(s.g, s.g.LegalActions()).Kind]++
	}
	if picks[engine.ActActivate] == 0 || picks[engine.ActMarch] == 0 {
		t.Errorf("affordable activation beside a March: picks %v, want both at least once", picks)
	}

	picks = map[engine.ActionKind]int{}
	for seed := uint64(1); seed <= 64; seed++ {
		s := newScene(t)
		emptyPool(s, true)
		s.put(0, engine.ZoneReserves, activator(1, 0))
		picks[ProgressPolicy(seed)(s.g, s.g.LegalActions()).Kind]++
	}
	if picks[engine.ActGenerate] == 0 || picks[engine.ActMarch] == 0 {
		t.Errorf("reachable activation beside a March: picks %v, want both at least once", picks)
	}
}

// Case 7: while a hand card is short of Requisition it can't reach, a Reserves
// unit that yields Requisition mostly stays, to fund it on a later turn, but
// still marches sometimes, so a card it may never afford doesn't pin it there.
// Anything else marches freely.
func TestProgress_SavingHoldsMarchesOutOfReserves(t *testing.T) {
	marches := func(build func(s *scene)) int {
		n := 0
		for seed := uint64(1); seed <= 64; seed++ {
			s := newScene(t)
			emptyPool(s, false)
			build(s)
			if ProgressPolicy(seed)(s.g, s.g.LegalActions()).Kind == engine.ActMarch {
				n++
			}
		}
		return n
	}
	worker := func(s *scene) { s.put(0, engine.ZoneReserves, unit("Worker", 1, 1, engine.Keywords{})) }

	if n := marches(func(s *scene) { s.give(0, costlyUnit(9)); worker(s) }); n == 0 || n > 32 {
		t.Errorf("saving: %d of 64 marches, want some and at most half", n)
	}
	if n := marches(worker); n <= 32 {
		t.Errorf("empty hand: %d of 64 marches, want more than half", n)
	}
	noYield := func(s *scene) {
		s.give(0, costlyUnit(9))
		w := unit("Worker", 1, 1, engine.Keywords{})
		w.ReqYield = 0
		s.put(0, engine.ZoneReserves, w)
	}
	if n := marches(noYield); n <= 32 {
		t.Errorf("a unit with no Requisition to save: %d of 64 marches, want more than half", n)
	}
	tacShort := func(s *scene) { s.give(0, drawManeuver("Costly", 0, 9)); worker(s) }
	if n := marches(tacShort); n <= 32 {
		t.Errorf("short of Tactics only: %d of 64 marches, want more than half", n)
	}
}

// Case 8: over real starter games the activated abilities and the deck search
// fire. The floors catch a return to the starvation of directive #122: before it,
// 200 games gave Boneyard Scavenger 2 activations and Cliff of Portents 1.
func TestProgress_StarterGamesFireActivations(t *testing.T) {
	counter := &activationCounter{names: map[[2]int]string{}, fired: map[string]int{}}
	opt := SimOptions{DeckNames: simDeckNames, Policy: ProgressPolicy,
		Observer: func(g int, s uint64) engine.Observer {
			counter.game = g
			return counter
		}}
	if _, err := SimulateWith(simStarterDecks(t), 1000, 1, 400, opt); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Cliff of Portents", "Sinuous Growth", "Boneyard Scavenger"} {
		t.Logf("%-20s %d activations", name, counter.fired[name])
		if counter.fired[name] < 25 {
			t.Errorf("%s activated %d times in 1000 games, want at least 25", name, counter.fired[name])
		}
	}
	t.Logf("deck searches: %d", counter.searches)
	if counter.searches < 25 {
		t.Errorf("%d deck searches in 1000 games, want at least 25", counter.searches)
	}
}

// activationCounter counts activate actions by card and deck-search decisions.
type activationCounter struct {
	game     int
	names    map[[2]int]string
	fired    map[string]int
	searches int
}

func (c *activationCounter) Event(e engine.Event) {
	key := [2]int{c.game, e.Source}
	switch {
	case e.Kind == "deploy":
		c.names[key] = e.Card
	case e.Kind == "action" && e.Detail == "Activate":
		c.fired[c.names[key]]++
	case e.Decision != nil && e.Decision.Kind == engine.DecideSearch.String():
		c.searches++
	}
}
func (c *activationCounter) Settled(s engine.Snapshot) {}
```

Run them and confirm all 8 fail, then commit only this file as `test(T-0185): ...`.

### 2. `pkg/engine/resources.go` and `pkg/engine/actions.go`
Rename `convertReq` to `ConvertReq` everywhere: its comment and declaration and `convert` in `resources.go`, and the two uses in `check`'s `ActConvert` case in `actions.go`. Nothing else changes there.

### 3. `pkg/toolchain/linter/fences/fences.go`
In `allowedImports`, this line becomes:
```go
	"pkg/sim":            {"pkg/engine", "pkg/engine/ability"},
```

### 4. `pkg/sim/policy.go`
Replace the import block with:
```go
import (
	"fmt"
	"maps"
	"math/rand/v2"
	"slices"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/engine/ability"
)
```
Replace everything from the comment `// ProgressPolicy is a random policy that favors progress over passing.` to the end of the file with:
```go
// marchPast is how often a unit marches out of Reserves anyway while a hand card
// is short of Requisition it can't reach: often enough that a card the player
// may never afford doesn't pin its units there.
const marchPast = 0.25

// ProgressPolicy is a random policy that favors progress over passing. It
// picks only from legal and is deterministic for seed. Each call it prefers:
//  1. a play from hand: a Deploy or an Enact, uniform over card and zone;
//  2. a Generate or Convert that funds a hand card's cost (funds);
//  3. an Attack, March, Scratch or Activate, or a Generate or Convert that funds
//     an activated ability's cost (probability 0.9, uniform), else EndTurn;
//  4. EndTurn.
//
// A unit in Reserves that yields Requisition stays there, except marchPast of
// the time, while a hand card is short of Requisition it can't reach this turn,
// so that its yield is there on a later turn.
func ProgressPolicy(seed uint64) Policy {
	r := rand.New(rand.NewPCG(seed, ^seed))
	return func(g *engine.Game, legal []engine.Action) engine.Action {
		var plays, closers, acts []engine.Action
		end := legal[len(legal)-1] // LegalActions always ends with EndTurn
		p := g.Players[g.Active]
		pool := resources{p.Req, p.Tac}
		taps := legalTaps(g, legal)
		hand := handCosts(g, p)
		activations := activationCosts(g)
		saving := slices.ContainsFunc(hand, func(c neededCost) bool {
			return c.req > pool.req && !reachable(pool, taps, c.resources)
		})
		for _, a := range legal {
			switch a.Kind {
			case engine.ActDeploy, engine.ActEnact:
				plays = append(plays, a)
			case engine.ActGenerate, engine.ActConvert:
				if fundsAny(a, pool, taps, hand) {
					closers = append(closers, a)
				} else if fundsAny(a, pool, taps, activations) {
					acts = append(acts, a)
				}
			case engine.ActMarch:
				as := g.Asset(a.Asset)
				if saving && as.Zone == engine.ZoneReserves && as.Printed().ReqYield > 0 && r.Float64() >= marchPast {
					continue
				}
				acts = append(acts, a)
			case engine.ActAttack, engine.ActScratch, engine.ActActivate:
				acts = append(acts, a)
			}
		}
		switch {
		case len(plays) > 0:
			return plays[r.IntN(len(plays))]
		case len(closers) > 0:
			return closers[r.IntN(len(closers))]
		case len(acts) > 0 && r.Float64() < 0.9:
			return acts[r.IntN(len(acts))]
		}
		return end
	}
}

// resources is an amount of Requisition and Tactics.
type resources struct{ req, tac int }

// tap is what one asset can yield if it generates now, per resource. A resource
// its zone doesn't allow (FR-CAP-001) or its printed yield lacks is 0.
type tap struct {
	asset int
	yield resources
}

// neededCost is a cost the active player may want to pay. source is the asset
// whose activated ability it is, which won't tap to fund it, or 0 for a hand
// card: no asset has that ID.
type neededCost struct {
	resources
	source int
}

// legalTaps reads the active player's taps from legal, so the engine alone says
// which zone yields which resource.
func legalTaps(g *engine.Game, legal []engine.Action) []tap {
	var taps []tap
	for _, a := range legal {
		if a.Kind != engine.ActGenerate {
			continue
		}
		i := slices.IndexFunc(taps, func(t tap) bool { return t.asset == a.Asset })
		if i < 0 {
			taps = append(taps, tap{asset: a.Asset})
			i = len(taps) - 1
		}
		printed := g.Asset(a.Asset).Printed()
		if a.Resource == engine.Requisition {
			taps[i].yield.req = printed.ReqYield
		} else {
			taps[i].yield.tac = printed.TacYield
		}
	}
	return taps
}

// handCosts lists what p's playable hand cards cost now.
func handCosts(g *engine.Game, p *engine.Player) []neededCost {
	var out []neededCost
	for _, inst := range p.Hand {
		c := inst.Card
		if !c.Supported() || (len(c.DeployZones) == 0 && c.Type != engine.TypeManeuver) {
			continue
		}
		req, tac := g.EffectiveCost(inst)
		out = append(out, neededCost{resources: resources{req, tac}})
	}
	return out
}

// activationCosts lists the resource costs of the active player's activated
// abilities, leaving out an ability whose source must exhaust and already has.
// It doesn't check targets, so a tap may fund an ability that then has none.
func activationCosts(g *engine.Game) []neededCost {
	var out []neededCost
	for _, as := range g.Assets() {
		def := as.Printed().Ability
		if as.Controller != g.Active || def == nil {
			continue
		}
		for _, act := range def.Activated {
			c := neededCost{source: as.ID}
			usable := true
			for _, k := range act.Costs {
				switch k.Kind {
				case ability.CostResources:
					c.req += k.Req
					c.tac += k.Tac
				case ability.CostExhaustSelf:
					usable = !as.Exhausted
				}
			}
			if usable {
				out = append(out, c)
			}
		}
	}
	return out
}

// fundsAny reports whether a funds one of costs.
func fundsAny(a engine.Action, pool resources, taps []tap, costs []neededCost) bool {
	return slices.ContainsFunc(costs, func(c neededCost) bool { return funds(a, pool, taps, c) })
}

// funds reports whether Generate or Convert a adds a resource c is short of and
// leaves c reachable this turn. Requisition counts toward a Tactics shortfall
// only when no asset can tap for Tactics, through conversion (FR-RES-003).
func funds(a engine.Action, pool resources, taps []tap, c neededCost) bool {
	shortReq, shortTac := c.req > pool.req, c.tac > pool.tac
	var rest []tap
	for _, t := range taps {
		if t.asset != a.Asset && t.asset != c.source {
			rest = append(rest, t)
		}
	}
	if a.Kind == engine.ActConvert {
		return shortTac && reachable(resources{pool.req - engine.ConvertReq, pool.tac + 1}, rest, c.resources)
	}
	i := slices.IndexFunc(taps, func(t tap) bool { return t.asset == a.Asset })
	if a.Asset == c.source || i < 0 {
		return false
	}
	y := taps[i].yield
	if a.Resource == engine.Tactics {
		return y.tac > 0 && shortTac && reachable(resources{pool.req, pool.tac + y.tac}, rest, c.resources)
	}
	viaConvert := shortTac && !slices.ContainsFunc(taps, func(t tap) bool { return t.yield.tac > 0 })
	return y.req > 0 && (shortReq || viaConvert) && reachable(resources{pool.req + y.req, pool.tac}, rest, c.resources)
}

// reachable reports whether pool and taps can pay c this turn: each asset taps
// for one resource, and Requisition beyond c's converts to Tactics.
func reachable(pool resources, taps []tap, c resources) bool {
	states := map[resources]bool{pool: true}
	for _, t := range taps {
		next := maps.Clone(states)
		for s := range states {
			next[resources{s.req + t.yield.req, s.tac}] = true
			next[resources{s.req, s.tac + t.yield.tac}] = true
		}
		states = next
	}
	for s := range states {
		if s.req >= c.req && s.tac+(s.req-c.req)/engine.ConvertReq >= c.tac {
			return true
		}
	}
	return false
}
```

## Allowed Files
- `pkg/sim/policy.go`
- `pkg/sim/policy_test.go` (Edit 1 only, in the test commit)
- `pkg/engine/resources.go`
- `pkg/engine/actions.go`
- `pkg/toolchain/linter/fences/fences.go`

The lead edits `docs/architecture/m2-ability-model.md` at close: #122 (c), closing Open Question 18 as provisional with the user's reopen trigger, and the import table's new entry. Don't edit it.

## Done When
- [x] `go test ./pkg/sim -run TestProgress` passes
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

`go tool task sim-compare` is expected to differ (Revision item 3), so don't run it.

## Log
- 2026-11-06 (lead-architect): Written for #122, `ready`. The Wave D.1 directives ahead of it are applied, and #123 still waits on the user. After the merge the lead re-measures with `go tool task sim-events` against the user's reopen trigger (Revision item 4) and records (c).
- 2026-11-06 (kernel-engineer): Implemented for #122, `review`.
  - Added 8 acceptance tests to `pkg/sim/policy_test.go` in commit `1df9059`, confirmed all 8 failed on baseline.
  - Exported `ConvertReq` in `pkg/engine/resources.go` and updated call sites in `pkg/engine/actions.go`.
  - Widened `allowedImports` in `pkg/toolchain/linter/fences/fences.go` to permit `pkg/sim` -> `pkg/engine/ability`.
  - Rewrote `ProgressPolicy` in `pkg/sim/policy.go` with reachable cost checks, `marchPast = 0.25`, activation budgeting into `acts`, and legal tap filtering.
  - Verification: `go test ./pkg/sim -run TestProgress` passed; `go tool task check` passed all vet, test, and linters; `go tool task sim` ran clean with matching metrics (P0 78.5% / P1 19.0% / draws 2.5%, avg 24.7 turns, 2778 attacks).
- 2026-11-06 (lead-architect): Approved. Run: one resume, after `git branch -a` was refused while hunting for the task file (the prompt didn't name its location; the resume did). The diff's `policy.go` and tests match the dry run byte for byte, and only `Allowed Files` changed. All 8 tests fail at 1df9059. Re-ran `go test ./pkg/sim -run TestProgress`, `go tool task check` and `go tool task sim`: exit 0. Merged as 3b92cc0. Re-measured on the merge, 200 games, `-seed 1`: P0 78.5%, P1 19.0%, draws 2.5%, 24.7 turns; *Sinuous Growth* 232, *Cliff of Portents* 20, *Boneyard Scavenger* 14, deck searches 20. The reopen trigger (Revision item 4) doesn't fire. #122 (c) recorded in the design note.
