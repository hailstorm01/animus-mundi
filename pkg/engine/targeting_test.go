package engine

import (
	"reflect"
	"sort"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for chosen targets on triggered abilities (T-0017): conjunction
// filters (FR-TGT-001), two-stage verification (FR-TGT-003), Ward
// (FR-KWD-002h), non-combat damage mitigation (FR-REP-004) and cascading
// triggers (FR-TRG-004). Cards here are synthetic: a unit deploys from hand
// into Reserves and its deploy trigger deals damage to chosen entities.
// Choices are observed through the scripted chooser's recorded Decisions.
// scene.putScratch(p) is a requested helper (see the task's Helpers Needed).

// tgt is one "chosen" slot with the given conjunction filter.
func tgt(a ability.Allegiance, k ability.Kinds, nonScratch bool) ability.Target {
	return ability.Target{Filter: ability.Filter{Allegiance: a, Kinds: k, NonScratch: nonScratch}}
}

// dmgTo is "deal n damage to chosen <slot>".
func dmgTo(n, slot int) ability.Step {
	return ability.Step{Op: ability.OpDealDamage, Amount: ability.Value{N: n}, Subject: ability.SubjectTarget, Target: slot}
}

// targetTrigger is one "When this deploys, <steps>" ability with chosen targets.
func targetTrigger(targets []ability.Target, steps ...ability.Step) ability.Trigger {
	return ability.Trigger{Hook: ability.HookDeploy, Effect: ability.Effect{Targets: targets, Steps: steps}}
}

// assetOf finds the battlefield asset whose card is c.
func assetOf(g *Game, c *Card) *Asset {
	for _, a := range g.Assets() {
		if a.Inst.Card == c {
			return a
		}
	}
	return nil
}

func refs(d Decision) []int {
	out := make([]int, len(d.Options))
	for i, o := range d.Options {
		out[i] = o.Ref
	}
	sort.Ints(out)
	return out
}

func ids(as ...*Asset) []int {
	out := make([]int, len(as))
	for i, a := range as {
		out[i] = a.ID
	}
	sort.Ints(out)
	return out
}

// totalDamage is the damage on every asset still on the battlefield.
func totalDamage(g *Game) int {
	n := 0
	for _, a := range g.Assets() {
		n += a.Damage
	}
	return n
}

// scriptRefs installs a chooser for seat p that answers DecideOrder with
// order and each DecideTarget by picking the option that refers to the next
// wanted asset. It records every Decision like scene.script.
func (s *scene) scriptRefs(p PlayerID, order []int, want ...*Asset) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[p] = nil
	next := 0
	s.g.choosers[p] = func(_ *Game, d Decision) Answer {
		s.asks[p] = append(s.asks[p], d)
		if d.Kind == DecideOrder {
			return Answer{Picks: order}
		}
		if next >= len(want) {
			s.t.Errorf("player %d chooser asked for a target past the end of its script: %q", p, d.Prompt)
			return Answer{Timeout: true}
		}
		w := want[next]
		next++
		for i, o := range d.Options {
			if o.Ref == w.ID {
				return Answer{Picks: []int{i}}
			}
		}
		s.t.Errorf("wanted asset %d is not among the options of %q", w.ID, d.Prompt)
		return Answer{Timeout: true}
	}
}

// tboard is one entity of each kind for each player (besides the Headquarters).
type tboard struct{ u0, s0, c0, u1, s1, c1 *Asset }

func (s *scene) targetBoard() tboard {
	return tboard{
		u0: s.put(0, ZoneMainTheater, unit("Ally Unit", 1, 5, Keywords{})),
		s0: s.put(0, ZoneMainTheater, structure("Ally Structure", 0, 5, Keywords{})),
		c0: s.putScratch(0),
		u1: s.put(1, ZoneMainTheater, unit("Enemy Unit", 1, 5, Keywords{})),
		s1: s.put(1, ZoneMainTheater, structure("Enemy Structure", 0, 5, Keywords{})),
		c1: s.putScratch(1),
	}
}

