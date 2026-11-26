package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0318 (Hellfire Armorers): "Ward" / "Allied units in the same zone have Aggressive 1."
// Player 0 controls the Armorers and is the active player. Every defender has power 0 and enough health to survive.

func hfAggressiveItems(items []StackItem) int {
	n := 0
	for _, it := range items {
		if it.FromKeyword == ability.KeywordTriggerAggressive {
			n++
		}
	}
	return n
}

func hfFoe(s *scene) *Asset {
	return s.put(1, ZoneMainTheater, unit("Dummy", 0, 20, ability.Keywords{}))
}

func TestCard_HellfireArmorers_AlliedUnitsInItsZone(t *testing.T) {
	// Covers: FR-LAY-001, FR-TGT-001 (SVY-R44), FR-KWD-002m
	s := newScene(t)
	arm := s.put(0, ZoneMainTheater, s.card("Hellfire Armorers"))
	u := s.put(0, ZoneMainTheater, unit("Ally", 2, 5, ability.Keywords{}))
	scratch := s.putScratch(0)
	s.moveTo(scratch, ZoneMainTheater)
	st := s.put(0, ZoneMainTheater, structure("Wall", 1, 5, ability.Keywords{}))
	res := s.put(0, ZoneReserves, unit("Reserve Ally", 2, 5, ability.Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 2, 5, ability.Keywords{}))
	s.wantPower(arm, 3, "the grant is a keyword, not power")
	s.wantPower(u, 2, "the grant is a keyword, not power")
	s.wantPower(scratch, 1, "the grant is a keyword, not power")
	for _, c := range []struct {
		a    *Asset
		want int
		why  string
	}{
		{arm, 1, "allied units includes the Armorers itself"},
		{u, 1, "an allied unit in the Armorers' zone has Aggressive 1"},
		{scratch, 1, "a face-down Scratch Company is a unit"},
		{st, 0, "an allied structure is not a unit"},
		{res, 0, "an allied unit in Reserves is not in the Armorers' zone"},
		{foe, 0, "an opposing unit is not allied"},
	} {
		if got := s.g.Keywords(c.a).Aggressive; got != c.want {
			t.Errorf("%s Aggressive = %d, want %d (%s)", c.a, got, c.want, c.why)
		}
	}
}

func TestCard_HellfireArmorers_HasWard(t *testing.T) {
	// Covers: FR-KWD-002h
	s := newScene(t)
	allyArm := s.put(0, ZoneMainTheater, s.card("Hellfire Armorers"))
	foeArm := s.put(1, ZoneMainTheater, s.card("Hellfire Armorers"))
	if !s.g.Keywords(allyArm).Ward {
		t.Errorf("Hellfire Armorers has Ward (FR-KWD-002h)")
	}
	stormChoose(s, func() *Asset { return allyArm })
	rifles := crDeployCard(s, s.card("Conscript Rifles"))
	d := crOneQuestion(t, s)
	crWantOffered(t, d, "player 0's own Warded Armorers is offered, the opposing one is not (FR-KWD-002h)", allyArm, assetOf(s.g, rifles.Card))
	s.wantDamage(foeArm, 0)
}

func TestCard_HellfireArmorers_AlliedAttackerGetsOneMorePower(t *testing.T) {
	// Covers: FR-KWD-002f, FR-CBT-002, FR-LAY-003
	s := newScene(t)
	s.put(0, ZoneMainTheater, s.card("Hellfire Armorers"))
	att := s.put(0, ZoneMainTheater, unit("Attacker", 2, 9, ability.Keywords{}))
	def := hfFoe(s)
	items := recordPlaced(s)
	s.do(s.attack(att, def))
	if n := hfAggressiveItems(*items); n != 1 {
		t.Errorf("one Aggressive item is placed for the granted Aggressive 1: %d (FR-KWD-002f)", n)
	}
	s.wantDamage(def, 3)
	s.wantPower(att, 3, "the Aggressive bonus lasts until end of turn")
	s.endTurn()
	s.wantPower(att, 2, "the bonus expires at end of turn (FR-LAY-003)")
}

func TestCard_HellfireArmorers_ItselfAttacksWithOneMore(t *testing.T) {
	// Covers: FR-KWD-002f
	s := newScene(t)
	arm := s.put(0, ZoneReserves, s.card("Hellfire Armorers"))
	s.moveTo(arm, ZoneMainTheater)
	def := hfFoe(s)
	s.do(s.attack(arm, def))
	s.wantDamage(def, 4)
}

