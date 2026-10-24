package engine

// DamageResult records what each mitigation stage did to one damage instance.
type DamageResult struct {
	Raw       int
	AfterRes  int // after Stage 2 (Resistant)
	AfterCap  int // after Stage 3 (Stalwart)
	Absorbed  int // Stage 4 (Armor)
	Committed int // Stage 5 (Health)
}

// Taken reports whether the target "takes damage" for trigger purposes:
// (armorAbsorbed + healthDamageCommitted) > 0 (FR-REP-004 Stage 5).
func (r DamageResult) Taken() bool { return r.Absorbed+r.Committed > 0 }

// dealDamage routes one discrete damage instance through the 5-stage
// mitigation sequence (FR-REP-004). It never destroys anything; destruction
// waits for the state-based death sweep (FR-LAY-002).
func (g *Game) dealDamage(source, target *Asset, raw int, combat bool) DamageResult {
	r := DamageResult{Raw: raw}
	kw := g.Keywords(target)
	armorBefore := target.Armor

	// Stage 1: target redirection. No executable card redirects damage yet.

	// Stage 2: flat Resistant N. Negative Resistant amplifies, even 0 raw damage.
	dmg := max(0, raw-kw.Resistant)
	r.AfterRes = dmg

	// Stage 3: Stalwart caps combat damage at 1; non-combat damage skips it.
	if combat && kw.Stalwart {
		dmg = min(dmg, 1)
	}
	r.AfterCap = dmg

	// Stage 4: dynamic Armor absorbs 1-for-1.
	r.Absorbed = min(dmg, target.Armor)
	target.Armor -= r.Absorbed
	dmg -= r.Absorbed

	// Stage 5: commit to Health.
	r.Committed = dmg
	target.Damage += dmg

	kind := "effect"
	if combat {
		kind = "combat"
	}
	// A source that has left the battlefield is passed as its last-known
	// record, so it is still named; a nil source never reaches the format.
	from := "an unknown source"
	if source != nil {
		from = source.String()
	}
	g.logEvent(Event{Kind: "damage", Player: -1, Source: assetID(source), Target: target.ID, Amount: r.Committed,
		Raw: raw, Combat: combat, Armor: armorBefore, Absorbed: r.Absorbed},
		"%s deals %d %s damage to %s (resist->%d, cap->%d, armor %d, health %d)",
		from, raw, kind, target, r.AfterRes, r.AfterCap, r.Absorbed, r.Committed)
	if r.Taken() {
		g.fireDamaged(target)
	}
	return r
}

// deathSweep destroys every lethal asset simultaneously, then applies the
// match-loss checks: HQ defeat and empty-deck draws (FR-LAY-002, FR-CBT-003,
// FR-WIN-001). It runs only when nothing is resolving, so an HQ damaged
// mid-resolution loses only if it is still lethal here, and every loss found
// by one sweep is simultaneous.
func (g *Game) deathSweep() {
	var dying []*Asset
	for _, a := range g.Assets() {
		if g.IsLethal(a) {
			dying = append(dying, a)
		}
	}

	var loses [2]bool
	var reasons [2]string
	for _, a := range dying {
		if a.Printed().Type == TypeHeadquarters {
			loses[a.Controller], reasons[a.Controller] = true, "Headquarters destroyed"
		}
	}
	// Every dying asset's last-known effective state is taken immediately prior
	// to departure, before any of them leaves, so the departure of one cannot
	// change what another is recorded with (FR-ZON-004, board directive #13).
	known := make([]EffectiveState, len(dying))
	for i, a := range dying {
		known[i] = g.effectiveState(a)
	}
	for i, a := range dying {
		g.fireKilled(a)
		g.leaveBattlefield(a, known[i])
		g.putCard(a.Inst, ZoneDiscard, false)
		g.logEvent(Event{Kind: "destroy", Player: -1, Target: a.ID, Card: a.Inst.Card.Name}, "%s is destroyed", a)
	}
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
