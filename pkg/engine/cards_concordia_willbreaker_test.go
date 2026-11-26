package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0309: Will Breaker (unit, R4 T0, 5/4, Reserves): "When this unit
// deploys, kill chosen opposing unit with Requisition cost 3 or less." The real
// card is loaded from data/cards. Player 0 deploys it from hand to Reserves with
// a pool of R4 and is the active player, unless a case says otherwise. Every
// other card is synthetic and carries the printed Requisition cost the case needs.

// wbUnit is a synthetic unit with the given printed Requisition cost.
func wbUnit(name string, cost int, kw ability.Keywords) *Card {
	c := unit(name, 1, 5, kw)
	c.ReqCost = cost
	return c
}

// wbStructure is a synthetic structure with the given printed Requisition cost.
func wbStructure(name string, cost int) *Card {
	c := structure(name, 0, 5, ability.Keywords{})
	c.ReqCost = cost
	return c
}

// wbDeployCard gives player 0 the card c and R4, and deploys it to Reserves.
func wbDeployCard(s *scene, c *Card) *Instance {
	s.t.Helper()
	inst := s.give(0, c)
	s.setPool(0, 4, 0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	return inst
}

// wbDeploy deploys a fresh real Will Breaker and returns its card.
func wbDeploy(s *scene) *Card {
	s.t.Helper()
	c := s.card("Will Breaker")
	wbDeployCard(s, c)
	return c
}

// wbDestroys counts the destroy events that name a.
func wbDestroys(s *scene, a *Asset) int {
	n := 0
	for _, e := range effectEvents(s.g, 0, "destroy") {
		if e.Target == a.ID {
			n++
		}
	}
	return n
}

func TestCard_WillBreaker_EntersReservesAsFiveFour(t *testing.T) {
	// Covers: FR-TRN-003, FR-SET-004
	s := newScene(t)
	s.script(0)
	c := wbDeploy(s)
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves || a.Controller != 0 {
		t.Fatalf("Will Breaker is on the battlefield in player 0's Reserves after deploying (FR-TRN-003)")
	}
	if p := a.Printed(); p.Power != 5 || p.Health != 4 {
		t.Errorf("Will Breaker is a 5/4: got %d/%d", p.Power, p.Health)
	}
	if a.Exhausted {
		t.Errorf("a newly deployed unit enters Readied (FR-SET-004)")
	}
	if s.g.Players[0].Req != 0 {
		t.Errorf("Will Breaker costs 4 Requisition: %d left", s.g.Players[0].Req)
	}
	if n := len(s.asked(0)); n != 0 || len(s.g.Stack) != 0 {
		t.Errorf("with no opposing unit nothing is asked or placed: asked %d, stack %d (FR-TGT-003)", n, len(s.g.Stack))
	}

	s2 := newScene(t)
	c2 := s2.card("Will Breaker")
	inst := s2.give(0, c2)
	s2.setPool(0, 4, 0)
	s2.reject(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater})
	if !inHand(s2.g.Players[0], inst) || assetOf(s2.g, c2) != nil {
		t.Errorf("deploying Will Breaker to the Main Theater is rejected and the card stays in hand (FR-TRN-003)")
	}

	s3 := newScene(t)
	c3 := s3.card("Will Breaker")
	inst3 := s3.give(0, c3)
	s3.setPool(0, 3, 0)
	s3.reject(Action{Kind: ActDeploy, Player: 0, Card: inst3.ID, Zone: ZoneReserves})
	if !inHand(s3.g.Players[0], inst3) || assetOf(s3.g, c3) != nil {
		t.Errorf("with 3 Requisition Will Breaker (cost 4) cannot be deployed (FR-TRN-003)")
	}
}

func TestCard_WillBreaker_KillsTheChosenUnit(t *testing.T) {
	// Covers: FR-LAY-002, FR-ZON-002
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, wbUnit("Foe", 3, ability.Keywords{}))
	inst := foe.Inst
	s.script(0)
	wbDeploy(s)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("one legal target is chosen without a question: asked %d times (FR-MOD-001)", n)
	}
	if s.alive(foe) {
		t.Fatalf("the cost 3 opposing unit is killed and leaves the battlefield (FR-LAY-002)")
	}
	if !inDiscardPile(s.g.Players[1], inst) {
		t.Errorf("a killed unit's card is in its owner's discard pile (FR-ZON-002)")
	}
	if n := wbDestroys(s, foe); n != 1 {
		t.Errorf("one destroy event names the killed unit, got %d (FR-LAY-002)", n)
	}
}