func TestCard_HellfireArmorers_AddsToPrintedAggressive(t *testing.T) {
	// Covers: FR-KWD-002f, FR-LAY-001
	s := newScene(t)
	s.put(0, ZoneMainTheater, s.card("Hellfire Armorers"))
	att := s.put(0, ZoneMainTheater, unit("Raider", 1, 9, ability.Keywords{Aggressive: 2}))
	def := hfFoe(s)
	if got := s.g.Keywords(att).Aggressive; got != 3 {
		t.Errorf("printed Aggressive 2 plus the granted 1 is Aggressive 3: got %d (FR-LAY-001)", got)
	}
	s.do(s.attack(att, def))
	s.wantDamage(def, 4)
}

func TestCard_HellfireArmorers_TwoArmorersAddUp(t *testing.T) {
	// Covers: FR-LAY-001, FR-KWD-002f
	s := newScene(t)
	s.put(0, ZoneMainTheater, s.card("Hellfire Armorers"))
	s.put(0, ZoneMainTheater, s.card("Hellfire Armorers"))
	att := s.put(0, ZoneMainTheater, unit("Attacker", 2, 9, ability.Keywords{}))
	def := hfFoe(s)
	s.script(0, Answer{Timeout: true})
	s.do(s.attack(att, def))
	s.wantDamage(def, 4)
}

func TestCard_HellfireArmorers_FollowsItsZone(t *testing.T) {
	// Covers: FR-TGT-001 (SVY-R44)
	s := newScene(t)
	arm := s.put(0, ZoneMainTheater, s.card("Hellfire Armorers"))
	main := s.put(0, ZoneMainTheater, unit("Main", 2, 5, ability.Keywords{}))
	res := s.put(0, ZoneReserves, unit("Res", 2, 5, ability.Keywords{}))
	s.moveTo(arm, ZoneReserves)
	foeRes := s.put(1, ZoneReserves, unit("FoeRes", 2, 5, ability.Keywords{}))
	for _, c := range []struct {
		a    *Asset
		want int
		why  string
	}{
		{res, 1, "the Armorers is now in player 0's Reserves"},
		{main, 0, "the Armorers left the Main Theater"},
		{arm, 1, "the Armorers still includes itself"},
		{foeRes, 0, "player 1's Reserves are not player 0's Reserves"},
	} {
		if got := s.g.Keywords(c.a).Aggressive; got != c.want {
			t.Errorf("%s Aggressive = %d, want %d (%s)", c.a, got, c.want, c.why)
		}
	}
}

func TestCard_HellfireArmorers_GoneBeforeTheAttackGrantsNothing(t *testing.T) {
	// Covers: FR-ZON-004, FR-KWD-002f
	s := newScene(t)
	arm := s.put(0, ZoneMainTheater, s.card("Hellfire Armorers"))
	att := s.put(0, ZoneMainTheater, unit("Attacker", 2, 9, ability.Keywords{}))
	def := hfFoe(s)
	stormKill(s, arm)
	if got := s.g.Keywords(att).Aggressive; got != 0 {
		t.Errorf("an Armorers that left the battlefield grants nothing: Aggressive %d (FR-ZON-004)", got)
	}
	items := recordPlaced(s)
	s.do(s.attack(att, def))
	if n := hfAggressiveItems(*items); n != 0 {
		t.Errorf("no Aggressive item is placed: %d (FR-KWD-002f)", n)
	}
	s.wantDamage(def, 2)
}

func TestCard_HellfireArmorers_BonusOutlastsTheAura(t *testing.T) {
	// Covers: FR-KWD-002f, FR-LAY-001 (One-Shot Modifier Snapshot Locking)
	s := newScene(t)
	arm := s.put(0, ZoneMainTheater, s.card("Hellfire Armorers"))
	att := s.put(0, ZoneMainTheater, unit("Attacker", 2, 9, ability.Keywords{}))
	def := hfFoe(s)
	s.do(s.attack(att, def))
	s.wantDamage(def, 3)
	stormKill(s, arm)
	s.wantPower(att, 3, "the resolved Aggressive bonus is locked in until end of turn")
	s.endTurn()
	s.wantPower(att, 2, "the bonus expires at end of turn")
}
