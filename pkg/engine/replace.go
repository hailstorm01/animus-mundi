// Replacement effects: destruction replacement at the death sweep (FR-REP-001, SVY-R20 to SVY-R22).

package engine

import (
	"slices"

	"animus-mundi/pkg/engine/ability"
)

// findReplacement returns the source and the replacement that apply to a's
// destruction now: the first, in battlefield order, oldest source first, of the
// replacements printed by a face-up asset on the battlefield whose filter a
// matches. A source that is itself lethal still applies until the sweep
// destroys it. Letting a's controller choose between several is T-0100's.
func (g *Game) findReplacement(a *Asset) (*Asset, *ability.Replacement) {
	for _, src := range g.assetsByID() {
		if src.FaceDown || src.Printed().Ability == nil {
			continue
		}
		reps := src.Printed().Ability.Replacements
		for i := range reps {
			r := &reps[i]
			if r.Event == ability.WouldBeDestroyed && g.matchesFilter(src.Controller, src.ID, r.Filter, a) {
				return src, r
			}
		}
	}
	return nil, nil
}

// replacedQueue returns the assets of dying that a replacement applies to, in
// the order their deaths are replaced: the active player's first (SVY-R22), then
// the non-active player's, each player's oldest asset first (SVY-R20). No player
// orders them.
func (g *Game) replacedQueue(dying []*Asset) []*Asset {
	var queued [2][]*Asset
	for _, a := range dying {
		if src, _ := g.findReplacement(a); src != nil {
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
// stack, with its source as their source. It reports whether it did.
func (g *Game) replaceDeath(a *Asset) bool {
	src, rep := g.findReplacement(a)
	if src == nil {
		return false
	}
	g.logEvent(Event{Kind: "replace", Player: int(src.Controller), Source: src.ID, Target: a.ID},
		"%s replaces the destruction of %s", src, a)
	it := &StackItem{Kind: StackAbility, Controller: src.Controller, Source: src.ID, Inst: src.Inst,
		Effect: ability.Effect{Steps: rep.Steps}}
	g.runSteps(&resolution{it: it, replaced: a})
	return true
}

// relocateReplaced returns the replaced asset from the battlefield to its
// owner's hand or deck, as one thing done. Its last-known state is kept
// (FR-ZON-004). The event carries no Card: a face-down asset's card is private
// in a hand or deck.
func (g *Game) relocateReplaced(r *resolution, st ability.Step) int {
	as := g.stepAssets(r, st)
	for _, a := range as {
		to, text := ZoneHand, "%s returns to its owner's hand"
		if st.Op == ability.OpPutOnDeck {
			to, text = ZoneDeck, "%s is put into its owner's deck"
		}
		g.logEvent(Event{Kind: "leave", Player: -1, Target: a.ID, Zone: to.String()}, text, a)
		g.leaveBattlefield(a, g.effectiveState(a))
		g.putCard(a.Inst, to, st.Bottom)
	}
	return len(as)
}
