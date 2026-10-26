package engine

import (
	"reflect"
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for the default order of a player's simultaneous triggers (T-0066,
// board directive #57, FR-MOD-001 sec. 6, SVY-R7, FR-TRG-003). When a player
// does not order their own triggers, they are placed by where each source stood
// when it fired: the Headquarters, then Reserves, then the Main Theater, each
// zone in the order its assets entered it. The first trigger in that order is
// placed first, so it resolves last. Neither the order a step happened to visit
// assets in (oldest first for an "all" step, pick order for a multi-target
// step) nor asset age (ID) changes it.
//
// Cards are synthetic. A "sweeper" unit's deploy trigger damages allied units;
// the Unbowed units it hits each fire a trigger in one window. In every case
// the source order and the age order of the Unbowed units disagree, so a
// fallback by buffer order or by ID would fail. Placement is read from the
// "place" events (one per trigger put on the stack, in placement order), which
// is the order FR-MOD-001 sec. 6 defines.

// pingedBowed is a unit with 9 Health printing one Unbowed trigger.
func pingedBowed(name string) *Card {
	return bowed(name, 1, 9, Keywords{}, unbowed(nil, scratchN(1)))
}

// allySweeper is a Reserves unit whose deploy trigger deals 1 damage to every
// allied unit, itself included (9 Health, so it survives and has no trigger of
// its own to place).
func allySweeper() *Card {
	c := unit("Sweeper", 1, 9, Keywords{})
	c.DeployZones = []Zone{ZoneReserves}
	step := allDamage(1, ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit})
	c.Ability = &ability.Def{Triggers: []ability.Trigger{deployTrigger(step)}}
	return c
}

// placedOrder returns the IDs of the given assets in the order their triggers
// were placed on the stack, from the match log. Other triggers are ignored.
func placedOrder(s *scene, of ...*Asset) []int {
	want := map[int]bool{}
	for _, a := range of {
		want[a.ID] = true
	}
	var out []int
	for _, e := range s.g.Log {
		if e.Kind == "place" && strings.Contains(e.Detail, "places ability") && want[e.Source] {
			out = append(out, e.Source)
		}
	}
	return out
}

func wantPlaced(t *testing.T, s *scene, why string, want []*Asset, of ...*Asset) {
	t.Helper()
	ids := make([]int, len(want))
	for i, a := range want {
		ids[i] = a.ID
	}
	if got := placedOrder(s, of...); !reflect.DeepEqual(got, ids) {
		t.Errorf("%s: triggers should be placed in the order of sources %v, got %v (FR-MOD-001 sec. 6, SVY-R7, FR-TRG-003)", why, ids, got)
	}
}

// olderInTheaterNewerInReserves builds the shared disagreement: the older unit
// (lower ID) is in the Main Theater, the newer one in Reserves.
func olderInTheaterNewerInReserves(s *scene) (older, newer *Asset) {
	older = s.put(0, ZoneReserves, pingedBowed("Older Bowed"))
	newer = s.put(0, ZoneReserves, pingedBowed("Newer Bowed"))
	s.moveTo(older, ZoneMainTheater)
	return older, newer
}

func TestFR_TRG_003_AllStepTriggersFallBackToSourcePositionNotAge(t *testing.T) {
	// Covers: FR-MOD-001, SVY-R7
	// The older unit is in the Main Theater and the newer one in Reserves, so
	// age (oldest first, as the "all" step visits them) and source order
	// disagree. Reserves come first, so the newer unit's trigger is placed first.
	s := newScene(t)
	older, newer := olderInTheaterNewerInReserves(s)
	if older.ID >= newer.ID {
		t.Fatalf("setup: the Main Theater unit must be older (lower ID)")
	}
	s.deployFromHand(0, allySweeper())
	s.wantDamage(older, 1)
	s.wantDamage(newer, 1)
	wantPlaced(t, s, "an \"all\" step visits oldest first, but unordered triggers go by source position, Reserves before the Main Theater",
		[]*Asset{newer, older}, older, newer)
}

func TestFR_TRG_003_MarchingIsAnArrival(t *testing.T) {
	// Covers: FR-MOD-001, SVY-R7
	// Both units end in the Main Theater. The older one is in Reserves and
	// marches in after the newer one was placed directly in the Main Theater,
	// so it arrived there last: "each zone in the order its assets entered it".
	s := newScene(t)
	older := s.put(0, ZoneReserves, pingedBowed("Older Bowed"))
	newer := s.put(0, ZoneMainTheater, pingedBowed("Newer Bowed"))
	s.do(Action{Kind: ActMarch, Player: 0, Asset: older.ID})
	if older.Zone != ZoneMainTheater || newer.Zone != ZoneMainTheater || older.ID >= newer.ID {
		t.Fatalf("setup: both units must be in the Main Theater, the marched one older: %s (%d), %s (%d)", older.Zone, older.ID, newer.Zone, newer.ID)
	}
	s.deployFromHand(0, allySweeper())
	s.wantDamage(older, 1)
	s.wantDamage(newer, 1)
	wantPlaced(t, s, "the marched unit entered the Main Theater last, so its trigger is placed last",
		[]*Asset{newer, older}, older, newer)
}

