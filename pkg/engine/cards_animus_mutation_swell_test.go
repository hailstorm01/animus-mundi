package engine

import (
	"slices"
	"sort"
	"testing"
)

// Card tests for Mutation Swell (T-0266), written from the CSV text
// (data/cards/maneuvers-animus.csv:12): Animus Maneuver, R0 T3, "As an
// additional cost to enact this maneuver, kill 2 or 4 allied Scratch Company
// units. For every 2 units killed this way, deploy a different chosen unit card
// from your discard pile for free." Player 0 is the active player and enacts.
// Pile cards are synthetic: pkCard prints only Reserves (no zone asked) and
// dfUnit sets printed zones. The Scratch Companies' own hidden cards are unit
// cards (scene.putScratch). The count and target helpers are those of
// per_kill_target_test.go.

// msScene gives player 0 the real Mutation Swell with exactly R0 T3, n allied
// Scratch Companies in Reserves and the given unit cards in the discard pile
// (the last on top).
func msScene(t *testing.T, n int, cards ...*Card) (*scene, *Instance, []*Asset, []*Instance) {
	t.Helper()
	s := newScene(t)
	pile := s.inDiscard(0, cards...)
	var scratch []*Asset
	for i := 0; i < n; i++ {
		scratch = append(scratch, s.putScratch(0))
	}
	inst := s.give(0, s.card("Mutation Swell"))
	s.setPool(0, 0, 3)
	return s, inst, scratch, pile
}

func msSortedRefs(d Decision) []int {
	r := kcRefs(d)
	sort.Ints(r)
	return r
}

func TestCard_MutationSwell_NotEnactableWithoutItsCostAndTargets(t *testing.T) {
	// Covers: FR-MNV-001, SVY-R29, SVY-R53
	t.Run("one allied Scratch Company", func(t *testing.T) {
		s, inst, _, _ := msScene(t, 1, pkCard("Fallen"))
		s.putScratch(1)
		if offersEnact(s.g, inst) {
			t.Errorf("one allied Scratch Company cannot pay a kill of 2 or 4, and the opponent's does not count: not offered (SVY-R53, SVY-R29)")
		}
		s.reject(enactAction(inst))
		if !inHand(s.g.Players[0], inst) {
			t.Errorf("a rejected Enact leaves Mutation Swell in hand (SVY-R29)")
		}
	})
	t.Run("no allied unit card in the pile", func(t *testing.T) {
		s, inst, _, _ := msScene(t, 2, maneuver("Spent", 0, 0, steps()))
		s.inDiscard(1, pkCard("Theirs"))
		if offersEnact(s.g, inst) {
			t.Errorf("a Maneuver in the pile and the opponent's unit card are not legal targets, so the count 2 cannot fill its target: not offered (FR-MNV-001, SVY-R29, SVY-R53)")
		}
		s.reject(enactAction(inst))
		if !inHand(s.g.Players[0], inst) {
			t.Errorf("a rejected Enact leaves Mutation Swell in hand (SVY-R29)")
		}
		s.inDiscard(0, pkCard("Fallen"))
		if !offersEnact(s.g, inst) {
			t.Errorf("with a unit card in player 0's pile the count 2 can be paid and filled, so Mutation Swell is offered (SVY-R53)")
		}
	})
}

func TestCard_MutationSwell_OnlyTwoOfferedWhenFourCantBeFilled(t *testing.T) {
	// Covers: SVY-R53, FR-MOD-001
	t.Run("three units", func(t *testing.T) {
		s, inst, scratch, pile := msScene(t, 3, pkCard("One"), pkCard("Two"))
		pkScript(s, 0, scratch[:2], pkPick(pile[1]))
		s.do(enactAction(inst))
		if got := pkAsks(s, DecideMode); len(got) != 0 {
			t.Errorf("only the count 2 can be paid, so no count is asked (FR-MOD-001 sec. 1): asked %v", got)
		}
		picks := pkAsks(s, DecidePick)
		if len(picks) != 1 {
			t.Fatalf("one kill pick is asked, got %d", len(picks))
		}
		if len(picks[0].Options) != 3 || picks[0].Min != 2 || picks[0].Max != 2 {
			t.Errorf("the kill pick offers the 3 Scratch Companies and takes exactly 2: got %d options, Min %d, Max %d (SVY-R53)", len(picks[0].Options), picks[0].Min, picks[0].Max)
		}
		if got := len(pkAsks(s, DecideTarget)); got != 1 {
			t.Errorf("one target is chosen for the count 2, got %d target decisions (SVY-R53)", got)
		}
		if s.alive(scratch[0]) || s.alive(scratch[1]) || !s.alive(scratch[2]) {
			t.Errorf("exactly the 2 picked Scratch Companies die (SVY-R53)")
		}
		if !pkDeployed(s, pile[1].Card) {
			t.Errorf("the chosen unit card is deployed (SVY-R53)")
		}
		if pkDeployed(s, pile[0].Card) || !inDiscardPile(s.g.Players[0], pile[0]) {
			t.Errorf("the other unit card stays in the pile (SVY-R53)")
		}
	})
	t.Run("one unit card", func(t *testing.T) {
		s, inst, scratch, pile := msScene(t, 4, pkCard("Fallen"))
		pkScript(s, 0, scratch[:2], nil)
		s.do(enactAction(inst))
		if got := pkAsks(s, DecideMode); len(got) != 0 {
			t.Errorf("count 4 needs 2 different cards and the pile has 1, so only count 2 is legal and is taken unasked (SVY-R53, FR-MOD-001 sec. 1): asked %v", got)
		}
		if got := pkAsks(s, DecideTarget); len(got) != 0 {
			t.Errorf("the one card is the only legal entity and is taken unasked (FR-MOD-001 sec. 1): asked %v", got)
		}
		if pkAlive(s, scratch) != 2 {
			t.Errorf("exactly 2 Scratch Companies die and 2 live (SVY-R53)")
		}
		if !pkDeployed(s, pile[0].Card) {
			t.Errorf("the one unit card is deployed (SVY-R53)")
		}
	})
}

