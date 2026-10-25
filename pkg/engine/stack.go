// The Resolution Stack: stack items, pushing, LIFO resolution and the Deploy and March items (FR-QUE-001).

package engine

import "animus-mundi/pkg/engine/ability"

// StackKind identifies what a Resolution Stack item does when it resolves.
type StackKind int

const (
	StackDeploy StackKind = iota
	StackMarch
	StackAttack
	StackScratch
	StackAggressive
	StackAbility // a triggered ability: runs Effect's steps for Controller
)

// StackItem is one entry on the shared LIFO Resolution Stack (FR-QUE-001).
type StackItem struct {
	Kind       StackKind
	Controller PlayerID

	Inst *Instance // StackDeploy: the card on the stack
	Zone Zone      // StackDeploy: declared zone; StackMarch: origin zone

	Source    int // asset ID: marching unit, attacker, scratching HQ, Aggressive source
	Target    int // StackAttack: primary defender
	Secondary int // StackAttack: Berserk secondary target, 0 if none
	Amount    int // StackAggressive: N

	Effect  ability.Effect // StackAbility: the effect resolving (Source is its source asset)
	Chosen  [][]int        // StackAbility: asset IDs chosen for each Effect.Targets slot, in pick order; empty if none
	Tracked bool           // StackAbility: a death trigger, so Inst is its card tracked into its owner's discard pile (FR-ZON-004)
}

// push puts item on the stack and records where it went: the one place event
// for every item kind, with the stack depth after the push as Amount.
func (g *Game) push(item *StackItem) {
	g.Stack = append(g.Stack, item)
	e := Event{Kind: "place", Player: int(item.Controller), Source: item.Source, Amount: len(g.Stack)}
	var what string
	switch item.Kind {
	case StackDeploy:
		e.Card, e.Inst = item.Inst.Card.Name, item.Inst.ID
		what = "Deploy of " + e.Card
	case StackMarch:
		what = "March"
	case StackAttack:
		e.Target, e.Secondary = item.Target, item.Secondary
		what = "Attack"
	case StackScratch:
		what = "Scratch"
	case StackAggressive:
		what = "Aggressive trigger"
	case StackAbility:
		e.Card, e.Inst = cardName(item.Inst), 0
		if item.Inst != nil {
			e.Inst = item.Inst.ID
		}
		what = "ability of " + e.Card
	}
	g.logEvent(e, "player %d places %s on the stack", item.Controller, what)
}

// resolveStack places any triggers the action fired, then resolves top-down
// until the stack is empty, sweeping for deaths after each item (FR-QUE-001,
// FR-LAY-002, FR-TRG-004). Placing first puts a March's or an Attack's
// triggers on top of it, and resolves a Generate's before the next action
// (SVY-R17).
func (g *Game) resolveStack() {
	g.flushTriggers()
	for len(g.Stack) > 0 && !g.Outcome.Over {
		top := g.Stack[len(g.Stack)-1]
		g.Stack = g.Stack[:len(g.Stack)-1]
		g.resolve(top)
		g.deathSweep()
		g.flushTriggers()
		g.settled()
	}
}

func (g *Game) resolve(it *StackItem) {
	switch it.Kind {
	case StackDeploy:
		g.resolveDeploy(it)
	case StackMarch:
		g.resolveMarch(it)
	case StackAttack:
		g.resolveAttack(it)
	case StackScratch:
		g.scratch(it.Controller, 1)
	case StackAbility:
		g.resolveAbility(it)
	case StackAggressive:
		// "it gets +N power until end of turn" (FR-KWD-002f). Fizzles if the
		// attacker already left the battlefield.
		if a := g.Asset(it.Source); a != nil {
			a.TempPower += it.Amount
			g.logEvent(Event{Kind: "trigger", Player: -1, Source: a.ID},
				"%s Aggressive %d: +%d power until end of turn", a, it.Amount, it.Amount)
		}
	}
}

func (g *Game) resolveDeploy(it *StackItem) {
	c := it.Inst.Card
	a := &Asset{ID: g.newID(), Inst: it.Inst, Controller: it.Controller}
	// Deploys Exhausted replaces entering Readied (FR-KWD-002l).
	a.Exhausted = c.Keywords.DeploysExhausted
	// Armor N initializes the Armor pool on entry (FR-LAY-002, FR-KWD-002k).
	a.Armor = c.Keywords.Armor
	g.placeOnBattlefield(a, it.Zone)
	g.logEvent(Event{Kind: "deploy", Player: int(it.Controller), Card: c.Name, Inst: it.Inst.ID, Source: a.ID, Zone: it.Zone.String()},
		"player %d deploys %s to %s", it.Controller, a, it.Zone)
	g.fireDeploy(a)
}

func (g *Game) resolveMarch(it *StackItem) {
	a := g.Asset(it.Source)
	if a == nil || a.Zone != it.Zone {
		return // left the battlefield or already moved: the March does nothing
	}
	g.removeFromBattlefield(a)
	dest := ZoneMainTheater
	if it.Zone == ZoneMainTheater {
		dest = ZoneReserves
	}
	// Damage, Armor and modifiers persist across battlefield zones (FR-ZON-004).
	g.placeOnBattlefield(a, dest)
	g.logEvent(Event{Kind: "march", Player: int(it.Controller), Source: a.ID, Zone: dest.String()},
		"%s marches to %s", a, dest)
}
