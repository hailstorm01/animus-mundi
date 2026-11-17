package engine

import (
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0267: HookEntersDiscard, a pile trigger that fires whenever its
// card is put into its owner's discard pile, from any zone and by any route
// (FR-ZON-004 Enter-Discard Triggers, FR-ZON-005, FR-ZON-002, FR-ZON-003,
// FR-ENT-102, FR-TRG-003, FR-MNV-001, FR-KWD-002d; SVY-R7, survey gap G1).
// Cards are synthetic. Player 0 is the active player.

// entersTrigger is one mandatory "When this card is put into your discard
// pile, <steps>" ability.
func entersTrigger(steps ...ability.Step) ability.Trigger {
	return ability.Trigger{Hook: ability.HookEntersDiscard, ActiveZone: ability.ZoneDiscard,
		Effect: ability.Effect{Steps: steps}}
}

// entersHealTrigger is "When this card is put into your discard pile, chosen
// asset heals 2".
func entersHealTrigger() ability.Trigger {
	return ability.Trigger{Hook: ability.HookEntersDiscard, ActiveZone: ability.ZoneDiscard,
		Effect: ability.Effect{
			Targets: []ability.Target{{}},
			Steps:   []ability.Step{{Op: ability.OpHeal, Amount: ability.Value{N: 2}, Subject: ability.SubjectTarget, Target: 0}},
		}}
}

// edMonk is a unit printing one enter-discard "draw 1" trigger.
func edMonk(name string) *Card {
	return mortal(name, 1, 1, entersTrigger(drawN(1)))
}

// enactAlliedStep gives player 0 an empty hand and a costless Maneuver whose
// one target is an allied unit and whose steps are st, and enacts it.
func (s *scene) enactAlliedStep(st ...ability.Step) {
	s.t.Helper()
	s.g.Players[0].Hand = nil
	m := s.give(0, maneuver("Purge", 0, 0, ability.Effect{
		Targets: []ability.Target{{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}}},
		Steps:   st,
	}))
	s.do(enactAction(m))
}

func killTarget() ability.Step {
	return ability.Step{Op: ability.OpKill, Subject: ability.SubjectTarget, Target: 0}
}

// firstPick answers every target decision with its first option and times
// out everything else.
func firstPick(d Decision) Answer {
	if d.Kind == DecideTarget {
		return Answer{Picks: []int{0}}
	}
	return yesAndTimeouts(d)
}

func TestFR_ZON_004_EntersDiscardFiresFromHandDiscard(t *testing.T) {
	s := newScene(t)
	s.byKind(0, yesAndTimeouts)
	s.byKind(1, yesAndTimeouts)
	inst := s.enactWith([]*Card{edMonk("Monk")}, discardN(1))[0]
	p := s.g.Players[0]
	if !inZone(p.Discard, inst) {
		t.Fatalf("the discarded card is in its owner's discard pile (FR-ZON-004)")
	}
	if got := pileLogPlaces(s, "Monk"); got != 1 {
		t.Errorf("a card discarded from hand fires its enter-discard trigger once, got %d placements (FR-ZON-004)", got)
	}
	if len(p.Hand) != 1 {
		t.Errorf("the resolved trigger drew 1 card into an empty hand, hand has %d (FR-ZON-004)", len(p.Hand))
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty once the trigger resolves: %d items", len(s.g.Stack))
	}
}

func TestFR_ZON_004_EntersDiscardFiresOnDestruction(t *testing.T) {
	// (a) combat
	s := newScene(t)
	s.byKind(0, yesAndTimeouts)
	s.byKind(1, yesAndTimeouts)
	before := len(s.g.Players[1].Hand)
	inst := killOnP1(s, edMonk("Monk"))
	p1 := s.g.Players[1]
	if !inZone(p1.Discard, inst) {
		t.Fatalf("the destroyed unit is in its owner's discard pile (FR-ZON-002)")
	}
	if got := pileLogPlaces(s, "Monk"); got != 1 {
		t.Errorf("a unit destroyed in combat fires its enter-discard trigger once, got %d placements (FR-ZON-004)", got)
	}
	if got := len(p1.Hand) - before; got != 1 {
		t.Errorf("player 1 drew %d cards from the trigger, want 1 (FR-ZON-004)", got)
	}

	// (b) a kill step
	s = newScene(t)
	s.byKind(0, firstPick)
	s.byKind(1, yesAndTimeouts)
	a := s.put(0, ZoneMainTheater, edMonk("Monk"))
	s.enactAlliedStep(killTarget())
	p0 := s.g.Players[0]
	if s.alive(a) || !inZone(p0.Discard, a.Inst) {
		t.Fatalf("the killed unit is in its owner's discard pile (FR-ZON-002)")
	}
	if got := pileLogPlaces(s, "Monk"); got != 1 {
		t.Errorf("a unit killed by a step fires its enter-discard trigger once, got %d placements (FR-ZON-004)", got)
	}
	if len(p0.Hand) != 1 {
		t.Errorf("player 0 drew %d cards from the trigger, want 1 (FR-ZON-004)", len(p0.Hand))
	}
}

