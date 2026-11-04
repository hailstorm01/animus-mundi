---
id: T-0162
title: A filter can ask for any of several subtypes, and for a Maneuver card (FR-TGT-001, wave D shape, sonnet)
kind: shape
tier: sonnet
wave: D
status: tests
branch: m2/autoloop-wave-d/T-0162-filter-subtypes-and-maneuver
depends_on: [T-0158]
claims: [filter:subtypes, kind:maneuver]
cards: []
requirements: [FR-TGT-001]
---

## Goal
A `Filter` can require any of a list of subtypes ("Fungus or Plant card"), and `Kinds` can name a Maneuver card ("maneuver card"). *Sinuous Growth* and *Boneyard Scavenger* need these two criteria for their pile targets. No card uses them yet.

## Inputs
- FRD 01: `FR-TGT-001` (line 352 on), Entity Classification (*"Card in a specified non-battlefield zone"*) and Readiness, Subtypes & Dynamic Stats (*"subtype (Infantry, Fungus, non-Spirit)"*). `FR-ZON-004` for a card's printed characteristics in a pile.
- Card text, for the wording only (`data/cards/*.csv`): *Sinuous Growth*, *"Return chosen Fungus or Plant card from your discard pile to your hand."*; *Boneyard Scavenger*, *"Return chosen maneuver card from your discard pile to your hand."* Both come later, in card tasks.
- Survey rows: wave D (`docs/design/m2_starter_set_ability_survey.md`). No card in this task.
- Rulings that apply: `SVY-R3`, for the difference between a target and a pick. None on subtypes.
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Filter`: *"a conjunction of criteria: ... subtype in/out, non-Scratch ..."*).
- Kernel today:
  - `Filter` and `Kinds` are in `ability/ability.go`. `kindOf` and `matchesPrinted` are in `targeting.go`, and every target, pick and pile check reaches `matchesPrinted`, for assets through `matchesFilter` and for pile cards through `matchesCard`.
  - `kindOf` is called only by `matchesPrinted`. It returns 0 for a Maneuver, so a pile filter with any `Kinds` set refuses Maneuvers, and one with `Kinds` empty accepts them. Both stay true with the new bit.
  - `Card.HasSubtype` already reads a subtype, case-insensitively, and `IsScratchCompany` calls it.
  - The switch on `f.Scratch` at the end of `matchesPrinted` returns from inside its cases, so a check placed after it would be skipped.

### Shape (lead decisions, binding)
- **`Filter.Subtypes []string`** (`ability/ability.go`): the entity must carry at least one of these printed subtypes. Empty means no subtype criterion. Its comment says "any of" and gives "Fungus or Plant card" as the example.
- **`KindManeuver`** (`ability/ability.go`): a new bit after `KindHeadquarters`, so the existing bits keep their values. Its comment says only a pile filter can meet it, because no asset is a Maneuver. `Kinds`' comment says it is a set of card types. `Filter.Kinds`' comment says zero means any asset, or any card in a pile.
- **`kindOf`** (`targeting.go`) maps `TypeManeuver` to `KindManeuver`.
- **`matchesPrinted`** (`targeting.go`) checks `Subtypes` through `Card.HasSubtype`, **before** the `f.Scratch` switch, so `Subtypes` and `Scratch` together are a conjunction. No second subtype comparison. Add the criterion to the comments on `matchesPrinted`, `matchesCard` and `matchesFilter`, which list what each reads.
- **`Filter.Scratch` stays.** Reuse check: the Scratch criterion reads `IsScratchCompany`, which reads `HasSubtype`, so one subtype read serves both, and the name "Scratch Company" stays in one place. `ScratchExcluded` is a "not this subtype", which `Subtypes` can't say. Its rule (`FR-TGT-001`'s *"non-Scratch"*, #78's one enum) is a different question from "any of these". No card definition may write `Subtypes: []string{"Scratch Company"}`: use `ScratchOnly`.
- **A subtype exclusion is not built.** The design note's "subtype out" has no reader: searching the Animus, Path and HQ text for `non-` finds only *non-Scratch* and *non-Headquarters*. Nothing is split off.
- **Sims:** no card definition sets either criterion, and `KindManeuver` has no reader, so play doesn't change. `go tool task sim-compare -- autoloop/m2-wave-d` should pass byte for byte. If it differs, don't change code to make it match: report the first differing game in the Log and hand off.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Synthetic cards only, in a new file `pkg/engine/filter_subtypes_test.go`. Pile cases can follow `pile_target_test.go`'s scene (a deploy trigger with a target slot). Give a synthetic card its subtypes by setting `Subtypes` on the card the helper returns. Every existing test passes unchanged.
- `TestFR_TGT_001_SubtypesMatchAnyOfTheList`: the pile holds a Fungus card, a Plant card, a "Plant, Fungus" card, a card with another subtype and a card with none. A slot with `Subtypes` Fungus and Plant lists exactly the first three, and the one chosen goes to hand.
- `TestFR_TGT_001_EmptySubtypesIsNoCriterion`: with `Subtypes` empty, the same pile lists all five.
- `TestFR_TGT_001_SubtypesAndKindsAreAConjunction`: with `Kinds` unit and `Subtypes` Fungus, a Fungus structure in the pile is not listed, and a Fungus unit is.
- `TestFR_TGT_001_SubtypesAndAllegianceAreAConjunction`: with allied allegiance and `Subtypes` Fungus, a Fungus card in the opponent's pile is not listed.
- `TestFR_TGT_001_SubtypesAndScratchAreAConjunction`: battlefield units, a table of two cases. `Subtypes` Infantry with `ScratchExcluded` lists a non-Scratch Infantry unit, and neither a Scratch Company Infantry unit nor a non-Infantry unit. `Subtypes` Fungus with `ScratchOnly` lists a Scratch Company unit that also carries Fungus, and not one without it.
- `TestFR_TGT_001_SubtypesReadABattlefieldAsset`: a battlefield slot with `Subtypes` Fungus lists a Fungus unit and not a unit with another subtype.
- `TestFR_TGT_001_SubtypesReadAFaceDownAssetsFace`: an asset face-down as Scratch Company whose card is a Fungus is not listed for `Subtypes` Fungus (`FR-ZON-005`, `FR-TGT-001`).
- `TestFR_TGT_001_PilePickReadsSubtypes`: an untargeted `SubjectPick` step with a pile filter and `Subtypes` Fungus asks a `DecidePick` over the matching cards only, and returns the one picked (`SVY-R3`).
- `TestFR_TGT_001_ManeuverKindListsOnlyManeuverCards`: the pile holds a Maneuver, a unit and a structure. A slot with `Kinds` `KindManeuver` lists only the Maneuver, and it goes to hand.
- `TestFR_TGT_001_OtherKindsStillRefuseManeuvers`: with `Kinds` `KindUnit`, or `KindUnit` and `KindStructure`, that pile's Maneuver is not listed. With `Kinds` empty, all three are.
- `TestFR_TGT_001_ManeuverKindMatchesNoAsset`: a battlefield slot with `Kinds` `KindManeuver`, with assets on the battlefield, has no legal entity, so no target is asked and its clause does nothing (`FR-TGT-003`).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test file may declare its own.

## Allowed Files
- `pkg/engine/ability/ability.go` (`Filter.Subtypes`, `KindManeuver`, their comments)
- `pkg/engine/targeting.go` (`kindOf`, `matchesPrinted`, the comments of `matchesCard` and `matchesFilter`)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_TGT_001_(Subtypes|EmptySubtypes|PilePickReadsSubtypes|ManeuverKind|OtherKinds)'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d` passes (no card uses either criterion; see Shape, Sims)
- [ ] `cards.NotYetExecutable` is unchanged.

## Log
- 2026-11-04 lead-architect (wave D run, session 5): no `main` commits to sync (`main` is the merge base, 58a2bc9). Not `WAVE-COMPLETE`: all four header cards are on `NotYetExecutable`. No `open: before` directive; #14 is `open` and orders the cards. Waiting directives checked, none fired: #6, #7 (mode is `serial`); #8 (this task doesn't touch `deck.go`); #15 (no turn-boundary hook); #71, #91 (2) (no such card); #104 (not due, `cards: []`, and neither question bears on a filter). The board's task table was empty, so this is the next wave D shape in T-0158's order. Reuse search: `grep Subtype|KindManeuver|kindOf` finds `HasSubtype` (reused) and no filter criterion; `Filter.Scratch` is kept, see Shape. A probe in the sync worktree, since reverted, showed that a slice field on `Filter` still passes `go build ./...`, and `go vet` of `pkg/engine`, `pkg/view`, `pkg/sim` and `pkg/eventlog` with their tests, so no code there compares a `Filter` with `==`. Two non-test files. Status `tests`; dispatching `rules-test-author`.