func TestCard_WillBreaker_OnlyCostThreeOrLessOffered(t *testing.T) {
	// Covers: FR-TGT-001
	s := newScene(t)
	two := s.put(1, ZoneMainTheater, wbUnit("Cost2", 2, ability.Keywords{}))
	three := s.put(1, ZoneReserves, wbUnit("Cost3", 3, ability.Keywords{}))
	four := s.put(1, ZoneMainTheater, wbUnit("Cost4", 4, ability.Keywords{}))
	ally := s.put(0, ZoneMainTheater, wbUnit("AllyCost1", 1, ability.Keywords{}))
	foeStruct := s.put(1, ZoneMainTheater, wbStructure("FoeStructure", 1))
	stormChoose(s, func() *Asset { return two })
	wbDeploy(s)
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("two legal targets are one question, asked %d times (FR-TGT-001)", len(asked))
	}
	d := asked[0]
	if d.Kind != DecideTarget || d.Player != 0 || d.Min != 1 || d.Max != 1 {
		t.Errorf("the question is a mandatory single DecideTarget for the deploying player: %+v", d)
	}
	wantOptions(t, d, "exactly the opposing units of Requisition cost 3 or less (FR-TGT-001)", two, three)
	if s.alive(two) {
		t.Errorf("the chosen cost 2 unit is killed")
	}
	for _, a := range []*Asset{three, four, ally, foeStruct} {
		if !s.alive(a) {
			t.Errorf("only the chosen unit is killed: %s is gone", a)
		}
	}
}

func TestCard_WillBreaker_MayChooseAScratchCompany(t *testing.T) {
	// Covers: SVY-R59, FR-ZON-005
	s := newScene(t)
	sc := s.putScratch(1)
	sc.Inst.Card.ReqCost = 5
	three := s.put(1, ZoneMainTheater, wbUnit("Cost3", 3, ability.Keywords{}))
	stormChoose(s, func() *Asset { return sc })
	wbDeploy(s)
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("two legal targets are one question, asked %d times", len(asked))
	}
	wantOptions(t, asked[0], "a face-down Scratch Company costs 0, so it is offered beside the cost 3 unit (SVY-R59)", sc, three)
	if s.alive(sc) {
		t.Errorf("the chosen Scratch Company is killed (SVY-R59)")
	}
	if !s.alive(three) {
		t.Errorf("the unchosen unit stays")
	}
}

func TestCard_WillBreaker_MustChooseTheOnlyScratchCompany(t *testing.T) {
	// Covers: SVY-R59, FR-TGT-003
	s := newScene(t)
	sc := s.putScratch(1)
	sc.Inst.Card.ReqCost = 5
	s.script(0)
	wbDeploy(s)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("the only legal target is chosen without a question: asked %d times (FR-MOD-001)", n)
	}
	if s.alive(sc) {
		t.Errorf("the only opposing unit, a face-down Scratch Company costing 0, is killed (SVY-R59)")
	}
}

func TestCard_WillBreaker_MayChooseAUnitTurnedFaceDown(t *testing.T) {
	// Covers: SVY-R59, FR-ENT-103
	s := newScene(t)
	down := s.put(1, ZoneMainTheater, wbUnit("Costly", 5, ability.Keywords{}))
	s.script(0)
	morphEnact(s, morphManeuver())
	if !down.FaceDown {
		t.Fatalf("setup: the cost 5 unit is turned face down (FR-ENT-103)")
	}
	up := s.put(1, ZoneMainTheater, wbUnit("CostlyUp", 5, ability.Keywords{}))
	s.script(0)
	wbDeploy(s)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("the face-down unit is the only legal target and is chosen without a question: asked %d times (SVY-R59)", n)
	}
	if s.alive(down) {
		t.Errorf("a unit turned face down costs 0 and is killed (SVY-R59, FR-ENT-103)")
	}
	if !s.alive(up) {
		t.Errorf("the face-up cost 5 unit is not a legal target and stays (FR-TGT-001)")
	}
}

func TestCard_WillBreaker_NoTargetDoesNothing(t *testing.T) {
	// Covers: FR-TGT-003
	s := newScene(t)
	four := s.put(1, ZoneMainTheater, wbUnit("Cost4", 4, ability.Keywords{}))
	nine := s.put(1, ZoneReserves, wbUnit("Cost9", 9, ability.Keywords{}))
	ally := s.put(0, ZoneMainTheater, wbUnit("AllyCost1", 1, ability.Keywords{}))
	s.script(0)
	placed := false
	s.ptPlaceHook(func(int, StackItem) { placed = true })
	c := wbDeploy(s)
	if placed {
		t.Errorf("a deploy trigger with no legal target is not placed (FR-TGT-003)")
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("nothing is asked: %d decisions", n)
	}
	for _, a := range []*Asset{four, nine, ally, assetOf(s.g, c)} {
		if a == nil || !s.alive(a) {
			t.Errorf("every unit stays on the battlefield (FR-TGT-003)")
		}
	}
}

