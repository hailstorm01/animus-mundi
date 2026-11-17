package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0264 (FR-TRN-003a, FR-TRN-003, FR-MNV-001, FR-MOD-001,
// FR-REP-001; SVY-R2, SVY-R15, SVY-R29, SVY-R53, SVY-R54): a Maneuver whose
// additional cost is "kill 2 or 4 allied Scratch Company units". The count is
// chosen on declaration among the counts that can be paid, and the cost is paid
// in full. Maneuvers are synthetic; Mutation Swell itself is T-0266. Player 0
// is the active player and enacts.

var alliedScratchFilter = ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Scratch: ability.ScratchOnly}

// killCountManeuver is a Maneuver that draws one card, with the printed cost
// "As an additional cost to enact this maneuver, kill 2 or 4 allied Scratch
// Company units."
func killCountManeuver() *Card {
	return killCountManeuverOf(alliedScratchFilter, steps(drawN(1)), 2, 4)
}

func killCountManeuverOf(f ability.Filter, e ability.Effect, counts ...int) *Card {
	c := maneuver("Mass Sacrifice", 0, 0, e)
	c.Ability.ExtraCosts = []ability.Cost{{Kind: ability.CostKill, Filter: f, Counts: counts}}
	return c
}

// kcCountScript installs a chooser for seat 0 that records every Decision. A
// DecideMode is answered with the option labelled "kill <count>" (timed out
// when count is 0), a DecidePick with the options referring to kill (timed out
// when kill is empty), and anything else is timed out.
func kcCountScript(s *scene, count int, kill ...*Asset) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		switch {
		case d.Kind == DecideMode && count > 0:
			label := "kill " + string(rune('0'+count))
			for i, o := range d.Options {
				if o.Label == label {
					return Answer{Picks: []int{i}}
				}
			}
			s.t.Errorf("no option labelled %q among the options of %q", label, d.Prompt)
		case d.Kind == DecidePick && len(kill) > 0:
			var picks []int
			for _, a := range kill {
				found := false
				for i, o := range d.Options {
					if o.Ref == a.ID {
						picks = append(picks, i)
						found = true
					}
				}
				if !found {
					s.t.Errorf("no option referring to %d among the options of %q", a.ID, d.Prompt)
				}
			}
			return Answer{Picks: picks}
		}
		return Answer{Timeout: true}
	}
}

func kcRefs(d Decision) []int {
	var r []int
	for _, o := range d.Options {
		r = append(r, o.Ref)
	}
	return r
}

func TestFR_TRN_003a_KillCountCostNeedsTheLowestCount(t *testing.T) {
	// Covers: SVY-R29, SVY-R53
	s := newScene(t)
	inst := s.give(0, killCountManeuver())
	s.putScratch(0)
	s.put(0, ZoneMainTheater, unit("Allied Regular", 1, 5, Keywords{}))
	s.putScratch(1)
	s.putScratch(1)
	if offersEnact(s.g, inst) {
		t.Errorf("one allied Scratch Company cannot pay a cost of 2 or 4: a non-Scratch ally and the opponent's Scratch Companies do not count (SVY-R53, SVY-R29)")
	}
	s.reject(enactAction(inst))
	if !inHand(s.g.Players[0], inst) {
		t.Errorf("a rejected Enact leaves the Maneuver in hand (SVY-R29)")
	}
	s.putScratch(0)
	if !offersEnact(s.g, inst) {
		t.Errorf("with two allied Scratch Companies the count 2 can be paid, so the Maneuver is offered (SVY-R53)")
	}
}

func TestFR_TRN_003a_KillCountCostOffersOnlyPayableCounts(t *testing.T) {
	// Covers: SVY-R53, FR-MOD-001
	s := newScene(t)
	a := s.putScratch(0)
	b := s.putScratch(0)
	c := s.putScratch(0)
	inst := s.give(0, killCountManeuver())
	p := s.g.Players[0]
	deck := len(p.Deck)
	if !offersEnact(s.g, inst) {
		t.Fatalf("with three allied Scratch Companies the count 2 can be paid, so the Maneuver is offered (SVY-R53)")
	}
	kcCountScript(s, 0, a, b)
	s.do(enactAction(inst))
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecidePick {
		t.Fatalf("only count 2 can be paid, so no count is asked (FR-MOD-001 sec. 1); the one decision is the kill pick: got %v", kindsOf(asks))
	}
	if len(asks[0].Options) != 3 || asks[0].Min != 2 || asks[0].Max != 2 {
		t.Errorf("the kill pick offers the 3 Scratch Companies and takes exactly 2: got %d options, Min %d, Max %d (SVY-R53)", len(asks[0].Options), asks[0].Min, asks[0].Max)
	}
	if s.alive(a) || s.alive(b) || !s.alive(c) {
		t.Errorf("exactly the two picked Scratch Companies are killed, the third lives (SVY-R53)")
	}
	if got := deck - len(p.Deck); got != 1 {
		t.Errorf("the cost was paid in full, so the Maneuver resolves: drew %d, want 1 (SVY-R53)", got)
	}
}

