package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for kill steps going through destruction replacement (T-0097, board
// directive #73 part 2; T-0101, directive #74): FR-REP-001 (inline
// substitution, SVY-R15 a replaced action still counts as done, SVY-R20 to
// SVY-R22 fixed order), FR-LAY-002 (nothing is swept while an item resolves)
// and SVY-R21: a replacement sees every death of one kill step, and a source
// left lethal is destroyed at the sweep after the item.
//
// Cards are synthetic and reuse replacement_test.go's guardians. A "killer"
// unit's deploy trigger runs the given steps.

// killStep is "kill each asset matching f".
func killStep(f ability.Filter) ability.Step {
	return ability.Step{Op: ability.OpKill, Subject: ability.SubjectAll, Filter: f}
}

// killer has player p deploy a unit whose deploy trigger runs steps.
func (s *scene) killer(p PlayerID, steps ...ability.Step) {
	s.t.Helper()
	s.deployFromHand(p, trigUnit("Killer", deployTrigger(steps...)))
}

func TestFR_REP_001_ReplacedKillReturnsTheUnitToItsOwnersHand(t *testing.T) {
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, guardianStructure(3))
	deathDraw := mortal("Martyr", 1, 1, killedTrigger(drawN(1)))
	u := s.put(0, ZoneReserves, deathDraw)
	deck := len(s.g.Players[0].Deck)
	s.killer(0, killStep(alliedUnits))
	if s.alive(u) {
		t.Errorf("the replaced unit should have left the battlefield (FR-REP-001)")
	}
	if !inHand(s.g.Players[0], u.Inst) {
		t.Errorf("a unit whose kill is replaced is returned to its owner's hand (FR-REP-001)")
	}
	if inDiscard(s.g.Players[0], u.Inst) {
		t.Errorf("a replaced kill does not put the unit into the discard pile (FR-REP-001)")
	}
	s.wantDamage(src, 1)
	wantReplaced(t, s, u)
	if eventSeq(s, "leave", u.ID) == 0 {
		t.Errorf("returning the unit to hand logs a leave event for it (FR-REP-001)")
	}
	if n := eventSeq(s, "destroy", u.ID); n != 0 {
		t.Errorf("the replaced destruction never occurs, so no destroy event for the unit (FR-REP-001), got seq %d", n)
	}
	if got := len(s.g.Players[0].Deck); got != deck {
		t.Errorf("the replaced event never occurs, so the unit's death trigger does not fire (no draw): deck %d -> %d (FR-REP-001)", deck, got)
	}
}

func TestFR_REP_001_ReplacedKillStillCountsAsDone(t *testing.T) {
	// Covers: SVY-R15. The Infested Apostle example: the kill is replaced, but
	// the player carried it out, so "if you do" runs.
	s := newScene(t)
	s.put(0, ZoneMainTheater, guardianStructure(3))
	u := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	deck := len(s.g.Players[0].Deck)
	draw := drawN(1)
	draw.IfDone = true
	s.killer(0, killStep(alliedUnits), draw)
	if s.alive(u) || !inHand(s.g.Players[0], u.Inst) {
		t.Fatalf("precondition: the unit's kill is replaced and it is in its owner's hand")
	}
	if got := len(s.g.Players[0].Deck); got != deck-1 {
		t.Errorf("a replaced kill still counts as done, so the \"if you do\" draw runs: deck %d -> %d, want %d (SVY-R15, FR-REP-001)", deck, got, deck-1)
	}
}

func TestFR_REP_001_OneKillStepsDeathsAreAllReplaced(t *testing.T) {
	// SVY-R21. The source has 1 Health: the first byproduct leaves
	// it lethal, but nothing is swept while the item resolves, so it still
	// replaces the second kill, and is destroyed at the sweep after the item.
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, guardianStructure(1))
	a := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	b := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	s.killer(0, killStep(alliedUnits))
	wantReplaced(t, s, a, b)
	for _, u := range []*Asset{a, b} {
		if s.alive(u) || !inHand(s.g.Players[0], u.Inst) || inDiscard(s.g.Players[0], u.Inst) {
			t.Errorf("%s is replaced and returns to hand although the source is lethal after the first kill (SVY-R21)", u)
		}
	}
	if s.alive(src) || !inDiscard(s.g.Players[0], src.Inst) {
		t.Errorf("the lethal source is destroyed at the sweep after the item (SVY-R21, FR-LAY-002)")
	}
	last := max(eventSeq(s, "leave", a.ID), eventSeq(s, "leave", b.ID))
	if d := eventSeq(s, "destroy", src.ID); d == 0 || d < last {
		t.Errorf("the source is destroyed (seq %d) only after both replaced kills (last leave seq %d) (SVY-R21)", d, last)
	}
}

