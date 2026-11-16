package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for an untargeted distributed pool (T-0247, board directive #157): a
// SubjectPick step with Split makes its Amount a pool its controller splits among the
// matching assets as it resolves (SVY-R51), and a timeout puts it all on the first
// in fixed order (SVY-R52).

// poolStructure is a structure whose start-of-your-turn trigger distributes 2 armor
// among allied units in its own zone, as Hammerhand Bivouac's text reads.
func poolStructure() *Card {
	return startStructure("Pool Bivouac", startSteps(ability.TurnYours, ability.Step{
		Op: ability.OpAddArmor, Amount: ability.Value{N: 2}, Subject: ability.SubjectPick, Split: true,
		Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, ZoneRef: ability.ZoneRefSame},
	}))
}

// poolScene puts the pool structure in player 0's Main Theater.
func poolScene(t *testing.T) *scene {
	s := newScene(t)
	s.put(0, ZoneMainTheater, poolStructure())
	return s
}

// runPool scripts player 0's answers, leaves player 1 unscripted-but-recording
// (it must never be asked), and runs to the start of player 0's next turn.
func (s *scene) runPool(answers ...Answer) (from int) {
	s.script(0, answers...)
	s.script(1)
	from = len(s.g.Log)
	s.endTurn()
	s.endTurn()
	return from
}

func picksAsked(s *scene) []Decision {
	var out []Decision
	for _, p := range []PlayerID{0, 1} {
		for _, d := range s.asked(p) {
			if d.Kind == DecidePick {
				out = append(out, d)
			}
		}
	}
	return out
}

func TestFR_TGT_002_UntargetedPoolSplitAsItResolves(t *testing.T) {
	// Covers: FR-MOD-001
	cases := []struct {
		name         string
		picks        []int
		wantA, wantB int
		events       []int // expected event amounts, A's first
	}{
		{"both on first", []int{0, 0}, 2, 0, []int{2}},
		{"one each", []int{0, 1}, 1, 1, []int{1, 1}},
		{"both on second", []int{1, 1}, 0, 2, []int{2}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := poolScene(t)
			a := s.put(0, ZoneMainTheater, unit("A", 1, 5, Keywords{}))
			b := s.put(0, ZoneMainTheater, unit("B", 1, 5, Keywords{}))
			from := s.runPool(Answer{Picks: c.picks})
			picks := picksAsked(s)
			if len(picks) != 1 {
				t.Fatalf("the pool asks its controller exactly one DecidePick, got %d decisions (SVY-R51)", len(picks))
			}
			d := picks[0]
			if d.Player != 0 || !d.Repeats || d.Min != 2 || d.Max != 2 || len(d.Options) != 2 ||
				d.Options[0].Ref != a.ID || d.Options[1].Ref != b.ID {
				t.Errorf("the pool is one repeated pick of exactly 2 (one per point) over A then B, got %+v (SVY-R51)", d)
			}
			for _, d := range s.asked(0) {
				if d.Kind == DecideTarget {
					t.Errorf("an untargeted pool is not a target choice (SVY-R51, FR-TGT-001)")
				}
			}
			if a.Armor != c.wantA || b.Armor != c.wantB {
				t.Errorf("armor A,B = %d,%d, want %d,%d: each unit gets the points picked for it (SVY-R51)", a.Armor, b.Armor, c.wantA, c.wantB)
			}
			var got []int
			for _, e := range effectEvents(s.g, from, EventArmor) {
				got = append(got, e.Amount)
			}
			if len(got) != len(c.events) {
				t.Fatalf("armor events amounts = %v, want one event per receiving unit with its whole share %v (SVY-R51)", got, c.events)
			}
			for i := range got {
				if got[i] != c.events[i] {
					t.Errorf("armor event amounts = %v, want %v (SVY-R51)", got, c.events)
				}
			}
		})
	}
}

