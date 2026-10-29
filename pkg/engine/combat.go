// Combat: attack legality (FR-CBT-001) and the 3-step combat sequence (FR-CBT-002).

package engine

// checkAttack returns nil if a, an Attack, is legal right now (FR-CBT-001):
// the attacker is a Readied asset of a.Player's on the battlefield
// (readyAsset, so an Exhausted asset cannot attack), it is a Unit or
// Headquarters in its combat zone (canDeclareAttack), and the defender is one
// it may attack (CanAttack). check calls it after the checks every action
// shares.
func (g *Game) checkAttack(a Action) error {
	att, err := g.readyAsset(a)
	if err != nil {
		return err
	}
	if !canDeclareAttack(att) {
		return illegal("%s cannot attack from %s (FR-CBT-001)", att, att.Zone)
	}
	def := g.Asset(a.Target)
	if def == nil || !g.CanAttack(att, def) {
		return illegal("illegal combat target (FR-CBT-001)")
	}
	return nil
}

// declareAttack is Steps 1 and 2 of combat (FR-CBT-002).
func (g *Game) declareAttack(a Action) {
	att := g.Asset(a.Asset)
	g.exhaust(att)
	g.push(&StackItem{Kind: StackAttack, Controller: a.Player, Source: att.ID, Target: a.Target})
	g.logEvent(Event{Kind: "attack", Player: int(a.Player), Source: att.ID, Target: a.Target},
		"%s attacks %s", att, g.Asset(a.Target))
	// Step 2: attack triggers enqueue on top of the Attack action; Apply's
	// resolveStack places them before anything resolves.
	g.fireAttacks(att)
}

func canDeclareAttack(a *Asset) bool {
	t := a.Printed().Type
	return (t == TypeUnit || t == TypeHeadquarters) && inCombatZone(a)
}

// CanAttack implements defender eligibility, Ethereal and Guardian (FR-CBT-001).
func (g *Game) CanAttack(att, def *Asset) bool {
	if def.Controller == att.Controller {
		return false
	}
	if def.Zone != ZoneMainTheater && def.Zone != ZoneHeadquarters {
		return false // Reserves are never attackable
	}
	if g.Keywords(def).Ethereal {
		return false
	}
	if g.hasGuardian(def.Controller) {
		return def.Zone == ZoneMainTheater && g.Keywords(def).Guardian
	}
	return true
}

// hasGuardian reports whether p controls a non-Ethereal Guardian in the Main Theater.
func (g *Game) hasGuardian(p PlayerID) bool {
	for _, a := range g.TheaterOf(p) {
		kw := g.Keywords(a)
		if kw.Guardian && !kw.Ethereal {
			return true
		}
	}
	return false
}

// inCombatZone reports whether a is where it can fight: a Headquarters in the
// Headquarters Zone, anything else in the Main Theater.
func inCombatZone(a *Asset) bool {
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
	// The defender must still oppose the attacker's current controller (its
	// last-known controller once it has left), not the declaring player
	// (FR-CBT-002 Step 3). A control change on either side fizzles the attack.
	attController := it.Controller
	if a := g.Asset(it.Source); a != nil {
		attController = a.Controller
	} else if lk := g.LastKnown(it.Source); lk != nil {
		attController = lk.Controller
	}
	if def == nil || !inCombatZone(def) || def.Controller == attController {
		g.logEvent(Event{Kind: "fizzle", Player: -1, Source: it.Source},
			"attack fizzles: the defender is gone")
		return
	}

	att := g.Asset(it.Source)
	if att == nil || !inCombatZone(att) {
		g.logEvent(Event{Kind: "fizzle", Player: -1, Source: it.Source}, "attack fizzles: the attacker is gone")
		return
	}

	// Lock effective values at the start of the simultaneous step (FR-LAY-003).
	attPower, defPower := g.Power(att), g.Power(def)
	g.logEvent(Event{Kind: "combat", Player: -1, Source: att.ID, Target: def.ID},
		"%s (%d) and %s (%d) exchange combat damage", att, attPower, def, defPower)
	g.dealDamage(att, def, attPower, true)
	g.dealDamage(def, att, defPower, true)
	// Only the declared attacker fires, and only for the defender: a
	// retaliation kill does not count (FR-CBT-002 Step 3).
	if g.IsLethal(def) {
		g.fireCombatKill(att)
	}
}