func TestFR_ZON_004_EntersDiscardTriggerChoosesItsTarget(t *testing.T) {
	// Covers: FR-KWD-002d, FR-TGT-001
	s := newScene(t)
	s.byKind(0, yesAndTimeouts)
	hurt := s.put(1, ZoneMainTheater, unit("Hurt", 1, 5, Keywords{}))
	hurt.Damage = 3
	s.byKind(1, func(d Decision) Answer {
		if d.Kind == DecideTarget {
			for i, o := range d.Options {
				if o.Ref == hurt.ID {
					return Answer{Picks: []int{i}}
				}
			}
			t.Errorf("the damaged unit is among the heal's target options: %+v", d.Options)
		}
		return yesAndTimeouts(d)
	})
	killOnP1(s, mortal("Monk", 1, 1, entersHealTrigger()))
	if got := askedKind(s, 1, DecideTarget); len(got) != 1 || got[0].Player != 1 {
		t.Fatalf("the card's owner, player 1, picks the heal's target once: %+v (FR-TGT-001)", got)
	}
	s.wantDamage(hurt, 1)
}

func TestFR_ZON_005_EntersDiscardFiresForADestroyedScratchCompany(t *testing.T) {
	// Covers: FR-ZON-004
	s := newScene(t)
	s.byKind(0, yesAndTimeouts)
	s.byKind(1, yesAndTimeouts)
	down := s.putScratch(1)
	down.Inst.Card = mortal("Hidden Monk", 1, 1, killedTrigger(drawN(1)), entersTrigger(drawN(1)))
	s.moveTo(down, ZoneMainTheater)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 5, 9, Keywords{}))
	before := len(s.g.Players[1].Hand)
	s.do(s.attack(att, down))
	if s.alive(down) {
		t.Fatalf("precondition: the Scratch Company is destroyed by the attack")
	}
	p1 := s.g.Players[1]
	if !inZone(p1.Discard, down.Inst) {
		t.Fatalf("the destroyed Scratch Company's card is in its owner's discard pile (FR-ZON-005)")
	}
	if got := pileLogPlaces(s, "Hidden Monk"); got != 1 {
		t.Errorf("only the enter-discard trigger is placed for the face-up card, got %d placements (FR-ZON-004)", got)
	}
	if got := len(p1.Hand) - before; got != 1 {
		t.Errorf("player 1 drew %d cards, want exactly 1: the face-down death trigger fires nothing, the enter-discard one does (FR-ZON-004)", got)
	}
}

func TestFR_ZON_004_EntersDiscardFiresWhenAManeuverResolves(t *testing.T) {
	// Covers: FR-MNV-001
	s := newScene(t)
	s.byKind(0, yesAndTimeouts)
	s.byKind(1, yesAndTimeouts)
	mc := maneuver("Rite", 0, 0, steps(drawN(1)))
	mc.Ability.Triggers = []ability.Trigger{entersTrigger(drawN(1))}
	s.g.Players[0].Hand = nil
	mi := s.give(0, mc)
	s.do(enactAction(mi))
	p := s.g.Players[0]
	if len(s.g.Stack) != 0 || !inZone(p.Discard, mi) {
		t.Fatalf("the resolved Maneuver is in its owner's discard pile and the stack is empty (FR-MNV-001)")
	}
	if len(p.Hand) != 2 {
		t.Errorf("player 0 drew its effect's 1 and the pile trigger's 1, hand has %d (FR-ZON-004)", len(p.Hand))
	}
}

func TestFR_ZON_003_EntersDiscardIsTheOwnersTrigger(t *testing.T) {
	// Covers: FR-ZON-002
	s := newScene(t)
	s.byKind(0, firstPick)
	s.byKind(1, yesAndTimeouts)
	stolen := s.put(1, ZoneMainTheater, edMonk("Monk"))
	stolen.Controller = 0
	before := len(s.g.Players[1].Hand)
	s.enactAlliedStep(killTarget())
	if s.alive(stolen) {
		t.Fatalf("precondition: the controlled unit is killed")
	}
	if !inZone(s.g.Players[1].Discard, stolen.Inst) || inZone(s.g.Players[0].Discard, stolen.Inst) {
		t.Fatalf("a destroyed asset goes to its owner's discard pile, player 1's (FR-ZON-002)")
	}
	if got := len(s.g.Players[1].Hand) - before; got != 1 {
		t.Errorf("the owner, player 1, drew %d cards from the trigger, want 1 (FR-ZON-003)", got)
	}
	if len(s.g.Players[0].Hand) != 0 {
		t.Errorf("the controller, player 0, draws nothing from it, hand has %d (FR-ZON-003)", len(s.g.Players[0].Hand))
	}
}

