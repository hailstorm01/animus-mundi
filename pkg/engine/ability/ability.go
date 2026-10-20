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
	HookDeploy  Hook = iota + 1 // this asset was deployed (its deploy resolved onto the battlefield)
	HookKilled                  // this asset was destroyed from the battlefield (FR-ZON-004 leaves-battlefield trigger)
	HookAttacks                 // this asset attacked (combat Step 2, FR-CBT-002)
)

// TurnOwner gates a trigger on whose turn it is when its event happens.
type TurnOwner int

const (
	TurnAny   TurnOwner = iota // the zero value: no gate
	TurnYours                  // only during the trigger controller's own turn
)

// Trigger is one printed triggered ability.
type Trigger struct {
	Hook   Hook
	Turn   TurnOwner
	Effect Effect
}

// Effect is what a resolving ability does: its steps, in order.
type Effect struct {
	Targets []Target // "chosen" clauses, chosen when the ability is placed on the stack (FR-TGT-001)
	Steps   []Step
}

// Target is one "chosen" entity: exactly one, for now ("up to N" comes later).
type Target struct {
	Filter Filter
}

// Filter is a conjunction: an entity must meet every set criterion (FR-TGT-001).
type Filter struct {
	Allegiance Allegiance // relative to the ability's controller
	Kinds      Kinds      // zero means any asset
	NonScratch bool       // excludes Scratch Company units
}

// Allegiance is whose side an entity is on, relative to the ability's controller.
type Allegiance int

const (
	AllegianceAny      Allegiance = iota // the zero value
	AllegianceAllied                     // controlled by the ability's controller
	AllegianceOpposing                   // controlled by an opponent
)

// Kinds is a set of asset types; "structure or Headquarters" sets two bits.
type Kinds int

const (
	KindUnit Kinds = 1 << iota
	KindStructure
	KindHeadquarters
)

// Op is one primitive step.
type Op int

const (
	OpScratch      Op = iota + 1 // its controller performs Deploy Scratch Company Amount times (FR-ZON-005)
	OpDraw                       // its controller draws Amount cards (FR-WIN-001 applies to an empty deck)
	OpReturnToHand               // put the Subject card into its owner's hand
	OpPutOnDeck                  // put the Subject card into its owner's deck: on top, or on the bottom if Bottom
	OpDealDamage                 // Amount non-combat damage to the Subject (FR-REP-004)
)

// Subject is what a step acts on.
type Subject int

const (
	SubjectController Subject = iota // the ability's controller (the zero value; Scratch and Draw)
	SubjectSelf                      // the ability's own source card, tracked by card instance (FR-ZON-004)
	SubjectOpponents                 // each opponent of the ability's controller (OpScratch, OpDraw)
	SubjectTarget                    // the entity chosen for Effect.Targets[Step.Target]
)

// Step applies one primitive.
type Step struct {
	Op      Op
	Amount  Value
	Subject Subject
	Bottom  bool // OpPutOnDeck: the bottom of the deck instead of the top
	Target  int  // SubjectTarget: index into Effect.Targets
}

// Value is a number a step uses. Later shape tasks add other sources; the
// zero source always means the fixed N.
type Value struct {
	N int
}
