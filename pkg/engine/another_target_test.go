package engine

import (
	"reflect"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0140 (FR-TGT-002 "Another" Targets, SVY-R34, SVY-R35, board
// directives #88 and #91 (3)): a slot that says "another chosen unit" must pick
// an entity different from the slot it follows in the same clause. The effect is
// "Chosen unit deals damage equal to its power to another chosen unit": slot 1 is
// the dealer, slot 2 (Another: 1) is dealt to. Cards are synthetic Maneuvers
// enacted by player 0.
//
// Units on the dealer board (newDealerBoard): X (power 4), Y (power 1) and Z
// (power 2), all health 30, player 1's, entering the Main Theater in that order,
// so the fixed order of FR-MOD-001 sec. 6 is X, Y, Z.

// anotherEffect is the non-modal "another" effect.
func anotherEffect() ability.Effect {
	e := dealerMode()
	e.Targets[1].Another = 1
	return e
}

func anotherKillMode() ability.Effect {
	return mode([]ability.Target{anyUnit()}, ability.Step{Op: ability.OpKill, Subject: ability.SubjectTarget})
}

// anotherTargetAsks are the DecideTarget decisions seat 0 was asked, in order.
func anotherTargetAsks(s *scene) []Decision {
	var out []Decision
	for _, d := range s.asks[0] {
		if d.Kind == DecideTarget {
			out = append(out, d)
		}
	}
	return out
}

func anotherOffers(d Decision, a *Asset) bool {
	for _, o := range d.Options {
		if o.Ref == a.ID {
			return true
		}
	}
	return false
}

func anotherTargetRecords(g *Game) []DecisionRecord {
	var out []DecisionRecord
	for _, e := range eventsOfKind(g.Log, "decision") {
		if e.Decision != nil && e.Decision.Kind == DecideTarget.String() {
			out = append(out, *e.Decision)
		}
	}
	return out
}

func TestFR_MNV_001_AnotherSlotNeedsTwoDifferentUnitsToEnact(t *testing.T) {
	// Covers: SVY-R34, SVY-R29, FR-TGT-002, FR-KWD-002h
	s := newScene(t)
	inst := s.give(0, maneuver("Ambush", 0, 0, anotherEffect()))
	s.put(1, ZoneMainTheater, unit("X", 4, 30, Keywords{}))
	if offersEnact(s.g, inst) {
		t.Errorf("an \"another\" slot with one unit on the battlefield cannot be filled with a different entity, so the Maneuver is not a legal action (SVY-R34, SVY-R29)")
	}
	s.reject(enactAction(inst))
	if !inHand(s.g.Players[0], inst) {
		t.Errorf("the rejected Maneuver stays in hand (SVY-R29)")
	}
	s.put(1, ZoneMainTheater, unit("Y", 1, 30, Keywords{}))
	if !offersEnact(s.g, inst) {
		t.Errorf("with two units both slots can be filled with different entities, so the Maneuver is a legal action (SVY-R34)")
	}

	s2 := newScene(t)
	inst2 := s2.give(0, maneuver("Ambush", 0, 0, anotherEffect()))
	s2.put(1, ZoneMainTheater, unit("X", 4, 30, Keywords{}))
	s2.put(1, ZoneMainTheater, unit("Warded", 1, 30, Keywords{Ward: true}))
	if offersEnact(s2.g, inst2) {
		t.Errorf("an opposing Ward unit is never a legal target, so one legal entity is left and the Maneuver is not a legal action (SVY-R34, FR-KWD-002h)")
	}
}

func TestFR_MOD_001_AnotherSlotNeedsTwoUnits(t *testing.T) {
	// Covers: SVY-R34, SVY-R28, SVY-R29, FR-MNV-001, FR-TGT-002
	build := func(units int) (*dealerBoard, *Instance) {
		s := newScene(t)
		b := &dealerBoard{s: s}
		b.x = s.put(1, ZoneMainTheater, unit("X", 4, 30, Keywords{}))
		if units > 1 {
			b.y = s.put(1, ZoneMainTheater, unit("Y", 1, 30, Keywords{}))
		}
		inst := s.give(0, maneuver("Ambush", 0, 0, ability.Effect{
			Modes: []ability.Effect{anotherKillMode(), anotherEffect()}, ChooseN: 1}))
		return b, inst
	}
	b, inst := build(1)
	if !offersEnact(b.s.g, inst) {
		t.Fatalf("with one unit the kill mode is legal, so the \"Choose one\" Maneuver is a legal action (SVY-R29)")
	}
	b.s.scriptModal(0, nil, nil, []Answer{modeAnswer(0)}, pick(b.x))
	b.s.do(enactAction(inst))
	for _, r := range modeRecords(b.s.g) {
		if len(r.Options) != 1 {
			t.Errorf("with one unit only the kill mode is offered, got %d options %v (SVY-R34, SVY-R28)", len(r.Options), r.Options)
		}
	}
	if b.s.alive(b.x) {
		t.Errorf("the only legal mode, kill, was the one taken and X died (SVY-R34)")
	}
	if len(nonCombatDamage(b.s.g.Log)) != 0 {
		t.Errorf("the \"another\" mode was not legal and did not resolve (SVY-R34)")
	}

	b2, inst2 := build(2)
	b2.s.scriptModal(0, nil, nil, []Answer{modeAnswer(1)}, pick(b2.x), pick(b2.y))
	b2.s.do(enactAction(inst2))
	recs := modeRecords(b2.s.g)
	if len(recs) != 1 || len(recs[0].Options) != 2 {
		t.Fatalf("with two units both modes are legal and offered: mode decisions %+v (SVY-R34, FR-MOD-001 sec. 5)", recs)
	}
	wantDamageFrom(t, b2.s.g, b2.y, b2.x, 4, "the \"another\" mode was chosen and resolved")
}

func TestFR_TGT_002_AnotherSlotExcludesEarlierPick(t *testing.T) {
	// Covers: SVY-R34, FR-MOD-001
	b := newDealerBoard(t)
	inst := b.s.give(0, maneuver("Ambush", 0, 0, anotherEffect()))
	b.s.scriptModal(0, nil, nil, nil, pick(b.y), pick(b.z))
	b.s.do(enactAction(inst))
	asks := anotherTargetAsks(b.s)
	if len(asks) != 2 {
		t.Fatalf("both slots are asked with three units, got %d target decisions (SVY-R34)", len(asks))
	}
	if len(asks[0].Options) != 3 {
		t.Errorf("the first slot is offered every legal unit, got %d options (SVY-R34)", len(asks[0].Options))
	}
	if len(asks[1].Options) != 2 || !anotherOffers(asks[1], b.x) || !anotherOffers(asks[1], b.z) || anotherOffers(asks[1], b.y) {
		t.Errorf("the \"another\" slot is offered X and Z but not the earlier pick Y: got options %v (SVY-R34)", labels(asks[1]))
	}
	wantDamageFrom(t, b.s.g, b.z, b.y, 1, "Y, picked first, deals its power to Z, picked for the \"another\" slot")
	noFormatErrors(t, b.s.g)
}

func TestFR_MOD_001_AnotherSlotTimeoutTakesNextDifferent(t *testing.T) {
	// Covers: SVY-R34, SVY-R16, FR-TGT-002
	b := newDealerBoard(t)
	inst := b.s.give(0, maneuver("Ambush", 0, 0, anotherEffect()))
	b.s.scriptModal(0, nil, nil, nil, timeoutPick, timeoutPick)
	b.s.do(enactAction(inst))
	wantDamageFrom(t, b.s.g, b.y, b.x, 4, "a timeout takes the first unit in fixed order (X) for the dealer slot and the next different one (Y) for the \"another\" slot")
	if got := len(nonCombatDamage(b.s.g.Log)); got != 1 {
		t.Errorf("exactly one damage event, got %d (SVY-R34)", got)
	}
	b.s.wantDamage(b.x, 0)
	noFormatErrors(t, b.s.g)
}

func TestFR_TGT_002_AnotherSlotSingleRemainingOptionIsTakenUnasked(t *testing.T) {
	// Covers: SVY-R34, FR-MOD-001
	s := newScene(t)
	x := s.put(1, ZoneMainTheater, unit("X", 4, 30, Keywords{}))
	y := s.put(1, ZoneMainTheater, unit("Y", 1, 30, Keywords{}))
	inst := s.give(0, maneuver("Ambush", 0, 0, anotherEffect()))
	s.scriptModal(0, nil, nil, nil, pick(y))
	s.do(enactAction(inst))
	recs := anotherTargetRecords(s.g)
	if len(recs) != 2 {
		t.Fatalf("two target decisions are recorded, got %+v", recs)
	}
	if recs[1].How != "auto" {
		t.Errorf("the \"another\" slot has a single remaining option and is taken unasked, recorded %q (SVY-R34, FR-MOD-001 sec. 1)", recs[1].How)
	}
	wantDamageFrom(t, s.g, x, y, 1, "Y deals its power to the other unit, X")
}

func TestFR_TGT_002_AnotherSlotInLaterModeFollowsItsOwnClause(t *testing.T) {
	// Covers: SVY-R34, SVY-R35, SVY-R25, FR-MOD-001
	b := newDealerBoard(t)
	inst := b.s.give(0, maneuver("Ambush", 0, 0, ability.Effect{
		Modes: []ability.Effect{anotherKillMode(), anotherEffect()}, ChooseN: 2}))
	b.s.scriptModal(0, nil, nil, []Answer{modeAnswer(0, 1)}, pick(b.x), pick(b.y), pick(b.x))
	b.s.do(enactAction(inst))
	asks := anotherTargetAsks(b.s)
	if len(asks) != 3 {
		t.Fatalf("three target decisions (kill, dealer, \"another\"), got %d (SVY-R25)", len(asks))
	}
	if len(asks[2].Options) != 2 || !anotherOffers(asks[2], b.x) || !anotherOffers(asks[2], b.z) || anotherOffers(asks[2], b.y) {
		t.Errorf("the \"another\" slot differs from its own clause's dealer pick Y, not from the kill mode's pick X: options %v (SVY-R34, SVY-R35)", labels(asks[2]))
	}
}

func TestFR_MOD_001_AnotherSlotDoesNotChangeTheCardDefinition(t *testing.T) {
	// Covers: SVY-R34, FR-TGT-002
	b := newDealerBoard(t)
	def := ability.Effect{Modes: []ability.Effect{anotherKillMode(), anotherEffect()}, ChooseN: 2}
	inst := b.s.give(0, maneuver("Ambush", 0, 0, def))
	b.s.scriptModal(0, nil, nil, []Answer{modeAnswer(0, 1)}, pick(b.x), pick(b.y), pick(b.z))
	b.s.do(enactAction(inst))
	if len(anotherTargetAsks(b.s)) != 3 {
		t.Fatalf("precondition: both modes were chosen and all three slots asked")
	}
	if got := def.Modes[1].Targets[1].Another; got != 1 {
		t.Errorf("the dealer mode's second slot still has Another 1 after the join, got %d (the join copies, never edits, the definition)", got)
	}
	want := ability.Effect{Modes: []ability.Effect{anotherKillMode(), anotherEffect()}, ChooseN: 2}
	if !reflect.DeepEqual(def, want) {
		t.Errorf("the card's definition is unchanged by enacting it:\n got %+v\nwant %+v", def, want)
	}
	for i, m := range def.Modes {
		for _, tg := range m.Targets[len(m.Targets):cap(m.Targets)] {
			if !reflect.DeepEqual(tg, ability.Target{}) {
				t.Errorf("mode %d: a target was appended onto the card's own Targets slice: %+v", i+1, tg)
			}
		}
	}
}
