// Costs: reading and paying a card's cost, and checking and paying a placement's additional costs (FR-CST-001, FR-TRN-003a).

package engine

import (
	"fmt"

	"animus-mundi/pkg/engine/ability"
)

// EffectiveCost returns what deploying or enacting inst costs now, in
// Requisition and Tactics, the one read of a card's cost (FR-CST-001). No card
// changes a cost yet, so it is the printed cost.
func (g *Game) EffectiveCost(inst *Instance) (req, tac int) {
	return inst.Card.ReqCost, inst.Card.TacCost
}

// payCost takes inst's cost (EffectiveCost) from p.
func (g *Game) payCost(p *Player, inst *Instance) {
	req, tac := g.EffectiveCost(inst)
	p.Req -= req
	p.Tac -= tac
}

// costDecision builds the pick of the allied asset that pays cost c, with the
// assets it lists in option order. It is an untargeted selection (DecidePick),
// so Ward doesn't apply, a timeout takes the first option in fixed order
// (FR-MOD-001 §6) and a single option is taken unasked (§1). Legality
// (costsPayable) and payment (payCosts) both build their decision with it. It
// is never called for a CostDiscard, which takes no asset.
func (g *Game) costDecision(pl placement, c ability.Cost) (Decision, []*Asset) {
	verb := "kill"
	if c.Kind == ability.CostExhaust {
		verb = "exhaust"
	}
	prompt := fmt.Sprintf("%s: choose an asset to %s (additional cost)", pl.label, verb)
	return g.assetDecision(pl.controller, DecidePick, prompt, 0, func(a *Asset) bool {
		if c.Kind == ability.CostExhaust && a.Exhausted {
			return false
		}
		return g.matchesFilter(pl.controller, pl.source, c.Filter, a)
	})
}

// chooseCostDiscards has pl's controller pick the hand cards its discard cost
// discards, before the modes and targets are chosen, so the number discarded is
// known when the targets are (FR-TRN-003a, SVY-R4). It moves nothing: payCosts
// discards them after the push. Without a discard cost it asks nothing and
// returns nil. A placement has at most one discard cost.
func (g *Game) chooseCostDiscards(pl placement) []*Instance {
	var picked []*Instance
	found := false
	for _, c := range pl.costs {
		if c.Kind != ability.CostDiscard {
			continue
		}
		if found {
			panic("engine: a placement has two discard costs")
		}
		found = true
		picked = g.chooseDiscards(pl.controller, c.UpTo, true)
	}
	return picked
}

// costsPayable reports whether every additional cost of pl has at least one
// option now (FR-TRN-003a, SVY-R29). A discard cost is always payable, because
// discarding none pays it.
func (g *Game) costsPayable(pl placement) bool {
	for _, c := range pl.costs {
		if c.Kind == ability.CostDiscard {
			continue // discarding none pays it
		}
		if _, opts := g.costDecision(pl, c); len(opts) == 0 {
			return false
		}
	}
	return true
}

// payCosts pays each additional cost of pl in printed order: its controller
// picks the asset (costDecision, pickFrom) and it is killed through kill
// (CostKill) or exhausted through exhaust (CostExhaust), while a discard cost
// moves the cards picked earlier (CostDiscard). A replaced kill (not the discard
// cost) still counts as paid (SVY-R15), and the triggers either fires wait in
// the pending buffer to go above the declared card (FR-TRN-003a). It runs after
// the Maneuver is pushed, so the card is already on the stack when the cost's
// triggers are queued (FR-TRN-003), and after its targets were chosen, so they
// were judged against the state before the cost moved anything (SVY-R2). A
// discard cost's cards were picked before the modes and targets
// (chooseCostDiscards) and are moved here, through discardCards.
func (g *Game) payCosts(pl placement) {
	for _, c := range pl.costs {
		if c.Kind == ability.CostDiscard {
			g.discardCards(pl.controller, pl.discards)
			continue
		}
		d, opts := g.costDecision(pl, c)
		for _, a := range pickFrom(g, d, opts) {
			switch c.Kind {
			case ability.CostKill:
				g.kill([]*Asset{a})
			case ability.CostExhaust:
				g.exhaust(a)
			default:
				panic(fmt.Sprintf("engine: unknown cost kind %d", c.Kind))
			}
		}
	}
}
