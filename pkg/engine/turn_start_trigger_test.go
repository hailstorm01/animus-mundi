package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for "At the start of your turn" triggers (T-0245, board directive #15 (1)):
// they are placed in APNAP order and resolve completely, cascades included, in the
// Ready Phase, before the Draw Phase (FR-TRN-001 1, FR-TRG-003, FR-TRG-004).

// startTrigger is one "At the start of <whose> turn, <effect>" ability.
func startTrigger(turn ability.TurnOwner, e ability.Effect) ability.Trigger {
	return ability.Trigger{Hook: ability.HookTurnStart, Turn: turn, Effect: e}
}

// startSteps is a start-of-turn trigger with no targets.
func startSteps(turn ability.TurnOwner, steps ...ability.Step) ability.Trigger {
	return startTrigger(turn, ability.Effect{Steps: steps})
}

// startStructure is a structure printing the given triggers in order.
func startStructure(name string, trigs ...ability.Trigger) *Card {
	c := structure(name, 1, 9, Keywords{})
	c.Ability = &ability.Def{Triggers: trigs}
	return c
}

func TestFR_TRN_001_StartOfTurnTriggerResolvesBeforeDraw(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, startStructure("Bivouac", startSteps(ability.TurnYours, scratchN(1))))
	s.endTurn()
	cards := s.onTop(0, vanilla("A", "B", "C")...)
	s.endTurn()
	p := s.g.Players[0]
	if !scratched(p, cards[0]) {
		t.Errorf("the start-of-turn Scratch resolves before the Draw Phase and takes the top card (FR-TRN-001 1)")
	}
	if !inHand(p, cards[1]) {
		t.Errorf("the Draw Phase then draws the card second from the top, because the trigger already took the top one (FR-TRN-001 1-2)")
	}
	if s.g.Phase != PhaseMain || len(s.g.Stack) != 0 {
		t.Errorf("after the turn start the phase is Main with an empty stack: phase %v, stack %d", s.g.Phase, len(s.g.Stack))
	}
}

func TestFR_TRN_001_StartOfYourTurnOnly(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, startStructure("Bivouac", startSteps(ability.TurnYours, scratchN(1))))
	p := s.g.Players[0]
	deck, faceDown := len(p.Deck), faceDownCount(p)
	s.endTurn() // player 1's turn starts
	if len(p.Deck) != deck || faceDownCount(p) != faceDown {
		t.Errorf("\"at the start of your turn\" does not fire when the opponent's turn starts (FR-TRN-001 1): deck %d -> %d, Scratch Companies %d -> %d",
			deck, len(p.Deck), faceDown, faceDownCount(p))
	}
}

func TestFR_TRN_001_StartOfTurnTriggerAfterReady(t *testing.T) {
	s := newScene(t)
	tapped := s.put(0, ZoneMainTheater, unit("Tired", 1, 5, Keywords{}))
	s.put(0, ZoneMainTheater, unit("Fresh", 1, 5, Keywords{}))
	allied := ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}
	s.put(0, ZoneMainTheater, startStructure("Bivouac",
		startTrigger(ability.TurnYours, ability.Effect{Targets: []ability.Target{{Filter: allied}}, Steps: []ability.Step{armorStep(1, 0)}})))
	tapped.Exhausted = true
	s.endTurn()
	if !tapped.Exhausted {
		t.Fatalf("precondition: the opponent's turn does not ready player 0's assets")
	}
	asked, readiedWhenAsked := 0, true
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		if d.Kind == DecideTarget {
			asked++
			if tapped.Exhausted {
				readiedWhenAsked = false
			}
		}
		return Answer{Picks: []int{0}}
	}
	s.endTurn()
	if asked == 0 {
		t.Fatalf("with two allied units the start trigger asks its controller for a target")
	}
	if !readiedWhenAsked {
		t.Errorf("start-of-turn triggers are placed after the active player's assets ready (FR-TRN-001 1): the unit was still Exhausted when the target was chosen")
	}
}

