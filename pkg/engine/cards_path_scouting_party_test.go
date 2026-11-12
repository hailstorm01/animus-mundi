package engine

import (
	"strings"
	"testing"
)

// Tests for T-0223: Devilhorn Scouting Party (Path Maneuver, R0 T2): "Look at
// the top 3 cards of your deck. You may deploy a unit card with Requisition
// cost 3 or less from among them for free. Put the rest on the bottom of your
// deck in any order. Then, Scratch 1." (FR-MOD-001 4 and 6, FR-TRN-003a,
// FR-KWD-002m, FR-ZON-004, SVY-R41, SVY-R42, SVY-R43, directive #140 (3)).

// sctLabel is a DecideLook option label: "Deploy <name> to <zone>; bottom: ..."
// (name "" is "Deploy nothing").
func sctLabel(name string, zone Zone, bottom ...string) string {
	deploy := "Deploy nothing"
	if name != "" {
		deploy = "Deploy " + name + " to " + zone.String()
	}
	return deploy + "; bottom: " + strings.Join(bottom, ", ")
}

func sctLabels(d Decision) []string {
	var out []string
	for _, o := range d.Options {
		out = append(out, o.Label)
	}
	return out
}

// sctScript scripts seat 0: a DecideLook is answered with the option labelled
// look (a timeout when look is empty); any other decision picks its first option.
func sctScript(s *scene, look string) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		if d.Kind != DecideLook {
			return Answer{Picks: []int{0}}
		}
		if look == "" {
			return Answer{Timeout: true}
		}
		for i, o := range d.Options {
			if o.Label == look {
				return Answer{Picks: []int{i}}
			}
		}
		s.t.Errorf("no DecideLook option is labelled %q; options: %v", look, sctLabels(d))
		return Answer{Timeout: true}
	}
}

// sctScene gives player 0 exactly the deck deckOf returns (first card on top)
// and Scouting Party with exactly R0 T2, scripts the look, and enacts the card.
// setup, if any, runs before the enactment.
func sctScene(t *testing.T, deckOf func(s *scene) []*Card, look string, setup func(s *scene)) (*scene, []*Instance, *Instance) {
	t.Helper()
	s := newScene(t)
	s.g.Players[0].Deck = nil
	insts := s.onTop(0, deckOf(s)...)
	if setup != nil {
		setup(s)
	}
	man := s.give(0, s.card("Devilhorn Scouting Party"))
	s.setPool(0, 0, 2)
	sctScript(s, look)
	s.do(enactAction(man))
	return s, insts, man
}

// sctReal builds a deck of real cards by name; "X" is a filler Maneuver.
func sctReal(names ...string) func(s *scene) []*Card {
	return func(s *scene) []*Card {
		var cs []*Card
		for _, n := range names {
			if n == "X" {
				cs = append(cs, ltOther("X"))
			} else {
				cs = append(cs, s.card(n))
			}
		}
		return cs
	}
}

func sctNoZoneAsked(t *testing.T, s *scene) {
	t.Helper()
	for _, d := range s.asked(0) {
		if d.Kind == DecideZone {
			t.Errorf("the zone of the free deploy is part of the look's prompt, no DecideZone is asked (SVY-R43): %q", d.Prompt)
		}
	}
}

func sctWantScratch(t *testing.T, s *scene, inst *Instance, why string) {
	t.Helper()
	a := ltAsset(s, inst)
	if a == nil || a.Controller != 0 || !a.FaceDown || a.Zone != ZoneReserves {
		t.Errorf("%s: want a face-down Scratch Company in player 0's Reserves (FR-KWD-002m), got %v", why, a)
	}
}

