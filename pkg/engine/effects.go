// Effects: running a resolved triggered ability's steps and their values (FR-TGT-003).

package engine

import (
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
			if a := g.Asset(id); a != nil && g.legalTarget(it.Controller, it.Source, it.Trigger.Effect.Targets[i].Filter, a) {
				chosen[i] = append(chosen[i], a)
			}
		}
	}
	for _, st := range it.Trigger.Effect.Steps {
		who := []PlayerID{it.Controller}
		if st.Subject == ability.SubjectOpponents {
			who = []PlayerID{it.Controller.Opponent()}
		}
		switch st.Op {
		case ability.OpDealDamage:
			// One instance per asset, all dealt before the item's death sweep,
			// so the deaths they cause are simultaneous (FR-REP-004, FR-LAY-002).
			n, src := g.value(it, st.Amount), g.itemSource(it)
			for _, a := range g.stepAssets(it, st, chosen) {
				g.dealDamage(src, a, n, false)
			}
		case ability.OpHeal, ability.OpAddArmor:
			n := g.value(it, st.Amount)
			for _, a := range g.stepAssets(it, st, chosen) {
				if st.Op == ability.OpHeal {
					a.Damage = max(0, a.Damage-n)
				} else {
					a.Armor += n
				}
			}
		case ability.OpScratch:
			for _, pid := range who {
				g.scratch(pid, g.value(it, st.Amount))
			}
		case ability.OpDraw:
			for _, pid := range who {
				g.draw(pid, g.value(it, st.Amount))
			}
		case ability.OpReturnToHand, ability.OpPutOnDeck:
			g.relocateSelf(it, st)
		}
	}
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
// battlefield asset matching st.Filter now, oldest first. "All" is not a
// target, so Ward excludes nothing and nothing is asked (FR-TGT-002,
// FR-KWD-002h, SVY-R3).
func (g *Game) stepAssets(it *StackItem, st ability.Step, chosen [][]*Asset) []*Asset {
	switch st.Subject {
	case ability.SubjectTarget:
		if st.Target >= 0 && st.Target < len(chosen) {
			return chosen[st.Target]
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
// for SubjectSelf on a HookKilled trigger.
func (g *Game) relocateSelf(it *StackItem, st ability.Step) {
	if st.Subject != ability.SubjectSelf || it.Trigger.Hook != ability.HookKilled || it.Inst == nil {
		return
	}
	owner := g.Players[it.Inst.Owner]
	if !slices.Contains(owner.Discard, it.Inst) {
		return
	}
	owner.Discard = remove(owner.Discard, it.Inst)
	switch {
	case st.Op == ability.OpReturnToHand:
		owner.Hand = append(owner.Hand, it.Inst)
	case st.Bottom:
		owner.Deck = append(owner.Deck, it.Inst)
	default:
		owner.Deck = append([]*Instance{it.Inst}, owner.Deck...)
	}
}
