---
id: T-0270
title: "Sim damage monitor: a redirected combat hit is still the exchange's attacker-to-defender instance (wave H harness, flash)"
kind: harness
tier: flash
wave: H
status: done
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
- FRD 01: `FR-CBT-002` (an exchange's two Combat damage instances, each raw = the dealer's effective power), `FR-REP-004` Stage 1 (target redirection; Stages 2 to 5 name the new target, the source is unchanged).
- The event T-0269 logs (`pkg/engine/replace.go:149-166`, `redirectDamage`, run branch at d8309229): `Event{Kind: EventReplace ("replace"), Player: <redirector's controller>, Source: <redirecting asset>, Target: <original target>}`, logged inside `dealDamage` (`pkg/engine/damage.go:22-27`) right before that instance's `damage` event, whose `Target` is the redirecting asset and whose `Source` and `Raw` are unchanged. When several replacements apply, a `decision` event comes before the `replace` event, never between it and its damage event. A destruction replacement (`replaceDeath`, `replace.go:111`) logs the same kind; it happens in the death sweep, after an exchange's two damage events, so it is never inside an open exchange.
- `pkg/sim/monitor.go:148-160`, the event switch in `Monitors.Event` (no `replace` case today).
- `pkg/sim/monitor_damage.go:1-52`: `combatPair`, `openCombat`, `checkCombatDamage`. The Armor check in `checkDamage` (`:66-76`) already reads `e.Target`, so a redirected event is mitigated against the redirector's pool with no change there.
- Test helpers in `pkg/sim/monitor_damage_test.go:20-68` (`damageDecks`, `powered`, `dmgEvent`, `combatEvent`, `damageCase`) and `pkg/sim/monitor_test.go:97-130` (`feed.event`, `wantViolation`, `wantClean`, `wantCounts`).

## Shape
- `combatPair` gains the pending redirect, the original target's ID and the redirecting asset's ID (both 0 when none). A `replace` event while an exchange is open (`left > 0`) records it; outside one it does nothing. The next damage event of the exchange consumes it, whether it matches or not.
- `checkCombatDamage`: the attacker half matches `Source == att` and a Target equal to `def`, or to the redirecting asset when the pending redirect's original target is `def`; the defender half the same with `def` and `att`. The raw check against the dealer's power is unchanged.
- No new coverage rule ID and no change to `monitorRules`: the starter and stress sims must stay byte-identical (no executable card redirects damage until T-0271).
- Not this task: a redirect of non-combat damage needs nothing (the non-combat path never pairs targets).

