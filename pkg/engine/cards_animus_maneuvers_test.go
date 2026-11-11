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

// Funeral Fungus (T-0130): "As an additional cost to enact this maneuver,
// exhaust a readied allied unit." / "Return chosen unit card from your discard
// pile to your hand." Costs 0 Requisition, 1 Tactics.

// ffChoose answers every DecideTarget and DecidePick (the cost payer) with the
// option referring to whichever of prefer is offered first (an *Asset or an
// *Instance, by ID), and every other decision with its first option. It
// records the decisions in s.asks[0].
func ffChoose(s *scene, prefer ...any) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		if d.Kind == DecideTarget || d.Kind == DecidePick {
			for _, want := range prefer {
				id := 0
				switch v := want.(type) {
				case *Asset:
					id = v.ID
				case *Instance:
					id = v.ID
				}
				for i, o := range d.Options {
					if o.Ref == id {
						return Answer{Picks: []int{i}}
					}
				}
			}
		}
		return Answer{Picks: []int{0}}
	}
}

func TestCard_FuneralFungus_ExhaustsAndReturnsTheChosenUnitCard(t *testing.T) {
	// Covers: FR-TRN-003a, FR-MNV-001, FR-TGT-001
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	other := s.put(0, ZoneReserves, unit("Bystander", 1, 5, Keywords{}))
	cards := s.inDiscard(0, unit("Older", 1, 1, Keywords{}), unit("Newer", 1, 1, Keywords{}))
	inst := enactFor(s, "Funeral Fungus")
	p := s.g.Players[0]
	ffChoose(s, cards[0], other)
	s.do(enactAction(inst))
	if !other.Exhausted {
		t.Errorf("the Readied allied unit the payer chose (Bystander) is exhausted as the additional cost (FR-TRN-003a)")
	}
	if ally.Exhausted {
		t.Errorf("no other asset is exhausted: the unchosen Readied allied unit stays Readied (FR-TRN-003a)")
	}
	if !inHand(p, cards[0]) || inDiscard(p, cards[0]) {
		t.Errorf("the chosen (older) unit card returns to its controller's hand, not the top one")
	}
	if !inDiscard(p, cards[1]) || inHand(p, cards[1]) {
		t.Errorf("the unchosen unit card stays in the discard pile")
	}
	if lastDiscard(p) != inst {
		t.Errorf("the Maneuver goes to its owner's discard pile after resolving (FR-MNV-001)")
	}
	if p.Req != 3 || p.Tac != 4 {
		t.Errorf("it costs 1 Tactics and no Requisition: pool R%d T%d, want R3 T4", p.Req, p.Tac)
	}
}

func TestCard_FuneralFungus_OnlyAReadiedAlliedUnitPays(t *testing.T) {
	// Covers: SVY-R29, FR-TRN-003a
	s := newScene(t)
	tired := s.put(0, ZoneMainTheater, unit("Tired Ally", 1, 5, Keywords{}))
	tired.Exhausted = true
	wall := s.put(0, ZoneMainTheater, structure("Ready Wall", 0, 5, Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, Keywords{}))
	card := s.inDiscard(0, unit("Fallen", 1, 1, Keywords{}))[0]
	inst := enactFor(s, "Funeral Fungus")
	p := s.g.Players[0]
	if offersEnact(s.g, inst) {
		t.Errorf("with no Readied allied unit the Maneuver is not offered (SVY-R29, FR-TRN-003a)")
	}
	s.reject(enactAction(inst))
	if !inHand(p, inst) || inHand(p, card) {
		t.Errorf("the rejected Maneuver stays in hand and returns nothing")
	}

	ready := s.put(0, ZoneMainTheater, unit("Ready Ally", 1, 5, Keywords{}))
	if !offersEnact(s.g, inst) {
		t.Fatalf("with a Readied allied unit the Maneuver is offered (FR-TRN-003a)")
	}
	ffChoose(s, card, ready)
	s.do(enactAction(inst))
	if !ready.Exhausted {
		t.Errorf("the Readied allied unit pays the cost by exhausting")
	}
	if foe.Exhausted || !tired.Exhausted {
		t.Errorf("no other unit changes readiness: foe exhausted=%v, tired exhausted=%v, want false/true", foe.Exhausted, tired.Exhausted)
	}
	if wall.Exhausted {
		t.Errorf("the allied structure is not exhausted (the cost needs a unit)")
	}
}

func TestCard_FuneralFungus_NeedsAUnitCardInItsControllersDiscardPile(t *testing.T) {
	// Covers: SVY-R29, FR-TGT-001
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	s.inDiscard(0, maneuver("Spent Spell", 0, 0, steps()), structure("Fallen Wall", 0, 3, Keywords{}))
	s.inDiscard(1, unit("Their Fallen", 1, 1, Keywords{}))
	inst := enactFor(s, "Funeral Fungus")
	p := s.g.Players[0]
	if offersEnact(s.g, inst) {
		t.Errorf("without a unit card in its controller's discard pile the Maneuver is not offered (SVY-R29, FR-TGT-001)")
	}
	s.reject(enactAction(inst))
	if !inHand(p, inst) {
		t.Errorf("the card stays in hand (SVY-R29)")
	}
	if ally.Exhausted {
		t.Errorf("the allied unit stays Readied: nothing was paid")
	}

	s.inDiscard(0, unit("Fallen", 1, 1, Keywords{}))
	if !offersEnact(s.g, inst) {
		t.Errorf("with a unit card in its controller's discard pile the Maneuver is offered")
	}
}

