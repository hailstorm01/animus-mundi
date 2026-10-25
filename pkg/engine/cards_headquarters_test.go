package engine

import "testing"

// Card tests for Headquarters abilities (T-0055), from the card text in
// data/cards/units-and-structures-HQ.csv.
//
// Animoceps Rager: "Whenever this Headquarters is dealt damage, Scratch 1.
// This ability triggers only once each turn." Headquarters, 3/20.
//
// The damage-taken moment is FR-REP-004 Stage 5; the once-each-turn limit
// resets at the start of every turn, either player's (FR-TRN-001). The Rager
// has no Unbowed, so these tests do not cite FR-KWD-002a. Armor and Resistant
// reducing the damage to zero are not tested: the starter set gives no
// Headquarters either keyword and the threshold is an unruled default
// (survey section 4). Lethal damage is not tested: the match ends at the
// sweep, so the trigger's resolution cannot be observed.

// hqScene starts a match like newScene, but with the named real
// Headquarters cards in seats 0 and 1. It sets state and asserts nothing.
func hqScene(t *testing.T, hq0, hq1 string) *scene {
	t.Helper()
	var hqs [2]*Card
	for i, name := range []string{hq0, hq1} {
		c, err := realCard(name)
		if err != nil {
			t.Fatal(err)
		}
		hqs[i] = c
	}
	g, err := NewGame(Config{
		Seed:      1,
		Decks:     [2]DeckList{{HQ: hqs[0], Main: filler(20)}, {HQ: hqs[1], Main: filler(20)}},
		PlayFirst: func(decider PlayerID) bool { return decider == 0 },
	})
	if err != nil {
		t.Fatal(err)
	}
	return &scene{t: t, g: g}
}

// wantSettled fails unless the stack is empty and the invariants hold.
func (s *scene) wantSettled(why string) {
	s.t.Helper()
	if n := len(s.g.Stack); n != 0 {
		s.t.Errorf("stack has %d items, want 0: %s", n, why)
	}
	if err := s.g.CheckInvariants(); err != nil {
		s.t.Errorf("invariants after %s: %v", why, err)
	}
}

const (
	rager    = "Animoceps Rager"
	watchman = "Eternal Watchman"
)

// --- firing ---

// Covers: FR-REP-004
func TestCard_AnimocepsRager_CombatDamageScratchesOneForItsController(t *testing.T) {
	s := hqScene(t, watchman, rager)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	r := s.g.Players[1].HQ
	s.do(s.attack(att, r))
	s.wantDamage(r, 2)
	s.wantScratches(1, 1, "combat damage to the Rager Scratches 1 for its controller, the defender, on the opponent's turn (FR-REP-004 Stage 5)")
	s.wantScratches(0, 0, "the attacker's controller gets no Scratch")
	s.wantSettled("the Scratch resolved through the stack")
}

// Covers: FR-REP-004, FR-TGT-001
func TestCard_AnimocepsRager_EffectDamageScratchesOne(t *testing.T) {
	s := hqScene(t, watchman, rager)
	r := s.g.Players[1].HQ
	s.scriptRefs(0, nil, r)
	deploySappers(s)
	s.wantDamage(r, 1)
	s.wantScratches(1, 1, "effect damage (Chainbreaker Sappers) to the Rager Scratches 1 for its controller (FR-REP-004 Stage 5)")
	s.wantScratches(0, 0, "the Sappers' controller gets no Scratch")
	s.wantSettled("the Scratch resolved through the stack")
}

// Covers: FR-REP-004, FR-CAP-001
func TestCard_AnimocepsRager_RetaliationDamageScratchesOneOnItsOwnTurn(t *testing.T) {
	s := hqScene(t, rager, watchman)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 2, 9, Keywords{}))
	r := s.g.Players[0].HQ
	s.do(s.attack(r, foe))
	s.wantDamage(r, 2)
	if !r.Exhausted {
		t.Errorf("a Headquarters that attacks is Exhausted (FR-CAP-001)")
	}
	s.wantScratches(0, 1, "retaliation damage to the attacking Rager Scratches 1 for its controller; the Scratch does not need it Readied (FR-REP-004 Stage 5)")
	s.wantScratches(1, 0, "the defender's controller gets no Scratch")
	s.wantSettled("the Scratch resolved through the stack")
}

