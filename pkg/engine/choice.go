package engine

import (
	"math/rand/v2"
	"slices"
)

// DecisionKind says what a Decision asks, and so which FR-MOD-001 §6 fallback applies.
type DecisionKind int

const (
	DecideTarget      DecisionKind = iota // mandatory targets: fallback picks the first legal options in fixed order (SVY-R16)
	DecideMode                            // "Choose N" modes: fallback picks the first listed options
	DecideOptional                        // "you may": fallback declines (no picks)
	DecideOrder                           // order own triggers (FR-TRG-003): Min==Max==len(Options), fallback is collection order
	DecideHand                            // choose cards from your own hand: not an inspection (fast path applies); fallback picks the cards most recently added (SVY-R16)
	DecidePick                            // untargeted pick from the battlefield or a discard pile as a step resolves (SVY-R3, SVY-R14): not a target, so Ward and the targeting checks skip it; same fallback as DecideTarget
	DecideReplacement                     // choose which of several replacements applies to one event (FR-REP-001, SVY-R23), asked of the affected asset's controller: same fallback as DecideTarget, the first source in fixed order (SVY-R16), a source's own replacements in listed (printed) order
)

func (k DecisionKind) String() string {
	return [...]string{"Target", "Mode", "Optional", "Order", "Hand", "Pick", "Replacement"}[k]
}

// Option is one legal answer element, in the order the card or rules list it.
type Option struct {
	Label string // for logs and cardtown play
	Ref   int    // the asset or card instance ID it refers to, or 0
}

// Decision is one question to one player. Options holds only legal options.
type Decision struct {
	Player   PlayerID
	Kind     DecisionKind
	Prompt   string
	Options  []Option
	Min, Max int  // how many picks a valid answer has
	Hidden   bool // the choice inspects hidden cards in a Hand or Deck (no fast path)
	Repeats  bool // the same option may be picked more than once (FR-MOD-001 §5)
}

// Answer is a Chooser's reply: indexes into Decision.Options, or a timeout.
type Answer struct {
	Picks   []int
	Timeout bool
}

// Chooser answers decisions for one seat.
type Chooser func(g *Game, d Decision) Answer

// Valid reports whether picks is a legal answer to d: Min to Max picks, each an index into Options, with no repeats unless Repeats (FR-MOD-001). The kernel applies the same rule to every Chooser answer.
func (d Decision) Valid(picks []int) bool {
	if len(picks) < d.Min || len(picks) > d.Max {
		return false
	}
	seen := make(map[int]bool, len(picks))
	for _, p := range picks {
		if p < 0 || p >= len(d.Options) || (seen[p] && !d.Repeats) {
			return false
		}
		seen[p] = true
	}
	return true
}

// satisfiable reports whether any valid answer to d exists.
func (d Decision) satisfiable() bool {
	if d.Min < 0 || d.Min > d.Max {
		return false
	}
	if d.Repeats {
		return len(d.Options) > 0 || d.Min == 0
	}
	return d.Min <= len(d.Options)
}

// soleAnswer returns the only valid answer to d, if there is exactly one
// (FR-MOD-001 §1). A "you may" decision always has two, so it never has one.
func (d Decision) soleAnswer() ([]int, bool) {
	if d.Kind == DecideOptional || d.Min != d.Max {
		return nil, false
	}
	switch {
	case !d.Repeats && d.Min == len(d.Options):
		picks := make([]int, len(d.Options))
		for i := range picks {
			picks[i] = i
		}
		return picks, true
	case d.Repeats && len(d.Options) == 1:
		return make([]int, d.Min), true // Min copies of option 0
	}
	return nil, false
}

// choose returns the picks for d (indexes into d.Options). It takes the
// single-option fast path itself, otherwise asks the deciding seat's Chooser,
// and applies the FR-MOD-001 §6 fallback on a nil chooser, a timeout or an
// invalid answer. With no valid answer at all it returns no picks and asks no one.
func (g *Game) choose(d Decision) []int {
	if !d.satisfiable() {
		return nil
	}
	// Order matters for DecideOrder, so with 2+ options there is no sole answer.
	if !d.Hidden && !(d.Kind == DecideOrder && len(d.Options) >= 2) {
		if picks, ok := d.soleAnswer(); ok {
			g.emitDecision(d, picks, "auto")
			return picks
		}
	}
	var ch Chooser
	if int(d.Player) >= 0 && int(d.Player) < len(g.choosers) {
		ch = g.choosers[d.Player]
	}
	how := "timeout"
	if ch != nil {
		a := ch(g, d)
		if !a.Timeout {
			if d.Valid(a.Picks) {
				g.emitDecision(d, a.Picks, "answer")
				return a.Picks
			}
			g.logf(EventInvalidAnswer, "player %d gave an invalid answer %v to %q", d.Player, a.Picks, d.Prompt)
			how = "invalid"
		}
	}
	picks := g.fallbackPicks(d)
	g.emitDecision(d, picks, how)
	return picks
}

