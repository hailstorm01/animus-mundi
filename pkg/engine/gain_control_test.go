package engine

import (
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0254: OpGainControl gives the ability's controller control of
// each Subject unit in the Main Theater that an opponent controls (FR-ZON-002,
// FR-ZON-003). Cards are synthetic. Player 0 is the active player.

var gainControlStep = ability.Step{Op: ability.OpGainControl, Subject: ability.SubjectTarget, Target: 0}

// gainOpposing has player 0 enact a free Maneuver that gains control of one
// chosen opposing unit in the Main Theater.
func gainOpposing(s *scene) {
	s.t.Helper()
	inst := s.give(0, maneuver("Take Over", 0, 0, ability.Effect{
		Targets: []ability.Target{tgtZone(ability.AllegianceOpposing, ability.KindUnit, false, ability.ZoneMainTheater)},
		Steps:   []ability.Step{gainControlStep},
	}))
	s.do(enactAction(inst))
}

func hasAction(g *Game, want Action) bool {
	for _, a := range g.LegalActions() {
		if a.Kind == want.Kind && a.Player == want.Player && a.Asset == want.Asset && a.Target == want.Target {
			return true
		}
	}
	return false
}

func inAssets(list []*Asset, a *Asset) bool {
	for _, x := range list {
		if x == a {
			return true
		}
	}
	return false
}

func TestFR_ZON_002_GainControlTakesOpposingUnit(t *testing.T) {
	// Covers: FR-ZON-003
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 3, 5, Keywords{}))
	foe.Damage, foe.Armor = 1, 2
	foe.Modifiers = []Modifier{{Layer: 5, Power: 1}}
	gainOpposing(s)
	got := s.g.Asset(foe.ID)
	if got != foe || !inAssets(s.g.Theater, got) {
		t.Fatalf("the unit is not the same asset still in the Main Theater (FR-ZON-002: a control change keeps its zone)")
	}
	if got.Controller != 0 {
		t.Errorf("controller %d, want 0 (the ability's controller gains control)", got.Controller)
	}
	if got.Inst.Owner != 1 {
		t.Errorf("owner %d, want 1 (ownership never changes, FR-ZON-003)", got.Inst.Owner)
	}
	if got.Zone != ZoneMainTheater {
		t.Errorf("zone %v, want Main Theater", got.Zone)
	}
	if got.Exhausted {
		t.Errorf("unit is Exhausted, want Readied (readiness is kept)")
	}
	if got.Damage != 1 || got.Armor != 2 {
		t.Errorf("damage %d, Armor %d, want 1 and 2 (state is kept)", got.Damage, got.Armor)
	}
	if p := s.g.Power(got); p != 4 {
		t.Errorf("power %d, want 4 (the Layer 5 modifier is kept)", p)
	}
}

func TestFR_ZON_002_GainControlKeepsReadiness(t *testing.T) {
	for _, exhausted := range []bool{false, true} {
		s := newScene(t)
		foe := s.put(1, ZoneMainTheater, unit("Foe", 3, 5, Keywords{}))
		foe.Exhausted = exhausted
		gainOpposing(s)
		if foe.Controller != 0 {
			t.Fatalf("exhausted=%v: controller %d, want 0", exhausted, foe.Controller)
		}
		if foe.Exhausted != exhausted {
			t.Errorf("exhausted before %v, after %v: readiness must not change (FR-ZON-002)", exhausted, foe.Exhausted)
		}
	}
}

func TestFR_SET_004_GainedUnitActsAtOnce(t *testing.T) {
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 3, 5, Keywords{}))
	gainOpposing(s)
	if !hasAction(s.g, Action{Kind: ActMarch, Player: 0, Asset: foe.ID}) {
		t.Errorf("player 0 may not March the gained unit; there is no summoning sickness (FR-SET-004)")
	}
	if !hasAction(s.g, Action{Kind: ActAttack, Player: 0, Asset: foe.ID, Target: s.g.Players[1].HQ.ID}) {
		t.Errorf("player 0 may not attack player 1's Headquarters with the gained unit (FR-SET-004)")
	}
}

