package engine

import (
	"fmt"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for the discard step and the own-hand decision (FR-MOD-001 §1, §6;
// SVY-R16). Synthetic cards only.

func discardN(n int) ability.Step {
	return ability.Step{Op: ability.OpDiscard, Amount: ability.Value{N: n}}
}

// discardScene gives player 0 an empty hand, then the named vanilla cards in
// order (the last is the most recently added), and returns the scene and
// those instances. Player 0 then deploys a unit whose deploy trigger runs steps.
func discardScene(t *testing.T, names []string) (*scene, []*Instance) {
	t.Helper()
	s := newScene(t)
	s.g.Players[0].Hand = nil
	insts := make([]*Instance, len(names))
	for i, c := range vanilla(names...) {
		insts[i] = s.give(0, c)
	}
	return s, insts
}

func (s *scene) runDiscardSource(steps ...ability.Step) {
	s.t.Helper()
	s.deployFromHand(0, trigUnit("Source", deployTrigger(steps...)))
}

func discardNames(p *Player) []string {
	out := make([]string, len(p.Discard))
	for i, d := range p.Discard {
		out[i] = d.Card.Name
	}
	return out
}

func TestFR_MOD_001_DiscardAsksTheControllerForACardFromTheirHand(t *testing.T) {
	s, insts := discardScene(t, []string{"A", "B", "C"})
	s.script(0, Answer{Picks: []int{1}})
	s.script(1)
	s.runDiscardSource(discardN(1))
	if len(s.asked(1)) != 0 {
		t.Errorf("the opponent is never asked: got %d decisions (FR-MOD-001)", len(s.asked(1)))
	}
	asks := s.asked(0)
	if len(asks) != 1 {
		t.Fatalf("the controller gets exactly one decision, got %d (FR-MOD-001)", len(asks))
	}
	d := asks[0]
	if d.Kind != DecideHand {
		t.Errorf("kind = %v, want Hand (SVY-R16)", d.Kind)
	}
	if d.Min != 1 || d.Max != 1 {
		t.Errorf("picking one card is required: Min/Max = %d/%d, want 1/1", d.Min, d.Max)
	}
	if len(d.Options) != 3 {
		t.Fatalf("options are exactly the 3 hand cards, got %d", len(d.Options))
	}
	for i, o := range d.Options {
		if o.Ref != insts[i].ID {
			t.Errorf("option %d refers to instance %d, want hand card %d", i, o.Ref, insts[i].ID)
		}
	}
	p := s.g.Players[0]
	if n := len(p.Discard); n == 0 || p.Discard[n-1] != insts[1] {
		t.Errorf("the picked card B ends on top of its owner's discard pile, got %v", discardNames(p))
	}
	if len(p.Hand) != 2 || p.Hand[0] != insts[0] || p.Hand[1] != insts[2] {
		t.Errorf("the other two cards stay in hand, got %d cards", len(p.Hand))
	}
}

func TestFR_MOD_001_DiscardFromAOneCardHandIsAutomatic(t *testing.T) {
	s, insts := discardScene(t, []string{"A"})
	s.script(0)
	s.runDiscardSource(discardN(1))
	if len(s.asked(0)) != 0 {
		t.Errorf("choosing from your own one-card hand is not an inspection: the fast path asks nothing, asked %d (FR-MOD-001 §1)", len(s.asked(0)))
	}
	p := s.g.Players[0]
	if len(p.Hand) != 0 || len(p.Discard) == 0 || p.Discard[len(p.Discard)-1] != insts[0] {
		t.Errorf("the only card is discarded, hand %d, discard %v", len(p.Hand), discardNames(p))
	}
}

func TestFR_MOD_001_DiscardFromAnEmptyHandDoesNothing(t *testing.T) {
	s, _ := discardScene(t, nil)
	s.script(0)
	discardBefore := len(s.g.Players[0].Discard)
	s.runDiscardSource(discardN(1))
	if len(s.asked(0)) != 0 {
		t.Errorf("an empty hand asks nothing, asked %d (FR-MOD-001)", len(s.asked(0)))
	}
	if len(s.g.Players[0].Discard) != discardBefore {
		t.Errorf("the discard pile is unchanged with an empty hand")
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the ability still resolves: stack has %d items", len(s.g.Stack))
	}
	if len(eventsOfKind(s.g.Log, "discard")) != 0 {
		t.Errorf("no discard event with nothing discarded")
	}
}

func TestFR_MOD_001_DiscardTimeoutDiscardsTheCardMostRecentlyAdded(t *testing.T) {
	s, insts := discardScene(t, []string{"A", "B", "C"})
	s.script(0, Answer{Timeout: true})
	s.runDiscardSource(discardN(1))
	p := s.g.Players[0]
	if n := len(p.Discard); n == 0 || p.Discard[n-1] != insts[2] {
		t.Errorf("a timeout discards the card most recently added (C), got discard %v (FR-MOD-001 §6, SVY-R16)", discardNames(p))
	}
	if len(p.Hand) != 2 {
		t.Errorf("hand = %d cards, want 2", len(p.Hand))
	}
}

func TestFR_MOD_001_InvalidDiscardAnswerFallsBackToTheCardMostRecentlyAdded(t *testing.T) {
	s, insts := discardScene(t, []string{"A", "B", "C"})
	s.script(0, Answer{Picks: []int{7}})
	s.runDiscardSource(discardN(1))
	p := s.g.Players[0]
	if n := len(p.Discard); n == 0 || p.Discard[n-1] != insts[2] {
		t.Errorf("an out-of-range pick falls back to the card most recently added (C), got discard %v (FR-MOD-001 §6)", discardNames(p))
	}
}

func TestFR_MOD_001_DrawThenDiscardTimeoutDiscardsTheCardJustDrawn(t *testing.T) {
	s, _ := discardScene(t, []string{"A", "B"})
	drawn := s.onTop(0, vanilla("D")...)[0]
	s.script(0, Answer{Timeout: true})
	s.runDiscardSource(drawN(1), discardN(1))
	p := s.g.Players[0]
	if n := len(p.Discard); n == 0 || p.Discard[n-1] != drawn {
		t.Errorf("a timeout discards the card just drawn (D), got discard %v (FR-MOD-001 §6)", discardNames(p))
	}
	if len(p.Hand) != 2 {
		t.Errorf("hand = %d cards, want the original 2", len(p.Hand))
	}
}

func TestFR_MOD_001_DrawThenDiscardFromAnEmptyHandDiscardsTheDrawnCardAutomatically(t *testing.T) {
	s, _ := discardScene(t, nil)
	drawn := s.onTop(0, vanilla("D")...)[0]
	s.script(0)
	s.runDiscardSource(drawN(1), discardN(1))
	if len(s.asked(0)) != 0 {
		t.Errorf("the drawn card is the only option, so nothing is asked, asked %d (FR-MOD-001 §1, §6)", len(s.asked(0)))
	}
	p := s.g.Players[0]
	if n := len(p.Discard); n == 0 || p.Discard[n-1] != drawn || len(p.Hand) != 0 {
		t.Errorf("the drawn card is discarded, discard %v, hand %d", discardNames(p), len(p.Hand))
	}
}

func TestFR_MOD_001_DiscardEmitsADiscardEvent(t *testing.T) {
	s, insts := discardScene(t, []string{"A", "B", "C"})
	s.script(0, Answer{Picks: []int{1}})
	s.runDiscardSource(discardN(1))
	decAt, discAt := -1, -1
	n := 0
	for i, e := range s.g.Log {
		if e.Kind == "decision" && e.Decision != nil && e.Decision.Kind == "Hand" && decAt < 0 {
			decAt = i
		}
		if e.Kind == "discard" {
			n++
			discAt = i
			if e.Player != 0 || e.Card != "B" || e.Inst != insts[1].ID {
				t.Errorf("discard event = player %d card %q inst %d, want player 0 card B inst %d", e.Player, e.Card, e.Inst, insts[1].ID)
			}
			if e.Zone != "Hand" {
				t.Errorf("discard event zone = %q, want Hand (where it came from)", e.Zone)
			}
			if want := fmt.Sprintf("player %d discards %s", 0, "B"); e.Detail != want {
				t.Errorf("discard event detail = %q, want %q", e.Detail, want)
			}
		}
	}
	if n != 1 {
		t.Fatalf("want one discard event, got %d", n)
	}
	if decAt < 0 || discAt < decAt {
		t.Errorf("the discard event comes after the step's decision event (decision at %d, discard at %d)", decAt, discAt)
	}
}
