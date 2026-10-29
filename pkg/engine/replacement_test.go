package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for destruction replacement at the death sweep (T-0096, board
// directive #73 part 1; T-0101, directive #74): FR-REP-001 (inline
// substitution, consequential event isolation, sequential per-entity death
// replacement, SVY-R20 to SVY-R22 same-moment deaths, fixed order and sweep
// passes), FR-LAY-002 (the death sweep) and FR-ZON-004.
//
// Cards are synthetic. A "guardian" prints: if another allied unit would be
// destroyed, return it to its owner's hand and deal 1 damage to this asset
// instead. Its Maximum Health sets how many deaths it can absorb. Most tests
// use a structure guardian so a unit sweeper never hits it. A "sweeper" unit's
// deploy trigger deals 1 damage to the units named by its filter, which leaves
// every 1-Health unit lethal at the same sweep.

var alliedUnits = ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Other: true}

func guardianReplacement() ability.Replacement {
	return ability.Replacement{
		Event:  ability.WouldBeDestroyed,
		Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Other: true},
		Steps: []ability.Step{
			{Op: ability.OpReturnToHand, Subject: ability.SubjectReplaced},
			{Op: ability.OpDealDamage, Amount: ability.Value{N: 1}, Subject: ability.SubjectSelf},
		},
	}
}

// guardianStructure is a structure guardian with the given Maximum Health and triggers.
func guardianStructure(health int, trigs ...ability.Trigger) *Card {
	c := structure("Guardian", 1, health, Keywords{})
	c.Ability = &ability.Def{Triggers: trigs, Replacements: []ability.Replacement{guardianReplacement()}}
	return c
}

// guardianUnit is a unit guardian with the given Maximum Health.
func guardianUnit(health int) *Card {
	c := unit("Guardian Unit", 1, health, Keywords{})
	c.Ability = &ability.Def{Replacements: []ability.Replacement{guardianReplacement()}}
	return c
}

// sweep has player p deploy a sweeper that deals 1 damage to each unit matching f.
func (s *scene) sweep(p PlayerID, f ability.Filter) {
	s.t.Helper()
	s.deployFromHand(p, trigUnit("Sweeper", deployTrigger(allDamage(1, f))))
}

// eventSeq returns the Seq of the first event of the kind aimed at asset id, or 0.
func eventSeq(s *scene, kind string, id int) int {
	for _, e := range s.g.Log {
		if e.Kind == kind && e.Target == id {
			return e.Seq
		}
	}
	return 0
}

// replaceTargets returns the replaced asset IDs of the "replace" events, in log order.
func replaceTargets(s *scene) []int {
	var out []int
	for _, e := range eventsOfKind(s.g.Log, "replace") {
		out = append(out, e.Target)
	}
	return out
}

func wantReplaced(t *testing.T, s *scene, want ...*Asset) {
	t.Helper()
	got := replaceTargets(s)
	if len(got) != len(want) {
		t.Fatalf("replaced deaths (replace events' targets) = %v, want %d of them (FR-REP-001, SVY-R20, SVY-R22)", got, len(want))
	}
	for i, a := range want {
		if got[i] != a.ID {
			t.Errorf("replaced death %d = asset %d, want %s (%d) (FR-REP-001, SVY-R20, SVY-R22)", i+1, got[i], a, a.ID)
		}
	}
}

func TestFR_REP_001_ReplacedDeathReturnsTheUnitToItsOwnersHand(t *testing.T) {
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, guardianStructure(3))
	u := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	s.sweep(0, alliedUnits)
	if s.alive(u) {
		t.Errorf("the replaced unit should have left the battlefield (FR-REP-001)")
	}
	if !inHand(s.g.Players[0], u.Inst) {
		t.Errorf("a unit whose destruction is replaced is returned to its owner's hand (FR-REP-001)")
	}
	if inDiscard(s.g.Players[0], u.Inst) {
		t.Errorf("a replaced death does not put the unit into the discard pile (FR-REP-001)")
	}
	s.wantDamage(src, 1)
	wantReplaced(t, s, u)
	if eventSeq(s, "leave", u.ID) == 0 {
		t.Errorf("returning the unit to hand logs a leave event for it (FR-REP-001)")
	}
	if n := eventSeq(s, "destroy", u.ID); n != 0 {
		t.Errorf("the replaced destruction never occurs, so no destroy event for the unit (FR-REP-001), got seq %d", n)
	}
}

