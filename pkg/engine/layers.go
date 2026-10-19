package engine

// Effective characteristics through the six-layer pipeline (FR-LAY-001).
// M1 has no type grants (Layer 2), keyword grants (Layer 3) or auras
// (Layer 4), so the pipeline reduces to Layer 1 + Layer 5 + Layer 6 clamping.

// Power returns effective Combat Power, the only stat that clamps at 0.
func (g *Game) Power(a *Asset) int {
	return max(0, a.Printed().Power+a.TempPower)
}

// MaxHealth returns effective Maximum Health. It does not clamp (FR-LAY-002).
func (g *Game) MaxHealth(a *Asset) int {
	return a.Printed().Health
}

// CurrentHealth is Effective Maximum Health minus accumulated damage.
func (g *Game) CurrentHealth(a *Asset) int {
	return g.MaxHealth(a) - a.Damage
}

// Keywords returns the asset's effective keywords. Face-down assets have none.
func (g *Game) Keywords(a *Asset) Keywords {
	return a.Printed().Keywords
}

// IsLethal reports whether the next death sweep destroys a (FR-LAY-002).
func (g *Game) IsLethal(a *Asset) bool {
	return g.MaxHealth(a) <= 0 || g.CurrentHealth(a) <= 0
}
