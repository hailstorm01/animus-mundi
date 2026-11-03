package engine

import (
	"reflect"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0123 (FR-MNV-001, FR-TGT-001, FR-TGT-003, FR-KWD-002h,
// FR-REP-001): the Path Maneuver Devilhorn Sanction, loaded from data/cards.
// Player 0 enacts; player 1 is the opponent.
//
// Devilhorn Sanction: "Kill chosen opposing unit." Cost 3 Tactics.

// chooseAsset installs a chooser for seat 0 that answers every decision by
// picking the option that refers to target, and records the decisions asked.
func chooseAsset(s *scene, target *Asset) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		for i, o := range d.Options {
			if o.Ref == target.ID {
				return Answer{Picks: []int{i}}
			}
		}
		s.t.Errorf("target %s is not offered in %q: %v", target, d.Prompt, d.Options)
		return Answer{Timeout: true}
	}
}

// targetOffers returns the Ref set of the first target decision seat 0 was asked.
func targetOffers(s *scene) map[int]bool {
	for _, d := range s.asked(0) {
		if d.Kind == DecideTarget {
			m := map[int]bool{}
			for _, o := range d.Options {
				m[o.Ref] = true
			}
			return m
		}
	}
	s.t.Fatalf("no target decision was asked of seat 0")
	return nil
}

func TestCard_DevilhornSanction_KillsChosenOpposingUnit(t *testing.T) {
	// Covers: FR-MNV-001, FR-TGT-001
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 3, Keywords{}))
	other := s.put(1, ZoneMainTheater, unit("Other", 1, 3, Keywords{}))
	inst := enactFor(s, "Devilhorn Sanction")
	chooseAsset(s, foe)
	s.do(enactAction(inst))
	if s.alive(foe) || !inDiscardPile(s.g.Players[1], foe.Inst) {
		t.Errorf("the chosen opposing unit is killed and goes to its owner's discard pile (FR-MNV-001)")
	}
	if !s.alive(other) {
		t.Errorf("the unchosen opposing unit survives (FR-TGT-001)")
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver ends in its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_DevilhornSanction_KillsInReservesAndFaceDown(t *testing.T) {
	// Covers: FR-TGT-001
	s := newScene(t)
	res := s.put(1, ZoneReserves, unit("Reserve", 1, 3, Keywords{}))
	chooseAsset(s, res)
	s.do(enactAction(enactFor(s, "Devilhorn Sanction")))
	if s.alive(res) || !inDiscardPile(s.g.Players[1], res.Inst) {
		t.Errorf("an opposing unit in Reserves can be chosen and is killed (FR-TGT-001)")
	}

	s = newScene(t)
	sc := s.putScratch(1)
	chooseAsset(s, sc)
	s.do(enactAction(enactFor(s, "Devilhorn Sanction")))
	if s.alive(sc) || !inDiscardPile(s.g.Players[1], sc.Inst) {
		t.Errorf("an opposing face-down Scratch Company is a unit: it is killed and goes to its owner's discard pile (FR-TGT-001, FR-ZON-005)")
	}
}

func TestCard_DevilhornSanction_IgnoresHealthAndArmor(t *testing.T) {
	// Covers: FR-MNV-001
	s := newScene(t)
	tank := s.put(1, ZoneMainTheater, unit("Tank", 1, 30, Keywords{Armor: 9}))
	chooseAsset(s, tank)
	s.do(enactAction(enactFor(s, "Devilhorn Sanction")))
	if s.alive(tank) {
		t.Errorf("a kill destroys the unit at once whatever its Health and Armor (FR-MNV-001)")
	}
	if n := damageEventsOn(s, tank); n != 0 {
		t.Errorf("a kill is not damage: %d damage events on the target, want 0", n)
	}
}

func TestCard_DevilhornSanction_OnlyOpposingUnitsAreOffered(t *testing.T) {
	// Covers: FR-TGT-001
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 3, Keywords{}))
	st := s.put(1, ZoneMainTheater, structure("EWall", 0, 3, Keywords{}))
	u1 := s.put(1, ZoneMainTheater, unit("E1", 1, 3, Keywords{}))
	u2 := s.put(1, ZoneReserves, unit("E2", 1, 3, Keywords{}))
	chooseAsset(s, u1)
	s.do(enactAction(enactFor(s, "Devilhorn Sanction")))
	got := targetOffers(s)
	if len(got) != 2 || !got[u1.ID] || !got[u2.ID] {
		t.Errorf("exactly the two opposing units are offered, got refs %v (FR-TGT-001)", got)
	}
	for _, a := range []*Asset{ally, st, s.g.Players[0].HQ, s.g.Players[1].HQ} {
		if got[a.ID] {
			t.Errorf("%s is not an opposing unit and must not be offered (FR-TGT-001)", a)
		}
	}
}

func TestCard_DevilhornSanction_WardUnitCannotBeChosen(t *testing.T) {
	// Covers: FR-KWD-002h
	s := newScene(t)
	ward := s.put(1, ZoneMainTheater, unit("Warded", 1, 3, Keywords{Ward: true}))
	plain := s.put(1, ZoneMainTheater, unit("Plain", 1, 3, Keywords{}))
	plain2 := s.put(1, ZoneMainTheater, unit("Plain2", 1, 3, Keywords{}))
	chooseAsset(s, plain)
	s.do(enactAction(enactFor(s, "Devilhorn Sanction")))
	got := targetOffers(s)
	if got[ward.ID] || len(got) != 2 || !got[plain.ID] || !got[plain2.ID] {
		t.Errorf("an opposing Ward unit is not offered, only the two plain ones are, got refs %v (FR-KWD-002h)", got)
	}
	if !s.alive(ward) {
		t.Errorf("the Ward unit survives (FR-KWD-002h)")
	}
}

