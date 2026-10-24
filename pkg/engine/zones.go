// Zones: shuffling, drawing and moving cards and assets between zones (FR-ZON-001, FR-ZON-004).

package engine

func (g *Game) shuffle(p *Player) {
	g.rng.Shuffle(len(p.Deck), func(i, j int) { p.Deck[i], p.Deck[j] = p.Deck[j], p.Deck[i] })
}

func (g *Game) moveTopToHand(p *Player) {
	if len(p.Deck) > 0 {
		p.Hand = append(p.Hand, p.Deck[0])
		p.Deck = p.Deck[1:]
	}
}

// draw draws n cards. A draw from an empty deck does nothing and marks the
// player to lose at the next death sweep (FR-WIN-001).
func (g *Game) draw(pid PlayerID, n int) {
	p := g.Players[pid]
	for range n {
		if len(p.Deck) == 0 {
			p.DrewFromEmpty = true
			g.logEvent(Event{Kind: "draw", Player: int(pid)}, "player %d must draw from an empty deck", pid)
			return
		}
		g.moveTopToHand(p)
	}
}

// scratch performs Deploy Scratch Company n times (FR-ZON-005, FR-KWD-002m).
func (g *Game) scratch(pid PlayerID, n int) {
	p := g.Players[pid]
	for range n {
		if len(p.Deck) == 0 {
			return // not a draw: no Empty Deck Loss
		}
		inst := p.Deck[0]
		p.Deck = p.Deck[1:]
		a := &Asset{ID: g.newID(), Inst: inst, Controller: pid, FaceDown: true}
		// The face prints Deploys Exhausted, so it enters Exhausted (FR-ZON-005, FR-KWD-002l).
		a.Exhausted = a.Printed().Keywords.DeploysExhausted
		g.placeOnBattlefield(a, ZoneReserves)
		g.logEvent(Event{Kind: "scratch", Player: int(pid), Source: a.ID}, "player %d deploys %s", pid, a)
		g.fireDeploy(a) // a scratch is a unit deploying (KIM-S05)
	}
}

// leaveBattlefield removes a from the battlefield to a non-battlefield zone,
// keeping a last-known record with eff, a's effective characteristics taken
// immediately prior to departure. A lone departure passes g.effectiveState(a);
// the death sweep takes every dying asset's state before any of them leaves.
func (g *Game) leaveBattlefield(a *Asset, eff EffectiveState) {
	if g.lastKnown == nil {
		g.lastKnown = make(map[int]*lastKnownRecord)
	}
	g.lastKnown[a.ID] = &lastKnownRecord{asset: *a, eff: eff}
	g.removeFromBattlefield(a)
}

// removeFromBattlefield detaches a from whichever battlefield zone holds it.
func (g *Game) removeFromBattlefield(a *Asset) {
	switch a.Zone {
	case ZoneMainTheater:
		g.Theater = remove(g.Theater, a)
	case ZoneReserves:
		p := g.Players[a.Controller]
		p.Reserves = remove(p.Reserves, a)
	case ZoneHeadquarters:
		g.Players[a.Controller].HQ = nil
	}
}

// placeOnBattlefield puts a into zone under its controller.
func (g *Game) placeOnBattlefield(a *Asset, zone Zone) {
	a.Zone = zone
	switch zone {
	case ZoneMainTheater:
		g.Theater = append(g.Theater, a)
	case ZoneReserves:
		p := g.Players[a.Controller]
		p.Reserves = append(p.Reserves, a)
	case ZoneHeadquarters:
		g.Players[a.Controller].HQ = a
	}
}

func remove[T comparable](s []T, v T) []T {
	for i, x := range s {
		if x == v {
			return append(s[:i:i], s[i+1:]...)
		}
	}
	return s
}

func handCard(p *Player, id int) *Instance {
	for _, inst := range p.Hand {
		if inst.ID == id {
			return inst
		}
	}
	return nil
}
