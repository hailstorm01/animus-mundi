package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for damage redirection (T-0269): a WouldBeDealtDamage replacement deals
// the damage to its source instead, in Stage 1 of damage (FR-REP-004), inline
// and off the stack (FR-REP-001), with the affected asset's controller choosing
// between several (FR-REP-001, SVY-R23) and "in the same zone" reading the
// source's current zone (SVY-R44).
//
// Cards are synthetic. A "redirector" prints: redirect all damage that would be
// dealt to allied units in the same zone to this structure.

var (
	alliedUnitsOnly = ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}
	redirectEvent   = ability.Replacement{
		Event:  ability.WouldBeDealtDamage,
		Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, ZoneRef: ability.ZoneRefSame},
	}
)

// redirector is a structure printing the redirection above, with the given
// Maximum Health and keywords, and any triggers.
func redirector(name string, health int, kw ability.Keywords, trigs ...ability.Trigger) *Card {
	c := structure(name, 0, health, kw)
	c.Ability = &ability.Def{Triggers: trigs, Replacements: []ability.Replacement{redirectEvent}}
	return c
}

// damageEvents returns the log's damage events, in order.
func redirectDamageEvents(s *scene) []Event { return eventsOfKind(s.g.Log, EventDamage) }

// dealToAlliedUnits has player 0 enact a free Maneuver dealing n damage to
// every allied unit, one damage instance per unit.
func (s *scene) dealToAlliedUnits(n int) { s.enactWith(nil, allDamage(n, alliedUnitsOnly)) }

func wantNoReplace(t *testing.T, s *scene) {
	t.Helper()
	if n := len(eventsOfKind(s.g.Log, EventReplace)); n != 0 {
		t.Errorf("the damage target is outside the redirection's filter, so no replace event; got %d (FR-REP-004 Stage 1)", n)
	}
}

func TestFR_REP_004_RedirectDealsEffectDamageToTheSource(t *testing.T) {
	s := newScene(t)
	red := s.put(0, ZoneMainTheater, redirector("Cloud", 9, ability.Keywords{}))
	u := s.put(0, ZoneMainTheater, unit("Ward", 1, 5, ability.Keywords{}))
	u.Damage = 4
	s.dealToAlliedUnits(3)
	if !s.alive(u) {
		t.Errorf("damage redirected away from the unit never destroys it (FR-REP-001 Damage vs. Destruction Precedence)")
	}
	s.wantDamage(u, 4)
	s.wantDamage(red, 3)
	reps := eventsOfKind(s.g.Log, EventReplace)
	if len(reps) != 1 || reps[0].Source != red.ID || reps[0].Target != u.ID {
		t.Fatalf("want one replace event from the redirector (%d) for the unit (%d), got %+v (FR-REP-004 Stage 1)", red.ID, u.ID, reps)
	}
	dmg := redirectDamageEvents(s)
	if len(dmg) != 1 || dmg[0].Target != red.ID {
		t.Fatalf("want one damage event, aimed at the redirector (%d), got %+v (FR-REP-004 Stage 1)", red.ID, dmg)
	}
	if reps[0].Seq > dmg[0].Seq {
		t.Errorf("the replace event (seq %d) precedes the damage event (seq %d) (FR-REP-004 Stage 1)", reps[0].Seq, dmg[0].Seq)
	}
}

func TestFR_REP_004_RedirectedDamageUsesTheNewTargetsDefenses(t *testing.T) {
	s := newScene(t)
	red := s.put(0, ZoneMainTheater, redirector("Cloud", 9, ability.Keywords{Resistant: 1, Armor: 1}))
	u := s.put(0, ZoneMainTheater, unit("Ward", 1, 5, ability.Keywords{Armor: 2}))
	s.dealToAlliedUnits(3)
	if u.Armor != 2 || u.Damage != 0 {
		t.Errorf("the original target's Armor and Damage are untouched: Armor %d (want 2), Damage %d (want 0) (FR-REP-004 Stage 1)", u.Armor, u.Damage)
	}
	if red.Armor != 0 {
		t.Errorf("the redirector's Armor of 1 absorbs 1 damage, Armor = %d, want 0 (FR-REP-004 Stages 2-5 use the new target's defenses)", red.Armor)
	}
	s.wantDamage(red, 1) // 3 - 1 Resistant = 2, 1 absorbed by Armor, 1 committed
}

