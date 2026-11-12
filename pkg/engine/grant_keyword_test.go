package engine

import (
	"fmt"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0219 (FR-LAY-001, FR-LAY-003, FR-TRN-003a, FR-ENT-103, wave F):
// OpGrantKeyword gives its Subject a keyword until end of turn (a Layer 3
// modifier that expires in Sub-Phase 4c with the Layer 5 power modifiers,
// SVY-R39), and a later step of one resolution acts on the asset a pile slot's
// card became when an earlier step deployed it. Cards are synthetic. Player 0
// is the active player and enacts a free Maneuver.

var gkAlliedUnit = ability.Target{Filter: ability.Filter{
	Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}}

var gkPile = ability.Target{Filter: ability.Filter{
	Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}}

var gkDeploy = ability.Step{Op: ability.OpDeployFree, Subject: ability.SubjectTarget, To: ability.ZoneMainTheater}

// grantStep is "<subject slot> gains kw until end of turn".
func grantStep(slot int, kw ability.Keyword) ability.Step {
	return ability.Step{Op: ability.OpGrantKeyword, Subject: ability.SubjectTarget, Target: slot, Keyword: kw}
}

// grantAlliedManeuver is a free Maneuver whose one target is an allied unit,
// with the given steps.
func grantAlliedManeuver(ss ...ability.Step) *Card {
	return maneuver("Grant", 0, 0, ability.Effect{Targets: []ability.Target{gkAlliedUnit}, Steps: ss})
}

func gkEnact(s *scene, c *Card) {
	s.t.Helper()
	s.do(enactAction(s.give(0, c)))
}

// gkLater is a free Maneuver whose one target is a unit card in the
// controller's discard pile, with the given steps.
func gkLater(ss ...ability.Step) *Card {
	return maneuver("Trance", 0, 0, ability.Effect{Targets: []ability.Target{gkPile}, Steps: ss})
}

// gkLaterSteps is first, then Stalwart and +2 power on the same pile slot.
func gkLaterSteps(first ability.Step) []ability.Step {
	return []ability.Step{first, grantStep(0, ability.KeywordStalwart), modifyTarget(2, 0)}
}

func TestFR_LAY_001_GrantedKeywordIsEffective(t *testing.T) {
	// Covers: FR-KWD-002c
	s := newScene(t)
	u := s.put(0, ZoneMainTheater, unit("Rider", 1, 9, Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 3, 20, Keywords{}))
	gkEnact(s, grantAlliedManeuver(grantStep(0, ability.KeywordStalwart)))
	if !s.g.Keywords(u).Stalwart {
		t.Errorf("the granted unit effectively has Stalwart (FR-LAY-001 Layer 3)")
	}
	if u.Printed().Keywords.Stalwart {
		t.Errorf("a grant changes no printed keyword (FR-LAY-001)")
	}
	s.do(s.attack(u, foe))
	s.wantDamage(u, 1)
}

func TestFR_LAY_001_GrantedKeywordAddsToCornered(t *testing.T) {
	s := newScene(t)
	card := unit("Dancer", 2, 6, Keywords{})
	card.Ability = &ability.Def{Cornered: ability.CorneredKeywords{Resistant: 1}}
	u := s.put(0, ZoneMainTheater, card)
	s.put(1, ZoneMainTheater, unit("Foe", 1, 5, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Second Foe", 1, 5, Keywords{}))
	gkEnact(s, grantAlliedManeuver(grantStep(0, ability.KeywordStalwart)))
	if got, want := s.g.Keywords(u), (Keywords{Resistant: 1, Stalwart: true}); got != want {
		t.Errorf("a Cornered unit granted Stalwart has its Cornered keyword and Stalwart: %+v, want %+v (FR-LAY-001 Layer 3)", got, want)
	}
}

func TestFR_LAY_003_GrantedKeywordExpiresInSubPhase4c(t *testing.T) {
	// Covers: FR-TRN-001, SVY-R39
	s := newScene(t)
	u := s.put(0, ZoneMainTheater, unit("Rider", 2, 9, Keywords{}))
	gkEnact(s, grantAlliedManeuver(grantStep(0, ability.KeywordStalwart), modifyTarget(2, 0)))
	if !s.g.Keywords(u).Stalwart || s.g.Power(u) != 4 {
		t.Fatalf("until the turn ends the unit has Stalwart and power 4: Stalwart %v, power %d (SVY-R39)", s.g.Keywords(u).Stalwart, s.g.Power(u))
	}
	s.endTurn()
	if !s.alive(u) || u.Zone != ZoneMainTheater {
		t.Fatalf("the unit is still in the Main Theater after the turn ends: %v", u)
	}
	if s.g.Keywords(u).Stalwart {
		t.Errorf("the granted Stalwart expires in Sub-Phase 4c (FR-LAY-003, FR-TRN-001)")
	}
	if got := s.g.Power(u); got != 2 {
		t.Errorf("the +2 power expires with it: power %d, want printed 2 (SVY-R39)", got)
	}
}

func TestFR_LAY_001_GrantedKeywordKeptThroughMove(t *testing.T) {
	s := newScene(t)
	u := s.put(0, ZoneReserves, unit("Rider", 2, 9, Keywords{}))
	gkEnact(s, grantAlliedManeuver(grantStep(0, ability.KeywordStalwart),
		moveStep(ability.SubjectTarget, 0, ability.Filter{}, ability.ZoneMainTheater)))
	if a := s.g.Asset(u.ID); a == nil || a.Zone != ZoneMainTheater {
		t.Fatalf("the unit moved to the Main Theater: %v", a)
	}
	if !s.g.Keywords(u).Stalwart {
		t.Errorf("a granted keyword is kept through a move between battlefield zones (FR-LAY-001)")
	}
}

func TestFR_ENT_103_FaceDownUnitKeepsGrantedKeyword(t *testing.T) {
	s := newScene(t)
	u := s.put(0, ZoneMainTheater, unit("Rider", 2, 9, Keywords{}))
	gkEnact(s, grantAlliedManeuver(grantStep(0, ability.KeywordStalwart),
		ability.Step{Op: ability.OpMorphScratch, Subject: ability.SubjectTarget}))
	if !u.FaceDown {
		t.Fatalf("the unit is face down")
	}
	if got, want := s.g.Keywords(u), (Keywords{DeploysExhausted: true, Stalwart: true}); got != want {
		t.Errorf("a face-down unit keeps its temporary turn modifiers, so it has Stalwart besides its Scratch face's Deploys Exhausted: %+v, want %+v (FR-ENT-103)", got, want)
	}
}

func TestFR_LAY_001_GrantingAggressiveGrantsNothing(t *testing.T) {
	s := newScene(t)
	u := s.put(0, ZoneMainTheater, unit("Rider", 2, 9, Keywords{}))
	from := len(s.g.Log)
	gkEnact(s, grantAlliedManeuver(grantStep(0, ability.KeywordAggressive)))
	if got := s.g.Keywords(u).Aggressive; got != 0 {
		t.Errorf("Aggressive comes only from the printed keyword, a grant gives none: Aggressive %d (FR-LAY-001)", got)
	}
	if len(u.Modifiers) != 0 {
		t.Errorf("a grant of Aggressive adds no modifier: %+v", u.Modifiers)
	}
	if ev := effectEvents(s.g, from, EventKeyword); len(ev) != 0 {
		t.Errorf("%d keyword events for a grant that granted nothing, want 0: %+v", len(ev), ev)
	}
}

func TestEffectEvents_KeywordGrant(t *testing.T) {
	// ADR-0005: one event per asset acted on, in order.
	s := newScene(t)
	s.put(0, ZoneMainTheater, unit("Rider", 2, 9, Keywords{}))
	s.put(0, ZoneReserves, unit("Sentry", 2, 3, Keywords{}))
	var want []*Asset
	for _, a := range s.g.Assets() {
		if a.Controller == 0 && a.Printed().Type == TypeUnit {
			want = append(want, a)
		}
	}
	all := maneuver("Mass Grant", 0, 0, steps(ability.Step{Op: ability.OpGrantKeyword, Subject: ability.SubjectAll,
		Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}, Keyword: ability.KeywordStalwart}))
	from := len(s.g.Log)
	gkEnact(s, all)
	ev := effectEvents(s.g, from, EventKeyword)
	if len(ev) != len(want) || len(want) != 2 {
		t.Fatalf("one keyword event per unit granted, in order: %d events for %d units: %+v", len(ev), len(want), ev)
	}
	for i, a := range want {
		wantDetail := fmt.Sprintf("%s gains Stalwart until end of turn", a)
		if ev[i].Target != a.ID || ev[i].Player != -1 || ev[i].Card != "" || ev[i].Detail != wantDetail {
			t.Errorf("event %d = %+v, want Target %d, Player -1, empty Card, detail %q", i, ev[i], a.ID, wantDetail)
		}
	}
}

func TestEffectEvents_KeywordGrantFaceDownStaysHidden(t *testing.T) {
	s := newScene(t)
	fd := s.putScratch(0)
	from := len(s.g.Log)
	gkEnact(s, grantAlliedManeuver(grantStep(0, ability.KeywordStalwart)))
	ev := effectEvents(s.g, from, EventKeyword)
	if len(ev) != 1 {
		t.Fatalf("one keyword event for the face-down unit, got %d", len(ev))
	}
	if want := fmt.Sprintf("%s gains Stalwart until end of turn", fd); ev[0].Detail != want || ev[0].Card != "" || ev[0].Target != fd.ID {
		t.Errorf("the event names the Scratch Company, not the hidden card: %+v, want detail %q", ev[0], want)
	}
}

func TestFR_TRN_003a_LaterStepActsOnTheDeployedCard(t *testing.T) {
	// Covers: FR-LAY-001
	s := newScene(t)
	c := dfUnit("Rider", ZoneReserves)
	s.inDiscard(0, c)
	s.script(0)
	from := len(s.g.Log)
	gkEnact(s, gkLater(gkLaterSteps(gkDeploy)...))
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneMainTheater {
		t.Fatalf("the card is deployed into the Main Theater (FR-TRN-003a): %v", a)
	}
	if !s.g.Keywords(a).Stalwart {
		t.Errorf("the later step grants Stalwart to the asset the card became (FR-TRN-003a)")
	}
	if got := s.g.Power(a); got != 4 {
		t.Errorf("power %d, want printed 2 plus 2", got)
	}
	for _, kind := range []string{EventKeyword, EventPower} {
		ev := effectEvents(s.g, from, kind)
		if len(ev) != 1 || ev[0].Target != a.ID {
			t.Errorf("one %s event naming the deployed asset %d, got %+v", kind, a.ID, ev)
		}
	}
}

func TestFR_TGT_003_LaterStepWithoutTheDeployDoesNothing(t *testing.T) {
	// Covers: FR-TRN-003a
	check := func(t *testing.T, s *scene, from int) {
		t.Helper()
		for _, a := range s.g.Assets() {
			if a.Printed().Type != TypeUnit {
				continue
			}
			if s.g.Keywords(a).Stalwart || s.g.Power(a) != a.Printed().Power {
				t.Errorf("%s gained something: Stalwart %v, power %d", a, s.g.Keywords(a).Stalwart, s.g.Power(a))
			}
		}
		if ev := effectEvents(s.g, from, EventKeyword); len(ev) != 0 {
			t.Errorf("no keyword event, got %+v", ev)
		}
	}
	t.Run("CardLeftThePile", func(t *testing.T) {
		s := newScene(t)
		s.put(0, ZoneMainTheater, unit("Bystander", 2, 5, Keywords{}))
		c := dfUnit("Leaves", ZoneReserves)
		inst := s.inDiscard(0, c)[0]
		man := s.give(0, gkLater(gkLaterSteps(gkDeploy)...))
		s.script(0)
		moved := false
		s.g.obs = &recorder{onEvent: func(e Event) {
			if e.Kind == "place" && !moved {
				moved = true
				if s.g.takeCard(inst) {
					s.g.putCard(inst, ZoneHand, false)
				}
			}
		}}
		from := len(s.g.Log)
		s.do(enactAction(man))
		if !moved {
			t.Fatalf("the Maneuver was placed on the stack")
		}
		if assetOf(s.g, c) != nil {
			t.Errorf("a card that left the pile is not deployed (FR-TGT-003)")
		}
		check(t, s, from)
	})
	t.Run("DeploySkippedUnsupported", func(t *testing.T) {
		s := newScene(t)
		s.put(0, ZoneMainTheater, unit("Bystander", 2, 5, Keywords{}))
		c := dfUnit("Odd", ZoneReserves)
		c.Unsupported = []string{"Some rule the kernel cannot run."}
		s.inDiscard(0, c)
		s.script(0)
		from := len(s.g.Log)
		gkEnact(s, gkLater(gkLaterSteps(gkDeploy)...))
		if assetOf(s.g, c) != nil {
			t.Errorf("an unsupported card is never deployed")
		}
		check(t, s, from)
	})
	t.Run("NotADeploy", func(t *testing.T) {
		s := newScene(t)
		s.put(0, ZoneMainTheater, unit("Bystander", 2, 5, Keywords{}))
		c := dfUnit("Rider", ZoneReserves)
		inst := s.inDiscard(0, c)[0]
		s.script(0)
		from := len(s.g.Log)
		gkEnact(s, gkLater(gkLaterSteps(ability.Step{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget})...))
		if !inHand(s.g.Players[0], inst) {
			t.Errorf("the card is in hand")
		}
		if assetOf(s.g, c) != nil {
			t.Errorf("a card returned to hand is no asset")
		}
		check(t, s, from)
	})
}
