package engine

import (
	"slices"
	"testing"
)

// Card tests for Allshan's Battle Trance (T-0220), written from the CSV text:
// Path Maneuver, R0 T4, "Deploy chosen unit card from your discard pile to the
// Main Theater. Until end of turn, it gains Stalwart and gets +2 power."
// Player 0 enacts. Synthetic pile units are Reserves-only (rrUnit), so the
// explicit zone is the only thing that puts them in the Main Theater.

// btScene gives player 0 Battle Trance with exactly R0 T4.
func btScene(t *testing.T) (*scene, *Instance) {
	t.Helper()
	s := newScene(t)
	inst := s.give(0, s.card("Allshan's Battle Trance"))
	s.setPool(0, 0, 4)
	return s, inst
}

func btZoneAsks(s *scene) int {
	n := 0
	for _, d := range s.asked(0) {
		if d.Kind == DecideZone {
			n++
		}
	}
	return n
}

func TestCard_AllshansBattleTrance_DeploysToTheMainTheater(t *testing.T) {
	// Covers: FR-TRN-003a, FR-CST-001, FR-LAY-001, FR-LAY-003
	s, bt := btScene(t)
	c := rrUnit("Costly", 3)
	c.ReqCost, c.TacCost = 2, 1
	pile := s.inDiscard(0, c)[0]
	s.script(0)
	s.do(enactAction(bt))
	p := s.g.Players[0]
	a := assetOf(s.g, c)
	if a == nil {
		t.Fatalf("the chosen unit card is on the battlefield after resolution (FR-TRN-003a)")
	}
	if a.Controller != 0 || a.Zone != ZoneMainTheater {
		t.Errorf("it enters the Main Theater under player 0's control: controller %d zone %v (FR-TRN-003a 2)", a.Controller, a.Zone)
	}
	if inDiscardPile(p, pile) {
		t.Errorf("the deployed card left the discard pile (FR-TRN-003a)")
	}
	if n := btZoneAsks(s); n != 0 {
		t.Errorf("an explicit zone is asked of nobody: %d DecideZone (FR-TRN-003a 2)", n)
	}
	if p.Req != 0 || p.Tac != 0 {
		t.Errorf("only the Maneuver's own T4 is paid, the deploy is free: pool R%d T%d, want R0 T0 (FR-CST-001)", p.Req, p.Tac)
	}
	if !s.g.Keywords(a).Stalwart {
		t.Errorf("the deployed unit has Stalwart until end of turn (FR-LAY-001)")
	}
	if got := s.g.Power(a); got != 5 {
		t.Errorf("power %d, want printed 3 plus 2 (FR-LAY-001)", got)
	}
	if a.Printed().Keywords.Stalwart {
		t.Errorf("the printed keywords are unchanged (FR-LAY-001)")
	}
	if !inDiscardPile(p, bt) {
		t.Errorf("the resolved Allshan's Battle Trance is in its controller's discard pile")
	}
}

func TestCard_AllshansBattleTrance_ExpiresInSubPhase4c(t *testing.T) {
	// Covers: FR-LAY-003, FR-TRN-001
	s, bt := btScene(t)
	c := rrUnit("Rider", 3)
	s.inDiscard(0, c)
	s.script(0)
	s.do(enactAction(bt))
	s.endTurn()
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneMainTheater {
		t.Fatalf("the unit stays in the Main Theater after the turn ends: %v", a)
	}
	if s.g.Keywords(a).Stalwart {
		t.Errorf("Stalwart expires with the turn, one duration for both effects (FR-LAY-003, FR-TRN-001 4c)")
	}
	if got := s.g.Power(a); got != 3 {
		t.Errorf("power %d after the turn ends, want printed 3 (FR-LAY-003, FR-TRN-001 4c)", got)
	}
}

func TestCard_AllshansBattleTrance_StalwartTakesOneCombatDamage(t *testing.T) {
	// Covers: FR-KWD-002c
	s, bt := btScene(t)
	c := unit("Rider", 2, 9, Keywords{})
	c.DeployZones = []Zone{ZoneReserves}
	s.inDiscard(0, c)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 3, 20, Keywords{}))
	s.script(0)
	s.do(enactAction(bt))
	a := assetOf(s.g, c)
	if a == nil {
		t.Fatalf("the unit is deployed")
	}
	s.do(s.attack(a, foe))
	s.wantDamage(a, 1) // 3 combat damage capped to 1 by Stalwart (FR-KWD-002c)
}

