package engine

import (
	"slices"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0162: a Filter that asks for any of several printed subtypes, and
// for a Maneuver card in a pile (FR-TGT-001). Cards are synthetic. Pile scenes
// follow pile_target_test.go: player 0 deploys a unit "Source" whose deploy
// trigger has the slot under test, and player 1 is the opponent. Every
// exactly-one slot below has a legal entity on declaration, except the trigger
// clause of TestFR_TGT_001_ManeuverKindMatchesNoAsset, so no test pins the open
// questions of board directive #104.

// fsSub gives c the printed subtypes subs and returns it.
func fsSub(c *Card, subs ...string) *Card {
	c.Subtypes = subs
	return c
}

// fsPileSlot is one target slot over allied cards in a discard pile.
func fsPileSlot(kinds ability.Kinds, subs ...string) ability.Target {
	return ability.Target{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: kinds, Subtypes: subs, Zone: ability.ZoneDiscard}}
}

// fsFieldSlot is one target slot over opposing assets on the battlefield.
func fsFieldSlot(kinds ability.Kinds, sc ability.ScratchCompany, subs ...string) ability.Target {
	return ability.Target{Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: kinds, Scratch: sc, Subtypes: subs}}
}

// fsAsk runs the deploy trigger "return chosen <slot> to hand", the seat
// answering pick, and returns the single target question's option refs.
func (s *scene) fsAsk(slot ability.Target, pick int) []int {
	s.t.Helper()
	s.script(0, Answer{Picks: []int{pick}})
	s.iydDeploy(iydTrigger(false, []ability.Target{slot}, ptStep(ability.OpReturnToHand, 0)))
	asks := s.asked(0)
	if len(asks) != 1 {
		s.t.Fatalf("one target question, got %d decisions", len(asks))
	}
	return ppRefs(asks[0])
}

// fsKillAsk runs the deploy trigger "destroy chosen <slot>", the seat taking
// the first option, and returns the single target question's option refs.
func (s *scene) fsKillAsk(slot ability.Target) []int {
	s.t.Helper()
	s.script(0, Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, []ability.Target{slot}, ability.Step{Op: ability.OpKill, Subject: ability.SubjectTarget, Target: 0}))
	asks := s.asked(0)
	if len(asks) != 1 {
		s.t.Fatalf("one target question, got %d decisions", len(asks))
	}
	return ppRefs(asks[0])
}

func fsWantRefs(t *testing.T, got, want []int, why string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("%s: options got %v, want %v", why, got, want)
	}
}

func fsSorted(xs ...int) []int {
	slices.Sort(xs)
	return xs
}

// fsFivePile is the pile of the first two tests, bottom first: a Fungus card, a
// Plant card, a "Plant, Fungus" card, a card of another subtype and a card with none.
func fsFivePile(s *scene) []*Instance {
	return s.inDiscard(0,
		fsSub(unit("Fungus", 1, 1, Keywords{}), "Fungus"),
		fsSub(unit("Plant", 1, 1, Keywords{}), "Plant"),
		fsSub(unit("Both", 1, 1, Keywords{}), "Plant", "Fungus"),
		fsSub(unit("Other", 1, 1, Keywords{}), "Insect"),
		unit("Bare", 1, 1, Keywords{}))
}

func TestFR_TGT_001_SubtypesMatchAnyOfTheList(t *testing.T) {
	s := newScene(t)
	p := fsFivePile(s)
	got := s.fsAsk(fsPileSlot(0, "Fungus", "Plant"), 2)
	fsWantRefs(t, got, ppInstIDs(p[0], p[1], p[2]), "a card carrying at least one listed subtype is listed, any other is not (FR-TGT-001)")
	pl := s.g.Players[0]
	if !inHand(pl, p[2]) || inDiscardPile(pl, p[2]) {
		t.Errorf("the chosen card goes to hand")
	}
	for _, i := range []int{0, 1, 3, 4} {
		if inHand(pl, p[i]) || !inDiscardPile(pl, p[i]) {
			t.Errorf("card %d was not chosen and stays in the discard pile", i)
		}
	}
}

func TestFR_TGT_001_EmptySubtypesIsNoCriterion(t *testing.T) {
	s := newScene(t)
	p := fsFivePile(s)
	got := s.fsAsk(fsPileSlot(0), 0)
	fsWantRefs(t, got, ppInstIDs(p...), "with no subtype listed every card matches (FR-TGT-001)")
}

