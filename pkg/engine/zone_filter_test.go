package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for the Zone criterion of a Filter (T-0052, FR-TGT-001 "in the Main
// Theater", FR-TGT-003 second-stage verification). Cards are synthetic: a unit
// deploys from hand into Reserves and its deploy trigger targets or hits assets
// "in the Main Theater". The tests do not compile until ability.Filter has a
// Zone field (lead decision 1 of T-0052).
//
// scene.moveTo(a, z) is a requested helper (see the task's Helpers Needed).

// tgtZone is one "chosen" slot with the given conjunction filter and zone.
func tgtZone(a ability.Allegiance, k ability.Kinds, nonScratch bool, z ability.Zone) ability.Target {
	return ability.Target{Filter: ability.Filter{Allegiance: a, Kinds: k, NonScratch: nonScratch, Zone: z}}
}

// zoneBoard is a mixed board: some assets in the Main Theater (mt*), some in
// Reserves (rs*), for both players.
type zoneBoard struct {
	mtUnit0, mtStruct0, mtUnit1, mtStruct1 *Asset
	rsUnit0, rsUnit1, rsScratch1           *Asset
}

func (s *scene) zoneBoard() zoneBoard {
	return zoneBoard{
		mtUnit0:    s.put(0, ZoneMainTheater, unit("Ally MT Unit", 1, 5, Keywords{})),
		mtStruct0:  s.put(0, ZoneMainTheater, structure("Ally MT Structure", 0, 5, Keywords{})),
		mtUnit1:    s.put(1, ZoneMainTheater, unit("Enemy MT Unit", 1, 5, Keywords{})),
		mtStruct1:  s.put(1, ZoneMainTheater, structure("Enemy MT Structure", 0, 5, Keywords{})),
		rsUnit0:    s.put(0, ZoneReserves, unit("Ally Reserves Unit", 1, 5, Keywords{})),
		rsUnit1:    s.put(1, ZoneReserves, unit("Enemy Reserves Unit", 1, 5, Keywords{})),
		rsScratch1: s.putScratch(1),
	}
}

// --- FR-TGT-001: the Zone criterion ---

func TestFR_TGT_001_MainTheaterZoneOffersOnlyMainTheaterAssetsOfEitherSide(t *testing.T) {
	s := newScene(t)
	b := s.zoneBoard()
	s.scriptRefs(0, nil, b.mtUnit1)
	s.deployTargeter(0, []ability.Target{tgtZone(ability.AllegianceAny, 0, false, ability.ZoneMainTheater)}, dmgTo(2, 0))
	asked := s.asked(0)
	if len(asked) != 1 || asked[0].Kind != DecideTarget {
		t.Fatalf("one target question is asked of the controller: %+v", asked)
	}
	wantOptions(t, asked[0], "an asset \"in the Main Theater\" excludes Reserves assets and both Headquarters (FR-TGT-001, FR-ZON-001)",
		b.mtUnit0, b.mtStruct0, b.mtUnit1, b.mtStruct1)
	s.wantDamage(b.mtUnit1, 2)
	if got := totalDamage(s.g); got != 2 {
		t.Errorf("only the chosen asset takes the damage: total %d, want 2", got)
	}
}

func TestFR_TGT_001_ZoneCombinesWithAllegianceAndKindsAsAConjunction(t *testing.T) {
	s := newScene(t)
	b := s.zoneBoard()
	extra := s.put(1, ZoneMainTheater, unit("Second Enemy MT Unit", 1, 5, Keywords{}))
	s.scriptRefs(0, nil, extra)
	s.deployTargeter(0, []ability.Target{tgtZone(ability.AllegianceOpposing, ability.KindUnit, true, ability.ZoneMainTheater)}, dmgTo(1, 0))
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one target question is asked: %+v", asked)
	}
	wantOptions(t, asked[0], "opposing non-Scratch unit in the Main Theater is the conjunction of all three criteria (FR-TGT-001)",
		b.mtUnit1, extra)
	s.wantDamage(extra, 1)
}

func TestFR_TGT_001_ReservesZoneOffersOnlyReservesAssets(t *testing.T) {
	s := newScene(t)
	b := s.zoneBoard()
	s.scriptRefs(0, nil, b.rsUnit0)
	s.deployTargeter(0, []ability.Target{{Filter: ability.Filter{Zone: ability.ZoneReserves, Other: true}}}, dmgTo(1, 0))
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one target question is asked: %+v", asked)
	}
	wantOptions(t, asked[0], "an other asset in Reserves excludes the Main Theater, the Headquarters and the source (FR-TGT-001)",
		b.rsUnit0, b.rsUnit1, b.rsScratch1)
}

// An unset Zone means any battlefield zone: existing definitions are unchanged.
func TestFR_TGT_001_NoZoneCriterionStillMeansAnyBattlefieldZone(t *testing.T) {
	s := newScene(t)
	b := s.zoneBoard()
	s.scriptRefs(0, nil, b.rsUnit1)
	s.deployTargeter(0, []ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(1, 0))
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one target question is asked: %+v", asked)
	}
	wantOptions(t, asked[0], "with no zone stated every opposing unit on the battlefield is an option (FR-TGT-001)",
		b.mtUnit1, b.rsUnit1, b.rsScratch1)
}

// --- FR-TGT-001 / FR-TGT-003: no legal target ---

