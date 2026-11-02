package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0129 (FR-TRN-003a, FR-TRN-003, FR-MNV-001, FR-CAP-001,
// FR-MOD-001; SVY-R16, SVY-R17, SVY-R29): a Maneuver whose additional cost is
// "exhaust a readied allied unit". The cost is paid on declaration, its
// exhaust triggers go on the stack above the Maneuver, and a Maneuver whose
// cost cannot be paid cannot be enacted. Exhausting an already-Exhausted asset
// does nothing (directive #58). Maneuvers are synthetic. Player 0 is the
// active player and enacts.

// exhaustCostManeuver is maneuver(...) with the printed cost "As an additional
// cost to enact this maneuver, exhaust a readied allied unit."
func exhaustCostManeuver(name string, req, tac int, e ability.Effect) *Card {
	c := maneuver(name, req, tac, e)
	c.Ability.ExtraCosts = []ability.Cost{{Kind: ability.CostExhaust, Filter: alliedUnitFilter}}
	return c
}

func TestFR_TRN_003a_ExhaustCostPaidOnDeclaration(t *testing.T) {
	// Covers: FR-TRN-003, FR-MNV-001, FR-CAP-001
	s := newScene(t)
	u1 := s.put(0, ZoneMainTheater, unit("First", 1, 5, Keywords{}))
	u2 := s.put(0, ZoneMainTheater, unit("Second", 1, 5, Keywords{}))
	inst := s.give(0, exhaustCostManeuver("Fungus", 1, 1, steps(drawN(1))))
	p := s.g.Players[0]
	p.Req, p.Tac = 1, 1
	deck := len(p.Deck)
	s.scriptRefs(0, nil, u2)
	s.do(enactAction(inst))
	if !s.alive(u2) || !u2.Exhausted {
		t.Errorf("the allied unit the controller chose is exhausted as the cost and stays on the battlefield (FR-TRN-003a)")
	}
	if !s.alive(u1) || u1.Exhausted {
		t.Errorf("only the chosen allied unit is exhausted; the other stays Readied (FR-TRN-003a)")
	}
	if got := deck - len(p.Deck); got != 1 {
		t.Errorf("the Maneuver's step still ran after the cost: drew %d, want 1 (FR-TRN-003a)", got)
	}
	if lastDiscard(p) != inst {
		t.Errorf("the Maneuver goes to its owner's discard pile once resolved (FR-MNV-001)")
	}
	if p.Req != 0 || p.Tac != 0 {
		t.Errorf("enacting pays the Maneuver's resource cost R1 T1: pool R%d T%d, want R0 T0 (FR-TRN-003)", p.Req, p.Tac)
	}
}

func TestFR_TRN_003a_ExhaustCostNeedsAReadiedUnit(t *testing.T) {
	// Covers: SVY-R29, FR-TRN-003
	s := newScene(t)
	tired := s.put(0, ZoneMainTheater, unit("Tired Ally", 1, 5, Keywords{}))
	tired.Exhausted = true
	s.put(0, ZoneMainTheater, structure("Allied Wall", 0, 5, Keywords{}))
	inst := s.give(0, exhaustCostManeuver("Fungus", 1, 1, steps(drawN(1))))
	p := s.g.Players[0]
	p.Req, p.Tac = 1, 1
	if offersEnact(s.g, inst) {
		t.Errorf("a Maneuver whose exhaust cost has no Readied allied unit to pay it is not offered; Headquarters, Structures and Exhausted units do not count (SVY-R29, FR-TRN-003a)")
	}
	s.reject(enactAction(inst))
	if !inHand(p, inst) || p.Req != 1 || p.Tac != 1 {
		t.Errorf("a rejected Enact changes nothing: in hand %v, Req %d, Tac %d, want true, 1, 1 (SVY-R29)", inHand(p, inst), p.Req, p.Tac)
	}
	s.put(0, ZoneReserves, unit("Fresh Ally", 1, 1, Keywords{}))
	if !offersEnact(s.g, inst) {
		t.Errorf("with a Readied allied unit to exhaust, the Maneuver is offered (FR-TRN-003a)")
	}
}

func TestFR_TRN_003a_ExhaustCostTriggerResolvesBeforeTheManeuver(t *testing.T) {
	// Covers: FR-TRN-003, FR-MNV-001, FR-CAP-001
	s := newScene(t)
	hqDamage := func(n int) ability.Step {
		return ability.Step{Op: ability.OpDealDamage, Amount: ability.Value{N: n}, Subject: ability.SubjectAll,
			Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindHeadquarters}}
	}
	payer := s.put(0, ZoneMainTheater, striker("Payer", 1, 5, Keywords{}, exhaustTrigger(hqDamage(2))))
	inst := s.give(0, exhaustCostManeuver("Fungus", 0, 0, steps(hqDamage(1))))
	s.do(enactAction(inst))
	if !payer.Exhausted {
		t.Fatalf("precondition: the only Readied allied unit was exhausted as the cost")
	}
	var manPlace, trigPlace, trigDmg, manDmg, discard int
	for _, e := range s.g.Log {
		switch {
		case e.Kind == "place" && e.Card == "Fungus":
			manPlace = e.Seq
			if e.Amount != 1 {
				t.Errorf("the Maneuver is placed at stack depth 1, got %d (FR-TRN-003)", e.Amount)
			}
		case e.Kind == "place" && e.Card == "Payer":
			trigPlace = e.Seq
			if e.Amount != 2 {
				t.Errorf("the cost's exhaust trigger is placed on top of the Maneuver, depth 2, got %d (FR-TRN-003a, FR-CAP-001)", e.Amount)
			}
		case e.Kind == "damage" && e.Target == s.g.Players[1].HQ.ID && e.Amount == 2 && manPlace != 0 && trigDmg == 0:
			trigDmg = e.Seq
		case e.Kind == "damage" && e.Target == s.g.Players[1].HQ.ID && e.Amount == 1 && manPlace != 0 && manDmg == 0:
			manDmg = e.Seq
		case e.Kind == "discard" && e.Inst == inst.ID:
			discard = e.Seq
		}
	}
	if manPlace == 0 || trigPlace == 0 || trigDmg == 0 || manDmg == 0 {
		t.Fatalf("missing events: Maneuver place %d, trigger place %d, trigger 2 damage %d, Maneuver 1 damage %d", manPlace, trigPlace, trigDmg, manDmg)
	}
	if trigPlace < manPlace {
		t.Errorf("the trigger goes onto the stack after the Maneuver was declared, above it (seq %d vs %d) (FR-TRN-003a)", trigPlace, manPlace)
	}
	if trigDmg > manDmg {
		t.Errorf("the cost's trigger resolves (2 damage, seq %d) before the Maneuver's step (1 damage, seq %d) (FR-TRN-003a)", trigDmg, manDmg)
	}
	if discard != 0 && trigDmg > discard {
		t.Errorf("the cost's trigger resolves before the Maneuver goes to the discard pile (FR-TRN-003a)")
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver still resolves and is discarded (FR-TRN-003a)")
	}
}

