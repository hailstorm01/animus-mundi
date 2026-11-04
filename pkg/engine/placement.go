// Placement: placing an effect on the stack, with its modes and then its targets chosen on placement (FR-MOD-001 §5, FR-TGT-001, SVY-R3).

package engine

import (
	"fmt"
	"slices"

	"animus-mundi/pkg/engine/ability"
)

// placement is an effect being placed on the stack, as choosing its modes and
// targets sees it: who controls it, its source and the label its prompts use.
// A fired trigger carries one (pendingTrigger); a played Maneuver does too (maneuverPlacement).
type placement struct {
	controller PlayerID
	source     int       // source asset ID
	inst       *Instance // source card instance, tracked by identity (FR-ZON-004)
	label      string
	costs      []ability.Cost // additional costs paid with this placement (FR-TRN-003a); a trigger has none
	discards   []*Instance    // hand cards its discard cost picked on declaration, in pick order; nil without one
}

// chooseOnPlacement chooses what e needs on placement, in rules order: its
// modes first, if it has any, then the targets of the effect that leaves
// (FR-MOD-001 §5, SVY-R3). It returns the effect the stack item resolves and
// the chosen targets, one slice of Refs per target slot.
func (g *Game) chooseOnPlacement(pl placement, e ability.Effect) (ability.Effect, [][]int) {
	if len(e.Modes) > 0 {
		e = g.chooseModes(pl, e)
	}
	return e, g.chooseTargets(pl, e)
}

// declare runs the declare sequence for item, an effect declared by a manual
// action, in rules order: the discard cost's cards are picked, e's modes and
// targets are chosen and filled into item, item goes on the stack, and the
// additional costs are paid (FR-TRN-003, FR-TRN-003a).
func (g *Game) declare(pl placement, e ability.Effect, item *StackItem) {
	pl.discards = g.chooseCostDiscards(pl)
	item.Effect, item.Chosen = g.chooseOnPlacement(pl, e)
	g.push(item)
	g.payCosts(pl)
}

// maneuverPlacement is the placement of the Maneuver inst that pid enacts, for
// both its legality check and the choosing of its modes and targets.
func maneuverPlacement(pid PlayerID, inst *Instance) placement {
	pl := placement{controller: pid, inst: inst, label: inst.Card.Name}
	if inst.Card.Ability != nil {
		pl.costs = inst.Card.Ability.ExtraCosts
	}
	return pl
}

// slotsLegal reports whether the exactly-one slots of ts can all be filled at
// once now: one legal entity each, with every "another" slot's entity
// different from the one assigned to the slot it follows when that slot is
// exactly-one too (Target.Another, SVY-R34). It checks a joint assignment, in
// slot order, not each slot alone. An "up to N" slot never makes it false
// (design note Open Question 11), nor does a PerDiscard slot: its controller
// can discard none, which needs no target, and the count is not known before
// the discard is picked (SVY-R29). An "another" slot that follows one of them
// needs only a legal entity of its own.
func (g *Game) slotsLegal(pl placement, ts []ability.Target) bool {
	opts := make([][]Option, len(ts))
	for i, t := range ts {
		opts[i] = g.targetDecision(pl, i, t).Options
	}
	assigned := make([]int, len(ts))
	exactlyOne := func(t ability.Target) bool { return t.UpTo == 0 && !t.PerDiscard }
	var fill func(i int) bool
	fill = func(i int) bool {
		if i == len(ts) {
			return true
		}
		if !exactlyOne(ts[i]) {
			return fill(i + 1)
		}
		k, hasPrev := anotherOf(ts, i)
		hasPrev = hasPrev && exactlyOne(ts[k])
		for _, o := range opts[i] {
			if hasPrev && o.Ref == assigned[k] {
				continue
			}
			assigned[i] = o.Ref
			if fill(i + 1) {
				return true
			}
		}
		return false
	}
	return fill(0)
}

// placeable reports whether e can be placed with everything it requires
// (SVY-R29): every additional cost payable (costsPayable), then at least
// ChooseN legal modes, or a legal entity for each exactly-one slot. It applies
// to enacting a Maneuver only: a trigger is placed whatever it finds (SVY-R30,
// FR-TGT-003).
func (g *Game) placeable(pl placement, e ability.Effect) bool {
	if !g.costsPayable(pl) {
		return false
	}
	if len(e.Modes) > 0 {
		return len(g.legalModes(pl, e)) >= e.ChooseN
	}
	return g.slotsLegal(pl, e.Targets)
}

// legalModes returns the indices of e's legal modes, in printed order. A mode
// is legal if each of its exactly-one target slots has a legal entity now (slotsLegal).
func (g *Game) legalModes(pl placement, e ability.Effect) []int {
	var legal []int
	for k, m := range e.Modes {
		if g.slotsLegal(pl, m.Targets) {
			legal = append(legal, k)
		}
	}
	return legal
}

// chooseModes asks pl's controller which modes of e to use, when it is placed
// and before its targets (FR-MOD-001 §5, SVY-R3), and returns the effect the
// item resolves: the chosen modes joined in printed order. Only legal modes
// (legalModes) are offered, so a timeout takes the first listed legal ones
// (§6) and a single legal mode is answered automatically (§1). With no legal
// mode nothing is asked and the effect is empty. With fewer legal modes than
// ChooseN all of them are chosen; an enacted Maneuver never gets here with
// fewer than ChooseN (placeable, SVY-R29), only a trigger does (SVY-R30).
func (g *Game) chooseModes(pl placement, e ability.Effect) ability.Effect {
	legal := g.legalModes(pl, e)
	if len(legal) == 0 {
		return ability.Effect{}
	}
	n := min(e.ChooseN, len(legal))
	prompt := fmt.Sprintf("%s: choose %d", pl.label, e.ChooseN)
	if e.ChooseN == 1 {
		prompt = pl.label + ": choose one"
	}
	d := Decision{Player: pl.controller, Kind: DecideMode, Prompt: prompt, Min: n, Max: n}
	for _, k := range legal {
		d.Options = append(d.Options, Option{Label: fmt.Sprintf("mode %d", k+1), Ref: pl.source})
	}
	picked := g.choose(d)
	slices.Sort(picked)
	modes := make([]ability.Effect, len(picked))
	for i, j := range picked {
		modes[i] = e.Modes[legal[j]]
	}
	return joinModes(modes)
}

// joinModes joins modes, in the order given, into the one effect the stack
// item resolves: each mode's targets and steps follow the last mode's, with
// their slot references renumbered for the joined Targets (shiftSlots).
func joinModes(modes []ability.Effect) ability.Effect {
	var joined ability.Effect
	for _, m := range modes {
		shifted := shiftSlots(m, len(joined.Targets))
		joined.Targets = append(joined.Targets, shifted.Targets...)
		joined.Steps = append(joined.Steps, shifted.Steps...)
	}
	return joined
}

// shiftSlots returns a copy of mode m with every reference to a target slot
// moved up by shift: a target's Another, and a step's Target and Dealer. New
// slices: a mode's own Targets and Steps belong to the card's definition and
// must never be written to.
func shiftSlots(m ability.Effect, shift int) ability.Effect {
	out := ability.Effect{Targets: slices.Clone(m.Targets), Steps: slices.Clone(m.Steps)}
	for i := range out.Targets {
		if out.Targets[i].Another > 0 {
			out.Targets[i].Another += shift
		}
	}
	for i := range out.Steps {
		st := &out.Steps[i]
		if st.Subject == ability.SubjectTarget {
			st.Target += shift
		}
		if st.Dealer > 0 {
			st.Dealer += shift
		}
	}
	return out
}
