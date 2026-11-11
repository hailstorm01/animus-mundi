package sim

import (
	"fmt"
	"maps"
	"math/rand/v2"
	"slices"

	"animus-mundi/pkg/engine"
)

// Policy picks one of the legal actions for the active player.
type Policy func(g *engine.Game, legal []engine.Action) engine.Action

// PolicyByName returns the constructor of the policy that a -policy flag
// names: progress or random.
func PolicyByName(name string) (func(seed uint64) Policy, error) {
	switch name {
	case "progress":
		return ProgressPolicy, nil
	case "random":
		return defaultRandomPolicy, nil
	}
	return nil, fmt.Errorf("unknown -policy %q (want progress or random)", name)
}

// defaultRandomPolicy is RandomPolicy at DefaultEndBias.
func defaultRandomPolicy(seed uint64) Policy { return RandomPolicy(seed, DefaultEndBias) }

// RandomPolicy picks uniformly among non-EndTurn actions, ending the turn
// with probability endBias or when nothing else is legal.
func RandomPolicy(seed uint64, endBias float64) Policy {
	r := rand.New(rand.NewPCG(seed, ^seed))
	return func(g *engine.Game, legal []engine.Action) engine.Action {
		if len(legal) == 1 || r.Float64() < endBias {
			return legal[len(legal)-1] // LegalActions always ends with EndTurn
		}
		return legal[r.IntN(len(legal)-1)]
	}
}

// marchPast is how often a unit marches out of Reserves anyway while a playable
// card is short of Requisition it can't reach: often enough that a card the player
// may never afford doesn't pin its units there.
const marchPast = 0.25

// ProgressPolicy is a random policy that favors progress over passing. It
// picks only from legal and is deterministic for seed. Each call it prefers:
//  1. a play: a Deploy or an Enact, from hand or a Deploy from the discard pile,
//     uniform over card and zone;
//  2. a Generate or Convert that funds a playable card's cost (funds, playCosts);
//  3. an Attack, March, Scratch or Activate, or a Generate or Convert that funds
//     an activated ability's cost (probability 0.9, uniform), else EndTurn;
//  4. EndTurn.
//
// A unit in Reserves that yields Requisition stays there, except marchPast of
// the time, while a playable card is short of Requisition it can't reach this turn,
// so that its yield is there on a later turn.
func ProgressPolicy(seed uint64) Policy {
	r := rand.New(rand.NewPCG(seed, ^seed))
	return func(g *engine.Game, legal []engine.Action) engine.Action {
		var plays, closers, acts []engine.Action
		end := legal[len(legal)-1] // LegalActions always ends with EndTurn
		p := g.Players[g.Active]
		pool := resources{p.Req, p.Tac}
		taps := legalTaps(g, legal)
		cards := playCosts(g, p)
		activations := activationCosts(g)
		saving := slices.ContainsFunc(cards, func(c neededCost) bool {
			return c.req > pool.req && !reachable(pool, taps, c.resources)
		})
		for _, a := range legal {
			switch a.Kind {
			case engine.ActDeploy, engine.ActEnact:
				plays = append(plays, a)
			case engine.ActGenerate, engine.ActConvert:
				if fundsAny(a, pool, taps, cards) {
					closers = append(closers, a)
				} else if fundsAny(a, pool, taps, activations) {
					acts = append(acts, a)
				}
			case engine.ActMarch:
				as := g.Asset(a.Asset)
				if saving && as.Zone == engine.ZoneReserves && as.Printed().ReqYield > 0 && r.Float64() >= marchPast {
					continue
				}
				acts = append(acts, a)
			case engine.ActAttack, engine.ActScratch, engine.ActActivate:
				acts = append(acts, a)
			}
		}
		switch {
		case len(plays) > 0:
			return plays[r.IntN(len(plays))]
		case len(closers) > 0:
			return closers[r.IntN(len(closers))]
		case len(acts) > 0 && r.Float64() < 0.9:
			return acts[r.IntN(len(acts))]
		}
		return end
	}
}

// resources is an amount of Requisition and Tactics.
type resources struct{ req, tac int }

// tap is what one asset can yield if it generates now, per resource. A resource
// its zone doesn't allow (FR-CAP-001) or its printed yield lacks is 0.
type tap struct {
	asset int
	yield resources
}

// neededCost is a cost the active player may want to pay. source is the asset
// whose activated ability it is, which won't tap to fund it, or 0 for a card to
// play: no asset has that ID.
type neededCost struct {
	resources
	source int
}

