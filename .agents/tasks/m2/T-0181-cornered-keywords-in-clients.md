---
id: T-0181
title: The view shows a keyword behind Cornered, and the monitors stay clean on it (directive #117 (a), (c), harness, flash, tests only)
kind: harness
tier: flash
wave: D.1
status: queued
branch: m2/autoloop-wave-d.1/T-0181-cornered-keywords-in-clients
depends_on: [T-0180]
claims: [test:view shows a keyword behind Cornered, test:monitors clean on a keyword behind Cornered]
cards: []
requirements: []
---

## Goal
The two cases of #117 that `rules-test-author` can't write, because it writes only under `pkg/engine` (D2.4, as for T-0164). A seat's view lists a keyword behind Cornered only while Cornered, and a real game with a `Cornered: Resistant 1` card breaks no rule monitor. No non-test code is expected to change.

## Covers (Inputs, Acceptance Cases and Allowed Files are written against the code of the day, before this task leaves `queued`)
- **`pkg/view`:** an `AssetView`'s `Keywords` shows `Aggressive 3` for a unit with `Cornered: {Aggressive: 3}` while its controller is Cornered, and doesn't when it isn't (#117 (a)).
- **`pkg/sim`:** simulated games of a deck holding a synthetic `Cornered: {Resistant: 1}` unit finish with no monitor violation (#117 (c)). The reviewer breaks a damage event's `Resistant` in `dealDamage`, since reverted, to see the test fail, because a test-only task has no `failed before` to check.
- `go tool task sim-compare -- autoloop/m2-wave-d.1` passes.

## Log
- 2026-11-05 lead-architect (wave D.1 run, session 11): split from T-0180 (D2.4). Status `queued`.
