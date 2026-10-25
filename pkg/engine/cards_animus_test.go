package engine

import (
	"reflect"
	"testing"
)

// Card tests for Animus cards. Each is written from the card's CSV rules text
// in data/cards/units-and-structures-animus.csv and loads the real card with
// s.card. Later Animus card tasks append their tests here.

// deployCreeper gives the real Spoil Forest Creeper (cost 3 Requisition) to
// player 0 with enough Requisition and declares its deploy into Reserves.
func deployCreeper(s *scene) *Instance {
	s.t.Helper()
	inst := s.give(0, s.card("Spoil Forest Creeper"))
	s.g.Players[0].Req = 3
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	return inst
}

// "When this unit deploys, Scratch 1."
// Covers: FR-KWD-002m, FR-ZON-005, FR-TRG-004
func TestCard_SpoilForestCreeper_DeployScratchesOne(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	top := s.onTop(0, vanilla("Top", "Next")...)
	deckBefore := len(p.Deck)
	deployCreeper(s)
	if !scratched(p, top[0]) {
		t.Errorf("deploying the Creeper scratches the top card of its controller's deck into Reserves as a face-down Scratch Company (FR-KWD-002m, FR-ZON-005)")
	}
	if n := faceDownCount(p); n != 1 {
		t.Errorf("Scratch 1 creates exactly one Scratch Company: got %d (FR-KWD-002m)", n)
	}
	if len(p.Deck) != deckBefore-1 || p.Deck[0] != top[1] {
		t.Errorf("Scratch 1 takes exactly one card, leaving the second card on top of the deck (FR-KWD-002m)")
	}
}

// Covers: FR-ZON-005
func TestCard_SpoilForestCreeper_ScratchCompanyEntersExhausted(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	top := s.onTop(0, vanilla("Top")...)
	deployCreeper(s)
	for _, a := range p.Reserves {
		if a.Inst == top[0] {
			if !a.FaceDown || !a.Exhausted || a.Zone != ZoneReserves {
				t.Errorf("a Scratch Company enters face-down and Exhausted in Reserves: faceDown=%v exhausted=%v zone=%v (FR-ZON-005)", a.FaceDown, a.Exhausted, a.Zone)
			}
			return
		}
	}
	t.Errorf("the scratched card is a Scratch Company in Reserves (FR-ZON-005)")
}

func TestCard_SpoilForestCreeper_ItselfIsFaceUpThreeTwoInReserves(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	inst := deployCreeper(s)
	for _, a := range p.Reserves {
		if a.Inst == inst {
			if a.FaceDown {
				t.Errorf("the deployed Creeper is face-up; only the scratched card is face-down")
			}
			if c := a.Printed(); c.Power != 3 || c.Health != 2 {
				t.Errorf("the Creeper is a 3/2 unit: got %d/%d", c.Power, c.Health)
			}
			if inHand(p, inst) {
				t.Errorf("the Creeper left the hand when deployed")
			}
			return
		}
	}
	t.Errorf("the Creeper is on the battlefield in Reserves after deploying (FR-TRG-004)")
}

func TestCard_SpoilForestCreeper_DeployPaysThreeRequisition(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	deployCreeper(s)
	if p.Req != 0 {
		t.Errorf("deploying with exactly 3 Requisition spends all of it: %d left", p.Req)
	}
}

// Covers: FR-KWD-002m
func TestCard_SpoilForestCreeper_ScratchOnEmptyDeckDoesNothing(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	p.Deck = nil
	inst := deployCreeper(s)
	if n := faceDownCount(p); n != 0 {
		t.Errorf("Scratch 1 on an empty deck creates no Scratch Company: got %d (FR-KWD-002m)", n)
	}
	if len(p.Reserves) != 1 || p.Reserves[0].Inst != inst {
		t.Errorf("only the Creeper itself is in Reserves after a Scratch on an empty deck: %d assets", len(p.Reserves))
	}
	if s.g.Outcome.Over {
		t.Errorf("Scratch on an empty deck is not a loss: %+v (FR-KWD-002m)", s.g.Outcome)
	}
}

// Covers: FR-ZON-005, FR-TRG-003
func TestCard_SpoilForestCreeper_FaceDownCopyDoesNotTrigger(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	cards := s.onTop(0, s.card("Spoil Forest Creeper"), vanilla("B")[0], vanilla("C")[0])
	s.script(0)
	s.do(Action{Kind: ActScratch, Player: 0, Asset: p.HQ.ID})
	if !scratched(p, cards[0]) {
		t.Fatalf("the Creeper on top of the deck became a face-down Scratch Company")
	}
	if n := faceDownCount(p); n != 1 {
		t.Errorf("a face-down Scratch Company has no printed abilities, so no second Scratch happens: %d face-down (FR-ZON-005)", n)
	}
	if len(p.Deck) == 0 || p.Deck[0] != cards[1] {
		t.Errorf("the face-down Creeper does not take another card from the deck (FR-ZON-005)")
	}
}

// --- Despoiler Larva: "Whenever this unit attacks, Scratch 1." ---

// larvaAttack puts the real Despoiler Larva (face-down if asked) into the Main
// Theater for player 0 and attacks a 0/10 dummy with it. top are the first
// cards of player 0's deck.
func larvaAttack(s *scene, faceDown bool, top ...*Card) (dummy *Asset, cards []*Instance) {
	s.t.Helper()
	cards = s.onTop(0, top...)
	larva := s.put(0, ZoneMainTheater, s.card("Despoiler Larva"))
	larva.FaceDown = faceDown
	dummy = s.put(1, ZoneMainTheater, unit("Dummy", 0, 10, Keywords{}))
	s.do(s.attack(larva, dummy))
	return dummy, cards
}

// Covers: FR-CBT-002, FR-KWD-002m
func TestCard_DespoilerLarva_AttackScratchesOne(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	_, cards := larvaAttack(s, false, vanilla("A", "B")...)
	if !scratched(p, cards[0]) {
		t.Errorf("attacking with the Larva scratches the top card of its controller's deck into a face-down Scratch Company (FR-KWD-002m)")
	}
	if n := faceDownCount(p); n != 1 {
		t.Errorf("one attack fires Scratch 1 once, creating exactly one Scratch Company: got %d (FR-CBT-002, FR-KWD-002m)", n)
	}
	if len(p.Deck) == 0 || p.Deck[0] != cards[1] {
		t.Errorf("Scratch 1 takes exactly one card, leaving the second card on top (FR-KWD-002m)")
	}
}

// Covers: FR-CBT-002
func TestCard_DespoilerLarva_ScratchResolvesBeforeCombatDamage(t *testing.T) {
	// The Larva deals lethal damage to the 19-damage Headquarters. The Step 2
	// Scratch resolves before that damage, so it happens although the match ends
	// in the same Attack.
	s := newScene(t)
	p := s.g.Players[0]
	cards := s.onTop(0, vanilla("A")...)
	s.g.Players[1].HQ.Damage = 19
	larva := s.put(0, ZoneMainTheater, s.card("Despoiler Larva"))
	s.do(s.attack(larva, s.g.Players[1].HQ))
	if !s.g.Outcome.Over || s.g.Outcome.Winner != 0 {
		t.Fatalf("precondition: the Larva's 1 damage destroys the 19-damage Headquarters: %+v", s.g.Outcome)
	}
	if !scratched(p, cards[0]) {
		t.Errorf("the attack trigger resolves in Step 2, before combat damage ends the match (FR-CBT-002)")
	}
}

// Covers: FR-KWD-002m
func TestCard_DespoilerLarva_ScratchOnEmptyDeckDoesNothing(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	p.Deck = nil
	dummy, _ := larvaAttack(s, false)
	if n := faceDownCount(p); n != 0 {
		t.Errorf("Scratch 1 on an empty deck creates no Scratch Company: got %d (FR-KWD-002m)", n)
	}
	if s.g.Outcome.Over {
		t.Errorf("Scratch on an empty deck is not a loss: %+v (FR-KWD-002m)", s.g.Outcome)
	}
	s.wantDamage(dummy, 1)
}

// Covers: FR-ZON-005
func TestCard_DespoilerLarva_FaceDownCopyAttackingDoesNotScratch(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	_, cards := larvaAttack(s, true, vanilla("A", "B")...)
	if scratched(p, cards[0]) || len(p.Deck) == 0 || p.Deck[0] != cards[0] {
		t.Errorf("a face-down Scratch Company has no printed abilities, so attacking does not Scratch (FR-ZON-005)")
	}
}

// Covers: FR-CBT-002
func TestCard_DespoilerLarva_DefendingDoesNotScratch(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[1]
	cards := s.onTop(1, vanilla("A")...)
	larva := s.put(1, ZoneMainTheater, s.card("Despoiler Larva"))
	hit := s.put(0, ZoneMainTheater, unit("Hitter", 0, 9, Keywords{}))
	s.do(s.attack(hit, larva))
	if scratched(p, cards[0]) {
		t.Errorf("the Larva's trigger fires only when it attacks, not when it is attacked (FR-CBT-002)")
	}
}