func TestCard_DevilhornScoutingParty_DeploysAUnitCostingThreeForFree(t *testing.T) {
	// Covers: FR-MOD-001, FR-TRN-003a, FR-CST-001, FR-SET-004, FR-KWD-002m
	s, insts, man := sctScene(t, sctReal("Chainbreaker Irregulars", "Devilhorn Guerrillas", "Devilhorn Sanction", "X"),
		sctLabel("Chainbreaker Irregulars", ZoneReserves, "Devilhorn Guerrillas", "Devilhorn Sanction"), nil)
	// 6 bottom orders deploying nothing + the Irregulars (Reserves only) x 2 orders.
	ltOneLook(t, s, 8)
	sctNoZoneAsked(t, s)
	a := ltAsset(s, insts[0])
	if a == nil || a.Controller != 0 || a.Zone != ZoneReserves || a.FaceDown {
		t.Fatalf("the Irregulars are an asset in player 0's Reserves (FR-TRN-003a): %v", a)
	}
	if a.Exhausted {
		t.Errorf("a newly deployed asset enters Readied (FR-SET-004)")
	}
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
		t.Errorf("pool is R0 T0: the Maneuver's T2 paid, the deploy free: R%d T%d (FR-CST-001)", p.Req, p.Tac)
	}
	sctWantScratch(t, s, insts[3], "Scratch 1 takes the deck's new top, X")
	ltWantDeck(t, s, insts, "the rest go to the bottom as chosen: Guerrillas, Sanction", 1, 2)
	if !inDiscardPile(s.g.Players[0], man) {
		t.Errorf("Scouting Party is in its owner's discard pile once resolved")
	}
}

func TestCard_DevilhornScoutingParty_OnlyUnitsCostingThreeOrLessAreOffered(t *testing.T) {
	// Covers: FR-ZON-004, FR-TRN-003a
	s, _, _ := sctScene(t, sctReal("Devilhorn Guerrillas", "Devilhorn Sanction", "Chainbreaker Sneak", "X"), "", nil)
	d := ltOneLook(t, s, 8) // 6 orders deploying nothing + the Sneak to the Main Theater x 2 orders
	nothing, sneak := 0, 0
	for _, o := range d.Options {
		switch {
		case strings.HasPrefix(o.Label, "Deploy nothing"):
			nothing++
		case strings.HasPrefix(o.Label, "Deploy Chainbreaker Sneak to "+ZoneMainTheater.String()+";"):
			sneak++
		default:
			t.Errorf("option %q deploys a card that is not a unit costing 3 or less (the printed Requisition cost, FR-ZON-004)", o.Label)
		}
	}
	if nothing != 6 || sneak != 2 {
		t.Errorf("want 6 options deploying nothing and 2 deploying the Sneak to its only zone, got %d and %d", nothing, sneak)
	}
}

func TestCard_DevilhornScoutingParty_NothingEligibleStillAsks(t *testing.T) {
	// Covers: FR-MOD-001
	s, insts, _ := sctScene(t, sctReal("Devilhorn Guerrillas", "Devilhorn Sanction", "Steel Rain", "X"),
		sctLabel("", 0, "Steel Rain", "Devilhorn Guerrillas", "Devilhorn Sanction"), nil)
	d := ltOneLook(t, s, 6) // a deck inspection has no single-option fast path (FR-MOD-001 1)
	for _, o := range d.Options {
		if !strings.HasPrefix(o.Label, "Deploy nothing") {
			t.Errorf("nothing is eligible, yet option %q deploys a card", o.Label)
		}
	}
	sctWantScratch(t, s, insts[3], "X is scratched after the look")
	ltWantDeck(t, s, insts, "the deck is Steel Rain, Guerrillas, Sanction as chosen", 2, 0, 1)
}

func TestCard_DevilhornScoutingParty_TimeoutDeploysNothingAndStillScratches(t *testing.T) {
	// Covers: SVY-R41, SVY-R42, FR-MOD-001
	s, insts, _ := sctScene(t, sctReal("Chainbreaker Irregulars", "Chainbreaker Sneak", "Devilhorn Guerrillas", "X"), "", nil)
	ltOneLook(t, s, 6+2+2) // 6 nothing + Irregulars (1 zone, 2 orders) + Sneak (1 zone, 2 orders); Guerrillas is R4
	for _, i := range []int{0, 1} {
		if a := ltAsset(s, insts[i]); a != nil {
			t.Errorf("a timed-out look deploys nothing, but card %d is on the battlefield (SVY-R41)", i)
		}
	}
	sctWantScratch(t, s, insts[3], "a timed-out Scouting Party still Scratches 1")
	ltWantDeck(t, s, insts, "the rest go to the bottom in the order seen, top first (SVY-R42)", 0, 1, 2)
}

