package engine

import (
	"reflect"
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for keywords printed behind "Cornered:" (T-0180, SVY-R18). A card lists
// the keywords it has while its controller is Cornered (FR-KWD-002e): the
// opponent controls strictly more Main Theater assets. They are Layer 3
// keyword grants (FR-LAY-001), the gate is locked for a simultaneous step
// (FR-LAY-003), and Aggressive read this way is the keyword trigger placed on
// attack (FR-KWD-002f). Synthetic cards set Def.Cornered through Ability.
//
// This file does not compile until ability.CorneredKeywords and
// ability.Def.Cornered exist.

// corneredCard is a unit that prints the given triggers and, while its
// controller is Cornered, the keywords c.
func corneredCard(name string, power, health int, c ability.CorneredKeywords, trigs ...ability.Trigger) *Card {
	card := unit(name, power, health, Keywords{})
	card.Ability = &ability.Def{Triggers: trigs, Cornered: c}
	return card
}

// aggressiveItems counts the Aggressive keyword items among placed.
func aggressiveItems(items []StackItem) int {
	n := 0
	for _, it := range items {
		if it.Kind == StackAbility && it.Keyword == ability.KeywordAggressive {
			n++
		}
	}
	return n
}

// snapKeywords returns the Snapshot's Keywords for the asset.
func snapKeywords(t *testing.T, s *scene, a *Asset) Keywords {
	t.Helper()
	for _, as := range s.g.Snapshot().Theater {
		if as.ID == a.ID {
			return as.Keywords
		}
	}
	t.Fatalf("asset %d is not in the Snapshot's Theater", a.ID)
	return Keywords{}
}

// damageEvent returns the damage event whose target is a.
func damageEvent(t *testing.T, s *scene, a *Asset) Event {
	t.Helper()
	for _, e := range eventsOfKind(s.g.Log, EventDamage) {
		if e.Target == a.ID {
			return e
		}
	}
	t.Fatalf("no damage event targeting %s", a)
	return Event{}
}

// pickTarget installs a chooser for seat p that answers a DecideTarget with the
// option referring to victim, and a DecideOrder with the Aggressive keyword
// item first (placed first, so it resolves last), then the rest.
func pickTarget(s *scene, p PlayerID, victim *Asset) {
	s.g.choosers[p] = func(_ *Game, d Decision) Answer {
		switch d.Kind {
		case DecideTarget:
			for i, o := range d.Options {
				if o.Ref == victim.ID {
					return Answer{Picks: []int{i}}
				}
			}
		case DecideOrder:
			var picks []int
			for i, o := range d.Options {
				if strings.Contains(o.Label, "Aggressive") {
					picks = append(picks, i)
				}
			}
			for i, o := range d.Options {
				if !strings.Contains(o.Label, "Aggressive") {
					picks = append(picks, i)
				}
			}
			return Answer{Picks: picks}
		}
		s.t.Errorf("unexpected decision %q", d.Prompt)
		return Answer{Timeout: true}
	}
}

// --- Allshan, Chainbreaker Revenant (SVY-R18) ---

func TestCard_Allshan_HasAggressiveOnlyWhileCornered(t *testing.T) {
	// Covers: FR-KWD-002e, FR-LAY-001
	s := newScene(t)
	al := s.put(0, ZoneMainTheater, s.card("Allshan, Chainbreaker Revenant"))
	s.put(1, ZoneMainTheater, unit("Enemy", 0, 5, Keywords{}))
	if got := s.g.Keywords(al).Aggressive; got != 0 {
		t.Errorf("1 against 1 is not Cornered, so Allshan has no Aggressive: got %d (SVY-R18)", got)
	}
	if got := snapKeywords(t, s, al).Aggressive; got != 0 {
		t.Errorf("Snapshot keywords agree: not Cornered, Aggressive 0, got %d", got)
	}
	extra := s.put(1, ZoneMainTheater, unit("Second Enemy", 0, 5, Keywords{}))
	if got := s.g.Keywords(al).Aggressive; got != 3 {
		t.Errorf("Cornered (1 against 2): Allshan has Aggressive 3, got %d (SVY-R18)", got)
	}
	if got := snapKeywords(t, s, al).Aggressive; got != 3 {
		t.Errorf("Snapshot keywords agree: Cornered, Aggressive 3, got %d", got)
	}
	s.moveTo(extra, ZoneReserves)
	if got := s.g.Keywords(al).Aggressive; got != 0 {
		t.Errorf("no longer Cornered once the extra enemy leaves the Main Theater: Aggressive 0, got %d (SVY-R18)", got)
	}
}

func TestCard_Allshan_InReservesHasAggressiveWhileItsControllerIsCornered(t *testing.T) {
	// Covers: FR-KWD-002e
	s := newScene(t)
	al := s.put(0, ZoneReserves, s.card("Allshan, Chainbreaker Revenant"))
	s.put(1, ZoneMainTheater, unit("Enemy", 0, 5, Keywords{}))
	if got := s.g.Keywords(al).Aggressive; got != 3 {
		t.Errorf("Cornered belongs to the controller and counts the Main Theater only (0 against 1): Aggressive 3 in Reserves, got %d", got)
	}
	s.put(0, ZoneMainTheater, unit("Ally", 0, 5, Keywords{}))
	if got := s.g.Keywords(al).Aggressive; got != 0 {
		t.Errorf("1 against 1 is not Cornered: Aggressive 0, got %d", got)
	}
}

func TestCard_Allshan_AggressiveComesFromTheKeyword(t *testing.T) {
	// Covers: FR-KWD-002e, SVY-R18
	t.Run("cornered", func(t *testing.T) {
		s := newScene(t)
		al := s.put(0, ZoneMainTheater, s.card("Allshan, Chainbreaker Revenant"))
		def := s.put(1, ZoneMainTheater, unit("Wall", 0, 20, Keywords{}))
		s.put(1, ZoneMainTheater, unit("Second Wall", 0, 20, Keywords{}))
		items := recordPlaced(s)
		s.do(s.attack(al, def))
		if n := aggressiveItems(*items); n != 1 {
			t.Fatalf("Cornered attack places one Aggressive keyword item: got %d in %+v (FR-KWD-002f)", n, *items)
		}
		for _, it := range *items {
			if it.Keyword == ability.KeywordAggressive && it.Source != al.ID {
				t.Errorf("the Aggressive item's source is Allshan: %+v", it)
			}
		}
		s.wantDamage(def, 8)
	})
	t.Run("not cornered", func(t *testing.T) {
		s := newScene(t)
		al := s.put(0, ZoneMainTheater, s.card("Allshan, Chainbreaker Revenant"))
		def := s.put(1, ZoneMainTheater, unit("Wall", 0, 20, Keywords{}))
		items := recordPlaced(s)
		s.do(s.attack(al, def))
		if n := aggressiveItems(*items); n != 0 {
			t.Errorf("not Cornered, no Aggressive item: got %d in %+v", n, *items)
		}
		s.wantDamage(def, 5)
	})
}

// --- FR-KWD-002f: the label ---

func TestFR_KWD_002f_CorneredAggressiveIsLabelledTheKeyword(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Timeout: true})
	att := s.put(0, ZoneMainTheater, corneredCard("Raider", 1, 9, ability.CorneredKeywords{Aggressive: 3}, attackTrigger(scratchN(1))))
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 20, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Second Dummy", 0, 20, Keywords{}))
	s.do(s.attack(att, def))
	d := s.asked(0)
	if len(d) != 1 || d[0].Kind != DecideOrder || len(d[0].Options) != 2 {
		t.Fatalf("want one DecideOrder with two options (printed trigger and Aggressive): %+v", d)
	}
	found := false
	for _, o := range d[0].Options {
		if o.Label == "Raider: Aggressive 3" {
			found = true
		}
	}
	if !found {
		t.Errorf("one option is labelled %q (FR-KWD-002f): %+v", "Raider: Aggressive 3", d[0].Options)
	}
}