// --- Fungal Evangelist: "When this unit deploys or is killed, Scratch 2." ---

func deployEvangelist(s *scene) {
	s.t.Helper()
	inst := s.give(0, s.card("Fungal Evangelist"))
	s.g.Players[0].Req = 6
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
}

// Covers: FR-KWD-002m
func TestCard_FungalEvangelist_DeployScratchesTwo(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	top := s.onTop(0, vanilla("A", "B", "C")...)
	deployEvangelist(s)
	if !scratched(p, top[0]) || !scratched(p, top[1]) {
		t.Errorf("deploying the Evangelist scratches the top two cards into face-down Scratch Companies (FR-KWD-002m)")
	}
	if n := faceDownCount(p); n != 2 {
		t.Errorf("Scratch 2 creates exactly two Scratch Companies: got %d (FR-KWD-002m)", n)
	}
	if len(p.Deck) == 0 || p.Deck[0] != top[2] {
		t.Errorf("Scratch 2 takes exactly two cards, leaving the third on top (FR-KWD-002m)")
	}
}

// Covers: FR-KWD-002m
func TestCard_FungalEvangelist_DeployWithOneCardLeftScratchesItAndDoesNotLose(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	p.Deck = nil
	top := s.onTop(0, vanilla("Only")...)
	deployEvangelist(s)
	if !scratched(p, top[0]) || faceDownCount(p) != 1 {
		t.Errorf("Scratch 2 on a one-card deck scratches that card and stops (FR-KWD-002m): %d face-down", faceDownCount(p))
	}
	if s.g.Outcome.Over {
		t.Errorf("Scratch on a short deck is not a loss: %+v (FR-KWD-002m)", s.g.Outcome)
	}
}

// Covers: FR-KWD-002m
func TestCard_FungalEvangelist_DeployOnEmptyDeckDoesNothing(t *testing.T) {
	s := newScene(t)
	s.g.Players[0].Deck = nil
	deployEvangelist(s)
	if n := faceDownCount(s.g.Players[0]); n != 0 || s.g.Outcome.Over {
		t.Errorf("Scratch 2 on an empty deck makes no Scratch Company and is no loss: %d face-down, outcome %+v (FR-KWD-002m)", n, s.g.Outcome)
	}
}

// killOnOpponentsTurn puts card c (face-down if asked) for player 0 into the
// Main Theater, passes the turn and has player 1's 9-power hitter destroy it.
func killOnOpponentsTurn(s *scene, c *Card, faceDown bool) {
	s.t.Helper()
	victim := s.put(0, ZoneMainTheater, c)
	victim.FaceDown = faceDown
	s.endTurn()
	hit := s.put(1, ZoneMainTheater, unit("Hitter", 9, 9, Keywords{}))
	s.do(s.attack(hit, victim))
	if s.alive(victim) {
		s.t.Fatalf("precondition: the victim is destroyed")
	}
}

// Covers: FR-ZON-004, FR-KWD-002m
func TestCard_FungalEvangelist_DeathScratchesTwoOnItsControllersTurn(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	top := s.onTop(0, vanilla("A", "B", "C")...)
	ev := s.put(0, ZoneMainTheater, s.card("Fungal Evangelist"))
	def := s.put(1, ZoneMainTheater, unit("Hitter", 9, 9, Keywords{}))
	s.do(s.attack(ev, def))
	if s.alive(ev) {
		t.Fatalf("precondition: the Evangelist is destroyed")
	}
	if !scratched(p, top[0]) || !scratched(p, top[1]) || faceDownCount(p) != 2 {
		t.Errorf("the Evangelist's death scratches two cards for its controller (FR-ZON-004, FR-KWD-002m): %d face-down", faceDownCount(p))
	}
}

// Covers: FR-ZON-004
func TestCard_FungalEvangelist_DeathOnOpponentsTurnScratchesForItsController(t *testing.T) {
	s := newScene(t)
	p0, p1 := s.g.Players[0], s.g.Players[1]
	top := s.onTop(0, vanilla("A", "B", "C")...)
	killOnOpponentsTurn(s, s.card("Fungal Evangelist"), false)
	if !scratched(p0, top[0]) || !scratched(p0, top[1]) || faceDownCount(p0) != 2 {
		t.Errorf("killed on the opponent's turn, the Evangelist still scratches two cards for its own controller (FR-ZON-004): %d face-down", faceDownCount(p0))
	}
	if faceDownCount(p1) != 0 {
		t.Errorf("the killer's controller does not Scratch: %d face-down (FR-ZON-004)", faceDownCount(p1))
	}
}

// Covers: FR-KWD-002m
func TestCard_FungalEvangelist_DeathOnEmptyDeckDoesNothing(t *testing.T) {
	s := newScene(t)
	s.g.Players[0].Deck = nil
	killOnOpponentsTurn(s, s.card("Fungal Evangelist"), false)
	if n := faceDownCount(s.g.Players[0]); n != 0 || s.g.Outcome.Over {
		t.Errorf("Scratch 2 on an empty deck makes no Scratch Company and is no loss: %d face-down, outcome %+v (FR-KWD-002m)", n, s.g.Outcome)
	}
}

// Covers: FR-ZON-004
func TestCard_FungalEvangelist_FaceDownCopyDoesNotScratchWhenKilled(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	top := s.onTop(0, vanilla("A", "B")...)
	killOnOpponentsTurn(s, s.card("Fungal Evangelist"), true)
	if scratched(p, top[0]) || len(p.Deck) == 0 || p.Deck[0] != top[0] {
		t.Errorf("a face-down Scratch Company has no abilities, so its death scratches nothing (FR-ZON-004)")
	}
}

// --- Titanic Spore Carrier: "When this unit is killed, Scratch 1 and draw a card." ---

// Covers: FR-ZON-004, FR-KWD-002m
func TestCard_TitanicSporeCarrier_DeathScratchesFirstCardThenDrawsSecond(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	top := s.onTop(0, vanilla("First", "Second", "Third")...)
	killOnOpponentsTurn(s, s.card("Titanic Spore Carrier"), false)
	if !scratched(p, top[0]) || inHand(p, top[0]) {
		t.Errorf("the top card becomes the Scratch Company, because Scratch comes first in the printed text (FR-KWD-002m)")
	}
	if !inHand(p, top[1]) || scratched(p, top[1]) {
		t.Errorf("the second card is the one drawn, because the draw follows the Scratch")
	}
	if inHand(p, top[2]) || scratched(p, top[2]) || len(p.Deck) == 0 || p.Deck[0] != top[2] {
		t.Errorf("exactly one card is scratched and one drawn; the third stays on top")
	}
}

// Covers: FR-WIN-001
func TestCard_TitanicSporeCarrier_DrawFromEmptyDeckLosesAtThatSweep(t *testing.T) {
	s := newScene(t)
	s.g.Players[0].Deck = nil
	killOnOpponentsTurn(s, s.card("Titanic Spore Carrier"), false)
	if o := s.g.Outcome; !o.Over || o.Draw || o.Winner != 1 {
		t.Errorf("the Carrier's required draw from an empty deck loses its controller at the trigger's sweep (FR-WIN-001): outcome %+v", o)
	}
}

// Covers: FR-WIN-001, FR-KWD-002m
func TestCard_TitanicSporeCarrier_ScratchTakesTheLastCardSoOnlyTheDrawLoses(t *testing.T) {
	// With one card left, Scratch (first) takes it, so the Draw finds an empty
	// deck and the controller loses. The Scratch itself is never the loss.
	s := newScene(t)
	p := s.g.Players[0]
	p.Deck = nil
	top := s.onTop(0, vanilla("Last")...)
	killOnOpponentsTurn(s, s.card("Titanic Spore Carrier"), false)
	if !scratched(p, top[0]) || inHand(p, top[0]) {
		t.Errorf("Scratch resolves first and takes the last card (FR-KWD-002m)")
	}
	if o := s.g.Outcome; !o.Over || o.Winner != 1 {
		t.Errorf("the draw after the Scratch finds an empty deck and loses (FR-WIN-001): outcome %+v", o)
	}
}

// Covers: FR-ZON-004
func TestCard_TitanicSporeCarrier_FaceDownCopyDoesNothingWhenKilled(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	top := s.onTop(0, vanilla("A", "B")...)
	killOnOpponentsTurn(s, s.card("Titanic Spore Carrier"), true)
	if scratched(p, top[0]) || inHand(p, top[0]) {
		t.Errorf("a face-down Scratch Company has no abilities, so its death neither scratches nor draws (FR-ZON-004)")
	}
}

// --- Portent Scrivener: "When this unit is killed on your turn, draw a card." ---

// scrivenerKilledOnOwnTurn puts the real Scrivener (face-down if asked) for
// player 0 and has it attack a 9/9 that destroys it on player 0's own turn.
func scrivenerKilledOnOwnTurn(s *scene, faceDown bool) {
	s.t.Helper()
	scr := s.put(0, ZoneMainTheater, s.card("Portent Scrivener"))
	scr.FaceDown = faceDown
	def := s.put(1, ZoneMainTheater, unit("Hitter", 9, 9, Keywords{}))
	s.do(s.attack(scr, def))
	if s.alive(scr) {
		s.t.Fatalf("precondition: the Scrivener is destroyed")
	}
}

