package engine

import (
	"reflect"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0306: Tempest City Ray (structure, R5 T0, 0/8, Main Theater):
// "T2, exhaust: Move chosen opposing unit to its controller's Reserves."
// The real card is loaded from data/cards. Player 0 controls the Ray and is the
// active player unless a case says otherwise.

var trayOpposingUnits = ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}

// trayRay is a scene with a Readied Ray put in player 0's Main Theater, pool R0 T2.
func trayRay(t *testing.T) (*scene, *Asset) {
	t.Helper()
	s := newScene(t)
	r := s.put(0, ZoneMainTheater, s.card("Tempest City Ray"))
	s.setPool(0, 0, 2)
	return s, r
}

func trayInReserves(p *Player, a *Asset) bool { return inAssets(p.Reserves, a) }

// trayWatcher is a unit printing a deploy trigger (+4 Armor) and a Moves
// trigger (+1 Armor) on itself, so which of them fires shows in its Armor.
func trayWatcher() *Card {
	c := unit("Watcher", 2, 6, ability.Keywords{})
	armor := func(n int) ability.Effect {
		return steps(ability.Step{Op: ability.OpAddArmor, Amount: ability.Value{N: n}, Subject: ability.SubjectSelf})
	}
	c.Ability = &ability.Def{Triggers: []ability.Trigger{
		{Hook: ability.HookDeploy, Effect: armor(4)},
		{Hook: ability.HookMoves, Effect: armor(1)},
	}}
	return c
}

// --- Deploying ---

func TestCard_TempestCityRay_EntersMainTheaterAsZeroEight(t *testing.T) {
	// Covers: FR-TRN-003
	s := newScene(t)
	c := s.card("Tempest City Ray")
	inst := s.give(0, c)
	s.setPool(0, 5, 0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater})
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneMainTheater {
		t.Fatalf("the Ray is on the battlefield in the Main Theater after deploying (FR-TRN-003)")
	}
	if p := a.Printed(); p.Power != 0 || p.Health != 8 || p.Type != TypeStructure {
		t.Errorf("the Ray is a 0/8 structure: got %d/%d type %v", p.Power, p.Health, p.Type)
	}
	if a.Exhausted {
		t.Errorf("a newly deployed structure enters Readied (FR-SET-004)")
	}
	if s.g.Players[0].Req != 0 {
		t.Errorf("the Ray costs 5 Requisition: %d left", s.g.Players[0].Req)
	}

	s2 := newScene(t)
	c2 := s2.card("Tempest City Ray")
	inst2 := s2.give(0, c2)
	s2.setPool(0, 5, 0)
	s2.reject(Action{Kind: ActDeploy, Player: 0, Card: inst2.ID, Zone: ZoneReserves})
	if !inHand(s2.g.Players[0], inst2) || assetOf(s2.g, c2) != nil {
		t.Errorf("deploying the Ray to Reserves is rejected and the card stays in hand (FR-TRN-003)")
	}

	s3 := newScene(t)
	c3 := s3.card("Tempest City Ray")
	inst3 := s3.give(0, c3)
	s3.setPool(0, 4, 0)
	s3.reject(Action{Kind: ActDeploy, Player: 0, Card: inst3.ID, Zone: ZoneMainTheater})
	if !inHand(s3.g.Players[0], inst3) || assetOf(s3.g, c3) != nil {
		t.Errorf("with 4 Requisition the Ray (cost 5) cannot be deployed (FR-TRN-003)")
	}
}

// --- Activating ---

