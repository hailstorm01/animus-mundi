package engine

// StackKind identifies what a Resolution Stack item does when it resolves.
type StackKind int

const (
	StackDeploy StackKind = iota
	StackMarch
	StackAttack
	StackScratch
	StackAggressive
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
	case StackAggressive:
		// "it gets +N power until end of turn" (FR-KWD-002f). Fizzles if the
		// attacker already left the battlefield.
		if a := g.Asset(it.Source); a != nil {
			a.TempPower += it.Amount
			g.logf("trigger", "%s Aggressive %d: +%d power until end of turn", a, it.Amount, it.Amount)
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
	g.logf("deploy", "player %d deploys %s to %s", it.Controller, a, it.Zone)
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
	g.logf("march", "%s marches to %s", a, dest)
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
		// OPEN(KIM-S01): FR-ZON-005 does not state the readiness of a new
		// Scratch Company; FR-SET-004's "Readied by default" is applied.
		a := &Asset{ID: g.newID(), Inst: inst, Controller: pid, FaceDown: true}
		g.placeOnBattlefield(a, ZoneReserves)
		g.logf("scratch", "player %d deploys %s", pid, a)
	}
}

// combatLegalZone reports whether a is still somewhere it can fight from.
func combatLegalZone(a *Asset) bool {
	if a.Printed().Type == TypeHeadquarters {
		return a.Zone == ZoneHeadquarters
	}
	return a.Zone == ZoneMainTheater
}

// resolveAttack is Step 3 of combat (FR-CBT-002).
func (g *Game) resolveAttack(it *StackItem) {
	att := g.Asset(it.Source)
	def := g.Asset(it.Target)
	if att == nil || def == nil || !combatLegalZone(att) || !combatLegalZone(def) || def.Controller == att.Controller {
		// OPEN(KIM-B03): whether Berserk cleave still happens when the attack
		// fizzles is undefined; the kernel treats the whole attack as fizzled.
		g.logf("fizzle", "attack fizzles")
		return
	}

	// Lock effective values at the start of the simultaneous step (FR-LAY-003).
	attPower, defPower := g.Power(att), g.Power(def)
	berserk := 0
	var sec *Asset
	if att.Printed().Type == TypeUnit && def.Zone == ZoneMainTheater && it.Secondary != 0 {
		berserk = g.Keywords(att).Berserk
		sec = g.Asset(it.Secondary)
		// Two-stage verification (FR-TGT-003): the secondary must still be an
		// opposing asset in the Main Theater, otherwise the cleave dissipates.
		if sec != nil && (sec.Zone != ZoneMainTheater || sec.Controller == att.Controller) {
			sec = nil
		}
	}

	g.logf("combat", "%s (%d) and %s (%d) exchange combat damage", att, attPower, def, defPower)
	g.dealDamage(att, def, attPower, true)
	g.dealDamage(def, att, defPower, true)
	if berserk > 0 && sec != nil {
		// Berserk cleave is non-combat damage committed within Step 3, before
		// the death sweep, whether or not either combatant dies (FR-KWD-002o).
		g.dealDamage(att, sec, berserk, false)
	}
}
