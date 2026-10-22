package sim

import "animus-mundi/pkg/engine"

// combatPair is the open combat exchange: a combat event, then its two
// Combat damage events (FR-CBT-002).
type combatPair struct {
	left             int  // damage events still due, 0 when no exchange is open
	att, def         int  // asset IDs
	attPow, defPow   int  // effective power in the last settled snapshot, -1 if not there
	attDone, defDone bool // the attacker-to-defender and defender-to-attacker events have come
}

// openCombat starts a combat exchange: the next two damage events are its
// Combat damage (FR-CBT-002).
func (m *Monitors) openCombat(e engine.Event) {
	m.count(monDamage, "FR-CBT-002")
	cp := combatPair{left: 2, att: e.Source, def: e.Target, attPow: -1, defPow: -1}
	if m.last != nil {
		if a := find(*m.last, e.Source); a != nil {
			cp.attPow = a.Power
		}
		if d := find(*m.last, e.Target); d != nil {
			cp.defPow = d.Power
		}
	}
	m.combat = cp
}

// checkCombatDamage matches a Combat damage event to the open exchange.
func (m *Monitors) checkCombatDamage(e engine.Event) {
	const mon = monDamage
	c := &m.combat
	if c.left == 0 {
		m.fail(mon, e.Seq, "combat damage from %d to %d outside a combat exchange (or a third one in it)", e.Source, e.Target)
		return
	}
	c.left--
	var want int
	switch {
	case e.Source == c.att && e.Target == c.def && !c.attDone:
		c.attDone, want = true, c.attPow
	case e.Source == c.def && e.Target == c.att && !c.defDone:
		c.defDone, want = true, c.defPow
	default:
		m.fail(mon, e.Seq, "combat damage from %d to %d is not the attacker %d / defender %d exchange", e.Source, e.Target, c.att, c.def)
		return
	}
	if want >= 0 && e.Raw != want {
		m.fail(mon, e.Seq, "combat damage from %d to %d has raw %d, want the dealer's effective power %d", e.Source, e.Target, e.Raw, want)
	}
}

// checkDamage checks one damage event against the five-stage sequence
// (FR-REP-004) using the target's printed characteristics.
func (m *Monitors) checkDamage(e engine.Event) {
	const mon = monDamage
	m.count(mon, "FR-REP-004")
	if e.Combat {
		m.checkCombatDamage(e)
	} else if m.combat.left > 0 {
		m.fail(mon, e.Seq, "non-combat damage from %d to %d inside a combat exchange", e.Source, e.Target)
		m.combat.left = 0
	}
	// The armor pool: the settled snapshot's, less what earlier damage absorbed.
	armorOK, wantArmor := false, 0
	if m.last != nil {
		if t := find(*m.last, e.Target); t != nil {
			armorOK, wantArmor = true, t.Armor-m.absorbed[e.Target]
		}
	}
	m.absorbed[e.Target] += e.Absorbed
	if armorOK && e.Armor != wantArmor {
		m.fail(mon, e.Seq, "damage to %d saw an Armor pool of %d, want %d from the settled state", e.Target, e.Armor, wantArmor)
	}

	check := func(s engine.Snapshot) bool {
		t := find(s, e.Target)
		if t == nil {
			return false
		}
		c := m.card(*t)
		if c == nil {
			m.fail(mon, e.Seq, "%s is not in its owner's decklist", t.Card)
			return true
		}
		kw := c.Keywords
		if kw.Resistant != 0 {
			m.count(mon, "FR-KWD-002b")
		}
		if kw.Stalwart {
			m.count(mon, "FR-KWD-002c")
		}
		if e.Armor > 0 {
			m.count(mon, "FR-KWD-002k")
		}
		r := max(0, e.Raw-kw.Resistant)
		capped := r
		if e.Combat && kw.Stalwart {
			capped = min(r, 1)
		}
		absorbed := min(capped, e.Armor)
		if e.Absorbed != absorbed || e.Amount != capped-absorbed {
			m.fail(mon, e.Seq, "damage %d (combat %t) to %s#%d (Resistant %d, Stalwart %t, Armor %d) absorbed %d and committed %d, want %d and %d",
				e.Raw, e.Combat, t.Card, t.ID, kw.Resistant, kw.Stalwart, e.Armor, e.Absorbed, e.Amount, absorbed, capped-absorbed)
		}
		return true
	}
	if m.last != nil && check(*m.last) {
		return
	}
	m.pending = append(m.pending, pendingCheck{seq: e.Seq, run: func(after engine.Snapshot, _ bool) { check(after) }})
}
