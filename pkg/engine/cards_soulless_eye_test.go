package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0322 (FR-TRN-003, FR-SET-004, FR-CAP-001, FR-CBT-001, FR-CBT-002,
// FR-TRG-003, FR-ZON-004, FR-WIN-001): Visionary's Eye (Soulless unit,
// Cybernetic, Scout, R1 T0, Reserves only, 1/1), loaded from data/cards:
// "Whenever this unit attacks, draw a card." Player 0 controls it and is the
// active player unless a case says otherwise. Rulings: SVY-R8, SVY-R49.

const eyeName = "Visionary's Eye"

// eyeCounts returns the sizes of player p's hand and deck.
func eyeCounts(s *scene, p PlayerID) (hand, deck int) {
	return len(s.g.Players[p].Hand), len(s.g.Players[p].Deck)
}

// eyeWantCounts fails unless player p's hand and deck have the given sizes.
func eyeWantCounts(s *scene, p PlayerID, hand, deck int, why string) {
	s.t.Helper()
	h, d := eyeCounts(s, p)
	if h != hand || d != deck {
		s.t.Errorf("%s: player %d hand %d deck %d, want hand %d deck %d", why, p, h, d, hand, deck)
	}
}

// eyeInMain puts a Readied real Eye into player 0's Main Theater.
func eyeInMain(s *scene) *Asset {
	return s.put(0, ZoneMainTheater, s.card(eyeName))
}

func TestCard_VisionarysEye_EntersReservesAsOneOne(t *testing.T) {
	// Covers: FR-TRN-003, FR-CAP-001, FR-SET-004
	s := newScene(t)
	c := s.card(eyeName)
	inst := s.give(0, c)
	s.setPool(0, 1, 0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves || a.Controller != 0 {
		t.Fatalf("the Eye is on the battlefield in player 0's Reserves after deploying (FR-TRN-003)")
	}
	if p := a.Printed(); p.Power != 1 || p.Health != 1 {
		t.Errorf("the Eye is a 1/1: got %d/%d", p.Power, p.Health)
	}
	if a.Exhausted {
		t.Errorf("a newly deployed unit enters Readied (FR-SET-004)")
	}
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
		t.Errorf("the Eye costs R1: pool R%d T%d, want R0 T0 (FR-TRN-003)", p.Req, p.Tac)
	}

	s2 := newScene(t)
	c2 := s2.card(eyeName)
	inst2 := s2.give(0, c2)
	s2.setPool(0, 1, 0)
	s2.reject(Action{Kind: ActDeploy, Player: 0, Card: inst2.ID, Zone: ZoneMainTheater})
	if !inHand(s2.g.Players[0], inst2) || assetOf(s2.g, c2) != nil {
		t.Errorf("deploying the Eye to the Main Theater is rejected and the card stays in hand (FR-CAP-001)")
	}

	s3 := newScene(t)
	c3 := s3.card(eyeName)
	inst3 := s3.give(0, c3)
	s3.setPool(0, 0, 0)
	s3.reject(Action{Kind: ActDeploy, Player: 0, Card: inst3.ID, Zone: ZoneReserves})
	if !inHand(s3.g.Players[0], inst3) || assetOf(s3.g, c3) != nil {
		t.Errorf("with R0 the Eye (cost R1) cannot be deployed (FR-TRN-003)")
	}
}

func TestCard_VisionarysEye_CannotAttackFromReserves(t *testing.T) {
	// Covers: FR-CBT-001, FR-CAP-001
	s := newScene(t)
	eye := s.put(0, ZoneReserves, s.card(eyeName))
	hand, deck := eyeCounts(s, 0)
	s.reject(s.attack(eye, s.g.Players[1].HQ))
	eyeWantCounts(s, 0, hand, deck, "a rejected attack from Reserves draws nothing (FR-CBT-001)")
	if eye.Exhausted {
		t.Errorf("the rejected attack leaves the Eye Readied (FR-CAP-001)")
	}
}

func TestCard_VisionarysEye_MarchesThenAttacksAndDraws(t *testing.T) {
	// Covers: FR-SET-004, FR-CAP-001
	s := newScene(t)
	c := s.card(eyeName)
	inst := s.give(0, c)
	s.setPool(0, 1, 0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	eye := assetOf(s.g, c)
	s.do(Action{Kind: ActMarch, Player: 0, Asset: eye.ID})
	if eye.Zone != ZoneMainTheater || !eye.Exhausted {
		t.Fatalf("the Eye may March the turn it deploys (FR-SET-004) and is then Exhausted in the Main Theater: zone %s exhausted %v", eye.Zone, eye.Exhausted)
	}
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 30, ability.Keywords{}))
	s.endTurn()
	s.endTurn()
	hand, deck := eyeCounts(s, 0)
	s.do(s.attack(eye, def))
	eyeWantCounts(s, 0, hand+1, deck-1, "the Eye's attack draws exactly one card (FR-CBT-002)")
}

func TestCard_VisionarysEye_AttackDrawsOneCard(t *testing.T) {
	// Covers: FR-CBT-002
	s := newScene(t)
	top := s.onTop(0, vanilla("Drawn")...)
	eye := eyeInMain(s)
	s.script(0)
	s.script(1)
	hq1 := s.g.Players[1].HQ
	h0, d0 := eyeCounts(s, 0)
	h1, d1 := eyeCounts(s, 1)
	s.do(s.attack(eye, hq1))
	eyeWantCounts(s, 0, h0+1, d0-1, "the attack trigger draws one card for the Eye's controller")
	if !inHand(s.g.Players[0], top[0]) {
		t.Errorf("the card drawn is the top card of player 0's deck")
	}
	if n := len(s.asked(0)) + len(s.asked(1)); n != 0 {
		t.Errorf("the draw is mandatory and has no targets, so no decision is asked: asked %d", n)
	}
	s.wantDamage(hq1, 1)
	if s.alive(eye) || !inDiscardPile(s.g.Players[0], eye.Inst) {
		t.Errorf("the Eye takes 3 retaliation and its card is in player 0's discard pile")
	}
	eyeWantCounts(s, 1, h1, d1, "the opponent's hand and deck are unchanged")
}