func TestFR_MOD_001_ExhaustCostTimeoutPicksFixedOrder(t *testing.T) {
	// Covers: SVY-R16, FR-TRN-003a
	s := newScene(t)
	mt := s.put(0, ZoneMainTheater, unit("Theater Unit", 1, 5, Keywords{})) // deployed first
	rs := s.put(0, ZoneReserves, unit("Reserves Unit", 1, 5, Keywords{}))   // deployed later
	inst := s.give(0, exhaustCostManeuver("Fungus", 0, 0, steps(drawN(1))))
	s.script(0, Answer{Timeout: true})
	s.do(enactAction(inst))
	if n := len(s.asked(0)); n != 1 {
		t.Errorf("two options for the cost are asked once, got %d decisions", n)
	}
	if !rs.Exhausted || mt.Exhausted {
		t.Errorf("a timed-out cost choice exhausts the first Readied allied unit in fixed order, the Reserves unit before the Main Theater unit (FR-MOD-001 sec. 6, SVY-R16)")
	}
}

func TestFR_MOD_001_ExhaustCostSingleOptionNotAsked(t *testing.T) {
	// Covers: FR-TRN-003a
	s := newScene(t)
	ready := s.put(0, ZoneMainTheater, unit("Ready Ally", 1, 5, Keywords{}))
	tired := s.put(0, ZoneReserves, unit("Tired Ally", 1, 5, Keywords{}))
	tired.Exhausted = true
	inst := s.give(0, exhaustCostManeuver("Fungus", 0, 0, steps(drawN(1))))
	s.script(0) // any decision asked is an error
	s.do(enactAction(inst))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a single legal option is taken without a decision; an Exhausted unit is not an option: got %d decisions (FR-MOD-001 sec. 1)", n)
	}
	if !ready.Exhausted {
		t.Errorf("the only Readied allied unit is exhausted as the cost (FR-TRN-003a)")
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver resolves and is discarded (FR-TRN-003a)")
	}
}

func TestFR_CAP_001_ExhaustingAnExhaustedAssetDoesNothing(t *testing.T) {
	// Covers: FR-CAP-001
	t.Run("already Exhausted fires nothing", func(t *testing.T) {
		s := newScene(t)
		w := s.put(0, ZoneMainTheater, exhaustWatcher("Watcher"))
		w.Exhausted = true
		s.g.exhaust(w)
		s.g.resolveStack()
		s.wantScratches(0, 0, "exhausting an asset that is already Exhausted is not a Readied to Exhausted change, so its exhaust trigger does not fire (SVY-R17)")
		if !w.Exhausted {
			t.Errorf("the asset stays Exhausted")
		}
		s.wantSettled("exhausting an already Exhausted asset")
	})
	t.Run("Readied fires once", func(t *testing.T) {
		s := newScene(t)
		w := s.put(0, ZoneMainTheater, exhaustWatcher("Watcher"))
		s.g.exhaust(w)
		s.g.resolveStack()
		s.wantScratches(0, 1, "control: exhausting a Readied asset fires its exhaust trigger once (SVY-R17)")
		if !w.Exhausted {
			t.Errorf("the asset is Exhausted")
		}
	})
}
