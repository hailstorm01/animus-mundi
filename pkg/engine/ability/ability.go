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
//     Allegiance, ScratchCompany, Subject, ValueSource).
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
	// Effect is a Maneuver's effect, resolved when the Maneuver resolves (FR-MNV-001).
	// It is zero for every other card type.
	Effect Effect
	// ExtraCosts lists the additional costs paid when the card is declared, in
	// printed order (FR-TRN-003a). It is zero for a card without one.
	ExtraCosts []Cost
	// Activated lists the printed activated abilities, in printed order
	// (FR-ACT-001).
	Activated []Activated
	// Cornered is what the card has while its controller is Cornered: keywords, as
	// in "Cornered: Aggressive 3.", and power, as in "Cornered: This unit has +2
	// power." (SVY-R18, FR-KWD-002e). It is zero for a card without one.
	Cornered WhileCornered
	// CostReduction is the card's own hand-active cost reduction (FR-CST-001),
	// read by Game.EffectiveCost. It is zero for a card without one.
	CostReduction CostReduction
	// Statics lists the printed static abilities, in printed order: a
	// discard-active deploy permission and a battlefield aura are the kinds
	// so far.
	Statics []Static
}

// StaticKind enumerates the kinds of printed static ability. It has no default:
// the zero value is invalid.
type StaticKind int

const (
	// StaticDeployPermission is "You may deploy this card from your discard pile
	// by ...": its owner may deploy it from ActiveZone as a manual Deploy, paying
	// Costs in addition to its deployment costs (FR-TRN-003, FR-CST-001).
	StaticDeployPermission StaticKind = iota + 1
	// StaticAura is a continuous aura of a face-up battlefield asset
	// (FR-LAY-001): every battlefield asset matching Filter, relative to the
	// source's controller, has +Power Combat Power (Layer 4) and +Resistant
	// Resistant (a Layer 3 keyword grant). A ZoneRefSame criterion reads
	// the source's current zone (SVY-R44). ActiveZone is ZoneAny, the battlefield.
	StaticAura
)

// Static is one printed static ability.
type Static struct {
	Kind StaticKind
	// ActiveZone is the zone the card must be in for the ability to apply
	// (ZoneDiscard for a deploy permission).
	ActiveZone Zone
	// Costs are a deploy permission's additional costs, in printed order, paid on
	// declaration (FR-TRN-003a). Their filters stay battlefield filters, relative
	// to the payer.
	Costs []Cost
	// Filter is an aura's: which assets it applies to, relative to the source's
	// controller; Other excludes the source. Its Stat stays StatNone, and it
	// stays free of keyword criteria, since a Layer 3 or 4 value read by its own
	// filter has no defined order.
	Filter Filter
	// Power is an aura's Layer 4 power.
	Power int
	// Resistant is an aura's keyword grant: +Resistant Resistant to every asset
	// Filter matches, a Layer 3 grant (FR-LAY-001, survey G2). It may be
	// negative, like printed Resistant; 0 grants nothing.
	Resistant int
}

// CostReduction lowers the card's own Requisition cost while it is in its owner's
// hand: Req off for each card in a discard pile that matches Per (Per.InPile; its
// Allegiance is relative to the card's owner). Requisition clamps at 0.
type CostReduction struct {
	Req int
	Per Filter
}

// WhileCornered is what a card has only while its controller is Cornered, 0 for
// absent. g.Keywords adds the keywords as a Layer 3 grant and g.Power the Power
// in Layer 4, on a face-up asset (FR-LAY-001); the gate is locked for a
// simultaneous step (FR-LAY-003). Power may be negative, like any static power;
// the clamp is Layer 6's.
type WhileCornered struct {
	Aggressive int
	Resistant  int
	Power      int
}

// Activated is one printed "[Cost]: [Effect]." ability of a battlefield asset.
// Its Costs are paid, in printed order, and its targets chosen when it is
// activated (FR-ACT-001, FR-ACT-002); its Effect resolves from the stack.
type Activated struct {
	Costs  []Cost
	Effect Effect
}

// CostKind is what a cost asks the payer to do. It has no default.
type CostKind int

