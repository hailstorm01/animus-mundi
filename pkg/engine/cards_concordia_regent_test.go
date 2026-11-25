package engine

import (
	"animus-mundi/pkg/engine/ability"
	"reflect"
	"testing"
)

// Tests for T-0300: Icewind Regent, "Exhaust: Return chosen unit to its owner's
// hand." (FR-ACT-001, FR-ACT-002, FR-TGT-001, FR-ZON-003, FR-ZON-004, FR-ZON-005,
// FR-KWD-002h, SVY-R36). The real card is loaded from data/cards. Player 0 is the
// active player and activates, with an empty pool.

// regentScene is a scene with a Readied Icewind Regent in player 0's Reserves.
func regentScene(t *testing.T) (*scene, *Asset) {
	t.Helper()
	s := newScene(t)
	r := s.put(0, ZoneReserves, s.card("Icewind Regent"))
	s.setPool(0, 0, 0)
	return s, r
}

// regentChoose makes seat 0 pick the option whose Ref is want, recording every
// decision it is asked in s.asked(0).
func regentChoose(s *scene, want *Asset) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		for i, o := range d.Options {
			if o.Ref == want.ID {
				return Answer{Picks: []int{i}}
			}
		}
		s.t.Errorf("the chosen asset %d is not among the options %v", want.ID, refs(d))
		return Answer{Timeout: true}
	}
}

func TestCard_IcewindRegent_ReturnsChosenOpposingUnit(t *testing.T) {
	// Covers: FR-ACT-001, FR-ZON-004
	s, r := regentScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 3, 5, ability.Keywords{}))
	foe.Damage = 2
	other := s.put(1, ZoneMainTheater, unit("Other", 2, 5, ability.Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 2, 5, ability.Keywords{}))
	regentChoose(s, foe)
	s.do(activate(r, 0))
	if s.alive(foe) || inAssets(s.g.Theater, foe) {
		t.Errorf("the chosen unit leaves the battlefield (FR-ZON-004)")
	}
	p1 := s.g.Players[1]
	if !inHand(p1, foe.Inst) {
		t.Errorf("the chosen unit's card is in its owner's hand (FR-ZON-004)")
	}
	if inDiscardPile(p1, foe.Inst) || inHand(s.g.Players[0], foe.Inst) {
		t.Errorf("the card is not in a discard pile or the controller's hand (FR-ZON-004)")
	}
	if !r.Exhausted || !s.alive(r) {
		t.Errorf("the Regent is Exhausted and stays on the battlefield (FR-ACT-001): exhausted %v", r.Exhausted)
	}
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
		t.Errorf("no Requisition or Tactics is spent: pool R%d T%d, want R0 T0 (FR-ACT-001)", p.Req, p.Tac)
	}
	if !s.alive(other) || other.Damage != 0 || !s.alive(ally) || ally.Damage != 0 {
		t.Errorf("the other units are untouched (FR-ZON-004)")
	}
}

func TestCard_IcewindRegent_AnyUnitInAnyZoneOffered(t *testing.T) {
	// Covers: FR-TGT-001
	s, r := regentScene(t)
	allyMain := s.put(0, ZoneMainTheater, unit("AllyMain", 2, 5, ability.Keywords{}))
	allyRes := s.put(0, ZoneReserves, unit("AllyRes", 2, 5, ability.Keywords{}))
	foeMain := s.put(1, ZoneMainTheater, unit("FoeMain", 2, 5, ability.Keywords{}))
	foeRes := s.put(1, ZoneReserves, unit("FoeRes", 2, 5, ability.Keywords{}))
	scratch := s.putScratch(1)
	s.put(0, ZoneMainTheater, structure("AllyWall", 0, 5, ability.Keywords{}))
	s.put(1, ZoneMainTheater, structure("FoeWall", 0, 5, ability.Keywords{}))
	regentChoose(s, foeMain)
	s.do(activate(r, 0))
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecideTarget {
		t.Fatalf("one target question is asked, got %d decisions (FR-TGT-001)", len(asks))
	}
	want := sortedCopy([]int{r.ID, allyMain.ID, allyRes.ID, foeMain.ID, foeRes.ID, scratch.ID})
	if got := refs(asks[0]); !reflect.DeepEqual(got, want) {
		t.Errorf("options are every unit on the battlefield (the Regent and a Scratch Company included; no structure or Headquarters): got %v, want %v (FR-TGT-001)", got, want)
	}
}

