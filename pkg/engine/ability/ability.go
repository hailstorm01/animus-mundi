// Package ability is the shape vocabulary for printed card abilities: plain
// data, no game logic and no card names (the fences linter enforces the
// latter). New types here arrive only through a `shape` task.
//
// A new enum here follows one of three rules, chosen by what its zero value
// means:
//   - A set of flags that combine is declared with 1 << iota, one bit per
//     member, so a value may hold several (Kinds).
//   - An enum with no default starts at iota + 1, so an unset field (the zero
//     value) is invalid and never passes for the first member (Hook, Op).
//   - An enum whose zero value is a meaningful default starts at iota, and the
//     first member's comment says "the zero value" (TurnOwner, Keyword, Zone,
//     Allegiance, Subject, ValueSource).
package ability

// Def is the executable definition of one card's printed abilities.
type Def struct {
	// Text lists the exact rules-text lines this definition implements.
	// Each must equal a non-keyword line printed on the card.
	Text []string
	// Triggers lists the printed triggered abilities, in printed order.
	Triggers []Trigger
	// Replacements lists the printed replacement effects, in printed order.
	Replacements []Replacement
}

// Would is the event a replacement replaces.
type Would int

const (
	WouldBeDestroyed Would = iota + 1 // an asset matching the replacement's Filter would be destroyed, by the death sweep (being lethal) or by a kill step (FR-REP-001, FR-LAY-002, SVY-R15)
)

// Replacement is one printed replacement effect. If an asset matching Filter
// would be destroyed (Event), do Steps instead (FR-REP-001). Filter is relative
// to the controller of the replacement's source. Steps run inline, never on the
// stack, with that source as their source: SubjectSelf is the source, and
// SubjectReplaced is the asset whose event is replaced.
type Replacement struct {
	Event  Would
	Filter Filter
	Steps  []Step
}

// Hook is the event a trigger watches.
type Hook int

const (
	HookDeploy     Hook = iota + 1 // this asset was deployed (its deploy resolved onto the battlefield)
	HookKilled                     // this asset was destroyed from the battlefield (FR-ZON-004 leaves-battlefield trigger)
	HookAttacks                    // this asset attacked (combat Step 2, FR-CBT-002)
	HookDamaged                    // this asset took damage: Armor absorbed plus Health committed was more than 0 (FR-REP-004 Stage 5, FR-KWD-002a)
	HookExhausted                  // this asset exhausted: changed from Readied to Exhausted, for any reason (SVY-R17, FR-CAP-001)
	HookCombatKill                 // this asset attacked, and its combat damage left the defender lethal, so the death sweep after the Attack destroys it (FR-CBT-002 Step 3)
)

// TurnOwner gates a trigger on whose turn it is when its event happens.
type TurnOwner int

const (
	TurnAny   TurnOwner = iota // the zero value: no gate
	TurnYours                  // only during the trigger controller's own turn
)

// Keyword names the keyword a trigger comes from, for a keyword that is a
// triggered ability. The zero value is a printed ability.
type Keyword int

const (
	KeywordNone       Keyword = iota // the zero value: a printed ability
	KeywordAggressive                // "Aggressive N" (FR-KWD-002f)
)

// String returns the keyword's printed name, or "" for KeywordNone.
func (k Keyword) String() string {
	if k == KeywordAggressive {
		return "Aggressive"
	}
	return ""
}

// Aggressive is the trigger the keyword "Aggressive N" means: whenever this
// unit attacks, it gets +N power until end of turn (FR-KWD-002f).
func Aggressive(n int) Trigger {
	return Trigger{Hook: HookAttacks, Keyword: KeywordAggressive,
		Effect: Effect{Steps: []Step{{Op: OpModifyStat, Subject: SubjectSelf, Amount: Value{N: n}}}}}
}

