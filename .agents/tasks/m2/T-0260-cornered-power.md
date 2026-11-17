---
id: T-0260
title: "A card's own Cornered power: `Def.Cornered` (renamed `ability.WhileCornered`) gains `Power`, a Layer 4 static read by `Game.Power` while the controller is Cornered (wave H shape, sonnet)"
kind: shape
tier: sonnet
wave: H
status: in-progress
branch: m2/autoloop-wave-h/T-0260-cornered-power
depends_on: [T-0180, T-0256]
claims: [shape:own Cornered power (WhileCornered.Power: a card's own "Cornered: This unit has +N power.", Layer 4, read by Game.Power), rename:ability.CorneredKeywords -> ability.WhileCornered]
cards: []
requirements: [FR-KWD-002e, FR-LAY-001, FR-LAY-003, FR-ZON-004, FR-ZON-005, SVY-R18]
---

## Goal
A card can have power, not only keywords, while its controller is Cornered: *"Cornered: This unit has +2 power."* (*Brigand Highlander*). The one shape for a card's own continuous "Cornered: …" (#117, T-0180: `Def.Cornered`, read by `Game.Keywords`) gains a `Power int`: while its controller is Cornered, a face-up battlefield asset has `+Power` Combat Power, read by `Game.Power`. It is a static ability, so it sits in Layer 4 (FRD 01's pipeline, `FR-LAY-001`: "Continuous Auras (Static Board State)"), added before Layer 6's clamp. Its gate is the same `g.cornered` that `Game.Keywords` reads, live, which is what `FR-LAY-003` asks for: combat reads both powers before any damage (`combat.go:105`) and deaths wait for the sweep, so the gate is locked for a simultaneous step without a snapshot.

Since the type then holds a power as well as keywords, `ability.CorneredKeywords` is renamed `ability.WhileCornered` ("what a card has while its controller is Cornered"). The rename is mechanical: the same fields, the same readers, the same tests.

*Brigand Highlander* itself is T-0261, queued (split off under the D2 size rule: the card adds `cards/cards.go` as a fifth non-test file). *Ancestral Monument* ("Other allied assets in the same zone have Cornered: Resistant 1.") is a different shape, a Cornered gate on T-0256's `StaticAura` inside `Game.auras`, and is its own task after T-0261.

## Inputs
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): the pipeline (`:476-498`; Layer 4 `:491`, "Continuous Auras (Static Board State)"); `FR-LAY-001` (`:500-504`): continuous evaluation and Combat Power the only stat that clamps, in Layer 6; `FR-LAY-003` (`:512-517`): Cornered is live, and locked at the start of a simultaneous damage step; `FR-KWD-002e` (`:617`): "Cornered: [Ability]." and `SVY-R18`; `FR-ZON-004`: last-known effective values from immediately before departure; `FR-ZON-005`: a face-down asset has no printed abilities.
- Survey (`docs/design/m2_starter_set_ability_survey.md`): *Brigand Highlander*'s row (`:203`, "self +2 power while Cornered", `FR-KWD-002e`, `FR-LAY-003`); `R18` (`:77`).
- Rulings: `SVY-R18`. No open question: the card's text says "this unit", its controller's Cornered, and the layer and lock are FRD 01's.
- Design note `docs/architecture/m2-ability-model.md`: `Static` (`:43`: "A card's own 'Cornered: keyword' is not a `Static`: it is `Def.Cornered` … 'Cornered: This unit has +2 power.' (*Brigand Highlander*) … not built yet"), Layers (`:55`).
- Directive #117 (archived, `archive/directives/directive_archive.md`): "Cornered: keyword" is one continuous shape, `Def.Cornered`. This task extends that shape instead of adding a second one.
- Code of the day (run branch at 54200ea):
  - `pkg/engine/ability/ability.go`: `Def.Cornered` (`:34-37`), `CorneredKeywords` (`:95-101`).
  - `pkg/engine/layers.go`: the file comment (`:1-6`), `Game.Power` (`:54-63`: printed + `TempPower` + `g.auras` Power, clamped), `Game.Keywords` (`:105-125`: reads `d.Cornered` with `g.cornered(a.Controller)`), `cornered` (`:140-145`).
  - `pkg/engine/cards/path.go`: *Allshan, Chainbreaker Revenant* (`:68`) and *Macabre Dancers* (`:640`) use `ability.CorneredKeywords`.
  - `pkg/engine/doc.go`: the Cornered keywords bullet (`:64-66`) and "Not built yet" (`:90-91`, which names "Cornered: This unit has +2 power.").
  - Every effective-power read goes through `Game.Power`: combat (`combat.go:105`), damage equal to power (`effects.go:420-423`), power-bound filters (`filter.go:72`), `effectiveState` (`layers.go:161`, the sweep's record) and the snapshot (`snapshot.go:64`, so the view). `Printed().Power` is read only inside `Game.Power`.
  - No recursion: `g.cornered` counts `TheaterOf`, which reads no power.
  - Searched `pkg/engine` (non-test) and the task archive for a Cornered power: none. `StaticAura` can't express "this unit" (its `Filter` has no self-only criterion, only `Other`), so a self power belongs to `Def.Cornered`, as the design note says, and no second scheme for a card's own Cornered abilities appears.

## Shape (the lead's plan)
- `pkg/engine/ability/ability.go`:
  - Rename `CorneredKeywords` to `WhileCornered`. It gains `Power int` after `Resistant`: the card's Layer 4 power while its controller is Cornered (`FR-LAY-001`, `FR-KWD-002e`), 0 for none; it may be negative, like any static power, and the clamp is Layer 6's. Its comment: what a card has only while its controller is Cornered; `g.Keywords` adds the keywords as a Layer 3 grant and `g.Power` the power in Layer 4, on a face-up asset; the gate is locked for a simultaneous step (`FR-LAY-003`).
  - `Def.Cornered`'s type becomes `WhileCornered`, and its comment names power as well as keywords ("Cornered: Aggressive 3.", "Cornered: This unit has +2 power.").
- `pkg/engine/layers.go`:
  - `Game.Power` adds `d.Cornered.Power` when `a`'s face has an `Ability` and `g.cornered(a.Controller)`, beside the auras' Layer 4 power, before the clamp. Its comment names it (Layer 4: the card's own Cornered power and auras).
  - `Game.Keywords` keeps its logic, with the renamed type in its zero-value comparison. Read the gate once, in one small helper both readers call (for example `func (g *Game) whileCornered(a *Asset) ability.WhileCornered`, which returns the zero value when the face has no `Ability` or the controller is not Cornered), so the "face-up, own controller, Cornered" rule is written once.
  - The file comment: Layer 4 also holds a card's own Cornered power.
- `pkg/engine/cards/path.go`: *Allshan* and *Macabre Dancers* use `ability.WhileCornered`, same values. Nothing else changes there.
- `pkg/engine/doc.go`: the Cornered keywords bullet becomes "A card's own Cornered abilities ("Cornered: Aggressive 3.", "Cornered: This unit has +2 power.", `Def.Cornered`)": `g.Keywords` adds the keywords as a Layer 3 grant and `g.Power` the power in Layer 4, on a face-up asset, while its controller is Cornered. Under "Not built yet", drop "and "Cornered: This unit has +2 power."", and the opening "other than a card's own Cornered keywords" becomes "other than a card's own Cornered abilities".
- No change to `combat.go`, `effects.go`, `filter.go`, `sweep.go`, `snapshot.go` or `pkg/view`: each reads `Game.Power` already.
- Not this task: *Brigand Highlander* (T-0261); *Ancestral Monument*'s Cornered-gated aura (its own task).

## Tests (shape tasks: written by rules-test-author; engineers never edit these)
**Edits to existing tests** (in the `test(T-0260)` commit, nothing else changed in them):
- `pkg/engine/cornered_keywords_test.go`: every `ability.CorneredKeywords` becomes `ability.WhileCornered` (the `corneredCard` helper's parameter and its call sites in `TestFR_KWD_002f_CorneredAggressiveIsLabelledTheKeyword`, `TestFR_LAY_003_CorneredAggressiveIsCheckedWhenTheAttackHappens`, `TestFR_KWD_002e_CorneredResistantReducesDamageOnlyWhileCornered`, `TestFR_LAY_003_SimultaneousDamageLocksCorneredKeywords` and `TestFR_ZON_005_FaceDownAssetHasNoCorneredKeyword`), and the file comment's "ability.CorneredKeywords" (`:18`) becomes "ability.WhileCornered".
- `pkg/engine/grant_keyword_test.go`: in `TestFR_LAY_001_GrantedKeywordAddsToCornered`, `ability.CorneredKeywords{Resistant: 1}` becomes `ability.WhileCornered{Resistant: 1}`.

**New tests** go in `pkg/engine/cornered_power_test.go`. Cards are synthetic, through `corneredCard` (`cornered_keywords_test.go:23`) with `ability.WhileCornered{Power: 2}`. Player 0 is Cornered when player 1 controls strictly more Main Theater assets (units and structures) than player 0. Read `s.g.Power(a)`.
- `TestFR_KWD_002e_CorneredPowerOnlyWhileCornered`: a printed-power-2 unit with `Power: 2` in player 0's Main Theater. With one player 1 Main Theater unit (1 against 1): power 2. Add a second player 1 Main Theater unit: power 4. Move one of them to player 1's Reserves (`s.moveTo`): power 2 again. It has no Cornered keywords either way (`s.g.Keywords` equals its printed keywords).
- `TestFR_KWD_002e_CorneredPowerInReservesFollowsTheController`: the same unit in player 0's Reserves while player 0 is Cornered (player 1 has a Main Theater asset, player 0 none): power 4. Headquarters don't count for either side.
- `TestFR_LAY_001_CorneredPowerAddsBeforeTheClamp`: a printed-power-1 unit with `Power: 2`, Cornered, given a Layer 5 power modifier of −4 (as other layer tests build one, for example `Modifier{Layer: 5, Power: -4}` appended to `Modifiers`): power 0 (1 + 2 − 4 clamps). With −2 instead: power 1. It adds to an aura's Layer 4 power as well: next to an allied T-0256 aura source giving it +1 (an `ability.StaticAura` whose filter matches it), Cornered and with no modifier, power 4.
- `TestFR_LAY_003_SimultaneousDamageLocksCorneredPower`: player 0's Readied unit (printed power 1, `Power: 2`, health 20) in the Main Theater; player 1 controls two Main Theater units, defender D (power 0, health 3) and a wall, so player 0 is Cornered. Player 0's unit attacks D: D takes 3 and is destroyed, though its death evens the counts (the gate is locked for the combat step, `FR-LAY-003`). After the sweep the counts are 1 against 1, so the attacker's power is 1.
- `TestFR_ZON_004_SweepRecordsCorneredPowerBeforeAnyLeaves`: player 0's unit U (`Power: 2`) and player 1's two Main Theater units A and B; U and A both lethal. After `s.g.deathSweep()`, both are gone and `s.g.LastKnownEffective(U.ID).Power` is printed + 2: the sweep records every dying asset before any leaves (`sweep.go:11`), so A's departure doesn't remove U's Cornered power from its record.
- `TestFR_ZON_005_FaceDownAssetHasNoCorneredPower`: the synthetic card put face down (as `s.putScratch` builds one, with the card as its `Inst.Card`) while player 0 is Cornered: its power is the face-down printed power, with no +2.

Existing tests (with the renames above) must still pass.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none expected. Helpers local to the new file live in it.

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/layers.go`
- `pkg/engine/cards/path.go` (the rename only)
- `pkg/engine/doc.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "Cornered"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (no card prints a Cornered power yet, and the rename changes no value, so every log is byte-identical)

## Log
- 2026-11-16 lead-architect (wave H run, session 21): written against the run branch at 54200ea. *Brigand Highlander* is next in Deliverable 6m's order (wave E, after T-0259's *Hulking Incongruity*). No ruling needed: "this unit" while its own controller is Cornered is `FR-KWD-002e`; a static ability's power is Layer 4 (`FR-LAY-001`'s pipeline); and the live gate already gives `FR-LAY-003`'s lock, since combat reads both powers before damage and deaths wait for the sweep. Reuse: the shape extends #117's `Def.Cornered`, the one shape for a card's own continuous Cornered abilities, rather than a self-only `StaticAura` (which would be a second way to write "Cornered: Resistant 1." on *Macabre Dancers*, a duplicate under D2.0); `Game.Power` is the one power read, so combat, damage, filters, the sweep and the view need no change. The rename to `WhileCornered` is there because the type no longer holds only keywords; its test edits are named above. Split from the card under the D2 size rule (five non-test files with `cards/cards.go`); T-0261, *Brigand Highlander*, queued in this commit. Four non-test files. Directives: none `open`; each `waiting` one was checked against the repo and none has fired (#6, #7: mode is `serial`; #15 (2): no draw or end-of-turn hook; #71: no each-player effect; #91 (2), #106, #120, #142 name cards or shapes no task touches yet). `main` has no commits the run branch lacks. Status `tests`.
- 2026-11-16 rules-test-author: tests written in `pkg/engine/cornered_power_test.go`: TestFR_KWD_002e_CorneredPowerOnlyWhileCornered, TestFR_KWD_002e_CorneredPowerInReservesFollowsTheController, TestFR_LAY_001_CorneredPowerAddsBeforeTheClamp (clamp and aura subtests), TestFR_LAY_003_SimultaneousDamageLocksCorneredPower, TestFR_ZON_004_SweepRecordsCorneredPowerBeforeAnyLeaves, TestFR_ZON_005_FaceDownAssetHasNoCorneredPower. The `ability.CorneredKeywords` -> `ability.WhileCornered` rename was applied in `cornered_keywords_test.go` and `grant_keyword_test.go` only. Helpers requested: none. Left out: none. Notes: the aura case sits in Reserves (source and unit) so the Main Theater counts keep the unit Cornered; the face-down case sits in Reserves for the same reason. `go vet` fails only on the missing `ability.WhileCornered`. `cardtown lint rules` passes.
- 2026-11-16 lead-architect (D4): tests approved, round 1. Each checks what the cited rules give: the +2 applies only while the controller is Cornered and follows the counts live (`FR-KWD-002e`), belongs to the controller wherever the unit is, and grants no keyword; it is summed with Layer 5 before the Layer 6 clamp and with an aura's Layer 4 power (`FR-LAY-001`); combat damage reads it locked for the step though the defender's death evens the counts (`FR-LAY-003`); the sweep's record keeps it although an enemy dies in the same sweep (`FR-ZON-004`); a face-down asset has none (`FR-ZON-005`). The edits to `cornered_keywords_test.go` and `grant_keyword_test.go` are the rename only, as the task names. Status `ready`; dispatched `kernel-engineer` (sonnet).