func TestFR_TGT_003_NoAssetInTheMainTheaterAsksNothingAndOtherClausesRun(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneReserves, unit("Enemy Reserves Unit", 1, 5, Keywords{}))
	s.put(0, ZoneReserves, unit("Ally Reserves Unit", 1, 5, Keywords{}))
	cards := s.onTop(0, vanilla("A")...)
	s.script(0)
	c := trigUnit("Source", targetTrigger([]ability.Target{tgtZone(ability.AllegianceAny, 0, false, ability.ZoneMainTheater)}, dmgTo(2, 0), drawN(1)))
	s.deployFromHand(0, c)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("no asset in the Main Theater means no legal target and no question: asked %d times (FR-TGT-003)", n)
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("the targeted clause does nothing: total damage %d (FR-TGT-003)", got)
	}
	if !inHand(s.g.Players[0], cards[0]) {
		t.Errorf("the non-targeted Draw clause of the same effect still resolves (FR-TGT-003)")
	}
	if assetOf(s.g, c) == nil {
		t.Errorf("the deployed unit stays on the battlefield (FR-TGT-003)")
	}
}

// --- FR-TGT-003: second-stage verification ---

// A chosen target "in the Main Theater" that leaves the Main Theater before the
// ability resolves is illegal. Hero's printed attack triggers are [hit, mayDraw];
// hit is placed first so it resolves last, after mayDraw's answer moves the target.
func TestFR_TGT_003_TargetThatLeavesTheMainTheaterBeforeResolutionIsIllegal(t *testing.T) {
	s := newScene(t)
	hit := attackTrig(false, []ability.Target{tgtZone(ability.AllegianceOpposing, ability.KindUnit, false, ability.ZoneMainTheater)}, dmgTo(3, 0))
	hero := s.put(0, ZoneMainTheater, striker("Hero", 0, 9, Keywords{}, hit, mayDraw()))
	dummy := s.put(1, ZoneMainTheater, unit("Dummy", 0, 9, Keywords{}))
	victim := s.put(1, ZoneMainTheater, unit("Victim", 0, 9, Keywords{}))
	s.midStack(func() { s.moveTo(victim, ZoneReserves) }, victim)
	s.do(s.attack(hero, dummy))
	if victim.Zone != ZoneReserves {
		t.Fatalf("precondition: the victim moved to Reserves before the hit resolved")
	}
	s.wantDamage(victim, 0)
	s.wantDamage(dummy, 0)
}

// A target that is destroyed before resolution is illegal too, and the damage
// is not redirected to another asset in the Main Theater.
func TestFR_TGT_003_TargetDestroyedBeforeResolutionTakesNoDamageAndNothingIsRedirected(t *testing.T) {
	s := newScene(t)
	victim := s.put(1, ZoneMainTheater, unit("Victim", 1, 1, Keywords{}))
	bystander := s.put(1, ZoneMainTheater, unit("Bystander", 1, 9, Keywords{}))
	inMT := []ability.Target{tgtZone(ability.AllegianceOpposing, ability.KindUnit, true, ability.ZoneMainTheater)}
	hit := targetTrigger(inMT, dmgTo(3, 0))
	killer := targetTrigger(inMT, dmgTo(1, 0))
	// Printed [hit, killer]: killer is placed last, so it resolves first and
	// destroys the victim that hit had chosen.
	s.scriptRefs(0, []int{0, 1}, victim, victim)
	s.deployFromHand(0, trigUnit("Source", hit, killer))
	if s.alive(victim) {
		t.Fatalf("precondition: the killer trigger destroyed the victim first")
	}
	s.wantDamage(bystander, 0)
	if len(s.g.Stack) != 0 {
		t.Errorf("both triggers resolved: %d items left", len(s.g.Stack))
	}
}

// --- FR-TGT-001: untargeted "all assets in the Main Theater" ---

func TestFR_TGT_001_AllAssetsInTheMainTheaterHitsExactlyThoseAssets(t *testing.T) {
	s := newScene(t)
	b := s.zoneBoard()
	src := trigUnit("Cataclysm", targetTrigger(nil, allDamage(1, ability.Filter{Zone: ability.ZoneMainTheater})))
	s.deployFromHand(0, src)
	for _, a := range []*Asset{b.mtUnit0, b.mtStruct0, b.mtUnit1, b.mtStruct1} {
		s.wantDamage(a, 1)
	}
	for _, a := range []*Asset{b.rsUnit0, b.rsUnit1, b.rsScratch1, assetOf(s.g, src), s.g.Players[0].HQ, s.g.Players[1].HQ} {
		if a.Damage != 0 {
			t.Errorf("%s is not in the Main Theater and takes no damage (FR-TGT-001, FR-ZON-001): damage %d", a, a.Damage)
		}
	}
}

func TestFR_TGT_001_AllOpposingAssetsInTheMainTheaterCombinesZoneWithAllegiance(t *testing.T) {
	s := newScene(t)
	b := s.zoneBoard()
	f := ability.Filter{Allegiance: ability.AllegianceOpposing, Zone: ability.ZoneMainTheater}
	s.deployFromHand(0, trigUnit("Cataclysm", targetTrigger(nil, allDamage(1, f))))
	s.wantDamage(b.mtUnit1, 1)
	s.wantDamage(b.mtStruct1, 1)
	if got := totalDamage(s.g); got != 2 {
		t.Errorf("only the opposing Main Theater assets take damage: total %d, want 2 (FR-TGT-001)", got)
	}
}
