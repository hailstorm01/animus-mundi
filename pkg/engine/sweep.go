// Death sweep: destroying lethal assets, replaced deaths first, and the match-loss checks (FR-LAY-002, FR-WIN-001).

package engine

// destroyAll destroys the assets as one simultaneous event. Every asset's
// last-known effective state is taken immediately prior to departure, before any
// of them leaves, so the departure of one cannot change what another is recorded
// with (FR-ZON-004, board directive #13). Then, for each in order: the killed
// hook fires, it leaves the battlefield, goes to its owner's discard pile, and
// the destroy event is logged.
func (g *Game) destroyAll(as []*Asset) {
	known := make([]EffectiveState, len(as))
	for i, a := range as {
		known[i] = g.effectiveState(a)
	}
	for i, a := range as {
		g.fireKilled(a)
		g.confirmDestroyed(a.ID)
		g.leaveBattlefield(a, known[i])
		g.putCard(a.Inst, ZoneDiscard, false)
		g.logEvent(Event{Kind: "destroy", Player: -1, Target: a.ID, Card: a.Inst.Card.Name}, "%s is destroyed", a)
	}
}

// kill kills each of as except a Headquarters, as one simultaneous event.
// A death a replacement applies to is replaced instead, in a fixed order,
// before the rest are destroyed (replaceDeaths), with no sweep between: a source a
// byproduct leaves lethal applies to every death of this kill and is destroyed at
// the sweep after the item (SVY-R21). A replaced kill still
// counts as done (SVY-R15). It returns how many it killed.
func (g *Game) kill(as []*Asset) int {
	var dying []*Asset
	for _, a := range as {
		if a.Printed().Type != TypeHeadquarters {
			dying = append(dying, a)
		}
	}
	g.destroyAll(g.replaceDeaths(dying, func(*Asset) bool { return true }, usedReplacements{}))
	return len(dying)
}

// deathSweep handles every lethal asset (sweepDeaths): deaths a replacement
// applies to are replaced one at a time in a fixed order (SVY-R20, SVY-R22),
// and the rest are destroyed simultaneously. Then it applies the match-loss
// checks: HQ defeat and empty-deck draws (FR-LAY-002, FR-CBT-003, FR-WIN-001).
// It runs only when nothing is resolving, so an HQ damaged
// mid-resolution loses only if it is still lethal here, and every loss found
// by one sweep is simultaneous.
func (g *Game) deathSweep() {
	var loses [2]bool
	var reasons [2]string
	g.sweepDeaths(usedReplacements{}, &loses, &reasons)
	for _, p := range g.Players {
		if p.DrewFromEmpty {
			p.DrewFromEmpty = false
			if !loses[p.ID] {
				loses[p.ID], reasons[p.ID] = true, "drew from an empty deck"
			}
		}
	}

	switch {
	case loses[0] && loses[1]:
		reason := "both players lost at the same death sweep"
		if reasons[0] == reasons[1] {
			reason = "both players lost simultaneously: " + reasons[0]
		}
		g.end(Outcome{Over: true, Draw: true, Reason: reason})
	case loses[0]:
		g.end(Outcome{Over: true, Winner: 1, Reason: reasons[0]})
	case loses[1]:
		g.end(Outcome{Over: true, Winner: 0, Reason: reasons[1]})
	}
}

// lethalAssets returns every lethal asset, in g.Assets() order.
func (g *Game) lethalAssets() []*Asset {
	var out []*Asset
	for _, a := range g.Assets() {
		if g.IsLethal(a) {
			out = append(out, a)
		}
	}
	return out
}

// destroyDying destroys the assets as one simultaneous event and notes each
// destroyed Headquarters as a loss at this sweep.
func (g *Game) destroyDying(dying []*Asset, loses *[2]bool, reasons *[2]string) {
	for _, a := range dying {
		if a.Printed().Type == TypeHeadquarters {
			loses[a.Controller], reasons[a.Controller] = true, "Headquarters destroyed"
		}
	}
	g.destroyAll(dying)
}

// sweepDeaths takes up the lethal assets in passes. In each pass, the deaths a
// replacement applies to are replaced first (replaceDeaths), then the rest die
// together. Each pass takes up every asset still lethal, a new moment (SVY-R20,
// FR-LAY-002): what a byproduct left lethal, and what a replacement left on the
// battlefield lethal. used holds the replacements already applied in this sweep;
// none applies to the same asset again (FR-REP-003, SVY-R24). The sweep ends when
// a pass finds no lethal asset, as every pass destroys an asset or uses up a
// replacement for one.
func (g *Game) sweepDeaths(used usedReplacements, loses *[2]bool, reasons *[2]string) {
	for {
		dying := g.lethalAssets()
		if len(dying) == 0 {
			return
		}
		g.destroyDying(g.replaceDeaths(dying, g.IsLethal, used), loses, reasons)
	}
}

func (g *Game) end(o Outcome) {
	if g.Outcome.Over {
		return
	}
	g.Outcome = o
	g.Phase = PhaseOver
	g.Stack = nil
	if o.Draw {
		g.logf("result", "match draw: %s", o.Reason)
	} else {
		g.logf("result", "player %d wins: %s", o.Winner, o.Reason)
	}
	g.emitMatchEnd(o)
}
