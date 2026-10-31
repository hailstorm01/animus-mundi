package engine

import "testing"

// Tests for T-0120 (FR-MNV-001, FR-TGT-002, FR-REP-004): the Animus Maneuvers
// Despoiler's Emergence and Deafening Aria, loaded from data/cards. Player 0
// enacts; player 1 is the opponent.
//
// Despoiler's Emergence: "Deal 4 damage to all opposing assets in the Main
// Theater." / "Deal 2 damage to all opposing non-Scratch assets in opponents'
// Reserves."
// Deafening Aria: "Deal 1 damage to all opposing units in the Main Theater."

// enactFor gives player 0 the named real Maneuver and enough resources, and
// returns the card instance. Req is 3 and Tac is 5 beforehand.
func enactFor(s *scene, name string) *Instance {
	inst := s.give(0, s.card(name))
	p := s.g.Players[0]
	p.Req, p.Tac = 3, 5
	return inst
}

// damageEventsOn returns how many damage events in the log hit asset a.
func damageEventsOn(s *scene, a *Asset) int {
	n := 0
	for _, e := range damageEvents(s.g.Log) {
		if e.Target == a.ID {
			n++
		}
	}
	return n
}

func TestCard_DespoilersEmergence_FourToEveryOpposingAssetInTheMainTheater(t *testing.T) {
	s := newScene(t)
	u1 := s.put(1, ZoneMainTheater, unit("E1", 1, 9, Keywords{}))
	u2 := s.put(1, ZoneMainTheater, unit("E2", 1, 9, Keywords{}))
	st := s.put(1, ZoneMainTheater, structure("EWall", 0, 9, Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{}))
	allySt := s.put(0, ZoneMainTheater, structure("AWall", 0, 9, Keywords{}))
	s.do(enactAction(enactFor(s, "Despoiler's Emergence")))
	for _, a := range []*Asset{u1, u2, st} {
		s.wantDamage(a, 4)
	}
	s.wantDamage(ally, 0)
	s.wantDamage(allySt, 0)
	if n := damageEventsOn(s, ally) + damageEventsOn(s, allySt); n != 0 {
		t.Errorf("allied assets are not opposing: %d damage events on them, want 0 (FR-MNV-001)", n)
	}
}

func TestCard_DespoilersEmergence_TwoToOpposingNonScratchAssetsInReserves(t *testing.T) {
	s := newScene(t)
	u := s.put(1, ZoneReserves, unit("RUnit", 1, 9, Keywords{}))
	st := s.put(1, ZoneReserves, structure("RWall", 0, 9, Keywords{}))
	scratch := s.putScratch(1)
	allyU := s.put(0, ZoneReserves, unit("AUnit", 1, 9, Keywords{}))
	allyScratch := s.putScratch(0)
	s.do(enactAction(enactFor(s, "Despoiler's Emergence")))
	s.wantDamage(u, 2)
	s.wantDamage(st, 2)
	s.wantDamage(scratch, 0)
	if n := damageEventsOn(s, scratch); n != 0 {
		t.Errorf("a face-down Scratch Company is excluded from the Reserves step: %d damage events, want 0", n)
	}
	s.wantDamage(allyU, 0)
	s.wantDamage(allyScratch, 0)
}

func TestCard_DespoilersEmergence_NoHeadquartersIsDamaged(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("E1", 1, 9, Keywords{}))
	s.do(enactAction(enactFor(s, "Despoiler's Emergence")))
	for i, p := range s.g.Players {
		s.wantDamage(p.HQ, 0)
		if n := damageEventsOn(s, p.HQ); n != 0 {
			t.Errorf("player %d's Headquarters is in neither zone: %d damage events, want 0", i, n)
		}
	}
}

