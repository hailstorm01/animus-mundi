// Replacement effects: destruction replacement at the death sweep (FR-REP-001, SVY-R20 to SVY-R22).

package engine

import (
	"fmt"
	"slices"

	"animus-mundi/pkg/engine/ability"
)

// replacementOption is one replacement that applies to a destruction, with the
// asset that prints it.
type replacementOption struct {
	src *Asset
	rep *ability.Replacement
}

// replacementsFor returns every replacement that applies to a's destruction now,
// oldest source first, each source's replacements in printed order: those printed
// by a face-up asset on the battlefield whose filter a matches. A source that is
// itself lethal still applies until the sweep destroys it (SVY-R20).
func (g *Game) replacementsFor(a *Asset) []replacementOption {
	var out []replacementOption
	for _, src := range g.assetsByID() {
		if src.FaceDown || src.Printed().Ability == nil {
			continue
		}
		reps := src.Printed().Ability.Replacements
		for i := range reps {
			r := &reps[i]
			if r.Event == ability.WouldBeDestroyed && g.matchesFilter(src.Controller, src.ID, r.Filter, a) {
				out = append(out, replacementOption{src: src, rep: r})
			}
		}
	}
	return out
}

// replacedQueue returns the assets of dying that a replacement applies to, in
// the order their deaths are replaced: the active player's first (SVY-R22), then
// the non-active player's, each player's oldest asset first (SVY-R20). No player
// orders them.
func (g *Game) replacedQueue(dying []*Asset) []*Asset {
	var queued [2][]*Asset
	for _, a := range dying {
		if len(g.replacementsFor(a)) > 0 {
			queued[a.Controller] = append(queued[a.Controller], a)
		}
	}
	var out []*Asset
	for _, pid := range []PlayerID{g.Active, g.Active.Opponent()} {
		mine := queued[pid]
		slices.SortFunc(mine, func(x, y *Asset) int { return x.ID - y.ID })
		out = append(out, mine...)
	}
	return out
}

// replaceDeaths replaces the deaths of dying that a replacement applies to, one at a
// time in replacedQueue's order, with nothing destroyed between them: a source that
// is lethal, or that a byproduct leaves lethal, still applies (SVY-R20, SVY-R21).
// It returns the rest, to be destroyed together: the assets of dying still on the
// battlefield, for which dies still holds, whose death it did not replace.
func (g *Game) replaceDeaths(dying []*Asset, dies func(*Asset) bool) []*Asset {
	replaced := map[int]bool{}
	for _, a := range g.replacedQueue(dying) {
		if g.Asset(a.ID) != a || !dies(a) {
			continue
		}
		if g.replaceDeath(a) {
			replaced[a.ID] = true
		}
	}
	var rest []*Asset
	for _, a := range dying {
		if !replaced[a.ID] && g.Asset(a.ID) == a && dies(a) {
			rest = append(rest, a)
		}
	}
	return rest
}

// replaceDeath replaces a's destruction in full, if a replacement still applies:
// the destroy never happens, and the replacement's steps run inline, off the
// stack, with its source as their source. It reports whether it did. When several
// apply, a's controller chooses one (chooseReplacement, SVY-R23).
func (g *Game) replaceDeath(a *Asset) bool {
	opts := g.replacementsFor(a)
	if len(opts) == 0 {
		return false
	}
	pick := opts[0]
	if len(opts) > 1 {
		pick = g.chooseReplacement(a, opts)
	}
	src, rep := pick.src, pick.rep
	g.logEvent(Event{Kind: "replace", Player: int(src.Controller), Source: src.ID, Target: a.ID},
		"%s replaces the destruction of %s", src, a)
	it := &StackItem{Kind: StackAbility, Controller: src.Controller, Source: src.ID, Inst: src.Inst,
		Effect: ability.Effect{Steps: rep.Steps}}
	g.runSteps(&resolution{it: it, replaced: a})
	return true
}

// chooseReplacement asks a's controller which of several replacements applies to
// its destruction (FR-REP-001, SVY-R23). A timeout picks the first source in the
// FR-MOD-001 §6 fixed order, a source's own replacements in printed order
// (DecideReplacement's fallback).
func (g *Game) chooseReplacement(a *Asset, opts []replacementOption) replacementOption {
	options := make([]Option, len(opts))
	for i, o := range opts {
		label := o.src.String()
		if n := len(o.src.Printed().Ability.Replacements); n > 1 {
			for j := range o.src.Printed().Ability.Replacements {
				if &o.src.Printed().Ability.Replacements[j] == o.rep {
					label = fmt.Sprintf("%s (replacement %d)", o.src, j+1)
				}
			}
		}
		options[i] = Option{Label: label, Ref: o.src.ID}
	}
	picks := g.choose(Decision{Player: a.Controller, Kind: DecideReplacement,
		Prompt:  fmt.Sprintf("choose which replacement applies to the destruction of %s", a),
		Options: options, Min: 1, Max: 1})
	return opts[picks[0]]
}
