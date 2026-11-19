package engine

import (
	"animus-mundi/pkg/engine/ability"
	"testing"
)

// Tests for T-0244: Drifting Concept (FR-MOD-001 §1, §3, §6; SVY-R16, SVY-R50;
// board directive #156; FR-KWD-002g).

// dcScene gives player 1 a hand of the named vanilla cards (the last is the
// most recently added) and empties player 0's hand. Player 0 has the pool to
// deploy Drifting Concept.
func dcScene(t *testing.T, names ...string) (*scene, []*Instance) {
	t.Helper()
	s := newScene(t)
	s.g.Players[0].Hand = nil
	s.g.Players[1].Hand = nil
	s.setPool(0, 5, 5)
	var insts []*Instance
	for _, c := range vanilla(names...) {
		insts = append(insts, s.give(1, c))
	}
	return s, insts
}

// dcDeploy has player 0 deploy the real Drifting Concept from hand.
func (s *scene) dcDeploy() {
	s.t.Helper()
	s.deployFromHand(0, s.card("Drifting Concept"))
}

// dcWantDiscarded checks that want, and only want, left player 1's hand for the
// top of player 1's discard pile, and that player 0's discard pile is unchanged.
func dcWantDiscarded(t *testing.T, s *scene, all []*Instance, want *Instance, why string) {
	t.Helper()
	p1 := s.g.Players[1]
	if n := len(p1.Discard); n != 1 || p1.Discard[n-1] != want {
		t.Errorf("%s: the card is on top of player 1's discard pile, got %v", why, discardNames(p1))
	}
	if inHand(p1, want) {
		t.Errorf("%s: the card has left player 1's hand", why)
	}
	if len(p1.Hand) != len(all)-1 {
		t.Errorf("%s: player 1's hand has %d cards, want %d", why, len(p1.Hand), len(all)-1)
	}
	if len(s.g.Players[0].Discard) != 0 {
		t.Errorf("%s: player 0's discard pile is unchanged, got %v", why, discardNames(s.g.Players[0]))
	}
	if len(s.g.Players[0].Hand) != 0 {
		t.Errorf("%s: player 0's hand is unchanged (empty), got %d cards", why, len(s.g.Players[0].Hand))
	}
}

func TestCard_DriftingConcept_ControllerChoosesFromOpponentsHand(t *testing.T) {
	// Covers: FR-MOD-001
	s, insts := dcScene(t, "A", "B", "C")
	s.script(0, Answer{Picks: []int{1}})
	s.script(1)
	s.dcDeploy()
	if len(s.asked(1)) != 0 {
		t.Errorf("the hand's owner is asked nothing: the controller chooses, got %d decisions (FR-MOD-001 §3)", len(s.asked(1)))
	}
	hands := askedKind(s, 0, DecideHand)
	if len(s.asked(0)) != 1 || len(hands) != 1 {
		t.Fatalf("player 0 is asked exactly one Hand decision, got %d decisions, %d Hand (FR-MOD-001 §3)", len(s.asked(0)), len(hands))
	}
	d := hands[0]
	if !d.Hidden {
		t.Errorf("choosing from an opponent's hand is an inspection: Hidden must be true (FR-MOD-001 §1, §3)")
	}
	if len(d.Options) != 3 {
		t.Fatalf("options are exactly player 1's 3 cards, got %d", len(d.Options))
	}
	for i, o := range d.Options {
		if o.Ref != insts[i].ID {
			t.Errorf("option %d refers to instance %d, want player 1's card %d", i, o.Ref, insts[i].ID)
		}
	}
	dcWantDiscarded(t, s, insts, insts[1], "the picked middle card B")
}

func TestCard_DriftingConcept_TimeoutDiscardsMostRecentlyAdded(t *testing.T) {
	// Covers: FR-MOD-001
	s, insts := dcScene(t, "A", "B", "C")
	s.script(0, Answer{Timeout: true})
	s.script(1)
	s.dcDeploy()
	dcWantDiscarded(t, s, insts, insts[2], "a timeout takes the card most recently added to the opponent's hand, C (FR-MOD-001 §6, SVY-R50)")
}

func TestCard_DriftingConcept_OneCardIsStillAsked(t *testing.T) {
	// Covers: FR-MOD-001
	s, insts := dcScene(t, "A")
	s.script(0, Answer{Picks: []int{0}})
	s.script(1)
	s.dcDeploy()
	hands := askedKind(s, 0, DecideHand)
	if len(hands) != 1 {
		t.Fatalf("an inspection of a one-card opponent hand is still asked, got %d Hand decisions (FR-MOD-001 §1, §3)", len(hands))
	}
	if len(hands[0].Options) != 1 {
		t.Errorf("the decision has one option, got %d (FR-MOD-001 §1)", len(hands[0].Options))
	}
	dcWantDiscarded(t, s, insts, insts[0], "the only card")
}

func TestCard_DriftingConcept_EmptyHandAsksNothing(t *testing.T) {
	// Covers: FR-MOD-001
	s, _ := dcScene(t)
	s.script(0)
	s.script(1)
	s.dcDeploy()
	if n := len(askedKind(s, 0, DecideHand)) + len(askedKind(s, 1, DecideHand)); n != 0 {
		t.Errorf("an empty hand asks no one a Hand decision, got %d (FR-MOD-001)", n)
	}
	if len(s.g.Players[1].Discard) != 0 || len(s.g.Players[0].Discard) != 0 {
		t.Errorf("no discard pile changes, got %v and %v", discardNames(s.g.Players[0]), discardNames(s.g.Players[1]))
	}
}

func TestCard_DriftingConcept_IsEthereal(t *testing.T) {
	// Covers: FR-KWD-002g
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 2, ability.Keywords{}))
	ghost := s.put(1, ZoneMainTheater, s.card("Drifting Concept"))
	s.reject(s.attack(att, ghost))
	s.do(s.attack(att, s.g.Players[1].HQ))
}