func TestCard_AllshansBattleTrance_GuerrillasDeployTriggerFires(t *testing.T) {
	// Covers: FR-TRN-003a, FR-LAY-001
	s, bt := btScene(t)
	g := s.card("Devilhorn Guerrillas")
	s.inDiscard(0, g)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 9, Keywords{}))
	s.script(0)
	s.do(enactAction(bt))
	if n := btZoneAsks(s); n != 0 {
		t.Errorf("the explicit zone overrides the printed ones, no DecideZone: %d (FR-TRN-003a 2)", n)
	}
	a := assetOf(s.g, g)
	if a == nil || a.Zone != ZoneMainTheater {
		t.Fatalf("the Guerrillas enter the Main Theater (FR-TRN-003a 2): %v", a)
	}
	s.wantDamage(foe, 2)
	if got := s.g.Power(a); got != 5 {
		t.Errorf("power %d, want printed 3 plus 2 (FR-LAY-001)", got)
	}
	if !s.g.Keywords(a).Stalwart {
		t.Errorf("the Guerrillas have Stalwart (FR-LAY-001)")
	}
}

func TestCard_AllshansBattleTrance_VeilRenderPaysNoKill(t *testing.T) {
	// Covers: FR-TRN-003a, SVY-R46
	s, bt := btScene(t)
	vr := s.card("Veil Render")
	s.inDiscard(0, vr)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 3, Keywords{}))
	s.script(0)
	s.do(enactAction(bt))
	a := assetOf(s.g, vr)
	if a == nil || a.Zone != ZoneMainTheater {
		t.Fatalf("Veil Render enters the Main Theater (FR-TRN-003a 2): %v", a)
	}
	if !s.alive(ally) {
		t.Errorf("the allied unit is alive, Veil Render's kill cost is not paid by another permission (SVY-R46)")
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("no decision is asked of player 0: %v (FR-TRN-003a, SVY-R46)", kindsOf(s.asked(0)))
	}
}

func TestCard_AllshansBattleTrance_NotEnactableWithoutAUnitCard(t *testing.T) {
	// Covers: SVY-R29, FR-MNV-001
	notOffered := func(t *testing.T, s *scene, bt *Instance, why string) {
		t.Helper()
		if offersEnact(s.g, bt) {
			t.Errorf("Enact is not offered with %s (FR-MNV-001)", why)
		}
		s.reject(enactAction(bt))
		if !slices.Contains(s.g.Players[0].Hand, bt) {
			t.Errorf("a rejected Enact leaves Allshan's Battle Trance in hand (FR-MNV-001)")
		}
	}
	t.Run("empty pile", func(t *testing.T) {
		s, bt := btScene(t)
		notOffered(t, s, bt, "an empty discard pile")
	})
	t.Run("only a Maneuver and a Structure card", func(t *testing.T) {
		s, bt := btScene(t)
		s.inDiscard(0, maneuver("Spell", 0, 0, steps()), structure("Wall1", 1, 3, Keywords{}))
		notOffered(t, s, bt, "only a Maneuver card and a Structure card in the pile")
	})
	t.Run("only the opponent's pile", func(t *testing.T) {
		s, bt := btScene(t)
		s.inDiscard(1, rrUnit("Theirs1", 1))
		notOffered(t, s, bt, "a unit card only in the opposing pile")
	})
	t.Run("offered with a power-5 unit card", func(t *testing.T) {
		s, bt := btScene(t)
		s.inDiscard(0, rrUnit("P5", 5))
		if !offersEnact(s.g, bt) {
			t.Errorf("Enact is offered with a power-5 unit card in the pile, no power limit (FR-MNV-001)")
		}
	})
}

func TestCard_AllshansBattleTrance_TargetLeftThePile(t *testing.T) {
	// Covers: FR-TGT-003, FR-TRN-003a
	s, bt := btScene(t)
	bystander := s.put(0, ZoneMainTheater, unit("Bystander", 2, 5, Keywords{}))
	c := rrUnit("Leaves", 3)
	inst := s.inDiscard(0, c)[0]
	s.script(0)
	moved := false
	s.g.obs = &recorder{onEvent: func(e Event) {
		if e.Kind == "place" && !moved {
			moved = true
			if s.g.takeCard(inst) {
				s.g.putCard(inst, ZoneHand, false)
			}
		}
	}}
	from := len(s.g.Log)
	s.do(enactAction(bt))
	if !moved {
		t.Fatalf("the Maneuver was placed on the stack")
	}
	if assetOf(s.g, c) != nil {
		t.Errorf("a card that left the pile is not deployed (FR-TGT-003)")
	}
	if s.g.Keywords(bystander).Stalwart || s.g.Power(bystander) != 2 {
		t.Errorf("the allied unit gained nothing: Stalwart %v, power %d (FR-TGT-003)", s.g.Keywords(bystander).Stalwart, s.g.Power(bystander))
	}
	if ev := effectEvents(s.g, from, EventKeyword); len(ev) != 0 {
		t.Errorf("no keyword event is logged, got %+v (FR-TGT-003)", ev)
	}
	if !inDiscardPile(s.g.Players[0], bt) {
		t.Errorf("the Maneuver still goes to its owner's discard pile (FR-TGT-003)")
	}
}
