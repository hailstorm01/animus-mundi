package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0308: Echoing Nightmare (unit, Spirit, Dread, R5 T0, 6/2, Reserves):
// "When this unit deploys, return chosen opposing unit in the Main Theater to its
// owner's hand. If a unit is returned this way, move this unit to the Main
// Theater." The real card is loaded from data/cards. Player 0 deploys it from
// hand to Reserves with R5 and is the active player unless a case says otherwise.

// ecnDeploy gives player 0 the card and R5 T0, deploys it to Reserves and
// returns its hand instance.
func ecnDeploy(s *scene, c *Card) *Instance {
	s.t.Helper()
	inst := s.give(0, c)
	s.setPool(0, 5, 0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	return inst
}

func ecnEvents(s *scene, kind string, id int) []Event {
	var out []Event
	for _, e := range eventsOfKind(s.g.Log, kind) {
		if e.Target == id {
			out = append(out, e)
		}
	}
	return out
}

// ecnStays fails unless the Nightmare is still in player 0's Reserves.
func ecnStays(t *testing.T, s *scene, c *Card, why string) {
	t.Helper()
	nm := assetOf(s.g, c)
	if nm == nil || nm.Zone != ZoneReserves || !inAssets(s.g.Players[0].Reserves, nm) || inAssets(s.g.Theater, nm) {
		t.Errorf("the Nightmare stays in player 0's Reserves: %s (SVY-R14)", why)
	}
}

func TestCard_EchoingNightmare_EntersReservesAsSixTwo(t *testing.T) {
	// Covers: FR-TRN-003, FR-SET-004
	s := newScene(t)
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, ability.Keywords{}))
	s.script(0)
	c := s.card("Echoing Nightmare")
	ecnDeploy(s, c)
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves || !inAssets(s.g.Players[0].Reserves, a) {
		t.Fatalf("the Nightmare is in player 0's Reserves after deploying (FR-TRN-003)")
	}
	if p := a.Printed(); p.Power != 6 || p.Health != 2 || p.Type != TypeUnit {
		t.Errorf("the Nightmare is a 6/2 unit: got %d/%d type %v", p.Power, p.Health, p.Type)
	}
	if a.Exhausted {
		t.Errorf("a newly deployed unit enters Readied (FR-SET-004)")
	}
	if s.g.Players[0].Req != 0 {
		t.Errorf("the Nightmare costs 5 Requisition: %d left", s.g.Players[0].Req)
	}
	if len(s.g.Stack) != 0 || len(s.asked(0)) != 0 {
		t.Errorf("with no opposing unit nothing is placed or asked: stack %d, asked %d (FR-TGT-003)", len(s.g.Stack), len(s.asked(0)))
	}

	s2 := newScene(t)
	c2 := s2.card("Echoing Nightmare")
	inst2 := s2.give(0, c2)
	s2.setPool(0, 5, 0)
	s2.reject(Action{Kind: ActDeploy, Player: 0, Card: inst2.ID, Zone: ZoneMainTheater})
	if !inHand(s2.g.Players[0], inst2) || assetOf(s2.g, c2) != nil {
		t.Errorf("deploying the Nightmare to the Main Theater is rejected and the card stays in hand (FR-TRN-003)")
	}

	s3 := newScene(t)
	c3 := s3.card("Echoing Nightmare")
	inst3 := s3.give(0, c3)
	s3.setPool(0, 4, 0)
	s3.reject(Action{Kind: ActDeploy, Player: 0, Card: inst3.ID, Zone: ZoneReserves})
	if !inHand(s3.g.Players[0], inst3) || assetOf(s3.g, c3) != nil {
		t.Errorf("with 4 Requisition the Nightmare (cost 5) cannot be deployed (FR-TRN-003)")
	}
}

