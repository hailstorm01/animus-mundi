package engine

import (
	"testing"
)

// Card tests for Devilhorn Captain (T-0052), written from the CSV rules text in
// data/cards/units-and-structures-path.csv: Unit, 2 Requisition, Reserves, 2/3,
// "Unbowed: Deal 1 damage to chosen asset in the Main Theater." Cards are the
// real ones loaded with s.card.

// captainBoard is player 1's Captain in the Main Theater facing player 0 (the
// active player), with units, structures and Reserves assets of both sides.
type captainBoard struct {
	cap                                      *Asset
	u0, s0, u1, s1, ward0, rs0, rs1, scratch *Asset
}

func (s *scene) captainBoard() captainBoard {
	return captainBoard{
		cap:     s.put(1, ZoneMainTheater, s.card("Devilhorn Captain")),
		u0:      s.put(0, ZoneMainTheater, unit("Enemy Unit", 1, 5, Keywords{})),
		s0:      s.put(0, ZoneMainTheater, structure("Enemy Structure", 0, 5, Keywords{})),
		u1:      s.put(1, ZoneMainTheater, unit("Ally Unit", 1, 5, Keywords{})),
		s1:      s.put(1, ZoneMainTheater, structure("Ally Structure", 0, 5, Keywords{})),
		ward0:   s.put(0, ZoneMainTheater, unit("Warded Enemy", 1, 5, Keywords{Ward: true})),
		rs0:     s.put(0, ZoneReserves, unit("Enemy Reserves Unit", 1, 5, Keywords{})),
		rs1:     s.put(1, ZoneReserves, unit("Ally Reserves Unit", 1, 5, Keywords{})),
		scratch: s.putScratch(0),
	}
}

// Control: a 2/3 that enters Reserves for 2 Requisition and does nothing on deploy.
func TestCard_DevilhornCaptain_EntersReservesAsTwoThreeForTwoRequisitionAndDoesNothing(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("Enemy", 1, 5, Keywords{}))
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	s.script(0)
	s.script(1)
	c := deployPath(s, "Devilhorn Captain", 2)
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves {
		t.Fatalf("Devilhorn Captain is on the battlefield in Reserves after deploying")
	}
	if p := a.Printed(); p.Power != 2 || p.Health != 3 {
		t.Errorf("Devilhorn Captain is a 2/3: got %d/%d", p.Power, p.Health)
	}
	if s.g.Players[0].Req != 0 {
		t.Errorf("Devilhorn Captain costs 2 Requisition: %d left", s.g.Players[0].Req)
	}
	if n := len(s.asked(0)) + len(s.asked(1)); n != 0 {
		t.Errorf("Unbowed is not a deploy trigger, so nothing is asked on deploy: asked %d times", n)
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("nothing happens on deploy: total damage %d", got)
	}
}

// Covers: FR-TGT-001, FR-KWD-002h, FR-KWD-002a
func TestCard_DevilhornCaptain_DamagedOffersEveryMainTheaterAssetOfEitherSideButNotOpposingWard(t *testing.T) {
	s := newScene(t)
	b := s.captainBoard()
	hitter := s.put(0, ZoneMainTheater, unit("Hitter", 1, 9, Keywords{}))
	s.scriptRefs(1, nil, b.u0)
	s.do(s.attack(hitter, b.cap))
	asked := s.asked(1)
	if len(asked) != 1 || asked[0].Kind != DecideTarget || asked[0].Player != 1 {
		t.Fatalf("the Captain's controller chooses the target: %+v", asked)
	}
	wantOptions(t, asked[0], "units and structures of either side in the Main Theater, the Captain included; no Reserves asset, no Headquarters, no opposing Ward unit",
		b.cap, b.u0, b.s0, b.u1, b.s1, hitter)
	s.wantDamage(b.u0, 1)
	s.wantDamage(b.ward0, 0)
	s.wantDamage(b.rs0, 0)
	s.wantDamage(b.rs1, 0)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("the opponent is not asked: asked %d times", n)
	}
}

// Covers: FR-TGT-001, FR-KWD-002a
func TestCard_DevilhornCaptain_CanDamageAnAlliedStructureInTheMainTheater(t *testing.T) {
	s := newScene(t)
	b := s.captainBoard()
	hitter := s.put(0, ZoneMainTheater, unit("Hitter", 1, 9, Keywords{}))
	s.scriptRefs(1, nil, b.s1)
	s.do(s.attack(hitter, b.cap))
	s.wantDamage(b.s1, 1)
	s.wantDamage(b.u0, 0)
}

// Covers: FR-KWD-002h
func TestCard_DevilhornCaptain_AlliedWardUnitInTheMainTheaterIsAnOption(t *testing.T) {
	s := newScene(t)
	captain := s.put(1, ZoneMainTheater, s.card("Devilhorn Captain"))
	allyWard := s.put(1, ZoneMainTheater, unit("Warded Ally", 1, 5, Keywords{Ward: true}))
	hitter := s.put(0, ZoneMainTheater, unit("Hitter", 1, 9, Keywords{}))
	s.scriptRefs(1, nil, allyWard)
	s.do(s.attack(hitter, captain))
	asked := s.asked(1)
	if len(asked) == 0 {
		t.Fatalf("the Captain's controller is asked")
	}
	wantOptions(t, asked[0], "Ward only protects against opposing players' targeting", captain, allyWard, hitter)
	s.wantDamage(allyWard, 1)
}

