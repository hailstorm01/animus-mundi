// Tests for T-0317 (FR-MNV-001, FR-TGT-001, FR-TGT-003, FR-MOD-001, FR-LAY-001,
// FR-LAY-003, FR-REP-004, FR-KWD-002b, FR-KWD-002h, FR-KWD-002k, FR-LAY-002,
// FR-CBT-002, FR-ZON-005): the Soulless Maneuver Fire for Effect, loaded from
// data/cards. Player 0 enacts with R0 T4 and is the active player; player 1 is
// the opponent.
//
// Fire for Effect: "Opposing assets in the Main Theater gain Resistant -1 until
// end of turn." / "Deal 1 damage to chosen asset in the Main Theater."
package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// ffeEnact gives player 0 Fire for Effect with exactly R0 T4.
func ffeEnact(s *scene) *Instance {
	s.t.Helper()
	inst := s.give(0, s.card("Fire for Effect"))
	s.setPool(0, 0, 4)
	return inst
}

// ffeResolve enacts Fire for Effect choosing target.
func ffeResolve(s *scene, target *Asset) *Instance {
	s.t.Helper()
	inst := ffeEnact(s)
	chooseAsset(s, target)
	s.do(enactAction(inst))
	return inst
}

// ffeBoard is the full mixed board of the targeting and grant cases.
type ffeBoard struct {
	allyRes, foeRes       *Asset // units in the Reserves
	allyUnit, foeUnit     *Asset // units in the Main Theater
	allyWall, foeWall     *Asset // structures in the Main Theater
	allyScratch, foeScrat *Asset // face-down Scratch Companies in the Main Theater
	allyWard, foeWard     *Asset // Warded units in the Main Theater
}

func ffeBoardOf(s *scene) ffeBoard {
	var b ffeBoard
	b.allyRes = s.put(0, ZoneReserves, unit("AllyRes", 1, 9, ability.Keywords{}))
	b.foeRes = s.put(1, ZoneReserves, unit("FoeRes", 1, 9, ability.Keywords{}))
	b.allyUnit = s.put(0, ZoneMainTheater, unit("AllyMain", 1, 9, ability.Keywords{}))
	b.foeUnit = s.put(1, ZoneMainTheater, unit("FoeMain", 1, 9, ability.Keywords{}))
	b.allyWall = s.put(0, ZoneMainTheater, structure("AllyWall", 0, 9, ability.Keywords{}))
	b.foeWall = s.put(1, ZoneMainTheater, structure("FoeWall", 0, 9, ability.Keywords{}))
	b.allyScratch = s.putScratch(0)
	s.moveTo(b.allyScratch, ZoneMainTheater)
	b.foeScrat = s.putScratch(1)
	s.moveTo(b.foeScrat, ZoneMainTheater)
	b.allyWard = s.put(0, ZoneMainTheater, unit("AllyWard", 1, 9, ability.Keywords{Ward: true}))
	b.foeWard = s.put(1, ZoneMainTheater, unit("FoeWard", 1, 9, ability.Keywords{Ward: true}))
	return b
}

func TestCard_FireForEffect_ChosenOpposingAssetTakesTwo(t *testing.T) {
	// Covers: FR-REP-004, FR-KWD-002b, FR-LAY-001
	s := newScene(t)
	hit := s.put(1, ZoneMainTheater, unit("Hit", 1, 9, ability.Keywords{}))
	other := s.put(1, ZoneMainTheater, unit("Other", 1, 9, ability.Keywords{}))
	ffeResolve(s, hit)
	s.wantDamage(hit, 2)
	if n := damageEventsOn(s, hit); n != 1 {
		t.Errorf("the chosen unit takes one damage instance: %d events, want 1 (FR-REP-004)", n)
	}
	for _, e := range damageEvents(s.g.Log) {
		if e.Target == hit.ID && (e.Combat || e.Raw != 1 || e.Resistant != -1) {
			t.Errorf("the instance is non-combat, raw 1, against Resistant -1 (the grant resolves first): %+v (FR-REP-004 Stage 2)", e)
		}
	}
	s.wantDamage(other, 0)
	s.wantResistant(hit, -1, "Fire for Effect grants Resistant -1 to opposing Main Theater assets")
	s.wantResistant(other, -1, "Fire for Effect grants Resistant -1 to opposing Main Theater assets")
}

func TestCard_FireForEffect_ChosenAlliedAssetTakesOne(t *testing.T) {
	// Covers: FR-TGT-001, FR-LAY-001
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, ability.Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 9, ability.Keywords{}))
	ffeResolve(s, ally)
	s.wantDamage(ally, 1)
	s.wantResistant(ally, 0, "the grant reaches only opposing assets")
	s.wantResistant(foe, -1, "the opposing unit gains Resistant -1")
	s.wantDamage(foe, 0)
}