func TestCard_FuneralFungus_ScratchCompanyCanPay(t *testing.T) {
	// Covers: FR-TRN-003a
	s := newScene(t)
	sc := s.putScratch(0)
	sc.Exhausted = false
	card := s.inDiscard(0, unit("Fallen", 1, 1, Keywords{}))[0]
	inst := enactFor(s, "Funeral Fungus")
	if !offersEnact(s.g, inst) {
		t.Fatalf("a Readied face-down Scratch Company is a Readied allied unit, so the cost is payable (FR-TRN-003a)")
	}
	ffChoose(s, card, sc)
	s.do(enactAction(inst))
	if !sc.Exhausted {
		t.Errorf("the Scratch Company is Exhausted after paying the cost")
	}
	if !inHand(s.g.Players[0], card) {
		t.Errorf("the unit card returns to hand")
	}
}

func TestCard_FuneralFungus_ShamblerTriggerResolvesFirst(t *testing.T) {
	// Covers: FR-TRN-003a, SVY-R17, FR-CAP-001
	s := newScene(t)
	sh := putShambler(s, 0, ZoneMainTheater, 0)
	hurt := s.put(0, ZoneMainTheater, structure("Hurt Wall", 0, 5, Keywords{}))
	hurt.Damage = 3
	card := s.inDiscard(0, unit("Fallen", 1, 1, Keywords{}))[0]
	inst := enactFor(s, "Funeral Fungus")
	p := s.g.Players[0]
	ffChoose(s, card, hurt, sh)

	s.do(enactAction(inst))

	if !sh.Exhausted {
		t.Errorf("the Shambler is exhausted as the cost")
	}
	if hurt.Damage != 2 {
		t.Errorf("the Shambler's trigger heals the damaged allied asset 1: damage %d, want 2 (SVY-R17)", hurt.Damage)
	}
	if !inHand(p, card) {
		t.Fatalf("the unit card returns to hand")
	}
	manPlace, trigPlace := 0, 0
	for _, e := range s.g.Log {
		switch {
		case e.Kind == "place" && e.Card == "Funeral Fungus":
			manPlace = e.Seq
			if e.Amount != 1 {
				t.Errorf("the Maneuver is placed at stack depth 1, got %d (FR-TRN-003a)", e.Amount)
			}
		case e.Kind == "place" && e.Card == "Animoceps Shambler":
			trigPlace = e.Seq
			if e.Amount != 2 {
				t.Errorf("the Shambler's trigger is placed on top of the Maneuver, depth 2, got %d (FR-TRN-003a, SVY-R17)", e.Amount)
			}
		}
	}
	if manPlace == 0 || trigPlace == 0 {
		t.Fatalf("missing place events: Maneuver %d, Shambler trigger %d", manPlace, trigPlace)
	}
	if trigPlace < manPlace {
		t.Errorf("the trigger goes onto the stack after the Maneuver was declared, above it, so it resolves first (FR-TRN-003a)")
	}
}

// Tests for T-0133 (FR-MNV-001, FR-TGT-001, FR-KWD-002h, FR-ZON-004,
// FR-ZON-005): the Animus Maneuver Disorienting Pulse, loaded from data/cards.
//
// Disorienting Pulse: "Return chosen opposing unit in the Main Theater to its
// owner's hand." Cost 1 Tactics.

