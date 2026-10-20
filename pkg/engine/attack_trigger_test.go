package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for attack triggers placed in the combat Step 2 window together with
// Aggressive (FR-CBT-002, FR-KWD-002f), the "on your turn" gate on triggers,
// and the opponents subject (T-0010). The turn gate and the opponents subject
// have no dedicated FR in FRD 01; those tests use the TestShape_ prefix.

// attackTrigger is one "Whenever this unit attacks, <steps>" ability.
func attackTrigger(steps ...ability.Step) ability.Trigger {
	return ability.Trigger{Hook: ability.HookAttacks, Effect: ability.Effect{Steps: steps}}
}

// striker is a unit with the given keywords printing the given triggers.
func striker(name string, power, health int, kw Keywords, trigs ...ability.Trigger) *Card {
	c := unit(name, power, health, kw)
	c.Ability = &ability.Def{Triggers: trigs}
	return c
}

func opponentsDraw(n int) ability.Step {
	return ability.Step{Op: ability.OpDraw, Subject: ability.SubjectOpponents, Amount: ability.Value{N: n}}
}

// --- FR-CBT-002 Step 2: attack triggers resolve before combat damage ---

func TestFR_CBT_002_AttackTriggerResolvesBeforeCombatDamage(t *testing.T) {
	// The attacker's Draw 1 attack trigger hits an empty deck. The attacker's
	// controller loses at that trigger's sweep, before the Attack resolves, so
	// the Headquarters takes no combat damage.
	s := newScene(t)
	s.g.Players[0].Deck = nil
	att := s.put(0, ZoneMainTheater, striker("Scout", 3, 3, Keywords{}, attackTrigger(drawN(1))))
	s.do(s.attack(att, s.g.Players[1].HQ))
	if o := s.g.Outcome; !o.Over || o.Draw || o.Winner != 1 {
		t.Errorf("a required draw from an empty deck in the Step 2 trigger loses the attacker's controller (FR-WIN-001): outcome %+v", o)
	}
	if d := s.g.Players[1].HQ.Damage; d != 0 {
		t.Errorf("attack triggers resolve before combat damage (FR-CBT-002 Step 2), so the defender took no damage: HQ damage %d", d)
	}
}

func TestFR_CBT_002_AttackTriggerScratchHappensBeforeCombat(t *testing.T) {
	s := newScene(t)
	cards := s.onTop(0, vanilla("A", "B")...)
	att := s.put(0, ZoneMainTheater, striker("Despoiler", 1, 5, Keywords{}, attackTrigger(scratchN(1))))
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 10, Keywords{}))
	s.do(s.attack(att, def))
	p := s.g.Players[0]
	if !scratched(p, cards[0]) || faceDownCount(p) != 1 {
		t.Errorf("the attack trigger puts the top card into Reserves as a face-down Scratch Company: face-down count %d", faceDownCount(p))
	}
	s.wantDamage(def, 1)
}

func TestFR_CBT_002_AttackTriggerFiresOncePerAttack(t *testing.T) {
	s := newScene(t)
	cards := s.onTop(0, vanilla("A", "B", "C")...)
	att := s.put(0, ZoneMainTheater, striker("Scout", 1, 9, Keywords{}, attackTrigger(drawN(1))))
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 10, Keywords{}))
	s.do(s.attack(att, def))
	p := s.g.Players[0]
	if !inHand(p, cards[0]) || inHand(p, cards[1]) {
		t.Errorf("one attack fires the Draw 1 attack trigger exactly once (FR-CBT-002)")
	}
}

func TestFR_CBT_002_OnlyTheAttackerFiresItsAttackTrigger(t *testing.T) {
	s := newScene(t)
	cards := s.onTop(1, vanilla("A")...)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 1, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, striker("Defender", 0, 10, Keywords{}, attackTrigger(drawN(1))))
	s.do(s.attack(att, def))
	if inHand(s.g.Players[1], cards[0]) {
		t.Errorf("a defender's attack trigger does not fire when it is attacked (FR-CBT-002)")
	}
}

