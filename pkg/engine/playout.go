package engine

import (
	"fmt"
	"math/rand/v2"
)

// Policy picks one of the legal actions for the active player.
type Policy func(g *Game, legal []Action) Action

// RandomPolicy picks uniformly among non-EndTurn actions, ending the turn
// with probability endBias or when nothing else is legal.
func RandomPolicy(seed uint64, endBias float64) Policy {
	r := rand.New(rand.NewPCG(seed, ^seed))
	return func(g *Game, legal []Action) Action {
		if len(legal) == 1 || r.Float64() < endBias {
			return legal[len(legal)-1] // LegalActions always ends with EndTurn
		}
		return legal[r.IntN(len(legal)-1)]
	}
}

// ProgressPolicy is a random policy that favors progress over passing. It
// picks only from legal and is deterministic for seed. Each call it prefers:
//  1. a Deploy (uniform over card and zone);
//  2. a Generate or Convert that yields the resource a deployable hand card
//     is short of;
//  3. an Attack, March or Scratch (probability 0.9, uniform), else EndTurn;
//  4. EndTurn.
func ProgressPolicy(seed uint64) Policy {
	r := rand.New(rand.NewPCG(seed, ^seed))
	return func(g *Game, legal []Action) Action {
		var deploys, closers, acts []Action
		end := legal[len(legal)-1] // LegalActions always ends with EndTurn
		p := g.Players[g.Active]
		needReq, needTac := false, false
		for _, inst := range p.Hand {
			c := inst.Card
			if !c.Supported() || len(c.DeployZones) == 0 {
				continue
			}
			needReq = needReq || c.ReqCost > p.Req
			needTac = needTac || c.TacCost > p.Tac
		}
		for _, a := range legal {
			switch a.Kind {
			case ActDeploy:
				deploys = append(deploys, a)
			case ActGenerate:
				printed := g.Asset(a.Asset).Printed()
				if (a.Resource == Requisition && needReq && printed.ReqYield > 0) ||
					(a.Resource == Tactics && needTac && printed.TacYield > 0) {
					closers = append(closers, a)
				}
			case ActConvert:
				if needTac {
					closers = append(closers, a)
				}
			case ActMarch, ActAttack, ActScratch:
				acts = append(acts, a)
			}
		}
		switch {
		case len(deploys) > 0:
			return deploys[r.IntN(len(deploys))]
		case len(closers) > 0:
			return closers[r.IntN(len(closers))]
		case len(acts) > 0 && r.Float64() < 0.9:
			return acts[r.IntN(len(acts))]
		}
		return end
	}
}

// PlayoutResult summarizes one simulated match.
type PlayoutResult struct {
	Outcome     Outcome
	Turns       int
	Actions     int
	FirstPlayer PlayerID
}

// Playout plays g to completion with one policy per seat, checking
// invariants after every action. It fails if the match exceeds maxTurns.
func Playout(g *Game, policies [2]Policy, maxTurns int) (PlayoutResult, error) {
	res := PlayoutResult{FirstPlayer: g.FirstPlayer}
	for !g.Outcome.Over {
		if g.Turn > maxTurns {
			return res, fmt.Errorf("match did not end within %d turns", maxTurns)
		}
		legal := g.LegalActions()
		a := policies[g.Active](g, legal)
		logged := len(g.Log)
		if err := g.Apply(a); err != nil {
			return res, fmt.Errorf("turn %d: policy chose an action LegalActions offered but Apply rejected: %w", g.Turn, err)
		}
		for _, e := range g.Log[logged:] {
			if e.Kind == "invalid-answer" {
				return res, fmt.Errorf("turn %d after %s: a chooser answer was rejected: %s", g.Turn, a.Kind, e.Detail)
			}
		}
		res.Actions++
		if err := g.CheckInvariants(); err != nil {
			return res, fmt.Errorf("turn %d after %s: %w", g.Turn, a.Kind, err)
		}
	}
	res.Outcome = g.Outcome
	res.Turns = g.Turn
	return res, nil
}

// CheckInvariants verifies properties that must hold between actions.
func (g *Game) CheckInvariants() error {
	seen := make(map[int]string)
	note := func(inst *Instance, where string) error {
		if prev, dup := seen[inst.ID]; dup {
			return fmt.Errorf("card instance %d is in both %s and %s", inst.ID, prev, where)
		}
		seen[inst.ID] = where
		return nil
	}
	for _, p := range g.Players {
		if p.Req < 0 || p.Tac < 0 {
			return fmt.Errorf("player %d has a negative resource pool (R%d T%d)", p.ID, p.Req, p.Tac)
		}
		if p.DrewFromEmpty && !g.Outcome.Over {
			return fmt.Errorf("player %d drew from an empty deck but no death sweep settled it", p.ID)
		}
		for _, zone := range []struct {
			name  string
			cards []*Instance
		}{{"deck", p.Deck}, {"hand", p.Hand}, {"discard", p.Discard}} {
			for _, inst := range zone.cards {
				if err := note(inst, fmt.Sprintf("player %d %s", p.ID, zone.name)); err != nil {
					return err
				}
			}
		}
	}
	for _, a := range g.Assets() {
		if err := note(a.Inst, a.Zone.String()); err != nil {
			return err
		}
		if a.Armor < 0 || a.Damage < 0 {
			return fmt.Errorf("%s has negative armor or damage", a)
		}
		if len(g.Stack) == 0 && g.IsLethal(a) {
			return fmt.Errorf("%s is lethal but survived the death sweep", a)
		}
	}
	for _, p := range g.Players {
		if p.HQ != nil && (p.HQ.Zone != ZoneHeadquarters || p.HQ.Controller != p.ID) {
			return fmt.Errorf("%s is player %d's HQ but records zone %s", p.HQ, p.ID, p.HQ.Zone)
		}
		for _, a := range p.Reserves {
			if a.Zone != ZoneReserves || a.Controller != p.ID {
				return fmt.Errorf("%s is in player %d's Reserves but records zone %s, controller %d", a, p.ID, a.Zone, a.Controller)
			}
		}
	}
	for _, a := range g.Theater {
		if a.Zone != ZoneMainTheater {
			return fmt.Errorf("%s is in the Main Theater but records zone %s", a, a.Zone)
		}
	}
	if !g.Outcome.Over && len(g.Stack) != 0 {
		return fmt.Errorf("stack not empty between actions")
	}
	if len(g.pending) != 0 {
		return fmt.Errorf("%d fired triggers left in the pending buffer between actions", len(g.pending))
	}
	return nil
}