func TestFR_REP_001_ReplacedDeathFiresNoDeathTrigger(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, guardianStructure(3))
	u := s.put(0, ZoneReserves, mortal("Martyr", 1, 1, killedTrigger(drawN(1))))
	deck := len(s.g.Players[0].Deck)
	s.sweep(0, alliedUnits)
	if s.alive(u) || !inHand(s.g.Players[0], u.Inst) {
		t.Fatalf("precondition: the unit's death is replaced and it is in its owner's hand")
	}
	if got := len(s.g.Players[0].Deck); got != deck {
		t.Errorf("the replaced event never occurs, so the unit's death trigger does not fire (no draw): deck %d -> %d (FR-REP-001)", deck, got)
	}
}

func TestFR_REP_001_ReplacementDoesNotApplyOutsideItsFilter(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, guardianStructure(3))
	foe := s.put(1, ZoneReserves, unit("Foe", 1, 1, Keywords{}))
	s.sweep(0, ability.Filter{Kinds: ability.KindUnit, Other: true})
	if s.alive(foe) {
		t.Fatalf("precondition: the opposing unit is lethal and leaves the battlefield")
	}
	if !inDiscard(s.g.Players[1], foe.Inst) {
		t.Errorf("an opposing unit is outside the allied filter, so it is destroyed into the discard pile (FR-REP-001)")
	}
	if len(replaceTargets(s)) != 0 {
		t.Errorf("no replacement applies to an opposing unit, got replace events for %v (FR-REP-001)", replaceTargets(s))
	}
}

// threeVictims builds a guardian structure of the given Health and three allied
// 1-Health units, oldest first. It returns the guardian and the units.
func threeVictims(s *scene, health int) (*Asset, [3]*Asset) {
	src := s.put(0, ZoneMainTheater, guardianStructure(health))
	var us [3]*Asset
	for i := range us {
		us[i] = s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	}
	return src, us
}

func TestFR_REP_001_SourceLethalInTheSameSweepReplacesEveryDeath(t *testing.T) {
	// SVY-R20 (FR-REP-001): a replacement source that is itself lethal in the same sweep stays on the
	// battlefield until the sweep's unreplaced deaths happen, so it replaces every
	// death it applies to and is then destroyed as one of the sweep's own deaths.
	// One "all allied units" damage step leaves the guardian (1 Health) and three
	// allied units lethal.
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, guardianUnit(1))
	var us [3]*Asset
	for i := range us {
		us[i] = s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	}
	s.sweep(0, alliedUnits)
	wantReplaced(t, s, us[0], us[1], us[2])
	for _, u := range us {
		if !inHand(s.g.Players[0], u.Inst) || inDiscard(s.g.Players[0], u.Inst) {
			t.Errorf("%s is replaced and returns to hand although the guardian is lethal too (SVY-R20)", u)
		}
	}
	if s.alive(src) || !inDiscard(s.g.Players[0], src.Inst) {
		t.Errorf("the lethal guardian is destroyed after replacing every death (SVY-R20)")
	}
	last := eventSeq(s, "leave", us[2].ID)
	if d := eventSeq(s, "destroy", src.ID); d == 0 || d < last {
		t.Errorf("the guardian is destroyed (seq %d) only after the last replaced death (seq %d) (SVY-R20)", d, last)
	}
}

func TestFR_REP_001_ByproductDamageFiresTheSourcesDamageTrigger(t *testing.T) {
	// FR-REP-001 Consequential Event Isolation: the byproduct damage is a separate
	// child event, so the guardian's Unbowed trigger fires, and it is placed after
	// the sweep (FR-TRG-004).
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, guardianStructure(3, unbowed(nil, healSelf(1))))
	u := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	s.sweep(0, alliedUnits)
	if s.alive(u) || !inHand(s.g.Players[0], u.Inst) {
		t.Fatalf("precondition: the unit's death is replaced")
	}
	s.wantDamage(src, 0)
	placed, left := 0, eventSeq(s, "leave", u.ID)
	for _, e := range s.g.Log {
		if e.Kind == "place" && e.Source == src.ID {
			placed = e.Seq
		}
	}
	if placed == 0 {
		t.Fatalf("the byproduct damage should fire the guardian's damage trigger (heals it back to 0 damage) (FR-REP-001)")
	}
	if placed < left {
		t.Errorf("the trigger is placed after the sweep (seq %d), not during the replaced death (leave at seq %d) (FR-TRG-004)", placed, left)
	}
}

