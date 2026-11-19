// Costs: reading and paying a card's cost, and checking and paying a placement's costs, a Maneuver's additional costs or an activated ability's (FR-CST-001, FR-TRN-003a, FR-ACT-001).

package engine

import (
	"fmt"
	"slices"

	"animus-mundi/pkg/engine/ability"
)

// EffectiveCost returns what deploying or enacting inst costs now, in
// Requisition and Tactics, the one read of a card's cost (FR-CST-001). It is the
// printed cost, less the card's own hand-active reduction (Def.CostReduction)
// while it is in its owner's hand, with Requisition clamped at 0. It is read on
// demand, so it follows the discard piles as they change.
func (g *Game) EffectiveCost(inst *Instance) (req, tac int) {
	req, tac = inst.Card.ReqCost, inst.Card.TacCost
	if inst.Card.Ability == nil {
		return req, tac
	}
	red := inst.Card.Ability.CostReduction
	if red.Req > 0 && slices.Contains(g.Players[inst.Owner].Hand, inst) {
		req = max(0, req-red.Req*g.countPileCards(inst, red.Per))
	}
	return req, tac
}

// countPileCards returns how many cards in the discard piles match the pile
// filter f for src's owner as controller (matchesCard).
func (g *Game) countPileCards(src *Instance, f ability.Filter) int {
	n := 0
	for _, p := range g.Players {
		for _, c := range p.Discard {
			if g.matchesCard(src.Owner, src, f, c) {
				n++
			}
		}
	}
	return n
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
	counted := c.Counted() && pl.count > 0
	if counted {
		prompt = fmt.Sprintf("%s: choose %d assets to kill (additional cost)", pl.label, pl.count)
	}
	d, opts := g.assetDecision(pl.controller, DecidePick, prompt, 0, func(a *Asset) bool {
		if c.Kind == ability.CostExhaust && a.Exhausted {
			return false
		}
		return g.matchesFilterAt(pl.controller, pl.source, nil, c.Filter, a)
	})
	if counted {
		d.Min, d.Max = pl.count, pl.count
	}
	return d, opts
}

// countedKillCost returns pl's kill cost that offers a choice of counts, or nil.
// A placement has at most one.
func countedKillCost(pl placement) *ability.Cost {
	var found *ability.Cost
	for i, c := range pl.costs {
		if !c.Counted() {
			continue
		}
		if found != nil {
			panic("engine: a placement has two counted kill costs")
		}
		found = &pl.costs[i]
	}
	return found
}

// payableCounts returns the counts of pl's counted kill cost that can be paid
// now, in printed order: those that at least that many assets matching its
// filter can pay (SVY-R53). It is the one judge of which counts can be paid. Nil
// without a counted kill cost.
func (g *Game) payableCounts(pl placement) []int {
	c := countedKillCost(pl)
	if c == nil {
		return nil
	}
	_, opts := g.costDecision(pl, *c)
	var out []int
	for _, n := range c.Counts {
		if len(opts) >= n {
			out = append(out, n)
		}
	}
	return out
}

// chooseCostCount has pl's controller choose the count its counted kill cost
// pays among counts, on declaration before the targets (SVY-R53). A single count
// is taken unasked (FR-MOD-001 §1) and a timeout takes the first listed (SVY-R54).
// It returns 0, asking nothing, when counts is empty.
func (g *Game) chooseCostCount(pl placement, counts []int) int {
	if len(counts) == 0 {
		return 0
	}
	d := Decision{Player: pl.controller, Kind: DecideMode, Prompt: pl.label + ": choose how many to kill (additional cost)", Min: 1, Max: 1}
	for _, n := range counts {
		d.Options = append(d.Options, Option{Label: fmt.Sprintf("kill %d", n), Ref: 0})
	}
	return counts[g.choose(d)[0]]
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
		picked = g.chooseHandCards(pl.controller, pl.controller, c.UpTo, true)
	}
	return picked
}

// costsPayable reports whether every cost of pl can be paid now: a resources
// cost needs its Requisition and Tactics (CostResources), and the others are
// otherCostsPayable's to judge (FR-TRN-003a, FR-ACT-001, SVY-R29).
func (g *Game) costsPayable(pl placement) bool {
	for _, c := range pl.costs {
		if c.Kind == ability.CostResources && !g.Players[pl.controller].canPay(c.Req, c.Tac) {
			return false
		}
	}
	return g.otherCostsPayable(pl)
}