// --- FR-LAY-003: checked when the attack happens ---

func TestFR_LAY_003_CorneredAggressiveIsCheckedWhenTheAttackHappens(t *testing.T) {
	// Covers: FR-KWD-002e, FR-KWD-002f, FR-CBT-002
	kill := func(opposing bool) ability.Trigger {
		a := ability.AllegianceAllied
		if opposing {
			a = ability.AllegianceOpposing
		}
		return targetedAttack([]ability.Target{tgt(a, ability.KindUnit, true)}, dmgTo(99, 0))
	}
	t.Run("cornered at declaration then evened by the printed trigger", func(t *testing.T) {
		s := newScene(t)
		att := s.put(0, ZoneMainTheater, corneredCard("Raider", 2, 9, ability.CorneredKeywords{Aggressive: 3}, kill(true)))
		def := s.put(1, ZoneMainTheater, unit("Wall", 0, 20, Keywords{}))
		victim := s.put(1, ZoneMainTheater, unit("Victim", 0, 1, Keywords{}))
		pickTarget(s, 0, victim)
		items := recordPlaced(s)
		s.do(s.attack(att, def))
		if s.alive(victim) {
			t.Fatalf("precondition: the printed trigger resolves first and kills the victim")
		}
		if n := aggressiveItems(*items); n != 1 {
			t.Errorf("Cornered when the attack happened: one Aggressive item placed, got %d", n)
		}
		s.wantDamage(def, 5) // power 2 + Aggressive 3; the item was already placed (FR-LAY-003)
	})
	t.Run("not cornered at declaration then cornered by Step 3", func(t *testing.T) {
		s := newScene(t)
		att := s.put(0, ZoneMainTheater, corneredCard("Raider", 2, 9, ability.CorneredKeywords{Aggressive: 3}, kill(false)))
		ally := s.put(0, ZoneMainTheater, unit("Ally", 0, 1, Keywords{}))
		s.put(0, ZoneMainTheater, unit("Second Ally", 0, 5, Keywords{}))
		def := s.put(1, ZoneMainTheater, unit("Wall", 0, 20, Keywords{}))
		s.put(1, ZoneMainTheater, unit("Second Wall", 0, 20, Keywords{}))
		s.put(1, ZoneMainTheater, unit("Third Wall", 0, 20, Keywords{}))
		pickTarget(s, 0, ally)
		items := recordPlaced(s)
		s.do(s.attack(att, def))
		if s.alive(ally) {
			t.Fatalf("precondition: the printed trigger kills the allied asset")
		}
		if n := aggressiveItems(*items); n != 0 {
			t.Errorf("not Cornered when the attack happened: no Aggressive item, got %d", n)
		}
		s.wantDamage(def, 2) // printed power only
	})
}

