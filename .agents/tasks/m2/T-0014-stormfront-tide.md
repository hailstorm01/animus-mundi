---
id: T-0014
title: Make Stormfront of Ideals and Unceasing Tide executable
kind: card
tier: flash
wave: A1 (board directive #4, task 4: card task 4 of 4)
status: tests-review
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
- [ ] The cards' tests pass: `go test ./pkg/engine -run 'TestCard_(StormfrontOfIdeals|UnceasingTide)'`
- [ ] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged
- [ ] `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean
- [ ] `go run ./cmd/cardtown lint --all` passes
- [ ] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0, and the Animus executable count rises by 5 (2 Stormfront, 3 Tide)
- [ ] `go run ./cmd/cardtown sim` exits 0. Report the before and after summary lines in the Log.

## Log
- 2026-10-20 lead-architect: Wrote the task (A1 card task 4 of 4). It stays `queued` until T-0013 is `done`, because they share `animus.go` and `cards.go`.
- 2026-10-20 lead-architect: T-0013 is `done`, so status `tests` for `rules-test-author`.
- 2026-10-20 rules-test-author: Wrote 8 card tests on m2/T-0014-stormfront-tide. Run: 6 fail (the cards have no trigger yet), 2 pass (the face-down guards). No helpers requested. Left out: the fizzle case (card already left discard when the trigger resolves). The trigger resolves at the death sweep, and state setup cannot move the card out of discard before then without guessing a resolution-ordering ruling that FRD 01 does not give. The fizzle rule itself is stated in FR-ZON-004 and needs a card or effect that moves the card first. Owner-vs-controller is tested because the bullet says "its owner's" deck and hand.