// ActivatedCost returns the Requisition and Tactics that a's printed activated
// ability i costs, and whether its other costs can be paid now (FR-ACT-001). It
// checks neither the resources against the pool nor the ability's targets, so a
// client can plan for an ability it can't activate yet. LegalActions says what
// can be activated now.
func (g *Game) ActivatedCost(a *Asset, i int) (req, tac int, ok bool) {
	act := a.Printed().Ability.Activated[i]
	for _, c := range act.Costs {
		if c.Kind == ability.CostResources {
			req += c.Req
			tac += c.Tac
		}
	}
	return req, tac, g.otherCostsPayable(g.activatedPlacement(a.Controller, a, act))
}

// DiscardDeployCost returns the Requisition and Tactics that deploying inst from
// its owner's discard pile costs, and whether it may be deployed from there: it
// is in that pile, prints a discard-active deploy permission (FR-TRN-003), and
// the permission's and the card's additional costs other than resources can be
// paid now (FR-TRN-003a). Like ActivatedCost, it checks neither the resources
// against the pool nor the zone, so a client can plan for a deploy it can't
// make yet. LegalActions says what can be deployed now.
func (g *Game) DiscardDeployCost(inst *Instance) (req, tac int, ok bool) {
	req, tac = g.EffectiveCost(inst)
	if pileCard(g.Players[inst.Owner], inst.ID, ZoneDiscard) == nil || deployPermission(inst.Card) == nil {
		return req, tac, false
	}
	// No cost depends on the zone, so Reserves stands in for either.
	return req, tac, g.otherCostsPayable(deployPlacement(inst.Owner, inst, ZoneReserves, ZoneDiscard))
}

// PlayCost returns the Requisition and Tactics that playing inst costs now, and
// whether a client should plan for it: a supported card in its owner's hand
// that has a deploy zone or is a Maneuver costs EffectiveCost, without checking
// its additional costs, and a supported card in its owner's discard pile that
// has a deploy zone costs DiscardDeployCost. Anywhere else ok is false. Like
// DiscardDeployCost, it checks neither the resources against the pool nor the
// zone. LegalActions says what can be played now.
func (g *Game) PlayCost(inst *Instance) (req, tac int, ok bool) {
	c, p := inst.Card, g.Players[inst.Owner]
	switch {
	case pileCard(p, inst.ID, ZoneHand) != nil:
		if !c.Supported() || (len(c.DeployZones) == 0 && c.Type != TypeManeuver) {
			return 0, 0, false
		}
		req, tac = g.EffectiveCost(inst)
		return req, tac, true
	case pileCard(p, inst.ID, ZoneDiscard) != nil:
		if !c.Supported() || len(c.DeployZones) == 0 {
			return 0, 0, false
		}
		return g.DiscardDeployCost(inst)
	}
	return 0, 0, false
}

// otherCostsPayable reports whether every cost of pl except its resources can be
// paid now: exhausting the source needs it Readied (CostExhaustSelf), and a kill
// or exhaust cost needs at least one option, a counted kill cost one payable
// count (SVY-R29, SVY-R53). A discard cost is always payable, because
// discarding none pays it, and so is killing the source (CostKillSelf).
func (g *Game) otherCostsPayable(pl placement) bool {
	for _, c := range pl.costs {
		switch c.Kind {
		case ability.CostResources:
			continue // not this check's to judge
		case ability.CostDiscard, ability.CostKillSelf:
			continue // discarding none pays it; killing the source always can
		case ability.CostExhaustSelf:
			if a := g.Asset(pl.source); a == nil || a.Exhausted {
				return false
			}
			continue
		}
		if c.Counted() {
			if len(g.payableCounts(pl)) == 0 {
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
		picked := pickFrom(g, d, opts)
		switch c.Kind {
		case ability.CostKill:
			g.kill(picked)
		case ability.CostExhaust:
			for _, a := range picked {
				g.exhaust(a)
			}
		default:
			panic(fmt.Sprintf("engine: unknown cost kind %d", c.Kind))
		}
	}
}
