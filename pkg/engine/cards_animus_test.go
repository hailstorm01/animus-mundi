package engine

import "testing"

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
