package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0118 (FR-TGT-003, SVY-R27, FR-MOD-001 §5, SVY-R31, board
// directive #88 part 3): a damage step can name one of its effect's chosen
// slots as the dealer ("Chosen unit deals damage equal to its power to another
// chosen unit"). The chosen asset deals the damage, attributed to it, equal to
// its current effective power as the step resolves. If it is gone, the step
// deals nothing and emits no damage event, and its last-known power is not
// used. Cards are synthetic Maneuvers enacted by player 0. Distinctness between
// two slots is not pinned: the scripts choose distinct units.
//
// Units on every board: X (power 4) is the dealer, Y (power 1) is dealt to, and
// Z (power 2) is the third unit. All have health 30 and belong to player 1.

type dealerBoard struct {
	s       *scene
	x, y, z *Asset
}

func newDealerBoard(t *testing.T) *dealerBoard {
	t.Helper()
	s := newScene(t)
	return &dealerBoard{
		s: s,
		x: s.put(1, ZoneMainTheater, unit("X", 4, 30, Keywords{})),
		y: s.put(1, ZoneMainTheater, unit("Y", 1, 30, Keywords{})),
		z: s.put(1, ZoneMainTheater, unit("Z", 2, 30, Keywords{})),
	}
}

func anyUnit() ability.Target { return tgt(ability.AllegianceAny, ability.KindUnit, false) }

// dealerStep is "the unit chosen for slot dealer deals damage equal to its
// power to the unit chosen for slot target". dealer is the 1-based Step.Dealer.
func dealerStep(dealer, target int) ability.Step {
	return ability.Step{Op: ability.OpDealDamage, Amount: ability.Value{From: ability.ValueSourcePower},
		Subject: ability.SubjectTarget, Target: target, Dealer: dealer}
}

// dealerMode is "Chosen unit deals damage equal to its power to another chosen
// unit": slot 0 deals, slot 1 is dealt to.
func dealerMode() ability.Effect {
	return mode([]ability.Target{anyUnit(), anyUnit()}, dealerStep(1, 1))
}

func returnMode() ability.Effect {
	return mode([]ability.Target{anyUnit()}, ability.Step{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget})
}

func deckTopMode() ability.Effect {
	return mode([]ability.Target{anyUnit()}, ability.Step{Op: ability.OpPutOnDeck, Subject: ability.SubjectTarget})
}

func wantDamageFrom(t *testing.T, g *Game, victim, dealer *Asset, amount int, why string) {
	t.Helper()
	ev := damageTo(g.Log, victim)
	if len(ev) != 1 || ev[0].Amount != amount || ev[0].Source != dealer.ID {
		t.Errorf("%s: damage events to %s = %+v, want one of %d from %s (SVY-R27)", why, victim, ev, amount, dealer)
	}
}

func TestFR_TGT_003_RemovedDealerDealsNothing(t *testing.T) {
	// Covers: SVY-R27, SVY-R25, FR-MOD-001, FR-MNV-001
	b := newDealerBoard(t)
	xInst := b.x.Inst
	inst := b.s.give(0, maneuver("Ambush", 0, 0, ability.Effect{Modes: []ability.Effect{returnMode(), dealerMode()}, ChooseN: 2}))
	b.s.scriptModal(0, nil, nil, []Answer{modeAnswer(0, 1)}, pick(b.x), pick(b.x), pick(b.y))
	b.s.do(enactAction(inst))
	if b.s.alive(b.x) || !inHand(b.s.g.Players[1], xInst) {
		t.Fatalf("precondition: the first mode returned X to its owner's hand")
	}
	if len(nonCombatDamage(b.s.g.Log)) != 0 {
		t.Errorf("a dealer removed by an earlier mode deals nothing and emits no damage event: got %+v (SVY-R27)", nonCombatDamage(b.s.g.Log))
	}
	b.s.wantDamage(b.y, 0)
	if lastDiscard(b.s.g.Players[0]) != inst {
		t.Errorf("the Maneuver still ends in its owner's discard pile (FR-MNV-001)")
	}
	noFormatErrors(t, b.s.g)
}

func TestFR_MOD_001_ChosenModesResolveInPrintedOrder(t *testing.T) {
	// Covers: SVY-R31, SVY-R27, FR-TGT-003
	b := newDealerBoard(t)
	inst := b.s.give(0, maneuver("Ambush", 0, 0, ability.Effect{Modes: []ability.Effect{dealerMode(), deckTopMode()}, ChooseN: 2}))
	// Picked in the reverse order; targets are asked in printed order.
	b.s.scriptModal(0, nil, nil, []Answer{modeAnswer(1, 0)}, pick(b.x), pick(b.y), pick(b.x))
	b.s.do(enactAction(inst))
	wantDamageFrom(t, b.s.g, b.y, b.x, 4, "the first printed mode resolves first, while X is still on the battlefield")
	owner := b.s.g.Players[1]
	if len(owner.Deck) == 0 || owner.Deck[0] != b.x.Inst {
		t.Errorf("the second printed mode then puts X on top of its owner's deck (SVY-R31)")
	}
	if b.s.alive(b.x) {
		t.Errorf("X is no longer on the battlefield after being put on its owner's deck")
	}
	noFormatErrors(t, b.s.g)
}

func TestFR_TGT_003_DealerInALaterModeIsItsOwnSlot(t *testing.T) {
	// Covers: SVY-R27, FR-MOD-001
	b := newDealerBoard(t)
	zInst := b.z.Inst
	inst := b.s.give(0, maneuver("Ambush", 0, 0, ability.Effect{Modes: []ability.Effect{returnMode(), dealerMode()}, ChooseN: 2}))
	b.s.scriptModal(0, nil, nil, []Answer{modeAnswer(0, 1)}, pick(b.z), pick(b.x), pick(b.y))
	b.s.do(enactAction(inst))
	if b.s.alive(b.z) || !inHand(b.s.g.Players[1], zInst) {
		t.Fatalf("precondition: the first mode returned Z to its owner's hand")
	}
	wantDamageFrom(t, b.s.g, b.y, b.x, 4, "the dealer mode's own dealer slot names X, not the earlier mode's Z")
	if got := len(nonCombatDamage(b.s.g.Log)); got != 1 {
		t.Errorf("exactly one damage event, got %d", got)
	}
	noFormatErrors(t, b.s.g)
}

func TestFR_TGT_003_DealerDealsItsCurrentPower(t *testing.T) {
	// Covers: SVY-R27, FR-LAY-001
	b := newDealerBoard(t)
	inst := b.s.give(0, maneuver("Rally Strike", 0, 0, ability.Effect{
		Targets: []ability.Target{anyUnit(), anyUnit()},
		Steps: []ability.Step{
			{Op: ability.OpModifyStat, Amount: ability.Value{N: 2}, Subject: ability.SubjectTarget, Target: 0},
			dealerStep(1, 1),
		},
	}))
	b.s.scriptModal(0, nil, nil, nil, pick(b.x), pick(b.y))
	b.s.do(enactAction(inst))
	wantDamageFrom(t, b.s.g, b.y, b.x, 6, "the dealer's power as the step resolves is its printed 4 plus the +2 from the earlier step")
	noFormatErrors(t, b.s.g)
}

func TestFR_TGT_003_DealerGoneBeforeResolutionDealsNothing(t *testing.T) {
	// Covers: SVY-R27, SVY-R14, SVY-R10, FR-ZON-004
	b := newDealerBoard(t)
	if b.s.g.Power(b.x) <= 0 {
		t.Fatalf("precondition: X has power above 0 when it leaves, so a last-known power would show")
	}
	inst := b.s.give(0, maneuver("Ambush", 0, 0, ability.Effect{
		Targets: []ability.Target{anyUnit(), anyUnit()},
		Steps:   []ability.Step{dealerStep(1, 1), {Op: ability.OpDraw, Amount: ability.Value{N: 1}, IfDone: true}},
	}))
	b.s.scriptModal(0, nil, nil, nil, pick(b.x), pick(b.y))
	r := &recorder{}
	r.onEvent = func(e Event) {
		// After the Maneuver is placed and before it resolves, X leaves.
		if e.Kind == "place" && e.Card == "Ambush" && b.s.alive(b.x) {
			b.s.g.removeFromBattlefield(b.x)
		}
	}
	b.s.g.obs = r
	b.s.onTop(0, vanilla("Top")...)
	p := b.s.g.Players[0]
	hand, deck := len(p.Hand), len(p.Deck)
	b.s.do(enactAction(inst))
	if b.s.alive(b.x) {
		t.Fatalf("precondition: X left the battlefield while the Maneuver waited on the stack")
	}
	if got := len(nonCombatDamage(b.s.g.Log)); got != 0 {
		t.Errorf("a dealer that is gone at resolution deals nothing, with no last-known power: %d damage events, want 0 (SVY-R27)", got)
	}
	b.s.wantDamage(b.y, 0)
	if len(p.Hand) != hand-1 || len(p.Deck) != deck {
		t.Errorf("the \"if you do\" draw does not apply after a step that did nothing: hand %d -> %d (the Maneuver left), deck %d -> %d (SVY-R14)", hand, len(p.Hand), deck, len(p.Deck))
	}
	if lastDiscard(p) != inst {
		t.Errorf("the Maneuver still ends in its owner's discard pile (FR-MNV-001)")
	}
}
