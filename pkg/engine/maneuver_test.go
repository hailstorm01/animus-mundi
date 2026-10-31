package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0116 (FR-MNV-001, FR-TRN-003, board directive #88 part 1):
// enacting a Maneuver from hand. Its costs are paid and the card goes on the
// Resolution Stack as it is declared, its modes and targets are chosen then,
// its steps resolve under two-stage verification, and the card then goes to
// its owner's discard pile, even when every target fizzled. Cards are
// synthetic Maneuvers. The legality of a declaration (SVY-R29) is the next
// task's and is deliberately not pinned here: every Maneuver below has a legal
// entity for each exactly-one slot, and the modal one has enough legal modes.
// Player 0 is the active player and the enacting player.

// maneuver is a Maneuver costing req Requisition and tac Tactics whose
// definition has effect e.
func maneuver(name string, req, tac int, e ability.Effect) *Card {
	return &Card{Name: name, Type: TypeManeuver, ReqCost: req, TacCost: tac,
		Ability: &ability.Def{Effect: e}}
}

func steps(ss ...ability.Step) ability.Effect { return ability.Effect{Steps: ss} }

// enactAction is player 0 enacting inst.
func enactAction(inst *Instance) Action {
	return Action{Kind: ActEnact, Player: 0, Card: inst.ID}
}

func offersEnact(g *Game, inst *Instance) bool {
	for _, a := range g.LegalActions() {
		if a.Kind == ActEnact && a.Card == inst.ID {
			return true
		}
	}
	return false
}

func lastDiscard(p *Player) *Instance {
	if len(p.Discard) == 0 {
		return nil
	}
	return p.Discard[len(p.Discard)-1]
}

func TestFR_MNV_001_EnactedManeuverResolvesThenGoesToDiscard(t *testing.T) {
	// Covers: FR-TRN-003, FR-ZON-001
	s := newScene(t)
	inst := s.give(0, maneuver("Study", 1, 1, steps(drawN(2))))
	p := s.g.Players[0]
	p.Req, p.Tac = 1, 1
	deck := len(p.Deck)
	s.do(enactAction(inst))
	if p.Req != 0 || p.Tac != 0 {
		t.Errorf("enacting pays the Maneuver's cost R1 T1: pool R%d T%d, want R0 T0 (FR-TRN-003)", p.Req, p.Tac)
	}
	if got := deck - len(p.Deck); got != 2 {
		t.Errorf("its effect \"draw 2 cards\" resolved: drew %d, want 2 (FR-MNV-001)", got)
	}
	if inHand(p, inst) {
		t.Errorf("the enacted card is no longer in hand (FR-TRN-003)")
	}
	if lastDiscard(p) != inst {
		t.Errorf("once resolved the Maneuver is the last card of its owner's discard pile (FR-MNV-001)")
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty after resolution: %d items (FR-QUE-001)", len(s.g.Stack))
	}
}

func TestFR_MNV_001_EnactIsALegalActionOnlyWhenPayable(t *testing.T) {
	s := newScene(t)
	inst := s.give(0, maneuver("Study", 1, 1, steps(drawN(1))))
	p := s.g.Players[0]
	p.Req, p.Tac = 1, 1
	if !offersEnact(s.g, inst) {
		t.Errorf("a supported Maneuver in hand whose cost the player can pay is a legal Enact (FR-MNV-001)")
	}
	p.Req, p.Tac = 1, 0
	if offersEnact(s.g, inst) {
		t.Errorf("Enact is not offered when the player cannot pay Tactics 1 (FR-TRN-003)")
	}
	s.reject(enactAction(inst))
	if !inHand(p, inst) || p.Req != 1 {
		t.Errorf("a rejected Enact changes nothing: card in hand %v, Req %d", inHand(p, inst), p.Req)
	}
	p.Req, p.Tac = 0, 1
	if offersEnact(s.g, inst) {
		t.Errorf("Enact is not offered when the player cannot pay Requisition 1 (FR-TRN-003)")
	}
}

func TestFR_MNV_001_UnsupportedOrNonManeuverCannotBeEnacted(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	p.Req, p.Tac = 5, 5
	noEffect := s.give(0, &Card{Name: "Empty", Type: TypeManeuver, Ability: &ability.Def{}})
	asUnit := s.give(0, unit("Soldier", 1, 1, Keywords{}))
	inOpp := s.give(1, maneuver("Theirs", 0, 0, steps(drawN(1))))
	good := s.give(0, maneuver("Study", 0, 0, steps(drawN(1))))
	for _, c := range []struct {
		name string
		inst *Instance
	}{{"a Maneuver with no effect", noEffect}, {"a Unit", asUnit}, {"a card in the opponent's hand", inOpp}} {
		if offersEnact(s.g, c.inst) {
			t.Errorf("Enact is not offered for %s (FR-MNV-001)", c.name)
		}
		s.reject(enactAction(c.inst))
		if !inHand(s.g.Players[c.inst.Owner], c.inst) {
			t.Errorf("rejected Enact of %s leaves it in hand", c.name)
		}
	}
	s.reject(Action{Kind: ActDeploy, Player: 0, Card: good.ID, Zone: ZoneReserves})
	if !inHand(p, good) {
		t.Errorf("a Maneuver cannot be deployed and stays in hand (FR-MNV-001)")
	}
}

func TestFR_MNV_001_OnlyTheActivePlayerEnacts(t *testing.T) {
	// Covers: FR-TRN-002
	s := newScene(t)
	inst := s.give(0, maneuver("Study", 0, 0, steps(drawN(1))))
	s.endTurn()
	if s.g.Active != 1 {
		t.Fatalf("precondition: player 1 is now active")
	}
	s.reject(enactAction(inst))
	if !inHand(s.g.Players[0], inst) {
		t.Errorf("the non-active player cannot enact; the card stays in hand (FR-TRN-002)")
	}
}

func TestFR_TGT_001_ManeuverTargetChosenOnDeclaration(t *testing.T) {
	// Covers: FR-MNV-001, FR-TGT-003
	s := newScene(t)
	e1 := s.put(1, ZoneMainTheater, unit("E1", 1, 5, Keywords{}))
	e2 := s.put(1, ZoneMainTheater, unit("E2", 1, 5, Keywords{}))
	inst := s.give(0, maneuver("Execute", 0, 0, ability.Effect{
		Targets: []ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)},
		Steps:   []ability.Step{{Op: ability.OpKill, Subject: ability.SubjectTarget}},
	}))
	s.script(0, Answer{Picks: []int{1}})
	s.do(enactAction(inst))
	var targets []Decision
	for _, d := range s.asked(0) {
		if d.Kind == DecideTarget {
			targets = append(targets, d)
		}
	}
	if len(targets) != 1 {
		t.Fatalf("the enacting player is asked one target decision, got %d (FR-TGT-001)", len(targets))
	}
	seen := map[int]bool{}
	for _, o := range targets[0].Options {
		seen[o.Ref] = true
	}
	if len(targets[0].Options) != 2 || !seen[e1.ID] || !seen[e2.ID] {
		t.Fatalf("the decision lists both opposing units: %v", targets[0].Options)
	}
	chosen, other := e2, e1
	if targets[0].Options[1].Ref != e2.ID {
		chosen, other = e1, e2
	}
	if s.alive(chosen) {
		t.Errorf("the chosen (second listed) unit is destroyed (FR-MNV-001)")
	}
	if !s.alive(other) {
		t.Errorf("the unchosen unit survives (FR-TGT-001)")
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver ends in its owner's discard pile (FR-MNV-001)")
	}
}