// Covers: FR-ZON-004
func TestCard_PortentScrivener_KilledOnYourTurnDrawsOne(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	top := s.onTop(0, vanilla("A", "B")...)
	scrivenerKilledOnOwnTurn(s, false)
	if !inHand(p, top[0]) {
		t.Errorf("killed on its controller's turn, the Scrivener's controller draws the top card (FR-ZON-004)")
	}
	if inHand(p, top[1]) || len(p.Deck) == 0 || p.Deck[0] != top[1] {
		t.Errorf("exactly one card is drawn; the second stays on top of the deck")
	}
}

// Covers: FR-ZON-004
func TestCard_PortentScrivener_KilledOnOpponentsTurnDrawsNothing(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	top := s.onTop(0, vanilla("A", "B")...)
	killOnOpponentsTurn(s, s.card("Portent Scrivener"), false)
	if inHand(p, top[0]) || len(p.Deck) == 0 || p.Deck[0] != top[0] {
		t.Errorf("killed on the opponent's turn, the Scrivener's 'on your turn' condition fails and nothing is drawn")
	}
	if s.g.Outcome.Over {
		t.Errorf("no draw happens, so no one loses: %+v", s.g.Outcome)
	}
}

// Covers: FR-WIN-001
func TestCard_PortentScrivener_DrawFromEmptyDeckLosesAtThatSweep(t *testing.T) {
	s := newScene(t)
	s.g.Players[0].Deck = nil
	scrivenerKilledOnOwnTurn(s, false)
	if o := s.g.Outcome; !o.Over || o.Draw || o.Winner != 1 {
		t.Errorf("the Scrivener's required draw from an empty deck loses its controller at the trigger's sweep (FR-WIN-001): outcome %+v", o)
	}
}

// Covers: FR-ZON-004
func TestCard_PortentScrivener_FaceDownCopyDoesNotDraw(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	top := s.onTop(0, vanilla("A", "B")...)
	scrivenerKilledOnOwnTurn(s, true)
	if inHand(p, top[0]) || len(p.Deck) == 0 || p.Deck[0] != top[0] {
		t.Errorf("a face-down Scratch Company has no abilities, so its death draws nothing (FR-ZON-004)")
	}
}

// --- Despoiler Worm: "When this unit is killed, all opponents draw 2 cards." ---

// wormKilled puts the real Worm (face-down if asked) for player 0 and has a
// 20-power hitter destroy it, on player 0's turn if ownTurn, else on player 1's.
// Player 1's deck is replaced by exactly oppDeck once the hitter's turn begins,
// so the turn-start draw cannot disturb it.
func wormKilled(s *scene, ownTurn, faceDown bool, oppDeck ...*Card) []*Instance {
	s.t.Helper()
	worm := s.put(0, ZoneMainTheater, s.card("Despoiler Worm"))
	worm.FaceDown = faceDown
	if !ownTurn {
		s.endTurn()
	}
	s.g.Players[1].Deck = nil
	insts := s.onTop(1, oppDeck...)
	hit := s.put(1, ZoneMainTheater, unit("Hitter", 20, 9, Keywords{}))
	if ownTurn {
		s.do(s.attack(worm, hit))
	} else {
		s.do(s.attack(hit, worm))
	}
	if s.alive(worm) {
		s.t.Fatalf("precondition: the Worm is destroyed")
	}
	return insts
}

// Covers: FR-ZON-004
func TestCard_DespoilerWorm_KilledMakesOpponentDrawTwo(t *testing.T) {
	s := newScene(t)
	opp := s.g.Players[1]
	top := wormKilled(s, false, false, vanilla("A", "B", "C")...)
	if !inHand(opp, top[0]) || !inHand(opp, top[1]) {
		t.Errorf("when the Worm is killed the opponent draws the top two cards (FR-ZON-004)")
	}
	if inHand(opp, top[2]) || len(opp.Deck) == 0 || opp.Deck[0] != top[2] {
		t.Errorf("exactly two cards are drawn; the third stays on top of the opponent's deck")
	}
}

// Covers: FR-ZON-004
func TestCard_DespoilerWorm_ControllerDrawsNothing(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	mine := s.onTop(0, vanilla("Mine1", "Mine2")...)
	wormKilled(s, false, false, vanilla("A", "B", "C")...)
	if inHand(p, mine[0]) || len(p.Deck) == 0 || p.Deck[0] != mine[0] {
		t.Errorf("only opponents draw from the Worm's trigger; its controller draws nothing")
	}
}

// Covers: FR-ZON-004
func TestCard_DespoilerWorm_TriggersOnItsControllersTurnToo(t *testing.T) {
	s := newScene(t)
	opp := s.g.Players[1]
	top := wormKilled(s, true, false, vanilla("A", "B", "C")...)
	if !inHand(opp, top[0]) || !inHand(opp, top[1]) || inHand(opp, top[2]) {
		t.Errorf("the Worm's trigger has no turn condition, so killed on its controller's turn the opponent still draws exactly two (FR-ZON-004)")
	}
}

// Covers: FR-WIN-001
func TestCard_DespoilerWorm_OpponentWithOneCardInDeckLosesAtThatSweep(t *testing.T) {
	s := newScene(t)
	wormKilled(s, false, false, vanilla("Only")...)
	if o := s.g.Outcome; !o.Over || o.Draw || o.Winner != 0 {
		t.Errorf("the opponent cannot draw its second card and loses at the trigger's sweep (FR-WIN-001): outcome %+v", o)
	}
}

// Covers: FR-WIN-001
func TestCard_DespoilerWorm_OpponentWithEmptyDeckLosesAtThatSweep(t *testing.T) {
	s := newScene(t)
	wormKilled(s, false, false)
	if o := s.g.Outcome; !o.Over || o.Draw || o.Winner != 0 {
		t.Errorf("the opponent's required draw from an empty deck loses at the trigger's sweep (FR-WIN-001): outcome %+v", o)
	}
}

func TestCard_DespoilerWorm_ResistantTwoReducesDamage(t *testing.T) {
	s := newScene(t)
	hit := s.put(0, ZoneMainTheater, unit("Hitter", 5, 9, Keywords{}))
	worm := s.put(1, ZoneMainTheater, s.card("Despoiler Worm"))
	s.do(s.attack(hit, worm))
	s.wantDamage(worm, 3) // 5 -> 3 (Resistant 2)
}

// Covers: FR-ZON-004
func TestCard_DespoilerWorm_FaceDownCopyDoesNotMakeOpponentDraw(t *testing.T) {
	s := newScene(t)
	opp := s.g.Players[1]
	top := wormKilled(s, false, true, vanilla("A", "B", "C")...)
	if inHand(opp, top[0]) || len(opp.Deck) == 0 || opp.Deck[0] != top[0] {
		t.Errorf("a face-down Scratch Company has no abilities, so its death makes no one draw (FR-ZON-004)")
	}
}

// --- Stormfront of Ideals and Unceasing Tide: self-relocating death triggers ---

func inDiscardPile(p *Player, inst *Instance) bool {
	for _, d := range p.Discard {
		if d == inst {
			return true
		}
	}
	return false
}

// bottomOfDeck reports whether inst is the last card of p's deck.
func bottomOfDeck(p *Player, inst *Instance) bool {
	return len(p.Deck) > 0 && p.Deck[len(p.Deck)-1] == inst
}

// selfRelocKilled puts the real named card (face-down if asked) for player 0
// and has a 20-power hitter destroy it, on player 0's turn if ownTurn, else on
// player 1's. It returns the destroyed card's instance.
func selfRelocKilled(s *scene, name string, ownTurn, faceDown bool) *Instance {
	s.t.Helper()
	victim := s.put(0, ZoneMainTheater, s.card(name))
	victim.FaceDown = faceDown
	inst := victim.Inst
	if !ownTurn {
		s.endTurn()
	}
	hit := s.put(1, ZoneMainTheater, unit("Hitter", 20, 9, Keywords{}))
	if ownTurn {
		s.do(s.attack(victim, hit))
	} else {
		s.do(s.attack(hit, victim))
	}
	if s.alive(victim) {
		s.t.Fatalf("precondition: %s is destroyed", name)
	}
	return inst
}

// stolenKilled puts the real named card owned by player 0 but controlled by
// player 1, and has player 0's 20-power hitter destroy it on player 0's turn.
func stolenKilled(s *scene, name string) *Instance {
	s.t.Helper()
	c := s.card(name)
	a := &Asset{ID: s.g.newID(), Inst: &Instance{ID: s.g.newID(), Card: c, Owner: 0}, Controller: 1, Armor: c.Keywords.Armor}
	s.g.placeOnBattlefield(a, ZoneMainTheater)
	hit := s.put(0, ZoneMainTheater, unit("Hitter", 20, 9, Keywords{}))
	s.do(s.attack(hit, a))
	if s.alive(a) {
		s.t.Fatalf("precondition: %s is destroyed", name)
	}
	return a.Inst
}

