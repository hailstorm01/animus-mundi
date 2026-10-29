# M2: Ability Model

- **Milestone:** M2 | **Source spec:** [FRD 01](../frd/01_Game_Rules_FRD.md) v3.5 (`FR-TRG-*`, `FR-TGT-*`, `FR-MOD-001`, `FR-ACT-*`, `FR-TRN-003a`, `FR-MNV-001`, `FR-LAY-*`, `FR-REP-*`, `FR-ZON-004`, `FR-SET-001`), build order from the [starter set survey](../design/m2_starter_set_ability_survey.md) | **Date:** 2026-10-19

FRD 02 is frozen. This note borrows its names for hooks and primitives (as the survey does) so that the M3 rewrite starts from familiar vocabulary. It does not design from FRD 02, and FRD 01 wins any conflict.

## Goal

Every Animus and Path card executes in `pkg/engine` as a Go value built from a small set of named shapes, and the kernel's shape vocabulary is the M2 write-up that FRD 02/03 are rewritten from in M3. We know it works when:

- every Animus and Path card, plus the starter Headquarters *Animoceps Rager* and *Eternal Watchman*, loads with `Unsupported` empty and has at least one `TestCard_*` test;
- `cardtown lint rules` shows every requirement the survey says the starter set reaches as pinned (27 of the 30 that are unpinned today, see survey §1);
- `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` runs clean on the construction-legal starter decks, and a human can finish a `cardtown play` match.

## Boundaries

