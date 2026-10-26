// Zones: shuffling, drawing and moving cards and assets between zones (FR-ZON-001, FR-ZON-004).

package engine

import (
	"fmt"
	"slices"
)

func (g *Game) shuffle(p *Player) {
	g.rng.Shuffle(len(p.Deck), func(i, j int) { p.Deck[i], p.Deck[j] = p.Deck[j], p.Deck[i] })
}

func (g *Game) moveTopToHand(p *Player) {
	if len(p.Deck) > 0 {
		inst := p.Deck[0]
		g.takeCard(inst)
		g.putCard(inst, ZoneHand, false)
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
		g.takeCard(inst)
		a := g.newAsset(inst, pid, true)
		g.placeOnBattlefield(a, ZoneReserves)
		g.logEvent(Event{Kind: "scratch", Player: int(pid), Source: a.ID}, "player %d deploys %s", pid, a)
		g.fireDeploy(a) // a scratch is a unit deploying (KIM-S05)
	}
}

// takeCard removes inst from whichever of its owner's piles (deck, hand or
// discard) holds it, and reports whether one did. A card on the battlefield,
// on the stack or set aside is in no pile.
func (g *Game) takeCard(inst *Instance) bool {
	p := g.Players[inst.Owner]
	for _, pile := range []*[]*Instance{&p.Deck, &p.Hand, &p.Discard} {
		if slices.Contains(*pile, inst) {
			*pile = remove(*pile, inst)
			return true
		}
	}
	return false
}

// putCard puts inst into its owner's pile to, which is ZoneDeck, ZoneHand or
// ZoneDiscard. On the deck, bottom chooses the bottom over the top (index 0).
// Hand and discard always add at the end.
func (g *Game) putCard(inst *Instance, to Zone, bottom bool) {
	p := g.Players[inst.Owner]
	switch to {
	case ZoneDeck:
		if bottom {
			p.Deck = append(p.Deck, inst)
		} else {
			p.Deck = append([]*Instance{inst}, p.Deck...)
		}
	case ZoneHand:
		p.Hand = append(p.Hand, inst)
	case ZoneDiscard:
		p.Discard = append(p.Discard, inst)
	default:
		panic(fmt.Sprintf("engine: putCard to zone %v, which is not a pile", to))
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

// newAsset builds the asset inst becomes as it enters the battlefield under
// controller, with a new ID. Every way onto the battlefield builds its asset
// here. The entry state comes from the printed keywords: Deploys Exhausted
// replaces entering Readied (FR-KWD-002l; the Scratch Company face prints it,
// FR-ZON-005), and Armor N fills the Armor pool (FR-LAY-002, FR-KWD-002k).
func (g *Game) newAsset(inst *Instance, controller PlayerID, faceDown bool) *Asset {
	a := &Asset{ID: g.newID(), Inst: inst, Controller: controller, FaceDown: faceDown}
	kw := a.Printed().Keywords
	a.Exhausted = kw.DeploysExhausted
	a.Armor = kw.Armor
	return a
}

// placeOnBattlefield puts a into zone under its controller.
func (g *Game) placeOnBattlefield(a *Asset, zone Zone) {
	g.arrivals++
	a.arrived = g.arrivals
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