func TestFR_TRN_001_StartOfTurnCascadeResolvesBeforeDraw(t *testing.T) {
	// Covers: FR-TRG-004
	s := newScene(t)
	allied := ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}
	s.put(0, ZoneMainTheater, mortal("Fragile", 1, 1, killedTrigger(scratchN(1))))
	s.put(0, ZoneMainTheater, startStructure("Bivouac", startSteps(ability.TurnYours, scratchN(1), allDamage(1, allied))))
	s.endTurn()
	cards := s.onTop(0, vanilla("A", "B", "C")...)
	s.endTurn()
	p := s.g.Players[0]
	if !scratched(p, cards[0]) || !scratched(p, cards[1]) {
		t.Errorf("the start trigger and the death trigger it causes both resolve before the Draw Phase, taking the top two cards (FR-TRN-001 1, FR-TRG-004)")
	}
	if !inHand(p, cards[2]) {
		t.Errorf("the Draw Phase draws the third card from the top (FR-TRN-001 2)")
	}
}

func TestFR_TRG_003_StartOfTurnTriggersInAPNAPOrder(t *testing.T) {
	// Both decks are empty and both triggers are Draw 1, so whichever resolves
	// first loses at its own sweep. The non-active player's trigger is placed on
	// top and resolves first, so player 1 loses and player 0 wins.
	s := newScene(t)
	s.endTurn()
	s.put(0, ZoneMainTheater, startStructure("Mine", startSteps(ability.TurnYours, drawN(1))))
	s.put(1, ZoneMainTheater, startStructure("Theirs", startSteps(ability.TurnAny, drawN(1))))
	s.g.Players[0].Deck = nil
	s.g.Players[1].Deck = nil
	s.endTurn()
	if !s.g.Outcome.Over || s.g.Outcome.Draw || s.g.Outcome.Winner != 0 {
		t.Errorf("the non-active player's start trigger resolves first (FR-TRG-003): player 1 draws from an empty deck and loses, outcome %+v", s.g.Outcome)
	}
}

func TestFR_TRG_003_TwoStartOfTurnTriggersAreOrdered(t *testing.T) {
	// Covers: SVY-R7
	// Options are in collection order: 0 = Scratch, 1 = Draw. Placing option 1
	// first puts Scratch on top, so it resolves first.
	run := func(picks []int) (*scene, []*Instance, []Decision) {
		s := newScene(t)
		s.put(0, ZoneMainTheater, startStructure("Twin",
			startSteps(ability.TurnYours, scratchN(1)), startSteps(ability.TurnYours, drawN(1))))
		s.endTurn()
		cards := s.onTop(0, vanilla("A", "B", "C")...)
		s.script(0, Answer{Picks: picks})
		s.endTurn()
		return s, cards, s.asked(0)
	}
	s, cards, asked := run([]int{1, 0})
	if len(asked) != 1 || asked[0].Kind != DecideOrder || asked[0].Player != 0 || len(asked[0].Options) != 2 {
		t.Fatalf("two simultaneous start triggers of one player are ordered with one DecideOrder of 2 options at turn start (FR-TRG-003, SVY-R7): %+v", asked)
	}
	p := s.g.Players[0]
	if !scratched(p, cards[0]) || !inHand(p, cards[1]) {
		t.Errorf("answer [1 0]: Scratch resolves first and takes the top card, then Draw takes the next (FR-TRG-003)")
	}
	s2, cards2, _ := run([]int{0, 1})
	p2 := s2.g.Players[0]
	if !inHand(p2, cards2[0]) || !scratched(p2, cards2[1]) {
		t.Errorf("answer [0 1]: Draw resolves first and takes the top card, then Scratch takes the next (FR-TRG-003)")
	}
}

func TestFR_TRN_001_MatchEndsDuringStartOfTurn(t *testing.T) {
	s := newScene(t)
	hqs := ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindHeadquarters}
	s.put(0, ZoneMainTheater, startStructure("Bivouac", startSteps(ability.TurnYours, allDamage(20, hqs))))
	s.endTurn()
	p := s.g.Players[0]
	hand, deck := len(p.Hand), len(p.Deck)
	s.endTurn()
	if !s.g.Outcome.Over || s.g.Outcome.Draw || s.g.Outcome.Winner != 0 {
		t.Fatalf("the start trigger defeats the opposing Headquarters, so player 0 wins: outcome %+v", s.g.Outcome)
	}
	if len(p.Hand) != hand || len(p.Deck) != deck {
		t.Errorf("the match ended during the Ready Phase, so no Draw Phase happens (FR-TRN-001): hand %d -> %d, deck %d -> %d",
			hand, len(p.Hand), deck, len(p.Deck))
	}
}
