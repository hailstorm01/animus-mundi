// Package ability is the shape vocabulary for printed card abilities: plain
// data, no game logic and no card names (the fences linter enforces the
// latter). New types here arrive only through a `shape` task.
package ability

// Def is the executable definition of one card's printed abilities.
type Def struct {
	// Text lists the exact rules-text lines this definition implements.
	// Each must equal a non-keyword line printed on the card.
	Text []string
	// Triggers lists the printed triggered abilities, in printed order.
	Triggers []Trigger
}

// Hook is the event a trigger watches.
type Hook int

const (
	HookDeploy Hook = iota + 1 // this asset was deployed (its deploy resolved onto the battlefield)
	HookKilled                 // this asset was destroyed from the battlefield (FR-ZON-004 leaves-battlefield trigger)
)

// Trigger is one printed triggered ability.
type Trigger struct {
	Hook   Hook
	Effect Effect
}

// Effect is what a resolving ability does: its steps, in order.
type Effect struct {
	Steps []Step
}

// Op is one primitive step.
type Op int

const (
	OpScratch      Op = iota + 1 // its controller performs Deploy Scratch Company Amount times (FR-ZON-005)
	OpDraw                       // its controller draws Amount cards (FR-WIN-001 applies to an empty deck)
	OpReturnToHand               // put the Subject card into its owner's hand
	OpPutOnDeck                  // put the Subject card into its owner's deck: on top, or on the bottom if Bottom
)

// Subject is what a step acts on.
type Subject int

const (
	SubjectController Subject = iota // the ability's controller (the zero value; Scratch and Draw)
	SubjectSelf                      // the ability's own source card, tracked by card instance (FR-ZON-004)
)

// Step applies one primitive.
type Step struct {
	Op      Op
	Amount  Value
	Subject Subject
	Bottom  bool // OpPutOnDeck: the bottom of the deck instead of the top
}

// Value is a number a step uses. Later shape tasks add other sources; the
// zero source always means the fixed N.
type Value struct {
	N int
}
