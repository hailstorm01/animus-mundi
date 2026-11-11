---
id: T-0211
title: OpMorphScratch, turning a battlefield unit face down in place (wave F shape, sonnet)
kind: shape
tier: sonnet
wave: F
status: tests-review
branch: m2/autoloop-wave-f/T-0211-morph-scratch
depends_on: []
claims: [shape:OpMorphScratch (ability.OpMorphScratch; stepMorphScratch; the face-down event EventFaceDown)]
cards: []
requirements: [FR-ENT-103, FR-ZON-004, FR-ZON-005, FR-LAY-001, FR-LAY-002, FR-KWD-002l]
---

## Goal
The `MORPH_SCRATCH` primitive (`FR-ENT-103`, "Face-Down Transformation"): a resolving step turns a battlefield unit face down where it stands. It becomes the 1/1/1/1 Infantry Scratch Company whose only printed ability is Deploys Exhausted, and keeps its readiness, accumulated damage, Armor and Layer 5 modifiers. A unit whose damage reaches its new Maximum Health of 1 dies in the death sweep after the item (`FR-LAY-002`). *Forced Mutation* (T-0212, `queued`) is the first card to print it.

The kernel already reads every characteristic through `Asset.Printed()` ([state.go:83-89](../../../pkg/engine/state.go)), which gives the Scratch face when `FaceDown` is set: Layer 1 (`FR-LAY-001`), keywords (`Keywords`, [layers.go:27](../../../pkg/engine/layers.go)), triggers (a face-down asset fires nothing, [trigger.go:60](../../../pkg/engine/trigger.go), [:90](../../../pkg/engine/trigger.go)), and the face-up reveal when it leaves for a discard pile, hand or deck (it keeps its `*Instance`, `FR-ZON-004`, `FR-ZON-005`). So the step only sets the flag and logs it.

Covers:
- `ability.OpMorphScratch`, acting on assets (`stepAssets`).
- `stepMorphScratch` in `effects.go`, dispatched from `runStep` by one `case` line.
- A `face-down` event kind, `EventFaceDown`, logged once per unit turned.
- `doc.go`'s step list names `OpMorphScratch`.
- Synthetic-card tests only. No card prints it until T-0212.

