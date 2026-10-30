package sim

import "animus-mundi/pkg/engine"

// watchEntry queues the entry-state check for a deploy or scratch event.
func (m *Monitors) watchEntry(e engine.Event) {
	const mon = monEntry
	m.pending = append(m.pending, pendingCheck{seq: e.Seq, asset: e.Source, run: func(after engine.Snapshot, targeted bool) {
		a := find(after, e.Source)
		if a == nil {
			return // left the battlefield before the state settled: skip
		}
		if e.Kind == engine.EventScratch {
			m.count(mon, "FR-ZON-005")
			if a.Zone != engine.ZoneReserves.String() || a.Controller != e.Player || !a.FaceDown || !a.Exhausted {
				m.fail(mon, e.Seq, "Scratch Company %d entered as zone %s, controller %d, face-down %t, Exhausted %t; want Reserves, controller %d, face-down, Exhausted",
					a.ID, a.Zone, a.Controller, a.FaceDown, a.Exhausted, e.Player)
			}
			return
		}
		c := m.card(*a)
		if c == nil {
			m.fail(mon, e.Seq, "%s is not in its owner's decklist", a.Card)
			return
		}
		if a.FaceDown || a.Zone != e.Zone {
			m.fail(mon, e.Seq, "%s#%d entered face-down %t in %s, want face-up in %s", a.Card, a.ID, a.FaceDown, a.Zone, e.Zone)
		}
		if c.Keywords.DeploysExhausted {
			m.count(mon, "FR-KWD-002l")
		} else {
			m.count(mon, "FR-SET-004")
		}
		if a.Exhausted != c.Keywords.DeploysExhausted {
			m.fail(mon, e.Seq, "%s#%d entered with Exhausted %t, want %t", a.Card, a.ID, a.Exhausted, c.Keywords.DeploysExhausted)
		}
		if n := c.Keywords.Armor; n > 0 && !targeted {
			m.count(mon, "FR-KWD-002k")
			if a.Armor != n {
				m.fail(mon, e.Seq, "%s#%d entered with Armor %d, want its printed Armor %d", a.Card, a.ID, a.Armor, n)
			}
		}
	}})
}
