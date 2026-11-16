package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for the "in the same zone" filter criterion (T-0246; SVY-R44, FRD 01
// FR-TGT-001 "In the Same Zone" Is Locked on Placement, FR-TGT-003, FR-ZON-004,
// FR-ACT-001, SVY-R36). A stack item whose effect reads
// ability.Filter{ZoneRef: ability.ZoneRefSame} locks the zone its source was
// in when the item was placed (the source's last-known zone if it had already
// left the battlefield), and reads that zone at both target checks and in
// untargeted steps.
//
// This file does not compile until ability has ZoneRefSame; that is expected
// until T-0246 is implemented.

var (
	sameUnits   = ability.Filter{Kinds: ability.KindUnit, ZoneRef: ability.ZoneRefSame}
	sameAllies  = ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, ZoneRef: ability.ZoneRefSame}
	sameAlliesO = ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, ZoneRef: ability.ZoneRefSame, Other: true}
)

// szDeploy deploys c for player p into zone z and returns its asset.
func szDeploy(s *scene, p PlayerID, z Zone, c *Card) *Asset {
	s.t.Helper()
	inst := &Instance{ID: s.g.newID(), Card: c, Owner: p}
	s.g.deployCard(inst, p, z)
	return assetOf(s.g, c)
}

// szTargetDeployer is a unit whose deploy trigger is "chosen unit in the same
// zone gains 1 armor".
func szTargetDeployer() *Card {
	return striker("Deployer", 1, 5, Keywords{}, ability.Trigger{
		Hook: ability.HookDeploy,
		Effect: ability.Effect{
			Targets: []ability.Target{{Filter: sameUnits}},
			Steps:   []ability.Step{armorStep(1, 0)},
		},
	})
}

func TestFR_TGT_001_SameZoneAllStepReadsSourceZone(t *testing.T) {
	// Covers: SVY-R44, FR-ZON-001, SVY-R3
	f := tzScene(t, unit("Idle", 1, 5, Keywords{}), ZoneReserves)
	tzScript(f.s)
	card := structure("Beacon", 1, 5, Keywords{})
	card.Ability = &ability.Def{Triggers: []ability.Trigger{deployTrigger(allStep(ability.OpAddArmor, 1, sameUnits))}}
	szDeploy(f.s, 0, ZoneMainTheater, card)
	f.s.g.flushTriggers()
	f.s.g.resolveStack()
	why := "\"each unit in the same zone\" acts on every unit, either player's, in the zone the source was in on placement (the Main Theater) and nothing else (SVY-R44)"
	tzWantArmor(t, f.main0, 1, why)
	tzWantArmor(t, f.main1, 1, why)
	tzWantArmor(t, f.res0, 0, why)
	tzWantArmor(t, f.res1, 0, why)
}

func TestFR_TGT_001_SameZoneLocksOwnReserves(t *testing.T) {
	// Covers: SVY-R44, FR-ZON-001
	f := tzScene(t, unit("Idle", 1, 5, Keywords{}), ZoneReserves)
	tzScript(f.s)
	card := striker("Deployer", 1, 5, Keywords{}, deployTrigger(allStep(ability.OpAddArmor, 1, sameUnits)))
	card.DeployZones = []Zone{ZoneReserves, ZoneMainTheater}
	szDeploy(f.s, 0, ZoneReserves, card)
	items := flushedItems(f.s)
	if len(items) != 1 {
		t.Fatalf("expected the deploy trigger placed once, got %d items", len(items))
	}
	wantLocked(t, items[0], ZoneReserves, 0, "a source in player 0's Reserves locks that player's own Reserves (SVY-R44, FR-ZON-001)")
	f.s.g.resolveStack()
	why := "\"in the same zone\" from player 0's Reserves reaches player 0's Reserves units only (FR-ZON-001, SVY-R44)"
	tzWantArmor(t, f.res0, 1, why)
	tzWantArmor(t, f.res1, 0, why)
	tzWantArmor(t, f.main0, 0, why)
	tzWantArmor(t, f.main1, 0, why)
}