func TestFR_ZON_002_GainedUnitMarchesToNewControllersReserves(t *testing.T) {
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 3, 5, Keywords{}))
	gainOpposing(s)
	s.do(Action{Kind: ActMarch, Player: 0, Asset: foe.ID})
	if !inAssets(s.g.Players[0].Reserves, foe) {
		t.Errorf("the unit is not in player 0's Reserves (a March goes to the current controller's Reserves, FR-ZON-002)")
	}
	if inAssets(s.g.Players[1].Reserves, foe) {
		t.Errorf("the unit is in player 1's Reserves, want the new controller's only")
	}
	if !foe.Exhausted {
		t.Errorf("the Marched unit is Readied, want Exhausted")
	}
}

func TestFR_ZON_002_GainedUnitDiesToOwnersDiscard(t *testing.T) {
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 3, 5, Keywords{}))
	gainOpposing(s)
	inst := s.give(0, maneuver("Cull", 0, 0, ability.Effect{
		Targets: []ability.Target{tgt(ability.AllegianceAllied, ability.KindUnit, false)},
		Steps:   []ability.Step{{Op: ability.OpKill, Subject: ability.SubjectTarget, Target: 0}},
	}))
	s.do(enactAction(inst))
	if s.alive(foe) {
		t.Fatalf("the gained unit survived the kill")
	}
	in := func(p PlayerID) bool {
		for _, i := range s.g.Players[p].Discard {
			if i == foe.Inst {
				return true
			}
		}
		return false
	}
	if !in(1) || in(0) {
		t.Errorf("card in owner's discard %v, in controller's discard %v; want the owner's only (FR-ZON-002 Destruction)", in(1), in(0))
	}
}

func TestFR_TRN_001_GainedUnitReadiesInNewControllersReadyPhase(t *testing.T) {
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 3, 5, Keywords{}))
	foe.Exhausted = true
	gainOpposing(s)
	s.endTurn()
	if !foe.Exhausted {
		t.Errorf("the unit readied in player 1's Ready Phase; only its controller's assets ready (FR-TRN-001)")
	}
	s.endTurn()
	if s.g.Active != 0 {
		t.Fatalf("precondition: player 0 active, got %d", s.g.Active)
	}
	if foe.Exhausted {
		t.Errorf("the unit stayed Exhausted through its new controller's Ready Phase (FR-TRN-001)")
	}
}

func TestFR_KWD_002e_GainedUnitCountsForNewController(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, unit("Mine", 1, 5, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Foe A", 1, 5, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Foe B", 1, 5, Keywords{}))
	if !s.g.cornered(0) || s.g.cornered(1) {
		t.Fatalf("precondition: player 0 Cornered (1 vs 2), player 1 not")
	}
	s.script(0, Answer{Picks: []int{0}}) // either of player 1's units
	gainOpposing(s)
	if len(s.g.TheaterOf(0)) != 2 || len(s.g.TheaterOf(1)) != 1 {
		t.Fatalf("Main Theater counts %d and %d, want 2 for player 0 and 1 for player 1", len(s.g.TheaterOf(0)), len(s.g.TheaterOf(1)))
	}
	if !s.g.cornered(1) || s.g.cornered(0) {
		t.Errorf("Cornered: player 0 %v, player 1 %v; want player 1 only (FR-KWD-002e)", s.g.cornered(0), s.g.cornered(1))
	}
}

