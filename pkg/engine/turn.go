// Turn phases: beginTurn and endTurn (FR-TRN-001).

package engine

// beginTurn runs the Ready and Draw phases, then opens the Main Phase (FR-TRN-001).
func (g *Game) beginTurn() {
	g.Phase = PhaseReady
	g.emitTurnStart()
	for _, a := range g.Assets() {
		if a.Controller == g.Active {
			a.Exhausted = false
		}
	}
	// No passive income (FR-RES-001). No executable card has start-of-turn triggers yet.

	g.Phase = PhaseDraw
	if !(g.Turn == 1 && g.Active == g.FirstPlayer) { // Turn 1 draw skip (FR-SET-004)
		g.draw(g.Active, 1)
		g.deathSweep() // nothing is resolving, so an empty-deck draw is settled now
		if g.Outcome.Over {
			return
		}
	}
	g.Phase = PhaseMain
}

// endTurn runs End Phase sub-phases 4a-4d and starts the next turn (FR-TRN-001).
func (g *Game) endTurn() {
	g.Phase = PhaseEnd
	// 4a/4b: no executable card has end-of-turn triggers yet.
	// 4c: pools reset, "until end of turn" modifiers expire.
	for _, p := range g.Players {
		p.Req, p.Tac = 0, 0
	}
	for _, a := range g.Assets() {
		a.TempPower = 0
	}
	// 4d: handover.
	g.Active = g.Active.Opponent()
	g.Turn++
	g.beginTurn()
}