func TestCard_TempestCityRay_MovesChosenOpposingUnitToReserves(t *testing.T) {
	// Covers: FR-ZON-002, FR-ZON-004, SVY-R1
	s, r := trayRay(t)
	foe := s.put(1, ZoneMainTheater, trayWatcher())
	foe.Damage, foe.Armor, foe.Exhausted = 1, 1, true
	foe.Modifiers = []Modifier{{Layer: 5, Power: 1}}
	wantMods := append([]Modifier(nil), foe.Modifiers...)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, ability.Keywords{}))
	foeStruct := s.put(1, ZoneMainTheater, structure("Wall", 0, 5, ability.Keywords{}))
	s.script(0, Answer{Picks: []int{0}})
	from := len(s.g.Log)
	s.do(activate(r, 0))
	if s.g.Asset(foe.ID) != foe {
		t.Fatalf("the unit is the same asset after the move (SVY-R1: a move is not a deploy)")
	}
	if foe.Zone != ZoneReserves || !trayInReserves(s.g.Players[1], foe) || inAssets(s.g.Theater, foe) {
		t.Errorf("the unit is in player 1's Reserves and not in the Main Theater (FR-ZON-002)")
	}
	if !foe.Exhausted || foe.Damage != 1 || !reflect.DeepEqual(foe.Modifiers, wantMods) {
		t.Errorf("a move keeps state: exhausted %v damage %d modifiers %v; want Exhausted, 1 damage, the +1 power modifier (SVY-R1)",
			foe.Exhausted, foe.Damage, foe.Modifiers)
	}
	if foe.Armor != 2 {
		t.Errorf("Armor %d, want 2: its 1 Armor plus 1 from its Moves trigger, and nothing from its deploy trigger (a move is not a deploy, SVY-R1)", foe.Armor)
	}
	evs := effectEvents(s.g, from, EventMove)
	if len(evs) != 1 || evs[0].Target != foe.ID {
		t.Errorf("want one move event naming the unit, got %+v", evs)
	}
	p := s.g.Players[0]
	if p.Req != 0 || p.Tac != 0 {
		t.Errorf("the 2 Tactics are spent: pool R%d T%d, want R0 T0 (FR-ACT-001)", p.Req, p.Tac)
	}
	if !s.alive(r) || !r.Exhausted || r.Zone != ZoneMainTheater {
		t.Errorf("the Ray stays on the battlefield, Exhausted: alive %v exhausted %v", s.alive(r), r.Exhausted)
	}
	if ally.Zone != ZoneMainTheater || foeStruct.Zone != ZoneMainTheater {
		t.Errorf("no other asset moves: ally %v, structure %v", ally.Zone, foeStruct.Zone)
	}

	t.Run("face-down Scratch Company", func(t *testing.T) {
		s, r := trayRay(t)
		sc := s.put(1, ZoneMainTheater, unit("Filler", 1, 1, ability.Keywords{}))
		sc.FaceDown = true
		s.script(0, Answer{Picks: []int{0}})
		s.do(activate(r, 0))
		if !s.alive(sc) || sc.Zone != ZoneReserves || !trayInReserves(s.g.Players[1], sc) || inAssets(s.g.Theater, sc) {
			t.Errorf("a face-down Scratch Company moves to player 1's Reserves (FR-ZON-002)")
		}
		if !sc.FaceDown {
			t.Errorf("it stays face down (SVY-R1)")
		}
	})
}

func TestCard_TempestCityRay_OnlyOpposingUnitsOffered(t *testing.T) {
	// Covers: FR-TGT-001, SVY-R57, FR-ZON-002
	s, r := trayRay(t)
	b := s.targetBoard()
	foeRes := s.put(1, ZoneReserves, unit("FoeRes", 1, 5, ability.Keywords{}))
	foeFD := s.put(1, ZoneMainTheater, unit("Filler", 1, 1, ability.Keywords{}))
	foeFD.FaceDown = true
	s.put(0, ZoneReserves, unit("AllyRes", 1, 5, ability.Keywords{}))
	allyFD := s.put(0, ZoneMainTheater, unit("Filler", 1, 1, ability.Keywords{}))
	allyFD.FaceDown = true
	s.putControlled(1, 0, unit("Stolen", 1, 5, ability.Keywords{}))
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(r, 0))
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one chosen target is one question: asked %d times (FR-TGT-001)", len(asked))
	}
	d := asked[0]
	if d.Kind != DecideTarget || d.Player != 0 || d.Min != 1 || d.Max != 1 {
		t.Errorf("the question is a mandatory single DecideTarget for player 0: %+v", d)
	}
	wantOptions(t, d, "exactly player 1's units, in the Main Theater and in Reserves, face-down Scratch Companies included (SVY-R57)",
		b.u1, b.c1, foeRes, foeFD)
}

