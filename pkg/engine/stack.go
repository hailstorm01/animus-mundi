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
	StackAbility // a triggered ability: runs Effect's steps for Controller
)

// StackItem is one entry on the shared LIFO Resolution Stack (FR-QUE-001).
type StackItem struct {
	Kind       StackKind
	Controller PlayerID

	Inst *Instance // StackDeploy: the card on the stack
	Zone Zone      // StackDeploy: declared zone; StackMarch: origin zone

	Source    int // asset ID: marching unit, attacker, scratching HQ
	Target    int // StackAttack: primary defender
	Secondary int // StackAttack: Berserk secondary target, 0 if none

	Effect  ability.Effect  // StackAbility: the effect resolving (Source is its source asset)
	Chosen  [][]int         // StackAbility: per Effect.Targets slot, in pick order, the asset IDs chosen, or for a PileSlot the card instance IDs; empty if none
	Tracked bool            // StackAbility: a death trigger, so Inst is its card tracked into its owner's discard pile (FR-ZON-004)
	Keyword ability.Keyword // StackAbility: the keyword its trigger came from (ability.KeywordAggressive); zero for a printed ability
}

// PileSlot reports whether Effect.Targets[i] chooses cards in discard piles, so
// Chosen[i] holds card instance IDs, not asset IDs. It is false when i is out
// of range of Effect.Targets.
func (it *StackItem) PileSlot(i int) bool {
	return i >= 0 && i < len(it.Effect.Targets) && it.Effect.Targets[i].Filter.InPile()
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
	case StackAbility:
		e.Card, e.Inst = cardName(item.Inst), 0
		if item.Inst != nil {
			e.Inst = item.Inst.ID
		}
		what = "ability of " + e.Card
		if item.Keyword != ability.KeywordNone {
			what = item.Keyword.String() + " trigger of " + e.Card
		}
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
	}
}

func (g *Game) resolveDeploy(it *StackItem) {
	c := it.Inst.Card
	a := g.newAsset(it.Inst, it.Controller, false)
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
