package engine

import (
	"fmt"

	"animus-mundi/pkg/engine/ability"
)

// pendingTrigger is a fired trigger waiting in the buffer for the flush.
type pendingTrigger struct {
	source     int // source asset ID
	controller PlayerID
	label      string
	trig       ability.Trigger
}

// fireDeploy buffers a's own deploy triggers (self hook only). The controller
// is a's controller now. They wait for the flush after the item's sweep.
func (g *Game) fireDeploy(a *Asset) {
	if a.FaceDown {
		return
	}
	def := a.Printed().Ability
	if def == nil {
		return
	}
	for i, t := range def.Triggers {
		if t.Hook != ability.HookDeploy {
			continue
		}
		g.pending = append(g.pending, pendingTrigger{
			source: a.ID, controller: a.Controller, trig: t,
			label: fmt.Sprintf("%s: trigger %d", a.Printed().Name, i+1),
		})
	}
}

// flushTriggers places the buffer on the stack: the active player's triggers
// first, then the non-active player's on top (FR-TRG-003). Each player orders
// their own triggers; a single trigger is not asked about. A match that ended
// discards the buffer.
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
			g.push(&StackItem{Kind: StackAbility, Controller: p.controller, Source: p.source, Trigger: p.trig})
		}
	}
}

// resolveAbility runs a triggered ability's steps in order. It resolves even if
// its source has left the battlefield: untargeted steps verify nothing.
func (g *Game) resolveAbility(it *StackItem) {
	for _, st := range it.Trigger.Effect.Steps {
		switch st.Op {
		case ability.OpScratch:
			g.scratch(it.Controller, st.Amount.N)
		case ability.OpDraw:
			g.draw(it.Controller, st.Amount.N)
		}
	}
}
