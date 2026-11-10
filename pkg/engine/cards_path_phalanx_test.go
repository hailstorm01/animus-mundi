package engine

import "testing"

// Card tests for Chainbreaker Phalanx (T-0204), written from the CSV rules text
// in data/cards/units-and-structures-path.csv: Path Unit, Infantry,
// Chainbreaker, R4 T0, deploys to Reserves, 4/5, "Whenever this unit moves to
// the Main Theater, you may move chosen allied unit from your Reserves to the
// Main Theater. That unit gains 2 armor."
// Rules: FR-ZON-002 and SVY-R1 (what counts as moving: a deploy does not),
// FR-TGT-001 and FR-TGT-003 (legal targets; SVY-R25, SVY-R26: the armor step
// still applies to the unit the move step just moved), FR-MOD-001 section 7
// and SVY-R9 ("you may" is asked as the trigger is placed, a decline or
// timeout means it is not placed), FR-TRG-003 and FR-TRG-004 (a moved unit
// fires its own trigger, placed after the first one resolves).

// phalanxChooser installs a chooser for seat 0 that answers "you may" with opt
// and a target question with the first option whose Ref is one of prefer, in
// the order given (a timeout if none is offered). It records every Decision.
func (s *scene) phalanxChooser(opt Answer, prefer ...*Asset) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		switch d.Kind {
		case DecideOptional:
			return opt
		case DecideTarget:
			for _, want := range prefer {
				for i, o := range d.Options {
					if o.Ref == want.ID {
						return Answer{Picks: []int{i}}
					}
				}
			}
			return Answer{Timeout: true}
		}
		return Answer{Timeout: true}
	}
}

// countKind counts the decisions of kind k.
func countKind(ds []Decision, k DecisionKind) int {
	n := 0
	for _, d := range ds {
		if d.Kind == k {
			n++
		}
	}
	return n
}

// phalanxScene has the real Phalanx and a plain allied unit X in player 0's
// Reserves.
func phalanxScene(t *testing.T) (s *scene, p, x *Asset) {
	t.Helper()
	s = newScene(t)
	p = s.put(0, ZoneReserves, s.card("Chainbreaker Phalanx"))
	x = s.put(0, ZoneReserves, unit("X", 2, 4, Keywords{}))
	return s, p, x
}

// Covers: FR-ZON-002, FR-TGT-003, SVY-R25
func TestCard_ChainbreakerPhalanx_MarchMovesChosenUnitAndArmorsIt(t *testing.T) {
	s, p, x := phalanxScene(t)
	s.phalanxChooser(yes, x)
	s.do(Action{Kind: ActMarch, Player: 0, Asset: p.ID})
	if got := s.g.Asset(p.ID); got == nil || got.Zone != ZoneMainTheater {
		t.Fatalf("precondition: the March moves the Phalanx to the Main Theater, got %v", got)
	}
	if got := s.g.Asset(x.ID); got == nil || got.Zone != ZoneMainTheater {
		t.Errorf("the chosen allied unit moves from Reserves to the Main Theater (FR-ZON-002): %v", got)
	}
	if x.Armor != 2 {
		t.Errorf("the moved unit gains 2 armor although it no longer matches the slot's zone (SVY-R25, SVY-R26): armor %d, want 2", x.Armor)
	}
	if p.Armor != 0 {
		t.Errorf("the Phalanx gains no armor, only the chosen unit does: armor %d, want 0", p.Armor)
	}
	if n := countKind(s.asked(0), DecideOptional); n != 1 {
		t.Errorf("'you may' is asked once, as the trigger is placed (SVY-R9): asked %d times", n)
	}
	s.wantSettled("a Phalanx March that moves a unit")
}

// Covers: SVY-R9, FR-MOD-001
func TestCard_ChainbreakerPhalanx_DeclineIsNotPlaced(t *testing.T) {
	cases := []struct {
		name   string
		answer Answer
	}{
		{"decline", no},
		{"timeout", Answer{Timeout: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, p, x := phalanxScene(t)
			s.phalanxChooser(tc.answer, x)
			items := recordPlaced(s)
			s.do(Action{Kind: ActMarch, Player: 0, Asset: p.ID})
			if got := s.g.Asset(p.ID); got == nil || got.Zone != ZoneMainTheater {
				t.Fatalf("precondition: the March moves the Phalanx, got %v", got)
			}
			if n := countKind(s.asked(0), DecideOptional); n != 1 {
				t.Errorf("'you may' is asked once: asked %d times (SVY-R9)", n)
			}
			if n := countKind(s.asked(0), DecideTarget); n != 0 {
				t.Errorf("a declined or timed-out 'you may' is not placed, so no target is asked for: asked %d times (SVY-R9, FR-MOD-001)", n)
			}
			for _, it := range *items {
				if it.Kind == StackAbility && it.Source == p.ID {
					t.Errorf("a declined or timed-out trigger is not placed: placed %+v (SVY-R9)", it)
				}
			}
			if got := s.g.Asset(x.ID); got == nil || got.Zone != ZoneReserves {
				t.Errorf("X stays in Reserves: %v", got)
			}
			if x.Armor != 0 {
				t.Errorf("X gains no armor: %d", x.Armor)
			}
			s.wantSettled("a declined Phalanx trigger")
		})
	}
}