func TestCard_IcewindRegent_MayReturnItself(t *testing.T) {
	// Covers: FR-ACT-001, FR-ZON-004
	s, r := regentScene(t)
	if !offersActivate(s.g, r, 0) {
		t.Fatalf("with no other unit on the battlefield the Regent is still offered, it may choose itself (FR-TGT-001)")
	}
	regentChoose(s, r)
	s.do(activate(r, 0))
	if s.alive(r) {
		t.Errorf("the Regent chose itself and leaves the battlefield (FR-ZON-004)")
	}
	if !inHand(s.g.Players[0], r.Inst) {
		t.Errorf("the Regent's card is in its owner's hand (FR-ZON-004)")
	}
	if inDiscardPile(s.g.Players[0], r.Inst) {
		t.Errorf("the Regent's card is not in the discard pile (FR-ZON-004)")
	}
}

func TestCard_IcewindRegent_ReturnsControlledUnitToItsOwner(t *testing.T) {
	// Covers: FR-ZON-003, FR-ZON-004
	s, r := regentScene(t)
	stolen := s.put(1, ZoneMainTheater, unit("Stolen", 2, 5, ability.Keywords{}))
	stolen.Controller = 0
	regentChoose(s, stolen)
	s.do(activate(r, 0))
	if s.alive(stolen) {
		t.Fatalf("the controlled unit leaves the battlefield (FR-ZON-004)")
	}
	if !inHand(s.g.Players[1], stolen.Inst) {
		t.Errorf("the card returns to its owner's (player 1's) hand (FR-ZON-003)")
	}
	if inHand(s.g.Players[0], stolen.Inst) {
		t.Errorf("the card does not go to the controller's hand (FR-ZON-003)")
	}
}

func TestCard_IcewindRegent_ReturnsScratchCompanyToOwnersHand(t *testing.T) {
	// Covers: FR-ZON-005
	s, r := regentScene(t)
	sc := s.putScratch(1)
	regentChoose(s, sc)
	s.do(activate(r, 0))
	if s.alive(sc) {
		t.Fatalf("the Scratch Company leaves the battlefield (FR-ZON-005)")
	}
	if !inHand(s.g.Players[1], sc.Inst) {
		t.Errorf("the Scratch Company's card is in its owner's hand (FR-ZON-005)")
	}
	if inDiscardPile(s.g.Players[1], sc.Inst) {
		t.Errorf("the card is not in the discard pile (FR-ZON-005)")
	}
}

func TestCard_IcewindRegent_WardedOpposingUnitNotOffered(t *testing.T) {
	// Covers: FR-KWD-002h
	s, r := regentScene(t)
	foeWard := s.put(1, ZoneMainTheater, unit("FoeWard", 2, 5, ability.Keywords{Ward: true}))
	allyWard := s.put(0, ZoneMainTheater, unit("AllyWard", 2, 5, ability.Keywords{Ward: true}))
	regentChoose(s, allyWard)
	s.do(activate(r, 0))
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecideTarget {
		t.Fatalf("one target question is asked, got %d decisions (FR-TGT-001)", len(asks))
	}
	got := refs(asks[0])
	for _, id := range got {
		if id == foeWard.ID {
			t.Errorf("an opposing unit with Ward is not an option, got %v (FR-KWD-002h)", got)
		}
	}
	want := sortedCopy([]int{r.ID, allyWard.ID})
	if !reflect.DeepEqual(got, want) {
		t.Errorf("options are the Regent and the allied Ward unit: got %v, want %v (FR-KWD-002h)", got, want)
	}
}

func TestCard_IcewindRegent_NotOfferedWhileExhausted(t *testing.T) {
	// Covers: FR-ACT-002
	s, r := regentScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 2, 5, ability.Keywords{}))
	r.Exhausted = true
	if offersActivate(s.g, r, 0) {
		t.Errorf("an Exhausted Regent cannot pay its exhaust cost, so Activate is not offered (FR-ACT-002)")
	}
	s.reject(activate(r, 0))
	p1 := s.g.Players[1]
	if !s.alive(r) || !s.alive(foe) || inHand(p1, foe.Inst) || len(p1.Discard) != 0 {
		t.Errorf("a rejected Activate moves nothing: Regent alive %v, Foe alive %v (FR-ACT-002)", s.alive(r), s.alive(foe))
	}
}
