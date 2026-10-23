package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for Trigger.MaxPerTurn (T-0054): a printed trigger limited to a fixed
// number of firings each turn, as on Animoceps Rager ("This ability triggers
// only once each turn"). FR-TRN-001 resets once-per-turn limits at the start of
// every turn, either player's. The limit counts firings (a trigger entering the
// pending buffer), not resolutions, and is kept per asset. Cards are synthetic:
// a damage-taken (HookDamaged) trigger that Scratches 1, observed with
// wantScratches. Damage-taken is FR-REP-004 Stage 5.
//
// This file does not compile until ability.Trigger has a MaxPerTurn int field;
// that is expected until T-0054 is implemented.

// limited is trig with MaxPerTurn n (0 means no limit).
func limited(n int, trig ability.Trigger) ability.Trigger {
	trig.MaxPerTurn = n
	return trig
}

// hitters puts n power-1 player-0 attackers on the battlefield.
func (s *scene) hitters(n int) []*Asset {
	out := make([]*Asset, n)
	for i := range out {
		out[i] = s.put(0, ZoneMainTheater, unit("Hitter", 1, 9, Keywords{}))
	}
	return out
}

// --- the limit within one turn ---

// Covers: FR-REP-004
func TestFR_TRN_001_LimitedTriggerFiresOnceForTwoDamageInstancesInOneTurn(t *testing.T) {
	s := newScene(t)
	h := s.hitters(2)
	def := s.put(1, ZoneMainTheater, bowed("Rager", 0, 9, Keywords{}, limited(1, unbowed(nil, scratchN(1)))))
	s.do(s.attack(h[0], def))
	s.wantScratches(1, 1, "the first damage instance fires the once-each-turn trigger (FR-REP-004 Stage 5)")
	s.do(s.attack(h[1], def))
	s.wantDamage(def, 2)
	s.wantScratches(1, 1, "a second damage instance in the same turn does not fire a trigger limited to once each turn (FR-TRN-001)")
}

func TestFR_TRN_001_UnlimitedTriggerStillFiresEveryTime(t *testing.T) {
	s := newScene(t)
	h := s.hitters(2)
	def := s.put(1, ZoneMainTheater, bowed("Bowed", 0, 9, Keywords{}, limited(0, unbowed(nil, scratchN(1)))))
	s.do(s.attack(h[0], def))
	s.do(s.attack(h[1], def))
	s.wantScratches(1, 2, "MaxPerTurn 0 means no limit: each damage instance fires the trigger")
}

func TestFR_TRN_001_LimitOfTwoAllowsTwoFiringsAndNotAThird(t *testing.T) {
	s := newScene(t)
	h := s.hitters(3)
	def := s.put(1, ZoneMainTheater, bowed("Twice", 0, 9, Keywords{}, limited(2, unbowed(nil, scratchN(1)))))
	s.do(s.attack(h[0], def))
	s.do(s.attack(h[1], def))
	s.wantScratches(1, 2, "MaxPerTurn 2 allows two firings in a turn")
	s.do(s.attack(h[2], def))
	s.wantDamage(def, 3)
	s.wantScratches(1, 2, "MaxPerTurn 2 does not allow a third firing in the same turn (FR-TRN-001)")
}

func TestFR_TRN_001_EachAssetHasItsOwnAllowance(t *testing.T) {
	s := newScene(t)
	h := s.hitters(4)
	trig := limited(1, unbowed(nil, scratchN(1)))
	a := s.put(1, ZoneMainTheater, bowed("Rager A", 0, 9, Keywords{}, trig))
	b := s.put(1, ZoneMainTheater, bowed("Rager B", 0, 9, Keywords{}, trig))
	s.do(s.attack(h[0], a))
	s.do(s.attack(h[1], a))
	s.wantScratches(1, 1, "asset A fired once and is spent for the turn")
	s.do(s.attack(h[2], b))
	s.wantScratches(1, 2, "asset B has its own allowance, unaffected by asset A's firing")
	s.do(s.attack(h[3], b))
	s.wantScratches(1, 2, "asset B is also limited to one firing this turn")
}

// --- the limit resets every turn (FR-TRN-001) ---

func TestFR_TRN_001_LimitResetsOnTheControllersNextTurn(t *testing.T) {
	s := newScene(t)
	h := s.hitters(2)
	def := s.put(1, ZoneMainTheater, bowed("Rager", 0, 9, Keywords{}, limited(1, unbowed(nil, scratchN(1)))))
	s.do(s.attack(h[0], def))
	s.wantScratches(1, 1, "fires on turn 1")
	s.endTurn() // player 1's turn
	s.endTurn() // player 0's turn again
	s.do(s.attack(h[1], def))
	s.wantScratches(1, 2, "once-per-turn limits reset at the start of each turn, so it fires again (FR-TRN-001)")
}

