package engine

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0252: OpModifyStat on Maximum Health. Step.Stat names the stat a
// modify step changes (zero value: power); ability.StatMaxHealth makes a Layer 5
// Maximum Health modifier for Step.Until (FR-LAY-002 Temporary Maximum Health
// Modifiers, FR-LAY-003). Cards are synthetic. Player 0 is the active player.

// mhStep is "chosen unit gets n Maximum Health" lasting until.
func mhStep(n int, until ability.Duration) ability.Step {
	s := powerFor(n, until)
	s.Stat = ability.StatMaxHealth
	return s
}

func mhWant(t *testing.T, s *scene, a *Asset, maxH, cur, dmg int, when string) {
	t.Helper()
	if got := s.g.MaxHealth(a); got != maxH {
		t.Errorf("%s: Maximum Health %d, want %d (FR-LAY-002)", when, got, maxH)
	}
	if got := s.g.CurrentHealth(a); got != cur {
		t.Errorf("%s: Current Health %d, want %d (FR-LAY-002)", when, got, cur)
	}
	if a.Damage != dmg {
		t.Errorf("%s: Damage %d, want %d (a Maximum Health change adds no damage counters, FR-LAY-002)", when, a.Damage, dmg)
	}
}

func mhEventsNaming(s *scene, from int, kind string, a *Asset) int {
	n := 0
	for _, e := range effectEvents(s.g, from, kind) {
		if e.Target == a.ID {
			n++
		}
	}
	return n
}

func TestFR_LAY_002_ModifyStepLowersMaxHealth(t *testing.T) {
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 2, 5, ability.Keywords{}))
	from := len(s.g.Log)
	enactOn(s, ability.AllegianceOpposing, mhStep(-3, ability.UntilEndOfTurn))
	mhWant(t, s, foe, 2, 2, 0, "after -3 Maximum Health")
	wantPower(t, s, foe, 2, "a Maximum Health change leaves power alone")
	if n := mhEventsNaming(s, from, EventPower, foe) + mhEventsNaming(s, from, EventDamage, foe); n != 0 {
		t.Errorf("%d power or damage events name the unit, want none (a Maximum Health change is not power or damage)", n)
	}
}

func TestFR_LAY_002_ModifyStepRaisesMaxHealth(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 3, ability.Keywords{}))
	ally.Damage = 2
	enactOn(s, ability.AllegianceAllied, mhStep(2, ability.UntilEndOfTurn))
	mhWant(t, s, ally, 5, 3, 2, "after +2 Maximum Health on a unit with 2 damage")
}

func TestFR_LAY_002_ModifyStepToZeroMaxHealthDestroysAtTheSweep(t *testing.T) {
	// Covers: FR-LAY-002 Current Health & Lethal Destruction
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 3, ability.Keywords{}))
	from := len(s.g.Log)
	enactOn(s, ability.AllegianceOpposing, mhStep(-3, ability.UntilEndOfTurn))
	if s.alive(foe) {
		t.Errorf("a unit whose Maximum Health is 0 is destroyed at the sweep after the Maneuver (FR-LAY-002)")
	}
	if p := s.g.Players[1]; !slices.Contains(p.Discard, foe.Inst) {
		t.Errorf("the destroyed unit's card is in its owner's discard pile (FR-LAY-002)")
	}
	if n := mhEventsNaming(s, from, EventDamage, foe); n != 0 {
		t.Errorf("%d damage events name the unit, want none: Maximum Health 0 is lethal with no damage (FR-LAY-002)", n)
	}

	s2 := newScene(t)
	tough := s2.put(1, ZoneMainTheater, unit("Tough", 1, 4, ability.Keywords{}))
	enactOn(s2, ability.AllegianceOpposing, mhStep(-3, ability.UntilEndOfTurn))
	if !s2.alive(tough) {
		t.Fatalf("a Health 4 unit given -3 Maximum Health survives (Current Health 1, FR-LAY-002)")
	}
	mhWant(t, s2, tough, 1, 1, 0, "Health 4 unit after -3")
}

