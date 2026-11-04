package sim

import (
	"fmt"
	"math/rand/v2"

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

// ProgressPolicy is a random policy that favors progress over passing. It
// picks only from legal and is deterministic for seed. Each call it prefers:
//  1. a play from hand: a Deploy or an Enact, uniform over card and zone;
//  2. a Generate or Convert that yields the resource a deployable or
//     enactable hand card is short of;
//  3. an Attack, March or Scratch (probability 0.9, uniform), else EndTurn;
//  4. EndTurn.
func ProgressPolicy(seed uint64) Policy {
	r := rand.New(rand.NewPCG(seed, ^seed))
	return func(g *engine.Game, legal []engine.Action) engine.Action {
		var plays, closers, acts []engine.Action
		end := legal[len(legal)-1] // LegalActions always ends with EndTurn
		p := g.Players[g.Active]
		needReq, needTac := false, false
		for _, inst := range p.Hand {
			c := inst.Card
			if !c.Supported() || (len(c.DeployZones) == 0 && c.Type != engine.TypeManeuver) {
				continue
			}
			req, tac := g.EffectiveCost(inst)
			needReq = needReq || req > p.Req
			needTac = needTac || tac > p.Tac
		}
		for _, a := range legal {
			switch a.Kind {
			case engine.ActDeploy, engine.ActEnact:
				plays = append(plays, a)
			case engine.ActGenerate:
				printed := g.Asset(a.Asset).Printed()
				if (a.Resource == engine.Requisition && needReq && printed.ReqYield > 0) ||
					(a.Resource == engine.Tactics && needTac && printed.TacYield > 0) {
					closers = append(closers, a)
				}
			case engine.ActConvert:
				if needTac {
					closers = append(closers, a)
				}
			case engine.ActMarch, engine.ActAttack, engine.ActScratch:
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
