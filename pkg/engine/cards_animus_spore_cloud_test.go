package engine

import (
	"reflect"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Card tests for Animoceps Spore Cloud (T-0271), written from the CSV rules
// text in data/cards/units-and-structures-animus.csv:18: Animus Structure
// (Fungus), Requisition 4, Tactics 0, Main Theater, power 0, Health 9, yields
// 2/2, rare, "Redirect all damage that would be dealt to allied units in the
// same zone to this structure."
// Rules: FR-REP-004 Stage 1 (redirection, before any defense), FR-REP-001
// (inline, off the stack), SVY-R23 (the affected asset's controller chooses),
// SVY-R44 (the source's current zone), FR-CBT-002, FR-ZON-005.
// Player 0 is the active player.

const sporeCloudName = "Animoceps Spore Cloud"

func sporeAlliedUnits() ability.Filter {
	return ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}
}

func sporeNoReplace(t *testing.T, s *scene) {
	t.Helper()
	if n := len(eventsOfKind(s.g.Log, EventReplace)); n != 0 {
		t.Errorf("the damage target is outside the Cloud's filter, so no replace event; got %d (FR-REP-004 Stage 1)", n)
	}
}

func TestCard_AnimocepsSporeCloud_PrintedStats(t *testing.T) {
	s := newScene(t)
	c := s.card(sporeCloudName)
	if c.Type != TypeStructure || c.Faction != "ANIMUS" {
		t.Errorf("an Animus Structure, got type %v faction %q", c.Type, c.Faction)
	}
	if !c.HasSubtype("Fungus") {
		t.Errorf("subtype Fungus, got %v", c.Subtypes)
	}
	if c.ReqCost != 4 || c.TacCost != 0 {
		t.Errorf("Requisition 4, Tactics 0, got %d and %d", c.ReqCost, c.TacCost)
	}
	if !reflect.DeepEqual(c.DeployZones, []Zone{ZoneMainTheater}) {
		t.Errorf("deploys to the Main Theater only, got %v", c.DeployZones)
	}
	if c.Power != 0 || c.Health != 9 {
		t.Errorf("0/9, got %d/%d", c.Power, c.Health)
	}
	if c.ReqYield != 2 || c.TacYield != 2 {
		t.Errorf("yields 2/2, got %d/%d", c.ReqYield, c.TacYield)
	}
	if !c.Supported() {
		t.Errorf("the card is executable, so Supported accepts it; unsupported lines: %v", c.Unsupported)
	}
}

// Covers: FR-REP-001
func TestCard_AnimocepsSporeCloud_EffectDamageToAlliedUnitLandsOnTheCloud(t *testing.T) {
	s := newScene(t)
	cloud := s.put(0, ZoneMainTheater, s.card(sporeCloudName))
	u := s.put(0, ZoneMainTheater, unit("Ward", 1, 5, Keywords{}))
	s.enactWith(nil, allDamage(3, sporeAlliedUnits()))
	s.wantDamage(u, 0)
	s.wantDamage(cloud, 3)
	reps := eventsOfKind(s.g.Log, EventReplace)
	if len(reps) != 1 || reps[0].Source != cloud.ID || reps[0].Target != u.ID {
		t.Fatalf("want one replace event from the Cloud (%d) for the unit (%d), got %+v (FR-REP-004 Stage 1)", cloud.ID, u.ID, reps)
	}
}

// Covers: FR-REP-004
func TestCard_AnimocepsSporeCloud_CombatHitOnAlliedDefenderLandsOnTheCloud(t *testing.T) {
	s := newScene(t)
	cloud := s.put(1, ZoneMainTheater, s.card(sporeCloudName))
	def := s.put(1, ZoneMainTheater, unit("Ward", 2, 5, Keywords{}))
	att := s.put(0, ZoneMainTheater, unit("Attacker", 4, 10, Keywords{}))
	s.do(s.attack(att, def))
	s.wantDamage(cloud, 4)
	s.wantDamage(def, 0)
	s.wantDamage(att, 2) // the defender's retaliation is its own instance, and the attacker is not the Cloud's ally
}

// Covers: FR-REP-004
func TestCard_AnimocepsSporeCloud_UnitInAnotherZoneIsNotCovered(t *testing.T) {
	s := newScene(t)
	cloud := s.put(0, ZoneMainTheater, s.card(sporeCloudName))
	u := s.put(0, ZoneReserves, unit("Reserve Ward", 1, 5, Keywords{}))
	s.enactWith(nil, allDamage(2, sporeAlliedUnits()))
	s.wantDamage(u, 2)
	s.wantDamage(cloud, 0)
	sporeNoReplace(t, s)
}

func TestCard_AnimocepsSporeCloud_EnemyUnitsAreNotCovered(t *testing.T) {
	s := newScene(t)
	cloud := s.put(0, ZoneMainTheater, s.card(sporeCloudName))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, Keywords{}))
	s.enactWith(nil, allDamage(2, ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}))
	s.wantDamage(foe, 2)
	s.wantDamage(cloud, 0)
	sporeNoReplace(t, s)
}

func TestCard_AnimocepsSporeCloud_AlliedStructuresAreNotCovered(t *testing.T) {
	s := newScene(t)
	cloud := s.put(0, ZoneMainTheater, s.card(sporeCloudName))
	wall := s.put(0, ZoneMainTheater, structure("Wall", 0, 5, Keywords{}))
	s.enactWith(nil, allDamage(2, ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindStructure}))
	s.wantDamage(wall, 2)
	s.wantDamage(cloud, 2) // its own damage as a structure, nothing redirected onto it
	sporeNoReplace(t, s)
}

// Covers: FR-REP-001
func TestCard_AnimocepsSporeCloud_TwoCloudsTheUnitsControllerChooses(t *testing.T) {
	s := newScene(t)
	a := s.put(1, ZoneMainTheater, s.card(sporeCloudName))
	b := s.put(1, ZoneMainTheater, s.card(sporeCloudName))
	u := s.put(1, ZoneMainTheater, unit("Ward", 1, 5, Keywords{}))
	s.script(1, Answer{Picks: []int{1}})
	s.enactWith(nil, allDamage(3, ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}))
	if n := len(replacementAsks(s, 0)); n != 0 {
		t.Errorf("player 0 (the damage's source) was asked %d replacement questions, want 0 (SVY-R23: the affected asset's controller chooses)", n)
	}
	asks := replacementAsks(s, 1)
	if len(asks) != 1 {
		t.Fatalf("player 1 was asked %d replacement questions, want 1 (FR-REP-001, SVY-R23)", len(asks))
	}
	wantOffered(t, asks[0], a.ID, b.ID)
	s.wantDamage(b, 3)
	s.wantDamage(a, 0)
	s.wantDamage(u, 0)
}

// Covers: FR-REP-004
func TestCard_AnimocepsSporeCloud_FaceDownCloudRedirectsNothing(t *testing.T) {
	s := newScene(t)
	down := s.putScratch(0)
	down.Inst.Card = s.card(sporeCloudName)
	s.moveTo(down, ZoneMainTheater)
	horror := unit("Horror", 1, 5, Keywords{})
	horror.Subtypes = []string{"Horror"}
	u := s.put(0, ZoneMainTheater, horror)
	f := sporeAlliedUnits()
	f.Subtypes = []string{"Horror"}
	s.enactWith(nil, allDamage(2, f))
	s.wantDamage(u, 2)
	s.wantDamage(down, 0)
	sporeNoReplace(t, s)
}