func TestCard_WillBreaker_WardedUnitNotOffered(t *testing.T) {
	// Covers: FR-KWD-002h
	s := newScene(t)
	ward := s.put(1, ZoneMainTheater, wbUnit("Warded", 2, ability.Keywords{Ward: true}))
	s.script(0)
	wbDeploy(s)
	if !s.alive(ward) || len(s.asked(0)) != 0 {
		t.Errorf("an opposing Warded unit is not a legal target, so it stays and nothing is asked (FR-KWD-002h)")
	}

	s2 := newScene(t)
	ward2 := s2.put(1, ZoneMainTheater, wbUnit("Warded", 2, ability.Keywords{Ward: true}))
	plain2 := s2.put(1, ZoneMainTheater, wbUnit("Plain2", 2, ability.Keywords{}))
	plain3 := s2.put(1, ZoneReserves, wbUnit("Plain3", 3, ability.Keywords{}))
	stormChoose(s2, func() *Asset { return plain2 })
	wbDeploy(s2)
	asked := s2.asked(0)
	if len(asked) != 1 {
		t.Fatalf("two legal targets are one question, asked %d times", len(asked))
	}
	wantOptions(t, asked[0], "the Warded unit is not offered beside plain units (FR-KWD-002h)", plain2, plain3)
	if !s2.alive(ward2) || s2.alive(plain2) {
		t.Errorf("only the chosen plain unit is killed (FR-KWD-002h)")
	}
}

func TestCard_WillBreaker_TargetGoneDoesNothing(t *testing.T) {
	// Covers: FR-TGT-003
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, wbUnit("Fleeting", 2, ability.Keywords{}))
	other := s.put(1, ZoneMainTheater, wbUnit("Other", 5, ability.Keywords{}))
	s.script(0)
	s.ptPlaceHook(func(int, StackItem) {
		if s.alive(foe) {
			stormKill(s, foe)
		}
	})
	c := wbDeploy(s)
	if !s.alive(other) || assetOf(s.g, c) == nil {
		t.Errorf("a target that left before resolution is skipped and nothing else is killed (FR-TGT-003)")
	}
	if n := wbDestroys(s, other); n != 0 {
		t.Errorf("no destroy event names any other unit: %d", n)
	}
}

func TestCard_WillBreaker_TargetChangedControllerDoesNothing(t *testing.T) {
	// Covers: FR-TGT-003, FR-ZON-002
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, wbUnit("Turncoat", 2, ability.Keywords{}))
	s.script(0)
	s.ptPlaceHook(func(int, StackItem) { foe.Controller = 0 })
	wbDeploy(s)
	if !s.alive(foe) {
		t.Errorf("a chosen unit that came under player 0's control is no longer an opposing unit, so it is not killed (FR-TGT-003)")
	}
	if inDiscardPile(s.g.Players[1], foe.Inst) || inDiscardPile(s.g.Players[0], foe.Inst) {
		t.Errorf("its card is not in a discard pile (FR-ZON-002)")
	}
}

func TestCard_WillBreaker_SourceGoneStillKills(t *testing.T) {
	// Covers: FR-ZON-004
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, wbUnit("Foe", 3, ability.Keywords{}))
	s.script(0)
	c := s.card("Will Breaker")
	s.ptPlaceHook(func(int, StackItem) {
		if wb := assetOf(s.g, c); wb != nil {
			stormKill(s, wb)
		}
	})
	inst := wbDeployCard(s, c)
	if assetOf(s.g, c) != nil {
		t.Fatalf("setup: Will Breaker should have left before resolution")
	}
	if s.alive(foe) {
		t.Errorf("the trigger still resolves from last-known information and kills the chosen unit (FR-ZON-004)")
	}
	if !inDiscardPile(s.g.Players[0], inst) {
		t.Errorf("Will Breaker's card is in its owner's discard pile (FR-ZON-004)")
	}
}

func TestCard_WillBreaker_FaceDownScratchDoesNothing(t *testing.T) {
	// Covers: FR-ZON-005
	s := newScene(t)
	top := s.onTop(0, s.card("Will Breaker"))
	foe := s.put(1, ZoneMainTheater, wbUnit("Foe", 1, ability.Keywords{}))
	s.script(0)
	s.do(enactAction(glEnact(s)))
	p := s.g.Players[0]
	if !scratched(p, top[0]) || faceDownCount(p) != 1 {
		t.Fatalf("setup: Will Breaker enters Reserves as a face-down Scratch Company (FR-ZON-005)")
	}
	if !s.alive(foe) {
		t.Errorf("a Scratch Company has no printed deploy trigger, so the opposing unit stays (FR-ZON-005)")
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("nothing is asked: %d decisions", n)
	}
}
