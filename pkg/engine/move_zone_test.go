package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0119 (FR-ZON-002, FR-ZON-004, FR-TGT-003, FR-MOD-001 sec. 6,
// SVY-R1, SVY-R7, SVY-R16, SVY-R26, board directive #88 test (e)): a step can
// move the units it acts on to their controller's Reserves or the Main
// Theater. A moved unit stays the same asset with its state kept, and a later
// step of the same effect still acts on it. Cards are synthetic Maneuvers
// enacted by player 0 (the active player) at no cost.

// moveStep is "move <subject> to z" (one OpMoveZone step).
func moveStep(subj ability.Subject, slot int, f ability.Filter, z ability.Zone) ability.Step {
	return ability.Step{Op: ability.OpMoveZone, Subject: subj, Target: slot, Filter: f, To: z}
}

var (
	mzAlliedUnits = ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}
	mzAllAssets   = ability.Filter{}
)

// choosesAssets installs a chooser for seat p that answers every decision by
// picking the options whose Ref is one of ids, in option order, and times out
// when none matches.
func (s *scene) choosesAssets(p PlayerID, ids ...int) {
	s.g.choosers[p] = func(_ *Game, d Decision) Answer {
		var picks []int
		for i, o := range d.Options {
			for _, id := range ids {
				if o.Ref == id {
					picks = append(picks, i)
				}
			}
		}
		if len(picks) == 0 {
			return Answer{Timeout: true}
		}
		return Answer{Picks: picks}
	}
}

// enactFree gives player 0 a free Maneuver with effect e and enacts it.
func (s *scene) enactFree(e ability.Effect) {
	s.t.Helper()
	inst := s.give(0, maneuver("Mover", 0, 0, e))
	s.do(enactAction(inst))
}

func TestFR_TGT_003_MovedTargetStillTakesLaterSteps(t *testing.T) {
	// Covers: SVY-R26
	s := newScene(t)
	x := s.put(1, ZoneMainTheater, unit("X", 3, 5, Keywords{}))
	s.choosesAssets(0, x.ID)
	s.enactFree(ability.Effect{
		Targets: []ability.Target{{Filter: ability.Filter{Kinds: ability.KindUnit, Zone: ability.ZoneMainTheater}}},
		Steps: []ability.Step{
			moveStep(ability.SubjectTarget, 0, ability.Filter{}, ability.ZoneReserves),
			modifyTarget(-1, 0),
		},
	})
	if got := s.g.Asset(x.ID); got == nil || got.Zone != ZoneReserves || got.Controller != 1 {
		t.Fatalf("the chosen unit moves to its controller's (the opponent's) Reserves (FR-ZON-002): %v", got)
	}
	if got := s.g.Power(x); got != 2 {
		t.Errorf("the later step still acts on the moved unit although it no longer matches the slot's zone: power %d, want 2 (FR-TGT-003, SVY-R26)", got)
	}
}

func TestFR_ZON_004_MovedUnitKeepsItsState(t *testing.T) {
	s := newScene(t)
	x := s.put(0, ZoneReserves, unit("X", 2, 6, Keywords{}))
	x.Damage, x.Armor, x.Modifiers, x.Exhausted = 2, 3, []Modifier{{Layer: 5, Power: 1}}, true
	power := s.g.Power(x)
	s.choosesAssets(0, x.ID)
	s.enactFree(ability.Effect{
		Targets: []ability.Target{{Filter: mzAlliedUnits}},
		Steps:   []ability.Step{moveStep(ability.SubjectTarget, 0, ability.Filter{}, ability.ZoneMainTheater)},
	})
	got := s.g.Asset(x.ID)
	if got != x {
		t.Fatalf("a moved unit is still the same asset (FR-ZON-004): %v", got)
	}
	if got.Zone != ZoneMainTheater {
		t.Errorf("the unit is in the Main Theater: %v (FR-ZON-002)", got.Zone)
	}
	if got.Damage != 2 || got.Armor != 3 || s.g.Power(got) != power || !got.Exhausted {
		t.Errorf("moving keeps damage 2, Armor 3, power %d and Exhausted: got damage %d, Armor %d, power %d, exhausted %v (FR-ZON-004)",
			power, got.Damage, got.Armor, s.g.Power(got), got.Exhausted)
	}
}

func TestFR_ZON_002_MoveAllAlliedUnits(t *testing.T) {
	s := newScene(t)
	r1 := s.put(0, ZoneReserves, unit("R1", 1, 3, Keywords{}))
	r2 := s.put(0, ZoneReserves, unit("R2", 1, 3, Keywords{}))
	there := s.put(0, ZoneMainTheater, unit("There", 1, 3, Keywords{}))
	opp := s.put(1, ZoneReserves, unit("Opp", 1, 3, Keywords{}))
	s.enactFree(steps(moveStep(ability.SubjectAll, 0, mzAlliedUnits, ability.ZoneMainTheater)))
	for _, a := range []*Asset{r1, r2} {
		if a.Zone != ZoneMainTheater {
			t.Errorf("allied unit %s moves to the Main Theater (FR-ZON-002)", a)
		}
	}
	if there.Zone != ZoneMainTheater || opp.Zone != ZoneReserves {
		t.Errorf("a unit already there and an opposing unit stay: %v, %v (FR-ZON-002)", there.Zone, opp.Zone)
	}
	moves := eventsOfKind(s.g.Log, "move")
	if len(moves) != 2 {
		t.Fatalf("one move event per moved unit: %d events, want 2 (FR-ZON-002)", len(moves))
	}
	seen := map[int]bool{}
	for _, e := range moves {
		seen[e.Target] = true
		if e.Zone != ZoneMainTheater.String() {
			t.Errorf("a move event's Zone is the destination %q, got %q", ZoneMainTheater.String(), e.Zone)
		}
	}
	if !seen[r1.ID] || !seen[r2.ID] {
		t.Errorf("the move events target the two moved units %d and %d: %v", r1.ID, r2.ID, seen)
	}
}