// Covers: FR-ZON-004
func TestCard_StormfrontOfIdeals_KilledGoesToBottomOfOwnersDeck(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	before := len(p.Deck)
	inst := selfRelocKilled(s, "Stormfront of Ideals", false, false)
	if !bottomOfDeck(p, inst) {
		t.Errorf("when Stormfront is killed that exact card goes to the bottom of its owner's deck (FR-ZON-004)")
	}
	if inDiscardPile(p, inst) {
		t.Errorf("the card leaves the discard pile for the deck; it must not stay in discard (FR-ZON-004)")
	}
	if len(p.Deck) != before+1 {
		t.Errorf("exactly one card is added to the deck: had %d, now %d", before, len(p.Deck))
	}
}

// Covers: FR-ZON-004
func TestCard_StormfrontOfIdeals_TriggersOnItsControllersTurnToo(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	inst := selfRelocKilled(s, "Stormfront of Ideals", true, false)
	if !bottomOfDeck(p, inst) || inDiscardPile(p, inst) {
		t.Errorf("Stormfront's trigger has no turn condition, so killed on its controller's turn it still goes to the bottom of the deck (FR-ZON-004)")
	}
}

// Covers: FR-ZON-004
func TestCard_StormfrontOfIdeals_OwnerGetsItNotController(t *testing.T) {
	s := newScene(t)
	owner, ctrl := s.g.Players[0], s.g.Players[1]
	inst := stolenKilled(s, "Stormfront of Ideals")
	if !bottomOfDeck(owner, inst) {
		t.Errorf("Stormfront goes to the bottom of its owner's deck, not its controller's (FR-ZON-004)")
	}
	if bottomOfDeck(ctrl, inst) || inDiscardPile(ctrl, inst) || inDiscardPile(owner, inst) {
		t.Errorf("the card must not remain in either discard pile or enter the controller's deck (FR-ZON-004)")
	}
}

// Covers: FR-ZON-004
func TestCard_StormfrontOfIdeals_FaceDownCopyStaysInDiscard(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	inst := selfRelocKilled(s, "Stormfront of Ideals", false, true)
	if !inDiscardPile(p, inst) || bottomOfDeck(p, inst) {
		t.Errorf("a face-down Scratch Company has no abilities, so it stays in discard and is not moved to the deck (FR-ZON-004)")
	}
}

// Covers: FR-ZON-004
func TestCard_UnceasingTide_KilledReturnsToOwnersHand(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	inst := selfRelocKilled(s, "Unceasing Tide", false, false)
	if !inHand(p, inst) {
		t.Errorf("when Unceasing Tide is killed that exact card returns to its owner's hand (FR-ZON-004)")
	}
	if inDiscardPile(p, inst) {
		t.Errorf("the card leaves the discard pile for the hand (FR-ZON-004)")
	}
}

// Covers: FR-ZON-004
func TestCard_UnceasingTide_TriggersOnItsControllersTurnToo(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	inst := selfRelocKilled(s, "Unceasing Tide", true, false)
	if !inHand(p, inst) || inDiscardPile(p, inst) {
		t.Errorf("Tide's trigger has no turn condition, so killed on its controller's turn it still returns to hand (FR-ZON-004)")
	}
}

// Covers: FR-ZON-004
func TestCard_UnceasingTide_OwnerGetsItNotController(t *testing.T) {
	s := newScene(t)
	owner, ctrl := s.g.Players[0], s.g.Players[1]
	inst := stolenKilled(s, "Unceasing Tide")
	if !inHand(owner, inst) {
		t.Errorf("Tide returns to its owner's hand, not its controller's (FR-ZON-004)")
	}
	if inHand(ctrl, inst) || inDiscardPile(ctrl, inst) || inDiscardPile(owner, inst) {
		t.Errorf("the card must not remain in either discard pile or enter the controller's hand (FR-ZON-004)")
	}
}

// Covers: FR-ZON-004
func TestCard_UnceasingTide_FaceDownCopyStaysInDiscard(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	inst := selfRelocKilled(s, "Unceasing Tide", false, true)
	if !inDiscardPile(p, inst) || inHand(p, inst) {
		t.Errorf("a face-down Scratch Company has no abilities, so it stays in discard and is not returned to hand (FR-ZON-004)")
	}
}

// --- Spoil Forest Colossus ---
// "Whenever this unit deploys or attacks, chosen asset heals 2."  Unit,
// 6 Requisition, Reserves, 6/7. Uses deployPath, colossusCycle and wantOptions
// from cards_path_test.go.

func TestCard_SpoilForestColossus_EntersReservesAsSixSeven(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Timeout: true})
	c := deployPath(s, "Spoil Forest Colossus", 6)
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves {
		t.Fatalf("Spoil Forest Colossus is on the battlefield in Reserves after deploying")
	}
	if p := a.Printed(); p.Power != 6 || p.Health != 7 {
		t.Errorf("Spoil Forest Colossus is a 6/7: got %d/%d", p.Power, p.Health)
	}
	if s.g.Players[0].Req != 0 {
		t.Errorf("Spoil Forest Colossus costs 6 Requisition: %d left", s.g.Players[0].Req)
	}
}

// Covers: FR-TGT-001, FR-KWD-002h
func TestCard_SpoilForestColossus_DeployOffersEveryAssetOfBothSidesButOpposingWard(t *testing.T) {
	s := newScene(t)
	b := s.targetBoard()
	s.put(1, ZoneMainTheater, unit("Warded", 1, 5, Keywords{Ward: true}))
	s.scriptRefs(0, nil, b.u0)
	c := deployPath(s, "Spoil Forest Colossus", 6)
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one question: asked %d times", len(asked))
	}
	wantOptions(t, asked[0], "chosen asset: every asset of both sides (both Headquarters, Structures, Scratch Companies and the Colossus itself included), no opposing Ward unit",
		s.g.Players[0].HQ, s.g.Players[1].HQ, b.u0, b.s0, b.c0, b.u1, b.s1, b.c1, assetOf(s.g, c))
}

// Covers: FR-KWD-002d
func TestCard_SpoilForestColossus_DeployHealsChosenAssetOf2(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{}))
	ally.Damage = 5
	s.scriptRefs(0, nil, ally)
	deployPath(s, "Spoil Forest Colossus", 6)
	s.wantDamage(ally, 3)
}

// Covers: FR-KWD-002d
func TestCard_SpoilForestColossus_CanHealAnOpposingAsset(t *testing.T) {
	s := newScene(t)
	enemy := s.put(1, ZoneMainTheater, unit("Enemy", 1, 9, Keywords{}))
	enemy.Damage = 4
	s.scriptRefs(0, nil, enemy)
	deployPath(s, "Spoil Forest Colossus", 6)
	s.wantDamage(enemy, 2)
}

// Covers: FR-KWD-002d
func TestCard_SpoilForestColossus_CanHealAHeadquarters(t *testing.T) {
	s := newScene(t)
	hq0 := s.g.Players[0].HQ
	hq0.Damage = 6
	s.scriptRefs(0, nil, hq0)
	deployPath(s, "Spoil Forest Colossus", 6)
	s.wantDamage(hq0, 4)
}

// Covers: FR-KWD-002d
func TestCard_SpoilForestColossus_HealNeverGoesBelowZero(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{}))
	ally.Damage = 1
	s.scriptRefs(0, nil, ally)
	deployPath(s, "Spoil Forest Colossus", 6)
	s.wantDamage(ally, 0)
}

// Covers: FR-KWD-002d
func TestCard_SpoilForestColossus_HealingAnUndamagedAssetChangesNothing(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{}))
	s.scriptRefs(0, nil, ally)
	c := deployPath(s, "Spoil Forest Colossus", 6)
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("healing an undamaged asset leaves no damage anywhere: total %d", got)
	}
	if assetOf(s.g, c) == nil || !s.alive(ally) {
		t.Errorf("both assets are still on the battlefield")
	}
}

// Covers: FR-KWD-002d
func TestCard_SpoilForestColossus_HealLeavesArmorAlone(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Armored Ally", 1, 9, Keywords{Armor: 2}))
	ally.Damage = 3
	s.scriptRefs(0, nil, ally)
	deployPath(s, "Spoil Forest Colossus", 6)
	s.wantDamage(ally, 1)
	if ally.Armor != 2 {
		t.Errorf("Heal removes damage and does not touch Armor: got %d, want 2 (FR-KWD-002d)", ally.Armor)
	}
}