func TestFR_ZON_002_GainControlOnlyMainTheaterUnits(t *testing.T) {
	s := newScene(t)
	src := s.effectSource(ability.Step{Op: ability.OpGainControl, Subject: ability.SubjectAll})
	mt := s.put(1, ZoneMainTheater, unit("Foe", 3, 5, Keywords{}))
	res := s.put(1, ZoneReserves, unit("Reserve Foe", 3, 5, Keywords{}))
	st := s.put(1, ZoneMainTheater, structure("Foe Wall", 0, 5, Keywords{}))
	hq1 := s.g.Players[1].HQ
	from := len(s.g.Log)
	s.do(activate(src, 0))
	if mt.Controller != 0 {
		t.Errorf("the Main Theater unit's controller is %d, want 0", mt.Controller)
	}
	if res.Controller != 1 || !inAssets(s.g.Players[1].Reserves, res) {
		t.Errorf("the Reserves unit changed controller or zone (FR-ZON-002: Main Theater only)")
	}
	if st.Controller != 1 || hq1.Controller != 1 {
		t.Errorf("a Structure or Headquarters changed controller (units only)")
	}
	if src.Controller != 0 {
		t.Errorf("the source's controller is %d, want 0", src.Controller)
	}
	if n := len(effectEvents(s.g, from, EventControl)); n != 1 {
		t.Errorf("%d control events, want exactly 1 (one per unit that changed)", n)
	}
}

func TestEffectEvents_ControlLogsTheChange(t *testing.T) {
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 3, 5, Keywords{}))
	from := len(s.g.Log)
	gainOpposing(s)
	evs := effectEvents(s.g, from, EventControl)
	if len(evs) != 1 {
		t.Fatalf("%d control events, want 1: %+v", len(evs), evs)
	}
	e := evs[0]
	if e.Player != 0 || e.Target != foe.ID || e.Card != "" {
		t.Errorf("player %d, target %d, card %q; want player 0, target %d, no card", e.Player, e.Target, e.Card, foe.ID)
	}
	if want := "player 0 gains control of " + foe.String(); e.Detail != want {
		t.Errorf("detail %q, want %q", e.Detail, want)
	}
}

func TestEffectEvents_ControlFaceDownAssetStaysHidden(t *testing.T) {
	s := newScene(t)
	fd := s.putScratch(1)
	s.moveTo(fd, ZoneMainTheater)
	from := len(s.g.Log)
	gainOpposing(s)
	if fd.Controller != 0 || !fd.FaceDown {
		t.Fatalf("controller %d, face down %v; want 0 and true", fd.Controller, fd.FaceDown)
	}
	evs := effectEvents(s.g, from, EventControl)
	if len(evs) != 1 {
		t.Fatalf("%d control events, want 1", len(evs))
	}
	e := evs[0]
	if e.Card != "" || strings.Contains(e.Detail, "Filler") || !strings.HasSuffix(e.Detail, fd.String()) {
		t.Errorf("the event names the hidden card: card %q, detail %q", e.Card, e.Detail)
	}
}

func TestFR_ZON_002_GainControlStepCountsForIfDone(t *testing.T) {
	for _, c := range []struct {
		name       string
		allegiance ability.Allegiance
		drawn      int
	}{
		{"opposing units", ability.AllegianceOpposing, 1},
		{"allied units only", ability.AllegianceAllied, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			s := newScene(t)
			src := s.put(0, ZoneMainTheater, actUnit("Engine", nil, actAbility(nil,
				ability.Step{Op: ability.OpGainControl, Subject: ability.SubjectAll,
					Filter: ability.Filter{Allegiance: c.allegiance, Kinds: ability.KindUnit}},
				ability.Step{Op: ability.OpDraw, Amount: ability.Value{N: 1}, IfDone: true})))
			s.put(1, ZoneMainTheater, unit("Foe", 3, 5, Keywords{}))
			before := len(s.g.Players[0].Hand)
			s.do(activate(src, 0))
			if got := len(s.g.Players[0].Hand) - before; got != c.drawn {
				t.Errorf("hand grew by %d, want %d (the step counts the units that changed controller)", got, c.drawn)
			}
		})
	}
}
