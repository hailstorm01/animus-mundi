package engine

import "testing"

// Tests for T-0226 (FR-ZON-004, FR-TRN-003a, FR-TRG-002, FR-MOD-001,
// FR-KWD-002l; SVY-R9; board directive #136): the Path unit Eerie Remnant,
// loaded from data/cards. "Deploys Exhausted. When this card is put into your
// discard pile from your hand, you may deploy it from your discard pile for
// free." Unit, Skeleton, Reaver, 1 Requisition, Reserves, 3/1.

// remnantChooser installs a chooser for seat p that records every decision.
// A hand question is answered with the pick (the first option when pick is
// nil), an optional question with opt, a target question with the option
// naming target (when set), and anything else times out.
func remnantChooser(s *scene, p PlayerID, pick *Instance, opt Answer, target *Asset) {
	s.byKind(p, func(d Decision) Answer {
		switch d.Kind {
		case DecideHand:
			if pick == nil {
				return Answer{Picks: []int{0}}
			}
			for i, o := range d.Options {
				if o.Ref == pick.ID {
					return Answer{Picks: []int{i}}
				}
			}
		case DecideOptional:
			return opt
		case DecideTarget:
			if target != nil {
				for i, o := range d.Options {
					if o.Ref == target.ID {
						return Answer{Picks: []int{i}}
					}
				}
			}
		}
		return Answer{Timeout: true}
	})
}

// remnantExhaustedInReserves reports whether inst is an Exhausted asset in p's
// Reserves, under p, and not in p's discard pile.
func remnantExhaustedInReserves(s *scene, p PlayerID, inst *Instance) bool {
	a := onField(s.g, inst)
	return a != nil && a.Controller == p && a.Zone == ZoneReserves && a.Exhausted &&
		!inZone(s.g.Players[p].Discard, inst)
}

// remnantOrdersScene gives player 0 a hand of Improvised Orders and Eerie
// Remnant, a pool of R0 T2 and three vanilla cards on top of the deck, then
// enacts the Orders, discarding Remnant and answering the optional question
// with opt. It returns the scene, the Orders, Remnant and the drawn cards.
func remnantOrdersScene(t *testing.T, opt Answer) (*scene, *Instance, *Instance, []*Instance) {
	t.Helper()
	s := newScene(t)
	g0 := s.g.Players[0]
	g0.Hand = nil
	orders := s.give(0, s.card("Improvised Orders"))
	rem := s.give(0, s.card("Eerie Remnant"))
	g0.Req, g0.Tac = 0, 2
	deck := s.onTop(0, vanilla("D1", "D2", "D3")...)
	remnantChooser(s, 0, rem, opt, nil)
	s.script(1)
	s.do(enactAction(orders))
	return s, orders, rem, deck
}

func TestCard_EerieRemnant_DeploysFromHandAsAThreeOneExhaustedForOneRequisition(t *testing.T) {
	// Covers: FR-KWD-002l
	s := newScene(t)
	s.script(0)
	s.script(1)
	c := deployPath(s, "Eerie Remnant", 1)
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves {
		t.Fatalf("Eerie Remnant is on the battlefield in Reserves after deploying")
	}
	if p := a.Printed(); p.Power != 3 || p.Health != 1 {
		t.Errorf("Eerie Remnant is a 3/1: got %d/%d", p.Power, p.Health)
	}
	if !a.Exhausted {
		t.Errorf("a card with Deploys Exhausted enters Exhausted (FR-KWD-002l)")
	}
	if p := s.g.Players[0]; p.Req != 0 {
		t.Errorf("Eerie Remnant costs 1 Requisition: pool R%d, want R0", p.Req)
	}
	if n := len(s.asked(0)) + len(s.asked(1)); n != 0 {
		t.Errorf("deploying from hand fires no discard trigger and asks nothing: asked %d times (FR-ZON-004)", n)
	}
}

func TestCard_EerieRemnant_ImprovisedOrdersDiscardDeploysItExhausted(t *testing.T) {
	// Covers: FR-ZON-004, FR-TRN-003a
	s, orders, rem, deck := remnantOrdersScene(t, yes)
	p := s.g.Players[0]
	if got := askedKind(s, 0, DecideOptional); len(got) != 1 {
		t.Fatalf("player 0 is asked one optional question for Remnant's trigger, got %d (FR-MOD-001)", len(got))
	}
	if !remnantExhaustedInReserves(s, 0, rem) {
		t.Errorf("the accepted trigger deploys Remnant into player 0's Reserves, Exhausted (FR-ZON-004, FR-KWD-002l)")
	}
	if p.Req != 0 || p.Tac != 0 {
		t.Errorf("the deploy is free: pool R%d T%d, want R0 T0 (FR-TRN-003a, FR-CST-001)", p.Req, p.Tac)
	}
	if len(p.Hand) != 3 || p.Hand[0] != deck[0] || p.Hand[1] != deck[1] || p.Hand[2] != deck[2] {
		t.Errorf("the three drawn cards are in hand, got %d cards", len(p.Hand))
	}
	if !inZone(p.Discard, orders) {
		t.Errorf("Improvised Orders is in the discard pile (FR-MNV-001)")
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty once everything resolves: %d items", len(s.g.Stack))
	}
}

