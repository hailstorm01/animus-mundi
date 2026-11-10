package engine

import (
	"slices"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0208 (FR-TRN-003a, wave F): OpDeployFree puts a unit card chosen
// from a discard pile onto the battlefield for free during resolution, with no
// stack item and no costs, into its printed zone, a zone its controller chooses
// when it prints both Reserves and the Main Theater (SVY-R43), or the zone the
// step names. Cards are synthetic. Player 0 is the active player and enacts a
// Maneuver whose one pile target is the unit card.

// dfManeuver is a Maneuver costing req and tac whose one pile target is a unit
// card in the controller's discard pile and whose one step is OpDeployFree
// into to (ZoneAny: the card's printed zones).
func dfManeuver(req, tac int, to ability.Zone) *Card {
	return maneuver("Free Deploy", req, tac, ability.Effect{
		Targets: []ability.Target{{Filter: ability.Filter{
			Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}}},
		Steps: []ability.Step{{Op: ability.OpDeployFree, Subject: ability.SubjectTarget, To: to}},
	})
}

// dfScene puts card c into player 0's discard pile and enacts a Maneuver that
// free-deploys it into to. Seat 0 is scripted with answers, one per decision
// actually asked: a single legal pile target is chosen without a question
// (FR-MOD-001 1). It returns the scene, the card's instance and the Maneuver.
func dfScene(t *testing.T, c *Card, to ability.Zone, answers ...Answer) (*scene, *Instance, *Instance) {
	t.Helper()
	s := newScene(t)
	inst := s.inDiscard(0, c)[0]
	man := s.give(0, dfManeuver(1, 1, to))
	p := s.g.Players[0]
	p.Req, p.Tac = 3, 3
	s.script(0, answers...)
	s.do(enactAction(man))
	return s, inst, man
}

func dfZoneAsks(s *scene) []Decision {
	var out []Decision
	for _, d := range s.asked(0) {
		if d.Kind == DecideZone {
			out = append(out, d)
		}
	}
	return out
}

func dfUnit(name string, zones ...Zone) *Card {
	c := unit(name, 2, 2, Keywords{})
	c.DeployZones = zones
	return c
}

func TestFR_TRN_003a_DeployFreeEntersFromDiscardWithoutCosts(t *testing.T) {
	// Covers: FR-CST-001, FR-ZON-003
	c := dfUnit("Costly", ZoneReserves)
	c.ReqCost, c.TacCost = 2, 2
	s, inst, _ := dfScene(t, c, ability.ZoneAny)
	a := assetOf(s.g, c)
	if a == nil {
		t.Fatalf("the unit card is on the battlefield after resolution (FR-TRN-003a)")
	}
	if a.Controller != 0 || a.Zone != ZoneReserves || !slices.Contains(s.g.Players[0].Reserves, a) {
		t.Errorf("a free deploy enters player 0's Reserves under player 0's control, its only printed zone: controller %d, zone %v (FR-TRN-003a, FR-ZON-003)", a.Controller, a.Zone)
	}
	if s.g.DiscardCard(inst.ID) != nil {
		t.Errorf("the deployed card left the discard pile (FR-TRN-003a)")
	}
	if p := s.g.Players[0]; p.Req != 2 || p.Tac != 2 {
		t.Errorf("only the Maneuver's own cost R1 T1 is paid, the deploy is free: pool R%d T%d, want R2 T2 (FR-CST-001)", p.Req, p.Tac)
	}
	for _, e := range eventsOfKind(s.g.Log, "place") {
		if e.Card == c.Name {
			t.Errorf("a free deploy puts no stack item for the deployed card: %+v (FR-TRN-003a)", e)
		}
	}
	if n := len(dfZoneAsks(s)); n != 0 {
		t.Errorf("a single printed zone is asked of nobody: %d DecideZone (FR-TRN-003a 1)", n)
	}
}

func TestFR_TRN_003a_DeployFreeAsksZoneWhenBothPrinted(t *testing.T) {
	// Covers: SVY-R43
	c := dfUnit("Either", ZoneMainTheater, ZoneReserves)
	s, _, _ := dfScene(t, c, ability.ZoneAny, Answer{Picks: []int{1}})
	if n := len(s.asked(0)); n != 1 {
		t.Errorf("the DecideZone is the only decision player 0 was asked, got %d", n)
	}
	asks := dfZoneAsks(s)
	if len(asks) != 1 {
		t.Fatalf("player 0 is asked one DecideZone, got %d (FR-TRN-003a 1)", len(asks))
	}
	d := asks[0]
	if d.Player != 0 || d.Min != 1 || d.Max != 1 || len(d.Options) != 2 {
		t.Fatalf("the controller picks exactly one of two zones: %+v", d)
	}
	if d.Options[0].Label != ZoneReserves.String() || d.Options[1].Label != ZoneMainTheater.String() {
		t.Errorf("options list Reserves first, then the Main Theater, whatever order the card prints (SVY-R43): %v", d.Options)
	}
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneMainTheater {
		t.Errorf("answering the second option deploys into the Main Theater (FR-TRN-003a 1): %v", a)
	}
}

func TestFR_MOD_001_DeployFreeZoneTimeoutEntersReserves(t *testing.T) {
	// Covers: SVY-R43, FR-TRN-003a
	c := dfUnit("Either", ZoneMainTheater, ZoneReserves)
	s, _, _ := dfScene(t, c, ability.ZoneAny, Answer{Timeout: true})
	if n := len(dfZoneAsks(s)); n != 1 {
		t.Fatalf("player 0 is asked one DecideZone, got %d", n)
	}
	if n := len(s.asked(0)); n != 1 {
		t.Errorf("the DecideZone is the only decision player 0 was asked, got %d", n)
	}

	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves {
		t.Errorf("a timed-out free-deploy zone choice enters Reserves (SVY-R43, FR-MOD-001 6): %v", a)
	}
}

func TestFR_TRN_003a_DeployFreeExplicitZoneOverridesPrinted(t *testing.T) {
	// Covers: FR-TRN-003a
	t.Run("MainTheaterOverReservesOnly", func(t *testing.T) {
		c := dfUnit("Home", ZoneReserves)
		s, _, _ := dfScene(t, c, ability.ZoneMainTheater)
		if a := assetOf(s.g, c); a == nil || a.Zone != ZoneMainTheater {
			t.Errorf("a step naming the Main Theater overrides the printed zones (FR-TRN-003a 2): %v", a)
		}
		if n := len(dfZoneAsks(s)); n != 0 {
			t.Errorf("an explicit zone asks nobody: %d DecideZone", n)
		}
	})
	t.Run("ReservesOverBoth", func(t *testing.T) {
		c := dfUnit("Either", ZoneMainTheater, ZoneReserves)
		s, _, _ := dfScene(t, c, ability.ZoneReserves)
		if a := assetOf(s.g, c); a == nil || a.Zone != ZoneReserves {
			t.Errorf("a step naming Reserves overrides the printed choice (FR-TRN-003a 2): %v", a)
		}
		if n := len(dfZoneAsks(s)); n != 0 {
			t.Errorf("an explicit zone asks nobody: %d DecideZone", n)
		}
	})
}

func TestFR_TRN_003a_DeployFreeFiresDeployTriggersNotMoves(t *testing.T) {
	// Covers: SVY-R1, FR-ZON-002
	c := dfUnit("Triggered", ZoneReserves)
	c.Ability = &ability.Def{Triggers: []ability.Trigger{
		{Hook: ability.HookDeploy, Effect: steps(drawN(1))},
		{Hook: ability.HookMoves, Effect: steps(ability.Step{Op: ability.OpAddArmor, Amount: ability.Value{N: 1}, Subject: ability.SubjectSelf})},
	}}
	s := newScene(t)
	s.inDiscard(0, c)
	man := s.give(0, dfManeuver(0, 0, ability.ZoneMainTheater))
	s.script(0)
	deck := len(s.g.Players[0].Deck)
	s.do(enactAction(man))
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneMainTheater {
		t.Fatalf("the unit is deployed into the Main Theater: %v", a)
	}
	if got := deck - len(s.g.Players[0].Deck); got != 1 {
		t.Errorf("its deploy trigger fires and resolves: player 0 drew %d cards, want 1 (FR-TRN-003a)", got)
	}
	if a.Armor != 0 {
		t.Errorf("entering the battlefield by a free deploy is not a move, so its move trigger never fires: armor %d, want 0 (SVY-R1)", a.Armor)
	}
}

func TestFR_TRN_003a_DeployFreeKeepsDeploysExhausted(t *testing.T) {
	// Covers: FR-KWD-002l
	c := unit("Tired", 2, 2, Keywords{DeploysExhausted: true})
	c.DeployZones = []Zone{ZoneReserves}
	s, _, _ := dfScene(t, c, ability.ZoneAny)
	a := assetOf(s.g, c)
	if a == nil {
		t.Fatalf("the unit is deployed")
	}
	if !a.Exhausted {
		t.Errorf("a free-deployed Deploys Exhausted unit enters Exhausted (FR-KWD-002l)")
	}
}

func TestFR_TGT_003_DeployFreeTargetLeftPileDeploysNothing(t *testing.T) {
	s := newScene(t)
	c := dfUnit("Leaves", ZoneReserves)
	inst := s.inDiscard(0, c)[0]
	man := s.give(0, dfManeuver(0, 0, ability.ZoneAny))
	s.script(0)
	// Once the Maneuver is placed with its target chosen, the card goes to hand.
	r := &recorder{}
	moved := false
	r.onEvent = func(e Event) {
		if e.Kind == "place" && !moved {
			moved = true
			if s.g.takeCard(inst) {
				s.g.putCard(inst, ZoneHand, false)
			}
		}
	}
	s.g.obs = r
	s.do(enactAction(man))
	if !moved {
		t.Fatalf("the Maneuver was placed on the stack")
	}
	if assetOf(s.g, c) != nil {
		t.Errorf("a target that left its discard pile before resolution is not deployed (FR-TGT-003)")
	}
	if !inHand(s.g.Players[0], inst) {
		t.Errorf("the card stays where it went, in hand (FR-TGT-003)")
	}
}

func TestFR_TRN_003a_DeployFreeSkipsUnsupportedCard(t *testing.T) {
	// Covers: FR-MNV-001
	c := dfUnit("Odd", ZoneReserves)
	c.Unsupported = []string{"Some rule the kernel cannot run."}
	s, inst, man := dfScene(t, c, ability.ZoneAny)
	if assetOf(s.g, c) != nil {
		t.Errorf("a card with text the kernel cannot run is never put onto the battlefield")
	}
	if s.g.DiscardCard(inst.ID) != inst {
		t.Errorf("the unsupported card stays in the discard pile")
	}
	if s.g.DiscardCard(man.ID) != man {
		t.Errorf("the Maneuver still goes to the discard pile (FR-MNV-001)")
	}
}
