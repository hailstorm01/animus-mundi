package engine

import (
	"animus-mundi/pkg/engine/ability"
	"testing"
)

// Tests for the real card Hammerhand Bivouac (T-0248, board directive #157): at the
// start of its controller's turn it distributes 2 armor among allied units in its own
// zone (SVY-R51, SVY-R52). The synthetic-card shape tests are in pool_test.go.

// bivouacScene puts the real Hammerhand Bivouac in player 0's Main Theater.
func bivouacScene(t *testing.T) *scene {
	s := newScene(t)
	s.put(0, ZoneMainTheater, s.card("Hammerhand Bivouac"))
	return s
}

func TestCard_HammerhandBivouac_SplitsAmongAlliedMainTheaterUnits(t *testing.T) {
	// Covers: FR-TGT-002
	cases := []struct {
		name         string
		picks        []int
		wantA, wantB int
		invalid      bool
	}{
		{"one each", []int{0, 1}, 1, 1, false},
		{"both on second", []int{1, 1}, 0, 2, false},
		{"invalid total falls back to first", []int{0}, 2, 0, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := bivouacScene(t)
			a := s.put(0, ZoneMainTheater, unit("A", 1, 5, ability.Keywords{}))
			b := s.put(0, ZoneMainTheater, unit("B", 1, 5, ability.Keywords{}))
			from := s.runPool(Answer{Picks: c.picks})
			picks := picksAsked(s)
			if len(picks) != 1 {
				t.Fatalf("Bivouac asks its controller exactly one DecidePick, got %d (SVY-R51)", len(picks))
			}
			d := picks[0]
			if d.Player != 0 || !d.Repeats || d.Min != 2 || d.Max != 2 || len(d.Options) != 2 ||
				d.Options[0].Ref != a.ID || d.Options[1].Ref != b.ID {
				t.Errorf("the pool is one repeated pick of exactly 2 over unit A then unit B (the structure itself is not offered), got %+v (SVY-R51)", d)
			}
			if a.Armor != c.wantA || b.Armor != c.wantB {
				t.Errorf("armor A,B = %d,%d, want %d,%d (SVY-R51)", a.Armor, b.Armor, c.wantA, c.wantB)
			}
			n := len(effectEvents(s.g, from, EventInvalidAnswer))
			if c.invalid && n != 1 {
				t.Errorf("an answer not totalling 2 is invalid: want one invalid-answer event, got %d (SVY-R51)", n)
			}
			if !c.invalid && n != 0 {
				t.Errorf("a valid answer logs no invalid-answer event, got %d (SVY-R51)", n)
			}
		})
	}
}

func TestCard_HammerhandBivouac_OneUnitTakesBothUnasked(t *testing.T) {
	s := bivouacScene(t)
	a := s.put(0, ZoneMainTheater, unit("A", 1, 5, ability.Keywords{}))
	s.runPool()
	if a.Armor != 2 {
		t.Errorf("with one allied unit the whole 2 armor goes on it, got %d (FR-MOD-001 1, SVY-R51)", a.Armor)
	}
	if n := len(picksAsked(s)); n != 0 {
		t.Errorf("with one match nothing is asked, got %d pick decisions (FR-MOD-001 1)", n)
	}
}

func TestCard_HammerhandBivouac_NoAlliedUnitDoesNothing(t *testing.T) {
	s := bivouacScene(t)
	enemy := s.put(1, ZoneMainTheater, unit("Enemy", 1, 5, ability.Keywords{}))
	from := s.runPool()
	if enemy.Armor != 0 {
		t.Errorf("the enemy unit gains no armor, got %d (SVY-R51)", enemy.Armor)
	}
	if n := len(effectEvents(s.g, from, EventArmor)); n != 0 {
		t.Errorf("no armor event is logged when no allied unit is in the zone, got %d (SVY-R51)", n)
	}
	if n := len(picksAsked(s)); n != 0 {
		t.Errorf("nothing is asked when nothing matches, got %d pick decisions (SVY-R51)", n)
	}
}

func TestCard_HammerhandBivouac_TimeoutPutsBothOnEarliestEntered(t *testing.T) {
	s := bivouacScene(t)
	late := s.put(0, ZoneReserves, unit("Late", 1, 5, ability.Keywords{}))
	x := s.put(0, ZoneMainTheater, unit("X", 1, 5, ability.Keywords{}))
	y := s.put(0, ZoneMainTheater, unit("Y", 1, 5, ability.Keywords{}))
	s.moveTo(late, ZoneMainTheater) // enters the Main Theater last
	s.runPool(Answer{Timeout: true})
	if x.Armor != 2 || y.Armor != 0 || late.Armor != 0 {
		t.Errorf("a timed-out Bivouac gives both armor to the allied unit that entered the Main Theater earliest (X): armor X,Y,Late = %d,%d,%d, want 2,0,0 (SVY-R52)",
			x.Armor, y.Armor, late.Armor)
	}
}

func TestCard_HammerhandBivouac_OffersOnlyMainTheaterUnits(t *testing.T) {
	s := bivouacScene(t)
	a := s.put(0, ZoneMainTheater, unit("A", 1, 5, ability.Keywords{}))
	s.put(0, ZoneReserves, unit("Reserve", 1, 5, ability.Keywords{}))
	s.put(0, ZoneHeadquarters, unit("HQUnit", 1, 5, ability.Keywords{}))
	s.put(1, ZoneMainTheater, unit("Enemy", 1, 5, ability.Keywords{}))
	b := s.put(0, ZoneMainTheater, unit("B", 1, 5, ability.Keywords{}))
	s.runPool(Answer{Picks: []int{0, 1}})
	picks := picksAsked(s)
	if len(picks) != 1 {
		t.Fatalf("the pool asks one decision, got %d (SVY-R51)", len(picks))
	}
	opts := picks[0].Options
	if len(opts) != 2 || opts[0].Ref != a.ID || opts[1].Ref != b.ID {
		t.Errorf("the options are exactly the two allied Main Theater units, not Reserves, Headquarters or enemy units: %+v (SVY-R44, SVY-R51)", opts)
	}
}

func TestCard_HammerhandBivouac_ArmorPersistsAndAccumulates(t *testing.T) {
	s := bivouacScene(t)
	a := s.put(0, ZoneMainTheater, unit("A", 1, 5, ability.Keywords{}))
	s.script(0)
	s.script(1)
	s.endTurn() // player 0's turn ends
	// Player 1's turn is now active.
	if a.Armor != 0 {
		t.Errorf("armor must not change at the start of player 1's turn, got %d (FR-TRN-001)", a.Armor)
	}
	s.endTurn() // player 1's turn ends; player 0's turn starts
	if a.Armor != 2 {
		t.Errorf("at the start of player 0's turn the unit gains 2 armor, got %d (FR-TRN-001, SVY-R51)", a.Armor)
	}
	s.endTurn() // player 0's turn ends
	if a.Armor != 2 {
		t.Errorf("armor persists through the end of player 0's turn and into player 1's, got %d (FR-MOD-001)", a.Armor)
	}
	s.endTurn() // player 1's turn ends; player 0's next turn starts
	if a.Armor != 4 {
		t.Errorf("at the start of player 0's following turn armor accumulates to 4, got %d (FR-MOD-001)", a.Armor)
	}
	if n := len(picksAsked(s)); n != 0 {
		t.Errorf("with one allied unit nothing is ever asked, got %d pick decisions (FR-MOD-001 1)", n)
	}
}