func TestFR_REP_001_KillStepDestroysWhatNoReplacementAppliesTo(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, guardianStructure(3))
	ally := s.put(0, ZoneReserves, unit("Ally", 1, 1, Keywords{}))
	foe := s.put(1, ZoneReserves, mortal("Foe", 1, 1, killedTrigger(drawN(1))))
	foeDeck := len(s.g.Players[1].Deck)
	s.killer(0, killStep(ability.Filter{Kinds: ability.KindUnit, Other: true}))
	if s.alive(ally) || !inHand(s.g.Players[0], ally.Inst) {
		t.Errorf("the allied unit's kill is replaced and it returns to hand (FR-REP-001)")
	}
	if s.alive(foe) || !inDiscard(s.g.Players[1], foe.Inst) {
		t.Errorf("an opposing unit is outside the allied filter, so the kill destroys it into the discard pile (FR-REP-001)")
	}
	wantReplaced(t, s, ally)
	if got := len(s.g.Players[1].Deck); got != foeDeck-1 {
		t.Errorf("the destroyed opposing unit's death trigger fires (draw 1): deck %d -> %d (FR-REP-001)", foeDeck, got)
	}
}

func TestFR_REP_001_SourceKilledByTheSameStepReplacesTheOthersFirst(t *testing.T) {
	// SVY-R21: the source is killed in the same moment as the
	// units it protects, yet it replaces their deaths, then is destroyed.
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, guardianUnit(3))
	a := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	b := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	s.killer(0, killStep(ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Other: true}))
	wantReplaced(t, s, a, b)
	for _, u := range []*Asset{a, b} {
		if s.alive(u) || !inHand(s.g.Players[0], u.Inst) {
			t.Errorf("%s is replaced and returns to hand though the source is killed by the same step (SVY-R21)", u)
		}
	}
	if s.alive(src) || !inDiscard(s.g.Players[0], src.Inst) {
		t.Errorf("the source, killed by the same step, is destroyed (FR-REP-001)")
	}
	last := max(eventSeq(s, "leave", a.ID), eventSeq(s, "leave", b.ID))
	if d := eventSeq(s, "destroy", src.ID); d == 0 || d < last {
		t.Errorf("the source is destroyed (seq %d) after both replaced kills (last leave seq %d) (SVY-R21)", d, last)
	}
}

func TestFR_REP_001_ReplacedKillsAreNotOrderedByThePlayer(t *testing.T) {
	// SVY-R20, SVY-R21. Player 0's scripted order answer would reverse the kills;
	// it is never asked for, and both kills are replaced oldest first.
	s := newScene(t)
	s.put(0, ZoneMainTheater, guardianStructure(3))
	a := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	b := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	s.script(0, Answer{Picks: []int{1, 0}})
	s.killer(0, killStep(alliedUnits))
	wantReplaced(t, s, a, b)
	noOrderAsked(t, s, 0)
}

func TestFR_REP_001_KillStepAsksForEachKill(t *testing.T) {
	// SVY-R21, SVY-R23: a kill step's replaced deaths are each a choice.
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, guardianStructure(4))
	b := s.put(0, ZoneMainTheater, guardianStructure(4))
	u1 := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	u2 := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	s.script(0, Answer{Picks: []int{0}}, Answer{Picks: []int{1}})
	s.killer(0, killStep(alliedUnits))
	asks := replacementAsks(s, 0)
	if len(asks) != 2 {
		t.Fatalf("player 0 was asked %d replacement questions, want 2, one per kill (SVY-R21, SVY-R23)", len(asks))
	}
	for _, d := range asks {
		wantOffered(t, d, a.ID, b.ID)
	}
	wantReplaced(t, s, u1, u2)
	for _, u := range []*Asset{u1, u2} {
		if s.alive(u) || !inHand(s.g.Players[0], u.Inst) || inDiscard(s.g.Players[0], u.Inst) {
			t.Errorf("%s's kill is replaced and it returns to hand (FR-REP-001)", u)
		}
	}
}
