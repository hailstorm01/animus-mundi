// Tests for T-0303 (FR-MNV-001, FR-TGT-001, FR-TGT-003, FR-REP-004): the
// Concordia Maneuver Soul Whip, loaded from data/cards. Player 0 enacts;
// player 1 is the opponent.
//
// Soul Whip: "Chosen allied unit deals damage equal to its power to chosen
// opposing unit." Cost 2 Tactics.
package engine

import (
	"reflect"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// swBoard is an allied dealer (power 3) and an opposing unit (power 1), both
// in the Main Theater with health 30.
type swBoard struct {
	s         *scene
	ally, foe *Asset
}

func newSWBoard(t *testing.T) *swBoard {
	t.Helper()
	s := newScene(t)
	return &swBoard{
		s:    s,
		ally: s.put(0, ZoneMainTheater, unit("Ally", 3, 30, ability.Keywords{})),
		foe:  s.put(1, ZoneMainTheater, unit("Foe", 1, 30, ability.Keywords{})),
	}
}

// swTargetAsks are the DecideTarget decisions seat 0 was asked, in order.
func swTargetAsks(s *scene) []Decision {
	var out []Decision
	for _, d := range s.asks[0] {
		if d.Kind == DecideTarget {
			out = append(out, d)
		}
	}
	return out
}

func swOffers(d Decision, a *Asset) bool {
	for _, o := range d.Options {
		if o.Ref == a.ID {
			return true
		}
	}
	return false
}

func TestCard_SoulWhip_ChosenAllyDealsItsPower(t *testing.T) {
	// Covers: FR-TGT-003, SVY-R27, FR-MNV-001
	b := newSWBoard(t)
	other := b.s.put(1, ZoneMainTheater, unit("Other", 2, 30, ability.Keywords{}))
	allyOther := b.s.put(0, ZoneMainTheater, unit("AllyOther", 5, 30, ability.Keywords{}))
	inst := enactFor(b.s, "Soul Whip")
	b.s.scriptModal(0, nil, nil, nil, pick(b.ally), pick(b.foe))
	b.s.do(enactAction(inst))
	wantDamageFrom(t, b.s.g, b.foe, b.ally, 3, "the opposing unit takes the dealer's power")
	if got := len(nonCombatDamage(b.s.g.Log)); got != 1 {
		t.Errorf("exactly one damage event, got %d (SVY-R27)", got)
	}
	b.s.wantDamage(b.ally, 0)
	if b.ally.Exhausted {
		t.Errorf("dealing damage does not Exhaust the dealer (FR-TGT-003)")
	}
	b.s.wantDamage(other, 0)
	b.s.wantDamage(allyOther, 0)
	if lastDiscard(b.s.g.Players[0]) != inst {
		t.Errorf("Soul Whip ends on top of its owner's discard pile (FR-MNV-001)")
	}
	noFormatErrors(t, b.s.g)
}

func TestCard_SoulWhip_TwoTargetQuestions(t *testing.T) {
	// Covers: FR-TGT-001
	s := newScene(t)
	a1 := s.put(0, ZoneMainTheater, unit("A1", 2, 30, ability.Keywords{}))
	a2 := s.put(0, ZoneReserves, unit("A2", 3, 30, ability.Keywords{}))
	aSc := s.putScratch(0)
	aStruct := s.put(0, ZoneMainTheater, structure("AStruct", 0, 30, ability.Keywords{}))
	f1 := s.put(1, ZoneMainTheater, unit("F1", 2, 30, ability.Keywords{}))
	f2 := s.put(1, ZoneReserves, unit("F2", 3, 30, ability.Keywords{}))
	fSc := s.putScratch(1)
	fStruct := s.put(1, ZoneMainTheater, structure("FStruct", 0, 30, ability.Keywords{}))
	inst := enactFor(s, "Soul Whip")
	s.scriptModal(0, nil, nil, nil, pick(a1), pick(f1))
	s.do(enactAction(inst))
	asks := swTargetAsks(s)
	if len(asks) != 2 {
		t.Fatalf("two target questions are asked, got %d (FR-TGT-001)", len(asks))
	}
	if got, want := refs(asks[0]), sortedCopy(ids(a1, a2, aSc)); !reflect.DeepEqual(got, want) {
		t.Errorf("the first question offers exactly player 0's units: got %v, want %v (FR-TGT-001)", got, want)
	}
	if got, want := refs(asks[1]), sortedCopy(ids(f1, f2, fSc)); !reflect.DeepEqual(got, want) {
		t.Errorf("the second question offers exactly player 1's units: got %v, want %v (FR-TGT-001)", got, want)
	}
	for _, d := range asks {
		if d.Min != 1 || d.Max != 1 {
			t.Errorf("each question takes exactly one pick: Min %d, Max %d (FR-TGT-001)", d.Min, d.Max)
		}
		if swOffers(d, aStruct) || swOffers(d, fStruct) {
			t.Errorf("structures are never offered (FR-TGT-001)")
		}
		for _, o := range d.Options {
			if x := s.g.Asset(o.Ref); x != nil && x.Inst.Card.Type == TypeHeadquarters {
				t.Errorf("a Headquarters is never offered: option %q (FR-TGT-001)", o.Label)
			}
		}
	}
}

func TestCard_SoulWhip_AnyZoneAndExhausted(t *testing.T) {
	// Covers: FR-TGT-001, SVY-R27
	s := newScene(t)
	dealer := s.put(0, ZoneReserves, unit("Dealer", 4, 30, ability.Keywords{}))
	dealer.Exhausted = true
	foe := s.put(1, ZoneReserves, unit("Foe", 1, 30, ability.Keywords{}))
	inst := enactFor(s, "Soul Whip")
	s.scriptModal(0, nil, nil, nil, pick(dealer), pick(foe))
	s.do(enactAction(inst))
	wantDamageFrom(t, s.g, foe, dealer, 4, "an Exhausted dealer in Reserves deals to an opposing unit in Reserves")
}

func TestCard_SoulWhip_DealerUsesItsCurrentPower(t *testing.T) {
	// Covers: SVY-R27, FR-LAY-001
	b := newSWBoard(t)
	boost := enactFor(b.s, "Chainbreaker's Tenacity")
	chooseAsset(b.s, b.ally)
	b.s.do(enactAction(boost))
	if got := b.s.g.Power(b.ally); got != 4 {
		t.Fatalf("precondition: Chainbreaker's Tenacity gives power 3 +1 = 4, got %d", got)
	}
	inst := enactFor(b.s, "Soul Whip")
	b.s.scriptModal(0, nil, nil, nil, pick(b.ally), pick(b.foe))
	b.s.do(enactAction(inst))
	wantDamageFrom(t, b.s.g, b.foe, b.ally, 4, "the dealer deals its current power, printed 3 plus 1")
}

func TestCard_SoulWhip_PowerReadOnResolution(t *testing.T) {
	// Covers: SVY-R27, SVY-R3
	b := newSWBoard(t)
	inst := enactFor(b.s, "Soul Whip")
	b.s.scriptModal(0, nil, nil, nil, pick(b.ally), pick(b.foe))
	r := &recorder{}
	r.onEvent = func(e Event) {
		// After Soul Whip is placed and before it resolves, the dealer's power changes.
		if e.Kind == "place" && e.Card == "Soul Whip" {
			b.ally.Inst.Card = unit("Ally", 7, 30, ability.Keywords{})
		}
	}
	b.s.g.obs = r
	b.s.do(enactAction(inst))
	wantDamageFrom(t, b.s.g, b.foe, b.ally, 7, "the dealer deals the power it has on resolution, not the 3 it had when chosen")
}

func TestCard_SoulWhip_DealerGoneDealsNothing(t *testing.T) {
	// Covers: FR-TGT-003, SVY-R27, FR-ZON-004
	b := newSWBoard(t)
	inst := enactFor(b.s, "Soul Whip")
	b.s.scriptModal(0, nil, nil, nil, pick(b.ally), pick(b.foe))
	r := &recorder{}
	r.onEvent = func(e Event) {
		if e.Kind == "place" && e.Card == "Soul Whip" && b.s.alive(b.ally) {
			b.s.g.removeFromBattlefield(b.ally)
		}
	}
	b.s.g.obs = r
	b.s.do(enactAction(inst))
	if b.s.alive(b.ally) {
		t.Fatalf("precondition: the dealer left the battlefield while Soul Whip waited on the stack")
	}
	if got := len(nonCombatDamage(b.s.g.Log)); got != 0 {
		t.Errorf("a dealer that is gone at resolution deals nothing, with no last-known power: %d damage events, want 0 (SVY-R27)", got)
	}
	b.s.wantDamage(b.foe, 0)
	if lastDiscard(b.s.g.Players[0]) != inst {
		t.Errorf("Soul Whip still ends in its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_SoulWhip_TargetGoneDealsNothing(t *testing.T) {
	// Covers: FR-TGT-003
	b := newSWBoard(t)
	inst := enactFor(b.s, "Soul Whip")
	b.s.scriptModal(0, nil, nil, nil, pick(b.ally), pick(b.foe))
	r := &recorder{}
	r.onEvent = func(e Event) {
		if e.Kind == "place" && e.Card == "Soul Whip" && b.s.alive(b.foe) {
			b.s.g.removeFromBattlefield(b.foe)
		}
	}
	b.s.g.obs = r
	b.s.do(enactAction(inst))
	if b.s.alive(b.foe) {
		t.Fatalf("precondition: the opposing unit left the battlefield while Soul Whip waited on the stack")
	}
	if got := len(nonCombatDamage(b.s.g.Log)); got != 0 {
		t.Errorf("an opposing unit that is gone at resolution takes nothing: %d damage events, want 0 (FR-TGT-003)", got)
	}
	if !b.s.alive(b.ally) {
		t.Errorf("the dealer stays on the battlefield (FR-TGT-003)")
	}
	b.s.wantDamage(b.ally, 0)
}

func TestCard_SoulWhip_StalwartDoesNotCapIt(t *testing.T) {
	// Covers: FR-REP-004
	s := newScene(t)
	dealer := s.put(0, ZoneMainTheater, unit("Dealer", 5, 30, ability.Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Bulwark", 1, 30, ability.Keywords{Stalwart: true}))
	inst := enactFor(s, "Soul Whip")
	s.scriptModal(0, nil, nil, nil, pick(dealer), pick(foe))
	s.do(enactAction(inst))
	s.wantDamage(foe, 5)
}

func TestCard_SoulWhip_ResistantAndArmorApply(t *testing.T) {
	// Covers: FR-REP-004, FR-KWD-002b, FR-KWD-002k
	s := newScene(t)
	dealer := s.put(0, ZoneMainTheater, unit("Dealer", 4, 30, ability.Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Plated", 1, 30, ability.Keywords{Resistant: 1, Armor: 1}))
	inst := enactFor(s, "Soul Whip")
	s.scriptModal(0, nil, nil, nil, pick(dealer), pick(foe))
	s.do(enactAction(inst))
	s.wantDamage(foe, 2)
	if foe.Armor != 0 {
		t.Errorf("power 4 less Resistant 1 leaves 3; Armor 1 absorbs one, so 2 damage is marked and the Armor is gone: Armor %d (FR-REP-004)", foe.Armor)
	}
}

func TestCard_SoulWhip_LethalDamageKills(t *testing.T) {
	// Covers: FR-LAY-002
	s := newScene(t)
	dealer := s.put(0, ZoneMainTheater, unit("Dealer", 4, 30, ability.Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 4, ability.Keywords{}))
	inst := enactFor(s, "Soul Whip")
	s.scriptModal(0, nil, nil, nil, pick(dealer), pick(foe))
	s.do(enactAction(inst))
	if s.alive(foe) || !inDiscardPile(s.g.Players[1], foe.Inst) {
		t.Errorf("a unit with Health equal to the damage dies and its card goes to its owner's discard pile (FR-LAY-002)")
	}
}

func TestCard_SoulWhip_UnbowedTriggers(t *testing.T) {
	// Covers: FR-KWD-002a
	s := newScene(t)
	dealer := s.put(0, ZoneMainTheater, unit("Dealer", 2, 30, ability.Keywords{}))
	foe := s.put(1, ZoneMainTheater, bowed("Bowed", 1, 30, ability.Keywords{}, unbowed(nil, scratchN(1))))
	inst := enactFor(s, "Soul Whip")
	s.scriptModal(0, nil, nil, nil, pick(dealer), pick(foe))
	s.do(enactAction(inst))
	s.wantDamage(foe, 2)
	s.wantScratches(1, 1, "the surviving opposing unit's Unbowed triggers once (FR-KWD-002a)")
	s.wantScratches(0, 0, "the trigger belongs to its asset's controller")
}

func TestCard_SoulWhip_WardedOpposingUnitNotOffered(t *testing.T) {
	// Covers: FR-KWD-002h, FR-TGT-001
	s := newScene(t)
	allyWard := s.put(0, ZoneMainTheater, unit("AllyWard", 2, 30, ability.Keywords{Ward: true}))
	foeWard := s.put(1, ZoneMainTheater, unit("FoeWard", 2, 30, ability.Keywords{Ward: true}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 2, 30, ability.Keywords{}))
	inst := enactFor(s, "Soul Whip")
	s.scriptModal(0, nil, nil, nil, pick(allyWard), pick(foe))
	s.do(enactAction(inst))
	asks := swTargetAsks(s)
	if len(asks) != 2 {
		t.Fatalf("two target questions are asked, got %d", len(asks))
	}
	if !swOffers(asks[0], allyWard) {
		t.Errorf("an allied Ward unit is offered as the dealer (FR-KWD-002h)")
	}
	if swOffers(asks[1], foeWard) || !swOffers(asks[1], foe) {
		t.Errorf("an opposing Ward unit is not offered beside one without Ward: options %v (FR-KWD-002h)", labels(asks[1]))
	}

	s2 := newScene(t)
	s2.put(0, ZoneMainTheater, unit("Ally", 2, 30, ability.Keywords{}))
	s2.put(1, ZoneMainTheater, unit("OnlyWard", 2, 30, ability.Keywords{Ward: true}))
	inst2 := enactFor(s2, "Soul Whip")
	if offersEnact(s2.g, inst2) {
		t.Errorf("with a Ward unit the only opposing unit, Soul Whip is not offered (FR-KWD-002h)")
	}
}

func TestCard_SoulWhip_NeedsBothUnits(t *testing.T) {
	// Covers: SVY-R29, FR-MNV-001
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("Foe", 1, 30, ability.Keywords{}))
	inst := enactFor(s, "Soul Whip")
	if offersEnact(s.g, inst) {
		t.Errorf("with no allied unit, Soul Whip is not offered (SVY-R29)")
	}
	s.reject(enactAction(inst))

	s2 := newScene(t)
	s2.put(0, ZoneMainTheater, unit("Ally", 2, 30, ability.Keywords{}))
	inst2 := enactFor(s2, "Soul Whip")
	if offersEnact(s2.g, inst2) {
		t.Errorf("with no opposing unit, Soul Whip is not offered (SVY-R29)")
	}
	s2.reject(enactAction(inst2))

	b := newSWBoard(t)
	inst3 := enactFor(b.s, "Soul Whip")
	if !offersEnact(b.s.g, inst3) {
		t.Errorf("with one allied and one opposing unit, Soul Whip is offered (SVY-R29)")
	}
}

func TestCard_SoulWhip_PaysTwoTactics(t *testing.T) {
	// Covers: FR-MNV-001
	b := newSWBoard(t)
	inst := enactFor(b.s, "Soul Whip")
	b.s.setPool(0, 5, 1)
	if offersEnact(b.s.g, inst) {
		t.Errorf("with 1 Tactic, Soul Whip costing 2 is not offered (FR-MNV-001)")
	}
	b.s.setPool(0, 3, 5)
	b.s.scriptModal(0, nil, nil, nil, pick(b.ally), pick(b.foe))
	b.s.do(enactAction(inst))
	if p := b.s.g.Players[0]; p.Req != 3 || p.Tac != 3 {
		t.Errorf("enacting costs 2 Tactics and no Requisition: pool R%d T%d, want R3 T3 (FR-MNV-001)", p.Req, p.Tac)
	}
}