// --- once each turn ---

// Covers: FR-REP-004
func TestCard_AnimocepsRager_TwoAttacksInOneTurnScratchOnlyOnce(t *testing.T) {
	s := hqScene(t, watchman, rager)
	h := s.hitters(2)
	r := s.g.Players[1].HQ
	s.do(s.attack(h[0], r))
	s.wantScratches(1, 1, "the first damage instance fires the Rager")
	s.do(s.attack(h[1], r))
	s.wantDamage(r, 2)
	s.wantScratches(1, 1, "the second damage instance in the same turn does not fire it: it triggers only once each turn (FR-TRN-001)")
	s.wantSettled("two attacks")
}

// Covers: FR-REP-004
func TestCard_AnimocepsRager_AttackThenSappersInOneTurnScratchOnlyOnce(t *testing.T) {
	s := hqScene(t, watchman, rager)
	h := s.hitters(1)
	r := s.g.Players[1].HQ
	s.do(s.attack(h[0], r))
	s.scriptRefs(0, nil, r)
	deploySappers(s)
	s.wantDamage(r, 2)
	s.wantScratches(1, 1, "combat damage then effect damage in one turn Scratch only once (FR-TRN-001)")
	s.wantSettled("attack then Sappers")
}

// Covers: FR-REP-004
func TestCard_AnimocepsRager_LimitResetsAtTheStartOfEveryTurn(t *testing.T) {
	s := hqScene(t, watchman, rager)
	h := s.hitters(2)
	foe := s.put(0, ZoneMainTheater, unit("Foe", 1, 9, Keywords{}))
	r := s.g.Players[1].HQ
	s.do(s.attack(h[0], r)) // turn 1, the opponent's turn
	s.wantScratches(1, 1, "damaged on turn 1 Scratches")
	s.endTurn()
	s.do(s.attack(r, foe)) // turn 2, its own turn: retaliation damage
	s.wantScratches(1, 2, "damaged again on the next turn, the limit has reset (FR-TRN-001)")
	s.endTurn()
	s.do(s.attack(h[1], r)) // turn 3, the opponent's turn again
	s.wantScratches(1, 3, "the limit resets at the start of the opponent's turn too (FR-TRN-001)")
	s.wantSettled("three turns")
}

// --- not firing ---

// Covers: FR-REP-004
func TestCard_AnimocepsRager_DamageToAnotherAssetDoesNotScratch(t *testing.T) {
	s := hqScene(t, rager, rager)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Bystander", 1, 9, Keywords{}))
	s.do(s.attack(att, def))
	s.wantDamage(def, 2)
	s.wantScratches(0, 0, "damage to another unit does not fire the Rager (card text: this Headquarters)")
	s.wantScratches(1, 0, "damage to another unit does not fire the other player's Rager")
}

// Covers: FR-REP-004
func TestCard_AnimocepsRager_DamageToTheOtherHeadquartersDoesNotScratchForIt(t *testing.T) {
	s := hqScene(t, rager, watchman)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	s.do(s.attack(att, s.g.Players[1].HQ))
	s.wantDamage(s.g.Players[1].HQ, 2)
	s.wantScratches(0, 0, "damage to the opposing Headquarters does not fire the Rager (card text: this Headquarters)")
	s.wantScratches(1, 0, "the Eternal Watchman has no damage trigger")
}

// Covers: FR-REP-004
func TestCard_AnimocepsRager_DamageToOpposingRagerDoesNotScratchForTheAttacker(t *testing.T) {
	s := hqScene(t, rager, rager)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	s.do(s.attack(att, s.g.Players[1].HQ))
	s.wantScratches(1, 1, "the damaged Rager's controller Scratches 1")
	s.wantScratches(0, 0, "the undamaged Rager does not fire")
}

// Covers: FR-REP-004
func TestCard_EternalWatchman_DamageDoesNotScratchForAnyone(t *testing.T) {
	s := hqScene(t, rager, watchman)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	s.do(s.attack(att, s.g.Players[1].HQ))
	s.wantDamage(s.g.Players[1].HQ, 2)
	s.wantScratches(1, 0, "the Eternal Watchman has no damage trigger")
	s.wantScratches(0, 0, "no Scratch for the attacker's controller either")
}
