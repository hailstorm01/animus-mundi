// Layers: effective characteristics through the six-layer pipeline (FR-LAY-001).
// Nothing grants types (Layer 2), and the only keyword grant (Layer 3) is a
// card's Cornered keywords. There are no auras (Layer 4) yet, so the pipeline
// reduces to Layer 1 + Layer 3 + Layer 5 + Layer 6 clamping.

package engine

import (
	"slices"

	"animus-mundi/pkg/engine/ability"
)

// Duration is how long a Modifier lasts (FR-LAY-003).
type Duration int

const (
	UntilEndOfTurn Duration = iota // expires in Sub-Phase 4c (FR-TRN-001)
)

// Modifier is one change a resolved effect made to an asset's characteristics.
// Layer is its FR-LAY-001 layer. The only kind built so far is a Layer 5 power
// modifier: Power is added to Combat Power.
type Modifier struct {
	Layer int
	Power int
	Until Duration
}

// TempPower is the sum of a's Layer 5 power modifiers.
func (a *Asset) TempPower() int {
	n := 0
	for _, m := range a.Modifiers {
		if m.Layer == 5 {
			n += m.Power
		}
	}
	return n
}

// expire removes a's modifiers that last d.
func (a *Asset) expire(d Duration) {
	a.Modifiers = slices.DeleteFunc(a.Modifiers, func(m Modifier) bool { return m.Until == d })
}

// Power returns effective Combat Power, the only stat that clamps at 0.
func (g *Game) Power(a *Asset) int {
	return max(0, a.Printed().Power+a.TempPower())
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

// cornered reports whether p is Cornered: the opponent controls strictly more
// battlefield assets (units and structures) in the Main Theater than p does.
// Headquarters are never in the Main Theater (FR-KWD-002e).
func (g *Game) cornered(p PlayerID) bool {
	return len(g.TheaterOf(p.Opponent())) > len(g.TheaterOf(p))
}

// IsLethal reports whether the next death sweep destroys a (FR-LAY-002).
func (g *Game) IsLethal(a *Asset) bool {
	return g.MaxHealth(a) <= 0 || g.CurrentHealth(a) <= 0
}

// EffectiveState is an asset's effective characteristics at one moment, read
// through the layer pipeline (FR-LAY-001).
type EffectiveState struct {
	Power     int
	MaxHealth int
	Keywords  Keywords
}

// effectiveState reads a's effective characteristics now.
func (g *Game) effectiveState(a *Asset) EffectiveState {
	return EffectiveState{Power: g.Power(a), MaxHealth: g.MaxHealth(a), Keywords: g.Keywords(a)}
}