func TestFR_TRN_003a_KillCountCostChoosesTheCount(t *testing.T) {
	// Covers: SVY-R53, FR-MOD-001
	four := func(s *scene) []*Asset {
		return []*Asset{s.putScratch(0), s.putScratch(0), s.putScratch(0), s.putScratch(0)}
	}
	t.Run("count 4 kills all four with no pick asked", func(t *testing.T) {
		s := newScene(t)
		units := four(s)
		inst := s.give(0, killCountManeuver())
		kcCountScript(s, 4)
		s.do(enactAction(inst))
		asks := s.asked(0)
		if len(asks) != 1 || asks[0].Kind != DecideMode {
			t.Fatalf("the count is chosen first and, with all four killed, nothing else is asked (FR-MOD-001 sec. 1): got %v", kindsOf(asks))
		}
		d := asks[0]
		if len(d.Options) != 2 || d.Options[0].Label != "kill 2" || d.Options[1].Label != "kill 4" || d.Min != 1 || d.Max != 1 {
			t.Errorf("the count decision offers 2 then 4 in printed order and takes one: got %v, Min %d, Max %d (SVY-R53)", d.Options, d.Min, d.Max)
		}
		for _, u := range units {
			if s.alive(u) {
				t.Errorf("count 4 kills all four allied Scratch Companies (SVY-R53)")
			}
		}
		if lastDiscard(s.g.Players[0]) != inst {
			t.Errorf("the Maneuver resolves after the cost is paid (SVY-R53)")
		}
	})
	t.Run("count 2 asks which two", func(t *testing.T) {
		s := newScene(t)
		units := four(s)
		inst := s.give(0, killCountManeuver())
		kcCountScript(s, 2, units[1], units[3])
		s.do(enactAction(inst))
		asks := s.asked(0)
		if len(asks) != 2 || asks[0].Kind != DecideMode || asks[1].Kind != DecidePick {
			t.Fatalf("the count decision comes before the kill pick: got %v", kindsOf(asks))
		}
		if len(asks[1].Options) != 4 || asks[1].Min != 2 || asks[1].Max != 2 {
			t.Errorf("the kill pick offers the four Scratch Companies and takes exactly 2: got %d options, Min %d, Max %d (SVY-R53)", len(asks[1].Options), asks[1].Min, asks[1].Max)
		}
		if s.alive(units[1]) || s.alive(units[3]) || !s.alive(units[0]) || !s.alive(units[2]) {
			t.Errorf("only the two picked Scratch Companies are killed (SVY-R53)")
		}
	})
}

func TestFR_MOD_001_KillCountCostTimeout(t *testing.T) {
	// Covers: SVY-R54, SVY-R16, SVY-R53
	s := newScene(t)
	r1 := s.putScratch(0)
	r2 := s.putScratch(0)
	m1 := s.putScratch(0)
	m2 := s.putScratch(0)
	s.moveTo(m1, ZoneMainTheater)
	s.moveTo(m2, ZoneMainTheater)
	inst := s.give(0, killCountManeuver())
	s.script(0, Answer{Timeout: true}, Answer{Timeout: true})
	s.do(enactAction(inst))
	if s.alive(r1) || s.alive(r2) {
		t.Errorf("a timed-out count is the first listed payable one, 2, and the kills take the first allied Scratch Companies in fixed order, Reserves first (SVY-R54, FR-MOD-001 sec. 6)")
	}
	if !s.alive(m1) || !s.alive(m2) {
		t.Errorf("a timed-out count of 2 leaves the two Main Theater Scratch Companies alive (SVY-R54)")
	}
}

