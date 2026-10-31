---
id: T-0126
title: Make Animoceps Spores executable (Scratch 2, then return an allied Scratch Company unit, wave C card, flash)
kind: card
tier: flash
wave: C
status: review
branch: m2/T-0126-animoceps-spores
depends_on: [T-0072, T-0116, T-0117, T-0121]
claims: [card:Animoceps Spores]
cards: [Animoceps Spores]
requirements: [FR-MNV-001, FR-KWD-002m, FR-ZON-005, FR-TGT-001, FR-MOD-001]
---

## Goal
*Animoceps Spores* (Animus starter deck, 3 copies) runs in the kernel. It is plain `ability` data in `pkg/engine/cards/animus.go`: a Maneuver `Effect` with no targets and two steps, `OpScratch` 2, then `OpReturnToHand` on an untargeted pick (`SubjectPick`, T-0072) of exactly one allied Scratch Company unit. No new shape.

## Inputs
- Card (`data/cards/maneuvers-animus.csv` line 7): *Animoceps Spores*, Animus Maneuver, cost 0 Requisition / 2 Tactics, one rules-text line:
  - "Scratch 2, then return an allied Scratch Company unit to its owner's hand."
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 134: `SCRATCH 2`, then `RETURN_TO_HAND` an allied Scratch Company (untargeted); timeout picks the first in fixed order (`SVY-R16`).
- FRD 01:
  - `FR-MNV-001` (lines 332-334): enacted from Hand, costs paid on declaration, the effect runs when it resolves, then the card goes to its owner's Discard pile. It has no target slot, so `SVY-R29` never stops it being enacted.
  - `FR-KWD-002m` (line 575) and `FR-ZON-005` (lines 227-230): Scratch 2 deploys the top card of the deck face-down into Reserves, twice, as `1/1` Infantry Scratch Company units with Deploys Exhausted. It is safe on an empty deck: no unit, no loss, and it is not a draw. A face-down Scratch Company returned to its owner's Hand reverts to its printed card there (line 230).
  - `FR-TGT-001`, *Targets Are Locked on Placement* (line 355) and *Untargeted Selections* (line 357, `SVY-R3`): the return is not a target. It is picked when the step resolves, after both scratches, from every allied Scratch Company unit then, new or older, in any zone. If none matches, only the return does nothing.
  - `FR-MOD-001` §6 (line 389), *Mandatory Targets and Untargeted Selections* (`SVY-R16`): a timeout picks the first option in fixed order: the choosing player's own assets, Headquarters first, then Reserves, then the Main Theater, each zone in the order its assets entered it.
- Shapes it uses, all `done`: `Def.Effect`, `ActEnact` and `StackManeuver` (T-0116), the `SVY-R29` enact legality (T-0117), `OpScratch` (`stepScratch`, `effects.go`, as *Despoiler Larva*'s "Scratch 1" trigger uses it), `SubjectPick` (T-0072, `pickAssets` in `targeting.go`), and `OpReturnToHand` on an asset (`stepRelocate`, `effects.go`, as *Shifting Impossibility*'s replacement returns an allied Scratch Company unit).
- Rulings: `SVY-R3`, `SVY-R16` as above. No other.

### Definition (the engineer adds it to `animusDefs`)
- *Animoceps Spores*: `Text` holds its line. `Effect`:
  - `Steps`: two, in order: `{Op: ability.OpScratch, Amount: ability.Value{N: 2}}` (`SubjectController`, the zero value), then `{Op: ability.OpReturnToHand, Subject: ability.SubjectPick, Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Scratch: ability.ScratchOnly}}` (`UpTo` 0: exactly one).
