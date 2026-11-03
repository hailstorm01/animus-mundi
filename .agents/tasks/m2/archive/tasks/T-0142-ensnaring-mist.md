---
id: T-0142
title: Make Ensnaring Mist executable (move a chosen unit to its controller's Reserves, then a chosen unit gets -1 power until end of turn; wave C card, flash)
kind: card
tier: flash
wave: C
status: done
branch: m2/autoloop-wave-c/T-0142-ensnaring-mist
depends_on: [T-0056, T-0069, T-0107, T-0116, T-0117, T-0119, T-0135]
claims: [card:Ensnaring Mist]
cards: [Ensnaring Mist]
requirements: [FR-MNV-001, FR-ZON-002, FR-LAY-001, FR-LAY-003, FR-TGT-001, FR-TGT-002]
---

## Goal
*Ensnaring Mist* (Animus Maneuver, not in a starter deck, kept in this run's `Cards` by #14 (c); 0 Requisition, 2 Tactics) is executable. It only combines shapes that are `done`: enacting a Maneuver (T-0116) with its required targets (T-0117), `OpMoveZone` to Reserves (T-0119), `OpModifyStat` until end of turn (T-0069) on a chosen unit (T-0135), and timed-out targets in fixed order (T-0056). It also carries #91 (4)'s same-unit test (`SVY-R35`).

## Inputs
- Card (`data/cards/maneuvers-animus.csv` lines 10-11): *Ensnaring Mist*, Animus Maneuver, cost 0 Requisition / 2 Tactics, two rules-text lines:
  - "Move chosen unit to its controller's Reserves."
  - "Chosen unit gets -1 power until end of turn."
- Survey row (`docs/design/m2_starter_set_ability_survey.md` line 134): `MOVE_ZONE` chosen unit → its Reserves; `MODIFY_STAT` −1 power until end of turn to a second chosen unit; rulings R26, R35.
- FRD 01 (run branch, v3.20):
  - `FR-MNV-001` (line 340, `SVY-R29`): enacted from Hand, costs paid on declaration; enactable only if each exactly-one slot has a legal entity; the card goes to its owner's Discard pile after resolving.
  - `FR-TGT-002`, Separate Target Clauses (`SVY-R35`, line 371): the two "chosen unit" clauses are two independent slots; both may pick the same unit, which is then moved and gets -1 power. *Ensnaring Mist* is the FRD's own example.
  - `FR-TGT-003` (`SVY-R26`, line 377): between the verification stages a step checks only presence, so the -1 step still acts on a unit the move step just moved.
  - `FR-ZON-002` (line 217): moving keeps the unit's state; a unit already in its controller's Reserves doesn't move (still a legal choice: "chosen unit" has no zone criterion).
  - `FR-LAY-001` (line 459): Effective Power = max(0, raw power); only power clamps. `FR-LAY-003` (line 471): an until-end-of-turn modifier expires in Sub-Phase 4c.
  - `FR-KWD-002h` (Ward): an opposing Ward unit is never a legal target, in either slot. "Chosen unit" has no allegiance, so allied units are legal; no structure or Headquarters is.
  - `FR-MOD-001` §6: a timed-out target takes the first legal entity in fixed order (the choosing player's assets first; Headquarters, Reserves, Main Theater, each zone in entry order). The two slots are independent, so both take that same unit.
- Shapes it uses, all `done`: `ActEnact` and `StackManeuver` (T-0116); `SVY-R29` legality (T-0117); `OpMoveZone` with `To: ZoneReserves` (T-0119, used with `SubjectAll` by *Rapturous Sermon*, T-0127; `stepMoveZone` goes through `stepAssets`, so `SubjectTarget` works); `OpModifyStat` on `SubjectTarget` (T-0069, T-0135, *Chainbreaker's Tenacity*); `TempPower` may go negative and `Power` clamps (`layers.go:8-10`); `stepAssets` presence checks (T-0107).

### Definition (the engineer adds it to `animusDefs`)
- *Ensnaring Mist*: `Text` holds the two lines. `Effect`:
  - `Targets`: two slots, each `ability.Target{Filter: ability.Filter{Kinds: ability.KindUnit}}`.
  - `Steps`: `{Op: ability.OpMoveZone, To: ability.ZoneReserves, Subject: ability.SubjectTarget, Target: 0}`, then `{Op: ability.OpModifyStat, Amount: ability.Value{N: -1}, Subject: ability.SubjectTarget, Target: 1}`.
- No `Triggers`, `ExtraCosts` or `Modes`. Model it on *Conversion Spike* (two slots, `animus.go:102`) and *Rapturous Sermon* (move and modify, `animus.go:370`).
- Remove the name from `cards.NotYetExecutable`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be added to `pkg/engine/cards_animus_maneuvers_test.go`, from the CSV text, loading the real card (`s.card("Ensnaring Mist")`). Player 0 enacts; player 1 is the opponent. Give player 0 the Tactics to pay. Answer each `DecideTarget` decision separately (slot 0 is the move, slot 1 the -1); a file-local chooser is fine.
- `TestCard_EnsnaringMist_MovesOneUnitAndWeakensAnother` (`FR-MNV-001`, `FR-ZON-002`, `FR-LAY-001`): opposing unit A in the Main Theater is moved to player 1's Reserves; opposing unit B (power 3) has effective power 2; A's power is unchanged; *Ensnaring Mist* is then the top of player 0's discard pile; 2 Tactics are spent.
- `TestCard_EnsnaringMist_SameUnitFillsBothSlots` (`SVY-R35`, `SVY-R26`, #91 (4)): both slots pick opposing unit A (power 3) in the Main Theater: A is in player 1's Reserves with effective power 2.
- `TestCard_EnsnaringMist_MovedUnitKeepsItsState` (`FR-ZON-002`, `FR-ZON-004`): A, damaged and exhausted, is moved: it keeps its damage and stays exhausted, and it is the same asset (same ID), not a new one.
- `TestCard_EnsnaringMist_MovesToItsControllersReserves`: an allied unit chosen for the move goes to player 0's Reserves, not player 1's.
- `TestCard_EnsnaringMist_UnitAlreadyInReservesStays`: a unit already in its controller's Reserves may be chosen for the move; it stays there, and no move event is emitted for it.
- `TestCard_EnsnaringMist_PowerClampsAtZero` (`FR-LAY-001`): the -1 slot on a unit with power 0 leaves its effective power at 0.
- `TestCard_EnsnaringMist_MinusOneExpiresAtEndOfTurn` (`FR-LAY-003`): after player 0's turn ends, B's power is its printed power again; A stays in Reserves.
- `TestCard_EnsnaringMist_ChosenUnitIsAnyUnit` (`FR-TGT-001`, `FR-KWD-002h`): each slot's options include allied and opposing units, in either zone, an allied Ward unit and a face-down Scratch Company unit, but no opposing Ward unit, no structure and no Headquarters.
- `TestCard_EnsnaringMist_LegalityNeedsAUnit` (`SVY-R29`, `FR-MNV-001`), table: with no unit on the battlefield (only Headquarters and a structure) it is not offered (`offersEnact`), enacting it is rejected (`s.reject`), and it stays in hand; with exactly one unit, it is offered, and both slots offer that unit.
- `TestCard_EnsnaringMist_TargetTimeoutsTakeFixedOrder` (`FR-MOD-001` §6, `SVY-R35`): both target decisions time out. Player 0 has allied unit P in the Main Theater (entered first) and player 1 has unit Q, so both slots take P: P is moved to player 0's Reserves and gets -1 power.
- `TestCard_EnsnaringMist_PaysTwoTactics`: with 1 Tactics it is not offered and enacting it is rejected; with 3 it is enacted, leaving 1 Tactics and spending no Requisition.
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestView_Case8_ExecutableFlag`, `TestProgress_Case4_Coverage`, `TestEvents_Case4_EveryActionKind` and the *Rapturous Sermon* and `OpMoveZone` tests.

## Helpers Needed
- None expected. The tests may add their own setup helper in the test file.

## Allowed Files
- `pkg/engine/cards/animus.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [x] The card tests pass: `go test ./pkg/engine -run TestCard_EnsnaringMist`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-11-03 lead-architect (wave C session 34): written against the run branch at 056fcde (no `main` commits to sync). Not `WAVE-COMPLETE`: *Ensnaring Mist* and *Pincer Tactics* are still on `NotYetExecutable`; this is the first of the two in the header's `Cards` order. A non-deck card kept by #14 (c): every shape it needs is `done`. #91 (4) goes into this task (the same-unit test); #91 (2) is about a cost on a mode, and this card has no modes. Waiting triggers #6, #7 (mode is `serial`), #8 (this task doesn't touch `deck.go`), #15 (no turn-start, turn-end or draw hook; a -1 power modifier can't make an asset lethal when it expires), #71 (no each-player effect) not fired. Status `tests`.
- 2026-11-03 rules-test-author: wrote 11 tests in pkg/engine/cards_animus_maneuvers_test.go (file-local helper emChoose, per-slot target chooser): TestCard_EnsnaringMist_MovesOneUnitAndWeakensAnother, _SameUnitFillsBothSlots, _MovedUnitKeepsItsState, _MovesToItsControllersReserves, _UnitAlreadyInReservesStays, _PowerClampsAtZero, _MinusOneExpiresAtEndOfTurn, _ChosenUnitIsAnyUnit, _LegalityNeedsAUnit, _TargetTimeoutsTakeFixedOrder, _PaysTwoTactics. Helpers requested: none. Cases left out: none. go vet clean, lint rules passes; all fail until the card is added. The one-unit case asserts offers only if a decision is asked (the rules do not say whether a single option is auto-picked). Status tests-review.
- 2026-11-03 lead-architect (wave C session 34): tests approved (b137a1f). Checked each test against FRD 01: `SVY-R29` (line 340) for legality with no unit and with one; `SVY-R35` (line 371) for one unit in both slots, including both timed-out slots taking P; `SVY-R26` (line 377) for the -1 on the unit just moved; `FR-ZON-002` for the controller's Reserves, the unmoved unit and kept state; `FR-LAY-001` for the clamp; `FR-LAY-003` for the expiry; `FR-KWD-002h` for allied versus opposing Ward. Run against a discarded trial definition (the Definition above): all 11 pass, and `go test ./pkg/engine/...` passes with it. Three mutations are each caught: both steps on slot 0 (2 tests fail), moving to the Main Theater (8) and an opposing-only filter (3). Status `ready`, dispatching `kernel-engineer` (flash).
- 2026-11-03 kernel-engineer: Leftover worktree cleanup: found leftover worktree from prior run with only in-progress status commit (90ded4d) and no uncommitted changes; removed it and opened scratch worktree. Reused existing shapes OpMoveZone with ZoneReserves (T-0119) and OpModifyStat with SubjectTarget (T-0069, T-0135). Added definition for Ensnaring Mist to animusDefs in pkg/engine/cards/animus.go with two KindUnit target slots, OpMoveZone to ZoneReserves on Target 0, and OpModifyStat -1 power on Target 1; removed Ensnaring Mist from NotYetExecutable in pkg/engine/cards/cards.go. Ran `go test ./pkg/engine -run TestCard_EnsnaringMist` (all 11 pass), `go test ./pkg/engine -run TestStarterSet_Executable` (pass), `go tool task sim` (clean on starter decks, all 200 matches terminated with invariants holding), and `go tool task check` (gofmt, go vet, all package tests, cardtown lint --all pass cleanly). Status review.
- 2026-11-03 lead-architect (wave C session 35): re-dispatched at `in-progress` after `main`'s 6727141 (the `Select-String` command for listing card definitions) was merged into the run branch (3b8f494). `agy` conversation `9f9bc0a9-34af-48e0-b6de-3f52e9f70eda` exited 0 with no refused command (its JSON has no `denied_actions` key; the transcript has no `permission check failed` step). Code approved: only the two Allowed Files changed after the test commit (b137a1f), and the definition is the one this task planned, laid out like *Conversion Spike*. No card name in engine code, no new shape, no duplicate. Re-run in a review worktree of the task branch: `go test ./pkg/engine -run TestCard_EnsnaringMist` passes, `go tool task check` passes, `go tool task sim` is clean (200 games, P0 82.5% / P1 17.5% / draws 0.0%, unchanged: a non-deck card). Merged into `autoloop/m2-wave-c` as e3e800d. #91 (4) is applied for this card; its *Pincer Tactics* part stays open. Status `done`.
