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
//     HookDamaged, HookExhausted, HookCombatKill, HookMoves,
//     HookDiscardedFromHand, HookEntersDiscard, HookTurnStart, which resolves in the Ready Phase before the
//     Draw Phase, FR-TRN-001) run through a pending-trigger buffer, in APNAP
//     order with cascades (FR-TRG-002 to FR-TRG-004). A trigger may be
//     optional, and its effects may choose targets ("up to N", "other", a zone)
//     that are verified when it is placed and again when it resolves
//     (FR-TGT-001, FR-TGT-003). A trigger may be limited to N firings each turn
//     per asset (Trigger.MaxPerTurn, FR-TRN-001), and Unbowed (FR-KWD-002a) fires
//     on damage taken. A trigger may work from its card's discard pile instead of
//     the battlefield (Trigger.ActiveZone, survey G1, FR-ZON-004). A trigger
//     may watch another asset's attack instead of its own (Trigger.Watches,
//     Trigger.Watch, FR-CBT-002), and its steps may act on that asset
//     (SubjectWatched). A trigger
//     may be modal ("choose N" modes, chosen on placement before its targets,
//     FR-MOD-001 §5). An optional trigger that could do nothing if accepted
//     (every step acts on a target slot, and no slot has a legal entity) is
//     neither asked nor placed (couldDoNothing, SVY-R48).
//   - Per-seat Chooser decisions with timeout fallbacks (FR-MOD-001), including
//     choosing from a hand, your own or an opponent's revealed one (DecideHand, SVY-R16, SVY-R50), and from your own deck (DecideSearch, SVY-R37, SVY-R38), the zone of a free deploy (DecideZone, SVY-R43) and a look at the top cards of your own deck (DecideLook, SVY-R41–SVY-R43).
//     Each decision names the card whose Deploy, Maneuver or ability is being
//     placed or resolved as it is asked, as both seats see it (Decision.Source,
//     FR-ZON-005), from the scope askFor opens, and its match record carries
//     that card's name (DecisionRecord.Source).
//   - Steps OpScratch, OpDraw, OpDiscard, OpReturnToHand, OpPutOnDeck,
//     OpDealDamage, OpHeal, OpAddArmor, OpModifyStat, OpKill (inline destruction), OpMoveZone, OpSearchDeck, OpDeployFree (a free deploy from a discard pile, FR-TRN-003a), OpMorphScratch (a face-down transformation, FR-ENT-103), OpGrantKeyword (a keyword granted until end of turn, FR-LAY-001) and OpLookTop (a look at the top cards with free deploys and a bottom order, FR-MOD-001 §4), with "all" steps, source-power values, the
//     deploy-zone gate, the TurnYours gate, the Cornered gate, SubjectSelf (tracked in discard,
//     FR-ZON-004) and SubjectOpponents. A later step on a pile slot acts on the asset an earlier
//     OpDeployFree of the same resolution made of its card. SubjectPick is an untargeted pick made as the
//     step resolves (SVY-R3), asked as a DecidePick and not a target (Step.Split makes a pick step's Amount a pool split among the matches as it resolves, SVY-R51, SVY-R52), and a step with IfDone
//     runs only if the step before it did something ("if you do", SVY-R14), and a step with
//     IfHandEmpty runs only if its controller has no cards in hand when it resolves. A damage
//     step with Step.Dealer has an asset chosen for one of its slots deal the damage (SVY-R27).
//   - Target slots that choose once for each card the placement's discard cost
//     discarded (Target.PerDiscard, SVY-R4) or once for every k assets its counted
//     kill cost kills (Target.PerKilled, SVY-R53), those picks different entities with
//     Target.Different (SVY-R54), or that must pick an entity different from an
//     earlier slot's (Target.Another, SVY-R34).
//   - Maneuvers (ActEnact, FR-MNV-001): modes and targets are chosen on placement,
//     as a trigger's are, the effect resolves as a StackManeuver item, and the card
//     then goes to its owner's discard pile. A Maneuver may print additional costs
//     to kill, exhaust or discard (Def.ExtraCosts), paid on declaration
//     (FR-TRN-003a); a kill cost may offer a choice of counts (Cost.Counts),
//     chosen on declaration before the targets and paid in full; the counts offered
//     are those that can be paid and whose targets can be filled (SVY-R53).
//   - Activated abilities (ActActivate, FR-ACT-001, FR-ACT-002): a printed
//     "[Cost]: [Effect]." ability of a battlefield asset (Def.Activated), activated
//     on an empty stack. On declaration its targets are chosen, it goes on the stack
//     as a StackAbility item and its costs are paid, including CostResources,
//     CostExhaustSelf and CostKillSelf (declare, SVY-R2). It is offered only if placeable holds: its
//     costs are payable and each required target or mode has a legal choice (SVY-R36).
//   - A card's own Cornered abilities ("Cornered: Aggressive 3.", "Cornered: This unit
//     has +2 power.", Def.Cornered): g.Keywords adds the keywords as a Layer 3 grant and
//     g.Power the power in Layer 4, on a face-up asset, while its controller is
//     Cornered (FR-LAY-001, FR-KWD-002e, SVY-R18).
//   - A card's own hand-active cost reduction ("costs R1 less to deploy for each
//     unit card in your discard pile", Def.CostReduction): EffectiveCost lowers its
//     Requisition cost while it is in its owner's hand (FR-CST-001).
//   - A card's own discard-active deploy permission ("You may deploy this card
//     from your discard pile by ...", Def.Statics with ability.StaticDeployPermission):
//     its owner deploys it as an ActDeploy with From: ZoneDiscard, and the
//     permission's Costs are paid on declaration (FR-TRN-003, FR-TRN-003a).
//   - Auras (Def.Statics with ability.StaticAura): a face-up battlefield
//     asset gives +Power in Layer 4 and its Keywords as a Layer 3 grant
//     to every battlefield asset its Filter matches, read by Game.Power and
//     Game.Keywords through Game.auras; "in the same zone" is the
//     source's current zone (FR-LAY-001, SVY-R44). An aura's Cornered grant
//     counts while the matched asset's controller is Cornered, through
//     Game.whileCornered (FR-KWD-002e).
//   - Ward on targets, and last-known effective state for effects whose
//     source has left the battlefield.
//   - Destruction replacement (FR-REP-001): a face-up asset's printed "would be
//     destroyed" replacement replaces a death at the death sweep or in a kill step,
//     every death of one moment seeing its source (SVY-R20, SVY-R21), in a fixed
//     order (SVY-R22). When several apply to one death, its controller chooses
//     (DecideReplacement, SVY-R23). A replacement applies at most once to a given
//     asset per sweep or kill step (SVY-R24).
//   - Damage redirection (FR-REP-004 Stage 1): a face-up asset's printed
//     WouldBeDealtDamage replacement deals each damage instance that would be
//     dealt to an asset its Filter matches to the source instead, before any
//     defense, so the source's defenses apply; "in the same zone" is the
//     source's current zone (SVY-R44). When several apply, the damaged asset's
//     controller chooses (DecideReplacement, SVY-R23).
//   - Triggered abilities on a Headquarters, through the same self hooks.
//
// Not built yet:
//   - Continuous abilities other than auras, a card's own Cornered abilities
//     and its own hand-active cost reduction, Regroup N, replacement effects
//     other than destruction replacement and damage redirection, and match
//     clocks. Cards whose text needs any of these load with Card.Unsupported
//     set and cannot be deployed or enacted. No in-scope M2 card is in that
//     state, so cards.NotYetExecutable is empty.
//
// Files:
//   - setup.go: Config and NewGame. deck.go: DeckList, validation, loading.
//   - turn.go: beginTurn and endTurn. zones.go: moving cards and assets, deploying a card, and the zone a shape names (zoneMatches).
//   - actions.go: Action, Apply, checks and LegalActions. combat.go: attack
//     legality and combat resolution. stack.go: the Resolution Stack, and
//     StackLabel, the one wording of what a stack item is.
//   - trigger.go: the pending-trigger buffer and firing. placement.go: placing
//     an effect, its modes and then its targets, the card it shows as its
//     source (shownCard, SourceCard), and whether an optional trigger could do
//     nothing (couldDoNothing). costs.go: a card's cost and
//     a placement's costs. resources.go: Requisition and Tactics, generating and
//     converting them, and paying with them. targeting.go: target choice and
//     untargeted picks. filter.go: checking assets and discard-pile cards against
//     an ability's filter. effects.go: running a resolved ability's steps.
//     freedeploy.go: free deploys and the look at the top cards.
//   - state.go: core state types and lookups, and exhausting an asset. card.go: card types.
//     cardload.go: the CSV loader. The keyword set, its names and line parser are in
//     pkg/engine/ability (keywords.go).
//   - events.go: Event, the Observer, match records and emitting. choice.go: decisions,
//     and the askFor scope that names their source card.
//     snapshot.go: Snapshot and its deep copies. buildinfo.go: BuildInfo.
//     damage.go: the damage pipeline. sweep.go: the death sweep, destruction
//     and match loss. replace.go: destruction replacement. layers.go: derived power, health and keywords,
//     the Cornered check and EffectiveState, and an asset's Modifier list with
//     each modifier's Duration.
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
// an unset marker, and ActionKind.String indexes an array by value. One field
// is the exception: Action.From's zero value reads as the hand, and inside
// this package only Action.pile reads it. The shape enums in pkg/engine/ability
// follow that package's three rules instead.
//
// Where FRD 01 leaves a case undefined, the kernel picks a behavior and marks
// it with an "OPEN(KIM-...)" comment pointing at docs/frd/keyword_interaction_matrix.md
// or an "OPEN(SVY-...)" comment pointing at the survey's ruling questions in
// docs/design/m2_starter_set_ability_survey.md. The fences linter
// (cardtown lint fences) checks both against those files.
package engine