func TestFR_ZON_004_FaceDownUnitReturnedToHandIsItsPrintedCard(t *testing.T) {
	// A face-down unit that leaves the battlefield for the hand reverts to its
	// printed card, and the leave event does not reveal a card name.
	s := newScene(t)
	s.put(0, ZoneMainTheater, guardianStructure(3))
	sc := s.putScratch(0)
	s.sweep(0, alliedUnits)
	if s.alive(sc) || !inHand(s.g.Players[0], sc.Inst) {
		t.Fatalf("precondition: the Scratch Company's death is replaced and it is in its owner's hand")
	}
	if sc.Inst.Card.Name != "Filler" {
		t.Errorf("the unit in hand is its printed card instance, got %q (FR-ZON-004, FR-ZON-005)", sc.Inst.Card.Name)
	}
	var seen bool
	for _, e := range eventsOfKind(s.g.Log, "leave") {
		if e.Target == sc.ID {
			seen = true
			if e.Card != "" {
				t.Errorf("a face-down card's name is private in the hand, but the leave event carries Card %q (FR-ZON-004)", e.Card)
			}
			if e.Zone != ZoneHand.String() {
				t.Errorf("the leave event names the destination zone %q, got %q (FR-ZON-004)", ZoneHand, e.Zone)
			}
		}
	}
	if !seen {
		t.Errorf("returning the face-down unit logs a leave event (FR-ZON-004)")
	}
}

func noOrderAsked(t *testing.T, s *scene, p PlayerID) {
	t.Helper()
	for _, d := range s.asked(p) {
		if d.Kind == DecideOrder {
			t.Errorf("replaced deaths are not ordered by the player, but player %d was asked %q (SVY-R20, SVY-R22)", p, d.Prompt)
		}
	}
}

func TestFR_REP_001_ReplacedDeathsAreNotOrderedByThePlayer(t *testing.T) {
	// SVY-R20. The guardian has 2 Health, so three byproducts of 1 damage leave
	// it lethal; the replaced deaths are still all replaced, oldest first, and
	// player 0's scripted order answer is never asked for.
	s := newScene(t)
	src, us := threeVictims(s, 2)
	s.script(0, Answer{Picks: []int{2, 0, 1}})
	s.sweep(0, alliedUnits)
	wantReplaced(t, s, us[0], us[1], us[2])
	for _, u := range us {
		if !inHand(s.g.Players[0], u.Inst) || inDiscard(s.g.Players[0], u.Inst) {
			t.Errorf("%s was replaced and returns to its owner's hand (FR-REP-001, SVY-R20)", u)
		}
	}
	noOrderAsked(t, s, 0)
	if s.alive(src) {
		t.Errorf("the guardian, left lethal by its three byproducts, is destroyed (SVY-R20)")
	}
	last := eventSeq(s, "leave", us[2].ID)
	if d := eventSeq(s, "destroy", src.ID); d == 0 || d < last {
		t.Errorf("the guardian is destroyed (seq %d) only after the third replaced death (leave seq %d) (SVY-R20)", d, last)
	}
}

func TestFR_REP_001_ByproductLethalSourceStillReplacesEveryDeathInThePass(t *testing.T) {
	// SVY-R20, FRD 01's second example: the guardian (4 Health) is not lethal
	// when the sweep starts, but five byproducts leave it at -1 Current Health.
	// All five deaths are replaced; it is destroyed in the next pass.
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, guardianStructure(4))
	var us [5]*Asset
	for i := range us {
		us[i] = s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	}
	s.sweep(0, alliedUnits)
	wantReplaced(t, s, us[0], us[1], us[2], us[3], us[4])
	for _, u := range us {
		if s.alive(u) || !inHand(s.g.Players[0], u.Inst) || inDiscard(s.g.Players[0], u.Inst) {
			t.Errorf("%s is replaced and returns to hand, never to the discard pile, although a byproduct left the source lethal (SVY-R20)", u)
		}
	}
	if s.alive(src) || !inDiscard(s.g.Players[0], src.Inst) {
		t.Errorf("the guardian, left at -1 Current Health, is destroyed in the next pass (SVY-R20, FR-LAY-002)")
	}
	last := eventSeq(s, "leave", us[4].ID)
	if d := eventSeq(s, "destroy", src.ID); d == 0 || d < last {
		t.Errorf("the guardian is destroyed (seq %d) only after the fifth replaced death (leave seq %d) (SVY-R20)", d, last)
	}
}

