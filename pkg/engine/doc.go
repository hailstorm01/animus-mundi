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
//     Ward, Versatile, Guardian, Armor N, Deploys Exhausted, Berserk N.
//   - Ability shapes as plain data in pkg/engine/ability, and card definitions
//     keyed by card name in pkg/engine/cards. Engine code never names a card
//     (cardtown lint fences).
//   - Triggered abilities on self hooks (HookDeploy, HookKilled, HookAttacks,
//     HookDamaged, HookExhausted, HookCombatKill) run through a pending-trigger buffer, in APNAP order with cascades
//     (FR-TRG-002 to FR-TRG-004). A trigger may be optional, and its effects
//     may choose targets ("up to N", "other", a zone) that are verified
//     when it is placed and again when it resolves (FR-TGT-001, FR-TGT-003). A
//     trigger may be limited to N firings each turn per asset (Trigger.MaxPerTurn,
//     FR-TRN-001), and Unbowed (FR-KWD-002a) fires on damage taken. A trigger
//     may be modal ("choose N" modes, chosen on placement before its targets,
//     FR-MOD-001 §5).
//   - Per-seat Chooser decisions with timeout fallbacks (FR-MOD-001), including
//     choosing from your own hand (DecideHand, SVY-R16).
//   - Steps OpScratch, OpDraw, OpDiscard, OpReturnToHand, OpPutOnDeck,
//     OpDealDamage, OpHeal, OpAddArmor, OpModifyStat and OpKill (inline destruction), with "all" steps, source-power values, the
//     deploy-zone gate, the TurnYours gate, the Cornered gate, SubjectSelf (tracked in discard,
//     FR-ZON-004) and SubjectOpponents. SubjectPick is an untargeted pick made as the
//     step resolves (SVY-R3), asked as a DecidePick and not a target, and a step with IfDone
//     runs only if the step before it did something ("if you do", SVY-R14).
//   - Ward on targets, and last-known effective state for effects whose
//     source has left the battlefield.
//   - Triggered abilities on a Headquarters, through the same self hooks.
//
// Not built yet:
//   - Maneuvers, activated and continuous abilities, continuous Cornered, Heal N,
//     Regroup N, replacement effects beyond the damage pipeline, and match
//     clocks. Cards whose text needs any of these load with Card.Unsupported set
//     and cannot be deployed.
//     cards.NotYetExecutable lists the in-scope M2 cards still in that state.
//
// Files:
//   - setup.go: Config and NewGame. deck.go: DeckList, validation, loading.
//   - turn.go: beginTurn and endTurn. zones.go: moving cards and assets.
//   - actions.go: Action, Apply, checks and LegalActions. combat.go: attack
//     legality and combat resolution. stack.go: the Resolution Stack.
//   - trigger.go: the pending-trigger buffer and firing. targeting.go: target
//     choice and filters. effects.go: running a resolved ability's steps.
//   - state.go: core state types and lookups. card.go: card types.
//     cardload.go: the CSV loader and keyword lines.
//   - events.go: Event, the Observer, match records and emitting. choice.go: decisions.
//     snapshot.go: Snapshot and its deep copies. buildinfo.go: BuildInfo.
//     damage.go: the damage pipeline. layers.go: derived power, health and keywords.
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
// Where FRD 01 leaves a case undefined, the kernel picks a behavior and marks
// it with an "OPEN(KIM-...)" comment pointing at docs/frd/keyword_interaction_matrix.md
// or an "OPEN(SVY-...)" comment pointing at the survey's ruling questions in
// docs/design/m2_starter_set_ability_survey.md. The fences linter
// (cardtown lint fences) checks both against those files.
package engine