// Covers: FR-KWD-002a
func TestCard_DevilhornCaptain_FiresInReservesAndStillOffersOnlyMainTheaterAssets(t *testing.T) {
	s := newScene(t)
	enemy := s.put(1, ZoneMainTheater, unit("Enemy Unit", 1, 5, Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally Unit", 1, 5, Keywords{}))
	s.put(0, ZoneReserves, unit("Ally Reserves Unit", 1, 5, Keywords{}))
	captain := s.put(0, ZoneReserves, s.card("Devilhorn Captain"))
	// Saboteur: the opposing clause has a sole target (enemy, not asked); the
	// allied clause is asked and picks the Captain in Reserves.
	s.scriptRefs(0, nil, captain, ally)
	deploySaboteur(s)
	asked := s.asked(0)
	if len(asked) != 2 {
		t.Fatalf("two questions: Saboteur's allied clause, then the Captain's Unbowed target: %+v", asked)
	}
	s.wantDamage(captain, 1)
	wantOptions(t, asked[1], "the Captain in Reserves still chooses only among Main Theater assets", ally, enemy)
	s.wantDamage(ally, 1)
	s.wantDamage(enemy, 2) // Saboteur's 2 only
}

// Covers: FR-TGT-001, FR-KWD-002a
func TestCard_DevilhornCaptain_CanChooseItselfAndItsOwnDamageFiresUnbowedAgain(t *testing.T) {
	s := newScene(t)
	b := s.captainBoard()
	hitter := s.put(0, ZoneMainTheater, unit("Hitter", 1, 9, Keywords{}))
	s.scriptRefs(1, nil, b.cap, b.u1)
	s.do(s.attack(hitter, b.cap))
	asked := s.asked(1)
	if len(asked) != 2 {
		t.Fatalf("combat damage fires Unbowed, and the 1 damage it deals to itself fires it again: %d questions, want 2 (FR-KWD-002a)", len(asked))
	}
	s.wantDamage(b.cap, 2) // 1 from combat, 1 from its own Unbowed
	s.wantDamage(b.u1, 1)  // the second trigger's target
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty once the cascade resolves: %d items", len(s.g.Stack))
	}
}

// Covers: FR-ZON-004, FR-KWD-002a
func TestCard_DevilhornCaptain_LethalDamageResolvesFromDiscardAndTheDeadCaptainIsNotAnOption(t *testing.T) {
	s := newScene(t)
	b := s.captainBoard()
	hitter := s.put(0, ZoneMainTheater, unit("Hitter", 5, 9, Keywords{}))
	inst := b.cap.Inst
	s.scriptRefs(1, nil, b.u1)
	s.do(s.attack(hitter, b.cap))
	if s.alive(b.cap) || !inZone(s.g.Players[1].Discard, inst) {
		t.Fatalf("precondition: 5 damage destroys the 2/3 Captain into its owner's Discard")
	}
	asked := s.asked(1)
	if len(asked) != 1 || asked[0].Kind != DecideTarget || asked[0].Player != 1 {
		t.Fatalf("the dead Captain's controller still chooses: %+v", asked)
	}
	wantOptions(t, asked[0], "the dead Captain is not on the battlefield, so not an option (FR-TGT-003)",
		b.u0, b.s0, b.u1, b.s1, hitter)
	s.wantDamage(b.u1, 1)
}

// Covers: FR-ZON-004
func TestCard_DevilhornCaptain_LethalDamageIsChosenByTheLastKnownController(t *testing.T) {
	s := newScene(t)
	s.endTurn() // player 1 is active
	hitter := s.put(1, ZoneMainTheater, unit("Hitter", 5, 9, Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally of Controller", 1, 5, Keywords{}))
	// Owned by player 1 and controlled by player 0 when it dies.
	captured := s.putControlled(1, 0, s.card("Devilhorn Captain"))
	s.scriptRefs(0, nil, hitter)
	s.script(1)
	s.do(s.attack(hitter, captured))
	if s.alive(captured) {
		t.Fatalf("precondition: the captured Captain is destroyed")
	}
	asked := s.asked(0)
	if len(asked) != 1 || asked[0].Player != 0 {
		t.Fatalf("the trigger is chosen by the Captain's last-known controller, player 0: %+v", asked)
	}
	wantOptions(t, asked[0], "both remaining Main Theater assets, whichever side they are on", ally, hitter)
	if n := len(s.asked(1)); n != 0 {
		t.Errorf("the card's owner is not asked: asked %d times", n)
	}
}

// Covers: FR-MOD-001
func TestCard_DevilhornCaptain_SoleLegalTargetIsNotAsked(t *testing.T) {
	s := newScene(t)
	hitter := s.put(0, ZoneMainTheater, unit("Hitter", 5, 9, Keywords{}))
	captain := s.put(1, ZoneMainTheater, s.card("Devilhorn Captain"))
	s.put(1, ZoneReserves, unit("Ally Reserves Unit", 1, 5, Keywords{}))
	s.script(0)
	s.script(1)
	s.do(s.attack(hitter, captain))
	if s.alive(captain) {
		t.Fatalf("precondition: the Captain is destroyed, leaving the hitter as the only Main Theater asset")
	}
	if n := len(s.asked(0)) + len(s.asked(1)); n != 0 {
		t.Errorf("a sole legal target is selected without a question: asked %d times (FR-MOD-001)", n)
	}
	s.wantDamage(hitter, 3) // 2 retaliation, 1 from the Captain's Unbowed
}

// Covers: FR-MOD-001
func TestCard_DevilhornCaptain_TimeoutDamagesExactlyOneLegalTarget(t *testing.T) {
	s := newScene(t)
	b := s.captainBoard()
	hitter := s.put(0, ZoneMainTheater, unit("Hitter", 5, 9, Keywords{}))
	s.script(1, Answer{Timeout: true})
	s.do(s.attack(hitter, b.cap))
	if s.alive(b.cap) {
		t.Fatalf("precondition: the Captain is destroyed, so its trigger cannot cascade")
	}
	got := b.u0.Damage + b.s0.Damage + b.u1.Damage + b.s1.Damage + (hitter.Damage - 2) // 2 is the retaliation
	if got != 1 {
		t.Errorf("a timeout picks one legal Main Theater target and deals it 1 damage: %d total on them (FR-MOD-001)", got)
	}
	for _, a := range []*Asset{b.ward0, b.rs0, b.rs1, b.scratch, s.g.Players[0].HQ, s.g.Players[1].HQ} {
		if a.Damage != 0 {
			t.Errorf("%s is not a legal option and takes no damage: damage %d", a, a.Damage)
		}
	}
}

// Covers: FR-REP-004, FR-KWD-002b
func TestCard_DevilhornCaptain_DamageReducedToZeroByResistantDoesNotFire(t *testing.T) {
	s := newScene(t)
	b := s.captainBoard()
	b.cap.Inst.Card.Keywords.Resistant = 2
	hitter := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	s.script(1)
	s.do(s.attack(hitter, b.cap))
	s.wantDamage(b.cap, 0)
	if n := len(s.asked(1)); n != 0 {
		t.Errorf("2 damage less Resistant 2 is 0, so Unbowed does not trigger: asked %d times (FR-REP-004 Stage 5)", n)
	}
	for _, a := range []*Asset{b.u0, b.s0, b.u1, b.s1} {
		s.wantDamage(a, 0)
	}
}

// Covers: FR-REP-004
func TestCard_DevilhornCaptain_ZeroPowerAttackerDoesNotFireIt(t *testing.T) {
	s := newScene(t)
	b := s.captainBoard()
	hitter := s.put(0, ZoneMainTheater, unit("Pacifist", 0, 9, Keywords{}))
	s.script(1)
	s.do(s.attack(hitter, b.cap))
	s.wantDamage(b.cap, 0)
	if n := len(s.asked(1)); n != 0 {
		t.Errorf("0 damage commits and absorbs nothing, so nothing is taken: asked %d times (FR-REP-004 Stage 5)", n)
	}
}

// Covers: FR-ZON-004
func TestCard_DevilhornCaptain_FaceDownCopyFiresNothingWhenDamaged(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	top := s.onTop(0, s.card("Devilhorn Captain"))
	s.script(0)
	s.script(1)
	s.do(Action{Kind: ActScratch, Player: 0, Asset: p.HQ.ID})
	var sc *Asset
	for _, a := range p.Reserves {
		if a.Inst == top[0] {
			sc = a
		}
	}
	if sc == nil || !sc.FaceDown {
		t.Fatalf("precondition: the Captain on top of the deck became a face-down Scratch Company")
	}
	s.endTurn()
	s.endTurn()
	s.do(Action{Kind: ActMarch, Player: 0, Asset: sc.ID})
	bystander0 := s.put(0, ZoneMainTheater, unit("Ally Bystander", 1, 5, Keywords{}))
	bystander1 := s.put(1, ZoneMainTheater, unit("Enemy Bystander", 1, 5, Keywords{}))
	s.endTurn()
	hit := s.put(1, ZoneMainTheater, unit("Hitter", 9, 9, Keywords{}))
	s.do(s.attack(hit, sc))
	if s.alive(sc) {
		t.Fatalf("precondition: the face-down Scratch Company is destroyed")
	}
	if n := len(s.asked(0)) + len(s.asked(1)); n != 0 {
		t.Errorf("a face-down Scratch Company has no printed abilities: asked %d times (FR-ZON-004)", n)
	}
	s.wantDamage(bystander0, 0)
	s.wantDamage(bystander1, 0)
}