func TestCard_TempestCityRay_UnitAlreadyInReservesStaysPut(t *testing.T) {
	// Covers: SVY-R57, FR-ZON-002
	s, r := trayRay(t)
	foe := s.put(1, ZoneReserves, unit("FoeRes", 2, 5, ability.Keywords{}))
	foe.Damage, foe.Exhausted = 2, true
	other := s.put(1, ZoneMainTheater, unit("FoeMain", 2, 5, ability.Keywords{}))
	// Two legal options, so the question is asked (FR-MOD-001 sec. 1): pick the
	// Reserves unit by its ref, not by position.
	s.script(0, Answer{Picks: []int{0}})
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		for i, o := range d.Options {
			if o.Ref == foe.ID {
				return Answer{Picks: []int{i}}
			}
		}
		t.Errorf("the Reserves unit is not among the options %v (SVY-R57)", refs(d))
		return Answer{Timeout: true}
	}
	from := len(s.g.Log)
	s.do(activate(r, 0))
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("an opposing unit in Reserves is a legal choice: asked %d times (SVY-R57)", len(asked))
	}
	if other.Zone != ZoneMainTheater || !inAssets(s.g.Theater, other) {
		t.Errorf("the other opposing unit was not chosen and stays in the Main Theater")
	}
	if s.g.Asset(foe.ID) != foe || foe.Zone != ZoneReserves || !trayInReserves(s.g.Players[1], foe) {
		t.Errorf("the unit is still the same asset in player 1's Reserves (SVY-R57)")
	}
	if foe.Damage != 2 || !foe.Exhausted {
		t.Errorf("its state is kept: damage %d exhausted %v", foe.Damage, foe.Exhausted)
	}
	if n := len(effectEvents(s.g, from, EventMove)); n != 0 {
		t.Errorf("a unit already in Reserves moves nothing, so no move event is logged: %d (FR-ZON-002)", n)
	}
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 || !r.Exhausted {
		t.Errorf("the costs are still paid: pool R%d T%d, Ray exhausted %v (FR-ACT-001)", p.Req, p.Tac, r.Exhausted)
	}
}

func TestCard_TempestCityRay_ControlledUnitGoesToItsControllersReserves(t *testing.T) {
	// Covers: SVY-R65, FR-ZON-003, SVY-R7
	s := newScene(t)
	ray := s.put(0, ZoneMainTheater, s.card("Tempest City Ray"))
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
	ray.Exhausted = false
	s.setPool(0, 0, 2)
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(ray, 0))
	if !s.alive(mine) || mine.Zone != ZoneReserves {
		t.Fatalf("the unit leaves the Main Theater for Reserves: alive %v zone %v", s.alive(mine), mine.Zone)
	}
	if !trayInReserves(s.g.Players[1], mine) || trayInReserves(s.g.Players[0], mine) {
		t.Errorf("it goes to its controller's (player 1's) Reserves, not its owner's (SVY-R65)")
	}
	if mine.Inst.Owner != 0 || mine.Controller != 1 {
		t.Errorf("owner %d controller %d, want owner 0 and controller 1 (FR-ZON-003, SVY-R7)", mine.Inst.Owner, mine.Controller)
	}
}

func TestCard_TempestCityRay_WardedOpposingUnitNotOffered(t *testing.T) {
	// Covers: FR-KWD-002h, SVY-R36
	s, r := trayRay(t)
	ward := s.put(1, ZoneMainTheater, unit("FoeWard", 1, 5, ability.Keywords{Ward: true}))
	plain := s.put(1, ZoneMainTheater, unit("FoePlain", 1, 5, ability.Keywords{}))
	// The plain unit is the only legal option, so the Single-Option Fast-Path
	// takes it and asks nothing (FR-MOD-001 sec. 1); assert the outcome.
	s.script(0)
	s.do(activate(r, 0))
	if plain.Zone != ZoneReserves || ward.Zone != ZoneMainTheater {
		t.Errorf("only the unit without Ward moves: plain %v, ward %v", plain.Zone, ward.Zone)
	}

	s2, r2 := trayRay(t)
	s2.put(1, ZoneMainTheater, unit("FoeWard", 1, 5, ability.Keywords{Ward: true}))
	actNotOffered(t, s2, r2, 0, "the only opposing unit has Ward")
}

func TestCard_TempestCityRay_NotOfferedWithoutAnOpposingUnit(t *testing.T) {
	// Covers: SVY-R36, FR-ACT-001
	s, r := trayRay(t)
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, ability.Keywords{}))
	s.put(0, ZoneReserves, unit("AllyRes", 1, 5, ability.Keywords{}))
	s.put(1, ZoneMainTheater, structure("FoeWall", 0, 5, ability.Keywords{}))
	actNotOffered(t, s, r, 0, "only allied units, an opposing structure and both Headquarters")
}