// --- FR-CBT-002 / FR-TRG-003: Aggressive and the attack hook share the Step 2 window ---

func TestFR_CBT_002_AggressiveAndAttackTriggerAreOrderedTogether(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Picks: []int{0, 1}})
	s.script(1)
	att := s.put(0, ZoneMainTheater, striker("Raider", 1, 9, Keywords{Aggressive: 2}, attackTrigger(scratchN(1))))
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 10, Keywords{}))
	s.do(s.attack(att, def))
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("Aggressive and a printed attack trigger are two simultaneous Step 2 triggers, so the controller is asked once: asked %d times (FR-CBT-002, FR-TRG-003)", len(asked))
	}
	if d := asked[0]; d.Kind != DecideOrder || d.Player != 0 || len(d.Options) != 2 || d.Min != 2 || d.Max != 2 {
		t.Errorf("the decision orders both triggers (DecideOrder, 2 options, Min=Max=2, Player 0): %+v", d)
	}
	if n := len(s.asked(1)); n != 0 {
		t.Errorf("the defender has no Step 2 triggers and is not asked: asked %d times", n)
	}
}

func TestFR_CBT_002_AggressiveAloneIsNotAskedAbout(t *testing.T) {
	s := newScene(t)
	s.script(0)
	att := s.put(0, ZoneMainTheater, unit("Brawler", 1, 9, Keywords{Aggressive: 2}))
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 10, Keywords{}))
	s.do(s.attack(att, def))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a single Step 2 trigger needs no ordering question (FR-TRG-003): asked %d times", n)
	}
}

func TestFR_CBT_002_AttackTriggerAloneIsNotAskedAbout(t *testing.T) {
	s := newScene(t)
	s.script(0)
	att := s.put(0, ZoneMainTheater, striker("Scout", 1, 9, Keywords{}, attackTrigger(scratchN(1))))
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 10, Keywords{}))
	s.do(s.attack(att, def))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a single Step 2 trigger needs no ordering question (FR-TRG-003): asked %d times", n)
	}
}

func TestFR_CBT_002_TwoPrintedAttackTriggersAreOrderedTogether(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Timeout: true})
	att := s.put(0, ZoneMainTheater, striker("Twin", 1, 9, Keywords{}, attackTrigger(scratchN(1)), attackTrigger(drawN(1))))
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 10, Keywords{}))
	s.do(s.attack(att, def))
	asked := s.asked(0)
	if len(asked) != 1 || asked[0].Kind != DecideOrder || len(asked[0].Options) != 2 {
		t.Errorf("two printed attack triggers are ordered with one DecideOrder of 2 options (FR-TRG-003): %+v", asked)
	}
}

func TestFR_CBT_002_OrderAnswerDecidesWhichPrintedAttackTriggerResolvesFirst(t *testing.T) {
	// Options are in collection order: 0 = Scratch, 1 = Draw (printed order).
	// The first pick is placed first and resolves last.
	for _, tc := range []struct {
		picks        []int
		scratchFirst bool
	}{{[]int{0, 1}, false}, {[]int{1, 0}, true}} {
		s := newScene(t)
		cards := s.onTop(0, vanilla("A", "B")...)
		s.script(0, Answer{Picks: tc.picks})
		att := s.put(0, ZoneMainTheater, striker("Twin", 1, 9, Keywords{}, attackTrigger(scratchN(1)), attackTrigger(drawN(1))))
		def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 10, Keywords{}))
		s.do(s.attack(att, def))
		p := s.g.Players[0]
		if tc.scratchFirst {
			if !scratched(p, cards[0]) || !inHand(p, cards[1]) {
				t.Errorf("answer %v: Scratch resolves first and takes the top card, then Draw (FR-TRG-003)", tc.picks)
			}
		} else if !inHand(p, cards[0]) || !scratched(p, cards[1]) {
			t.Errorf("answer %v: Draw resolves first and takes the top card, then Scratch (FR-TRG-003)", tc.picks)
		}
	}
}

// --- FR-KWD-002f: Aggressive still gives +N for the combat ---

