package engine

import (
	"reflect"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for the self-deploy trigger pipeline (FR-TRG-003, FR-TRG-004) and the
// DecideOrder decision kind. Cross-player APNAP and the deep-cascade case of
// FR-TRG-004 are out of scope here (T-0008, board directive #11).

func scratchN(n int) ability.Step {
	return ability.Step{Op: ability.OpScratch, Amount: ability.Value{N: n}}
}
func drawN(n int) ability.Step { return ability.Step{Op: ability.OpDraw, Amount: ability.Value{N: n}} }

// deployTrigger is one "When this deploys, <steps>" ability.
func deployTrigger(steps ...ability.Step) ability.Trigger {
	return ability.Trigger{Hook: ability.HookDeploy, Effect: ability.Effect{Steps: steps}}
}

// trigUnit is a free Reserves unit printing the given deploy triggers in order.
func trigUnit(name string, trigs ...ability.Trigger) *Card {
	c := unit(name, 1, 1, Keywords{})
	c.DeployZones = []Zone{ZoneReserves}
	c.Ability = &ability.Def{Triggers: trigs}
	return c
}

func vanilla(names ...string) []*Card {
	out := make([]*Card, len(names))
	for i, n := range names {
		out[i] = unit(n, 1, 1, Keywords{})
	}
	return out
}

func (s *scene) deployFromHand(p PlayerID, c *Card) {
	s.t.Helper()
	inst := s.give(p, c)
	s.do(Action{Kind: ActDeploy, Player: p, Card: inst.ID, Zone: ZoneReserves})
}

func inHand(p *Player, inst *Instance) bool {
	for _, h := range p.Hand {
		if h == inst {
			return true
		}
	}
	return false
}

// scratched reports whether inst is one of p's face-down Scratch Companies.
func scratched(p *Player, inst *Instance) bool {
	for _, a := range p.Reserves {
		if a.FaceDown && a.Inst == inst {
			return true
		}
	}
	return false
}

func faceDownCount(p *Player) int {
	n := 0
	for _, a := range p.Reserves {
		if a.FaceDown {
			n++
		}
	}
	return n
}

// --- deploy triggers resolve after the card is on the battlefield ---

func TestFR_TRG_004_DeployTriggerResolvesAfterCardIsOnBattlefield(t *testing.T) {
	s := newScene(t)
	cards := s.onTop(0, vanilla("A", "B")...)
	p := s.g.Players[0]
	handBefore := len(p.Hand)
	var sawSource bool
	var handAtAsk int
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		for _, a := range p.Reserves {
			if a.ID == d.Options[0].Ref {
				sawSource = true
			}
		}
		handAtAsk = len(p.Hand)
		return Answer{Timeout: true}
	}
	s.deployFromHand(0, trigUnit("Source", deployTrigger(drawN(1)), deployTrigger(scratchN(1))))
	if !sawSource {
		t.Errorf("triggers are ordered after the deploy resolved, with the source asset already on the battlefield (FR-TRG-004)")
	}
	if handAtAsk != handBefore {
		t.Errorf("no trigger has resolved when the player orders them: hand %d, want %d (FR-TRG-004)", handAtAsk, handBefore)
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty once the triggers have resolved: %d items", len(s.g.Stack))
	}
	if !inHand(p, cards[0]) && !inHand(p, cards[1]) {
		t.Errorf("the Draw trigger resolved: neither top card is in hand")
	}
	if len(p.Reserves) != 2 {
		t.Errorf("source plus one Scratch Company in Reserves, got %d assets", len(p.Reserves))
	}
}

func TestFR_TRG_004_SingleTriggerResolvesWithoutQuestion(t *testing.T) {
	s := newScene(t)
	cards := s.onTop(0, vanilla("A", "B", "C")...)
	s.script(0)
	s.deployFromHand(0, trigUnit("Source", deployTrigger(drawN(2))))
	p := s.g.Players[0]
	if !inHand(p, cards[0]) || !inHand(p, cards[1]) || inHand(p, cards[2]) {
		t.Errorf("a Draw 2 trigger draws exactly the top two cards")
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a player with one trigger is not asked to order it: asked %d times (FR-TRG-003)", n)
	}
}

