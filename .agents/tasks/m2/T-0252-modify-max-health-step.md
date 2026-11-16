---
id: T-0252
title: "OpModifyStat on Maximum Health: Step.Stat, ability.StatMaxHealth, a max-health event (wave H shape, sonnet)"
kind: shape
tier: sonnet
wave: H
status: ready
branch: m2/autoloop-wave-h/T-0252-modify-max-health-step
depends_on: [T-0251]
claims: [shape:OpModifyStat on Maximum Health (Step.Stat naming the stat a modify step changes; ability.StatMaxHealth in the one Stat vocabulary, read by filters too; EventMaxHealth)]
cards: []
requirements: [FR-LAY-001, FR-LAY-002, FR-LAY-003, FR-TGT-001]
---

## Goal
A step can give its subjects a temporary Maximum Health change: `OpModifyStat` reads a new `Step.Stat` naming the stat it changes, and with `ability.StatMaxHealth` it makes T-0251's Layer 5 `Modifier{MaxHealth: n}` for `Step.Until` (`FR-LAY-002` Temporary Maximum Health Modifiers, `FR-LAY-003`). The zero value keeps meaning power, so every existing definition reads as before. Each asset changed logs one new `max-health` event, as `power` does (ADR-0005 addendum, written with this task).

`StatMaxHealth` joins the one `ability.Stat` vocabulary that filters already bound (`FR-TGT-001` Dynamic Stats, which tells Maximum Health from Current Health), so a filter on it reads printed Health for a card in a pile (`FR-LAY-001` Non-Battlefield Evaluation Boundary) and `Game.MaxHealth` for an asset. A second stat enum for modify steps would be a duplicate (D2.0).

*Poisoned Rations*, the card, is T-0253.

