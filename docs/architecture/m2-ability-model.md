# M2: Ability Model

- **Milestone:** M2 | **Source spec:** [FRD 01](../frd/01_Game_Rules_FRD.md) v3.5 (`FR-TRG-*`, `FR-TGT-*`, `FR-MOD-001`, `FR-ACT-*`, `FR-TRN-003a`, `FR-MNV-001`, `FR-LAY-*`, `FR-REP-*`, `FR-ZON-004`, `FR-SET-001`), build order from the [starter set survey](../design/m2_starter_set_ability_survey.md) | **Date:** 2026-10-19

FRD 02 is frozen. This note borrows its names for hooks and primitives (as the survey does) so that the M3 rewrite starts from familiar vocabulary. It does not design from FRD 02, and FRD 01 wins any conflict.

## Goal

Every Animus and Path card executes in `pkg/engine` as a Go value built from a small set of named shapes, and the kernel's shape vocabulary is the M2 write-up that FRD 02/03 are rewritten from in M3. We know it works when:

- every Animus, Path and starter Headquarters card loads with `Unsupported` empty and has at least one `TestCard_*` test;
- `cardtown lint rules` shows every requirement the survey says the starter set reaches as pinned (27 of the 30 that are unpinned today, see survey §1);
- `cardtown sim -p0 ANIMUS -p1 PATH` runs clean on construction-legal decks, and a human can finish a `cardtown play` match.

## Boundaries

Three packages. Imports go one way only: `engine` → `cards` → `ability`. That lets package-`engine` scenario tests use real card definitions without an import cycle.

| Package | Holds | Card names? | Changed by |
| :--- | :--- | :--- | :--- |
| `pkg/engine/ability` | The shape vocabulary: plain data types and enums with no game logic. Its `go doc` output is the M2 write-up. | No (fenced) | `shape` tasks |
| `pkg/engine/cards` | One `ability.Def` per card, keyed by card name, in `animus.go`, `path.go` and `headquarters.go`. | Yes (only place) | `card` tasks |
| `pkg/engine` | The interpreter: events, trigger collection, stack, choices, layers, replacements. | No (fenced) | `shape` tasks |

**Key types in `ability`** (one line each; fields are decided in code):

- `Def`: everything one card prints besides standalone keywords. `Text []string` holds the exact rules-text lines it implements, then `Triggers`, `Activated`, `Statics`, `Replacements`, `Effect` (Maneuvers) and `ExtraCosts` (declaration costs, `FR-TRN-003a`).
- `Trigger`: a `Hook`, gate conditions (`TurnOwner`, `ToZone`, `Cornered`, `MaxPerTurn`), `ActiveZone` (battlefield by default; survey gap G1), `Optional`, an optional non-resource `Cost` (`FR-TRG-005`), and an `Effect`.
- `Hook`: an enum of events a trigger watches: deploy, killed, attacks, damage taken, exhausts, turn start, combat kill, opposing unit attacks, moves zone, enters discard, discarded from hand.
- `Activated`: `Costs` plus an `Effect` (`FR-ACT-001`, `FR-ACT-002`).
- `Cost`: resources, exhaust self, kill self (survey G3), kill/exhaust a chosen allied asset, discard N, move a card between zones.
- `Effect`: upfront `Targets`, then either `Steps` or `Modes` with `ChooseN` (`FR-MOD-001` §5, `SVY-R6`).
- `Target`: a `Filter`, `Min`/`Max`, `Timing` (on placement, or on resolution for dependent clauses per `SVY-R3`/`SVY-R4`), `PerResult` (one choice per earlier outcome), and `Pool` for distributed values (`FR-TGT-002`).
- `Filter`: a conjunction of criteria (`FR-TGT-001`): allegiance, card types, zone (including "same zone as source"), subtype in/out, non-Scratch, non-HQ, exclude self, readiness, stat comparisons.
- `Step`: one primitive (`Op`: `Scratch`, `Draw`, `ReturnToHand`, `PutOnDeck`, `DealDamage`, `Heal`, `AddArmor`, `Kill`, `MoveZone`, `ModifyStat`, `GainControl`, `SearchDeck`, `RevealHand`, `Discard`, `DeployFree`, `LookTop`, `GrantKeyword`, `MorphScratch`, …) applied to a `Subject` (target slot, self, all matching a filter, or a player) with an `Amount Value` and an optional `Duration`.
- `Value`: fixed N, source power, last-known power, a count of matching cards, or the count of results from an earlier step ("for each card discarded this way").
- `Static`: a Layer 3/4 aura, a hand-active cost reduction or a discard-active deploy permission, with `ActiveZone` and `Filter` (`FR-LAY-001`, `FR-CST-001`).
- `Replacement`: watches `WouldTakeDamage` or `WouldBeDestroyed` on assets matching a `Filter`, and substitutes `Steps` (`FR-REP-001`–`FR-REP-003`).

