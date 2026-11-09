package engine

import "testing"

// Card tests for Macabre Dancers (T-0199), written from the CSV rules text in
// data/cards/units-and-structures-path.csv: Path Unit, 2 Requisition, Reserves,
// 3/6, "Unbowed: Heal 2." and "Cornered: Resistant 1."
// Rules: FR-KWD-002a (Unbowed, lethal Unbowed fizzles), FR-KWD-002d (Heal N),
// FR-KWD-002e and FR-KWD-002b (Cornered keyword), FR-REP-004, SVY-R13, SVY-R18.

// dancersScene puts Macabre Dancers on player 1's Main Theater facing one
// player 0 hitter of the given power (health 20, so it survives the Dancers'
// 3 power). If cornered, player 0 controls a second Main Theater asset, so
// player 1 has 1 against 2 (FR-KWD-002e); otherwise 1 against 1.
func dancersScene(t *testing.T, cornered bool, hitPower int) (s *scene, dancers, hitter *Asset) {
	t.Helper()
	s = newScene(t)
	dancers = s.put(1, ZoneMainTheater, s.card("Macabre Dancers"))
	hitter = s.put(0, ZoneMainTheater, unit("Hitter", hitPower, 20, Keywords{}))
	if cornered {
		s.put(0, ZoneMainTheater, unit("Second Enemy", 0, 5, Keywords{}))
	}
	return s, dancers, hitter
}

// Covers: FR-KWD-002a, FR-KWD-002d
func TestCard_MacabreDancers_UnbowedHealsTwoAfterSurvivingDamage(t *testing.T) {
	s, d, h := dancersScene(t, false, 3)
	s.do(s.attack(h, d))
	if !s.alive(d) {
		t.Fatalf("precondition: 3 damage on 6 Health leaves Macabre Dancers alive")
	}
	s.wantDamage(d, 1) // 3 taken, Unbowed: Heal 2
	if len(s.g.Stack) != 0 {
		t.Errorf("the Unbowed trigger has resolved: %d items left on the stack", len(s.g.Stack))
	}
}

// Covers: FR-KWD-002d
func TestCard_MacabreDancers_HealNeverRemovesMoreThanItsDamage(t *testing.T) {
	s, d, h := dancersScene(t, false, 1)
	s.do(s.attack(h, d))
	s.wantDamage(d, 0) // 1 taken, Heal 2 removes only 1
	if cur, max := s.g.CurrentHealth(d), s.g.MaxHealth(d); cur != max {
		t.Errorf("Heal never leaves damage below 0 nor Current Health above Effective Maximum Health: Current %d, Max %d", cur, max)
	}
	if max := s.g.MaxHealth(d); max != 6 {
		t.Errorf("Macabre Dancers has 6 Health: Effective Maximum Health %d", max)
	}
}

// Covers: FR-KWD-002a, FR-TGT-003, FR-ZON-004
func TestCard_MacabreDancers_LethalUnbowedTriggersAndFizzles(t *testing.T) {
	s, d, h := dancersScene(t, false, 6)
	inst := d.Inst
	var placedWhileAlive []bool
	r := &recorder{}
	r.onEvent = func(e Event) {
		if e.Kind == "place" {
			if it := s.g.Stack[len(s.g.Stack)-1]; it.Kind == StackAbility && it.Source == d.ID {
				placedWhileAlive = append(placedWhileAlive, s.alive(d))
			}
		}
	}
	s.g.obs = r
	s.do(s.attack(h, d))
	if s.alive(d) {
		t.Fatalf("Heal 2 cannot save the asset: 6 damage on 6 Health destroys it at the death sweep (SVY-R13)")
	}
	if !inZone(s.g.Players[1].Discard, inst) {
		t.Errorf("Macabre Dancers is in its owner's discard pile and stays there: the self-aimed Heal fizzles (FR-KWD-002a)")
	}
	if len(placedWhileAlive) != 1 {
		t.Fatalf("a lethal Unbowed still triggers: want 1 trigger placed, got %d", len(placedWhileAlive))
	}
	if placedWhileAlive[0] {
		t.Errorf("the lethal Unbowed trigger is placed after the death sweep, when the asset is already gone (SVY-R13)")
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the trigger resolved: %d items left on the stack", len(s.g.Stack))
	}
	if s.g.Asset(d.ID) != nil {
		t.Errorf("the fizzled Heal does not bring Macabre Dancers back to the battlefield")
	}
	if err := s.g.CheckInvariants(); err != nil {
		t.Errorf("state after the fizzled Heal: %v", err)
	}
}

// Covers: FR-KWD-002e, FR-KWD-002b, FR-REP-004, SVY-R18
func TestCard_MacabreDancers_CorneredResistantReducesDamage(t *testing.T) {
	s, d, h := dancersScene(t, true, 3)
	if got := s.g.Keywords(d).Resistant; got != 1 {
		t.Fatalf("Cornered (1 against 2): Macabre Dancers has Resistant 1, got %d (SVY-R18)", got)
	}
	s.do(s.attack(h, d))
	if e := damageEvent(t, s, d); e.Amount != 2 {
		t.Errorf("3 damage less Resistant 1 is 2 taken: damage event amount %d", e.Amount)
	}
	s.wantDamage(d, 0) // 2 taken, Heal 2 removes both
}

// Covers: FR-KWD-002a, FR-REP-004, SVY-R18
func TestCard_MacabreDancers_CorneredResistantStopsOneDamageAndUnbowed(t *testing.T) {
	s, d, h := dancersScene(t, true, 1)
	items := recordPlaced(s)
	s.do(s.attack(h, d))
	s.wantDamage(d, 0)
	for _, it := range *items {
		if it.Kind == StackAbility && it.Source == d.ID {
			t.Errorf("1 damage less Resistant 1 is 0, so no damage is taken and Unbowed does not trigger: placed %+v (FR-REP-004, FR-KWD-002a)", it)
		}
	}
}

// Covers: SVY-R18
func TestCard_MacabreDancers_NotCorneredHasNoResistant(t *testing.T) {
	s, d, h := dancersScene(t, false, 1)
	if got := s.g.Keywords(d).Resistant; got != 0 {
		t.Fatalf("not Cornered (1 against 1): Resistant 0, got %d (SVY-R18)", got)
	}
	items := recordPlaced(s)
	s.do(s.attack(h, d))
	if e := damageEvent(t, s, d); e.Amount != 1 {
		t.Errorf("without Resistant, 1 damage is 1 taken: damage event amount %d", e.Amount)
	}
	n := 0
	for _, it := range *items {
		if it.Kind == StackAbility && it.Source == d.ID {
			n++
		}
	}
	if n != 1 {
		t.Errorf("1 damage taken places one Unbowed trigger: got %d", n)
	}
	s.wantDamage(d, 0) // Heal 2 removes the 1
}
