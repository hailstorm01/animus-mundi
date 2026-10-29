// Targeting: choosing targets on placement, untargeted picks on resolution, and checking entities against filters (FR-TGT-001, FR-TGT-003, SVY-R14).

package engine

import (
	"fmt"

	"animus-mundi/pkg/engine/ability"
)

// chooseTargets asks pl's controller to choose each of e's Targets, in
// slot order, when the effect is placed on the stack (FR-TGT-001). Only legal
// entities are offered. A single-entity slot is not asked about with one
// option, and a timeout picks the first legal one in fixed order (SVY-R16).
// An "up to N" slot is always asked (Min 0 != Max), its picks are distinct
// and a timeout picks none (SVY-R11). A slot with no legal entity stays empty
// and asks nothing: deploying never requires a target (FR-TGT-003). A slot
// whose filter names ZoneDiscard chooses cards in discard piles, and its Refs
// are card instance IDs.
func (g *Game) chooseTargets(pl placement, e ability.Effect) [][]int {
	targets := e.Targets
	if len(targets) == 0 {
		return nil
	}
	chosen := make([][]int, len(targets))
	for i, t := range targets {
		d := g.targetDecision(pl, i, t)
		if t.UpTo > 0 && len(d.Options) == 0 {
			continue
		}
		for _, k := range g.choose(d) {
			chosen[i] = append(chosen[i], d.Options[k].Ref)
		}
	}
	return chosen
}

// targetDecision builds the decision for target slot i (clause t) of the
// effect pl places: the legal entities of the slot, as pileDecision or
// assetDecision list them. chooseTargets asks it, and legalModes reads its
// Options (FR-MOD-001 §5), so both use the same filters.
func (g *Game) targetDecision(pl placement, i int, t ability.Target) Decision {
	prompt := fmt.Sprintf("%s: choose target %d", pl.label, i+1)
	if t.UpTo > 0 {
		prompt = fmt.Sprintf("%s: choose up to %d for target %d", pl.label, t.UpTo, i+1)
	}
	var d Decision
	if t.Filter.InPile() {
		d, _ = g.pileDecision(pl.controller, DecideTarget, prompt, t.UpTo, func(inst *Instance) bool {
			return g.matchesCard(pl.controller, pl.inst, t.Filter, inst)
		})
	} else {
		d, _ = g.assetDecision(pl.controller, DecideTarget, prompt, t.UpTo, func(a *Asset) bool {
			return g.legalTarget(pl.controller, pl.source, t.Filter, a)
		})
	}
	return d
}

// assetDecision builds a decision of the given kind (DecideTarget for targets,
// DecidePick for untargeted picks) for player p over the battlefield
// assets for which ok holds, and returns it with those assets in option order.
// Options are listed oldest asset first (by ID), not by zone (SVY-R11). With
// upTo 0 exactly one is picked (Min = Max = 1); with upTo N up to N distinct
// assets are, possibly none (Min 0, Max min(N, len)). Callers handle an empty
// option list themselves.
func (g *Game) assetDecision(p PlayerID, kind DecisionKind, prompt string, upTo int, ok func(*Asset) bool) (Decision, []*Asset) {
	d := Decision{Player: p, Kind: kind, Prompt: prompt}
	var as []*Asset
	for _, a := range g.assetsByID() {
		if ok(a) {
			as = append(as, a)
			d.Options = append(d.Options, Option{Label: a.String(), Ref: a.ID})
		}
	}
	d.Min, d.Max = pickBounds(upTo, len(as))
	return d, as
}

// pickBounds is the bounds rule of a pick over n options: with upTo 0 exactly
// one is picked (1, 1); with upTo N up to N distinct ones are, possibly none
// (0, min(N, n)). Asset and pile decisions share it.
func pickBounds(upTo, n int) (lo, hi int) {
	if upTo > 0 {
		return 0, min(upTo, n)
	}
	return 1, 1
}

// pileDecision builds a decision of the given kind for player p over the cards
// in discard piles for which ok holds, and returns it with those cards in
// option order: p's pile first, then the opponent's, each from the bottom (the
// order of Player.Discard). A discard pile is public, so the decision is not
// Hidden. Bounds are pickBounds'. Callers handle an empty option list.
func (g *Game) pileDecision(p PlayerID, kind DecisionKind, prompt string, upTo int, ok func(*Instance) bool) (Decision, []*Instance) {
	d := Decision{Player: p, Kind: kind, Prompt: prompt}
	var cs []*Instance
	for _, id := range []PlayerID{p, p.Opponent()} {
		for _, inst := range g.Players[id].Discard {
			if ok(inst) {
				cs = append(cs, inst)
				d.Options = append(d.Options, Option{Label: inst.Card.Name, Ref: inst.ID})
			}
		}
	}
	d.Min, d.Max = pickBounds(upTo, len(cs))
	return d, cs
}

