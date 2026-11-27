package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0321 (FR-TRN-003, FR-SET-004, FR-ACT-001, FR-ACT-002, FR-TGT-001,
// FR-TGT-003, FR-ZON-001, FR-ZON-004, FR-LAY-003, FR-MOD-001, FR-KWD-002h):
// Mobile Armory (Soulless unit, R3 T0, Main Theater or Reserves, 1/3), loaded
// from data/cards: "Exhaust: Chosen allied unit in the same zone gets +1 power
// until end of turn." Player 0 controls it and is the active player.

// maScene is a scene with a Readied Armory in player 0's Main Theater, pool R0 T0.
func maScene(t *testing.T) (*scene, *Asset) {
	t.Helper()
	s := newScene(t)
	m := s.put(0, ZoneMainTheater, s.card("Mobile Armory"))
	s.setPool(0, 0, 0)
	return s, m
}

func maWantPower(t *testing.T, s *scene, a *Asset, want int, why string) {
	t.Helper()
	if got := s.g.Power(a); got != want {
		t.Errorf("%s: power %d, want %d", why, got, want)
	}
}

func TestCard_MobileArmory_EntersEitherZoneAsOneThree(t *testing.T) {
	// Covers: FR-TRN-003
	for _, z := range []Zone{ZoneMainTheater, ZoneReserves} {
		s := newScene(t)
		c := s.card("Mobile Armory")
		inst := s.give(0, c)
		s.setPool(0, 3, 0)
		s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: z})
		a := assetOf(s.g, c)
		if a == nil || a.Zone != z || a.Controller != 0 {
			t.Fatalf("the Armory is on the battlefield in the chosen zone %v of player 0 (FR-TRN-003)", z)
		}
		if p := a.Printed(); p.Power != 1 || p.Health != 3 {
			t.Errorf("the Armory is a 1/3: got %d/%d", p.Power, p.Health)
		}
		if a.Exhausted {
			t.Errorf("a newly deployed unit enters Readied (FR-SET-004)")
		}
		if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
			t.Errorf("the Armory costs R3: pool R%d T%d, want R0 T0 (FR-TRN-003)", p.Req, p.Tac)
		}
	}

	s2 := newScene(t)
	c2 := s2.card("Mobile Armory")
	inst2 := s2.give(0, c2)
	s2.setPool(0, 2, 0)
	s2.reject(Action{Kind: ActDeploy, Player: 0, Card: inst2.ID, Zone: ZoneMainTheater})
	if !inHand(s2.g.Players[0], inst2) || assetOf(s2.g, c2) != nil {
		t.Errorf("with R2 the Armory (cost R3) cannot be deployed (FR-TRN-003)")
	}
}

func TestCard_MobileArmory_ChosenAllyGetsPlusOnePower(t *testing.T) {
	// Covers: FR-ACT-001, FR-LAY-001
	s, m := maScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 2, 5, ability.Keywords{}))
	other := s.put(0, ZoneMainTheater, unit("Other", 2, 5, ability.Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 2, 5, ability.Keywords{}))
	activateChoosing(s, m, ally)
	maWantPower(t, s, ally, 3, "the chosen ally gets +1 power (FR-ACT-001)")
	maWantPower(t, s, m, 1, "the Armory's own power is unchanged")
	maWantPower(t, s, other, 2, "an unchosen ally is unchanged")
	maWantPower(t, s, foe, 2, "an opposing unit is unchanged")
	if !m.Exhausted || !s.alive(m) {
		t.Errorf("the Armory is Exhausted and on the battlefield: exhausted %v alive %v (FR-ACT-001)", m.Exhausted, s.alive(m))
	}
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
		t.Errorf("the ability costs no resources: pool R%d T%d, want R0 T0", p.Req, p.Tac)
	}
}

func TestCard_MobileArmory_OffersAlliedUnitsInItsZoneOnly(t *testing.T) {
	// Covers: FR-TGT-001, SVY-R44, FR-KWD-002h
	s, m := maScene(t)
	a1 := s.put(0, ZoneMainTheater, unit("AllyMain", 1, 5, ability.Keywords{}))
	a2 := s.put(0, ZoneMainTheater, unit("AllyWard", 1, 5, ability.Keywords{Ward: true}))
	s.put(0, ZoneReserves, unit("AllyRes", 1, 5, ability.Keywords{}))
	s.put(0, ZoneMainTheater, structure("AllyWall", 0, 5, ability.Keywords{}))
	s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	s.put(1, ZoneMainTheater, unit("FoeWard", 1, 5, ability.Keywords{Ward: true}))
	s.put(1, ZoneReserves, unit("FoeRes", 1, 5, ability.Keywords{}))
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(m, 0))
	d := targetAsk(t, s)
	if d.Player != 0 || d.Min != 1 || d.Max != 1 {
		t.Errorf("the question is a mandatory single target decision for player 0: %+v", d)
	}
	wantOptions(t, d, "only the allied units in the Armory's zone (including the Warded one) and the Armory; no Headquarters, structure, Reserves unit or opposing unit",
		a1, a2, m)
}

