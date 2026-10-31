---
id: T-0123
title: Make Devilhorn Sanction executable (kill a chosen opposing unit, wave C card, flash)
kind: card
tier: flash
wave: C
status: in-progress
branch: m2/T-0123-devilhorn-sanction
depends_on: [T-0116, T-0117, T-0121]
claims: [card:Devilhorn Sanction]
cards: [Devilhorn Sanction]
requirements: [FR-MNV-001, FR-TGT-001, FR-TGT-003, FR-KWD-002h, FR-REP-001]
---

## Goal
*Devilhorn Sanction* (Path starter deck, 2 copies) runs in the kernel. It is plain `ability` data in `pkg/engine/cards/path.go`: a Maneuver `Effect` with one exactly-one target slot, a chosen opposing unit, and one `OpKill` step on that slot. No new shape.

## Inputs
- Card (`data/cards/maneuvers-path.csv` line 8): *Devilhorn Sanction*, Path Maneuver, cost 0 Requisition / 3 Tactics, one rules-text line:
  - "Kill chosen opposing unit."
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 172: `KILL` chosen opposing unit.
- FRD 01:
  - `FR-MNV-001` (lines 332-334): enacted from Hand, costs paid on declaration, the effect runs when it resolves, then the card goes to its owner's Discard pile. It can be enacted only if its exactly-one target slot has a legal entity on declaration (`SVY-R29`); otherwise it stays in hand. A target that is illegal at resolution is skipped, and the card still goes to the Discard pile.
  - `FR-TGT-001` (line 346 on): "chosen opposing unit" is a target: any unit the opponent controls, in either battlefield zone (no zone is named), face-up or a face-down Scratch Company (a face-down Scratch Company is a unit, `FR-ZON-005`). Never a structure, a Headquarters or an allied unit.
  - `FR-KWD-002h` (line 570): an opposing unit with Ward can't be chosen. `FR-TGT-003` (line 367): Ward is checked on placement and again at resolution.
  - `FR-REP-001` (lines 475-480): a destruction replacement applies to a kill step's death (`SVY-R21`).
- Shapes it uses, all `done`: `Def.Effect`, `ActEnact` and `StackManeuver` (T-0116), the `SVY-R29` enact legality (T-0117), `Effect.Targets` with an allegiance and kind `Filter`, `OpKill` on `SubjectTarget` (`stepKill`, `effects.go`; one `kill`, `sweep.go`, T-0108), Ward in `legalTarget` (`targeting.go`), the kill-step replacement (T-0097, T-0101).
- Rulings: none beyond the FRD text above. A kill ignores Armor and remaining health: the unit is destroyed at once, not dealt damage.

### Definition (the engineer adds it to `pathDefs`)
- *Devilhorn Sanction*: `Text` holds its line. `Effect`:
  - `Targets`: one, `{Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}}`.
  - `Steps`: one, `{Op: ability.OpKill, Subject: ability.SubjectTarget, Target: 0}`.
