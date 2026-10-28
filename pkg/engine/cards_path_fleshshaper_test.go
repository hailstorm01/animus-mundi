package engine

import (
	"reflect"
	"testing"
)

// Card tests for Rogue Fleshshaper (T-0084), written from the CSV rules text in
// data/cards/units-and-structures-path.csv: Unit, 2 Requisition, Reserves, 2/3,
//   "When this unit deploys, choose one:
//    - Chosen asset heals 1.
//    - Deal 1 damage to chosen non-Scratch asset."
// "Chosen asset" is any asset of either side, Headquarters included, and the
// Fleshshaper itself. Cards are the real ones loaded with s.card. Mode 1 is
// answered with index 0 and mode 2 with index 1 (printed order).

// fleshBoard is player 0's deploying Fleshshaper's battlefield before it deploys:
// a damaged allied unit and opposing unit, one Scratch Company each, and an
// allied and an opposing unit with Ward. Damage is set directly.
type fleshBoard struct {
	s                *scene
	a1, e1, sc0, sc1 *Asset
	wardA, wardE     *Asset
	hq0, hq1         *Asset
}

func newFleshBoard(t *testing.T) *fleshBoard {
	t.Helper()
	s := newScene(t)
	b := &fleshBoard{
		s:   s,
		a1:  s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{})),
		e1:  s.put(1, ZoneMainTheater, unit("Enemy", 1, 5, Keywords{})),
		sc0: s.putScratch(0),
		sc1: s.putScratch(1),
		hq0: s.g.Players[0].HQ,
		hq1: s.g.Players[1].HQ,
	}
	return b
}

func (b *fleshBoard) withWard() {
	b.wardA = b.s.put(0, ZoneMainTheater, unit("Warded Ally", 1, 5, Keywords{Ward: true}))
	b.wardE = b.s.put(1, ZoneMainTheater, unit("Warded Enemy", 1, 5, Keywords{Ward: true}))
}

// targetAsk returns the one DecideTarget question put to seat 0.
func targetAsk(t *testing.T, s *scene) Decision {
	t.Helper()
	var out []Decision
	for _, d := range s.asked(0) {
		if d.Kind == DecideTarget {
			out = append(out, d)
		}
	}
	if len(out) != 1 {
		t.Fatalf("the chosen mode has one target slot, so exactly one target question: asked %d", len(out))
	}
	return out[0]
}

// Covers: FR-MOD-001, FR-TGT-001, SVY-R3
func TestCard_RogueFleshshaper_DeployAsksForAModeThenItsTarget(t *testing.T) {
	b := newFleshBoard(t)
	places := b.s.scriptModal(0, nil, nil, []Answer{modeAnswer(0)}, pick(b.a1))
	c := deployPath(b.s, "Rogue Fleshshaper", 2)
	asked := b.s.asked(0)
	if want := []DecisionKind{DecideMode, DecideTarget}; !reflect.DeepEqual(kindsOf(asked), want) {
		t.Fatalf("the mode is asked first, then the chosen mode's target: got %v, want %v", kindsOf(asked), want)
	}
	m := asked[0]
	if m.Player != 0 || m.Min != 1 || m.Max != 1 {
		t.Errorf("'choose one' is a mandatory single mode decision for the controller: %+v", m)
	}
	if want := []string{"mode 1", "mode 2"}; !reflect.DeepEqual(labels(m), want) {
		t.Errorf("mode options = %v, want %v in printed order", labels(m), want)
	}
	if asked[1].Player != 0 || asked[1].Min != 1 || asked[1].Max != 1 {
		t.Errorf("'chosen asset' is exactly one, asked of the controller: %+v", asked[1])
	}
	for i, n := range *places {
		if n != 1 {
			t.Errorf("question %d is asked before the trigger's place event: %d place events logged, want 1 (the deploy)", i+1, n)
		}
	}
	if got := placeCount(b.s.g); got != 2 {
		t.Errorf("the deploy and the trigger are 2 place events, got %d", got)
	}
	if p := assetOf(b.s.g, c).Printed(); p.Power != 2 || p.Health != 3 {
		t.Errorf("Rogue Fleshshaper is a 2/3: got %d/%d", p.Power, p.Health)
	}
}