func TestFR_QUE_001_ManeuverIsPlacedOnTheStack(t *testing.T) {
	// Covers: FR-MNV-001, FR-TRN-003
	s := newScene(t)
	inst := s.give(0, maneuver("Study", 0, 0, steps(drawN(1))))
	s.do(enactAction(inst))
	places := eventsOfKind(s.g.Log, "place")
	if len(places) != 1 {
		t.Fatalf("enacting emits one place event, got %d (FR-QUE-001)", len(places))
	}
	pe := places[0]
	if pe.Card != "Study" || pe.Inst != inst.ID {
		t.Errorf("the place event names the card: Card %q Inst %d, want Study %d", pe.Card, pe.Inst, inst.ID)
	}
	if pe.Amount != 1 {
		t.Errorf("the Maneuver is placed at stack depth 1, got %d (FR-QUE-001)", pe.Amount)
	}
	draws := eventsOfKind(s.g.Log, "draw")
	if len(draws) == 0 {
		t.Fatalf("precondition: the effect drew a card")
	}
	for _, e := range draws {
		if e.Seq < pe.Seq {
			t.Errorf("the effect's draw (seq %d) happens after the Maneuver is placed (seq %d)", e.Seq, pe.Seq)
		}
	}
}

func TestFR_MOD_001_ModalManeuverResolvesOnlyItsChosenMode(t *testing.T) {
	// Covers: FR-MNV-001, FR-TGT-001
	s := newScene(t)
	u := s.put(1, ZoneMainTheater, unit("Victim", 1, 5, Keywords{}))
	uInst := u.Inst
	drawMode := mode(nil, drawN(1))
	backMode := mode([]ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)},
		ability.Step{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget})
	inst := s.give(0, maneuver("Either", 0, 0, ability.Effect{Modes: []ability.Effect{drawMode, backMode}, ChooseN: 1}))
	deck := len(s.g.Players[0].Deck)
	s.scriptModal(0, nil, nil, []Answer{{Picks: []int{1}}}, pick(u))
	s.do(enactAction(inst))
	if s.alive(u) || !inHand(s.g.Players[1], uInst) {
		t.Errorf("mode 2 returned the chosen unit's card to its owner's hand (FR-MOD-001 §5)")
	}
	if got := len(s.g.Players[0].Deck); got != deck {
		t.Errorf("mode 1 was not chosen, so no card was drawn: deck %d -> %d (FR-MOD-001 §5)", deck, got)
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver is in its owner's discard pile (FR-MNV-001)")
	}
}