func TestCard_DevilhornSanction_NotEnactableWithoutLegalTarget(t *testing.T) {
	// Covers: SVY-R29, FR-MNV-001
	cases := []struct {
		name  string
		setup func(s *scene)
		want  bool
	}{
		{"no opposing unit", func(s *scene) {}, false},
		{"only an opposing Ward unit", func(s *scene) {
			s.put(1, ZoneMainTheater, unit("Warded", 1, 3, Keywords{Ward: true}))
		}, false},
		{"only allied units and opposing structures", func(s *scene) {
			s.put(0, ZoneMainTheater, unit("Ally", 1, 3, Keywords{}))
			s.put(1, ZoneMainTheater, structure("EWall", 0, 3, Keywords{}))
		}, false},
		{"one plain opposing unit", func(s *scene) {
			s.put(1, ZoneReserves, unit("Foe", 1, 3, Keywords{}))
		}, true},
	}
	for _, c := range cases {
		s := newScene(t)
		c.setup(s)
		inst := enactFor(s, "Devilhorn Sanction")
		if got := offersEnact(s.g, inst); got != c.want {
			t.Errorf("%s: Enact offered = %v, want %v (SVY-R29)", c.name, got, c.want)
		}
		if !c.want {
			s.reject(enactAction(inst))
			if !inHand(s.g.Players[0], inst) {
				t.Errorf("%s: the card stays in hand (SVY-R29)", c.name)
			}
		}
	}
}

func TestCard_DevilhornSanction_ShiftingImpossibilityReplacesTheKill(t *testing.T) {
	// Covers: FR-REP-001, SVY-R21
	s := newScene(t)
	si := s.put(1, ZoneReserves, s.card("Shifting Impossibility"))
	sc := s.putScratch(1)
	inst := enactFor(s, "Devilhorn Sanction")
	chooseAsset(s, sc)
	s.do(enactAction(inst))
	if s.alive(sc) || !inHand(s.g.Players[1], sc.Inst) || inDiscardPile(s.g.Players[1], sc.Inst) {
		t.Errorf("the kill of the Scratch Company is replaced: it returns to its owner's hand, not the discard pile (FR-REP-001)")
	}
	s.wantDamage(si, 1)
	wantReplaced(t, s, sc)
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver still ends in its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_DevilhornSanction_PaysThreeTactics(t *testing.T) {
	// Covers: FR-MNV-001
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 3, Keywords{}))
	chooseAsset(s, foe)
	s.do(enactAction(enactFor(s, "Devilhorn Sanction")))
	if p := s.g.Players[0]; p.Tac != 2 || p.Req != 3 {
		t.Errorf("enacting costs 3 Tactics and no Requisition: pool R%d T%d, want R3 T2 (FR-MNV-001)", p.Req, p.Tac)
	}
}

// Tests for T-0125 (FR-MNV-001, FR-MOD-001, FR-WIN-001): the Path Maneuver
// Improvised Orders, loaded from data/cards. Player 0 enacts; player 1 is the
// opponent.
//
// Improvised Orders: "Draw 3 cards, then discard a card." Cost 2 Tactics.

// improvisedScene gives player 0 the named vanilla hand cards (the last is
// most recently added), puts the named vanilla cards on top of player 0's
// deck (the first on top) and gives player 0 the real Improvised Orders. Seat 0
// starts with an empty hand and the given deck only; the deck below the named
// cards is emptied when emptyDeck is set. It returns the scene, the Maneuver
// and the hand and deck instances.
func improvisedScene(t *testing.T, hand, deck []string, emptyDeck bool) (*scene, *Instance, []*Instance, []*Instance) {
	t.Helper()
	s := newScene(t)
	s.g.Players[0].Hand = nil
	if emptyDeck {
		s.g.Players[0].Deck = nil
	}
	hs := make([]*Instance, len(hand))
	for i, c := range vanilla(hand...) {
		hs[i] = s.give(0, c)
	}
	ds := s.onTop(0, vanilla(deck...)...)
	return s, enactFor(s, "Improvised Orders"), hs, ds
}

// handDiscards returns the instances of the cards discarded from a hand, in order.
func handDiscards(s *scene) []int {
	var out []int
	for _, e := range s.g.Log {
		if e.Kind == EventDiscard && e.Zone == ZoneHand.String() {
			out = append(out, e.Inst)
		}
	}
	return out
}

func TestCard_ImprovisedOrders_DrawsThreeThenDiscardsTheCardPicked(t *testing.T) {
	// Covers: FR-MNV-001, FR-MOD-001
	s, inst, hand, deck := improvisedScene(t, []string{"A"}, []string{"D1", "D2", "D3"}, false)
	s.script(0, Answer{Picks: []int{0}})
	s.script(1)
	topBefore := s.g.Players[0].Deck[3]
	s.do(enactAction(inst))
	p := s.g.Players[0]
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecideHand {
		t.Fatalf("the controller is asked once, a Hand decision, for the discard: got %d decisions (FR-MOD-001)", len(asks))
	}
	if len(asks[0].Options) != 4 {
		t.Fatalf("the discard is chosen from all 4 cards in hand, the three drawn included: got %d options (FR-MOD-001)", len(asks[0].Options))
	}
	want := []*Instance{hand[0], deck[0], deck[1], deck[2]}
	for i, w := range want {
		if asks[0].Options[i].Ref != w.ID {
			t.Errorf("option %d should be hand card %d (earlier card, then the 3 drawn in order), got ref %d", i, w.ID, asks[0].Options[i].Ref)
		}
	}
	if !inDiscardPile(p, hand[0]) {
		t.Errorf("the picked card is in its owner's discard pile (FR-MNV-001)")
	}
	if len(p.Hand) != 3 || p.Hand[0] != deck[0] || p.Hand[1] != deck[1] || p.Hand[2] != deck[2] {
		t.Errorf("the three drawn cards stay in hand, got %d cards", len(p.Hand))
	}
	if len(p.Deck) == 0 || p.Deck[0] != topBefore {
		t.Errorf("exactly the top 3 cards were drawn (FR-MNV-001)")
	}
	if lastDiscard(p) != inst {
		t.Errorf("the Maneuver ends in its owner's discard pile after resolving (FR-MNV-001)")
	}
}