// legalTaps reads the active player's taps from legal, so the engine alone says
// which zone yields which resource.
func legalTaps(g *engine.Game, legal []engine.Action) []tap {
	var taps []tap
	for _, a := range legal {
		if a.Kind != engine.ActGenerate {
			continue
		}
		i := slices.IndexFunc(taps, func(t tap) bool { return t.asset == a.Asset })
		if i < 0 {
			taps = append(taps, tap{asset: a.Asset})
			i = len(taps) - 1
		}
		printed := g.Asset(a.Asset).Printed()
		if a.Resource == engine.Requisition {
			taps[i].yield.req = printed.ReqYield
		} else {
			taps[i].yield.tac = printed.TacYield
		}
	}
	return taps
}

// playCosts lists what p's playable cards cost now: its hand cards, and the
// cards it may deploy from its discard pile by their own permission, leaving out
// one whose other costs can't be paid now (Game.DiscardDeployCost, FR-TRN-003).
func playCosts(g *engine.Game, p *engine.Player) []neededCost {
	var out []neededCost
	for _, inst := range p.Hand {
		c := inst.Card
		if !c.Supported() || (len(c.DeployZones) == 0 && c.Type != engine.TypeManeuver) {
			continue
		}
		req, tac := g.EffectiveCost(inst)
		out = append(out, neededCost{resources: resources{req, tac}})
	}
	for _, inst := range p.Discard {
		if !inst.Card.Supported() || len(inst.Card.DeployZones) == 0 {
			continue
		}
		if req, tac, ok := g.DiscardDeployCost(inst); ok {
			out = append(out, neededCost{resources: resources{req, tac}})
		}
	}
	return out
}

// activationCosts lists the resource costs of the active player's activated
// abilities, leaving out an ability whose other costs can't be paid now
// (ActivatedCost). It doesn't check targets, so a tap may fund an ability that
// then has none.
func activationCosts(g *engine.Game) []neededCost {
	var out []neededCost
	for _, as := range g.Assets() {
		def := as.Printed().Ability
		if as.Controller != g.Active || def == nil {
			continue
		}
		for i := range def.Activated {
			if req, tac, ok := g.ActivatedCost(as, i); ok {
				out = append(out, neededCost{resources: resources{req, tac}, source: as.ID})
			}
		}
	}
	return out
}

// fundsAny reports whether a funds one of costs.
func fundsAny(a engine.Action, pool resources, taps []tap, costs []neededCost) bool {
	return slices.ContainsFunc(costs, func(c neededCost) bool { return funds(a, pool, taps, c) })
}

// funds reports whether Generate or Convert a adds a resource c is short of and
// leaves c reachable this turn. Requisition counts toward a Tactics shortfall
// only when no asset can tap for Tactics, through conversion (FR-RES-003).
func funds(a engine.Action, pool resources, taps []tap, c neededCost) bool {
	shortReq, shortTac := c.req > pool.req, c.tac > pool.tac
	var rest []tap
	for _, t := range taps {
		if t.asset != a.Asset && t.asset != c.source {
			rest = append(rest, t)
		}
	}
	if a.Kind == engine.ActConvert {
		return shortTac && reachable(resources{pool.req - engine.ConvertReq, pool.tac + engine.ConvertTac}, rest, c.resources)
	}
	i := slices.IndexFunc(taps, func(t tap) bool { return t.asset == a.Asset })
	if a.Asset == c.source || i < 0 {
		return false
	}
	y := taps[i].yield
	if a.Resource == engine.Tactics {
		return y.tac > 0 && shortTac && reachable(resources{pool.req, pool.tac + y.tac}, rest, c.resources)
	}
	viaConvert := shortTac && !slices.ContainsFunc(taps, func(t tap) bool { return t.yield.tac > 0 })
	return y.req > 0 && (shortReq || viaConvert) && reachable(resources{pool.req + y.req, pool.tac}, rest, c.resources)
}

// reachable reports whether pool and taps can pay c this turn: each asset taps
// for one resource, and Requisition beyond c's converts to Tactics.
func reachable(pool resources, taps []tap, c resources) bool {
	states := map[resources]bool{pool: true}
	for _, t := range taps {
		next := maps.Clone(states)
		for s := range states {
			next[resources{s.req + t.yield.req, s.tac}] = true
			next[resources{s.req, s.tac + t.yield.tac}] = true
		}
		states = next
	}
	for s := range states {
		if s.req >= c.req && s.tac+(s.req-c.req)/engine.ConvertReq*engine.ConvertTac >= c.tac {
			return true
		}
	}
	return false
}
