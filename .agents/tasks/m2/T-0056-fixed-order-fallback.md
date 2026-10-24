---
id: T-0056
title: No random fallbacks - a timed-out target picks the first legal options in fixed order (SVY-R16, directive #47 part 3, shape, sonnet)
kind: shape
tier: sonnet
wave: B
status: tests
branch: m2/T-0056-fixed-order-fallback
depends_on: [T-0054]
claims: [fallback:DecideTarget, helper:fixed-order]
cards: []
requirements: [FR-MOD-001]
---

## Goal
No `FR-MOD-001` §6 fallback is random. A mandatory target choice that times out, has no chooser or gets an invalid answer picks the first `d.Min` legal options in §6's **fixed order** (`SVY-R16`), not random ones, so a player can always predict what a timeout does. This is board directive #47 part 3. It is the only part of #47 that contradicts merged code. Parts 1, 2 and 4 go to the tasks that build their cards. Part 5 went into the design note with this task.

## Inputs
- FRD 01 v3.11+, `FR-MOD-001` §6 (line 352), "Mandatory Targets and Untargeted Selections": *"Selects the required number of legal options, the first ones in **fixed order**. Assets: the choosing player's own assets first, then the opponent's. Within each player, the Headquarters first, then Reserves, then the Main Theater, each zone in the order its assets entered it (as `SVY-R7`). … An "up to N" selection selects none."* Revision summary v3.10 → v3.11, item 3.
- Rulings: `SVY-R16` (survey line 75), `SVY-R11` ("up to N": options oldest first, a timeout picks none; unchanged).
- Design note: `docs/architecture/m2-ability-model.md`, Decisions (the ADR-0002 line, rewritten for this task) and Open Question 8(b).
- Board directive #47 part (3).
- Existing code: `fallbackPicks` and `randomPicks` in `pkg/engine/choice.go`, and `chooseTargets` in `pkg/engine/targeting.go`, which lists options oldest asset first (`assetsByID`). `Game.Assets` in `state.go` walks the HQs, each player's `Reserves` and `g.Theater`. `placeOnBattlefield` appends to those slices, so slice order is arrival order.

### Lead decisions (technical)
1. **Options keep their listed order.** `chooseTargets` still lists legal options oldest first, by ID. `TestFR_TGT_002_UpToOptionsAreOldestFirst` and many `refs(d)` checks pin that order. The fallback applies the fixed order itself. So a `Chooser` that answers sees the same `Decision` as today, and `cardtown sim`, whose seats always answer through `RandomChooser`, should not change.
2. **`fallbackPicks` for `DecideTarget`** returns the indexes of the first `d.Min` options in fixed order, in that order. It draws nothing from `g.rng`. `randomPicks` stays, because `RandomChooser`, a bot policy, uses it. `RandomChooser` is unchanged.
3. **Fixed order, relative to `d.Player`, for an option whose `Ref` is a battlefield asset:**
   - `d.Player`'s assets come before the opponent's.
   - Within each player: the Headquarters, then that player's `Reserves` in slice order, then the assets in `g.Theater` that the player controls, in slice order.
   - "Own" means controlled by the player, not owned.
   - Options whose `Ref` is not a battlefield asset come after those that are, in listed order. Synthetic decisions (`opts(n)`, `Ref` 100+) therefore fall back to the first listed options, like `DecideMode`.
4. **Where the helper lives** is the engineer's call: `choice.go`, `targeting.go` or `state.go`. One function gives an asset's fixed-order position for a choosing player. Untargeted picks (Animoceps Spores, Mutation Swell, Infested Apostle's kill) will reuse it. Choosing from your own hand (Chainbreaker Irregulars) will have its own rule, the card most recently added, in that card's task. Neither is built here.
5. **Comments:** the `DecideTarget` comment in `choice.go` and the `chooseTargets` doc comment in `targeting.go` stop saying "random".
6. **"Up to N"** is unchanged: `Min` is 0, so the fallback picks none (`SVY-R11`).

### What the tests can observe (for rules-test-author)
Use a synthetic trigger with one chosen-target `DealDamage`, answered with a timeout (`s.script(p, Answer{Timeout: true})`) unless stated otherwise:
- **Own before opponent's:** the filter allows either player's units. The opponent's unit is older (lower ID) than the choosing player's, and both are in the Main Theater. The timeout damages the choosing player's unit.
- **Headquarters, then Reserves, then Main Theater:** the filter allows any kind (or unit plus Headquarters). The choosing player's Main Theater unit is older than their Reserves unit. The timeout picks the Headquarters. With Headquarters excluded by the filter, it picks the Reserves unit over the older Main Theater unit.
- **Within a zone by arrival, not ID:** an older unit moved into the Main Theater (`s.moveTo`) after a newer unit was put there comes second, so the timeout damages the newer unit.
- **Relative to the choosing player, not seat 0 or the active player:** the same layout with the trigger controlled by player 1, on player 0's turn, picks player 1's asset first.
- **Opponent's assets in the same order:** with only opposing assets legal, the timeout picks the opponent's Headquarters before their Reserves, and their Reserves before their Main Theater.
- **No chooser and an invalid answer** give the same pick as a timeout.
- **No randomness drawn:** two scenes with the same seed. In one, a target fallback fires. Afterwards both draw the same next value from the game's random source, such as a shuffle of identical decks. Optional; drop it if it can only be written against internals.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
New tests in `pkg/engine/fallback_order_test.go`. Names may be refined; record the final regex here.
- `TestFR_MOD_001_TargetTimeoutPicksOwnAssetBeforeOpponents`
- `TestFR_MOD_001_TargetTimeoutPicksHeadquartersThenReservesThenMainTheater`
- `TestFR_MOD_001_TargetTimeoutOrdersAZoneByArrivalNotByID`
- `TestFR_MOD_001_TargetTimeoutOrderIsRelativeToTheChoosingPlayer`
- `TestFR_MOD_001_TargetTimeoutOrdersOpponentsAssetsTheSameWay`
- `TestFR_MOD_001_TargetNoChooserAndInvalidAnswerPickLikeATimeout`
- `TestFR_MOD_001_TargetFallbackDrawsNoRandomness` (optional, see above)

**Authorized edits to existing tests** (in the `test(T-0056)` commit only):
- `pkg/engine/choice_test.go`:
  - Replace `TestFR_MOD_001_TargetFallbackIsRandomNotFixed` with `TestFR_MOD_001_TargetFallbackPicksFirstInFixedOrder`. It uses a synthetic decision (`opts(6)`): `Min` 1 gives `[0]` and `Min` 2 gives `[0 1]`, on every call.
  - **Delete** `TestFR_MOD_001_TargetFallbackDeterministicForSameSeed`. A fallback with no randomness has no seed to depend on, and the replacement asserts the exact picks, so it covers more.
  - `TestFR_MOD_001_TargetFallbackWithNoChooserPicksLegalTarget` and `TestFR_MOD_001_TargetFallbackOnTimeoutPicksLegalDistinctTargets`: reword the "random" messages. You may tighten them to the exact picks (`[0]`, and `[0 1]`).
  - Remove imports that are no longer used.
- **Message-only rewording**, with assertions unchanged. Each says "a timeout picks one random legal target". Make it "the first legal target in fixed order (FR-MOD-001 sec. 6, SVY-R16)":
  - `TestFR_TGT_001_TimeoutOnTwoOptionsDamagesExactlyOneOfThem` (`targeting_test.go`);
  - `TestCard_Saboteur_TimeoutOnOpposingChoiceDamagesExactlyOneLegalUnit`, `TestCard_ChainbreakerSappers_TimeoutDamagesExactlyOneLegalTarget`, `TestCard_DevilhornGuerrillas_TimeoutDamagesExactlyOneLegalUnit` and `TestCard_NecropolisHunt_TimeoutDamagesExactlyOneLegalTarget` (`cards_path_test.go`);
  - the comment in `TestFR_MOD_001_UpToTimeoutAndNoChooserPickNone` (`optional_upto_test.go`).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
None expected. Request one here if a test needs it.

## Allowed Files
- `pkg/engine/choice.go`
- `pkg/engine/targeting.go`
- `pkg/engine/state.go`

## Done When
- [ ] The listed tests pass, along with the edited ones: `go test ./pkg/engine -run 'TestFR_MOD_001_Target|TestFR_TGT_001_TimeoutOnTwoOptions|TestCard_.*_Timeout'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean. `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` output before and after is byte-identical, or differs only in games where a fallback fired: a `decision` event with `how` `timeout` or `invalid`. Record the `cmp` in the Log.
- [ ] Any other existing test that fails is a stop, not an edit. Set `blocked` and name it (a test that relied on a seed's random pick).

## Log
- 2026-10-23 lead-architect: Wrote the task for board directive #47 part 3, with status `tests` for `rules-test-author`. Applied #47 part 5 and #48 part 5 to the design note in the same commit.
  - The directive is ordered ahead of T-0055 (user, 2026-10-23).
  - The trigger of directive #48 has happened: FRD 01 v3.12, the Scratch Company CSV row, and the outcome `Status: applied`. Its task, T-0057, is written and stays `queued`.
  - The other waiting triggers have not fired. #8: no `deck.go` change. #15: no turn hook. #20: this is wave B.
  - Starter-deck cards not yet executable: 28, plus both Headquarters.
