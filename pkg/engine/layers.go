// Layers: effective characteristics through the six-layer pipeline (FR-LAY-001).
// Nothing grants types (Layer 2), and Layer 3 holds a card's Cornered keywords,
// the keywords a resolved step grants until end of turn, and auras' keyword
// grants (`Game.auras`) and their Cornered keyword grants. Layer 5 holds power
// and Maximum Health modifiers. Layer 4 holds power auras (`Game.auras`), their
// Cornered power and a card's own Cornered power. The power
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
	Keywords  ability.Keywords
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
// of Layer 1 (printed), Layer 4 (the card's own Cornered power, auras and their
// Cornered grants) and Layer 5 (modifiers), clamped by Layer 6.
func (g *Game) Power(a *Asset) int {
	n := a.Printed().Power + a.TempPower() + g.whileCornered(a).Power
	for _, st := range g.auras(a) {
		n += st.Power
	}
	return max(0, n)
}

// auras returns every aura that applies to battlefield asset a, in
// g.Assets() order, which the sums of Layers 3 and 4 do not depend on (FR-LAY-001). It
// is the one reader of which auras apply: a face-up battlefield source's printed
// StaticAura applies when its Filter, relative to the
// source's current controller, matches a. A ZoneRefSame criterion reads the
// source's current zone through zoneLock (SVY-R44). A face-down source has no
// Ability, so it gives none. Each caller reads the field it needs (Power, Resistant).
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
// card's Cornered keywords while its controller is Cornered (SVY-R18), plus the
// Resistant of each aura that applies (Game.auras, Layer 3). The Cornered part
// includes auras' Cornered grants. A face-down asset has no Cornered keywords of
// its own, since its face has no Ability, but it still receives auras' grants.
func (g *Game) Keywords(a *Asset) ability.Keywords {
	kw := a.Printed().Keywords
	wc := g.whileCornered(a)
	kw.Aggressive += wc.Aggressive
	kw.Resistant += wc.Resistant
	for _, st := range g.auras(a) {
		kw.Resistant += st.Resistant
	}
	for _, m := range a.Modifiers {
		if m.Layer == 3 {
			kw.Add(m.Keywords)
		}
	}
	return kw
}

// whileCornered returns what a has while its controller is Cornered: the zero
// value when it is not, else the card's own Cornered abilities (none for a
// face-down asset, whose face has no Ability) plus the Cornered of each aura
// that applies (Game.auras). It is the one reader of that gate, live, for
// Keywords and Power (FR-KWD-002e, FR-LAY-003).
func (g *Game) whileCornered(a *Asset) ability.WhileCornered {
	var wc ability.WhileCornered
	if !g.cornered(a.Controller) {
		return wc
	}
	if d := a.Printed().Ability; d != nil {
		wc = d.Cornered
	}
	for _, st := range g.auras(a) {
		wc = wc.Add(st.Cornered)
	}
	return wc
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
	Keywords  ability.Keywords
}

// effectiveState reads a's effective characteristics now.
func (g *Game) effectiveState(a *Asset) EffectiveState {
	return EffectiveState{Power: g.Power(a), MaxHealth: g.MaxHealth(a), Keywords: g.Keywords(a)}
}