func TestFR_TRG_003_MultiTargetTriggersFallBackToSourcePositionNotPickOrder(t *testing.T) {
	// Covers: FR-MOD-001, SVY-R7
	// One deploy trigger damages two chosen allied Unbowed units. The picks are
	// the older Main Theater unit first, then the newer Reserves unit: the
	// opposite of source order. The trigger-order question is left to time out.
	s := newScene(t)
	older, newer := olderInTheaterNewerInReserves(s)
	if older.ID >= newer.ID {
		t.Fatalf("setup: the Main Theater unit must be older (lower ID)")
	}
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		if d.Kind != DecideTarget {
			return Answer{Timeout: true}
		}
		var picks []int
		for _, want := range []int{older.ID, newer.ID} {
			for i, o := range d.Options {
				if o.Ref == want {
					picks = append(picks, i)
				}
			}
		}
		return Answer{Picks: picks}
	}
	two := ability.Target{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, NonScratch: true, Other: true}, UpTo: 2}
	c := unit("Pinger", 1, 9, Keywords{})
	c.DeployZones = []Zone{ZoneReserves}
	c.Ability = &ability.Def{Triggers: []ability.Trigger{targetTrigger([]ability.Target{two}, dmgTo(1, 0))}}
	s.deployFromHand(0, c)
	s.wantDamage(older, 1)
	s.wantDamage(newer, 1)
	wantPlaced(t, s, "the step damaged the older unit first, but unordered triggers go by source position, not pick order",
		[]*Asset{newer, older}, older, newer)
}

func TestFR_TRG_003_DeathAndUnbowedTriggersFallBackToSourcePosition(t *testing.T) {
	// Covers: FR-MOD-001, SVY-R7, SVY-R13
	// One sweep: the survivor's Unbowed is buffered when it takes damage, the
	// dying unit's death trigger after the sweep, so buffer order is survivor then
	// dying. The dying unit stood in Reserves and the survivor (older) in the
	// Main Theater, so source order is dying then survivor (Reserves before the
	// Main Theater; SVY-R13, directive #46: one window). The two units are in
	// different zones, so the sweep's index shifts are not tested here; see
	// TestFR_TRG_003_SweepRemovalsDoNotReorderTriggersInOneZone.
	s := newScene(t)
	survivor := s.put(0, ZoneMainTheater, pingedBowed("Survivor"))
	plain := s.put(0, ZoneReserves, unit("Plain", 1, 1, Keywords{}))
	dying := s.put(0, ZoneReserves, mortal("Dying", 1, 1, killedTrigger(scratchN(1))))
	if survivor.ID >= dying.ID {
		t.Fatalf("setup: the survivor must be older than the dying unit")
	}
	s.deployFromHand(0, allySweeper())
	if s.alive(plain) || s.alive(dying) || !s.alive(survivor) {
		t.Fatalf("setup: Plain and Dying die, Survivor lives (plain alive=%v, dying alive=%v, survivor alive=%v)", s.alive(plain), s.alive(dying), s.alive(survivor))
	}
	wantPlaced(t, s, "the dead unit's death trigger and the survivor's Unbowed share one window, ordered by where each source stood",
		[]*Asset{dying, survivor}, survivor, dying)
}

func TestFR_TRG_003_OrderDecisionListsTriggersInSourceOrder(t *testing.T) {
	// Covers: FR-MOD-001, SVY-R7
	// The ordering question lists the player's triggers in source order, and the
	// answer's indexes refer to that list. Picking [1, 0] places the second
	// listed trigger first.
	s := newScene(t)
	older, newer := olderInTheaterNewerInReserves(s)
	s.script(0, Answer{Picks: []int{1, 0}})
	s.deployFromHand(0, allySweeper())
	var orders []Decision
	for _, d := range s.asked(0) {
		if d.Kind == DecideOrder {
			orders = append(orders, d)
		}
	}
	if len(orders) != 1 {
		t.Fatalf("two simultaneous triggers are ordered with one question: asked %d times (FR-TRG-003)", len(orders))
	}
	d := orders[0]
	if len(d.Options) != 2 || d.Min != 2 || d.Max != 2 {
		t.Fatalf("the question has two options and takes both: %+v", d)
	}
	if !strings.Contains(d.Options[0].Label, "Newer Bowed") || !strings.Contains(d.Options[1].Label, "Older Bowed") {
		t.Errorf("options are listed in source order, Reserves before the Main Theater (newer unit first), got %q then %q (FR-MOD-001 sec. 6, SVY-R7)", d.Options[0].Label, d.Options[1].Label)
	}
	wantPlaced(t, s, "pick [1, 0] places the second listed trigger (the Main Theater unit's) first",
		[]*Asset{older, newer}, older, newer)
}

func TestFR_TRG_003_SweepRemovalsDoNotReorderTriggersInOneZone(t *testing.T) {
	// Covers: FR-MOD-001, SVY-R7, SVY-R13
	// This passes on main today, because buffer order already matches source
	// order here. It guards the rule against a fallback that reads zone positions
	// as the sweep removes assets. All four units are in Reserves, in this order:
	// PlainA, PlainB (both die), Survivor, Dying (dies). Survivor entered Reserves
	// before Dying, so its trigger is placed first. After the two Plain units
	// leave, Dying would stand ahead of where Survivor stood when its Unbowed fired.
	s := newScene(t)
	plainA := s.put(0, ZoneReserves, unit("PlainA", 1, 1, Keywords{}))
	plainB := s.put(0, ZoneReserves, unit("PlainB", 1, 1, Keywords{}))
	survivor := s.put(0, ZoneReserves, pingedBowed("Survivor"))
	dying := s.put(0, ZoneReserves, mortal("Dying", 1, 1, killedTrigger(scratchN(1))))
	s.deployFromHand(0, allySweeper())
	if s.alive(plainA) || s.alive(plainB) || s.alive(dying) || !s.alive(survivor) {
		t.Fatalf("setup: PlainA, PlainB and Dying die, Survivor lives (plainA alive=%v, plainB alive=%v, dying alive=%v, survivor alive=%v)", s.alive(plainA), s.alive(plainB), s.alive(dying), s.alive(survivor))
	}
	wantPlaced(t, s, "within one zone, the unit that entered Reserves first is placed first, however the sweep shifts positions",
		[]*Asset{survivor, dying}, survivor, dying)
}
