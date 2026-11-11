package engine

import (
	"fmt"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0211 (FR-ENT-103, directive #142): OpMorphScratch turns a
// battlefield unit face down in place. Cards are synthetic. Player 0 is the
// active player and enacts a free Maneuver whose first step morphs the one
// opposing unit (a single legal target is chosen without a question,
// FR-MOD-001 1).

var morphOpposingUnit = ability.Target{Filter: ability.Filter{
	Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}}

// morphManeuver is a Maneuver whose first step morphs its one opposing-unit
// target, followed by the given steps.
func morphManeuver(then ...ability.Step) *Card {
	first := ability.Step{Op: ability.OpMorphScratch, Subject: ability.SubjectTarget}
	return maneuver("Mutate", 0, 0, ability.Effect{
		Targets: []ability.Target{morphOpposingUnit},
		Steps:   append([]ability.Step{first}, then...),
	})
}

func morphEnact(s *scene, c *Card) {
	s.t.Helper()
	s.do(enactAction(s.give(0, c)))
}

func TestFR_ENT_103_MorphTurnsUnitFaceDownInPlace(t *testing.T) {
	// Covers: FR-LAY-001
	s := newScene(t)
	card := unit("Bulwark", 4, 5, Keywords{Stalwart: true})
	u := s.put(1, ZoneMainTheater, card)
	u.Exhausted, u.Armor, u.TempPower = true, 3, 2
	id := u.ID
	morphEnact(s, morphManeuver())
	a := s.g.Asset(id)
	if a == nil || a.Zone != ZoneMainTheater || a.Controller != 1 {
		t.Fatalf("the morphed unit stays the same asset on the battlefield in player 1's Main Theater (FR-ENT-103): %v", a)
	}
	if !a.FaceDown {
		t.Errorf("the unit is face down (FR-ENT-103)")
	}
	if !a.Exhausted || a.Armor != 3 || a.TempPower != 2 {
		t.Errorf("it keeps its readiness, Armor and Layer 5 modifier: Exhausted %v, Armor %d, TempPower %d; want true, 3, 2 (FR-ENT-103)", a.Exhausted, a.Armor, a.TempPower)
	}
	if got := s.g.Power(a); got != 3 {
		t.Errorf("effective power = %d, want 3: Scratch Company 1 plus the kept +2 (FR-LAY-001 Layer 1, FR-ENT-103)", got)
	}
	if got := s.g.MaxHealth(a); got != 1 {
		t.Errorf("effective Maximum Health = %d, want 1 (FR-LAY-001 Layer 1)", got)
	}
	if got := s.g.Keywords(a); got != (Keywords{DeploysExhausted: true}) {
		t.Errorf("effective keywords = %+v, want Deploys Exhausted only, no Stalwart (FR-ENT-103)", got)
	}
	if st := a.Printed().Subtypes; len(st) != 2 || st[0] != "Infantry" || st[1] != "Scratch Company" {
		t.Errorf("subtypes = %v, want [Infantry Scratch Company] (FR-ENT-103)", st)
	}
}

func TestFR_ENT_103_MorphKeepsAReadiedUnitReadied(t *testing.T) {
	// Covers: FR-KWD-002l
	s := newScene(t)
	u := s.put(1, ZoneMainTheater, unit("Bulwark", 4, 5, Keywords{}))
	morphEnact(s, morphManeuver())
	if !u.FaceDown || u.Exhausted {
		t.Errorf("a Readied unit turns face down and stays Readied (Deploys Exhausted does nothing on the battlefield, FR-ENT-103, FR-KWD-002l): FaceDown %v, Exhausted %v", u.FaceDown, u.Exhausted)
	}
}

func TestFR_ENT_103_MorphWithDamageIsDestroyed(t *testing.T) {
	// Covers: FR-LAY-002, FR-ZON-004, FR-ZON-005
	s := newScene(t)
	card := unit("Martyr", 3, 4, Keywords{})
	card.Ability = &ability.Def{Triggers: []ability.Trigger{{Hook: ability.HookKilled, Effect: steps(drawN(1))}}}
	u := s.put(1, ZoneMainTheater, card)
	u.Damage, u.Armor = 1, 2
	inst := u.Inst
	hand := len(s.g.Players[1].Hand)
	morphEnact(s, morphManeuver())
	if s.alive(u) {
		t.Fatalf("1 damage on a Maximum Health of 1 is lethal: the unit is destroyed in the sweep after the item (FR-ENT-103, FR-LAY-002)")
	}
	if lastDiscard(s.g.Players[1]) != inst || inst.Card != card {
		t.Errorf("its own card is in player 1's discard pile as its printed card %q (FR-ZON-005)", card.Name)
	}
	if d := eventsOfKind(s.g.Log, EventDestroy); len(d) != 1 {
		t.Errorf("%d destroy events, want 1 (FR-LAY-002): %+v", len(d), d)
	}
	if got := len(s.g.Players[1].Hand); got != hand {
		t.Errorf("player 1's hand %d -> %d: a face-down unit has no printed death trigger (FR-ZON-004)", hand, got)
	}
}

func TestFR_ENT_103_MorphedUnitFiresNoPrintedTriggers(t *testing.T) {
	// Covers: FR-ZON-005
	s := newScene(t)
	card := unit("Wanderer", 2, 5, Keywords{})
	card.Ability = &ability.Def{Triggers: []ability.Trigger{{Hook: ability.HookMoves, Effect: steps(drawN(1))}}}
	u := s.put(1, ZoneMainTheater, card)
	hand := len(s.g.Players[1].Hand)
	morphEnact(s, morphManeuver(ability.Step{Op: ability.OpMoveZone, Subject: ability.SubjectTarget, To: ability.ZoneReserves}))
	if !u.FaceDown || u.Zone != ZoneReserves || !s.alive(u) {
		t.Fatalf("precondition: the unit is face down in player 1's Reserves: FaceDown %v, zone %v", u.FaceDown, u.Zone)
	}
	if got := len(s.g.Players[1].Hand); got != hand {
		t.Errorf("player 1's hand %d -> %d: a face-down unit's printed Moves trigger does not fire (FR-ENT-103, FR-ZON-005)", hand, got)
	}
}

func TestFR_ENT_103_MorphedUnitReturnsToHandAsPrintedCard(t *testing.T) {
	// Covers: FR-ZON-004
	s := newScene(t)
	card := unit("Bulwark", 4, 5, Keywords{Stalwart: true})
	u := s.put(1, ZoneMainTheater, card)
	inst := u.Inst
	morphEnact(s, morphManeuver(ability.Step{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget}))
	if s.alive(u) {
		t.Fatalf("the unit left the battlefield (FR-ZON-004)")
	}
	if !inHand(s.g.Players[1], inst) || inst.Card != card {
		t.Errorf("player 1's hand holds the unit's own instance as its printed card, not a Scratch Company (FR-ENT-103, FR-ZON-005)")
	}
}

func TestFR_ENT_103_MorphSkipsUnitAlreadyFaceDown(t *testing.T) {
	then := ability.Step{Op: ability.OpDraw, Subject: ability.SubjectController, Amount: ability.Value{N: 1}, IfDone: true}

	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("Bulwark", 4, 5, Keywords{}))
	deck := len(s.g.Players[0].Deck)
	morphEnact(s, morphManeuver(then))
	if got := deck - len(s.g.Players[0].Deck); got != 1 {
		t.Errorf("turning a face-up unit face down did something, so the \"if you do\" draw ran: drew %d, want 1 (FR-ENT-103)", got)
	}

	s = newScene(t)
	sc := s.putScratch(1)
	deck = len(s.g.Players[0].Deck)
	from := len(s.g.Log)
	morphEnact(s, morphManeuver(then))
	if !sc.FaceDown {
		t.Errorf("an already face-down unit stays face down")
	}
	if ev := effectEvents(s.g, from, EventFaceDown); len(ev) != 0 {
		t.Errorf("%d face-down events for a unit already face down, want 0: %+v", len(ev), ev)
	}
	if got := deck - len(s.g.Players[0].Deck); got != 0 {
		t.Errorf("the step turned nothing, so the \"if you do\" draw did not run: drew %d, want 0 (FR-ENT-103)", got)
	}
}

func TestEffectEvents_MorphLogsTheFaceDownChange(t *testing.T) {
	// ADR-0005: one event per asset acted on.
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("Bulwark", 4, 5, Keywords{}))
	s.put(1, ZoneReserves, unit("Sentry", 2, 3, Keywords{}))
	var want []*Asset
	for _, a := range s.g.Assets() {
		if a.Controller == 1 && a.Printed().Type == TypeUnit {
			want = append(want, a)
		}
	}
	names := map[int]string{}
	for _, a := range want {
		names[a.ID] = fmt.Sprintf("%s#%d turns face down", a.Printed().Name, a.ID)
	}
	all := maneuver("Mass Mutate", 0, 0, steps(ability.Step{Op: ability.OpMorphScratch, Subject: ability.SubjectAll,
		Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}}))
	from := len(s.g.Log)
	morphEnact(s, all)
	got := effectEvents(s.g, from, EventFaceDown)
	if len(got) != len(want) || len(want) != 2 {
		t.Fatalf("%d face-down events, want one per unit turned (2): %+v", len(got), got)
	}
	for _, a := range want {
		var mine []Event
		for _, e := range got {
			if e.Target == a.ID {
				mine = append(mine, e)
			}
		}
		if len(mine) != 1 {
			t.Errorf("unit %d has %d face-down events, want exactly 1: %+v", a.ID, len(mine), got)
			continue
		}
		e := mine[0]
		if e.Player != -1 || e.Card != "" {
			t.Errorf("event for unit %d: player %d, card %q; want player -1, no card", a.ID, e.Player, e.Card)
		}
		if w := names[a.ID]; e.Detail != w {
			t.Errorf("event for unit %d: detail %q, want %q (naming the printed card, logged before the flip)", a.ID, e.Detail, w)
		}
		if !a.FaceDown {
			t.Errorf("unit %d is face down after the step", a.ID)
		}
	}
}