// deployTargeter deploys a free Reserves unit with one deploy trigger of the
// given targets and steps, and returns the new asset.
func (s *scene) deployTargeter(p PlayerID, targets []ability.Target, steps ...ability.Step) *Asset {
	s.t.Helper()
	c := trigUnit("Source", targetTrigger(targets, steps...))
	s.deployFromHand(p, c)
	return assetOf(s.g, c)
}

// --- FR-TGT-001: conjunction filtering ---

func TestFR_TGT_001_OptionsAreExactlyTheEntitiesMeetingEveryCriterion(t *testing.T) {
	const (
		allied   = ability.AllegianceAllied
		opposing = ability.AllegianceOpposing
		anyone   = ability.AllegianceAny
		unitK    = ability.KindUnit
		structK  = ability.KindStructure
		hqK      = ability.KindHeadquarters
	)
	type want func(b tboard, src *Asset, s *scene) []*Asset
	cases := []struct {
		name       string
		a          ability.Allegiance
		k          ability.Kinds
		nonScratch bool
		want       want
	}{
		{"allied unit", allied, unitK, false, func(b tboard, src *Asset, s *scene) []*Asset { return []*Asset{b.u0, b.c0, src} }},
		{"opposing unit", opposing, unitK, false, func(b tboard, src *Asset, s *scene) []*Asset { return []*Asset{b.u1, b.c1} }},
		{"allied non-Scratch unit", allied, unitK, true, func(b tboard, src *Asset, s *scene) []*Asset { return []*Asset{b.u0, src} }},
		{"opposing non-Scratch unit", opposing, unitK, true, func(b tboard, src *Asset, s *scene) []*Asset { return []*Asset{b.u1} }},
		{"allied structure", allied, structK, false, func(b tboard, src *Asset, s *scene) []*Asset { return []*Asset{b.s0} }},
		{"opposing structure", opposing, structK, false, func(b tboard, src *Asset, s *scene) []*Asset { return []*Asset{b.s1} }},
		{"allied Headquarters", allied, hqK, false, func(b tboard, src *Asset, s *scene) []*Asset { return []*Asset{s.g.Players[0].HQ} }},
		{"opposing Headquarters", opposing, hqK, false, func(b tboard, src *Asset, s *scene) []*Asset { return []*Asset{s.g.Players[1].HQ} }},
		{"opposing structure or Headquarters", opposing, structK | hqK, false, func(b tboard, src *Asset, s *scene) []*Asset {
			return []*Asset{b.s1, s.g.Players[1].HQ}
		}},
		{"allied structure or Headquarters", allied, structK | hqK, false, func(b tboard, src *Asset, s *scene) []*Asset {
			return []*Asset{b.s0, s.g.Players[0].HQ}
		}},
		{"opposing asset (no kind set)", opposing, 0, false, func(b tboard, src *Asset, s *scene) []*Asset {
			return []*Asset{b.u1, b.c1, b.s1, s.g.Players[1].HQ}
		}},
		{"opposing non-Scratch asset", opposing, 0, true, func(b tboard, src *Asset, s *scene) []*Asset {
			return []*Asset{b.u1, b.s1, s.g.Players[1].HQ}
		}},
		{"any unit", anyone, unitK, false, func(b tboard, src *Asset, s *scene) []*Asset {
			return []*Asset{b.u0, b.c0, b.u1, b.c1, src}
		}},
		{"any non-Scratch unit", anyone, unitK, true, func(b tboard, src *Asset, s *scene) []*Asset {
			return []*Asset{b.u0, b.u1, src}
		}},
		{"any structure or Headquarters", anyone, structK | hqK, false, func(b tboard, src *Asset, s *scene) []*Asset {
			return []*Asset{b.s0, b.s1, s.g.Players[0].HQ, s.g.Players[1].HQ}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newScene(t)
			b := s.targetBoard()
			single := len(tc.want(b, nil, s)) == 1
			if single {
				s.script(0)
			} else {
				s.script(0, Answer{Picks: []int{0}})
			}
			src := s.deployTargeter(0, []ability.Target{tgt(tc.a, tc.k, tc.nonScratch)}, dmgTo(1, 0))
			want := tc.want(b, src, s)
			if single {
				if n := len(s.asked(0)); n != 0 {
					t.Fatalf("%s: one legal target is chosen without a question (FR-MOD-001): asked %d times", tc.name, n)
				}
				s.wantDamage(want[0], 1)
				if got := totalDamage(s.g); got != 1 {
					t.Errorf("%s: exactly the sole legal target takes the damage: total damage %d, want 1", tc.name, got)
				}
				return
			}
			asked := s.asked(0)
			if len(asked) != 1 {
				t.Fatalf("%s: one chosen target is one question to the trigger's controller: asked %d times", tc.name, len(asked))
			}
			d := asked[0]
			if d.Kind != DecideTarget || d.Player != 0 || d.Min != 1 || d.Max != 1 {
				t.Errorf("%s: the question is a mandatory single DecideTarget for player 0: %+v", tc.name, d)
			}
			if got, w := refs(d), ids(want...); !reflect.DeepEqual(got, w) {
				t.Errorf("%s: options are exactly the entities meeting every criterion (FR-TGT-001): got %v, want %v", tc.name, got, w)
			}
		})
	}
}

