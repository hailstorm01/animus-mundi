package engine

import (
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0225: a triggered ability that works from its card's discard
// pile (survey G1, Trigger.ActiveZone), here "When this card is put into your
// discard pile from your hand, ..." (FR-ZON-004 Enter-Discard Triggers and
// Self-Relocating Trigger Tracking, FR-TRG-002, FR-TRG-003, FR-MOD-001; SVY-R7,
// SVY-R9; board directive #136). Cards are synthetic. Player 0 is the active
// player and enacts a Maneuver that discards from a hand holding only the
// card(s) under test, so the discard itself is not asked (SVY-R16).

// pileTrigger is one "When this card is put into your discard pile from your
// hand, <steps>" ability.
func pileTrigger(optional bool, steps ...ability.Step) ability.Trigger {
	return ability.Trigger{Hook: ability.HookDiscardedFromHand, ActiveZone: ability.ZoneDiscard,
		Optional: optional, Effect: ability.Effect{Steps: steps}}
}

func deploySelf() ability.Step {
	return ability.Step{Op: ability.OpDeployFree, Subject: ability.SubjectSelf}
}

// remnant is a unit costing 1 that prints only Reserves and the optional
// "deploy it" pile trigger.
func remnant(name string) *Card {
	c := unit(name, 1, 1, Keywords{})
	c.ReqCost = 1
	c.DeployZones = []Zone{ZoneReserves}
	c.Ability = &ability.Def{Triggers: []ability.Trigger{pileTrigger(true, deploySelf())}}
	return c
}

// byKind installs a chooser for seat p that records every decision and
// answers with fn.
func (s *scene) byKind(p PlayerID, fn func(d Decision) Answer) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[p] = nil
	s.g.choosers[p] = func(_ *Game, d Decision) Answer {
		s.asks[p] = append(s.asks[p], d)
		return fn(d)
	}
}

// yesAndTimeouts accepts every optional question and lets everything else
// (an order, a hand pick, a target) time out.
func yesAndTimeouts(d Decision) Answer {
	if d.Kind == DecideOptional {
		return Answer{Picks: []int{0}}
	}
	return Answer{Timeout: true}
}

func askedKind(s *scene, p PlayerID, k DecisionKind) []Decision {
	var out []Decision
	for _, d := range s.asked(p) {
		if d.Kind == k {
			out = append(out, d)
		}
	}
	return out
}

// onField returns the battlefield asset of inst, or nil.
func onField(g *Game, inst *Instance) *Asset {
	for _, p := range g.Players {
		for _, list := range [][]*Asset{{p.HQ}, p.Reserves, p.MainTheater} {
			for _, a := range list {
				if a != nil && a.Inst == inst {
					return a
				}
			}
		}
	}
	return nil
}

// placeSeq is the Seq of the first place event naming card at stack depth
// depth (0: any depth), or 0.
func placeSeq(s *scene, card string, depth int) int {
	for _, e := range s.g.Log {
		if e.Kind == "place" && e.Card == card && (depth == 0 || e.Amount == depth) {
			return e.Seq
		}
	}
	return 0
}

func pileLogPlaces(s *scene, card string) int {
	n := 0
	for _, e := range s.g.Log {
		if e.Kind == "place" && e.Card == card {
			n++
		}
	}
	return n
}

// enactWith gives player 0 an empty hand, then the given cards in order, then
// a costless Maneuver with no targets whose steps are st, and enacts it. It
// returns the instances of the given cards.
func (s *scene) enactWith(hand []*Card, st ...ability.Step) []*Instance {
	s.t.Helper()
	s.g.Players[0].Hand = nil
	insts := make([]*Instance, len(hand))
	for i, c := range hand {
		insts[i] = s.give(0, c)
	}
	m := s.give(0, maneuver("Purge", 0, 0, steps(st...)))
	s.do(enactAction(m))
	return insts
}

func TestFR_ZON_004_DiscardedFromHandTriggerDeploysItself(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Picks: []int{0}})
	s.script(1)
	inst := s.enactWith([]*Card{remnant("Remnant")}, discardN(1))[0]
	p := s.g.Players[0]
	if got := askedKind(s, 0, DecideOptional); len(got) != 1 {
		t.Fatalf("player 0 is asked one optional question for Remnant's trigger, got %d (FR-ZON-004, FR-MOD-001)", len(got))
	}
	a := onField(s.g, inst)
	if a == nil || a.Controller != 0 {
		t.Fatalf("the accepted trigger deploys Remnant under player 0 (FR-ZON-004)")
	}
	if !inZone(units(p.Reserves), inst) {
		t.Errorf("Remnant enters Reserves, the only zone it prints (FR-ZON-004)")
	}
	if inZone(p.Discard, inst) {
		t.Errorf("Remnant has left the discard pile (FR-ZON-004)")
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty once the trigger resolves: %d items (FR-TRG-002)", len(s.g.Stack))
	}
	found := false
	for _, e := range s.g.Log {
		if e.Kind == "place" && e.Card == "Remnant" && e.Inst == inst.ID {
			found = true
		}
	}
	if !found {
		t.Errorf("the trigger is placed on the stack naming the card Remnant and its instance (FR-TRG-002)")
	}
}

