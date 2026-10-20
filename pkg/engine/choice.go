package engine

import "math/rand/v2"

// DecisionKind says what a Decision asks, and so which FR-MOD-001 §6 fallback applies.
type DecisionKind int

const (
	DecideTarget   DecisionKind = iota // mandatory targets: fallback picks random legal options
	DecideMode                         // "Choose N" modes: fallback picks the first listed options
	DecideOptional                     // "you may": fallback declines (no picks)
	DecideOrder                        // order own triggers (FR-TRG-003): Min==Max==len(Options), fallback is collection order
)

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

// valid reports whether picks is a valid answer to d.
func (d Decision) valid(picks []int) bool {
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
			return picks
		}
	}
	var ch Chooser
	if int(d.Player) >= 0 && int(d.Player) < len(g.choosers) {
		ch = g.choosers[d.Player]
	}
	if ch != nil {
		a := ch(g, d)
		if !a.Timeout {
			if d.valid(a.Picks) {
				return a.Picks
			}
			g.logf("invalid-answer", "player %d gave an invalid answer %v to %q", d.Player, a.Picks, d.Prompt)
		}
	}
	return g.fallbackPicks(d)
}

// fallbackPicks is the FR-MOD-001 §6 answer for a missing, timed-out or invalid reply.
func (g *Game) fallbackPicks(d Decision) []int {
	switch d.Kind {
	case DecideTarget:
		return randomPicks(g.rng, d, d.Min)
	case DecideOrder:
		picks := make([]int, d.Min)
		for i := range picks {
			picks[i] = i
		}
		return picks
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