func TestCard_VisionarysEye_DrawResolvesBeforeCombatDamage(t *testing.T) {
	// Covers: FR-CBT-002, FR-WIN-001
	s := newScene(t)
	s.g.Players[0].Deck = nil
	eye := eyeInMain(s)
	s.do(s.attack(eye, s.g.Players[1].HQ))
	if o := s.g.Outcome; !o.Over || o.Draw || o.Winner != 1 {
		t.Errorf("a draw from an empty deck in the Step 2 trigger loses player 0 at that sweep (FR-WIN-001): outcome %+v", o)
	}
	s.wantDamage(s.g.Players[1].HQ, 0)
}

func TestCard_VisionarysEye_DrawsEvenWhenItDiesInTheTrade(t *testing.T) {
	// Covers: FR-CBT-002
	s := newScene(t)
	eye := eyeInMain(s)
	def := s.put(1, ZoneMainTheater, unit("Trader", 1, 1, ability.Keywords{}))
	h, d := eyeCounts(s, 0)
	s.do(s.attack(eye, def))
	if s.alive(eye) || !inDiscardPile(s.g.Players[0], eye.Inst) {
		t.Errorf("the Eye is destroyed and its card is in player 0's discard pile")
	}
	if s.alive(def) || !inDiscardPile(s.g.Players[1], def.Inst) {
		t.Errorf("the defender is destroyed and its card is in player 1's discard pile")
	}
	eyeWantCounts(s, 0, h+1, d-1, "the draw happened before the trade (FR-CBT-002 Step 2)")
}

func TestCard_VisionarysEye_ZeroPowerAttackStillDraws(t *testing.T) {
	// Covers: FR-CBT-001
	s := newScene(t)
	eye := eyeInMain(s)
	eye.Modifiers = append(eye.Modifiers, Modifier{Layer: 5, Power: -1})
	h, d := eyeCounts(s, 0)
	s.do(s.attack(eye, s.g.Players[1].HQ))
	eyeWantCounts(s, 0, h+1, d-1, "a zero-power attack is allowed and fires attack triggers (FR-CBT-001)")
	s.wantDamage(s.g.Players[1].HQ, 0)
}

func TestCard_VisionarysEye_KilledBeforeItsTriggerStillDraws(t *testing.T) {
	// Covers: FR-TRG-003, FR-ZON-004, FR-CBT-002, SVY-R49, SVY-R8
	s := newScene(t)
	eye := eyeInMain(s)
	defile := s.put(1, ZoneMainTheater, s.card("Blind Defile"))
	s.script(0)
	s.script(1, yes)
	h, d := eyeCounts(s, 0)
	s.do(s.attack(eye, s.g.Players[1].HQ))
	if s.alive(defile) || s.alive(eye) {
		t.Fatalf("setup: the Defile's trigger kills the Defile and the Eye (SVY-R49): Defile alive %v, Eye alive %v", s.alive(defile), s.alive(eye))
	}
	eyeWantCounts(s, 0, h+1, d-1, "the Eye's trigger still resolves with its last-known state (FR-ZON-004)")
	s.wantDamage(s.g.Players[1].HQ, 0)
	if !inDiscardPile(s.g.Players[0], eye.Inst) {
		t.Errorf("the Eye's card is in player 0's discard pile")
	}
}

func TestCard_VisionarysEye_EachAttackDrawsAgain(t *testing.T) {
	// Covers: FR-CBT-002
	s := newScene(t)
	eye := eyeInMain(s)
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 30, ability.Keywords{}))
	h, d := eyeCounts(s, 0)
	s.do(s.attack(eye, def))
	eyeWantCounts(s, 0, h+1, d-1, "the first attack draws one card")
	s.endTurn()
	s.endTurn()
	h, d = eyeCounts(s, 0)
	s.do(s.attack(eye, def))
	eyeWantCounts(s, 0, h+1, d-1, "the second attack draws one card more than the turn's draws")
}

func TestCard_VisionarysEye_DefendingDoesNotDraw(t *testing.T) {
	// Covers: FR-CBT-002
	s := newScene(t)
	eye := eyeInMain(s)
	s.endTurn()
	hitter := s.put(1, ZoneMainTheater, unit("Hitter", 1, 30, ability.Keywords{}))
	h, d := eyeCounts(s, 0)
	s.do(s.attack(hitter, eye))
	eyeWantCounts(s, 0, h, d, "only attacking triggers the Eye")
}

func TestCard_VisionarysEye_OnlyItsOwnAttackDraws(t *testing.T) {
	// Covers: FR-CBT-002
	s := newScene(t)
	eye1 := eyeInMain(s)
	eyeInMain(s)
	other := s.put(0, ZoneMainTheater, unit("Other", 1, 30, ability.Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 30, ability.Keywords{}))
	h, d := eyeCounts(s, 0)
	s.do(s.attack(other, def))
	eyeWantCounts(s, 0, h, d, "another unit's attack draws nothing")
	s.do(s.attack(eye1, def))
	eyeWantCounts(s, 0, h+1, d-1, "one Eye's attack draws exactly one card; the other Eye does not trigger")
}