func TestFR_TGT_001_AllegianceIsRelativeToTheTriggersController(t *testing.T) {
	s := newScene(t)
	s.endTurn() // player 1 is active
	u0 := s.put(0, ZoneMainTheater, unit("P0 Unit A", 1, 5, Keywords{}))
	u0b := s.put(0, ZoneMainTheater, unit("P0 Unit B", 1, 5, Keywords{}))
	s.put(1, ZoneMainTheater, unit("P1 Unit", 1, 5, Keywords{}))
	s.script(1, Answer{Picks: []int{0}})
	s.script(0)
	s.deployTargeter(1, []ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, true)}, dmgTo(1, 0))
	asked := s.asked(1)
	if len(asked) != 1 || asked[0].Player != 1 {
		t.Fatalf("the deploying player (1) chooses: %+v", asked)
	}
	if got, w := refs(asked[0]), ids(u0, u0b); !reflect.DeepEqual(got, w) {
		t.Errorf("for player 1, opposing units are player 0's: got %v, want %v (FR-TGT-001)", got, w)
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("the other player is not asked: %d", n)
	}
}

// --- FR-KWD-002h: Ward ---

func TestFR_KWD_002h_OpposingWardUnitIsNotAnOption(t *testing.T) {
	s := newScene(t)
	plain := s.put(1, ZoneMainTheater, unit("Plain", 1, 5, Keywords{}))
	plain2 := s.put(1, ZoneMainTheater, unit("Plain 2", 1, 5, Keywords{}))
	ward := s.put(1, ZoneMainTheater, unit("Warded", 1, 5, Keywords{Ward: true}))
	s.script(0, Answer{Picks: []int{0}})
	s.deployTargeter(0, []ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(1, 0))
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("two non-Warded options are asked once: %d", len(asked))
	}
	if got, w := refs(asked[0]), ids(plain, plain2); !reflect.DeepEqual(got, w) {
		t.Errorf("an opposing Ward unit cannot be chosen by a triggered ability (FR-KWD-002h): got %v, want %v", got, w)
	}
	s.wantDamage(ward, 0)
}

func TestFR_KWD_002h_AlliedWardUnitIsAnOption(t *testing.T) {
	s := newScene(t)
	ward := s.put(0, ZoneMainTheater, unit("Warded Ally", 1, 5, Keywords{Ward: true}))
	plain := s.put(0, ZoneMainTheater, unit("Plain Ally", 1, 5, Keywords{}))
	s.script(0, Answer{Picks: []int{0}})
	src := s.deployTargeter(0, []ability.Target{tgt(ability.AllegianceAllied, ability.KindUnit, true)}, dmgTo(1, 0))
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("three allied options are asked once: %d", len(asked))
	}
	if got, w := refs(asked[0]), ids(ward, plain, src); !reflect.DeepEqual(got, w) {
		t.Errorf("Ward restricts only opponents, so the controller's own Ward unit is an option (FR-KWD-002h): got %v, want %v", got, w)
	}
}