func TestFR_TRG_004_StepsOfOneEffectRunInPrintedOrder(t *testing.T) {
	s := newScene(t)
	cards := s.onTop(0, vanilla("A", "B")...)
	s.script(0)
	s.deployFromHand(0, trigUnit("Source", deployTrigger(scratchN(1), drawN(1))))
	p := s.g.Players[0]
	if !scratched(p, cards[0]) || !inHand(p, cards[1]) {
		t.Errorf("an effect's steps run in order: Scratch takes the top card, then Draw takes the next")
	}
}

func TestFR_TRG_004_ScratchAmountPerformsScratchCompanyThatManyTimes(t *testing.T) {
	s := newScene(t)
	s.onTop(0, vanilla("A", "B", "C")...)
	s.script(0)
	s.deployFromHand(0, trigUnit("Source", deployTrigger(scratchN(2))))
	if n := faceDownCount(s.g.Players[0]); n != 2 {
		t.Errorf("Scratch 2 deploys two face-down Scratch Companies (FR-ZON-005): got %d", n)
	}
}

func TestFR_TRG_004_TriggerControllerIsTheDeployingPlayer(t *testing.T) {
	s := newScene(t)
	s.endTurn() // player 1 is now active
	mine := s.onTop(1, vanilla("A")...)
	theirs := s.onTop(0, vanilla("Z")...)
	s.script(1)
	s.script(0)
	s.deployFromHand(1, trigUnit("Source", deployTrigger(drawN(1))))
	if !inHand(s.g.Players[1], mine[0]) || inHand(s.g.Players[0], theirs[0]) {
		t.Errorf("a trigger's controller is its source's controller: only the deploying player draws")
	}
}

// --- FR-TRG-003: ordering the player's own simultaneous triggers ---

func TestFR_TRG_003_TwoTriggersAskOneOrderDecision(t *testing.T) {
	s := newScene(t)
	s.onTop(0, vanilla("A", "B")...)
	s.script(0, Answer{Picks: []int{0, 1}})
	s.script(1)
	s.deployFromHand(0, trigUnit("Source", deployTrigger(drawN(1)), deployTrigger(scratchN(1))))
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("a player with two simultaneous triggers is asked exactly once: asked %d times", len(asked))
	}
	d := asked[0]
	if d.Kind != DecideOrder || len(d.Options) != 2 || d.Min != 2 || d.Max != 2 || d.Player != 0 {
		t.Errorf("the decision orders both triggers (DecideOrder, 2 options, Min=Max=2) for the controller: %+v", d)
	}
	src := s.g.Players[0].Reserves[0]
	for i, o := range d.Options {
		if o.Ref != src.ID {
			t.Errorf("option %d names the source asset %d, got Ref %d", i, src.ID, o.Ref)
		}
		if o.Label == "" {
			t.Errorf("option %d has no label", i)
		}
	}
	if n := len(s.asked(1)); n != 0 {
		t.Errorf("the other player has no triggers and is not asked: %d", n)
	}
}

func TestFR_TRG_003_FirstPickIsPlacedFirstAndResolvesLast(t *testing.T) {
	// Triggers: 0 = Draw, 1 = Scratch. Answer [0,1]: Draw is placed first and
	// Scratch on top of it, so Scratch resolves first and takes the top card.
	s := newScene(t)
	cards := s.onTop(0, vanilla("A", "B")...)
	s.script(0, Answer{Picks: []int{0, 1}})
	s.deployFromHand(0, trigUnit("Source", deployTrigger(drawN(1)), deployTrigger(scratchN(1))))
	p := s.g.Players[0]
	if !scratched(p, cards[0]) || !inHand(p, cards[1]) {
		t.Errorf("answer [0 1]: trigger 1 (Scratch) resolves first and takes the top card, trigger 0 (Draw) takes the next")
	}
}

func TestFR_TRG_003_ReversedAnswerReversesResolution(t *testing.T) {
	s := newScene(t)
	cards := s.onTop(0, vanilla("A", "B")...)
	s.script(0, Answer{Picks: []int{1, 0}})
	s.deployFromHand(0, trigUnit("Source", deployTrigger(drawN(1)), deployTrigger(scratchN(1))))
	p := s.g.Players[0]
	if !inHand(p, cards[0]) || !scratched(p, cards[1]) {
		t.Errorf("answer [1 0]: trigger 0 (Draw) resolves first and takes the top card, trigger 1 (Scratch) takes the next")
	}
}

