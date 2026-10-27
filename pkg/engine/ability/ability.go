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
	HookDeploy    Hook = iota + 1 // this asset was deployed (its deploy resolved onto the battlefield)
	HookKilled                    // this asset was destroyed from the battlefield (FR-ZON-004 leaves-battlefield trigger)
	HookAttacks                   // this asset attacked (combat Step 2, FR-CBT-002)
	HookDamaged                   // this asset took damage: Armor absorbed plus Health committed was more than 0 (FR-REP-004 Stage 5, FR-KWD-002a)
	HookExhausted                 // this asset exhausted: changed from Readied to Exhausted, for any reason (SVY-R17, FR-CAP-001)
)

// TurnOwner gates a trigger on whose turn it is when its event happens.
type TurnOwner int

const (
	TurnAny   TurnOwner = iota // the zero value: no gate
	TurnYours                  // only during the trigger controller's own turn
)

// Trigger is one printed triggered ability.
type Trigger struct {
	Hook     Hook
	Turn     TurnOwner
	ToZone   Zone // HookDeploy only: fires only if the asset was deployed into this zone
	Cornered bool // "Cornered: ...": fires only if its controller is Cornered when the event happens (FR-KWD-002e, FR-LAY-003)
	Optional bool // "you may": asked when the trigger is placed, before its targets; declined, it is not placed (SVY-R9)
	Effect   Effect

	// MaxPerTurn limits how often this trigger fires each turn, per asset (FR-TRN-001);
	// 0 means no limit. A firing whose clauses later fizzle still counts.
	MaxPerTurn int
}

// Zone names a battlefield zone a shape refers to.
type Zone int

const (
	ZoneAny Zone = iota // the zero value: no zone condition
	ZoneReserves
	ZoneMainTheater
)

// Effect is what a resolving ability does: its steps, in order.
type Effect struct {
	Targets []Target // "chosen" clauses, chosen when the ability is placed on the stack (FR-TGT-001)
	Steps   []Step
}

// Target is one "chosen" clause: exactly one entity, or with UpTo "up to N".
type Target struct {
	Filter Filter
	UpTo   int // 0: exactly one chosen entity. N > 0: up to N distinct entities, possibly none (SVY-R11)
}

// Filter is a conjunction: an entity must meet every set criterion (FR-TGT-001).
type Filter struct {
	Allegiance Allegiance // relative to the ability's controller
	Kinds      Kinds      // zero means any asset
	NonScratch bool       // excludes Scratch Company units
	Other      bool       // excludes the ability's own source asset ("other")
	Zone       Zone       // ZoneAny (zero) means any battlefield zone; else only assets in that zone
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
	OpHeal                       // remove up to Amount accumulated damage from the Subject (FR-KWD-002d)
	OpAddArmor                   // add Amount to the Subject's Armor pool (FR-LAY-002)
	OpDiscard                    // its controller discards Amount cards of their choice from their own hand (FR-MOD-001 §1, §6; SVY-R16)
	OpModifyStat                 // the Subject gets +Amount power until end of turn (Layer 5, FR-LAY-001; expires in Sub-Phase 4c, FR-LAY-003). Other stats and durations come later
)

// Subject is what a step acts on.
type Subject int

const (
	SubjectController Subject = iota // the ability's controller (the zero value; Scratch and Draw)
	SubjectSelf                      // the ability's own source card, tracked by card instance (FR-ZON-004)
	SubjectOpponents                 // each opponent of the ability's controller (OpScratch, OpDraw)
	SubjectTarget                    // each entity chosen for Effect.Targets[Step.Target], in pick order
	SubjectAll                       // every battlefield asset matching Step.Filter when the step resolves; not a target (FR-TGT-002)
)

// Step applies one primitive.
type Step struct {
	Op      Op
	Amount  Value
	Subject Subject
	Bottom  bool   // OpPutOnDeck: the bottom of the deck instead of the top
	Target  int    // SubjectTarget: index into Effect.Targets
	Filter  Filter // SubjectAll: which assets
}

// ValueSource is where a Value's number comes from.
type ValueSource int

const (
	ValueFixed       ValueSource = iota // the zero value: N
	ValueSourcePower                    // the source's effective power when the step resolves, or its last-known effective power (SVY-R10); N is ignored
)

// Value is a number a step uses, read when the step resolves. The zero source
// always means the fixed N.
type Value struct {
	N    int
	From ValueSource
}
