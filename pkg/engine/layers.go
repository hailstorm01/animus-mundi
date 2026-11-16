// Layers: effective characteristics through the six-layer pipeline (FR-LAY-001).
// Nothing grants types (Layer 2), and Layer 3 holds a card's Cornered keywords
// and the keywords a resolved step grants until end of turn. Layer 5 holds power
// and Maximum Health modifiers. Layer 4 holds power auras (auras). The power
// pipeline is Layer 1 + Layer 4 + Layer 5, clamped at 0 by Layer 6.

package engine

import (
	"slices"

	"animus-mundi/pkg/engine/ability"
)

// Modifier is one change a resolved effect made to an asset's characteristics.
// Layer is its FR-LAY-001 layer. A Layer 5 modifier adds Power to Combat Power
// and MaxHealth to Maximum Health (FR-LAY-002; it adds no damage); a Layer 3 modifier grants Keywords.
type Modifier struct {
	Layer     int
	Power     int
	MaxHealth int
	Keywords  Keywords
	Until     ability.Duration
	By        PlayerID // the controller of the ability whose step made it; read only for UntilEndOfOpponentTurn
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

// expireAtEndOf removes the modifiers Sub-Phase 4c of active's turn expires
// (FR-TRN-001, FR-LAY-003): every one that lasts until end of turn, and every
// one that lasts until the end of its maker's opponent's turn whose maker is
// not active.
func (a *Asset) expireAtEndOf(active PlayerID) {
	a.Modifiers = slices.DeleteFunc(a.Modifiers, func(m Modifier) bool {
		switch m.Until {
		case ability.UntilEndOfOpponentTurn:
			return m.By != active
		default:
			return true
		}
	})
}

// Power returns effective Combat Power, the only stat that clamps at 0: the sum
// of Layer 1 (printed), Layer 4 (auras) and Layer 5 (modifiers), clamped by
// Layer 6.
func (g *Game) Power(a *Asset) int {
	n := a.Printed().Power + a.TempPower()
	for _, st := range g.auras(a) {
		n += st.Power
	}
	return max(0, n)
}

// auras returns every power aura that applies to battlefield asset a, oldest
// source first (FR-LAY-001). It is the one reader of which auras apply: a face-up
// battlefield source's printed StaticAura applies when its Filter, relative to the
// source's current controller, matches a. A ZoneRefSame criterion reads the
// source's current zone through zoneLock (SVY-R44). A face-down source has no
// Ability, so it gives none.
func (g *Game) auras(a *Asset) []ability.Static {
	var out []ability.Static
	for _, src := range g.Assets() {
		d := src.Printed().Ability
		if d == nil {
			continue
		}
		for _, st := range d.Statics {
			if st.Kind == ability.StaticAura && g.matchesFilterAt(src.Controller, src.ID, zoneLock(src), st.Filter, a) {
				out = append(out, st)
			}
		}
	}
	return out
}

// MaxHealth returns effective Maximum Health: printed Health plus the Layer 5
// max-health modifiers. It does not clamp (FR-LAY-001, FR-LAY-002).
func (g *Game) MaxHealth(a *Asset) int {
	n := a.Printed().Health
	for _, m := range a.Modifiers {
		if m.Layer == 5 {
			n += m.MaxHealth
		}
	}
	return n
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
	for _, m := range a.Modifiers {
		if m.Layer == 3 {
			kw.add(m.Keywords)
		}
	}
	return kw
}

// add folds o into k: a flag is set, a number is added. It walks keywordDefs,
// so a later keyword needs no new line here.
func (k *Keywords) add(o Keywords) {
	for _, d := range keywordDefs {
		switch {
		case d.num != nil:
			*d.num(k) += *d.num(&o)
		case d.flag != nil:
			*d.flag(k) = *d.flag(k) || *d.flag(&o)
		}
	}
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