func TestFR_TGT_001_SameZoneLockedOnPlacement(t *testing.T) {
	// Covers: SVY-R44
	f := tzScene(t, unit("Idle", 1, 5, Keywords{}), ZoneReserves)
	tzScript(f.s)
	card := striker("Mover", 1, 5, Keywords{}, deployTrigger(
		ability.Step{Op: ability.OpMoveZone, Subject: ability.SubjectSelf, To: ability.ZoneReserves},
		allStep(ability.OpAddArmor, 1, sameUnits)))
	src := szDeploy(f.s, 0, ZoneMainTheater, card)
	f.s.g.flushTriggers()
	f.s.g.resolveStack()
	if src == nil || src.Zone != ZoneReserves {
		t.Fatalf("precondition: the source moved itself to Reserves before the armor step")
	}
	why := "the zone is the one the source was in when the item was placed (the Main Theater), not where it is when the step resolves (SVY-R44)"
	tzWantArmor(t, f.main0, 1, why)
	tzWantArmor(t, f.main1, 1, why)
	tzWantArmor(t, f.res0, 0, why)
	tzWantArmor(t, f.res1, 0, why)
}

func TestFR_ZON_004_SameZoneLastKnownZone(t *testing.T) {
	// Covers: SVY-R44, FR-TGT-001
	f := tzScene(t, unit("Idle", 1, 5, Keywords{}), ZoneReserves)
	victim := f.s.put(0, ZoneReserves, striker("Martyr", 1, 5, Keywords{}, killedTrigger(allStep(ability.OpAddArmor, 1, sameAllies))))
	killer := striker("Killer", 1, 5, Keywords{}, ability.Trigger{
		Hook: ability.HookDeploy,
		Effect: ability.Effect{
			Targets: []ability.Target{{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneReserves, Other: true}}},
			Steps:   []ability.Step{{Op: ability.OpKill, Subject: ability.SubjectTarget, Target: 0}},
		},
	})
	tzScript(f.s, victim.ID)
	szDeploy(f.s, 0, ZoneMainTheater, killer)
	for i := 0; i < 3; i++ {
		f.s.g.flushTriggers()
		f.s.g.resolveStack()
	}
	if f.s.alive(victim) {
		t.Fatalf("precondition: the Martyr in Reserves was killed")
	}
	why := "a death trigger is placed after its source left the battlefield and locks the source's last-known zone, player 0's Reserves (SVY-R44, FR-ZON-004)"
	tzWantArmor(t, f.res0, 1, why)
	tzWantArmor(t, f.main0, 0, why)
	tzWantArmor(t, f.res1, 0, why)
	tzWantArmor(t, f.main1, 0, why)
}

func TestFR_TGT_001_SameZoneTargetOffersSourceZoneOnly(t *testing.T) {
	// Covers: SVY-R44, FR-ZON-001
	f := tzScene(t, unit("Idle", 1, 5, Keywords{}), ZoneReserves)
	tzScript(f.s, f.main0.ID)
	src := szDeploy(f.s, 0, ZoneMainTheater, szTargetDeployer())
	f.s.g.flushTriggers()
	tzWantOffered(t, f.s, DecideTarget, "a targeted \"chosen unit in the same zone\" offers only the units in the source's zone, both players', and none in Reserves (SVY-R44, FR-TGT-001)",
		src, f.main0, f.main1)
}

func TestFR_TGT_003_SameZoneTargetOutlivesSource(t *testing.T) {
	// Covers: SVY-R44, FR-ZON-004
	f := tzScene(t, unit("Idle", 1, 5, Keywords{}), ZoneReserves)
	tzScript(f.s, f.main1.ID)
	src := szDeploy(f.s, 0, ZoneMainTheater, szTargetDeployer())
	f.s.g.flushTriggers()
	f.s.g.leaveBattlefield(src, f.s.g.effectiveState(src))
	f.s.g.resolveStack()
	tzWantArmor(t, f.main1, 1, "the second check reads the zone locked on placement, so a target in it stays legal after the source left (FR-TGT-003, SVY-R44)")
}

func TestFR_TGT_003_SameZoneTargetLeftZoneIsIllegal(t *testing.T) {
	// Covers: SVY-R44
	f := tzScene(t, unit("Idle", 1, 5, Keywords{}), ZoneReserves)
	tzScript(f.s, f.main1.ID)
	szDeploy(f.s, 0, ZoneMainTheater, szTargetDeployer())
	f.s.g.flushTriggers()
	f.s.moveTo(f.main1, ZoneReserves)
	f.s.g.resolveStack()
	tzWantArmor(t, f.main1, 0, "a target that left the locked zone is illegal at the second check and takes no effect (FR-TGT-003, SVY-R44)")
}

