package engine

import (
	"slices"
	"testing"
)

// Tests for the one place that moves a card between deck, hand and discard
// (takeCard and putCard).

// piles is a snapshot of every pile of both players, by instance.
type piles [2][3][]*Instance

func snapPiles(g *Game) piles {
	var out piles
	for i, p := range g.Players {
		out[i] = [3][]*Instance{slices.Clone(p.Deck), slices.Clone(p.Hand), slices.Clone(p.Discard)}
	}
	return out
}

func (s *scene) loose(p PlayerID) *Instance {
	return &Instance{ID: s.g.newID(), Card: unit("Loose", 1, 1, Keywords{}), Owner: p}
}

func samePiles(a, b piles) bool {
	for i := range a {
		for j := range a[i] {
			if !slices.Equal(a[i][j], b[i][j]) {
				return false
			}
		}
	}
	return true
}

func TestPutCard_DeckTop(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	before := slices.Clone(p.Deck)
	x := s.loose(0)
	s.g.putCard(x, ZoneDeck, false)
	if p.Deck[0] != x || !slices.Equal(p.Deck[1:], before) {
		t.Errorf("deck top: x not first, or the others reordered")
	}
}

func TestPutCard_DeckBottom(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	before := slices.Clone(p.Deck)
	x := s.loose(0)
	s.g.putCard(x, ZoneDeck, true)
	if p.Deck[len(p.Deck)-1] != x || !slices.Equal(p.Deck[:len(p.Deck)-1], before) {
		t.Errorf("deck bottom: x not last, or the others reordered")
	}
}

func TestPutCard_HandAndDiscardAddAtEnd(t *testing.T) {
	for _, bottom := range []bool{false, true} {
		s := newScene(t)
		p := s.g.Players[0]
		hand, discard := slices.Clone(p.Hand), slices.Clone(p.Discard)
		x, y := s.loose(0), s.loose(0)
		s.g.putCard(x, ZoneHand, bottom)
		s.g.putCard(y, ZoneDiscard, bottom)
		if p.Hand[len(p.Hand)-1] != x || !slices.Equal(p.Hand[:len(hand)], hand) {
			t.Errorf("bottom=%v: hand does not end with x", bottom)
		}
		if p.Discard[len(p.Discard)-1] != y || !slices.Equal(p.Discard[:len(discard)], discard) {
			t.Errorf("bottom=%v: discard does not end with y", bottom)
		}
	}
}

func TestPutCard_OwnersPileWhenOtherPlayerActive(t *testing.T) {
	s := newScene(t)
	s.g.Active = 1
	before := snapPiles(s.g)
	x := s.loose(0)
	s.g.putCard(x, ZoneHand, false)
	after := snapPiles(s.g)
	if after[0][1][len(after[0][1])-1] != x {
		t.Errorf("x is not at the end of its owner's hand")
	}
	after[0][1] = after[0][1][:len(after[0][1])-1]
	if !samePiles(before, after) {
		t.Errorf("another pile changed")
	}
}

func TestTakeCard_FromEachPile(t *testing.T) {
	for pile, name := range []string{"deck", "hand", "discard"} {
		s := newScene(t)
		p := s.g.Players[0]
		p.Discard = append(p.Discard, s.loose(0), s.loose(0), s.loose(0))
		all := [3]*[]*Instance{&p.Deck, &p.Hand, &p.Discard}
		src := *all[pile]
		x := src[1]
		want := slices.Clone(slices.DeleteFunc(slices.Clone(src), func(i *Instance) bool { return i == x }))
		if !s.g.takeCard(x) {
			t.Errorf("%s: takeCard = false, want true", name)
		}
		for _, q := range all {
			if slices.Contains(*q, x) {
				t.Errorf("%s: x still in a pile", name)
			}
		}
		if !slices.Equal(*all[pile], want) {
			t.Errorf("%s: the rest of the pile changed order", name)
		}
	}
}

func TestTakeCard_InNoPile(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneReserves, unit("Soldier", 1, 1, Keywords{}))
	before := snapPiles(s.g)
	if s.g.takeCard(a.Inst) {
		t.Errorf("takeCard on a battlefield asset's instance = true, want false")
	}
	if !samePiles(before, snapPiles(s.g)) {
		t.Errorf("a pile changed")
	}
}

func TestPutCard_PanicsOnOtherZones(t *testing.T) {
	for _, z := range []Zone{ZoneStack, ZoneReserves} {
		s := newScene(t)
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("putCard to zone %v did not panic", z)
				}
			}()
			s.g.putCard(s.loose(0), z, false)
		}()
	}
}