- No `Targets`, `Triggers` or `Modes`.
- Remove the name from `cards.NotYetExecutable`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be added to `pkg/engine/cards_animus_maneuvers_test.go`, from the CSV text, loading the real card (`s.card`). Player 0 enacts; player 1 is the opponent. Give player 0 enough Tactics to pay (`enactFor`). Set the decks so the scratched cards are known. The pick is answered with a scripted decision.
- `TestCard_AnimocepsSpores_ScratchesTwoThenReturnsThePickedOne`: player 0 has at least 2 cards in deck and no other unit. After it resolves, the top 2 deck cards became two Scratch Company units in player 0's Reserves; one `DecidePick` decision was asked of player 0, offering exactly those two; the picked unit has left the battlefield and its card (the printed deck card) is in player 0's hand; the other stays in Reserves, Exhausted; *Animoceps Spores* is in its owner's discard pile.
- `TestCard_AnimocepsSpores_MayReturnAnOlderScratchCompany`: with an allied Scratch Company unit already in the Main Theater, the pick offers it and the two new ones; picking it returns it, and both new units stay.
- `TestCard_AnimocepsSpores_TimeoutReturnsTheFirstInFixedOrder`: with an older allied Scratch Company unit in the Main Theater, a scripted timeout returns the first scratched unit (Reserves come before the Main Theater, and it entered Reserves first; `FR-MOD-001` §6, `SVY-R16`).
- `TestCard_AnimocepsSpores_OnlyAlliedScratchCompanyUnitsAreOffered`: an opposing Scratch Company unit and an allied non-Scratch unit are never among the options.
- `TestCard_AnimocepsSpores_OneCardDeckReturnsItWithoutAsking`: with 1 card in player 0's deck and no other allied Scratch Company unit, one unit is scratched and returned to hand with no `DecidePick` asked of the Chooser (its sole option), and player 0 has not lost (a scratch is not a draw, `FR-KWD-002m`).
- `TestCard_AnimocepsSpores_EmptyDeckDoesNothing`: with an empty deck and no allied Scratch Company unit, nothing is scratched, nothing is asked, nothing returns, player 0 has not lost, and the Maneuver still goes to its owner's discard pile.
- `TestCard_AnimocepsSpores_PaysTwoTactics`: enacting it spends 2 Tactics and no Requisition, and it is offered with no unit on the battlefield (no target slot, `SVY-R29`).
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestProgress_Case4_Coverage` and `TestEvents_Case4_EveryActionKind`.

## Helpers Needed
- None expected. The tests use existing helpers (`newScene`, `give`, `card`, `do`, `script`, `asked`), those in `maneuver_test.go` (`enactAction`, `offersEnact`) and `enactFor` in `cards_animus_maneuvers_test.go`; the test file may add its own setup helpers.

## Allowed Files
- `pkg/engine/cards/animus.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [x] The card tests pass: `go test ./pkg/engine -run TestCard_AnimocepsSpores`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-10-31 lead-architect: wave C run, session 13. No `main` commits to sync (`git log autoloop/m2-wave-c..main` is empty). Not `WAVE-COMPLETE`: 19 deck cards are on `NotYetExecutable`, 6 of them in this wave's `Cards` (*Rapturous Sermon*, *Animoceps Spores*, *Rhapsody's Embrace*, *Funeral Fungus*, *Ambush Tactics*, *Steel Rain*). Directives read on the run branch: no `open: before` directive. #14 (a): deck cards first. *Animoceps Spores* is the deck card session 11 named whose every shape is `done`, so it needs no shape task. No non-deck card has its shapes (*Disorienting Pulse* and *Tactical Withdrawal* return a chosen unit, a target, not a pick), so none joins under #14 (c). #20 (b) stays open for *Steel Rain*'s discard-cost task. Waiting triggers checked: #8 (no `deck.go` change), #15 (a scratch step is not a turn-boundary or draw hook), #58 (no exhaust in `claims:`), #59 (2) (no target slot), #61's fifth question (no modal card in `cards:`), #71 (acts on its controller only), #88's distinctness (no slot, no *Ambush Tactics*). None has fired. Duplicates searched: no new code; `OpScratch` goes through the one `stepScratch`, the pick through the one `pickAssets`, the return through the one `stepRelocate`. Two non-test files. Status `tests`; dispatching `rules-test-author`.
- 2026-10-31 rules-test-author: wrote 7 tests in `pkg/engine/cards_animus_maneuvers_test.go`: `TestCard_AnimocepsSpores_ScratchesTwoThenReturnsThePickedOne`, `_MayReturnAnOlderScratchCompany`, `_TimeoutReturnsTheFirstInFixedOrder`, `_OnlyAlliedScratchCompanyUnitsAreOffered`, `_OneCardDeckReturnsItWithoutAsking`, `_EmptyDeckDoesNothing`, `_PaysTwoTactics`. Helpers requested: none (the file adds its own `sp*` setup helpers). `go vet` is clean, `cardtown lint rules` passes, and all 7 fail as expected (card not yet executable). Cases left out: none. The first test does not assume which of the two scratched units the scripted pick (index 0) returns; it checks that exactly one went to hand and the other stayed.
- 2026-10-31 lead-architect: tests approved. The seven check outcomes against the card text and FRD 01. The top 2 deck cards become two face-down Scratch Company units, and one `DecidePick` asked of player 0 offers exactly those two. One returns to hand as its printed card, and the other stays in Reserves, Exhausted (`FR-KWD-002m`, `FR-ZON-005`). An older allied Scratch Company in the Main Theater is offered too, and picking it returns that one (`SVY-R3`). A timeout returns the first scratched unit, because Reserves come before the Main Theater in fixed order (`FR-MOD-001` §6, `SVY-R16`, `fixedOrderPositions`). An opposing Scratch Company and an allied non-Scratch unit are never offered. With a 1-card deck the sole option returns without asking the Chooser, and nobody loses. With an empty deck nothing happens and the Maneuver still goes to its owner's discard pile. It costs 2 Tactics and no Requisition, and it is offered on an empty battlefield (`SVY-R29`). The asserts read zones, pools, decisions and the outcome, not internals. Setup helpers are local to the test file, and no existing test is edited. All seven fail on the missing definition (`Enact rejected: … needs rules the kernel does not implement yet`, re-run in a review worktree). Status `ready`; dispatching `kernel-engineer` (flash) through `agy`.
- 2026-10-31 kernel-engineer: Implemented. Diff: added ability definition for "Animoceps Spores" to `animusDefs` in `pkg/engine/cards/animus.go` (two steps: `OpScratch` with amount 2, and `OpReturnToHand` on `SubjectPick` filtering for `AllegianceAllied`, `KindUnit`, and `ScratchOnly`); removed "Animoceps Spores" from `NotYetExecutable` in `pkg/engine/cards/cards.go`. Reused: existing shapes `ability.OpScratch`, `ability.OpReturnToHand`, `ability.SubjectPick`, `ability.Filter` (`AllegianceAllied`, `KindUnit`, `ScratchOnly`). Commands run: `go test ./pkg/engine -run TestCard_AnimocepsSpores` (ok, all 7 passed), `go test ./pkg/engine -run TestStarterSet_Executable` (ok), `go test ./pkg/engine -run "TestProgress_Case4_Coverage|TestEvents_Case4_EveryActionKind"` (ok), `go tool task check` (all vet, tests, and linters passed), `go tool task sim` (200 games clean, P0 starter deck executable cards increased from 33/50 to 36/50). Reviewer: check that the two steps run in sequence as declared with no target slots. Status `review`.
