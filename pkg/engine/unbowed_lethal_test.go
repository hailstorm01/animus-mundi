package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for ruling SVY-R13 (FRD 01 v3.10, directive #46): lethal Unbowed is
// buffered like any trigger and placed after the death sweep, in the same APNAP
// window (FR-TRG-003) as the death triggers from the same event. The kernel is
// expected to pass these when they are written (T-0050). Cards are synthetic.
// The window is observed through the order decision offered, which player's
// trigger resolves first, and a self-aimed step that fizzles.

// bowedMortal is a unit printing an Unbowed trigger and then a death trigger.
func bowedMortal(name string, power, health int, u, k ability.Trigger) *Card {
	return bowed(name, power, health, Keywords{}, u, k)
}

// Covers: FR-TRG-003, FR-TRG-004
func TestFR_KWD_002a_LethalUnbowedAndDeathTriggerShareOneOrderingWindow(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, bowedMortal("Martyr", 1, 1, unbowed(nil, scratchN(1)), killedTrigger(drawN(1))))
	hand := len(s.g.Players[1].Hand)
	s.script(0)
	s.script(1, Answer{Picks: []int{0, 1}})
	s.do(s.attack(att, def))
	if s.alive(def) {
		t.Fatalf("precondition: the defender is destroyed by 3 damage on 1 Health")
	}
	asked := s.asked(1)
	if len(asked) != 1 || asked[0].Kind != DecideOrder || asked[0].Player != 1 || len(asked[0].Options) != 2 {
		t.Errorf("the dead asset's Unbowed and its death trigger are placed in one window, so its controller orders both with one question (FR-KWD-002a SVY-R13, FR-TRG-003): %+v", asked)
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("the opponent has no trigger in the window and is asked nothing: asked %d times", n)
	}
	s.wantScratches(1, 1, "the lethal Unbowed Scratch 1 resolves from the discard pile (FR-KWD-002a)")
	if got := len(s.g.Players[1].Hand) - hand; got != 1 {
		t.Errorf("the death trigger's Draw 1 resolves too: drew %d, want 1", got)
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("both triggers resolved: %d items left", len(s.g.Stack))
	}
}

// Covers: FR-TRG-003
func TestFR_TRG_003_LethalUnbowedOfTheNonActivePlayerResolvesBeforeTheActivePlayersDeathTrigger(t *testing.T) {
	// Both units die in the same combat. The active player's death trigger and
	// the non-active player's lethal Unbowed are in one APNAP window, so the
	// non-active player's resolves first (Draw 1 from an empty deck loses): player
	// 1 loses. Had the Unbowed been placed before the death sweep, player 0's
	// death trigger would sit on top and player 0 would lose instead.
	s := newScene(t)
	s.g.Players[0].Deck = nil
	s.g.Players[1].Deck = nil
	att := s.put(0, ZoneMainTheater, mortal("Attacker", 1, 1, killedTrigger(drawN(1))))
	def := s.put(1, ZoneMainTheater, bowed("Defender", 1, 1, Keywords{}, unbowed(nil, drawN(1))))
	s.do(s.attack(att, def))
	if s.alive(att) || s.alive(def) {
		t.Fatalf("precondition: both units are destroyed (attacker alive=%v, defender alive=%v)", s.alive(att), s.alive(def))
	}
	if o := s.g.Outcome; !o.Over || o.Draw || o.Winner != 0 {
		t.Errorf("the non-active player's lethal Unbowed resolves first in the shared window (FR-TRG-003, SVY-R13): player 1 loses, outcome %+v", o)
	}
}

// Covers: FR-TRG-003
func TestFR_TRG_003_ActivePlayersLethalUnbowedResolvesAfterTheNonActivePlayersDeathTrigger(t *testing.T) {
	s := newScene(t)
	s.g.Players[0].Deck = nil
	s.g.Players[1].Deck = nil
	att := s.put(0, ZoneMainTheater, bowed("Attacker", 1, 1, Keywords{}, unbowed(nil, drawN(1))))
	def := s.put(1, ZoneMainTheater, mortal("Defender", 1, 1, killedTrigger(drawN(1))))
	s.do(s.attack(att, def))
	if s.alive(att) || s.alive(def) {
		t.Fatalf("precondition: both units are destroyed (attacker alive=%v, defender alive=%v)", s.alive(att), s.alive(def))
	}
	if o := s.g.Outcome; !o.Over || o.Draw || o.Winner != 0 {
		t.Errorf("the non-active player's death trigger resolves before the active player's lethal Unbowed (FR-TRG-003, SVY-R13): player 1 loses, outcome %+v", o)
	}
}

// Covers: FR-TGT-003, FR-ZON-004
func TestFR_KWD_002a_LethalUnbowedSelfHealFizzlesAndTheAssetIsDestroyedAnyway(t *testing.T) {
	// 3 damage on 2 Health. Resolving Heal 3 before the death sweep would leave the
	// asset alive; resolved after the sweep it fizzles, the asset stays in the
	// discard pile and its death trigger still resolves.
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, bowedMortal("Dancer", 1, 2, unbowed(nil, healSelf(3)), killedTrigger(scratchN(1))))
	inst := def.Inst
	s.script(1, Answer{Picks: []int{0, 1}})
	s.do(s.attack(att, def))
	if s.alive(def) {
		t.Fatalf("a heal in the Unbowed trigger cannot save the asset: it is destroyed at the death sweep (FR-TRG-004, SVY-R13)")
	}
	if !inZone(s.g.Players[1].Discard, inst) {
		t.Errorf("the self-aimed heal fizzles and the card stays in its owner's discard pile (FR-KWD-002a, FR-TGT-003)")
	}
	s.wantScratches(1, 1, "the death trigger from the same event resolves (FR-TRG-003)")
	if err := s.g.CheckInvariants(); err != nil {
		t.Errorf("state after the fizzled step: %v", err)
	}
}