func TestCard_DespoilersEmergence_ArmorAbsorbsEachInstance(t *testing.T) {
	// Covers: FR-REP-004
	s := newScene(t)
	armored := s.put(1, ZoneMainTheater, unit("Armored", 1, 9, Keywords{Armor: 3}))
	plain := s.put(1, ZoneMainTheater, unit("Plain", 1, 9, Keywords{}))
	s.do(enactAction(enactFor(s, "Despoiler's Emergence")))
	if armored.Armor != 0 {
		t.Errorf("Armor 3 absorbs 3 of the 4 damage first: Armor %d, want 0 (FR-REP-004)", armored.Armor)
	}
	s.wantDamage(armored, 1)
	s.wantDamage(plain, 4)
}

func TestCard_DespoilersEmergence_LethalAssetsDieAfterItResolves(t *testing.T) {
	s := newScene(t)
	mt := s.put(1, ZoneMainTheater, unit("Frail", 1, 4, Keywords{}))
	res := s.put(1, ZoneReserves, unit("Frailer", 1, 2, Keywords{}))
	survivor := s.put(1, ZoneReserves, unit("Sturdy", 1, 3, Keywords{}))
	inst := enactFor(s, "Despoiler's Emergence")
	s.do(enactAction(inst))
	if s.alive(mt) || s.alive(res) {
		t.Errorf("assets left lethal (4 in the Main Theater, 2 in Reserves) die in the sweep after it resolves")
	}
	if !s.alive(survivor) {
		t.Errorf("a 3-health Reserves unit survives 2 damage")
	}
	opp := s.g.Players[1]
	if !inDiscardPile(opp, mt.Inst) || !inDiscardPile(opp, res.Inst) {
		t.Errorf("dead assets go to their owner's discard pile")
	}
	if !inDiscardPile(s.g.Players[0], inst) {
		t.Errorf("the resolved Maneuver ends in its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_DespoilersEmergence_EnactableWithNoOpposingAsset(t *testing.T) {
	// Covers: SVY-R29
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{}))
	inst := enactFor(s, "Despoiler's Emergence")
	if !offersEnact(s.g, inst) {
		t.Fatalf("a Maneuver with no target slot is a legal Enact with no opposing asset (SVY-R29)")
	}
	s.do(enactAction(inst))
	s.wantDamage(ally, 0)
	if n := len(damageEvents(s.g.Log)); n != 0 {
		t.Errorf("nothing matched, so nothing is damaged: %d damage events, want 0", n)
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver goes to its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_DespoilersEmergence_AsksNothing(t *testing.T) {
	// Covers: FR-TGT-002
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("E1", 1, 9, Keywords{Ward: true}))
	s.put(1, ZoneReserves, unit("E2", 1, 9, Keywords{}))
	s.script(0)
	s.script(1)
	s.do(enactAction(enactFor(s, "Despoiler's Emergence")))
	for p := PlayerID(0); p < 2; p++ {
		if got := s.asked(p); len(got) != 0 {
			t.Errorf("\"all\" is not targeted: player %d was asked %d decisions, want none (FR-TGT-002)", p, len(got))
		}
	}
}

func TestCard_DespoilersEmergence_PaysFourTactics(t *testing.T) {
	s := newScene(t)
	s.do(enactAction(enactFor(s, "Despoiler's Emergence")))
	p := s.g.Players[0]
	if p.Tac != 1 || p.Req != 3 {
		t.Errorf("enacting costs 4 Tactics and no Requisition: pool R%d T%d, want R3 T1 (FR-MNV-001)", p.Req, p.Tac)
	}
}

func TestCard_DeafeningAria_OneToEveryOpposingUnitInTheMainTheater(t *testing.T) {
	s := newScene(t)
	u1 := s.put(1, ZoneMainTheater, unit("E1", 1, 9, Keywords{}))
	u2 := s.put(1, ZoneMainTheater, unit("E2", 1, 9, Keywords{}))
	st := s.put(1, ZoneMainTheater, structure("EWall", 0, 9, Keywords{}))
	res := s.put(1, ZoneReserves, unit("ERes", 1, 9, Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{}))
	s.do(enactAction(enactFor(s, "Deafening Aria")))
	s.wantDamage(u1, 1)
	s.wantDamage(u2, 1)
	s.wantDamage(st, 0)
	s.wantDamage(res, 0)
	s.wantDamage(ally, 0)
	for _, a := range []*Asset{st, res, ally, s.g.Players[1].HQ} {
		if n := damageEventsOn(s, a); n != 0 {
			t.Errorf("%s is not an opposing Main Theater unit: %d damage events, want 0", a, n)
		}
	}
	if p := s.g.Players[0]; p.Tac != 3 || p.Req != 3 {
		t.Errorf("enacting costs 2 Tactics: pool R%d T%d, want R3 T3", p.Req, p.Tac)
	}
}

func TestCard_DeafeningAria_KillsOneHealthUnits(t *testing.T) {
	s := newScene(t)
	one := s.put(1, ZoneMainTheater, unit("OneHealth", 1, 1, Keywords{}))
	two := s.put(1, ZoneMainTheater, unit("TwoHealth", 1, 2, Keywords{}))
	inst := enactFor(s, "Deafening Aria")
	s.do(enactAction(inst))
	if s.alive(one) {
		t.Errorf("a 1-health unit left lethal dies in the sweep after the Maneuver resolves")
	}
	if !inDiscardPile(s.g.Players[1], one.Inst) {
		t.Errorf("the dead unit goes to its owner's discard pile")
	}
	if !s.alive(two) {
		t.Errorf("a 2-health unit survives 1 damage")
	}
	if !inDiscardPile(s.g.Players[0], inst) {
		t.Errorf("the resolved Maneuver ends in its owner's discard pile (FR-MNV-001)")
	}
}

// Tests for T-0126 (FR-MNV-001, FR-KWD-002m, FR-ZON-005, FR-TGT-001, FR-MOD-001):
// Animoceps Spores, loaded from data/cards. Player 0 enacts; player 1 is the
// opponent.
//
// Animoceps Spores: "Scratch 2, then return an allied Scratch Company unit to
// its owner's hand."

// spFaceDown returns p's face-down Scratch Company assets in Reserves, in the
// order they entered.
func spFaceDown(p *Player) []*Asset {
	var out []*Asset
	for _, a := range p.Reserves {
		if a.FaceDown {
			out = append(out, a)
		}
	}
	return out
}

// spScript installs a chooser for seat 0 that records every Decision. On a
// DecidePick it answers with choose(d), which returns the option index, or -1
// for a timeout.
func spScript(s *scene, choose func(d Decision) int) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		if d.Kind == DecidePick {
			if i := choose(d); i >= 0 {
				return Answer{Picks: []int{i}}
			}
		}
		return Answer{Timeout: true}
	}
}

