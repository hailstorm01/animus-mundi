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
	StackAbility // a triggered ability: runs Trigger's steps for Controller
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

	Trigger ability.Trigger // StackAbility: the ability resolving (Source is its source asset)
	Chosen  []int           // StackAbility: asset ID chosen for each Trigger.Effect.Targets slot, 0 if none
}

func (g *Game) push(item *StackItem) {
	g.Stack = append(g.Stack, item)
}

// resolveStack resolves top-down until the stack is empty, sweeping for
// deaths after each item (FR-QUE-001, FR-LAY-002, FR-TRG-004).
func (g *Game) resolveStack() {
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

// scratch performs Deploy Scratch Company n times (FR-ZON-005, FR-KWD-002m).
func (g *Game) scratch(pid PlayerID, n int) {
	p := g.Players[pid]
	for range n {
		if len(p.Deck) == 0 {
			return // not a draw: no Empty Deck Loss
		}
		inst := p.Deck[0]
		p.Deck = p.Deck[1:]
		// A new Scratch Company always enters Exhausted (FR-ZON-005).
		a := &Asset{ID: g.newID(), Inst: inst, Controller: pid, FaceDown: true, Exhausted: true}
		g.placeOnBattlefield(a, ZoneReserves)
		g.logEvent(Event{Kind: "scratch", Player: int(pid), Source: a.ID}, "player %d deploys %s", pid, a)
	}
}

// combatLegalZone reports whether a is still somewhere it can fight from.
func combatLegalZone(a *Asset) bool {
	if a.Printed().Type == TypeHeadquarters {
		return a.Zone == ZoneHeadquarters
	}
	return a.Zone == ZoneMainTheater
}

// resolveAttack is Step 3 of combat (FR-CBT-002). The target was committed at
// declaration: Guardian or Ethereal changes since then do not matter, only
// whether both combatants are still where they can fight.
func (g *Game) resolveAttack(it *StackItem) {
	def := g.Asset(it.Target)
	// OPEN(SVY-R8): fizzles if the defender changed controller before resolution.
	if def == nil || !combatLegalZone(def) || def.Controller == it.Controller {
		g.logEvent(Event{Kind: "fizzle", Player: -1, Source: it.Source},
			"attack fizzles: the defender is gone, so there is no Berserk cleave")
		return
	}

	att := g.Asset(it.Source)
	if att == nil || !combatLegalZone(att) {
		// Attacker-side fizzle: no combat damage, but Berserk cleave still
		// happens from the attacker's last-known characteristics (FR-CBT-002).
		src := att
		if src == nil {
			src = g.LastKnown(it.Source)
		}
		g.logEvent(Event{Kind: "fizzle", Player: -1, Source: it.Source}, "attack fizzles: the attacker is gone")
		g.cleave(it, src, def)
		return
	}

	// Lock effective values at the start of the simultaneous step (FR-LAY-003).
	attPower, defPower := g.Power(att), g.Power(def)
	g.logf("combat", "%s (%d) and %s (%d) exchange combat damage", att, attPower, def, defPower)
	g.dealDamage(att, def, attPower, true)
	g.dealDamage(def, att, defPower, true)
	// Berserk cleave is committed within Step 3, before the death sweep,
	// whether or not either combatant dies (FR-KWD-002o).
	g.cleave(it, att, def)
}

// cleave deals Berserk N non-combat damage to the attack's secondary target if
// it is still legal (two-stage verification, FR-TGT-003).
func (g *Game) cleave(it *StackItem, att, def *Asset) {
	if att == nil || it.Secondary == 0 || att.Printed().Type != TypeUnit || def.Zone != ZoneMainTheater {
		return
	}
	n := g.Keywords(att).Berserk
	sec := g.Asset(it.Secondary)
	if n <= 0 || sec == nil || !g.legalSecondary(att, sec) {
		return
	}
	g.dealDamage(att, sec, n, false)
}
