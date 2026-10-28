package engine

import (
	"slices"
	"testing"
)

// Card tests for the Eternal Watchman (T-0082), from the card text in
// data/cards/units-and-structures-HQ.csv.
//
// Eternal Watchman: "Whenever this Headquarters attacks and kills an opposing
// unit or structure, return chosen unit card from your discard pile to your
// hand." Headquarters, 3/20.
//
// Only the declared attacker's lethal combat damage counts (FR-CBT-002 Step 3).
// The trigger is placed after the death sweep (FR-TRG-004), so the defender it
// killed is already in its owner's pile. The target is chosen on placement
// (SVY-R3). "Chosen" without "up to" is exactly one and there is no "you may",
// so the trigger is not optional. The Watchman's printed power is 3.

// wmScene starts a match with the Watchman in seat 0 and the Rager in seat 1.
func wmScene(t *testing.T) *scene { return hqScene(t, watchman, rager) }

// wmAttack has seat 0's Watchman attack def, with the given scripted answers
// for seat 0. It returns the number of decisions seat 0 had been asked when
// the first ability item was placed (-1 if none was placed) and how many
// ability items were placed.
func (s *scene) wmAttack(def *Asset, answers ...Answer) (askedAtPlace, places int) {
	s.script(0, answers...)
	askedAtPlace = -1
	s.ptPlaceHook(func(asked int, _ StackItem) {
		if places == 0 {
			askedAtPlace = asked
		}
		places++
	})
	s.do(s.attack(s.g.Players[0].HQ, def))
	return askedAtPlace, places
}

