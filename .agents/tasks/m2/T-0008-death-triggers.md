---
id: T-0008
title: Fire death triggers from the death sweep, placed in APNAP order
kind: shape
tier: sonnet
wave: A1 (board directive #4, task 2b)
status: in-progress
branch: m2/T-0008-death-triggers
depends_on: [T-0007]
claims: [hook:killed, func:fire-killed, invariant:pending-buffer-empty, invariant:no-invalid-answer]
cards: []
requirements: [FR-TRG-003, FR-TRG-002, FR-ZON-004]
---

## Goal
A card can print "When this is killed, …". Every asset destroyed by one death sweep fires its death triggers together. The flush then places them in APNAP order: the active player's first, the non-active player's on top, so the non-active player's resolve first. Each player orders their own, the non-active player through their own chooser. Playouts now also fail if a trigger is left in the buffer between actions, or if a chooser answer is rejected.

## Inputs
- FRD 01:
  - `FR-TRG-003` (lines 333–337 of `docs/frd/01_Game_Rules_FRD.md`): both bullets, now including the cross-player half.
  - `FR-TRG-002` (lines 330–331).
  - `FR-TRG-004` (line 340).
  - `FR-ZON-004` (lines 156–162): the leaves-battlefield trigger bullet, including "a face-down Scratch Company … never triggers printed leaves-battlefield death abilities".
  - Destruction (line 154) and death sweep timing (`FR-LAY-002`, line 381).
- Built on: T-0007 (`pkg/engine/trigger.go`: the buffer, `flushTriggers`, `DecideOrder`, `StackAbility`, `OpScratch`, `OpDraw`) and T-0006 (choices).
- Board directive #10: the ordering fallback is collection order. For one sweep that means battlefield order (`g.Assets()`), then each card's printed order.
- Design note: `docs/architecture/m2-ability-model.md`, Test Plan ("Fuzz": the new invariants).
- Rulings that apply: none beyond directive #10.

### Vocabulary (`pkg/engine/ability/ability.go`)
```go
const (
	HookDeploy Hook = iota + 1
	HookKilled // this asset was destroyed from the battlefield (FR-ZON-004 leaves-battlefield trigger)
)
```

### Lead decisions (technical, not rules)
- **Where they fire:** `deathSweep` fires each dying asset's `HookKilled` triggers before the asset leaves the battlefield. It reads `a.Printed().Ability` and uses `a.Controller` as the trigger's controller, which is the asset's last-known state on the battlefield. A face-down asset fires nothing (`FR-ZON-004`).
- **Collection order (directive #10):** `deathSweep` already builds `dying` in `g.Assets()` order, so firing in that loop gives battlefield order, then each card's printed order. The flush must keep that order within each player.
- **One sweep is one event:** every death trigger from one sweep goes to the same flush. That flush places the active player's triggers first and the non-active player's on top (T-0007's `flushTriggers`, unchanged).
- **Match over:** if the sweep ended the match (an HQ destroyed, or an empty-deck draw), the buffer is discarded as in T-0007, and no death trigger resolves.
- **Source gone:** a death trigger resolves even though its source has left the battlefield (T-0007's decision: untargeted steps verify nothing).
- **Effective state is T-0009:** death triggers read the printed `Def`. Nothing can grant or remove abilities yet, so printed and last-known effective abilities are the same today. T-0009 replaces this read with the last-known effective state, for when grants exist and for values like last-known power.
- **Turn-start sweep:** `beginTurn`'s sweep cannot destroy an asset (it only settles empty-deck draws), so it does not flush. The new buffer invariant would catch any death trigger left behind there.
- **Invariants (`playout.go`):**
  - `CheckInvariants` reports a non-empty pending buffer between actions, as it already reports a non-empty stack.
  - `Playout` returns an error naming the turn and the logged detail when an action adds an `invalid-answer` event to `Game.Log`. The design note's rule is that a chooser answer is never rejected in a playout.

### What the tests can observe
- **Cross-player order:** a combat in which the attacker and defender kill each other gives one death trigger to each player. Both decks are empty and both triggers are `Draw 1`, so whoever resolves first loses at that trigger's sweep. The non-active player resolving first means the active player wins.
- **Non-active player ordering (`FR-TRG-002`):** a Berserk attacker kills the defender and the secondary target, both the non-active player's, in one sweep. The non-active player is asked one `DecideOrder` through their own chooser, and the active player is not asked. Order shows on distinct named cards on top of that player's deck, as in T-0007.
- **Collection order:** with no chooser, two simultaneous deaths of one player resolve as directive #10 says, so the earlier asset in battlefield order is placed first and resolves last.
- **Playout failure:** a playout whose chooser answers an ordering question invalidly returns an error. Build it from synthetic decks whose cards print two deploy triggers, and a policy that deploys.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
All in `pkg/engine/death_trigger_test.go` (needs only `ability.HookKilled`; `go vet` fails on nothing else).
- `TestFR_TRG_003_NonActivePlayersDeathTriggerResolvesFirst`, `TestFR_TRG_003_NonActiveIsRelativeToTheTurn`: mutual kill, empty decks, Draw 1 each; the non-active player loses, from either seat.
- `TestFR_TRG_002_NonActivePlayerOrdersTheirOwnDeathTriggers`: Berserk kills two of player 1's assets; player 1 gets one `DecideOrder` (Player 1, 2 options), player 0 none.
- `TestFR_TRG_003_NonActivePlayersOrderAnswerIsApplied`: answers [1 0] and [0 1] change which trigger resolves first.
- `TestFR_TRG_002_BothPlayersWithTwoTriggersAreEachAskedOnceAboutTheirOwn`: each seat asked once about its own two; no order of questions pinned.
- `TestFR_TRG_003_DeathTriggersOfOneAssetFallBackToPrintedOrder`, `TestFR_TRG_003_SimultaneousDeathsFallBackToBattlefieldOrder`: directive #10 with no chooser.
- `TestFR_TRG_004_DeathTriggersAreOrderedAfterTheSweepAndResolveWithoutTheirSource`: at ordering time the source is gone and nothing has resolved; both resolve after.
- `TestFR_TRG_004_SingleDeathTriggerResolvesWithoutQuestion`
- `TestFR_TRG_004_MatchEndedBySweepResolvesNoDeathTrigger`: attacker and nearly dead HQ die together; no trigger resolves.
- `TestFR_ZON_004_FaceDownScratchCompanyFiresNoDeathTrigger`: face-down unit fires nothing and is not asked; face-up control fires.
- `TestPlayout_InvalidOrderAnswerIsAnError`, `TestPlayout_ValidAndTimedOutOrderAnswersAreNotErrors`: synthetic two-deploy-trigger decks, deploy policy.
- Left out: the pending-buffer-empty invariant (not observable from outside). The face-down test sets `FaceDown` on a Main Theater asset via `put`, since Reserves assets cannot be attacked.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. `s.onTop`, `s.script`, `s.asked`, `s.put`, `s.attack` and `s.cleave` exist.

## Allowed Files
- `pkg/engine/ability/ability.go` (`HookKilled`)
- `pkg/engine/trigger.go` (firing `HookKilled` triggers into the buffer)
- `pkg/engine/damage.go` (the call from `deathSweep`, before each asset leaves)
- `pkg/engine/playout.go` (the two invariants)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_TRG_00[234]|TestFR_ZON_004|TestPlayout'`
- [ ] T-0006 and T-0007 tests still pass unchanged
- [ ] `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean
- [ ] `go run ./cmd/cardtown lint --all` passes
- [ ] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0, and plain `sim` output is unchanged from the integration branch

## Log
- 2026-10-20 lead-architect: Wrote the task (board directive #4, task 2b). Applies directive #10 to deaths in one sweep, settling the open point in T-0007's review: fire in `deathSweep`'s battlefield-order loop, with no sort in the flush. Set `status: tests` for `rules-test-author`. Don't pin which player is *asked* first when both have two or more triggers. FRD 01 fixes placement order, not the order of the questions.
- 2026-10-20 rules-test-author: Wrote the T-0008 tests in `pkg/engine/death_trigger_test.go` (listed under Tests). No helpers requested. Left out: the pending-buffer invariant (unobservable externally). Not pinned: which player is asked first when both have 2+ triggers. Option indexing in the Berserk ordering test assumes collection order (battlefield order, then printed), per directive #10.
- 2026-10-20 lead-architect: Test review: **approved**, status `ready`. The tests match FR-TRG-002, FR-TRG-003, FR-TRG-004 and the leaves-battlefield bullet of FR-ZON-004, plus directive #10. The empty-deck "who loses" construction is a clean observation of cross-player APNAP, and it is checked on both players' turns. The Berserk constructions rely on Theater `put` order (Berserker, Defender, Bystander), which is battlefield order. The face-down case has a face-up control. The playout controls end by decking, because every deploy draws and scratches. Setting `s.g.choosers[1]` directly is setup only. The pending-buffer invariant can't be seen from outside, so it is reviewed in code. Next: `kernel-engineer` (tier sonnet).