func TestCard_DisorientingPulse_ReturnsChosenOpposingUnitToItsOwnersHand(t *testing.T) {
	// Covers: FR-MNV-001, FR-TGT-001, FR-ZON-004
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 3, Keywords{}))
	other := s.put(1, ZoneMainTheater, unit("Other", 1, 3, Keywords{}))
	inst := enactFor(s, "Disorienting Pulse")
	chooseAsset(s, foe)
	s.do(enactAction(inst))
	if s.alive(foe) || !inHand(s.g.Players[1], foe.Inst) {
		t.Errorf("the chosen opposing unit leaves the battlefield and is in its owner's hand (FR-MNV-001, FR-ZON-004)")
	}
	if !s.alive(other) {
		t.Errorf("the unchosen opposing unit stays on the battlefield (FR-TGT-001)")
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver ends in its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_DisorientingPulse_OnlyOpposingMainTheaterUnitsAreOffered(t *testing.T) {
	// Covers: FR-TGT-001
	s := newScene(t)
	u1 := s.put(1, ZoneMainTheater, unit("E1", 1, 3, Keywords{}))
	u2 := s.put(1, ZoneMainTheater, unit("E2", 1, 3, Keywords{}))
	res := s.put(1, ZoneReserves, unit("ERes", 1, 3, Keywords{}))
	st := s.put(1, ZoneMainTheater, structure("EWall", 0, 3, Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 3, Keywords{}))
	chooseAsset(s, u1)
	s.do(enactAction(enactFor(s, "Disorienting Pulse")))
	got := targetOffers(s)
	if len(got) != 2 || !got[u1.ID] || !got[u2.ID] {
		t.Errorf("exactly the two opposing Main Theater units are offered, got refs %v (FR-TGT-001)", got)
	}
	for _, a := range []*Asset{res, st, ally, s.g.Players[0].HQ, s.g.Players[1].HQ} {
		if got[a.ID] {
			t.Errorf("%s is not an opposing Main Theater unit and must not be offered (FR-TGT-001)", a)
		}
	}
}

func TestCard_DisorientingPulse_WardUnitCannotBeChosen(t *testing.T) {
	// Covers: FR-KWD-002h
	s := newScene(t)
	ward := s.put(1, ZoneMainTheater, unit("Warded", 1, 3, Keywords{Ward: true}))
	p1 := s.put(1, ZoneMainTheater, unit("Plain", 1, 3, Keywords{}))
	p2 := s.put(1, ZoneMainTheater, unit("Plain2", 1, 3, Keywords{}))
	chooseAsset(s, p1)
	s.do(enactAction(enactFor(s, "Disorienting Pulse")))
	got := targetOffers(s)
	if got[ward.ID] || len(got) != 2 || !got[p1.ID] || !got[p2.ID] {
		t.Errorf("an opposing Ward unit is not offered, only the two plain ones are, got refs %v (FR-KWD-002h)", got)
	}
	if !s.alive(ward) {
		t.Errorf("the Ward unit stays on the battlefield (FR-KWD-002h)")
	}
}

func TestCard_DisorientingPulse_NotEnactableWithoutLegalTarget(t *testing.T) {
	// Covers: SVY-R29, FR-MNV-001
	cases := []struct {
		name  string
		setup func(s *scene)
		want  bool
	}{
		{"opposing units only in Reserves", func(s *scene) {
			s.put(1, ZoneReserves, unit("Res", 1, 3, Keywords{}))
		}, false},
		{"only an opposing Ward unit in the Main Theater", func(s *scene) {
			s.put(1, ZoneMainTheater, unit("Warded", 1, 3, Keywords{Ward: true}))
		}, false},
		{"only allied units in the Main Theater", func(s *scene) {
			s.put(0, ZoneMainTheater, unit("Ally", 1, 3, Keywords{}))
		}, false},
		{"one plain opposing unit in the Main Theater", func(s *scene) {
			s.put(1, ZoneMainTheater, unit("Foe", 1, 3, Keywords{}))
		}, true},
	}
	for _, c := range cases {
		s := newScene(t)
		c.setup(s)
		inst := enactFor(s, "Disorienting Pulse")
		if got := offersEnact(s.g, inst); got != c.want {
			t.Errorf("%s: Enact offered = %v, want %v (SVY-R29)", c.name, got, c.want)
		}
		if !c.want {
			s.reject(enactAction(inst))
			if !inHand(s.g.Players[0], inst) {
				t.Errorf("%s: the card stays in hand (SVY-R29)", c.name)
			}
		}
	}
}

func TestCard_DisorientingPulse_FaceDownScratchCompanyReturnsAsItsCard(t *testing.T) {
	// Covers: FR-ZON-004, FR-ZON-005, FR-TGT-001
	s := newScene(t)
	sc := s.putScratch(1)
	s.moveTo(sc, ZoneMainTheater)
	chooseAsset(s, sc)
	s.do(enactAction(enactFor(s, "Disorienting Pulse")))
	if s.alive(sc) || !inHand(s.g.Players[1], sc.Inst) {
		t.Errorf("an opposing face-down Scratch Company is a unit: it is chosen and its printed card goes to its owner's hand (FR-ZON-004, FR-ZON-005)")
	}
}

func TestCard_DisorientingPulse_PaysOneTactics(t *testing.T) {
	// Covers: FR-MNV-001
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 3, Keywords{}))
	chooseAsset(s, foe)
	s.do(enactAction(enactFor(s, "Disorienting Pulse")))
	if p := s.g.Players[0]; p.Tac != 4 || p.Req != 3 {
		t.Errorf("enacting costs 1 Tactics and no Requisition: pool R%d T%d, want R3 T4 (FR-MNV-001)", p.Req, p.Tac)
	}
}

// Tests for T-0212 (FR-ENT-103, FR-MNV-001, FR-TGT-001, FR-LAY-001,
// FR-LAY-002, FR-KWD-002h): the Animus Maneuver Forced Mutation, loaded from
// data/cards. Player 0 enacts; player 1 is the opponent.
//
// Forced Mutation: "Turn chosen opposing unit face down. It becomes a Scratch
// Company unit." Cost 4 Tactics.

func TestCard_ForcedMutation_TurnsChosenOpposingUnitFaceDownInPlace(t *testing.T) {
	// Covers: FR-ENT-103, FR-MNV-001, FR-LAY-001
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Bulwark", 4, 5, Keywords{Stalwart: true}))
	foe.Exhausted, foe.Armor = true, 2
	other := s.put(1, ZoneMainTheater, unit("Other", 1, 3, Keywords{}))
	id := foe.ID
	inst := enactFor(s, "Forced Mutation")
	chooseAsset(s, foe)
	s.do(enactAction(inst))
	a := s.g.Asset(id)
	if a == nil || a.Zone != ZoneMainTheater || a.Controller != 1 {
		t.Fatalf("the chosen unit stays the same asset in player 1's Main Theater (FR-ENT-103): %v", a)
	}
	if !a.FaceDown || !a.Exhausted || a.Armor != 2 {
		t.Errorf("it is face down and keeps its readiness and Armor: FaceDown %v, Exhausted %v, Armor %d; want true, true, 2 (FR-ENT-103)", a.FaceDown, a.Exhausted, a.Armor)
	}
	if got := s.g.Power(a); got != 1 {
		t.Errorf("effective power = %d, want 1 (FR-LAY-001, FR-ENT-103)", got)
	}
	if got := s.g.MaxHealth(a); got != 1 {
		t.Errorf("effective Maximum Health = %d, want 1 (FR-LAY-001, FR-ENT-103)", got)
	}
	if got := s.g.Keywords(a); got != (Keywords{DeploysExhausted: true}) {
		t.Errorf("effective keywords = %+v, want Deploys Exhausted only (FR-ENT-103)", got)
	}
	if other.FaceDown {
		t.Errorf("the unchosen opposing unit is not turned face down (FR-TGT-001)")
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver ends in its owner's discard pile (FR-MNV-001)")
	}
	if p := s.g.Players[0]; p.Tac != 1 || p.Req != 3 {
		t.Errorf("enacting costs 4 Tactics and no Requisition: pool R%d T%d, want R3 T1 (FR-MNV-001)", p.Req, p.Tac)
	}
}