func TestFR_REP_004_RedirectedCombatDamageStaysCombatDamage(t *testing.T) {
	s := newScene(t)
	red := s.put(1, ZoneMainTheater, redirector("Cloud", 9, ability.Keywords{Stalwart: true}))
	def := s.put(1, ZoneMainTheater, unit("Ward", 2, 5, ability.Keywords{}))
	att := s.put(0, ZoneMainTheater, unit("Attacker", 4, 10, ability.Keywords{}))
	s.do(s.attack(att, def))
	s.wantDamage(red, 1) // Stalwart caps combat damage at 1, so the damage stayed combat damage
	s.wantDamage(def, 0)
	s.wantDamage(att, 2) // the defender's retaliation is its own instance, and the attacker isn't the redirector's ally
	var redirected *Event
	for _, e := range redirectDamageEvents(s) {
		if e.Target == red.ID {
			e := e
			redirected = &e
		}
	}
	if redirected == nil {
		t.Fatalf("no damage event for the redirector (FR-REP-004 Stage 1)")
	}
	if !redirected.Combat || redirected.Source != att.ID {
		t.Errorf("the redirected damage is still combat damage dealt by the attacker (%d): %+v (FR-REP-004 Stage 3, FR-CBT-002)", att.ID, *redirected)
	}
}

// Covers: FR-CBT-002
func TestFR_REP_004_RedirectedLethalCombatDamageIsNoCombatKill(t *testing.T) {
	s := newScene(t)
	red := s.put(1, ZoneMainTheater, redirector("Cloud", 3, ability.Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Ward", 2, 5, ability.Keywords{}))
	att := s.put(0, ZoneMainTheater, striker("Slayer", 4, 10, ability.Keywords{}, combatKillTrigger(drawN(1))))
	deck := len(s.g.Players[0].Deck)
	s.do(s.attack(att, def))
	if s.alive(red) {
		t.Errorf("the redirector took lethal combat damage and is destroyed at the sweep (FR-REP-004 Stage 5, FR-LAY-002)")
	}
	if !s.alive(def) {
		t.Errorf("the defending unit took nothing and survives (FR-REP-004 Stage 1)")
	}
	if got := len(s.g.Players[0].Deck); got != deck {
		t.Errorf("the defender was not destroyed, so there is no combat kill and no draw: deck %d -> %d (FR-CBT-002 Step 3)", deck, got)
	}
}

func TestFR_REP_004_RedirectReadsTheSourcesCurrentZone(t *testing.T) {
	// SVY-R44: a continuous ability isn't on the stack and uses its source's current zone.
	cases := []struct {
		name       string
		redZone    Zone
		moveRedTo  Zone // zero: never moved
		unitZone   Zone
		redirected bool
	}{
		{"a: source in Main Theater, unit in Reserves", ZoneMainTheater, 0, ZoneReserves, false},
		{"b: source in Reserves, unit in Reserves", ZoneReserves, 0, ZoneReserves, true},
		{"c: source in Reserves, unit in Main Theater", ZoneReserves, 0, ZoneMainTheater, false},
		{"d1: source moved to Reserves, unit in Main Theater", ZoneMainTheater, ZoneReserves, ZoneMainTheater, false},
		{"d2: source moved to Reserves, unit in Reserves", ZoneMainTheater, ZoneReserves, ZoneReserves, true},
		{"e: source in Main Theater, unit in Main Theater", ZoneMainTheater, 0, ZoneMainTheater, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newScene(t)
			red := s.put(0, c.redZone, redirector("Cloud", 9, ability.Keywords{}))
			u := s.put(0, c.unitZone, unit("Ward", 1, 5, ability.Keywords{}))
			if c.moveRedTo != 0 {
				s.moveTo(red, c.moveRedTo)
			}
			s.dealToAlliedUnits(2)
			if c.redirected {
				s.wantDamage(red, 2)
				s.wantDamage(u, 0)
			} else {
				s.wantDamage(red, 0)
				s.wantDamage(u, 2)
			}
			n := len(eventsOfKind(s.g.Log, EventReplace))
			if c.redirected && n != 1 || !c.redirected && n != 0 {
				t.Errorf("replace events = %d, redirected want %v (SVY-R44: the source's current zone)", n, c.redirected)
			}
		})
	}
}

func TestFR_REP_004_RedirectAppliesOnlyToWhatItsFilterMatches(t *testing.T) {
	t.Run("an enemy unit", func(t *testing.T) {
		s := newScene(t)
		red := s.put(0, ZoneMainTheater, redirector("Cloud", 9, ability.Keywords{}))
		foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
		s.enactWith(nil, allDamage(2, ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}))
		s.wantDamage(foe, 2)
		s.wantDamage(red, 0)
		wantNoReplace(t, s)
	})
	t.Run("another allied structure", func(t *testing.T) {
		s := newScene(t)
		red := s.put(0, ZoneMainTheater, redirector("Cloud", 9, ability.Keywords{}))
		wall := s.put(0, ZoneMainTheater, structure("Wall", 0, 9, ability.Keywords{}))
		s.enactWith(nil, allDamage(2, ability.Filter{Kinds: ability.KindStructure}))
		s.wantDamage(wall, 2)
		s.wantDamage(red, 2) // its own damage as a structure, nothing redirected onto it
		wantNoReplace(t, s)
	})
	t.Run("the allied Headquarters", func(t *testing.T) {
		s := newScene(t)
		red := s.put(0, ZoneMainTheater, redirector("Cloud", 9, ability.Keywords{}))
		s.enactWith(nil, allDamage(2, ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindHeadquarters}))
		s.wantDamage(s.g.Players[0].HQ, 2)
		s.wantDamage(red, 0)
		wantNoReplace(t, s)
	})
}

