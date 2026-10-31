package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0128 (FR-TRN-003a, FR-TRN-003, FR-MNV-001, FR-TGT-001,
// FR-MOD-001, FR-REP-001; SVY-R2, SVY-R15, SVY-R16, SVY-R29): a Maneuver whose
// additional cost is "kill an allied unit". The cost is paid on declaration,
// its triggers go on the stack above the Maneuver, a replaced kill still pays
// it, and a Maneuver whose cost cannot be paid cannot be enacted. Maneuvers are
// synthetic; the real cards are tested in cards_animus_maneuvers_test.go.
// Player 0 is the active player and enacts.

var alliedUnitFilter = ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}

// killCostManeuver is maneuver(...) with the printed cost "As an additional
// cost to enact this maneuver, kill an allied unit."
func killCostManeuver(name string, req, tac int, e ability.Effect) *Card {
	c := maneuver(name, req, tac, e)
	c.Ability.ExtraCosts = []ability.Cost{{Kind: ability.CostKill, Filter: alliedUnitFilter}}
	return c
}

// kcScript installs a chooser for seat 0 that records every Decision. A
// DecideTarget is answered with the option referring to target (an *Asset or
// an *Instance, by ID) and a DecidePick with the option referring to payer;
// either that is nil is timed out. A missing option is a test error.
func kcScript(s *scene, target any, payer *Asset) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	answer := func(d Decision, ref int) Answer {
		for i, o := range d.Options {
			if o.Ref == ref {
				return Answer{Picks: []int{i}}
			}
		}
		s.t.Errorf("no option referring to %d among the options of %q", ref, d.Prompt)
		return Answer{Timeout: true}
	}
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		switch {
		case d.Kind == DecideTarget && target != nil:
			switch v := target.(type) {
			case *Asset:
				return answer(d, v.ID)
			case *Instance:
				return answer(d, v.ID)
			}
		case d.Kind == DecidePick && payer != nil:
			return answer(d, payer.ID)
		}
		return Answer{Timeout: true}
	}
}

func TestFR_TRN_003a_KillCostPaidOnDeclaration(t *testing.T) {
	// Covers: FR-TRN-003, FR-MNV-001
	s := newScene(t)
	u1 := s.put(0, ZoneMainTheater, unit("First", 1, 5, Keywords{}))
	u2 := s.put(0, ZoneMainTheater, unit("Second", 1, 5, Keywords{}))
	inst := s.give(0, killCostManeuver("Sacrifice", 0, 0, steps(drawN(1))))
	p := s.g.Players[0]
	deck := len(p.Deck)
	s.scriptRefs(0, nil, u2)
	s.do(enactAction(inst))
	if s.alive(u2) || !inDiscard(p, u2.Inst) {
		t.Errorf("the allied unit the controller chose is killed as the cost, into its owner's discard pile (FR-TRN-003a)")
	}
	if !s.alive(u1) {
		t.Errorf("only the chosen allied unit is killed; the other stays on the battlefield (FR-TRN-003a)")
	}
	if got := deck - len(p.Deck); got != 1 {
		t.Errorf("the Maneuver's step still ran after the cost: drew %d, want 1 (FR-TRN-003a)", got)
	}
	if lastDiscard(p) != inst {
		t.Errorf("the Maneuver goes to its owner's discard pile once resolved (FR-MNV-001)")
	}
}