func TestCard_RogueFleshshaper_HealModeHealsTheChosenAsset(t *testing.T) {
	b := newFleshBoard(t)
	b.a1.Damage, b.e1.Damage = 2, 2
	b.s.scriptModal(0, nil, nil, []Answer{modeAnswer(0)}, pick(b.e1))
	deployPath(b.s, "Rogue Fleshshaper", 2)
	b.s.wantDamage(b.e1, 1)
	b.s.wantDamage(b.a1, 2)
	wantNoNonCombatDamage(t, b.s.g, "mode 1 only heals")
}

func TestCard_RogueFleshshaper_HealModeOffersEveryAsset(t *testing.T) {
	b := newFleshBoard(t)
	b.s.scriptModal(0, nil, nil, []Answer{modeAnswer(0)}, pick(b.a1))
	c := deployPath(b.s, "Rogue Fleshshaper", 2)
	wantOptions(t, targetAsk(t, b.s), "heal mode: any asset of either side, Scratch Company, Headquarters and the Fleshshaper itself",
		b.a1, b.e1, b.sc0, b.sc1, b.hq0, b.hq1, assetOf(b.s.g, c))
}

func TestCard_RogueFleshshaper_DamageModeDealsOneToTheChosenAsset(t *testing.T) {
	b := newFleshBoard(t)
	b.s.scriptModal(0, nil, nil, []Answer{modeAnswer(1)}, pick(b.e1))
	deployPath(b.s, "Rogue Fleshshaper", 2)
	wantDealt(t, b.s.g, b.e1, 1, 1, "mode 2 deals 1 damage to the chosen opposing asset")
	b.s.wantDamage(b.e1, 1)
	if got := totalDamage(b.s.g); got != 1 {
		t.Errorf("nothing else is damaged: total damage %d, want 1", got)
	}

	d := newFleshBoard(t)
	weak := d.s.put(1, ZoneMainTheater, unit("Weak Enemy", 1, 1, Keywords{}))
	d.s.scriptModal(0, nil, nil, []Answer{modeAnswer(1)}, pick(weak))
	deployPath(d.s, "Rogue Fleshshaper", 2)
	if d.s.alive(weak) {
		t.Errorf("1 damage destroys the chosen 1-health asset")
	}
}

func TestCard_RogueFleshshaper_DamageModeDoesNotOfferScratchCompany(t *testing.T) {
	b := newFleshBoard(t)
	b.s.scriptModal(0, nil, nil, []Answer{modeAnswer(1)}, pick(b.e1))
	c := deployPath(b.s, "Rogue Fleshshaper", 2)
	wantOptions(t, targetAsk(t, b.s), "damage mode: non-Scratch assets of either side, Headquarters and the Fleshshaper included, no Scratch Company",
		b.a1, b.e1, b.hq0, b.hq1, assetOf(b.s.g, c))
}

// Covers: FR-KWD-002h, FR-TGT-003
func TestCard_RogueFleshshaper_OpposingWardAssetIsNotOffered(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode int
	}{{"heal mode", 0}, {"damage mode", 1}} {
		b := newFleshBoard(t)
		b.withWard()
		b.s.scriptModal(0, nil, nil, []Answer{modeAnswer(tc.mode)}, pick(b.wardA))
		c := deployPath(b.s, "Rogue Fleshshaper", 2)
		want := []*Asset{b.a1, b.e1, b.wardA, b.hq0, b.hq1, assetOf(b.s.g, c)}
		if tc.mode == 0 {
			want = append(want, b.sc0, b.sc1)
		}
		wantOptions(t, targetAsk(t, b.s), tc.name+": the opposing Ward unit is not a legal target, the allied Ward unit is", want...)
	}
}

// Covers: FR-MOD-001
func TestCard_RogueFleshshaper_TimeoutChoosesTheHealMode(t *testing.T) {
	b := newFleshBoard(t)
	b.a1.Damage = 2
	b.s.scriptModal(0, nil, nil, []Answer{{Timeout: true}}, pick(b.a1))
	deployPath(b.s, "Rogue Fleshshaper", 2)
	b.s.wantDamage(b.a1, 1)
	wantNoNonCombatDamage(t, b.s.g, "a timed-out mode choice is the first listed mode, heal, not the damage mode")
	if recs := modeRecords(b.s.g); len(recs) != 1 || recs[0].How != "timeout" {
		t.Errorf("the mode decision is recorded as a timeout: %+v", recs)
	}
}
