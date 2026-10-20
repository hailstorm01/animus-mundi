package engine

import (
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
	trig       ability.Trigger
}

// fireDeploy buffers a's own deploy triggers (self hook only). The controller
// is a's controller now. They wait for the flush after the item's sweep.
func (g *Game) fireDeploy(a *Asset) { g.fireHook(a, ability.HookDeploy) }

// fireKilled buffers a's own death triggers. The deathSweep calls it before a
// leaves the battlefield, so the controller is a's last-known controller. A
// face-down asset fires nothing (FR-ZON-004).
func (g *Game) fireKilled(a *Asset) { g.fireHook(a, ability.HookKilled) }

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
		if t.Hook != hook {
			continue
		}
		g.pending = append(g.pending, pendingTrigger{
			source: a.ID, inst: a.Inst, controller: a.Controller, trig: t,
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
			g.push(&StackItem{Kind: StackAbility, Controller: p.controller, Source: p.source, Inst: p.inst, Trigger: p.trig})
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
		case ability.OpReturnToHand, ability.OpPutOnDeck:
			g.relocateSelf(it, st)
		}
	}
}

// relocateSelf moves a death trigger's own card from its owner's discard pile
// to the owner's hand or deck. It tracks the exact instance: if that instance
// is no longer in the discard pile, it does nothing (FR-ZON-004). Only defined
// for SubjectSelf on a HookKilled trigger.
func (g *Game) relocateSelf(it *StackItem, st ability.Step) {
	if st.Subject != ability.SubjectSelf || it.Trigger.Hook != ability.HookKilled || it.Inst == nil {
		return
	}
	owner := g.Players[it.Inst.Owner]
	if !slices.Contains(owner.Discard, it.Inst) {
		return
	}
	owner.Discard = remove(owner.Discard, it.Inst)
	switch {
	case st.Op == ability.OpReturnToHand:
		owner.Hand = append(owner.Hand, it.Inst)
	case st.Bottom:
		owner.Deck = append(owner.Deck, it.Inst)
	default:
		owner.Deck = append([]*Instance{it.Inst}, owner.Deck...)
	}
}