func TestFR_TRN_003a_KillCostMustBePayable(t *testing.T) {
	// Covers: SVY-R29, FR-TRN-003
	s := newScene(t)
	s.put(0, ZoneMainTheater, structure("Allied Wall", 0, 5, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Foe", 1, 5, Keywords{}))
	inst := s.give(0, killCostManeuver("Sacrifice", 1, 1, steps(drawN(1))))
	p := s.g.Players[0]
	p.Req, p.Tac = 1, 1
	if offersEnact(s.g, inst) {
		t.Errorf("a Maneuver whose kill cost has no allied unit to pay it is not offered (SVY-R29, FR-TRN-003a)")
	}
	s.reject(enactAction(inst))
	if !inHand(p, inst) || p.Req != 1 || p.Tac != 1 {
		t.Errorf("a rejected Enact changes nothing: in hand %v, Req %d, Tac %d, want true, 1, 1 (SVY-R29)", inHand(p, inst), p.Req, p.Tac)
	}
	s.put(0, ZoneReserves, unit("Ally", 1, 1, Keywords{}))
	if !offersEnact(s.g, inst) {
		t.Errorf("with an allied unit to kill, the Maneuver is offered (FR-TRN-003a)")
	}
}

func TestFR_TRN_003a_CostTriggersResolveBeforeTheManeuver(t *testing.T) {
	// Covers: FR-TRN-003, FR-MNV-001
	s := newScene(t)
	martyr := s.put(0, ZoneMainTheater, mortal("Martyr", 1, 1, killedTrigger(ability.Step{
		Op: ability.OpDealDamage, Amount: ability.Value{N: 2}, Subject: ability.SubjectAll,
		Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindHeadquarters}})))
	inst := s.give(0, killCostManeuver("Sacrifice", 0, 0, steps(drawN(1))))
	s.do(enactAction(inst))
	if s.alive(martyr) {
		t.Fatalf("precondition: the only allied unit was killed as the cost")
	}
	var sacPlace, trigPlace, dmg, draw, discard int
	for _, e := range s.g.Log {
		switch {
		case e.Kind == "place" && e.Card == "Sacrifice":
			sacPlace = e.Seq
			if e.Amount != 1 {
				t.Errorf("the Maneuver is placed at stack depth 1, got %d (FR-TRN-003)", e.Amount)
			}
		case e.Kind == "place" && e.Card == "Martyr":
			trigPlace = e.Seq
			if e.Amount != 2 {
				t.Errorf("the cost's trigger is placed on top of the Maneuver, depth 2, got %d (FR-TRN-003a)", e.Amount)
			}
		case e.Kind == "damage" && e.Target == s.g.Players[1].HQ.ID:
			dmg = e.Seq
		case e.Kind == "draw" && sacPlace != 0 && draw == 0:
			draw = e.Seq
		case e.Kind == "discard" && e.Inst == inst.ID:
			discard = e.Seq
		}
	}
	if sacPlace == 0 || trigPlace == 0 || dmg == 0 || draw == 0 {
		t.Fatalf("missing events: Maneuver place %d, trigger place %d, trigger damage %d, Maneuver draw %d", sacPlace, trigPlace, dmg, draw)
	}
	if trigPlace < sacPlace {
		t.Errorf("the trigger goes onto the stack after the Maneuver was declared, above it (seq %d vs %d) (FR-TRN-003a)", trigPlace, sacPlace)
	}
	if dmg > draw {
		t.Errorf("the cost's trigger resolves (seq %d) before the Maneuver's step (seq %d) (FR-TRN-003a)", dmg, draw)
	}
	if discard != 0 && dmg > discard {
		t.Errorf("the cost's trigger resolves before the Maneuver goes to the discard pile (FR-TRN-003a)")
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver still resolves and is discarded (FR-TRN-003a)")
	}
}

func TestFR_TRN_003a_ReplacedKillCostStaysPaid(t *testing.T) {
	// Covers: SVY-R15, FR-REP-001
	s := newScene(t)
	si := s.put(0, ZoneReserves, s.card("Shifting Impossibility"))
	sc := s.putScratch(0)
	inst := s.give(0, killCostManeuver("Sacrifice", 0, 0, steps(drawN(1))))
	p := s.g.Players[0]
	deck := len(p.Deck)
	s.scriptRefs(0, nil, sc)
	s.do(enactAction(inst))
	if s.alive(sc) || !inHand(p, sc.Inst) || inDiscard(p, sc.Inst) {
		t.Errorf("the Scratch Company killed to pay the cost is returned to its owner's hand instead (FR-REP-001)")
	}
	s.wantDamage(si, 1)
	if got := deck - len(p.Deck); got != 1 {
		t.Errorf("the replaced kill still pays the cost, so the Maneuver resolves: drew %d, want 1 (SVY-R15)", got)
	}
	if lastDiscard(p) != inst {
		t.Errorf("the Maneuver is in its owner's discard pile (SVY-R15)")
	}
}

func TestFR_TRN_003a_CostPaidAfterTargetsAreChosen(t *testing.T) {
	// Covers: SVY-R2, FR-TGT-001
	pileSlot := ability.Target{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}}
	card := func() *Card {
		return killCostManeuver("Recall", 0, 0, ability.Effect{
			Targets: []ability.Target{pileSlot},
			Steps:   []ability.Step{ptStep(ability.OpReturnToHand, 0)},
		})
	}
	t.Run("the card the cost would put in the pile does not count", func(t *testing.T) {
		s := newScene(t)
		s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
		inst := s.give(0, card())
		if offersEnact(s.g, inst) {
			t.Errorf("with no unit card in the discard pile yet, the Maneuver is not offered: the card the cost puts there is not a target (SVY-R2)")
		}
		s.reject(enactAction(inst))
		if !inHand(s.g.Players[0], inst) {
			t.Errorf("a rejected Enact leaves the card in hand (SVY-R29)")
		}
	})
	t.Run("targets are judged before the cost kills", func(t *testing.T) {
		s := newScene(t)
		u1 := s.put(0, ZoneMainTheater, unit("Ally One", 1, 5, Keywords{}))
		u2 := s.put(0, ZoneMainTheater, unit("Ally Two", 1, 5, Keywords{}))
		piled := s.inDiscard(0, unit("C", 1, 1, Keywords{}), unit("D", 1, 1, Keywords{}))
		inst := s.give(0, card())
		if !offersEnact(s.g, inst) {
			t.Fatalf("with a unit card in the discard pile and an allied unit to kill, the Maneuver is offered")
		}
		kcScript(s, piled[0], u2)
		s.do(enactAction(inst))
		asks := s.asked(0)
		if len(asks) != 2 || asks[0].Kind != DecideTarget || asks[1].Kind != DecidePick {
			t.Fatalf("the target decision is asked before the cost's: got %d decisions %v (SVY-R2)", len(asks), kindsOf(asks))
		}
		got, want := ppRefs(asks[0]), ppInstIDs(piled[0], piled[1])
		if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
			t.Errorf("the target options are the unit cards already in the pile, not the unit the cost then kills: got %v, want %v (SVY-R2)", got, want)
		}
		p := s.g.Players[0]
		if !inHand(p, piled[0]) {
			t.Errorf("the chosen card C is returned to hand (SVY-R2)")
		}
		if s.alive(u2) || !s.alive(u1) || !inDiscard(p, u2.Inst) {
			t.Errorf("the cost killed the chosen unit and only it (FR-TRN-003a)")
		}
	})
}