// units lists the instances of a slice of assets.
func units(as []*Asset) []*Instance {
	out := make([]*Instance, len(as))
	for i, a := range as {
		out[i] = a.Inst
	}
	return out
}

func TestFR_MOD_001_DiscardedFromHandTriggerDeclinedLeavesTheCard(t *testing.T) {
	// Covers: FR-ZON-004; SVY-R9
	s := newScene(t)
	s.script(0, Answer{Timeout: true})
	s.script(1)
	inst := s.enactWith([]*Card{remnant("Remnant")}, discardN(1))[0]
	if got := askedKind(s, 0, DecideOptional); len(got) != 1 {
		t.Fatalf("player 0 is asked one optional question, got %d (FR-MOD-001)", len(got))
	}
	if onField(s.g, inst) != nil {
		t.Errorf("an optional trigger that times out is declined: Remnant is not deployed (FR-MOD-001 sec. 6, SVY-R9)")
	}
	if !inZone(s.g.Players[0].Discard, inst) {
		t.Errorf("Remnant stays in player 0's discard pile (FR-ZON-004)")
	}
}

func TestFR_TRG_002_DiscardCostFiresTheTriggerAboveTheItem(t *testing.T) {
	// Covers: FR-TRN-003a, FR-ZON-004
	s := newScene(t)
	s.script(0, Answer{Picks: []int{0}}, Answer{Picks: []int{0}}) // the cost's hand pick, then accept
	s.script(1)
	s.g.Players[0].Hand = nil
	rem := s.give(0, remnant("Remnant"))
	m := s.give(0, discardCostManeuver("Steel Rain", 0, 0, 1, steps(hqPing())))
	s.do(enactAction(m))
	if onField(s.g, rem) == nil {
		t.Fatalf("Remnant discarded as a cost fires its trigger and is deployed (FR-TRN-003a, FR-ZON-004)")
	}
	var deploy, dmg int
	for _, e := range s.g.Log {
		switch {
		case e.Kind == "deploy" && e.Inst == rem.ID && deploy == 0:
			deploy = e.Seq
		case e.Kind == "damage" && e.Target == s.g.Players[1].HQ.ID && e.Amount == 1 && dmg == 0:
			dmg = e.Seq
		}
	}
	if deploy == 0 || dmg == 0 {
		t.Fatalf("missing events: Remnant deploy %d, Maneuver damage %d", deploy, dmg)
	}
	if deploy > dmg {
		t.Errorf("the cost's trigger is placed above the Maneuver and resolves first: deploy seq %d, Maneuver damage seq %d (FR-TRG-002)", deploy, dmg)
	}
}

func TestFR_TRG_003_OpponentsDiscardFiresTheOwnersTrigger(t *testing.T) {
	s := newScene(t)
	s.script(0)
	s.script(1, Answer{Picks: []int{0}})
	s.g.Players[1].Hand = nil
	inst := s.give(1, remnant("Remnant"))
	s.enactWith(nil, ability.Step{Op: ability.OpDiscard, Subject: ability.SubjectOpponents, Amount: ability.Value{N: 1}})
	if len(s.asked(0)) != 0 {
		t.Errorf("player 0 is never asked: got %d decisions (FR-TRG-003)", len(s.asked(0)))
	}
	if got := askedKind(s, 1, DecideOptional); len(got) != 1 {
		t.Fatalf("the card's owner, player 1, is asked the optional question, got %d (FR-TRG-003)", len(got))
	}
	a := onField(s.g, inst)
	if a == nil || a.Controller != 1 || !inZone(units(s.g.Players[1].Reserves), inst) {
		t.Errorf("Remnant enters player 1's Reserves under player 1 (FR-ZON-004, FR-TRG-003)")
	}
}