func TestCard_MutationSwell_KillFourDeploysTwoDifferent(t *testing.T) {
	// Covers: SVY-R53, SVY-R54, FR-TRN-003a
	s, inst, scratch, pile := msScene(t, 4, dfUnit("A", ZoneMainTheater), pkCard("B"), pkCard("C"))
	pkScript(s, 4, nil, pkPick(pile[0], pile[1]))
	s.do(enactAction(inst))
	modes := pkAsks(s, DecideMode)
	if len(modes) != 1 || len(modes[0].Options) != 2 || modes[0].Options[0].Label != "kill 2" || modes[0].Options[1].Label != "kill 4" {
		t.Errorf("the count decision offers \"kill 2\" then \"kill 4\" (SVY-R53): got %v", modes)
	}
	ts := pkAsks(s, DecideTarget)
	if len(ts) != 1 {
		t.Fatalf("count 4 asks the slot once for both repetitions, got %d target decisions (SVY-R53)", len(ts))
	}
	d := ts[0]
	if d.Min != 2 || d.Max != 2 || d.Repeats {
		t.Errorf("Min/Max/Repeats = %d/%d/%v, want 2/2/false: two different picks (SVY-R53, SVY-R54)", d.Min, d.Max, d.Repeats)
	}
	want := []int{pile[0].ID, pile[1].ID, pile[2].ID}
	sort.Ints(want)
	if got := msSortedRefs(d); !slices.Equal(got, want) {
		t.Errorf("the options are exactly the three unit cards in the pile, got refs %v want %v (SVY-R53)", got, want)
	}
	if pkAlive(s, scratch) != 0 {
		t.Errorf("count 4 kills all four Scratch Companies (SVY-R53)")
	}
	aa, ba := assetOf(s.g, pile[0].Card), assetOf(s.g, pile[1].Card)
	if aa == nil || aa.Zone != ZoneMainTheater || aa.Controller != 0 {
		t.Errorf("A is deployed into player 0's Main Theater, its printed zone (FR-TRN-003a 1)")
	}
	if ba == nil || ba.Zone != ZoneReserves || ba.Controller != 0 {
		t.Errorf("B is deployed into player 0's Reserves, its printed zone (FR-TRN-003a 1)")
	}
	if pkDeployed(s, pile[2].Card) || !inDiscardPile(s.g.Players[0], pile[2]) {
		t.Errorf("C, not chosen, stays in the pile (SVY-R53)")
	}
	if !inDiscardPile(s.g.Players[0], inst) {
		t.Errorf("Mutation Swell is in player 0's discard pile after resolving (FR-TRN-003)")
	}
}

func TestCard_MutationSwell_KilledUnitsCardsAreNotTargets(t *testing.T) {
	// Covers: SVY-R2
	s, inst, scratch, pile := msScene(t, 2, pkCard("X"), pkCard("Y"))
	pkScript(s, 0, scratch, pkPick(pile[0]))
	s.do(enactAction(inst))
	ts := pkAsks(s, DecideTarget)
	if len(ts) != 1 {
		t.Fatalf("the one target is asked, got %d target decisions (SVY-R2)", len(ts))
	}
	want := []int{pile[0].ID, pile[1].ID}
	sort.Ints(want)
	if got := msSortedRefs(ts[0]); !slices.Equal(got, want) {
		t.Errorf("targets are chosen on declaration, so the options are exactly X and Y and not the killed Scratch Companies' cards: got refs %v want %v (SVY-R2)", got, want)
	}
	for _, sc := range scratch {
		if s.alive(sc) || !inDiscardPile(s.g.Players[0], sc.Inst) {
			t.Errorf("a killed Scratch Company's card is in the discard pile (SVY-R2)")
		}
	}
	if !pkDeployed(s, pile[0].Card) {
		t.Errorf("the chosen card X is deployed (SVY-R2)")
	}
}