// Covers: FR-KWD-002h
func TestCard_SpoilForestColossus_OpposingWardUnitCannotBeChosenButAlliedCan(t *testing.T) {
	s := newScene(t)
	enemyWard := s.put(1, ZoneMainTheater, unit("Warded Enemy", 1, 9, Keywords{Ward: true}))
	enemyWard.Damage = 4
	allyWard := s.put(0, ZoneMainTheater, unit("Warded Ally", 1, 9, Keywords{Ward: true}))
	allyWard.Damage = 4
	s.scriptRefs(0, nil, allyWard)
	deployPath(s, "Spoil Forest Colossus", 6)
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one question: asked %d times", len(asked))
	}
	for _, o := range asked[0].Options {
		if o.Ref == enemyWard.ID {
			t.Errorf("an opposing Ward asset cannot be chosen by a triggered ability (FR-KWD-002h)")
		}
	}
	s.wantDamage(allyWard, 2)
	s.wantDamage(enemyWard, 4)
}

// Covers: FR-CBT-002, FR-KWD-002d
func TestCard_SpoilForestColossus_HealsAgainEachTimeItAttacks(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 20, Keywords{}))
	ally.Damage = 9
	dummy := s.put(1, ZoneMainTheater, unit("Dummy", 0, 30, Keywords{}))
	s.scriptRefs(0, nil, ally, ally, ally)
	c := deployPath(s, "Spoil Forest Colossus", 6)
	a := assetOf(s.g, c)
	s.wantDamage(ally, 7)
	colossusCycle(s, a)
	s.do(s.attack(a, dummy))
	s.wantDamage(ally, 5)
	s.endTurn()
	s.endTurn()
	s.do(s.attack(a, dummy))
	s.wantDamage(ally, 3)
	if n := len(s.asked(0)); n != 3 {
		t.Errorf("one question for the deploy and one per attack: asked %d times, want 3", n)
	}
}

// Covers: FR-CBT-002
func TestCard_SpoilForestColossus_AttackHealResolvesBeforeCombatDamage(t *testing.T) {
	// The Colossus has 6 damage on 7 health. It heals itself 2 at Step 2 (4
	// damage), then takes 2 retaliation and ends at 6. Without a Step 2 heal it
	// would take 8 and die.
	s := newScene(t)
	s.script(0, Answer{Timeout: true})
	c := deployPath(s, "Spoil Forest Colossus", 6)
	a := assetOf(s.g, c)
	colossusCycle(s, a)
	a.Damage = 6
	def := s.put(1, ZoneMainTheater, unit("Defender", 2, 30, Keywords{}))
	s.scriptRefs(0, nil, a)
	s.do(s.attack(a, def))
	if !s.alive(a) {
		t.Fatalf("the Step 2 heal resolves before combat damage, so the Colossus survives 2 retaliation (FR-CBT-002)")
	}
	s.wantDamage(a, 6)
}

// Covers: FR-CBT-002
func TestCard_SpoilForestColossus_DefendingDoesNotHeal(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Timeout: true})
	c := deployPath(s, "Spoil Forest Colossus", 6)
	a := assetOf(s.g, c)
	s.do(Action{Kind: ActMarch, Player: 0, Asset: a.ID})
	s.endTurn()
	a.Damage = 3
	hitter := s.put(1, ZoneMainTheater, unit("Hitter", 1, 30, Keywords{}))
	s.script(0)
	s.do(s.attack(hitter, a))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("only attacking triggers the Colossus: asked %d times while defending", n)
	}
	s.wantDamage(a, 4)
}

// --- Craterfall Ancient ---
// "Whenever this unit attacks, you may deal damage equal to its power to each
// of up to 2 other chosen assets."  "When this unit is killed, deal 4 damage to
// all assets."  Unit, 10 Requisition, Reserves, 6/13. Rulings SVY-R9 (you may),
// SVY-R10 (its power), SVY-R11 (up to N, distinct). Uses deployPath,
// colossusCycle and wantOptions from cards_path_test.go, and scriptPlan, pick,
// yes, no, wantDealt and kindsOf from optional_upto_test.go.

// ancientBoard is the battlefield around the Ancient. Every asset has 20 or
// more health, so nothing dies in a sweep or from the Ancient's 6 or 8 damage.
type ancientBoard struct {
	ally, enemy, enemy2, ward, dummy *Asset
}

// ancientScene puts the board on the battlefield, deploys the real Ancient for
// 10 Requisition (a question on deploy fails the test) and cycles it to the Main
// Theater, Readied, with player 0 to act. Player 0 then has an empty plan.
func ancientScene(t *testing.T) (*scene, *Asset, ancientBoard) {
	t.Helper()
	s := newScene(t)
	b := ancientBoard{
		ally:   s.put(0, ZoneMainTheater, unit("Ally", 0, 20, Keywords{})),
		enemy:  s.put(1, ZoneMainTheater, unit("Enemy", 0, 20, Keywords{})),
		enemy2: s.put(1, ZoneMainTheater, unit("Enemy Two", 0, 20, Keywords{})),
		ward:   s.put(1, ZoneMainTheater, unit("Warded", 0, 20, Keywords{Ward: true})),
		dummy:  s.put(1, ZoneMainTheater, unit("Dummy", 0, 30, Keywords{})),
	}
	s.scriptPlan(0, nil, nil)
	c := deployPath(s, "Craterfall Ancient", 10)
	anc := assetOf(s.g, c)
	if anc == nil {
		t.Fatalf("precondition: the Ancient is on the battlefield after deploying")
	}
	colossusCycle(s, anc)
	s.scriptPlan(0, nil, nil)
	return s, anc, b
}

// ancientKilled passes to player 1's turn, where a 13-power Hitter destroys the
// Readied Ancient, and returns the Hitter. The Hitter has 30 health, so it
// survives the 6 retaliation and the 4 of the death trigger.
func ancientKilled(s *scene, anc *Asset) *Asset {
	s.t.Helper()
	s.endTurn()
	hit := s.put(1, ZoneMainTheater, unit("Hitter", 13, 30, Keywords{}))
	s.do(s.attack(hit, anc))
	if s.alive(anc) {
		s.t.Fatalf("precondition: the Ancient is destroyed")
	}
	return hit
}

func TestCard_CraterfallAncient_EntersReservesAsSixThirteenForTenRequisition(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("Enemy", 0, 20, Keywords{}))
	s.scriptPlan(0, nil, nil)
	c := deployPath(s, "Craterfall Ancient", 10)
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves {
		t.Fatalf("Craterfall Ancient is on the battlefield in Reserves after deploying")
	}
	if p := a.Printed(); p.Power != 6 || p.Health != 13 {
		t.Errorf("Craterfall Ancient is a 6/13: got %d/%d", p.Power, p.Health)
	}
	if s.g.Players[0].Req != 0 {
		t.Errorf("Craterfall Ancient costs 10 Requisition: %d left", s.g.Players[0].Req)
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("deploying asks no question, the triggers are on attack and death: asked %d times", n)
	}
	if got := nonCombatDamage(s.g.Log); len(got) != 0 {
		t.Errorf("deploying deals no damage: %d damage events", len(got))
	}
}

// Covers: FR-CBT-002, FR-TGT-001, FR-KWD-002h
func TestCard_CraterfallAncient_AttackAsksMayThenOffersEveryOtherAssetButOpposingWard(t *testing.T) {
	s, anc, b := ancientScene(t)
	s.scriptPlan(0, nil, []Answer{yes}, pick(b.enemy))
	s.do(s.attack(anc, b.dummy))
	asked := s.asked(0)
	if want := []DecisionKind{DecideOptional, DecideTarget}; !reflect.DeepEqual(kindsOf(asked), want) {
		t.Fatalf("'you may' is asked first, then the targets: %v (SVY-R9)", kindsOf(asked))
	}
	if d := asked[1]; d.Min != 0 || d.Max != 2 {
		t.Errorf("up to 2 targets: Min 0, Max 2, got Min %d, Max %d (SVY-R11)", d.Min, d.Max)
	}
	wantOptions(t, asked[1], "other chosen assets: both Headquarters and every asset of both sides, not the Ancient, not the opposing Ward unit",
		s.g.Players[0].HQ, s.g.Players[1].HQ, b.ally, b.enemy, b.enemy2, b.dummy)
}

// Covers: SVY-R9, SVY-R11
func TestCard_CraterfallAncient_YesDealsPowerToEachOfTwoAssetsOwnHeadquartersIncluded(t *testing.T) {
	s, anc, b := ancientScene(t)
	hq0 := s.g.Players[0].HQ
	s.scriptPlan(0, nil, []Answer{yes}, pick(hq0, b.enemy2))
	s.do(s.attack(anc, b.dummy))
	wantDealt(t, s.g, hq0, 1, 6, "the Ancient's own Headquarters is a legal choice and takes its power, 6")
	wantDealt(t, s.g, b.enemy2, 1, 6, "the second chosen asset takes 6")
	wantDealt(t, s.g, b.enemy, 0, 0, "an unchosen asset takes none")
	wantDealt(t, s.g, b.ward, 0, 0, "an unchosen Ward unit takes none")
	wantDealt(t, s.g, anc, 0, 0, "the Ancient itself is not a legal choice")
	if got := len(nonCombatDamage(s.g.Log)); got != 2 {
		t.Errorf("two chosen assets are two non-combat damage instances: %d", got)
	}
	for _, e := range nonCombatDamage(s.g.Log) {
		if e.Source != anc.ID {
			t.Errorf("the damage comes from the Ancient: source %d, want %d", e.Source, anc.ID)
		}
	}
}