func TestFR_ZON_004_EntersDiscardFiresOnlyIntoTheDiscardPile(t *testing.T) {
	// (a) returned to its owner's hand
	s := newScene(t)
	s.byKind(0, firstPick)
	s.byKind(1, yesAndTimeouts)
	deck := len(s.g.Players[0].Deck)
	s.put(0, ZoneMainTheater, edMonk("Monk"))
	s.enactAlliedStep(ability.Step{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget, Target: 0})
	if got := pileLogPlaces(s, "Monk"); got != 0 {
		t.Errorf("a card returned to hand fires no enter-discard trigger, got %d placements (FR-ZON-004)", got)
	}
	if len(s.g.Players[0].Hand) != 1 || len(s.g.Players[0].Deck) != deck {
		t.Errorf("only the returned card is in hand and nothing is drawn: hand %d, deck %d of %d (FR-ZON-004)",
			len(s.g.Players[0].Hand), len(s.g.Players[0].Deck), deck)
	}

	// (b) put on its owner's deck
	s = newScene(t)
	s.byKind(0, firstPick)
	s.byKind(1, yesAndTimeouts)
	deck = len(s.g.Players[0].Deck)
	s.put(0, ZoneMainTheater, edMonk("Monk"))
	s.enactAlliedStep(ability.Step{Op: ability.OpPutOnDeck, Subject: ability.SubjectTarget, Target: 0})
	if got := pileLogPlaces(s, "Monk"); got != 0 {
		t.Errorf("a card put on a deck fires no enter-discard trigger, got %d placements (FR-ZON-004)", got)
	}
	if len(s.g.Players[0].Hand) != 0 || len(s.g.Players[0].Deck) != deck+1 {
		t.Errorf("nothing is drawn and the card is on the deck: hand %d, deck %d, want 0 and %d (FR-ZON-004)",
			len(s.g.Players[0].Hand), len(s.g.Players[0].Deck), deck+1)
	}

	// (c) a battlefield trigger (no ActiveZone) on the hook, destroyed
	s = newScene(t)
	s.byKind(0, yesAndTimeouts)
	s.byKind(1, yesAndTimeouts)
	before := len(s.g.Players[1].Hand)
	bf := mortal("Field Monk", 1, 1, ability.Trigger{Hook: ability.HookEntersDiscard,
		Effect: ability.Effect{Steps: []ability.Step{drawN(1)}}})
	inst := killOnP1(s, bf)
	if !inZone(s.g.Players[1].Discard, inst) {
		t.Fatalf("precondition: the unit is in its owner's discard pile")
	}
	if pileLogPlaces(s, "Field Monk") != 0 || len(s.g.Players[1].Hand) != before {
		t.Errorf("a trigger with no ActiveZone works from the battlefield only, so a card entering the pile fires nothing (G1)")
	}
}

func TestFR_TRG_003_EntersDiscardTriggersFollowThePileOrder(t *testing.T) {
	// Covers: SVY-R7
	s := newScene(t)
	s.byKind(0, yesAndTimeouts)
	s.byKind(1, yesAndTimeouts)
	s.put(1, ZoneMainTheater, edMonk("Monk A"))
	s.put(1, ZoneMainTheater, edMonk("Monk B"))
	s.g.Players[0].Hand = nil
	m := s.give(0, maneuver("Purge", 0, 0, steps(ability.Step{Op: ability.OpKill, Subject: ability.SubjectAll,
		Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}})))
	s.do(enactAction(m))
	pile := s.g.Players[1].Discard
	if len(pile) != 2 {
		t.Fatalf("both Monks are in player 1's discard pile: %d cards (FR-ZON-002)", len(pile))
	}
	orders := askedKind(s, 1, DecideOrder)
	if len(orders) != 1 || len(orders[0].Options) != 2 {
		t.Fatalf("player 1 orders the two pile triggers with one decision: %+v (FR-TRG-003)", orders)
	}
	o := orders[0].Options
	if !strings.Contains(o[0].Label, pile[0].Card.Name) || !strings.Contains(o[1].Label, pile[1].Card.Name) {
		t.Errorf("the options follow the pile order, %q then %q: got %q, %q (SVY-R7)",
			pile[0].Card.Name, pile[1].Card.Name, o[0].Label, o[1].Label)
	}
	first, second := placeSeq(s, pile[0].Card.Name, 0), placeSeq(s, pile[1].Card.Name, 0)
	if first == 0 || second == 0 || first > second {
		t.Errorf("by default the card lower in the pile is placed first: %q seq %d, %q seq %d (SVY-R7)",
			pile[0].Card.Name, first, pile[1].Card.Name, second)
	}
}

func TestFR_ZON_004_DiscardFromHandFiresBothHooks(t *testing.T) {
	s := newScene(t)
	s.byKind(0, yesAndTimeouts)
	s.byKind(1, yesAndTimeouts)
	c := trigUnit("Twin", pileTrigger(false, drawN(1)), entersTrigger(drawN(1)))
	s.enactWith([]*Card{c}, discardN(1))
	if got := pileLogPlaces(s, "Twin"); got != 2 {
		t.Errorf("a discard from hand fires both the from-hand and the enter-discard trigger, got %d placements (FR-ZON-004)", got)
	}
	if got := len(s.g.Players[0].Hand); got != 2 {
		t.Errorf("both triggers drew 1, hand has %d (FR-ZON-004)", got)
	}
}