// Covers: FR-ZON-002
func TestCard_ChainbreakerPhalanx_MarchToReservesFiresNothing(t *testing.T) {
	s := newScene(t)
	p := s.put(0, ZoneMainTheater, s.card("Chainbreaker Phalanx"))
	x := s.put(0, ZoneReserves, unit("X", 2, 4, Keywords{}))
	s.phalanxChooser(yes, x)
	s.do(Action{Kind: ActMarch, Player: 0, Asset: p.ID})
	if got := s.g.Asset(p.ID); got == nil || got.Zone != ZoneReserves {
		t.Fatalf("precondition: the March moves the Phalanx to Reserves, got %v", got)
	}
	if len(s.asked(0)) != 0 {
		t.Errorf("moving to Reserves is not moving to the Main Theater, so nothing is asked: %v (FR-ZON-002)", kindsOf(s.asked(0)))
	}
	if got := s.g.Asset(x.ID); got == nil || got.Zone != ZoneReserves || x.Armor != 0 {
		t.Errorf("X does not move and gains no armor: %v armor %d", got, x.Armor)
	}
	s.wantSettled("a Phalanx March to Reserves")
}

// Covers: SVY-R1
func TestCard_ChainbreakerPhalanx_DeployFiresNothing(t *testing.T) {
	s := newScene(t)
	x := s.put(0, ZoneReserves, unit("X", 2, 4, Keywords{}))
	s.g.Players[0].Req, s.g.Players[0].Tac = 4, 0
	s.phalanxChooser(yes, x)
	s.deployFromHand(0, s.card("Chainbreaker Phalanx"))
	if len(s.asked(0)) != 0 {
		t.Errorf("deploying is entering the battlefield, not moving, so nothing is asked: %v (SVY-R1)", kindsOf(s.asked(0)))
	}
	if got := s.g.Asset(x.ID); got == nil || got.Zone != ZoneReserves || x.Armor != 0 {
		t.Errorf("X does not move and gains no armor: %v armor %d", got, x.Armor)
	}
	s.wantSettled("a Phalanx deploy")
}

// Covers: FR-TGT-001, FR-TGT-003
func TestCard_ChainbreakerPhalanx_OnlyAlliedUnitsInReserves(t *testing.T) {
	s := newScene(t)
	p := s.put(0, ZoneReserves, s.card("Chainbreaker Phalanx"))
	enemy := s.put(1, ZoneReserves, unit("Enemy", 2, 4, Keywords{}))
	front := s.put(0, ZoneMainTheater, unit("Front", 2, 4, Keywords{}))
	s.phalanxChooser(yes, enemy, front)
	s.do(Action{Kind: ActMarch, Player: 0, Asset: p.ID})
	for _, d := range s.asked(0) {
		if d.Kind != DecideTarget {
			continue
		}
		for _, o := range d.Options {
			if o.Ref == enemy.ID || o.Ref == front.ID {
				t.Errorf("only allied units in your Reserves are legal targets: option %+v offered (FR-TGT-001)", o)
			}
		}
	}
	if got := s.g.Asset(enemy.ID); got == nil || got.Zone != ZoneReserves || enemy.Armor != 0 {
		t.Errorf("the opposing unit does not move and gains no armor: %v armor %d", got, enemy.Armor)
	}
	if got := s.g.Asset(front.ID); got == nil || got.Zone != ZoneMainTheater || front.Armor != 0 {
		t.Errorf("the allied Main Theater unit gains no armor: %v armor %d", got, front.Armor)
	}
	if p.Armor != 0 {
		t.Errorf("the Phalanx gains no armor: %d", p.Armor)
	}
	s.wantSettled("a Phalanx March with no legal target")
}

// Covers: FR-TRG-004, FR-TRG-003
func TestCard_ChainbreakerPhalanx_ChainsIntoAnotherPhalanx(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneReserves, s.card("Chainbreaker Phalanx"))
	b := s.put(0, ZoneReserves, s.card("Chainbreaker Phalanx"))
	c := s.put(0, ZoneReserves, unit("C", 2, 4, Keywords{}))
	s.phalanxChooser(yes, b, c)
	// Record, as each ability is placed, its source, the stack depth and
	// whether Phalanx B is already in the Main Theater with its armor.
	type placed struct {
		src     int
		depth   int
		bArmor  int
		bInMain bool
	}
	var seen []placed
	r := &recorder{}
	r.onEvent = func(e Event) {
		if e.Kind == "place" {
			it := s.g.Stack[len(s.g.Stack)-1]
			if it.Kind == StackAbility {
				seen = append(seen, placed{it.Source, len(s.g.Stack), b.Armor, b.Zone == ZoneMainTheater})
			}
		}
	}
	s.g.obs = r
	s.do(Action{Kind: ActMarch, Player: 0, Asset: a.ID})
	if b.Zone != ZoneMainTheater || b.Armor != 2 {
		t.Fatalf("precondition: A's trigger moves Phalanx B to the Main Theater with 2 armor: zone %v armor %d", b.Zone, b.Armor)
	}
	if c.Zone != ZoneMainTheater || c.Armor != 2 {
		t.Errorf("B's own trigger fires and moves C to the Main Theater with 2 armor (FR-TRG-004): zone %v armor %d", c.Zone, c.Armor)
	}
	if a.Armor != 0 {
		t.Errorf("A gains no armor: %d", a.Armor)
	}
	if len(seen) != 2 || seen[0].src != a.ID || seen[1].src != b.ID {
		t.Fatalf("A's trigger is placed, then B's: got %+v", seen)
	}
	if seen[1].depth != 1 || !seen[1].bInMain || seen[1].bArmor != 2 {
		t.Errorf("B's trigger is placed after A's has resolved, so the stack holds only it and B already has its armor (FR-TRG-003): %+v", seen[1])
	}
	if n := countKind(s.asked(0), DecideOptional); n != 2 {
		t.Errorf("'you may' is asked once for each Phalanx trigger: %d times", n)
	}
	s.wantSettled("a chain of two Phalanxes")
}