// Covers: SVY-R9, SVY-R11
func TestCard_CraterfallAncient_YesCanHitTheOpposingHeadquartersAndAnAlliedUnit(t *testing.T) {
	s, anc, b := ancientScene(t)
	hq1 := s.g.Players[1].HQ
	s.scriptPlan(0, nil, []Answer{yes}, pick(hq1, b.ally))
	s.do(s.attack(anc, b.dummy))
	wantDealt(t, s.g, hq1, 1, 6, "the opposing Headquarters takes 6")
	wantDealt(t, s.g, b.ally, 1, 6, "an allied unit takes 6: the targets are not limited to opposing assets")
}

// Covers: SVY-R11
func TestCard_CraterfallAncient_YesWithOneAssetDealsOneInstance(t *testing.T) {
	s, anc, b := ancientScene(t)
	s.scriptPlan(0, nil, []Answer{yes}, pick(b.enemy))
	s.do(s.attack(anc, b.dummy))
	wantDealt(t, s.g, b.enemy, 1, 6, "the one chosen asset takes 6")
	if got := len(nonCombatDamage(s.g.Log)); got != 1 {
		t.Errorf("up to 2 allows choosing 1, which is one instance: %d", got)
	}
}

// Covers: SVY-R11
func TestCard_CraterfallAncient_YesWithZeroAssetsDamagesNothing(t *testing.T) {
	s, anc, b := ancientScene(t)
	s.scriptPlan(0, nil, []Answer{yes}, pick())
	s.do(s.attack(anc, b.dummy))
	wantNoNonCombatDamage(t, s.g, "up to 2 allows choosing none")
	wantNoInvalid(t, s.g)
}

// Covers: SVY-R9
func TestCard_CraterfallAncient_NoDealsNothingAndAsksNoTarget(t *testing.T) {
	s, anc, b := ancientScene(t)
	s.scriptPlan(0, nil, []Answer{no})
	s.do(s.attack(anc, b.dummy))
	wantNoNonCombatDamage(t, s.g, "the 'you may' was declined")
	if want := []DecisionKind{DecideOptional}; !reflect.DeepEqual(kindsOf(s.asked(0)), want) {
		t.Errorf("a declined trigger asks no target: %v", kindsOf(s.asked(0)))
	}
}

// Covers: SVY-R9
func TestCard_CraterfallAncient_TimeoutOnTheQuestionDeclines(t *testing.T) {
	s, anc, b := ancientScene(t)
	s.scriptPlan(0, nil, []Answer{{Timeout: true}})
	s.do(s.attack(anc, b.dummy))
	wantNoNonCombatDamage(t, s.g, "a timeout on 'you may' declines")
	if want := []DecisionKind{DecideOptional}; !reflect.DeepEqual(kindsOf(s.asked(0)), want) {
		t.Errorf("a timed-out trigger asks no target: %v", kindsOf(s.asked(0)))
	}
}

// Covers: SVY-R10
func TestCard_CraterfallAncient_DamageEqualsEffectivePowerAtResolution(t *testing.T) {
	s, anc, b := ancientScene(t)
	anc.TempPower = 2 // stands in for an until-end-of-turn bonus, active this turn
	s.scriptPlan(0, nil, []Answer{yes}, pick(b.enemy, b.enemy2))
	s.do(s.attack(anc, b.dummy))
	wantDealt(t, s.g, b.enemy, 1, 8, "power 6 plus 2 temporary is 8, not the printed 6")
	wantDealt(t, s.g, b.enemy2, 1, 8, "each chosen asset takes the same 8")
}

// Covers: FR-CBT-002
func TestCard_CraterfallAncient_AttackDamageResolvesBeforeCombatSoAKilledDefenderFizzlesTheAttack(t *testing.T) {
	s, anc, _ := ancientScene(t)
	frail := s.put(1, ZoneMainTheater, unit("Frail Defender", 5, 6, Keywords{}))
	s.scriptPlan(0, nil, []Answer{yes}, pick(frail))
	s.do(s.attack(anc, frail))
	if s.alive(frail) {
		t.Fatalf("6 damage at Step 2 destroys the 6-health defender")
	}
	s.wantDamage(anc, 0) // a destroyed defender deals no retaliation (FR-CBT-002)
}

// Covers: FR-CBT-002
func TestCard_CraterfallAncient_AsksAgainOnEachAttack(t *testing.T) {
	s, anc, b := ancientScene(t)
	s.scriptPlan(0, nil, []Answer{yes, yes}, pick(b.enemy), pick(b.enemy))
	s.do(s.attack(anc, b.dummy))
	s.endTurn()
	s.endTurn()
	s.do(s.attack(anc, b.dummy))
	wantDealt(t, s.g, b.enemy, 2, 12, "each attack fires the trigger once")
	if n := len(s.asked(0)); n != 4 {
		t.Errorf("one 'you may' and one target question per attack: asked %d times, want 4", n)
	}
}

// Covers: FR-CBT-002
func TestCard_CraterfallAncient_DefendingDoesNotTrigger(t *testing.T) {
	s, anc, _ := ancientScene(t)
	s.endTurn()
	hit := s.put(1, ZoneMainTheater, unit("Hitter", 1, 30, Keywords{}))
	s.do(s.attack(hit, anc))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("only attacking triggers the Ancient: asked %d times while defending", n)
	}
	wantNoNonCombatDamage(t, s.g, "the attack trigger does not fire when defending")
}

// Covers: FR-ZON-004
func TestCard_CraterfallAncient_FaceDownCopyAttackingAsksNothing(t *testing.T) {
	s := newScene(t)
	fd := s.put(0, ZoneMainTheater, s.card("Craterfall Ancient"))
	fd.FaceDown = true
	dummy := s.put(1, ZoneMainTheater, unit("Dummy", 0, 30, Keywords{}))
	s.scriptPlan(0, nil, nil)
	s.do(s.attack(fd, dummy))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a face-down Scratch Company has no printed trigger: asked %d times (FR-ZON-004)", n)
	}
	wantNoNonCombatDamage(t, s.g, "a face-down copy fires no attack trigger")
}

func TestCard_CraterfallAncient_DeathDealsFourToEveryAssetOfBothSidesWithoutAQuestion(t *testing.T) {
	s, anc, b := ancientScene(t)
	hit := ancientKilled(s, anc)
	hq0, hq1 := s.g.Players[0].HQ, s.g.Players[1].HQ
	for _, a := range []*Asset{hq0, hq1, b.ally, b.enemy, b.enemy2, b.dummy, hit} {
		wantDealt(t, s.g, a, 1, 4, "every asset on the battlefield takes one 4-damage instance, allied, opposing and Headquarters")
	}
	wantDealt(t, s.g, b.ward, 1, 4, "'all assets' is not a target, so an opposing Ward unit still takes 4 (FR-TGT-002)")
	if got := len(nonCombatDamage(s.g.Log)); got != 8 {
		t.Errorf("eight assets are left, so eight instances: %d", got)
	}
	for _, e := range nonCombatDamage(s.g.Log) {
		if e.Source != anc.ID {
			t.Errorf("the damage comes from the dead Ancient: source %d, want %d", e.Source, anc.ID)
		}
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("the death trigger asks no question: asked %d times", n)
	}
}

func TestCard_CraterfallAncient_DeathKillsAssetsWithFourOrLessHealthAndSparesTheRest(t *testing.T) {
	s, anc, _ := ancientScene(t)
	frailAlly := s.put(0, ZoneMainTheater, unit("Frail Ally", 0, 4, Keywords{}))
	frailEnemy := s.put(1, ZoneMainTheater, unit("Frail Enemy", 0, 4, Keywords{}))
	sturdy := s.put(1, ZoneMainTheater, unit("Sturdy Enemy", 0, 5, Keywords{}))
	ancientKilled(s, anc)
	if s.alive(frailAlly) || s.alive(frailEnemy) {
		t.Errorf("4 damage destroys 4-health assets of both sides in the same sweep")
	}
	if !s.alive(sturdy) {
		t.Fatalf("a 5-health asset survives 4 damage")
	}
	s.wantDamage(sturdy, 4)
}

func TestCard_CraterfallAncient_DeathAppliesResistantAndArmorPerInstance(t *testing.T) {
	s, anc, _ := ancientScene(t)
	resist := s.put(1, ZoneMainTheater, unit("Resistant Unit", 0, 20, Keywords{Resistant: 2}))
	armored := s.put(0, ZoneMainTheater, unit("Armored Unit", 0, 20, Keywords{Armor: 3}))
	ancientKilled(s, anc)
	s.wantDamage(resist, 2) // 4 - Resistant 2 (FR-REP-004)
	s.wantDamage(armored, 1)
	if armored.Armor != 0 {
		t.Errorf("3 Armor absorbs 3 of the 4 damage 1-for-1 and is gone: Armor %d (FR-REP-004)", armored.Armor)
	}
}