func TestFR_TRG_003_ThreeTriggersFollowThePlayersPermutation(t *testing.T) {
	// 0 = Draw 1, 1 = Scratch 1, 2 = Draw 2. Answer [2,0,1] places 2 first, so
	// resolution is 1, 0, 2: Scratch takes A, Draw takes B, Draw 2 takes C and D.
	// (The identity answer would resolve 2 first and draw A.)
	s := newScene(t)
	cards := s.onTop(0, vanilla("A", "B", "C", "D")...)
	s.script(0, Answer{Picks: []int{2, 0, 1}})
	s.deployFromHand(0, trigUnit("Source", deployTrigger(drawN(1)), deployTrigger(scratchN(1)), deployTrigger(drawN(2))))
	p := s.g.Players[0]
	if n := len(s.asked(0)); n != 1 {
		t.Errorf("three triggers are ordered in one decision: asked %d times", n)
	}
	if !scratched(p, cards[0]) {
		t.Errorf("answer [2 0 1]: the Scratch trigger resolves first and takes the top card")
	}
	for _, i := range []int{1, 2, 3} {
		if !inHand(p, cards[i]) {
			t.Errorf("card %d should have been drawn", i)
		}
	}
}

func TestFR_TRG_003_FallbackIsCollectionOrder(t *testing.T) {
	// Collection order is printed order (board directive #10): the identity answer.
	// 0 = Draw, 1 = Scratch: Scratch resolves first and takes the top card.
	cases := map[string]func(s *scene){
		"timeout":        func(s *scene) { s.script(0, Answer{Timeout: true}) },
		"no chooser":     func(s *scene) {},
		"invalid repeat": func(s *scene) { s.script(0, Answer{Picks: []int{1, 1}}) },
		"invalid range":  func(s *scene) { s.script(0, Answer{Picks: []int{0, 5}}) },
		"too few":        func(s *scene) { s.script(0, Answer{Picks: []int{1}}) },
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			s := newScene(t)
			cards := s.onTop(0, vanilla("A", "B")...)
			setup(s)
			s.deployFromHand(0, trigUnit("Source", deployTrigger(drawN(1)), deployTrigger(scratchN(1))))
			p := s.g.Players[0]
			if !scratched(p, cards[0]) || !inHand(p, cards[1]) {
				t.Errorf("%s: falls back to collection (printed) order, placing trigger 0 first so trigger 1 resolves first", name)
			}
		})
	}
}

func TestFR_TRG_003_InvalidOrderAnswerIsLoggedAsInvalid(t *testing.T) {
	s := newScene(t)
	s.onTop(0, vanilla("A", "B")...)
	s.script(0, Answer{Picks: []int{0, 0}})
	s.deployFromHand(0, trigUnit("Source", deployTrigger(drawN(1)), deployTrigger(scratchN(1))))
	if n := len(invalidLogs(s.g)); n != 1 {
		t.Errorf("an ordering answer that is not a permutation is logged as invalid-answer: %d events, want 1", n)
	}
}

// --- FR-TRG-004: buffering and the death sweep ---

func TestFR_TRG_004_EmptyDeckDrawLosesAtItsOwnSweepBeforeNextTrigger(t *testing.T) {
	// The deck holds one card. Collection order [Scratch, Draw 2] resolves Draw 2
	// first: its second draw is from an empty deck, and the match is lost at that
	// trigger's own sweep, so the Scratch trigger below it never resolves.
	s := newScene(t)
	p := s.g.Players[0]
	p.Deck = nil
	cards := s.onTop(0, vanilla("A")...)
	s.script(0, Answer{Timeout: true})
	s.deployFromHand(0, trigUnit("Source", deployTrigger(scratchN(1)), deployTrigger(drawN(2))))
	if !s.g.Outcome.Over || s.g.Outcome.Draw || s.g.Outcome.Winner != 1 {
		t.Errorf("drawing from an empty deck loses at that trigger's sweep (FR-WIN-001): outcome %+v", s.g.Outcome)
	}
	if !inHand(p, cards[0]) {
		t.Errorf("the last card was drawn before the empty draw")
	}
	if faceDownCount(p) != 0 {
		t.Errorf("the match ended at the Draw trigger's sweep, so the later Scratch trigger never resolved")
	}
}

