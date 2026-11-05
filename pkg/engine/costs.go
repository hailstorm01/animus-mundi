// Costs: reading and paying a card's cost, and checking and paying a placement's costs, a Maneuver's additional costs or an activated ability's (FR-CST-001, FR-TRN-003a, FR-ACT-001).

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

// canPay reports whether p holds req Requisition and tac Tactics, the one read
// of "can pay".
func (p *Player) canPay(req, tac int) bool {
	return p.Req >= req && p.Tac >= tac
}

// pay takes req Requisition and tac Tactics from p, the one payment.
func (p *Player) pay(req, tac int) {
	p.Req -= req
	p.Tac -= tac
}

// payCost takes inst's cost (EffectiveCost) from p.
func (g *Game) payCost(p *Player, inst *Instance) {
	req, tac := g.EffectiveCost(inst)
	p.pay(req, tac)
}

// costDecision builds the pick of the allied asset that pays cost c, with the
// assets it lists in option order. It is an untargeted selection (DecidePick),
// so Ward doesn't apply, a timeout takes the first option in fixed order
// (FR-MOD-001 §6) and a single option is taken unasked (§1). Legality
// (costsPayable) and payment (payCosts) both build their decision with it. It
// is never called for a kind that takes no chosen asset (CostDiscard,
// CostResources, CostExhaustSelf, CostKillSelf).
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

// costsPayable reports whether every cost of pl can be paid now: a resources
// cost needs its Requisition and Tactics (CostResources), exhausting the source
// needs it Readied (CostExhaustSelf), and a kill or exhaust cost needs at least
// one option (FR-TRN-003a, FR-ACT-001, SVY-R29). A discard cost is always
// payable, because discarding none pays it, and so is killing the source
// (CostKillSelf).
func (g *Game) costsPayable(pl placement) bool {
	for _, c := range pl.costs {
		switch c.Kind {
		case ability.CostDiscard, ability.CostKillSelf:
			continue // discarding none pays it; killing the source always can
		case ability.CostResources:
			if !g.Players[pl.controller].canPay(c.Req, c.Tac) {
				return false
			}
			continue
		case ability.CostExhaustSelf:
			if a := g.Asset(pl.source); a == nil || a.Exhausted {
				return false
			}
			continue
		}
		if _, opts := g.costDecision(pl, c); len(opts) == 0 {
			return false
		}
	}
	return true
}

// payCosts pays each cost of pl in printed order: its controller picks the asset
// (costDecision, pickFrom) and it is killed through kill (CostKill) or exhausted
// through exhaust (CostExhaust), the source itself is killed or exhausted the
// same way (CostKillSelf, CostExhaustSelf), a resources cost takes its
// Requisition and Tactics (CostResources), and a discard cost moves the cards
// picked earlier (CostDiscard). A replaced kill (not the discard cost) still
// counts as paid (SVY-R15), and the triggers a kill or an exhaust fires wait in
// the pending buffer to go above the declared item (FR-TRN-003a). It runs after
// the item is pushed, so it is already on the stack when the cost's triggers are
// queued (FR-TRN-003), and after its targets were chosen, so they were judged
// against the state before the cost moved anything (SVY-R2). A discard cost's
// cards were picked before the modes and targets (chooseCostDiscards) and are
// moved here, through discardCards.
func (g *Game) payCosts(pl placement) {
	for _, c := range pl.costs {
		if c.Kind == ability.CostDiscard {
			g.discardCards(pl.controller, pl.discards)
			continue
		}
		switch c.Kind {
		case ability.CostResources:
			g.Players[pl.controller].pay(c.Req, c.Tac)
			continue
		case ability.CostExhaustSelf:
			g.exhaust(g.Asset(pl.source))
			continue
		case ability.CostKillSelf:
			g.kill([]*Asset{g.Asset(pl.source)})
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
