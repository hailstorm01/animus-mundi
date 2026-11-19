package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0265 (FR-TGT-001, FR-TRN-003a, FR-MNV-001, FR-MOD-001, FR-TGT-003;
// SVY-R4, SVY-R29, SVY-R53, SVY-R54): a target slot marked PerKilled: 2 is
// chosen once for every 2 assets the Maneuver's counted kill cost kills, and
// with Different the picks are different entities. The Maneuvers are synthetic:
// "As an additional cost, kill 2 or 4 allied Scratch Company units. For every 2
// units killed this way, deploy a different chosen unit card from your discard
// pile for free." Mutation Swell itself is T-0266. Player 0 is the active
// player and enacts. The killed units' own cards are never relied on.

var pkPileFilter = ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}

// pkManeuver is the Maneuver above; different false leaves the repetitions
// independent.
func pkManeuver(different bool) *Card {
	return killCountManeuverOf(alliedScratchFilter, ability.Effect{
		Targets: []ability.Target{{Filter: pkPileFilter, PerKilled: 2, Different: different}},
		Steps:   []ability.Step{{Op: ability.OpDeployFree, Subject: ability.SubjectTarget, To: ability.ZoneAny}},
	}, 2, 4)
}

// pkCard is a unit card that prints only Reserves, so deploying it asks no zone.
func pkCard(name string) *Card {
	c := unit(name, 1, 1, ability.Keywords{})
	c.DeployZones = []Zone{ZoneReserves}
	return c
}

// pkScene gives player 0 the Maneuver, n allied Scratch Companies in Reserves
// and one unit card in the discard pile for each name (the last name on top).
func pkScene(t *testing.T, different bool, n int, names ...string) (*scene, *Instance, []*Asset, []*Instance, []*Card) {
	t.Helper()
	s := newScene(t)
	var cards []*Card
	for _, name := range names {
		cards = append(cards, pkCard(name))
	}
	pile := s.inDiscard(0, cards...)
	var scratch []*Asset
	for i := 0; i < n; i++ {
		scratch = append(scratch, s.putScratch(0))
	}
	inst := s.give(0, pkManeuver(different))
	return s, inst, scratch, pile, cards
}

// pkTop is the pile's cards, top first.
func pkTop(s *scene) []*Instance {
	d := s.g.Players[0].Discard
	out := make([]*Instance, len(d))
	for i, in := range d {
		out[len(d)-1-i] = in
	}
	return out
}

// pkScript installs seat 0's chooser and records every Decision. A DecideMode
// is answered with "kill <count>" (timed out when count is 0), a DecidePick by
// the options referring to kill (timed out when kill is empty), a DecideTarget
// by target (timed out when nil), anything else is timed out.
func pkScript(s *scene, count int, kill []*Asset, target func(Decision) Answer) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		switch {
		case d.Kind == DecideMode && count > 0:
			label := "kill " + string(rune('0'+count))
			for i, o := range d.Options {
				if o.Label == label {
					return Answer{Picks: []int{i}}
				}
			}
			s.t.Errorf("no option labelled %q among the options of %q", label, d.Prompt)
		case d.Kind == DecidePick && len(kill) > 0:
			var picks []int
			for _, a := range kill {
				for i, o := range d.Options {
					if o.Ref == a.ID {
						picks = append(picks, i)
					}
				}
			}
			return Answer{Picks: picks}
		case d.Kind == DecideTarget && target != nil:
			return target(d)
		}
		return Answer{Timeout: true}
	}
}

// pkPick answers a target decision with the options referring to insts, in order.
func pkPick(insts ...*Instance) func(Decision) Answer {
	return func(d Decision) Answer {
		var picks []int
		for _, in := range insts {
			for i, o := range d.Options {
				if o.Ref == in.ID {
					picks = append(picks, i)
				}
			}
		}
		return Answer{Picks: picks}
	}
}

