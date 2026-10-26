---
id: T-0066
title: Simultaneous triggers fall back to source position, not step order (directive #57, shape, sonnet)
kind: shape
tier: sonnet
wave: none (board directive #57, FRD 01 review finding 1)
status: tests
branch: m2/T-0066-trigger-source-order
depends_on: [T-0065]
claims: [order:simultaneous-trigger-fallback, field:engine.Asset.arrived]
cards: []
requirements: [FR-TRG-003, FR-MOD-001]
---

## Goal
When a player doesn't order their own simultaneous triggers, they are placed by where each source stood when it fired: the Headquarters, then Reserves, then the Main Theater, each zone in arrival order. Within one source, the existing order stays: keyword triggers first, then printed order. Today the fallback is the order the triggers were buffered. That is right for the death sweep, which walks the battlefield in source order, but wrong for triggers a step fires as it visits assets: an "all" step visits oldest first, and a multi-target step visits in pick order.

## Inputs
- FRD 01 `FR-MOD-001` §6, "Ordering Own Simultaneous Triggers (`SVY-R7`, `FR-TRG-003`)" (line 358); `FR-TRG-003` (APNAP placement).
- Rulings that apply: `SVY-R7` (default trigger order), `SVY-R13` and board directive #46 (lethal Unbowed is placed in one window with the death triggers), `SVY-R11` (an "all" step acts on assets oldest first: **unchanged**; this task changes only where the triggers it causes are placed, not the order damage is dealt).
- Board directive #57.
- Kernel today (at `main` ab86a44):
  - `fallbackPicks` for `DecideOrder` is the identity, i.e. buffer order (`choice.go:134`).
  - The death sweep fires `fireKilled` in `Assets()` order, then removes each dying asset one at a time (`damage.go:72`, `damage.go:93-94`).
  - `dealDamage` fires `fireDamaged` as it hits each asset (`damage.go:60`), so an "all" step (`effects.go:140`, `assetsByID`) buffers Unbowed triggers oldest first, and a multi-target step buffers them in pick order.
  - Every test helper (`scene.put`, `scene.putScratch`) places assets through `placeOnBattlefield`.
- The lead prototyped this shape in a throwaway worktree: every existing test passed unchanged, and `go tool task sim-compare` passed (no starter or stress match reaches the divergent case today).

### Shape (lead decisions)
- **Why not sort by current position at the flush:** a dead source has left the battlefield by then, and the death sweep removes assets one at a time, so zone indexes shift between firings. The position is recorded when the trigger fires, using an arrival number that never shifts.
- **`Asset.arrived int`** (unexported, in `state.go`), commented `// arrival number: when it entered its current battlefield zone, counted across the match (SVY-R7)`. It is not added to `AssetSnap`.
- **`Game.arrivals int`** (unexported, in `state.go`, beside `nextID`): the counter.
- **`placeOnBattlefield`** (`zones.go`) increments `g.arrivals` and sets `a.arrived` before placing the asset. A March places the asset again, so a marched asset arrives in its new zone last.
- **`sourcePos`** in `trigger.go`:
  ```go
  // sourcePos is where a trigger's source stood when it fired: its zone's rank
  // (Headquarters 0, Reserves 1, Main Theater 2) and its arrival number. A
  // player's simultaneous triggers fall back to this order (SVY-R7, FR-MOD-001 §6).
  type sourcePos struct{ zone, arrived int }
  ```
  with `func posOf(a *Asset) sourcePos` and `func (p sourcePos) compare(q sourcePos) int` (zone first, then arrival; use `cmp.Compare`).
- **`pendingTrigger.pos sourcePos`**, set from `posOf(a)` at both places a trigger is buffered: `fireHook` and the Aggressive keyword trigger in `fireAttacks`.
- **`flushTriggers`** stable-sorts each player's `mine` by `pos` (`slices.SortStableFunc`) before it builds the order decision. The stable sort keeps one source's keyword trigger ahead of its printed ones, and printed order. The decision's options are listed in this order, so the fallback (identity) and a chooser's indexes both refer to source order. Its doc comment adds one sentence: `Each player's triggers are listed, and by default placed, in source order (sourcePos).`
- **`state.go` battlefield-orders comment:** the `Assets` paragraph becomes exactly:
  ```go
  // Assets: the Headquarters, then every player's Reserves, then the Main
  // Theater, each zone in arrival order. The death sweep uses it. Simultaneous
  // triggers are placed in the same order, by where each source stood when it
  // fired (sourcePos in trigger.go), so a step that visits assets in another
  // order doesn't change it (SVY-R7, FR-TRG-003). Uses where order doesn't
  // matter (lookups, readying, invariants, legal actions) also use it.
  ```

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Written in a new file `pkg/engine/trigger_order_test.go`, with synthetic cards only. Each case runs with no chooser (or a timeout) unless it says otherwise, and observes placement by resolution order: the trigger placed first resolves last (`FR-MOD-001` §6). In each case the IDs (age) and the source order must disagree, or the test proves nothing; say so in a comment.
- `TestFR_TRG_003_AllStepTriggersFallBackToSourcePositionNotAge`: two allied Unbowed units, the older one in the Main Theater (it marched there) and the newer one in Reserves. An "all" damage step hits both, and both survive. The Reserves unit's trigger is placed first and resolves last.
- `TestFR_TRG_003_MarchingIsAnArrival`: both allied Unbowed units in the Main Theater. The older one was deployed to Reserves and marched in after the newer one was deployed directly to the Main Theater. An "all" damage step hits both. The newer unit (first to arrive in the Main Theater) has its trigger placed first.
- `TestFR_TRG_003_MultiTargetTriggersFallBackToSourcePositionNotPickOrder`: one effect damages two chosen allied Unbowed units, with the picks scripted in the opposite of source order. The fallback for the trigger order still follows source position.
- `TestFR_TRG_003_DeathAndUnbowedTriggersFallBackToSourcePosition`: in one flush, the same player has an Unbowed survivor and a dying ally with a death trigger. Another asset ahead of the dying one in its zone dies in the same sweep, so zone indexes shift during the sweep. The two triggers follow source position (`SVY-R13`, directive #46: both in one window).
- `TestFR_TRG_003_OrderDecisionListsTriggersInSourceOrder`: with a recording chooser, the order decision's options come in source order, not buffer order, and an answer's indexes refer to that list.
- `TestFR_TRG_003_OneSourcesKeywordTriggerStaysAheadOfItsPrintedTriggers`: an Aggressive attacker with a printed attack trigger, plus another allied source whose trigger fires in the same window from an earlier source position. Source order between the two sources, and keyword before printed within the attacker. If FRD 01 offers no way to fire a second allied trigger in that window with existing shapes, leave this case out and say so in the Log.
- **Left out:** ordering across players (APNAP, already pinned); exhaust triggers from a step over several assets (no effect exhausts yet).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
None expected. The test file defines its own synthetic cards and reuses `scene.put`, `scene.script` and the existing march and damage helpers. If a helper is missing, list it here and stop at `tests-review`.

## Allowed Files
- `pkg/engine/state.go` (`Asset.arrived`, `Game.arrivals`, the battlefield-orders comment)
- `pkg/engine/zones.go` (`placeOnBattlefield` sets the arrival number)
- `pkg/engine/trigger.go` (`sourcePos`, `posOf`, `compare`, `pendingTrigger.pos`, the sort in `flushTriggers`)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_TRG_003_(AllStepTriggers|MarchingIsAnArrival|MultiTargetTriggers|DeathAndUnbowedTriggers|OrderDecisionListsTriggers|OneSourcesKeywordTrigger)'`
- [ ] Every existing test passes unchanged (`git diff main -- '*_test.go'` shows only the new file).
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks.
- [ ] `go tool task sim-compare` passes. The lead's prototype matched byte for byte. If yours doesn't, record each difference in the Log and explain why it is only a trigger order.

## Log
- 2026-10-26 lead-architect: written from directive #57, after T-0065 (`newAsset`). Prototyped in a throwaway worktree to check the blast radius: all existing tests and both sims unchanged. Status `tests`; dispatching `rules-test-author`.
- 2026-10-26 rules-test-author: tests written in `pkg/engine/trigger_order_test.go` (all fail now, each on the order assertion; setup preconditions pass). Placement is read from the match log's `place` events (the order FR-MOD-001 sec. 6 defines; the first placed resolves last), using a Scratch trigger on each Unbowed unit. Tests:
  - `TestFR_TRG_003_AllStepTriggersFallBackToSourcePositionNotAge`
  - `TestFR_TRG_003_MarchingIsAnArrival` (uses a real `ActMarch`)
  - `TestFR_TRG_003_MultiTargetTriggersFallBackToSourcePositionNotPickOrder` (own chooser picks older then newer; order question times out)
  - `TestFR_TRG_003_DeathAndUnbowedTriggersFallBackToSourcePosition` (a plain unit ahead of the dying one also dies; survivor older and in the Main Theater)
  - `TestFR_TRG_003_OrderDecisionListsTriggersInSourceOrder` (options checked by label, pick [1,0] then checked by placement)
  - Helpers requested: none. The file defines `pingedBowed`, `allySweeper`, `placedOrder`, `wantPlaced`, `olderInTheaterNewerInReserves`.
  - Left out: `TestFR_TRG_003_OneSourcesKeywordTriggerStaysAheadOfItsPrintedTriggers`. Every trigger hook is a self hook, so no existing shape fires a second allied source's trigger in the Aggressive attacker's Step 2 window; keyword-before-printed within one source is already pinned by `TestFR_CBT_002_AggressiveAndAttackTriggerAreOrderedTogether`. Also left out as the task says: cross-player order and exhaust triggers.
- 2026-10-26 lead-architect (tests-review, round 1): **Back to `tests`.** Four of the five tests are right as written: correct expected outcomes against FRD 01 §6 line 358, setup preconditions asserted, and age and source order disagree in each. Reading `place` events is fine: they are the match log's structured events (ADR 0005), as `event_facts_test.go` uses them. Leaving out the keyword-ordering case is accepted: every hook is a self hook, and `TestFR_CBT_002_AggressiveAndAttackTriggerAreOrderedTogether` already pins keyword before printed. Independence checked: the transcript reads only the task, FRD 01, test files, `go doc` and `ability` types. Notes:
  1. `TestFR_TRG_003_DeathAndUnbowedTriggersFallBackToSourcePosition`: keep the test, but fix its comment. The survivor (Main Theater) and the dying unit (Reserves) are in different zones, so the shifting Reserves indexes never change their relative order. Drop the claim that the index shift is tested here. Keeping the `Plain` unit is fine.
  2. Add `TestFR_TRG_003_SweepRemovalsDoNotReorderTriggersInOneZone`, the case the task's test (c) was meant to pin (the lead's spec put it badly). One player, all in Reserves, put in this order: `PlainA` and `PlainB` (1/1, no ability, both die), then `Survivor` (`pingedBowed`), then `Dying` (`mortal`, 1/1, death trigger Scratch 1). Then deploy `allySweeper()`. Assert the setup (PlainA, PlainB and Dying die; Survivor lives), then `wantPlaced(..., []*Asset{survivor, dying}, survivor, dying)`: Survivor entered Reserves before Dying, so its trigger is placed first. This test **passes on main today**, because buffer order already matches. Say so in its comment: it guards the rule against a fallback that reads zone positions as the sweep removes assets. After the two Plain units leave, Dying would stand ahead of where Survivor stood when its Unbowed fired.
