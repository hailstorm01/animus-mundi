---
id: T-0310
title: Make Bring It Down executable (2 damage to a chosen non-Scratch asset; first Soulless card, wave A, card, flash)
kind: card
tier: flash
wave: A
status: tests-review
branch: m3/autoloop-wave-a/T-0310-bring-it-down
depends_on: [T-0309]
claims: [card:Bring It Down, file:pkg/engine/cards/soulless.go]
cards: [Bring It Down]
requirements: [FR-MNV-001, FR-TGT-001, FR-TGT-003, FR-REP-004, FR-LAY-002, FR-KWD-002h, FR-KWD-002k, FR-ZON-005, FR-ENT-103, FR-CBT-003]
---

## Goal
*Bring It Down* (a Soulless Maneuver, R0 T1; in the Soulless starter deck, 4 copies): *"Deal 2 damage to chosen non-Scratch asset."* It only combines shapes that are `done`: a Maneuver's `Effect` with one target slot whose filter is `Scratch: ability.ScratchExcluded` and nothing else (the second slot of *Pincer Tactics* in `pkg/engine/cards/path.go`, which prints the same clause), and one `OpDealDamage` step on that target. A Headquarters is a legal non-Scratch asset (`SVY-R29`), so the Maneuver is always enactable.

Covers:
- `pkg/engine/cards/soulless.go`, a new file with the definition, as under Definition below. It is the first Soulless card file (survey §2, wave A; run file header).
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/maneuvers-soulless.csv` line 2): Soulless Maneuver, R0 T1.
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Soulless: deck" (*Bring It Down*).
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): `FR-MNV-001` and `SVY-R29` (lines 394-395), `FR-TGT-001` (line 408), `FR-TGT-003` (line 435), `FR-LAY-002` (line 534), `FR-REP-004` (line 562, Stage 2 at line 585), `FR-KWD-002h` Ward (line 648), `FR-KWD-002k` Armor (line 651), `FR-ZON-005` (line 288), `FR-ENT-103` (line 297), `FR-CBT-003` (line 630).
- Rulings that apply: `SVY-R29` (a Headquarters is a legal non-Scratch asset), `SVY-R26` (presence only between the two verifications).
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it)
`soulless.go` has the same form as `concordia.go`: `package cards`, the `ability` import, an `init` that copies `soullessDefs` into `Defs`, and `soullessDefs` with the comment `// soullessDefs holds ability definitions for Soulless cards.` Its one entry:
- `"Bring It Down"`, `Text`: `"Deal 2 damage to chosen non-Scratch asset."` (one string).
- `Effect`:
  - `Targets`: one slot, `Filter: ability.Filter{Scratch: ability.ScratchExcluded}`.
  - `Steps`: `{Op: ability.OpDealDamage, Amount: ability.Value{N: 2}, Subject: ability.SubjectTarget, Target: 0}`.

