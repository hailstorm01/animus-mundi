---
id: T-0215
title: Replay and ProgressPolicy handle a deploy from a discard pile (wave F harness, flash)
kind: harness
tier: flash
wave: F
status: done
branch: m2/autoloop-wave-f/T-0215-discard-deploy-clients
depends_on: [T-0214]
claims: [file:pkg/eventlog/replay.go (actionOf reads Event.From through one zoneNamed; diff compares From), file:pkg/sim/policy.go (playCosts counts discard deploys), export:engine.Game.DiscardDeployCost]
cards: []
requirements: [FR-TRN-003, FR-TRN-003a]
---

## Goal
A recorded deploy from a discard pile replays, and `ProgressPolicy` budgets toward one, so *Veil Render* (T-0216) can be simulated and replayed. Part of directive #128's last part: replay, and `ProgressPolicy` counting the deploy's cost. The play screen's label is T-0217, split off here under the D2 size rule.

## Inputs
- Board directive #128 (structural review 9) and its Acceptance: "a recorded discard deploy replays with no divergence". #122 (a): `handCosts` counts "discard-pile deploy costs when Wave F adds them".
- T-0214 (merge d7d9858) added `Action.From` and the action event's `From` (`"Discard"` for a discard deploy, empty for the hand; `pkg/engine/events.go:166-169`), and `LegalActions` offers discard deploys (`pkg/engine/actions.go:304-308`). Today:
  - `actionOf` (`pkg/eventlog/replay.go:255`) never reads `e.From`, so a replayed discard deploy is a hand deploy and is rejected ("card N is not in your hand"). `diff` (`:218-219`) and `structured` (`:245`) don't compare `From`.
  - `handCosts` (`pkg/sim/policy.go:145`) reads only the hand. The sim package can't call `deployPermission` (`pkg/engine/placement.go:85`), which is unexported and is the one reader of the permission.