func TestCard_ImprovisedOrders_MayDiscardACardJustDrawn(t *testing.T) {
	// Covers: FR-MOD-001
	s, inst, hand, deck := improvisedScene(t, []string{"A"}, []string{"D1", "D2", "D3"}, false)
	s.script(0, Answer{Picks: []int{2}})
	s.do(enactAction(inst))
	p := s.g.Players[0]
	if !inDiscardPile(p, deck[1]) {
		t.Errorf("a card just drawn is a legal discard: the picked D2 is discarded (FR-MOD-001)")
	}
	if len(p.Hand) != 3 || p.Hand[0] != hand[0] || p.Hand[1] != deck[0] || p.Hand[2] != deck[2] {
		t.Errorf("the earlier hand card stays, with D1 and D3, got %d cards", len(p.Hand))
	}
}

func TestCard_ImprovisedOrders_TimeoutDiscardsTheLastCardDrawn(t *testing.T) {
	// Covers: FR-MOD-001
	s, inst, hand, deck := improvisedScene(t, []string{"A"}, []string{"D1", "D2", "D3"}, false)
	s.script(0, Answer{Timeout: true})
	s.do(enactAction(inst))
	p := s.g.Players[0]
	if !inDiscardPile(p, deck[2]) {
		t.Errorf("a timeout discards the card most recently added to the hand, the third card drawn (FR-MOD-001 §6, SVY-R16)")
	}
	if len(p.Hand) != 3 || p.Hand[0] != hand[0] || p.Hand[1] != deck[0] || p.Hand[2] != deck[1] {
		t.Errorf("hand is the earlier card, D1 and D2 after the timeout discard, got %d cards", len(p.Hand))
	}
}

func TestCard_ImprovisedOrders_ItselfIsNeverOffered(t *testing.T) {
	// Covers: FR-MNV-001
	s, inst, _, _ := improvisedScene(t, []string{"A"}, []string{"D1", "D2", "D3"}, false)
	s.script(0, Answer{Picks: []int{0}})
	s.do(enactAction(inst))
	for _, d := range s.asked(0) {
		for _, o := range d.Options {
			if o.Ref == inst.ID {
				t.Errorf("the Maneuver is on the Resolution Stack, not in Hand, so it is never offered for the discard (FR-MNV-001)")
			}
		}
	}
	for _, id := range handDiscards(s) {
		if id == inst.ID {
			t.Errorf("the Maneuver is not the discarded card: its own discard is the end of its resolution (FR-MNV-001)")
		}
	}
	if len(handDiscards(s)) != 1 {
		t.Errorf("exactly one card is discarded from hand, got %d", len(handDiscards(s)))
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver ends in its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_ImprovisedOrders_ShortDeckDrawsWhatIsThereThenLoses(t *testing.T) {
	// Covers: FR-WIN-001, FR-MOD-001
	s, inst, _, deck := improvisedScene(t, nil, []string{"D1"}, true)
	s.script(0)
	s.script(1)
	s.do(enactAction(inst))
	p := s.g.Players[0]
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("the one drawn card is the only card in hand, so it is discarded without asking: asked %d times (FR-MOD-001 §1, SVY-R16)", n)
	}
	if !inDiscardPile(p, deck[0]) || len(p.Hand) != 0 || len(p.Deck) != 0 {
		t.Errorf("the one card in the deck is drawn and then discarded, leaving hand and deck empty (FR-WIN-001)")
	}
	if o := s.g.Outcome; !o.Over || o.Draw || o.Winner != 1 {
		t.Fatalf("player 0 drew from an empty deck and loses at the death sweep after resolving, outcome %+v (FR-WIN-001)", o)
	}
	disc, end := -1, -1
	for _, e := range s.g.Log {
		if e.Kind == EventDiscard && e.Inst == deck[0].ID {
			disc = e.Seq
		}
		if e.Kind == EventMatchEnd {
			end = e.Seq
		}
	}
	if disc < 0 || end < 0 || disc > end {
		t.Errorf("the discard happens before the loss: discard seq %d, match-end seq %d (FR-WIN-001)", disc, end)
	}
}

func TestCard_ImprovisedOrders_OnlyTheControllersHandAndDeckAreUsed(t *testing.T) {
	// Covers: FR-MNV-001
	s, inst, _, _ := improvisedScene(t, []string{"A"}, []string{"D1", "D2", "D3"}, false)
	foe := s.g.Players[1]
	hand, deck, disc := len(foe.Hand), len(foe.Deck), len(foe.Discard)
	s.script(0, Answer{Picks: []int{0}})
	s.script(1)
	s.do(enactAction(inst))
	if len(foe.Hand) != hand || len(foe.Deck) != deck || len(foe.Discard) != disc {
		t.Errorf("the opponent's hand, deck and discard are unchanged: hand %d->%d deck %d->%d discard %d->%d (FR-MNV-001)",
			hand, len(foe.Hand), deck, len(foe.Deck), disc, len(foe.Discard))
	}
	if n := len(s.asked(1)); n != 0 {
		t.Errorf("the opponent is never asked: asked %d times", n)
	}
}

func TestCard_ImprovisedOrders_PaysTwoTactics(t *testing.T) {
	// Covers: FR-MNV-001
	s, inst, _, _ := improvisedScene(t, []string{"A"}, []string{"D1", "D2", "D3"}, false)
	if !offersEnact(s.g, inst) {
		t.Fatalf("with an empty battlefield it is offered: it has no target slot (SVY-R29, FR-MNV-001)")
	}
	s.script(0, Answer{Picks: []int{0}})
	s.do(enactAction(inst))
	p := s.g.Players[0]
	if p.Tac != 3 || p.Req != 3 {
		t.Errorf("enacting costs 2 Tactics and no Requisition: pool R%d T%d, want R3 T3 (FR-MNV-001)", p.Req, p.Tac)
	}
}

// Tests for T-0133 (FR-MNV-001, FR-TGT-001, FR-TGT-002, FR-KWD-002h,
// FR-ZON-004, FR-ZON-005): the Path Maneuver Tactical Withdrawal, loaded from
// data/cards. Player 0 enacts; player 1 is the opponent.
//
// Tactical Withdrawal: "Return up to one chosen allied unit to its owner's
// hand. Draw a card." Cost 1 Tactics.

func TestCard_TacticalWithdrawal_ReturnsChosenAlliedUnitThenDraws(t *testing.T) {
	// Covers: FR-MNV-001, FR-TGT-001, FR-ZON-004
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 3, Keywords{}))
	other := s.put(0, ZoneMainTheater, unit("Other", 1, 3, Keywords{}))
	top := s.onTop(0, unit("Top", 1, 1, Keywords{}))[0]
	inst := enactFor(s, "Tactical Withdrawal")
	chooseAsset(s, ally)
	s.do(enactAction(inst))
	p := s.g.Players[0]
	if s.alive(ally) || !inHand(p, ally.Inst) {
		t.Errorf("the chosen allied unit leaves the battlefield and is in its owner's hand (FR-MNV-001, FR-ZON-004)")
	}
	if !s.alive(other) {
		t.Errorf("the unchosen allied unit stays on the battlefield (FR-TGT-001)")
	}
	if !inHand(p, top) {
		t.Errorf("the top card of the deck is drawn (FR-MNV-001)")
	}
	if n := len(p.Hand); n < 2 || p.Hand[n-2] != ally.Inst || p.Hand[n-1] != top {
		t.Errorf("the unit is returned before the card is drawn: hand %v, want the unit then the drawn card last (FR-MNV-001)", p.Hand)
	}
	if lastDiscard(p) != inst {
		t.Errorf("the Maneuver ends in its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_TacticalWithdrawal_ChoosingNoneStillDraws(t *testing.T) {
	// Covers: FR-TGT-002, SVY-R11
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 3, Keywords{}))
	top := s.onTop(0, unit("Top", 1, 1, Keywords{}))[0]
	s.script(0, Answer{Picks: []int{}})
	s.do(enactAction(enactFor(s, "Tactical Withdrawal")))
	p := s.g.Players[0]
	if !s.alive(ally) || inHand(p, ally.Inst) {
		t.Errorf("choosing no unit returns nothing (FR-TGT-002)")
	}
	if !inHand(p, top) {
		t.Errorf("the card is still drawn when no unit is chosen (FR-MNV-001)")
	}
}