func TestCard_MutationSwell_ReplacedKillsStillDeploy(t *testing.T) {
	// Covers: SVY-R15, SVY-R21, FR-REP-001
	s, inst, scratch, pile := msScene(t, 4, pkCard("One"), pkCard("Two"))
	s.put(0, ZoneReserves, s.card("Shifting Impossibility"))
	pkScript(s, 4, nil, pkPick(pile[0], pile[1]))
	s.do(enactAction(inst))
	for _, sc := range scratch {
		if s.alive(sc) || !inHand(s.g.Players[0], sc.Inst) || inDiscardPile(s.g.Players[0], sc.Inst) {
			t.Errorf("a replaced kill still pays the cost: the Scratch Company leaves the battlefield and is returned to its owner's hand, not the pile (SVY-R15, SVY-R21, FR-REP-001)")
		}
	}
	if !pkDeployed(s, pile[0].Card) || !pkDeployed(s, pile[1].Card) {
		t.Errorf("the cost was paid in full, so both chosen unit cards are deployed (SVY-R15, SVY-R21)")
	}
}

func TestCard_MutationSwell_GoneTargetSkippedOtherDeploys(t *testing.T) {
	// Covers: FR-TGT-003, SVY-R53
	s, inst, _, pile := msScene(t, 4, pkCard("P"), pkCard("Q"))
	pkScript(s, 4, nil, pkPick(pile[0], pile[1]))
	r := &recorder{}
	moved := false
	r.onEvent = func(e Event) {
		if e.Kind == "place" && !moved {
			moved = true
			if s.g.takeCard(pile[0]) {
				s.g.putCard(pile[0], ZoneHand, false)
			}
		}
	}
	s.g.obs = r
	s.do(enactAction(inst))
	if !moved {
		t.Fatalf("Mutation Swell was placed on the stack")
	}
	if pkDeployed(s, pile[0].Card) || !inHand(s.g.Players[0], pile[0]) {
		t.Errorf("P left the pile before resolution, so it is not deployed and stays in hand (FR-TGT-003, SVY-R53)")
	}
	if !pkDeployed(s, pile[1].Card) {
		t.Errorf("the other target Q is still deployed (SVY-R53)")
	}
}

func TestCard_MutationSwell_Timeout(t *testing.T) {
	// Covers: SVY-R54, SVY-R16, FR-MOD-001, SVY-R43
	t.Run("everything timed out", func(t *testing.T) {
		s, inst, _, pile := msScene(t, 0, pkCard("One"), pkCard("Two"), pkCard("Three"))
		r1 := s.putScratch(0)
		r2 := s.putScratch(0)
		m1 := s.putScratch(0)
		m2 := s.putScratch(0)
		s.moveTo(m1, ZoneMainTheater)
		s.moveTo(m2, ZoneMainTheater)
		pkScript(s, 0, nil, nil)
		s.do(enactAction(inst))
		if s.alive(r1) || s.alive(r2) || !s.alive(m1) || !s.alive(m2) {
			t.Errorf("a timed-out count is 2 and the kills take the first allied Scratch Companies in fixed order, Reserves first (SVY-R54, SVY-R16)")
		}
		if !pkDeployed(s, pile[2].Card) || pkDeployed(s, pile[1].Card) || pkDeployed(s, pile[0].Card) {
			t.Errorf("a timed-out target takes the top card of the pile only (SVY-R54, FR-MOD-001 sec. 6)")
		}
	})
	t.Run("count 4 answered, target timed out", func(t *testing.T) {
		s, inst, _, pile := msScene(t, 4, pkCard("One"), pkCard("Two"), dfUnit("Three", ZoneMainTheater, ZoneReserves))
		pkScript(s, 4, nil, nil)
		s.do(enactAction(inst))
		if !pkDeployed(s, pile[2].Card) || !pkDeployed(s, pile[1].Card) || pkDeployed(s, pile[0].Card) {
			t.Errorf("a timed-out target of 2 takes the top two different cards, the third stays (SVY-R54, FR-MOD-001 sec. 6)")
		}
		if zs := pkAsks(s, DecideZone); len(zs) != 1 {
			t.Errorf("the both-zones card asks one DecideZone, got %d (SVY-R43)", len(zs))
		}
		if a := assetOf(s.g, pile[2].Card); a == nil || a.Zone != ZoneReserves {
			t.Errorf("a timed-out zone choice enters Reserves (SVY-R43)")
		}
	})
}