func TestFR_TRG_003_PileTriggerRanksAfterBattlefieldTriggers(t *testing.T) {
	// Covers: FR-TRG-002; SVY-R7, board directive #136
	s := newScene(t)
	s.byKind(0, yesAndTimeouts)
	s.script(1)
	watcher := mortal("Watcher", 1, 5, ability.Trigger{Hook: ability.HookDamaged,
		Effect: ability.Effect{Steps: []ability.Step{drawN(1)}}})
	w := s.put(0, ZoneMainTheater, watcher)
	s.g.Players[0].Hand = nil
	rem := s.give(0, remnant("Remnant"))
	m := maneuver("Purge", 0, 0, ability.Effect{
		Targets: []ability.Target{{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}}},
		Steps: []ability.Step{
			{Op: ability.OpDealDamage, Subject: ability.SubjectTarget, Target: 0, Amount: ability.Value{N: 1}},
			discardN(1),
		}})
	s.do(enactAction(s.give(0, m)))
	orders := askedKind(s, 0, DecideOrder)
	if len(orders) != 1 {
		t.Fatalf("player 0 orders the two triggers with one decision, got %d (FR-TRG-003)", len(orders))
	}
	o := orders[0].Options
	if len(o) != 2 || o[0].Ref != w.ID || o[1].Ref != 0 {
		t.Fatalf("options list Watcher's trigger (its asset) first, then Remnant's (Ref 0): %+v (SVY-R7, #136)", o)
	}
	ws, rs := placeSeq(s, "Watcher", 0), placeSeq(s, "Remnant", 0)
	if ws == 0 || rs == 0 {
		t.Fatalf("both triggers are placed: Watcher %d, Remnant %d", ws, rs)
	}
	if ws > rs {
		t.Errorf("an unanswered order goes by source position, a pile ranking after the battlefield: Watcher placed first (seq %d), Remnant (seq %d) resolves first (SVY-R7, #136)", ws, rs)
	}
	if onField(s.g, rem) == nil {
		t.Errorf("Remnant, accepted, is deployed (FR-ZON-004)")
	}
}

func TestFR_TRG_003_PileTriggersFollowTheDiscardOrder(t *testing.T) {
	// Covers: SVY-R7
	s := newScene(t)
	s.byKind(0, func(d Decision) Answer {
		if d.Kind == DecideHand {
			return Answer{Picks: []int{1, 0}} // B, then A
		}
		return yesAndTimeouts(d)
	})
	s.script(1)
	s.enactWith([]*Card{remnant("Remnant A"), remnant("Remnant B"), unit("Filler", 1, 1, Keywords{})}, discardN(2))
	if got := askedKind(s, 0, DecideHand); len(got) != 1 {
		t.Fatalf("the discard of 2 from 3 cards is asked once, got %d (SVY-R16)", len(got))
	}
	orders := askedKind(s, 0, DecideOrder)
	if len(orders) != 1 || len(orders[0].Options) != 2 {
		t.Fatalf("player 0 orders the two pile triggers: %+v (FR-TRG-003)", orders)
	}
	o := orders[0].Options
	if !strings.Contains(o[0].Label, "Remnant B") || !strings.Contains(o[1].Label, "Remnant A") {
		t.Errorf("among one player's pile sources the order is the order the cards entered the pile, B then A: %q, %q (SVY-R7)", o[0].Label, o[1].Label)
	}
	b, a := placeSeq(s, "Remnant B", 0), placeSeq(s, "Remnant A", 0)
	if b == 0 || a == 0 || b > a {
		t.Errorf("by default B's trigger is placed first, A's second: B seq %d, A seq %d (SVY-R7)", b, a)
	}
}

func TestFR_ZON_004_PileTriggerFizzlesIfItsCardLeftThePile(t *testing.T) {
	// FRD 01 line 251: a trigger that moves "it" tracks the card in the pile
	// and does nothing once it has left.
	s := newScene(t)
	s.byKind(0, func(d Decision) Answer {
		if d.Kind == DecideOrder {
			return Answer{Picks: []int{1, 0}} // the first pick is placed first and resolves last
		}
		return Answer{Timeout: true}
	})
	s.script(1)
	c := unit("Wanderer", 1, 1, Keywords{})
	c.DeployZones = []Zone{ZoneReserves}
	c.Ability = &ability.Def{Triggers: []ability.Trigger{
		pileTrigger(false, returnSelf()),
		pileTrigger(false, deploySelf()),
	}}
	inst := s.enactWith([]*Card{c}, discardN(1))[0]
	if got := askedKind(s, 0, DecideOrder); len(got) != 1 || len(got[0].Options) != 2 {
		t.Fatalf("player 0 orders Wanderer's two triggers: %+v (FR-TRG-003)", got)
	}
	if !inHand(s.g.Players[0], inst) {
		t.Errorf("the return trigger resolves first and puts Wanderer in its owner's hand (FR-ZON-004)")
	}
	if onField(s.g, inst) != nil || inZone(s.g.Players[0].Discard, inst) {
		t.Errorf("the deploy trigger finds Wanderer out of the pile and does nothing (FR-ZON-004)")
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("both triggers resolved: %d items on the stack", len(s.g.Stack))
	}
}