func TestFR_KWD_002f_AggressiveStillGivesPowerWithAnAttackTrigger(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Timeout: true})
	att := s.put(0, ZoneMainTheater, striker("Raider", 1, 9, Keywords{Aggressive: 2}, attackTrigger(scratchN(1))))
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 10, Keywords{}))
	s.do(s.attack(att, def))
	s.wantDamage(def, 3)
	s.endTurn()
	if p := s.g.Power(att); p != 1 {
		t.Errorf("the Aggressive bonus expires at end of turn: power %d, want 1", p)
	}
}

func TestFR_KWD_002f_AggressiveAloneStillGivesPowerForTheCombat(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Brawler", 2, 9, Keywords{Aggressive: 3}))
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 10, Keywords{}))
	s.do(s.attack(att, def))
	s.wantDamage(def, 5)
}

// --- FR-WIN-001 with the opponents subject ---

func TestFR_WIN_001_OpponentsDrawFromEmptyDeckLosesAtThatSweep(t *testing.T) {
	s := newScene(t)
	s.g.Players[1].Deck = nil
	att := s.put(0, ZoneMainTheater, mortal("Worm", 1, 1, killedTrigger(opponentsDraw(2))))
	def := s.put(1, ZoneMainTheater, unit("Wall", 5, 5, Keywords{}))
	s.do(s.attack(att, def))
	if s.alive(att) {
		t.Fatalf("precondition: the Worm dies in combat")
	}
	if o := s.g.Outcome; !o.Over || o.Draw || o.Winner != 0 {
		t.Errorf("the opponent's required draw from an empty deck loses them at the sweep (FR-WIN-001): outcome %+v", o)
	}
}

func TestFR_WIN_001_ControllersOwnEmptyDeckIsUnaffectedByOpponentsDraw(t *testing.T) {
	s := newScene(t)
	s.g.Players[0].Deck = nil
	s.onTop(1, vanilla("A", "B")...)
	att := s.put(0, ZoneMainTheater, mortal("Worm", 1, 1, killedTrigger(opponentsDraw(2))))
	def := s.put(1, ZoneMainTheater, unit("Wall", 5, 5, Keywords{}))
	s.do(s.attack(att, def))
	if s.g.Outcome.Over {
		t.Errorf("only the opponent draws, so the controller's empty deck is not a loss (FR-WIN-001): outcome %+v", s.g.Outcome)
	}
}

func TestShape_OpponentsDrawIsPerformedByTheOpponentNotTheController(t *testing.T) {
	s := newScene(t)
	mine := s.onTop(0, vanilla("X")...)
	theirs := s.onTop(1, vanilla("A", "B", "C")...)
	att := s.put(0, ZoneMainTheater, mortal("Worm", 1, 1, killedTrigger(opponentsDraw(2))))
	def := s.put(1, ZoneMainTheater, unit("Wall", 5, 5, Keywords{}))
	s.do(s.attack(att, def))
	p0, p1 := s.g.Players[0], s.g.Players[1]
	if !inHand(p1, theirs[0]) || !inHand(p1, theirs[1]) || inHand(p1, theirs[2]) {
		t.Errorf("\"all opponents draw 2\" gives the opponent exactly their top two cards")
	}
	if inHand(p0, mine[0]) {
		t.Errorf("the controller draws nothing from an opponents-subject draw")
	}
}

func TestShape_OpponentsScratchIsPerformedByTheOpponent(t *testing.T) {
	s := newScene(t)
	mine := s.onTop(0, vanilla("X")...)
	theirs := s.onTop(1, vanilla("A", "B")...)
	att := s.put(0, ZoneMainTheater, mortal("Worm", 1, 1, killedTrigger(
		ability.Step{Op: ability.OpScratch, Subject: ability.SubjectOpponents, Amount: ability.Value{N: 1}})))
	def := s.put(1, ZoneMainTheater, unit("Wall", 5, 5, Keywords{}))
	s.do(s.attack(att, def))
	if !scratched(s.g.Players[1], theirs[0]) {
		t.Errorf("the opponent scratches their own top card")
	}
	if scratched(s.g.Players[0], mine[0]) {
		t.Errorf("the controller does not scratch for an opponents-subject Scratch")
	}
}

