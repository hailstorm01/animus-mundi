// Triggers: the pending-trigger buffer, hook firing and the flush in APNAP order (FR-TRG-002 to FR-TRG-004).

package engine

import (
	"cmp"
	"fmt"
	"slices"

	"animus-mundi/pkg/engine/ability"
)

// pendingTrigger is a fired trigger waiting in the buffer for the flush.
type pendingTrigger struct {
	source     int       // source asset ID
	inst       *Instance // source card instance, tracked by identity (FR-ZON-004)
	controller PlayerID
	label      string
	pos        sourcePos // where the source stood when it fired
	trig       ability.Trigger
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

// fireCombatKill buffers att's own combat-kill triggers. resolveAttack calls it after
// the declared attacker's combat damage left the defender lethal.
func (g *Game) fireCombatKill(att *Asset) { g.fireHook(att, ability.HookCombatKill) }

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
			source: a.ID, inst: a.Inst, controller: a.Controller, trig: t,
			label: fmt.Sprintf("%s: trigger %d", a.Printed().Name, i+1), pos: posOf(a),
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
			source: a.ID, inst: a.Inst, controller: a.Controller,
			trig:  ability.Aggressive(n),
			label: fmt.Sprintf("%s: Aggressive %d", a.Printed().Name, n), pos: posOf(a),
		})
	}
	g.fireHook(a, ability.HookAttacks)
}

// flushTriggers places the buffer on the stack: the active player's triggers
// first, then the non-active player's on top (FR-TRG-003). Each player orders
// their own triggers; a single trigger is not asked about. An optional trigger
// is asked about in its turn to be placed, before its modes and targets; declined, it is
// not placed (SVY-R9). Each player's triggers are listed, and by default placed,
// in source order (sourcePos). A match that ended discards the buffer.
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
			if len(p.trig.Effect.Modes) > 0 {
				p.trig.Effect = g.chooseModes(p)
			}
			g.push(&StackItem{Kind: StackAbility, Controller: p.controller, Source: p.source, Inst: p.inst,
				Effect: p.trig.Effect, Tracked: p.trig.Hook == ability.HookKilled, Keyword: p.trig.Keyword,
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

// chooseModes asks p's controller which modes of a modal trigger to use, when
// it is placed and before its targets (FR-MOD-001 §5, SVY-R3), and returns the
// effect the item resolves: the chosen modes joined in printed order. A mode
// is legal if each of its exactly-one target slots has a legal entity now; an
// "up to N" slot never makes it illegal. Only legal modes are offered, so a
// timeout takes the first listed legal ones (§6) and a single legal mode is
// answered automatically (§1). With no legal mode nothing is asked and the
// effect is empty. With fewer legal modes than ChooseN all of them are chosen.
func (g *Game) chooseModes(p pendingTrigger) ability.Effect {
	var legal []int
	for k, m := range p.trig.Effect.Modes {
		ok := true
		for i, t := range m.Targets {
			if t.UpTo == 0 && len(g.targetDecision(p, i, t).Options) == 0 {
				ok = false
				break
			}
		}
		if ok {
			legal = append(legal, k)
		}
	}
	if len(legal) == 0 {
		return ability.Effect{}
	}
	n := min(p.trig.Effect.ChooseN, len(legal))
	prompt := fmt.Sprintf("%s: choose %d", p.label, p.trig.Effect.ChooseN)
	if p.trig.Effect.ChooseN == 1 {
		prompt = p.label + ": choose one"
	}
	d := Decision{Player: p.controller, Kind: DecideMode, Prompt: prompt, Min: n, Max: n}
	for _, k := range legal {
		d.Options = append(d.Options, Option{Label: fmt.Sprintf("mode %d", k+1), Ref: p.source})
	}
	picked := g.choose(d)
	slices.Sort(picked)
	// New slices throughout: a mode's own Targets and Steps belong to the
	// card's definition and must never be appended to.
	var joined ability.Effect
	for _, j := range picked {
		m := p.trig.Effect.Modes[legal[j]]
		shift := len(joined.Targets)
		joined.Targets = append(joined.Targets, m.Targets...)
		for _, st := range m.Steps {
			if st.Subject == ability.SubjectTarget {
				st.Target += shift
			}
			joined.Steps = append(joined.Steps, st)
		}
	}
	return joined
}