func TestCard_DevilhornScoutingParty_ChosenBottomOrderKept(t *testing.T) {
	// Covers: FR-MOD-001, FR-KWD-002m
	s, insts, _ := sctScene(t, sctReal("Devilhorn Guerrillas", "Devilhorn Sanction", "Steel Rain"),
		sctLabel("", 0, "Devilhorn Sanction", "Steel Rain", "Devilhorn Guerrillas"), nil)
	ltOneLook(t, s, 6)
	sctWantScratch(t, s, insts[1], "the Sanction, the highest of the chosen bottom order, is scratched")
	ltWantDeck(t, s, insts, "the deck is Steel Rain, Guerrillas", 2, 0)
}

func TestCard_DevilhornScoutingParty_ZoneIsPartOfTheLook(t *testing.T) {
	// Covers: SVY-R43, FR-TRN-003a
	s, insts, _ := sctScene(t, func(*scene) []*Card {
		return []*Card{ltEligible("Twin"), ltOther("M1"), ltOther("M2"), ltOther("X")}
	}, sctLabel("Twin", ZoneMainTheater, "M1", "M2"), nil)
	d := ltOneLook(t, s, 6+4) // 6 nothing + Twin x 2 zones x 2 orders
	for _, z := range []Zone{ZoneReserves, ZoneMainTheater} {
		found := false
		for _, o := range d.Options {
			if strings.HasPrefix(o.Label, "Deploy Twin to "+z.String()+";") {
				found = true
			}
		}
		if !found {
			t.Errorf("no option deploys the unit to %v, a printed Deployment Zone (FR-TRN-003a 1)", z)
		}
	}
	sctNoZoneAsked(t, s)
	a := ltAsset(s, insts[0])
	if a == nil || a.Controller != 0 || a.Zone != ZoneMainTheater {
		t.Errorf("the unit is an asset in player 0's Main Theater: %v", a)
	}
}

func TestCard_DevilhornScoutingParty_DeployTriggerFires(t *testing.T) {
	// Covers: SVY-R1
	var enemy, ally *Asset
	s, insts, _ := sctScene(t, func(s *scene) []*Card {
		return []*Card{s.card("Saboteur"), ltOther("M1"), ltOther("M2"), ltOther("X")}
	}, sctLabel("Saboteur", ZoneReserves, "M1", "M2"), func(s *scene) {
		enemy = s.put(1, ZoneMainTheater, unit("Enemy", 1, 5, Keywords{}))
		ally = s.put(0, ZoneReserves, unit("Ally", 1, 5, Keywords{}))
	})
	sab := ltAsset(s, insts[0])
	if sab == nil || sab.Zone != ZoneReserves {
		t.Fatalf("Saboteur is deployed into Reserves: %v", sab)
	}
	s.wantDamage(enemy, 2)
	if got := ally.Damage + sab.Damage; got != 1 {
		t.Errorf("its deploy trigger deals 1 damage to one allied unit: ally %d, Saboteur %d (a deploy is not a move, SVY-R1)", ally.Damage, sab.Damage)
	}
}

func TestCard_DevilhornScoutingParty_EmptyDeck(t *testing.T) {
	// Covers: FR-KWD-002m, FR-MOD-001
	s, _, man := sctScene(t, sctReal(), "", nil)
	if n := len(ltLooks(s)); n != 0 {
		t.Errorf("an empty deck asks no DecideLook, got %d", n)
	}
	for _, a := range s.g.Assets() {
		if a.Controller == 0 && a.FaceDown {
			t.Errorf("Scratch on an empty deck deploys nothing: %v", a)
		}
	}
	if s.g.Outcome.Over {
		t.Errorf("a Scratch on an empty deck is safe, but the game is over: %+v (FR-KWD-002m)", s.g.Outcome)
	}
	if !inDiscardPile(s.g.Players[0], man) {
		t.Errorf("Scouting Party resolves and is in its owner's discard pile")
	}
}
