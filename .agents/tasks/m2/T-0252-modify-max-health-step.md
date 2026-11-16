---
id: T-0252
title: "OpModifyStat on Maximum Health: Step.Stat, ability.StatMaxHealth, a max-health event (wave H shape, sonnet)"
kind: shape
tier: sonnet
wave: H
status: queued
branch: m2/autoloop-wave-h/T-0252-modify-max-health-step
depends_on: [T-0251]
claims: [shape:OpModifyStat on Maximum Health (Step.Stat naming the stat a modify step changes; ability.StatMaxHealth in the one Stat vocabulary, read by filters too; EventMaxHealth)]
cards: []
requirements: [FR-LAY-001, FR-LAY-002, FR-LAY-003, FR-TGT-001]
---

## Goal
A step can give its subjects a temporary Maximum Health change: `OpModifyStat` reads a new `Step.Stat` naming the stat it changes, and with `ability.StatMaxHealth` it makes T-0251's Layer 5 `Modifier{MaxHealth: n}` for `Step.Until` (`FR-LAY-002`, `FR-LAY-003`). The zero value keeps meaning power, so every existing definition reads as before. Each asset changed logs one new `max-health` event, as `power` does (ADR-0005 addendum).

`StatMaxHealth` joins the one `ability.Stat` vocabulary that filters already bound (`FR-TGT-001` Dynamic Stats), so a filter on it reads printed Health for a card in a pile and `Game.MaxHealth` for an asset. A second stat enum for modify steps would be a duplicate (D2.0).

*Poisoned Rations*, the card, is T-0253.

## Covers (write Inputs, Shape, Tests and Allowed Files against the code of the day before this leaves `queued`)
- `pkg/engine/ability/ability.go`: `StatMaxHealth` in the `Stat` const block (`:290-296` at 0647f4f; update the `Stat` comment, which names health as "a later stat"); `Step.Stat` (`OpModifyStat` only: `StatNone`, the zero value, and `StatPower` are Combat Power; `StatMaxHealth` is Maximum Health); the `OpModifyStat` comment (`:376`).
- `pkg/engine/filter.go`: `cardStat` and `assetStat` (`:53-75` at 0647f4f) read `StatMaxHealth` (printed `Health`; `g.MaxHealth(a)`).
- `pkg/engine/effects.go`: `stepModifyStat` (`:263-273` at 0647f4f) builds the modifier through `(*resolution).modifier` with `MaxHealth: n` when `st.Stat` is `StatMaxHealth`, and logs `EventMaxHealth` with text `<asset> gets <±n> max health <Until>` through the asset's `String`.
- `pkg/engine/events.go`: `EventMaxHealth = "max-health"`. The lead adds an ADR-0005 addendum (`docs/architecture/decisions/0005-structured-game-events-and-observer.md`) when writing the task.
- Tests to pin: a synthetic Maneuver's `-3` max-health step on a unit (Maximum Health printed − 3, power unchanged, lethal at 0 and destroyed at the sweep after it resolves, restored after 4c); a positive one; `UntilEndOfOpponentTurn` on a max-health step lasts through the opponent's turn; one `max-health` event per asset with `Target`, `Amount`, `Player` -1 and a face-down asset hidden (extend `TestEffectEvents_FaceDownAssetStaysHidden`'s table in `effect_events_test.go:104-110` only if the task names that edit); `TestEventKinds_ValuesAreTheSchema` gains the new kind (a named edit); a filter with `Stat: StatMaxHealth, AtMost: n` matches by effective Maximum Health on the battlefield and by printed Health in a discard pile; existing power steps (no `Stat`) unchanged.
- `go tool task sim-compare -- autoloop/m2-wave-h` passes: no starter or stress card uses it.

## Log
- 2026-11-15 lead-architect (wave H run, session 12): queued, split off T-0251 under the D2 size rule (the whole max-health shape touches six non-test files). Waits for T-0251.