func TestFR_TGT_001_SubtypesAndKindsAreAConjunction(t *testing.T) {
	s := newScene(t)
	p := s.inDiscard(0,
		fsSub(unit("Fungus Unit", 1, 1, Keywords{}), "Fungus"),
		fsSub(structure("Fungus Wall", 0, 3, Keywords{}), "Fungus"),
		fsSub(unit("Fungus Unit 2", 1, 1, Keywords{}), "Fungus"))
	got := s.fsAsk(fsPileSlot(ability.KindUnit, "Fungus"), 0)
	fsWantRefs(t, got, ppInstIDs(p[0], p[2]), "a Fungus structure fails the unit criterion, a Fungus unit passes both (FR-TGT-001)")
}

func TestFR_TGT_001_SubtypesAndAllegianceAreAConjunction(t *testing.T) {
	s := newScene(t)
	mine := s.inDiscard(0, fsSub(unit("Mine 1", 1, 1, Keywords{}), "Fungus"), fsSub(unit("Mine 2", 1, 1, Keywords{}), "Fungus"))
	s.inDiscard(1, fsSub(unit("Theirs", 1, 1, Keywords{}), "Fungus"))
	got := s.fsAsk(fsPileSlot(0, "Fungus"), 0)
	fsWantRefs(t, got, ppInstIDs(mine...), "a Fungus card in the opponent's pile fails the allied criterion (FR-TGT-001)")
}

func TestFR_TGT_001_SubtypesAndScratchAreAConjunction(t *testing.T) {
	const sc = "Scratch Company"
	tests := []struct {
		name    string
		sc      ability.ScratchCompany
		sub     string
		matches []*Card
		refused []*Card
	}{
		{"infantry and not Scratch", ability.ScratchExcluded, "Infantry",
			[]*Card{fsSub(unit("Inf A", 1, 1, Keywords{}), "Infantry"), fsSub(unit("Inf B", 1, 1, Keywords{}), "Infantry")},
			[]*Card{fsSub(unit("Scratch Inf", 1, 1, Keywords{}), "Infantry", sc), fsSub(unit("Not Inf", 1, 1, Keywords{}), "Fungus")}},
		{"fungus and Scratch", ability.ScratchOnly, "Fungus",
			[]*Card{fsSub(unit("Scratch Fungus A", 1, 1, Keywords{}), sc, "Fungus"), fsSub(unit("Scratch Fungus B", 1, 1, Keywords{}), sc, "Fungus")},
			[]*Card{fsSub(unit("Scratch Bare", 1, 1, Keywords{}), sc), fsSub(unit("Fungus", 1, 1, Keywords{}), "Fungus")}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newScene(t)
			var want []int
			for _, c := range tc.matches {
				want = append(want, s.put(1, ZoneReserves, c).ID)
			}
			for _, c := range tc.refused {
				s.put(1, ZoneReserves, c)
			}
			got := s.fsKillAsk(fsFieldSlot(ability.KindUnit, tc.sc, tc.sub))
			fsWantRefs(t, fsSorted(got...), fsSorted(want...), "a unit must pass both the subtype list and the Scratch criterion (FR-TGT-001)")
		})
	}
}

func TestFR_TGT_001_SubtypesReadABattlefieldAsset(t *testing.T) {
	s := newScene(t)
	a := s.put(1, ZoneMainTheater, fsSub(unit("Fungus A", 1, 1, Keywords{}), "Fungus"))
	b := s.put(1, ZoneReserves, fsSub(unit("Fungus B", 1, 1, Keywords{}), "Fungus"))
	s.put(1, ZoneMainTheater, fsSub(unit("Other", 1, 1, Keywords{}), "Insect"))
	got := s.fsKillAsk(fsFieldSlot(ability.KindUnit, ability.ScratchAny, "Fungus"))
	fsWantRefs(t, fsSorted(got...), fsSorted(a.ID, b.ID), "a battlefield slot lists the units carrying the subtype and no other (FR-TGT-001)")
}

func TestFR_TGT_001_SubtypesReadAFaceDownAssetsFace(t *testing.T) {
	// Covers: FR-ZON-005
	s := newScene(t)
	a := s.put(1, ZoneMainTheater, fsSub(unit("Fungus A", 1, 1, Keywords{}), "Fungus"))
	b := s.put(1, ZoneMainTheater, fsSub(unit("Fungus B", 1, 1, Keywords{}), "Fungus"))
	down := s.put(1, ZoneReserves, fsSub(unit("Hidden Fungus", 1, 1, Keywords{}), "Fungus"))
	down.FaceDown = true
	down.Exhausted = true
	got := s.fsKillAsk(fsFieldSlot(ability.KindUnit, ability.ScratchAny, "Fungus"))
	fsWantRefs(t, fsSorted(got...), fsSorted(a.ID, b.ID), "a face-down asset has only the Scratch Company face, so its card's Fungus subtype is not read (FR-ZON-005)")
}