Six packages. The rules core imports one way only: `engine` → `cards` → `ability`. That lets package-`engine` scenario tests use real card definitions without an import cycle. Three client packages sit on top of `engine` and are never imported by it (board directive #22, T-0024; directive #24, T-0036). They live at the top of `pkg/`, beside `pkg/engine` rather than inside it (directive #31, T-0039), so that `pkg/engine/...` is exactly the rules core. `sim` and `eventlog` import `engine` only, `view` imports `engine` and `eventlog`, and `cmd/cardtown` imports all of them. `cardtown lint fences` checks these import directions, and the card-name fence covers every package under `pkg/` except `pkg/engine/cards` and `pkg/toolchain` (directive #30, T-0038).

| Package | Holds | Card names? | Changed by |
| :--- | :--- | :--- | :--- |
| `pkg/engine/ability` | The shape vocabulary: plain data types and enums with no game logic. Its `go doc` output is the M2 write-up. | No (fenced) | `shape` tasks |
| `pkg/engine/cards` | One `ability.Def` per card, keyed by card name, in `animus.go` and `path.go`, and in `headquarters.go` once Headquarters abilities land (board directive #45). `cards.go` holds the `Defs` map and the `NotYetExecutable` allowlist. | Yes (only place) | `card` tasks |
| `pkg/engine` | The interpreter: events, trigger collection, stack, choices, layers, replacements. | No (fenced) | `shape` tasks |
| `pkg/sim` | Match driving: `Policy`, `RandomPolicy`, `ProgressPolicy`, `Playout`, `Simulate`/`SimulateWith`, `FactionDeck` and `SimDecks`, and the rule monitors (`Monitors`, `Violation`, `Coverage`, `ReplayCommand`). `CheckInvariants` stays in `engine`. | No (fenced) | `harness` tasks |
| `pkg/eventlog` | Event output: `FormatEvent`, `WriteEventsJSONL`, and the reader and replay glue (T-0023). The event and record types stay in `engine` (ADR-0005). | No (fenced) | `harness` tasks |
| `pkg/view` | One seat's view: `PlayerView` (own hand, counts for the opponent's hand and both decks, public zones, face-down assets as Scratch Company to both seats) and `LogLines` (the event log that seat may read). It imports `engine` and `eventlog` only (board directive #24, T-0036). | No (fenced) | `harness` tasks |

**Key types in `ability`** (one line each; fields are decided in code):

- `Def`: everything one card prints besides standalone keywords. `Text []string` holds the exact rules-text lines it implements, then `Triggers`, `Activated`, `Statics`, `Replacements`, `Effect` (Maneuvers) and `ExtraCosts` (declaration costs, `FR-TRN-003a`).
- `Trigger`: a `Hook`, gate conditions (`TurnOwner`, `ToZone`, `Cornered`, `MaxPerTurn`), `ActiveZone` (battlefield by default; survey gap G1), `Optional`, and an `Effect`. No trigger `Cost` is built: "you may X. If you do, Y" (Infested Apostle, Arkha) is not a cost (`FR-TRG-005`) but a step X on resolution, with Y running only if X was done (`SVY-R14`, `SVY-R15`; board directive #47).
- `Hook`: an enum of events a trigger watches: deploy, killed, attacks, damage taken, exhausts, turn start, combat kill, opposing unit attacks, moves zone, enters discard, discarded from hand.
- `Activated`: `Costs` plus an `Effect` (`FR-ACT-001`, `FR-ACT-002`).
- `Cost`: resources, exhaust self, kill self (survey G3), kill/exhaust a chosen allied asset, discard N, move a card between zones.
- `Effect`: either `Targets` and `Steps`, or `Modes` (each an `Effect`) with `ChooseN` (`FR-MOD-001` §5, `SVY-R6`). A trigger's modes are chosen on placement, before its targets, from the legal ones in printed order, and only the chosen modes' targets are chosen. The stack item carries the chosen modes joined into one `Effect` (T-0083).
- `Target`: a `Filter`, `Min`/`Max`, always chosen on placement (`SVY-R3`), `PerResult` (one choice per cost outcome, such as each card discarded), and `Pool` for distributed values (`FR-TGT-002`). With `Filter.Zone` `ZoneDiscard` it chooses cards in discard piles, re-verified at resolution as still in a pile and matching, and the card steps move the chosen cards still legal (T-0081).
- Untargeted pick ("return an allied Scratch Company unit"): `SubjectPick`, a step's `Filter` and `UpTo` count chosen as the step resolves, from everything that matches then; not a `Target`, so Ward and `FR-TGT-003` do not apply (`SVY-R3`). It asks a `DecidePick` decision, whose fallback `FR-MOD-001` §6 shares with mandatory targets (T-0072, T-0076). With `Filter.Zone` `ZoneDiscard` it picks cards in a discard pile instead of assets, and the card steps (`ReturnToHand`, `PutOnDeck`) move the picked cards (T-0077).
- `Filter`: a conjunction of criteria (`FR-TGT-001`): allegiance, card types, zone (including "same zone as source"), subtype in/out, non-Scratch, non-HQ, exclude self, readiness, stat comparisons.
- `Step`: one primitive (`Op`: `Scratch`, `Draw`, `ReturnToHand`, `PutOnDeck`, `DealDamage`, `Heal`, `AddArmor`, `Kill`, `MoveZone`, `ModifyStat`, `GainControl`, `SearchDeck`, `RevealHand`, `Discard`, `DeployFree`, `LookTop`, `GrantKeyword`, `MorphScratch`, …) applied to a `Subject` (target slot, self, all matching a filter, an untargeted pick, or a player) with an `Amount Value` and an optional `Duration`. `IfDone` runs a step only if the step before it did at least one thing ("if you do", `SVY-R14`).
- `Value`: fixed N, source power, last-known power, a count of matching cards, or the count of results from an earlier step ("for each card discarded this way").
- `Static`: a Layer 3/4 aura, a hand-active cost reduction or a discard-active deploy permission, with `ActiveZone` and `Filter` (`FR-LAY-001`, `FR-CST-001`).
- `Replacement`: watches `WouldTakeDamage` or `WouldBeDestroyed` on assets matching a `Filter`, and substitutes `Steps` (`FR-REP-001`–`FR-REP-003`).

New vocabulary is added only by a `shape` task, which names it in `claims:`. A `card` task may only combine values that already exist, so no card task can contain logic.

**Changes in `engine`:**

- `Card.Ability *ability.Def`. `LoadCards` attaches `cards.Defs[name]` and removes each `Def.Text` line from `Unsupported`. A `Text` line missing from the CSV is a load error, so a reworded card falls back to unsupported instead of silently running stale logic. `Supported()` accepts Maneuvers that have an `Effect`.
- **Events and triggers:** state changes emit internal events. Matching triggers go into a pending buffer that is flushed to the stack in APNAP order after each item's resolution and death sweep (`FR-TRG-003`, `FR-TRG-004`), and after declaration costs (`FR-TRN-003a`).
- **Stack:** a `StackAbility` item carries its source asset ID and card instance, the `ability.Effect` it resolves, the asset IDs chosen per target slot (card instance IDs for a slot in a discard pile, `StackItem.PileSlot`; the view shows those by card name, T-0081), `Tracked` (a death trigger, whose card is tracked into its owner's discard pile, `FR-ZON-004`), and `Keyword`, the keyword its trigger came from (zero for a printed ability). A keyword that is a triggered ability is an `ability.Trigger` built by one constructor in `ability` and placed as a `StackAbility` item; Aggressive N is the first (`ability.Aggressive`, `FR-KWD-002f`, board directive #60). Last-known state lives on `Game`, not on the item. Maneuvers resolve an `Effect` the same way; their stack kind, modes and pool allocations come with the tasks that build them.
- **Choices:** an exported `Decision` and a per-seat `Chooser` in `Config`. See [ADR-0003](decisions/0003-synchronous-chooser-for-player-choices.md).
- **Actions:** `ActEnact` (Maneuver), `ActActivate` (asset plus ability index), and a `From` zone on `ActDeploy` for discard permissions. Targets, costs and modes for these come through the `Chooser`, so `LegalActions` offers a declaration only if at least one complete, legal set of choices exists.
- **Layers:** `Asset.TempPower` becomes a list of timestamped modifiers (Layers 2, 3 and 5, with a duration), and Layer 4 is computed on demand from battlefield `Statics`. The last-known snapshot stores effective values (power, keywords) at departure, not just the raw asset, because death triggers read them (`FR-ZON-004`).
- **Replacements:** `dealDamage` Stage 1 asks redirection replacements. The death sweep becomes per-entity, with the controller ordering replaced deaths and a sweep between them (`SVY-R5`).
- **Effect ops** (board directive #40, T-0050; #51, T-0071 and T-0072): `resolveAbility` re-verifies chosen targets and only dispatches. Each op's body is its own `(g *Game) step<Op>(r *resolution, st)` method, where `resolution` holds the item, the chosen assets and what each step did so far. Every `step<Op>` returns how many things it did (assets acted on, cards moved), and `resolution` records it per step, for `IfDone` and later for counts from an earlier step. A shape task that adds an op adds its `step<Op>`, and never grows the switch with a body.
- **Fences:** `cardtown lint fences` scans every non-test Go file under `pkg/` except `pkg/engine/cards` and `pkg/toolchain` for card names and `OPEN(…)` markers, and checks the core and client packages' imports against a table (harness deliverable 1, widened by T-0038).

**Rule monitors** (board directive #23, T-0025). A rule monitor is an `engine.Observer` in `pkg/sim` that checks named FRD 01 rules in every `cardtown sim` game. It sees only `Event` and `Settled` calls plus the decklists' card data, never a `*Game`, so it cannot change a game: sim output, `-events` and `-log` are byte-identical with monitors on or off (`-monitors=false`, `SimOptions.NoMonitors`). "Before" an action is the most recent settled snapshot before its `action` event, and "after" is the next one; a check still waiting when the game ends is dropped. A monitor records only the first violation of a game, and later events are still counted. Each check adds 1 to the count of every rule ID it exercised, and `Coverage` lists every monitor and rule ID, zero counts included, so `cardtown sim` can print which rules the sim reaches and which it never reaches. A `Violation` names the monitor, all its rule IDs, the game, the seed and the event `Seq`; `ReplayCommand` turns it into the two commands that replay the game to that event (`cardtown sim … -seed S -games 1 -events monitor-S.jsonl`, then `cardtown replay`). Monitors read printed characteristics from the decklist, which equal the effective ones until a task grants or removes keywords. A later `shape` task may add a monitor for the rule it implements.

## Decisions

- [ADR-0003](decisions/0003-synchronous-chooser-for-player-choices.md): player choices go through a synchronous `Chooser` called during resolution, not a suspendable state machine.
- [ADR-0004](decisions/0004-abilities-as-data-in-go.md): abilities are data values interpreted by the kernel, and card definitions are Go values in `pkg/engine/cards`.
- [ADR-0002](decisions/0002-single-prng-stream-in-m1.md) still holds: shuffles and the first-player decision draw from the game PRNG. No `FR-MOD-001` fallback does since `SVY-R16` (T-0056). A timeout picks the first legal options in §6's fixed order: the choosing player's assets before the opponent's, each player's Headquarters, then Reserves, then Main Theater, each zone by arrival. Decision builders may list options in any stable order (targets are listed oldest first, `SVY-R11`). The fallback applies the fixed order itself. Choosing from your own hand will fall back to the card most recently added. Cards in a discard pile rank after every battlefield asset, the choosing player's pile first, each from the top down (directive #59 (1), T-0077).

## Test Plan

- **Rules tests:** one `TestFR_<DOM>_<NNN>_<Behavior>` per behavior, using synthetic test cards built in test files from `ability` values. A synthetic card can exercise a shape before any real card uses it.
- **Card tests:** `TestCard_<CardSlug>_<Behavior>`, at least one per card, played through the real definition via a scenario helper that loads the card from `data/cards` (for example `s.card(name)`). They are written from the CSV text, never from `cards/*.go`.
- **Definition coverage:** `TestDefs_MatchRulesText` checks that every `Def.Text` line exists verbatim on its card. `TestStarterSet_Executable` compares every Animus, Path and starter-HQ card still unsupported against the allowlist `cards.NotYetExecutable` in `pkg/engine/cards/cards.go`. Both tests are protected. The allowlist lives outside the test file, so card tasks shrink it without touching a protected test, and every card task lists `cards.go` under `Allowed Files`. M2 is done when the allowlist is empty.
- **Choices:** a scripted `Chooser` for scenario tests (answers in order, fails the test if asked something unexpected). A random `Chooser` for playouts.
- **Fuzz:** playouts use random policy and random choosers. New invariants: the stack and pending-trigger buffer are empty between actions, no modifier outlives its duration, and a `Chooser` answer is never rejected.
- **Sim:** `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` on the starter decklists, run in every review (Runbook D5).

## Deliverables (in order)

1. **Scaffold (harness):** create the `ability` and `cards` packages, `Card.Ability`, text attachment in `LoadCards`, the two coverage tests and the scenario helper that loads a real card; widen the fence scope. Done when the existing tests pass, the 9 already-executable starter cards are unchanged, and `cardtown lint --all` passes.
2. **Starter decks (harness):** a decklist format under `data/decks/`, `engine.LoadDeck` and `ValidateDeck` (`FR-SET-001`), the two lists in [Starter Decks](#starter-decks) transcribed into it, and a `-deck` option on `cardtown sim`. Done when `TestFR_SET_001_*` pass and both lists validate.
3. **Wave A1:** events, triggers on self hooks, APNAP placement and cascade, the `Chooser` with the scripted and random implementations, self-tracking death triggers (`FR-ZON-004`), and untargeted steps. The last-known effective-value snapshot waits for its first reader (board directive #13). Done when the 8 A1 cards pass and `FR-TRG-003`, `FR-TRG-004` and `FR-ZON-004` death-trigger tests pass.
4. **Wave A2:** targets, filters, two-stage verification, optional triggers, `FR-MOD-001` fast-path and fallbacks, `DealDamage`, `Heal` and `AddArmor`. Done when the 7 A2 cards pass.
4a. **Cleanup (harness), added 2026-10-21 from the user's audit (board directive #26):**
   - T-0032 splits package `engine`'s files by responsibility, as a pure move.
   - T-0033 gives one deck-naming scheme and one build-revision reader.
   - T-0034 removes duplicated rule code.
   - T-0035 splits `cmd/cardtown/main.go` into one file per command.

   Done when all four are merged, with `cardtown sim` output unchanged.
5. **`cardtown play` (text mode):** a terminal `Chooser` and action menu against the random policy. Done when a human finishes a match with the cards executable so far. It goes early so physical-prototype comparison starts while waves land.
6. **Waves B, C, D, E, F, G** in the survey's order (C–G may be reordered). Each is done when its survey rows pass and `cardtown sim` stays clean.
6a. **Wave B.1, cleanup (harness), added 2026-10-28 from structural review 4 of wave B (board directives #62–#67).** It goes between waves B and C, because wave C's Maneuvers build on it. It has no survey row and no cards, so a wave run of it (`go tool task wave-run -- -Wave B.1`) writes `none` on its run file's `Cards`, `Deferred` and `Kept` lines. Its tasks are the directives, in board order:
   - #62: placing modes and targets takes a `placement`, not a `pendingTrigger`, and mode legality is one function (`placement.go`).
   - #63, with #66's doc lines: one generic untargeted pick and one "filter chooses cards in a pile" predicate.
   - #64: one function discards from hand.
   - #65, with #67's doc line: `pkg/view` asks the engine for a chosen pile card.

   Every task is a pure move. Done when #62–#67 are `applied`, each task's `go tool task sim-compare` passed, and `go tool task check` and `go tool task sim` pass.
6b. **Wave B.2, cleanup (harness), added 2026-10-28 from the user's code review (board directives #68–#71).** It goes after wave B.1 and before wave C. Like B.1 it has no survey row and no cards, so a wave run of it writes `none` on its run file's `Cards`, `Deferred` and `Kept` lines. Its tasks, in board order:
   - #68: Berserk leaves the kernel (the keyword, the secondary target, cleave, and their fields and readers), in five parts under the size rule (part (d) became T-0092 and T-0093). The user expects the keyword to be redesigned (survey `R12`), so it is not kept half-built. FRD 01's Berserk text waits for `R12`.
   - #69: attack legality moves from `actions.go`'s `check` to `combat.go` (pure move).
   - #70: the enum convention in `ability`'s package comment, and `ValueFixed` renamed to `ValueSourceFixed`.

   #71 (a step on each player) and the costs question added to #61 are recorded with the wave but wait for their own triggers. Done when #68–#70 are `applied`, the starter sim is byte-identical to wave B.1's end through #68 and `go tool task sim-compare` passes for #69 and #70, and `go tool task check` and `go tool task sim` pass.
6c. **Wave G's destruction replacement goes before wave C (user, 2026-10-29, for board directive #73).** #73's test needs *Shifting Impossibility*, a deck card, replacing a death. The rest of wave G (damage redirection, *Animoceps Spore Cloud*) stays deferred by #14 (b). In order, serial:
   - T-0096: replacement at the death sweep, one death at a time in its controller's order (`SVY-R5`).
   - T-0097: kill steps go through replacement (`SVY-R15`, Open Question 13).
   - T-0100: the affected asset's controller chooses between replacements (`FR-REP-001`, `FR-REP-003`), and `doc.go` lists replacements as built. Split from T-0097 under the D2 size rule (six non-test files), and numbered after T-0099 so that T-0098 and T-0099 keep the ids the board already gives them.
   - T-0098: a Scratch Company filter, and *Shifting Impossibility*.
   - T-0099: #73. The combat-kill hook fires only when the sweep destroys the defender (`SVY-R19`).

   Done when #73 is `applied` and `go tool task check` and `go tool task sim` pass.
7. **Write-up:** a short `docs/design/m2_ability_shapes.md` that indexes `go doc ./pkg/engine/ability` and records what changed from FRD 02 (survey gaps G1–G3 plus anything later). Done when the M2 done-when checks in `PRODUCT_SUMMARY.md` §3.2 hold.

## Open Questions

1. **Starter decklists: answered 2026-10-19.** See [Starter Decks](#starter-decks). The other 38 Animus and Path cards stay in M2 under the current done-when.
2. **Survey gaps G1–G3:** decided in the kernel as the survey proposes. G1 is `Trigger.ActiveZone`, G2 puts keyword grants in Layer 3 per `FR-LAY-001`, and G3 is a `KillSelf` cost. No ruling is needed.
3. **Action clocks:** M2 has no clocks. A `Chooser` may answer "timeout", which applies the `FR-MOD-001` §6 fallback. Clocks and time banks wait for M4/M5.
4. **Hidden information:** in M2 a `Chooser` can read the whole `Game`. `cardtown play` (T-0037) draws the human's screen from `view.PlayerView` and `view.LogLines` (T-0036), which show one seat only what it may see. That boundary is a first draft for M4/M5: the fog-of-war boundary itself is M5 work (FRD 04 rewrite).
5. **Partitioned inspections** (`FR-MOD-001` §4, Devilhorn Scouting Party) are one `Decision` with a per-card destination and a bottom order. The Wave F task fixes the answer encoding.
6. **Trigger-order timeout: answered.** A player who does not order their own simultaneous triggers (`FR-TRG-003`) gets them in the order the kernel collected them (battlefield order, then each card's own trigger order). Accepted by the user 2026-10-19 and in FRD 01 `FR-MOD-001` §6 as `SVY-R7` since v3.6.
7. **A seat cannot end a match (M4/M5 gap, board directive #43).** A `Chooser` can answer or time out, but it cannot concede or abandon. `cardtown play` quits by panicking through the engine's call stack, and `runPlay` recovers only that value. That works for one terminal, not for a networked client. When the M4/M5 client boundary is designed, it needs a way for a seat to concede or leave, together with the asynchronous choosers that board directive #24 left out of M2.

8. **Three wave B rulings: answered 2026-10-23** (`SVY-R14`–`SVY-R17`, FRD 01 v3.11; board directive #47). The lead's earlier recommendations were overruled where they differ.
   - **(a) Infested Apostle and Arkha, "you may X. If you do, Y" (`SVY-R14`, `SVY-R15`):** "you may" is answered and Y's target locked on placement. X is an untargeted pick and step on resolution, not a cost: for the Apostle any allied unit, itself included; for Arkha a unit card in your discard pile, put on the bottom of your deck. Y runs only if X was done, and a replaced X still counts as done. A timeout or no match for X means no Y. Units killed by X have their death triggers placed after the trigger finishes, so Y resolves first.
   - **(b) Chainbreaker Irregulars and every timeout (`SVY-R16`):** no fallback is random. Mandatory targets and untargeted picks time out to the first legal options in `FR-MOD-001` §6's fixed order (see [Decisions](#decisions)). Choosing from your own hand is not an inspection: one card is picked without a question, a timeout picks the card most recently added (for Irregulars, the card just drawn), and an empty hand discards nothing.
   - **(c) Animoceps Shambler (`SVY-R17`):** every change from Readied to Exhausted counts: March, Attack, Generate and exhaust costs. Entering the battlefield exhausted (Deploys Exhausted, which the Scratch Company face prints) does not. A March or Attack trigger is placed on top of that action. A Generate trigger is placed after the generation and resolves before the next action.

9. **A combat-kill attacker that dies in the same sweep (T-0080, `FR-CBT-002` Step 3).** `HookCombatKill` fires at the end of Step 3, before the death sweep, so an attacker that dies alongside its defender still fires, as Unbowed does on lethal damage (`FR-KWD-002a`). FRD 01 doesn't say whether such an attacker "attacks and kills". No M2 card reaches the difference: the only M2 card with the hook is *Eternal Watchman*, a Headquarters, whose death ends the match. Before the first unit with the hook is built (*Dread Manifestation*, Concordia), ask for a ruling.
10. **A card that leaves a discard pile and comes back while it is targeted (T-0081, `FR-ZON-004`, `FR-TGT-003`).** A card keeps its `*Instance` across pile moves, so a pile target that leaves the pile and returns before its item resolves is still legal, and a `Tracked` death trigger still finds its card. `FR-ZON-004` makes a card that changed zones a new object, so both should fizzle. No M2 card moves a card out of a discard pile and back while such an item waits. Before the first one that can, give `Instance` a zone-change count, recorded with the target or the tracked card and compared at resolution.
11. **Legal modes, and the order of several chosen modes (T-0083, `FR-MOD-001` §5, §6).** §6 times out to "the first listed legal mode(s)" but FRD 01 doesn't say what makes a mode illegal. The kernel offers a mode only if each of its exactly-one target slots has a legal entity on placement (an "up to N" slot never makes a mode illegal), the lead's reading of §6, pinned by T-0083's tests. A trigger with no legal mode is placed with no mode and does nothing, as a deploy trigger without legal targets (`FR-TGT-003`), and with fewer legal modes than `ChooseN` all the legal ones are chosen. Several chosen modes resolve in printed order, not pick order. No M2 trigger reaches the last three: both of *Rogue Fleshshaper*'s modes can always choose Fleshshaper itself, and the opposing Headquarters is always a non-Scratch asset. Before the *Ambush Tactics* task (a Maneuver, "Choose two", wave C, whose modes need units that may not exist), ask the user to confirm the legality reading and rule on the other three.
12. **Overkill, M3 work (board directive #72, `SVY-R12`, [outcome](../design/outcomes/2026-10-28-berserk-replaced-by-overkill.md)).** Berserk left the kernel in wave B.2 (#68), and Overkill replaces it. FRD 01 gains its requirement when `pm-frd-drafter` applies the outcome. No M2 card prints it, so it has no M2 task. When the M3 card tasks for *Tempest Surge* and *Redline Infusion* are planned, Overkill is its own hook, not a trigger on `HookCombatKill` (corrected by board directive #73, `SVY-R12` wording). It fires from Step 3 when the attacker's combat damage leaves the defender's Current Health below 0, so it still fires when the defender's death is replaced, and it fixes an "excess" value when it fires. Redline Infusion also needs a new shape: a granted quoted triggered ability until end of turn. Settling tests for that work: `TestFR_KWD_002p_ExactlyLethalIsNotOverkill`, `_ExcessIsHealthBelowZero`, `_ArmorAbsorbsBeforeTheExcess`, `_StalwartDefenderIsNeverOverkilled`, `_FiresWhenTheDefendersDeathIsReplaced`, `_FiresWhenTheAttackerDiesInTheSameSweep`, `_RetaliationNeverOverkills`, `_EffectDamageIsNonCombatAndDoesNotChain`, `_TwoInstancesTriggerSeparately`.
13. **A replacement source lethal in the same sweep: answered 2026-10-29 (user; T-0096, `FR-REP-001`, `SVY-R5`).** *Craterfall Ancient* deals 4 damage to all assets, so *Shifting Impossibility* (4 Health) and three allied Scratch Companies are all lethal. Impossibility replaces all three deaths, then is destroyed. Its own death is one of the sweep's simultaneous events, so it stays on the battlefield until the unreplaced deaths happen. Between replaced deaths, the sweep destroys only assets a byproduct newly left lethal. FRD 01 doesn't record this yet: `game-designer` or `pm-frd-drafter` should add it to `SVY-R5`'s text as a ruling. One reading is still the lead's: when two players' replacements apply in one sweep, the active player's queued deaths are processed first, then the non-active player's, each player ordering their own (T-0096). No M2 matchup reaches it, because only Animus prints a destruction replacement. Ask before the first card outside Animus that does. **Kill steps, answered 2026-10-29 (user; T-0097, `SVY-R15`):** a replacement sees every death that happens at the same moment, as a dies-trigger in other card games sees the creatures that die alongside its source. So when one kill step kills two allied Scratch Companies while *Shifting Impossibility* has 1 Health, both kills are replaced. Nothing is swept while an effect is resolving, so Impossibility is destroyed at the sweep after the item. Record this with the sweep ruling above.

## Starter Decks

Supplied by the user on 2026-10-19. Both are construction-legal under `FR-SET-001`: 50-card main decks, a faction Headquarters, and no card over its rarity limit. Names are as printed in `data/cards`. Two were corrected from the supplied text: *Despoiler's Emergence* and *Devilhorn Guerrillas*. Wave is the survey wave, and "exec" means the card already executes.

**Animus**, Headquarters *Animoceps Rager* (B)

| Card | Copies | Wave | Card | Copies | Wave |
| :--- | :---: | :---: | :--- | :---: | :---: |
| Despoiler Larva | 4 | A1 | Spoil Forest Amalgam | 1 | exec |
| Portent Scrivener | 4 | A1 | Spoil Forest Colossus | 2 | A2 |
| Shifting Impossibility | 2 | G | Despoiler Worm | 1 | A1 |
| Animoceps Shambler | 3 | B | Craterfall Ancient | 2 | A2 |
| Sinuous Growth | 4 | D | Despoiler's Emergence | 1 | C |
| Cliff of Portents | 1 | D | Forced Mutation | 2 | F |
| Spoil Forest Creeper | 4 | A1 | Rapturous Sermon | 2 | C |
| Stormfront of Ideals | 2 | A1 | Animoceps Spores | 3 | C |
| Unceasing Tide | 3 | A1 | Rhapsody's Embrace | 2 | C |
| Veil Render | 1 | F | Funeral Fungus | 2 | C |
| Infested Apostle | 4 | B | | | |

**Path**, Headquarters *Eternal Watchman* (B)

| Card | Copies | Wave | Card | Copies | Wave |
| :--- | :---: | :---: | :--- | :---: | :---: |
| Chainbreaker Sneak | 4 | exec | Devilhorn Guerrillas | 3 | A2 |
| Eerie Remnant | 3 | F | Allshan, Chainbreaker Revenant | 1 | B |
| Hammerhand Brigands | 3 | B | Hammerhand Brigade | 2 | exec |
| Broken Horn Armorer | 2 | F | Arkha, Sun's Fury | 1 | B |
| Macabre Dancers | 2 | E | Boneyard Amalgam | 2 | E |
| Devilhorn Captain | 3 | B | Allshan's Battle Trance | 1 | F |
| Chainbreaker Irregulars | 4 | B | Ambush Tactics | 2 | C |
| Necropolis Hunt | 1 | A2 | Steel Rain | 2 | C |
| Saboteur | 3 | A2 | Devilhorn Sanction | 2 | C |
| Boneyard Scavenger | 2 | D | Improvised Orders | 1 | C |
| Chainbreaker Phalanx | 2 | F | Reaver's Rush | 2 | F |
| | | | Devilhorn Scouting Party | 2 | F |

The decks use 46 distinct cards (3 already executable) and reach every wave from A1 to G. So a full `cardtown play` match with both decks needs all seven waves, and the survey order stands.