func TestFR_ACT_001_SameZoneActivatedAbility(t *testing.T) {
	// Covers: SVY-R44, SVY-R36, FR-TGT-001
	armorAlly := func(f ability.Filter) *Card {
		c := unit("Tender", 1, 5, Keywords{})
		c.Ability = &ability.Def{Activated: []ability.Activated{{
			Costs: []ability.Cost{actExhaust},
			Effect: ability.Effect{
				Targets: []ability.Target{{Filter: f}},
				Steps:   []ability.Step{armorStep(1, 0)},
			},
		}}}
		return c
	}
	t.Run("offers the source's zone only", func(t *testing.T) {
		f := tzScene(t, armorAlly(sameAllies), ZoneReserves)
		tzScript(f.s, f.res0.ID)
		if !offersActivate(f.s.g, f.w, 0) {
			t.Fatalf("the ability has a legal target in its zone and is offered (FR-ACT-001, SVY-R36)")
		}
		f.s.do(activate(f.w, 0))
		tzWantOffered(t, f.s, DecideTarget, "activating from Reserves offers player 0's Reserves allies, the source included, and no Main Theater ally (SVY-R44, FR-ZON-001)",
			f.w, f.res0)
	})
	t.Run("the source itself qualifies", func(t *testing.T) {
		s := newScene(t)
		src := s.put(0, ZoneMainTheater, armorAlly(sameAllies))
		res := s.put(0, ZoneReserves, unit("Res0", 1, 5, Keywords{}))
		s.put(1, ZoneMainTheater, unit("Main1", 1, 5, Keywords{}))
		tzScript(s, src.ID)
		if !offersActivate(s.g, src, 0) {
			t.Fatalf("without \"other\" the source is its own legal target in its zone, so the ability is offered (SVY-R36)")
		}
		s.do(activate(src, 0))
		s.g.resolveStack()
		tzWantArmor(t, src, 1, "the source in its own zone is a legal \"allied unit in the same zone\" (SVY-R44)")
		tzWantArmor(t, res, 0, "the Reserves ally is not in the source's zone (SVY-R44)")
	})
	t.Run("other with no ally in the zone", func(t *testing.T) {
		s := newScene(t)
		src := s.put(0, ZoneMainTheater, armorAlly(sameAlliesO))
		s.put(0, ZoneReserves, unit("Res0", 1, 5, Keywords{}))
		s.put(1, ZoneMainTheater, unit("Main1", 1, 5, Keywords{}))
		if offersActivate(s.g, src, 0) {
			t.Errorf("no other allied unit shares the source's zone, so a required target has no legal entity and the ability can't be activated (SVY-R36, SVY-R44)")
		}
	})
}

func TestFR_TGT_001_SameZoneWithoutSourceMatchesNothing(t *testing.T) {
	// Covers: SVY-R44
	s := newScene(t)
	main0 := s.put(0, ZoneMainTheater, unit("Main0", 1, 5, Keywords{}))
	res0 := s.put(0, ZoneReserves, unit("Res0", 1, 5, Keywords{}))
	main1 := s.put(1, ZoneMainTheater, unit("Main1", 1, 5, Keywords{}))
	pile := unit("Remnant", 1, 1, Keywords{})
	pile.Ability = &ability.Def{Triggers: []ability.Trigger{pileTrigger(false, drawN(1), allStep(ability.OpAddArmor, 1, sameUnits))}}
	s.enactWith([]*Card{pile}, discardN(1))
	s.g.resolveStack()
	if got := len(s.g.Players[0].Hand); got != 1 {
		t.Fatalf("precondition: the pile trigger resolved and its first step drew 1 card, so player 0's hand is 1, got %d", got)
	}
	why := "a trigger from a card in a discard pile has no battlefield source, locks nothing, and \"in the same zone\" matches nothing (SVY-R44)"
	tzWantArmor(t, main0, 0, why)
	tzWantArmor(t, res0, 0, why)
	tzWantArmor(t, main1, 0, why)
}