## Acceptance Cases (harness: the engineer writes these as one new test, `TestDamage_RedirectedCombat` in `pkg/sim/monitor_damage_test.go`, through `damageCase`; every case is one sub-check in it)
Board for every case: `baseSnap(0, 0, powered(10, "Vanilla", 0, 3, 0), powered(11, "Vanilla", 1, 2, 0), powered(12, "Plated", 1, 2, 2), powered(13, "Vanilla", 0, 3, 0))`. `rep(src, tgt)` below is `engine.Event{Kind: "replace", Player: -1, Source: src, Target: tgt}` (a local helper in the test file is fine).
1. **Redirected attacker half:** `combatEvent(10, 11)`, `rep(12, 11)`, `dmgEvent(10, 12, 3, true, 2, 2, 1)`, `dmgEvent(11, 10, 2, true, 0, 0, 2)` → `wantClean`, and `wantCounts` `FR-CBT-002` and `FR-KWD-002k` (the redirected hit is mitigated by 12's Armor).
2. **Redirected defender half:** `combatEvent(10, 11)`, `dmgEvent(10, 11, 3, true, 0, 0, 3)`, `rep(13, 10)`, `dmgEvent(11, 13, 2, true, 0, 0, 2)` → `wantClean`.
3. **Redirected hit with the wrong raw:** `combatEvent(10, 11)`, `dmgEvent(11, 10, 2, true, 0, 0, 2)`, `rep(12, 11)`, `dmgEvent(10, 12, 5, true, 2, 2, 3)` → `wantViolation(..., "damage-sequence", damageRules, <seq of the last event>)`, and the violation's `Detail` contains `want the dealer's effective power 3` (the raw check, not the pairing check).
4. **No replace event, no redirect (today's behavior):** `combatEvent(10, 11)`, `dmgEvent(10, 12, 3, true, 2, 2, 1)` → `wantViolation` at that event's seq, `Detail` containing `is not the attacker 10 / defender 11 exchange`.
5. **A redirect from another target doesn't apply:** `combatEvent(10, 11)`, `rep(12, 13)`, `dmgEvent(10, 12, 3, true, 2, 2, 1)` → `wantViolation` at the damage event's seq, `Detail` containing `is not the attacker 10 / defender 11 exchange`.
6. **A redirect is consumed by the next damage event, matched or not:** `combatEvent(10, 11)`, `rep(12, 10)`, `dmgEvent(10, 11, 3, true, 0, 0, 3)`, `dmgEvent(11, 12, 2, true, 2, 2, 0)` → `wantViolation` at the last event's seq, `Detail` containing `is not the attacker 10 / defender 11 exchange` (the redirect of 10 was used up by the unredirected attacker half, so it can't pair the next event; the engine never logs this order, but the rule is "the next damage event").
7. **A replace event outside an exchange opens nothing:** `rep(12, 11)`, `dmgEvent(10, 12, 3, false, 2, 2, 1)` → `wantClean`; then, in a second feed, `rep(12, 11)`, `dmgEvent(10, 12, 3, true, 2, 2, 1)` → `wantViolation` at that damage event's seq, `Detail` containing `outside a combat exchange`.

Cases 1, 2 and 7's first half fail before the implementation (1 and 2 report a pairing violation today; 7's first half is clean today too, so it guards against a regression only, which is fine inside the one test function).

## Allowed Files
- `pkg/sim/monitor.go` (the `replace` case in the event switch)
- `pkg/sim/monitor_damage.go`

## Done When
- [x] The acceptance cases' tests pass: `go test ./pkg/sim -run Damage`
- [x] `go tool task check` passes
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-compare -- autoloop/m2-wave-h` passes

## Log
- 2026-11-18 lead-architect (wave H run, session 31): queued with T-0269, split off under the D2 size rule (T-0269 already has four non-test files). Depends on T-0269 for the `replace` event's shape. Goes before T-0271, so the card never runs under a monitor that misreads it.
- 2026-11-18 lead-architect (wave H run, session 32): Inputs, Shape and Acceptance Cases written against the run branch at d8309229; status `ready`, dispatched to Flash. Branch created from the run branch.
- 2026-11-18 kernel-engineer: added TestDamage_RedirectedCombat in pkg/sim/monitor_damage_test.go covering acceptance cases 1-7 (confirmed failing on cases 1, 2, 3 before implementation and committed as test(T-0270)). Added recordReplace in pkg/sim/monitor_damage.go and wired engine.EventReplace in pkg/sim/monitor.go event switch. combatPair records pending redirection (redOrig, redTo) when left > 0, and checkCombatDamage consumes the pending redirection and allows matching redirected attacker or defender damage to the redirecting asset while verifying the dealer's power. Ran go test ./pkg/sim -run Damage (pass), go tool task check (pass), go tool task sim (clean), and go tool task sim-compare -- autoloop/m2-wave-h (pass, starter and stress sims match byte-for-byte). Status set to review.
- 2026-11-18 lead-architect (wave H run, session 32): reviewed `git diff autoloop/m2-wave-h...` and approved, round 1. Scope: `monitor.go`, `monitor_damage.go` and the new `TestDamage_RedirectedCombat` in the `test(T-0270)` commit only; all seven acceptance cases are in it, and at f099275d it fails (cases 1, 2, 3). Asset IDs start at 1 (`state.go:193`), so 0 is a safe "no redirect" value. Re-ran in a review worktree: `go tool task check` (exit 0), `go tool task sim` (clean, P0 58.5%, P1 39.5%, draws 2.0%), `go tool task sim-compare -- autoloop/m2-wave-h` (byte-identical). Merged as 85a4ee7. Next: T-0271, *Animoceps Spore Cloud*.
