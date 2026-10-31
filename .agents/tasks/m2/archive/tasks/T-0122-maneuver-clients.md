---
id: T-0122
title: The view, viewtest and the play screen show an enacted Maneuver (directive #89, harness, flash)
kind: harness
tier: flash
wave: C
status: done
branch: m2/T-0122-maneuver-clients
depends_on: [T-0116, T-0121]
claims: [view:stackView StackManeuver, view:stackKinds Maneuver, viewtest:PublicNames StackManeuver, play-screen:actionLabel Enact, play-screen:stack line Maneuver]
cards: []
requirements: [FR-MNV-001, FR-ZON-005]
---

## Goal
Before the first starter-deck Maneuver is executable (T-0120), every client of the engine's state shows an enacted Maneuver: `view.PlayerView` gives its stack item a kind, its card and its chosen targets; `viewtest.PublicNames` counts its card as public while it is on the stack; and the play screen labels the Enact action and the stack line with the card's name. This is directive #89 (a)–(d). No engine code changes.

## Inputs
- Directive #89 in `BOARD.md` (T-0116's Clients, less what T-0121 took).
- `pkg/view/view.go`: `StackView` (lines 75-87), `stackKinds` (lines 89-92, no `StackManeuver` label), `stackView` (lines 159-193: only the `StackAbility` branch fills `Card`, `Chosen` and `ChosenCards`).
- `pkg/view/viewtest/viewtest.go` `PublicNames` (lines 24-28): only `StackDeploy` and `StackAbility` items' cards are public.
- `cmd/cardtown/play_screen.go`: the stack lines in `printScreen` (lines 79-87, `"  %s (Player %d, Source %d)\n"`), and `actionLabel` (lines 112-169: no `ActEnact` case, so it falls back to `a.Kind.String()`, a bare "Enact").
- `pkg/engine/stack.go`: a `StackManeuver` item has `Inst` (the card), `Effect`, `Chosen`, and `Source` 0 (line 27); `resolveManeuver` (lines 110-116) puts the card into its owner's discard pile with no event.
- `pkg/engine/actions.go` line 88: `Apply` for `ActEnact` pushes `&StackItem{Kind: StackManeuver, Controller, Inst, Effect, Chosen}`.
- Existing view tests to mirror: `TestView_Case1_Contents` (stack items built by hand, lines 148-151 and 261-272), `TestView_PileTargetShowsCardNamesNotInstanceIDs` (lines 708-), and the synthetic Maneuver built in `pkg/sim/policy_test.go` `TestProgress_Case6_EnactsManeuvers` (T-0121) for a real enact.
- FRD 01 `FR-MNV-001` (a Maneuver resolves, then goes to its owner's discard pile) and `FR-ZON-005` (the stack is public).

**Directive #89 (d), decided here: no stack-to-discard event.** The view is a snapshot of state: `side` reads both discard piles from `Player.Discard`, so after the Maneuver resolves its card is in its owner's `Discard` in both seats' views without any event (case 5 checks this). The log already names the card twice (the Enact action event since T-0121, and the `place` event "player N places Maneuver X on the stack"), and `eventlog.Replay` rebuilds the Enact from its action event (`actionOf` reads `Card` from `Inst`), so it needs nothing either. A death sweep's move to the discard pile has its own events because it is a rules event (`FR-ZON-004`); a resolved Maneuver's move is not.

## Acceptance Cases (written by the lead; the engineer turns each into a test)
1. **A Maneuver stack item shows its kind, card and asset targets.** Scene (built by hand, as `TestView_Case1_Contents` does): the stack holds one `StackManeuver` item, controller 0, `Inst` a synthetic Maneuver instance named "Alpha Volley" that is in no zone, `Effect.Targets` one asset slot, `Chosen` `[[<an opposing unit's asset ID>]]` → in both seats' `PlayerView`, the one `StackView` has `Kind` "Maneuver", `Controller` 0, `Source` 0, `Card` "Alpha Volley", `Chosen` `[[<that ID>]]`, `ChosenCards` `[[]]`, `Keyword` "", `Zone` "", `Target` 0. New test in `pkg/view/view_test.go` (for example `TestView_ManeuverItemShowsItsCardAndTargets`).
2. **A Maneuver's pile slot shows card names, as an Ability's does.** Scene: as `TestView_PileTargetShowsCardNamesNotInstanceIDs`, but the item is a `StackManeuver` with `Inst` set and `Source` 0, and one of the two chosen pile cards has left the discard pile → `Chosen` for that slot is `[]`, `ChosenCards` holds only the name of the card still in a discard pile, and no instance ID appears in either seat's view. Same new test or its own.
3. **One fill, no copy.** The per-slot `Chosen` / `ChosenCards` loop in `stackView` (lines 177-190) moves into one function that the `StackAbility` branch and the new `StackManeuver` branch both call. The Ability branch keeps its own `Keyword` and face-down source-card logic; the Maneuver branch takes `Card` from `it.Inst.Card.Name`. Every existing test in `pkg/view` passes unchanged. (Checked at review against the diff: a second copy of the loop is a `changes` note.)
4. **Comments say Maneuver.** `StackView`'s field comments name the Maneuver: `Kind` lists "Maneuver"; `Source` says it is 0 for a Maneuver; `Card` says "Maneuver: the card on the stack"; `Chosen` and `ChosenCards` say "Ability, Maneuver". `stackKinds` maps `engine.StackManeuver` to "Maneuver".
5. **After resolution, the card is in its owner's discard pile in both views, with no new event.** Scene: player 0, in its main phase with the resources to pay, enacts a supported synthetic Maneuver with one untargeted step (for example `Draw 1`, as in T-0121's case 1) through `g.Apply` → afterwards `PlayerView(g, 0).You.Discard` and `PlayerView(g, 1).Opponent.Discard` both end with that card, both views' `Stack` is empty, and player 0's `You.HandCount` is unchanged (the Maneuver left the hand, the step drew one). `LogLines` for each seat has at least one line containing the card's name. No `engine` file changes. Same test file (for example `TestView_EnactedManeuverEndsInItsOwnersDiscard`).
6. **viewtest counts a Maneuver on the stack as public.** Scene: the stack holds one `StackManeuver` item whose card is named "Alpha Volley" and is in no pile and not on the battlefield → `viewtest.PublicNames(g)["Alpha Volley"]` is true, and `viewtest.HiddenNames(g, 0)` does not contain it. Before the change it is false. The `PublicNames` doc comment is unchanged in meaning ("the cards on the stack"). Test in `pkg/view/view_test.go` (for example `TestViewtest_ManeuverOnTheStackIsPublic`).
7. **The play screen labels the Enact action.** `actionLabel(v, engine.Action{Kind: engine.ActEnact, Player: 0, Card: <inst ID of "Alpha Volley" in v.You.Hand>})` returns `Enact: Alpha Volley`. With an instance ID not in the hand it returns `Enact: Card#<id>`, as Deploy's fallback does. Every other kind's label is unchanged. New test in `cmd/cardtown/play_test.go` (for example `TestPlay_ActionLabelNamesTheEnactedCard`).
8. **The play screen's stack line names a Maneuver's card.** `printScreen` with a `view.View` whose `Stack` is `[{Kind: "Maneuver", Controller: 0, Card: "Alpha Volley"}, {Kind: "Attack", Controller: 1, Source: 7, Target: 3}]` prints, under `Resolution Stack:`, exactly the lines `  Maneuver: Alpha Volley (Player 0)` and `  Attack (Player 1, Source 7)`, in that order. A Maneuver line has no `Source` (a Maneuver is not an asset). Every other kind's line is unchanged. Same new test file.
9. **The sims are unchanged.** No starter Maneuver is executable on this branch and no engine code changes: `go tool task sim-compare -- autoloop/m2-wave-c` is byte-identical.

## Allowed Files
- `pkg/view/view.go`
- `pkg/view/viewtest/viewtest.go`
- `cmd/cardtown/play_screen.go`

Tests: `pkg/view/view_test.go` and `cmd/cardtown/play_test.go`, new tests only. No existing test is edited.

## Done When
- [x] Every acceptance case has a test (case 3 is checked by the existing view tests and at review, case 4 at review, case 9 by `sim-compare`), and each new test failed before the implementation commit.
- [x] `go test ./pkg/view/...`
- [x] `go test ./cmd/cardtown`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-compare -- autoloop/m2-wave-c` passes (starter and stress sims byte-identical to the run branch)

## Log
- 2026-10-30 lead-architect: wave C run, session 8. No `main` commits to sync. Directive #89 is `open: before T-0120`, so it is this prompt's task (D2.0), ahead of T-0120. Duplicates searched: `stackView`'s Ability branch is the one place that fills a stack item's chosen targets, so the Maneuver branch shares it (case 3); `PublicNames` is the one public-name oracle; `actionLabel`'s `findHandCard` already names Deploy's card. #89 (d) decided above: no new event, the view reads discard piles from state. `eventlog.Replay` already rebuilds an Enact (`actionOf`), so it isn't touched. Three non-test files, all in the packages that own them (`pkg/view` the client view, `cmd/cardtown` the play screen). Tier `flash`: small, no engine change, no random draws. Harness, so `ready`; branch `m2/T-0122-maneuver-clients` from `autoloop/m2-wave-c`.
- 2026-10-30 kernel-engineer: acceptance tests added in test(T-0122) commit db9b261 (all cases covered; cases 1, 2, 6, 7, 8 verified failing prior to implementation). Implemented in feat(view) commit 5c75b01:
  - pkg/view/view.go: updated StackView comments (naming Maneuver, Source 0 for Maneuver, Card, Chosen/ChosenCards), added engine.StackManeuver to stackKinds map, extracted chosenTargets helper to share target slot resolution across StackAbility and StackManeuver without code duplication (case 3), wired StackManeuver to populate Card and Chosen/ChosenCards.
  - pkg/view/viewtest/viewtest.go: included engine.StackManeuver in PublicNames stack item loop.
  - cmd/cardtown/play_screen.go: added engine.ActEnact case to actionLabel reusing findHandCard, updated printScreen to format Maneuver stack line as "  Maneuver: %s (Player %d)\n" without Source.
  - Commands run and verified:
    - go test ./pkg/view/... (PASS)
    - go test ./cmd/cardtown (PASS)
    - go tool task check (gofmt, go vet, all tests, cardtown lint --all all PASS)
    - go tool task sim (PASS)
    - go tool task sim-compare -- autoloop/m2-wave-c (PASS: byte-identical starter and stress sims)
  - Worktree cleaned up. Task ready for review.
- 2026-10-30 lead-architect: review round 1, approved, merged into `autoloop/m2-wave-c` (cb08f45). Checked against the diff: only the three Allowed Files and the two test files changed, and no test changed after `test(T-0122)` (db9b261). Case 3: the per-slot loop moved into `chosenTargets`, which the Ability and Maneuver branches both call, with no second copy; the Ability branch keeps its `Keyword` and face-down logic. Case 4: the `StackView` comments and the `stackKinds` label are in. Cases 7 and 8: `actionLabel` reuses `findHandCard`, and only the Maneuver stack line changed. Re-run by the reviewer: at db9b261 the tests for cases 1, 2, 6, 7 and 8 fail. Case 5's test passes there too, as intended: it pins #89 (d), that the view needs no new event. At 6bf1592 `go tool task check` passes, `go tool task sim` is clean (200 games, P0 67.0% / P1 32.5% / draws 0.5%, no monitor violation), and `go tool task sim-compare -- autoloop/m2-wave-c` passes. Directive #89 applied. T-0120 is no longer behind #89: its branch takes in the run branch before its 8305cd5 is reviewed.