- No `Triggers` or `Modes`.
- Remove the name from `cards.NotYetExecutable`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written in a new file, `pkg/engine/cards_path_maneuvers_test.go`, from the CSV text, loading the real card (`s.card`). Player 0 enacts; player 1 is the opponent. Give player 0 enough Tactics to pay (`enactFor` in `cards_animus_maneuvers_test.go` does). The target is answered with a scripted decision.
- `TestCard_DevilhornSanction_KillsChosenOpposingUnit`: an opposing unit in the Main Theater, chosen, goes to its owner's discard pile once the Maneuver resolves, and the Maneuver ends in its owner's discard pile.
- `TestCard_DevilhornSanction_KillsInReservesAndFaceDown`: an opposing unit in Reserves can be chosen and is killed; so can an opposing face-down Scratch Company unit, which goes to its owner's discard pile face-up as its printed card.
- `TestCard_DevilhornSanction_IgnoresHealthAndArmor`: an opposing unit with high health and Armor is killed outright; it takes no damage event.
- `TestCard_DevilhornSanction_OnlyOpposingUnitsAreOffered`: with an allied unit, an opposing structure, both Headquarters and two opposing units on the battlefield, the target decision offers exactly the two opposing units.
- `TestCard_DevilhornSanction_WardUnitCannotBeChosen`: an opposing Ward unit is not offered while another opposing unit is.
- `TestCard_DevilhornSanction_NotEnactableWithoutLegalTarget`: with no opposing unit, or with only an opposing Ward unit, or with only allied units and opposing structures, enacting it is not a legal action and the card stays in hand (`SVY-R29`). With one plain opposing unit, it is.
- `TestCard_DevilhornSanction_ShiftingImpossibilityReplacesTheKill`: player 1 controls *Shifting Impossibility* and a face-down Scratch Company unit; the Sanction kills that Scratch Company. The replacement applies: the card goes to its owner's hand, not the discard pile, and *Shifting Impossibility* takes 1 damage (`FR-REP-001`, `SVY-R21`).
- `TestCard_DevilhornSanction_PaysThreeTactics`: enacting it spends 3 Tactics and no Requisition.
- Not a card test: a target that turns illegal between placement and resolution (`FR-TGT-003`). Nothing in a scene runs between a Maneuver's placement and its resolution (`midStack` needs an optional trigger), and the general rule is pinned on synthetic cards (`TestFR_TGT_003_TargetThatGainsWardBeforeResolutionIsIllegal`, T-0053).
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestProgress_Case4_Coverage` and `TestEvents_Case4_EveryActionKind`.

## Helpers Needed
- None expected. The tests use existing helpers (`newScene`, `put`, `putScratch`, `give`, `card`, `do`, `script`, `asked`, `alive`, `wantDamage`), those in `maneuver_test.go` (`enactAction`, `offersEnact`, `lastDiscard`) and `enactFor`, `damageEventsOn` in `cards_animus_maneuvers_test.go`; the new file may add its own setup helpers.

## Allowed Files
- `pkg/engine/cards/path.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_DevilhornSanction`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-10-30 lead-architect: wave C run, session 9. No `main` commits to sync (`git log autoloop/m2-wave-c..main` is empty). Not `WAVE-COMPLETE`: 21 deck cards are on `NotYetExecutable`, 8 of them in this wave's `Cards` (*Rapturous Sermon*, *Animoceps Spores*, *Rhapsody's Embrace*, *Funeral Fungus*, *Ambush Tactics*, *Steel Rain*, *Devilhorn Sanction*, *Improvised Orders*). Directives read on the run branch: no `open: before` directive. #14 (a): deck cards first; *Devilhorn Sanction* is a deck card whose every shape is `done`, so it needs no shape task. No non-deck card has its shape (*Disorienting Pulse* returns to hand, *Fortify* adds Armor), so none joins under #14 (c). #20 (b) stays open for *Steel Rain*'s discard-cost task. Waiting triggers checked: #8 (no `deck.go` change), #15 (no turn hook), #58 (no exhaust in `claims:`), #59 (2) (one exactly-one slot, no repeated targets), #61's fifth question (no modal card in `cards:`), #71 (acts on a unit, not on each player), #88's distinctness (one slot, no *Ambush Tactics*). None has fired. Duplicates searched: no new code; the definition is `OpKill` on `SubjectTarget`, which `stepKill` already serves through `stepAssets` and the one `kill` (T-0108), as T-0117's synthetic "Kill chosen unit" Maneuver does. Two non-test files. Status `tests`; dispatching `rules-test-author`.
- 2026-10-30 rules-test-author: tests written in `pkg/engine/cards_path_maneuvers_test.go`: `TestCard_DevilhornSanction_` KillsChosenOpposingUnit, KillsInReservesAndFaceDown, IgnoresHealthAndArmor, OnlyOpposingUnitsAreOffered, WardUnitCannotBeChosen, NotEnactableWithoutLegalTarget, ShiftingImpossibilityReplacesTheKill, PaysThreeTactics. All fail now with "needs rules the kernel does not implement yet" (one asserts Enact offered). The file defines its own helpers `chooseAsset` and `targetOffers`. Helpers requested: none. Left out: the FR-TGT-003 resolution-time case, as the Tests section says. The face-down case checks the discard pile only (the test cannot see the face-up state). Status `tests-review`.
- 2026-10-30 lead-architect: tests approved. The eight check their outcomes against the card text and FRD 01: the chosen opposing unit dies into its owner's discard pile and the unchosen one lives; an opposing unit in Reserves and a face-down Scratch Company can be chosen and are killed (`FR-TGT-001`; the face-down card reaching the pile as its own instance is enough, since a pile holds printed cards); a kill ignores 30 health and 9 Armor and emits no damage event; the target decision offers exactly the opposing units, never an allied unit, a structure or a Headquarters; a Ward unit isn't offered (`FR-KWD-002h`); not enactable with no opposing unit, only a Ward one, or only allies and structures, and enactable with one plain unit (`SVY-R29`); *Shifting Impossibility* replaces the kill, sending the Scratch Company to hand and taking 1 damage (`FR-REP-001`, `SVY-R21`); 3 Tactics and no Requisition. The asserts read zones, pools, offered options and the log, not internals. No existing test edited. All eight fail on the missing definition (seven on "needs rules the kernel does not implement yet", one on Enact not offered). Status `ready`; dispatching `kernel-engineer` (flash) through `agy`.