const (
	CostKill        CostKind = iota + 1 // kill one allied asset matching Cost.Filter, chosen by the payer (FR-TRN-003a)
	CostExhaust                         // exhaust one Readied allied asset matching Cost.Filter, chosen by the payer (FR-TRN-003a, SVY-R17)
	CostDiscard                         // discard up to Cost.UpTo cards of the payer's choice from their own hand, possibly none (FR-TRN-003a)
	CostResources                       // pay Cost.Req Requisition and Cost.Tac Tactics (FR-ACT-001)
	CostExhaustSelf                     // "exhaust": exhaust the ability's source, which must be Readied (FR-ACT-002, SVY-R17)
	CostKillSelf                        // "kill this unit": kill the ability's source through the kill path (FR-ACT-001, SVY-R15)
)

// Cost is one cost, paid when its card or ability is declared (FR-TRN-003a,
// FR-ACT-001). Filter is relative to the payer and is a battlefield filter
// (never InPile); the kill and exhaust costs read it. UpTo is the most cards
// the discard cost discards (UpTo > 0); no other cost reads it. Req and Tac are
// the Requisition and Tactics the resources cost pays; no other cost reads them.
// Choosing what pays a cost is not targeting, so Ward doesn't restrict it.
type Cost struct {
	Kind   CostKind
	Filter Filter
	UpTo   int
	Req    int
	Tac    int
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
	HookDeploy            Hook = iota + 1 // this asset was deployed (its deploy resolved onto the battlefield)
	HookKilled                            // this asset was destroyed from the battlefield (FR-ZON-004 leaves-battlefield trigger)
	HookAttacks                           // this asset attacked (combat Step 2, FR-CBT-002)
	HookDamaged                           // this asset took damage: Armor absorbed plus Health committed was more than 0 (FR-REP-004 Stage 5, FR-KWD-002a)
	HookExhausted                         // this asset exhausted: changed from Readied to Exhausted, for any reason (SVY-R17, FR-CAP-001)
	HookCombatKill                        // this asset attacked, its combat damage left the defender lethal, and the death sweep after the Attack destroyed it; a replaced death is not a kill (FR-CBT-002 Step 3, SVY-R19)
	HookMoves                             // this unit moved to a new battlefield zone, by March or by a step that moves it (OpMoveZone); entering the battlefield, by any deploy, is not moving (SVY-R1, FR-ZON-002)
	HookDiscardedFromHand                 // this card was put into its owner's discard pile from their hand, by a discard step or a discard cost (FR-ZON-004); fires only triggers with ActiveZone ZoneDiscard
	HookTurnStart                         // a turn began: fired in the Ready Phase after the active player's assets ready, for every battlefield asset; "at the start of your turn" is this hook with Turn: TurnYours (FR-TRN-001)
)

// TurnOwner gates a trigger on whose turn it is when its event happens.
type TurnOwner int

const (
	TurnAny   TurnOwner = iota // the zero value: no gate
	TurnYours                  // only during the trigger controller's own turn
)

// Keyword names a keyword: the one a trigger comes from (Aggressive, a keyword
// that is a triggered ability) or one a step grants (Stalwart, OpGrantKeyword).
// The zero value is a printed ability.
type Keyword int

const (
	KeywordNone       Keyword = iota // the zero value: a printed ability
	KeywordAggressive                // "Aggressive N" (FR-KWD-002f)
	KeywordStalwart                  // "Stalwart" (FR-KWD-002c)
)

// String returns the keyword's printed name, or "" for KeywordNone.
func (k Keyword) String() string {
	switch k {
	case KeywordAggressive:
		return "Aggressive"
	case KeywordStalwart:
		return "Stalwart"
	}
	return ""
}

// Duration is how long a temporary effect a step makes lasts (FR-LAY-003).
type Duration int

const (
	UntilEndOfTurn         Duration = iota // the zero value: expires in Sub-Phase 4c of the current turn (FR-TRN-001)
	UntilEndOfOpponentTurn                 // "until the end of your opponent's turn": expires in Sub-Phase 4c of a turn whose active player is an opponent of the ability's controller (FR-LAY-003, FR-TRN-001)
)

// String returns the duration as log text.
func (d Duration) String() string {
	if d == UntilEndOfOpponentTurn {
		return "until the end of its controller's opponent's turn"
	}
	return "until end of turn"
}

// Aggressive is the trigger the keyword "Aggressive N" means: whenever this
// unit attacks, it gets +N power until end of turn (FR-KWD-002f).
func Aggressive(n int) Trigger {
	return Trigger{Hook: HookAttacks, Keyword: KeywordAggressive,
		Effect: Effect{Steps: []Step{{Op: OpModifyStat, Subject: SubjectSelf, Amount: Value{N: n}}}}}
}