func TestFR_TRN_001_LimitResetsOnTheOpponentsNextTurn(t *testing.T) {
	// Player 0's Rager takes retaliation on player 0's turn, then takes
	// combat damage from player 1's attack on player 1's turn.
	s := newScene(t)
	rager := s.put(0, ZoneMainTheater, bowed("Rager", 1, 9, Keywords{}, limited(1, unbowed(nil, scratchN(1)))))
	wall := s.put(1, ZoneMainTheater, unit("Wall", 2, 9, Keywords{}))
	s.do(s.attack(rager, wall)) // retaliation damages the Rager
	s.wantScratches(0, 1, "fires on player 0's own turn from retaliation")
	s.endTurn()
	s.do(s.attack(wall, rager))
	s.wantScratches(0, 2, "the limit resets for player 1's turn, so damage on the opponent's turn fires it again (FR-TRN-001, every turn not only the controller's)")
}

func TestFR_TRN_001_OneFiringOnTheOpponentsTurnAndOneOnTheControllersNext(t *testing.T) {
	s := newScene(t)
	h := s.hitters(1)
	def := s.put(1, ZoneMainTheater, bowed("Rager", 2, 9, Keywords{}, limited(1, unbowed(nil, scratchN(1)))))
	foe := s.put(0, ZoneMainTheater, unit("Foe", 1, 9, Keywords{}))
	s.do(s.attack(h[0], def)) // opponent's turn (player 0's)
	s.wantScratches(1, 1, "one firing on the opponent's turn")
	s.endTurn()
	s.do(s.attack(def, foe)) // controller's turn: retaliation damages the Rager
	s.wantScratches(1, 2, "one more firing on the controller's own turn (FR-TRN-001)")
}

// --- the limit counts firings, not resolutions ---

// Covers: FR-TGT-003
func TestFR_TRN_001_AFizzledFiringStillUsesTheTurnsAllowance(t *testing.T) {
	// The limited trigger deals 1 damage to a chosen opposing unit. On the
	// first firing the chosen unit leaves the battlefield before the trigger
	// resolves, so it fizzles. A second damage instance in the same turn must
	// not fire it again: the unit left to choose, Spare, takes no damage.
	s := newScene(t)
	h := s.hitters(2)
	fragile := s.put(0, ZoneMainTheater, unit("Fragile", 1, 1, Keywords{}))
	spare := s.put(0, ZoneMainTheater, unit("Spare", 1, 9, Keywords{}))
	trig := limited(1, unbowed([]ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, true)}, dmgTo(1, 0)))
	def := s.put(1, ZoneMainTheater, bowed("Rager", 0, 9, Keywords{}, trig))
	first := true
	s.g.choosers[1] = func(_ *Game, d Decision) Answer {
		if d.Kind == DecideOrder {
			return Answer{}
		}
		want := spare
		if first {
			want = fragile
			first = false
		}
		for i, o := range d.Options {
			if o.Ref == want.ID {
				if want == fragile {
					s.g.removeFromBattlefield(fragile) // gone before resolution
				}
				return Answer{Picks: []int{i}}
			}
		}
		t.Errorf("wanted asset %d is not among the options of %q", want.ID, d.Prompt)
		return Answer{Timeout: true}
	}
	s.do(s.attack(h[0], def))
	if first {
		t.Fatalf("precondition: the first firing asked for a target")
	}
	s.do(s.attack(h[1], def))
	s.wantDamage(spare, 0)
	s.wantDamage(def, 2)
}

// --- Headquarters ---

// Covers: FR-REP-004
func TestFR_TRN_001_HeadquartersLimitedTriggerFiresOnceWhenAttackedTwice(t *testing.T) {
	s := newScene(t)
	h := s.hitters(2)
	hq1 := s.g.Players[1].HQ
	hq1.Inst.Card.Ability = &ability.Def{Triggers: []ability.Trigger{limited(1, unbowed(nil, scratchN(1)))}}
	s.do(s.attack(h[0], hq1))
	s.wantScratches(1, 1, "a Headquarters dealt damage fires its damage-taken trigger (FR-REP-004 Stage 5)")
	s.do(s.attack(h[1], hq1))
	s.wantDamage(hq1, 2)
	s.wantScratches(1, 1, "the second attack on the Headquarters in the same turn does not fire the once-each-turn trigger (FR-TRN-001)")
}

// Covers: FR-REP-004
func TestFR_TRN_001_HeadquartersFiresAgainFromItsOwnAttacksRetaliationNextTurn(t *testing.T) {
	s := newScene(t)
	h := s.hitters(1)
	foe := s.put(0, ZoneMainTheater, unit("Foe", 3, 9, Keywords{}))
	hq1 := s.g.Players[1].HQ
	hq1.Inst.Card.Ability = &ability.Def{Triggers: []ability.Trigger{limited(1, unbowed(nil, scratchN(1)))}}
	s.do(s.attack(h[0], hq1))
	s.wantScratches(1, 1, "fires when attacked on turn 1")
	s.endTurn()
	s.do(s.attack(hq1, foe)) // the Headquarters attacks; retaliation damages it
	s.wantScratches(1, 2, "the limit resets for the new turn, so retaliation damage to the attacking Headquarters fires it again (FR-TRN-001)")
}