func TestCard_EchoingNightmare_ReturnsTheUnitAndMoves(t *testing.T) {
	// Covers: FR-ZON-002, SVY-R14
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 2, 6, ability.Keywords{}))
	foe.Damage, foe.Exhausted = 3, true
	foe.Modifiers = []Modifier{{Layer: 5, Power: 1}}
	inst := foe.Inst
	s.script(0)
	c := s.card("Echoing Nightmare")
	ecnDeploy(s, c)
	if s.alive(foe) || inAssets(s.g.Theater, foe) {
		t.Errorf("the returned unit is off the battlefield (FR-ZON-002)")
	}
	if !inHand(s.g.Players[1], inst) || inDiscardPile(s.g.Players[1], inst) {
		t.Errorf("the unit's card is in player 1's hand, not a discard pile (FR-ZON-002)")
	}
	nm := assetOf(s.g, c)
	if nm == nil || nm.Zone != ZoneMainTheater || !inAssets(s.g.Theater, nm) || inAssets(s.g.Players[0].Reserves, nm) {
		t.Fatalf("the Nightmare is in the Main Theater after the unit is returned (SVY-R14)")
	}
	if nm.Controller != 0 {
		t.Errorf("player 0 still controls the Nightmare: %d", nm.Controller)
	}
	if nm.Exhausted {
		t.Errorf("a moved unit keeps its state: the Nightmare is still Readied (SVY-R1)")
	}
	leaves, moves := ecnEvents(s, EventLeave, foe.ID), ecnEvents(s, EventMove, nm.ID)
	if len(leaves) != 1 || len(moves) != 1 {
		t.Fatalf("one leave event naming the unit and one move event naming the Nightmare: got %d and %d (FR-ZON-002)", len(leaves), len(moves))
	}
	if leaves[0].Seq >= moves[0].Seq {
		t.Errorf("the return happens before the move: leave seq %d, move seq %d (SVY-R14)", leaves[0].Seq, moves[0].Seq)
	}
}

func TestCard_EchoingNightmare_MovedNightmareMayAttack(t *testing.T) {
	// Covers: FR-SET-004
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	s.script(0)
	c := s.card("Echoing Nightmare")
	ecnDeploy(s, c)
	nm := assetOf(s.g, c)
	if nm == nil || nm.Zone != ZoneMainTheater {
		t.Fatalf("setup: the Nightmare moved to the Main Theater")
	}
	s.do(s.attack(nm, s.g.Players[1].HQ))
	s.wantDamage(s.g.Players[1].HQ, 6)
}

func TestCard_EchoingNightmare_OnlyOpposingMainTheaterUnitsOffered(t *testing.T) {
	// Covers: FR-TGT-001
	s := newScene(t)
	foeA := s.put(1, ZoneMainTheater, unit("FoeA", 1, 5, ability.Keywords{}))
	foeB := s.put(1, ZoneMainTheater, unit("FoeB", 1, 5, ability.Keywords{}))
	foeRes := s.put(1, ZoneReserves, unit("FoeRes", 1, 5, ability.Keywords{}))
	wall := s.put(1, ZoneMainTheater, structure("FoeWall", 0, 5, ability.Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, ability.Keywords{}))
	stormChoose(s, func() *Asset { return foeB })
	c := s.card("Echoing Nightmare")
	ecnDeploy(s, c)
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one chosen target is one question: asked %d times (FR-TGT-001)", len(asked))
	}
	d := asked[0]
	if d.Kind != DecideTarget || d.Player != 0 || d.Min != 1 || d.Max != 1 {
		t.Errorf("the question is a mandatory single DecideTarget for player 0: %+v", d)
	}
	wantOptions(t, d, "exactly the two opposing units in the Main Theater", foeA, foeB)
	if !inHand(s.g.Players[1], foeB.Inst) || s.alive(foeB) {
		t.Errorf("the chosen unit returns to player 1's hand (FR-ZON-002)")
	}
	if !s.alive(foeA) || foeA.Zone != ZoneMainTheater || foeRes.Zone != ZoneReserves || !s.alive(wall) || !s.alive(ally) {
		t.Errorf("no other asset is returned (FR-TGT-001)")
	}
	if nm := assetOf(s.g, c); nm == nil || nm.Zone != ZoneMainTheater {
		t.Errorf("the Nightmare moves to the Main Theater (SVY-R14)")
	}
}

func TestCard_EchoingNightmare_NoTargetStaysInReserves(t *testing.T) {
	// Covers: FR-TGT-003
	s := newScene(t)
	foeRes := s.put(1, ZoneReserves, unit("FoeRes", 1, 5, ability.Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, ability.Keywords{}))
	s.script(0)
	c := s.card("Echoing Nightmare")
	ecnDeploy(s, c)
	ecnStays(t, s, c, "no opposing unit is in the Main Theater (FR-TGT-003, Deployment Triggers Without Legal Targets)")
	if !s.alive(foeRes) || foeRes.Zone != ZoneReserves || !s.alive(ally) || ally.Zone != ZoneMainTheater {
		t.Errorf("nothing returns or moves (FR-TGT-003)")
	}
	if len(s.g.Stack) != 0 || len(s.asked(0)) != 0 {
		t.Errorf("no trigger is placed and nothing is asked: stack %d, asked %d (FR-TGT-003)", len(s.g.Stack), len(s.asked(0)))
	}
}