func TestCard_FireForEffect_OffersEveryMainTheaterAsset(t *testing.T) {
	// Covers: FR-TGT-001, FR-KWD-002h, FR-ZON-005, SVY-R59
	s := newScene(t)
	b := ffeBoardOf(s)
	ffeResolve(s, b.allyUnit)
	var asks []Decision
	for _, d := range s.asked(0) {
		if d.Kind == DecideTarget {
			asks = append(asks, d)
		}
	}
	if len(asks) != 1 {
		t.Fatalf("one target slot, so exactly one target question: asked %d (FR-MNV-001)", len(asks))
	}
	wantOptions(t, asks[0], "the Main Theater units, structures, Scratch Companies and the allied Warded unit",
		b.allyUnit, b.foeUnit, b.allyWall, b.foeWall, b.allyScratch, b.foeScrat, b.allyWard)
}

func TestCard_FireForEffect_GrantReachesOnlyOpposingMainTheaterAssets(t *testing.T) {
	// Covers: FR-LAY-001, FR-ZON-005
	s := newScene(t)
	b := ffeBoardOf(s)
	ffeResolve(s, b.allyUnit)
	for _, a := range []*Asset{b.foeUnit, b.foeWall, b.foeScrat, b.foeWard} {
		s.wantResistant(a, -1, "an opposing Main Theater asset gains Resistant -1 (the grant is not a target, so Ward does not stop it)")
	}
	for _, a := range []*Asset{b.foeRes, s.g.Players[1].HQ, s.g.Players[0].HQ, b.allyRes,
		b.allyUnit, b.allyWall, b.allyScratch, b.allyWard} {
		s.wantResistant(a, 0, "only opposing Main Theater assets gain the grant")
	}
}

func TestCard_FireForEffect_AddsToPrintedResistant(t *testing.T) {
	// Covers: FR-KWD-002b, FR-REP-004
	s := newScene(t)
	one := s.put(1, ZoneMainTheater, unit("One", 1, 9, ability.Keywords{Resistant: 1}))
	ffeResolve(s, one)
	s.wantResistant(one, 0, "printed Resistant 1 plus the grant of -1")
	s.wantDamage(one, 1)

	s2 := newScene(t)
	two := s2.put(1, ZoneMainTheater, unit("Two", 1, 9, ability.Keywords{Resistant: 2}))
	ffeResolve(s2, two)
	s2.wantResistant(two, 1, "printed Resistant 2 plus the grant of -1")
	s2.wantDamage(two, 0)
}

func TestCard_FireForEffect_ArmorAbsorbs(t *testing.T) {
	// Covers: FR-REP-004, FR-KWD-002k
	s := newScene(t)
	arm := s.put(1, ZoneMainTheater, unit("Armored", 1, 9, ability.Keywords{Armor: 1}))
	ffeResolve(s, arm)
	if arm.Armor != 0 {
		t.Errorf("the amplified 2 damage removes the 1 Armor: Armor %d, want 0 (FR-KWD-002k)", arm.Armor)
	}
	s.wantDamage(arm, 1)
}

func TestCard_FireForEffect_LethalDamageKills(t *testing.T) {
	// Covers: FR-LAY-002
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Frail", 1, 2, ability.Keywords{}))
	ffeResolve(s, foe)
	if s.alive(foe) {
		t.Errorf("a unit with health 2 takes 2 damage and is destroyed (FR-LAY-002)")
	}
	if !inDiscardPile(s.g.Players[1], foe.Inst) {
		t.Errorf("the destroyed unit's card is in its owner's discard pile (FR-LAY-002)")
	}
}

func TestCard_FireForEffect_LaterDamageThisTurnIsAmplified(t *testing.T) {
	// Covers: FR-KWD-002b, FR-REP-004, FR-CBT-002
	s := newScene(t)
	chosen := s.put(1, ZoneMainTheater, unit("Chosen", 1, 20, ability.Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Defender", 1, 20, ability.Keywords{}))
	att := s.put(0, ZoneMainTheater, unit("Attacker", 3, 20, ability.Keywords{}))
	ping := s.put(0, ZoneMainTheater, actUnit("Pinger", nil, actAbility(nil, ability.Step{
		Op: ability.OpDealDamage, Amount: ability.Value{N: 1}, Subject: ability.SubjectAll,
		Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Zone: ability.ZoneMainTheater}})))
	ffeResolve(s, chosen)
	s.do(s.attack(att, def))
	var combat *Event
	for _, e := range damageEvents(s.g.Log) {
		if e.Target == def.ID && e.Combat {
			e := e
			combat = &e
		}
	}
	if combat == nil || combat.Amount != 4 {
		t.Errorf("3 combat damage against Resistant -1 deals 4: got %+v (FR-KWD-002b, FR-CBT-002)", combat)
	}
	before := def.Damage
	s.do(activate(ping, 0))
	if got := def.Damage - before; got != 2 {
		t.Errorf("a later non-combat 1 damage this turn deals 2 against Resistant -1: got %d (FR-REP-004 Stage 2)", got)
	}
}

func TestCard_FireForEffect_GrantEndsAtEndOfTurn(t *testing.T) {
	// Covers: FR-LAY-003
	s := newScene(t)
	plain := s.put(1, ZoneMainTheater, unit("Plain", 1, 9, ability.Keywords{}))
	res := s.put(1, ZoneMainTheater, unit("Res", 1, 9, ability.Keywords{Resistant: 1}))
	ffeResolve(s, plain)
	s.wantResistant(plain, -1, "setup: the grant is active this turn")
	s.endTurn()
	s.wantResistant(plain, 0, "the grant ends at end of turn")
	s.wantResistant(res, 1, "the printed Resistant 1 is back at end of turn")
}

func TestCard_FireForEffect_SnapshotLocksTheGrant(t *testing.T) {
	// Covers: FR-LAY-001
	s := newScene(t)
	fresh := s.put(1, ZoneReserves, unit("Fresh", 1, 9, ability.Keywords{}))
	left := s.put(1, ZoneMainTheater, unit("Leaver", 1, 9, ability.Keywords{}))
	chosen := s.put(1, ZoneMainTheater, unit("Chosen", 1, 9, ability.Keywords{}))
	ffeResolve(s, chosen)
	s.moveTo(fresh, ZoneMainTheater)
	s.moveTo(left, ZoneReserves)
	s.wantResistant(fresh, 0, "a unit entering the Main Theater after resolution is not granted (snapshot locking)")
	s.wantResistant(left, -1, "a unit that had the grant keeps it after leaving the Main Theater until end of turn")
}

func TestCard_FireForEffect_OneLegalTargetAsksNothing(t *testing.T) {
	// Covers: FR-MOD-001
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Only", 1, 9, ability.Keywords{}))
	inst := ffeEnact(s)
	s.script(0)
	s.do(enactAction(inst))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("one legal target is taken without asking: asked %d times (Single-Option Fast-Path)", n)
	}
	s.wantDamage(foe, 2)
}

