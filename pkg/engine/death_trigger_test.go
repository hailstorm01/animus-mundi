package engine

import (
	"fmt"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for death ("When this is killed, ...") triggers fired by the death
// sweep (T-0008): cross-player APNAP (FR-TRG-003), the non-active player's own
// ordering (FR-TRG-002), buffering until the sweep (FR-TRG-004) and the
// face-down rule (FR-ZON-004). Collection order is battlefield order, then
// printed order (board directive #10).

// killedTrigger is one "When this is killed, <steps>" ability.
func killedTrigger(steps ...ability.Step) ability.Trigger {
	return ability.Trigger{Hook: ability.HookKilled, Effect: ability.Effect{Steps: steps}}
}

// mortal is a unit printing the given death triggers in order.
func mortal(name string, power, health int, trigs ...ability.Trigger) *Card {
	c := unit(name, power, health, Keywords{})
	c.Ability = &ability.Def{Triggers: trigs}
	return c
}

// --- FR-TRG-003: cross-player APNAP ---

func TestFR_TRG_003_NonActivePlayersDeathTriggerResolvesFirst(t *testing.T) {
	// Both units kill each other. Both decks are empty and both triggers are
	// Draw 1, so whichever resolves first loses at its own sweep. The non-active
	// player's trigger sits on top of the stack, so player 1 loses and 0 wins.
	s := newScene(t)
	s.g.Players[0].Deck = nil
	s.g.Players[1].Deck = nil
	att := s.put(0, ZoneMainTheater, mortal("Attacker", 1, 1, killedTrigger(drawN(1))))
	def := s.put(1, ZoneMainTheater, mortal("Defender", 1, 1, killedTrigger(drawN(1))))
	s.do(s.attack(att, def))
	if s.alive(att) || s.alive(def) {
		t.Fatalf("precondition: both units are destroyed by the combat (attacker alive=%v, defender alive=%v)", s.alive(att), s.alive(def))
	}
	if !s.g.Outcome.Over || s.g.Outcome.Draw || s.g.Outcome.Winner != 0 {
		t.Errorf("the non-active player's death trigger resolves first (FR-TRG-003): player 1 draws from an empty deck and loses, outcome %+v", s.g.Outcome)
	}
}

func TestFR_TRG_003_NonActiveIsRelativeToTheTurn(t *testing.T) {
	// The same combat on player 1's turn: now player 0 is non-active and resolves first.
	s := newScene(t)
	s.endTurn()
	s.g.Players[0].Deck = nil
	s.g.Players[1].Deck = nil
	att := s.put(1, ZoneMainTheater, mortal("Attacker", 1, 1, killedTrigger(drawN(1))))
	def := s.put(0, ZoneMainTheater, mortal("Defender", 1, 1, killedTrigger(drawN(1))))
	s.do(s.attack(att, def))
	if !s.g.Outcome.Over || s.g.Outcome.Draw || s.g.Outcome.Winner != 1 {
		t.Errorf("on player 1's turn player 0 is non-active and resolves first, so player 0 loses (FR-TRG-003): outcome %+v", s.g.Outcome)
	}
}

// --- FR-TRG-002 / FR-TRG-003: each player orders their own death triggers ---

// berserkKillsTwo has player 0's Berserker kill both of player 1's units in one
// sweep. Player 1's deck top is A, B.
func berserkKillsTwo(s *scene, d, sec *Card) []*Instance {
	s.t.Helper()
	cards := s.onTop(1, vanilla("A", "B")...)
	att := s.put(0, ZoneMainTheater, unit("Berserker", 5, 9, Keywords{Berserk: 2}))
	dv := s.put(1, ZoneMainTheater, d)
	sv := s.put(1, ZoneMainTheater, sec)
	s.do(s.cleave(att, dv, sv))
	if s.alive(dv) || s.alive(sv) {
		s.t.Fatalf("precondition: the defender and the secondary target both die in one sweep")
	}
	return cards
}

func TestFR_TRG_002_NonActivePlayerOrdersTheirOwnDeathTriggers(t *testing.T) {
	s := newScene(t)
	s.script(0)
	s.script(1, Answer{Picks: []int{0, 1}})
	berserkKillsTwo(s,
		mortal("Defender", 1, 1, killedTrigger(drawN(1))),
		mortal("Bystander", 0, 1, killedTrigger(scratchN(1))))
	asked := s.asked(1)
	if len(asked) != 1 {
		t.Fatalf("the non-active player with two simultaneous triggers is asked once, through their own chooser: asked %d times (FR-TRG-002)", len(asked))
	}
	if d := asked[0]; d.Kind != DecideOrder || d.Player != 1 || len(d.Options) != 2 || d.Min != 2 || d.Max != 2 {
		t.Errorf("the decision orders both triggers for the non-active player (DecideOrder, 2 options, Min=Max=2, Player 1): %+v", d)
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("the active player has no triggers and is not asked about the non-active player's: asked %d times (FR-TRG-002)", n)
	}
}

func TestFR_TRG_003_NonActivePlayersOrderAnswerIsApplied(t *testing.T) {
	// Options are in collection order: 0 = Defender's Draw, 1 = Bystander's Scratch.
	// Answer [1 0] places Scratch first and Draw on top, so Draw resolves first.
	s := newScene(t)
	s.script(0)
	s.script(1, Answer{Picks: []int{1, 0}})
	cards := berserkKillsTwo(s,
		mortal("Defender", 1, 1, killedTrigger(drawN(1))),
		mortal("Bystander", 0, 1, killedTrigger(scratchN(1))))
	p := s.g.Players[1]
	if !inHand(p, cards[0]) || !scratched(p, cards[1]) {
		t.Errorf("answer [1 0]: the Draw trigger resolves first and takes the top card, then Scratch takes the next (FR-TRG-003)")
	}

	s2 := newScene(t)
	s2.script(0)
	s2.script(1, Answer{Picks: []int{0, 1}})
	cards2 := berserkKillsTwo(s2,
		mortal("Defender", 1, 1, killedTrigger(drawN(1))),
		mortal("Bystander", 0, 1, killedTrigger(scratchN(1))))
	p2 := s2.g.Players[1]
	if !scratched(p2, cards2[0]) || !inHand(p2, cards2[1]) {
		t.Errorf("answer [0 1]: the Scratch trigger resolves first and takes the top card, then Draw takes the next (FR-TRG-003)")
	}
}

func TestFR_TRG_002_BothPlayersWithTwoTriggersAreEachAskedOnceAboutTheirOwn(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Timeout: true})
	s.script(1, Answer{Timeout: true})
	att := s.put(0, ZoneMainTheater, mortal("Attacker", 1, 1, killedTrigger(drawN(1)), killedTrigger(drawN(1))))
	def := s.put(1, ZoneMainTheater, mortal("Defender", 1, 1, killedTrigger(drawN(1)), killedTrigger(drawN(1))))
	s.do(s.attack(att, def))
	for p := PlayerID(0); p < 2; p++ {
		asked := s.asked(p)
		if len(asked) != 1 {
			t.Fatalf("player %d has two simultaneous triggers and is asked once: asked %d times (FR-TRG-003)", p, len(asked))
		}
		if asked[0].Kind != DecideOrder || asked[0].Player != p || len(asked[0].Options) != 2 {
			t.Errorf("player %d orders only their own two triggers: %+v", p, asked[0])
		}
	}
}