// fallbackPicks is the FR-MOD-001 §6 answer for a missing, timed-out or invalid reply.
func (g *Game) fallbackPicks(d Decision) []int {
	switch d.Kind {
	case DecideTarget, DecidePick, DecideReplacement:
		return g.fixedOrderPicks(d)
	case DecideOrder:
		picks := make([]int, d.Min)
		for i := range picks {
			picks[i] = i
		}
		return picks
	case DecideHand:
		return g.handPicks(d)
	case DecideMode:
		picks := make([]int, d.Min)
		for i := range picks {
			if d.Repeats {
				picks[i] = i % len(d.Options)
			} else {
				picks[i] = i
			}
		}
		return picks
	}
	return nil
}

// handPicks returns the indexes of the d.Min options whose cards come latest in
// the deciding player's hand, most recent first (FR-MOD-001 §6, SVY-R16). It
// draws nothing from g.rng.
func (g *Game) handPicks(d Decision) []int {
	hand := g.Players[d.Player].Hand
	pos := make(map[int]int, len(hand))
	for i, inst := range hand {
		pos[inst.ID] = i
	}
	idx := make([]int, len(d.Options))
	for i := range idx {
		idx[i] = i
	}
	rank := func(i int) int {
		if r, ok := pos[d.Options[i].Ref]; ok {
			return r
		}
		return -1 // not in hand: after all that are
	}
	slices.SortStableFunc(idx, func(a, b int) int { return rank(b) - rank(a) })
	return idx[:min(d.Min, len(idx))]
}

// fixedOrderPicks returns the indexes of the first d.Min options in the FR-MOD-001
// §6 fixed order (SVY-R16), in that order. It draws nothing from g.rng.
func (g *Game) fixedOrderPicks(d Decision) []int {
	if d.Min == 0 {
		return nil
	}
	pos := g.fixedOrderPositions(d.Player)
	idx := make([]int, len(d.Options))
	for i := range idx {
		idx[i] = i
	}
	rank := func(i int) int {
		if r, ok := pos[d.Options[i].Ref]; ok {
			return r
		}
		return len(pos) // not a battlefield asset: after all that are
	}
	slices.SortStableFunc(idx, func(a, b int) int { return rank(a) - rank(b) })
	picks := make([]int, d.Min)
	for i := range picks {
		if d.Repeats {
			picks[i] = idx[i%len(idx)]
		} else {
			picks[i] = idx[i]
		}
	}
	return picks
}

// randomPicks returns n uniformly random picks for d: distinct unless d.Repeats.
func randomPicks(r *rand.Rand, d Decision, n int) []int {
	if n == 0 {
		return nil
	}
	picks := make([]int, n)
	if d.Repeats {
		for i := range picks {
			picks[i] = r.IntN(len(d.Options))
		}
		return picks
	}
	idx := make([]int, len(d.Options))
	for i := range idx {
		idx[i] = i
	}
	for i := 0; i < n; i++ { // partial Fisher-Yates
		j := i + r.IntN(len(idx)-i)
		idx[i], idx[j] = idx[j], idx[i]
	}
	copy(picks, idx[:n])
	return picks
}

// RandomChooser returns a Chooser that gives a uniformly random valid answer to
// every Decision, deterministic for a seed. Playouts and cardtown sim use it.
func RandomChooser(seed uint64) Chooser {
	r := rand.New(rand.NewPCG(seed, ^seed^0xC401CE))
	return func(_ *Game, d Decision) Answer {
		if !d.satisfiable() {
			return Answer{Timeout: true}
		}
		hi := d.Max
		if !d.Repeats && hi > len(d.Options) {
			hi = len(d.Options)
		}
		if d.Repeats && len(d.Options) == 0 {
			hi = 0
		}
		return Answer{Picks: randomPicks(r, d, d.Min+r.IntN(hi-d.Min+1))}
	}
}