func TestFR_LAY_002_ModifyStepMaxHealthExpiresAndRestores(t *testing.T) {
	// Covers: FR-LAY-003
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, ability.Keywords{}))
	ally.Damage = 2
	enactOn(s, ability.AllegianceAllied, mhStep(-2, ability.UntilEndOfTurn))
	if !s.alive(ally) {
		t.Fatalf("Health 5 with 2 damage given -2 Maximum Health has Current Health 1 and lives (FR-LAY-002)")
	}
	mhWant(t, s, ally, 3, 1, 2, "after -2")
	s.endTurn()
	if !s.alive(ally) {
		t.Fatalf("the unit survives the end of the turn when the modifier expires (FR-LAY-003)")
	}
	mhWant(t, s, ally, 5, 3, 2, "after the turn ended: the modifier expired and the damage stays")
}

func TestFR_LAY_003_ModifyStepMaxHealthUntilEndOfOpponentTurn(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 3, ability.Keywords{}))
	enactOn(s, ability.AllegianceAllied, mhStep(2, ability.UntilEndOfOpponentTurn))
	mhWant(t, s, ally, 5, 5, 0, "right after resolution")
	s.endTurn()
	if s.g.Active != 1 {
		t.Fatalf("precondition: player 1's turn, got active %d", s.g.Active)
	}
	mhWant(t, s, ally, 5, 5, 0, "through the opponent's turn start (FR-LAY-003)")
	s.endTurn()
	mhWant(t, s, ally, 3, 3, 0, "after the opponent's turn ended (FR-LAY-003)")

	s2 := newScene(t)
	a2 := s2.put(0, ZoneMainTheater, unit("Ally", 1, 3, ability.Keywords{}))
	enactOn(s2, ability.AllegianceAllied, mhStep(2, ability.UntilEndOfTurn))
	mhWant(t, s2, a2, 5, 5, 0, "zero Until, right after resolution")
	s2.endTurn()
	mhWant(t, s2, a2, 3, 3, 0, "zero Until means until end of turn (FR-LAY-003)")
}

func TestEffectEvents_MaxHealthLogsTheChange(t *testing.T) {
	// Covers: FR-LAY-002
	for _, n := range []int{2, -1} {
		t.Run(fmt.Sprintf("%+d", n), func(t *testing.T) {
			step := onAlliedKinds(ability.OpModifyStat, n, ability.KindStructure)
			step.Stat = ability.StatMaxHealth
			s := newScene(t)
			src := s.effectSource(step)
			a, b := s.wall(), s.wall()
			from := len(s.g.Log)
			s.do(activate(src, 0))
			wantAssetEvents(t, effectEvents(s.g, from, EventMaxHealth), []*Asset{a, b}, []int{n, n}, "%s gets %+d max health until end of turn")
			if got := len(effectEvents(s.g, from, EventPower)); got != 0 {
				t.Errorf("%d power events, want none for a Maximum Health step", got)
			}
			for _, w := range []*Asset{a, b} {
				if s.g.MaxHealth(w) != 9+n || s.g.Power(w) != 0 {
					t.Errorf("%s: Maximum Health %d, power %d, want %d and 0", w, s.g.MaxHealth(w), s.g.Power(w), 9+n)
				}
			}
		})
	}
}

func TestEffectEvents_MaxHealthFaceDownAssetStaysHidden(t *testing.T) {
	step := onAlliedKinds(ability.OpModifyStat, 1, ability.KindUnit)
	step.Stat = ability.StatMaxHealth
	s := newScene(t)
	src := s.effectSource(step)
	fd := s.putScratch(0)
	from := len(s.g.Log)
	s.do(activate(src, 0))
	found := false
	for _, e := range effectEvents(s.g, from, EventMaxHealth) {
		if e.Target != fd.ID {
			continue
		}
		found = true
		if strings.Contains(e.Detail, "Filler") || e.Card != "" || !strings.HasPrefix(e.Detail, fd.String()) {
			t.Errorf("the event names the hidden card: card %q, detail %q", e.Card, e.Detail)
		}
	}
	if !found {
		t.Fatalf("no max-health event for the face-down asset")
	}
}