func TestFR_TGT_001_PilePickReadsSubtypes(t *testing.T) {
	// Covers: SVY-R3
	s := newScene(t)
	p := fsFivePile(s)
	s.ppScript(Answer{Picks: []int{1}})
	pick := ppPick(ability.OpReturnToHand, ability.Filter{Allegiance: ability.AllegianceAllied, Subtypes: []string{"Fungus"}, Zone: ability.ZoneDiscard}, 0)
	s.iydDeploy(iydTrigger(false, nil, pick))
	asks := ppPicks(s)
	if len(asks) != 1 {
		t.Fatalf("one pick question, got %d", len(asks))
	}
	fsWantRefs(t, ppRefs(asks[0]), ppInstIDs(p[0], p[2]), "the pick offers the cards carrying Fungus only (FR-TGT-001, SVY-R3)")
	pl := s.g.Players[0]
	if !inHand(pl, p[2]) {
		t.Errorf("the card picked returns to hand")
	}
	if inHand(pl, p[0]) || !inDiscardPile(pl, p[0]) {
		t.Errorf("the card not picked stays in the discard pile")
	}
}

// fsManeuverPile is a pile, bottom first: a unit, a Maneuver, a structure, a unit.
func fsManeuverPile(s *scene) []*Instance {
	return s.inDiscard(0,
		unit("Unit", 1, 1, Keywords{}),
		maneuver("Move", 0, 0, steps()),
		structure("Wall", 0, 3, Keywords{}),
		unit("Unit 2", 1, 1, Keywords{}))
}

func TestFR_TGT_001_ManeuverKindListsOnlyManeuverCards(t *testing.T) {
	s := newScene(t)
	p := fsManeuverPile(s)
	// The Maneuver is the only legal card, so nothing is asked and it is chosen.
	s.script(0)
	s.iydDeploy(iydTrigger(false, []ability.Target{fsPileSlot(ability.KindManeuver)}, ptStep(ability.OpReturnToHand, 0)))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("only the Maneuver is a legal card, so nothing is asked: %d decisions (FR-MOD-001 sec. 1)", n)
	}
	pl := s.g.Players[0]
	if !inHand(pl, p[1]) || inDiscardPile(pl, p[1]) {
		t.Errorf("the Maneuver card goes to hand (FR-TGT-001)")
	}
	for _, i := range []int{0, 2, 3} {
		if inHand(pl, p[i]) || !inDiscardPile(pl, p[i]) {
			t.Errorf("card %d is not a Maneuver card and stays in the pile", i)
		}
	}
}

func TestFR_TGT_001_OtherKindsStillRefuseManeuvers(t *testing.T) {
	tests := []struct {
		name  string
		kinds ability.Kinds
		want  []int // indexes into the pile
	}{
		{"unit", ability.KindUnit, []int{0, 3}},
		{"unit or structure", ability.KindUnit | ability.KindStructure, []int{0, 2, 3}},
		{"any", 0, []int{0, 1, 2, 3}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newScene(t)
			p := fsManeuverPile(s)
			var want []int
			for _, i := range tc.want {
				want = append(want, p[i].ID)
			}
			got := s.fsAsk(fsPileSlot(tc.kinds), 0)
			fsWantRefs(t, got, want, "a Maneuver is listed only when Kinds is empty (FR-TGT-001)")
		})
	}
}

func TestFR_TGT_001_ManeuverKindMatchesNoAsset(t *testing.T) {
	// Covers: FR-TGT-003
	s := newScene(t)
	mine := s.put(0, ZoneMainTheater, unit("Mine", 1, 1, Keywords{}))
	theirs := s.put(1, ZoneMainTheater, unit("Theirs", 1, 1, Keywords{}))
	wall := s.put(1, ZoneMainTheater, structure("Wall", 0, 3, Keywords{}))
	s.script(0)
	slot := ability.Target{Filter: ability.Filter{Kinds: ability.KindManeuver}}
	kill := ability.Step{Op: ability.OpKill, Subject: ability.SubjectTarget, Target: 0}
	s.iydDeploy(iydTrigger(false, []ability.Target{slot}, kill, scratchN(1)))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("no asset is a Maneuver, so the slot has no legal entity and nothing is asked: %d decisions", n)
	}
	s.iydWantAlive("a Maneuver slot matches no asset, so its clause does nothing (FR-TGT-003)", mine, theirs, wall)
	if n := faceDownCount(s.g.Players[0]); n != 1 {
		t.Errorf("the rest of the ability still resolves, so Scratch happens: %d Scratch Companies, want 1", n)
	}
}
