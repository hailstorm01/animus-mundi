package engine

import "testing"

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
