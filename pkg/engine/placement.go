// Placement: placing an effect on the stack, with its modes and then its targets chosen on placement (FR-MOD-001 §5, FR-TGT-001, SVY-R3).

package engine

import (
	"fmt"
	"slices"

	"animus-mundi/pkg/engine/ability"
)

// placement is an effect being placed on the stack, as choosing its modes and
// targets sees it: who controls it, its source and the label its prompts use.
// A fired trigger carries one (pendingTrigger); a played Maneuver will too.
type placement struct {
	controller PlayerID
	source     int       // source asset ID
	inst       *Instance // source card instance, tracked by identity (FR-ZON-004)
	label      string
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

// legalModes returns the indices of e's legal modes, in printed order. A mode
// is legal if each of its exactly-one target slots has a legal entity now; an
// "up to N" slot never makes it illegal (design note Open Question 11).
func (g *Game) legalModes(pl placement, e ability.Effect) []int {
	var legal []int
	for k, m := range e.Modes {
		ok := true
		for i, t := range m.Targets {
			if t.UpTo == 0 && len(g.targetDecision(pl, i, t).Options) == 0 {
				ok = false
				break
			}
		}
		if ok {
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
// ChooseN all of them are chosen.
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
	// New slices throughout: a mode's own Targets and Steps belong to the
	// card's definition and must never be appended to.
	var joined ability.Effect
	for _, j := range picked {
		m := e.Modes[legal[j]]
		shift := len(joined.Targets)
		joined.Targets = append(joined.Targets, m.Targets...)
		for _, st := range m.Steps {
			if st.Subject == ability.SubjectTarget {
				st.Target += shift
			}
			joined.Steps = append(joined.Steps, st)
		}
	}
	return joined
}