// Trigger is one printed triggered ability.
type Trigger struct {
	Hook     Hook
	Turn     TurnOwner
	ToZone   Zone    // HookDeploy only: fires only if the asset was deployed into this zone
	Cornered bool    // "Cornered: ...": fires only if its controller is Cornered when the event happens (FR-KWD-002e, FR-LAY-003)
	Optional bool    // "you may": asked when the trigger is placed, before its targets; declined, it is not placed (SVY-R9)
	Keyword  Keyword // the keyword this trigger comes from; zero for a printed ability
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
	ZoneDiscard // a discard pile, not the battlefield: the entities are cards in it (FR-TGT-001 Zone). Filter.Allegiance names whose pile: allied is the controller's
)

// Effect is what a resolving ability does: its steps, in order. A modal effect
// lists Modes instead: each mode is an Effect of its own, and ChooseN of them
// are chosen when the ability is placed, before any target (FR-MOD-001 §5). An
// effect with Modes has no Targets or Steps of its own.
type Effect struct {
	Targets []Target // "chosen" clauses, chosen when the ability is placed on the stack (FR-TGT-001)
	Steps   []Step
	Modes   []Effect // "choose N": the modes in printed order; an effect with Modes has no Targets or Steps of its own (FR-MOD-001 §5)
	ChooseN int      // with Modes: how many different modes are chosen ("choose one" is 1, SVY-R6)
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
	Scratch    bool       // only Scratch Company units
	Other      bool       // excludes the ability's own source asset ("other")
	Zone       Zone       // ZoneAny (zero) means any battlefield zone; ZoneDiscard means cards in a discard pile; else only assets in that zone
}

// InPile reports whether f chooses cards in a discard pile rather than assets
// on the battlefield (Zone is ZoneDiscard). Every target, pick and filter check
// that tells the two apart asks this, not the zone.
func (f Filter) InPile() bool { return f.Zone == ZoneDiscard }

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
	OpKill                       // destroy the Subject at once, into its owner's discard pile (FR-LAY-002, FR-ZON-002); a Headquarters is skipped (no M2 card kills one)
)

// Subject is what a step acts on.
type Subject int

const (
	SubjectController Subject = iota // the ability's controller (the zero value; Scratch and Draw)
	SubjectSelf                      // the ability's own source card, tracked by card instance (FR-ZON-004)
	SubjectOpponents                 // each opponent of the ability's controller (OpScratch, OpDraw)
	SubjectTarget                    // each entity chosen for Effect.Targets[Step.Target], in pick order
	SubjectAll                       // every battlefield asset matching Step.Filter when the step resolves; not a target (FR-TGT-002)
	SubjectPick                      // the entities matching Step.Filter, picked by the ability's controller as the step resolves: battlefield assets, or cards in a discard pile if Filter.InPile. One, or with Step.UpTo up to N, possibly none. Not a target (SVY-R3, SVY-R14)
	SubjectReplaced                  // the asset whose event a replacement replaces (Replacement.Steps only)
)

// Step applies one primitive.
type Step struct {
	Op      Op
	Amount  Value
	Subject Subject
	Bottom  bool   // OpPutOnDeck: the bottom of the deck instead of the top
	Target  int    // SubjectTarget: index into Effect.Targets
	Filter  Filter // SubjectAll: which assets. SubjectPick: which assets, or which cards in a discard pile (Filter.InPile)
	UpTo    int    // SubjectPick: 0 picks exactly one; N > 0 picks up to N distinct entities (assets or cards), possibly none
	IfDone  bool   // runs only if the step just before it did at least one thing ("if you do", SVY-R14, SVY-R15)
}

// ValueSource is where a Value's number comes from.
type ValueSource int

const (
	ValueSourceFixed ValueSource = iota // the zero value: N
	ValueSourcePower                    // the source's effective power when the step resolves, or its last-known effective power (SVY-R10); N is ignored
)

// Value is a number a step uses, read when the step resolves; From says where
// it comes from. It is a struct, not a bare int, so that each source can carry
// what it needs: N is read only for ValueSourceFixed (the zero source), and a
// later source, such as "the number of X", adds its own field here without
// changing Step.
type Value struct {
	N    int
	From ValueSource
}