func TestCard_ForcedMutation_DamagedTargetIsDestroyed(t *testing.T) {
	// Covers: FR-ENT-103, FR-LAY-002
	s := newScene(t)
	card := unit("Martyr", 3, 4, Keywords{})
	foe := s.put(1, ZoneMainTheater, card)
	foe.Damage = 1
	inst := foe.Inst
	s.do(enactAction(enactFor(s, "Forced Mutation")))
	if s.alive(foe) {
		t.Fatalf("1 damage on a Maximum Health of 1 is lethal: the unit is destroyed in the sweep after the item (FR-ENT-103, FR-LAY-002)")
	}
	if lastDiscard(s.g.Players[1]) != inst {
		t.Errorf("its own card is in player 1's discard pile (FR-LAY-002)")
	}
}

func TestCard_ForcedMutation_OnlyOpposingUnitsAreOffered(t *testing.T) {
	// Covers: FR-TGT-001, FR-KWD-002h
	s := newScene(t)
	main := s.put(1, ZoneMainTheater, unit("EMain", 1, 3, Keywords{}))
	res := s.put(1, ZoneReserves, unit("ERes", 1, 3, Keywords{}))
	st := s.put(1, ZoneMainTheater, structure("EWall", 0, 3, Keywords{}))
	ward := s.put(1, ZoneMainTheater, unit("Warded", 1, 3, Keywords{Ward: true}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 3, Keywords{}))
	chooseAsset(s, res)
	s.do(enactAction(enactFor(s, "Forced Mutation")))
	got := targetOffers(s)
	if len(got) != 2 || !got[main.ID] || !got[res.ID] {
		t.Errorf("exactly the two opposing non-Ward units, in either zone, are offered, got refs %v (FR-TGT-001)", got)
	}
	for _, a := range []*Asset{st, ward, ally, s.g.Players[0].HQ, s.g.Players[1].HQ} {
		if got[a.ID] {
			t.Errorf("%s is not a legal target and must not be offered (FR-TGT-001, FR-KWD-002h)", a)
		}
	}
	if !res.FaceDown || res.Zone != ZoneReserves {
		t.Errorf("the chosen Reserves unit is face down in Reserves: FaceDown %v, zone %v (FR-ENT-103)", res.FaceDown, res.Zone)
	}
	if main.FaceDown {
		t.Errorf("the Main Theater unit stays face up (FR-TGT-001)")
	}
}

func TestCard_ForcedMutation_NotEnactableWithoutLegalTarget(t *testing.T) {
	// Covers: SVY-R29, FR-MNV-001
	cases := []struct {
		name  string
		setup func(s *scene)
		want  bool
	}{
		{"no units on the battlefield", func(s *scene) {}, false},
		{"only an opposing Structure", func(s *scene) {
			s.put(1, ZoneMainTheater, structure("EWall", 0, 3, Keywords{}))
		}, false},
		{"only an allied unit", func(s *scene) {
			s.put(0, ZoneMainTheater, unit("Ally", 1, 3, Keywords{}))
		}, false},
		{"only an opposing Ward unit", func(s *scene) {
			s.put(1, ZoneMainTheater, unit("Warded", 1, 3, Keywords{Ward: true}))
		}, false},
		{"one plain opposing unit in Reserves", func(s *scene) {
			s.put(1, ZoneReserves, unit("Res", 1, 3, Keywords{}))
		}, true},
	}
	for _, c := range cases {
		s := newScene(t)
		c.setup(s)
		inst := enactFor(s, "Forced Mutation")
		if got := offersEnact(s.g, inst); got != c.want {
			t.Errorf("%s: Enact offered = %v, want %v (SVY-R29)", c.name, got, c.want)
		}
		if !c.want {
			s.reject(enactAction(inst))
			if !inHand(s.g.Players[0], inst) {
				t.Errorf("%s: the card stays in hand (SVY-R29)", c.name)
			}
		}
	}
}

// Tests for T-0134 (FR-MNV-001, FR-TGT-001, FR-KWD-002d, FR-KWD-002h,
// FR-REP-004): the Animus Maneuver Conversion Spike, loaded from data/cards.
// Player 0 enacts; player 1 is the opponent.
//
// Conversion Spike: "Deal 1 damage to chosen opposing unit, and chosen allied
// unit or structure heals 1." Cost 2 Tactics.

// csChoose installs a chooser for seat 0 that answers each decision by picking
// the first offered option whose Ref is one of want, and records the decisions.
func csChoose(s *scene, want ...*Asset) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		for i, o := range d.Options {
			for _, w := range want {
				if o.Ref == w.ID {
					return Answer{Picks: []int{i}}
				}
			}
		}
		s.t.Errorf("none of the wanted targets is offered in %q: %v", d.Prompt, d.Options)
		return Answer{Timeout: true}
	}
}

// csTargetOffers returns the Ref set of the n-th (0-based) target decision
// asked of seat 0.
func csTargetOffers(s *scene, n int) map[int]bool {
	s.t.Helper()
	i := 0
	for _, d := range s.asked(0) {
		if d.Kind != DecideTarget {
			continue
		}
		if i == n {
			m := map[int]bool{}
			for _, o := range d.Options {
				m[o.Ref] = true
			}
			return m
		}
		i++
	}
	s.t.Fatalf("target decision %d was not asked of seat 0", n)
	return nil
}

// csWantOffers fails unless got is exactly the set of want.
func csWantOffers(t *testing.T, got map[int]bool, why string, want ...*Asset) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: offered refs %v, want exactly %d entities (FR-TGT-001)", why, got, len(want))
	}
	for _, a := range want {
		if !got[a.ID] {
			t.Errorf("%s: %s must be offered (FR-TGT-001)", why, a)
		}
	}
}