// Covers: FR-CBT-003
func TestCard_CraterfallAncient_DeathWithBothHeadquartersAtFourOrLessIsAMatchDraw(t *testing.T) {
	s, anc, _ := ancientScene(t)
	s.g.Players[0].HQ.Damage = 16
	s.g.Players[1].HQ.Damage = 16
	ancientKilled(s, anc)
	if o := s.g.Outcome; !o.Over || !o.Draw {
		t.Errorf("both Headquarters die in the one sweep: outcome %+v, want Match Draw (FR-CBT-003)", o)
	}
}

// Covers: FR-CBT-003
func TestCard_CraterfallAncient_DeathWithOneHeadquartersAtFourOrLessLosesThatPlayer(t *testing.T) {
	for _, loser := range []PlayerID{0, 1} {
		s, anc, _ := ancientScene(t)
		s.g.Players[loser].HQ.Damage = 16
		ancientKilled(s, anc)
		if o := s.g.Outcome; !o.Over || o.Draw || o.Winner != 1-loser {
			t.Errorf("player %d's Headquarters dies alone: outcome %+v, want player %d wins (FR-CBT-003)", loser, o, 1-loser)
		}
	}
}

// Covers: FR-ZON-004
func TestCard_CraterfallAncient_FaceDownCopyFiresNothingWhenKilled(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 0, 20, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Enemy", 0, 20, Keywords{}))
	s.scriptPlan(0, nil, nil)
	killOnOpponentsTurn(s, s.card("Craterfall Ancient"), true)
	wantNoNonCombatDamage(t, s.g, "a face-down Scratch Company has no printed death trigger (FR-ZON-004)")
	s.wantDamage(ally, 0)
	s.wantDamage(s.g.Players[0].HQ, 0)
}

// Covers: FR-CBT-002
func TestCard_CraterfallAncient_AttackThenLethalRetaliationFiresBothTriggers(t *testing.T) {
	s, anc, b := ancientScene(t)
	brute := s.put(1, ZoneMainTheater, unit("Brute", 13, 30, Keywords{}))
	s.scriptPlan(0, nil, []Answer{yes}, pick(b.enemy, b.ally))
	s.do(s.attack(anc, brute))
	if s.alive(anc) {
		t.Fatalf("precondition: 13 retaliation destroys the 13-health Ancient")
	}
	wantDealt(t, s.g, b.enemy, 2, 10, "6 from the attack trigger, then 4 from the death trigger")
	wantDealt(t, s.g, b.ally, 2, 10, "6 from the attack trigger, then 4 from the death trigger")
	wantDealt(t, s.g, b.dummy, 1, 4, "an unchosen asset takes only the death trigger's 4")
	wantDealt(t, s.g, brute, 1, 4, "the defender takes the death trigger's 4 on top of the combat damage")
	wantDealt(t, s.g, s.g.Players[0].HQ, 1, 4, "the death trigger hits the Headquarters")
	if want := []DecisionKind{DecideOptional, DecideTarget}; !reflect.DeepEqual(kindsOf(s.asked(0)), want) {
		t.Errorf("one 'you may' and one target question, none for the death: %v", kindsOf(s.asked(0)))
	}
}

// --- Animoceps Shambler ---
// "Whenever this unit exhausts, chosen allied asset in the Main Theater heals
// 1."  Unit, 2 Requisition, Reserves, 2/2, yields 1R 2T (T-0064).
//
// FR-CAP-001 and SVY-R17: every change from Readied to Exhausted fires it
// (March, Attack, Generate). A March's or Attack's trigger is placed on top of
// the action and resolves before it; a Generate's trigger is placed after the
// generation completes and resolves before the player acts again. SVY-R3: the
// target is chosen when the trigger is placed, and FR-TGT-003 checks it again
// when the trigger resolves. SVY-R16: a timeout picks the first legal option
// in fixed order. Heal is FR-KWD-002d.

const shambler = "Animoceps Shambler"

// putShambler puts a Readied real Animoceps Shambler for p into zone z with
// the given damage. It sets state and asserts nothing.
func putShambler(s *scene, p PlayerID, z Zone, damage int) *Asset {
	a := s.put(p, z, s.card(shambler))
	a.Damage = damage
	return a
}

// putDamaged puts a Readied vanilla unit for p into zone z with the given
// damage. It sets state and asserts nothing.
func putDamaged(s *scene, p PlayerID, z Zone, name string, health, damage int) *Asset {
	a := s.put(p, z, unit(name, 0, health, Keywords{}))
	a.Damage = damage
	return a
}

// firstTargetDecision is the log position of the first recorded target
// decision (answered, fast-pathed or timed out), or -1.
func firstTargetDecision(g *Game) int {
	for i, e := range g.Log {
		if e.Kind == "decision" && e.Decision != nil && e.Decision.Kind == "Target" {
			return i
		}
	}
	return -1
}

func TestCard_AnimocepsShambler_DeploysToReservesReadiedAsTwoTwoAndAsksNothing(t *testing.T) {
	s := newScene(t)
	s.script(0)
	inst := s.give(0, s.card(shambler))
	s.g.Players[0].Req = 2
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	a := assetOf(s.g, inst.Card)
	if a == nil || a.Zone != ZoneReserves {
		t.Fatalf("the Shambler is on the battlefield in Reserves after deploying, got %v", a)
	}
	if p := a.Printed(); p.Power != 2 || p.Health != 2 || p.ReqYield != 1 || p.TacYield != 2 {
		t.Errorf("the Shambler is a 2/2 yielding 1R 2T: got %d/%d, %dR %dT", p.Power, p.Health, p.ReqYield, p.TacYield)
	}
	if s.g.Players[0].Req != 0 {
		t.Errorf("the Shambler costs 2 Requisition: %d left", s.g.Players[0].Req)
	}
	if a.Exhausted {
		t.Errorf("the Shambler enters Readied (FR-CAP-001)")
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("deploying is not exhausting, so the trigger does not fire: asked %d times (SVY-R17)", n)
	}
}

// Covers: FR-CAP-001, FR-QUE-001, FR-TGT-003
func TestCard_AnimocepsShambler_MarchTriggerResolvesBeforeTheMarch(t *testing.T) {
	// The Shambler marches from the Main Theater back to Reserves and is the
	// only allied asset in the Main Theater, so it is chosen. Its trigger sits
	// on top of the March and resolves first, while the Shambler is still in
	// the Main Theater, so it heals. Had the March resolved first, the
	// Shambler would be in Reserves and the heal would fizzle (FR-TGT-003).
	s := newScene(t)
	sh := putShambler(s, 0, ZoneMainTheater, 1)
	s.do(Action{Kind: ActMarch, Player: 0, Asset: sh.ID})
	if sh.Zone != ZoneReserves {
		t.Fatalf("precondition: the March still resolves after the trigger: Shambler in %s, want Reserves", sh.Zone)
	}
	if sh.Damage != 0 {
		t.Errorf("a March's exhaust trigger is placed on top of the March and resolves first, while its target is still in the Main Theater, so the Shambler heals 1 (SVY-R17): damage %d, want 0", sh.Damage)
	}
	s.wantSettled("a March with the Shambler's trigger")
}

// Covers: FR-CAP-001, FR-QUE-001
func TestCard_AnimocepsShambler_GenerateTriggerResolvesBeforeTheNextAction(t *testing.T) {
	s := newScene(t)
	sh := putShambler(s, 0, ZoneMainTheater, 0)
	ally := putDamaged(s, 0, ZoneMainTheater, "Ally", 5, 1)
	tac := s.g.Players[0].Tac
	s.scriptRefs(0, nil, ally)
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: sh.ID, Resource: Tactics})
	if got, want := s.g.Players[0].Tac, tac+2; got != want {
		t.Errorf("the Shambler in the Main Theater generates its Tactics Yield of 2: Tactics %d, want %d (FR-CAP-001)", got, want)
	}
	gen, dec := logIndex(s.g, "generate"), firstTargetDecision(s.g)
	if gen < 0 || dec < 0 {
		t.Fatalf("expected a generate event and a target decision in the log, got generate at %d, decision at %d", gen, dec)
	}
	if dec < gen {
		t.Errorf("a Generate's exhaust trigger is placed, and its target chosen, after the generation completes (SVY-R17, SVY-R3): decision event %d came before generate event %d", dec, gen)
	}
	if ally.Damage != 0 {
		t.Errorf("the trigger has resolved by the time the Generate returns, before the player acts again (SVY-R17): ally damage %d, want 0", ally.Damage)
	}
	s.wantSettled("a Generate with the Shambler's trigger")
	s.do(Action{Kind: ActMarch, Player: 0, Asset: ally.ID})
	s.wantDamage(ally, 0)
}