func pkAsks(s *scene, k DecisionKind) []Decision {
	var out []Decision
	for _, d := range s.asked(0) {
		if d.Kind == k {
			out = append(out, d)
		}
	}
	return out
}

func pkDeployed(s *scene, c *Card) bool { return assetOf(s.g, c) != nil }

func pkAlive(s *scene, as []*Asset) int {
	n := 0
	for _, a := range as {
		if s.alive(a) {
			n++
		}
	}
	return n
}

func TestFR_MNV_001_PerKilledSlotNeedsItsTargets(t *testing.T) {
	// Covers: SVY-R29, SVY-R53
	s, inst, _, _, _ := pkScene(t, true, 2)
	if offersEnact(s.g, inst) {
		t.Errorf("2 Scratch Companies and no unit card in the pile: the count 2 cannot fill its one target, so the Maneuver is not offered (SVY-R29, SVY-R53)")
	}
	s.reject(enactAction(inst))
	if !inHand(s.g.Players[0], inst) {
		t.Errorf("a rejected Enact leaves the Maneuver in hand (SVY-R29)")
	}
	s.inDiscard(0, pkCard("Fallen"))
	if !offersEnact(s.g, inst) {
		t.Errorf("with one unit card in the pile the count 2 can fill its target, so the Maneuver is offered (SVY-R53)")
	}
}

func TestFR_TGT_001_PerKilledCountsNarrowedToFillableTargets(t *testing.T) {
	// Covers: SVY-R53, FR-MOD-001
	s, inst, scratch, _, cards := pkScene(t, true, 4, "Fallen")
	if !offersEnact(s.g, inst) {
		t.Fatalf("count 2 can be paid and filled, so the Maneuver is offered (SVY-R53)")
	}
	pkScript(s, 0, scratch[:2], nil)
	s.do(enactAction(inst))
	if got := pkAsks(s, DecideMode); len(got) != 0 {
		t.Errorf("count 4 needs 2 different cards and the pile has 1, so only count 2 is legal and is taken unasked (SVY-R53, FR-MOD-001 sec. 1): asked %v", got)
	}
	if got := pkAsks(s, DecideTarget); len(got) != 0 {
		t.Errorf("the one card is the only legal entity for the one repetition, taken unasked (FR-MOD-001 sec. 1): asked %v", got)
	}
	if got := pkAlive(s, scratch); got != 2 {
		t.Errorf("exactly 2 Scratch Companies are killed, got %d alive of 4 (SVY-R53)", got)
	}
	if !pkDeployed(s, cards[0]) {
		t.Errorf("the slot is filled with the one unit card and it is deployed (SVY-R53)")
	}
}

func TestFR_TGT_001_PerKilledCountsBothOfferedWithEnoughTargets(t *testing.T) {
	// Covers: SVY-R53
	s, inst, _, _, _ := pkScene(t, true, 4, "Fallen One", "Fallen Two")
	pkScript(s, 0, nil, nil)
	s.do(enactAction(inst))
	modes := pkAsks(s, DecideMode)
	if len(modes) == 0 {
		t.Fatalf("both counts are legal, so the count is asked (SVY-R53)")
	}
	d := modes[0]
	if len(d.Options) != 2 || d.Options[0].Label != "kill 2" || d.Options[1].Label != "kill 4" {
		t.Errorf("4 units and 2 unit cards: the count decision offers \"kill 2\" and \"kill 4\" in printed order, got %v (SVY-R53)", d.Options)
	}
}

