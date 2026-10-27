---
id: T-0078
title: Make Arkha, Sun's Fury executable (wave B card, flash)
kind: card
tier: flash
wave: B
status: tests
branch: m2/T-0078-arkha
depends_on: [T-0077]
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
