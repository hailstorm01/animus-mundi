// Replacement effects: destruction replacement at the death sweep (FR-REP-001, SVY-R5).

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
// the order their deaths are replaced: each player orders the deaths of the
// assets they control, the active player's first (SVY-R5). The fallback order
// is oldest asset first.
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

		if len(mine) >= 2 {
			d := Decision{Player: pid, Kind: DecideOrder, Prompt: "order the deaths your replacement effects apply to (first pick is replaced first)", Min: len(mine), Max: len(mine)}
			for _, a := range mine {
				d.Options = append(d.Options, Option{Label: a.String(), Ref: a.ID})
			}
			for _, i := range g.choose(d) {
				out = append(out, mine[i])
			}
			continue
		}
		out = append(out, mine...)
	}
	return out
}

// replaceDeaths replaces the deaths of dying that a replacement applies to, one at
// a time in their controller's order (replacedQueue, SVY-R5), and calls between
// after each one it replaces. It returns the rest, to be destroyed together: the
// assets of dying still on the battlefield, for which dies still holds, whose
// death it did not replace.
func (g *Game) replaceDeaths(dying []*Asset, dies func(*Asset) bool, between func()) []*Asset {
	replaced := map[int]bool{}
	for _, a := range g.replacedQueue(dying) {
		if g.Asset(a.ID) != a || !dies(a) {
			continue
		}
		if g.replaceDeath(a) {
			replaced[a.ID] = true
			between()
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