// Covers: FR-TGT-001
func TestCard_AnimocepsShambler_MarchingFromReservesCannotChooseItself(t *testing.T) {
	// SVY-R3: the target is chosen when the trigger is placed. The Shambler is
	// still in Reserves then, so it is not an asset in the Main Theater.
	s := newScene(t)
	sh := putShambler(s, 0, ZoneReserves, 1)
	a1 := putDamaged(s, 0, ZoneMainTheater, "Ally One", 5, 1)
	a2 := putDamaged(s, 0, ZoneMainTheater, "Ally Two", 5, 1)
	s.scriptRefs(0, nil, a2)
	s.do(Action{Kind: ActMarch, Player: 0, Asset: sh.ID})
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one target question: asked %d times", len(asked))
	}
	wantOptions(t, asked[0], "chosen allied asset in the Main Theater, placed while the Shambler is in Reserves: the two allies, not the Shambler (SVY-R3)", a1, a2)
	s.wantDamage(a2, 0)
	s.wantDamage(a1, 1)
	s.wantDamage(sh, 1)
	if sh.Zone != ZoneMainTheater {
		t.Errorf("the March resolves after the trigger: Shambler in %s, want the Main Theater", sh.Zone)
	}
}

// Covers: FR-TGT-001, FR-CBT-002
func TestCard_AnimocepsShambler_AttackingFromTheMainTheaterMayChooseItself(t *testing.T) {
	s := newScene(t)
	sh := putShambler(s, 0, ZoneMainTheater, 1)
	ally := putDamaged(s, 0, ZoneMainTheater, "Ally", 5, 1)
	def := s.put(1, ZoneMainTheater, unit("Defender", 0, 9, Keywords{}))
	s.scriptRefs(0, nil, sh)
	s.do(s.attack(sh, def))
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one target question: asked %d times", len(asked))
	}
	wantOptions(t, asked[0], "an attacking Shambler is an allied asset in the Main Theater, so it is offered with the ally (FR-TGT-001)", sh, ally)
	s.wantDamage(sh, 0)
	s.wantDamage(ally, 1)
	s.wantDamage(def, 2)
}

// Covers: FR-TGT-001, FR-KWD-002h
func TestCard_AnimocepsShambler_OffersOnlyAlliedAssetsInTheMainTheater(t *testing.T) {
	s := newScene(t)
	b := s.targetBoard()
	sh := putShambler(s, 0, ZoneMainTheater, 0)
	s.put(0, ZoneReserves, unit("Ally In Reserves", 1, 5, Keywords{}))
	scratchMT := s.putScratch(0)
	s.moveTo(scratchMT, ZoneMainTheater)
	allyWard := s.put(0, ZoneMainTheater, unit("Warded Ally", 1, 5, Keywords{Ward: true}))
	s.scriptRefs(0, nil, b.u0)
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: sh.ID, Resource: Tactics})
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one target question: asked %d times", len(asked))
	}
	wantOptions(t, asked[0], "chosen allied asset in the Main Theater: allied units, structures and Scratch Companies there (the Shambler and an allied Ward unit included); no Reserves asset, no Headquarters, no opposing asset",
		sh, b.u0, b.s0, scratchMT, allyWard)
}

// Covers: FR-TGT-003
func TestCard_AnimocepsShambler_NoLegalTargetAsksNothingAndHealsNothing(t *testing.T) {
	// Marching from Reserves with no other allied asset in the Main Theater:
	// only damaged assets that are not legal targets exist.
	s := newScene(t)
	sh := putShambler(s, 0, ZoneReserves, 1)
	res := putDamaged(s, 0, ZoneReserves, "Ally In Reserves", 5, 1)
	enemy := putDamaged(s, 1, ZoneMainTheater, "Enemy", 5, 1)
	s.g.Players[0].HQ.Damage = 1
	s.script(0)
	s.do(Action{Kind: ActMarch, Player: 0, Asset: sh.ID})
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("with no legal target nothing is asked: asked %d times (FR-TGT-003)", n)
	}
	s.wantDamage(sh, 1)
	s.wantDamage(res, 1)
	s.wantDamage(enemy, 1)
	s.wantDamage(s.g.Players[0].HQ, 1)
	if sh.Zone != ZoneMainTheater {
		t.Errorf("the March still resolves: Shambler in %s, want the Main Theater", sh.Zone)
	}
	s.wantSettled("a March whose trigger has no legal target")
}

// Covers: FR-KWD-002d
func TestCard_AnimocepsShambler_HealsExactlyOne(t *testing.T) {
	s := newScene(t)
	sh := putShambler(s, 0, ZoneMainTheater, 0)
	ally := putDamaged(s, 0, ZoneMainTheater, "Ally", 9, 3)
	s.scriptRefs(0, nil, ally)
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: sh.ID, Resource: Tactics})
	s.wantDamage(ally, 2)
}

// Covers: FR-KWD-002d
func TestCard_AnimocepsShambler_HealingAnUndamagedAssetChangesNothing(t *testing.T) {
	s := newScene(t)
	sh := putShambler(s, 0, ZoneMainTheater, 0)
	ally := putDamaged(s, 0, ZoneMainTheater, "Ally", 5, 0)
	s.scriptRefs(0, nil, ally)
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: sh.ID, Resource: Tactics})
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("healing an undamaged asset leaves no damage anywhere: total %d (FR-KWD-002d)", got)
	}
	if !s.alive(ally) || !s.alive(sh) {
		t.Errorf("both assets are still on the battlefield")
	}
}

// Covers: FR-CBT-002, FR-KWD-002d
func TestCard_AnimocepsShambler_AttackHealResolvesBeforeCombatDamage(t *testing.T) {
	// The 2/2 Shambler has 1 damage and is the only allied asset in the Main
	// Theater, so it heals itself to full at Step 2, then takes 1 retaliation
	// and survives with 1 damage. Healed after combat, it would have died.
	s := newScene(t)
	sh := putShambler(s, 0, ZoneMainTheater, 1)
	def := s.put(1, ZoneMainTheater, unit("Defender", 1, 9, Keywords{}))
	s.do(s.attack(sh, def))
	if !s.alive(sh) {
		t.Fatalf("the Step 2 heal resolves before combat damage, so the Shambler survives 1 retaliation (SVY-R17, FR-CBT-002)")
	}
	s.wantDamage(sh, 1)
	s.wantDamage(def, 2)
}

// Covers: FR-MOD-001
func TestCard_AnimocepsShambler_TimeoutHealsTheFirstAlliedAssetToEnterTheMainTheater(t *testing.T) {
	// SVY-R16: own assets first; Headquarters, then Reserves, then the Main
	// Theater, each zone in entry order. Only Main Theater assets are legal,
	// so the first one to enter it is healed.
	s := newScene(t)
	s.g.Players[0].HQ.Damage = 1
	res := putDamaged(s, 0, ZoneReserves, "Ally In Reserves", 5, 1)
	first := putDamaged(s, 0, ZoneMainTheater, "First", 5, 1)
	sh := putShambler(s, 0, ZoneMainTheater, 1)
	last := putDamaged(s, 0, ZoneMainTheater, "Last", 5, 1)
	s.script(0, Answer{Timeout: true})
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: sh.ID, Resource: Tactics})
	s.wantDamage(first, 0)
	s.wantDamage(sh, 1)
	s.wantDamage(last, 1)
	s.wantDamage(res, 1)
	s.wantDamage(s.g.Players[0].HQ, 1)
}

// Covers: FR-CAP-001
func TestCard_AnimocepsShambler_HealsAgainOnEachExhaust(t *testing.T) {
	// Generate Requisition in Reserves on one turn, March on the next: two
	// exhausts, two heals. The ally is the only legal target both times.
	s := newScene(t)
	sh := putShambler(s, 0, ZoneReserves, 0)
	ally := putDamaged(s, 0, ZoneMainTheater, "Ally", 9, 3)
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: sh.ID, Resource: Requisition})
	s.wantDamage(ally, 2)
	s.endTurn()
	s.endTurn()
	s.do(Action{Kind: ActMarch, Player: 0, Asset: sh.ID})
	s.wantDamage(ally, 1)
}

// Covers: FR-CBT-002
func TestCard_AnimocepsShambler_DefendingDoesNotHeal(t *testing.T) {
	// Being attacked does not exhaust the Shambler, so nothing fires.
	s := newScene(t)
	sh := putShambler(s, 0, ZoneMainTheater, 0)
	ally := putDamaged(s, 0, ZoneMainTheater, "Ally", 5, 1)
	s.endTurn()
	hitter := s.put(1, ZoneMainTheater, unit("Hitter", 1, 9, Keywords{}))
	s.script(0)
	s.do(s.attack(hitter, sh))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("only the Shambler exhausting fires its trigger: asked %d times while defending", n)
	}
	s.wantDamage(ally, 1)
	s.wantDamage(sh, 1)
}

// Covers: FR-ZON-005
func TestCard_AnimocepsShambler_FaceDownCopyExhaustingHealsNothing(t *testing.T) {
	s := newScene(t)
	sh := putShambler(s, 0, ZoneMainTheater, 0)
	sh.FaceDown = true
	ally := putDamaged(s, 0, ZoneMainTheater, "Ally", 5, 1)
	s.script(0)
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: sh.ID, Resource: Tactics})
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a face-down Scratch Company has no printed abilities: asked %d times (FR-ZON-005)", n)
	}
	s.wantDamage(ally, 1)
}