// --- collection order fallback (directive #10) ---

func TestFR_TRG_003_DeathTriggersOfOneAssetFallBackToPrintedOrder(t *testing.T) {
	// 0 = Draw, 1 = Scratch. Printed order places Draw first, so Scratch resolves
	// first and takes the top card.
	s := newScene(t)
	cards := s.onTop(1, vanilla("A", "B")...)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 5, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, mortal("Defender", 1, 1, killedTrigger(drawN(1)), killedTrigger(scratchN(1))))
	s.do(s.attack(att, def))
	p := s.g.Players[1]
	if !scratched(p, cards[0]) || !inHand(p, cards[1]) {
		t.Errorf("without a chooser one asset's death triggers are placed in printed order, so the last printed resolves first")
	}
}

func TestFR_TRG_003_SimultaneousDeathsFallBackToBattlefieldOrder(t *testing.T) {
	// The earlier asset in battlefield order (Defender, Draw) is placed first and
	// resolves last; the later one (Bystander, Scratch) resolves first.
	s := newScene(t)
	cards := berserkKillsTwo(s,
		mortal("Defender", 1, 1, killedTrigger(drawN(1))),
		mortal("Bystander", 0, 1, killedTrigger(scratchN(1))))
	p := s.g.Players[1]
	if !scratched(p, cards[0]) || !inHand(p, cards[1]) {
		t.Errorf("collection order is battlefield order: the earlier asset's trigger is placed first and resolves last (directive #10)")
	}
}

