---
id: T-0229
title: "One read of which pile an action names: Action.pile() (directive #145 (b), wave F.1 harness, flash, pure move)"
kind: harness
tier: flash
wave: F.1
status: queued
branch: m2/autoloop-wave-f.1/T-0229-action-pile
depends_on: [T-0228]
claims: [func:Action.pile, func:pileCard (ZoneDeck case), func:emitAction (pile read)]
cards: []
requirements: []
---

## Goal
"A Deploy takes its card from `a.From`, an Enact from the hand" is decided in one place, `func (a Action) pile() Zone`, and `pileCard` no longer reads `ZoneDeck` (`Action.From`'s zero value) as the hand (#145 (b)). A pure move: same behavior, same test names, same sim output. Split from T-0228 (#145 (a)) by the D2.4 size rule.

## Covers (write the Inputs, Edits, Acceptance Cases and Allowed Files against the code of the day before this leaves `queued`)
- `Action.pile()` in `pkg/engine/actions.go`. **Lead correction to the directive's wording, to keep it a pure move:** #145 says `pile()` returns `ZoneDiscard` for a Deploy from the discard pile "else `ZoneHand`". That would make a Deploy whose `From` names no pile (for example `ZoneMainTheater`) take its card from the hand, which `TestFR_TRN_003_DeployFromAnotherZoneIsIllegal` rejects today. So `pile()` returns `ZoneHand` for an Enact and for a Deploy whose `From` is `ZoneDeck` (the zero value), and a Deploy's `From` as it is otherwise; `pileCard` finds no card in a zone that is not a pile, as now. The lead dry-ran this at abe2ea2 with `handCard` already gone: `go tool task check` exit 0, `sim-compare` `[PASS]`, `cardtown lint fences` exit 0.
- `Apply` (Deploy and Enact), `check` (Deploy's `canPlay`, its permission check and its `deployPlacement`; Enact's `canPlay`) and `emitAction` (`pkg/engine/events.go`) call `a.pile()`; `emitAction` still writes `From` only for a discard deploy (`ZoneDiscard.String()`), so events are unchanged. Afterwards `git grep -n 'a\.From' -- pkg/engine ':!*_test.go'` finds only `pile()`'s own lines.
- `pileCard` (in `zones.go` after T-0228) drops its `ZoneDeck` case and its comment says `Action.pile` says which pile an action names.
- `doc.go`'s Enums paragraph names `Action.From` as the one field whose zero value reads as the hand, read only by `Action.pile`; `Action.From`'s comment points at `pile()`.
- Allowed files (four non-test files): `pkg/engine/actions.go`, `pkg/engine/zones.go`, `pkg/engine/events.go`, `pkg/engine/doc.go`.
- Acceptance as in #145: no test changes; the same test names pass; `go tool task check`; `go tool task sim-compare -- autoloop/m2-wave-f.1`.

## Log
- 2026-11-13, lead-architect (wave F.1 run, session 2): queued with T-0228, which takes #145 (a). Written out in full (Inputs, exact Edits, Acceptance Cases) when T-0228 is `done`.
