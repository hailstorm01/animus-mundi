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
