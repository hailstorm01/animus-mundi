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
			g.logEvent(Event{Kind: EventDraw, Player: int(pid)}, "player %d must draw from an empty deck", pid)
			return
		}
		g.moveTopToHand(p)
	}
}

// deployCard puts inst, already out of its pile, onto the battlefield face-up in
// zone under controller, with its own deploy event and deploy triggers, and
// returns the new asset. A free deploy enters this way during resolution, with
// no stack item and no costs (FR-TRN-003a). scratch has its own path, because it
// deploys face-down and logs its own event.
func (g *Game) deployCard(inst *Instance, controller PlayerID, zone Zone) *Asset {
	a := g.newAsset(inst, controller, false)
	g.placeOnBattlefield(a, zone)
	g.logEvent(Event{Kind: EventDeploy, Player: int(controller), Card: inst.Card.Name, Inst: inst.ID, Source: a.ID, Zone: zone.String()},
		"player %d deploys %s to %s", controller, a, zone)
	g.fireDeploy(a)
	return a
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
		g.logEvent(Event{Kind: EventScratch, Player: int(pid), Source: a.ID}, "player %d deploys %s", pid, a)
		g.fireDeploy(a) // a scratch is a unit deploying (KIM-S05)
	}
}

// discardFromHand has player pid discard n cards of their choice from their
// own hand, or the whole hand if it holds fewer, and returns the cards
// discarded, in pick order: chooseDiscards, then discardCards.
func (g *Game) discardFromHand(pid PlayerID, n int, upTo bool) []*Instance {
	picked := g.chooseDiscards(pid, n, upTo)
	g.discardCards(pid, picked)
	return picked
}

// chooseDiscards has player pid pick n cards of their choice from their own
// hand, or the whole hand if it holds fewer, and returns them in pick order.
// It moves nothing. Choosing from your own hand is not an inspection, so a hand
// of exactly that many cards is picked without asking (FR-MOD-001 §1, §6;
// SVY-R16). With upTo it picks up to n cards, possibly none: the choice is
// asked whenever the hand isn't empty, and a timeout picks none (FR-MOD-001 §6,
// FR-TRN-003a). Every hand-discard decision is built here.
func (g *Game) chooseDiscards(pid PlayerID, n int, upTo bool) []*Instance {
	hand := g.Players[pid].Hand
	n = min(n, len(hand))
	if n <= 0 {
		return nil
	}
	d := Decision{Player: pid, Kind: DecideHand, Prompt: "Discard a card", Min: n, Max: n}
	if n > 1 {
		d.Prompt = fmt.Sprintf("Discard %d cards", n)
	}
	if upTo {
		d.Min = 0
		d.Prompt = fmt.Sprintf("Discard up to %d cards", n)
		if n == 1 {
			d.Prompt = "Discard up to 1 card"
		}
	}
	for _, inst := range hand {
		d.Options = append(d.Options, Option{Label: inst.Card.Name, Ref: inst.ID})
	}
	var picked []*Instance
	for _, i := range g.choose(d) {
		picked = append(picked, hand[i])
	}
	return picked
}

// discardCards moves each of cards, in the given order, from its pile to its
// owner's discard pile and logs the discard. Every discard from hand is moved
// here.
func (g *Game) discardCards(pid PlayerID, cards []*Instance) {
	for _, inst := range cards {
		g.takeCard(inst)
		g.putCard(inst, ZoneDiscard, false)
		g.logEvent(Event{Kind: EventDiscard, Player: int(pid), Card: inst.Card.Name, Inst: inst.ID, Zone: ZoneHand.String()},
			"player %d discards %s", pid, inst.Card.Name)
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

// DiscardCard returns the card with instance ID id if it is in a discard pile,
// else nil. It is the one lookup of a chosen pile card by ID: the engine
// re-verifies a pile target with it (pileTarget, FR-TGT-003), and a client
// uses it to show a chosen card that is still there.
func (g *Game) DiscardCard(id int) *Instance {
	for _, pl := range g.Players {
		for _, inst := range pl.Discard {
			if inst.ID == id {
				return inst
			}
		}
	}
	return nil
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

// moveOnBattlefield moves a from its battlefield zone to the battlefield zone to.
// Damage, Armor and modifiers persist across battlefield zones (FR-ZON-004), and
// a gets a new arrival number on placement (SVY-R7).
func (g *Game) moveOnBattlefield(a *Asset, to Zone) {
	g.removeFromBattlefield(a)
	g.placeOnBattlefield(a, to)
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
