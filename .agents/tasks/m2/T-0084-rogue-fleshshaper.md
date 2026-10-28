---
id: T-0084
title: Make Rogue Fleshshaper executable (wave B card, flash)
kind: card
tier: flash
wave: B
status: review
branch: m2/T-0084-rogue-fleshshaper
depends_on: [T-0083]
claims: [card:Rogue Fleshshaper]
cards: [Rogue Fleshshaper]
requirements: []
---

## Goal
*Rogue Fleshshaper* (Path unit, not in a starter deck) runs in the kernel. It is plain `ability` data in `pkg/engine/cards/path.go`, the first card on the modal shape (T-0083). Its rules text is three lines:
- "When this unit deploys, choose one:"
- "- Chosen asset heals 1."
- "- Deal 1 damage to chosen non-Scratch asset."

## Inputs
- Card (`data/cards/units-and-structures-path.csv`, lines 13-15): Path unit, Specialist, Reaver; costs R2, deploys to Reserves, 2 power / 3 health.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 152: deploy → choose one: `HEAL 1` chosen asset, or `DEAL_DAMAGE 1` chosen non-Scratch asset.
- FRD 01:
  - `FR-MOD-001` §1 (line 350), §5 (line 354) and §6 (line 357): one mode is chosen; a timeout takes the first listed legal mode.
  - `FR-TGT-001` (lines 319-326, `SVY-R3`): the mode, then its target, are chosen when the trigger is placed.
  - `FR-TGT-003` (lines 339-343): the target is verified again at resolution; Ward makes an opposing asset an illegal target (`FR-KWD-002h`).
  - `FR-KWD-002d` (heal removes up to the amount of marked damage).
- Shapes it uses: `HookDeploy`, `Effect.Modes` with `ChooseN` 1 (T-0083), exactly-one target slots, `Filter.NonScratch`, `OpHeal` and `OpDealDamage` with `SubjectTarget`. All exist on the run branch.
- Rulings: `SVY-R3`, `SVY-R6` (different modes), `SVY-R16` (fixed-order timeout). "Chosen" with no "up to" is exactly one, and there is no "you may", so the trigger is not optional. "Asset" with no allegiance or kind is any asset on the battlefield of either side, the Headquarters included, and the Fleshshaper itself (no "other").
- Design note Open Question 11: both modes can always choose the Fleshshaper itself, so the no-legal-mode cases stay unreached.

### Definition (the engineer adds it to `pathDefs`)
- One trigger: `Hook: ability.HookDeploy`. Not `Optional`.
- Its `Effect` has no `Targets` or `Steps` of its own: `ChooseN: 1`, and `Modes`, in printed order:
  1. `Targets`: one slot, `Filter{}` (any asset, exactly one). Step: `{Op: ability.OpHeal, Amount: ability.Value{N: 1}, Subject: ability.SubjectTarget, Target: 0}`.
  2. `Targets`: one slot, `Filter{NonScratch: true}` (exactly one). Step: `{Op: ability.OpDealDamage, Amount: ability.Value{N: 1}, Subject: ability.SubjectTarget, Target: 0}`.
- `Text` holds the three printed lines, exactly as above (the loader claims each CSV line separately).
- Remove "Rogue Fleshshaper" from `cards.NotYetExecutable`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written in a new file, `pkg/engine/cards_path_fleshshaper_test.go`, from the CSV text. `deployPath` and `wantOptions` are in `cards_path_test.go`.
- `TestCard_RogueFleshshaper_DeployAsksForAModeThenItsTarget`: deploying it asks its controller a `DecideMode` with options `"mode 1"` and `"mode 2"`, then the chosen mode's `DecideTarget`, both before the item's `place` event.
- `TestCard_RogueFleshshaper_HealModeHealsTheChosenAsset`: with a damaged asset on each side, mode 1 removes 1 damage from the chosen one, and no other asset changes.
- `TestCard_RogueFleshshaper_HealModeOffersEveryAsset`: mode 1's target options include an allied unit, an opposing unit, a Scratch Company unit, both Headquarters and the Fleshshaper itself.
- `TestCard_RogueFleshshaper_DamageModeDealsOneToTheChosenAsset`: mode 2 deals 1 damage to the chosen opposing asset, and a 1-health target dies.
- `TestCard_RogueFleshshaper_DamageModeDoesNotOfferScratchCompany`: mode 2's options leave out Scratch Company units of both sides and include the Fleshshaper itself and both Headquarters.
- `TestCard_RogueFleshshaper_OpposingWardAssetIsNotOffered`: an opposing Ward unit is in neither mode's options; an allied Ward unit is.
- `TestCard_RogueFleshshaper_TimeoutChoosesTheHealMode`: a timeout at the mode decision chooses mode 1 (`FR-MOD-001` §6), and no damage is dealt.
- **Existing tests:** every test in `pkg/engine` passes unchanged, including `modal_test.go`.