func TestFR_LAY_001_ModifyStepStatPowerIsPower(t *testing.T) {
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 4, ability.Keywords{}))
	step := powerFor(2, ability.UntilEndOfTurn)
	step.Stat = ability.StatPower
	from := len(s.g.Log)
	enactOn(s, ability.AllegianceOpposing, step)
	wantPower(t, s, foe, 3, "StatPower is Combat Power")
	if foe.TempPower() != 2 {
		t.Errorf("TempPower %d, want 2", foe.TempPower())
	}
	if got := len(effectEvents(s.g, from, EventPower)); got != 1 {
		t.Errorf("%d power events, want 1", got)
	}
	if got := len(effectEvents(s.g, from, EventMaxHealth)); got != 0 {
		t.Errorf("%d max-health events, want none for a power step", got)
	}
	mhWant(t, s, foe, 4, 4, 0, "power step leaves Maximum Health at printed")
}

// mhMax3 is "with Maximum Health 3 or less".
func mhMax3(f ability.Filter) ability.Filter {
	f.Stat = ability.StatMaxHealth
	f.AtMost = 3
	return f
}

func TestFR_TGT_001_MaxHealthAtMostReadsEffectiveMaxHealthOnTheBattlefield(t *testing.T) {
	// Covers: FR-LAY-002
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("Plain", 1, 5, ability.Keywords{}))
	lowered := s.put(1, ZoneMainTheater, unit("Lowered", 1, 5, ability.Keywords{}))
	lowered.Modifiers = []Modifier{{Layer: 5, MaxHealth: -2}} // effective 3
	hurt3 := s.put(1, ZoneMainTheater, unit("Hurt3", 1, 3, ability.Keywords{}))
	hurt3.Damage = 2 // Maximum Health 3, Current Health 1
	hurt4 := s.put(1, ZoneMainTheater, unit("Hurt4", 1, 4, ability.Keywords{}))
	hurt4.Damage = 3 // Maximum Health 4, Current Health 1
	slot := ability.Target{Filter: mhMax3(ability.Filter{Kinds: ability.KindUnit, Other: true})}
	s.script(0, Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, []ability.Target{slot}, armorStep(1, 0)))
	d := paTargetAsk(t, s)
	want := ids(lowered, hurt3)
	if got := refs(d); !slices.Equal(paSorted(got), paSorted(want)) {
		t.Errorf("'Maximum Health 3 or less' offers the unit lowered to 3 and the Health 3 unit with damage, not Health 5 nor Maximum Health 4 with Current Health 1: got %v, want %v (FR-TGT-001 Dynamic Stats)", got, want)
	}
}

func TestFR_TGT_001_MaxHealthAtMostOffersPileCardsByPrintedHealth(t *testing.T) {
	// Covers: FR-LAY-001, FR-ZON-004
	s := newScene(t)
	mine := s.inDiscard(0, unit("H2", 1, 2, ability.Keywords{}), unit("H3", 1, 3, ability.Keywords{}), unit("H4", 1, 4, ability.Keywords{}))
	pile := ability.Target{Filter: mhMax3(ppAllyUnitCards)}
	s.script(0, Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, []ability.Target{pile}, ptStep(ability.OpReturnToHand, 0)))
	d := paTargetAsk(t, s)
	want := ppInstIDs(mine[0], mine[1])
	if got := ppRefs(d); !slices.Equal(paSorted(got), paSorted(want)) {
		t.Errorf("'Maximum Health 3 or less' offers the discard-pile units of printed Health 2 and 3 (inclusive bound) and not Health 4: got %v, want %v (FR-LAY-001, FR-TGT-001)", got, want)
	}
}
