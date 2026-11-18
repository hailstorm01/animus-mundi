---
id: T-0270
title: "Sim damage monitor: a redirected combat hit is still the exchange's attacker-to-defender instance (wave H harness, flash)"
kind: harness
tier: flash
wave: H
status: queued
branch: m2/autoloop-wave-h/T-0270-monitor-redirected-combat-damage
depends_on: [T-0269]
claims: [monitor:checkCombatDamage accepts a combat damage event redirected by the replace event before it]
cards: []
requirements: [FR-CBT-002, FR-REP-004]
---

## Goal
`pkg/sim`'s damage monitor (`monitor_damage.go`) expects a combat exchange's two damage events to be attacker → defender and defender → attacker. Once T-0269 lands, a combat instance that a `WouldBeDealtDamage` replacement redirects is logged as a `replace` event (Source: the redirecting asset, Target: the original target) followed by a `damage` event whose Target is the redirecting asset. The monitor must take that event as the redirected half of the exchange, still checking its raw amount against the dealer's power. Without this, any sim that plays *Animoceps Spore Cloud* (T-0271; the Animus stress deck takes every executable Animus card, `pkg/sim/sim.go:25-53`) reports a false `FR-CBT-002` violation.

Covers: the monitor reads a `replace` event inside an open combat exchange and remembers the redirect for the next damage event; `checkCombatDamage` matches `Source == att && Target == redirectedTo(def)` as the attacker half (and the same for the defender half); a `replace` event outside a combat exchange (a destruction replacement, or a non-combat redirect) changes nothing.

## Inputs
To be written against the code of the day when this task leaves `queued`: the exact `replace` event T-0269 logs, `monitor.go`'s event switch (`:150-173`, run branch at e008353c), `monitor_damage.go` (`combatPair`, `openCombat`, `checkCombatDamage`, `:1-52`), and the feed helpers in `monitor_damage_test.go` (`dmgEvent`, `combatEvent`, `damageCase`) and `monitor_test.go`.

## Acceptance Cases
To be written when the task leaves `queued`. At least:
1. combat (att 1, def 2), replace (Source 9, Target 2), damage 1 → 9 combat with raw = att's power, damage 2 → 1 combat: clean.
2. The same with the redirected event's raw not the attacker's power: an `FR-CBT-002` violation at that event.
3. combat (1, 2), then damage 1 → 9 combat with no replace event before it: still a violation, as today.
4. A replace event outside a combat exchange, then a non-combat damage event: clean, and no combat exchange is opened.

## Allowed Files
- `pkg/sim/monitor.go` (the `replace` case in the event switch)
- `pkg/sim/monitor_damage.go`

## Done When
- [ ] The acceptance cases' tests pass: `go test ./pkg/sim -run Damage`
- [ ] `go tool task check` passes
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes

## Log
- 2026-11-18 lead-architect (wave H run, session 31): queued with T-0269, split off under the D2 size rule (T-0269 already has four non-test files). Depends on T-0269 for the `replace` event's shape. Goes before T-0271, so the card never runs under a monitor that misreads it.
