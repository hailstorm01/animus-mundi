// Triggers: the pending-trigger buffer, hook firing and the flush in APNAP order (FR-TRG-002 to FR-TRG-004).

package engine

import (
	"fmt"

	"animus-mundi/pkg/engine/ability"
)

// pendingTrigger is a fired trigger waiting in the buffer for the flush.
type pendingTrigger struct {
	source     int       // source asset ID
	inst       *Instance // source card instance, tracked by identity (FR-ZON-004)
	controller PlayerID
	label      string
	trig       ability.Trigger
	item       *StackItem // a ready-made stack item (keyword trigger); used instead of trig when set
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
		g.pending = append(g.pending, pendingTrigger{
			source: a.ID, inst: a.Inst, controller: a.Controller, trig: t,
			label: fmt.Sprintf("%s: trigger %d", a.Printed().Name, i+1),
		})
	}
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

// fireAttacks buffers a's Step 2 triggers: its Aggressive keyword trigger, if
// any, then its printed attack triggers (FR-CBT-002).
func (g *Game) fireAttacks(a *Asset) {
	if n := g.Keywords(a).Aggressive; n > 0 {
		g.pending = append(g.pending, pendingTrigger{
			source: a.ID, inst: a.Inst, controller: a.Controller,
			label: fmt.Sprintf("%s: Aggressive %d", a.Printed().Name, n),
			item:  &StackItem{Kind: StackAggressive, Controller: a.Controller, Source: a.ID, Amount: n},
		})
	}
	g.fireHook(a, ability.HookAttacks)
}

// flushTriggers places the buffer on the stack: the active player's triggers
// first, then the non-active player's on top (FR-TRG-003). Each player orders
// their own triggers; a single trigger is not asked about. An optional trigger
// is asked about in its turn to be placed, before its targets; declined, it is
// not placed (SVY-R9). A match that ended discards the buffer.
func (g *Game) flushTriggers() {
	buf := g.pending
	g.pending = nil
	if g.Outcome.Over {
		return
	}
	for _, pid := range []PlayerID{g.Active, g.Active.Opponent()} {
		var mine []pendingTrigger
		for _, p := range buf {
			if p.controller == pid {
				mine = append(mine, p)
			}
		}
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
			if p.item != nil {
				g.push(p.item)
				continue
			}
			if p.trig.Optional && !g.accepts(p) {
				continue
			}
			g.push(&StackItem{Kind: StackAbility, Controller: p.controller, Source: p.source, Inst: p.inst, Trigger: p.trig,
				Chosen: g.chooseTargets(p)})
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