func TestFR_TRN_003a_KillCountCostPaidAfterTargets(t *testing.T) {
	// Covers: SVY-R2, SVY-R53
	s := newScene(t)
	pileFilter := ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}
	x := s.inDiscard(0, unit("X", 1, 1, Keywords{}))[0]
	a := s.putScratch(0)
	b := s.putScratch(0)
	inst := s.give(0, killCountManeuverOf(alliedScratchFilter,
		ability.Effect{Targets: []ability.Target{{Filter: pileFilter}}}, 2, 4))
	kcCountScript(s, 0)
	s.do(enactAction(inst))
	for _, d := range s.asked(0) {
		if d.Kind == DecideTarget {
			if refs := kcRefs(d); len(refs) != 1 || refs[0] != x.ID {
				t.Errorf("the target options are only the unit card already in the pile, not the cost-killed units' cards: got %v, want [%d] (SVY-R2)", refs, x.ID)
			}
		}
	}
	p := s.g.Players[0]
	if s.alive(a) || s.alive(b) || !inDiscard(p, a.Inst) || !inDiscard(p, b.Inst) {
		t.Errorf("both cost-killed units' cards are in their owner's discard pile after the Maneuver resolves (SVY-R53)")
	}
	if lastDiscard(p) != inst {
		t.Errorf("the Maneuver is discarded once resolved (FR-MNV-001)")
	}
}

func TestFR_TRN_003a_KillCountCostTriggersGoAboveTheManeuver(t *testing.T) {
	// Covers: FR-TRN-003, FR-MNV-001
	s := newScene(t)
	hqDamage := func(n int) ability.Step {
		return ability.Step{Op: ability.OpDealDamage, Amount: ability.Value{N: n}, Subject: ability.SubjectAll,
			Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindHeadquarters}}
	}
	m1 := s.put(0, ZoneMainTheater, mortal("Martyr One", 1, 1, killedTrigger(hqDamage(2))))
	m2 := s.put(0, ZoneMainTheater, mortal("Martyr Two", 1, 1, killedTrigger(hqDamage(2))))
	inst := s.give(0, killCountManeuverOf(alliedUnitFilter, steps(hqDamage(1)), 2))
	s.script(0) // both units are taken for a count of 2: nothing is asked
	s.do(enactAction(inst))
	if s.alive(m1) || s.alive(m2) {
		t.Fatalf("precondition: both allied units were killed as the cost")
	}
	var sacPlace, sacDmg, trigPlaces, trigDmgs int
	for _, e := range s.g.Log {
		switch {
		case e.Kind == "place" && e.Card == "Mass Sacrifice":
			sacPlace = e.Seq
		case e.Kind == "place" && (e.Card == "Martyr One" || e.Card == "Martyr Two"):
			trigPlaces++
			if sacPlace == 0 || e.Seq < sacPlace || e.Amount < 2 {
				t.Errorf("each cost kill's trigger is placed above the declared Maneuver (FR-TRN-003a): seq %d, depth %d", e.Seq, e.Amount)
			}
		case e.Kind == "damage" && e.Target == s.g.Players[1].HQ.ID && e.Amount == 2 && sacDmg == 0:
			trigDmgs++
		case e.Kind == "damage" && e.Target == s.g.Players[1].HQ.ID && e.Amount == 1 && sacDmg == 0:
			sacDmg = e.Seq
		}
	}
	if trigPlaces != 2 || sacDmg == 0 {
		t.Fatalf("missing events: %d trigger places (want 2), Maneuver damage seq %d", trigPlaces, sacDmg)
	}
	if trigDmgs != 2 {
		t.Errorf("both cost kills' triggers resolve (2 damage each) before the Maneuver's step: %d did (FR-TRN-003a)", trigDmgs)
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver still resolves and is discarded (FR-TRN-003a)")
	}
}

func TestFR_TRN_003a_ReplacedKillCountCostStaysPaid(t *testing.T) {
	// Covers: SVY-R15, SVY-R21, FR-REP-001
	s := newScene(t)
	si := s.put(0, ZoneReserves, s.card("Shifting Impossibility"))
	a := s.putScratch(0)
	b := s.putScratch(0)
	inst := s.give(0, killCountManeuver())
	p := s.g.Players[0]
	deck := len(p.Deck)
	s.script(0) // both Scratch Companies are taken without a decision
	s.do(enactAction(inst))
	for _, sc := range []*Asset{a, b} {
		if s.alive(sc) || !inHand(p, sc.Inst) || inDiscard(p, sc.Inst) {
			t.Errorf("each Scratch Company killed to pay the cost is returned to its owner's hand instead (FR-REP-001)")
		}
	}
	s.wantDamage(si, 2)
	if got := deck - len(p.Deck); got != 1 {
		t.Errorf("the replaced kills still pay the cost, so the Maneuver resolves: drew %d, want 1 (SVY-R15)", got)
	}
	if lastDiscard(p) != inst {
		t.Errorf("the Maneuver is in its owner's discard pile (SVY-R15)")
	}
}
