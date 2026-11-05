// Layers: effective characteristics through the six-layer pipeline (FR-LAY-001).
// Nothing grants types (Layer 2), and the only keyword grant (Layer 3) is a
// card's Cornered keywords. There are no auras (Layer 4) yet, so the pipeline
// reduces to Layer 1 + Layer 3 + Layer 5 + Layer 6 clamping.

package engine

import "animus-mundi/pkg/engine/ability"

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

// Keywords returns the asset's effective keywords: the printed ones, plus the
// card's Cornered keywords while its controller is Cornered (SVY-R18). A
// face-down asset has none, since its face has no Ability.
func (g *Game) Keywords(a *Asset) Keywords {
	kw := a.Printed().Keywords
	if d := a.Printed().Ability; d != nil && d.Cornered != (ability.CorneredKeywords{}) && g.cornered(a.Controller) {
		kw.Aggressive += d.Cornered.Aggressive
		kw.Resistant += d.Cornered.Resistant
	}
	return kw
}

// IsLethal reports whether the next death sweep destroys a (FR-LAY-002).
func (g *Game) IsLethal(a *Asset) bool {
	return g.MaxHealth(a) <= 0 || g.CurrentHealth(a) <= 0
}