func TestFR_KWD_002h_OnlyWardedTargetMeansNoLegalTarget(t *testing.T) {
	s := newScene(t)
	ward := s.put(1, ZoneMainTheater, unit("Warded", 1, 5, Keywords{Ward: true}))
	s.script(0)
	src := s.deployTargeter(0, []ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(2, 0))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("no legal option means no question: asked %d times", n)
	}
	s.wantDamage(ward, 0)
	if !s.alive(src) {
		t.Errorf("the source still deploys when its trigger has no legal target (FR-TGT-003)")
	}
}

// --- FR-MOD-001 fast path and fallback, applied to targets ---

func TestFR_TGT_001_SingleLegalTargetIsNotAskedAndTakesTheDamage(t *testing.T) {
	s := newScene(t)
	victim := s.put(1, ZoneMainTheater, unit("Victim", 1, 5, Keywords{}))
	s.script(0)
	s.deployTargeter(0, []ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(2, 0))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a single legal option is chosen without a question (FR-MOD-001 sec. 1): asked %d times", n)
	}
	s.wantDamage(victim, 2)
}

func TestFR_TGT_001_TimeoutOnTwoOptionsDamagesExactlyOneOfThem(t *testing.T) {
	s := newScene(t)
	a := s.put(1, ZoneMainTheater, unit("Victim A", 1, 5, Keywords{}))
	b := s.put(1, ZoneMainTheater, unit("Victim B", 1, 5, Keywords{}))
	s.script(0, Answer{Timeout: true})
	s.deployTargeter(0, []ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(2, 0))
	if n := len(s.asked(0)); n != 1 {
		t.Fatalf("two options are asked once: %d", n)
	}
	if (a.Damage == 2) == (b.Damage == 2) || a.Damage+b.Damage != 2 {
		t.Errorf("a timeout picks the first legal target in fixed order (FR-MOD-001 sec. 6, SVY-R16): damage A=%d B=%d, want exactly one with 2", a.Damage, b.Damage)
	}
}

// --- FR-REP-004: non-combat damage ---

func TestFR_KWD_002b_ResistantReducesTargetedDamage(t *testing.T) {
	s := newScene(t)
	v := s.put(1, ZoneMainTheater, unit("Resistant", 1, 5, Keywords{Resistant: 1}))
	s.script(0)
	s.deployTargeter(0, []ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(3, 0))
	s.wantDamage(v, 2)
}

func TestFR_KWD_002k_ArmorAbsorbsTargetedDamageOneForOne(t *testing.T) {
	s := newScene(t)
	v := s.put(1, ZoneMainTheater, unit("Armored", 1, 5, Keywords{Armor: 2}))
	s.script(0)
	s.deployTargeter(0, []ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(3, 0))
	s.wantDamage(v, 1)
	if v.Armor != 0 {
		t.Errorf("Armor 2 absorbs 2 of 3 damage and is consumed: Armor = %d, want 0 (FR-KWD-002k)", v.Armor)
	}
}

func TestFR_KWD_002c_StalwartDoesNotReduceNonCombatDamage(t *testing.T) {
	s := newScene(t)
	v := s.put(1, ZoneMainTheater, unit("Stalwart", 1, 5, Keywords{Stalwart: true}))
	s.script(0)
	s.deployTargeter(0, []ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(3, 0))
	s.wantDamage(v, 3)
}

func TestFR_TGT_001_LethalDamageKillsAtThatItemsSweep(t *testing.T) {
	s := newScene(t)
	v := s.put(1, ZoneMainTheater, unit("Doomed", 1, 1, Keywords{}))
	inst := v.Inst
	s.script(0)
	s.deployTargeter(0, []ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(1, 0))
	if s.alive(v) || !inZone(s.g.Players[1].Discard, inst) {
		t.Errorf("a unit taking lethal damage from the ability is destroyed into its owner's Discard (FR-ZON-002)")
	}
}