// --- turn gate ---

// gatedKilled is a death trigger gated by the given turn owner that draws one card.
func gatedKilled(turn ability.TurnOwner) ability.Trigger {
	return ability.Trigger{Hook: ability.HookKilled, Turn: turn, Effect: ability.Effect{Steps: []ability.Step{drawN(1)}}}
}

func TestShape_TurnYoursDeathTriggerFiresOnItsControllersTurn(t *testing.T) {
	s := newScene(t)
	cards := s.onTop(0, vanilla("A", "B")...)
	att := s.put(0, ZoneMainTheater, striker("Scrivener", 1, 1, Keywords{}, gatedKilled(ability.TurnYours)))
	def := s.put(1, ZoneMainTheater, unit("Wall", 5, 5, Keywords{}))
	s.do(s.attack(att, def))
	if s.alive(att) {
		t.Fatalf("precondition: the attacker dies")
	}
	if !inHand(s.g.Players[0], cards[0]) {
		t.Errorf("an on-your-turn death trigger fires when its controller's unit dies during their own turn")
	}
}

func TestShape_TurnYoursDeathTriggerDoesNotFireOnTheOpponentsTurn(t *testing.T) {
	s := newScene(t)
	cards := s.onTop(0, vanilla("A", "B")...)
	s.endTurn()
	mine := s.put(0, ZoneMainTheater, striker("Scrivener", 1, 1, Keywords{}, gatedKilled(ability.TurnYours)))
	hitter := s.put(1, ZoneMainTheater, unit("Hitter", 5, 5, Keywords{}))
	s.do(s.attack(hitter, mine))
	if s.alive(mine) {
		t.Fatalf("precondition: the defender dies")
	}
	if inHand(s.g.Players[0], cards[0]) {
		t.Errorf("an on-your-turn death trigger does not fire when the unit dies on the opponent's turn")
	}
}

func TestShape_UngatedDeathTriggerFiresOnEitherTurn(t *testing.T) {
	s := newScene(t)
	cards := s.onTop(0, vanilla("A", "B")...)
	att := s.put(0, ZoneMainTheater, striker("Twin", 1, 1, Keywords{}, gatedKilled(ability.TurnAny)))
	def := s.put(1, ZoneMainTheater, unit("Wall", 5, 5, Keywords{}))
	s.do(s.attack(att, def))
	if !inHand(s.g.Players[0], cards[0]) {
		t.Errorf("an ungated death trigger fires on its controller's turn")
	}

	s2 := newScene(t)
	cards2 := s2.onTop(0, vanilla("A", "B")...)
	s2.endTurn()
	mine := s2.put(0, ZoneMainTheater, striker("Twin", 1, 1, Keywords{}, gatedKilled(ability.TurnAny)))
	hitter := s2.put(1, ZoneMainTheater, unit("Hitter", 5, 5, Keywords{}))
	s2.do(s2.attack(hitter, mine))
	if !inHand(s2.g.Players[0], cards2[0]) {
		t.Errorf("an ungated death trigger also fires on the opponent's turn")
	}
}

func TestShape_TurnYoursAttackTriggerFiresWhenAttackingOnYourTurn(t *testing.T) {
	// An attack always happens on the attacker's controller's turn, so a
	// gated attack trigger still fires.
	s := newScene(t)
	cards := s.onTop(0, vanilla("A")...)
	trig := ability.Trigger{Hook: ability.HookAttacks, Turn: ability.TurnYours, Effect: ability.Effect{Steps: []ability.Step{drawN(1)}}}
	att := s.put(0, ZoneMainTheater, striker("Scout", 1, 9, Keywords{}, trig))
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 10, Keywords{}))
	s.do(s.attack(att, def))
	if !inHand(s.g.Players[0], cards[0]) {
		t.Errorf("an on-your-turn attack trigger fires when its controller attacks on their turn")
	}
}
