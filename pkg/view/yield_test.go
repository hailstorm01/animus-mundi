package view_test

import (
	"animus-mundi/pkg/engine/ability"
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/view"
)

// Tests for T-0236 (directive #150): every battlefield asset shows its printed
// resource yields to both seats, and a face-down asset shows the Scratch
// face's (FR-ZON-005).

// everyAsset is every battlefield asset in v: both HQs, both Reserves and the
// Main Theater.
func everyAsset(v view.View) []view.AssetView {
	var out []view.AssetView
	for _, s := range []view.Side{v.You, v.Opponent} {
		if s.HQ != nil {
			out = append(out, *s.HQ)
		}
		out = append(out, s.Reserves...)
	}
	return append(out, v.Theater...)
}

func TestView_AssetShowsItsPrintedYields(t *testing.T) {
	s := newScene(t)
	g := s.g
	unit := testUnit("Alpha Yielder", 2, 3, ability.Keywords{})
	unit.ReqYield, unit.TacYield = 2, 0
	a0 := s.put(0, engine.ZoneReserves, unit)
	depot := &engine.Card{Name: "Beta Depot", Type: engine.TypeStructure, Health: 4, ReqYield: 0, TacYield: 3,
		DeployZones: []engine.Zone{engine.ZoneReserves}}
	a1 := s.put(1, engine.ZoneReserves, depot)
	hidden := testUnit("Alpha Hidden Miner", 4, 4, ability.Keywords{})
	hidden.ReqYield, hidden.TacYield = 0, 3
	fd := &engine.Asset{ID: s.id(), Inst: s.inst(0, hidden), Controller: 0, Zone: engine.ZoneMainTheater, FaceDown: true}
	g.Theater = append(g.Theater, fd)
	theirs := s.put(1, engine.ZoneMainTheater, testUnit("Beta Raider", 3, 3, ability.Keywords{}))

	want := map[int][2]int{
		g.Players[0].HQ.ID: {1, 2}, // testHQ prints R1 T2
		g.Players[1].HQ.ID: {1, 2},
		a0.ID:              {2, 0}, // (a) a face-up unit's printed yields
		a1.ID:              {0, 3}, // (c) a Structure's
		fd.ID:              {1, 1}, // (b) the Scratch face's, not the card's R0 T3
		theirs.ID:          {1, 1}, // testUnit prints R1 T1
	}
	for seat := range engine.PlayerID(2) {
		v := view.PlayerView(g, seat)
		seen := map[int]bool{}
		for _, a := range everyAsset(v) {
			w, ok := want[a.ID]
			if !ok {
				continue
			}
			seen[a.ID] = true
			if a.ReqYield != w[0] || a.TacYield != w[1] {
				t.Errorf("seat %d sees %s#%d with yields R%d T%d, want R%d T%d", seat, a.Card, a.ID, a.ReqYield, a.TacYield, w[0], w[1])
			}
		}
		// (d) each seat sees every asset's yields, the opponent's included.
		if len(seen) != len(want) {
			t.Errorf("seat %d sees %d of the %d assets", seat, len(seen), len(want))
		}
	}
}