func TestFR_MOD_001_KillCostTimeoutPicksFixedOrder(t *testing.T) {
	// Covers: SVY-R16, FR-TRN-003a
	s := newScene(t)
	mt := s.put(0, ZoneMainTheater, unit("Theater Unit", 1, 5, Keywords{})) // deployed first
	rs := s.put(0, ZoneReserves, unit("Reserves Unit", 1, 5, Keywords{}))   // deployed later
	inst := s.give(0, killCostManeuver("Sacrifice", 0, 0, steps(drawN(1))))
	s.script(0, Answer{Timeout: true})
	s.do(enactAction(inst))
	if n := len(s.asked(0)); n != 1 {
		t.Errorf("two options for the cost are asked once, got %d decisions", n)
	}
	if s.alive(rs) || !s.alive(mt) {
		t.Errorf("a timed-out cost choice kills the first allied unit in fixed order, the Reserves unit before the Main Theater unit (FR-MOD-001 sec. 6, SVY-R16)")
	}
}

func TestFR_MOD_001_KillCostSingleOptionNotAsked(t *testing.T) {
	// Covers: FR-TRN-003a
	s := newScene(t)
	u := s.put(0, ZoneMainTheater, unit("Only Ally", 1, 5, Keywords{}))
	inst := s.give(0, killCostManeuver("Sacrifice", 0, 0, steps(drawN(1))))
	s.script(0) // any decision asked is an error
	s.do(enactAction(inst))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a single legal option is taken without a decision, got %d decisions (FR-MOD-001 sec. 1)", n)
	}
	if s.alive(u) || !inDiscard(s.g.Players[0], u.Inst) {
		t.Errorf("the only allied unit is killed as the cost (FR-TRN-003a)")
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver resolves and is discarded (FR-TRN-003a)")
	}
}