func TestFR_TGT_001_DamageToHeadquartersToZeroLosesTheMatchAtThatSweep(t *testing.T) {
	s := newScene(t)
	p0 := s.g.Players[0]
	s.g.Players[1].HQ.Damage = 19
	cards := s.onTop(0, vanilla("A")...)
	s.script(0, Answer{Picks: []int{0, 1}})
	// Printed 1 (the damage) resolves first; printed 0 (Scratch) is beneath it.
	c := trigUnit("Source",
		deployTrigger(scratchN(1)),
		targetTrigger([]ability.Target{tgt(ability.AllegianceOpposing, ability.KindHeadquarters, false)}, dmgTo(1, 0)))
	s.deployFromHand(0, c)
	if !s.g.Outcome.Over || s.g.Outcome.Draw || s.g.Outcome.Winner != 0 {
		t.Errorf("an opposing HQ at 0 Current Health loses at the sweep after the item (FR-LAY-002, FR-CBT-003): outcome %+v", s.g.Outcome)
	}
	if scratched(p0, cards[0]) {
		t.Errorf("the match ended at the damage trigger's sweep, so the trigger beneath it never resolved")
	}
}

// --- two targets in one effect ---

func TestFR_TGT_001_TwoSlotsAskInSlotOrderAndEachStepHitsItsOwnSlot(t *testing.T) {
	s := newScene(t)
	u1a := s.put(1, ZoneMainTheater, unit("Enemy A", 1, 5, Keywords{}))
	u1b := s.put(1, ZoneMainTheater, unit("Enemy B", 1, 5, Keywords{}))
	s0a := s.put(0, ZoneMainTheater, structure("Ally Struct A", 0, 5, Keywords{}))
	s0b := s.put(0, ZoneMainTheater, structure("Ally Struct B", 0, 5, Keywords{}))
	s.script(0, Answer{Picks: []int{1}}, Answer{Picks: []int{0}})
	// The steps are printed in the reverse of slot order.
	s.deployTargeter(0, []ability.Target{
		tgt(ability.AllegianceOpposing, ability.KindUnit, true),
		tgt(ability.AllegianceAllied, ability.KindStructure, false),
	}, dmgTo(1, 1), dmgTo(2, 0))
	asked := s.asked(0)
	if len(asked) != 2 {
		t.Fatalf("two chosen targets are two questions: asked %d times", len(asked))
	}
	if got, w := refs(asked[0]), ids(u1a, u1b); !reflect.DeepEqual(got, w) {
		t.Errorf("the first question is slot 0 (opposing non-Scratch unit): got %v, want %v", got, w)
	}
	if got, w := refs(asked[1]), ids(s0a, s0b); !reflect.DeepEqual(got, w) {
		t.Errorf("the second question is slot 1 (allied structure): got %v, want %v", got, w)
	}
	hit0 := s.g.Asset(asked[0].Options[1].Ref)
	hit1 := s.g.Asset(asked[1].Options[0].Ref)
	s.wantDamage(hit0, 2)
	s.wantDamage(hit1, 1)
	if got := totalDamage(s.g); got != 3 {
		t.Errorf("only the two chosen entities take damage: total %d, want 3", got)
	}
}

// --- FR-TGT-003: no legal target ---

func TestFR_TGT_003_NoLegalTargetAsksNothingDamagesNothingAndDeploys(t *testing.T) {
	s := newScene(t)
	cards := s.onTop(0, vanilla("A")...)
	s.script(0)
	c := trigUnit("Source", targetTrigger([]ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(2, 0), drawN(1)))
	s.deployFromHand(0, c)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("no legal target means no question: asked %d times", n)
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("the targeted clause does nothing: total damage %d (FR-TGT-003)", got)
	}
	if !inHand(s.g.Players[0], cards[0]) {
		t.Errorf("the non-targeted Draw clause of the same effect still resolves (FR-TGT-003)")
	}
	if assetOf(s.g, c) == nil {
		t.Errorf("the deployed unit stays on the battlefield (FR-TGT-003)")
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty after resolution: %d items", len(s.g.Stack))
	}
}

// --- FR-TGT-003: second verification at resolution ---

