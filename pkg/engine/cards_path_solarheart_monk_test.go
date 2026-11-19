package engine

import (
	"reflect"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Card tests for Solarheart Monk (T-0268), written from the CSV rules text in
// data/cards/units-and-structures-path.csv:32: Path Unit (Specialist,
// Devilhorn), Requisition 8, Tactics 0, Reserves, power 6, Health 8,
// "When this card is put into your discard pile, chosen asset heals 2."
// Rules: FR-ZON-004 (Enter-Discard Triggers), FR-ZON-005, FR-ZON-002,
// FR-ZON-003, FR-KWD-002d (Heal N), FR-TGT-001 (an asset is any Unit,
// Structure or Headquarters on the battlefield, either side).
// Player 0 is the active player.

const monkName = "Solarheart Monk"

// monkPicker answers a target decision with the first option whose Ref is one
// of prefer, in the order given; any other decision gets yesAndTimeouts.
func monkPicker(t *testing.T, prefer ...int) func(Decision) Answer {
	return func(d Decision) Answer {
		if d.Kind != DecideTarget {
			return yesAndTimeouts(d)
		}
		for _, want := range prefer {
			for i, o := range d.Options {
				if o.Ref == want {
					return Answer{Picks: []int{i}}
				}
			}
		}
		t.Errorf("none of the wanted targets %v is among the options %+v", prefer, d.Options)
		return Answer{Timeout: true}
	}
}

func hasOptionRef(d Decision, ref int) bool {
	for _, o := range d.Options {
		if o.Ref == ref {
			return true
		}
	}
	return false
}

func hurtUnit(s *scene, p PlayerID, z Zone, name string) *Asset {
	a := s.put(p, z, unit(name, 1, 5, ability.Keywords{}))
	a.Damage = 3
	return a
}

func TestCard_SolarheartMonk_PrintedStats(t *testing.T) {
	s := newScene(t)
	c := s.card(monkName)
	if c.Type != TypeUnit || c.Faction != "PATH" {
		t.Errorf("a Path Unit, got type %v faction %q", c.Type, c.Faction)
	}
	if !c.HasSubtype("Specialist") || !c.HasSubtype("Devilhorn") {
		t.Errorf("subtypes Specialist and Devilhorn, got %v", c.Subtypes)
	}
	if c.ReqCost != 8 || c.TacCost != 0 {
		t.Errorf("Requisition 8, Tactics 0, got %d and %d", c.ReqCost, c.TacCost)
	}
	if !reflect.DeepEqual(c.DeployZones, []Zone{ZoneReserves}) {
		t.Errorf("deploys to Reserves only, got %v", c.DeployZones)
	}
	if c.Power != 6 || c.Health != 8 {
		t.Errorf("6/8, got %d/%d", c.Power, c.Health)
	}
	if c.ReqYield != 2 || c.TacYield != 2 {
		t.Errorf("yields 2/2, got %d/%d", c.ReqYield, c.TacYield)
	}
	if !reflect.DeepEqual(c.Keywords, ability.Keywords{}) {
		t.Errorf("prints no keywords, got %+v", c.Keywords)
	}
	if !c.Supported() {
		t.Errorf("the card is executable, so Supported accepts it; unsupported lines: %v", c.Unsupported)
	}
}

// Covers: FR-KWD-002d, FR-TGT-001
func TestCard_SolarheartMonk_DiscardedFromHandHealsChosenAsset(t *testing.T) {
	s := newScene(t)
	allied := hurtUnit(s, 0, ZoneMainTheater, "Allied Hurt")
	enemy := hurtUnit(s, 1, ZoneMainTheater, "Enemy Hurt")
	s.byKind(0, monkPicker(t, enemy.ID))
	s.byKind(1, yesAndTimeouts)
	insts := s.enactWith([]*Card{s.card(monkName)}, discardN(1))
	if !inZone(s.g.Players[0].Discard, insts[0]) {
		t.Fatalf("the discarded Monk is in player 0's discard pile (FR-ZON-004)")
	}
	got := askedKind(s, 0, DecideTarget)
	if len(got) != 1 {
		t.Fatalf("player 0 picks the heal's target once, got %d decisions (FR-TGT-001)", len(got))
	}
	for name, ref := range map[string]int{"the allied damaged unit": allied.ID, "the enemy damaged unit": enemy.ID,
		"player 0's Headquarters": s.g.Players[0].HQ.ID, "player 1's Headquarters": s.g.Players[1].HQ.ID} {
		if !hasOptionRef(got[0], ref) {
			t.Errorf("%s is a legal target of 'chosen asset' (FR-TGT-001): %+v", name, got[0].Options)
		}
	}
	s.wantDamage(enemy, 1)
	s.wantDamage(allied, 3)
}

// Covers: FR-ZON-002, FR-KWD-002d
func TestCard_SolarheartMonk_DestroyedOnTheBattlefieldHeals(t *testing.T) {
	s := newScene(t)
	monk := s.put(0, ZoneMainTheater, s.card(monkName))
	hurt := hurtUnit(s, 0, ZoneReserves, "Allied Hurt")
	s.byKind(0, monkPicker(t, monk.ID, hurt.ID))
	s.byKind(1, yesAndTimeouts)
	s.enactAlliedStep(killTarget())
	if s.alive(monk) || !inZone(s.g.Players[0].Discard, monk.Inst) {
		t.Fatalf("the killed Monk is in its owner's discard pile (FR-ZON-002)")
	}
	if got := pileLogPlaces(s, monkName); got != 1 {
		t.Errorf("the enter-discard trigger is placed once, got %d placements (FR-ZON-004)", got)
	}
	s.wantDamage(hurt, 1)
}

// Covers: FR-ZON-004, FR-KWD-002d
func TestCard_SolarheartMonk_DestroyedScratchCompanyHeals(t *testing.T) {
	s := newScene(t)
	s.byKind(0, yesAndTimeouts)
	hurt := hurtUnit(s, 1, ZoneReserves, "Own Hurt")
	s.byKind(1, monkPicker(t, hurt.ID))
	down := s.putScratch(1)
	down.Inst.Card = s.card(monkName)
	s.moveTo(down, ZoneMainTheater)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 5, 9, ability.Keywords{}))
	s.do(s.attack(att, down))
	if s.alive(down) {
		t.Fatalf("precondition: the Scratch Company is destroyed by the attack")
	}
	p1 := s.g.Players[1]
	if !inZone(p1.Discard, down.Inst) || down.Inst.Card.Name != monkName {
		t.Fatalf("the destroyed Scratch Company's Monk is revealed in player 1's discard pile (FR-ZON-005)")
	}
	if got := pileLogPlaces(s, monkName); got != 1 {
		t.Errorf("the face-up Monk's enter-discard trigger is placed once, got %d placements (FR-ZON-004)", got)
	}
	if got := askedKind(s, 1, DecideTarget); len(got) != 1 {
		t.Errorf("player 1 picks the heal's target once, got %d decisions (FR-ZON-004)", len(got))
	}
	s.wantDamage(hurt, 1)
}

