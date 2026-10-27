// Effects: running a resolved triggered ability's steps and their values (FR-TGT-003).

package engine

import (
	"fmt"
	"slices"

	"animus-mundi/pkg/engine/ability"
)

// resolveAbility runs a triggered ability's steps in order. It resolves even if
// its source has left the battlefield: untargeted steps verify nothing.
func (g *Game) resolveAbility(it *StackItem) {
	// Second verification stage (FR-TGT-003), entity by entity: one that is
	// gone or no longer legal is dropped, and the slot keeps the rest (partial
	// invalidation).
	chosen := make([][]*Asset, len(it.Chosen))
	for i, ids := range it.Chosen {
		for _, id := range ids {
			if a := g.Asset(id); a != nil && g.legalTarget(it.Controller, it.Source, it.Effect.Targets[i].Filter, a) {
				chosen[i] = append(chosen[i], a)
			}
		}
	}
	for _, st := range it.Effect.Steps {
		switch st.Op {
		case ability.OpDealDamage:
			g.stepDealDamage(it, st, chosen)
		case ability.OpHeal:
			g.stepHeal(it, st, chosen)
		case ability.OpAddArmor:
			g.stepAddArmor(it, st, chosen)
		case ability.OpScratch:
			g.stepScratch(it, st, chosen)
		case ability.OpDraw:
			g.stepDraw(it, st, chosen)
		case ability.OpDiscard:
			g.stepDiscard(it, st, chosen)
		case ability.OpModifyStat:
			g.stepModifyStat(it, st, chosen)
		case ability.OpReturnToHand, ability.OpPutOnDeck:
			g.stepRelocate(it, st, chosen)
		}
	}
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
func (g *Game) stepDealDamage(it *StackItem, st ability.Step, chosen [][]*Asset) {
	n, src := g.value(it, st.Amount), g.itemSource(it)
	for _, a := range g.stepAssets(it, st, chosen) {
		g.dealDamage(src, a, n, false)
	}
}

// stepHeal removes damage from each asset the step acts on.
func (g *Game) stepHeal(it *StackItem, st ability.Step, chosen [][]*Asset) {
	n := g.value(it, st.Amount)
	for _, a := range g.stepAssets(it, st, chosen) {
		a.Damage = max(0, a.Damage-n)
	}
}

// stepAddArmor gives each asset the step acts on Armor.
func (g *Game) stepAddArmor(it *StackItem, st ability.Step, chosen [][]*Asset) {
	n := g.value(it, st.Amount)
	for _, a := range g.stepAssets(it, st, chosen) {
		a.Armor += n
	}
}

// stepScratch has each step player scratch.
func (g *Game) stepScratch(it *StackItem, st ability.Step, _ [][]*Asset) {
	for _, pid := range stepPlayers(it, st) {
		g.scratch(pid, g.value(it, st.Amount))
	}
}

// stepDraw has each step player draw.
func (g *Game) stepDraw(it *StackItem, st ability.Step, _ [][]*Asset) {
	for _, pid := range stepPlayers(it, st) {
		g.draw(pid, g.value(it, st.Amount))
	}
}

// stepModifyStat gives each Subject +Amount power until end of turn: the Layer 5
// TempPower, cleared in Sub-Phase 4c (FR-LAY-001, FR-LAY-003). Power clamps at 0.
func (g *Game) stepModifyStat(it *StackItem, st ability.Step, chosen [][]*Asset) {
	for _, a := range g.stepAssets(it, st, chosen) {
		a.TempPower += g.value(it, st.Amount)
	}
}

// stepDiscard has each step player discard cards of their choice from their own
// hand. Choosing from your own hand is not an inspection, so a hand of exactly
// that many cards is discarded without asking (FR-MOD-001 §1, §6; SVY-R16).
func (g *Game) stepDiscard(it *StackItem, st ability.Step, _ [][]*Asset) {
	n := g.value(it, st.Amount)
	for _, pid := range stepPlayers(it, st) {
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
		}
	}
}

// stepRelocate moves the trigger's own card out of discard (see relocateSelf).
func (g *Game) stepRelocate(it *StackItem, st ability.Step, _ [][]*Asset) { g.relocateSelf(it, st) }

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
func (g *Game) stepAssets(it *StackItem, st ability.Step, chosen [][]*Asset) []*Asset {
	switch st.Subject {
	case ability.SubjectTarget:
		if st.Target >= 0 && st.Target < len(chosen) {
			return chosen[st.Target]
		}
	case ability.SubjectSelf:
		// The source itself, if it is still on the battlefield; once it has
		// left, the step fizzles (FR-TGT-003).
		if a := g.Asset(it.Source); a != nil {
			return []*Asset{a}
		}
	case ability.SubjectAll:
		var out []*Asset
		for _, a := range g.assetsByID() {
			if g.matchesFilter(it.Controller, it.Source, st.Filter, a) {
				out = append(out, a)
			}
		}
		return out
	}
	return nil
}

// relocateSelf moves a death trigger's own card from its owner's discard pile
// to the owner's hand or deck. It tracks the exact instance: if that instance
// is no longer in the discard pile, it does nothing (FR-ZON-004). Only defined
// for SubjectSelf on a Tracked item.
func (g *Game) relocateSelf(it *StackItem, st ability.Step) {
	if st.Subject != ability.SubjectSelf || !it.Tracked || it.Inst == nil {
		return
	}
	owner := g.Players[it.Inst.Owner]
	if !slices.Contains(owner.Discard, it.Inst) {
		return
	}
	g.takeCard(it.Inst)
	if st.Op == ability.OpReturnToHand {
		g.putCard(it.Inst, ZoneHand, false)
	} else {
		g.putCard(it.Inst, ZoneDeck, st.Bottom)
	}
}
