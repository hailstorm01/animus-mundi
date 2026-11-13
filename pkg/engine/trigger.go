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
// (Headquarters 0, Reserves 1, Main Theater 2, a discard pile 3: after every
// battlefield zone, and arrived is then the card's position in that pile; SVY-R7,
// directive #136) and its arrival number. A
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

// triggerSource is what fires a trigger, read when it fires: the face it
// prints, its controller, its zone and its rank in the flush's source order,
// and the IDs its placement and MaxPerTurn counter use. A source is a battlefield
// asset (assetSource) or a card in its owner's discard pile (pileSource); id is 0
// for the latter.
type triggerSource struct {
	card       *Card     // its printed face (the Scratch face if face-down)
	id         int       // source asset ID: the placement's source and the MaxPerTurn key
	inst       *Instance // source card instance (FR-ZON-004)
	controller PlayerID
	zone       Zone
	pos        sourcePos
}

// assetSource is a as a trigger source, as it stands now.
func assetSource(a *Asset) triggerSource {
	return triggerSource{card: a.Printed(), id: a.ID, inst: a.Inst, controller: a.Controller, zone: a.Zone, pos: posOf(a)}
}

// pileSource is inst, a card in its owner's discard pile, as a trigger source: its
// printed face (FR-ZON-004), no asset, controlled by its owner and ranked after
// every battlefield asset, by its position in the pile (SVY-R7, directive #136).
func pileSource(g *Game, inst *Instance) triggerSource {
	return triggerSource{card: inst.Card, inst: inst, controller: inst.Owner, zone: ZoneDiscard,
		pos: sourcePos{zone: 3, arrived: slices.Index(g.Players[inst.Owner].Discard, inst)}}
}

// activeIn reports whether a trigger with ActiveZone z works from a source in
// zone src: ZoneDiscard only from a discard pile, ZoneAny only from the battlefield.
func activeIn(src Zone, z ability.Zone) bool {
	if z == ability.ZoneDiscard {
		return src == ZoneDiscard
	}
	return src != ZoneDiscard
}

// fireDiscarded buffers inst's discarded-from-hand triggers (FR-ZON-004).
// discardCards calls it after inst is in its owner's discard pile.
func (g *Game) fireDiscarded(inst *Instance) {
	g.fireHook(pileSource(g, inst), ability.HookDiscardedFromHand)
}

// fireDeploy buffers a's own deploy triggers (self hook only). The controller
// is a's controller now. They wait for the flush after the item's sweep.
func (g *Game) fireDeploy(a *Asset) { g.fireHook(assetSource(a), ability.HookDeploy) }

// fireMoved buffers a's own move triggers (self hook only). moveOnBattlefield
// calls it after a is placed in its new zone (SVY-R1, FR-ZON-002).
// Each trigger locks the zone a moved to, its own value, fixed by the event (SVY-R45).
func (g *Game) fireMoved(a *Asset) {
	added := g.fireHook(assetSource(a), ability.HookMoves)
	for i := range added {
		added[i].locked = &LockedZone{Zone: a.Zone, Player: a.Controller}
	}
}

// fireKilled buffers a's own death triggers. The deathSweep calls it before a
// leaves the battlefield, so the controller is a's last-known controller. A
// face-down asset fires nothing (FR-ZON-004).
func (g *Game) fireKilled(a *Asset) { g.fireHook(assetSource(a), ability.HookKilled) }

// fireDamaged buffers a's own damage-taken triggers (Unbowed). dealDamage calls it
// when a took damage, so the controller is a's controller now, kept if a then dies.
func (g *Game) fireDamaged(a *Asset) { g.fireHook(assetSource(a), ability.HookDamaged) }

// fireExhausted buffers a's own exhaust triggers. exhaust calls it after a becomes Exhausted.
func (g *Game) fireExhausted(a *Asset) { g.fireHook(assetSource(a), ability.HookExhausted) }

// fireCombatKill buffers att's own combat-kill triggers. resolveAttack calls it when
// the attacker's combat damage left def lethal; the triggers wait on def's
// destruction, because a replaced death is not a kill (SVY-R19).
func (g *Game) fireCombatKill(att, def *Asset) {
	added := g.fireHook(assetSource(att), ability.HookCombatKill)
	for i := range added {
		added[i].awaits = def.ID
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

// fireHook buffers src's own triggers for hook in printed order: the one place
// a trigger's gates are checked. A face-down asset has no printed abilities.
// It returns the entries it added, a slice of the buffer, so a caller can set a
// field on each before the flush; the other callers ignore it.
func (g *Game) fireHook(src triggerSource, hook ability.Hook) []pendingTrigger {
	n := len(g.pending)
	def := src.card.Ability
	if def == nil {
		return nil
	}
	for i, t := range def.Triggers {
		if t.Hook != hook || !activeIn(src.zone, t.ActiveZone) || (t.Turn == ability.TurnYours && g.Active != src.controller) {
			continue
		}
		if (hook == ability.HookDeploy || hook == ability.HookMoves) && !deployedTo(src.zone, t.ToZone) {
			continue
		}
		// Cornered is an enqueue gate: checked here, when the event happens,
		// and before the MaxPerTurn counter (FR-KWD-002e, FR-LAY-003).
		if t.Cornered && !g.cornered(src.controller) {
			continue
		}
		if t.MaxPerTurn > 0 {
			k := firingKey{asset: src.id, trigger: i, turn: g.Turn}
			if g.firings[k] >= t.MaxPerTurn {
				continue
			}
			if g.firings == nil {
				g.firings = map[firingKey]int{}
			}
			g.firings[k]++
		}
		g.bufferTrigger(src, fmt.Sprintf("%s: trigger %d", src.card.Name, i+1), t)
	}
	return g.pending[n:]
}

// bufferTrigger adds the fired trigger t to the buffer: src is its source, src's
// controller is its controller, and label names it in the flush's prompts.
func (g *Game) bufferTrigger(src triggerSource, label string, t ability.Trigger) {
	g.pending = append(g.pending, pendingTrigger{
		placement: placement{source: src.id, inst: src.inst, controller: src.controller, label: label},
		trig:      t,
		pos:       src.pos,
	})
}

// fireAttacks buffers a's Step 2 triggers: its Aggressive keyword trigger (an
// ability.Aggressive trigger, read now, not at card load), if any, then its
// printed attack triggers (FR-CBT-002).
func (g *Game) fireAttacks(a *Asset) {
	src := assetSource(a)
	if n := g.Keywords(a).Aggressive; n > 0 {
		g.bufferTrigger(src, fmt.Sprintf("%s: %s %d", src.card.Name, ability.KeywordAggressive.String(), n), ability.Aggressive(n))
	}
	g.fireHook(src, ability.HookAttacks)
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
	for _, pid := range g.apnap() {
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
			item := p.stackItem(StackAbility)
			item.Effect, item.Chosen = g.chooseOnPlacement(p.placement, p.trig.Effect)
			item.Tracked, item.Keyword = p.trig.Hook == ability.HookKilled || p.trig.ActiveZone == ability.ZoneDiscard, p.trig.Keyword
			g.push(item)
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