func TestCard_TacticalWithdrawal_EnactableWithNoAlliedUnit(t *testing.T) {
	// Covers: SVY-R29, FR-TGT-002
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 3, Keywords{}))
	wall := s.put(0, ZoneMainTheater, structure("AWall", 0, 3, Keywords{}))
	top := s.onTop(0, unit("Top", 1, 1, Keywords{}))[0]
	inst := enactFor(s, "Tactical Withdrawal")
	if !offersEnact(s.g, inst) {
		t.Fatalf("an \"up to\" slot never stops an Enact, even with no allied unit (SVY-R29)")
	}
	s.script(0)
	s.do(enactAction(inst))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("no legal target, so no decision is asked: asked %d times (FR-TGT-002)", n)
	}
	if !s.alive(foe) || !s.alive(wall) {
		t.Errorf("nothing leaves the battlefield")
	}
	if !inHand(s.g.Players[0], top) {
		t.Errorf("the card is still drawn (FR-MNV-001)")
	}
}

func TestCard_TacticalWithdrawal_OnlyAlliedUnitsAreOffered(t *testing.T) {
	// Covers: FR-TGT-001
	s := newScene(t)
	u1 := s.put(0, ZoneReserves, unit("A1", 1, 3, Keywords{}))
	u2 := s.put(0, ZoneMainTheater, unit("A2", 1, 3, Keywords{}))
	st := s.put(0, ZoneMainTheater, structure("AWall", 0, 3, Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 3, Keywords{}))
	chooseAsset(s, u1)
	s.do(enactAction(enactFor(s, "Tactical Withdrawal")))
	got := targetOffers(s)
	if len(got) != 2 || !got[u1.ID] || !got[u2.ID] {
		t.Errorf("exactly the two allied units are offered, got refs %v (FR-TGT-001)", got)
	}
	for _, a := range []*Asset{st, foe, s.g.Players[0].HQ, s.g.Players[1].HQ} {
		if got[a.ID] {
			t.Errorf("%s is not an allied unit and must not be offered (FR-TGT-001)", a)
		}
	}
}

func TestCard_TacticalWithdrawal_AlliedWardUnitCanBeChosen(t *testing.T) {
	// Covers: FR-KWD-002h
	s := newScene(t)
	ward := s.put(0, ZoneMainTheater, unit("Warded", 1, 3, Keywords{Ward: true}))
	chooseAsset(s, ward)
	s.do(enactAction(enactFor(s, "Tactical Withdrawal")))
	if !targetOffers(s)[ward.ID] {
		t.Errorf("an allied Ward unit is offered: Ward stops only opponents (FR-KWD-002h)")
	}
	if s.alive(ward) || !inHand(s.g.Players[0], ward.Inst) {
		t.Errorf("the chosen allied Ward unit is returned to its owner's hand (FR-KWD-002h)")
	}
}

