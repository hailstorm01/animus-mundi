---
id: T-0220
title: "Allshan's Battle Trance: deployed from the discard pile to the Main Theater, with Stalwart and +2 power until end of turn (wave F card, flash; #140 (2))"
kind: card
tier: flash
wave: F
status: tests-review
branch: m2/autoloop-wave-f/T-0220-allshans-battle-trance
depends_on: [T-0219]
claims: [card:Allshan's Battle Trance]
cards: [Allshan's Battle Trance]
requirements: [FR-TRN-003a, FR-LAY-001, FR-LAY-003, FR-MNV-001]
---

## Goal
*Allshan's Battle Trance* (Path Maneuver, R0 T4, mythic; in `data/decks/path-starter.txt`; FRD 02 benchmark #11): "Deploy chosen unit card from your discard pile to the Main Theater. Until end of turn, it gains Stalwart and gets +2 power." is executable from shapes that are `done`: a pile slot (as *Reaver's Rush*), `OpDeployFree` with `To` the Main Theater, and T-0219's `OpGrantKeyword` Stalwart and `OpModifyStat` +2 on the same slot. It leaves `cards.NotYetExecutable`.

## Inputs
- Card text: `data/cards/maneuvers-path.csv:12` (Path Maneuver, R0 T4).
- FRD 01 sections:
  - `FR-TRN-003a` (lines 363-365): an in-resolution free deploy pays no costs. Item 2: an explicit zone overrides the printed ones. *Battle Trance* is its example, and so is *Veil Render* deployed through it with no kill.
  - `FR-MNV-001` / `SVY-R29` (line 361): a Maneuver needs a legal entity for each exactly-one slot when it is declared.
  - `FR-TGT-003` (line 397): a target that left its zone before resolution is skipped, and the Maneuver still goes to the discard pile.
  - Immediate Readiness (line 169): a newly deployed asset enters Readied unless it has Deploys Exhausted, so the deployed unit can attack that turn.
  - `FR-LAY-001` (line 484 on): Layer 3 keyword grants, Layer 5 power.
  - `FR-LAY-003` (lines 496-497, `SVY-R39`): one duration covers both effects of its sentence.
  - `FR-TRN-001` Sub-Phase 4c (line 304): "until end of turn" modifiers expire.
  - `FR-KWD-002c` (line 599): Stalwart.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 207.
- Rulings that apply: `SVY-R29`, `SVY-R39`, `SVY-R40` (wording only), `SVY-R46` (*Veil Render*'s kill belongs only to its own discard-pile permission).
- Directive #140 (2) (on `BOARD.md`): its test is `TestCard_AllshansBattleTrance_ExpiresInSubPhase4c` below.
- Shapes it combines, all `done`:
  - a pile slot `Filter{Allegiance: AllegianceAllied, Kinds: KindUnit, Zone: ZoneDiscard}` (as *Reaver's Rush*, `pkg/engine/cards/path.go:561`, without its power limit);
  - `OpDeployFree` with `To: ZoneMainTheater` (T-0208);
  - `OpGrantKeyword` with `Keyword: KeywordStalwart` and `OpModifyStat` with `Amount{N: 2}` on the same slot, which act on the asset the card became (T-0219). `TestFR_TRN_003a_LaterStepActsOnTheDeployedCard` (`pkg/engine/grant_keyword_test.go:179`) is the synthetic form of this card.

## Shape (the lead's plan)
- `pkg/engine/cards/path.go`: the definition, in alphabetical order (before *Allshan, Chainbreaker Revenant*). `Text` is the CSV text, one line. One target, the pile slot above. Three steps on `SubjectTarget` slot 0, in this order: `OpDeployFree` with `To: ability.ZoneMainTheater`, `OpGrantKeyword` with `Keyword: ability.KeywordStalwart`, then `OpModifyStat` with `Amount: ability.Value{N: 2}`. No new shape, no new field.
- `pkg/engine/cards/cards.go`: remove "Allshan's Battle Trance" from `NotYetExecutable`.
- `pkg/engine/doc.go`: the shape list (line 30) adds `OpGrantKeyword (a keyword granted until end of turn, FR-LAY-001)` after `OpMorphScratch`, and says that a later step on a pile slot acts on the asset an earlier `OpDeployFree` of the same resolution made of its card. Comment only.
- The lead tried this definition locally on 2026-11-11, on the run branch at c2ad364, without committing it. Every existing test passed with the card executable, `TestView_Case8_ExecutableFlag` included (since T-0216 it swaps seat 0's last hand card for an unsupported copy), and `go tool task sim` ran clean. So no existing test needs an edit.

## Tests (card tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/cards_path_battle_trance_test.go`. Use the real card (`s.card("Allshan's Battle Trance")`), with player 0 enacting it from exactly R0 T4 (`setPool`). Synthetic pile units are Reserves-only (`rrUnit` in `cards_path_reavers_rush_test.go`, or `dfUnit`), so the explicit zone is the only thing that puts them in the Main Theater.
- `TestCard_AllshansBattleTrance_DeploysToTheMainTheater`: a Reserves-only unit card with a resource cost, in player 0's discard pile. After resolution:
  - it is an asset in player 0's Main Theater under player 0's control, and out of the pile (`FR-TRN-003a` 2);
  - no `DecideZone` was asked;
  - the pool is R0 T0, because the deploy is free (`FR-CST-001`);
  - it has Stalwart (`g.Keywords`) and its printed power + 2 (`g.Power`), while its printed keywords are unchanged;
  - *Battle Trance* is in its owner's discard pile.
- `TestCard_AllshansBattleTrance_ExpiresInSubPhase4c` (#140 (2)): after the same enact, player 0 ends the turn. The unit is still in the Main Theater, without Stalwart and at its printed power (`SVY-R39`, `FR-TRN-001` 4c).
- `TestCard_AllshansBattleTrance_StalwartTakesOneCombatDamage`: the deployed unit enters Readied (line 169) and attacks an opposing Main Theater unit of power 3 the same turn. It takes 1 combat damage (`FR-KWD-002c`).
- `TestCard_AllshansBattleTrance_GuerrillasDeployTriggerFires`: the real *Devilhorn Guerrillas* (it prints both zones) in the pile, and one opposing unit in the Main Theater. No `DecideZone` is asked. The Guerrillas enter the Main Theater, their "if it was deployed to the Main Theater" trigger deals 2 damage to the opposing unit, and they have power 5 (printed 3 + 2) and Stalwart.
- `TestCard_AllshansBattleTrance_VeilRenderPaysNoKill`: the real *Veil Render* in the pile, and an allied unit in the Main Theater. *Veil Render* enters the Main Theater, the allied unit is alive, and no cost decision was asked (`FR-TRN-003a`'s example, `SVY-R46`).
- `TestCard_AllshansBattleTrance_NotEnactableWithoutAUnitCard` (`SVY-R29`, `FR-MNV-001`), with subtests as in `TestCard_ReaversRush_NotEnactableWithoutLegalCard`:
  - an empty pile; only a Maneuver card and a Structure card in the pile; a unit card only in the opponent's pile. In each, Enact is not offered, and a rejected Enact leaves the card in hand.
  - Enact is offered with one unit card of power 5 in the pile (no power limit, unlike *Reaver's Rush*).
- `TestCard_AllshansBattleTrance_TargetLeftThePile` (`FR-TGT-003`): the chosen card leaves the discard pile after the Maneuver is placed (as in T-0219's `TestFR_TGT_003_LaterStepWithoutTheDeployDoesNothing/CardLeftThePile`, with an observer on the `place` event). Nothing is deployed, an allied unit already in the Main Theater gains neither Stalwart nor power, no `keyword` event is logged, and *Battle Trance* is in its owner's discard pile.
- `TestDefs_MatchRulesText` already checks `Text` against the CSV. Nothing to add there.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none expected. Existing helpers: newScene, card, give, setPool, inDiscard, put, script, do, reject, endTurn, attack, wantDamage, alive, asked, kindsOf, enactAction, offersEnact, assetOf, dcAssetOf, inDiscardPile, effectEvents, recorder, rrUnit, dfUnit, unit, structure, maneuver, steps.

## Allowed Files
- `pkg/engine/cards/path.go`
- `pkg/engine/cards/cards.go`
- `pkg/engine/doc.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "AllshansBattleTrance|TestDefs"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-11, lead-architect (wave F run, session 21): queued, split off T-0219 under the D2 size rule (D2.4). Takes #140 (2) from T-0219, which now holds only the shape.
- 2026-11-11, lead-architect (wave F run, session 22): written in full against the run branch at c2ad364 (Inputs, Shape, Tests, Allowed Files). It is a card task with three non-test files and no new shape: it uses only `done` ones (T-0208, T-0210's pile slot, T-0219). #140 (2) is its `ExpiresInSubPhase4c` test. Waiting triggers checked: #15 doesn't fire (the keyword and the power that expire in 4c make nothing lethal), and neither do #6, #7, #71, #91 (2), #106, #120, #136, #141 (3) or #142. Set to `tests`.
- 2026-11-11, rules-test-author: tests written in `pkg/engine/cards_path_battle_trance_test.go`: TestCard_AllshansBattleTrance_DeploysToTheMainTheater, _ExpiresInSubPhase4c, _StalwartTakesOneCombatDamage, _GuerrillasDeployTriggerFires, _VeilRenderPaysNoKill, _NotEnactableWithoutAUnitCard (4 subtests), _TargetLeftThePile. Helpers requested: none (reuses rrUnit). `go vet` is clean and `cardtown lint rules` passes. All 7 fail now because the card is not executable. Left out: nothing. Note: the Stalwart-combat test uses a custom Reserves-only unit (power 2, health 9) so the 3-power foe's damage is capped to 1. Status set to tests-review.