- **The reader:** `Game.DiscardDeployCost(inst) (req, tac int, ok bool)` in `pkg/engine/costs.go`, beside `ActivatedCost`, which it mirrors. The cost is `EffectiveCost`, and `ok` says the card is in its owner's discard pile, prints the permission, and the permission's (and the card's `ExtraCosts`) non-resource costs can be paid now (`otherCostsPayable` of `deployPlacement`). It calls the existing readers, so the permission is still read in one place. Placement: an exported client-planning reader in `engine`, like `ActivatedCost` and `EffectiveCost`.
- **One zone-name parser:** `actionOf`'s zone loop becomes `zoneNamed`, used for both `Zone` and `From`, so there is one copy of it.
- `FormatEvent` (`pkg/eventlog/eventlog.go:24`) and the view's log (`pkg/view/log.go:44`) print an action event's `Detail` only ("Deploy"), never its card or zone, so neither needs a change.
- Size: three non-test files (`replay.go`, `policy.go`, `costs.go`). The play screen's label needs two more (`pkg/view/view.go`, whose discard cards carry no `Inst`, and `cmd/cardtown/play_screen.go`), so they are T-0217.
- **Dry run:** the lead made exactly these edits and tests in a throwaway worktree of the run branch at 6f69ee6. `gofmt -l pkg cmd` clean; `go tool task check` exit 0; `go tool task sim-compare -- autoloop/m2-wave-f` prints `[PASS]` (no printed card has the permission until T-0216). With the tests added and the three non-test files as at 6f69ee6, every new test fails: `TestReplay_DiscardDeploysRoundTrip` and the two divergence tests fail (the record is rejected at its first discard deploy), `TestProgress_DiscardDeployBudgets` fails at seed 1 (no Generate), and the engine tests don't compile (`DiscardDeployCost` is undefined). Blobs after the edits: `pkg/eventlog/replay.go` 130e4e8, `pkg/sim/policy.go` 4a39bf3, `pkg/engine/costs.go` f1e9b9e, `pkg/engine/discard_deploy_cost_test.go` eede490, `pkg/eventlog/replay_test.go` 9777c57, `pkg/sim/policy_test.go` 1a2a444.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log, for the test commit (each test fails) and for the implementation commit (each passes). The reviewer re-runs them.
1. **A recorded discard deploy replays with no divergence** (#128): `go -C W test ./pkg/eventlog -run TestReplay_DiscardDeploysRoundTrip -count=1` passes.
2. **An unknown `From` is an error, as an unknown zone is:** `go -C W test ./pkg/eventlog -run TestReplay_DiscardDeployFromAnUnknownZoneIsAnError -count=1` passes.
3. **A replayed action event whose `From` differs from the record diverges:** `go -C W test ./pkg/eventlog -run TestReplay_DeployWithAnotherFromDiverges -count=1` passes.
4. **The reader:** `go -C W test ./pkg/engine -run TestDiscardDeployCost -count=1` passes (two tests).
5. **`ProgressPolicy` budgets toward a discard deploy it may make, and not toward one it can't:** `go -C W test ./pkg/sim -run TestProgress_DiscardDeployBudgets -count=1` passes.
6. **No other change to the sims:** `go -C W tool task sim-compare -- autoloop/m2-wave-f` prints `[PASS] starter and stress sims match autoloop/m2-wave-f byte for byte, exit codes included`.
7. **Only the listed files:** `git -C W diff autoloop/m2-wave-f --stat` lists only the six Allowed Files and this task file.
8. `go -C W tool task check` and `go -C W tool task sim` pass.

## Tests (copy exactly; the `test(T-0215)` commit)
**(a)** Create `pkg/engine/discard_deploy_cost_test.go` with exactly:
```go
package engine

import "testing"

// Tests for T-0215 (FR-TRN-003, FR-TRN-003a, directive #128): DiscardDeployCost,
// the reader a client plans a discard deploy with. It uses T-0214's ddUnit,
// whose permission costs killing an allied unit. Player 0 owns every card.

func TestDiscardDeployCost_PermissionAndPayableCosts(t *testing.T) {
	s := newScene(t)
	inst := s.inDiscard(0, ddUnit("Phoenix", 2, 1))[0]
	if req, tac, ok := s.g.DiscardDeployCost(inst); req != 2 || tac != 1 || ok {
		t.Errorf("with no allied unit to kill: got R%d T%d %v, want R2 T1 false (FR-TRN-003a)", req, tac, ok)
	}
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	if req, tac, ok := s.g.DiscardDeployCost(inst); req != 2 || tac != 1 || !ok {
		t.Errorf("with an allied unit to kill: got R%d T%d %v, want R2 T1 true (FR-TRN-003)", req, tac, ok)
	}
	s.setPool(0, 0, 0)
	if _, _, ok := s.g.DiscardDeployCost(inst); !ok {
		t.Errorf("an empty pool doesn't change ok: it doesn't check the resources against the pool")
	}
}

func TestDiscardDeployCost_OnlyAPermittedCardInItsOwnersPile(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	inHand := s.give(0, ddUnit("Phoenix", 2, 0))
	if _, _, ok := s.g.DiscardDeployCost(inHand); ok {
		t.Errorf("a permitted card in hand: ok true, want false")
	}
	plain := unit("Plain", 1, 1, Keywords{})
	plain.ReqCost = 1
	if _, _, ok := s.g.DiscardDeployCost(s.inDiscard(0, plain)[0]); ok {
		t.Errorf("a discard card with no permission: ok true, want false (FR-TRN-003)")
	}
}
```

**(b)** Append to the end of `pkg/eventlog/replay_test.go`, after one blank line (its imports already cover `fmt`, `strings` and `ability`):
```go
// Tests for T-0215 (directive #128, FR-TRN-003): a deploy from a discard pile
// replays from its action event's From.

// revenantDeck is a synthetic deck whose units may be deployed from their
// owner's discard pile by killing an allied unit.
func revenantDeck() engine.DeckList {
	hq := &engine.Card{Name: "Revenant HQ", Type: engine.TypeHeadquarters, Power: 3, Health: 20, ReqYield: 3, TacYield: 2,
		DeployZones: []engine.Zone{engine.ZoneHeadquarters}}
	u := &engine.Card{Name: "Revenant", Type: engine.TypeUnit, Power: 2, Health: 1, ReqCost: 1, ReqYield: 1, TacYield: 1,
		DeployZones: []engine.Zone{engine.ZoneReserves, engine.ZoneMainTheater},
		Ability: &ability.Def{Statics: []ability.Static{{
			Kind: ability.StaticDeployPermission, ActiveZone: ability.ZoneDiscard,
			Costs: []ability.Cost{{Kind: ability.CostKill,
				Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}}},
		}}}}
	main := make([]*engine.Card, 30)
	for i := range main {
		main[i] = u
	}
	return engine.DeckList{HQ: hq, Main: main}
}

var revenantNames = [2]string{"revenant", "revenant"}

func revenantResolve(name string) (engine.DeckList, error) {
	if name != "revenant" {
		return engine.DeckList{}, fmt.Errorf("unknown deck %q", name)
	}
	return revenantDeck(), nil
}

func revenantGames(t *testing.T) [][]engine.Event {
	t.Helper()
	d := revenantDeck()
	games, err := play([2]engine.DeckList{d, d}, revenantNames, 30, 7)
	if err != nil {
		t.Fatalf("revenant games: %v", err)
	}
	return games
}

// isDeploy reports whether e is a Deploy action event from the pile from names
// ("" for the hand).
func isDeploy(e engine.Event, from string) bool {
	return e.Kind == "action" && e.Detail == "Deploy" && e.From == from
}

func TestReplay_DiscardDeploysRoundTrip(t *testing.T) {
	seen := 0
	for gi, evs := range revenantGames(t) {
		for _, e := range evs {
			if isDeploy(e, "Discard") {
				seen++
			}
		}
		if _, err := eventlog.Replay(evs, len(evs), revenantResolve); err != nil {
			t.Fatalf("game %d: %v", gi+1, err)
		}
	}
	if seen == 0 {
		t.Errorf("no Deploy action event with From %q over 30 games", "Discard")
	}
}

func TestReplay_DiscardDeployFromAnUnknownZoneIsAnError(t *testing.T) {
	for _, evs := range revenantGames(t) {
		for i, e := range evs {
			if !isDeploy(e, "Discard") {
				continue
			}
			cp := append([]engine.Event{}, evs...)
			cp[i].From = "Nowhere"
			_, err := eventlog.Replay(cp, len(cp), revenantResolve)
			want := fmt.Sprintf("replay diverged at seq %d: unknown zone %q", e.Seq, "Nowhere")
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("From recorded as Nowhere: error %v, want it to contain %q", err, want)
			}
			return
		}
	}
	t.Fatal("no game holds a Deploy event with From Discard")
}

func TestReplay_DeployWithAnotherFromDiverges(t *testing.T) {
	for _, evs := range revenantGames(t) {
		for i, e := range evs {
			if !isDeploy(e, "") {
				continue
			}
			cp := append([]engine.Event{}, evs...)
			cp[i].From = "Hand"
			_, err := eventlog.Replay(cp, len(cp), revenantResolve)
			want := fmt.Sprintf("replay diverged at seq %d: action event fields", e.Seq)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("a hand Deploy recorded with From Hand: error %v, want it to contain %q", err, want)
			}
			return
		}
	}
	t.Fatal("no game holds a Deploy event from hand")
}
```

**(c)** Append to the end of `pkg/sim/policy_test.go`, after one blank line (its imports already cover `ability`):
```go
// Tests for T-0215 (directive #128, #122 (a), FR-TRN-003): ProgressPolicy
// counts a discard deploy's cost with the hand cards'.

// revenant is a vanilla R1 unit that may be deployed from its owner's discard
// pile, paying costs (none for nil) in addition to its own.
func revenant(costs []ability.Cost) *engine.Card {
	c := costlyUnit(1)
	c.Ability = &ability.Def{Statics: []ability.Static{{
		Kind: ability.StaticDeployPermission, ActiveZone: ability.ZoneDiscard, Costs: costs}}}
	return c
}

// inDiscard puts c into player 0's discard pile.
func inDiscard(s *scene, c *engine.Card) *engine.Instance {
	inst := &engine.Instance{ID: s.id(), Card: c, Owner: 0}
	s.g.Players[0].Discard = append(s.g.Players[0].Discard, inst)
	return inst
}

// A discard card it may deploy (R1) makes the Readied HQ tap for it, and the
// deploy follows. A card with no permission, and one whose kill cost has no
// allied unit to kill, tap nothing.
func TestProgress_DiscardDeployBudgets(t *testing.T) {
	for seed := uint64(1); seed <= 16; seed++ {
		s := newScene(t)
		emptyPool(s, true)
		inst := inDiscard(s, revenant(nil))
		policy := ProgressPolicy(seed)
		got := policy(s.g, s.g.LegalActions())
		if got.Kind != engine.ActGenerate || got.Resource != engine.Requisition {
			t.Fatalf("seed %d: R1 from the discard pile is in reach: got %+v, want a Requisition Generate", seed, got)
		}
		if err := s.g.Apply(got); err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		got = policy(s.g, s.g.LegalActions())
		if got.Kind != engine.ActDeploy || got.Card != inst.ID || got.From != engine.ZoneDiscard {
			t.Errorf("seed %d: after the Generate, got %+v, want the Deploy of card %d from the discard pile", seed, got, inst.ID)
		}
	}
	kill := []ability.Cost{{Kind: ability.CostKill,
		Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}}}
	for _, c := range []*engine.Card{costlyUnit(1), revenant(kill)} {
		for seed := uint64(1); seed <= 16; seed++ {
			s := newScene(t)
			emptyPool(s, true)
			inDiscard(s, c)
			if got := ProgressPolicy(seed)(s.g, s.g.LegalActions()); got.Kind == engine.ActGenerate {
				t.Errorf("seed %d: %s can't be deployed from the discard pile now: got %+v, want no Generate", seed, c.Name, got)
			}
		}
	}
}
```
Commit these three files alone as `test(T-0215): discard deploys replay and are budgeted`, then run cases 1-5 and record that each fails.

## Edits (copy exactly; the implementation commit)
Line numbers are at 6f69ee6. The hunks below are the whole change to the three non-test files, tabs as in the surrounding code. Make them with your file-editing tool, not with a shell command.
```diff
diff --git a/pkg/engine/costs.go b/pkg/engine/costs.go
index 0405807..f1e9b9e 100644
--- a/pkg/engine/costs.go
+++ b/pkg/engine/costs.go
@@ -116,6 +116,22 @@ func (g *Game) ActivatedCost(a *Asset, i int) (req, tac int, ok bool) {
 	return req, tac, g.otherCostsPayable(activatedPlacement(a.Controller, a, act))
 }
 
+// DiscardDeployCost returns the Requisition and Tactics that deploying inst from
+// its owner's discard pile costs, and whether it may be deployed from there: it
+// is in that pile, prints a discard-active deploy permission (FR-TRN-003), and
+// the permission's and the card's additional costs other than resources can be
+// paid now (FR-TRN-003a). Like ActivatedCost, it checks neither the resources
+// against the pool nor the zone, so a client can plan for a deploy it can't
+// make yet. LegalActions says what can be deployed now.
+func (g *Game) DiscardDeployCost(inst *Instance) (req, tac int, ok bool) {
+	req, tac = g.EffectiveCost(inst)
+	if pileCard(g.Players[inst.Owner], inst.ID, ZoneDiscard) == nil || deployPermission(inst.Card) == nil {
+		return req, tac, false
+	}
+	// No cost depends on the zone, so Reserves stands in for either.
+	return req, tac, g.otherCostsPayable(deployPlacement(inst.Owner, inst, ZoneReserves, ZoneDiscard))
+}
+
 // otherCostsPayable reports whether every cost of pl except its resources can be
 // paid now: exhausting the source needs it Readied (CostExhaustSelf), and a kill
 // or exhaust cost needs at least one option (SVY-R29). A discard cost is always
diff --git a/pkg/eventlog/replay.go b/pkg/eventlog/replay.go
index 1af842e..130e4e8 100644
--- a/pkg/eventlog/replay.go
+++ b/pkg/eventlog/replay.go
@@ -216,7 +216,7 @@ func diff(rec, got engine.Event) string {
 	case rec.Turn != got.Turn || rec.Phase != got.Phase:
 		return fmt.Sprintf("turn and phase: recorded T%d %s, replayed T%d %s", rec.Turn, rec.Phase, got.Turn, got.Phase)
 	case rec.Player != got.Player || rec.Card != got.Card || rec.Inst != got.Inst || rec.Source != got.Source ||
-		rec.Target != got.Target || rec.Zone != got.Zone || rec.Amount != got.Amount || rec.Resource != got.Resource:
+		rec.Target != got.Target || rec.Zone != got.Zone || rec.From != got.From || rec.Amount != got.Amount || rec.Resource != got.Resource:
 		return fmt.Sprintf("%s event fields: recorded %+v, replayed %+v", rec.Kind, structured(rec), structured(got))
 	}
 	if (rec.Decision == nil) != (got.Decision == nil) ||
@@ -242,8 +242,8 @@ func diff(rec, got engine.Event) string {
 	return ""
 }
 
-func structured(e engine.Event) [8]any {
-	return [8]any{e.Player, e.Card, e.Inst, e.Source, e.Target, e.Zone, e.Amount, e.Resource}
+func structured(e engine.Event) [9]any {
+	return [9]any{e.Player, e.Card, e.Inst, e.Source, e.Target, e.Zone, e.From, e.Amount, e.Resource}
 }
 
 func decisionEqual(a, b engine.DecisionRecord) bool {
@@ -265,16 +265,27 @@ func actionOf(e engine.Event) (engine.Action, error) {
 	if a.Kind == engine.ActGenerate && e.Resource == engine.Tactics.String() {
 		a.Resource = engine.Tactics
 	}
+	var err error
 	if e.Zone != "" {
-		zoneFound := false
-		for z := engine.ZoneDeck; z <= engine.ZoneStack; z++ {
-			if z.String() == e.Zone {
-				a.Zone, zoneFound = z, true
-			}
+		if a.Zone, err = zoneNamed(e.Zone); err != nil {
+			return a, err
 		}
-		if !zoneFound {
-			return a, fmt.Errorf("unknown zone %q", e.Zone)
+	}
+	if e.From != "" {
+		if a.From, err = zoneNamed(e.From); err != nil {
+			return a, err
 		}
 	}
 	return a, nil
 }
+
+// zoneNamed returns the zone whose String is name: an action event's Zone or
+// From.
+func zoneNamed(name string) (engine.Zone, error) {
+	for z := engine.ZoneDeck; z <= engine.ZoneStack; z++ {
+		if z.String() == name {
+			return z, nil
+		}
+	}
+	return 0, fmt.Errorf("unknown zone %q", name)
+}
diff --git a/pkg/sim/policy.go b/pkg/sim/policy.go
index 004b1e9..4a39bf3 100644
--- a/pkg/sim/policy.go
+++ b/pkg/sim/policy.go
@@ -39,21 +39,22 @@ func RandomPolicy(seed uint64, endBias float64) Policy {
 	}
 }
 
-// marchPast is how often a unit marches out of Reserves anyway while a hand card
-// is short of Requisition it can't reach: often enough that a card the player
+// marchPast is how often a unit marches out of Reserves anyway while a playable
+// card is short of Requisition it can't reach: often enough that a card the player
 // may never afford doesn't pin its units there.
 const marchPast = 0.25
 
 // ProgressPolicy is a random policy that favors progress over passing. It
 // picks only from legal and is deterministic for seed. Each call it prefers:
-//  1. a play from hand: a Deploy or an Enact, uniform over card and zone;
-//  2. a Generate or Convert that funds a hand card's cost (funds);
+//  1. a play: a Deploy or an Enact, from hand or a Deploy from the discard pile,
+//     uniform over card and zone;
+//  2. a Generate or Convert that funds a playable card's cost (funds, playCosts);
 //  3. an Attack, March, Scratch or Activate, or a Generate or Convert that funds
 //     an activated ability's cost (probability 0.9, uniform), else EndTurn;
 //  4. EndTurn.
 //
 // A unit in Reserves that yields Requisition stays there, except marchPast of
-// the time, while a hand card is short of Requisition it can't reach this turn,
+// the time, while a playable card is short of Requisition it can't reach this turn,
 // so that its yield is there on a later turn.
 func ProgressPolicy(seed uint64) Policy {
 	r := rand.New(rand.NewPCG(seed, ^seed))
@@ -63,9 +64,9 @@ func ProgressPolicy(seed uint64) Policy {
 		p := g.Players[g.Active]
 		pool := resources{p.Req, p.Tac}
 		taps := legalTaps(g, legal)
-		hand := handCosts(g, p)
+		cards := playCosts(g, p)
 		activations := activationCosts(g)
-		saving := slices.ContainsFunc(hand, func(c neededCost) bool {
+		saving := slices.ContainsFunc(cards, func(c neededCost) bool {
 			return c.req > pool.req && !reachable(pool, taps, c.resources)
 		})
 		for _, a := range legal {
@@ -73,7 +74,7 @@ func ProgressPolicy(seed uint64) Policy {
 			case engine.ActDeploy, engine.ActEnact:
 				plays = append(plays, a)
 			case engine.ActGenerate, engine.ActConvert:
-				if fundsAny(a, pool, taps, hand) {
+				if fundsAny(a, pool, taps, cards) {
 					closers = append(closers, a)
 				} else if fundsAny(a, pool, taps, activations) {
 					acts = append(acts, a)
@@ -111,8 +112,8 @@ type tap struct {
 }
 
 // neededCost is a cost the active player may want to pay. source is the asset
-// whose activated ability it is, which won't tap to fund it, or 0 for a hand
-// card: no asset has that ID.
+// whose activated ability it is, which won't tap to fund it, or 0 for a card to
+// play: no asset has that ID.
 type neededCost struct {
 	resources
 	source int
@@ -141,8 +142,10 @@ func legalTaps(g *engine.Game, legal []engine.Action) []tap {
 	return taps
 }
 
-// handCosts lists what p's playable hand cards cost now.
-func handCosts(g *engine.Game, p *engine.Player) []neededCost {
+// playCosts lists what p's playable cards cost now: its hand cards, and the
+// cards it may deploy from its discard pile by their own permission, leaving out
+// one whose other costs can't be paid now (Game.DiscardDeployCost, FR-TRN-003).
+func playCosts(g *engine.Game, p *engine.Player) []neededCost {
 	var out []neededCost
 	for _, inst := range p.Hand {
 		c := inst.Card
@@ -152,6 +155,14 @@ func handCosts(g *engine.Game, p *engine.Player) []neededCost {
 		req, tac := g.EffectiveCost(inst)
 		out = append(out, neededCost{resources: resources{req, tac}})
 	}
+	for _, inst := range p.Discard {
+		if !inst.Card.Supported() || len(inst.Card.DeployZones) == 0 {
+			continue
+		}
+		if req, tac, ok := g.DiscardDeployCost(inst); ok {
+			out = append(out, neededCost{resources: resources{req, tac}})
+		}
+	}
 	return out
 }
 
```
Then run `gofmt -l pkg cmd`; it must print nothing.

## Allowed Files
- `pkg/eventlog/replay.go`
- `pkg/sim/policy.go`
- `pkg/engine/costs.go`
- `pkg/engine/discard_deploy_cost_test.go` (new)
- `pkg/eventlog/replay_test.go` (new tests appended only; no existing test changes)
- `pkg/sim/policy_test.go` (new tests appended only; no existing test changes)

## Done When
- [x] Cases 1-8 hold, with their commands and results in the Log, each test failing at the test commit
- [x] `go tool task sim-compare -- autoloop/m2-wave-f` passes, byte for byte
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-10 lead-architect (wave F run, session 13): queued with T-0213, split off *Veil Render* under the D2 size rule (D2.4). Depends on T-0214's `Action.From` and `Event.From`.
- 2026-11-11 lead-architect (wave F run, session 16): wrote the Inputs, Acceptance Cases, Tests and Edits against the run branch at 6f69ee6, and dry-ran them (Inputs). The original Covers needed five non-test files, so under the D2 size rule the play screen's label, with the view's `Inst` on the seat's own discard cards, went to T-0217 (`queued`). `FormatEvent` and the view's log need no change (Inputs). Harness, so it skips `tests` and goes to `ready`; tier `flash`, because the edits and tests are exact and dry-run. Every command here is on the headless allow-list, one per line. Directive #128 stays `open: before T-0216` for T-0217. Next: `kernel-engineer` on Flash.
- 2026-11-11 kernel-engineer (wave F run, session 17): implemented T-0215.
  - Test commit c102b21: added pkg/engine/discard_deploy_cost_test.go, appended discard deploy replay tests to pkg/eventlog/replay_test.go, and appended discard deploy budget tests to pkg/sim/policy_test.go. Verified cases 1-5 failed:
    - Case 1: `go -C W test ./pkg/eventlog -run TestReplay_DiscardDeploysRoundTrip -count=1` -> FAIL (replay diverged at seq 101: the recorded action was rejected: illegal action: card 53 is not in your hand)
    - Case 2: `go -C W test ./pkg/eventlog -run TestReplay_DiscardDeployFromAnUnknownZoneIsAnError -count=1` -> FAIL (From recorded as Nowhere: error replay diverged at seq 101: the recorded action was rejected: illegal action: card 53 is not in your hand, want it to contain "unknown zone \"Nowhere\"")
    - Case 3: `go -C W test ./pkg/eventlog -run TestReplay_DeployWithAnotherFromDiverges -count=1` -> FAIL (a hand Deploy recorded with From Hand: error replay diverged at seq 101: the recorded action was rejected: illegal action: card 53 is not in your hand, want it to contain "action event fields")
    - Case 4: `go -C W test ./pkg/engine -run TestDiscardDeployCost -count=1` -> FAIL (build failed: s.g.DiscardDeployCost undefined)
    - Case 5: `go -C W test ./pkg/sim -run TestProgress_DiscardDeployBudgets -count=1` -> FAIL (seed 1: R1 from the discard pile is in reach: got Scratch, want a Requisition Generate)
  - Search for reuse: DiscardDeployCost reuses EffectiveCost, pileCard, deployPermission, deployPlacement, and otherCostsPayable. zoneNamed in replay.go consolidates zone-name resolution for both Zone and From fields. playCosts in policy.go reuses DiscardDeployCost and Card.Supported/DeployZones.
  - Implementation commit 922247a:
    - pkg/engine/costs.go: implemented DiscardDeployCost mirroring ActivatedCost.
    - pkg/eventlog/replay.go: diff and structured compare Event.From; actionOf reads From and Zone through consolidated zoneNamed.
    - pkg/sim/policy.go: replaced handCosts with playCosts, considering both hand cards and playable discard cards.
  - Verification & acceptance results:
    - Case 1: `go -C W test ./pkg/eventlog -run TestReplay_DiscardDeploysRoundTrip -count=1` -> PASS (ok animus-mundi/pkg/eventlog 1.281s)
    - Case 2: `go -C W test ./pkg/eventlog -run TestReplay_DiscardDeployFromAnUnknownZoneIsAnError -count=1` -> PASS (ok animus-mundi/pkg/eventlog 1.241s)
    - Case 3: `go -C W test ./pkg/eventlog -run TestReplay_DeployWithAnotherFromDiverges -count=1` -> PASS (ok animus-mundi/pkg/eventlog 1.223s)
    - Case 4: `go -C W test ./pkg/engine -run TestDiscardDeployCost -count=1` -> PASS (ok animus-mundi/pkg/engine 1.024s)
    - Case 5: `go -C W test ./pkg/sim -run TestProgress_DiscardDeployBudgets -count=1` -> PASS (ok animus-mundi/pkg/sim 0.663s)
    - Case 6: `go -C W tool task sim-compare -- autoloop/m2-wave-f` -> PASS ([PASS] starter and stress sims match autoloop/m2-wave-f byte for byte, exit codes included)
    - Case 7: `git -C W diff autoloop/m2-wave-f --stat` -> lists only 6 Allowed Files plus task file
    - Case 8: `go -C W tool task check` -> PASS (gofmt, go vet, all package tests, all linters clean); `go -C W tool task sim` -> PASS (200 starter playouts, all invariants held).
- 2026-11-11 lead-architect (wave F run, session 16, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-f` as 49a1894. Flash run `c55d138e-cdcd-4b8c-8be8-6e3094641f40`: exit 0, `status: SUCCESS`, no `denied_actions` key, and the transcript has no `permission check failed` step; `.err` holds only agy's "root agent idle" notice. (The engineer's Log calls this "session 17"; it is this session's dispatch.) Review of `git diff autoloop/m2-wave-f...m2/autoloop-wave-f/T-0215-discard-deploy-clients` (ba53f81):
  - Scope: only the six Allowed Files and this file changed. The `test(T-0215)` commit c102b21 holds only the three test files; the implementation commit 922247a holds only `costs.go`, `replay.go` and `policy.go` and leaves every test file as c102b21 has it.
  - Checked against the diff, not the Log: all six files are byte-identical to the lead's dry run (blobs 130e4e8, 4a39bf3, f1e9b9e, eede490, 9777c57, 1a2a444), so the hunks are exactly the Edits and Tests. The dry run showed each new test failing without the implementation, as the Log also reports.
  - Reuse and fences: `DiscardDeployCost` calls `EffectiveCost`, `pileCard`, `deployPermission`, `deployPlacement` and `otherCostsPayable`, so the permission keeps one reader; `zoneNamed` replaces `actionOf`'s zone loop and serves `Zone` and `From`. No card name in package code. `engine` gains one exported client-planning reader beside `ActivatedCost`.
  - Re-run by the reviewer in a scratch worktree of the task branch: `gofmt -l pkg cmd` clean; `go tool task check` exit 0; `go tool task sim` exit 0, clean, 200 games, P0 74.0% / P1 25.0% / draws 1.0%, 49/50 and 44/50 main cards executable; `go tool task sim-compare -- autoloop/m2-wave-f` `[PASS]`, byte for byte. The merge's tree matches the task tip's.
  - #128 stays `open: before T-0216` for T-0217, its last piece.