- Not this task: *Forced Mutation* (T-0212, `queued`); the purge of cards tucked under the unit (`FR-ENT-103`'s second item, `FR-ENT-102`). The kernel has no Attached sub-zone, because no Animus or Path card tucks (only *NO-AH-5*-style Soulless cards do), so the purge has nothing to act on. Directive #142 holds it for the first task that adds tucking. The lead adds the ADR-0005 addendum for the `face-down` event when this task merges.

## Inputs
- FRD 01 v3.23: `FR-ENT-103` (lines 260–262), above all its first item; `FR-ZON-005` (lines 250–254: the Scratch face, and the reveal when a face-down unit leaves for a public or private zone); `FR-ZON-004` (line 241) and its leaves-battlefield item (line 245: a face-down Scratch Company has no printed triggers on the battlefield, so its death triggers never fire); `FR-LAY-001`'s Layer 1 row (line 463: the face-down Layer 1 values); `FR-LAY-002` (lines 487–488: lethal damage is destroyed in the state-based sweep after the resolving item finishes); `FR-KWD-002l` (line 605: Deploys Exhausted does nothing for an asset already on the battlefield).
- Survey: `docs/design/m2_starter_set_ability_survey.md` line 159 (*Forced Mutation*: "`MORPH_SCRATCH` chosen opposing unit", `FR-ENT-103`).
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Step`, line 40, names `MorphScratch`).
- ADR-0005 (`docs/architecture/decisions/0005-structured-game-events-and-observer.md`) and its T-0191 addendum (line 32): an effect logs one event per asset it acts on, with `Target`, `Player` -1 and no `Card`, its text through the asset's `String`.
- Code: `Asset.FaceDown`, `Printed` and `String` ([state.go:66-97](../../../pkg/engine/state.go)); `scratchFace` ([card.go:133-150](../../../pkg/engine/card.go)); `runStep` ([effects.go:76-104](../../../pkg/engine/effects.go)); `stepMoveZone` ([effects.go:181-197](../../../pkg/engine/effects.go)) for a unit-only asset step that skips what it can't change and counts the rest; `stepAddArmor` ([effects.go:244-253](../../../pkg/engine/effects.go)) for the event form; the event kinds ([events.go:253-284](../../../pkg/engine/events.go)); `ability.Op` ([ability.go:283-300](../../../pkg/engine/ability/ability.go)).

### Shape (the engineer builds it)
- **`ability.OpMorphScratch`**, the next `Op` after `OpDeployFree`. Its comment: turn each Subject unit face down in place. It becomes the 1/1/1/1 Infantry Scratch Company whose only ability is Deploys Exhausted, and keeps its zone, controller, readiness, damage, Armor and Layer 5 modifiers (`FR-ENT-103`). A unit already face-down, a Structure and a Headquarters are skipped.
- **`stepMorphScratch(r, st)`** in `effects.go`, one `case` line in `runStep`. For each asset of `stepAssets(r, st)`, in order: skip it if `a.Printed().Type != TypeUnit` or it is already `FaceDown`; otherwise log the event, then set `a.FaceDown = true`. Nothing else changes: not `Exhausted`, `Damage`, `Armor`, `TempPower`, `Zone`, `Controller`, `ID` or `arrived`. It returns how many units it turned.
- **`EventFaceDown = "face-down"`** in `events.go`'s kind list, after `EventPower`, with a comment: an effect turned a battlefield unit face down (`FR-ENT-103`). The event is `Event{Kind: EventFaceDown, Player: -1, Target: a.ID}` with the text `"%s turns face down"` of `a`, logged before the flag is set, so the text names the card while it is still public. No `Card`, as the other effect events.
- **`doc.go`:** the step list names `OpMorphScratch` (a face-down transformation, `FR-ENT-103`).
- Behavior doesn't change for any printed card: none uses `OpMorphScratch`. `go tool task sim-compare -- autoloop/m2-wave-f` must be byte-identical (a gate here).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/morph_scratch_test.go`, with a synthetic Maneuver (`maneuver`, `maneuver_test.go`) for player 0 whose one target slot is `ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}`, whose first step is `{Op: ability.OpMorphScratch, Subject: ability.SubjectTarget}`, and whose later steps (on the same slot, or on player 0) are as each test says. Synthetic units (`unit`) for player 1 on the battlefield, with printed power, health, keywords and triggers per test, and their state (`Exhausted`, `Damage`, `Armor`, `TempPower`) set directly before the Maneuver is enacted. Enact with `enactAction` and resolve the stack. Script only the decisions actually asked: a single legal target is chosen without a question (`FR-MOD-001` §1).
- `TestFR_ENT_103_MorphTurnsUnitFaceDownInPlace` (`FR-ENT-103`, `FR-LAY-001`): a 4-power, 5-health player 1 unit with Stalwart in the Main Theater, Exhausted, no damage, 3 Armor and `TempPower` 2. After resolution it is the same asset (same `ID`, still on the battlefield in player 1's Main Theater, controlled by player 1), `FaceDown`, still Exhausted, with Armor 3 and `TempPower` 2. Its effective power is 3 (1 + 2), its effective Maximum Health is 1, its effective keywords are the Scratch face's (Deploys Exhausted only: no Stalwart), and its printed subtypes are Infantry and Scratch Company. A second case: the same unit Readied stays Readied.
- `TestFR_ENT_103_MorphWithDamageIsDestroyed` (`FR-ENT-103`, `FR-LAY-002`, `FR-ZON-004`, `FR-ZON-005`): a 3-power, 4-health player 1 unit with 1 damage and 2 Armor, printing a death trigger (for example `OpDraw` 1 for its controller). After resolution it is no longer on the battlefield; its card (the same `*Instance`) is in player 1's discard pile under its printed name; one `destroy` event names it; and the death trigger did not fire (player 1's hand size is unchanged): a face-down unit has no printed triggers.
- `TestFR_ENT_103_MorphedUnitFiresNoPrintedTriggers` (`FR-ENT-103`, `FR-ZON-005`): the target unit, in the Main Theater, prints a `HookMoves` trigger (for example `OpDraw` 1 for its controller). The Maneuver's second step is `{Op: ability.OpMoveZone, Subject: ability.SubjectTarget, To: ability.ZoneReserves}`. After resolution the unit is face-down in player 1's Reserves and its trigger did not fire.
- `TestFR_ENT_103_MorphedUnitReturnsToHandAsPrintedCard` (`FR-ENT-103`, `FR-ZON-004`): the Maneuver's second step is `{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget}`. After resolution player 1's hand holds the unit's own `*Instance`, whose card is the printed unit card, and the unit is off the battlefield.
- `TestFR_ENT_103_MorphSkipsUnitAlreadyFaceDown` (`FR-ENT-103`): the Maneuver's second step is `{Op: ability.OpDraw, Subject: ability.SubjectController, Amount: 1, IfDone: true}`. On a face-up player 1 unit, player 0 draws one card. On a face-down Scratch Company of player 1 (`s.putScratch(1)`, the only opposing unit), it stays face-down, no `face-down` event is logged, and player 0 draws nothing (the step turned nothing).
- `TestEffectEvents_MorphLogsTheFaceDownChange` (ADR-0005): in `effect_events_test.go`'s style, a step turning two player 1 units face down (an untargeted `SubjectAll` with the same filter, or two target slots) logs one `face-down` event per unit, in step order, each with `Target` the asset's `ID`, `Player` -1, no `Card`, and the text `"<Card name>#<id> turns face down"`, naming the printed card. Logged before the flag is set.
- **Authorized edit:** `TestEventKinds_ValuesAreTheSchema` in `pkg/engine/eventkinds_test.go` gains the case `{EventFaceDown, "face-down"}`, after `{EventPower, "power"}`. No other change to that test. It lands in the `test(T-0211)` commit.
- **Existing tests:** every other test passes unchanged.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test author keeps any small builder in the test file.

## Allowed Files
- `pkg/engine/ability/ability.go` (`OpMorphScratch`)
- `pkg/engine/effects.go` (`stepMorphScratch`, its `runStep` case)
- `pkg/engine/events.go` (`EventFaceDown`)
- `pkg/engine/doc.go` (the step list)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "Morph|EventKinds"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f` passes (no printed card uses `OpMorphScratch`)

## Log
- 2026-11-10 lead-architect (wave F run, session 11): written against the run branch at de1b1c2, for *Forced Mutation* (`FR-ENT-103`), the next of wave F's deck cards. Searched `pkg/engine` and the task archive for an existing equivalent: no step changes `FaceDown` (only `newAsset` sets it, for a scratch, [zones.go:212](../../../pkg/engine/zones.go)), and `Printed()` already reads the Scratch face for every characteristic, so the step sets the flag and logs it. *Forced Mutation* would need five non-test files with its shape, so under the D2 size rule it splits into this task and T-0212 (*Forced Mutation*, card, flash, `queued`). `FR-ENT-103`'s purge of tucked cards has nothing to act on in M2 (no Attached sub-zone, no Animus or Path card tucks), so it is directive #142, `waiting` for the first task that adds tucking (D2.4). Set to `tests`.
- 2026-11-10 rules-test-author: tests written, `status: tests-review`. New `pkg/engine/morph_scratch_test.go`: `TestFR_ENT_103_MorphTurnsUnitFaceDownInPlace`, `TestFR_ENT_103_MorphKeepsAReadiedUnitReadied` (FR-KWD-002l), `TestFR_ENT_103_MorphWithDamageIsDestroyed`, `TestFR_ENT_103_MorphedUnitFiresNoPrintedTriggers`, `TestFR_ENT_103_MorphedUnitReturnsToHandAsPrintedCard`, `TestFR_ENT_103_MorphSkipsUnitAlreadyFaceDown`, `TestEffectEvents_MorphLogsTheFaceDownChange`. Authorized edit: `{EventFaceDown, "face-down"}` in `TestEventKinds_ValuesAreTheSchema`. Helpers requested: none (small builders `morphManeuver`, `morphEnact` live in the test file). Vet fails only on the missing `ability.OpMorphScratch` and `EventFaceDown`; `cardtown lint rules` passes. Notes: the two-unit event test expects events in `g.Assets()` order; the effective-keywords check expects `Keywords{DeploysExhausted: true}` from `Game.Keywords` for a face-down unit, per this task (the `go doc` comment says a face-down asset has "none", which the engineer should reconcile: FR-ENT-103 says the Scratch face's Deploys Exhausted). No case left out.
