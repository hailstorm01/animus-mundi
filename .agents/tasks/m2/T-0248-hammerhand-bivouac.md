---
id: T-0248
title: "Hammerhand Bivouac: at the start of your turn, distribute 2 armor among allied units in the same zone (wave H card, flash; directive #157)"
kind: card
tier: flash
wave: H
status: done
branch: m2/autoloop-wave-h/T-0248-hammerhand-bivouac
depends_on: [T-0245, T-0246, T-0247]
claims: [card:Hammerhand Bivouac]
cards: [Hammerhand Bivouac]
requirements: [FR-TRN-001, FR-TGT-001, FR-TGT-002, FR-MOD-001]
---

## Goal
*Hammerhand Bivouac* (Path structure, `data/cards/units-and-structures-path.csv:17`; Territory, Hammerhand; R3 T0, Main Theater, 0/7, yields 2/2; in neither starter deck): "At the start of your turn, distribute 2 armor among allied units in the same zone." becomes executable from shapes that are `done`: T-0245's `HookTurnStart` with `Turn: TurnYours`, T-0246's `ZoneRefSame` and T-0247's `Step.Split` on `OpAddArmor`. The card leaves `cards.NotYetExecutable`. With it, directive #157 is fully applied.

## Inputs
- Card text: `data/cards/units-and-structures-path.csv:17`.
- Board directive #157 (`SVY-R51`, `SVY-R52`), from the [outcome](../../../docs/design/outcomes/2026-11-15-wave-h-bivouac-untargeted-pool.md) (`Status: applied: FRD 01 v3.28`). T-0247 (merge 732c9c5) built the pool shape, with tests (a)–(e) on a synthetic card in `pkg/engine/pool_test.go`. This task adds the real card and pins (a)–(f) on it.
- FRD 01 v3.28: `FR-TRN-001` start-of-turn triggers (line 307); `SVY-R44` "in the same zone" locked on placement (line 398); `FR-TGT-002` *Distributed Pool Allocations* (line 409, `SVY-R51`); `FR-MOD-001` §1 (line 421) and §6 *Splitting a Pool* (line 437, `SVY-R52`: "A timed-out *Hammerhand Bivouac* gives both armor to the allied unit that entered the Main Theater earliest").
- Code of the day (run branch at 4ee33d7): `Step.Split` (`pkg/engine/ability/ability.go:398`), `HookTurnStart` (`:149`), `TurnYours` (`:157`), `ZoneRefSame` (`:290`); `splitAssets` (`pkg/engine/targeting.go`), `stepAddArmor` (`pkg/engine/effects.go`); `NotYetExecutable` (`pkg/engine/cards/cards.go:20`).

## Shape (the lead's plan)
- `pkg/engine/cards/path.go`: the definition, in alphabetical order (after the entry before it, before *Hammerhand Brigands*, `:499`). `Text` is the CSV ability line: `"At the start of your turn, distribute 2 armor among allied units in the same zone."`. One trigger `{Hook: ability.HookTurnStart, Turn: ability.TurnYours}` whose effect has no targets and one step `{Op: ability.OpAddArmor, Amount: ability.Value{N: 2}, Subject: ability.SubjectPick, Split: true, Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, ZoneRef: ability.ZoneRefSame}}`. No new shape, no new field.
- `pkg/engine/cards/cards.go`: remove "Hammerhand Bivouac" from `NotYetExecutable`.
- The lead dry-ran the definition and the `NotYetExecutable` removal on the run branch at 4ee33d7 (2026-11-15, wave H session 9): `gofmt -l pkg cmd` clean, `go tool task check` exit 0 with no existing test edited, `go tool task sim` exit 0 (P0 58.5%, P1 39.5%, draws 2.0%), and `go tool task sim-compare -- autoloop/m2-wave-h` `[PASS]` (byte-identical: the card is in neither starter deck).

