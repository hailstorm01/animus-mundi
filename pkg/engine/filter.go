// Filters: checking assets and cards in discard piles against an ability's filter (FR-TGT-001, FR-ZON-004, FR-KWD-002h).

package engine

import "animus-mundi/pkg/engine/ability"

// kindOf maps a printed card type to its filter kind.
func kindOf(t CardType) ability.Kinds {
	switch t {
	case TypeUnit:
		return ability.KindUnit
	case TypeStructure:
		return ability.KindStructure
	case TypeHeadquarters:
		return ability.KindHeadquarters
	case TypeManeuver:
		return ability.KindManeuver
	}
	return 0
}

// matchesCard reports whether card inst, in a discard pile, meets every
// criterion of the pile filter f for an ability controlled by ctl whose own
// card is src: allegiance by the card's owner, Kinds, Subtypes and Scratch by its
// printed characteristics (FR-ZON-004), Stat by its printed value, Other not
// src. f must be InPile.
func (g *Game) matchesCard(ctl PlayerID, src *Instance, f ability.Filter, inst *Instance) bool {
	return f.InPile() && g.cardMeets(ctl, src, f, inst)
}

// cardMeets reports whether card inst, wherever it is, meets every card-level
// criterion of f: Other (not src), Stat by its printed value, then allegiance
// by the card's owner, Kinds, Subtypes and Scratch by its printed
// characteristics. It reads no zone criterion. The look at the top cards
// (T-0222) checks a deck card with it.
func (g *Game) cardMeets(ctl PlayerID, src *Instance, f ability.Filter, inst *Instance) bool {
	if f.Other && inst == src {
		return false
	}
	if !withinStat(f, cardStat(f.Stat, inst.Card)) {
		return false
	}
	return matchesPrinted(ctl, f, inst.Owner, inst.Card)
}

// withinStat reports whether v meets f's stat criterion: true when f has none,
// else v is f.AtMost or less.
func withinStat(f ability.Filter, v int) bool {
	return f.Stat == ability.StatNone || v <= f.AtMost
}

// cardStat is a card's printed value of s (FR-ZON-004); 0 for StatNone.
func cardStat(s ability.Stat, c *Card) int {
	switch s {
	case ability.StatPower:
		return c.Power
	case ability.StatRequisition:
		return c.ReqCost
	}
	return 0
}

// assetStat is an asset's effective value of s (Layer 6, FR-TGT-001); 0 for
// StatNone. Requisition cost is the cost printed on what the asset is now (a
// face-down unit's is the Scratch Company face's, 0): no layer changes a
// Requisition cost, so printed is effective.
func (g *Game) assetStat(s ability.Stat, a *Asset) int {
	switch s {
	case ability.StatPower:
		return g.Power(a)
	case ability.StatRequisition:
		return a.Printed().ReqCost
	}
	return 0
}

// legalTarget reports whether battlefield asset a meets every criterion of f
// for an ability controlled by ctl whose source is asset src. An opponent's
// Ward asset is never legal (FR-KWD-002h).
func (g *Game) legalTarget(ctl PlayerID, src int, lock *LockedZone, f ability.Filter, a *Asset) bool {
	if a.Controller != ctl && g.Keywords(a).Ward {
		return false
	}
	return g.matchesFilterAt(ctl, src, lock, f, a)
}

// matchesPrinted is the card-level part of a filter, shared by assets and
// cards in piles: Allegiance by side, the player the entity belongs to (relative
// to ctl), then Kinds, Subtypes and Scratch by the card's printed characteristics.
func matchesPrinted(ctl PlayerID, f ability.Filter, side PlayerID, c *Card) bool {
	switch f.Allegiance {
	case ability.AllegianceAllied:
		if side != ctl {
			return false
		}
	case ability.AllegianceOpposing:
		if side == ctl {
			return false
		}
	}
	if f.Kinds != 0 && f.Kinds&kindOf(c.Type) == 0 {
		return false
	}
	if len(f.Subtypes) > 0 {
		found := false
		for _, st := range f.Subtypes {
			if c.HasSubtype(st) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	switch f.Scratch {
	case ability.ScratchOnly:
		return c.IsScratchCompany()
	case ability.ScratchExcluded:
		return !c.IsScratchCompany()
	}
	return true
}

// inLocked reports whether a is in the zone lock names: for Reserves or the
// Headquarters Zone also that zone's owner's, not the shared Main Theater's
// (FR-ZON-001). A nil lock holds for nothing.
func inLocked(a *Asset, lock *LockedZone) bool {
	if lock == nil || a.Zone != lock.Zone {
		return false
	}
	return lock.Zone == ZoneMainTheater || a.Controller == lock.Player
}

// matchesFilter is matchesFilterAt without a locked zone, for callers that have
// no stack item (costs, replacements).
func (g *Game) matchesFilter(ctl PlayerID, src int, f ability.Filter, a *Asset) bool {
	return g.matchesFilterAt(ctl, src, nil, f, a)
}

// matchesFilterAt reports whether battlefield asset a meets every criterion of f
// (allegiance relative to ctl, kinds, subtypes, Scratch, Other: not the source asset
// src, Stat: its effective value, Zone: currently in that battlefield zone, ZoneRef: in lock's zone, never the source's current one). Targeting adds Ward on top (legalTarget); an untargeted "all" step uses
// this alone.
func (g *Game) matchesFilterAt(ctl PlayerID, src int, lock *LockedZone, f ability.Filter, a *Asset) bool {
	if f.Other && a.ID == src {
		return false
	}
	if f.ZoneRef != ability.ZoneRefNone && !inLocked(a, lock) {
		return false
	}
	if f.InPile() || !deployedTo(a, f.Zone) {
		return false
	}
	if !withinStat(f, g.assetStat(f.Stat, a)) {
		return false
	}
	return matchesPrinted(ctl, f, a.Controller, a.Printed())
}