func TestFR_TGT_003_TargetGoneAtResolutionTakesNoDamageButOtherStepsRun(t *testing.T) {
	s := newScene(t)
	p0 := s.g.Players[0]
	victim := s.put(1, ZoneMainTheater, unit("Victim", 1, 1, Keywords{}))
	inst := victim.Inst
	hand := len(p0.Hand)
	s.onTop(0, vanilla("A", "B")...)
	s.script(0, Answer{Picks: []int{0, 1}})
	hit := targetTrigger([]ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(1, 0), drawN(1))
	s.deployFromHand(0, trigUnit("Source", hit, hit))
	if n := len(s.asked(0)); n != 1 {
		t.Errorf("only the order question is asked, both targets being forced: %d questions", n)
	}
	if s.alive(victim) || !inZone(s.g.Players[1].Discard, inst) {
		t.Fatalf("the first trigger to resolve kills the 1-health unit at its sweep")
	}
	if got := len(p0.Hand) - hand; got != 2 {
		t.Errorf("the second trigger's damage step does nothing but its Draw still happens: drew %d, want 2 (FR-TGT-003)", got)
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("nothing else takes the second trigger's damage: total %d", got)
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("both triggers resolved: %d items left", len(s.g.Stack))
	}
}

func TestFR_TGT_003_PartialInvalidationResolvesForTheRemainingLegalTarget(t *testing.T) {
	s := newScene(t)
	p0 := s.g.Players[0]
	victim := s.put(1, ZoneMainTheater, unit("Victim", 1, 1, Keywords{}))
	wall := s.put(1, ZoneMainTheater, structure("Wall", 0, 5, Keywords{}))
	hand := len(p0.Hand)
	s.onTop(0, vanilla("A")...)
	s.script(0, Answer{Picks: []int{0, 1}})
	// Printed 0 is the two-slot composite; printed 1 resolves first and kills
	// the unit chosen for the composite's slot 0.
	composite := targetTrigger([]ability.Target{
		tgt(ability.AllegianceOpposing, ability.KindUnit, true),
		tgt(ability.AllegianceOpposing, ability.KindStructure, false),
	}, dmgTo(1, 0), dmgTo(2, 1), drawN(1))
	killer := targetTrigger([]ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, true)}, dmgTo(1, 0))
	s.deployFromHand(0, trigUnit("Source", composite, killer))
	if s.alive(victim) {
		t.Fatalf("precondition: the killer trigger destroyed the unit before the composite resolved")
	}
	s.wantDamage(wall, 2)
	if got := len(p0.Hand) - hand; got != 1 {
		t.Errorf("the composite's Draw still happens: drew %d, want 1", got)
	}
}

// --- FR-TRG-004: deep cascade ---

func TestFR_TRG_004_DeathTriggerFromTargetedDamageResolvesBeforeTheNextStackItem(t *testing.T) {
	s := newScene(t)
	p0 := s.g.Players[0]
	victim := s.put(0, ZoneMainTheater, mortal("Victim", 1, 1, killedTrigger(drawN(1))))
	cards := s.onTop(0, vanilla("A", "B")...)
	// Printed [Y, X]; answer [0 1] places Y first and X above it, so X resolves
	// first. X kills the allied victim, whose death trigger Draws; Y Scratches
	// (a Scratch lands apart from the hand, so the two takes are told apart).
	y := deployTrigger(scratchN(1))
	x := targetTrigger([]ability.Target{tgt(ability.AllegianceAllied, ability.KindUnit, true)}, dmgTo(1, 0))
	s.scriptRefs(0, []int{0, 1}, victim)
	s.deployFromHand(0, trigUnit("Source", y, x))
	if s.alive(victim) {
		t.Fatalf("precondition: X's lethal damage destroyed the allied unit")
	}
	if !inHand(p0, cards[0]) {
		t.Errorf("the death trigger resolves completely before Y, so its Draw takes the top card (FR-TRG-004)")
	}
	if !scratched(p0, cards[1]) {
		t.Errorf("Y resolves after the death trigger, so its Scratch takes the second card (FR-TRG-004)")
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty after the cascade: %d items", len(s.g.Stack))
	}
}
