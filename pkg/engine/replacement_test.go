package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for destruction replacement at the death sweep (T-0096, board
// directive #73 part 1): FR-REP-001 (inline substitution, consequential event
// isolation, sequential per-entity death replacement, SVY-R5 ordering and
// mid-sweep destruction), FR-LAY-002 (the death sweep) and FR-ZON-004.
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
		t.Fatalf("replaced deaths (replace events' targets) = %v, want %d of them (FR-REP-001, SVY-R5)", got, len(want))
	}
	for i, a := range want {
		if got[i] != a.ID {
			t.Errorf("replaced death %d = asset %d, want %s (%d) (FR-REP-001, SVY-R5)", i+1, got[i], a, a.ID)
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

func TestFR_REP_001_ControllerOrdersTheReplacedDeaths(t *testing.T) {
	// SVY-R5. The guardian has 2 Health, so it absorbs two deaths. The controller
	// picks the third unit, then the first. The second byproduct leaves the
	// guardian lethal, so it is destroyed before the unreplaced second unit.
	s := newScene(t)
	src, us := threeVictims(s, 2)
	s.script(0, Answer{Picks: []int{2, 0, 1}})
	s.sweep(0, alliedUnits)
	wantReplaced(t, s, us[2], us[0])
	for _, u := range []*Asset{us[0], us[2]} {
		if !inHand(s.g.Players[0], u.Inst) {
			t.Errorf("%s was replaced and returns to its owner's hand (FR-REP-001, SVY-R5)", u)
		}
	}
	if !inDiscard(s.g.Players[0], us[1].Inst) {
		t.Errorf("the unit left after the guardian is gone is destroyed normally (SVY-R5)")
	}
	if s.alive(src) {
		t.Errorf("the guardian is lethal after the second byproduct and is destroyed (SVY-R5)")
	}
	srcSeq, thirdSeq := eventSeq(s, "destroy", src.ID), eventSeq(s, "destroy", us[1].ID)
	if srcSeq == 0 || thirdSeq == 0 || srcSeq > thirdSeq {
		t.Errorf("the guardian is destroyed (seq %d) before the third death (seq %d) (SVY-R5, FR-LAY-002)", srcSeq, thirdSeq)
	}
	var orderAsks int
	for _, d := range s.asked(0) {
		if d.Kind == DecideOrder {
			orderAsks++
			if len(d.Options) != 3 {
				t.Errorf("the order decision has one option per queued death, want 3, got %d (SVY-R5)", len(d.Options))
			}
		}
	}
	if orderAsks != 1 {
		t.Errorf("the controller is asked once to order the replaced deaths, got %d order decisions (SVY-R5)", orderAsks)
	}
}

func TestFR_REP_001_ReplacedDeathOrderFallsBackToOldestFirst(t *testing.T) {
	// SVY-R5. The controller gives no answer: the two oldest units are replaced.
	s := newScene(t)
	_, us := threeVictims(s, 2)
	s.script(0, Answer{Timeout: true})
	s.sweep(0, alliedUnits)
	wantReplaced(t, s, us[0], us[1])
	if !inDiscard(s.g.Players[0], us[2].Inst) {
		t.Errorf("the newest unit is left after the guardian is destroyed and dies normally (SVY-R5)")
	}
}

func TestFR_REP_001_SourceLeftLethalByAByproductStopsApplying(t *testing.T) {
	// SVY-R5. The guardian has 1 Health: the first byproduct kills it, and the
	// sweep destroys it before the second death, which is no longer replaced.
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, guardianStructure(1))
	a := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	b := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	s.sweep(0, alliedUnits)
	wantReplaced(t, s, a)
	if !inHand(s.g.Players[0], a.Inst) {
		t.Errorf("the first death is replaced and the unit returns to hand (FR-REP-001)")
	}
	if s.alive(src) {
		t.Errorf("the byproduct damage leaves the guardian lethal and the sweep destroys it (SVY-R5)")
	}
	if !inDiscard(s.g.Players[0], b.Inst) || inHand(s.g.Players[0], b.Inst) {
		t.Errorf("the second unit dies after its replacement's source is destroyed, so it goes to the discard pile (SVY-R5)")
	}
	if ds, db := eventSeq(s, "destroy", src.ID), eventSeq(s, "destroy", b.ID); ds == 0 || db == 0 || ds > db {
		t.Errorf("the guardian is destroyed (seq %d) before the second unit (seq %d) (SVY-R5)", ds, db)
	}
}

func TestFR_REP_001_SourceLethalInTheSameSweepReplacesEveryDeath(t *testing.T) {
	// FR-REP-001 with the user ruling of 2026-10-29 (not yet in FRD 01): a
	// replacement source that is itself lethal in the same sweep stays on the
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
			t.Errorf("%s is replaced and returns to hand although the guardian is lethal too (user ruling 2026-10-29)", u)
		}
	}
	if s.alive(src) || !inDiscard(s.g.Players[0], src.Inst) {
		t.Errorf("the lethal guardian is destroyed after replacing every death (user ruling 2026-10-29)")
	}
	last := eventSeq(s, "leave", us[2].ID)
	if d := eventSeq(s, "destroy", src.ID); d == 0 || d < last {
		t.Errorf("the guardian is destroyed (seq %d) only after the last replaced death (seq %d) (user ruling 2026-10-29)", d, last)
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

func TestFR_REP_001_TheAffectedAssetsControllerOrdersTheirDeaths(t *testing.T) {
	// SVY-R5. Player 1's guardian and two of their units; player 0, the active
	// player, deploys the sweeper. Player 1 is asked, and orders the deaths.
	s := newScene(t)
	s.put(1, ZoneMainTheater, guardianStructure(3))
	a := s.put(1, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	b := s.put(1, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	s.script(0)
	s.script(1, Answer{Picks: []int{1, 0}})
	s.sweep(0, ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit})
	for _, d := range s.asked(0) {
		if d.Kind == DecideOrder {
			t.Errorf("the active player is not asked to order the other player's replaced deaths: %q (SVY-R5)", d.Prompt)
		}
	}
	var asks int
	for _, d := range s.asked(1) {
		if d.Kind == DecideOrder {
			asks++
			if d.Player != 1 || len(d.Options) != 2 {
				t.Errorf("the order decision is player 1's with one option per death, got player %d, %d options (SVY-R5)", d.Player, len(d.Options))
			}
		}
	}
	if asks != 1 {
		t.Fatalf("the controller of the replacements and the affected units is asked once, got %d order decisions (SVY-R5)", asks)
	}
	wantReplaced(t, s, b, a)
}