## Inputs
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): `FR-LAY-002` (line 506): Current Health & Lethal Destruction (`:507`, which names *Poisoned Rations*), State-Based Death Sweep Timing (`:508`), Temporary Maximum Health Modifiers (`:509`); `FR-LAY-001` (line 500): Non-Battlefield Evaluation Boundary (`:502`), Exclusive Combat Power Clamping (`:503`: Maximum Health doesn't clamp); `FR-LAY-003` (line 512): Full-Round Durations (`:514`); `FR-TGT-001` (line 386): Readiness, Subtypes & Dynamic Stats (`:391`).
- ADR-0005 (`docs/architecture/decisions/0005-structured-game-events-and-observer.md`): the T-0191 addendum (`:32`: `heal`, `armor`, `power`) and the T-0252 addendum added in this task's commit (`max-health`).
- Code of the day (run branch at 34b604d):
  - `pkg/engine/ability/ability.go`: `Stat` and its const block (`:287-296`; the comment calls health "a later stat"), `Filter.Stat`/`AtMost` (`:283-284`), the `OpModifyStat` comment (`:376`, "Other stats and durations come later"), `Step` (`:399-418`, `Until` at `:415`).
  - `pkg/engine/filter.go`: `withinStat` (`:48`), `cardStat` (`:53-61`), `assetStat` (`:67-75`).
  - `pkg/engine/effects.go`: `(*resolution).modifier` (`:255-260`), `stepModifyStat` (`:263-273`), which always makes `Modifier{Layer: 5, Power: n}` and logs `EventPower`.
  - `pkg/engine/layers.go`: `Modifier.MaxHealth` (`:21`) and `Game.MaxHealth` (`:60`), summed and unclamped (T-0251). `expireAtEndOf` already expires any modifier by its `Until`.
  - `pkg/engine/events.go`: the event-kind constants (`:265-294`), `EventPower` at `:288`.
  - Searched `pkg/engine` (non-test) for `MaxHealth` and `Health`: only `layers.go`'s field and reader, their callers, and the card loader. No step makes a max-health modifier, and no filter reads health. No existing equivalent.

## Shape (the lead's plan)
- `pkg/engine/ability/ability.go`:
  - `StatMaxHealth` after `StatRequisition` in the `Stat` const block: Maximum Health (`FR-LAY-002`), not Current Health. Rewrite the `Stat` comment: it is the one stat vocabulary that a filter bound (`Filter.Stat`) and a modify step (`Step.Stat`) both use; a filter reads a card in a pile by its printed value and an asset by its effective one.
  - `Step.Stat Stat`: `OpModifyStat` only, the stat the step changes. `StatNone` (the zero value) and `StatPower` are Combat Power; `StatMaxHealth` is Maximum Health. No other stat is valid on a modify step.
  - The `OpModifyStat` comment: the Subject gets +Amount to `Step.Stat` (power, or Maximum Health) for `Step.Until` (Layer 5, `FR-LAY-001`, `FR-LAY-002`, `FR-LAY-003`). Drop "Other stats and durations come later".
- `pkg/engine/filter.go`: `cardStat` returns `c.Health` for `StatMaxHealth`; `assetStat` returns `g.MaxHealth(a)`. Update `assetStat`'s comment so it covers Maximum Health.
- `pkg/engine/effects.go` (`stepModifyStat`): one loop, the modifier chosen by `st.Stat`: `StatNone`/`StatPower` → `Modifier{Layer: 5, Power: n}` with today's `power` event; `StatMaxHealth` → `Modifier{Layer: 5, MaxHealth: n}` through `r.modifier`, logged as `Event{Kind: EventMaxHealth, Player: -1, Target: a.ID, Amount: n}` with text `"%s gets %+d max health %s"` (the asset, `n`, `st.Until`). Any other stat panics (`engine: OpModifyStat on stat %d`, a definition bug, as `costs.go:223` does for an unknown cost kind). Update the function comment: power clamps at 0, Maximum Health doesn't (`FR-LAY-001`), and Maximum Health 0 or less is lethal at the sweep after the item (`FR-LAY-002`). No sweep call here: the sweep after each stack item already runs.
- `pkg/engine/events.go`: `EventMaxHealth = "max-health"` after `EventPower`, commented: an effect changed an asset's Maximum Health for a duration (`FR-LAY-002`).
- Not this task: the card (T-0253). No change to `layers.go` or `turn.go` (T-0251 made the modifier, its expiry and the 4c sweep).

## Tests (shape tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/modify_max_health_test.go`. Cards are synthetic (`unit`, `structure` in `scenario_test.go`). Player 0 is the active player. For a chosen unit, follow `opponent_turn_duration_test.go:17-29` (`powerFor`, `enactOn`: a free Maneuver with one unit target, built with `tgt` and enacted with `enactAction`), with a local helper for the max-health step `{Op: ability.OpModifyStat, Stat: ability.StatMaxHealth, Amount: ability.Value{N: n}, Subject: ability.SubjectTarget, Target: 0, Until: until}`. For "every allied asset" steps, use `onAlliedKinds` and `effectSource` (`effect_events_test.go:26-34`) and set the returned step's `Stat`. Read stats with `s.g.MaxHealth`, `s.g.CurrentHealth`, `s.g.Power`; use `s.alive`, `s.endTurn()`, `effectEvents`, `wantAssetEvents`.

- `TestFR_LAY_002_ModifyStepLowersMaxHealth`: player 1's unit with power 2, Health 5 gets -3 max health: Maximum Health 2, Current Health 2, `Damage` 0 (no damage counters, `FR-LAY-002` Temporary Maximum Health Modifiers), power still 2, and no `power` or `damage` event is logged for it.
- `TestFR_LAY_002_ModifyStepRaisesMaxHealth`: a Health 3 unit with `Damage` 2 gets +2: Maximum Health 5, Current Health 3, `Damage` 2.
- `TestFR_LAY_002_ModifyStepToZeroMaxHealthDestroysAtTheSweep`: player 1's unit with Health 3 and no damage gets -3: after the Maneuver resolves it is off the battlefield and its card is in player 1's discard pile, and no `damage` event names it (`FR-LAY-002` Current Health & Lethal Destruction: Maximum Health 0 is lethal with no damage). A Health 4 unit given -3 in its own scene survives, Current Health 1.
- `TestFR_LAY_002_ModifyStepMaxHealthExpiresAndRestores`: player 0's Health 5 unit with `Damage` 2 gets -2: alive, Current Health 1. After `s.endTurn()`: alive, Maximum Health 5, Current Health 3, `Damage` 2.
- `TestFR_LAY_003_ModifyStepMaxHealthUntilEndOfOpponentTurn`: player 0's Health 3 unit gets +2 with `Until: ability.UntilEndOfOpponentTurn`: Maximum Health 5 after resolution, still 5 after player 0's `s.endTurn()` (now player 1's turn), and 3 after player 1's `s.endTurn()`. In a second scene the same step with the zero `Until` is back to 3 after the first `s.endTurn()`.
- `TestEffectEvents_MaxHealthLogsTheChange`: for `n` in {2, -1}, as subtests: an `effectSource` step on every allied structure (`onAlliedKinds(ability.OpModifyStat, n, ability.KindStructure)` with `Stat` set to `ability.StatMaxHealth`) over two `s.wall()`s, activated with `activate(src, 0)`: `wantAssetEvents(t, effectEvents(s.g, from, EventMaxHealth), []*Asset{a, b}, []int{n, n}, "%s gets %+d max health until end of turn")`, no `EventPower` event from `from`, and both walls have Maximum Health 9 + n and unchanged power.
- `TestEffectEvents_MaxHealthFaceDownAssetStaysHidden`: one case of `TestEffectEvents_FaceDownAssetStaysHidden` (`effect_events_test.go:101-130`), written as a new test (don't edit that table): a +1 max-health step on every allied unit, with a face-down Scratch Company (`s.putScratch(0)`) on the battlefield: its `max-health` event has `Card` empty, a `Detail` that starts with `fd.String()`, and no hidden card name.
- `TestFR_LAY_001_ModifyStepStatPowerIsPower`: a step with `Stat: ability.StatPower` and Amount 2 on a chosen unit gives +2 power (`TempPower` 2), logs one `power` event and no `max-health` event, and leaves Maximum Health at printed. The existing zero-`Stat` power tests (`TestEffectEvents_PowerLogsThePowerChange`, `cornered_test.go:155-220`, `opponent_turn_duration_test.go`) stay as they are and cover the zero value.
- `TestFR_TGT_001_MaxHealthAtMostReadsEffectiveMaxHealthOnTheBattlefield`: as `TestFR_TGT_001_PowerAtMostReadsEffectivePowerOnTheBattlefield` (`power_at_most_test.go:63`, with its `paTargetAsk`-style decision read), a target filter `Stat: ability.StatMaxHealth, AtMost: 3` on units: of two Health 5 units, the one given `Modifier{Layer: 5, MaxHealth: -2}` directly is offered and the other isn't; a Health 3 unit with `Damage` 2 (Maximum Health 3) is offered, and a Health 4 unit with `Damage` 3 (Current Health 1, Maximum Health 4) isn't: the bound is on Maximum Health, not Current Health (`FR-TGT-001` Dynamic Stats).
- `TestFR_TGT_001_MaxHealthAtMostOffersPileCardsByPrintedHealth`: as `TestFR_TGT_001_PowerAtMostOffersPileCardsByPrintedPower` (`power_at_most_test.go:47`): a discard-pile filter with `Stat: ability.StatMaxHealth, AtMost: 3` offers a unit card with Health 3 and not one with Health 4.
- Named edit of an existing test, in the `test(T-0252)` commit: `TestEventKinds_ValuesAreTheSchema` (`pkg/engine/eventkinds_test.go`) gains `{EventMaxHealth, "max-health"}` right after `{EventPower, "power"}`. No other existing test changes.

Existing tests stay as they are and must still pass.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none expected. Helpers local to the new file live in it.

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/filter.go`
- `pkg/engine/effects.go`
- `pkg/engine/events.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "MaxHealth|ModifyStep|EventKinds"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (no starter or stress card sets `Step.Stat` or a health filter, so every log is byte-identical)

## Log
- 2026-11-15 lead-architect (wave H run, session 12): queued, split off T-0251 under the D2 size rule (the whole max-health shape touches six non-test files). Waits for T-0251.
- 2026-11-15 lead-architect (wave H run, session 13): T-0251 is `done` (merge 6676322), so this task is written against the code of the day (run branch at 34b604d). Directives: none `open`. The `waiting` ones were checked against the repo and none has fired (#15 (2): no draw or end-of-turn hook here; #71, #91 (2), #106, #120, #141 (3) aura half and #142 name cards or shapes no task touches yet). Reuse: `Step.Stat` takes the existing `ability.Stat` type, and `StatMaxHealth` is read by the existing `cardStat`/`assetStat`, so one stat vocabulary serves filters and modify steps; the modifier goes through `r.modifier` and T-0251's `Modifier.MaxHealth`. Searched `pkg/engine` and the task archive for a max-health step or a health filter: none. Four non-test files, within the D2 size rule. ADR-0005 addendum for `max-health` added in this commit. No ruling needed: `FR-LAY-002` settles the modifier and its lethality, and `FR-TGT-001` lets a stat bound be on Maximum Health. Re-read `BOARD.md` before committing: no directive added since. Status `tests`.
- 2026-11-16 rules-test-author: tests written in pkg/engine/modify_max_health_test.go: TestFR_LAY_002_ModifyStepLowersMaxHealth, TestFR_LAY_002_ModifyStepRaisesMaxHealth, TestFR_LAY_002_ModifyStepToZeroMaxHealthDestroysAtTheSweep, TestFR_LAY_002_ModifyStepMaxHealthExpiresAndRestores, TestFR_LAY_003_ModifyStepMaxHealthUntilEndOfOpponentTurn, TestEffectEvents_MaxHealthLogsTheChange (+2/-1 subtests), TestEffectEvents_MaxHealthFaceDownAssetStaysHidden, TestFR_LAY_001_ModifyStepStatPowerIsPower, TestFR_TGT_001_MaxHealthAtMostReadsEffectiveMaxHealthOnTheBattlefield, TestFR_TGT_001_MaxHealthAtMostOffersPileCardsByPrintedHealth. Edited TestEventKinds_ValuesAreTheSchema (eventkinds_test.go) as named. Helpers requested: none (reused powerFor, enactOn, onAlliedKinds, effectSource, paTargetAsk, etc.). Missing production symbols only: Step.Stat, ability.StatMaxHealth, EventMaxHealth. Cases left out: none. cardtown lint rules passes.
- 2026-11-16 lead-architect (wave H run, session 14): D4 tests review passed, round 1. Session 13 ended on a usage limit after dispatching the test author, with nothing committed; this session re-dispatched it. Checked each test against `FR-LAY-002` (`:507` Maximum Health 0 or less is lethal at the sweep with no damage; `:509` no damage counters, restored at 4c), `FR-LAY-003` (`:514`), `FR-LAY-001` (`:503` no clamp) and `FR-TGT-001` (`:391` Maximum Health vs Current Health): the expected outcomes match, every listed case is there (the Health 4 survivor, the zero-`Until` scene, the `StatPower` case, both filter tests), the walls are the only allied structures (`effectSource` is a unit), and the one named edit to `TestEventKinds_ValuesAreTheSchema` is in the `test(T-0252)` commit. Status `ready`; dispatched `kernel-engineer` (sonnet).