func TestFR_REP_001_ActivePlayersReplacedDeathsGoFirst(t *testing.T) {
	// SVY-R22. Player 1's units are older than player 0's, yet player 0, the
	// active player, has their deaths replaced first.
	s := newScene(t)
	s.put(0, ZoneMainTheater, guardianStructure(3))
	s.put(1, ZoneMainTheater, guardianStructure(3))
	var p1, p0 [2]*Asset
	for i := range p1 {
		p1[i] = s.put(1, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	}
	for i := range p0 {
		p0[i] = s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	}
	s.script(0)
	s.script(1)
	s.sweep(0, ability.Filter{Kinds: ability.KindUnit, Other: true})
	wantReplaced(t, s, p0[0], p0[1], p1[0], p1[1])
	for _, u := range append(p0[:], p1[:]...) {
		if s.alive(u) || !inHand(s.g.Players[u.Inst.Owner], u.Inst) {
			t.Errorf("%s was replaced and returns to its owner's hand (FR-REP-001, SVY-R22)", u)
		}
	}
	noOrderAsked(t, s, 0)
	noOrderAsked(t, s, 1)
}

// replacementAsks returns the DecideReplacement questions (kind "Replacement")
// put to player p, in order. It compares the kind's string so the helper reads
// the decision events' kind the way replay does.
func replacementAsks(s *scene, p PlayerID) []Decision {
	var out []Decision
	for _, d := range s.asked(p) {
		if d.Kind.String() == "Replacement" {
			out = append(out, d)
		}
	}
	return out
}

// replaceSources returns the Source asset IDs of the "replace" events, in log order.
func replaceSources(s *scene) []int {
	var out []int
	for _, e := range eventsOfKind(s.g.Log, "replace") {
		out = append(out, e.Source)
	}
	return out
}

// wantOffered checks that decision d offers exactly the sources with the given IDs, in order.
func wantOffered(t *testing.T, d Decision, ids ...int) {
	t.Helper()
	if len(d.Options) != len(ids) {
		t.Fatalf("replacement question %q offers %d options, want %d (one per applicable replacement) (FR-REP-001, SVY-R23)", d.Prompt, len(d.Options), len(ids))
	}
	for i, id := range ids {
		if d.Options[i].Ref != id {
			t.Errorf("option %d of %q refers to asset %d, want %d (SVY-R11, SVY-R23)", i+1, d.Prompt, d.Options[i].Ref, id)
		}
	}
}

func TestFR_REP_001_ControllerChoosesBetweenReplacementsForEachDeath(t *testing.T) {
	// SVY-R23: with two replacements applying, the dying asset's controller
	// chooses, death by death.
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, guardianStructure(4))
	b := s.put(0, ZoneMainTheater, guardianStructure(4))
	u1 := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	u2 := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	s.script(0, Answer{Picks: []int{0}}, Answer{Picks: []int{1}})
	s.sweep(0, alliedUnits)
	asks := replacementAsks(s, 0)
	if len(asks) != 2 {
		t.Fatalf("player 0 was asked %d replacement questions, want 2, one per death (SVY-R23)", len(asks))
	}
	for _, d := range asks {
		wantOffered(t, d, a.ID, b.ID)
	}
	for _, u := range []*Asset{u1, u2} {
		if s.alive(u) || !inHand(s.g.Players[0], u.Inst) || inDiscard(s.g.Players[0], u.Inst) {
			t.Errorf("%s is replaced and returns to its owner's hand (FR-REP-001)", u)
		}
	}
	wantReplaced(t, s, u1, u2)
	if got := replaceSources(s); len(got) != 2 || got[0] != a.ID || got[1] != b.ID {
		t.Errorf("replace events' sources = %v, want [%d %d]: the first death's chosen replacement is A's, the second's is B's (SVY-R23)", got, a.ID, b.ID)
	}
	s.wantDamage(a, 1)
	s.wantDamage(b, 1)
}

func TestFR_REP_001_ReplacementChoiceTimesOutToFixedOrder(t *testing.T) {
	// SVY-R23, SVY-R16: a timeout picks the first source in FR-MOD-001 §6's
	// fixed order, Reserves before the Main Theater, though the Main Theater
	// guardian is older. The guardians are allied units too, so the sweep hits
	// each for 1; they have Health 5 to survive that and two byproducts.
	s := newScene(t)
	old := s.put(0, ZoneMainTheater, guardianUnit(5))
	newer := s.put(0, ZoneReserves, guardianUnit(5))
	u1 := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	u2 := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	s.script(0, Answer{Timeout: true}, Answer{Timeout: true})
	s.sweep(0, alliedUnits)
	if n := len(replacementAsks(s, 0)); n != 2 {
		t.Fatalf("player 0 was asked %d replacement questions, want 2 (SVY-R23)", n)
	}
	wantReplaced(t, s, u1, u2)
	if got := replaceSources(s); len(got) != 2 || got[0] != newer.ID || got[1] != newer.ID {
		t.Errorf("replace events' sources = %v, want both %d: a timeout picks the Reserves guardian first in §6's fixed order (SVY-R23, SVY-R16)", got, newer.ID)
	}
	s.wantDamage(newer, 3) // 1 from the sweep, 2 byproducts
	s.wantDamage(old, 1)   // 1 from the sweep only
}

