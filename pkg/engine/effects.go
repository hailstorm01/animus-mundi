// Effects: resolving a stack item's effect, its steps in order and their values (FR-TGT-003). Untargeted picks are targeting.go's.

package engine

import (
	"slices"

	"animus-mundi/pkg/engine/ability"
)

// resolution is one resolving ability as its steps see it: the stack item and,
// for each target slot, the chosen entities still legal after the second
// verification stage, in pick order (board directive #51): chosen holds an
// asset slot's assets, cards a pile slot's cards (the other is empty). Then how many things
// each step did so far, one entry per step (IfDone, SVY-R14).
type resolution struct {
	it     *StackItem
	chosen [][]*Asset
	cards  [][]*Instance
	done   []int
	// replaced is the asset whose event a replacement replaces: SubjectReplaced (replace.go).
	replaced *Asset
}

// resolveAbility runs a triggered ability's steps in order. It resolves even if
// its source has left the battlefield: untargeted steps verify nothing.
func (g *Game) resolveAbility(it *StackItem) {
	// Second verification stage (FR-TGT-003), entity by entity: one that is
	// gone or no longer legal is dropped, and the slot keeps the rest (partial
	// invalidation).
	r := &resolution{it: it, chosen: make([][]*Asset, len(it.Chosen)), cards: make([][]*Instance, len(it.Chosen))}
	for i, ids := range it.Chosen {
		for _, id := range ids {
			if it.PileSlot(i) {
				if c := g.pileTarget(it.Controller, it.Inst, it.Effect.Targets[i].Filter, id); c != nil {
					r.cards[i] = append(r.cards[i], c)
				}
				continue
			}
			if a := g.Asset(id); a != nil && g.legalTarget(it.Controller, it.Source, it.Effect.Targets[i].Filter, a) {
				r.chosen[i] = append(r.chosen[i], a)
			}
		}
	}
	g.runSteps(r)
}

// runSteps runs the steps of r's item in order, recording how many things each
// did. A resolving ability and an inline replacement (replace.go) share it.
func (g *Game) runSteps(r *resolution) {
	for i, st := range r.it.Effect.Steps {
		n := 0
		// "If you do": skipped, and counted as 0, unless the step just before it did something.
		if !st.IfDone || (i > 0 && r.done[i-1] > 0) {
			n = g.runStep(r, st)
		}
		r.done = append(r.done, n)
	}
}

// runStep runs one step and returns how many things it did.
func (g *Game) runStep(r *resolution, st ability.Step) int {
	switch st.Op {
	case ability.OpDealDamage:
		return g.stepDealDamage(r, st)
	case ability.OpHeal:
		return g.stepHeal(r, st)
	case ability.OpAddArmor:
		return g.stepAddArmor(r, st)
	case ability.OpScratch:
		return g.stepScratch(r, st)
	case ability.OpDraw:
		return g.stepDraw(r, st)
	case ability.OpDiscard:
		return g.stepDiscard(r, st)
	case ability.OpModifyStat:
		return g.stepModifyStat(r, st)
	case ability.OpKill:
		return g.stepKill(r, st)
	case ability.OpReturnToHand, ability.OpPutOnDeck:
		return g.stepRelocate(r, st)
	}
	return 0
}

// stepPlayers returns the players a player step acts on: the opponent for
// SubjectOpponents, else the item's controller.
func stepPlayers(it *StackItem, st ability.Step) []PlayerID {
	if st.Subject == ability.SubjectOpponents {
		return []PlayerID{it.Controller.Opponent()}
	}
	return []PlayerID{it.Controller}
}

// stepDealDamage deals one instance per asset, all before the item's death
// sweep, so the deaths they cause are simultaneous (FR-REP-004, FR-LAY-002).
// Like every step it returns how many things it did: here, the assets it hit.
func (g *Game) stepDealDamage(r *resolution, st ability.Step) int {
	n, src := g.value(r.it, st.Amount), g.itemSource(r.it)
	as := g.stepAssets(r, st)
	for _, a := range as {
		g.dealDamage(src, a, n, false)
	}
	return len(as)
}

// stepHeal removes damage from each asset the step acts on.
func (g *Game) stepHeal(r *resolution, st ability.Step) int {
	n := g.value(r.it, st.Amount)
	as := g.stepAssets(r, st)
	for _, a := range as {
		a.Damage = max(0, a.Damage-n)
	}
	return len(as)
}

// stepAddArmor gives each asset the step acts on Armor.
func (g *Game) stepAddArmor(r *resolution, st ability.Step) int {
	n := g.value(r.it, st.Amount)
	as := g.stepAssets(r, st)
	for _, a := range as {
		a.Armor += n
	}
	return len(as)
}

// stepScratch has each step player scratch. It returns the cards that left the
// step players' decks.
func (g *Game) stepScratch(r *resolution, st ability.Step) int {
	return g.deckDrop(r, st, g.scratch)
}

// stepDraw has each step player draw. It returns the cards that left the step
// players' decks: a draw from an empty deck counts nothing.
func (g *Game) stepDraw(r *resolution, st ability.Step) int {
	return g.deckDrop(r, st, g.draw)
}

// deckDrop runs take for each step player and returns how many cards left their decks.
func (g *Game) deckDrop(r *resolution, st ability.Step, take func(PlayerID, int)) int {
	total := 0
	for _, pid := range stepPlayers(r.it, st) {
		before := len(g.Players[pid].Deck)
		take(pid, g.value(r.it, st.Amount))
		total += before - len(g.Players[pid].Deck)
	}
	return total
}

