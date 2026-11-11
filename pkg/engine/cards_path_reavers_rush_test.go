package engine

import (
	"slices"
	"testing"
)

// Card tests for Reaver's Rush (T-0210), written from the CSV text: Path
// Maneuver, R0 T3, "Deploy chosen unit card with power 3 or less from your
// discard pile for free." Player 0 enacts. Pile cards other than the real
// Devilhorn Guerrillas and Chainbreaker Phalanx are synthetic, Reserves-only
// unless a test says otherwise, so no deploy trigger asks anything.

// rrUnit is a Reserves-only unit card of the given power.
func rrUnit(name string, power int) *Card {
	c := unit(name, power, 3, Keywords{})
	c.DeployZones = []Zone{ZoneReserves}
	return c
}

// rrScene gives player 0 Reaver's Rush with exactly R0 T3.
func rrScene(t *testing.T) (*scene, *Instance) {
	t.Helper()
	s := newScene(t)
	inst := s.give(0, s.card("Reaver's Rush"))
	s.setPool(0, 0, 3)
	return s, inst
}

func rrZoneAsks(s *scene) []Decision {
	var out []Decision
	for _, d := range s.asked(0) {
		if d.Kind == DecideZone {
			out = append(out, d)
		}
	}
	return out
}

func TestCard_ReaversRush_DeploysChosenUnitCardForFree(t *testing.T) {
	// Covers: FR-TRN-003a, FR-CST-001
	s, rush := rrScene(t)
	c := rrUnit("Costly", 3)
	c.ReqCost, c.TacCost = 2, 1
	pile := s.inDiscard(0, c)[0]
	s.script(0)
	s.do(enactAction(rush))
	p := s.g.Players[0]
	a := assetOf(s.g, c)
	if a == nil {
		t.Fatalf("the chosen unit card is on the battlefield after resolution (FR-TRN-003a)")
	}
	if a.Controller != 0 || a.Zone != ZoneReserves || !slices.Contains(p.Reserves, a) {
		t.Errorf("it enters player 0's Reserves under player 0's control: controller %d zone %v (FR-TRN-003a)", a.Controller, a.Zone)
	}
	if inDiscardPile(p, pile) {
		t.Errorf("the deployed card left the discard pile (FR-TRN-003a)")
	}
	if p.Req != 0 || p.Tac != 0 {
		t.Errorf("only the Maneuver's own T3 is paid, the deploy is free: pool R%d T%d, want R0 T0 (FR-CST-001)", p.Req, p.Tac)
	}
	if n := len(rrZoneAsks(s)); n != 0 {
		t.Errorf("a single printed zone is asked of nobody: %d DecideZone (FR-TRN-003a 1)", n)
	}
	if !inDiscardPile(p, rush) {
		t.Errorf("the resolved Reaver's Rush is in its controller's discard pile")
	}
}

func TestCard_ReaversRush_OffersOnlyUnitCardsPowerThreeOrLess(t *testing.T) {
	// Covers: FR-TGT-001, FR-ZON-004
	s, rush := rrScene(t)
	mine := s.inDiscard(0, rrUnit("P3", 3), rrUnit("P2", 2), rrUnit("P4", 4),
		structure("Wall1", 1, 3, Keywords{}), maneuver("Spell", 0, 0, steps()))
	s.inDiscard(1, rrUnit("Theirs1", 1))
	// Pick the power-2 card by its position among the offers, whatever the order.
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		for i, o := range d.Options {
			if o.Ref == mine[1].ID {
				return Answer{Picks: []int{i}}
			}
		}
		return Answer{Timeout: true}
	}
	s.asks = map[PlayerID][]Decision{}
	s.do(enactAction(rush))
	d := paTargetAsk(t, s)
	want := ppInstIDs(mine[0], mine[1])
	if got := ppRefs(d); !slices.Equal(paSorted(got), paSorted(want)) {
		t.Errorf("offers exactly the power-3 and power-2 unit cards of its controller's pile: got %v, want %v (FR-TGT-001, FR-ZON-004)", got, want)
	}
	p := s.g.Players[0]
	if assetOf(s.g, mine[1].Card) == nil || inDiscardPile(p, mine[1]) {
		t.Errorf("choosing the power-2 card deploys it (FR-TRN-003a)")
	}
	if assetOf(s.g, mine[0].Card) != nil || !inDiscardPile(p, mine[0]) {
		t.Errorf("the power-3 card, not chosen, stays in the pile (FR-TRN-003a)")
	}
}