func TestCard_ConversionSpike_DamagesChosenOpposingUnitAndHealsChosenAlly(t *testing.T) {
	// Covers: FR-MNV-001, FR-TGT-001, FR-KWD-002d
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, Keywords{}))
	otherFoe := s.put(1, ZoneMainTheater, unit("OtherFoe", 1, 5, Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	otherAlly := s.put(0, ZoneMainTheater, unit("OtherAlly", 1, 5, Keywords{}))
	ally.Damage, otherAlly.Damage = 2, 2
	inst := enactFor(s, "Conversion Spike")
	csChoose(s, foe, ally)
	s.do(enactAction(inst))
	s.wantDamage(foe, 1)
	s.wantDamage(otherFoe, 0)
	s.wantDamage(ally, 1)
	s.wantDamage(otherAlly, 2)
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver ends in its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_ConversionSpike_HealsAChosenAlliedStructure(t *testing.T) {
	// Covers: FR-TGT-001, FR-KWD-002d
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, Keywords{}))
	wall := s.put(0, ZoneMainTheater, structure("AWall", 0, 5, Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	wall.Damage, ally.Damage = 3, 3
	csChoose(s, foe, wall)
	s.do(enactAction(enactFor(s, "Conversion Spike")))
	s.wantDamage(wall, 2)
	s.wantDamage(ally, 3)
	s.wantDamage(foe, 1)
}

func TestCard_ConversionSpike_OpposingSlotOffersOnlyOpposingUnits(t *testing.T) {
	// Covers: FR-TGT-001, FR-ZON-005
	s := newScene(t)
	main := s.put(1, ZoneMainTheater, unit("EMain", 1, 5, Keywords{}))
	res := s.put(1, ZoneReserves, unit("ERes", 1, 5, Keywords{}))
	sc := s.putScratch(1)
	st := s.put(1, ZoneMainTheater, structure("EWall", 0, 5, Keywords{}))
	a1 := s.put(0, ZoneMainTheater, unit("A1", 1, 5, Keywords{}))
	s.put(0, ZoneReserves, unit("A2", 1, 5, Keywords{}))
	csChoose(s, main, a1)
	s.do(enactAction(enactFor(s, "Conversion Spike")))
	got := csTargetOffers(s, 0)
	csWantOffers(t, got, "first slot (opposing unit)", main, res, sc)
	for _, a := range []*Asset{st, a1, s.g.Players[0].HQ, s.g.Players[1].HQ} {
		if got[a.ID] {
			t.Errorf("%s is not an opposing unit and must not be offered for the damage (FR-TGT-001)", a)
		}
	}
}

func TestCard_ConversionSpike_AlliedSlotOffersOnlyAlliedUnitsAndStructures(t *testing.T) {
	// Covers: FR-TGT-001
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, Keywords{}))
	s.put(1, ZoneReserves, unit("Foe2", 1, 5, Keywords{}))
	eSt := s.put(1, ZoneMainTheater, structure("EWall", 0, 5, Keywords{}))
	aMain := s.put(0, ZoneMainTheater, unit("AMain", 1, 5, Keywords{}))
	aRes := s.put(0, ZoneReserves, unit("ARes", 1, 5, Keywords{}))
	aSt := s.put(0, ZoneMainTheater, structure("AWall", 0, 5, Keywords{}))
	csChoose(s, foe, aMain)
	s.do(enactAction(enactFor(s, "Conversion Spike")))
	got := csTargetOffers(s, 1)
	csWantOffers(t, got, "second slot (allied unit or structure)", aMain, aRes, aSt)
	for _, a := range []*Asset{s.g.Players[0].HQ, s.g.Players[1].HQ, foe, eSt} {
		if got[a.ID] {
			t.Errorf("%s is not an allied unit or structure and must not be offered for the heal (FR-TGT-001)", a)
		}
	}
}

func TestCard_ConversionSpike_OpposingWardUnitCannotBeChosen(t *testing.T) {
	// Covers: FR-KWD-002h
	s := newScene(t)
	ward := s.put(1, ZoneMainTheater, unit("Warded", 1, 5, Keywords{Ward: true}))
	p1 := s.put(1, ZoneMainTheater, unit("Plain", 1, 5, Keywords{}))
	p2 := s.put(1, ZoneMainTheater, unit("Plain2", 1, 5, Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	csChoose(s, p1, ally)
	s.do(enactAction(enactFor(s, "Conversion Spike")))
	csWantOffers(t, csTargetOffers(s, 0), "first slot with an opposing Ward unit", p1, p2)
	s.wantDamage(ward, 0)
}

func TestCard_ConversionSpike_AlliedWardUnitCanBeHealed(t *testing.T) {
	// Covers: FR-KWD-002h
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, Keywords{}))
	foe2 := s.put(1, ZoneMainTheater, unit("Foe2", 1, 5, Keywords{}))
	ward := s.put(0, ZoneMainTheater, unit("AWarded", 1, 5, Keywords{Ward: true}))
	plain := s.put(0, ZoneMainTheater, unit("APlain", 1, 5, Keywords{}))
	ward.Damage = 2
	csChoose(s, foe, ward)
	s.do(enactAction(enactFor(s, "Conversion Spike")))
	csWantOffers(t, csTargetOffers(s, 1), "second slot with an allied Ward unit", ward, plain)
	s.wantDamage(ward, 1)
	s.wantDamage(foe2, 0)
}