func TestFR_ZON_002_StructuresAndHeadquartersAreNotMoved(t *testing.T) {
	s := newScene(t)
	st := s.put(0, ZoneReserves, structure("Wall", 0, 5, Keywords{}))
	u := s.put(0, ZoneReserves, unit("U", 1, 3, Keywords{}))
	hq0, hq1 := s.g.Players[0].HQ, s.g.Players[1].HQ
	s.enactFree(steps(moveStep(ability.SubjectAll, 0, mzAllAssets, ability.ZoneMainTheater)))
	if st.Zone != ZoneReserves {
		t.Errorf("a Structure is immune to effects that move between battlefield zones: %v (FR-ZON-002)", st.Zone)
	}
	if hq0.Zone != ZoneHeadquarters || hq1.Zone != ZoneHeadquarters {
		t.Errorf("a Headquarters is not moved: %v, %v (FR-ZON-002)", hq0.Zone, hq1.Zone)
	}
	if u.Zone != ZoneMainTheater {
		t.Errorf("a unit in the Reserves does move: %v (FR-ZON-002)", u.Zone)
	}
	moves := eventsOfKind(s.g.Log, "move")
	if len(moves) != 1 || moves[0].Target != u.ID {
		t.Errorf("only the unit has a move event: %v (FR-ZON-002)", moves)
	}
}

func TestFR_ZON_002_MovingIsNotDeploying(t *testing.T) {
	// Covers: SVY-R1
	s := newScene(t)
	c := unit("Deployer", 1, 3, Keywords{})
	c.Ability = &ability.Def{Triggers: []ability.Trigger{{Hook: ability.HookDeploy, Effect: steps(drawN(1))}}}
	u := s.put(0, ZoneReserves, c)
	deck := len(s.g.Players[0].Deck)
	s.enactFree(steps(moveStep(ability.SubjectAll, 0, mzAlliedUnits, ability.ZoneMainTheater)))
	if u.Zone != ZoneMainTheater {
		t.Fatalf("precondition: the unit moved")
	}
	if n := len(eventsOfKind(s.g.Log, "deploy")); n != 0 {
		t.Errorf("moving a unit is not deploying it: %d deploy events (SVY-R1)", n)
	}
	if got := len(s.g.Players[0].Deck); got != deck {
		t.Errorf("its deploy trigger did not fire (no draw): deck %d -> %d (SVY-R1)", deck, got)
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("no trigger was placed: %d stack items (SVY-R1)", len(s.g.Stack))
	}
}

func TestFR_ZON_002_AUnitAlreadyThereDoesNotMove(t *testing.T) {
	s := newScene(t)
	u := s.put(0, ZoneMainTheater, unit("U", 1, 3, Keywords{}))
	deck := len(s.g.Players[0].Deck)
	s.choosesAssets(0, u.ID)
	more := drawN(1)
	more.IfDone = true
	s.enactFree(ability.Effect{
		Targets: []ability.Target{{Filter: mzAlliedUnits}},
		Steps:   []ability.Step{moveStep(ability.SubjectTarget, 0, ability.Filter{}, ability.ZoneMainTheater), more},
	})
	if n := len(eventsOfKind(s.g.Log, "move")); n != 0 {
		t.Errorf("a unit already in the destination does not move: %d move events (FR-ZON-002)", n)
	}
	if got := len(s.g.Players[0].Deck); got != deck {
		t.Errorf("nothing moved, so \"if you do\" does not run: deck %d -> %d (SVY-R14)", deck, got)
	}
}

func TestFR_MOD_001_MovedUnitIsLastInItsNewZone(t *testing.T) {
	// Covers: SVY-R7, SVY-R16
	s := newScene(t)
	y := s.put(0, ZoneMainTheater, unit("Y", 1, 5, Keywords{}))
	x := s.put(0, ZoneReserves, unit("X", 1, 5, Keywords{}))
	z := s.put(0, ZoneMainTheater, unit("Z", 1, 5, Keywords{}))
	s.g.choosers[0] = func(_ *Game, d Decision) Answer { return Answer{Timeout: true} }
	pick := func() ability.Step {
		return ability.Step{Op: ability.OpKill, Subject: ability.SubjectPick,
			Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneMainTheater}}
	}
	s.enactFree(steps(
		moveStep(ability.SubjectAll, 0, ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneReserves}, ability.ZoneMainTheater),
		pick(), pick()))
	if x.Zone != ZoneMainTheater {
		t.Fatalf("precondition: X moved to the Main Theater")
	}
	if s.alive(y) {
		t.Errorf("the first fallback pick is Y, the oldest arrival in the Main Theater (SVY-R7, SVY-R16)")
	}
	if s.alive(z) {
		t.Errorf("the second fallback pick is Z, which arrived before the moved X (SVY-R7, SVY-R16)")
	}
	if !s.alive(x) {
		t.Errorf("the moved X entered the Main Theater last, so it is picked last (SVY-R7, FR-MOD-001 sec. 6)")
	}
}