func TestFR_TGT_001_PerKilledSlotChoosesOncePerTwoKilled(t *testing.T) {
	// Covers: SVY-R53, SVY-R54
	t.Run("count 2 chooses once", func(t *testing.T) {
		s, inst, scratch, pile, cards := pkScene(t, true, 4, "One", "Two", "Three")
		pkScript(s, 2, scratch[:2], pkPick(pile[1]))
		s.do(enactAction(inst))
		ts := pkAsks(s, DecideTarget)
		if len(ts) != 1 {
			t.Fatalf("count 2 asks the slot once, got %d target decisions (SVY-R53)", len(ts))
		}
		if ts[0].Min != 1 || ts[0].Max != 1 || len(ts[0].Options) != 3 {
			t.Errorf("count 2: Min/Max/options = %d/%d/%d, want 1/1/3 (SVY-R53)", ts[0].Min, ts[0].Max, len(ts[0].Options))
		}
		if !pkDeployed(s, cards[1]) || pkDeployed(s, cards[0]) || pkDeployed(s, cards[2]) {
			t.Errorf("only the answered card is deployed (SVY-R53)")
		}
	})
	t.Run("count 4 chooses twice, different", func(t *testing.T) {
		s, inst, scratch, pile, cards := pkScene(t, true, 4, "One", "Two", "Three")
		pkScript(s, 4, nil, pkPick(pile[0], pile[2]))
		s.do(enactAction(inst))
		ts := pkAsks(s, DecideTarget)
		if len(ts) != 1 {
			t.Fatalf("count 4 asks the slot once for both repetitions, got %d target decisions (SVY-R53)", len(ts))
		}
		d := ts[0]
		if d.Min != 2 || d.Max != 2 || d.Repeats || len(d.Options) != 3 {
			t.Errorf("count 4: Min/Max/Repeats/options = %d/%d/%v/%d, want 2/2/false/3 (SVY-R53, SVY-R54)", d.Min, d.Max, d.Repeats, len(d.Options))
		}
		if pkAlive(s, scratch) != 0 {
			t.Errorf("count 4 kills all four units (SVY-R53)")
		}
		if !pkDeployed(s, cards[0]) || !pkDeployed(s, cards[2]) || pkDeployed(s, cards[1]) {
			t.Errorf("both answered cards are deployed and the third is not (SVY-R53)")
		}
	})
}

func TestFR_TGT_001_PerKilledDifferentRejectsARepeat(t *testing.T) {
	// Covers: SVY-R54, FR-MOD-001
	s, inst, _, _, cards := pkScene(t, true, 4, "One", "Two", "Three")
	top := pkTop(s)
	pkScript(s, 4, nil, func(Decision) Answer { return Answer{Picks: []int{0, 0}} })
	s.do(enactAction(inst))
	if len(eventsOfKind(s.g.Log, EventInvalidAnswer)) != 1 {
		t.Errorf("picking one card twice for a different slot is an invalid answer (SVY-R54, FR-MOD-001 sec. 6)")
	}
	if !pkDeployed(s, cards[2]) || !pkDeployed(s, cards[1]) || pkDeployed(s, cards[0]) {
		t.Errorf("the fallback takes the top two cards of the pile, %q and %q (SVY-R54, FR-MOD-001 sec. 6)", top[0].Card.Name, top[1].Card.Name)
	}
}

func TestFR_MOD_001_PerKilledDifferentTimeout(t *testing.T) {
	// Covers: SVY-R54, SVY-R53
	t.Run("everything timed out", func(t *testing.T) {
		s, inst, _, _, cards := pkScene(t, true, 0, "One", "Two", "Three")
		r1 := s.putScratch(0)
		r2 := s.putScratch(0)
		m1 := s.putScratch(0)
		m2 := s.putScratch(0)
		s.moveTo(m1, ZoneMainTheater)
		s.moveTo(m2, ZoneMainTheater)
		pkScript(s, 0, nil, nil)
		s.do(enactAction(inst))
		if s.alive(r1) || s.alive(r2) || !s.alive(m1) || !s.alive(m2) {
			t.Errorf("a timed-out count is 2 and the first two allied Scratch Companies in fixed order, Reserves first, are killed (SVY-R54)")
		}
		if !pkDeployed(s, cards[2]) || pkDeployed(s, cards[1]) || pkDeployed(s, cards[0]) {
			t.Errorf("a timed-out slot takes the top card of the pile (SVY-R54, FR-MOD-001 sec. 6)")
		}
	})
	t.Run("count answered, target timed out", func(t *testing.T) {
		s, inst, _, _, cards := pkScene(t, true, 4, "One", "Two", "Three")
		pkScript(s, 4, nil, nil)
		s.do(enactAction(inst))
		if !pkDeployed(s, cards[2]) || !pkDeployed(s, cards[1]) || pkDeployed(s, cards[0]) {
			t.Errorf("a timed-out slot of 2 takes the top two different cards of the pile (SVY-R54, FR-MOD-001 sec. 6)")
		}
	})
}