func TestCard_ConversionSpike_HealingAnUndamagedAllyDoesNothing(t *testing.T) {
	// Covers: FR-KWD-002d
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	s.do(enactAction(enactFor(s, "Conversion Spike")))
	s.wantDamage(ally, 0)
	s.wantDamage(foe, 1)
	if !s.alive(ally) {
		t.Errorf("healing an undamaged unit leaves it on the battlefield at 0 damage (FR-KWD-002d)")
	}
}

func TestCard_ConversionSpike_ArmorAbsorbsTheDamage(t *testing.T) {
	// Covers: FR-REP-004
	s := newScene(t)
	armored := s.put(1, ZoneMainTheater, unit("Armored", 1, 5, Keywords{Armor: 1}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	s.do(enactAction(enactFor(s, "Conversion Spike")))
	if armored.Armor != 0 {
		t.Errorf("1 Armor absorbs the 1 damage: Armor %d, want 0 (FR-REP-004)", armored.Armor)
	}
	s.wantDamage(armored, 0)
	s.wantDamage(ally, 0)
}

func TestCard_ConversionSpike_KillsAOneHealthUnit(t *testing.T) {
	// Covers: FR-MNV-001
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 1, Keywords{}))
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	s.do(enactAction(enactFor(s, "Conversion Spike")))
	if s.alive(foe) {
		t.Errorf("a chosen opposing unit with 1 health dies once the Maneuver resolves (FR-MNV-001)")
	}
}

func TestCard_ConversionSpike_NotEnactableWithoutBothTargets(t *testing.T) {
	// Covers: SVY-R29, FR-MNV-001
	cases := []struct {
		name  string
		setup func(s *scene)
		want  bool
	}{
		{"allied unit but no opposing unit", func(s *scene) {
			s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
		}, false},
		{"allied unit but only an opposing Ward unit", func(s *scene) {
			s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
			s.put(1, ZoneMainTheater, unit("Warded", 1, 5, Keywords{Ward: true}))
		}, false},
		{"opposing unit but only the allied Headquarters", func(s *scene) {
			s.put(1, ZoneMainTheater, unit("Foe", 1, 5, Keywords{}))
		}, false},
		{"one plain opposing unit and one allied unit", func(s *scene) {
			s.put(1, ZoneMainTheater, unit("Foe", 1, 5, Keywords{}))
			s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
		}, true},
	}
	for _, c := range cases {
		s := newScene(t)
		c.setup(s)
		inst := enactFor(s, "Conversion Spike")
		if got := offersEnact(s.g, inst); got != c.want {
			t.Errorf("%s: Enact offered = %v, want %v (SVY-R29)", c.name, got, c.want)
		}
		if !c.want {
			s.reject(enactAction(inst))
			if !inHand(s.g.Players[0], inst) {
				t.Errorf("%s: the card stays in hand (SVY-R29)", c.name)
			}
		}
	}
}

func TestCard_ConversionSpike_PaysTwoTactics(t *testing.T) {
	// Covers: FR-MNV-001
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("Foe", 1, 5, Keywords{}))
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	s.do(enactAction(enactFor(s, "Conversion Spike")))
	if p := s.g.Players[0]; p.Tac != 3 || p.Req != 3 {
		t.Errorf("enacting costs 2 Tactics and no Requisition: pool R%d T%d, want R3 T3 (FR-MNV-001)", p.Req, p.Tac)
	}
}

// Tests for T-0142 (FR-MNV-001, FR-ZON-002, FR-LAY-001, FR-LAY-003, FR-TGT-001,
// FR-TGT-002): the Animus Maneuver Ensnaring Mist, loaded from data/cards.
// Player 0 enacts; player 1 is the opponent.
//
// Ensnaring Mist: "Move chosen unit to its controller's Reserves." / "Chosen
// unit gets -1 power until end of turn." Cost 2 Tactics. The two "chosen unit"
// clauses are two independent target slots (SVY-R35).

// emChoose installs a chooser for seat 0 that answers the n-th target decision
// by picking the option naming picks[n] (a nil or missing entry times the
// decision out), and records the decisions in s.asks.
func emChoose(s *scene, picks ...*Asset) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	n := 0
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		if d.Kind != DecideTarget {
			return Answer{Timeout: true}
		}
		i := n
		n++
		if i >= len(picks) || picks[i] == nil {
			return Answer{Timeout: true}
		}
		for j, o := range d.Options {
			if o.Ref == picks[i].ID {
				return Answer{Picks: []int{j}}
			}
		}
		s.t.Errorf("target decision %d does not offer %s: %v", i, picks[i], d.Options)
		return Answer{Timeout: true}
	}
}

func TestCard_EnsnaringMist_MovesOneUnitAndWeakensAnother(t *testing.T) {
	// Covers: FR-MNV-001, FR-ZON-002, FR-LAY-001
	s := newScene(t)
	a := s.put(1, ZoneMainTheater, unit("A", 2, 5, Keywords{}))
	b := s.put(1, ZoneMainTheater, unit("B", 3, 5, Keywords{}))
	inst := enactFor(s, "Ensnaring Mist")
	emChoose(s, a, b)
	s.do(enactAction(inst))
	if a.Zone != ZoneReserves || !containsAsset(s.g.Players[1].Reserves, a) || containsAsset(s.g.TheaterOf(1), a) {
		t.Errorf("the first chosen unit moves to its controller's Reserves (FR-ZON-002)")
	}
	if b.Zone != ZoneMainTheater || !containsAsset(s.g.TheaterOf(1), b) {
		t.Errorf("the second chosen unit is not moved (FR-MNV-001)")
	}
	if got := s.g.Power(b); got != 2 {
		t.Errorf("the second chosen unit gets -1 power: effective power %d, want 2 (FR-LAY-001)", got)
	}
	if got := s.g.Power(a); got != 2 {
		t.Errorf("the moved unit does not get -1 power: effective power %d, want 2 (FR-MNV-001)", got)
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver ends on top of its owner's discard pile (FR-MNV-001)")
	}
	if p := s.g.Players[0]; p.Tac != 3 || p.Req != 3 {
		t.Errorf("enacting costs 2 Tactics and no Requisition: pool R%d T%d, want R3 T3 (FR-MNV-001)", p.Req, p.Tac)
	}
}

