// Targeting: choosing targets and checking them against filters (FR-TGT-001, FR-TGT-003).

package engine

import (
	"fmt"
	"slices"

	"animus-mundi/pkg/engine/ability"
)

// chooseTargets asks p's controller to choose each of its Effect.Targets, in
// slot order, when the trigger is placed on the stack (FR-TGT-001). Only legal
// entities are offered. A single-entity slot is not asked about with one
// option, and a timeout picks the first legal one in fixed order (SVY-R16). An "up to N" slot is always
// asked (Min 0 != Max), its picks are distinct and a timeout picks none
// (SVY-R11). A slot with no legal entity stays empty and asks nothing:
// deploying never requires a target (FR-TGT-003).
func (g *Game) chooseTargets(p pendingTrigger) [][]int {
	targets := p.trig.Effect.Targets
	if len(targets) == 0 {
		return nil
	}
	chosen := make([][]int, len(targets))
	for i, t := range targets {
		d := Decision{Player: p.controller, Kind: DecideTarget, Min: 1, Max: 1,
			Prompt: fmt.Sprintf("%s: choose target %d", p.label, i+1)}
		// Options are listed oldest asset first (by ID), not by zone.
		for _, a := range g.assetsByID() {
			if g.legalTarget(p.controller, p.source, t.Filter, a) {
				d.Options = append(d.Options, Option{Label: a.String(), Ref: a.ID})
			}
		}
		if t.UpTo > 0 {
			if len(d.Options) == 0 {
				continue
			}
			d.Prompt = fmt.Sprintf("%s: choose up to %d for target %d", p.label, t.UpTo, i+1)
			d.Min, d.Max = 0, min(t.UpTo, len(d.Options))
		}
		for _, k := range g.choose(d) {
			chosen[i] = append(chosen[i], d.Options[k].Ref)
		}
	}
	return chosen
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
	if !deployedTo(a, f.Zone) {
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

// assetsByID returns every battlefield asset, oldest (lowest ID) first.
func (g *Game) assetsByID() []*Asset {
	assets := g.Assets()
	slices.SortFunc(assets, func(x, y *Asset) int { return x.ID - y.ID })
	return assets
}
