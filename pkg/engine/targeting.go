// Targeting: choosing targets and checking them against filters (FR-TGT-001, FR-TGT-003).

package engine

import (
	"fmt"

	"animus-mundi/pkg/engine/ability"
)

// chooseTargets asks p's controller to choose each of its Effect.Targets, in
// slot order, when the trigger is placed on the stack (FR-TGT-001). Only legal
// entities are offered. A single-entity slot is not asked about with one
// option, and a timeout picks the first legal one in fixed order (SVY-R16).
// An "up to N" slot is always asked (Min 0 != Max), its picks are distinct
// and a timeout picks none (SVY-R11). A slot with no legal entity stays empty
// and asks nothing: deploying never requires a target (FR-TGT-003).
func (g *Game) chooseTargets(p pendingTrigger) [][]int {
	targets := p.trig.Effect.Targets
	if len(targets) == 0 {
		return nil
	}
	chosen := make([][]int, len(targets))
	for i, t := range targets {
		prompt := fmt.Sprintf("%s: choose target %d", p.label, i+1)
		if t.UpTo > 0 {
			prompt = fmt.Sprintf("%s: choose up to %d for target %d", p.label, t.UpTo, i+1)
		}
		d, _ := g.assetDecision(p.controller, DecideTarget, prompt, t.UpTo, func(a *Asset) bool {
			return g.legalTarget(p.controller, p.source, t.Filter, a)
		})
		if t.UpTo > 0 && len(d.Options) == 0 {
			continue
		}
		for _, k := range g.choose(d) {
			chosen[i] = append(chosen[i], d.Options[k].Ref)
		}
	}
	return chosen
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
// card is src: allegiance by the card's owner, Kinds and NonScratch by its
// printed characteristics (FR-ZON-004), Other not src. Zone must be ZoneDiscard.
func (g *Game) matchesCard(ctl PlayerID, src *Instance, f ability.Filter, inst *Instance) bool {
	if f.Zone != ability.ZoneDiscard || (f.Other && inst == src) {
		return false
	}
	switch f.Allegiance {
	case ability.AllegianceAllied:
		if inst.Owner != ctl {
			return false
		}
	case ability.AllegianceOpposing:
		if inst.Owner == ctl {
			return false
		}
	}
	if f.Kinds != 0 && f.Kinds&kindOf(inst.Card.Type) == 0 {
		return false
	}
	return !(f.NonScratch && inst.Card.HasSubtype("Scratch Company"))
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

// matchesFilter reports whether battlefield asset a meets every criterion of f
// (allegiance relative to ctl, kinds, NonScratch, Other: not the source asset
// src, Zone: currently in that battlefield zone). Targeting adds Ward on top (legalTarget); an untargeted "all" step uses
// this alone.
func (g *Game) matchesFilter(ctl PlayerID, src int, f ability.Filter, a *Asset) bool {
	if f.Other && a.ID == src {
		return false
	}
	if f.Zone == ability.ZoneDiscard || !deployedTo(a, f.Zone) {
		return false
	}
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
	if f.Kinds != 0 {
		if f.Kinds&kindOf(a.Printed().Type) == 0 {
			return false
		}
	}
	return !(f.NonScratch && a.Printed().HasSubtype("Scratch Company"))
}
