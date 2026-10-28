---
id: T-0082
title: Make Eternal Watchman executable (wave B card, flash)
kind: card
tier: flash
wave: B
status: tests-review
branch: m2/T-0082-eternal-watchman
depends_on: [T-0080, T-0081]
claims: [card:Eternal Watchman]
cards: [Eternal Watchman]
requirements: []
---

## Goal
*Eternal Watchman* (the Path starter Headquarters) runs in the kernel. It is plain `ability` data in `pkg/engine/cards/headquarters.go`. Its one rules-text line is:
- "Whenever this Headquarters attacks and kills an opposing unit or structure, return chosen unit card from your discard pile to your hand."

## Inputs
- Card (`data/cards/units-and-structures-HQ.csv`, line 7): Path Headquarters, 3/20, with the line above.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 153: `EVENT_COMBAT_KILL` as attacker → `RETURN_TO_HAND` chosen unit card from discard.
- FRD 01:
  - `FR-CBT-002` Step 3, "Post-Attack & Combat-Kill Triggers" (lines 512-519): only the declared attacker's lethal combat damage counts; a retaliation kill does not.
  - `FR-TRG-004` (line 386): the trigger is placed after the death sweep, so the defender it killed is already in its owner's discard pile.
  - `FR-TGT-001` (lines 319-326): a card in a named non-battlefield zone is a legal target, and the target is chosen when the trigger is placed (`SVY-R3`).
  - `FR-TGT-003` (lines 339-340): verified again at resolution.
  - `FR-MOD-001` §1 and §6 (lines 350, 359): one legal option is chosen without asking; a timeout takes the top card of the pile.
- Shapes it uses: `HookCombatKill` (T-0080), a target slot with `Filter.Zone: ability.ZoneDiscard` (T-0081), `OpReturnToHand` with `SubjectTarget` (T-0081).
- Rulings: `SVY-R3` (target on placement), `SVY-R16` (fixed-order timeout). "Chosen" with no "up to" is exactly one, and there is no "you may", so the trigger is not optional.

### Definition (the engineer adds it to `headquartersDefs`)
- One trigger: `Hook: ability.HookCombatKill`. Not `Optional`.
- Its `Effect`:
  - `Targets`: one slot, `Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}` (exactly one: `UpTo` 0).
  - Step 1: `{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget, Target: 0}`.
- `Text` holds the printed line.
- Remove "Eternal Watchman" from `cards.NotYetExecutable`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written in a new file, `pkg/engine/cards_headquarters_watchman_test.go`, from the CSV text. `hqScene` and the `watchman` constant are in `cards_headquarters_test.go`; `inDiscard` is in the test helpers. The Watchman's printed power is 3, so a defender it kills has at most 3 health left.
- `TestCard_EternalWatchman_KillingAUnitReturnsTheChosenUnitCardToHand`: with two unit cards in the Watchman controller's discard pile, the Watchman attacks and kills an opposing unit. A `Target` decision offering both is asked before the trigger's `place` event, and the chosen card is the last card of its owner's hand. The other card stays in the pile.
- `TestCard_EternalWatchman_KillingAStructureReturnsAUnitCard`: the same, with an opposing structure as the defender.
- `TestCard_EternalWatchman_OnlyYourUnitCardsAreOffered`: a non-unit card in the Watchman controller's pile, a unit card in the opponent's pile, and the opposing unit just killed (now in its owner's pile) are not offered.
- `TestCard_EternalWatchman_ADefenderThatSurvivesReturnsNothing`: the defender survives the Attack. No decision is asked and the hand and pile are unchanged.
- `TestCard_EternalWatchman_ARetaliationKillReturnsNothing`: an opposing unit attacks the Watchman and dies to its retaliation. No decision is asked and nothing moves.
- `TestCard_EternalWatchman_NoUnitCardInYourPileReturnsNothing`: the Watchman kills a defender with no unit card in its controller's pile. No decision is asked, the trigger resolves, and nothing moves.
- `TestCard_EternalWatchman_TimeoutReturnsTheTopUnitCard`: with several unit cards in the pile, a timeout at the target decision returns the top one (`FR-MOD-001` §6), which is not the first option.
- **Existing tests:** every test in `cards_headquarters_test.go` passes unchanged, including `TestCard_EternalWatchman_DamageDoesNotScratchForAnyone`.

## Helpers Needed
- None. The tests use existing helpers only (`hqScene`, `inDiscard`, `ptPlaceHook`, `ppRefs`, `inDiscardPile`, `inHand`); the new file adds its own `wmScene` and `wmAttack` helpers.

## Allowed Files
- `pkg/engine/cards/headquarters.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run 'TestCard_EternalWatchman'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-10-27 lead-architect (wave run B, session 19): Wrote the task, third of three for the Watchman (T-0080 hook, T-0081 pile target, this card). The run branch has everything `main` has (`git log autoloop/m2-wave-b..main` is empty), and no `open: before` directive is on the board. Waiting triggers checked against `autoloop/m2-wave-b` deda79d: #6 and #7 (mode is `serial`), #8 (no `deck.go` change), #15 (no turn hook), #20 (b) (no wave C shape task), #58 (no exhaust by a cost or an effect) and #59 (2) (no repeated targets). None has fired. Directive #14 is open and followed: the Watchman is a deck card (the Path Headquarters), so it goes before Rogue Fleshshaper.
  - Every shape the Definition names exists on the run branch: `HookCombatKill`, `ZoneDiscard` target slots, `OpReturnToHand` with `SubjectTarget`, `KindUnit`, `AllegianceAllied`. No new shape, so this is a card task.
  - The sim's targeting monitor (`pkg/sim/monitor_targeting.go`) checks that a `Target` decision is followed by its item's `place` event, which a pile target does (`SVY-R3`), and flags only refs that are Ward assets. Instance and asset IDs share one counter, so a card ref can't match one.
  - Claims don't overlap any open task, since none is open. Status `tests`; dispatching `rules-test-author`.
- 2026-10-27 rules-test-author: Wrote `pkg/engine/cards_headquarters_watchman_test.go` with the seven tests named in the Tests section: `TestCard_EternalWatchman_KillingAUnitReturnsTheChosenUnitCardToHand`, `..._KillingAStructureReturnsAUnitCard`, `..._OnlyYourUnitCardsAreOffered`, `..._ADefenderThatSurvivesReturnsNothing`, `..._ARetaliationKillReturnsNothing`, `..._NoUnitCardInYourPileReturnsNothing`, `..._TimeoutReturnsTheTopUnitCard`. They cover FR-CBT-002, FR-TRG-004, FR-TGT-001, FR-TGT-003, FR-MOD-001, SVY-R3, SVY-R16. No helpers requested. `go vet` and `cardtown lint rules` are clean. Five of the seven fail now because the Watchman has no trigger; the survive and retaliation tests pass vacuously until it exists. No cases left out. Status `tests-review`.