// Trigger is one printed triggered ability.
type Trigger struct {
	Hook   Hook
	Turn   TurnOwner
	ToZone Zone // HookDeploy and HookMoves only: fires only if the asset was deployed or moved into this zone
	// ActiveZone is where the trigger works from (survey G1): ZoneAny, the zero value, is the
	// battlefield; ZoneDiscard is its card's owner's discard pile. No other value is used.
	ActiveZone Zone
	Cornered   bool    // "Cornered: ...": fires only if its controller is Cornered when the event happens (FR-KWD-002e, FR-LAY-003)
	Optional   bool    // "you may": asked when the trigger is placed, before its targets; declined, it is not placed (SVY-R9)
	Keyword    Keyword // the keyword this trigger comes from; zero for a printed ability
	// Watches makes the trigger fire on the event of each battlefield asset matching Watch
	// (judged against the trigger's controller and source when the event happens), never
	// on its own source's event unless the source matches Watch. False, the zero value, is
	// a self hook. Read only for HookAttacks so far.
	Watches bool
	Watch   Filter
	Effect  Effect

	// MaxPerTurn limits how often this trigger fires each turn, per asset, so only for a battlefield trigger; no pile trigger prints a limit (FR-TRN-001);
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
	// PerDiscard chooses the slot once for each card the placement's discard
	// cost discarded: each an exactly-one pick, independent of the others, so
	// the same entity may be chosen more than once (SVY-R4, FR-TGT-001). UpTo
	// stays 0 with it. With no cards discarded, or no legal entity, the slot
	// stays empty.
	PerDiscard bool
	// Another 0 means none. k > 0 makes this slot pick an entity different from
	// the one chosen for Effect.Targets[k-1], the earlier slot of the same
	// clause ("another chosen X", SVY-R34); k-1 is less than this slot's own
	// index. "Another" with no earlier slot in its clause is Filter.Other, not
	// this field.
	Another int
}

// AnotherSlot returns the Effect.Targets index of the slot this one must
// differ from, and whether it has one.
func (t Target) AnotherSlot() (slot int, ok bool) {
	return t.Another - 1, t.Another > 0
}

// Filter is a conjunction: an entity must meet every set criterion (FR-TGT-001).
type Filter struct {
	Allegiance Allegiance     // relative to the ability's controller
	Kinds      Kinds          // zero means any asset, or any card in a pile
	Subtypes   []string       // the entity carries at least one of these printed subtypes ("Fungus or Plant card"); empty means no subtype criterion
	Scratch    ScratchCompany // ScratchAny (zero) means any unit; else only Scratch Company units, or none
	Other      bool           // excludes the ability's own source asset ("other")
	Zone       Zone           // ZoneAny (zero) means any battlefield zone; ZoneDiscard means cards in a discard pile; else only assets in that zone
	ZoneRef    ZoneRef        // ZoneRefNone (zero) means no source-relative zone criterion; else only assets in the zone the stack item locked (an aura: its source's current zone)
	Stat       Stat           // StatNone (zero) means no stat criterion; else the entity's value of that stat must be AtMost or less
	AtMost     int            // the bound of Stat ("power 3 or less" is StatPower, 3); means nothing while Stat is StatNone
}

// Stat is a stat of the one vocabulary that a filter bound (Filter.Stat, FR-TGT-001
// "Dynamic Stats") and a modify step (Step.Stat) both use. A filter reads a card in a
// pile by its printed value and an asset on the battlefield by its effective one.
type Stat int

const (
	StatNone        Stat = iota // the zero value: no stat criterion
	StatPower                   // Combat Power
	StatRequisition             // Requisition cost (FR-CRD-101)
	StatMaxHealth               // Maximum Health (FR-LAY-002), not Current Health
)

// ZoneRef is a zone criterion read against the zone a stack item locked, not
// the source's current zone (FR-TGT-001, FR-TGT-003). A stack item locks one
// zone, set by the ZoneRef its effect reads. A continuous aura is not a stack
// item: it reads its source's current zone (SVY-R44).
type ZoneRef int