New vocabulary is added only by a `shape` task, which names it in `claims:`. A `card` task may only combine values that already exist, so no card task can contain logic.

**Changes in `engine`:**

- `Card.Ability *ability.Def`. `LoadCards` attaches `cards.Defs[name]` and removes each `Def.Text` line from `Unsupported`. A `Text` line missing from the CSV is a load error, so a reworded card falls back to unsupported instead of silently running stale logic. `Supported()` accepts Maneuvers that have an `Effect`.
- **Events and triggers:** state changes emit internal events. Matching triggers go into a pending buffer that is flushed to the stack in APNAP order after each item's resolution and death sweep (`FR-TRG-003`, `FR-TRG-004`), and after declaration costs (`FR-TRN-003a`).
- **Stack:** new `StackAbility` and `StackManeuver` kinds carry an ability instance: the source ID, a last-known snapshot, chosen targets, modes and pool allocations.
- **Choices:** an exported `Decision` and a per-seat `Chooser` in `Config`. See [ADR-0003](decisions/0003-synchronous-chooser-for-player-choices.md).
- **Actions:** `ActEnact` (Maneuver), `ActActivate` (asset plus ability index), and a `From` zone on `ActDeploy` for discard permissions. Targets, costs and modes for these come through the `Chooser`, so `LegalActions` offers a declaration only if at least one complete, legal set of choices exists.
- **Layers:** `Asset.TempPower` becomes a list of timestamped modifiers (Layers 2, 3 and 5, with a duration), and Layer 4 is computed on demand from battlefield `Statics`. The last-known snapshot stores effective values (power, keywords) at departure, not just the raw asset, because death triggers read them (`FR-ZON-004`).
- **Replacements:** `dealDamage` Stage 1 asks redirection replacements. The death sweep becomes per-entity, with the controller ordering replaced deaths and a sweep between them (`SVY-R5`).
- **Fences:** `cardtown lint fences` also scans `pkg/engine/ability` and keeps exempting `pkg/engine/cards` (harness deliverable 1).

## Decisions

- [ADR-0003](decisions/0003-synchronous-chooser-for-player-choices.md): player choices go through a synchronous `Chooser` called during resolution, not a suspendable state machine.
- [ADR-0004](decisions/0004-abilities-as-data-in-go.md): abilities are data values interpreted by the kernel, and card definitions are Go values in `pkg/engine/cards`.
- [ADR-0002](decisions/0002-single-prng-stream-in-m1.md) still holds. The `FR-MOD-001` random-target fallback draws from the game PRNG.

## Test Plan

