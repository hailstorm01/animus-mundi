package engine

import (
	"slices"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0222 (FR-MOD-001 4 and 6, FR-TRN-003a, FR-CST-001, wave F):
// OpLookTop has its controller look at the top N cards of their deck and, in
// one hidden DecideLook prompt, choose which are deployed for free (up to UpTo,
// each meeting Filter), the zone of each, and the order the rest go to the
// bottom (SVY-R41, SVY-R42, SVY-R43). Each option is a whole outcome, listed:
// deploy sets smaller first, then zones per card, then bottom orders, each in
// look order. Cards are synthetic. Player 0 enacts a Maneuver with no targets
// whose steps are the look and then a Scratch 1, which takes the deck's new top.

// ltFilter is "a unit costing 3 or less".
var ltFilter = ability.Filter{Kinds: ability.KindUnit, Stat: ability.StatRequisition, AtMost: 3}

// ltEligible is an eligible unit: Requisition cost 3, printing both zones.
func ltEligible(name string) *Card {
	c := unit(name, 2, 2, Keywords{})
	c.ReqCost = 3
	return c
}

// ltOther is a card no ltFilter look can deploy: a Maneuver.
func ltOther(name string) *Card {
	return maneuver(name, 0, 0, ability.Effect{})
}

// ltScene makes the deck exactly cards (cards[0] on top) and enacts a Maneuver
// whose steps are the look (amount cards, upTo, ltFilter), then any extra
// steps given before the closing Scratch 1. Seat 0 is scripted with answers.
func ltScene(t *testing.T, amount, upTo int, cards []*Card, mid []ability.Step, answers ...Answer) (*scene, []*Instance) {
	t.Helper()
	s := newScene(t)
	s.g.Players[0].Deck = nil
	insts := s.onTop(0, cards...)
	look := ability.Step{Op: ability.OpLookTop, Amount: ability.Value{N: amount}, Filter: ltFilter, UpTo: upTo}
	ss := append([]ability.Step{look}, mid...)
	ss = append(ss, scratchN(1))
	man := s.give(0, maneuver("Look Top", 0, 0, ability.Effect{Steps: ss}))
	s.script(0, answers...)
	s.do(enactAction(man))
	return s, insts
}

func ltLooks(s *scene) []Decision {
	var out []Decision
	for _, d := range s.asked(0) {
		if d.Kind == DecideLook {
			out = append(out, d)
		}
	}
	return out
}

// ltDeck is player 0's deck as indexes into insts, top first.
func ltDeck(s *scene, insts []*Instance) []int {
	var out []int
	for _, d := range s.g.Players[0].Deck {
		out = append(out, slices.Index(insts, d))
	}
	return out
}

// ltWantDeck checks the deck's top-first order, as indexes into insts.
func ltWantDeck(t *testing.T, s *scene, insts []*Instance, why string, want ...int) {
	t.Helper()
	if got := ltDeck(s, insts); !slices.Equal(got, want) {
		t.Errorf("deck top first (indexes of the listed cards) = %v, want %v: %s", got, want, why)
	}
}

// ltAsset is the battlefield asset a card instance became, or nil.
func ltAsset(s *scene, inst *Instance) *Asset {
	for _, a := range s.g.Assets() {
		if a.Inst == inst {
			return a
		}
	}
	return nil
}

func ltOneLook(t *testing.T, s *scene, wantOptions int) Decision {
	t.Helper()
	looks := ltLooks(s)
	if len(looks) != 1 {
		t.Fatalf("player 0 is asked one DecideLook, got %d (FR-MOD-001 4)", len(looks))
	}
	d := looks[0]
	if d.Player != 0 || d.Min != 1 || d.Max != 1 || !d.Hidden || len(d.Options) != wantOptions {
		t.Fatalf("one hidden pick of one whole outcome out of %d options: %+v (FR-MOD-001 4)", wantOptions, d)
	}
	return d
}

// ltLookEvent is the recorded decision of the DecideLook, or nil.
func ltLookEvent(s *scene) *DecisionRecord {
	for _, e := range eventsOfKind(s.g.Log, "decision") {
		if e.Decision != nil && e.Decision.Kind == DecideLook.String() {
			return e.Decision
		}
	}
	return nil
}

func TestFR_MOD_001_LookTopTimeoutDeploysNothingKeepsOrderSeen(t *testing.T) {
	// Covers: SVY-R41, SVY-R42
	deck := []*Card{ltEligible("A"), ltEligible("B"), ltEligible("C"), ltEligible("D")}
	s, insts := ltScene(t, 3, 1, deck, nil, Answer{Timeout: true})
	if n := len(s.asked(0)); n != 1 {
		t.Errorf("the DecideLook is the only decision player 0 was asked, got %d", n)
	}
	// 6 bottom orders with nothing deployed + 3 cards x 2 zones x 2 orders = 18.
	ltOneLook(t, s, 18)
	if rec := ltLookEvent(s); rec == nil || !slices.Equal(rec.Picks, []int{0}) {
		t.Errorf("a timed-out look is recorded as option 0, the first listed: %v (SVY-R41, SVY-R42)", rec)
	}
	for _, i := range []int{0, 1, 2} {
		if a := ltAsset(s, insts[i]); a != nil {
			t.Errorf("a timed-out optional look deploys no card, but card %d is on the battlefield (SVY-R41)", i)
		}
	}
	a := ltAsset(s, insts[3])
	if a == nil || !a.FaceDown || a.Zone != ZoneReserves {
		t.Errorf("the Scratch takes the new top card D into Reserves face down: %v", a)
	}
	ltWantDeck(t, s, insts, "the looked cards go to the bottom in the order seen, top first (SVY-R42)", 0, 1, 2)
}

func TestFR_MOD_001_LookTopKeepsThePlayersBottomOrder(t *testing.T) {
	// Covers: FR-MOD-001
	deck := []*Card{ltEligible("A"), ltEligible("B"), ltEligible("C"), ltEligible("D")}
	// Deploy nothing: orders ABC, ACB, BAC, BCA, CAB (4), CBA. Option 4 is C, A, B.
	s, insts := ltScene(t, 3, 1, deck, nil, Answer{Picks: []int{4}})
	ltOneLook(t, s, 18)
	for _, i := range []int{0, 1, 2} {
		if a := ltAsset(s, insts[i]); a != nil {
			t.Errorf("the answer deploys nothing, but card %d is on the battlefield", i)
		}
	}
	ltWantDeck(t, s, insts, "the answer puts the looked cards on the bottom as C, A, B (FR-MOD-001 4)", 2, 0, 1)
}

func TestFR_MOD_001_LookTopZoneChosenInTheSamePrompt(t *testing.T) {
	// Covers: SVY-R43, FR-TRN-003a, FR-CST-001
	dear := ltEligible("C")
	dear.ReqCost = 4
	deck := []*Card{ltOther("A"), ltEligible("B"), dear, ltEligible("D")}
	// 0-5 deploy nothing; 6, 7: B to Reserves (bottom A, C; C, A); 8, 9: B to the
	// Main Theater (A, C; C, A). Option 9: B to the Main Theater, bottom C, A.
	s, insts := ltScene(t, 3, 1, deck, nil, Answer{Picks: []int{9}})
	ltOneLook(t, s, 10)
	if n := len(s.asked(0)); n != 1 {
		t.Errorf("the zone is chosen in the look's prompt, the only decision asked, got %d (SVY-R43)", n)
	}
	a := ltAsset(s, insts[1])
	if a == nil || a.Controller != 0 || a.Zone != ZoneMainTheater || a.FaceDown {
		t.Errorf("B enters the Main Theater under player 0 (FR-TRN-003a 1): %v", a)
	}
	ltWantDeck(t, s, insts, "C, A are the bottom order the answer chose", 2, 0)
}

func TestFR_MOD_001_LookTopOffersOnlyEligibleCards(t *testing.T) {
	// Covers: FR-TRN-003a, FR-TGT-001
	cheap := ltEligible("Cheap")
	cheap.ReqCost = 1
	cheap.Unsupported = []string{"Rules text the kernel cannot run."}
	dear := ltEligible("Dear")
	dear.ReqCost = 4
	deck := []*Card{ltOther("Spell"), dear, cheap, ltEligible("D")}
	// Nothing is eligible, so only the 6 bottom orders; the last (5) is the reverse.
	s, insts := ltScene(t, 3, 1, deck, nil, Answer{Picks: []int{5}})
	ltOneLook(t, s, 6)
	for _, i := range []int{0, 1, 2} {
		if a := ltAsset(s, insts[i]); a != nil {
			t.Errorf("card %d is not an eligible deploy (not a supported unit costing 3 or less) yet it is deployed (FR-TRN-003a)", i)
		}
	}
	ltWantDeck(t, s, insts, "the looked cards in reverse look order", 2, 1, 0)
}

func TestFR_MOD_001_LookTopRequiredPickTimesOutToTheFirstEligible(t *testing.T) {
	// Covers: SVY-R41
	deck := []*Card{ltOther("A"), ltEligible("B"), ltEligible("C"), ltEligible("D")}
	s, insts := ltScene(t, 3, 0, deck, nil, Answer{Timeout: true})
	// Exactly one eligible card: B or C, 2 zones each, 2 bottom orders each.
	ltOneLook(t, s, 8)
	a := ltAsset(s, insts[1])
	if a == nil || a.Controller != 0 || a.Zone != ZoneReserves {
		t.Errorf("a timed-out required pick deploys the first eligible card from the top, B, into Reserves (SVY-R41): %v", a)
	}
	if a := ltAsset(s, insts[2]); a != nil {
		t.Errorf("only one card is deployed by a required pick, but C is on the battlefield (SVY-R41)")
	}
	ltWantDeck(t, s, insts, "the rest go to the bottom in the order seen, A then C (SVY-R41)", 0, 2)
}

func TestFR_MOD_001_LookTopShortDeckLooksAtWhatIsThere(t *testing.T) {
	// Covers: FR-MOD-001
	t.Run("TwoCards", func(t *testing.T) {
		s, _ := ltScene(t, 3, 1, []*Card{ltEligible("A"), ltEligible("B")}, nil, Answer{Timeout: true})
		d := ltOneLook(t, s, 6) // 2 + 2 + 2
		if d.Prompt != "Look at the top 2 cards of your deck" {
			t.Errorf("the prompt names the 2 cards there are: %q (FR-MOD-001 4)", d.Prompt)
		}
	})
	t.Run("EmptyDeck", func(t *testing.T) {
		s, _ := ltScene(t, 3, 1, nil, nil)
		if n := len(s.asked(0)); n != 0 {
			t.Errorf("an empty deck asks no decision, got %d", n)
		}
		for _, a := range s.g.Assets() {
			if a.Controller == 0 && a.FaceDown {
				t.Errorf("the Scratch after a look at an empty deck deploys nothing: %v", a)
			}
		}
	})
}

func TestFR_MOD_001_LookTopIsHiddenSoNoFastPath(t *testing.T) {
	// Covers: FR-MOD-001
	s, _ := ltScene(t, 3, 1, []*Card{ltOther("A")}, nil, Answer{Picks: []int{0}})
	ltOneLook(t, s, 1)
	rec := ltLookEvent(s)
	if rec == nil || rec.How != "answer" {
		t.Errorf("a deck inspection takes no fast path even with one option: want How answer, got %v (FR-MOD-001 1)", rec)
	}
}

func TestFR_TRN_003a_LookTopDeployCountsForIfDone(t *testing.T) {
	// Covers: SVY-R14
	handChange := func(answer int) int {
		t.Helper()
		draw := drawN(1)
		draw.IfDone = true
		deck := []*Card{ltEligible("A"), ltOther("B"), ltOther("C"), ltOther("D"), ltOther("E")}
		s := newScene(t)
		s.g.Players[0].Deck = nil
		s.onTop(0, deck...)
		look := ability.Step{Op: ability.OpLookTop, Amount: ability.Value{N: 3}, Filter: ltFilter, UpTo: 1}
		man := s.give(0, maneuver("Look Top", 0, 0, ability.Effect{Steps: []ability.Step{look, draw, scratchN(1)}}))
		s.script(0, Answer{Picks: []int{answer}})
		before := len(s.g.Players[0].Hand)
		s.do(enactAction(man))
		return len(s.g.Players[0].Hand) - before
	}
	// Option 0 deploys nothing; option 6 is A into Reserves (bottom B, C).
	if got := handChange(0); got != -1 {
		t.Errorf("deploying none draws none: hand changed by %d, want -1 (the enacted card) (SVY-R14)", got)
	}
	if got := handChange(6); got != 0 {
		t.Errorf("deploying a card from the look draws one: hand changed by %d, want 0 (SVY-R14)", got)
	}
}

func TestFR_TRN_003a_LookTopDeployFiresDeployTriggers(t *testing.T) {
	// Covers: SVY-R1
	trig := ltEligible("Triggered")
	trig.Ability = &ability.Def{Triggers: []ability.Trigger{deployTrigger(drawN(1))}}
	deck := []*Card{trig, ltOther("B"), ltOther("C"), ltOther("D"), ltOther("E")}
	// Option 6: Triggered into Reserves, bottom B, C.
	s := newScene(t)
	s.g.Players[0].Deck = nil
	insts := s.onTop(0, deck...)
	look := ability.Step{Op: ability.OpLookTop, Amount: ability.Value{N: 3}, Filter: ltFilter, UpTo: 1}
	man := s.give(0, maneuver("Look Top", 0, 0, ability.Effect{Steps: []ability.Step{look}}))
	s.script(0, Answer{Picks: []int{6}})
	before := len(s.g.Players[0].Hand)
	s.do(enactAction(man))
	if a := ltAsset(s, insts[0]); a == nil || a.Zone != ZoneReserves {
		t.Fatalf("the looked unit is deployed into Reserves: %v", a)
	}
	if got := len(s.g.Players[0].Hand) - before; got != 0 {
		t.Errorf("its deploy trigger fires and resolves after the Maneuver: hand changed by %d, want 0 (-1 enacted, +1 drawn) (SVY-R1)", got)
	}
}