const (
	ZoneRefNone ZoneRef = iota // the zero value: no source-relative zone criterion
	ZoneRefThat                // "in that zone": the zone locked as a move's destination (SVY-R45)
	ZoneRefSame                // "in the same zone": the zone the item's source was in when the item was placed (SVY-R44)
)

// ReadsZoneRef reports whether any filter of e reads r: its Targets' filters,
// its Steps' filters, and each of its Modes, recursively.
func (e Effect) ReadsZoneRef(r ZoneRef) bool {
	for _, t := range e.Targets {
		if t.Filter.ZoneRef == r {
			return true
		}
	}
	for _, s := range e.Steps {
		if s.Filter.ZoneRef == r {
			return true
		}
	}
	for _, m := range e.Modes {
		if m.ReadsZoneRef(r) {
			return true
		}
	}
	return false
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

// Kinds is a set of card types; "structure or Headquarters" sets two bits.
type Kinds int

const (
	KindUnit Kinds = 1 << iota
	KindStructure
	KindHeadquarters
	KindManeuver // only a pile filter can meet it: no asset is a Maneuver
)

// ScratchCompany is a filter's Scratch Company criterion: whether a unit with
// the Scratch Company subtype matches (FR-TGT-001).
type ScratchCompany int

const (
	ScratchAny      ScratchCompany = iota // the zero value: any unit, Scratch Company or not
	ScratchOnly                           // only Scratch Company units ("allied Scratch Company unit")
	ScratchExcluded                       // no Scratch Company units ("non-Scratch unit")
)

// Op is one primitive step.
type Op int

const (
	OpScratch      Op = iota + 1 // its controller performs Deploy Scratch Company Amount times (FR-ZON-005)
	OpDraw                       // its controller draws Amount cards (FR-WIN-001 applies to an empty deck)
	OpReturnToHand               // put the Subject card or asset into its owner's hand
	OpPutOnDeck                  // put the Subject card or asset into its owner's deck: on top, or on the bottom if Bottom
	OpDealDamage                 // Amount non-combat damage to the Subject (FR-REP-004)
	OpHeal                       // remove up to Amount accumulated damage from the Subject (FR-KWD-002d)
	OpAddArmor                   // add Amount to the Subject's Armor pool (FR-LAY-002)
	OpDiscard                    // its controller discards Amount cards of their choice from their own hand (FR-MOD-001 §1, §6; SVY-R16)
	OpModifyStat                 // the Subject gets +Amount to Step.Stat (power, or Maximum Health) for Step.Until (Layer 5, FR-LAY-001, FR-LAY-002, FR-LAY-003)
	OpKill                       // destroy the Subject at once, into its owner's discard pile (FR-LAY-002, FR-ZON-002); a Headquarters is skipped (no M2 card kills one)
	OpMoveZone                   // move each Subject unit to Step.To, in its controller's battlefield, keeping its state (FR-ZON-002, FR-ZON-004). A unit already there, a Structure and a Headquarters don't move
	OpSearchDeck                 // its controller chooses a card from their deck, the rest is shuffled and the card goes to their hand (FR-DEK-001, SVY-R37, SVY-R38). An empty deck is still shuffled
	OpDeployFree                 // put each Subject card, still in its discard pile, onto the battlefield for free under the ability's controller, with no stack item and no costs (FR-TRN-003a, FR-CST-001): into Step.To if set, else its printed zone, chosen by the controller when it prints both Reserves and the Main Theater (SVY-R43)
	OpMorphScratch               // turn each Subject unit face down in place: it becomes the 1/1/1/1 Infantry Scratch Company whose only ability is Deploys Exhausted, and keeps its zone, controller, readiness, damage, Armor and Layer 5 modifiers (FR-ENT-103). A unit already face down, a Structure and a Headquarters are skipped
	OpGrantKeyword               // the Subject gains Step.Keyword for Step.Until (Layer 3, FR-LAY-001, FR-LAY-003)
	OpLookTop                    // its controller looks at the top Amount cards of their deck and, in one hidden prompt, chooses which are deployed for free (up to UpTo, or exactly one when UpTo is 0 and one is eligible), each card matching Filter's card-level criteria, the zone of each, and the order the rest go to the bottom (FR-MOD-001 §4, SVY-R41-SVY-R43)
	OpGainControl                // its controller gains control of each Subject unit in the Main Theater that an opponent controls; it keeps its zone, readiness, owner and state (FR-ZON-002, FR-ZON-003). A unit in Reserves, a Structure, a Headquarters and a unit the controller already controls are skipped
)

// Subject is what a step acts on.
type Subject int

const (
	SubjectController Subject = iota // the ability's controller (the zero value; OpScratch, OpDraw, OpDiscard, OpSearchDeck)
	SubjectSelf                      // the ability's own source card, tracked by card instance (FR-ZON-004)
	SubjectOpponents                 // each opponent of the ability's controller (OpScratch, OpDraw, OpDiscard, OpSearchDeck)
	SubjectTarget                    // each entity chosen for Effect.Targets[Step.Target], in pick order; for an asset step, a pile slot's card that an earlier step of this resolution deployed is the asset it became
	SubjectAll                       // every battlefield asset matching Step.Filter when the step resolves; not a target (FR-TGT-002)
	SubjectPick                      // the entities matching Step.Filter, picked by the ability's controller as the step resolves: battlefield assets, or cards in a discard pile if Filter.InPile. One, or with Step.UpTo up to N, possibly none. Not a target (SVY-R3, SVY-R14). With Step.Split, Amount is a pool split among the matches (SVY-R51)
	SubjectReplaced                  // the asset whose event a replacement replaces (Replacement.Steps only)
	SubjectWatched                   // the asset whose event fired a watching trigger (Trigger.Watches), if that same asset is still on the battlefield; not a target
)

// Step applies one primitive.
type Step struct {
	Op                Op
	Amount            Value
	Subject           Subject
	Bottom            bool     // OpPutOnDeck: the bottom of the deck instead of the top
	Target            int      // SubjectTarget: index into Effect.Targets
	Filter            Filter   // SubjectAll: which assets. SubjectPick: which assets, or which cards in a discard pile (Filter.InPile). OpLookTop: the card-level criteria a deployed card must meet
	UpTo              int      // SubjectPick, OpLookTop: 0 picks exactly one; N > 0 picks up to N distinct entities (assets or cards), possibly none
	Dealer            int      // OpDealDamage only. 0: the ability's source deals the damage. k > 0: the asset chosen for the exactly-one slot Effect.Targets[k-1] deals it, and ValueSourcePower reads that asset's power (SVY-R27)
	To                Zone     // OpMoveZone: ZoneReserves or ZoneMainTheater, the unit's controller's. OpDeployFree: ZoneReserves or ZoneMainTheater overrides the card's printed Deployment Zones (FR-TRN-003a 2). OpLookTop: the same override
	Keyword           Keyword  // OpGrantKeyword only: the keyword granted
	IfDone            bool     // runs only if the step just before it did at least one thing ("if you do", SVY-R14, SVY-R15)
	ControllerChooses bool     // OpDiscard only: the ability's controller, not each step player, chooses the cards from each step player's hand; with SubjectOpponents the opponent's hand, revealed to the controller for the choice, so always asked even with one card (FR-MOD-001 §3, SVY-R50). False: each step player chooses from their own hand
	Split             bool     // SubjectPick over assets with OpAddArmor only: Amount is a pool the ability's controller splits among the assets matching Filter as the step resolves, not a target (SVY-R51): each picked asset gets a positive whole share and the shares add up to Amount; one match takes it all unasked, none does nothing, and a timeout puts it all on the first in fixed order (SVY-R52). UpTo must be 0
	Stat              Stat     // OpModifyStat only: the stat the step changes. StatNone (zero) and StatPower are Combat Power; StatMaxHealth is Maximum Health; no other stat is valid
	Until             Duration // OpModifyStat, OpGrantKeyword: how long the modifier lasts; the zero value is until end of turn
	IfHandEmpty       bool     // runs only if the ability's controller has no cards in hand when the step resolves ("if you have no cards in hand")
}

// DealerSlot returns the Effect.Targets index of the step's chosen dealer, and
// whether it has one.
func (s Step) DealerSlot() (slot int, ok bool) {
	return s.Dealer - 1, s.Dealer > 0
}

// ValueSource is where a Value's number comes from.
type ValueSource int

const (
	ValueSourceFixed ValueSource = iota // the zero value: N
	ValueSourcePower                    // the effective power of the step's dealer when the step resolves: the source, read with its last-known effective power once it has left (SVY-R10), or a chosen dealer (Step.Dealer), read only while it is on the battlefield (SVY-R27); N is ignored
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
