---
id: T-0052
title: A zone criterion on Filter ("in the Main Theater", FR-TGT-001), with Devilhorn Captain (wave B shape 2, sonnet)
kind: shape
tier: sonnet
wave: B
status: queued
branch: m2/T-0052-filter-zone
depends_on: [T-0050]
claims: [filter:Zone, card:Devilhorn Captain]
cards: [Devilhorn Captain]
requirements: [FR-TGT-001]
---

## Goal
A `Filter` can restrict its matches to one battlefield zone (`FR-TGT-001`'s Zone criterion, "in the Main Theater"). The first card to use it, *Devilhorn Captain* (`Unbowed: Deal 1 damage to chosen asset in the Main Theater.`, 3 copies in the Path starter deck), becomes executable in the same task. *Animoceps Shambler* (Animus deck card, "chosen allied asset in Main Theater") will reuse it once its exhaust hook exists.

## Inputs
- FRD 01: `FR-TGT-001` (the Zone criterion, line 300); `FR-TGT-003` (second-stage verification at resolution); `FR-KWD-002a` (line 505 and its bullets; line 506 names this card); `FR-ZON-001` (zones: an asset in the Headquarters Zone or Reserves is not in the Main Theater).
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 140.
- Card (`data/cards/units-and-structures-path.csv`): Path Unit (Officer, Devilhorn), cost 2R 0T, deploys to Reserves, 2/3, yields 1R 2T, rarity Uncommon.
- Existing machinery to reuse: `ability.Zone` (`ZoneAny`, `ZoneReserves`, `ZoneMainTheater`) and the `deployedTo` mapping in `trigger.go`; `matchesFilter` in `targeting.go`, which both targeting and `SubjectAll` use; `HookDamaged` (T-0050).

### Lead decisions (technical)
1. `ability.Filter` gains `Zone ability.Zone`. `ZoneAny` (the zero value) means any battlefield zone, so every existing definition is unchanged. `matchesFilter` checks it, so it applies on placement, at resolution re-verification and to `SubjectAll` steps alike.
2. Only fixed zones. "Same zone as this source" is a separate shape that waits for the same-zone auras and Hammerhand Bivouac (directive #14 defers them).
3. The Captain's definition: a `HookDamaged` trigger with one target `{Filter: {Zone: ZoneMainTheater}}` (either side, units and structures; Headquarters are never in the Main Theater) and a `DealDamage` 1 step on `SubjectTarget` slot 0.
4. Update the stale "Not built yet" list in `pkg/engine/doc.go`: Unbowed landed with T-0050.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written when the task moves to `tests`.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- To be filled in by rules-test-author.

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/targeting.go`
- `pkg/engine/doc.go`
- `pkg/engine/cards/path.go`
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_TGT_001|TestCard_DevilhornCaptain'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks; report the before and after summary lines in the Log

## Log
- 2026-10-23 lead-architect: Wrote the task, `queued` behind T-0051 (serial mode). Split from the Brigands card task because Devilhorn Captain needs the new zone criterion.