func TestFR_TGT_002_UntargetedPoolRejectsWrongTotals(t *testing.T) {
	cases := []struct {
		name  string
		picks []int
	}{
		{"total of one", []int{0}},
		{"total of three", []int{0, 0, 1}},
		{"index out of range", []int{0, 2}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := poolScene(t)
			a := s.put(0, ZoneMainTheater, unit("A", 1, 5, Keywords{}))
			b := s.put(0, ZoneMainTheater, unit("B", 1, 5, Keywords{}))
			from := s.runPool(Answer{Picks: c.picks})
			if n := len(effectEvents(s.g, from, EventInvalidAnswer)); n != 1 {
				t.Errorf("an answer whose picks do not total exactly the pool is invalid, so one invalid-answer event is logged, got %d (SVY-R51, FR-MOD-001 6)", n)
			}
			if a.Armor != 2 || b.Armor != 0 {
				t.Errorf("an invalid answer falls back as a timeout: the whole pool on A, the first in fixed order; armor A,B = %d,%d, want 2,0 (SVY-R52)", a.Armor, b.Armor)
			}
		})
	}
}

func TestFR_MOD_001_UntargetedPoolOneMatchTakesAll(t *testing.T) {
	s := poolScene(t)
	a := s.put(0, ZoneMainTheater, unit("A", 1, 5, Keywords{}))
	s.runPool()
	if a.Armor != 2 {
		t.Errorf("with one match the pool's whole 2 armor goes on it, got %d (FR-MOD-001 1, SVY-R51)", a.Armor)
	}
	if n := len(picksAsked(s)); n != 0 {
		t.Errorf("with one match nothing is asked, got %d pick decisions (FR-MOD-001 1)", n)
	}
}

func TestFR_TGT_002_UntargetedPoolNoMatchDoesNothing(t *testing.T) {
	s := poolScene(t)
	enemy := s.put(1, ZoneMainTheater, unit("Enemy", 1, 5, Keywords{}))
	reserve := s.put(0, ZoneReserves, unit("Reserve", 1, 5, Keywords{}))
	from := s.runPool()
	if enemy.Armor != 0 || reserve.Armor != 0 {
		t.Errorf("with no allied unit in the structure's zone the clause does nothing; armor enemy,reserve = %d,%d (SVY-R51)", enemy.Armor, reserve.Armor)
	}
	if n := len(effectEvents(s.g, from, EventArmor)); n != 0 {
		t.Errorf("no armor event is logged when nothing matches, got %d (SVY-R51)", n)
	}
	if n := len(picksAsked(s)); n != 0 {
		t.Errorf("nothing is asked when nothing matches, got %d pick decisions (SVY-R51)", n)
	}
}

func TestFR_MOD_001_UntargetedPoolTimeoutPutsAllOnFirstInFixedOrder(t *testing.T) {
	// Covers: FR-TGT-002
	s := poolScene(t)
	late := s.put(0, ZoneReserves, unit("Late", 1, 5, Keywords{})) // lowest ID, listed first in Options
	x := s.put(0, ZoneMainTheater, unit("X", 1, 5, Keywords{}))
	y := s.put(0, ZoneMainTheater, unit("Y", 1, 5, Keywords{}))
	s.moveTo(late, ZoneMainTheater) // enters the Main Theater last
	s.runPool(Answer{Timeout: true})
	if x.Armor != 2 || y.Armor != 0 || late.Armor != 0 {
		t.Errorf("a timeout puts the whole pool on the first unit in fixed order, the one that entered the Main Theater earliest (X): armor X,Y,Late = %d,%d,%d, want 2,0,0 (SVY-R52, FR-MOD-001 6)",
			x.Armor, y.Armor, late.Armor)
	}
}

func TestFR_TGT_002_UntargetedPoolOffersOnlyLockedZone(t *testing.T) {
	// Covers: FR-MOD-001
	s := poolScene(t)
	a := s.put(0, ZoneMainTheater, unit("A", 1, 5, Keywords{}))
	s.put(0, ZoneReserves, unit("Reserve", 1, 5, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Enemy", 1, 5, Keywords{}))
	b := s.put(0, ZoneMainTheater, unit("B", 1, 5, Keywords{}))
	s.runPool(Answer{Picks: []int{0, 1}})
	picks := picksAsked(s)
	if len(picks) != 1 {
		t.Fatalf("the pool asks one decision, got %d (SVY-R51)", len(picks))
	}
	opts := picks[0].Options
	if len(opts) != 2 || opts[0].Ref != a.ID || opts[1].Ref != b.ID {
		t.Errorf("the options are exactly the two allied units in the structure's own zone, not the Reserves unit, the enemy or a Headquarters: %+v (SVY-R44, SVY-R51)", opts)
	}
}
