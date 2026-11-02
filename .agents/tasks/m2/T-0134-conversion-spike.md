---
id: T-0134
title: Make Conversion Spike executable (damage a chosen opposing unit, heal a chosen allied unit or structure, wave C card, flash)
kind: card
tier: flash
wave: C
status: tests-review
branch: m2/autoloop-wave-c/T-0134-conversion-spike
depends_on: [T-0019, T-0116, T-0117, T-0121]
claims: [card:Conversion Spike]
cards: [Conversion Spike]
requirements: [FR-MNV-001, FR-TGT-001, FR-TGT-003, FR-KWD-002d, FR-KWD-002h, FR-REP-004]
---

## Goal
*Conversion Spike* (Animus), a non-deck card kept in this run (#14 (c)), runs in the kernel. It is plain `ability` data: a Maneuver `Effect` with two exactly-one target slots, an `OpDealDamage` 1 step on the first and an `OpHeal` 1 step on the second. No new shape.

## Inputs
- Card: `data/cards/maneuvers-animus.csv` line 8: *Conversion Spike*, Animus Maneuver, cost 0 Requisition / 2 Tactics: "Deal 1 damage to chosen opposing unit, and chosen allied unit or structure heals 1."
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 128 (`DEAL_DAMAGE 1` chosen opposing unit; `HEAL 1` chosen allied unit or structure; "Two independent targets").
- FRD 01:
  - `FR-MNV-001` (lines 332-334): enacted from Hand, costs paid on declaration, the effect runs when it resolves, then the card goes to its owner's Discard pile. It can be enacted only if each exactly-one target slot has a legal entity on declaration (`SVY-R29`). Both of this card's slots are exactly-one, so it needs a legal opposing unit **and** a legal allied unit or structure.
  - `FR-TGT-001` (lines 345-359): every criterion must hold. "Chosen opposing unit" is a unit an opponent controls, in either battlefield zone, face-up or a face-down Scratch Company (`FR-ZON-005`), never a structure, a Headquarters or an allied unit. "Chosen allied unit or structure" is a unit or a structure its controller controls, in either battlefield zone, never a Headquarters (the classification line lists *Structure* and *Headquarters* apart) and never an opposing asset. Targets are chosen on placement (`SVY-R3`).
  - `FR-KWD-002d` (line 566): heal removes up to N accumulated damage; an undamaged target is still a legal choice and stays at 0 damage.
  - `FR-KWD-002h` (line 570): Ward stops only opponents. An opposing Ward unit can't be chosen for the damage; an allied Ward unit can be chosen for the heal.
  - `FR-REP-004` (line 484 on): the 1 damage is non-combat damage from the Maneuver, and Armor absorbs it.
  - `FR-TGT-003` (line 367 on): a target illegal at resolution is skipped and the rest of the effect still runs.
- Shapes it uses, all `done`: `Def.Effect`, `ActEnact` and `StackManeuver` (T-0116), the `SVY-R29` enact legality (T-0117, including `TestFR_MNV_001_SecondRequiredSlotAlsoNeeded`), `Effect.Targets` with an allegiance and kind `Filter` (`Kinds: ability.KindUnit | ability.KindStructure`, as in `pkg/engine/cards/path.go:277`), `OpDealDamage` and `OpHeal` on `SubjectTarget` (T-0019, `stepHeal` in `effects.go`), Ward in `legalTarget` (`targeting.go`).
- Rulings: none beyond the FRD text above. The two slots have opposite allegiances, so no entity can fill both: neither distinctness between slots (design note Open Question 16) nor repeated targets (#59 (2)) arises.

### Definition (the engineer adds it)
- *Conversion Spike*, in `animusDefs` (`pkg/engine/cards/animus.go`): `Text` holds its line. `Effect`:
  - `Targets`: two, in printed order: `{Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}}`, then `{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit | ability.KindStructure}}`.
  - `Steps`: `{Op: ability.OpDealDamage, Amount: ability.Value{N: 1}, Subject: ability.SubjectTarget, Target: 0}`, then `{Op: ability.OpHeal, Amount: ability.Value{N: 1}, Subject: ability.SubjectTarget, Target: 1}`.
- No `Triggers` or `Modes`. Remove the name from `cards.NotYetExecutable`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
From the CSV text, loading the real card (`s.card`). Player 0 enacts; player 1 is the opponent. Give player 0 enough Tactics to pay (`enactFor` in `cards_animus_maneuvers_test.go` does). Targets are answered with a scripted decision. Where a test checks which entities a slot offers, put at least two legal entities in that slot so a decision is asked (T-0123 round 2). All appended to `pkg/engine/cards_animus_maneuvers_test.go`:
- `TestCard_ConversionSpike_DamagesChosenOpposingUnitAndHealsChosenAlly`: a chosen opposing unit takes 1 damage and a chosen allied unit with 2 damage is left with 1 once the Maneuver resolves; the Maneuver ends in player 0's discard pile.
- `TestCard_ConversionSpike_HealsAChosenAlliedStructure`: an allied structure with damage can be chosen for the heal and loses 1 damage.
- `TestCard_ConversionSpike_OpposingSlotOffersOnlyOpposingUnits`: with opposing units in the Main Theater and in Reserves (one a face-down Scratch Company), an opposing structure, allied units and both Headquarters, the first target decision offers exactly the opposing units, in both zones.
- `TestCard_ConversionSpike_AlliedSlotOffersOnlyAlliedUnitsAndStructures`: with allied units in both zones, an allied structure, the allied Headquarters, opposing units and an opposing structure, the second target decision offers exactly the allied units and the allied structure, never the Headquarters.
- `TestCard_ConversionSpike_OpposingWardUnitCannotBeChosen`: an opposing Ward unit is not offered for the damage while two plain opposing units are.
- `TestCard_ConversionSpike_AlliedWardUnitCanBeHealed`: an allied Ward unit is offered for the heal and, when chosen, loses 1 damage (`FR-KWD-002h`: Ward stops only opponents).
- `TestCard_ConversionSpike_HealingAnUndamagedAllyDoesNothing`: an undamaged allied unit can be chosen; it stays at 0 damage, and the damage step still hits the opposing unit (`FR-KWD-002d`).
- `TestCard_ConversionSpike_ArmorAbsorbsTheDamage`: an opposing unit with 1 Armor loses its Armor and takes no damage (`FR-REP-004`).
- `TestCard_ConversionSpike_KillsAOneHealthUnit`: a chosen opposing unit with 1 current health dies once the Maneuver resolves.
- `TestCard_ConversionSpike_NotEnactableWithoutBothTargets`: with an allied unit but no opposing unit (or only an opposing Ward unit), enacting it is not a legal action; with an opposing unit but no allied unit or structure (only the allied Headquarters), it is not a legal action either; the card stays in hand (`SVY-R29`). With one plain opposing unit and one allied unit, it is.
- `TestCard_ConversionSpike_PaysTwoTactics`: enacting it spends 2 Tactics and no Requisition.
- Not a card test: a target that turns illegal between placement and resolution (`FR-TGT-003`), for the reason T-0123 gives (nothing in a scene runs between a Maneuver's placement and its resolution; the rule is pinned on synthetic cards by `TestFR_TGT_003_UpToResolvesForTheSurvivingChosenEntity` and T-0053).
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestProgress_Case4_Coverage` and `TestEvents_Case4_EveryActionKind`.

## Helpers Needed
- None expected. The tests use existing helpers (`newScene`, `put`, `putScratch`, `give`, `card`, `do`, `script`, `asked`, `wantDamage`, `alive`), those in `maneuver_test.go` (`enactAction`, `offersEnact`, `lastDiscard`), `enactFor` and `damageEventsOn` in `cards_animus_maneuvers_test.go`, and `targetOffers` in `cards_path_maneuvers_test.go`; a test file may add its own setup helpers.

## Allowed Files
- `pkg/engine/cards/animus.go` (*Conversion Spike*'s definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_ConversionSpike`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-11-02 lead-architect: wave C run, session 22. No `main` commits to sync (`git log autoloop/m2-wave-c..main` is empty). Not `WAVE-COMPLETE`: 15 deck cards are on `NotYetExecutable`, two of them in this wave's `Cards` (*Ambush Tactics*, *Steel Rain*), plus the kept *Conversion Spike*, *Ensnaring Mist*, *Chainbreaker's Tenacity*, *Fortify* and *Pincer Tactics*. Directives read on the run branch: no `open: before` directive. Both remaining wave C deck cards still wait on a ruling asked before their next task (*Steel Rain*: #59 (2); *Ambush Tactics*: #61's fifth question and #88's distinctness), so, as in rows 4, 5 and 21, ruling-free work goes first: under #14 (c), a kept non-deck card whose shapes all exist. *Conversion Spike* is the next kept card in the survey's wave C order (line 128). *Fortify* and *Chainbreaker's Tenacity* share a different shape (Armor on a chosen allied unit) and get their own task; *Ensnaring Mist* and *Pincer Tactics* stay out (two slots that could name the same unit, Open Question 16). Waiting triggers checked: #6, #7 (mode is `serial`), #8 (no `deck.go` change), #15 (no turn hook), #59 (2) (no repeated targets), #61's fifth question (no modal card in `cards:`), #71 (no effect on each player), #88's distinctness (the two slots have opposite allegiances). None has fired. Duplicates searched: no new code; the "unit or structure" kind is the existing `Kinds` bit set (`path.go:277`), and damage and heal on a slot go through `stepDealDamage` and `stepHeal`. Two non-test files. Status `tests`; dispatching `rules-test-author`.
- 2026-11-02 rules-test-author: wrote 11 tests, all appended to `pkg/engine/cards_animus_maneuvers_test.go`, one per bullet under Tests: `TestCard_ConversionSpike_DamagesChosenOpposingUnitAndHealsChosenAlly`, `_HealsAChosenAlliedStructure`, `_OpposingSlotOffersOnlyOpposingUnits`, `_AlliedSlotOffersOnlyAlliedUnitsAndStructures`, `_OpposingWardUnitCannotBeChosen`, `_AlliedWardUnitCanBeHealed`, `_HealingAnUndamagedAllyDoesNothing`, `_ArmorAbsorbsTheDamage`, `_KillsAOneHealthUnit`, `_NotEnactableWithoutBothTargets`, `_PaysTwoTactics`. Helpers requested: none. The file adds its own setup helpers `csChoose`, `csTargetOffers` and `csWantOffers` (two-slot scripted chooser and per-decision offer lookup). `go vet` and `cardtown lint rules` are clean; the tests fail now because the card is not yet executable. Left out, as the task says: `FR-TGT-003` illegal-at-resolution (no scene hook between placement and resolution).