func TestCard_EchoingNightmare_WardedUnitNotOffered(t *testing.T) {
	// Covers: FR-KWD-002h
	s := newScene(t)
	ward := s.put(1, ZoneMainTheater, unit("FoeWard", 1, 5, ability.Keywords{Ward: true}))
	s.script(0)
	c := s.card("Echoing Nightmare")
	ecnDeploy(s, c)
	if !s.alive(ward) || ward.Zone != ZoneMainTheater {
		t.Errorf("a Ward unit cannot be chosen by an opposing effect and stays on the battlefield (FR-KWD-002h)")
	}
	ecnStays(t, s, c, "the only opposing unit has Ward (FR-KWD-002h)")

	s2 := newScene(t)
	ward2 := s2.put(1, ZoneMainTheater, unit("FoeWard", 1, 5, ability.Keywords{Ward: true}))
	plain := s2.put(1, ZoneMainTheater, unit("FoePlain", 1, 5, ability.Keywords{}))
	s2.script(0)
	c2 := s2.card("Echoing Nightmare")
	ecnDeploy(s2, c2)
	if s2.alive(plain) || !inHand(s2.g.Players[1], plain.Inst) {
		t.Errorf("the unit without Ward is the only legal target and returns (FR-KWD-002h)")
	}
	if !s2.alive(ward2) || ward2.Zone != ZoneMainTheater {
		t.Errorf("the Ward unit stays (FR-KWD-002h)")
	}
	if nm := assetOf(s2.g, c2); nm == nil || nm.Zone != ZoneMainTheater {
		t.Errorf("the Nightmare moves to the Main Theater (SVY-R14)")
	}
}

func TestCard_EchoingNightmare_TargetGoneNoMove(t *testing.T) {
	// Covers: FR-TGT-003, SVY-R14
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Fleeting", 1, 5, ability.Keywords{}))
	other := s.put(1, ZoneMainTheater, unit("Other", 1, 5, ability.Keywords{}))
	stormChoose(s, func() *Asset { return foe })
	s.ptPlaceHook(func(int, StackItem) {
		if s.alive(foe) {
			stormKill(s, foe)
		}
	})
	c := s.card("Echoing Nightmare")
	ecnDeploy(s, c)
	if s.alive(foe) {
		t.Fatalf("setup: the chosen unit should have died before resolution")
	}
	if !inDiscardPile(s.g.Players[1], foe.Inst) || inHand(s.g.Players[1], foe.Inst) {
		t.Errorf("the dead unit is in its owner's discard pile, not returned to hand (FR-TGT-003)")
	}
	if !s.alive(other) || other.Zone != ZoneMainTheater || inHand(s.g.Players[1], other.Inst) {
		t.Errorf("the other opposing unit is not returned and stays in the Main Theater (FR-TGT-003)")
	}
	ecnStays(t, s, c, "the return did nothing, so the If-done move does not apply (SVY-R14)")
}

func TestCard_EchoingNightmare_TargetLeftTheMainTheaterNoMove(t *testing.T) {
	// Covers: FR-TGT-003
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Fleeting", 1, 5, ability.Keywords{}))
	s.put(1, ZoneMainTheater, unit("Other", 1, 5, ability.Keywords{}))
	stormChoose(s, func() *Asset { return foe })
	s.ptPlaceHook(func(int, StackItem) {
		if foe.Zone == ZoneMainTheater {
			s.moveTo(foe, ZoneReserves)
		}
	})
	c := s.card("Echoing Nightmare")
	ecnDeploy(s, c)
	if foe.Zone != ZoneReserves {
		t.Fatalf("setup: the chosen unit should have moved to Reserves before resolution")
	}
	if !s.alive(foe) || !inAssets(s.g.Players[1].Reserves, foe) || inHand(s.g.Players[1], foe.Inst) {
		t.Errorf("the unit is no longer in the Main Theater, so it is illegal at the second check and stays in player 1's Reserves (FR-TGT-003)")
	}
	ecnStays(t, s, c, "the return did nothing (SVY-R14)")
}

func TestCard_EchoingNightmare_SourceGoneStillReturns(t *testing.T) {
	// Covers: FR-ZON-004, FR-TGT-003
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	s.script(0)
	c := s.card("Echoing Nightmare")
	s.ptPlaceHook(func(int, StackItem) {
		if nm := assetOf(s.g, c); nm != nil {
			stormKill(s, nm)
		}
	})
	inst := ecnDeploy(s, c)
	if assetOf(s.g, c) != nil {
		t.Fatalf("setup: the Nightmare should have left before resolution")
	}
	if s.alive(foe) || !inHand(s.g.Players[1], foe.Inst) {
		t.Errorf("the chosen unit still returns to player 1's hand when the source is gone (FR-ZON-004, FR-TGT-003)")
	}
	if !inDiscardPile(s.g.Players[0], inst) || inHand(s.g.Players[0], inst) {
		t.Errorf("the Nightmare's card is in player 0's discard pile and nothing moves it (FR-ZON-004)")
	}
}

