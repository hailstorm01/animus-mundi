package engine

import (
	"reflect"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0138 (FR-TGT-001, FR-MOD-001, FR-MNV-001, FR-TRN-003a, FR-KWD-002h;
// SVY-R3, SVY-R4, SVY-R11, SVY-R29, SVY-R32): a target slot marked PerDiscard
// is chosen once for each card the Maneuver's discard cost discarded, each
// choice independent, so the same asset may be chosen more than once. The
// Maneuvers are synthetic: "As an additional cost, discard up to 3 cards. For
// each card discarded this way, deal 2 damage to chosen asset."

// perDiscardManeuver is the Maneuver above, its slot filtered by f.
func perDiscardManeuver(f ability.Filter) *Card {
	return discardCostManeuver("Steel Rain", 0, 0, 3, ability.Effect{
		Targets: []ability.Target{{Filter: f, PerDiscard: true}},
		Steps:   []ability.Step{dmgTo(2, 0)},
	})
}

// perDiscardScene gives player 0 the Maneuver (slot filtered by f) and n
// other cards, and returns them in hand order.
func perDiscardScene(t *testing.T, f ability.Filter, n int) (*scene, *Instance, []*Instance) {
	s := newScene(t)
	s.g.Players[0].Hand = nil
	inst := s.give(0, perDiscardManeuver(f))
	var others []*Instance
	for i := 0; i < n; i++ {
		others = append(others, s.give(0, unit(string(rune('A'+i)), 1, 1, Keywords{})))
	}
	return s, inst, others
}

// pdScript installs seat 0's chooser and records every Decision. A DecideHand
// is answered with the leading discards options; a DecideTarget with the
// options referring to targets in order, or a timeout when timeoutTargets.
func pdScript(s *scene, discards int, timeoutTargets bool, targets ...*Asset) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		switch d.Kind {
		case DecideHand:
			picks := []int{}
			for i := 0; i < discards && i < len(d.Options); i++ {
				picks = append(picks, i)
			}
			return Answer{Picks: picks}
		case DecideTarget:
			if timeoutTargets {
				return Answer{Timeout: true}
			}
			picks := []int{}
			for _, a := range targets {
				found := -1
				for i, o := range d.Options {
					if o.Ref == a.ID {
						found = i
					}
				}
				if found < 0 {
					s.t.Errorf("no option referring to asset %d among the options of %q", a.ID, d.Prompt)
				}
				picks = append(picks, found)
			}
			return Answer{Picks: picks}
		}
		return Answer{Timeout: true}
	}
}

// maneuverDamage lists the Amount and Target of each damage event in the log.
func maneuverDamage(s *scene) (amounts, targets []int) {
	for _, e := range s.g.Log {
		if e.Kind == EventDamage {
			amounts = append(amounts, e.Amount)
			targets = append(targets, e.Target)
		}
	}
	return
}

func TestFR_TGT_001_PerDiscardSlotChoosesOncePerDiscard(t *testing.T) {
	// Covers: SVY-R3, SVY-R4
	for _, n := range []int{0, 1, 3} {
		s, inst, _ := perDiscardScene(t, ability.Filter{}, 3)
		u1 := s.put(1, ZoneMainTheater, unit("Victim One", 1, 9, Keywords{}))
		u2 := s.put(1, ZoneMainTheater, unit("Victim Two", 1, 9, Keywords{}))
		targets := make([]*Asset, n)
		for i := range targets {
			targets[i] = u1
		}
		pdScript(s, n, false, targets...)
		s.do(enactAction(inst))
		asks := s.asked(0)
		if n == 0 {
			if len(asks) != 1 || asks[0].Kind != DecideHand {
				t.Errorf("0 discarded: only the discard is asked, no target (SVY-R4), got %v", kindsOf(asks))
			}
			if got := totalDamage(s.g); got != 0 {
				t.Errorf("0 discarded: no damage is dealt, got %d (SVY-R4)", got)
			}
			continue
		}
		if len(asks) != 2 || asks[0].Kind != DecideHand || asks[1].Kind != DecideTarget {
			t.Errorf("%d discarded: want [Hand Target] (the discard is paid first, SVY-R3), got %v", n, kindsOf(asks))
			continue
		}
		d := asks[1]
		if d.Min != n || d.Max != n || !d.Repeats {
			t.Errorf("%d discarded: target decision Min/Max/Repeats = %d/%d/%v, want %d/%d/true (SVY-R4)", n, d.Min, d.Max, d.Repeats, n, n)
		}
		if got := u1.Damage; got != 2*n {
			t.Errorf("%d discarded: the chosen unit takes %d, got %d (SVY-R4)", n, 2*n, got)
		}
		s.wantDamage(u2, 0)
	}
}

func TestFR_TGT_001_PerDiscardSameAssetEachTime(t *testing.T) {
	// Covers: SVY-R4, FR-REP-004
	s, inst, _ := perDiscardScene(t, ability.Filter{}, 3)
	u1 := s.put(1, ZoneMainTheater, unit("Victim One", 1, 9, Keywords{}))
	u2 := s.put(1, ZoneMainTheater, unit("Victim Two", 1, 9, Keywords{}))
	pdScript(s, 3, false, u1, u1, u1)
	s.do(enactAction(inst))
	s.wantDamage(u1, 6)
	s.wantDamage(u2, 0)
	amounts, targets := maneuverDamage(s)
	if !reflect.DeepEqual(amounts, []int{2, 2, 2}) || !reflect.DeepEqual(targets, []int{u1.ID, u1.ID, u1.ID}) {
		t.Errorf("each repetition is its own 2-damage instance on the chosen unit (FR-REP-004): got amounts %v on %v", amounts, targets)
	}
}