func TestFR_MNV_001_FizzledManeuverStillGoesToDiscard(t *testing.T) {
	// Covers: SVY-R25, FR-TGT-003
	s := newScene(t)
	u := s.put(1, ZoneMainTheater, unit("Victim", 1, 5, Keywords{}))
	uInst := u.Inst
	inst := s.give(0, maneuver("Bounce", 0, 0, ability.Effect{
		Targets: []ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)},
		Steps: []ability.Step{
			{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget},
			{Op: ability.OpKill, Subject: ability.SubjectTarget},
		},
	}))
	s.script(0, Answer{Picks: []int{0}})
	s.do(enactAction(inst))
	if !inHand(s.g.Players[1], uInst) {
		t.Errorf("the unit's card was returned to its owner's hand")
	}
	if n := len(eventsOfKind(s.g.Log, "destroy")); n != 0 {
		t.Errorf("the kill step finds the unit gone and destroys nothing: %d destroy events (SVY-R25)", n)
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("a Maneuver whose targets are gone still goes to its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_SupportedManeuverNeedsAnEffect(t *testing.T) {
	// Covers: FR-MNV-001
	eff := steps(drawN(1))
	modal := ability.Effect{Modes: []ability.Effect{mode(nil, drawN(1))}, ChooseN: 1}
	cases := []struct {
		name string
		c    *Card
		want bool
	}{
		{"steps", &Card{Type: TypeManeuver, Ability: &ability.Def{Effect: eff}}, true},
		{"modes", &Card{Type: TypeManeuver, Ability: &ability.Def{Effect: modal}}, true},
		{"nil Ability", &Card{Type: TypeManeuver}, false},
		{"empty Effect", &Card{Type: TypeManeuver, Ability: &ability.Def{}}, false},
		{"unsupported line", &Card{Type: TypeManeuver, Unsupported: []string{"Do something."}, Ability: &ability.Def{Effect: eff}}, false},
	}
	for _, tc := range cases {
		if got := tc.c.Supported(); got != tc.want {
			t.Errorf("Maneuver with %s: Supported() = %v, want %v", tc.name, got, tc.want)
		}
	}
}
