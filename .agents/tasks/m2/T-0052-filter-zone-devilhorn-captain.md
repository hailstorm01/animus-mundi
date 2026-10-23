---
id: T-0052
title: A zone criterion on Filter ("in the Main Theater", FR-TGT-001), with Devilhorn Captain (wave B shape 2, sonnet)
kind: shape
tier: sonnet
wave: B
status: tests
branch: m2/T-0052-filter-zone
depends_on: [T-0050]
claims: [filter:Zone, card:Devilhorn Captain]
cards: [Devilhorn Captain]
requirements: [FR-TGT-001, FR-TGT-003, FR-KWD-002a, FR-TRG-003, FR-TRG-004]
---

## Goal
A `Filter` can restrict its matches to one battlefield zone (`FR-TGT-001`'s Zone criterion, "in the Main Theater"). The first card to use it, *Devilhorn Captain* (`Unbowed: Deal 1 damage to chosen asset in the Main Theater.`, 3 copies in the Path starter deck), becomes executable in the same task. *Animoceps Shambler* (Animus deck card, "chosen allied asset in Main Theater") will reuse it once its exhaust hook exists.

## Inputs
- FRD 01: `FR-TGT-001` (the Zone criterion, line 300); `FR-TGT-003` (second-stage verification at resolution); `FR-KWD-002a` (line 505 and its bullets; line 506 names this card); `FR-ZON-001` (zones: an asset in the Headquarters Zone or Reserves is not in the Main Theater).
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 140.
- Card (`data/cards/units-and-structures-path.csv`): Path Unit (Officer, Devilhorn), cost 2R 0T, deploys to Reserves, 2/3, yields 1R 2T, rarity Uncommon.
- Existing machinery to reuse: `ability.Zone` (`ZoneAny`, `ZoneReserves`, `ZoneMainTheater`) and `deployedTo` in `trigger.go`, which despite its name checks an asset's current zone against an `ability.Zone` (call it from `targeting.go`; same package); `matchesFilter` in `targeting.go`, which both targeting and `SubjectAll` use; `HookDamaged` (T-0050).
- Ruling `SVY-R13` (FRD 01 v3.10 changelog, line 18; `FR-KWD-002a` line 505): lethal Unbowed is buffered and placed after the death sweep, in one APNAP window with the death triggers from the same event.

### Lead decisions (technical)
1. `ability.Filter` gains `Zone ability.Zone`. `ZoneAny` (the zero value) means any battlefield zone, so every existing definition is unchanged. `matchesFilter` checks it, so it applies on placement, at resolution re-verification and to `SubjectAll` steps alike.
2. Only fixed zones. "Same zone as this source" is a separate shape that waits for the same-zone auras and Hammerhand Bivouac (directive #14 defers them).
3. The Captain's definition: a `HookDamaged` trigger with one target `{Filter: {Zone: ZoneMainTheater}}` (either side, units and structures; Headquarters are never in the Main Theater) and a `DealDamage` 1 step on `SubjectTarget` slot 0.
4. **Board directive #46 (user ruling `SVY-R13`), folded in here:** `rules-test-author` adds `TestFR_KWD_002a_*` or `TestFR_TRG_004_*` tests pinning that a lethally damaged asset's Unbowed trigger is placed after the death sweep, in one APNAP window with the death triggers from the same event. Use synthetic cards: a dying asset with both an Unbowed and a death trigger, and a case where another asset dies in the same sweep. The kernel already behaves this way (T-0050), so these tests should pass when written.
5. Not in this task: refreshing the stale "Not built yet" list in `pkg/engine/doc.go` (Unbowed landed with T-0050). It would make five non-test files (D2.4), so the next shape task that touches `engine` carries it.

### What the tests can observe (for rules-test-author)
**The zone criterion (synthetic cards, `TestFR_TGT_001_*`, `TestFR_TGT_003_*`):**
- A chosen target "in the Main Theater" offers only assets in the Main Theater, of either side unless an allegiance is also given. Reserves assets and both Headquarters are never options.
- With no asset in the Main Theater, the clause has no legal target: nothing is asked and nothing happens (existing T-0017 behavior).
- Second-stage verification: a target that leaves the Main Theater, or the battlefield, before the ability resolves is illegal, and that clause does nothing (`FR-TGT-003`). The Chooser-mutation technique from `attack_reverify_test.go` (`midStack`) can move it mid-stack.
- A zone criterion combines with the others (conjunction, `FR-TGT-001`): for example opposing and Main Theater.
- An untargeted "all assets in the Main Theater" step (`SubjectAll` with the filter) hits exactly those assets.

**Devilhorn Captain (`TestCard_DevilhornCaptain_*`, from the CSV text):**
- It enters Reserves as a 2/3 for 2 Requisition. Nothing happens on deploy.
- Whenever it takes damage (`FR-REP-004` Stage 5), its controller chooses one asset in the Main Theater, of either side, and deals it 1 damage. Opposing Ward assets are not options.
- It fires in Reserves too (for example from *Saboteur*'s allied 1 damage), and still only Main Theater assets are options.
- It can choose itself while it is alive and in the Main Theater (no "other"). That 1 damage fires its Unbowed again.
- **Lethal damage:** the trigger resolves from the discard pile (`FR-KWD-002a` second bullet). Its last-known controller chooses, and the dead Captain is not an option.
- Damage reduced to 0 (Resistant, a 0-power attacker) does not fire it. A face-down Captain fires nothing (`FR-ZON-004`).
- A sole legal target is not asked. A timeout picks a legal target (`FR-MOD-001`).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written by rules-test-author.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- To be filled in by rules-test-author.

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/targeting.go`
- `pkg/engine/cards/path.go`
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_TGT_00|TestFR_KWD_002a|TestFR_TRG_00|TestCard_DevilhornCaptain'`
- [ ] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged; Path's executable count rises by 3, from 16/50 to 19/50
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks; report the before and after summary lines in the Log

## Log
- 2026-10-23 lead-architect: Wrote the task, `queued` behind T-0051 (serial mode). Split from the Brigands card task because Devilhorn Captain needs the new zone criterion.
- 2026-10-23 lead-architect: Moved to `tests`. Checked every waiting directive's trigger first (#6, #7, #8, #15 and #20 have not fired). Folded in directive #46 (decision 4) and added the test guidance. Took `doc.go` out of Allowed Files to keep to four non-test files (decision 5). Starter-deck cards not yet executable: 29, plus both Headquarters; expected after: 28.