func TestCard_MobileArmory_FromReservesOffersOwnReservesOnly(t *testing.T) {
	// Covers: SVY-R44, FR-ZON-001
	s := newScene(t)
	m := s.put(0, ZoneReserves, s.card("Mobile Armory"))
	s.setPool(0, 0, 0)
	res := s.put(0, ZoneReserves, unit("AllyRes", 1, 5, ability.Keywords{}))
	sc := s.putScratch(0)
	s.put(0, ZoneMainTheater, unit("AllyMain", 1, 5, ability.Keywords{}))
	s.put(1, ZoneReserves, unit("FoeRes", 1, 5, ability.Keywords{}))
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(m, 0))
	wantOptions(t, targetAsk(t, s), "player 0's own Reserves units: the unit, the Scratch Company and the Armory", res, sc, m)

	s2 := newScene(t)
	m2 := s2.put(0, ZoneReserves, s2.card("Mobile Armory"))
	sc2 := s2.putScratch(0)
	s2.put(0, ZoneReserves, unit("AllyRes", 1, 5, ability.Keywords{}))
	activateChoosing(s2, m2, sc2)
	maWantPower(t, s2, sc2, 2, "the Scratch Company is a unit and gets +1 power")
}

func TestCard_MobileArmory_MayChooseItself(t *testing.T) {
	// Covers: FR-TGT-001
	s, m := maScene(t)
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, ability.Keywords{}))
	activateChoosing(s, m, m)
	maWantPower(t, s, m, 2, "the Armory may choose itself (no \"other\")")
	if !m.Exhausted {
		t.Errorf("the Armory is Exhausted (FR-ACT-001)")
	}
}

func TestCard_MobileArmory_AloneInItsZoneBuffsItselfWithoutAsking(t *testing.T) {
	// Covers: SVY-R36, FR-MOD-001
	s, m := maScene(t)
	other := s.put(0, ZoneReserves, unit("AllyRes", 1, 5, ability.Keywords{}))
	s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	if !offersActivate(s.g, m, 0) {
		t.Fatalf("the Armory is its own legal target, so Activate is offered (SVY-R36)")
	}
	chooseAsset(s, m)
	s.do(activate(m, 0))
	for _, d := range s.asked(0) {
		if d.Kind == DecideTarget {
			t.Errorf("a single legal target is taken without asking (FR-MOD-001 Single-Option Fast-Path): %+v", d)
		}
	}
	maWantPower(t, s, m, 2, "the Armory buffs itself")
	maWantPower(t, s, other, 1, "the Reserves unit is unchanged")
}

func TestCard_MobileArmory_PowerExpiresAtEndOfTurn(t *testing.T) {
	// Covers: FR-LAY-003
	s, m := maScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 2, 5, ability.Keywords{}))
	activateChoosing(s, m, ally)
	maWantPower(t, s, ally, 3, "the buff applies")
	s.endTurn()
	maWantPower(t, s, ally, 2, "the buff ends at end of turn (FR-LAY-003)")
}

func TestCard_MobileArmory_BuffStaysWhenTheAllyMovesAfterwards(t *testing.T) {
	// Covers: SVY-R44, FR-ZON-004
	s, m := maScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 2, 5, ability.Keywords{}))
	activateChoosing(s, m, ally)
	s.moveTo(ally, ZoneReserves)
	maWantPower(t, s, ally, 3, "the zone is checked only on placement and resolution, so the buff stays")
}

func TestCard_MobileArmory_NotOfferedWhileExhausted(t *testing.T) {
	// Covers: FR-ACT-002
	s, m := maScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 2, 5, ability.Keywords{}))
	m.Exhausted = true
	if offersActivate(s.g, m, 0) {
		t.Errorf("an Exhausted Armory cannot pay its exhaust cost, so Activate is not offered (FR-ACT-002)")
	}
	s.reject(activate(m, 0))
	maWantPower(t, s, ally, 2, "a rejected Activate changes no power")
	maWantPower(t, s, m, 1, "a rejected Activate changes no power")
	if !m.Exhausted {
		t.Errorf("the Armory stays Exhausted (FR-ACT-002)")
	}
}

func TestCard_MobileArmory_NotOfferedWhileTheStackIsNotEmpty(t *testing.T) {
	// Covers: FR-ACT-001
	s, m := maScene(t)
	s.put(0, ZoneMainTheater, unit("Ally", 2, 5, ability.Keywords{}))
	s.g.Stack = append(s.g.Stack, &StackItem{Kind: StackScratch, Controller: 0})
	s.reject(activate(m, 0))
	if offersActivate(s.g, m, 0) {
		t.Errorf("no Activate is offered with an item on the stack (FR-ACT-001)")
	}
	if m.Exhausted {
		t.Errorf("a rejected Activate exhausts nothing (FR-ACT-001)")
	}
}

