---
id: T-0078
title: Make Arkha, Sun's Fury executable (wave B card, flash)
kind: card
tier: flash
wave: B
status: review
branch: m2/T-0078-arkha
depends_on: [T-0077, T-0079]
claims: [card:Arkha, Sun's Fury]
cards: [Arkha, Sun's Fury]
requirements: []
---

## Goal
*Arkha, Sun's Fury* (Path unit, one copy in the Path starter deck) runs in the kernel. It is plain `ability` data in `pkg/engine/cards/path.go`. Its one rules-text line is:
- "Whenever this unit attacks, you may put a unit card from your discard pile on the bottom of your deck. If you do, deal 3 damage to chosen opposing Headquarters."

## Inputs
- Card (`data/cards/units-and-structures-path.csv`, line 30): Path unit, Officer, Devilhorn, Mythic, deploys to the Main Theater, with the line above.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 151. `SVY-R14`: the put is not a cost but an untargeted pick on resolution.
- Shapes it uses: `HookAttacks` and `Trigger.Optional` (A1, A2), a target slot (A2), `OpDealDamage`, `OpPutOnDeck` with `Bottom`, `SubjectPick` with `UpTo` and `Step.IfDone` (T-0072), and the pile pick with `ability.ZoneDiscard` (T-0077).
- Rulings: `SVY-R9` ("you may" is answered on placement), `SVY-R3` (the Headquarters target is chosen then), `SVY-R14` (the put happens on resolution, and a timeout or no unit card means no damage), `SVY-R16`.

### Definition (the engineer adds it to the Path defs)
- One trigger: `Hook: ability.HookAttacks`, `Optional: true`.
- Its `Effect`:
  - `Targets`: one slot, `Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindHeadquarters}` (exactly one).
  - Step 1: `{Op: ability.OpPutOnDeck, Bottom: true, Subject: ability.SubjectPick, UpTo: 1, Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}}`.
  - Step 2: `{Op: ability.OpDealDamage, Subject: ability.SubjectTarget, Target: 0, Amount: ability.Value{N: 3}, IfDone: true}`.
- `Text` holds the printed line.
- Remove "Arkha, Sun's Fury" from `cards.NotYetExecutable`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written in `pkg/engine/cards_path_arkha_test.go`, from the CSV text:
- `TestCard_ArkhasSunsFury_PutsAUnitCardOnTheBottomAndDamagesTheOpposingHeadquarters`: on an attack, "you may" and the opposing Headquarters are asked on placement and the pick on resolution. The picked unit card is the bottom card of Arkha's controller's deck, and the opposing Headquarters takes 3 damage.
- `TestCard_ArkhasSunsFury_OnlyUnitCardsInYourDiscardPileAreOffered`: a non-unit card in that pile and a unit card in the opponent's pile are not offered.
- `TestCard_ArkhasSunsFury_NoUnitCardMeansNoDamage`: with no unit card in the pile, nothing is picked and the Headquarters takes no damage from the trigger.
- `TestCard_ArkhasSunsFury_TimeoutAtThePickDealsNoDamage`: a timeout at the pick moves nothing and deals nothing.
- `TestCard_ArkhasSunsFury_DecliningDoesNothing`: declining "you may" asks nothing more and does nothing, and the attack goes on.

## Helpers Needed
None. (The test file defines its own `arkhaChooser` and `newArkhaScene`.) Originally: none expected. `inDiscard` comes with T-0077.

## Allowed Files
- `pkg/engine/cards/path.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run 'TestCard_ArkhasSunsFury'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-10-27 lead-architect (wave run B, session 13): Wrote the task, `queued` behind T-0077 (its pile pick). Split from T-0077 by the size rule (D2.4).
- 2026-10-27 lead-architect (wave run B, session 14): Synced `main` into the run branch first (2312e28).
  - Directives: no `open: before` directive is on the board. Waiting triggers checked against `autoloop/m2-wave-b` 2312e28: #6 and #7 (mode is `serial`), #8 (no `deck.go` change), #15 (no turn hook), #20 (b) (no wave C shape task), #58 (no exhaust by a cost or an effect) and #59 (2) (no repeated targets). None has fired.
  - Every shape the Definition names exists on the run branch: `ability.ZoneDiscard`, `OpPutOnDeck` with `Bottom`, `SubjectPick` with `UpTo`, `Step.IfDone`, `KindHeadquarters`, `AllegianceOpposing`. `inDiscard` is in the test helpers. No new shape, so this stays a card task.
  - Claims don't overlap any open task, since none is open. Status `tests`; dispatching `rules-test-author`.
- 2026-10-27 rules-test-author: Wrote `pkg/engine/cards_path_arkha_test.go` with the five listed tests: PutsAUnitCardOnTheBottomAndDamagesTheOpposingHeadquarters, OnlyUnitCardsInYourDiscardPileAreOffered, NoUnitCardMeansNoDamage, TimeoutAtThePickDealsNoDamage, DecliningDoesNothing. They fail now because the card is not yet executable. No helpers requested.
  - Deviation from the task text: the Headquarters target is not asserted as a question. There is exactly one opposing Headquarters, so FR-MOD-001 chooses it without asking (as in the Infested Apostle tests). The tests expect asked kinds `[Optional, Pick]`.
  - Arkha attacks an opposing 0/20 dummy so all Headquarters damage comes from the trigger. The timeout test relies on the `UpTo: 1` pick selecting none on a timeout (FR-MOD-001 sec. 6, SVY-R16).
- 2026-10-27 lead-architect (tests-review, round 1): **Changes, back to `tests`.** I checked the tests against `SVY-R14` (FRD 01 line 330), `SVY-R16` and `FR-MOD-001` §1 and §6 (lines 350, 359), and the CSV line.
  - **Right:** the asked kinds `[Optional, Pick]` (one opposing Headquarters is chosen by the single-option fast path, §1, so the deviation from the task text is correct); the picked card on the bottom of the deck and the other one left in the pile; only your own unit cards offered; no unit card asks no pick and deals nothing; an "up to" timeout selects none, so nothing moves and nothing is dealt (§6); declining asks nothing more.
  - **Note:**
    1. **Wrong expected damage in `TestCard_ArkhasSunsFury_DecliningDoesNothing`.** `s.wantDamage(a.dummy, 6)` expects 6, but Arkha's printed power is 8 (`data/cards/units-and-structures-path.csv` line 30, `power` column; checked by loading the card in a scratch test: power 8, the dummy takes 8). Expect 8, or read it from `a.arkha.Printed().Power`. Change nothing else.
- 2026-10-27 rules-test-author (round 1 fix): Note 1 done. `DecliningDoesNothing` now expects the dummy's damage to equal `a.arkha.Printed().Power` (8) instead of 6. Nothing else changed.
- 2026-10-27 lead-architect (tests-review, round 2): **Approved, `ready`.** Note 1 is done: 10b30dc changes only the dummy's expected damage, now `a.arkha.Printed().Power` (8). No other test changed since 7a9602d. Next: dispatch `kernel-engineer` (tier `flash`) headless.
- 2026-10-27 kernel-engineer: BLOCKED on `go tool task check` (TestView_Case4_NoLeak).
  - Implementation: added ability definition for "Arkha, Sun's Fury" to `pathDefs` in `pkg/engine/cards/path.go` (Trigger: HookAttacks, Optional: true, Targets: slot 0 Filter{Allegiance: AllegianceOpposing, Kinds: KindHeadquarters}, Step 1: OpPutOnDeck, Bottom: true, SubjectPick, UpTo: 1, Filter{Allegiance: AllegianceAllied, Kinds: KindUnit, Zone: ZoneDiscard}, Step 2: OpDealDamage, Amount: Value{N: 3}, SubjectTarget: slot 0, IfDone: true; Text: printed CSV line). Removed "Arkha, Sun's Fury" from `NotYetExecutable` in `pkg/engine/cards/cards.go` while keeping alphabetical sort order.
  - Reused: HookAttacks, Trigger.Optional, Effect.Targets, Filter (AllegianceOpposing, KindHeadquarters), OpPutOnDeck, Bottom, SubjectPick, UpTo, Filter (AllegianceAllied, KindUnit, ZoneDiscard), OpDealDamage, SubjectTarget, IfDone. No new shapes or engine logic added.
  - Verification:
    - `go test ./pkg/engine -run 'TestCard_ArkhasSunsFury'` PASS (5/5 tests pass).
    - `go test ./pkg/engine -run 'TestStarterSet_Executable'` PASS.
    - `go test ./pkg/engine` PASS (all engine unit tests pass).
    - `cardtown lint --all` PASS (clean).
    - `go tool task sim` PASS (200 games ANIMUS vs PATH starter playout clean, all invariants hold, Path executable cards 24/50 -> 25/50).
  - Block cause:
    - `go tool task check` runs `go test ./...` which fails at `pkg/view/view_test.go:537` (`TestView_Case4_NoLeak`):
      `view_test.go:537: seed 13, settled point 61 (turn 11): seat 0's log names "Hammerhand Brigands", hidden by player 1: 122 T11 main decision: player 1, Pick, "Arkha, Sun's Fury: pick up to 1": picked [Hammerhand Brigands] (answer)`
    - `TestView_Case4_NoLeak` checks that a seat's new log lines do not name cards hidden in the opponent's hand or deck, unless they are currently in `PublicNames(g)` (discard piles, face-up battlefield assets, or stack). When Arkha's trigger resolves, Player 1 picks a unit card from their discard pile to put on the bottom of their deck. The decision event `decision: player 1, Pick, "Arkha, Sun's Fury: pick up to 1": picked [...]` is emitted while the card is in the discard pile, but the card is immediately moved to the deck by `OpPutOnDeck`. At the subsequent settled point (point 61), the card is in Player 1's deck and no longer in `PublicNames(g)`, so `TestView_Case4_NoLeak` treats the event log line as a leak of a card in the opponent's deck.
    - Resolving this failure requires updating `pkg/view/view_test.go` (or `pkg/view/viewtest` / `pkg/view`), which are outside `Allowed Files` for T-0078. Status set to `blocked`.
- 2026-10-27 lead-architect (wave run B, session 14, D6): **Requeued, `queued` behind T-0079.** `agy` conversation `cdea209f-bdf6-4bb9-bb89-eac6390e88f0` finished with no denied action.
  - The block is a fault in the leak oracle, not in this task. The picked card was in the public discard pile when the `decision` line was written (FRD 01 line 176, `SVY-R14`). `TestView_Case4_NoLeak` checks the line only at the next settled point, when the card is already on the bottom of the deck. T-0079 makes the check also accept a name that was public at the line's own event.
  - 4123001 matches the Definition: the trigger, its target slot, both steps and `Text` in `path.go`, and the name removed from `NotYetExecutable`. Only the two Allowed Files changed, and the tests are unchanged since 10b30dc.
  - Next, once T-0079 is `done`: merge the run branch into this branch (T-0079 changes only `pkg/view/view_test.go`), set `review`, and review 4123001 at D5 with every gate re-run. The approved tests stand, so no new test round is needed.
- 2026-10-27 lead-architect (wave run B, session 16, D5): **Approved.** T-0079 is `done` (bea61cf), so I merged the run branch into this branch (6d21110, no conflict) and set `review`. The approved tests stand, so there was no new test round.
  - Directives: no `open: before` directive is on the board, and `main` has nothing the run branch lacks. Waiting triggers checked against `autoloop/m2-wave-b` 996f0c9: #6 and #7 (mode is `serial`), #8 (no `deck.go` change), #15 (no turn hook), #20 (b) (no wave C shape task), #58 (no exhaust by a cost or an effect) and #59 (2) (no repeated targets). None has fired.
  - **Scope:** `git diff autoloop/m2-wave-b...HEAD` changes only `pkg/engine/cards/path.go`, `pkg/engine/cards/cards.go`, this task file and the test file. `cards_path_arkha_test.go` is unchanged since 10b30dc.
  - **Fences and reuse:** the definition is plain `ability` data that matches the Definition: the trigger, its opposing-Headquarters slot, the bottom-of-deck pile pick and the `IfDone` damage step. It adds no shape. No existing helper builds an opposing-Headquarters slot, so there is nothing to reuse.
  - **Rules fidelity:** "you may" is answered on placement (`SVY-R9`) and the target is locked then (`SVY-R3`). The put is an `UpTo: 1` pick on resolution, and no pick means no damage (`SVY-R14`, `SVY-R16`).
  - **Gates, re-run by the reviewer in a review worktree (6d21110):** `go tool task check` exits 0, which includes the 5 card tests, `TestStarterSet_Executable` and `TestView_Case4_NoLeak`. `go tool task sim` runs clean: P0 65.0% / P1 34.0% / draws 1.0% (was 70.5% / 28.5% / 1.0%), every invariant holds and the monitors found no violation.
