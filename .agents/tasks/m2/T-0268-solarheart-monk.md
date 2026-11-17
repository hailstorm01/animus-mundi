---
id: T-0268
title: "Solarheart Monk: an enter-discard trigger, chosen asset heals 2 (wave H card, flash)"
kind: card
tier: flash
wave: H
status: review
branch: m2/autoloop-wave-h/T-0268-solarheart-monk
depends_on: [T-0267]
claims: [card:Solarheart Monk]
cards: [Solarheart Monk]
requirements: [FR-ZON-002, FR-ZON-003, FR-ZON-004, FR-ZON-005, FR-KWD-002d, FR-TGT-001]
---

## Goal
*Solarheart Monk* (Path unit, "When this card is put into your discard pile, chosen asset heals 2.") is executable: one `HookEntersDiscard` trigger (T-0267) with `ActiveZone: ZoneDiscard`, one target slot over any asset and one `OpHeal` 2, written as *Spoil Forest Colossus* writes its heal (`cards/animus.go:690-707`). It leaves `cards.NotYetExecutable`. No new shape.

## Inputs
- Card text (`data/cards/units-and-structures-path.csv`): Unit, Path, Specialist and Devilhorn, Requisition 8, Tactics 0, Reserves, 6/8, yields 2/2, common. It is in no starter or stress deck.
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`):
  - `FR-ZON-004` *Enter-Discard Triggers* (`:267`, naming *Solarheart Monk*): evaluated on the face-up printed card entering the pile, by any route; a destroyed Scratch Company entering the pile face-up triggers it.
  - `FR-ZON-005` (`:270` on): the destroyed Scratch Company is revealed in the pile.
  - `FR-ZON-002` (`:254`), `FR-ZON-003` (`:230`): a destroyed asset goes to its owner's pile, and the pile trigger is its owner's.
  - `FR-KWD-002d` (`:622`): Heal N. `FR-TGT-001` (`:390`, `:393`): "asset" is any Unit, Structure or Headquarters on the battlefield, either side.
- Survey (`docs/design/m2_starter_set_ability_survey.md`): its row (`:210`), gap G1 (`:47`).
- Code of the day (run branch at `47c8c9f`):
  - `ability.HookEntersDiscard` (`pkg/engine/ability/ability.go:198`), fired from `putCard` (`pkg/engine/trigger.go:92`).
  - *Spoil Forest Colossus*' heal (`pkg/engine/cards/animus.go:690-707`): the same target slot and step on a different hook. *Eerie Remnant* (`pkg/engine/cards/path.go:496`): a pile trigger with `ActiveZone: ZoneDiscard`.
  - `NotYetExecutable` (`pkg/engine/cards/cards.go:13-16`): *Solarheart Monk* is its second entry.
  - T-0267 pinned the hook on synthetic cards in `pkg/engine/enters_discard_test.go` (every route in, the owner's trigger, only into the pile, pile order). This task pins the real card.

## Shape (the lead's plan)
- `pkg/engine/cards/path.go`: one definition, after *Saboteur* and before *Steel Rain*. `Text` is the CSV line, as one string.
  - `Triggers`: one `{Hook: ability.HookEntersDiscard, ActiveZone: ability.ZoneDiscard, Effect: ...}`, not `Optional`.
  - `Effect.Targets`: one `{Filter: ability.Filter{}}` (any asset).
  - `Effect.Steps`: one `{Op: ability.OpHeal, Amount: ability.Value{N: 2}, Subject: ability.SubjectTarget, Target: 0}`.
- `pkg/engine/cards/cards.go`: remove `"Solarheart Monk"` from `NotYetExecutable`.
- The lead dry-ran this definition and the removal on the run branch at `47c8c9f` (2026-11-17, wave H session 30):
  - `gofmt -l pkg cmd` clean.
  - `go tool task check` exit 0.
  - `go tool task sim` exit 0.
  - `go tool task sim-compare -- autoloop/m2-wave-h` `[PASS]`.

## Tests (card tasks: written by rules-test-author; engineers never edit these)
New tests go in a new file `pkg/engine/cards_path_solarheart_monk_test.go`, with a file comment naming the card, its CSV line and T-0268, as `cards_path_monument_test.go` does.
- Use the real card (`s.card("Solarheart Monk")`). Player 0 is the active player.
- The Monk is 6/8, so `killOnP1`'s 5-power hitter does not kill it. Kill it with a kill step (`killTarget()` through `s.enactAlliedStep`, or a Maneuver of your own whose slot can name an enemy unit), or by an attacker with power 8 or more.
- Damaged assets: a unit with Health 5 and `Damage = 3`; `s.wantDamage` checks it.
- Helpers from T-0267 and earlier may be reused: `enactWith`, `discardN`, `killTarget`, `enactAlliedStep`, `pileLogPlaces`, `askedKind`, `putScratch`, `moveTo`, `inDiscard`.

Tests:
- `TestCard_SolarheartMonk_PrintedStats`: the real card is a Path unit, Specialist and Devilhorn, Requisition 8, Tactics 0, prints only Reserves, 6/8, and `Card.Supported` accepts it.
- `TestCard_SolarheartMonk_DiscardedFromHandHealsChosenAsset` (`FR-ZON-004`, `FR-KWD-002d`, `FR-TGT-001`): player 0 discards the Monk from hand by a discard step. One damaged allied unit and one damaged enemy unit are on the battlefield. Player 0 is asked one target decision, whose options include both damaged units and both Headquarters. Answer the enemy unit: it has 1 damage left and the allied unit still has 3. The Monk is in player 0's pile.
- `TestCard_SolarheartMonk_DestroyedOnTheBattlefieldHeals` (`FR-ZON-004`, `FR-ZON-002`): the Monk in player 0's Main Theater is killed by a kill step; its trigger is placed once (`pileLogPlaces`), and the chosen damaged allied unit has 1 damage left.
- `TestCard_SolarheartMonk_DestroyedScratchCompanyHeals` (`FR-ZON-004` `:267`, `FR-ZON-005`): player 1's face-down Scratch Company, whose hidden card is the real Monk, moved to the Main Theater and destroyed by player 0's attacker. The Monk is face-up in player 1's pile, player 1 picks the heal's target once, and the chosen damaged unit of player 1's has 1 damage left.
- `TestCard_SolarheartMonk_ControlledMonkIsItsOwnersTrigger` (`FR-ZON-003`, `FR-ZON-002`): a Monk owned by player 1 and controlled by player 0 (`Controller = 0`) is killed. It is in player 1's pile, not player 0's. Player 1 is asked the target decision, and player 0 is asked none.
- `TestCard_SolarheartMonk_ReturnedToHandHealsNothing` (`FR-ZON-004`): the Monk returned from the battlefield to its owner's hand places no trigger (`pileLogPlaces` 0), and a damaged unit keeps its 3 damage.
- `TestDefs_MatchRulesText` already checks `Text` against the CSV.
- Not this task: pile order between several pile triggers, a Maneuver entering the pile, and a battlefield trigger on the hook. T-0267 pinned each on synthetic cards, and the card adds nothing to them.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none. The tests use existing helpers only; setup local to the new file stays in it.

## Allowed Files
- `pkg/engine/cards/path.go`
- `pkg/engine/cards/cards.go`

## Done When
- [x] The listed tests pass: `go test ./pkg/engine -run "SolarheartMonk|TestDefs"`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-compare -- autoloop/m2-wave-h` passes (the card is in no starter or stress deck)