// --- FR-TRG-004: death triggers wait for the sweep, and resolve without their source ---

func TestFR_TRG_004_DeathTriggersAreOrderedAfterTheSweepAndResolveWithoutTheirSource(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[1]
	cards := s.onTop(1, vanilla("A", "B")...)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 5, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, mortal("Defender", 1, 1, killedTrigger(drawN(1)), killedTrigger(scratchN(1))))
	handBefore := len(p.Hand)
	var handAtAsk int
	var aliveAtAsk bool
	s.g.choosers[1] = func(_ *Game, _ Decision) Answer {
		handAtAsk = len(p.Hand)
		aliveAtAsk = s.alive(def)
		return Answer{Picks: []int{0, 1}}
	}
	s.do(s.attack(att, def))
	if aliveAtAsk {
		t.Errorf("the sweep has completed before the death triggers are placed (FR-TRG-004): source still on the battlefield")
	}
	if handAtAsk != handBefore {
		t.Errorf("no death trigger has resolved when the player orders them: hand %d, want %d (FR-TRG-004)", handAtAsk, handBefore)
	}
	if !scratched(p, cards[0]) || !inHand(p, cards[1]) {
		t.Errorf("both death triggers resolve although their source has left the battlefield")
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty once the death triggers have resolved: %d items", len(s.g.Stack))
	}
}

func TestFR_TRG_004_SingleDeathTriggerResolvesWithoutQuestion(t *testing.T) {
	s := newScene(t)
	cards := s.onTop(1, vanilla("A", "B")...)
	s.script(0)
	s.script(1)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 5, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, mortal("Defender", 1, 1, killedTrigger(drawN(1))))
	s.do(s.attack(att, def))
	p := s.g.Players[1]
	if !inHand(p, cards[0]) || inHand(p, cards[1]) {
		t.Errorf("a Draw 1 death trigger draws exactly the top card")
	}
	if len(s.asked(0))+len(s.asked(1)) != 0 {
		t.Errorf("one trigger needs no ordering question (FR-TRG-003)")
	}
}

func TestFR_TRG_004_MatchEndedBySweepResolvesNoDeathTrigger(t *testing.T) {
	// The attacker kills the nearly dead HQ and dies in the same sweep. The match
	// is over, so the attacker's death trigger never resolves.
	s := newScene(t)
	cards := s.onTop(0, vanilla("A")...)
	s.g.Players[1].HQ.Damage = 19
	att := s.put(0, ZoneMainTheater, mortal("Attacker", 1, 1, killedTrigger(drawN(1))))
	s.do(s.attack(att, s.g.Players[1].HQ))
	if s.alive(att) {
		t.Fatalf("precondition: the attacker dies to the HQ's combat damage")
	}
	if !s.g.Outcome.Over || s.g.Outcome.Winner != 0 {
		t.Fatalf("precondition: destroying the HQ wins the match for player 0: %+v", s.g.Outcome)
	}
	if inHand(s.g.Players[0], cards[0]) {
		t.Errorf("the match ended at the sweep, so the buffered death trigger is discarded (FR-TRG-004)")
	}
}

// --- FR-ZON-004: a face-down Scratch Company has no abilities ---