func TestFR_ZON_004_PileTriggerMovesOnlyItsOwnInstance(t *testing.T) {
	// FR-TRG-006: tracking is by card instance, not by name.
	s := newScene(t)
	s.script(0, Answer{Picks: []int{0}})
	s.script(1)
	c := remnant("Remnant")
	other := s.inDiscard(0, c)[0]
	inst := s.enactWith([]*Card{c}, discardN(1))[0]
	if onField(s.g, inst) == nil {
		t.Fatalf("the discarded copy is deployed (FR-ZON-004)")
	}
	if onField(s.g, other) != nil || !inZone(s.g.Players[0].Discard, other) {
		t.Errorf("the other copy of the same card stays in the discard pile (FR-ZON-004, FR-TRG-006)")
	}
}

func TestFR_ZON_004_OnlyADiscardFromHandFiresIt(t *testing.T) {
	// A unit destroyed on the battlefield goes to the discard pile, but not
	// from the hand.
	s := newScene(t)
	s.script(0)
	s.script(1)
	inst := killOnP1(s, remnant("Remnant"))
	if len(s.asked(1)) != 0 || len(s.asked(0)) != 0 {
		t.Errorf("a card put into the discard pile from the battlefield fires no discarded-from-hand trigger: asked %d, %d (FR-ZON-004)", len(s.asked(0)), len(s.asked(1)))
	}
	if pileLogPlaces(s, "Remnant") != 0 {
		t.Errorf("nothing is placed for Remnant (FR-ZON-004)")
	}
	if !inZone(s.g.Players[1].Discard, inst) {
		t.Errorf("the destroyed Remnant stays in its owner's discard pile (FR-ZON-004)")
	}

	// A Maneuver card carrying the same trigger goes to the pile on resolution.
	s = newScene(t)
	s.script(0)
	s.script(1)
	mc := maneuver("Remnant Maneuver", 0, 0, steps(drawN(1)))
	mc.Ability.Triggers = []ability.Trigger{pileTrigger(true, deploySelf())}
	s.g.Players[0].Hand = nil
	mi := s.give(0, mc)
	s.do(enactAction(mi))
	if len(s.asked(0)) != 0 {
		t.Errorf("a Maneuver going to the discard pile on resolution fires nothing: asked %d (FR-ZON-004)", len(s.asked(0)))
	}
	if got := pileLogPlaces(s, "Remnant Maneuver"); got != 1 {
		t.Errorf("only the Maneuver's own placement is on the log, got %d place events (FR-ZON-004)", got)
	}
	if !inZone(s.g.Players[0].Discard, mi) {
		t.Errorf("the Maneuver card is in its owner's discard pile (FR-MNV-001)")
	}
}

func TestFR_ZON_004_ActiveZoneGatesTheSource(t *testing.T) {
	// A battlefield trigger (no ActiveZone) on the discarded-from-hand hook
	// does not fire from the discard pile.
	s := newScene(t)
	s.script(0)
	s.script(1)
	bf := unit("Battlefield Hook", 1, 1, Keywords{})
	bf.DeployZones = []Zone{ZoneReserves}
	bf.Ability = &ability.Def{Triggers: []ability.Trigger{{Hook: ability.HookDiscardedFromHand, Optional: true,
		Effect: ability.Effect{Steps: []ability.Step{deploySelf()}}}}}
	inst := s.enactWith([]*Card{bf}, discardN(1))[0]
	if len(s.asked(0)) != 0 || pileLogPlaces(s, "Battlefield Hook") != 0 {
		t.Errorf("a trigger with no ActiveZone works from the battlefield only, so a card in the discard pile fires nothing (G1)")
	}
	if onField(s.g, inst) != nil || !inZone(s.g.Players[0].Discard, inst) {
		t.Errorf("the card stays in the discard pile (G1)")
	}

	// A discard-pile trigger on another hook does not fire from the battlefield.
	s = newScene(t)
	s.script(0)
	s.script(1)
	dp := trigUnit("Pile Deploy", ability.Trigger{Hook: ability.HookDeploy, ActiveZone: ability.ZoneDiscard, Optional: true,
		Effect: ability.Effect{Steps: []ability.Step{drawN(1)}}})
	s.deployFromHand(0, dp)
	if len(s.asked(0)) != 0 || pileLogPlaces(s, "Pile Deploy") != 0 {
		t.Errorf("a discard-pile trigger does not fire for a card on the battlefield (G1)")
	}
}