A lead probe with the definition on the run branch at 2a79a87d, since reverted, passed `go tool task check`, `go tool task sim` (P0 58.5% / P1 39.5% / draws 2.0%, unchanged) and `go tool task sim-matrix`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_soulless_maneuvers_test.go`, from the CSV text, loading the real card (`s.card`), in the style of `cards_concordia_soulwhip_test.go` and the *Pincer Tactics* tests in `cards_path_maneuvers_test.go`. Player 0 enacts with exactly R0 T1 and is the active player, unless the case says otherwise. Reuse the existing test helpers (`s.setPool`, `s.put`, `s.putScratch`, `s.give`, `s.script`, `s.asked`, `s.alive`, `s.ptPlaceHook`, `stormKill`, `enactAction`, `offersEnact`, `inDiscardPile`, `effectEvents`, `targetAsk`, `wantOptions`, `unit`, `morphManeuver`/`morphEnact` from `morph_scratch_test.go`). The other cards may be synthetic.

- `TestCard_BringItDown_DealsTwoToTheChosenAsset` (`FR-MNV-001`, `FR-LAY-002`): with two opposing units of health 5, choosing one leaves it with 2 damage (health 3) and the other undamaged. The T1 is paid, and *Bring It Down* is in player 0's discard pile.
- `TestCard_BringItDown_OffersEveryNonScratchAsset` (`FR-TGT-001`, `SVY-R29`): with an allied and an opposing unit in each of the Reserves and the Main Theater, an opposing structure, and a face-down Scratch Company on each side (`s.putScratch`), the target decision offers exactly the four units, the structure and both Headquarters. Neither Scratch Company is offered.
- `TestCard_BringItDown_UnitTurnedFaceDownNotOffered` (`FR-ENT-103`, `FR-TGT-001`): an opposing unit turned face down (`morphEnact` with `morphManeuver()`) is a Scratch Company, so it is not offered; a face-up opposing unit beside it is.
- `TestCard_BringItDown_EnactableWithOnlyHeadquarters` (`SVY-R29`, `FR-MNV-001`): with only face-down Scratch Companies on the battlefield besides the Headquarters, the Enact is offered (`offersEnact`) and the decision offers exactly the two Headquarters. Choosing player 0's own Headquarters deals it 2.
- `TestCard_BringItDown_DamagesTheOpposingHeadquarters` (`FR-CBT-003`): choosing player 1's Headquarters lowers its remaining Health by 2, and the match isn't over.
- `TestCard_BringItDown_LethalDamageKills` (`FR-LAY-002`): a chosen opposing unit with health 2 is destroyed, and its card is in player 1's discard pile.
- `TestCard_BringItDown_ResistantAndArmorApply` (`FR-REP-004`, `FR-KWD-002k`): a chosen unit with Resistant 1 takes 1 damage; a chosen unit with Armor 1 loses its Armor and takes 1 damage to its Health.
- `TestCard_BringItDown_WardedOpposingAssetNotOffered` (`FR-KWD-002h`): an opposing Warded unit is not offered. An allied Warded unit is (Ward stops only opponents).
- `TestCard_BringItDown_TargetGoneDealsNothing` (`FR-TGT-003`): the chosen unit is removed between the Maneuver's placement and its resolution: a `recorder` on `s.g.obs` whose `onEvent` removes it on the `place` event for *Bring It Down* (`e.Kind == "place" && e.Card == "Bring It Down"`), as in `TestCard_SoulWhip_TargetGoneDealsNothing`. (`s.ptPlaceHook` fires only for an ability on the stack, never for a Maneuver.) Nothing else is dealt damage, and *Bring It Down* is in player 0's discard pile.
- `TestCard_BringItDown_PaysOneTactics` (`FR-MNV-001`): with R0 T0 the Enact is not offered; with R0 T1 it is, and the pool is empty after it resolves.
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/soulless.go` (new: the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_BringItDown`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [ ] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-25 lead-architect (wave A run, session 12): written against 2a79a87d, with the lead's probe above. `main` has no commits the run branch lacks. Board read twice (D2.0): no directive is due here (#185 has only (4)'s *Awaken Nightmare* half left, wave F), and no waiting trigger has fired. Searched `pkg/engine` and the task archives: every part of the Definition exists (*Pincer Tactics*' second slot and step are the same clause), so no kernel change. Set to `tests`.
- 2026-11-25 rules-test-author: wrote `pkg/engine/cards_soulless_maneuvers_test.go` with all ten listed tests: TestCard_BringItDown_DealsTwoToTheChosenAsset, _OffersEveryNonScratchAsset, _UnitTurnedFaceDownNotOffered, _EnactableWithOnlyHeadquarters, _DamagesTheOpposingHeadquarters, _LethalDamageKills, _ResistantAndArmorApply, _WardedOpposingAssetNotOffered, _TargetGoneDealsNothing, _PaysOneTactics. No helpers requested; no cases left out. `go vet` is clean and `cardtown lint rules` passes. All ten fail now because the card has no definition ("needs rules the kernel does not implement yet"). Set to `tests-review`.
- 2026-11-25 lead-architect: tests returned (D4, round 1). A lead probe with the Definition, since reverted, passes 9 of the 10, and each of 8 mutants (no Scratch filter, 1 damage, 3 damage, units only, no Headquarters, opposing only, Main Theater only, kill in place of damage) fails at least one of those 9. Note:
  1. `TestCard_BringItDown_TargetGoneDealsNothing` fails its setup check ("the chosen unit should have left before resolution") with the Definition in place: `s.ptPlaceHook` (`pile_target_test.go:27`) calls its function only when the top stack item is a `StackAbility`, so it never fires for a Maneuver. The helper was the lead's suggestion in the task text, now corrected under Tests. Remove the unit through a `recorder` on `s.g.obs` whose `onEvent` acts on `e.Kind == "place" && e.Card == "Bring It Down"` while the unit is alive (`stormKill` or `s.g.removeFromBattlefield`), as `TestCard_SoulWhip_TargetGoneDealsNothing` (`cards_concordia_soulwhip_test.go:185`) does. Keep the assertions. Change nothing else. Set to `tests`.
- 2026-11-25 rules-test-author: applied lead note 1 (D4 round 1). `TestCard_BringItDown_TargetGoneDealsNothing` now removes the chosen unit through a `recorder` on `s.g.obs` whose `onEvent` calls `stormKill` on the `place` event for *Bring It Down* while the unit is alive, in place of `s.ptPlaceHook`. Assertions unchanged, nothing else changed, no helpers requested. `go vet` is clean; the test fails only because the card has no definition yet. Set to `tests-review`.