## Log
- 2026-11-17 lead-architect (wave H run, session 29): queued, split off T-0267 under the D2 size rule. Write its Inputs, Tests and Allowed Files against the code of the day once T-0267 is `done`.
- 2026-11-17 lead-architect (wave H run, session 30): T-0267 is `done` (merge b66cdee) with the hook it planned (`HookEntersDiscard`, fired from `putCard`).
  - Wrote Inputs, Shape, Tests, Allowed Files and Done When against the run branch at `47c8c9f`, and dry-ran the definition (above). `requirements:` gains `FR-ZON-002` and `FR-ZON-003`.
  - `main` has no commits the run branch lacks, so no sync.
  - Directives: none is `open`. I checked each `waiting` one, and none has fired:
    - #6, #7: mode is `serial`.
    - #15 (2): no draw or end-of-turn hook here.
    - #71: no each-player effect.
    - #91 (2): no cost on a mode.
    - #106, #120, #142: no deck search with criteria, no mandatory discard cost, no tuck.
  - No new shape: the card is T-0267's `HookEntersDiscard` with *Spoil Forest Colossus*' heal slot and step.
  - Re-read `BOARD.md` before committing: no directive added since.
  - Status `tests`.
- 2026-11-17 rules-test-author: wrote `pkg/engine/cards_path_solarheart_monk_test.go`: `TestCard_SolarheartMonk_PrintedStats`, `_DiscardedFromHandHealsChosenAsset`, `_DestroyedOnTheBattlefieldHeals`, `_DestroyedScratchCompanyHeals`, `_ControlledMonkIsItsOwnersTrigger`, `_ReturnedToHandHealsNothing`.
  - Helpers requested: none. File-local helpers: `monkPicker`, `hasOptionRef`, `hurtUnit`.
  - Left out: nothing beyond the task's "Not this task" list.
  - `go vet` clean, `cardtown lint rules` PASS. Five tests fail until the card is executable; the returned-to-hand test passes already (it pins a negative).
  - Status `tests-review`.