// Covers: FR-ZON-002, FR-ZON-003
func TestCard_SolarheartMonk_ControlledMonkIsItsOwnersTrigger(t *testing.T) {
	s := newScene(t)
	monk := s.put(1, ZoneMainTheater, s.card(monkName))
	monk.Controller = 0
	hurt := hurtUnit(s, 1, ZoneReserves, "Owner Hurt")
	s.byKind(0, monkPicker(t, monk.ID))
	s.byKind(1, monkPicker(t, hurt.ID))
	s.enactAlliedStep(killTarget())
	if s.alive(monk) {
		t.Fatalf("precondition: the controlled Monk is killed")
	}
	if !inZone(s.g.Players[1].Discard, monk.Inst) || inZone(s.g.Players[0].Discard, monk.Inst) {
		t.Fatalf("a destroyed asset goes to its owner's pile, player 1's (FR-ZON-002)")
	}
	if got := askedKind(s, 1, DecideTarget); len(got) != 1 {
		t.Errorf("the owner, player 1, picks the heal's target once, got %d decisions (FR-ZON-003)", len(got))
	}
	for _, d := range askedKind(s, 0, DecideTarget) {
		if hasOptionRef(d, hurt.ID) {
			t.Errorf("the controller, player 0, is not asked the heal's target (FR-ZON-003): %+v", d.Options)
		}
	}
	s.wantDamage(hurt, 1)
}

// Covers: FR-ZON-004
func TestCard_SolarheartMonk_ReturnedToHandHealsNothing(t *testing.T) {
	s := newScene(t)
	monk := s.put(0, ZoneMainTheater, s.card(monkName))
	hurt := hurtUnit(s, 0, ZoneReserves, "Allied Hurt")
	s.byKind(0, monkPicker(t, monk.ID, hurt.ID))
	s.byKind(1, yesAndTimeouts)
	s.enactAlliedStep(ability.Step{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget, Target: 0})
	if s.alive(monk) || !inHand(s.g.Players[0], monk.Inst) {
		t.Fatalf("precondition: the Monk is returned to its owner's hand")
	}
	if got := pileLogPlaces(s, monkName); got != 0 {
		t.Errorf("a card returned to hand fires no enter-discard trigger, got %d placements (FR-ZON-004)", got)
	}
	s.wantDamage(hurt, 3)
}