func TestCard_FireForEffect_TimeoutTakesTheFirstOption(t *testing.T) {
	// Covers: FR-MOD-001, SVY-R16
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 9, ability.Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, ability.Keywords{}))
	inst := ffeEnact(s)
	s.script(0, Answer{Timeout: true})
	s.do(enactAction(inst))
	s.wantDamage(ally, 1)
	s.wantDamage(foe, 0)
	if got := totalDamage(s.g); got != 1 {
		t.Errorf("a timed-out target decision takes the first option in fixed order (the allied unit): total damage %d, want 1 (SVY-R16)", got)
	}
}

func TestCard_FireForEffect_NotEnactableWithAnEmptyMainTheater(t *testing.T) {
	// Covers: SVY-R29, FR-MNV-001
	s := newScene(t)
	s.put(0, ZoneReserves, unit("AllyRes", 1, 9, ability.Keywords{}))
	s.put(1, ZoneReserves, unit("FoeRes", 1, 9, ability.Keywords{}))
	inst := ffeEnact(s)
	if offersEnact(s.g, inst) {
		t.Errorf("with no asset in the Main Theater there is no legal target, so the Enact is not offered (SVY-R29)")
	}
	s.put(0, ZoneMainTheater, unit("AllyMain", 1, 9, ability.Keywords{}))
	if !offersEnact(s.g, inst) {
		t.Errorf("a single allied Main Theater unit is a legal target, so the Enact is offered (FR-MNV-001)")
	}
}

func TestCard_FireForEffect_TargetGoneStillGrants(t *testing.T) {
	// Covers: FR-TGT-003, SVY-R25
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Fleeting", 1, 9, ability.Keywords{}))
	other := s.put(1, ZoneMainTheater, unit("Other", 1, 9, ability.Keywords{}))
	inst := ffeEnact(s)
	chooseAsset(s, foe)
	r := &recorder{}
	r.onEvent = func(e Event) {
		if e.Kind == "place" && e.Card == "Fire for Effect" && s.alive(foe) {
			stormKill(s, foe)
		}
	}
	s.g.obs = r
	s.do(enactAction(inst))
	if s.alive(foe) {
		t.Fatalf("setup: the chosen unit should have left before resolution")
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("the target is gone, so nothing is dealt damage: total %d (FR-TGT-003)", got)
	}
	s.wantResistant(other, -1, "the untargeted grant still applies when the target is gone (FR-TGT-003)")
	if !inDiscardPile(s.g.Players[0], inst) {
		t.Errorf("Fire for Effect still ends in its owner's discard pile (FR-TGT-003)")
	}
}

func TestCard_FireForEffect_PaysFourTactics(t *testing.T) {
	// Covers: FR-MNV-001
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 9, ability.Keywords{}))
	inst := s.give(0, s.card("Fire for Effect"))
	s.setPool(0, 0, 3)
	if offersEnact(s.g, inst) {
		t.Errorf("with R0 T3 the Enact is not offered (FR-MNV-001)")
	}
	s.setPool(0, 0, 4)
	if !offersEnact(s.g, inst) {
		t.Fatalf("with R0 T4 the Enact is offered (FR-MNV-001)")
	}
	chooseAsset(s, foe)
	s.do(enactAction(inst))
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
		t.Errorf("the pool is empty after it resolves: R%d T%d (FR-MNV-001)", p.Req, p.Tac)
	}
}
