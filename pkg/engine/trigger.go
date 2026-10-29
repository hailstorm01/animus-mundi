// Triggers: the pending-trigger buffer, hook firing and the flush in APNAP order (FR-TRG-002 to FR-TRG-004). Placing each one is placement.go's.

package engine

import (
	"cmp"
	"fmt"
	"slices"

	"animus-mundi/pkg/engine/ability"
)

// pendingTrigger is a fired trigger waiting in the buffer for the flush.
type pendingTrigger struct {
	placement
	pos  sourcePos // where the source stood when it fired
	trig ability.Trigger
	// awaits is an asset this trigger waits on: placed only if destroyAll destroys
	// it before the flush (SVY-R19); 0 for none.
	awaits int
}

// sourcePos is where a trigger's source stood when it fired: its zone's rank
// (Headquarters 0, Reserves 1, Main Theater 2) and its arrival number. A
// player's simultaneous triggers fall back to this order (SVY-R7, FR-MOD-001 §6).
type sourcePos struct{ zone, arrived int }

func posOf(a *Asset) sourcePos {
	rank := 0
	switch a.Zone {
	case ZoneReserves:
		rank = 1
	case ZoneMainTheater:
		rank = 2
	}
	return sourcePos{zone: rank, arrived: a.arrived}
}

func (p sourcePos) compare(q sourcePos) int {
	return cmp.Or(cmp.Compare(p.zone, q.zone), cmp.Compare(p.arrived, q.arrived))
}

// fireDeploy buffers a's own deploy triggers (self hook only). The controller
// is a's controller now. They wait for the flush after the item's sweep.
func (g *Game) fireDeploy(a *Asset) { g.fireHook(a, ability.HookDeploy) }

// fireKilled buffers a's own death triggers. The deathSweep calls it before a
// leaves the battlefield, so the controller is a's last-known controller. A
// face-down asset fires nothing (FR-ZON-004).
func (g *Game) fireKilled(a *Asset) { g.fireHook(a, ability.HookKilled) }

// fireDamaged buffers a's own damage-taken triggers (Unbowed). dealDamage calls it
// when a took damage, so the controller is a's controller now, kept if a then dies.
func (g *Game) fireDamaged(a *Asset) { g.fireHook(a, ability.HookDamaged) }

// fireExhausted buffers a's own exhaust triggers. exhaust calls it after a becomes Exhausted.
func (g *Game) fireExhausted(a *Asset) { g.fireHook(a, ability.HookExhausted) }

// fireCombatKill buffers att's own combat-kill triggers. resolveAttack calls it when
// the attacker's combat damage left def lethal; the triggers wait on def's
// destruction, because a replaced death is not a kill (SVY-R19).
func (g *Game) fireCombatKill(att, def *Asset) {
	n := len(g.pending)
	g.fireHook(att, ability.HookCombatKill)
	for i := n; i < len(g.pending); i++ {
		g.pending[i].awaits = def.ID
	}
}

// confirmDestroyed clears the wait on asset id for every buffered trigger that awaits it.
func (g *Game) confirmDestroyed(id int) {
	for i := range g.pending {
		if g.pending[i].awaits == id {
			g.pending[i].awaits = 0
		}
	}
}

// fireHook buffers a's own triggers for hook in printed order. A face-down
// asset has no printed abilities.
func (g *Game) fireHook(a *Asset, hook ability.Hook) {
	if a.FaceDown {
		return
	}
	def := a.Printed().Ability
	if def == nil {
		return
	}
	for i, t := range def.Triggers {
		if t.Hook != hook || (t.Turn == ability.TurnYours && g.Active != a.Controller) {
			continue
		}
		if hook == ability.HookDeploy && !deployedTo(a, t.ToZone) {
			continue
		}
		// Cornered is an enqueue gate: checked here, when the event happens,
		// and before the MaxPerTurn counter (FR-KWD-002e, FR-LAY-003).
		if t.Cornered && !g.cornered(a.Controller) {
			continue
		}
		if t.MaxPerTurn > 0 {
			k := firingKey{asset: a.ID, trigger: i, turn: g.Turn}
			if g.firings[k] >= t.MaxPerTurn {
				continue
			}
			if g.firings == nil {
				g.firings = map[firingKey]int{}
			}
			g.firings[k]++
		}
		g.pending = append(g.pending, pendingTrigger{
			placement: placement{source: a.ID, inst: a.Inst, controller: a.Controller,
				label: fmt.Sprintf("%s: trigger %d", a.Printed().Name, i+1)},
			trig: t, pos: posOf(a),
		})
	}
}