func TestFR_ZON_004_FaceDownScratchCompanyFiresNoDeathTrigger(t *testing.T) {
	for _, faceDown := range []bool{true, false} {
		s := newScene(t)
		cards := s.onTop(1, vanilla("A", "B")...)
		s.script(1, Answer{Timeout: true})
		att := s.put(0, ZoneMainTheater, unit("Hitter", 5, 9, Keywords{}))
		def := s.put(1, ZoneMainTheater, mortal("Printed", 1, 1, killedTrigger(drawN(1)), killedTrigger(drawN(1))))
		def.FaceDown = faceDown
		s.do(s.attack(att, def))
		if s.alive(def) {
			t.Fatalf("faceDown=%v: precondition: the unit is destroyed", faceDown)
		}
		p := s.g.Players[1]
		fired := inHand(p, cards[0])
		if faceDown && (fired || len(s.asked(1)) != 0) {
			t.Errorf("a face-down Scratch Company never fires printed death triggers, and is not asked to order any (FR-ZON-004)")
		}
		if !faceDown && !fired {
			t.Errorf("control: the same face-up unit does fire its printed death triggers")
		}
	}
}

// --- playouts: a rejected chooser answer is an error ---

// deployPolicy deploys to Reserves whenever it can, else ends the turn.
func deployPolicy(_ *Game, legal []Action) Action {
	for _, a := range legal {
		if a.Kind == ActDeploy && a.Zone == ZoneReserves {
			return a
		}
	}
	for _, a := range legal {
		if a.Kind == ActEndTurn {
			return a
		}
	}
	return legal[0]
}

func twoTriggerDeck() DeckList {
	main := make([]*Card, 50)
	for i := range main {
		main[i] = trigUnit("Twin", deployTrigger(drawN(1)), deployTrigger(scratchN(1)))
	}
	return DeckList{HQ: hq(3, 20), Main: main}
}

// playDeploy plays g to completion with deployPolicy for both seats, as sim.Playout
// would (sim imports engine, so it cannot be called here): it checks invariants
// after every action and returns an error on a rejected chooser answer.
func playDeploy(g *Game) error {
	for !g.Outcome.Over {
		if g.Turn > 400 {
			return fmt.Errorf("match did not end within %d turns", 400)
		}
		a := deployPolicy(g, g.LegalActions())
		logged := len(g.Log)
		if err := g.Apply(a); err != nil {
			return fmt.Errorf("turn %d: policy chose an action LegalActions offered but Apply rejected: %w", g.Turn, err)
		}
		for _, e := range g.Log[logged:] {
			if e.Kind == "invalid-answer" {
				return fmt.Errorf("turn %d after %s: a chooser answer was rejected: %s", g.Turn, a.Kind, e.Detail)
			}
		}
		if err := g.CheckInvariants(); err != nil {
			return fmt.Errorf("turn %d after %s: %w", g.Turn, a.Kind, err)
		}
	}
	return nil
}

func playoutWith(t *testing.T, ch Chooser) error {
	t.Helper()
	g, err := NewGame(Config{Seed: 3, Decks: [2]DeckList{twoTriggerDeck(), twoTriggerDeck()}, Choosers: [2]Chooser{ch, ch}})
	if err != nil {
		t.Fatal(err)
	}
	return playDeploy(g)
}

func TestPlayout_InvalidOrderAnswerIsAnError(t *testing.T) {
	bad := func(_ *Game, _ Decision) Answer { return Answer{Picks: []int{0, 0}} }
	if err := playoutWith(t, bad); err == nil {
		t.Errorf("a playout whose chooser gives an invalid ordering answer must return an error")
	}
}

func TestPlayout_ValidAndTimedOutOrderAnswersAreNotErrors(t *testing.T) {
	good := func(_ *Game, d Decision) Answer {
		picks := make([]int, len(d.Options))
		for i := range picks {
			picks[i] = i
		}
		return Answer{Picks: picks}
	}
	if err := playoutWith(t, good); err != nil {
		t.Errorf("control: valid ordering answers must not fail a playout: %v", err)
	}
	timeout := func(_ *Game, _ Decision) Answer { return Answer{Timeout: true} }
	if err := playoutWith(t, timeout); err != nil {
		t.Errorf("control: a timed-out answer is not a rejected answer: %v", err)
	}
}
