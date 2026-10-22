package sim

import (
	"fmt"

	"animus-mundi/pkg/engine"
)

// PlayoutResult summarizes one simulated match.
type PlayoutResult struct {
	Outcome     engine.Outcome
	Turns       int
	Actions     int
	FirstPlayer engine.PlayerID
}

// Playout plays g to completion with one policy per seat, checking
// invariants after every action. It fails if the match exceeds maxTurns.
func Playout(g *engine.Game, policies [2]Policy, maxTurns int) (PlayoutResult, error) {
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
