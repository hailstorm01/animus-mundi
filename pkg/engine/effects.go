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

// resolveAbility runs the stack item's effect, a triggered ability's or an
// enacted Maneuver's, step by step in order. It resolves even if
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
		if g.stepRuns(r, i, st) {
			n = g.runStep(r, st)
		}
		r.done = append(r.done, n)
	}
}

// stepRuns reports whether step i may run: every gate it sets must hold. A
// skipped step counts as 0 in r.done. Each gate is read as the step is reached.
func (g *Game) stepRuns(r *resolution, i int, st ability.Step) bool {
	// "If you do": the step just before it must have done something.
	if st.IfDone && (i == 0 || r.done[i-1] == 0) {
		return false
	}
	// "If you have no cards in hand": the controller's hand, never an opponent's.
	if st.IfHandEmpty && len(g.Players[r.it.Controller].Hand) != 0 {
		return false
	}
	return true
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
	case ability.OpMoveZone:
		return g.stepMoveZone(r, st)
	}
	return 0
}

// stepMoveZone moves each unit of stepAssets to st.To in its controller's
// battlefield and returns how many moved. A Structure, a Headquarters and a
// unit already there are skipped and don't count (FR-ZON-002). The unit keeps
// its state (FR-ZON-004); moving is not deploying, so no deploy trigger fires
// (SVY-R1). The event carries no Card, as stepRelocate's leave doesn't.
func (g *Game) stepMoveZone(r *resolution, st ability.Step) int {
	to, ok := battlefieldZone(st.To)
	if !ok {
		return 0
	}
	n := 0
	for _, a := range g.stepAssets(r, st) {
		if a.Printed().Type != TypeUnit || a.Zone == to {
			continue
		}
		g.moveOnBattlefield(a, to)
		g.logEvent(Event{Kind: EventMove, Player: -1, Target: a.ID, Zone: to.String()}, "%s moves to %s", a, to)
		n++
	}
	return n
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
// With a chosen dealer (Step.Dealer) that asset deals the damage, at its current
// power; with none left the step deals nothing and logs no event (SVY-R27).
func (g *Game) stepDealDamage(r *resolution, st ability.Step) int {
	src, dealer := g.itemSource(r.it), (*Asset)(nil)
	if slot, ok := st.DealerSlot(); ok {
		ds := g.slotAssets(r, slot)
		if len(ds) == 0 {
			return 0
		}
		dealer = ds[0]
		src = dealer
	}
	n := g.value(r.it, st.Amount, dealer)
	as := g.stepAssets(r, st)
	for _, a := range as {
		g.dealDamage(src, a, n, false)
	}
	return len(as)
}

// stepHeal removes damage from each asset the step acts on.
func (g *Game) stepHeal(r *resolution, st ability.Step) int {
	n := g.value(r.it, st.Amount, nil)
	as := g.stepAssets(r, st)
	for _, a := range as {
		a.Damage = max(0, a.Damage-n)
	}
	return len(as)
}

// stepAddArmor gives each asset the step acts on Armor.
func (g *Game) stepAddArmor(r *resolution, st ability.Step) int {
	n := g.value(r.it, st.Amount, nil)
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
		take(pid, g.value(r.it, st.Amount, nil))
		total += before - len(g.Players[pid].Deck)
	}
	return total
}

// stepModifyStat gives each Subject +Amount power until end of turn: the Layer 5
// TempPower, cleared in Sub-Phase 4c (FR-LAY-001, FR-LAY-003). Power clamps at 0.
func (g *Game) stepModifyStat(r *resolution, st ability.Step) int {
	as := g.stepAssets(r, st)
	for _, a := range as {
		a.TempPower += g.value(r.it, st.Amount, nil)
	}
	return len(as)
}

// stepKill kills the step assets as one moment (kill). It returns how many it killed.
func (g *Game) stepKill(r *resolution, st ability.Step) int {
	return g.kill(g.stepAssets(r, st))
}

// stepDiscard has each step player discard Amount cards of their choice from
// their own hand (discardFromHand). It returns how many cards they discarded.
func (g *Game) stepDiscard(r *resolution, st ability.Step) int {
	n, total := g.value(r.it, st.Amount, nil), 0
	for _, pid := range stepPlayers(r.it, st) {
		total += len(g.discardFromHand(pid, n, false))
	}
	return total
}

// relocateTo maps a relocate step to where its cards go: its owner's hand, or
// its owner's deck (the bottom if st.Bottom).
func relocateTo(st ability.Step) (to Zone, bottom bool) {
	if st.Op == ability.OpPutOnDeck {
		return ZoneDeck, st.Bottom
	}
	return ZoneHand, false
}

// stepRelocate moves each asset of stepAssets, then each card of stepCards, to
// its owner's hand or deck, and returns how many moved. An asset leaves the
// battlefield with its last-known state (FR-ZON-004); the event carries no
// Card, as a face-down asset's card is private in a hand or deck.
func (g *Game) stepRelocate(r *resolution, st ability.Step) int {
	to, bottom := relocateTo(st)
	n := 0
	for _, a := range g.stepAssets(r, st) {
		text := "%s returns to its owner's hand"
		if to == ZoneDeck {
			text = "%s is put into its owner's deck"
		}
		g.logEvent(Event{Kind: EventLeave, Player: -1, Target: a.ID, Zone: to.String()}, text, a)
		g.leaveBattlefield(a, g.effectiveState(a))
		g.putCard(a.Inst, to, bottom)
		n++
	}
	for _, inst := range g.stepCards(r, st) {
		if !g.takeCard(inst) {
			continue
		}
		g.putCard(inst, to, bottom)
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

// slotAssets returns slot i's chosen assets that are still on the battlefield,
// in pick order, or nil for an index out of range (FR-ZON-004, SVY-R27).
func (g *Game) slotAssets(r *resolution, i int) []*Asset {
	if i < 0 || i >= len(r.chosen) {
		return nil
	}
	var out []*Asset
	for _, a := range r.chosen[i] {
		if g.Asset(a.ID) == a {
			out = append(out, a)
		}
	}
	return out
}

// value reads v for it as its step resolves (SVY-R10). With a nil dealer,
// source power is the source's effective power if it is on the battlefield,
// else the effective power recorded as it left (FR-ZON-004), else 0. With a
// chosen dealer it is that asset's current power, never a last-known value
// (SVY-R27).
func (g *Game) value(it *StackItem, v ability.Value, dealer *Asset) int {
	if v.From != ability.ValueSourcePower {
		return v.N
	}
	if dealer != nil {
		return g.Power(dealer)
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
// its slot that are still legal and still on the battlefield, in pick order (a
// target an earlier step or mode removed is skipped: FR-ZON-004, FR-TGT-003
// partial invalidation), or for SubjectAll every
// battlefield asset matching st.Filter now, oldest first, or for SubjectSelf the
// source if it is still on the battlefield. "All" is not a
// target, so Ward excludes nothing and nothing is asked (FR-TGT-002,
// FR-KWD-002h, SVY-R3).
func (g *Game) stepAssets(r *resolution, st ability.Step) []*Asset {
	switch st.Subject {
	case ability.SubjectTarget:
		return g.slotAssets(r, st.Target)
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