// --- FR-REP-004: Resistant behind Cornered ---

func TestFR_KWD_002e_CorneredResistantReducesDamageOnlyWhileCornered(t *testing.T) {
	// Covers: FR-REP-004, FR-LAY-001
	setup := func(t *testing.T, cornered bool, attacker *Card) (s *scene, u, wall, a *Asset) {
		s = newScene(t)
		u = s.put(1, ZoneMainTheater, corneredCard("Tough", 0, 20, ability.CorneredKeywords{Resistant: 1}))
		wall = s.put(1, ZoneMainTheater, unit("Wall", 0, 20, Keywords{}))
		a = s.put(0, ZoneMainTheater, attacker)
		if cornered {
			s.put(0, ZoneMainTheater, unit("Ally", 0, 5, Keywords{}))
			s.put(0, ZoneMainTheater, unit("Second Ally", 0, 5, Keywords{}))
		}
		return s, u, wall, a
	}
	for _, tc := range []struct {
		name      string
		cornered  bool
		wantDmg   int
		wantResis int
	}{{"cornered", true, 2, 1}, {"not cornered", false, 3, 0}} {
		t.Run("combat hit "+tc.name, func(t *testing.T) {
			s, u, _, a := setup(t, tc.cornered, unit("Hitter", 3, 9, Keywords{}))
			s.do(s.attack(a, u))
			s.wantDamage(u, tc.wantDmg)
			if e := damageEvent(t, s, u); e.Resistant != tc.wantResis {
				t.Errorf("the damage event reports Resistant %d, got %d", tc.wantResis, e.Resistant)
			}
		})
		t.Run("effect damage "+tc.name, func(t *testing.T) {
			zap := striker("Zapper", 0, 9, Keywords{}, targetedAttack([]ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, true)}, dmgTo(3, 0)))
			s, u, wall, a := setup(t, tc.cornered, zap)
			pickTarget(s, 0, u)
			s.do(s.attack(a, wall))
			s.wantDamage(u, tc.wantDmg)
			if e := damageEvent(t, s, u); e.Resistant != tc.wantResis {
				t.Errorf("the damage event reports Resistant %d, got %d", tc.wantResis, e.Resistant)
			}
		})
	}
}

// --- FR-LAY-003: the gate is locked for a simultaneous step ---