- **Rules tests:** one `TestFR_<DOM>_<NNN>_<Behavior>` per behavior, using synthetic test cards built in test files from `ability` values. A synthetic card can exercise a shape before any real card uses it.
- **Card tests:** `TestCard_<CardSlug>_<Behavior>`, at least one per card, played through the real definition via a scenario helper that loads the card from `data/cards` (for example `s.card(name)`). They are written from the CSV text, never from `cards/*.go`.
- **Definition coverage:** `TestDefs_MatchRulesText` checks that every `Def.Text` line exists verbatim on its card. `TestStarterSet_Executable` lists every Animus, Path and starter-HQ card still unsupported, against an allowlist that each card task shrinks. M2 is done when the allowlist is empty.
- **Choices:** a scripted `Chooser` for scenario tests (answers in order, fails the test if asked something unexpected). A random `Chooser` for playouts.
- **Fuzz:** playouts use random policy and random choosers. New invariants: the stack and pending-trigger buffer are empty between actions, no modifier outlives its duration, and a `Chooser` answer is never rejected.
- **Sim:** `cardtown sim -p0 ANIMUS -p1 PATH` on starter decklists, run in every review once decks exist (Runbook D5).

## Deliverables (in order)

1. **Scaffold (harness):** create the `ability` and `cards` packages, `Card.Ability`, text attachment in `LoadCards`, the two coverage tests and the scenario helper that loads a real card; widen the fence scope. Done when the existing tests pass, the 9 already-executable starter cards are unchanged, and `cardtown lint --all` passes.
2. **Starter decks (harness):** a decklist format under `data/decks/`, `engine.LoadDeck` and `ValidateDeck` (`FR-SET-001`), and a `-deck` option on `cardtown sim`. Done when `TestFR_SET_001_*` pass. Real lists wait on open question 1, and synthetic lists cover the tests until then.
3. **Wave A1:** events, triggers on self hooks, APNAP placement and cascade, the `Chooser` with the scripted and random implementations, last-known effective state, and untargeted steps. Done when the 8 A1 cards pass and `FR-TRG-003`, `FR-TRG-004` and `FR-ZON-004` death-trigger tests pass.
4. **Wave A2:** targets, filters, two-stage verification, optional triggers, `FR-MOD-001` fast-path and fallbacks, `DealDamage`, `Heal` and `AddArmor`. Done when the 7 A2 cards pass.
5. **`cardtown play` (text mode):** a terminal `Chooser` and action menu against the random policy. Done when a human finishes a match with the cards executable so far. It goes early so physical-prototype comparison starts while waves land.
6. **Waves B, C, D, E, F, G** in the survey's order (C–G may be reordered). Each is done when its survey rows pass and `cardtown sim` stays clean.
7. **Write-up:** a short `docs/design/m2_ability_shapes.md` that indexes `go doc ./pkg/engine/ability` and records what changed from FRD 02 (survey gaps G1–G3 plus anything later). Done when the M2 done-when checks in `PRODUCT_SUMMARY.md` §3.2 hold.

## Open Questions

1. **Starter decklists (needs the user).** No decklist exists in the repo. Which 50+ card Animus and Path lists did the physical playtests use, and which Headquarters? Recommendation: transcribe the playtest lists. If they no longer exist, `ValidateDeck` plus maximum legal copies of each starter card is a stand-in for `sim`, not for `play`.
2. **Survey gaps G1–G3:** decided in the kernel as the survey proposes. G1 is `Trigger.ActiveZone`, G2 puts keyword grants in Layer 3 per `FR-LAY-001`, and G3 is a `KillSelf` cost. No ruling is needed.
3. **Action clocks:** M2 has no clocks. A `Chooser` may answer "timeout", which applies the `FR-MOD-001` §6 fallback. Clocks and time banks wait for M4/M5.
4. **Hidden information:** in M2 a `Chooser` can read the whole `Game`. The fog-of-war boundary is M5 work (FRD 04 rewrite). `cardtown play` shows the human only their own zones by convention.
5. **Partitioned inspections** (`FR-MOD-001` §4, Devilhorn Scouting Party) are one `Decision` with a per-card destination and a bottom order. The Wave F task fixes the answer encoding.