func TestCard_EnsnaringMist_SameUnitFillsBothSlots(t *testing.T) {
	// Covers: SVY-R35, SVY-R26
	s := newScene(t)
	a := s.put(1, ZoneMainTheater, unit("A", 3, 5, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Other", 3, 5, Keywords{}))
	emChoose(s, a, a)
	s.do(enactAction(enactFor(s, "Ensnaring Mist")))
	if a.Zone != ZoneReserves || !containsAsset(s.g.Players[1].Reserves, a) {
		t.Errorf("a unit chosen in both slots is moved to its controller's Reserves (SVY-R35)")
	}
	if got := s.g.Power(a); got != 2 {
		t.Errorf("the unit just moved still gets -1 power: effective power %d, want 2 (SVY-R26, SVY-R35)", got)
	}
}

func TestCard_EnsnaringMist_MovedUnitKeepsItsState(t *testing.T) {
	// Covers: FR-ZON-002, FR-ZON-004
	s := newScene(t)
	a := s.put(1, ZoneMainTheater, unit("A", 2, 5, Keywords{}))
	b := s.put(1, ZoneMainTheater, unit("B", 2, 5, Keywords{}))
	a.Damage, a.Exhausted = 2, true
	id := a.ID
	emChoose(s, a, b)
	s.do(enactAction(enactFor(s, "Ensnaring Mist")))
	if a.Zone != ZoneReserves {
		t.Fatalf("the chosen unit moves to the Reserves (FR-ZON-002)")
	}
	if s.g.Asset(id) != a {
		t.Errorf("a moved unit is the same asset, not a new one (FR-ZON-002)")
	}
	if a.Damage != 2 {
		t.Errorf("a moved unit keeps its damage: %d, want 2 (FR-ZON-004)", a.Damage)
	}
	if !a.Exhausted {
		t.Errorf("a unit moved between the Main Theater and Reserves stays Exhausted (FR-ZON-004)")
	}
}

func TestCard_EnsnaringMist_MovesToItsControllersReserves(t *testing.T) {
	// Covers: FR-ZON-002
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 2, 5, Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 2, 5, Keywords{}))
	emChoose(s, ally, foe)
	s.do(enactAction(enactFor(s, "Ensnaring Mist")))
	if ally.Zone != ZoneReserves || !containsAsset(s.g.Players[0].Reserves, ally) || containsAsset(s.g.Players[1].Reserves, ally) {
		t.Errorf("an allied unit goes to player 0's Reserves, its controller's (FR-ZON-002)")
	}
	if ally.Controller != 0 {
		t.Errorf("moving a unit does not change its controller (FR-ZON-002)")
	}
}

func TestCard_EnsnaringMist_UnitAlreadyInReservesStays(t *testing.T) {
	// Covers: FR-ZON-002
	s := newScene(t)
	a := s.put(1, ZoneReserves, unit("A", 2, 5, Keywords{}))
	b := s.put(1, ZoneMainTheater, unit("B", 2, 5, Keywords{}))
	emChoose(s, a, b)
	s.do(enactAction(enactFor(s, "Ensnaring Mist")))
	if a.Zone != ZoneReserves || !containsAsset(s.g.Players[1].Reserves, a) {
		t.Errorf("a unit already in its controller's Reserves stays there (FR-ZON-002)")
	}
	if n := rsMoveEventsNaming(s, a); n != 0 {
		t.Errorf("a unit that does not move has no move event: %d, want 0 (FR-ZON-002)", n)
	}
}

func TestCard_EnsnaringMist_PowerClampsAtZero(t *testing.T) {
	// Covers: FR-LAY-001
	s := newScene(t)
	a := s.put(1, ZoneMainTheater, unit("A", 2, 5, Keywords{}))
	zero := s.put(1, ZoneMainTheater, unit("Zero", 0, 5, Keywords{}))
	emChoose(s, a, zero)
	s.do(enactAction(enactFor(s, "Ensnaring Mist")))
	if got := s.g.Power(zero); got != 0 {
		t.Errorf("effective power is never below 0: got %d, want 0 (FR-LAY-001)", got)
	}
}

func TestCard_EnsnaringMist_MinusOneExpiresAtEndOfTurn(t *testing.T) {
	// Covers: FR-LAY-003
	s := newScene(t)
	a := s.put(1, ZoneMainTheater, unit("A", 2, 5, Keywords{}))
	b := s.put(1, ZoneMainTheater, unit("B", 3, 5, Keywords{}))
	emChoose(s, a, b)
	s.do(enactAction(enactFor(s, "Ensnaring Mist")))
	if got := s.g.Power(b); got != 2 {
		t.Fatalf("B has -1 power this turn: %d, want 2 (FR-LAY-001)", got)
	}
	s.endTurn()
	if got := s.g.Power(b); got != 3 {
		t.Errorf("the -1 power expires at the end of the turn: power %d, want 3 (FR-LAY-003)", got)
	}
	if a.Zone != ZoneReserves {
		t.Errorf("the moved unit stays in the Reserves when the turn ends (FR-ZON-002)")
	}
}