// Covers: FR-CBT-002, FR-TRG-004, FR-TGT-001
func TestCard_EternalWatchman_KillingAUnitReturnsTheChosenUnitCardToHand(t *testing.T) {
	s := wmScene(t)
	cards := s.inDiscard(0, unit("A", 1, 1, Keywords{}), unit("B", 1, 1, Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Victim", 1, 3, Keywords{}))
	atPlace, places := s.wmAttack(def, Answer{Picks: []int{1}})
	if s.alive(def) {
		t.Fatalf("setup: the Watchman's 3 power should kill the Victim")
	}
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecideTarget {
		t.Fatalf("the controller is asked one Target decision, got %+v", asks)
	}
	if got, want := ppRefs(asks[0]), ppInstIDs(cards...); !slices.Equal(got, want) {
		t.Errorf("both unit cards in the pile are offered: got %v, want %v", got, want)
	}
	if places != 1 || atPlace != 1 {
		t.Errorf("the target is chosen before the trigger is placed (SVY-R3): %d placed, %d asked by then, want 1 and 1", places, atPlace)
	}
	p := s.g.Players[0]
	if ptLast(p.Hand) != cards[1] || inDiscardPile(p, cards[1]) {
		t.Errorf("the chosen card is the last card of its owner's hand and leaves the pile")
	}
	if !inDiscardPile(p, cards[0]) || inHand(p, cards[0]) {
		t.Errorf("the other card stays in the discard pile")
	}
	s.wantSettled("the Watchman's trigger resolved")
}

// Covers: FR-CBT-002, FR-TRG-004, FR-TGT-001
func TestCard_EternalWatchman_KillingAStructureReturnsAUnitCard(t *testing.T) {
	s := wmScene(t)
	cards := s.inDiscard(0, unit("A", 1, 1, Keywords{}), unit("B", 1, 1, Keywords{}))
	def := s.put(1, ZoneMainTheater, structure("Wall", 0, 3, Keywords{}))
	atPlace, _ := s.wmAttack(def, Answer{Picks: []int{0}})
	if s.alive(def) {
		t.Fatalf("setup: the Watchman's 3 power should destroy the Wall")
	}
	if len(s.asked(0)) != 1 || atPlace != 1 {
		t.Errorf("killing a structure asks one Target decision before placement: %d asked, %d at place", len(s.asked(0)), atPlace)
	}
	p := s.g.Players[0]
	if ptLast(p.Hand) != cards[0] || inDiscardPile(p, cards[0]) {
		t.Errorf("the chosen unit card is returned to hand")
	}
	if !inDiscardPile(p, cards[1]) {
		t.Errorf("the other card stays in the discard pile")
	}
	s.wantSettled("the Watchman's trigger resolved")
}

// Covers: FR-TGT-001, FR-TRG-004
func TestCard_EternalWatchman_OnlyYourUnitCardsAreOffered(t *testing.T) {
	s := wmScene(t)
	mine := s.inDiscard(0, unit("A", 1, 1, Keywords{}), structure("Wall", 0, 3, Keywords{}), unit("B", 1, 1, Keywords{}))
	theirs := s.inDiscard(1, unit("Theirs", 1, 1, Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Victim", 1, 3, Keywords{}))
	victim := def.Inst
	s.wmAttack(def, Answer{Picks: []int{0}})
	asks := s.asked(0)
	if len(asks) != 1 {
		t.Fatalf("one Target decision, got %d", len(asks))
	}
	if got, want := ppRefs(asks[0]), ppInstIDs(mine[0], mine[2]); !slices.Equal(got, want) {
		t.Errorf("only unit cards in your own pile are offered, not a structure card, the opponent's card or the killed Victim: got %v, want %v", got, want)
	}
	if !inDiscardPile(s.g.Players[1], theirs[0]) || !inDiscardPile(s.g.Players[1], victim) {
		t.Errorf("the opponent's pile is untouched")
	}
	if !inDiscardPile(s.g.Players[0], mine[1]) {
		t.Errorf("the structure card stays in the pile")
	}
}

// Covers: FR-CBT-002
func TestCard_EternalWatchman_ADefenderThatSurvivesReturnsNothing(t *testing.T) {
	s := wmScene(t)
	cards := s.inDiscard(0, unit("A", 1, 1, Keywords{}), unit("B", 1, 1, Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Tough", 1, 9, Keywords{}))
	handBefore := len(s.g.Players[0].Hand)
	_, places := s.wmAttack(def)
	if !s.alive(def) {
		t.Fatalf("setup: the Tough defender should survive")
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("no decision is asked when nothing was killed: %d", n)
	}
	if places != 0 {
		t.Errorf("the trigger does not fire without a kill: %d items placed", places)
	}
	p := s.g.Players[0]
	if len(p.Hand) != handBefore || !inDiscardPile(p, cards[0]) || !inDiscardPile(p, cards[1]) {
		t.Errorf("the hand and the pile are unchanged")
	}
}

// Covers: FR-CBT-002
func TestCard_EternalWatchman_ARetaliationKillReturnsNothing(t *testing.T) {
	s := hqScene(t, rager, watchman)
	cards := s.inDiscard(1, unit("A", 1, 1, Keywords{}), unit("B", 1, 1, Keywords{}))
	att := s.put(0, ZoneMainTheater, unit("Raider", 1, 3, Keywords{}))
	s.script(1)
	handBefore := len(s.g.Players[1].Hand)
	s.do(s.attack(att, s.g.Players[1].HQ))
	if s.alive(att) {
		t.Fatalf("setup: the Watchman's retaliation of 3 should kill the Raider")
	}
	if n := len(s.asked(1)); n != 0 {
		t.Errorf("a retaliation kill does not fire the Watchman, so no decision is asked: %d (FR-CBT-002 Step 3)", n)
	}
	p := s.g.Players[1]
	if len(p.Hand) != handBefore || !inDiscardPile(p, cards[0]) || !inDiscardPile(p, cards[1]) {
		t.Errorf("nothing moves after a retaliation kill")
	}
	s.wantSettled("the retaliation kill")
}

// Covers: FR-TGT-003, SVY-R16
func TestCard_EternalWatchman_NoUnitCardInYourPileReturnsNothing(t *testing.T) {
	s := wmScene(t)
	wall := s.inDiscard(0, structure("Wall", 0, 3, Keywords{}))
	theirs := s.inDiscard(1, unit("Theirs", 1, 1, Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Victim", 1, 3, Keywords{}))
	handBefore := len(s.g.Players[0].Hand)
	_, places := s.wmAttack(def)
	if s.alive(def) {
		t.Fatalf("setup: the Victim should die")
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("no decision is asked when no unit card is a legal choice: %d", n)
	}
	if places != 1 {
		t.Errorf("the trigger is still placed and resolves: %d items placed, want 1", places)
	}
	p := s.g.Players[0]
	if len(p.Hand) != handBefore || !inDiscardPile(p, wall[0]) || !inDiscardPile(s.g.Players[1], theirs[0]) {
		t.Errorf("nothing moves when nothing can be chosen")
	}
	s.wantSettled("the empty trigger resolved")
}

// Covers: FR-MOD-001, SVY-R16
func TestCard_EternalWatchman_TimeoutReturnsTheTopUnitCard(t *testing.T) {
	s := wmScene(t)
	cards := s.inDiscard(0, unit("A", 1, 1, Keywords{}), unit("B", 1, 1, Keywords{}), unit("C", 1, 1, Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Victim", 1, 3, Keywords{}))
	s.wmAttack(def, Answer{Timeout: true})
	if n := len(s.asked(0)); n != 1 {
		t.Fatalf("three matches are asked once: %d", n)
	}
	p := s.g.Players[0]
	if !inHand(p, cards[2]) {
		t.Errorf("a timeout takes the top card of the pile, C (FR-MOD-001 sec. 6)")
	}
	if inHand(p, cards[0]) || inHand(p, cards[1]) {
		t.Errorf("only the top card is returned, not the first option (A)")
	}
}
