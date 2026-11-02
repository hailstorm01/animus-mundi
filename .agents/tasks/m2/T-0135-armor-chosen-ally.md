---
id: T-0135
title: Make Chainbreaker's Tenacity and Fortify executable (a chosen allied unit gains armor, Tenacity also +1 power until end of turn, wave C card, flash)
kind: card
tier: flash
wave: C
status: tests-review
branch: m2/autoloop-wave-c/T-0135-armor-chosen-ally
depends_on: [T-0019, T-0069, T-0116, T-0117, T-0121]
claims: [card:Chainbreaker's Tenacity, card:Fortify]
cards: [Chainbreaker's Tenacity, Fortify]
requirements: [FR-MNV-001, FR-TGT-001, FR-KWD-002h, FR-KWD-002k, FR-LAY-001, FR-LAY-002, FR-LAY-003, FR-REP-004, FR-ZON-005]
---

## Goal
*Chainbreaker's Tenacity* and *Fortify* (both Path), non-deck cards kept in this run (#14 (c)), run in the kernel. Each is plain `ability` data: a Maneuver `Effect` with one exactly-one allied-unit target slot, an `OpAddArmor` step on it, and for *Chainbreaker's Tenacity* an `OpModifyStat` +1 step before it. No new shape.

## Inputs
- Cards:
  - `data/cards/maneuvers-path.csv` line 2: *Chainbreaker's Tenacity*, Path Maneuver, cost 0 Requisition / 1 Tactics: "Chosen allied unit gets +1 power until end of turn and gains 2 armor."
  - `data/cards/maneuvers-path.csv` line 3: *Fortify*, Path Maneuver, cost 0 Requisition / 1 Tactics: "Chosen allied unit gains 3 armor."
- Survey rows: `docs/design/m2_starter_set_ability_survey.md` line 167 (*Chainbreaker's Tenacity*: chosen allied unit, `MODIFY_STAT` +1 until end of turn, `ADD_ARMOR 2`) and line 168 (*Fortify*: `ADD_ARMOR 3` chosen allied unit).
- FRD 01:
  - `FR-MNV-001` (line 332 on): enacted from Hand, costs paid on declaration, the effect runs when it resolves, then the card goes to its owner's Discard pile. It can be enacted only if each exactly-one target slot has a legal entity on declaration (`SVY-R29`).
  - `FR-TGT-001` (line 346 on): "chosen allied unit" is a unit its controller controls in either battlefield zone, face-up or a face-down Scratch Company (a face-down Scratch Company is a unit, `FR-ZON-005`, line 227), never a structure, a Headquarters or an opposing unit.
  - `FR-KWD-002h` (line 570): Ward stops only opponents, so an allied Ward unit can be chosen.
  - `FR-KWD-002k` (line 573) and `FR-LAY-002` (line 461): armor granted by an effect adds to the asset's Armor pool, stacks with armor it already has, never expires at end of turn, and absorbs damage 1-for-1 before Health.
  - `FR-LAY-001` (line 451) and `FR-LAY-003` (line 463): "+1 power until end of turn" is a Layer 5 modifier that expires in Sub-Phase 4c of this turn's End Phase.
- Shapes it uses, all `done`: `Def.Effect`, `ActEnact` and `StackManeuver` (T-0116), the `SVY-R29` enact legality (T-0117), the progress policy enacting Maneuvers (T-0121), `Effect.Targets` with an allegiance and kind `Filter`, `OpAddArmor` on `stepAssets` (T-0019, `stepAddArmor` in `effects.go`), `OpModifyStat` on `stepAssets` (T-0069, `stepModifyStat` in `effects.go`, `TempPower` cleared in `turn.go`), Ward in `legalTarget` (`targeting.go`).
- Rulings: none beyond the FRD text above. Each card has one target slot, so neither distinctness between slots (design note Open Question 16) nor repeated targets (#59 (2)) arises.

### Definitions (the engineer adds them, in `pathDefs`, `pkg/engine/cards/path.go`)
- *Chainbreaker's Tenacity*: `Text` holds its line. `Effect`:
  - `Targets`: one, `{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}}`.
  - `Steps`: `{Op: ability.OpModifyStat, Amount: ability.Value{N: 1}, Subject: ability.SubjectTarget, Target: 0}`, then `{Op: ability.OpAddArmor, Amount: ability.Value{N: 2}, Subject: ability.SubjectTarget, Target: 0}`.
- *Fortify*: `Text` holds its line. `Effect`:
  - `Targets`: the same single slot.
  - `Steps`: one, `{Op: ability.OpAddArmor, Amount: ability.Value{N: 3}, Subject: ability.SubjectTarget, Target: 0}`.
- No `Triggers` or `Modes`. Remove both names from `cards.NotYetExecutable`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
From the CSV text, loading the real cards (`s.card`), appended to `pkg/engine/cards_path_maneuvers_test.go`. Player 0 enacts; player 1 is the opponent. Give player 0 enough Tactics to pay. Targets are answered with a scripted decision. A target decision is asked only when two or more entities are legal, so every test that checks what is offered, or that a particular unit was chosen, puts at least two legal allied units on the battlefield (T-0123's round 2, T-0134's round 1).

*Chainbreaker's Tenacity*:
- `TestCard_ChainbreakersTenacity_ChosenAllyGetsPlusOnePowerAndTwoArmor`: with two allied units, the chosen one has its printed power +1 and 2 more armor once the Maneuver resolves; the other is unchanged; the Maneuver ends in player 0's discard pile.
- `TestCard_ChainbreakersTenacity_PowerExpiresAtEndOfTurnArmorStays`: after the turn ends (Sub-Phase 4c), the chosen unit is back at its printed power and still has the 2 armor (`FR-LAY-003`, `FR-KWD-002k`).
- `TestCard_ChainbreakersTenacity_OnlyAlliedUnitsAreOffered`: with two allied units (one in Reserves, one in the Main Theater), an allied structure, an opposing unit and both Headquarters, the target decision offers exactly the two allied units.
- `TestCard_ChainbreakersTenacity_AlliedWardUnitCanBeChosen`: an allied unit with Ward is offered beside a plain allied unit and, when chosen, gets the +1 power and 2 armor (`FR-KWD-002h`).
- `TestCard_ChainbreakersTenacity_NotEnactableWithoutAlliedUnit`: with only an allied structure and opposing units, enacting it is not a legal action and the card stays in hand (`SVY-R29`). With one allied unit, it is.
- `TestCard_ChainbreakersTenacity_FaceDownScratchCompanyCanBeChosen`: an allied face-down Scratch Company in Reserves can be chosen, and has power 2 and 2 armor once it resolves (`FR-ZON-005`).
- `TestCard_ChainbreakersTenacity_PaysOneTactics`: enacting it spends 1 Tactics and no Requisition.

*Fortify*:
- `TestCard_Fortify_ChosenAllyGainsThreeArmor`: with two allied units, the chosen one has 3 more armor once the Maneuver resolves and its power is unchanged; the other is unchanged; the Maneuver ends in player 0's discard pile.
- `TestCard_Fortify_ArmorStacksAndOutlastsTheTurn`: a chosen allied unit that already has armor (1 or more, from a grant or a printed Armor N) ends with the sum, and still has it after the turn ends (`FR-LAY-002`, `FR-KWD-002k`).
- `TestCard_Fortify_ArmorAbsorbsLaterDamage`: after *Fortify* resolves on a unit, 2 damage dealt to it (by any existing source, such as combat or a damage Maneuver) removes 2 armor and leaves its Health untouched (`FR-REP-004` Stage 4).
- `TestCard_Fortify_OnlyAlliedUnitsAreOffered`: the same setup as Tenacity's; the decision offers exactly the two allied units.
- `TestCard_Fortify_NotEnactableWithoutAlliedUnit`: as Tenacity's (`SVY-R29`).
- `TestCard_Fortify_PaysOneTactics`: enacting it spends 1 Tactics and no Requisition.
- Not a card test: a target that turns illegal between placement and resolution (`FR-TGT-003`), for the reason T-0123 gives (nothing in a scene runs between a Maneuver's placement and its resolution; the rule is pinned on synthetic cards).
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with both names removed, `TestProgress_Case4_Coverage` and `TestEvents_Case4_EveryActionKind`.

## Helpers Needed
- None expected. The tests use existing helpers (`newScene`, `put`, `putScratch`, `give`, `card`, `do`, `script`, `asked`), those in `maneuver_test.go` (`enactAction`, `offersEnact`, `lastDiscard`), `enactFor` in `cards_animus_maneuvers_test.go`, and `chooseAsset`, `targetOffers` in `cards_path_maneuvers_test.go`; a test file may add its own setup helpers.

## Allowed Files
- `pkg/engine/cards/path.go` (both definitions)
- `pkg/engine/cards/cards.go` (remove both names from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run "TestCard_ChainbreakersTenacity|TestCard_Fortify"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] The cards load with no unsupported lines (`TestStarterSet_Executable` passes with both names removed)

## Log
- 2026-11-02 lead-architect: wave C run, session 24. No `main` commits to sync (`git log autoloop/m2-wave-c..main` is empty). Not `WAVE-COMPLETE`: 15 deck cards are on `NotYetExecutable`, two of them in this wave's `Cards` (*Ambush Tactics*, *Steel Rain*), plus the kept *Ensnaring Mist*, *Chainbreaker's Tenacity*, *Fortify*, *Pincer Tactics*. Directives read on the run branch: no `open: before` directive. Both remaining wave C deck cards still wait on a ruling asked before their next task (#59 (2) for *Steel Rain*; #61's fifth question and #88's distinctness for *Ambush Tactics*), so, as in rows 21 to 23, ruling-free work goes first: under #14 (c), kept cards whose shapes all exist. *Chainbreaker's Tenacity* is the next kept card in the survey's wave C order, and *Fortify* uses a subset of its shapes (one allied-unit slot, `OpAddArmor`), so the two share one card task (D2.4). *Ensnaring Mist* and *Pincer Tactics* stay out (two slots that could name the same unit, Open Question 16). Waiting triggers checked: #6, #7 (mode is `serial`), #8 (no `deck.go` change), #15 (no turn hook; a power modifier expiring in 4c can't make an asset lethal), #59 (2) (no repeated targets), #61's fifth question (no modal card), #71 (no each-player effect), #88's distinctness (one slot per card). None has fired. Duplicates searched: no new code; the slot, `stepAddArmor` and `stepModifyStat` already serve these steps (`OpAddArmor` on a target slot runs today in Ironclad Colossus; `OpModifyStat` reads its subject through the same `stepAssets`, Rapturous Sermon with `SubjectAll`). Two non-test files. Status `tests`; dispatching `rules-test-author`.
- 2026-11-02 rules-test-author: wrote 13 card tests appended to `pkg/engine/cards_path_maneuvers_test.go` (plus file-local helpers `armorPair`, `armorOffers`, `armorNotEnactable`): TestCard_ChainbreakersTenacity_{ChosenAllyGetsPlusOnePowerAndTwoArmor, PowerExpiresAtEndOfTurnArmorStays, OnlyAlliedUnitsAreOffered, AlliedWardUnitCanBeChosen, NotEnactableWithoutAlliedUnit, FaceDownScratchCompanyCanBeChosen, PaysOneTactics} and TestCard_Fortify_{ChosenAllyGainsThreeArmor, ArmorStacksAndOutlastsTheTurn, ArmorAbsorbsLaterDamage, OnlyAlliedUnitsAreOffered, NotEnactableWithoutAlliedUnit, PaysOneTactics}. Helpers requested: none. The damage test uses a synthetic allied-damage Maneuver (existing `maneuver`, `tgt`, `dmgTo`). Left out: FR-TGT-003 illegal-at-resolution (as the task says). All fail now ("needs rules the kernel does not implement yet"); vet and `cardtown lint rules` pass. Status `tests-review`.