- 2026-11-17 lead-architect (D4, session 30): tests approved, round 1.
  - Each test checks what the cited rules give, on the real card:
    - Printed stats: Path unit, Specialist and Devilhorn, Requisition 8, Tactics 0, Reserves only, 6/8, yields 2/2, `Supported`.
    - Discarded from hand: one target decision for player 0 over any asset, both sides and both Headquarters (`FR-TGT-001`); the chosen enemy unit heals 2 and the other keeps its damage (`FR-KWD-002d`).
    - Killed by a step on the battlefield, and destroyed as a face-down Scratch Company whose card is the Monk: one placement each, the heal lands (`FR-ZON-004` `:267`, `FR-ZON-005`).
    - Controlled by player 0, owned by player 1: in player 1's pile, player 1 picks the target (`FR-ZON-002`, `FR-ZON-003`).
    - Returned to hand: no placement, no heal (`FR-ZON-004`).
  - Only the new file; no existing test edited.
  - Sanity check in the review worktree: with the Shape's definition applied, `go test ./pkg/engine -run "SolarheartMonk|TestDefs"` passes; without it, five of the six fail (the returned-to-hand negative passes either way). The change was reverted.
  - Status `ready`, dispatched to `kernel-engineer` (flash).
- 2026-11-17 kernel-engineer (flash): implemented Solarheart Monk.
  - Changes:
    - Added definition for "Solarheart Monk" in `pkg/engine/cards/path.go` between Saboteur and Steel Rain, with Text matching the CSV, a single Trigger on `HookEntersDiscard` with `ActiveZone: ZoneDiscard`, one Target slot with `Filter: ability.Filter{}` (any asset, FR-TGT-001), and one Step with `Op: ability.OpHeal`, `Amount: Value{N: 2}`, `Subject: SubjectTarget`, `Target: 0` (FR-KWD-002d).
    - Removed "Solarheart Monk" from `NotYetExecutable` in `pkg/engine/cards/cards.go`.
  - Reuse:
    - `HookEntersDiscard` from T-0267.
    - Target slot and Step structure as in Spoil Forest Colossus (`cards/animus.go:697-707`).
    - Pile trigger active zone `ZoneDiscard` as in Eerie Remnant (`cards/path.go:501`).
  - Verification:
    - `go test ./pkg/engine -run "SolarheartMonk|TestDefs"` PASS.
    - `go tool task check` PASS (gofmt, go vet, all unit & package tests, full cardtown lint suite).
    - `go tool task sim` PASS (200 games clean playout, all invariants held).
    - `go tool task sim-compare -- autoloop/m2-wave-h` PASS (starter and stress sims match autoloop/m2-wave-h byte for byte, exit codes included).
  - Status `review`.