// stepModifyStat gives each Subject +Amount power until end of turn: the Layer 5
// TempPower, cleared in Sub-Phase 4c (FR-LAY-001, FR-LAY-003). Power clamps at 0.
func (g *Game) stepModifyStat(r *resolution, st ability.Step) int {
	as := g.stepAssets(r, st)
	for _, a := range as {
		a.TempPower += g.value(r.it, st.Amount)
	}
	return len(as)
}

// stepKill kills each step asset except a Headquarters, as one simultaneous event.
// A death a replacement applies to is replaced instead, in its controller's order,
// before the rest are destroyed (replaceDeaths), with no sweep between: a source a
// byproduct leaves lethal applies to every death of this step and is destroyed at
// the sweep after the item (design note Open Question 13). A replaced kill still
// counts as done (SVY-R15). It returns how many it killed.
func (g *Game) stepKill(r *resolution, st ability.Step) int {
	var dying []*Asset
	for _, a := range g.stepAssets(r, st) {
		if a.Printed().Type != TypeHeadquarters {
			dying = append(dying, a)
		}
	}
	g.destroyAll(g.replaceDeaths(dying, func(*Asset) bool { return true }, func() {}))
	return len(dying)
}

// stepDiscard has each step player discard Amount cards of their choice from
// their own hand (discardFromHand). It returns how many cards they discarded.
func (g *Game) stepDiscard(r *resolution, st ability.Step) int {
	n, total := g.value(r.it, st.Amount), 0
	for _, pid := range stepPlayers(r.it, st) {
		total += len(g.discardFromHand(pid, n))
	}
	return total
}

// stepRelocate moves each card of stepCards to its owner's hand or deck (the
// bottom if st.Bottom), and returns how many moved.
func (g *Game) stepRelocate(r *resolution, st ability.Step) int {
	if st.Subject == ability.SubjectReplaced {
		return g.relocateReplaced(r, st)
	}
	n := 0
	for _, inst := range g.stepCards(r, st) {
		if !g.takeCard(inst) {
			continue
		}
		if st.Op == ability.OpReturnToHand {
			g.putCard(inst, ZoneHand, false)
		} else {
			g.putCard(inst, ZoneDeck, st.Bottom)
		}
		n++
	}
	return n
}

// stepCards returns the cards a card step acts on. SubjectSelf is a death
// trigger's own card: a Tracked item whose exact instance is still in its
// owner's discard pile (FR-ZON-004). SubjectPick with a pile filter is the
// cards picked from discard piles, in pick order. SubjectTarget is the cards
// chosen for its pile slot that are still legal. Anything else has none.
func (g *Game) stepCards(r *resolution, st ability.Step) []*Instance {
	switch st.Subject {
	case ability.SubjectSelf:
		it := r.it
		if it.Tracked && it.Inst != nil && slices.Contains(g.Players[it.Inst.Owner].Discard, it.Inst) {
			return []*Instance{it.Inst}
		}
	case ability.SubjectTarget:
		if st.Target >= 0 && st.Target < len(r.cards) {
			return r.cards[st.Target]
		}
	case ability.SubjectPick:
		if st.Filter.InPile() {
			return g.pickCards(r, st)
		}
	}
	return nil
}

// value reads v for it as its step resolves (SVY-R10). Source power is the
// source's effective power if it is on the battlefield, else the effective
// power recorded as it left (FR-ZON-004), else 0.
func (g *Game) value(it *StackItem, v ability.Value) int {
	if v.From != ability.ValueSourcePower {
		return v.N
	}
	if a := g.Asset(it.Source); a != nil {
		return g.Power(a)
	}
	if k := g.LastKnownEffective(it.Source); k != nil {
		return k.Power
	}
	return 0
}

// itemSource returns the item's source asset, or its last-known record once it has
// left the battlefield, so a damage event still names and attributes it
// (FR-REP-004 Stage 5). It is nil only for a source that was never an asset.
func (g *Game) itemSource(it *StackItem) *Asset {
	if a := g.Asset(it.Source); a != nil {
		return a
	}
	return g.LastKnown(it.Source)
}

// stepAssets returns the assets an asset step acts on: the entities chosen for
// its slot that are still legal, in pick order, or for SubjectAll every
// battlefield asset matching st.Filter now, oldest first, or for SubjectSelf the
// source if it is still on the battlefield. "All" is not a
// target, so Ward excludes nothing and nothing is asked (FR-TGT-002,
// FR-KWD-002h, SVY-R3).
func (g *Game) stepAssets(r *resolution, st ability.Step) []*Asset {
	switch st.Subject {
	case ability.SubjectTarget:
		if st.Target >= 0 && st.Target < len(r.chosen) {
			return r.chosen[st.Target]
		}
	case ability.SubjectReplaced:
		if a := r.replaced; a != nil && g.Asset(a.ID) == a {
			return []*Asset{a}
		}
	case ability.SubjectSelf:
		// The source itself, if it is still on the battlefield; once it has
		// left, the step fizzles (FR-TGT-003).
		if a := g.Asset(r.it.Source); a != nil {
			return []*Asset{a}
		}
	case ability.SubjectAll:
		var out []*Asset
		for _, a := range g.assetsByID() {
			if g.matchesFilter(r.it.Controller, r.it.Source, st.Filter, a) {
				out = append(out, a)
			}
		}
		return out
	case ability.SubjectPick:
		if !st.Filter.InPile() {
			return g.pickAssets(r, st)
		}
	}
	return nil
}
