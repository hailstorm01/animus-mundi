package sim

import "animus-mundi/pkg/engine"

func (m *Monitors) checkTurnStart(e engine.Event) {
	m.count(monResources, "FR-RES-002")
	if e.Snapshot == nil {
		return
	}
	for p, ps := range e.Snapshot.Players {
		if ps.Req != 0 || ps.Tac != 0 {
			m.fail(monResources, e.Seq, "player %d starts turn %d with Requisition %d and Tactics %d, want both 0", p, e.Snapshot.Turn, ps.Req, ps.Tac)
		}
	}
}

// pool is player p's Requisition or Tactics in s.
func pool(s engine.Snapshot, p int, resource string) int {
	if resource == "Requisition" {
		return s.Players[p].Req
	}
	return s.Players[p].Tac
}

func (m *Monitors) checkGenerate(e engine.Event) {
	const mon = monResources
	m.count(mon, "FR-CAP-001")
	b := m.last
	if b == nil {
		m.fail(mon, e.Seq, "Generate before any settled snapshot")
		return
	}
	a := find(*b, e.Source)
	if a == nil {
		m.fail(mon, e.Seq, "Generate from asset %d, which is not on the battlefield", e.Source)
		return
	}
	c := m.card(*a)
	if c == nil {
		m.fail(mon, e.Seq, "%s is not in its owner's decklist", a.Card)
		return
	}
	if a.Controller != e.Player {
		m.fail(mon, e.Seq, "%s#%d is not controlled by the generating player %d", a.Card, a.ID, e.Player)
	}
	if a.Exhausted {
		m.fail(mon, e.Seq, "%s#%d was Exhausted before it generated", a.Card, a.ID)
	}
	res, other := e.Resource, "Tactics"
	yield := c.TacYield
	if res == "Requisition" {
		other, yield = "Tactics", c.ReqYield
	} else if res == "Tactics" {
		other = "Requisition"
	} else {
		m.fail(mon, e.Seq, "Generate names resource %q", res)
		return
	}
	if c.Type == engine.TypeUnit {
		reserves := engine.ZoneReserves.String()
		switch {
		case c.Keywords.Versatile:
			m.count(mon, "FR-KWD-002i")
		case a.Zone == reserves && res != "Requisition":
			m.fail(mon, e.Seq, "%s#%d in Reserves generated %s, not Requisition", a.Card, a.ID, res)
		case a.Zone != reserves && res != "Tactics":
			m.fail(mon, e.Seq, "%s#%d in %s generated %s, not Tactics", a.Card, a.ID, a.Zone, res)
		}
	}
	if yield == 0 {
		m.count(mon, "KIM-S03")
	}
	p, id, name := e.Player, a.ID, a.Card
	beforeRes, beforeOther := pool(*b, p, res), pool(*b, p, other)
	m.pending = append(m.pending, pendingCheck{seq: e.Seq, run: func(after engine.Snapshot, _ bool) {
		if n := find(after, id); n == nil || !n.Exhausted {
			m.fail(mon, e.Seq, "%s#%d is not Exhausted after it generated", name, id)
		}
		if got := pool(after, p, res) - beforeRes; got != yield {
			m.fail(mon, e.Seq, "%s rose by %d after %s#%d generated, want its yield %d", res, got, name, id, yield)
		}
		if got := pool(after, p, other) - beforeOther; got != 0 {
			m.fail(mon, e.Seq, "%s changed by %d when %s#%d generated %s", other, got, name, id, res)
		}
	}})
}

func (m *Monitors) checkConvert(e engine.Event) {
	const mon = monResources
	m.count(mon, "FR-RES-003")
	b := m.last
	if b == nil {
		m.fail(mon, e.Seq, "Convert before any settled snapshot")
		return
	}
	p := e.Player
	req, tac := b.Players[p].Req, b.Players[p].Tac
	if req < 2 {
		m.fail(mon, e.Seq, "player %d converted with %d Requisition, want at least 2", p, req)
	}
	m.pending = append(m.pending, pendingCheck{seq: e.Seq, run: func(after engine.Snapshot, _ bool) {
		if after.Players[p].Req != req-2 || after.Players[p].Tac != tac+1 {
			m.fail(mon, e.Seq, "Convert took player %d from R%d T%d to R%d T%d, want R%d T%d",
				p, req, tac, after.Players[p].Req, after.Players[p].Tac, req-2, tac+1)
		}
	}})
}