func TestFR_TGT_001_PerDiscardSplitAcrossAssets(t *testing.T) {
	// Covers: SVY-R4
	s, inst, _ := perDiscardScene(t, ability.Filter{}, 3)
	a := s.put(1, ZoneMainTheater, unit("Victim A", 1, 9, Keywords{}))
	b := s.put(1, ZoneMainTheater, unit("Victim B", 1, 9, Keywords{}))
	pdScript(s, 3, false, a, b, a)
	s.do(enactAction(inst))
	s.wantDamage(a, 4)
	s.wantDamage(b, 2)
	_, targets := maneuverDamage(s)
	if !reflect.DeepEqual(targets, []int{a.ID, b.ID, a.ID}) {
		t.Errorf("damage events follow pick order A, B, A (SVY-R4): got %v", targets)
	}
}

func TestFR_MOD_001_PerDiscardTimeoutConcentrates(t *testing.T) {
	// Covers: SVY-R32
	s, inst, _ := perDiscardScene(t, ability.Filter{}, 3)
	u1 := s.put(1, ZoneMainTheater, unit("Victim One", 1, 9, Keywords{}))
	u2 := s.put(1, ZoneMainTheater, unit("Victim Two", 1, 9, Keywords{}))
	pdScript(s, 3, true)
	s.do(enactAction(inst))
	first := s.g.Players[0].HQ
	if first.Damage != 6 {
		t.Errorf("a timed-out per-discard choice puts every repetition on the first legal entity in fixed order, the choosing player's Headquarters: got %d damage there, want 6 (FR-MOD-001 sec. 6, SVY-R32)", first.Damage)
	}
	s.wantDamage(u1, 0)
	s.wantDamage(u2, 0)
	s.wantDamage(s.g.Players[1].HQ, 0)
}

func TestFR_KWD_002h_PerDiscardSlotNeverOffersOpposingWard(t *testing.T) {
	s, inst, _ := perDiscardScene(t, ability.Filter{Kinds: ability.KindUnit}, 2)
	oppWard := s.put(1, ZoneMainTheater, unit("Warded Foe", 1, 9, Keywords{Ward: true}))
	oppPlain := s.put(1, ZoneMainTheater, unit("Plain Foe", 1, 9, Keywords{}))
	ownWard := s.put(0, ZoneMainTheater, unit("Warded Ally", 1, 9, Keywords{Ward: true}))
	pdScript(s, 2, false, oppPlain, ownWard)
	s.do(enactAction(inst))
	asks := s.asked(0)
	if len(asks) != 2 || asks[1].Kind != DecideTarget {
		t.Fatalf("want [Hand Target], got %v", kindsOf(asks))
	}
	got := map[int]bool{}
	for _, o := range asks[1].Options {
		got[o.Ref] = true
	}
	if got[oppWard.ID] {
		t.Errorf("an opponent's Ward unit is never a legal target (FR-KWD-002h)")
	}
	if !got[ownWard.ID] || !got[oppPlain.ID] {
		t.Errorf("the controller's own Ward unit and a plain opposing unit are options (FR-KWD-002h): options %v", refs(asks[1]))
	}
}

func TestFR_MNV_001_PerDiscardSlotNeverBlocksEnact(t *testing.T) {
	// The lead's reading of SVY-R29: a per-discard slot is not an exactly-one slot.
	s, inst, o := perDiscardScene(t, ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}, 2)
	if !offersEnact(s.g, inst) {
		t.Fatalf("the Maneuver is a legal action though no opposing unit exists: discarding none needs no target (SVY-R29)")
	}
	pdScript(s, 2, false)
	s.do(enactAction(inst))
	for _, d := range s.asked(0) {
		if d.Kind == DecideTarget {
			t.Errorf("no legal entity: the slot stays empty and no target is asked (FR-TGT-003)")
		}
	}
	p := s.g.Players[0]
	if !inDiscard(p, o[0]) || !inDiscard(p, o[1]) {
		t.Errorf("both picked cards end in the discard pile (FR-TRN-003a)")
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("no damage is dealt with an empty slot, got %d", got)
	}
}

func TestFR_TGT_001_PerDiscardSlotOnTriggerChoosesNothing(t *testing.T) {
	// Covers: FR-TRN-003a (only an enacted card pays additional costs)
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("Victim", 1, 9, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Victim Two", 1, 9, Keywords{}))
	s.script(0)
	trig := targetTrigger([]ability.Target{{PerDiscard: true}}, dmgTo(2, 0))
	s.deployFromHand(0, trigUnit("Source", trig))
	for _, d := range s.asked(0) {
		if d.Kind == DecideTarget {
			t.Errorf("a trigger pays no discard cost, so a per-discard slot has zero repetitions and asks nothing (FR-TRN-003a)")
		}
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("no repetitions, no damage: got %d (FR-TRN-003a)", got)
	}
}
