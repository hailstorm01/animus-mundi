---
id: T-0226
title: "Eerie Remnant: when put into your discard pile from your hand, you may deploy it from there for free (wave F deck card, flash; #136's card part)"
kind: card
tier: flash
wave: F
status: done
branch: m2/autoloop-wave-f/T-0226-eerie-remnant
depends_on: [T-0225]
claims: [card:Eerie Remnant]
cards: [Eerie Remnant]
requirements: [FR-ZON-004, FR-TRN-003a, FR-MOD-001, FR-KWD-002l]
---

## Goal
*Eerie Remnant* (Path unit, `data/cards/units-and-structures-path.csv:4`; 3 copies in `data/decks/path-starter.txt`): "Deploys Exhausted / When this card is put into your discard pile from your hand, you may deploy it from your discard pile for free." is executable from shapes that are `done`: T-0225's `HookDiscardedFromHand` with `ActiveZone` `ZoneDiscard`, `Optional`, and T-0208's `OpDeployFree` on `SubjectSelf`. It leaves `cards.NotYetExecutable`, the last wave F deck card to do so. With it, directive #136 is fully applied.

## Inputs
- Card text: `data/cards/units-and-structures-path.csv:4-5` (Path unit, Skeleton, Reaver, R1 T0, Reserves, 3/1, yields 0/0). "Deploys Exhausted" is its keyword, read from the CSV (`Card.DeploysExhausted`), not part of the definition's `Text`.
- FRD 01 v3.24:
  - `FR-ZON-004` *Enter-Discard Triggers* (line 250): evaluated against the card's printed face as it enters the discard pile. *Self-Relocating Trigger Tracking* (line 251): a trigger that moves "it" out of the discard pile tracks that card there. *Leaves-Battlefield Triggers* (line 248): a card killed on the battlefield goes to the discard pile, but not from the hand, so this trigger doesn't fire.
  - `FR-TRN-003a` (line 362: a discard cost; line 363: a free deploy pays no costs, `FR-CST-001`, and with no zone in the text the card enters one of its printed zones, here only Reserves).
  - `FR-SET-004` *Immediate Readiness* (line 169) and `FR-KWD-002l`: a deployed asset enters Readied unless it has Deploys Exhausted.
  - `FR-MOD-001` §6 (line 421) and §7: an optional trigger is asked on placement and times out to declined (`SVY-R9`, line 88).
  - `FR-TRG-002`, `FR-TRG-003`: a cost's trigger is placed above the item it was paid for, and resolves first.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 201 (gap G1, line 47); line 177 (*Chainbreaker Irregulars*' discard "can trigger Eerie Remnant").
- Directive #136 (`BOARD.md`): the card part, its shape applied by T-0225 (merge d7c70c9). This task applies the rest.
- Shapes it combines, all `done`: `HookDiscardedFromHand`, `Trigger.ActiveZone`, `pileSource` and the `Tracked` pile item (T-0225); `OpDeployFree` with `SubjectSelf` (T-0208). `pkg/engine/pile_trigger_test.go`'s "Remnant" is the synthetic form of this card.
- Deck cards that discard from hand, used below: *Improvised Orders* (Maneuver, R0 T2: "Draw 3 cards, then discard a card."), *Steel Rain* (Maneuver, R0 T4: a discard of up to 3 as an additional cost), *Chainbreaker Irregulars* (Unbowed: draw a card, then discard a card).

## Shape (the lead's plan)
- `pkg/engine/cards/path.go`: the definition, in alphabetical order (after *Devilhorn Scouting Party*, before *Fortify*). `Text` is the CSV's second line only: `"When this card is put into your discard pile from your hand, you may deploy it from your discard pile for free."`. One trigger: `Hook: ability.HookDiscardedFromHand`, `ActiveZone: ability.ZoneDiscard`, `Optional: true`, and an effect of one step, `Op: ability.OpDeployFree`, `Subject: ability.SubjectSelf`, no `To` (the card prints Reserves only). No targets, no new shape, no new field.
- `pkg/engine/cards/cards.go`: remove "Eerie Remnant" from `NotYetExecutable`.
- `pkg/engine/doc.go`, comment only: in the triggers bullet, the hook list `(HookDeploy, HookKilled, HookAttacks, HookDamaged, HookExhausted, HookCombatKill, HookMoves)` becomes `(HookDeploy, HookKilled, HookAttacks, HookDamaged, HookExhausted, HookCombatKill, HookMoves, HookDiscardedFromHand)`, and after `fires on damage taken.` add the sentence `A trigger may work from its card's discard pile instead of the battlefield (Trigger.ActiveZone, survey G1, FR-ZON-004).` Nothing else in `doc.go` changes.
- The definition was tried locally on 2026-11-12, on the run branch at dfda605 (by the session before this one, which hit a usage limit before committing anything; its trial worktree held only this definition and the `NotYetExecutable` removal). `go tool task check` passed with no existing test edited (`TestView_Case8_ExecutableFlag` and the replay and sim-policy tests unchanged), and `go tool task sim` ran clean: Path 50/50 executable, P0 63.5% / P1 35.0% / draws 1.5%, every invariant holding. The sim's policy has no case for `DecideOptional`, so its chooser answers through the default; nothing to edit there.

