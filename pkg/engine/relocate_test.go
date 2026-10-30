package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0102: OpReturnToHand and OpPutOnDeck move every battlefield asset
// their step acts on, whatever the subject (FR-ZON-004, FR-TGT-003, SVY-R3).
// Cards are synthetic. Every scene deploys a unit "Source" from player 0's
// hand (see iydDeploy); its deploy trigger runs the steps under test. Player 1
// is the opponent.

var rlEnemyTarget = []ability.Target{{Filter: iydEnemy}}

func rlStep(op ability.Op, subject ability.Subject, bottom bool) ability.Step {
	return ability.Step{Op: op, Subject: subject, Bottom: bottom}
}

func rlLeaveTo(s *scene, id int, zone string) (Event, bool) {
	for _, e := range eventsOfKind(s.g.Log, "leave") {
		if e.Target == id && e.Zone == zone {
			return e, true
		}
	}
	return Event{}, false
}

func TestFR_ZON_004_ReturnedTargetLeavesWithItsLastKnownState(t *testing.T) {
	s := newScene(t)
	e := s.put(1, ZoneMainTheater, unit("Enemy", 2, 6, Keywords{}))
	e.Damage = 3
	e.TempPower = 2
	inst := e.Inst
	s.script(0, Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, rlEnemyTarget, rlStep(ability.OpReturnToHand, ability.SubjectTarget, false)))
	if s.alive(e) {
		t.Fatalf("the returned unit leaves the battlefield (FR-ZON-004)")
	}
	if !inHand(s.g.Players[1], inst) {
		t.Errorf("the returned unit's card is in its owner's hand")
	}
	if _, ok := rlLeaveTo(s, e.ID, ZoneHand.String()); !ok {
		t.Errorf("a leave event names the unit with zone hand")
	}
	lk := s.g.LastKnownEffective(e.ID)
	if lk == nil {
		t.Fatalf("last-known effective state is kept for an asset that left (FR-ZON-004)")
	}
	if lk.Power != 4 {
		t.Errorf("last-known power includes the temporary modifier: got %d, want 4", lk.Power)
	}
	if inst.Card.Power != 2 || inst.Card.Health != 6 {
		t.Errorf("in hand the card has its printed characteristics: %d/%d", inst.Card.Power, inst.Card.Health)
	}
	// Deploying it again gives a new asset with no damage and printed power.
	s.endTurn()
	s.do(Action{Kind: ActDeploy, Player: 1, Card: inst.ID, Zone: ZoneMainTheater})
	n := assetOf(s.g, inst.Card)
	if n == nil {
		t.Fatalf("the card can be deployed again")
	}
	if n.ID == e.ID || n.Damage != 0 || n.TempPower != 0 {
		t.Errorf("a redeployed card is a new asset with no damage or modifier: id %d (old %d), damage %d, temp power %d", n.ID, e.ID, n.Damage, n.TempPower)
	}
}

func rlPutOnDeck(t *testing.T, bottom bool, subject ability.Subject) (*scene, *Asset, []*Instance) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 4, Keywords{}))
	top := s.onTop(0, unit("Top", 1, 1, Keywords{}))
	step := rlStep(ability.OpPutOnDeck, subject, bottom)
	var trig ability.Trigger
	if subject == ability.SubjectPick {
		step.Filter = iydAlly
		step.UpTo = 1
		s.ppScript(Answer{Picks: []int{0}})
		trig = iydTrigger(false, nil, step)
	} else {
		s.script(0, Answer{Picks: []int{0}})
		trig = iydTrigger(false, []ability.Target{{Filter: iydAlly}}, step)
	}
	s.iydDeploy(trig)
	return s, ally, top
}

func TestFR_ZON_004_PickedUnitGoesOnTheBottomOfItsOwnersDeck(t *testing.T) {
	// Covers: SVY-R3
	s, ally, top := rlPutOnDeck(t, true, ability.SubjectPick)
	if s.alive(ally) {
		t.Fatalf("the picked unit leaves the battlefield")
	}
	d := s.g.Players[0].Deck
	if n := len(d); n == 0 || d[n-1] != ally.Inst {
		t.Errorf("the picked unit's card is the last card of its owner's deck")
	}
	if d[0] != top[0] {
		t.Errorf("the top card of the deck is unchanged")
	}
	if _, ok := rlLeaveTo(s, ally.ID, ZoneDeck.String()); !ok {
		t.Errorf("a leave event names the unit with zone deck")
	}
}