func TestFR_LAY_003_SimultaneousDamageLocksCorneredKeywords(t *testing.T) {
	// Covers: FR-REP-004, FR-KWD-002e, FR-CBT-002
	t.Run("combat", func(t *testing.T) {
		s := newScene(t)
		u := s.put(0, ZoneMainTheater, corneredCard("Tough", 5, 20, ability.CorneredKeywords{Resistant: 1}))
		def := s.put(1, ZoneMainTheater, unit("Brittle", 4, 1, Keywords{}))
		s.put(1, ZoneMainTheater, unit("Wall", 0, 20, Keywords{}))
		s.do(s.attack(u, def))
		if s.alive(def) {
			t.Fatalf("precondition: the defender is destroyed, which evens the counts")
		}
		s.wantDamage(u, 3) // retaliation 4 less Resistant 1 (FR-LAY-003 lock)
		if got := s.g.Keywords(u).Resistant; got != 0 {
			t.Errorf("after the sweep the counts are 1 against 1: Resistant 0, got %d", got)
		}
	})
	t.Run("area damage", func(t *testing.T) {
		s := newScene(t)
		weak := s.put(1, ZoneMainTheater, unit("Weak", 0, 3, Keywords{}))
		u := s.put(0, ZoneMainTheater, corneredCard("Tough", 0, 20, ability.CorneredKeywords{Resistant: 1}))
		src := s.put(0, ZoneMainTheater, striker("Sweeper", 0, 20, Keywords{},
			attackTrigger(allDamage(3, ability.Filter{Kinds: ability.KindUnit}))))
		def := s.put(1, ZoneMainTheater, unit("Wall", 0, 20, Keywords{}))
		s.put(1, ZoneMainTheater, unit("Second Wall", 0, 20, Keywords{}))
		s.do(s.attack(src, def))
		if s.alive(weak) {
			t.Fatalf("precondition: the lethal area damage destroys the enemy listed before the unit")
		}
		s.wantDamage(u, 2) // 3 less Resistant 1, though the enemy's death comes first in fixed order
		if got := s.g.Keywords(u).Resistant; got != 0 {
			t.Errorf("after the sweep the counts are 2 against 2: Resistant 0, got %d", got)
		}
	})
}

// --- FR-ZON-005: face-down ---

func TestFR_ZON_005_FaceDownAssetHasNoCorneredKeyword(t *testing.T) {
	// Covers: FR-KWD-002e, FR-REP-004, FR-KWD-002f
	t.Run("attack", func(t *testing.T) {
		s := newScene(t)
		c := corneredCard("Hidden", 1, 1, ability.CorneredKeywords{Aggressive: 3, Resistant: 1}, attackTrigger(scratchN(1)))
		f := &Asset{ID: s.g.newID(), Inst: &Instance{ID: s.g.newID(), Card: c, Owner: 0}, Controller: 0, FaceDown: true}
		s.g.placeOnBattlefield(f, ZoneMainTheater)
		def := s.put(1, ZoneMainTheater, unit("Wall", 0, 20, Keywords{}))
		s.put(1, ZoneMainTheater, unit("Second Wall", 0, 20, Keywords{}))
		if got, want := s.g.Keywords(f), f.Printed().Keywords; !reflect.DeepEqual(got, want) {
			t.Errorf("a face-down asset has the Scratch Company face's keywords %+v, got %+v (FR-ZON-005)", want, got)
		}
		items := recordPlaced(s)
		s.do(s.attack(f, def))
		if n := aggressiveItems(*items); n != 0 {
			t.Errorf("face-down, no Aggressive item: got %d", n)
		}
	})
	t.Run("hit", func(t *testing.T) {
		s := newScene(t)
		c := corneredCard("Hidden", 1, 1, ability.CorneredKeywords{Aggressive: 3, Resistant: 1})
		f := &Asset{ID: s.g.newID(), Inst: &Instance{ID: s.g.newID(), Card: c, Owner: 1}, Controller: 1, FaceDown: true}
		s.g.placeOnBattlefield(f, ZoneMainTheater)
		a := s.put(0, ZoneMainTheater, unit("Hitter", 1, 9, Keywords{}))
		s.put(0, ZoneMainTheater, unit("Ally", 0, 5, Keywords{}))
		s.do(s.attack(a, f))
		if e := damageEvent(t, s, f); e.Resistant != 0 || e.Amount != 1 {
			t.Errorf("a face-down asset has no Resistant, so the hit of 1 is not reduced: Resistant %d, committed %d", e.Resistant, e.Amount)
		}
	})
}