func TestFR_TRG_004_MatchEndedAtDeploySweepDiscardsBufferedTriggers(t *testing.T) {
	s := newScene(t)
	cards := s.onTop(0, vanilla("A")...)
	s.g.Players[1].DrewFromEmpty = true // player 1 loses at the next death sweep
	s.script(0)
	s.deployFromHand(0, trigUnit("Source", deployTrigger(drawN(1))))
	if !s.g.Outcome.Over || s.g.Outcome.Winner != 0 {
		t.Fatalf("the deploy item's sweep ends the match with player 1 losing: %+v", s.g.Outcome)
	}
	if inHand(s.g.Players[0], cards[0]) {
		t.Errorf("triggers buffered during an item are placed only after its sweep; the match ended there, so they are discarded (FR-TRG-004)")
	}
}

// --- self hook: a face-down Scratch Company never triggers ---

func TestFR_TRG_003_FaceDownScratchCompanyNeverTriggers(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	top := trigUnit("Printed Trigger", deployTrigger(drawN(1)), deployTrigger(drawN(1)))
	rest := vanilla("B", "C")
	inst := s.onTop(0, append([]*Card{top}, rest...)...)
	s.script(0)
	handBefore := len(p.Hand)
	s.do(Action{Kind: ActScratch, Player: 0, Asset: p.HQ.ID})
	if !scratched(p, inst[0]) {
		t.Fatalf("the top card became a face-down Scratch Company")
	}
	if len(p.Hand) != handBefore || inHand(p, inst[1]) {
		t.Errorf("a face-down Scratch Company has no abilities, so its printed trigger does not fire")
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("no triggers means no ordering question: asked %d times", n)
	}
}

// --- DecideOrder against the FR-MOD-001 choice rules ---

func orderDecision(n int) Decision {
	return Decision{Player: 0, Kind: DecideOrder, Prompt: "order", Options: opts(n), Min: n, Max: n}
}

func TestFR_TRG_003_OrderDecisionWithManyOptionsIsNeverFastPathed(t *testing.T) {
	for _, want := range [][]int{{1, 0}, {2, 0, 1}} {
		s := newScene(t)
		s.script(0, Answer{Picks: want})
		got := s.g.choose(orderDecision(len(want)))
		if !reflect.DeepEqual(got, want) {
			t.Errorf("n=%d: the player's permutation is used as given: got %v, want %v", len(want), got, want)
		}
		if c := len(s.asked(0)); c != 1 {
			t.Errorf("n=%d: order matters, so Min==Max==len(Options) does not take the take-everything fast path: asked %d times", len(want), c)
		}
	}
}

func TestFR_TRG_003_OrderDecisionFallbackIsCollectionOrder(t *testing.T) {
	cases := map[string]func(s *scene){
		"timeout":    func(s *scene) { s.script(0, Answer{Timeout: true}) },
		"no chooser": func(s *scene) {},
		"repeat":     func(s *scene) { s.script(0, Answer{Picks: []int{2, 2, 0}}) },
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			s := newScene(t)
			setup(s)
			got := s.g.choose(orderDecision(3))
			if !reflect.DeepEqual(got, []int{0, 1, 2}) {
				t.Errorf("%s: an order decision falls back to collection order: got %v, want [0 1 2]", name, got)
			}
		})
	}
}

func TestRandomChooser_OrderDecisionIsAValidPermutation(t *testing.T) {
	s := newScene(t)
	ch := RandomChooser(11)
	for n := 2; n <= 6; n++ {
		d := orderDecision(n)
		for i := 0; i < 20; i++ {
			a := ch(s.g, d)
			if a.Timeout || !validPicks(d, a.Picks) || len(a.Picks) != n {
				t.Fatalf("n=%d: random order answer %v is not a permutation of the options", n, a.Picks)
			}
		}
	}
}

func TestRandomChooser_OrderDecisionIsNotAlwaysTheIdentity(t *testing.T) {
	s := newScene(t)
	ch := RandomChooser(5)
	d := orderDecision(5)
	identity := []int{0, 1, 2, 3, 4}
	for i := 0; i < 50; i++ {
		if !reflect.DeepEqual(ch(s.g, d).Picks, identity) {
			return
		}
	}
	t.Errorf("a random chooser must give random permutations for an order decision, but 50 draws were all the identity")
}
