package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Shifting Impossibility (T-0098): "If an allied Scratch Company unit would be
// killed, return it to its owner's hand and deal 1 damage to this unit instead."
// Rulings: FR-REP-001 (inline substitution, consequential event isolation,
// replaced actions still count as done, SVY-R15; same-moment deaths and
// replacement within a death sweep, SVY-R20 to SVY-R22; SVY-R23 choosing
// between replacements) and FR-ZON-005 (a face-down Scratch Company returned to
// its owner's hand). The card is a 3/4 Horror unit. A Scratch Company is the
// face-down 1/1 from scene.putScratch; a sweeper keeps off Shifting Impossibility
// by naming Scratch in its filter, unless the test says otherwise.

var alliedScratch = ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Scratch: true}

// noEvent reports a failure if the log holds an event of one of the kinds aimed at asset a.
func noEvent(t *testing.T, s *scene, a *Asset, why string, kinds ...string) {
	t.Helper()
	for _, k := range kinds {
		if n := eventSeq(s, k, a.ID); n != 0 {
			t.Errorf("%s: %q event for %s at seq %d, want none (FR-REP-001)", why, k, a, n)
		}
	}
}

// Covers: FR-ZON-005
func TestCard_ShiftingImpossibility_ReturnsAScratchCompanyKilledAtTheSweep(t *testing.T) {
	s := newScene(t)
	si := s.put(0, ZoneReserves, s.card("Shifting Impossibility"))
	sc := s.putScratch(0)
	s.sweep(0, alliedScratch)
	if s.alive(sc) {
		t.Fatalf("the Scratch Company leaves the battlefield (FR-REP-001)")
	}
	if !inHand(s.g.Players[0], sc.Inst) {
		t.Errorf("an allied Scratch Company that would be killed is returned to its owner's hand (FR-REP-001, FR-ZON-005)")
	}
	if inDiscard(s.g.Players[0], sc.Inst) {
		t.Errorf("a replaced kill does not put the Scratch Company into the discard pile (FR-REP-001)")
	}
	if sc.Inst.Card.Name != "Filler" {
		t.Errorf("the card in hand is the Scratch Company's printed card, got %q (FR-ZON-005)", sc.Inst.Card.Name)
	}
	s.wantDamage(si, 1)
	wantReplaced(t, s, sc)
	noEvent(t, s, sc, "the replaced kill never occurs", "killed", "destroy")
	if !s.alive(si) {
		t.Errorf("Shifting Impossibility survives its 1 damage (4 Health)")
	}
}

func TestCard_ShiftingImpossibility_ReplacesOnlyAlliedScratchCompanies(t *testing.T) {
	s := newScene(t)
	si := s.put(0, ZoneReserves, s.card("Shifting Impossibility"))
	ally := s.put(0, ZoneReserves, unit("Ally", 1, 1, Keywords{}))
	foeSc := s.putScratch(1)
	// The sweep hits every other unit, Shifting Impossibility too, which takes
	// exactly that 1 damage and none from a replacement.
	s.sweep(0, ability.Filter{Kinds: ability.KindUnit, Other: true})
	if s.alive(ally) || !inDiscard(s.g.Players[0], ally.Inst) {
		t.Errorf("an allied unit that is not a Scratch Company is destroyed into the discard pile (FR-REP-001)")
	}
	if s.alive(foeSc) || !inDiscard(s.g.Players[1], foeSc.Inst) {
		t.Errorf("an opposing Scratch Company is not 'allied', so it is destroyed into its owner's discard pile (FR-REP-001)")
	}
	if len(replaceTargets(s)) != 0 {
		t.Errorf("no replacement applies to either death, got replace events for %v (FR-REP-001)", replaceTargets(s))
	}
	s.wantDamage(si, 1) // the sweep's own damage only
}

// Covers: SVY-R15, FR-MOD-001
func TestCard_ShiftingImpossibility_ReplacedKillStillCountsAsDone(t *testing.T) {
	s := newScene(t)
	si := s.put(0, ZoneReserves, s.card("Shifting Impossibility"))
	sc := s.putScratch(0)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 0, 5, Keywords{}))
	s.scriptPlan(0, nil, []Answer{yes}, pick(sc))
	deployApostle(s)
	if s.alive(sc) || !inHand(s.g.Players[0], sc.Inst) || inDiscard(s.g.Players[0], sc.Inst) {
		t.Errorf("the Apostle's kill of the Scratch Company is replaced: it returns to its owner's hand (FR-REP-001)")
	}
	s.wantDamage(si, 1)
	wantReplaced(t, s, sc)
	s.wantDamage(foe, 2) // the kill was carried out, so 'if you do' holds (SVY-R15)
	if d := damageTo(s.g.Log, foe); len(d) != 1 {
		t.Errorf("the Apostle deals its 2 damage to the chosen opposing unit exactly once: %d damage events (SVY-R15)", len(d))
	}
}

func TestCard_ShiftingImpossibility_ItsOwnLethalDamageStillReplacesEveryDeathInThePass(t *testing.T) {
	// SVY-R20: at 1 Current Health (3 damage of 4), the first byproduct leaves
	// it lethal, yet it still replaces both deaths and is destroyed afterwards.
	s := newScene(t)
	si := s.put(0, ZoneReserves, s.card("Shifting Impossibility"))
	si.Damage = 3
	sc1 := s.putScratch(0)
	sc2 := s.putScratch(0)
	s.sweep(0, alliedScratch)
	wantReplaced(t, s, sc1, sc2)
	for _, sc := range []*Asset{sc1, sc2} {
		if s.alive(sc) || !inHand(s.g.Players[0], sc.Inst) || inDiscard(s.g.Players[0], sc.Inst) {
			t.Errorf("%s is replaced and returns to its owner's hand although Shifting Impossibility is lethal (SVY-R20)", sc)
		}
	}
	if s.alive(si) || !inDiscard(s.g.Players[0], si.Inst) {
		t.Errorf("Shifting Impossibility is destroyed once the sweep's replaced deaths are done (SVY-R20)")
	}
	last := eventSeq(s, "leave", sc2.ID)
	if d := eventSeq(s, "destroy", si.ID); d == 0 || d < last {
		t.Errorf("Shifting Impossibility is destroyed (seq %d) only after the last replaced death (leave seq %d) (SVY-R20)", d, last)
	}
}

func TestCard_ShiftingImpossibility_TwoCopiesAskWhichOneReplaces(t *testing.T) {
	// SVY-R23: both copies apply to the one death, so its controller chooses.
	s := newScene(t)
	older := s.put(0, ZoneReserves, s.card("Shifting Impossibility"))
	newer := s.put(0, ZoneReserves, s.card("Shifting Impossibility"))
	sc := s.putScratch(0)
	s.script(0, Answer{Picks: []int{1}})
	s.sweep(0, alliedScratch)
	asks := replacementAsks(s, 0)
	if len(asks) != 1 {
		t.Fatalf("player 0 was asked %d replacement questions, want 1: two copies apply to the one death (SVY-R23)", len(asks))
	}
	wantOffered(t, asks[0], older.ID, newer.ID)
	if s.alive(sc) || !inHand(s.g.Players[0], sc.Inst) {
		t.Errorf("the Scratch Company's death is replaced and it returns to hand (FR-REP-001)")
	}
	if got := replaceSources(s); len(got) != 1 || got[0] != newer.ID {
		t.Errorf("replace events' sources = %v, want [%d]: the newer copy the controller picked (SVY-R23)", got, newer.ID)
	}
	s.wantDamage(newer, 1)
	s.wantDamage(older, 0)
}
