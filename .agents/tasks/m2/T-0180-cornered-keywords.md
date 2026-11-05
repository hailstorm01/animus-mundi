---
id: T-0180
title: A keyword behind Cornered is one continuous shape (Def.Cornered, read in g.Keywords; directive #117, wave D.1 shape, sonnet)
kind: shape
tier: sonnet
wave: D.1
status: tests-review
branch: m2/autoloop-wave-d.1/T-0180-cornered-keywords
depends_on: [T-0179]
claims: [shape:Def.Cornered, shape:CorneredKeywords, layer:3 keywords behind Cornered]
cards: [Allshan, Chainbreaker Revenant]
requirements: [FR-KWD-002e, FR-KWD-002f, FR-LAY-001, FR-LAY-003]
---

## Goal
A card lists the keywords it has while its controller is Cornered, and `g.Keywords` adds them, on a face-up asset. *Allshan, Chainbreaker Revenant* moves to it from a gated trigger, so `g.Keywords` shows its Aggressive 3 exactly while Cornered, as `SVY-R18` says. *Macabre Dancers*' "Cornered: Resistant 1." (wave E) then needs no new shape.

## Inputs
- FRD 01 v3.21: `FR-KWD-002e` (line 583, including `SVY-R18`), `FR-KWD-002f` (line 584), the layer list (line 455, Layer 3) and `FR-LAY-001` (lines 467-471), `FR-LAY-003` (lines 479-483: the enqueue gate for triggers, and continuous gates locked for a simultaneous step), `FR-CBT-002` Step 3 (lines 559-565), `FR-REP-004` (line 500), `FR-ZON-005` for a face-down asset.
- Survey rows: `docs/design/m2_starter_set_ability_survey.md`, R18 (line 77) and Allshan (line 168).
- Rulings that apply: `SVY-R18`. Directive #117 (structural review 8, finding 2).
- Design note: `docs/architecture/m2-ability-model.md`, Key types. #121 (d) updates its `Static` line and Layers paragraph; that is not this task.
- Card text: *Allshan, Chainbreaker Revenant*'s lines are "Unbowed: This unit gains 2 armor." and "Cornered: Aggressive 3." in `data/cards/units-and-structures-path.csv`.
- Kernel today:
  - `g.Keywords` ([layers.go:23](../../../pkg/engine/layers.go)) returns the printed keywords. `a.Printed()` is the Scratch Company face for a face-down asset, and that face has no `Ability`.
  - `g.cornered` ([trigger.go:128](../../../pkg/engine/trigger.go)) is the one read of Cornered. It stays where it is.
  - `fireAttacks` ([trigger.go:135-137](../../../pkg/engine/trigger.go)) already reads `g.Keywords(a).Aggressive` when the attack happens and buffers `ability.Aggressive(n)`, labelled `<name>: Aggressive <n>`. That is the moment `R18` and `FR-KWD-002e` name, so it needs no change.
  - `resolveAttack` ([combat.go:106-111](../../../pkg/engine/combat.go)) and `dealDamage` ([damage.go:23](../../../pkg/engine/damage.go)) remove nothing from the battlefield, and destruction waits for the sweep. So the Main Theater counts can't change during a simultaneous damage step, and `FR-LAY-003`'s lock holds with no new code. The tests below pin that.
  - `ability` can't import `engine`, so `engine.Keywords` isn't available to a `Def`.

### Shape (lead decisions, binding)
- **`ability.CorneredKeywords`** (`ability/ability.go`): a struct with two `int` fields, `Aggressive` and `Resistant`, 0 for absent. Those are the keywords M2 cards print behind "Cornered:" (*Allshan*, *Macabre Dancers*, and the grant on *Ancestral Monument*). Its comment cites `FR-LAY-001` and `FR-LAY-003`. Moving `engine.Keywords` into `ability` for this is not worth it yet: wave E's auras decide what a grant looks like, and a card's `CorneredKeywords{Aggressive: 3}` would survive that move with a one-word edit.
- **`Def.Cornered CorneredKeywords`** (`ability/ability.go`), the last field of `Def`. Its comment quotes "Cornered: Aggressive 3." and cites `SVY-R18` and `FR-KWD-002e`, and says it is zero for a card without one.
- **`g.Keywords`** (`layers.go`) starts from `a.Printed().Keywords`. If `a.Printed().Ability` is non-nil, its `Cornered` is not the zero value, and `g.cornered(a.Controller)` holds, it adds both fields to the result. The zero-value check comes before `g.cornered`, because `g.Keywords` is read for every asset in every snapshot and `g.cornered` walks the Main Theater. The addition matches `ParseKeywordLine`'s `+=`. No card prints a keyword both ways, so no test pins it. The doc comment says what it returns and cites `SVY-R18`. The file's header comment stops saying nothing grants keywords. `layers.go` imports `ability`.
- **Allshan** (`cards/path.go`): gains `Cornered: ability.CorneredKeywords{Aggressive: 3}`, loses `cornered(ability.Aggressive(3))` from `Triggers`, and keeps the "Cornered: Aggressive 3." line in `Text`. The `cornered` helper in `path.go` goes, since nothing else uses it.
- **Not this task:**
  - `Trigger.Cornered`, its gate in `fireHook` and the `TestFR_KWD_002e_*` tests in `cornered_test.go` stay unedited, although no card then uses the field.
  - `newAsset`'s read of printed keywords stays: `Armor` and `Deploys Exhausted` are entry state, and neither can sit behind Cornered here.
  - `doc.go`'s "not built" list is #121 (c).
  - *Macabre Dancers* stays on `NotYetExecutable`; its Unbowed `HEAL` line is wave E's.
  - The view and monitor cases of #117 are T-0181, because `rules-test-author` writes only under `pkg/engine` (D2.4, as for T-0164).
- **Sims:** `go tool task sim-compare -- autoloop/m2-wave-d.1` should pass byte for byte, because Allshan has one attack trigger and only a prompt label changes. If it differs, don't change code to make it match: report the first differing game in the Log and hand off.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cornered_keywords_test.go`. The names `ability.CorneredKeywords` and `ability.Def.Cornered` come from the Shape above, so the tests won't compile until the code exists. A synthetic card sets them through `striker`'s or `unit`'s `*Card`, by setting `Ability` directly. Cornered is 1 asset against 2 in the Main Theater; counts are strict, so 1 against 1 is not Cornered. Every existing test passes unchanged, including every Allshan test in `cards_path_allshan_test.go`.
- `TestCard_Allshan_HasAggressiveOnlyWhileCornered` (`SVY-R18`; real card): `s.g.Keywords(allshan).Aggressive` is 3 while Cornered and 0 when it isn't, in both directions on one scene (an enemy added, then removed). Allshan in Reserves reads the same, since Cornered belongs to its controller and counts only the Main Theater. A `Snapshot`'s `AssetSnap.Keywords` for Allshan agrees in both states.
- `TestCard_Allshan_AggressiveComesFromTheKeyword` (`FR-KWD-002f`; real card): attacking while Cornered places one `StackAbility` item with `Keyword` `ability.KeywordAggressive` whose source is Allshan, and combat deals 8. Not Cornered, no such item is placed and combat deals 5.
- `TestFR_KWD_002f_CorneredAggressiveIsLabelledTheKeyword`: a synthetic attacker with `Cornered: {Aggressive: 3}` and a printed attack trigger, while Cornered. The `DecideOrder` decision lists two options, and one is labelled `<name>: Aggressive 3`.
- `TestFR_LAY_003_CorneredAggressiveIsCheckedWhenTheAttackHappens`, two subtests with a synthetic attacker whose printed attack trigger kills a chosen enemy or allied asset:
  - Cornered at declaration, and the printed trigger resolves first and evens the counts: the Aggressive item already on the stack still gives +3, and combat deals power + 3.
  - Not Cornered at declaration, and the printed trigger kills an allied asset so the attacker's controller is Cornered by Step 3: no Aggressive item is placed, and combat deals printed power.
- `TestFR_KWD_002e_CorneredResistantReducesDamageOnlyWhileCornered` (`FR-REP-004`): a synthetic unit with `Cornered: {Resistant: 1}` takes 1 less from a combat hit and from an effect's damage while Cornered, and the full amount when it isn't. The `EventDamage` reports `Resistant` 1 and 0 in the two states.
- `TestFR_LAY_003_SimultaneousDamageLocksCorneredKeywords`, two subtests with a synthetic `Cornered: {Resistant: 1}` unit:
  - Combat: it attacks while Cornered and kills the defender, which would even the counts. Its retaliation damage is still reduced by 1. After the sweep, `g.Keywords` shows Resistant 0.
  - Area damage: one step deals damage to every asset, lethal to an enemy, and the unit, listed after that enemy in fixed order, still takes 1 less. After the sweep, `g.Keywords` shows Resistant 0.
- `TestFR_ZON_005_FaceDownAssetHasNoCorneredKeyword`: a face-down asset of a card with `Cornered: {Aggressive: 3, Resistant: 1}`, its controller Cornered. `g.Keywords` equals the Scratch Company face's keywords, its attack places no Aggressive item, and a hit isn't reduced.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test file may declare its own.

## Allowed Files
- `pkg/engine/ability/ability.go` (`CorneredKeywords`, `Def.Cornered`)
- `pkg/engine/layers.go` (`g.Keywords`, the header comment)
- `pkg/engine/cards/path.go` (Allshan, the `cornered` helper)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'Allshan|TestFR_KWD_002|TestFR_LAY_003|TestFR_ZON_005_FaceDown'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d.1` passes (see Shape, Sims)
- [ ] `cards.NotYetExecutable` is unchanged

## Log
- 2026-11-05 lead-architect (wave D.1 run, session 11): no `main` commits to sync (`git log autoloop/m2-wave-d.1..main` is empty). Not `WAVE-COMPLETE`: #117-#119 and #121 are `open: before the next Deliverables task`. Next by board order: #117. #122 reads `open: after Wave D.1 (#121), before Wave E`, so it doesn't apply until #121 is done. Waiting directives checked, none fired: #6, #7 (mode is `serial`); #8 (no task touches `deck.go`); #15 (no turn-boundary hook); #71, #91 (2), #106, #118 (b), #120 (no such card). #14 stays `open`, not blocking: 10 deck cards on `NotYetExecutable` (#14 (e)). Reuse search: `grep -i cornered` finds `g.cornered` (reused) and `Trigger.Cornered` (kept, #117), and nothing continuous; `ability` has no keyword-value type, and `engine.Keywords` can't be imported there. A lead probe of this Shape, since reverted, passes `go tool task check`, and `go tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`. The view and monitor cases of #117 (a), (c) are split off as T-0181, `queued` (D2.4). Three non-test files. Status `tests`; dispatching `rules-test-author`.
- 2026-11-05 rules-test-author: tests written in `pkg/engine/cornered_keywords_test.go`: `TestCard_Allshan_HasAggressiveOnlyWhileCornered`, `TestCard_Allshan_InReservesHasAggressiveWhileItsControllerIsCornered`, `TestCard_Allshan_AggressiveComesFromTheKeyword`, `TestFR_KWD_002f_CorneredAggressiveIsLabelledTheKeyword`, `TestFR_LAY_003_CorneredAggressiveIsCheckedWhenTheAttackHappens` (2 subtests), `TestFR_KWD_002e_CorneredResistantReducesDamageOnlyWhileCornered` (combat and effect, cornered and not), `TestFR_LAY_003_SimultaneousDamageLocksCorneredKeywords` (combat, area), `TestFR_ZON_005_FaceDownAssetHasNoCorneredKeyword` (attack, hit). Test-file helpers are declared in the file (`corneredCard`, `pickTarget`, `aggressiveItems`, `snapKeywords`, `damageEvent`). Helpers requested: none. `go vet` fails only on the undefined `ability.CorneredKeywords`/`Def.Cornered`; `cardtown lint rules` passes. Nothing deliberately left out. Note: the area-damage subtest relies on fixed order = placement order (enemy put before the unit), per the task text.