func TestCard_MobileArmory_MayActivateTheTurnItDeploys(t *testing.T) {
	// Covers: FR-SET-004
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 2, 5, ability.Keywords{}))
	c := s.card("Mobile Armory")
	inst := s.give(0, c)
	s.setPool(0, 3, 0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater})
	m := assetOf(s.g, c)
	if m == nil {
		t.Fatalf("the Armory is deployed")
	}
	if !offersActivate(s.g, m, 0) {
		t.Fatalf("a unit deployed this turn may activate right away (FR-SET-004)")
	}
	activateChoosing(s, m, ally)
	maWantPower(t, s, ally, 3, "the ally gets +1 power")
}

func TestCard_MobileArmory_TimeoutTakesTheFirstOption(t *testing.T) {
	// Covers: FR-MOD-001, SVY-R16
	s, m := maScene(t)
	a1 := s.put(0, ZoneMainTheater, unit("AllyOne", 2, 5, ability.Keywords{}))
	a2 := s.put(0, ZoneMainTheater, unit("AllyTwo", 2, 5, ability.Keywords{}))
	s.script(0, Answer{Timeout: true})
	s.do(activate(m, 0))
	d := targetAsk(t, s)
	if len(d.Options) == 0 {
		t.Fatalf("the target decision has options")
	}
	first := d.Options[0].Ref
	for _, a := range []*Asset{m, a1, a2} {
		want := a.Printed().Power
		if a.ID == first {
			want++
		}
		maWantPower(t, s, a, want, "a timeout buffs only the first option in fixed order (FR-MOD-001 sec. 6, SVY-R16)")
	}
}

// maHookScene is a scene whose Armory carries a synthetic exhaust trigger with
// step; "Ally" is the chosen unit, in the Main Theater.
func maHookScene(t *testing.T, step ability.Step) (*scene, *Asset, *Asset) {
	t.Helper()
	s := newScene(t)
	m := s.put(0, ZoneMainTheater, s.cardWithTrigger("Mobile Armory", exhaustTrigger(step)))
	s.setPool(0, 0, 0)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 2, 5, ability.Keywords{}))
	return s, m, ally
}

var otherAlliedMainUnits = ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Other: true, Zone: ability.ZoneMainTheater}

func TestCard_MobileArmory_TargetLeftTheZoneGetsNothing(t *testing.T) {
	// Covers: FR-TGT-003, SVY-R44
	s, m, ally := maHookScene(t, ability.Step{Op: ability.OpMoveZone, Subject: ability.SubjectAll, Filter: otherAlliedMainUnits, To: ability.ZoneReserves})
	activateChoosing(s, m, ally)
	if ally.Zone != ZoneReserves {
		t.Fatalf("setup: the synthetic trigger should have moved the target to Reserves before the ability resolves")
	}
	maWantPower(t, s, ally, 2, "a target that left the locked zone is illegal at the second check (FR-TGT-003)")
	if !m.Exhausted {
		t.Errorf("the cost stays paid: the Armory stays Exhausted (FR-TGT-003)")
	}
}

func TestCard_MobileArmory_SourceMovedStillBuffsInTheLockedZone(t *testing.T) {
	// Covers: SVY-R44, FR-TGT-003
	s, m, ally := maHookScene(t, ability.Step{Op: ability.OpMoveZone, Subject: ability.SubjectSelf, To: ability.ZoneReserves})
	activateChoosing(s, m, ally)
	if m.Zone != ZoneReserves {
		t.Fatalf("setup: the Armory should have moved to Reserves before the ability resolves")
	}
	maWantPower(t, s, ally, 3, "the zone was locked on placement, so the ally still in the Main Theater gets +1 power (SVY-R44)")
}

func TestCard_MobileArmory_SourceGoneStillBuffs(t *testing.T) {
	// Covers: FR-ZON-004, SVY-R44
	s, m, ally := maHookScene(t, ability.Step{Op: ability.OpKill, Subject: ability.SubjectSelf})
	activateChoosing(s, m, ally)
	if s.alive(m) {
		t.Fatalf("setup: the Armory should be dead before the ability resolves")
	}
	if !inDiscardPile(s.g.Players[0], m.Inst) {
		t.Errorf("the Armory's card is in its owner's discard pile (FR-ZON-004)")
	}
	maWantPower(t, s, ally, 3, "the ability resolves from last-known information (SVY-R44)")
}

func TestCard_MobileArmory_TargetGoneGetsNothing(t *testing.T) {
	// Covers: FR-TGT-003, SVY-R25
	s, m, ally := maHookScene(t, ability.Step{Op: ability.OpReturnToHand, Subject: ability.SubjectAll, Filter: otherAlliedMainUnits})
	activateChoosing(s, m, ally)
	if s.alive(ally) || !inHand(s.g.Players[0], ally.Inst) {
		t.Fatalf("setup: the target should be back in player 0's hand before the ability resolves")
	}
	maWantPower(t, s, m, 1, "no asset's power changes when the target is gone (FR-TGT-003, SVY-R25)")
	if !m.Exhausted {
		t.Errorf("the cost stays paid (FR-TGT-003)")
	}
}