## Tests (card tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/cards_path_bivouac_test.go`. Use the real card (`s.card("Hammerhand Bivouac")`), put under player 0 in the Main Theater (`s.put`), and run to the start of player 0's next turn (`s.endTurn()` twice), as `pool_test.go`'s `runPool` does (its helpers `runPool` and `picksAsked` are in the same package and may be reused; don't edit `pool_test.go`). Answer through `s.script`; check asked decisions with `s.asked`. Units have no printed Armor. "Nothing is asked" means no decision of kind `DecidePick` was asked of either seat.
- `TestCard_HammerhandBivouac_SplitsAmongAlliedMainTheaterUnits` (`SVY-R51`, #157 (a)): two allied units A (older) and B in the Main Theater. Player 0 is asked exactly one `DecidePick` decision, `Repeats`, `Min == Max == 2`, options A then B (the Bivouac itself, a structure, is not offered). Subtests: answer `[0, 1]`: each gains 1; `[1, 1]`: B gains 2, A 0. One subtest with an invalid total (`[0]`): an `EventInvalidAnswer` is logged and A, the first in fixed order, gains 2.
- `TestCard_HammerhandBivouac_OneUnitTakesBothUnasked` (`FR-MOD-001` §1, #157 (b)): one allied unit in the Main Theater: it gains 2, and nothing is asked.
- `TestCard_HammerhandBivouac_NoAlliedUnitDoesNothing` (`SVY-R51`, #157 (c)): one enemy unit in the Main Theater and no allied unit there: no asset's Armor changes, no `EventArmor` is logged, and nothing is asked.
- `TestCard_HammerhandBivouac_TimeoutPutsBothOnEarliestEntered` (`SVY-R52`, #157 (d)): three allied units in the Main Theater, where the one listed first in `Options` entered the Main Theater last (put it in Reserves first, put X then Y in the Main Theater, then `s.moveTo` it there). Player 0 times out (`Answer{Timeout: true}`): X gains 2, the other two 0.
- `TestCard_HammerhandBivouac_OffersOnlyMainTheaterUnits` (`SVY-R44`, #157 (e)): two allied units in the Main Theater, one allied unit in player 0's Reserves, one in player 0's Headquarters Zone, and one enemy unit in the Main Theater. The one decision's options are exactly the two allied Main Theater units.
- `TestCard_HammerhandBivouac_ArmorPersistsAndAccumulates` (#157 (f)): one allied unit in the Main Theater. At the start of player 0's turn it gains 2; it still has 2 after player 0's turn ends and throughout player 1's turn (no trigger fires at the start of player 1's turn: Armor unchanged and nothing asked then), and at the start of player 0's following turn it has 4.
- `TestDefs_MatchRulesText` already checks `Text` against the CSV.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none expected. If the test author needs one, it names it here.

## Allowed Files
- `pkg/engine/cards/path.go`
- `pkg/engine/cards/cards.go`

## Done When
- [x] The listed tests pass: `go test ./pkg/engine -run "HammerhandBivouac|TestDefs"`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-compare -- autoloop/m2-wave-h` passes (*Hammerhand Bivouac* is in neither starter deck)

## Log
- 2026-11-15 lead-architect (wave H run, session 8): queued, split off T-0247 under the D2 size rule (shape and card together touch six non-test files). Waits for T-0247.
- 2026-11-15 lead-architect (wave H run, session 9): T-0247 is `done` (merge 732c9c5), and the field names it merged are the ones planned. Wrote Inputs, Shape, Tests and Allowed Files against the run branch at 4ee33d7 and dry-ran the definition (above). No new shape: searched `pkg/engine/cards` for another pool card, none. Re-read `BOARD.md` before committing: no directive added since; #157 is the one this task applies. Status `tests`.
- 2026-11-15 rules-test-author: wrote pkg/engine/cards_path_bivouac_test.go: TestCard_HammerhandBivouac_SplitsAmongAlliedMainTheaterUnits (3 subtests), _OneUnitTakesBothUnasked, _NoAlliedUnitDoesNothing, _TimeoutPutsBothOnEarliestEntered, _OffersOnlyMainTheaterUnits, _ArmorPersistsAndAccumulates. Reuses runPool/picksAsked from pool_test.go. Helpers requested: none. Cases left out: none. go vet clean, cardtown lint rules PASS; tests fail until the card is defined (NoAlliedUnitDoesNothing passes trivially now).
- 2026-11-15 lead-architect (D4): tests approved. Each checks what `SVY-R44`, `SVY-R51`, `SVY-R52` and #157 give, on the real card: one `DecidePick`, `Repeats`, `Min == Max == 2`, over the two allied Main Theater units oldest first (the Bivouac itself not offered), with each answer's shares and an invalid total falling back to the first in fixed order (a); one unit takes both unasked (b); no allied unit does nothing and asks nothing (c); a timeout puts both on the unit that entered the Main Theater earliest, not the first option (d); Reserves, Headquarters and enemy units aren't offered (e); nothing at the start of player 1's turn, and the armor persists and reaches 4 at player 0's following turn (f). `runPool` and `picksAsked` reused from `pool_test.go`, which is unchanged. Status `ready`, dispatched to `kernel-engineer` (flash).
- 2026-11-15 kernel-engineer: defined Hammerhand Bivouac in pkg/engine/cards/path.go with HookTurnStart, TurnYours, OpAddArmor Amount 2, SubjectPick, Split true, and Filter targeting allied units in ZoneRefSame; removed Hammerhand Bivouac from NotYetExecutable in pkg/engine/cards/cards.go. Reused HookTurnStart (T-0245), ZoneRefSame (T-0246), and Split pool logic (T-0247). Verified: go test ./pkg/engine -run "HammerhandBivouac|TestDefs" PASS; go tool task check PASS; go tool task sim PASS (200 games, P0 58.5%, P1 39.5%, draws 2.0%); go tool task sim-compare -- autoloop/m2-wave-h PASS. Status review.
- 2026-11-15 lead-architect (D5): approved. The diff after the test commit touches only the two Allowed Files and is the planned definition: one `HookTurnStart` trigger with `Turn: TurnYours` and one `OpAddArmor` 2 `SubjectPick` step with `Split` over allied units at `ZoneRefSame`, and the `NotYetExecutable` removal. No new shape, no test edited, no card name in engine code. Re-ran in a review worktree: `go test ./pkg/engine -run "HammerhandBivouac|TestDefs"` ok, `go tool task check` exit 0, `go tool task sim` exit 0 (P0 58.5%, P1 39.5%, draws 2.0%), `go tool task sim-compare -- autoloop/m2-wave-h` `[PASS]`. Merged into `autoloop/m2-wave-h`. #157 fully applied. Status `done`.
