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

	// Stage 1: target redirection. No M1 card redirects damage.

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
	g.logf("damage", "%s deals %d %s damage to %s (resist->%d, cap->%d, armor %d, health %d)",
		source, raw, kind, target, r.AfterRes, r.AfterCap, r.Absorbed, r.Committed)
	return r
}

// deathSweep destroys every lethal asset simultaneously, then checks HQ
// defeat (FR-LAY-002, FR-CBT-003). It runs only between stack items.
// OPEN(KIM-S02): FR-CBT-003 says an HQ loses "the moment" it reaches 0;
// the kernel checks at the sweep, which is equivalent while HQs only take
// combat damage.
func (g *Game) deathSweep() {
	var dying []*Asset
	for _, a := range g.Assets() {
		if g.IsLethal(a) {
			dying = append(dying, a)
		}
	}
	if len(dying) == 0 {
		return
	}

	hqDown := [2]bool{}
	for _, a := range dying {
		if a.Printed().Type == TypeHeadquarters {
			hqDown[a.Controller] = true
		}
	}
	for _, a := range dying {
		g.removeFromBattlefield(a)
		owner := g.Players[a.Inst.Owner]
		owner.Discard = append(owner.Discard, a.Inst)
		g.logf("destroy", "%s is destroyed", a)
	}

	switch {
	case hqDown[0] && hqDown[1]:
		g.end(Outcome{Over: true, Draw: true, Reason: "both Headquarters destroyed simultaneously"})
	case hqDown[0]:
		g.end(Outcome{Over: true, Winner: 1, Reason: "Headquarters destroyed"})
	case hqDown[1]:
		g.end(Outcome{Over: true, Winner: 0, Reason: "Headquarters destroyed"})
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
}