// pickAssets is the untargeted pick: the item's controller picks from every
// battlefield asset matching st.Filter now, oldest first. It is not a target, so
// Ward excludes nothing (SVY-R3, SVY-R14). It asks a DecidePick, which has the
// same FR-MOD-001 section 6 fallback as a target: UpTo 0 picks exactly one (a single option is not
// asked, a timeout picks the first in fixed order); UpTo N picks up to N, always
// asked, and a timeout picks none. With no match nothing is asked.
func (g *Game) pickAssets(r *resolution, st ability.Step) []*Asset {
	d, opts := g.assetDecision(r.it.Controller, DecidePick, pickPrompt(r.it, st), st.UpTo, func(a *Asset) bool {
		return g.matchesFilter(r.it.Controller, r.it.Source, st.Filter, a)
	})
	return pickFrom(g, d, opts)
}

// pickPrompt is the prompt of an untargeted pick step.
func pickPrompt(it *StackItem, st ability.Step) string {
	name := cardName(it.Inst)
	if st.UpTo > 0 {
		return fmt.Sprintf("%s: pick up to %d", name, st.UpTo)
	}
	return name + ": pick one"
}

// pickCards is the untargeted pick over discard piles: the item's controller
// picks from every card matching st.Filter now, as pickAssets does for assets
// (SVY-R3, SVY-R14). The options are fixed before any card moves.
func (g *Game) pickCards(r *resolution, st ability.Step) []*Instance {
	d, opts := g.pileDecision(r.it.Controller, DecidePick, pickPrompt(r.it, st), st.UpTo, func(inst *Instance) bool {
		return g.matchesCard(r.it.Controller, r.it.Inst, st.Filter, inst)
	})
	return pickFrom(g, d, opts)
}

// pickFrom asks d and returns the options picked, in pick order: the body of
// every untargeted pick, over assets or cards. opts are d's options, in order
// (as assetDecision and pileDecision return them). With no option nothing is
// asked.
func pickFrom[T any](g *Game, d Decision, opts []T) []T {
	if len(opts) == 0 {
		return nil
	}
	var out []T
	for _, k := range g.choose(d) {
		out = append(out, opts[k])
	}
	return out
}

// pileTarget returns the card with instance ID id if it is still in a discard
// pile (DiscardCard) and still meets the pile filter f (FR-TGT-003), else nil.
// A card goes only to its owner's pile, so this is also "still in the zone the
// filter names". A card in a pile has no Ward, so only matchesCard is checked.
func (g *Game) pileTarget(ctl PlayerID, src *Instance, f ability.Filter, id int) *Instance {
	if inst := g.DiscardCard(id); inst != nil && g.matchesCard(ctl, src, f, inst) {
		return inst
	}
	return nil
}

// kindOf maps a printed card type to its filter kind.
func kindOf(t CardType) ability.Kinds {
	switch t {
	case TypeUnit:
		return ability.KindUnit
	case TypeStructure:
		return ability.KindStructure
	case TypeHeadquarters:
		return ability.KindHeadquarters
	}
	return 0
}

// matchesCard reports whether card inst, in a discard pile, meets every
// criterion of the pile filter f for an ability controlled by ctl whose own
// card is src: allegiance by the card's owner, Kinds, Scratch and NonScratch by its
// printed characteristics (FR-ZON-004), Other not src. f must be InPile.
func (g *Game) matchesCard(ctl PlayerID, src *Instance, f ability.Filter, inst *Instance) bool {
	if !f.InPile() || (f.Other && inst == src) {
		return false
	}
	return matchesPrinted(ctl, f, inst.Owner, inst.Card)
}

// legalTarget reports whether battlefield asset a meets every criterion of f
// for an ability controlled by ctl whose source is asset src. An opponent's
// Ward asset is never legal (FR-KWD-002h).
func (g *Game) legalTarget(ctl PlayerID, src int, f ability.Filter, a *Asset) bool {
	if a.Controller != ctl && g.Keywords(a).Ward {
		return false
	}
	return g.matchesFilter(ctl, src, f, a)
}

// matchesPrinted is the card-level part of a filter, shared by assets and
// cards in piles: Allegiance by side, the player the entity belongs to (relative
// to ctl), then Kinds, Scratch and NonScratch by the card's printed characteristics.
func matchesPrinted(ctl PlayerID, f ability.Filter, side PlayerID, c *Card) bool {
	switch f.Allegiance {
	case ability.AllegianceAllied:
		if side != ctl {
			return false
		}
	case ability.AllegianceOpposing:
		if side == ctl {
			return false
		}
	}
	if f.Kinds != 0 && f.Kinds&kindOf(c.Type) == 0 {
		return false
	}
	scratch := c.HasSubtype("Scratch Company")
	return !(f.NonScratch && scratch) && !(f.Scratch && !scratch)
}

// matchesFilter reports whether battlefield asset a meets every criterion of f
// (allegiance relative to ctl, kinds, Scratch and NonScratch, Other: not the source asset
// src, Zone: currently in that battlefield zone). Targeting adds Ward on top (legalTarget); an untargeted "all" step uses
// this alone.
func (g *Game) matchesFilter(ctl PlayerID, src int, f ability.Filter, a *Asset) bool {
	if f.Other && a.ID == src {
		return false
	}
	if f.InPile() || !deployedTo(a, f.Zone) {
		return false
	}
	return matchesPrinted(ctl, f, a.Controller, a.Printed())
}