func TestCard_TacticalWithdrawal_FaceDownScratchCompanyReturnsAsItsCard(t *testing.T) {
	// Covers: FR-ZON-004, FR-ZON-005, FR-TGT-001
	s := newScene(t)
	sc := s.putScratch(0)
	chooseAsset(s, sc)
	s.do(enactAction(enactFor(s, "Tactical Withdrawal")))
	if s.alive(sc) || !inHand(s.g.Players[0], sc.Inst) {
		t.Errorf("an allied face-down Scratch Company is a unit: it is chosen and its printed card goes to its owner's hand (FR-ZON-004, FR-ZON-005)")
	}
}

func TestCard_TacticalWithdrawal_PaysOneTactics(t *testing.T) {
	// Covers: FR-MNV-001
	s := newScene(t)
	s.script(0)
	s.do(enactAction(enactFor(s, "Tactical Withdrawal")))
	if p := s.g.Players[0]; p.Tac != 4 || p.Req != 3 {
		t.Errorf("enacting costs 1 Tactics and no Requisition: pool R%d T%d, want R3 T4 (FR-MNV-001)", p.Req, p.Tac)
	}
}

// Tests for T-0135 (FR-MNV-001, FR-TGT-001, FR-KWD-002h, FR-KWD-002k,
// FR-LAY-001, FR-LAY-002, FR-LAY-003, FR-REP-004, FR-ZON-005): the Path
// Maneuvers Chainbreaker's Tenacity and Fortify, loaded from data/cards.
// Player 0 enacts; player 1 is the opponent.
//
// Chainbreaker's Tenacity: "Chosen allied unit gets +1 power until end of turn
// and gains 2 armor." Cost 1 Tactics.
// Fortify: "Chosen allied unit gains 3 armor." Cost 1 Tactics.

// armorPair puts two plain allied units (printed power 2, Health 5) in the Main Theater.
func armorPair(s *scene) (*Asset, *Asset) {
	a := s.put(0, ZoneMainTheater, unit("Ally", 2, 5, Keywords{}))
	b := s.put(0, ZoneMainTheater, unit("Other", 2, 5, Keywords{}))
	return a, b
}

// armorOffers enacts the named Maneuver with two allied units, an allied
// structure, an opposing unit and both Headquarters on the battlefield, and
// checks that exactly the two allied units are offered.
func armorOffers(t *testing.T, name string) {
	t.Helper()
	s := newScene(t)
	u1 := s.put(0, ZoneReserves, unit("A1", 1, 3, Keywords{}))
	u2 := s.put(0, ZoneMainTheater, unit("A2", 1, 3, Keywords{}))
	st := s.put(0, ZoneMainTheater, structure("AWall", 0, 3, Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 3, Keywords{}))
	chooseAsset(s, u1)
	s.do(enactAction(enactFor(s, name)))
	got := targetOffers(s)
	if len(got) != 2 || !got[u1.ID] || !got[u2.ID] {
		t.Errorf("exactly the two allied units are offered, got refs %v (FR-TGT-001)", got)
	}
	for _, a := range []*Asset{st, foe, s.g.Players[0].HQ, s.g.Players[1].HQ} {
		if got[a.ID] {
			t.Errorf("%s is not an allied unit and must not be offered (FR-TGT-001)", a)
		}
	}
}

// armorNotEnactable checks SVY-R29 for the named Maneuver.
func armorNotEnactable(t *testing.T, name string) {
	t.Helper()
	cases := []struct {
		name  string
		setup func(s *scene)
		want  bool
	}{
		{"no unit at all", func(s *scene) {}, false},
		{"only an allied structure and opposing units", func(s *scene) {
			s.put(0, ZoneMainTheater, structure("AWall", 0, 3, Keywords{}))
			s.put(1, ZoneMainTheater, unit("Foe", 1, 3, Keywords{}))
		}, false},
		{"one allied unit", func(s *scene) {
			s.put(0, ZoneReserves, unit("Ally", 1, 3, Keywords{}))
		}, true},
	}
	for _, c := range cases {
		s := newScene(t)
		c.setup(s)
		inst := enactFor(s, name)
		if got := offersEnact(s.g, inst); got != c.want {
			t.Errorf("%s: Enact offered = %v, want %v (SVY-R29)", c.name, got, c.want)
		}
		if !c.want {
			s.reject(enactAction(inst))
			if !inHand(s.g.Players[0], inst) {
				t.Errorf("%s: the card stays in hand (SVY-R29)", c.name)
			}
		}
	}
}

