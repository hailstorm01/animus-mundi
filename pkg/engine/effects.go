// Effects: running a resolved triggered ability's steps and their values (FR-TGT-003).

package engine

import (
	"fmt"
	"slices"

	"animus-mundi/pkg/engine/ability"
)

// resolution is one resolving ability as its steps see it: the stack item and,
// for each target slot, the chosen entities still legal after the second
// verification stage, in pick order (board directive #51), and how many things
// each step did so far, one entry per step (IfDone, SVY-R14).
type resolution struct {
	it     *StackItem
	chosen [][]*Asset
	done   []int
}

// resolveAbility runs a triggered ability's steps in order. It resolves even if
// its source has left the battlefield: untargeted steps verify nothing.
func (g *Game) resolveAbility(it *StackItem) {
	// Second verification stage (FR-TGT-003), entity by entity: one that is
	// gone or no longer legal is dropped, and the slot keeps the rest (partial
	// invalidation).
	r := &resolution{it: it, chosen: make([][]*Asset, len(it.Chosen))}
	for i, ids := range it.Chosen {
		for _, id := range ids {
			if a := g.Asset(id); a != nil && g.legalTarget(it.Controller, it.Source, it.Effect.Targets[i].Filter, a) {
				r.chosen[i] = append(r.chosen[i], a)
			}
		}
	}
	for i, st := range it.Effect.Steps {
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

// stepKill destroys each asset the step acts on at once, inline (FR-LAY-002,
// FR-ZON-002), not at the death sweep. A Headquarters is skipped: the match-loss
// check lives in deathSweep and no M2 card kills one. It returns the assets killed.
func (g *Game) stepKill(r *resolution, st ability.Step) int {
	var dying []*Asset
	for _, a := range g.stepAssets(r, st) {
		if a.Printed().Type != TypeHeadquarters {
			dying = append(dying, a)
		}
	}
	g.destroyAll(dying)
	return len(dying)
}

// stepDiscard has each step player discard cards of their choice from their own
// hand. Choosing from your own hand is not an inspection, so a hand of exactly
// that many cards is discarded without asking (FR-MOD-001 §1, §6; SVY-R16).
func (g *Game) stepDiscard(r *resolution, st ability.Step) int {
	n, total := g.value(r.it, st.Amount), 0
	for _, pid := range stepPlayers(r.it, st) {
		hand := g.Players[pid].Hand
		n := min(n, len(hand))
		if n <= 0 {
			continue
		}
		d := Decision{Player: pid, Kind: DecideHand, Prompt: "Discard a card", Min: n, Max: n}
		if n > 1 {
			d.Prompt = fmt.Sprintf("Discard %d cards", n)
		}
		for _, inst := range hand {
			d.Options = append(d.Options, Option{Label: inst.Card.Name, Ref: inst.ID})
		}
		// The options are fixed before any card moves, so take the cards first.
		var picked []*Instance
		for _, i := range g.choose(d) {
			picked = append(picked, hand[i])
		}
		for _, inst := range picked {
			g.takeCard(inst)
			g.putCard(inst, ZoneDiscard, false)
			g.logEvent(Event{Kind: "discard", Player: int(pid), Card: inst.Card.Name, Inst: inst.ID, Zone: ZoneHand.String()},
				"player %d discards %s", pid, inst.Card.Name)
			total++
		}
	}
	return total
}

// stepRelocate moves the trigger's own card out of discard (see relocateSelf).
// It returns 1 if the card moved, else 0.
func (g *Game) stepRelocate(r *resolution, st ability.Step) int {
	if g.relocateSelf(r.it, st) {
		return 1
	}
	return 0
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
		return g.pickAssets(r, st)
	}
	return nil
}

// pickAssets is the untargeted pick: the item's controller picks from every
// battlefield asset matching st.Filter now, oldest first. It is not a target, so
// Ward excludes nothing (SVY-R3, SVY-R14). It asks a DecidePick, which has the
// same FR-MOD-001 section 6 fallback as a target: UpTo 0 picks exactly one (a single option is not
// asked, a timeout picks the first in fixed order); UpTo N picks up to N, always
// asked, and a timeout picks none. With no match nothing is asked.
func (g *Game) pickAssets(r *resolution, st ability.Step) []*Asset {
	name := cardName(r.it.Inst)
	prompt := name + ": pick one"
	if st.UpTo > 0 {
		prompt = fmt.Sprintf("%s: pick up to %d", name, st.UpTo)
	}
	d, opts := g.assetDecision(r.it.Controller, DecidePick, prompt, st.UpTo, func(a *Asset) bool {
		return g.matchesFilter(r.it.Controller, r.it.Source, st.Filter, a)
	})
	if len(opts) == 0 {
		return nil
	}
	var out []*Asset
	for _, k := range g.choose(d) {
		out = append(out, opts[k])
	}
	return out
}

// relocateSelf moves a death trigger's own card from its owner's discard pile
// to the owner's hand or deck. It tracks the exact instance: if that instance
// is no longer in the discard pile, it does nothing (FR-ZON-004). Only defined
// for SubjectSelf on a Tracked item. It reports whether the card moved.
func (g *Game) relocateSelf(it *StackItem, st ability.Step) bool {
	if st.Subject != ability.SubjectSelf || !it.Tracked || it.Inst == nil {
		return false
	}
	owner := g.Players[it.Inst.Owner]
	if !slices.Contains(owner.Discard, it.Inst) {
		return false
	}
	g.takeCard(it.Inst)
	if st.Op == ability.OpReturnToHand {
		g.putCard(it.Inst, ZoneHand, false)
	} else {
		g.putCard(it.Inst, ZoneDeck, st.Bottom)
	}
	return true
}
