// Package engine is the executable form of the Game Rules FRD
// (docs/frd/01_Game_Rules_FRD.md). It is the rules oracle: when the FRD and
// this package disagree, one of them has a bug, and a scenario test named
// after the FR ID (TestFR_CBT_002_...) settles which.
//
// What runs now:
//   - Match setup, first-player choice, partial mulligan, turn phases.
//   - Dual-resource economy: Generate Resources, 2R -> 1T conversion.
//   - Deploying Unit and Structure cards, March, Attack, Deploy Scratch Company.
//   - The LIFO Resolution Stack with clean-stack manual actions.
//   - 3-step combat, symmetric retaliation, 5-stage damage mitigation,
//     state-based death sweeps, HQ defeat, empty-deck loss and Match Draws.
//   - Standalone keywords: Resistant N, Stalwart, Aggressive N, Ethereal,
//     Ward, Versatile, Guardian, Armor N, Deploys Exhausted.
//   - Ability shapes as plain data in pkg/engine/ability, and card definitions
//     keyed by card name in pkg/engine/cards. Engine code never names a card
//     (cardtown lint fences).
//   - Triggered abilities on self hooks (HookDeploy, HookKilled, HookAttacks,
//     HookDamaged, HookExhausted, HookCombatKill, HookMoves) run through a pending-trigger buffer, in APNAP order with cascades
//     (FR-TRG-002 to FR-TRG-004). A trigger may be optional, and its effects
//     may choose targets ("up to N", "other", a zone) that are verified
//     when it is placed and again when it resolves (FR-TGT-001, FR-TGT-003). A
//     trigger may be limited to N firings each turn per asset (Trigger.MaxPerTurn,
//     FR-TRN-001), and Unbowed (FR-KWD-002a) fires on damage taken. A trigger
//     may be modal ("choose N" modes, chosen on placement before its targets,
//     FR-MOD-001 §5).
//   - Per-seat Chooser decisions with timeout fallbacks (FR-MOD-001), including
//     choosing from your own hand (DecideHand, SVY-R16) and from your own deck (DecideSearch, SVY-R37, SVY-R38) and the zone of a free deploy (DecideZone, SVY-R43).
//   - Steps OpScratch, OpDraw, OpDiscard, OpReturnToHand, OpPutOnDeck,
//     OpDealDamage, OpHeal, OpAddArmor, OpModifyStat, OpKill (inline destruction), OpMoveZone, OpSearchDeck, OpDeployFree (a free deploy from a discard pile, FR-TRN-003a), OpMorphScratch (a face-down transformation, FR-ENT-103) and OpGrantKeyword (a keyword granted until end of turn, FR-LAY-001), with "all" steps, source-power values, the
//     deploy-zone gate, the TurnYours gate, the Cornered gate, SubjectSelf (tracked in discard,
//     FR-ZON-004) and SubjectOpponents. A later step on a pile slot acts on the asset an earlier
//     OpDeployFree of the same resolution made of its card. SubjectPick is an untargeted pick made as the
//     step resolves (SVY-R3), asked as a DecidePick and not a target, and a step with IfDone
//     runs only if the step before it did something ("if you do", SVY-R14), and a step with
//     IfHandEmpty runs only if its controller has no cards in hand when it resolves. A damage
//     step with Step.Dealer has an asset chosen for one of its slots deal the damage (SVY-R27).
//   - Target slots that choose once for each card the placement's discard cost
//     discarded (Target.PerDiscard, SVY-R4), or that must pick an entity different
//     from an earlier slot's (Target.Another, SVY-R34).
//   - Maneuvers (ActEnact, FR-MNV-001): modes and targets are chosen on placement,
//     as a trigger's are, the effect resolves as a StackManeuver item, and the card
//     then goes to its owner's discard pile. A Maneuver may print additional costs
//     to kill, exhaust or discard (Def.ExtraCosts), paid on declaration (FR-TRN-003a).
//   - Activated abilities (ActActivate, FR-ACT-001, FR-ACT-002): a printed
//     "[Cost]: [Effect]." ability of a battlefield asset (Def.Activated), activated
//     on an empty stack. On declaration its targets are chosen, it goes on the stack
//     as a StackAbility item and its costs are paid, including CostResources,
//     CostExhaustSelf and CostKillSelf (declare, SVY-R2). It is offered only if placeable holds: its
//     costs are payable and each required target or mode has a legal choice (SVY-R36).
//   - A card's own Cornered keywords ("Cornered: Aggressive 3.", Def.Cornered):
//     g.Keywords adds them as a Layer 3 grant on a face-up asset while its controller
//     is Cornered (FR-LAY-001, FR-KWD-002e, SVY-R18).
//   - A card's own hand-active cost reduction ("costs R1 less to deploy for each
//     unit card in your discard pile", Def.CostReduction): EffectiveCost lowers its
//     Requisition cost while it is in its owner's hand (FR-CST-001).
//   - A card's own discard-active deploy permission ("You may deploy this card
//     from your discard pile by ...", Def.Statics with ability.StaticDeployPermission):
//     its owner deploys it as an ActDeploy with From: ZoneDiscard, and the
//     permission's Costs are paid on declaration (FR-TRN-003, FR-TRN-003a).
//   - Ward on targets, and last-known effective state for effects whose
//     source has left the battlefield.
//   - Destruction replacement (FR-REP-001): a face-up asset's printed "would be
//     destroyed" replacement replaces a death at the death sweep or in a kill step,
//     every death of one moment seeing its source (SVY-R20, SVY-R21), in a fixed
//     order (SVY-R22). When several apply to one death, its controller chooses
//     (DecideReplacement, SVY-R23). A replacement applies at most once to a given
//     asset per sweep or kill step (SVY-R24).
//   - Triggered abilities on a Headquarters, through the same self hooks.
//
// Not built yet:
//   - Continuous abilities other than a card's own Cornered keywords and its own
//     hand-active cost reduction (auras, and "Cornered: This unit has +2 power."),
//     Regroup N, replacement effects other than destruction replacement and the
//     damage pipeline (damage redirection), and match
//     clocks. Cards whose text needs any of these load with Card.Unsupported set
//     and cannot be deployed or enacted.
//     cards.NotYetExecutable lists the in-scope M2 cards still in that state.
//
// Files:
//   - setup.go: Config and NewGame. deck.go: DeckList, validation, loading.
//   - turn.go: beginTurn and endTurn. zones.go: moving cards and assets, deploying a card, and the zone a shape names.
//   - actions.go: Action, Apply, checks and LegalActions. combat.go: attack
//     legality and combat resolution. stack.go: the Resolution Stack.
//   - trigger.go: the pending-trigger buffer and firing. placement.go: placing
//     an effect, its modes and then its targets. costs.go: a card's cost and
//     a placement's costs. resources.go: Requisition and Tactics, generating and
//     converting them, and paying with them. targeting.go: target choice and
//     untargeted picks. filter.go: checking assets and discard-pile cards against
//     an ability's filter. effects.go: running a resolved ability's steps.
//   - state.go: core state types and lookups, and exhausting an asset. card.go: card types.
//     cardload.go: the CSV loader and keyword lines.
//   - events.go: Event, the Observer, match records and emitting. choice.go: decisions.
//     snapshot.go: Snapshot and its deep copies. buildinfo.go: BuildInfo.
//     damage.go: the damage pipeline. sweep.go: the death sweep, destruction
//     and match loss. replace.go: destruction replacement. layers.go: derived power, health and keywords,
//     the Cornered check and EffectiveState.
//     invariants.go: state checks.
//
// Package boundary: this package holds the rules, the game state, the Observer
// and the event and record types, and nothing that drives, prints or presents
// a match. Three client packages import it, and it never imports them:
// pkg/sim (match driving: policies, playouts, simulation, simulation decks and
// rule monitors), pkg/eventlog (event output and replay) and pkg/view (one
// seat's view of the game and its event log). sim and eventlog import only
// engine from this module, and view imports engine and eventlog; cardtown lint
// fences checks those imports.
//
// Enums: this package's enums (ActionKind, Resource, CardType, DecisionKind,
// StackKind, Zone, Phase) are plain iota lists. Their zero value is an
// ordinary first member (ActGenerate, ZoneDeck, PhaseSetup), not a default or
// an unset marker, and ActionKind.String indexes an array by value. The shape
// enums in pkg/engine/ability follow that package's three rules instead.
//
// Where FRD 01 leaves a case undefined, the kernel picks a behavior and marks
// it with an "OPEN(KIM-...)" comment pointing at docs/frd/keyword_interaction_matrix.md
// or an "OPEN(SVY-...)" comment pointing at the survey's ruling questions in
// docs/design/m2_starter_set_ability_survey.md. The fences linter
// (cardtown lint fences) checks both against those files.
package engine