func TestCard_ReaversRush_NotEnactableWithoutLegalCard(t *testing.T) {
	// Covers: SVY-R29, FR-MNV-001
	notOffered := func(t *testing.T, s *scene, rush *Instance, why string) {
		t.Helper()
		if offersEnact(s.g, rush) {
			t.Errorf("Enact is not offered with %s (FR-MNV-001)", why)
		}
		s.reject(enactAction(rush))
		if !slices.Contains(s.g.Players[0].Hand, rush) {
			t.Errorf("a rejected Enact leaves Reaver's Rush in hand (FR-MNV-001)")
		}
	}
	t.Run("empty pile", func(t *testing.T) {
		s, rush := rrScene(t)
		notOffered(t, s, rush, "an empty discard pile")
	})
	t.Run("only power 4", func(t *testing.T) {
		s, rush := rrScene(t)
		s.inDiscard(0, s.card("Chainbreaker Phalanx"))
		notOffered(t, s, rush, "only a power-4 unit card in the pile")
	})
	t.Run("only a Maneuver card", func(t *testing.T) {
		s, rush := rrScene(t)
		s.inDiscard(0, maneuver("Spell", 0, 0, steps()))
		notOffered(t, s, rush, "only a Maneuver card in the pile")
	})
	t.Run("only the opponent's pile", func(t *testing.T) {
		s, rush := rrScene(t)
		s.inDiscard(1, rrUnit("Theirs1", 1))
		notOffered(t, s, rush, "a legal card only in the opposing pile")
	})
	t.Run("offered with a power-3 unit card", func(t *testing.T) {
		s, rush := rrScene(t)
		s.inDiscard(0, rrUnit("P3", 3))
		if !offersEnact(s.g, rush) {
			t.Errorf("Enact is offered with a power-3 unit card in the pile (FR-MNV-001)")
		}
	})
}

// rrGuerrillas sets up the real Devilhorn Guerrillas in player 0's pile and an
// opposing unit in the Main Theater, answers the one DecideZone with ans, and
// enacts. It returns the scene, the Guerrillas card and the opposing unit.
func rrGuerrillas(t *testing.T, ans Answer) (*scene, *Card, *Asset) {
	t.Helper()
	s, rush := rrScene(t)
	g := s.card("Devilhorn Guerrillas")
	s.inDiscard(0, g)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 9, Keywords{}))
	s.script(0, ans)
	s.do(enactAction(rush))
	if n := len(s.asked(0)); n != 1 || len(rrZoneAsks(s)) != 1 {
		t.Fatalf("the one decision player 0 is asked is a DecideZone: %v (FR-TRN-003a 1)", kindsOf(s.asked(0)))
	}
	return s, g, foe
}

func TestCard_ReaversRush_GuerrillasZoneTimeoutEntersReserves(t *testing.T) {
	// Covers: SVY-R43, FR-MOD-001
	s, g, foe := rrGuerrillas(t, Answer{Timeout: true})
	a := assetOf(s.g, g)
	if a == nil || a.Zone != ZoneReserves || a.Controller != 0 {
		t.Fatalf("a timed-out free-deploy zone choice enters player 0's Reserves (SVY-R43): %v", a)
	}
	s.wantDamage(foe, 0)
}

func TestCard_ReaversRush_GuerrillasMainTheaterFiresDeployTrigger(t *testing.T) {
	// Covers: FR-TRN-003a
	s, g, foe := rrGuerrillas(t, Answer{Picks: []int{1}})
	a := assetOf(s.g, g)
	if a == nil || a.Zone != ZoneMainTheater {
		t.Fatalf("answering the second option deploys into the Main Theater (FR-TRN-003a 1): %v", a)
	}
	s.wantDamage(foe, 2)
}
