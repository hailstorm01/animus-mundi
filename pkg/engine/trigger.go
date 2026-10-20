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
	item       *StackItem // a ready-made stack item (keyword trigger); used instead of trig when set
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
		if t.Hook != hook || (t.Turn == ability.TurnYours && g.Active != a.Controller) {
			continue
		}
		g.pending = append(g.pending, pendingTrigger{
			source: a.ID, inst: a.Inst, controller: a.Controller, trig: t,
			label: fmt.Sprintf("%s: trigger %d", a.Printed().Name, i+1),
		})
	}
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
			if p.item != nil {
				g.push(p.item)
				continue
			}
			g.push(&StackItem{Kind: StackAbility, Controller: p.controller, Source: p.source, Inst: p.inst, Trigger: p.trig,
				Chosen: g.chooseTargets(p)})
		}
	}
}

// resolveAbility runs a triggered ability's steps in order. It resolves even if
// its source has left the battlefield: untargeted steps verify nothing.
func (g *Game) resolveAbility(it *StackItem) {
	// Second verification stage (FR-TGT-003): a chosen entity that is gone or
	// no longer legal is treated like an empty slot.
	chosen := make([]*Asset, len(it.Chosen))
	for i, id := range it.Chosen {
		if a := g.Asset(id); a != nil && g.legalTarget(it.Controller, it.Trigger.Effect.Targets[i].Filter, a) {
			chosen[i] = a
		}
	}
	for _, st := range it.Trigger.Effect.Steps {
		who := []PlayerID{it.Controller}
		if st.Subject == ability.SubjectOpponents {
			who = []PlayerID{it.Controller.Opponent()}
		}
		switch st.Op {
		case ability.OpDealDamage:
			if st.Subject != ability.SubjectTarget || st.Target < 0 || st.Target >= len(chosen) || chosen[st.Target] == nil {
				continue
			}
			g.dealDamage(g.Asset(it.Source), chosen[st.Target], st.Amount.N, false) // nil source if it left
		case ability.OpScratch:
			for _, pid := range who {
				g.scratch(pid, st.Amount.N)
			}
		case ability.OpDraw:
			for _, pid := range who {
				g.draw(pid, st.Amount.N)
			}
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

// chooseTargets asks p's controller to choose each of its Effect.Targets, in
// slot order, when the trigger is placed on the stack (FR-TGT-001). Only legal
// entities are offered, so a single one is not asked about and a timeout picks
// a random legal one. A slot with no legal entity stays 0: deploying never
// requires a target (FR-TGT-003).
func (g *Game) chooseTargets(p pendingTrigger) []int {
	targets := p.trig.Effect.Targets
	if len(targets) == 0 {
		return nil
	}
	chosen := make([]int, len(targets))
	for i, t := range targets {
		d := Decision{Player: p.controller, Kind: DecideTarget, Min: 1, Max: 1,
			Prompt: fmt.Sprintf("%s: choose target %d", p.label, i+1)}
		// Options are listed oldest asset first (by ID), not by zone.
		assets := g.Assets()
		slices.SortFunc(assets, func(x, y *Asset) int { return x.ID - y.ID })
		for _, a := range assets {
			if g.legalTarget(p.controller, t.Filter, a) {
				d.Options = append(d.Options, Option{Label: a.String(), Ref: a.ID})
			}
		}
		if picks := g.choose(d); len(picks) == 1 {
			chosen[i] = d.Options[picks[0]].Ref
		}
	}
	return chosen
}

// legalTarget reports whether battlefield asset a meets every criterion of f
// for an ability controlled by ctl. An opponent's Ward asset is never legal
// (FR-KWD-002h).
func (g *Game) legalTarget(ctl PlayerID, f ability.Filter, a *Asset) bool {
	switch f.Allegiance {
	case ability.AllegianceAllied:
		if a.Controller != ctl {
			return false
		}
	case ability.AllegianceOpposing:
		if a.Controller == ctl {
			return false
		}
	}
	if a.Controller != ctl && g.Keywords(a).Ward {
		return false
	}
	if f.Kinds != 0 {
		var k ability.Kinds
		switch a.Printed().Type {
		case TypeUnit:
			k = ability.KindUnit
		case TypeStructure:
			k = ability.KindStructure
		case TypeHeadquarters:
			k = ability.KindHeadquarters
		}
		if f.Kinds&k == 0 {
			return false
		}
	}
	return !(f.NonScratch && a.Printed().HasSubtype("Scratch Company"))
}