// Covers: FR-REP-003
func TestFR_REP_004_EveryInstanceIsRedirectedSeparately(t *testing.T) {
	s := newScene(t)
	red := s.put(0, ZoneMainTheater, redirector("Cloud", 9, ability.Keywords{}))
	u1 := s.put(0, ZoneMainTheater, unit("Ward", 1, 5, ability.Keywords{}))
	u2 := s.put(0, ZoneMainTheater, unit("Ward", 1, 5, ability.Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	s.enactWith(nil, allDamage(1, ability.Filter{Kinds: ability.KindUnit}))
	s.wantDamage(u1, 0)
	s.wantDamage(u2, 0)
	s.wantDamage(foe, 1)
	s.wantDamage(red, 2) // one instance per allied unit
	if n := len(eventsOfKind(s.g.Log, EventReplace)); n != 2 {
		t.Errorf("replace events = %d, want 2, one per redirected damage instance (FR-REP-004 Stage 1)", n)
	}
}

func TestFR_REP_004_DamageTakenTriggersFireOnTheNewTarget(t *testing.T) {
	s := newScene(t)
	damaged := ability.Trigger{Hook: ability.HookDamaged, Effect: ability.Effect{Steps: []ability.Step{drawN(1)}}}
	s.put(0, ZoneMainTheater, redirector("Cloud", 9, ability.Keywords{}, damaged))
	s.put(0, ZoneMainTheater, striker("Ward", 1, 5, ability.Keywords{}, damaged))
	deck := len(s.g.Players[0].Deck)
	s.dealToAlliedUnits(2)
	if got := deck - len(s.g.Players[0].Deck); got != 1 {
		t.Errorf("player 0 drew %d cards, want exactly 1: only the redirector, which took the damage, fires its damage-taken trigger (FR-REP-004 Stage 5)", got)
	}
}

// Covers: SVY-R23
func TestFR_REP_001_TwoRedirectionsTheAffectedControllerChooses(t *testing.T) {
	setup := func(t *testing.T) (s *scene, a, b, u *Asset) {
		s = newScene(t)
		a = s.put(0, ZoneMainTheater, redirector("Cloud A", 9, ability.Keywords{}))
		b = s.put(0, ZoneMainTheater, redirector("Cloud B", 9, ability.Keywords{}))
		u = s.put(0, ZoneMainTheater, unit("Ward", 1, 5, ability.Keywords{}))
		return
	}
	t.Run("the controller picks the second", func(t *testing.T) {
		s, a, b, u := setup(t)
		s.script(0, Answer{Picks: []int{1}})
		s.dealToAlliedUnits(2)
		asks := replacementAsks(s, 0)
		if len(asks) != 1 {
			t.Fatalf("player 0 was asked %d replacement questions, want 1 (FR-REP-001, SVY-R23)", len(asks))
		}
		wantOffered(t, asks[0], a.ID, b.ID)
		s.wantDamage(b, 2)
		s.wantDamage(a, 0)
		s.wantDamage(u, 0)
		if n := len(eventsOfKind(s.g.Log, EventReplace)); n != 1 {
			t.Errorf("replace events = %d, want 1 (FR-REP-003 single application)", n)
		}
	})
	t.Run("a timeout takes the first", func(t *testing.T) {
		s, a, b, _ := setup(t)
		s.script(0, Answer{Timeout: true})
		s.dealToAlliedUnits(2)
		s.wantDamage(a, 2)
		s.wantDamage(b, 0)
	})
}

// Covers: FR-REP-001, FR-LAY-002
func TestFR_REP_004_KillIsNotRedirected(t *testing.T) {
	s := newScene(t)
	red := s.put(0, ZoneMainTheater, redirector("Cloud", 9, ability.Keywords{}))
	u := s.put(0, ZoneMainTheater, unit("Ward", 1, 5, ability.Keywords{}))
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		if d.Kind == DecideTarget {
			return Answer{Picks: []int{0}}
		}
		return Answer{Timeout: true}
	}
	s.enactAlliedStep(killTarget())
	if s.alive(u) {
		t.Errorf("a kill is not damage, so the allied unit is destroyed (FR-REP-004 Stage 1 redirects damage only)")
	}
	s.wantDamage(red, 0)
	wantNoReplace(t, s)
}