func TestCard_TempestCityRay_NotOfferedWithoutTwoTactics(t *testing.T) {
	// Covers: FR-ACT-001
	s, r := trayRay(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	s.setPool(0, 5, 1)
	if offersActivate(s.g, r, 0) {
		t.Errorf("with R5 T1 the Ray's T2 cost cannot be paid, so Activate is not offered (FR-ACT-001)")
	}
	s.reject(activate(r, 0))
	p := s.g.Players[0]
	if foe.Zone != ZoneMainTheater || r.Exhausted || p.Req != 5 || p.Tac != 1 {
		t.Errorf("a rejected Activate moves nothing: foe %v exhausted %v pool R%d T%d (FR-ACT-001)", foe.Zone, r.Exhausted, p.Req, p.Tac)
	}
}

func TestCard_TempestCityRay_NotOfferedWhileExhausted(t *testing.T) {
	// Covers: FR-ACT-002
	s, r := trayRay(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	s.setPool(0, 0, 4)
	r.Exhausted = true
	if offersActivate(s.g, r, 0) {
		t.Errorf("an Exhausted Ray cannot pay its exhaust cost, so Activate is not offered (FR-ACT-002)")
	}
	s.reject(activate(r, 0))
	if p := s.g.Players[0]; foe.Zone != ZoneMainTheater || p.Tac != 4 {
		t.Errorf("a rejected Activate moves nothing: foe %v, Tactics %d (FR-ACT-002)", foe.Zone, p.Tac)
	}
}

func TestCard_TempestCityRay_MayActivateTheTurnItDeploys(t *testing.T) {
	// Covers: FR-SET-004
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	c := s.card("Tempest City Ray")
	inst := s.give(0, c)
	s.setPool(0, 5, 2)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater})
	ray := assetOf(s.g, c)
	if ray == nil {
		t.Fatalf("the Ray is deployed")
	}
	if !offersActivate(s.g, ray, 0) {
		t.Errorf("a Ray deployed this turn may activate right away: there is no summoning sickness (FR-SET-004)")
	}
}

func TestCard_TempestCityRay_TargetGoneMovesNothing(t *testing.T) {
	// Covers: FR-TGT-003
	s := newScene(t)
	r := s.put(0, ZoneMainTheater, s.cardWithTrigger("Tempest City Ray", exhaustTrigger(ability.Step{Op: ability.OpReturnToHand, Subject: ability.SubjectAll, Filter: trayOpposingUnits})))
	s.setPool(0, 0, 2)
	foe := s.put(1, ZoneMainTheater, unit("Fleeting", 1, 5, ability.Keywords{}))
	s.script(0, Answer{Picks: []int{0}})
	from := len(s.g.Log)
	s.do(activate(r, 0))
	if s.alive(foe) {
		t.Fatalf("setup: the unit should have left the battlefield before resolution")
	}
	if n := len(effectEvents(s.g, from, EventMove)); n != 0 {
		t.Errorf("a target that left before resolution is skipped and no move event is logged: %d (FR-TGT-003)", n)
	}
	if !s.alive(r) || !r.Exhausted {
		t.Errorf("the Ray stays on the battlefield, Exhausted: alive %v exhausted %v", s.alive(r), r.Exhausted)
	}
}

func TestCard_TempestCityRay_TargetNowAlliedMovesNothing(t *testing.T) {
	// Covers: FR-TGT-003, FR-ZON-002
	s := newScene(t)
	r := s.put(0, ZoneMainTheater, s.cardWithTrigger("Tempest City Ray", exhaustTrigger(ability.Step{Op: ability.OpGainControl, Subject: ability.SubjectAll, Filter: trayOpposingUnits})))
	s.setPool(0, 0, 2)
	foe := s.put(1, ZoneMainTheater, unit("Turncoat", 1, 5, ability.Keywords{}))
	s.script(0, Answer{Picks: []int{0}})
	from := len(s.g.Log)
	s.do(activate(r, 0))
	if foe.Controller != 0 {
		t.Fatalf("setup: player 0 gained control of the unit before resolution: controller %d", foe.Controller)
	}
	if !s.alive(foe) || foe.Zone != ZoneMainTheater || !inAssets(s.g.Theater, foe) {
		t.Errorf("the unit no longer matches \"opposing\", so it is illegal at the second check and stays in the Main Theater (FR-TGT-003, FR-ZON-002)")
	}
	if n := len(effectEvents(s.g, from, EventMove)); n != 0 {
		t.Errorf("no move event is logged: %d (FR-TGT-003)", n)
	}
}

func TestCard_TempestCityRay_SourceGoneStillMoves(t *testing.T) {
	// Covers: FR-TGT-003, FR-ZON-004
	s := newScene(t)
	r := s.put(0, ZoneMainTheater, s.cardWithTrigger("Tempest City Ray", exhaustTrigger(ability.Step{Op: ability.OpKill, Subject: ability.SubjectSelf})))
	s.setPool(0, 0, 2)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(r, 0))
	if s.alive(r) {
		t.Fatalf("setup: the Ray should be dead before resolution")
	}
	if !s.alive(foe) || foe.Zone != ZoneReserves || !trayInReserves(s.g.Players[1], foe) {
		t.Errorf("the chosen unit still moves to player 1's Reserves when the source is gone (FR-TGT-003, FR-ZON-004)")
	}
}