## Helpers Needed
- None expected. If a test needs one, `rules-test-author` names it here.

## Allowed Files
- `pkg/engine/cards/path.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [x] The card tests pass: `go test ./pkg/engine -run 'TestCard_RogueFleshshaper'`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-10-27 lead-architect (wave run B, session 21): Wrote the task, the last wave B card (T-0083 built its modal shape). The run branch has everything `main` has (`git log autoloop/m2-wave-b..main` is empty), and no `open: before` directive is on the board. Waiting triggers checked against `autoloop/m2-wave-b` 357fb43: #6 and #7 (mode is `serial`), #8 (no `deck.go` change), #15 (no turn hook), #20 (b) (no wave C shape task), #58 (no exhaust by a cost or an effect) and #59 (2) (no repeated targets). None has fired. Directive #14 is open and followed: the Fleshshaper is not a deck card, but the run header keeps it (its modal shape is also Ambush Tactics', #14 (b), (c)), every shape it needs exists, and it is the only wave B card left on `NotYetExecutable`.
  - Every shape the Definition names exists on the run branch: `HookDeploy`, `Effect.Modes`/`ChooseN`, `Filter.NonScratch`, `OpHeal`, `OpDealDamage`, `SubjectTarget`. No new shape, so this is a card task.
  - `ChooseN` is 1, so T-0083's review note (`chooseModes` sorts the chooser's own picks slice) changes nothing here; it stays for Ambush Tactics.
  - Claims don't overlap any open task, since none is open. Status `tests`; dispatching `rules-test-author`.
- 2026-10-27 rules-test-author: Wrote `pkg/engine/cards_path_fleshshaper_test.go`, the seven tests named under Tests (`TestCard_RogueFleshshaper_DeployAsksForAModeThenItsTarget`, `_HealModeHealsTheChosenAsset`, `_HealModeOffersEveryAsset`, `_DamageModeDealsOneToTheChosenAsset`, `_DamageModeDoesNotOfferScratchCompany`, `_OpposingWardAssetIsNotOffered`, `_TimeoutChoosesTheHealMode`). `go vet ./pkg/engine` is clean, `cardtown lint rules` passes, and all seven fail now because the card is not yet executable.
  - Helpers requested: none (used `deployPath`, `scriptModal`, `wantOptions`; the file adds its own `fleshBoard` scene builder).
  - Left out: nothing. Not tested because FRD 01 does not need it: the no-legal-mode cases (Open Question 11).
- 2026-10-27 lead-architect (wave run B, session 21, D4): **Tests approved, `ready`.** I checked the seven tests in 7321808 against the CSV text, `FR-MOD-001` §1, §5 and §6, `FR-TGT-001` (`SVY-R3`), `FR-TGT-003` and `FR-KWD-002h`. The mode is asked before its target and both before the trigger's `place` event; mode 1 heals 1 from the chosen asset only; its options are every asset of either side, Scratch Company, both Headquarters and the Fleshshaper itself; mode 2 deals 1 to the chosen asset and kills a 1-health one; its options leave out Scratch Company of both sides; an opposing Ward unit is in neither mode's options and an allied one is in both; a timed-out mode choice is mode 1. Each follows from the cited text. `wantOptions` compares sorted IDs, so the option order is not pinned. Re-run on the task branch: all seven fail, because the card has no definition yet. No test asserts on implementation details. Next: dispatch `kernel-engineer` (tier `flash`) headless.
- 2026-10-28 kernel-engineer (flash): Implemented Rogue Fleshshaper.
  - Diff: added "Rogue Fleshshaper" definition to `pathDefs` in `pkg/engine/cards/path.go` with one non-optional `HookDeploy` trigger having modal `Effect` (`ChooseN: 1`, two modes in printed order: mode 1 with target slot `Filter{}` and `OpHeal` 1 to `SubjectTarget` target 0; mode 2 with target slot `Filter{NonScratch: true}` and `OpDealDamage` 1 to `SubjectTarget` target 0); removed "Rogue Fleshshaper" from `NotYetExecutable` in `pkg/engine/cards/cards.go`.
  - Reused: `ability.HookDeploy`, `ability.Effect` (`Modes`, `ChooseN`), `ability.Target`, `ability.Filter` (`NonScratch`), `ability.OpHeal`, `ability.OpDealDamage`, `ability.SubjectTarget`.
  - Commands: `go test ./pkg/engine -run TestCard_RogueFleshshaper` (all 7 tests passed); `go test ./pkg/engine -run TestStarterSet_Executable` (passed); `go tool task check` passed (gofmt, `go vet`, all unit & scenario tests, `cardtown lint --all`); `go tool task sim` passed clean (200 random games, all invariants held).
  - Reviewer: each mode in `Modes` defines its own single target slot, so each mode's step addresses `Target: 0`, and the modal trigger resolution joins and remaps the chosen mode's targets.