func TestFR_ZON_004_PutOnDeckGoesOnTopWithoutBottom(t *testing.T) {
	s, ally, _ := rlPutOnDeck(t, false, ability.SubjectTarget)
	if s.alive(ally) {
		t.Fatalf("the chosen unit leaves the battlefield")
	}
	if d := s.g.Players[0].Deck; len(d) == 0 || d[0] != ally.Inst {
		t.Errorf("without Bottom the unit's card is the first card of its owner's deck")
	}
}

func TestFR_TGT_003_TargetKilledByAnEarlierStepIsNotMoved(t *testing.T) {
	s := newScene(t)
	e := s.put(1, ZoneMainTheater, unit("Enemy", 2, 6, Keywords{}))
	inst := e.Inst
	kill := ability.Step{Op: ability.OpKill, Subject: ability.SubjectTarget}
	// The kill step destroys the chosen unit first; the return step then acts on
	// the same slot, whose unit is no longer on the battlefield.
	s.script(0, Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, rlEnemyTarget, kill,
		rlStep(ability.OpReturnToHand, ability.SubjectTarget, false),
		iydIfDone(scratchN(1))))
	if s.alive(e) {
		t.Fatalf("the kill step destroyed the chosen unit")
	}
	p := s.g.Players[1]
	if !inDiscardPile(p, inst) {
		t.Errorf("the unit that left before the step stays in the discard pile (FR-TGT-003)")
	}
	if inHand(p, inst) {
		t.Errorf("a target that is no longer on the battlefield is not returned to hand")
	}
	if _, ok := rlLeaveTo(s, e.ID, ZoneHand.String()); ok {
		t.Errorf("no leave event to hand is logged for it")
	}
	if n := faceDownCount(s.g.Players[0]); n != 0 {
		t.Errorf("the return step counts 0, so the IfDone step does not run: %d Scratch Companies", n)
	}
}

func TestFR_ZON_004_FaceDownAssetReturnsAsItsCard(t *testing.T) {
	// Covers: FR-ZON-004
	s := newScene(t)
	c := s.putScratch(1)
	inst := c.Inst
	s.script(0, Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, []ability.Target{{Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}}},
		rlStep(ability.OpReturnToHand, ability.SubjectTarget, false)))
	if s.alive(c) {
		t.Fatalf("the face-down unit leaves the battlefield")
	}
	if !inHand(s.g.Players[1], inst) {
		t.Errorf("it goes to its owner's hand as its own printed card instance")
	}
	e, ok := rlLeaveTo(s, c.ID, ZoneHand.String())
	if !ok {
		t.Fatalf("a leave event to hand is logged")
	}
	if e.Card != "" {
		t.Errorf("a face-down asset's card is private in a hand, so the leave event names no card: %q", e.Card)
	}
}

func TestFR_ZON_004_SourceStillOnTheBattlefieldReturnsItself(t *testing.T) {
	s := newScene(t)
	c := s.iydDeploy(iydTrigger(false, nil, rlStep(ability.OpReturnToHand, ability.SubjectSelf, false)))
	var inst *Instance
	for _, h := range s.g.Players[0].Hand {
		if h.Card == c {
			inst = h
		}
	}
	if assetOf(s.g, c) != nil {
		t.Errorf("the source leaves the battlefield")
	}
	if inst == nil {
		t.Errorf("the source's card is in its owner's hand")
	}
}

func TestFR_ZON_004_ReturnAllMatchingUnits(t *testing.T) {
	s := newScene(t)
	e1 := s.put(1, ZoneMainTheater, unit("E1", 1, 3, Keywords{}))
	e2 := s.put(1, ZoneReserves, unit("E2", 1, 3, Keywords{}))
	al := s.put(0, ZoneMainTheater, unit("Ally", 1, 3, Keywords{}))
	ret := ability.Step{Op: ability.OpReturnToHand, Subject: ability.SubjectAll, Filter: iydEnemy}
	s.iydDeploy(iydTrigger(false, nil, ret, iydIfDone(scratchN(1))))
	for _, e := range []*Asset{e1, e2} {
		if s.alive(e) {
			t.Errorf("%s leaves the battlefield", e)
		}
		if !inHand(s.g.Players[1], e.Inst) {
			t.Errorf("%s's card is in its owner's hand", e)
		}
	}
	if !s.alive(al) {
		t.Errorf("an allied unit is not returned")
	}
	if n := faceDownCount(s.g.Players[0]); n != 1 {
		t.Errorf("the step counts what it moved, so the IfDone step runs: %d Scratch Companies, want 1", n)
	}
}
