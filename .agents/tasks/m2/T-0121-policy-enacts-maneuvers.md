---
id: T-0121
title: The progress policy enacts Maneuvers, and the coverage tests count them (harness, sonnet)
kind: harness
tier: sonnet
wave: C
status: done
branch: m2/T-0121-policy-enacts-maneuvers
depends_on: [T-0116]
claims: [func:ProgressPolicy, event-field:action.Card for Enact]
cards: []
requirements: [FR-MNV-001]
---

## Goal
Once a starter-deck Maneuver is executable, `cardtown sim` plays it and the two coverage tests can see it. Today `ProgressPolicy` (`pkg/sim/policy.go`) never picks `ActEnact` and never generates resources for a Maneuver, `TestProgress_Case4_Coverage` counts only `deploy` events, and `TestEvents_Case4_EveryActionKind` expects a uniform random playout of 10 seeds to enact a starter Maneuver. T-0120 (*Despoiler's Emergence*, a 1-of costing 4 Tactics) blocked on both tests: its Flash run (`agy` conversation `ae65880f-af81-403f-94b6-d0a499e26f01`) found `Despoiler's Emergence deployed 0 times in 1000 games` and `no action event for Enact`. These are gaps in the lead's spec for T-0116, not in T-0120.

## Inputs
- `pkg/sim/policy.go` `ProgressPolicy` (lines 24-71): prefers a Deploy, then a Generate or Convert closing a deployable hand card's shortfall (the scan skips cards with no `DeployZones`, which every Maneuver has), then Attack, March or Scratch.
- `pkg/sim/policy_test.go` `TestProgress_Case4_Coverage` (lines 117-144) and its `deployCounter` (lines 146-153).
- `pkg/engine/events_test.go` `TestEvents_Case4_EveryActionKind` (lines 306-334), with T-0116's `starterManeuver` exemption.
- `pkg/engine/events.go` `emitAction` (lines 141-163): an `ActDeploy` action event carries `Card` and `Inst`; an `ActEnact` one carries neither.
- FRD 01 `FR-MNV-001` (lines 332-334): a Maneuver is enacted from Hand, paying its costs on declaration.

## Acceptance Cases (written by the lead; the engineer turns each into a test)
1. **Enact is a play from hand.** `ProgressPolicy`'s first preference becomes "a play from hand": uniform over every legal `ActDeploy` and `ActEnact` together. Scene: player 0 has an empty hand plus one affordable synthetic Maneuver with an untargeted step (for example `Draw 1`), no deployable card, and a readied asset → for seeds 1..8 the policy returns `ActEnact` for that card. In a new test in `pkg/sim/policy_test.go` (for example `TestProgress_Case6_EnactsManeuvers`); `TestProgress_Case2_Preferences` is unchanged.
2. **Both kinds are offered together.** Scene: one affordable unit and one affordable Maneuver in hand → over seeds 1..16 the policy returns both `ActDeploy` and `ActEnact` at least once, and nothing else. Same new test.
3. **Resources close a Maneuver's shortfall.** Scene: the only hand card is a supported Maneuver costing 2 Tactics, the pool is 0/0, and a readied asset yields Tactics → for seeds 1..8 the policy returns `ActGenerate` of Tactics. A Maneuver the kernel can't execute (`!Supported()`) closes nothing, as an unsupported unit doesn't today. Same new test.
4. **Nothing changes while no Maneuver is executable.** On this branch no starter-deck Maneuver is executable, so when a turn offers no `ActEnact`, the policy makes the same random draws as before (no extra `r` call): `go tool task sim-compare -- autoloop/m2-wave-c` is byte-identical.
5. **The Enact action event names its card.** `emitAction` fills `Card` and `Inst` for `ActEnact` as it does for `ActDeploy` (the card is still in hand when the event is emitted). Scene in a new test in `pkg/engine/events_test.go` (for example `TestEvents_Case4_EnactActionEventNamesItsCard`): enacting a synthetic Maneuver logs exactly one `action` event with `Detail` "Enact", `Player` 0, and the card's `Card` and `Inst`, and it comes before the Maneuver's `place` event.
6. **Authorized edit to `TestEvents_Case4_EveryActionKind`.** Remove the `starterManeuver` exemption and its loop, and skip `ActEnact` unconditionally, with the comment: Enact is covered by `TestEvents_Case4_EnactActionEventNamesItsCard`, because a uniform random playout of 10 seeds doesn't reliably reach a 1-of Maneuver. Every other kind is still checked. This edit goes in the `test(T-0121)` commit.
7. **Authorized edit to `TestProgress_Case4_Coverage`.** `deployCounter` becomes a play counter: it counts a `deploy` event by `Card` as now, and an `action` event whose `Detail` is "Enact" by its `Card` (case 5). The log line reads `%-40s %d plays`, and the error `%s played %d times in 1000 games, want at least 10`. Every supported starter-deck card, Maneuvers included, still needs at least 10. This edit goes in the `test(T-0121)` commit. (With no executable starter Maneuver on this branch, it passes as before; T-0120 is where a Maneuver first reaches it.)
8. `ProgressPolicy`'s doc comment lists the new first preference ("a play from hand: a Deploy or an Enact, uniform over card and zone").

## Allowed Files
- `pkg/sim/policy.go`
- `pkg/engine/events.go` (the Enact branch of `emitAction` only)

Tests: `pkg/sim/policy_test.go` and `pkg/engine/events_test.go` (new tests, plus the two authorized edits above).

## Done When
- [x] Every acceptance case has a test, and each new or edited test failed before the implementation commit (case 4 is checked by `sim-compare`, cases 6 and 7 by the edited tests passing).
- [x] `go test ./pkg/sim -run TestProgress`
- [x] `go test ./pkg/engine -run TestEvents_Case4`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-compare -- autoloop/m2-wave-c` passes (starter and stress sims byte-identical to the run branch)

## Log
- 2026-10-30 lead-architect: wave C run, session 6, D6. Written for T-0120, which blocked on `TestProgress_Case4_Coverage` and `TestEvents_Case4_EveryActionKind` once *Despoiler's Emergence* became the first executable starter Maneuver (the Flash run finished with no denied action). Its implementation commit (8305cd5) passes its own ten tests and `go tool task sim`; only these two harness tests fail, and both need files outside its Allowed Files. Duplicates searched: `ProgressPolicy` is the only policy that ranks actions (`RandomPolicy` already picks Enact uniformly); `emitAction` is the one action-event writer. Tier `sonnet`: the policy must keep its random draws unchanged when no Enact is legal (case 4). Two non-test files. `queued`; T-0120 is requeued behind it. Wave run: `STOP: blocked task`, as wave B's T-0078/T-0079 (row 14).
- 2026-10-30 lead-architect: wave C run, session 7. Synced `main` (directive #89, merge 4e2336a). Inputs re-checked against the run branch (`policy.go`, `emitAction`): unchanged. #89 (the view and play screen, `open: before T-0120`) does not overlap these claims, so it is written after this task. Harness, so `ready`; dispatched to `kernel-engineer` (sonnet) on `m2/T-0121-policy-enacts-maneuvers`, branched from `autoloop/m2-wave-c`.
- 2026-10-30 kernel-engineer: test commit first (new `TestProgress_Case6_EnactsManeuvers`, `TestEvents_Case4_EnactActionEventNamesItsCard`, the two authorized edits); the new tests failed before the implementation (16 policy failures, Enact event had no Card/Inst). Implementation: `emitAction` fills Card and Inst for `ActEnact` (same as Deploy); `ProgressPolicy` merges Deploy and Enact into one `plays` list (no extra `r` draw when no Enact is legal) and the shortfall scan now includes supported Maneuvers. Reuse: `handCard`, existing Deploy branch shape. Ran: `go tool task check` pass, `go tool task sim` clean, `go tool task sim-compare -- autoloop/m2-wave-c` byte-identical. Case 3's unsupported check asserts only that no Generate/Enact is returned (random Attack/March can still be picked).
- 2026-10-30 lead-architect: review round 1, approved, merged into `autoloop/m2-wave-c` (62bac04). Checked against the diff: cases 1–3 in `TestProgress_Case6_EnactsManeuvers`, case 5 in `TestEvents_Case4_EnactActionEventNamesItsCard` (one Enact action event with `Card` and `Inst`, before the `place` event), cases 6 and 7 are the authorized edits in `test(T-0121)` (6869f13) and nowhere later, case 8 in the doc comment. Only Allowed Files and their tests changed; the `emitAction` change is its Enact branch only. Case 4: with no `ActEnact` in `legal`, `plays` holds exactly the old `deploys` and the shortfall scan adds only supported Maneuvers (none on this branch), so the draws are unchanged. Reuse: the Enact branch repeats Deploy's two-line `handCard` lookup; it's too small to count as a duplicate under D2.0. Re-run by the reviewer: at 6869f13 both new tests fail (policy picks March or Scratch, the event has no `Card`); at d0756e6 `go tool task check` passes, `go tool task sim` is clean (200 games, P0 67.0% / P1 32.5% / draws 0.5%) and `go tool task sim-compare -- autoloop/m2-wave-c` passes. The engineer's note that case 3's unsupported-Maneuver check can't assert EndTurn is accepted: the policy may still pick an Attack or March there, and the check that matters (no Generate, no Enact) holds.