// spPicks returns the DecidePick decisions asked of seat 0.
func spPicks(s *scene) []Decision {
	var out []Decision
	for _, d := range s.asked(0) {
		if d.Kind == DecidePick {
			out = append(out, d)
		}
	}
	return out
}

// spOptionFor returns the index of the option referring to asset a, or -1.
func spOptionFor(d Decision, a *Asset) int {
	for i, o := range d.Options {
		if o.Ref == a.ID {
			return i
		}
	}
	return -1
}

func TestCard_AnimocepsSpores_ScratchesTwoThenReturnsThePickedOne(t *testing.T) {
	// Covers: FR-KWD-002m, FR-ZON-005, FR-TGT-001
	s := newScene(t)
	top := s.onTop(0, unit("TopA", 1, 1, Keywords{}), unit("TopB", 1, 1, Keywords{}))
	spScript(s, func(d Decision) int { return 0 })
	inst := enactFor(s, "Animoceps Spores")
	s.do(enactAction(inst))
	p := s.g.Players[0]
	picks := spPicks(s)
	if len(picks) != 1 || picks[0].Player != 0 {
		t.Fatalf("one DecidePick is asked of player 0 for the return, got %d (FR-TGT-001)", len(picks))
	}
	if len(picks[0].Options) != 2 {
		t.Errorf("the pick offers exactly the two new Scratch Company units, got %d options (FR-TGT-001)", len(picks[0].Options))
	}
	h0, h1 := inHand(p, top[0]), inHand(p, top[1])
	if h0 == h1 {
		t.Fatalf("exactly one of the two scratched cards is returned to hand: top0 in hand=%v, top1 in hand=%v (FR-ZON-005)", h0, h1)
	}
	stay := top[0]
	if h0 {
		stay = top[1]
	}
	rest := spFaceDown(p)
	if len(rest) != 1 || rest[0].Inst != stay || !rest[0].Exhausted {
		t.Errorf("the other Scratch Company stays in Reserves, face-down and Exhausted (FR-KWD-002m)")
	}
	if !inDiscardPile(p, inst) {
		t.Errorf("the resolved Maneuver ends in its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_AnimocepsSpores_MayReturnAnOlderScratchCompany(t *testing.T) {
	// Covers: SVY-R3
	s := newScene(t)
	s.onTop(0, unit("TopA", 1, 1, Keywords{}), unit("TopB", 1, 1, Keywords{}))
	old := s.put(0, ZoneMainTheater, unit("Older", 1, 1, Keywords{}))
	old.FaceDown = true
	spScript(s, func(d Decision) int { return spOptionFor(d, old) })
	s.do(enactAction(enactFor(s, "Animoceps Spores")))
	p := s.g.Players[0]
	picks := spPicks(s)
	if len(picks) != 1 {
		t.Fatalf("one pick question, got %d", len(picks))
	}
	if len(picks[0].Options) != 3 || spOptionFor(picks[0], old) < 0 {
		t.Errorf("the pick offers the older Scratch Company and both new ones: %d options (SVY-R3)", len(picks[0].Options))
	}
	if s.alive(old) || !inHand(p, old.Inst) {
		t.Errorf("the picked older Scratch Company returns to its owner's hand (FR-ZON-005)")
	}
	if n := len(spFaceDown(p)); n != 2 {
		t.Errorf("both new Scratch Companies stay in Reserves, got %d", n)
	}
}

func TestCard_AnimocepsSpores_TimeoutReturnsTheFirstInFixedOrder(t *testing.T) {
	// Covers: SVY-R16
	s := newScene(t)
	top := s.onTop(0, unit("TopA", 1, 1, Keywords{}), unit("TopB", 1, 1, Keywords{}))
	old := s.put(0, ZoneMainTheater, unit("Older", 1, 1, Keywords{}))
	old.FaceDown = true
	spScript(s, func(d Decision) int { return -1 })
	s.do(enactAction(enactFor(s, "Animoceps Spores")))
	p := s.g.Players[0]
	if !inHand(p, top[0]) {
		t.Errorf("a timeout returns the first scratched unit: Reserves come before the Main Theater and it entered first (FR-MOD-001 §6, SVY-R16)")
	}
	if inHand(p, top[1]) || !s.alive(old) {
		t.Errorf("only the first option in fixed order returns on a timeout (SVY-R16)")
	}
}

func TestCard_AnimocepsSpores_OnlyAlliedScratchCompanyUnitsAreOffered(t *testing.T) {
	// Covers: SVY-R3
	s := newScene(t)
	s.onTop(0, unit("TopA", 1, 1, Keywords{}), unit("TopB", 1, 1, Keywords{}))
	foe := s.putScratch(1)
	plain := s.put(0, ZoneMainTheater, unit("Plain", 1, 1, Keywords{}))
	spScript(s, func(d Decision) int { return 0 })
	s.do(enactAction(enactFor(s, "Animoceps Spores")))
	picks := spPicks(s)
	if len(picks) != 1 {
		t.Fatalf("one pick question, got %d", len(picks))
	}
	if spOptionFor(picks[0], foe) >= 0 {
		t.Errorf("an opposing Scratch Company is not allied and is never offered")
	}
	if spOptionFor(picks[0], plain) >= 0 {
		t.Errorf("an allied non-Scratch unit is never offered")
	}
	if len(picks[0].Options) != 2 {
		t.Errorf("only the two new allied Scratch Companies are offered, got %d", len(picks[0].Options))
	}
	if !s.alive(foe) || !s.alive(plain) {
		t.Errorf("the opposing and non-Scratch units are untouched")
	}
}

func TestCard_AnimocepsSpores_OneCardDeckReturnsItWithoutAsking(t *testing.T) {
	// Covers: FR-KWD-002m, FR-TGT-001
	s := newScene(t)
	top := s.onTop(0, unit("OnlyCard", 1, 1, Keywords{}))
	s.g.Players[0].Deck = top
	spScript(s, func(d Decision) int { return 0 })
	s.do(enactAction(enactFor(s, "Animoceps Spores")))
	p := s.g.Players[0]
	if got := spPicks(s); len(got) != 0 {
		t.Errorf("a sole option is not asked of the Chooser, got %d pick questions (FR-TGT-001)", len(got))
	}
	if !inHand(p, top[0]) || len(spFaceDown(p)) != 0 {
		t.Errorf("the one scratched unit is returned to hand (FR-KWD-002m, FR-ZON-005)")
	}
	if s.g.Outcome.Over {
		t.Errorf("a scratch is not a draw: player 0 has not lost (FR-KWD-002m)")
	}
}

func TestCard_AnimocepsSpores_EmptyDeckDoesNothing(t *testing.T) {
	// Covers: FR-KWD-002m
	s := newScene(t)
	s.g.Players[0].Deck = nil
	spScript(s, func(d Decision) int { return 0 })
	inst := enactFor(s, "Animoceps Spores")
	handBefore := len(s.g.Players[0].Hand)
	s.do(enactAction(inst))
	p := s.g.Players[0]
	if len(spFaceDown(p)) != 0 {
		t.Errorf("an empty deck scratches nothing (FR-KWD-002m)")
	}
	if got := s.asked(0); len(got) != 0 {
		t.Errorf("nothing is asked, got %d decisions", len(got))
	}
	if len(p.Hand) != handBefore-1 {
		t.Errorf("nothing returns to hand: only the enacted Maneuver left it, hand %d -> %d", handBefore, len(p.Hand))
	}
	if s.g.Outcome.Over {
		t.Errorf("a scratch is not a draw: player 0 has not lost (FR-KWD-002m)")
	}
	if !inDiscardPile(p, inst) {
		t.Errorf("the Maneuver still goes to its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_AnimocepsSpores_PaysTwoTactics(t *testing.T) {
	// Covers: SVY-R29
	s := newScene(t)
	inst := enactFor(s, "Animoceps Spores")
	if !offersEnact(s.g, inst) {
		t.Fatalf("with no unit on the battlefield it is still offered: it has no target slot (SVY-R29)")
	}
	s.do(enactAction(inst))
	if p := s.g.Players[0]; p.Tac != 3 || p.Req != 3 {
		t.Errorf("enacting costs 2 Tactics and no Requisition: pool R%d T%d, want R3 T3 (FR-MNV-001)", p.Req, p.Tac)
	}
}

// Tests for T-0127 (FR-MNV-001, FR-ZON-002, FR-ZON-004, FR-LAY-001, FR-LAY-003):
// Rapturous Sermon, loaded from data/cards. Player 0 enacts.
//
// Rapturous Sermon: "Move all allied Scratch Company units to the Main
// Theater. Allied Scratch Company units get +2 power until end of turn."

// rsScratchInMain puts a face-down Scratch Company unit of p in the Main Theater.
func rsScratchInMain(s *scene, p PlayerID) *Asset {
	a := s.putScratch(p)
	s.moveTo(a, ZoneMainTheater)
	return a
}

// rsMoveEventsNaming counts move events in the log that name asset a.
func rsMoveEventsNaming(s *scene, a *Asset) int {
	n := 0
	for _, e := range eventsOfKind(s.g.Log, EventMove) {
		if e.Target == a.ID || e.Source == a.ID {
			n++
		}
	}
	return n
}

func TestCard_RapturousSermon_MovesAlliedScratchCompanyToTheMainTheater(t *testing.T) {
	// Covers: FR-ZON-001, FR-ZON-002
	s := newScene(t)
	r1, r2 := s.putScratch(0), s.putScratch(0)
	m := rsScratchInMain(s, 0)
	s.do(enactAction(enactFor(s, "Rapturous Sermon")))
	for _, a := range []*Asset{r1, r2, m} {
		if s.g.Asset(a.ID) != a {
			t.Errorf("a moved unit is still the same asset (FR-ZON-002)")
		}
		if a.Zone != ZoneMainTheater || !containsAsset(s.g.TheaterOf(0), a) {
			t.Errorf("every allied Scratch Company unit ends in player 0's Main Theater (FR-ZON-002)")
		}
	}
	if len(s.g.Players[0].Reserves) != 0 {
		t.Errorf("no allied unit is left in the Reserves, got %d (FR-ZON-001)", len(s.g.Players[0].Reserves))
	}
	for _, a := range []*Asset{r1, r2} {
		if n := rsMoveEventsNaming(s, a); n != 1 {
			t.Errorf("a unit that moved has exactly one move event, got %d (FR-ZON-002)", n)
		}
	}
	if n := rsMoveEventsNaming(s, m); n != 0 {
		t.Errorf("a unit already in the Main Theater does not move: %d move events, want 0 (FR-ZON-002)", n)
	}
}

func containsAsset(list []*Asset, a *Asset) bool {
	for _, x := range list {
		if x == a {
			return true
		}
	}
	return false
}

func TestCard_RapturousSermon_AlliedScratchCompanyGetPlusTwo(t *testing.T) {
	s := newScene(t)
	r1, r2 := s.putScratch(0), s.putScratch(0)
	m := rsScratchInMain(s, 0)
	s.do(enactAction(enactFor(s, "Rapturous Sermon")))
	for _, a := range []*Asset{r1, r2, m} {
		if got := s.g.Power(a); got != 3 {
			t.Errorf("each allied Scratch Company unit has power 3 (printed 1, +2), got %d", got)
		}
	}
}

func TestCard_RapturousSermon_OthersAreUntouched(t *testing.T) {
	s := newScene(t)
	s.putScratch(0)
	plain := s.put(0, ZoneReserves, unit("Plain", 2, 3, Keywords{}))
	foe := s.putScratch(1)
	s.do(enactAction(enactFor(s, "Rapturous Sermon")))
	if plain.Zone != ZoneReserves || !containsAsset(s.g.Players[0].Reserves, plain) {
		t.Errorf("an allied non-Scratch unit stays in the Reserves")
	}
	if foe.Zone != ZoneReserves || !containsAsset(s.g.Players[1].Reserves, foe) {
		t.Errorf("an opposing Scratch Company unit stays in its controller's Reserves")
	}
	if got := s.g.Power(plain); got != 2 {
		t.Errorf("an allied non-Scratch unit keeps its printed power 2, got %d", got)
	}
	if got := s.g.Power(foe); got != 1 {
		t.Errorf("an opposing Scratch Company unit keeps its printed power 1, got %d", got)
	}
	for _, a := range []*Asset{plain, foe} {
		if n := rsMoveEventsNaming(s, a); n != 0 {
			t.Errorf("a unit the card does not select has no move event, got %d", n)
		}
	}
}

func TestCard_RapturousSermon_MovedUnitsKeepTheirState(t *testing.T) {
	// Covers: FR-ZON-004
	s := newScene(t)
	a := s.putScratch(0)
	a.Exhausted = true
	a.Armor = 1
	s.do(enactAction(enactFor(s, "Rapturous Sermon")))
	if a.Zone != ZoneMainTheater {
		t.Fatalf("the unit moves to the Main Theater")
	}
	if !a.Exhausted {
		t.Errorf("a unit moved between Reserves and the Main Theater keeps its readiness (FR-ZON-004)")
	}
	if a.Armor != 1 {
		t.Errorf("a moved unit keeps its Armor: got %d, want 1 (FR-ZON-004)", a.Armor)
	}
	if got := s.g.Power(a); got != 3 {
		t.Errorf("the moved unit has power 3, got %d", got)
	}
}

func TestCard_RapturousSermon_LaterScratchCompanyDoesNotGetPlusTwo(t *testing.T) {
	// Covers: FR-LAY-001
	s := newScene(t)
	early := s.putScratch(0)
	s.do(enactAction(enactFor(s, "Rapturous Sermon")))
	later := s.putScratch(0) // enters the battlefield after the +2 resolved, same turn
	if got := s.g.Power(later); got != 1 {
		t.Errorf("a unit that becomes an allied Scratch Company later gets no +2 (FR-LAY-001): power %d, want 1", got)
	}
	if got := s.g.Power(early); got != 3 {
		t.Errorf("the earlier unit keeps its +2 (FR-LAY-001): power %d, want 3", got)
	}
}

func TestCard_RapturousSermon_PlusTwoExpiresAtEndOfTurn(t *testing.T) {
	// Covers: FR-LAY-003
	s := newScene(t)
	r := s.putScratch(0)
	m := rsScratchInMain(s, 0)
	s.do(enactAction(enactFor(s, "Rapturous Sermon")))
	s.endTurn()
	for _, a := range []*Asset{r, m} {
		if got := s.g.Power(a); got != 1 {
			t.Errorf("the +2 expires at the end of the turn: power %d, want 1 (FR-LAY-003)", got)
		}
		if a.Zone != ZoneMainTheater {
			t.Errorf("the units stay in the Main Theater when the turn ends (FR-ZON-002)")
		}
	}
}

func TestCard_RapturousSermon_EnactableWithNoScratchCompany(t *testing.T) {
	// Covers: SVY-R29
	s := newScene(t)
	inst := enactFor(s, "Rapturous Sermon")
	if !offersEnact(s.g, inst) {
		t.Fatalf("a Maneuver with no target slot is a legal Enact with no Scratch Company unit (SVY-R29)")
	}
	s.do(enactAction(inst))
	if n := len(eventsOfKind(s.g.Log, EventMove)); n != 0 {
		t.Errorf("nothing matched, so nothing moves: %d move events, want 0", n)
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver goes to its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_RapturousSermon_AsksNothingAndPaysFourTactics(t *testing.T) {
	// Covers: FR-TGT-002
	s := newScene(t)
	s.putScratch(0)
	s.script(0)
	s.script(1)
	s.do(enactAction(enactFor(s, "Rapturous Sermon")))
	for p := PlayerID(0); p < 2; p++ {
		if got := s.asked(p); len(got) != 0 {
			t.Errorf("\"all\" is not targeted: player %d was asked %d decisions, want none (FR-TGT-002)", p, len(got))
		}
	}
	pl := s.g.Players[0]
	if pl.Tac != 1 || pl.Req != 3 {
		t.Errorf("enacting costs 4 Tactics and no Requisition: pool R%d T%d, want R3 T1 (FR-MNV-001)", pl.Req, pl.Tac)
	}
}

// Rhapsody's Embrace (T-0128): "As an additional cost to enact this maneuver,
// kill an allied unit." / "Kill chosen opposing unit." Costs 0 Requisition, 3
// Tactics. Feed the Despoiler: the same cost / "Draw 2 cards." Costs 1 Tactics.

func TestFR_TRN_003a_RhapsodysEmbrace(t *testing.T) {
	// Covers: SVY-R29, FR-TGT-001
	s := newScene(t)
	a1 := s.put(0, ZoneMainTheater, unit("Ally One", 1, 5, Keywords{}))
	a2 := s.put(0, ZoneMainTheater, unit("Ally Two", 1, 5, Keywords{}))
	f1 := s.put(1, ZoneMainTheater, unit("Foe One", 1, 5, Keywords{}))
	f2 := s.put(1, ZoneMainTheater, unit("Foe Two", 1, 5, Keywords{}))
	inst := enactFor(s, "Rhapsody's Embrace")
	kcScript(s, f2, a2)
	s.do(enactAction(inst))
	p0, p1 := s.g.Players[0], s.g.Players[1]
	if s.alive(a2) || !inDiscard(p0, a2.Inst) || !s.alive(a1) {
		t.Errorf("the chosen allied unit is killed as the cost and no other (FR-TRN-003a)")
	}
	if s.alive(f2) || !inDiscard(p1, f2.Inst) || !s.alive(f1) {
		t.Errorf("the chosen opposing unit is killed and no other (printed effect)")
	}
	if lastDiscard(p0) != inst {
		t.Errorf("the Maneuver is in its owner's discard pile")
	}
	if p0.Req != 3 || p0.Tac != 2 {
		t.Errorf("it costs 3 Tactics and no Requisition: pool R%d T%d, want R3 T2", p0.Req, p0.Tac)
	}

	t.Run("no opposing unit", func(t *testing.T) {
		s := newScene(t)
		s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
		inst := enactFor(s, "Rhapsody's Embrace")
		if offersEnact(s.g, inst) {
			t.Errorf("without an opposing unit to kill the Maneuver is not offered, though an allied unit could pay (SVY-R29)")
		}
		s.reject(enactAction(inst))
		if !inHand(s.g.Players[0], inst) {
			t.Errorf("the card stays in hand (SVY-R29)")
		}
	})
	t.Run("no allied unit", func(t *testing.T) {
		s := newScene(t)
		s.put(1, ZoneMainTheater, unit("Foe", 1, 5, Keywords{}))
		inst := enactFor(s, "Rhapsody's Embrace")
		if offersEnact(s.g, inst) {
			t.Errorf("without an allied unit to kill the Maneuver is not offered (SVY-R29, FR-TRN-003a)")
		}
		s.reject(enactAction(inst))
		if !inHand(s.g.Players[0], inst) {
			t.Errorf("the card stays in hand (SVY-R29)")
		}
	})
}

func TestFR_TRN_003a_FeedTheDespoiler(t *testing.T) {
	// Covers: SVY-R29
	s := newScene(t)
	a1 := s.put(0, ZoneMainTheater, unit("Ally One", 1, 5, Keywords{}))
	a2 := s.put(0, ZoneReserves, unit("Ally Two", 1, 5, Keywords{}))
	inst := enactFor(s, "Feed the Despoiler")
	p := s.g.Players[0]
	hand, deck := len(p.Hand), len(p.Deck)
	kcScript(s, nil, a1)
	s.do(enactAction(inst))
	if s.alive(a1) || !inDiscard(p, a1.Inst) || !s.alive(a2) {
		t.Errorf("the chosen allied unit is killed as the cost and no other (FR-TRN-003a)")
	}
	if got := len(p.Hand) - (hand - 1); got != 2 {
		t.Errorf("its controller draws 2 cards: hand grew by %d beyond the Maneuver leaving, want 2", got)
	}
	if got := deck - len(p.Deck); got != 2 {
		t.Errorf("2 cards left the deck, got %d", got)
	}
	if p.Req != 3 || p.Tac != 4 {
		t.Errorf("it costs 1 Tactics: pool R%d T%d, want R3 T4", p.Req, p.Tac)
	}

	t.Run("a Scratch Company can pay", func(t *testing.T) {
		s := newScene(t)
		sc := s.putScratch(0)
		inst := enactFor(s, "Feed the Despoiler")
		if !offersEnact(s.g, inst) {
			t.Fatalf("a face-down Scratch Company is an allied unit, so the cost is payable (FR-TRN-003a)")
		}
		s.do(enactAction(inst))
		if s.alive(sc) || !inDiscard(s.g.Players[0], sc.Inst) {
			t.Errorf("the Scratch Company is killed as the cost (FR-TRN-003a)")
		}
	})
	t.Run("no allied unit", func(t *testing.T) {
		s := newScene(t)
		inst := enactFor(s, "Feed the Despoiler")
		if offersEnact(s.g, inst) {
			t.Errorf("without an allied unit to kill the Maneuver is not offered (SVY-R29)")
		}
	})
}