// cornered reports whether p is Cornered: the opponent controls strictly more
// battlefield assets (units and structures) in the Main Theater than p does.
// Headquarters are never in the Main Theater (FR-KWD-002e).
func (g *Game) cornered(p PlayerID) bool {
	return len(g.TheaterOf(p.Opponent())) > len(g.TheaterOf(p))
}

// deployedTo reports whether a is in zone z now, which for a deploy trigger is
// the zone its deploy put it in. ZoneAny always holds.
func deployedTo(a *Asset, z ability.Zone) bool {
	switch z {
	case ability.ZoneReserves:
		return a.Zone == ZoneReserves
	case ability.ZoneMainTheater:
		return a.Zone == ZoneMainTheater
	}
	return true
}

// fireAttacks buffers a's Step 2 triggers: its Aggressive keyword trigger (an
// ability.Aggressive trigger, read now, not at card load), if any, then its
// printed attack triggers (FR-CBT-002).
func (g *Game) fireAttacks(a *Asset) {
	if n := g.Keywords(a).Aggressive; n > 0 {
		g.pending = append(g.pending, pendingTrigger{
			placement: placement{source: a.ID, inst: a.Inst, controller: a.Controller,
				label: fmt.Sprintf("%s: Aggressive %d", a.Printed().Name, n)},
			trig: ability.Aggressive(n), pos: posOf(a),
		})
	}
	g.fireHook(a, ability.HookAttacks)
}

// flushTriggers places the buffer on the stack: the active player's triggers
// first, then the non-active player's on top (FR-TRG-003). Each player orders
// their own triggers; a single trigger is not asked about. An optional trigger
// is asked about in its turn to be placed, before its modes and targets; declined, it is
// not placed (SVY-R9). A trigger still waiting on an asset's destruction is dropped,
// because the sweep did not destroy it (SVY-R19). Each player's triggers are listed, and by default placed,
// in source order (sourcePos). A match that ended discards the buffer.
func (g *Game) flushTriggers() {
	buf := g.pending
	g.pending = nil
	if g.Outcome.Over {
		return
	}
	buf = slices.DeleteFunc(buf, func(p pendingTrigger) bool { return p.awaits != 0 })
	for _, pid := range []PlayerID{g.Active, g.Active.Opponent()} {
		var mine []pendingTrigger
		for _, p := range buf {
			if p.controller == pid {
				mine = append(mine, p)
			}
		}
		slices.SortStableFunc(mine, func(x, y pendingTrigger) int { return x.pos.compare(y.pos) })
		order := make([]int, len(mine))
		for i := range order {
			order[i] = i
		}
		if len(mine) >= 2 {
			d := Decision{Player: pid, Kind: DecideOrder, Prompt: "order your triggers (first pick is placed first and resolves last)", Min: len(mine), Max: len(mine)}
			for _, p := range mine {
				d.Options = append(d.Options, Option{Label: p.label, Ref: p.source})
			}
			order = g.choose(d)
		}
		for _, i := range order {
			p := mine[i]
			if p.trig.Optional && !g.accepts(p) {
				continue
			}
			eff, chosen := g.chooseOnPlacement(p.placement, p.trig.Effect)
			g.push(&StackItem{Kind: StackAbility, Controller: p.controller, Source: p.source, Inst: p.inst,
				Effect: eff, Tracked: p.trig.Hook == ability.HookKilled, Keyword: p.trig.Keyword,
				Chosen: chosen})
		}
	}
}

// accepts asks p's controller the "you may" question of an optional trigger
// (SVY-R9). A decline, a timeout, no chooser or an invalid answer all decline
// (FR-MOD-001 §6).
func (g *Game) accepts(p pendingTrigger) bool {
	d := Decision{Player: p.controller, Kind: DecideOptional, Min: 0, Max: 1,
		Prompt:  fmt.Sprintf("%s: you may use it", p.label),
		Options: []Option{{Label: "yes", Ref: p.source}}}
	return len(g.choose(d)) == 1
}
