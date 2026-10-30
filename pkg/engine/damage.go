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
	g.logEvent(Event{Kind: EventDamage, Player: -1, Source: assetID(source), Target: target.ID, Amount: r.Committed,
		Raw: raw, Combat: combat, Armor: armorBefore, Absorbed: r.Absorbed},
		"%s deals %d %s damage to %s (resist->%d, cap->%d, armor %d, health %d)",
		from, raw, kind, target, r.AfterRes, r.AfterCap, r.Absorbed, r.Committed)
	if r.Taken() {
		g.fireDamaged(target)
	}
	return r
}
