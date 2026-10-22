// Combat: attack legality (FR-CBT-001) and the 3-step combat sequence (FR-CBT-002).

package engine

import "animus-mundi/pkg/engine/ability"

// declareAttack is Steps 1 and 2 of combat (FR-CBT-002).
func (g *Game) declareAttack(a Action) {
	att := g.Asset(a.Asset)
	att.Exhausted = true
	g.push(&StackItem{Kind: StackAttack, Controller: a.Player, Source: att.ID, Target: a.Target, Secondary: a.Secondary})
	g.logEvent(Event{Kind: "attack", Player: int(a.Player), Source: att.ID, Target: a.Target},
		"%s attacks %s", att, g.Asset(a.Target))
	// Step 2: attack triggers enqueue on top of the Attack action.
	g.fireAttacks(att)
	g.flushTriggers()
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

// CanCleave reports whether sec is a legal Berserk secondary target.
func (g *Game) CanCleave(att, def, sec *Asset) bool {
	if att.Printed().Type != TypeUnit || g.Keywords(att).Berserk <= 0 {
		return false
	}
	if def.Zone != ZoneMainTheater || sec.ID == def.ID {
		return false // Berserk only cleaves when attacking a Main Theater asset
	}
	return g.legalSecondary(att, sec)
}

// hasCleaveTarget reports whether any legal Berserk secondary exists for this
// attack, which makes choosing one mandatory (FR-KWD-002o).
func (g *Game) hasCleaveTarget(att, def *Asset) bool {
	for _, sec := range g.Theater {
		if g.CanCleave(att, def, sec) {
			return true
		}
	}
	return false
}

// legalSecondary applies normal targeting rules to a Berserk secondary target:
// an opposing Main Theater asset the attacker's controller may choose. Ward
// blocks it; Ethereal and Guardian restrict only attacks (FR-KWD-002o).
func (g *Game) legalSecondary(att, sec *Asset) bool {
	return sec.Zone == ZoneMainTheater &&
		g.legalTarget(att.Controller, att.ID, ability.Filter{Allegiance: ability.AllegianceOpposing}, sec)
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
	// SVY-R8 (FRD 01 v3.8) compares with the attacker's current controller, not
	// the declaring player; board directive #29 brings this check in line.
	if def == nil || !inCombatZone(def) || def.Controller == it.Controller {
		g.logEvent(Event{Kind: "fizzle", Player: -1, Source: it.Source},
			"attack fizzles: the defender is gone, so there is no Berserk cleave")
		return
	}

	att := g.Asset(it.Source)
	if att == nil || !inCombatZone(att) {
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
	g.logEvent(Event{Kind: "combat", Player: -1, Source: att.ID, Target: def.ID},
		"%s (%d) and %s (%d) exchange combat damage", att, attPower, def, defPower)
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