func TestCard_EchoingNightmare_ControlledUnitGoesToItsOwnersHand(t *testing.T) {
	// Covers: FR-ZON-002, FR-ZON-003
	s := newScene(t)
	mine := s.put(0, ZoneMainTheater, unit("Mine", 2, 5, ability.Keywords{}))
	s.endTurn()
	if s.g.Active != 1 {
		t.Fatalf("setup: player 1 is the active player")
	}
	hollow := s.put(1, ZoneMainTheater, s.card("Rhapsody's Hollow"))
	s.setPool(1, 0, 3)
	s.script(1, Answer{Picks: []int{0}})
	s.do(activate(hollow, 0))
	if mine.Controller != 1 || mine.Zone != ZoneMainTheater {
		t.Fatalf("setup: Rhapsody's Hollow gave player 1 control of the unit in the Main Theater: controller %d zone %v", mine.Controller, mine.Zone)
	}
	s.endTurn()
	if s.g.Active != 0 {
		t.Fatalf("setup: player 0 is the active player again")
	}
	s.script(0)
	c := s.card("Echoing Nightmare")
	ecnDeploy(s, c)
	if s.alive(mine) {
		t.Fatalf("the controlled unit is a legal target and leaves the battlefield (FR-TGT-001)")
	}
	if !inHand(s.g.Players[0], mine.Inst) {
		t.Errorf("it goes to its owner's (player 0's) hand (FR-ZON-002, FR-ZON-003)")
	}
	if inHand(s.g.Players[1], mine.Inst) {
		t.Errorf("it does not go to its controller's (player 1's) hand (FR-ZON-003)")
	}
	if nm := assetOf(s.g, c); nm == nil || nm.Zone != ZoneMainTheater {
		t.Errorf("the Nightmare moves to the Main Theater (SVY-R14)")
	}
}

func TestCard_EchoingNightmare_MoveFiresMovesHooks(t *testing.T) {
	// Covers: SVY-R1, FR-ZON-002
	watcher := func(s *scene) *Card {
		c := *s.card("Echoing Nightmare")
		if c.Ability == nil {
			s.t.Fatalf("the Nightmare has no executable definition")
		}
		d := *c.Ability
		d.Triggers = append(append([]ability.Trigger(nil), d.Triggers...),
			ability.Trigger{Hook: ability.HookMoves, Effect: steps(armorSelf(1))})
		c.Ability = &d
		return &c
	}

	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	s.script(0)
	c := watcher(s)
	ecnDeploy(s, c)
	nm := assetOf(s.g, c)
	if nm == nil || nm.Zone != ZoneMainTheater {
		t.Fatalf("setup: the Nightmare moved to the Main Theater")
	}
	if nm.Armor != 1 {
		t.Errorf("Armor %d, want 1: the move fired its Moves trigger once and the deploy did not (SVY-R1)", nm.Armor)
	}

	s2 := newScene(t)
	s2.script(0)
	c2 := watcher(s2)
	ecnDeploy(s2, c2)
	if nm := assetOf(s2.g, c2); nm == nil || nm.Zone != ZoneReserves || nm.Armor != 0 {
		t.Errorf("with no target the Nightmare does not move and gains no Armor (SVY-R1)")
	}
}

func TestCard_EchoingNightmare_FaceDownScratchDoesNothing(t *testing.T) {
	// Covers: FR-ZON-005
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	top := s.onTop(0, s.card("Echoing Nightmare"))
	s.script(0)
	inst := glEnact(s)
	s.do(enactAction(inst))
	p := s.g.Players[0]
	if !scratched(p, top[0]) || faceDownCount(p) != 1 {
		t.Fatalf("setup: the Nightmare became a face-down Scratch Company in player 0's Reserves")
	}
	if !s.alive(foe) || foe.Zone != ZoneMainTheater || inHand(s.g.Players[1], foe.Inst) {
		t.Errorf("a face-down Scratch Company has no deploy trigger, so the opposing unit stays (FR-ZON-005)")
	}
	for _, a := range s.g.Assets() {
		if a.Controller == 0 && a.Zone == ZoneMainTheater {
			t.Errorf("nothing moves to player 0's Main Theater: %s (FR-ZON-005)", a)
		}
	}
	if len(s.asked(0)) != 0 {
		t.Errorf("nothing is asked (FR-ZON-005)")
	}
}
