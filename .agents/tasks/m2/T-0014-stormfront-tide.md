---
id: T-0014
title: Make Stormfront of Ideals and Unceasing Tide executable
kind: card
tier: flash
wave: A1 (board directive #4, task 4: card task 4 of 4)
status: review
branch: m2/T-0014-stormfront-tide
depends_on: [T-0013]
claims: [card:Stormfront of Ideals, card:Unceasing Tide]
cards: [Stormfront of Ideals, Unceasing Tide]
requirements: []
---

## Goal
Two Animus units whose death triggers move the card itself now run in the kernel. The trigger tracks that exact card in its owner's discard pile and does nothing if the card has left. Each is plain `ability` data added to `animusDefs` in `pkg/engine/cards/animus.go`. With this task, every A1 card is executable.

## Inputs
- Cards (`data/cards/units-and-structures-animus.csv`). Each has exactly one unsupported line today, quoted here:
  - *Stormfront of Ideals*: Unit, 3R 0T, Reserves, 6/5. "When this unit is killed, put it on the bottom of its owner's deck."
  - *Unceasing Tide*: Unit, 3R 0T, Reserves, 3/2. "When this unit is killed, return it to its owner's hand."
- Survey rows: `docs/design/m2_starter_set_ability_survey.md` (both "Self-tracking fizzle").
- Shapes they use (all `done`): `HookKilled` (T-0008), `SubjectSelf`, `OpPutOnDeck` with `Bottom: true`, and `OpReturnToHand` (T-0009).
- Requirements for context: `FR-ZON-004`, the Self-Relocating Trigger Tracking bullet. FRD 01 names both cards.
- Rulings that apply: none.

### Definitions (the engineer adds them to `animusDefs`)
- *Stormfront of Ideals*: one `HookKilled` trigger, `{Op: OpPutOnDeck, Subject: SubjectSelf, Bottom: true}`.
- *Unceasing Tide*: one `HookKilled` trigger, `{Op: OpReturnToHand, Subject: SubjectSelf}`.
- Each `Text` is the quoted line exactly. No new types, fields or logic. If a card seems to need one, set `blocked`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
All in `pkg/engine/cards_animus_test.go`, loading the real cards; each covers FR-ZON-004 (Self-Relocating Trigger Tracking).
- `TestCard_StormfrontOfIdeals_KilledGoesToBottomOfOwnersDeck`: that exact instance is the last card of the deck, not in discard, deck grew by one (killed on the opponent's turn).
- `TestCard_StormfrontOfIdeals_TriggersOnItsControllersTurnToo`: killed on its controller's own turn, same result.
- `TestCard_StormfrontOfIdeals_OwnerGetsItNotController`: owner 0, controller 1; it goes to the owner's deck bottom, not the controller's.
- `TestCard_StormfrontOfIdeals_FaceDownCopyStaysInDiscard` (guard): a face-down copy has no ability and stays in discard.
- `TestCard_UnceasingTide_KilledReturnsToOwnersHand`, `_TriggersOnItsControllersTurnToo`, `_OwnerGetsItNotController`, `_FaceDownCopyStaysInDiscard`: the same four for Tide, with hand as the destination.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None. (Tests build the owner/controller split inline with `Asset{...}` and `placeOnBattlefield`, as `put` does.)

## Allowed Files
- `pkg/engine/cards/animus.go`
- `pkg/engine/cards/cards.go` (remove the two names from `NotYetExecutable`)

## Done When
- [x] The cards' tests pass: `go test ./pkg/engine -run 'TestCard_(StormfrontOfIdeals|UnceasingTide)'`
- [x] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged
- [x] `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean
- [x] `go run ./cmd/cardtown lint --all` passes
- [x] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0, and the Animus executable count rises by 5 (2 Stormfront, 3 Tide)
- [x] `go run ./cmd/cardtown sim` exits 0. Report the before and after summary lines in the Log.

## Log
- 2026-10-20 lead-architect: Wrote the task (A1 card task 4 of 4). It stays `queued` until T-0013 is `done`, because they share `animus.go` and `cards.go`.
- 2026-10-20 lead-architect: T-0013 is `done`, so status `tests` for `rules-test-author`.
- 2026-10-20 rules-test-author: Wrote 8 card tests on m2/T-0014-stormfront-tide. Run: 6 fail (the cards have no trigger yet), 2 pass (the face-down guards). No helpers requested. Left out: the fizzle case (card already left discard when the trigger resolves). The trigger resolves at the death sweep, and state setup cannot move the card out of discard before then without guessing a resolution-ordering ruling that FRD 01 does not give. The fizzle rule itself is stated in FR-ZON-004 and needs a card or effect that moves the card first. Owner-vs-controller is tested because the bullet says "its owner's" deck and hand.
- 2026-10-20 lead-architect: Test review: **approved**, status `ready`. `test(T-0014)` 19bf43e appends 8 tests to `cards_animus_test.go` without changing any line of an existing test. They use the real cards and follow FR-ZON-004's Self-Relocating Trigger Tracking bullet, which names both cards.
  - Each test checks that exact instance's destination (bottom of the owner's deck, the owner's hand) and that it left discard. The Stormfront test also checks that the deck grew by exactly one card.
  - Both turn cases are covered, because neither card has a turn condition.
  - The owner/controller split follows "its owner's" literally. `stolenKilled` mirrors `s.put` and changes only `Controller`.
  - The two face-down guards pass today (FR-ZON-004, line 163).
  - Leaving out the fizzle is accepted. The shape tests already pin it (`TestFR_ZON_004_ReturnResolvingFirstFizzlesThePutOnBottom` and `..._PutOnBottomResolvingFirstFizzlesTheReturn` in `self_tracking_test.go`, T-0009). These cards' one-step triggers add no new case.
  - Minor, no change requested: `inDiscardPile` duplicates `inZone(p.Discard, inst)` from `self_tracking_test.go`. Later card tasks should reuse `inZone`.
  - Re-run: 6 fail (no trigger yet), 2 guards pass, and no other test fails.
  - Next: **the user runs `@kernel-engineer T-0014` in the Antigravity agent (Gemini Flash)**.
- 2026-10-20 kernel-engineer: Implemented. animus.go: added definitions to animusDefs for "Stormfront of Ideals" (HookKilled, OpPutOnDeck, Subject: SubjectSelf, Bottom: true) and "Unceasing Tide" (HookKilled, OpReturnToHand, Subject: SubjectSelf) in alphabetical order. cards.go: removed "Stormfront of Ideals" and "Unceasing Tide" from NotYetExecutable allowlist while preserving sort order. Reused HookKilled, SubjectSelf, OpPutOnDeck with Bottom: true, OpReturnToHand; no new shapes or engine logic added. Gates: go test ./pkg/engine -run 'TestCard_(StormfrontOfIdeals|UnceasingTide)' passed (8/8), TestDefs_MatchRulesText and TestStarterSet_Executable passed, go test ./... ok, go vet ./... clean, gofmt -l pkg cmd clean, cardtown lint --all passed. Starter deck sim reported P0 (Animus starter) main executable cards rose from 14/50 to 19/50 (2 Stormfront, 3 Tide). Default sim before and after: "SOULLESS (Colossus Platform) vs CONCORDIA (Enclave), P0 wins 66.0% P1 wins 30.0% draws 4.0% first player wins 47.0%, Avg match length: 30.2 turns" (unchanged as default sim runs Soulless vs Concordia). Reviewer: check that Stormfront of Ideals sets Bottom: true with OpPutOnDeck and SubjectSelf, and Unceasing Tide sets OpReturnToHand with SubjectSelf.