func TestCard_ChainbreakersTenacity_ChosenAllyGetsPlusOnePowerAndTwoArmor(t *testing.T) {
	// Covers: FR-MNV-001, FR-TGT-001, FR-LAY-001, FR-KWD-002k
	s := newScene(t)
	ally, other := armorPair(s)
	inst := enactFor(s, "Chainbreaker's Tenacity")
	chooseAsset(s, ally)
	s.do(enactAction(inst))
	if got := s.g.Power(ally); got != 3 {
		t.Errorf("the chosen unit has printed power 2 +1: power %d, want 3 (FR-LAY-001)", got)
	}
	if ally.Armor != 2 {
		t.Errorf("the chosen unit gains 2 armor: Armor %d, want 2 (FR-KWD-002k)", ally.Armor)
	}
	if got := s.g.Power(other); got != 2 || other.Armor != 0 {
		t.Errorf("the unchosen unit is unchanged: power %d Armor %d, want 2 and 0 (FR-TGT-001)", got, other.Armor)
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver ends in its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_ChainbreakersTenacity_PowerExpiresAtEndOfTurnArmorStays(t *testing.T) {
	// Covers: FR-LAY-003, FR-KWD-002k
	s := newScene(t)
	ally, _ := armorPair(s)
	chooseAsset(s, ally)
	s.do(enactAction(enactFor(s, "Chainbreaker's Tenacity")))
	s.endTurn()
	if got := s.g.Power(ally); got != 2 {
		t.Errorf("the +1 power expires at the end of the turn: power %d, want 2 (FR-LAY-003)", got)
	}
	if ally.Armor != 2 {
		t.Errorf("armor never expires at end of turn: Armor %d, want 2 (FR-KWD-002k)", ally.Armor)
	}
}

func TestCard_ChainbreakersTenacity_OnlyAlliedUnitsAreOffered(t *testing.T) {
	// Covers: FR-TGT-001
	armorOffers(t, "Chainbreaker's Tenacity")
}

func TestCard_ChainbreakersTenacity_AlliedWardUnitCanBeChosen(t *testing.T) {
	// Covers: FR-KWD-002h
	s := newScene(t)
	ward := s.put(0, ZoneMainTheater, unit("Warded", 2, 5, Keywords{Ward: true}))
	plain := s.put(0, ZoneMainTheater, unit("Plain", 2, 5, Keywords{}))
	chooseAsset(s, ward)
	s.do(enactAction(enactFor(s, "Chainbreaker's Tenacity")))
	got := targetOffers(s)
	if !got[ward.ID] || !got[plain.ID] {
		t.Errorf("an allied Ward unit is offered beside the plain one: Ward stops only opponents, got refs %v (FR-KWD-002h)", got)
	}
	if p := s.g.Power(ward); p != 3 || ward.Armor != 2 {
		t.Errorf("the chosen Ward unit gets +1 power and 2 armor: power %d Armor %d, want 3 and 2 (FR-KWD-002h)", p, ward.Armor)
	}
}

func TestCard_ChainbreakersTenacity_NotEnactableWithoutAlliedUnit(t *testing.T) {
	// Covers: SVY-R29, FR-MNV-001
	armorNotEnactable(t, "Chainbreaker's Tenacity")
}

func TestCard_ChainbreakersTenacity_FaceDownScratchCompanyCanBeChosen(t *testing.T) {
	// Covers: FR-ZON-005, FR-TGT-001
	s := newScene(t)
	sc := s.putScratch(0)
	s.put(0, ZoneMainTheater, unit("Other", 2, 5, Keywords{}))
	chooseAsset(s, sc)
	s.do(enactAction(enactFor(s, "Chainbreaker's Tenacity")))
	if got := s.g.Power(sc); got != 2 || sc.Armor != 2 {
		t.Errorf("an allied face-down Scratch Company is a unit: power %d Armor %d, want 2 and 2 (FR-ZON-005)", got, sc.Armor)
	}
}

func TestCard_ChainbreakersTenacity_PaysOneTactics(t *testing.T) {
	// Covers: FR-MNV-001
	s := newScene(t)
	ally, _ := armorPair(s)
	chooseAsset(s, ally)
	s.do(enactAction(enactFor(s, "Chainbreaker's Tenacity")))
	if p := s.g.Players[0]; p.Tac != 4 || p.Req != 3 {
		t.Errorf("enacting costs 1 Tactics and no Requisition: pool R%d T%d, want R3 T4 (FR-MNV-001)", p.Req, p.Tac)
	}
}

func TestCard_Fortify_ChosenAllyGainsThreeArmor(t *testing.T) {
	// Covers: FR-MNV-001, FR-TGT-001, FR-KWD-002k
	s := newScene(t)
	ally, other := armorPair(s)
	inst := enactFor(s, "Fortify")
	chooseAsset(s, ally)
	s.do(enactAction(inst))
	if ally.Armor != 3 {
		t.Errorf("the chosen unit gains 3 armor: Armor %d, want 3 (FR-KWD-002k)", ally.Armor)
	}
	if got := s.g.Power(ally); got != 2 {
		t.Errorf("Fortify does not change power: %d, want 2", got)
	}
	if got := s.g.Power(other); got != 2 || other.Armor != 0 {
		t.Errorf("the unchosen unit is unchanged: power %d Armor %d, want 2 and 0 (FR-TGT-001)", got, other.Armor)
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver ends in its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_Fortify_ArmorStacksAndOutlastsTheTurn(t *testing.T) {
	// Covers: FR-LAY-002, FR-KWD-002k
	s := newScene(t)
	armored := s.put(0, ZoneMainTheater, unit("Armored", 2, 5, Keywords{Armor: 2}))
	s.put(0, ZoneMainTheater, unit("Other", 2, 5, Keywords{}))
	chooseAsset(s, armored)
	s.do(enactAction(enactFor(s, "Fortify")))
	if armored.Armor != 5 {
		t.Errorf("granted armor adds to armor the unit already has: Armor %d, want 5 (FR-LAY-002)", armored.Armor)
	}
	s.endTurn()
	if armored.Armor != 5 {
		t.Errorf("armor does not expire at end of turn: Armor %d, want 5 (FR-KWD-002k)", armored.Armor)
	}
}

func TestCard_Fortify_ArmorAbsorbsLaterDamage(t *testing.T) {
	// Covers: FR-REP-004, FR-LAY-002
	s := newScene(t)
	ally, _ := armorPair(s)
	chooseAsset(s, ally)
	s.do(enactAction(enactFor(s, "Fortify")))
	zap := s.give(0, maneuver("Zap", 0, 0, ability.Effect{
		Targets: []ability.Target{tgt(ability.AllegianceAllied, ability.KindUnit, false)},
		Steps:   []ability.Step{dmgTo(2, 0)},
	}))
	s.do(enactAction(zap))
	if ally.Armor != 1 {
		t.Errorf("2 damage removes 2 of the 3 armor 1-for-1: Armor %d, want 1 (FR-REP-004)", ally.Armor)
	}
	s.wantDamage(ally, 0)
}

func TestCard_Fortify_OnlyAlliedUnitsAreOffered(t *testing.T) {
	// Covers: FR-TGT-001
	armorOffers(t, "Fortify")
}

func TestCard_Fortify_NotEnactableWithoutAlliedUnit(t *testing.T) {
	// Covers: SVY-R29, FR-MNV-001
	armorNotEnactable(t, "Fortify")
}

func TestCard_Fortify_PaysOneTactics(t *testing.T) {
	// Covers: FR-MNV-001
	s := newScene(t)
	ally, _ := armorPair(s)
	chooseAsset(s, ally)
	s.do(enactAction(enactFor(s, "Fortify")))
	if p := s.g.Players[0]; p.Tac != 4 || p.Req != 3 {
		t.Errorf("enacting costs 1 Tactics and no Requisition: pool R%d T%d, want R3 T4 (FR-MNV-001)", p.Req, p.Tac)
	}
}

// Tests for T-0139 (FR-TRN-003a, FR-MNV-001, FR-TGT-001, FR-MOD-001; SVY-R3,
// SVY-R4, SVY-R16, SVY-R32): the Path Maneuver Steel Rain, loaded from
// data/cards. Player 0 enacts; player 1 is the opponent.
//
// Steel Rain: "As an additional cost to enact this maneuver, discard up to 3
// cards. For each card discarded this way, deal 2 damage to chosen asset. Then,
// if you have no cards in hand, draw 2 cards." Cost 4 Tactics.

// steelRainScene gives player 0 an empty hand, then the real Steel Rain and n
// vanilla cards (in that hand order), and enough resources to enact. It
// returns the scene, the Maneuver and the other hand cards.
func steelRainScene(t *testing.T, n int) (*scene, *Instance, []*Instance) {
	t.Helper()
	s := newScene(t)
	s.g.Players[0].Hand = nil
	inst := enactFor(s, "Steel Rain")
	var others []*Instance
	for i := 0; i < n; i++ {
		others = append(others, s.give(0, unit(string(rune('A'+i)), 1, 1, Keywords{})))
	}
	return s, inst, others
}

func TestCard_SteelRain_DealsTwoToAChosenAssetPerDiscard(t *testing.T) {
	// Covers: FR-TRN-003a, FR-MNV-001, FR-TGT-001, SVY-R4
	s, inst, others := steelRainScene(t, 4)
	a := s.put(1, ZoneMainTheater, unit("Victim A", 1, 9, Keywords{}))
	b := s.put(1, ZoneMainTheater, unit("Victim B", 1, 9, Keywords{}))
	pdScript(s, 2, false, a, b)
	s.do(enactAction(inst))
	s.wantDamage(a, 2)
	s.wantDamage(b, 2)
	amounts, targets := maneuverDamage(s)
	if !reflect.DeepEqual(amounts, []int{2, 2}) || !reflect.DeepEqual(targets, []int{a.ID, b.ID}) {
		t.Errorf("two separate 2-damage events in pick order A, B (SVY-R4): got amounts %v on %v", amounts, targets)
	}
	p := s.g.Players[0]
	if !inDiscard(p, others[0]) || !inDiscard(p, others[1]) {
		t.Errorf("the two discarded cards are in the discard pile (FR-TRN-003a)")
	}
	if lastDiscard(p) != inst {
		t.Errorf("the Maneuver goes to the discard pile after the discarded cards (FR-MNV-001)")
	}
	if len(p.Hand) != 2 || p.Hand[0] != others[2] || p.Hand[1] != others[3] {
		t.Errorf("two cards stay in hand and nothing is drawn (hand not empty): got %d cards", len(p.Hand))
	}
}

func TestCard_SteelRain_SameAssetForEveryDiscard(t *testing.T) {
	// Covers: SVY-R4, FR-REP-004
	s, inst, _ := steelRainScene(t, 3)
	u1 := s.put(1, ZoneMainTheater, unit("Victim One", 1, 9, Keywords{}))
	u2 := s.put(1, ZoneMainTheater, unit("Victim Two", 1, 9, Keywords{}))
	pdScript(s, 3, false, u1, u1, u1)
	s.do(enactAction(inst))
	s.wantDamage(u1, 6)
	s.wantDamage(u2, 0)
	amounts, targets := maneuverDamage(s)
	if !reflect.DeepEqual(amounts, []int{2, 2, 2}) || !reflect.DeepEqual(targets, []int{u1.ID, u1.ID, u1.ID}) {
		t.Errorf("each repetition is its own 2-damage instance on the one unit (FR-REP-004): got amounts %v on %v", amounts, targets)
	}
}

func TestCard_SteelRain_ChosenAssetIsAnyAsset(t *testing.T) {
	// Covers: FR-TGT-001, FR-KWD-002h
	s, inst, _ := steelRainScene(t, 2)
	own, opp := s.g.Players[0].HQ, s.g.Players[1].HQ
	allyUnit := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{}))
	allyStruct := s.put(0, ZoneMainTheater, structure("Ally Wall", 0, 9, Keywords{}))
	allyWard := s.put(0, ZoneMainTheater, unit("Warded Ally", 1, 9, Keywords{Ward: true}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 9, Keywords{}))
	foeWard := s.put(1, ZoneMainTheater, unit("Warded Foe", 1, 9, Keywords{Ward: true}))
	pdScript(s, 2, false, foe, foe)
	s.do(enactAction(inst))
	var d *Decision
	for _, x := range s.asked(0) {
		if x.Kind == DecideTarget {
			x := x
			d = &x
		}
	}
	if d == nil {
		t.Fatalf("a target decision is asked after the discard (SVY-R3)")
	}
	got := map[int]bool{}
	for _, o := range d.Options {
		got[o.Ref] = true
	}
	for name, a := range map[string]*Asset{"own Headquarters": own, "opposing Headquarters": opp, "allied unit": allyUnit, "allied structure": allyStruct, "allied Ward unit": allyWard, "opposing unit": foe} {
		if !got[a.ID] {
			t.Errorf("%s is a legal chosen asset (FR-TGT-001)", name)
		}
	}
	if got[foeWard.ID] {
		t.Errorf("an opponent's Ward unit is never a legal target (FR-KWD-002h)")
	}
}

func TestCard_SteelRain_DiscardingNoneDealsNothing(t *testing.T) {
	// Covers: SVY-R4, FR-TRN-003a
	s, inst, others := steelRainScene(t, 2)
	s.put(1, ZoneMainTheater, unit("Victim One", 1, 9, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Victim Two", 1, 9, Keywords{}))
	pdScript(s, 0, false)
	s.do(enactAction(inst))
	for _, d := range s.asked(0) {
		if d.Kind == DecideTarget {
			t.Errorf("discarding none asks no target (SVY-R4)")
		}
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("discarding none deals no damage, got %d (SVY-R4)", got)
	}
	if p := s.g.Players[0]; len(p.Hand) != 2 || p.Hand[0] != others[0] || p.Hand[1] != others[1] {
		t.Errorf("both cards stay in hand and nothing is drawn (the hand is not empty): got %d cards", len(p.Hand))
	}
}

func TestCard_SteelRain_EmptyHandAfterTheDiscardDrawsTwo(t *testing.T) {
	// Covers: FR-MNV-001, FR-TRN-003a
	for _, tc := range []struct {
		discards int
		draws    bool
	}{{3, true}, {2, false}} {
		s, inst, _ := steelRainScene(t, 3)
		s.put(1, ZoneMainTheater, unit("Victim", 1, 9, Keywords{}))
		top := s.onTop(0, vanilla("D1", "D2", "D3")...)
		foeHQ := s.g.Players[1].HQ
		pdScript(s, tc.discards, false, foeHQ, foeHQ, foeHQ)
		s.do(enactAction(inst))
		p := s.g.Players[0]
		if tc.draws {
			if len(p.Hand) != 2 || p.Hand[0] != top[0] || p.Hand[1] != top[1] {
				t.Errorf("discarding all three empties the hand, so the top two cards of the deck are drawn: got %d cards in hand (FR-MNV-001)", len(p.Hand))
			}
		} else if len(p.Hand) != 1 {
			t.Errorf("discarding two of three leaves a card in hand, so nothing is drawn: got %d cards in hand (FR-MNV-001)", len(p.Hand))
		}
	}
}

func TestCard_SteelRain_AloneInHandDrawsTwo(t *testing.T) {
	// Covers: FR-MNV-001, FR-TRN-003a, SVY-R16
	s, inst, _ := steelRainScene(t, 0)
	s.put(1, ZoneMainTheater, unit("Victim", 1, 9, Keywords{}))
	top := s.onTop(0, vanilla("D1", "D2", "D3")...)
	if !offersEnact(s.g, inst) {
		t.Fatalf("Steel Rain alone in hand is offered: it may discard up to 3 cards, i.e. none")
	}
	pdScript(s, 3, false)
	s.do(enactAction(inst))
	if asks := s.asked(0); len(asks) != 0 {
		t.Errorf("with no other card there is nothing to discard and no target to choose: asked %v", kindsOf(asks))
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("no damage is dealt, got %d", got)
	}
	if p := s.g.Players[0]; len(p.Hand) != 2 || p.Hand[0] != top[0] || p.Hand[1] != top[1] {
		t.Errorf("the Maneuver left the hand when enacted, so the hand is empty as the draw resolves and 2 cards are drawn: got %d in hand (FR-MNV-001)", len(p.Hand))
	}
}

func TestCard_SteelRain_TimeoutConcentratesOnTheFirstLegalAsset(t *testing.T) {
	// Covers: SVY-R32, FR-MOD-001
	s, inst, _ := steelRainScene(t, 3)
	u1 := s.put(1, ZoneMainTheater, unit("Victim One", 1, 9, Keywords{}))
	u2 := s.put(1, ZoneMainTheater, unit("Victim Two", 1, 9, Keywords{}))
	pdScript(s, 3, true)
	s.do(enactAction(inst))
	s.wantDamage(s.g.Players[0].HQ, 6)
	s.wantDamage(s.g.Players[1].HQ, 0)
	s.wantDamage(u1, 0)
	s.wantDamage(u2, 0)
}

func TestCard_SteelRain_PaysFourTactics(t *testing.T) {
	// Covers: FR-MNV-001
	s, inst, _ := steelRainScene(t, 1)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 9, Keywords{}))
	p := s.g.Players[0]
	p.Tac = 3
	if offersEnact(s.g, inst) {
		t.Errorf("with 3 Tactics Steel Rain is not offered (FR-MNV-001)")
	}
	s.reject(enactAction(inst))
	p.Tac = 5
	if !offersEnact(s.g, inst) {
		t.Fatalf("with enough Tactics Steel Rain is offered (FR-MNV-001)")
	}
	pdScript(s, 1, false, foe)
	s.do(enactAction(inst))
	if p.Tac != 1 || p.Req != 3 {
		t.Errorf("enacting costs 4 Tactics and no Requisition: pool R%d T%d, want R3 T1 (FR-MNV-001)", p.Req, p.Tac)
	}
}
