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
	StackAbility  // a triggered or activated ability: runs Effect's steps for Controller
	StackManeuver // an enacted Maneuver: runs Effect's steps for Controller, then Inst goes to its owner's discard pile (FR-MNV-001)
)

// LockedZone is the zone a stack item's source-relative filters read, locked
// once (SVY-R44, SVY-R45). Player is the owner of a Reserves or Headquarters
// Zone; it is unused for the shared Main Theater (FR-ZON-001).
type LockedZone struct {
	Zone   Zone
	Player PlayerID
}

// StackItem is one entry on the shared LIFO Resolution Stack (FR-QUE-001).
type StackItem struct {
	Kind       StackKind
	Controller PlayerID

	Inst *Instance // StackDeploy, StackManeuver: the card on the stack
	Zone Zone      // StackDeploy: declared zone; StackMarch: origin zone

	Source int // asset ID: marching unit, attacker, scratching HQ; 0 for a StackManeuver (a Maneuver is not an asset)
	Target int // StackAttack: the defender

	Effect  ability.Effect  // StackAbility, StackManeuver: the effect resolving (for StackAbility, Source is its source asset)
	Chosen  [][]int         // StackAbility, StackManeuver: per Effect.Targets slot, in pick order, the asset IDs chosen, or for a PileSlot the card instance IDs; empty if none
	Tracked bool            // StackAbility: a death trigger, so Inst is its card tracked into its owner's discard pile (FR-ZON-004); false for an activated ability
	Keyword ability.Keyword // StackAbility: the keyword its trigger came from (ability.KeywordAggressive); zero for a printed trigger or an activated ability
	Locked  *LockedZone     // StackAbility, StackManeuver: the zone locked for its source-relative filters (SVY-R44, SVY-R45); nil if it locked none
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
	e := Event{Kind: EventPlace, Player: int(item.Controller), Source: item.Source, Amount: len(g.Stack)}
	var what string
	switch item.Kind {
	case StackDeploy:
		e.Card, e.Inst = item.Inst.Card.Name, item.Inst.ID
		what = "Deploy of " + e.Card
	case StackMarch:
		what = "March"
	case StackAttack:
		e.Target = item.Target
		what = "Attack"
	case StackScratch:
		what = "Scratch"
	case StackManeuver:
		e.Card, e.Inst = item.Inst.Card.Name, item.Inst.ID
		what = "Maneuver " + e.Card
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
	case StackManeuver:
		g.resolveManeuver(it)
	}
}

// resolveManeuver resolves an enacted Maneuver's effect as resolveAbility does
// (two-stage target verification, then the steps), then puts the card into its
// owner's discard pile, even if every target fizzled (FR-MNV-001).
func (g *Game) resolveManeuver(it *StackItem) {
	g.resolveAbility(it)
	g.putCard(it.Inst, ZoneDiscard, false)
}

func (g *Game) resolveDeploy(it *StackItem) {
	g.deployCard(it.Inst, it.Controller, it.Zone)
}

func (g *Game) resolveMarch(it *StackItem) {
	a := g.Asset(it.Source)
	if a == nil || a.Zone != it.Zone {
		return // left the battlefield or already moved: the March does nothing
	}
	dest := ZoneMainTheater
	if it.Zone == ZoneMainTheater {
		dest = ZoneReserves
	}
	g.moveOnBattlefield(a, dest)
	g.logEvent(Event{Kind: EventMarch, Player: int(it.Controller), Source: a.ID, Zone: dest.String()},
		"%s marches to %s", a, dest)
}