func TestCard_EnsnaringMist_ChosenUnitIsAnyUnit(t *testing.T) {
	// Covers: FR-TGT-001, FR-KWD-002h
	s := newScene(t)
	aMain := s.put(0, ZoneMainTheater, unit("AMain", 2, 5, Keywords{}))
	aRes := s.put(0, ZoneReserves, unit("ARes", 2, 5, Keywords{}))
	aWard := s.put(0, ZoneMainTheater, unit("AWard", 2, 5, Keywords{Ward: true}))
	aScratch := s.putScratch(0)
	fMain := s.put(1, ZoneMainTheater, unit("FMain", 2, 5, Keywords{}))
	fRes := s.put(1, ZoneReserves, unit("FRes", 2, 5, Keywords{}))
	fWard := s.put(1, ZoneMainTheater, unit("FWard", 2, 5, Keywords{Ward: true}))
	aWall := s.put(0, ZoneMainTheater, structure("AWall", 0, 5, Keywords{}))
	fWall := s.put(1, ZoneMainTheater, structure("FWall", 0, 5, Keywords{}))
	emChoose(s, fMain, fMain)
	s.do(enactAction(enactFor(s, "Ensnaring Mist")))
	for n := 0; n < 2; n++ {
		got := csTargetOffers(s, n)
		csWantOffers(t, got, "a chosen-unit slot", aMain, aRes, aWard, aScratch, fMain, fRes)
		for _, x := range []*Asset{fWard, aWall, fWall, s.g.Players[0].HQ, s.g.Players[1].HQ} {
			if got[x.ID] {
				t.Errorf("slot %d: %s is not a legal chosen unit and must not be offered (FR-TGT-001, FR-KWD-002h)", n, x)
			}
		}
	}
}

func TestCard_EnsnaringMist_LegalityNeedsAUnit(t *testing.T) {
	// Covers: SVY-R29, FR-MNV-001
	t.Run("no unit on the battlefield", func(t *testing.T) {
		s := newScene(t)
		s.put(0, ZoneMainTheater, structure("AWall", 0, 5, Keywords{}))
		s.put(1, ZoneMainTheater, structure("FWall", 0, 5, Keywords{}))
		inst := enactFor(s, "Ensnaring Mist")
		if offersEnact(s.g, inst) {
			t.Errorf("with no unit to choose it is not offered (SVY-R29)")
		}
		s.reject(enactAction(inst))
		if !inHand(s.g.Players[0], inst) {
			t.Errorf("a rejected Enact leaves the card in hand (SVY-R29)")
		}
	})
	t.Run("exactly one unit", func(t *testing.T) {
		s := newScene(t)
		u := s.put(1, ZoneMainTheater, unit("Only", 3, 5, Keywords{}))
		inst := enactFor(s, "Ensnaring Mist")
		if !offersEnact(s.g, inst) {
			t.Fatalf("with exactly one unit it is offered (SVY-R29)")
		}
		emChoose(s, u, u)
		s.do(enactAction(inst))
		for _, d := range s.asked(0) {
			if d.Kind != DecideTarget {
				continue
			}
			for _, o := range d.Options {
				if o.Ref != u.ID {
					t.Errorf("only the single unit may be offered in each slot, got ref %d (FR-TGT-001)", o.Ref)
				}
			}
		}
		if u.Zone != ZoneReserves || s.g.Power(u) != 2 {
			t.Errorf("both slots take the only unit: it is moved to Reserves (zone %v) with power %d, want 2 (SVY-R35)", u.Zone, s.g.Power(u))
		}
	})
}

func TestCard_EnsnaringMist_TargetTimeoutsTakeFixedOrder(t *testing.T) {
	// Covers: FR-MOD-001, SVY-R35
	s := newScene(t)
	p := s.put(0, ZoneMainTheater, unit("P", 3, 5, Keywords{}))
	q := s.put(1, ZoneMainTheater, unit("Q", 3, 5, Keywords{}))
	emChoose(s)
	s.do(enactAction(enactFor(s, "Ensnaring Mist")))
	if p.Zone != ZoneReserves || !containsAsset(s.g.Players[0].Reserves, p) {
		t.Errorf("a timed-out target takes the first legal entity in fixed order, the chooser's own unit P, which is moved to its Reserves (FR-MOD-001 6)")
	}
	if got := s.g.Power(p); got != 2 {
		t.Errorf("both independent slots take P, so it gets -1 power: %d, want 2 (SVY-R35)", got)
	}
	if q.Zone != ZoneMainTheater || s.g.Power(q) != 3 {
		t.Errorf("Q is untouched (FR-MOD-001 6)")
	}
}

func TestCard_EnsnaringMist_PaysTwoTactics(t *testing.T) {
	// Covers: FR-MNV-001
	t.Run("one Tactics is too few", func(t *testing.T) {
		s := newScene(t)
		s.put(1, ZoneMainTheater, unit("Foe", 2, 5, Keywords{}))
		inst := enactFor(s, "Ensnaring Mist")
		s.g.Players[0].Tac = 1
		if offersEnact(s.g, inst) {
			t.Errorf("with 1 Tactics it is not offered (FR-MNV-001)")
		}
		s.reject(enactAction(inst))
	})
	t.Run("three Tactics", func(t *testing.T) {
		s := newScene(t)
		s.put(1, ZoneMainTheater, unit("Foe", 2, 5, Keywords{}))
		inst := enactFor(s, "Ensnaring Mist")
		s.g.Players[0].Tac = 3
		s.do(enactAction(inst))
		if p := s.g.Players[0]; p.Tac != 1 || p.Req != 3 {
			t.Errorf("enacting costs 2 Tactics and no Requisition: pool R%d T%d, want R3 T1 (FR-MNV-001)", p.Req, p.Tac)
		}
	})
}