func TestFR_TGT_001_PerKilledWithoutDifferentRepeats(t *testing.T) {
	// Covers: SVY-R4
	t.Run("one card, taken twice unasked", func(t *testing.T) {
		s, inst, scratch, _, cards := pkScene(t, false, 4, "Only")
		pkScript(s, 4, nil, nil)
		s.do(enactAction(inst))
		modes := pkAsks(s, DecideMode)
		if len(modes) != 1 || len(modes[0].Options) != 2 {
			t.Errorf("without Different both counts are legal with one card, so both are offered (SVY-R4, SVY-R53): got %v", modes)
		}
		if got := pkAsks(s, DecideTarget); len(got) != 0 {
			t.Errorf("one legal card for repeated picks is taken without asking (FR-MOD-001 sec. 1, SVY-R4): asked %v", got)
		}
		if pkAlive(s, scratch) != 0 {
			t.Errorf("count 4 kills all four units (SVY-R53)")
		}
		if !pkDeployed(s, cards[0]) {
			t.Errorf("the one card is chosen for the repetitions and deployed (SVY-R4)")
		}
	})
	t.Run("two cards, the decision repeats", func(t *testing.T) {
		s, inst, _, pile, _ := pkScene(t, false, 4, "One", "Two")
		pkScript(s, 4, nil, pkPick(pile[1], pile[1]))
		s.do(enactAction(inst))
		ts := pkAsks(s, DecideTarget)
		if len(ts) != 1 {
			t.Fatalf("count 4 asks the slot once, got %d target decisions (SVY-R4)", len(ts))
		}
		if ts[0].Min != 2 || ts[0].Max != 2 || !ts[0].Repeats {
			t.Errorf("Min/Max/Repeats = %d/%d/%v, want 2/2/true: the repetitions are independent (SVY-R4)", ts[0].Min, ts[0].Max, ts[0].Repeats)
		}
		if n := len(eventsOfKind(s.g.Log, EventInvalidAnswer)); n != 0 {
			t.Errorf("picking the same card twice is a valid answer without Different, got %d invalid answers (SVY-R4)", n)
		}
	})
}

func TestFR_TGT_001_PerKilledSlotOnTriggerChoosesNothing(t *testing.T) {
	// Covers: FR-TGT-003
	s := newScene(t)
	pile := s.inDiscard(0, pkCard("One"), pkCard("Two"))
	s.script(0)
	trig := targetTrigger([]ability.Target{{Filter: pkPileFilter, PerKilled: 2, Different: true}},
		ability.Step{Op: ability.OpDeployFree, Subject: ability.SubjectTarget, To: ability.ZoneAny})
	s.deployFromHand(0, trigUnit("Source", trig))
	for _, d := range s.asked(0) {
		if d.Kind == DecideTarget {
			t.Errorf("a trigger pays no counted kill cost, so a per-kill slot has zero repetitions and asks nothing (FR-TGT-003)")
		}
	}
	if !inDiscard(s.g.Players[0], pile[0]) || !inDiscard(s.g.Players[0], pile[1]) {
		t.Errorf("no repetitions, so no pile card is deployed (FR-TGT-003)")
	}
}