## Tests (card tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/cards_path_remnant_test.go`. Use the real cards (`s.card("Eerie Remnant")` and the deck cards named), give player 0 an exact hand and pool (`setPool`), and set decks exactly with `onTop` where a draw happens. Answer the optional question with `Answer` values through `s.script`; check the asked decisions with `s.asked`. "In Reserves, Exhausted" means an `*Asset` for Remnant's instance under the named player, in `ZoneReserves`, with `Exhausted` true, and the instance is no longer in that player's discard pile.
- `TestCard_EerieRemnant_DeploysFromHandAsAThreeOneExhaustedForOneRequisition` (`FR-KWD-002l`): deployed normally from hand with R1, it is a 3/1 in Reserves, Exhausted, the pool R0, and no question is asked.
- `TestCard_EerieRemnant_ImprovisedOrdersDiscardDeploysItExhausted` (`FR-ZON-004` line 250, `FR-TRN-003a` line 363): player 0's hand is *Improvised Orders* and Remnant, the pool R0 T2, and the deck three vanilla cards. Player 0 enacts the Orders, picks Remnant in the `DecideHand`, and accepts the one `DecideOptional`. After everything resolves: Remnant is in Reserves, Exhausted; the pool is R0 T0 (the deploy is free); the three drawn cards are in hand; *Improvised Orders* is in the discard pile.
- `TestCard_EerieRemnant_SteelRainCostDiscardFiresAboveTheManeuver` (`FR-TRN-003a` line 362, `FR-TRG-002`): player 0's hand is *Steel Rain*, Remnant and one vanilla card; the pool R0 T4; player 1 has a unit in the Main Theater with health above 2. Player 0 enacts *Steel Rain*, discarding only Remnant to its cost, with the one damage target player 1's unit, and accepts the optional question. Remnant's trigger is placed above *Steel Rain* and resolves first: when Remnant enters the battlefield, player 1's unit has taken no damage yet (check through the event log's order, the deploy before the damage, or by resolving one item at a time). After everything: Remnant is in Reserves, Exhausted, and player 1's unit took 2.
- `TestCard_EerieRemnant_IrregularsUnbowedDiscardDeploysIt` (survey line 177): *Chainbreaker Irregulars* under player 1 in the Main Theater; player 1's hand is Remnant only, its deck one vanilla card. Player 0 attacks the Irregulars with a unit that leaves them alive. Player 1 draws, discards Remnant (picked in the `DecideHand`), and accepts the `DecideOptional`: Remnant is in player 1's Reserves, Exhausted, under player 1.
- `TestCard_EerieRemnant_DeclinedStaysInThePile` (`FR-MOD-001` §7): as `_ImprovisedOrdersDiscardDeploysItExhausted`, but player 0 answers no to the optional question: Remnant is in player 0's discard pile and is not an asset.
- `TestCard_EerieRemnant_TimeoutStaysInThePile` (`FR-MOD-001` §6, `SVY-R9`): the same, but seat 0 times out on the `DecideOptional`: nothing is placed, and Remnant stays in the discard pile.
- `TestCard_EerieRemnant_KilledOnTheBattlefieldFiresNothing` (`FR-ZON-004` line 248): Remnant is an asset under player 0 (`s.put`), and player 1's attack or a damage step kills it. It is in player 0's discard pile, no `DecideOptional` is asked of either player, and it is not an asset afterwards.
- `TestDefs_MatchRulesText` already checks `Text` against the CSV. Nothing to add there.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none. Note: `setPool` does not exist; the tests set `Players[0].Req`/`Tac` directly. Existing helpers used: newScene, card, give, byKind, askedKind, onField, inZone, onTop, put, inDiscard, script, do, asked, attack, enactAction, deployPath, assetOf, inDiscardPile, alive, wantDamage, unit, vanilla.

## Allowed Files
- `pkg/engine/cards/path.go`
- `pkg/engine/cards/cards.go`
- `pkg/engine/doc.go`

## Done When
- [x] The listed tests pass: `go test ./pkg/engine -run "EerieRemnant|TestDefs"`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-12, lead-architect (wave F run, session 27): split from T-0225 under the D2 size rule, `queued` on it, Covers only.
- 2026-11-12, lead-architect (wave F run, session 28): written in full against the run branch at dfda605 (Inputs, Shape, Tests, Allowed Files). A card task with three non-test files and no new shape. The local trial found no existing test to edit. Waiting triggers checked: #6, #7, #15 (a discard hook is not a turn-start, turn-end or draw hook), #71, #91 (2), #106, #120, #141 (3) and #142 don't fire. Set to `tests`.
- 2026-11-12, rules-test-author (wave F run): wrote `pkg/engine/cards_path_remnant_test.go`, 7 tests, no helpers requested (local helpers `remnantChooser`, `remnantExhaustedInReserves`, `remnantOrdersScene` live in the test file; `setPool` named in the plan does not exist, so the pool is set through `Players[p].Req/Tac`). Tests: `TestCard_EerieRemnant_DeploysFromHandAsAThreeOneExhaustedForOneRequisition`, `_ImprovisedOrdersDiscardDeploysItExhausted`, `_SteelRainCostDiscardFiresAboveTheManeuver` (deploy event seq before the damage event seq), `_IrregularsUnbowedDiscardDeploysIt`, `_DeclinedStaysInThePile` (answer `no`), `_TimeoutStaysInThePile`, `_KilledOnTheBattlefieldFiresNothing` (Remnant attacks a 5-power unit and dies). Current result: 6 fail as expected (card not executable yet); the killed-on-battlefield test passes today as a regression guard. `go vet` clean, `cardtown lint rules` passes. Nothing left out.
- 2026-11-12, lead-architect (wave F run, session 28): tests reviewed (81f9167). All seven match the Tests list and the FRD 01 text they cite (`FR-ZON-004` lines 248, 250; `FR-TRN-003a` lines 362-363; `FR-TRG-002`, `FR-TRG-003`; `FR-MOD-001` §6, §7, `SVY-R9`; `FR-KWD-002l`), and they assert outcomes: the Steel Rain case compares the deploy and damage events' sequence numbers. Six fail on the task branch (the card is on `NotYetExecutable`); `_KilledOnTheBattlefieldFiresNothing` passes already, as a regression guard. All seven and `TestDefs_MatchRulesText` pass against the Shape's definition, tried in the review worktree and discarded. (`setPool` does exist, in `activated_test.go:63`; setting the pool directly is equivalent.) Tests approved; set to `ready`.
- 2026-11-12, kernel-engineer (wave F run, T-0226): implemented Eerie Remnant ability definition in pkg/engine/cards/path.go (Hook: HookDiscardedFromHand, ActiveZone: ZoneDiscard, Optional: true, Step: OpDeployFree with SubjectSelf), removed Eerie Remnant from NotYetExecutable in pkg/engine/cards/cards.go, and updated doc comment in pkg/engine/doc.go (HookDiscardedFromHand added to hook list, ActiveZone/discard pile triggers documented). Reused existing shapes HookDiscardedFromHand, ZoneDiscard, and OpDeployFree with SubjectSelf. Ran `go test ./pkg/engine -run "EerieRemnant|TestDefs"` (all 7 scenario tests + TestDefs pass), `go tool task check` (gofmt, go vet, all tests, full linter pass), and `go tool task sim` (clean run on starter decks: Path 50/50 executable, all invariants held). Diff strictly limited to allowed files. Set to `review`.
- 2026-11-12, lead-architect (wave F run, session 28): code reviewed (e72986c; `agy` conversation `3fc3dced-8384-4dac-9684-07e20fd2cbd5`, no denied actions). Only the three Allowed Files changed, and no test after the `test(T-0226)` commit. The definition is the Shape's exactly; `doc.go` gained the hook and the sentence as asked, with its bullet re-wrapped (comment only). No card name in engine code, no new shape. Re-run in a review worktree: `go tool task check` passes (fences included) and `go tool task sim` runs clean (Path 50/50 executable, P0 63.5% / P1 35.0% / draws 1.5%, every invariant holding). Approved, merged into `autoloop/m2-wave-f` (5776555). #136 is fully applied. `done`.