func TestCard_EerieRemnant_SteelRainCostDiscardFiresAboveTheManeuver(t *testing.T) {
	// Covers: FR-TRN-003a, FR-TRG-002
	s := newScene(t)
	p := s.g.Players[0]
	p.Hand = nil
	rain := s.give(0, s.card("Steel Rain"))
	rem := s.give(0, s.card("Eerie Remnant"))
	s.give(0, unit("Other", 1, 1, Keywords{}))
	p.Req, p.Tac = 0, 4
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 9, Keywords{}))
	remnantChooser(s, 0, rem, yes, foe)
	s.script(1)
	s.do(enactAction(rain))
	if !remnantExhaustedInReserves(s, 0, rem) {
		t.Fatalf("Remnant discarded to Steel Rain's cost fires its trigger and is deployed Exhausted (FR-TRN-003a, FR-ZON-004)")
	}
	s.wantDamage(foe, 2)
	var deploy, dmg int
	for _, e := range s.g.Log {
		switch {
		case e.Kind == "deploy" && e.Inst == rem.ID && deploy == 0:
			deploy = e.Seq
		case e.Kind == EventDamage && e.Target == foe.ID && dmg == 0:
			dmg = e.Seq
		}
	}
	if deploy == 0 || dmg == 0 {
		t.Fatalf("missing events: Remnant deploy %d, Steel Rain damage %d", deploy, dmg)
	}
	if deploy > dmg {
		t.Errorf("the cost's trigger is placed above Steel Rain and resolves first: deploy seq %d, damage seq %d (FR-TRG-002)", deploy, dmg)
	}
}

func TestCard_EerieRemnant_IrregularsUnbowedDiscardDeploysIt(t *testing.T) {
	// Covers: FR-ZON-004, FR-TRG-003
	s := newScene(t)
	p1 := s.g.Players[1]
	p1.Hand = nil
	rem := s.give(1, s.card("Eerie Remnant"))
	s.onTop(1, vanilla("D")...)
	def := s.put(1, ZoneMainTheater, s.card("Chainbreaker Irregulars"))
	att := s.put(0, ZoneMainTheater, unit("Hitter", 1, 9, Keywords{}))
	s.script(0)
	remnantChooser(s, 1, rem, yes, nil)
	s.do(s.attack(att, def))
	if !s.alive(def) {
		t.Fatalf("precondition: the Irregulars survive 1 damage")
	}
	if got := askedKind(s, 1, DecideOptional); len(got) != 1 {
		t.Fatalf("player 1, the card's owner, is asked the optional question, got %d (FR-TRG-003)", len(got))
	}
	if !remnantExhaustedInReserves(s, 1, rem) {
		t.Errorf("Remnant discarded by Unbowed is deployed into player 1's Reserves, Exhausted, under player 1 (FR-ZON-004)")
	}
}

func TestCard_EerieRemnant_DeclinedStaysInThePile(t *testing.T) {
	// Covers: FR-MOD-001
	s, _, rem, _ := remnantOrdersScene(t, no)
	if got := askedKind(s, 0, DecideOptional); len(got) != 1 {
		t.Fatalf("player 0 is asked one optional question, got %d (FR-MOD-001)", len(got))
	}
	if onField(s.g, rem) != nil {
		t.Errorf("a declined trigger deploys nothing: Remnant is not an asset (FR-MOD-001 sec. 7)")
	}
	if !inZone(s.g.Players[0].Discard, rem) {
		t.Errorf("Remnant stays in player 0's discard pile (FR-ZON-004)")
	}
}

func TestCard_EerieRemnant_TimeoutStaysInThePile(t *testing.T) {
	// Covers: FR-MOD-001, SVY-R9
	s, _, rem, _ := remnantOrdersScene(t, Answer{Timeout: true})
	if got := askedKind(s, 0, DecideOptional); len(got) != 1 {
		t.Fatalf("player 0 is asked one optional question, got %d (FR-MOD-001)", len(got))
	}
	if onField(s.g, rem) != nil {
		t.Errorf("an optional trigger that times out is declined: Remnant is not an asset (FR-MOD-001 sec. 6, SVY-R9)")
	}
	if !inZone(s.g.Players[0].Discard, rem) {
		t.Errorf("Remnant stays in player 0's discard pile (FR-ZON-004)")
	}
}

func TestCard_EerieRemnant_KilledOnTheBattlefieldFiresNothing(t *testing.T) {
	// Covers: FR-ZON-004
	s := newScene(t)
	rem := s.put(0, ZoneMainTheater, s.card("Eerie Remnant"))
	foe := s.put(1, ZoneMainTheater, unit("Brute", 5, 9, Keywords{}))
	inst := rem.Inst
	s.script(0)
	s.script(1)
	s.do(s.attack(rem, foe))
	if s.alive(rem) {
		t.Fatalf("precondition: the 3/1 Remnant is destroyed by the 5-power defender")
	}
	if n := len(askedKind(s, 0, DecideOptional)) + len(askedKind(s, 1, DecideOptional)); n != 0 {
		t.Errorf("a card put into the discard pile from the battlefield, not the hand, fires no trigger: %d optional questions (FR-ZON-004)", n)
	}
	if !inZone(s.g.Players[0].Discard, inst) {
		t.Errorf("the killed Remnant is in player 0's discard pile (FR-ZON-004)")
	}
	if onField(s.g, inst) != nil {
		t.Errorf("the killed Remnant is not an asset afterwards (FR-ZON-004)")
	}
}