func TestFR_REP_001_ReplacementChoiceTimesOutToPrintedOrderWithinASource(t *testing.T) {
	// SVY-R23: one source prints two applicable replacements; a timeout
	// applies the first printed one (return to hand), not the second (deck).
	s := newScene(t)
	src := structure("Twin Guardian", 1, 4, Keywords{})
	toHand := ability.Replacement{Event: ability.WouldBeDestroyed, Filter: alliedUnits,
		Steps: []ability.Step{{Op: ability.OpReturnToHand, Subject: ability.SubjectReplaced}}}
	toDeck := ability.Replacement{Event: ability.WouldBeDestroyed, Filter: alliedUnits,
		Steps: []ability.Step{{Op: ability.OpPutOnDeck, Subject: ability.SubjectReplaced}}}
	src.Ability = &ability.Def{Replacements: []ability.Replacement{toHand, toDeck}}
	g := s.put(0, ZoneMainTheater, src)
	u := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	s.script(0, Answer{Timeout: true})
	s.sweep(0, alliedUnits)
	asks := replacementAsks(s, 0)
	if len(asks) != 1 {
		t.Fatalf("player 0 was asked %d replacement questions, want 1: two replacements apply to the one death (SVY-R23)", len(asks))
	}
	wantOffered(t, asks[0], g.ID, g.ID)
	wantReplaced(t, s, u)
	if s.alive(u) || !inHand(s.g.Players[0], u.Inst) {
		t.Errorf("a timeout applies the source's first printed replacement, returning the unit to hand (SVY-R23)")
	}
	for _, in := range s.g.Players[0].Deck {
		if in == u.Inst {
			t.Errorf("the second printed replacement (put on deck) applied, want the first (SVY-R23)")
		}
	}
}

func TestFR_REP_001_OneApplicableReplacementAsksNothing(t *testing.T) {
	// SVY-R23: a choice only exists when several replacements apply.
	s := newScene(t)
	g := s.put(0, ZoneMainTheater, guardianStructure(3))
	u := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	s.script(0, Answer{Picks: []int{1}})
	s.sweep(0, alliedUnits)
	for p := PlayerID(0); p < 2; p++ {
		if n := len(replacementAsks(s, p)); n != 0 {
			t.Errorf("player %d was asked %d replacement questions, want none: only one replacement applies (SVY-R23)", p, n)
		}
	}
	wantReplaced(t, s, u)
	if got := replaceSources(s); len(got) != 1 || got[0] != g.ID {
		t.Errorf("replace events' sources = %v, want [%d] (FR-REP-001)", got, g.ID)
	}
	if s.alive(u) || !inHand(s.g.Players[0], u.Inst) {
		t.Errorf("the death is replaced and the unit returns to hand (FR-REP-001)")
	}
}

func TestFR_REP_001_LethalSourceIsStillAChoice(t *testing.T) {
	// SVY-R20, SVY-R23: a lethal source still applies until the sweep destroys
	// it, so it is offered. The unit is oldest, so its death is replaced first.
	s := newScene(t)
	u := s.put(0, ZoneMainTheater, unit("Victim", 1, 1, Keywords{}))
	a := s.put(0, ZoneMainTheater, guardianUnit(1))
	b := s.put(0, ZoneReserves, guardianUnit(4))
	s.script(0, Answer{Picks: []int{0}})
	s.sweep(0, alliedUnits)
	asks := replacementAsks(s, 0)
	if len(asks) != 1 {
		t.Fatalf("player 0 was asked %d replacement questions, want 1: only the unit's death has two replacements (SVY-R23)", len(asks))
	}
	wantOffered(t, asks[0], a.ID, b.ID)
	wantReplaced(t, s, u, a)
	if got := replaceSources(s); len(got) != 2 || got[0] != a.ID || got[1] != b.ID {
		t.Errorf("replace events' sources = %v, want [%d %d]: the unit's chosen lethal guardian A, then B for A's own death (SVY-R20, SVY-R23)", got, a.ID, b.ID)
	}
	for _, x := range []*Asset{u, a} {
		if s.alive(x) || !inHand(s.g.Players[0], x.Inst) || inDiscard(s.g.Players[0], x.Inst) {
			t.Errorf("%s is replaced and returns to its owner's hand (SVY-R20)", x)
		}
	}
}
