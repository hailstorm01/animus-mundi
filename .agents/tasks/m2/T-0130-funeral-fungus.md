---
id: T-0130
title: Make Funeral Fungus executable (exhaust a readied allied unit as an additional cost, then return chosen unit card from your discard pile to your hand, wave C card, flash)
kind: card
tier: flash
wave: C
status: review
branch: m2/autoloop-wave-c/T-0130-funeral-fungus
depends_on: [T-0081, T-0128, T-0129]
claims: [card:Funeral Fungus]
cards: [Funeral Fungus]
requirements: [FR-TRN-003a, FR-MNV-001, FR-TGT-001]
---

## Goal
*Funeral Fungus* (a deck card, 0 Requisition, 1 Tactics): *"As an additional cost to enact this maneuver, exhaust a readied allied unit."* / *"Return chosen unit card from your discard pile to your hand."* Split from T-0128 under the D2 size rule; it only combines T-0129's exhaust cost with T-0081's discard-pile target and `OpReturnToHand`.

Covers:
- The definition in `pkg/engine/cards/animus.go`: the exhaust cost on a Readied allied unit, and one exactly-one target slot "unit card in your discard pile" with an `OpReturnToHand` step on it.
- The name off `cards.NotYetExecutable`.
- Benchmark blueprint #8 (`docs/frd/reference/02_Benchmark_Card_Blueprints.md` line 307) as a test source: no Readied allied unit, or no unit card in its controller's discard pile, means it can't be enacted (`SVY-R29`).

## Inputs
- Card (`data/cards/maneuvers-animus.csv` line 5): *Funeral Fungus*, Animus Maneuver, cost 0 Requisition / 1 Tactics, two rules-text lines:
  - "As an additional cost to enact this maneuver, exhaust a readied allied unit."
  - "Return chosen unit card from your discard pile to your hand."
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 127: exhaust readied allied unit; `RETURN_TO_HAND` chosen unit card from discard; `FR-TRN-003a`; Benchmark #8.
- Benchmark blueprint #8 (`docs/frd/reference/02_Benchmark_Card_Blueprints.md` lines 307-345): `EXHAUST_ASSET` cost with filter allied, unit, Readied; one `CHOSEN` target, allied, unit card, in the discard pile, count 1; `RETURN_TO_HAND` on it.
- FRD 01:
  - `FR-TRN-003a` (line 335, which names this card): the additional cost is paid on declaration. Abilities triggered by paying it go on the stack on top of the Maneuver and resolve before it.
  - `FR-MNV-001` (line 333): enacted from Hand, costs paid on declaration, the effect runs when it resolves, then the card goes to its owner's Discard pile.
  - `SVY-R29` (`FR-MNV-001`, line 334): it can be enacted only if its exactly-one target slot has a legal entity on declaration. A Maneuver that can't be enacted stays in hand.
  - `FR-TGT-001` (line 347): conjunction eligibility. "Your discard pile" is the controller's; "unit card" excludes Maneuvers and Structures.
  - `SVY-R17` (`FR-CAP-001`, line 55): paying the cost is a change from Readied to Exhausted, so "whenever this unit exhausts" triggers fire, on top of the Maneuver.
  - `FR-TGT-003`: two-stage verification; a target that left the discard pile before resolution is skipped and the card still goes to the discard pile.
- Shapes it uses, all `done`: `Def.Effect`, `ActEnact` and `StackManeuver` (T-0116), the `SVY-R29` enact legality (T-0117, `placeable` and `slotsLegal`, `placement.go`), `Def.ExtraCosts` and `ability.Cost` (T-0128), `ability.CostExhaust` (T-0129: `costDecision` offers only Readied assets, `placement.go:49-61`), a chosen target in a discard pile (T-0081: `ability.ZoneDiscard`, `Filter.InPile`), `OpReturnToHand` on `SubjectTarget` (as *Eternal Watchman*, `cards/headquarters.go:30-55`) and the progress policy's enacts (T-0121).
- Rulings: `SVY-R2` (targets and costs are committed together), `SVY-R17`, `SVY-R29`. No other.

### Definition (the engineer adds it to `animusDefs`)
- *Funeral Fungus*: `Text` holds both lines. `ExtraCosts`: one `{Kind: ability.CostExhaust, Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}}` (Readied is what `CostExhaust` means, so the filter has no readiness field). `Effect`:
  - `Targets`: one exactly-one slot, `Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}`.
  - `Steps`: one, `{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget, Target: 0}`.
- No `Triggers` or `Modes`. Model it on *Feed the Despoiler*'s `ExtraCosts` and *Eternal Watchman*'s target and step.
- Remove the name from `cards.NotYetExecutable`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be added to `pkg/engine/cards_animus_maneuvers_test.go`, from the CSV text, loading the real card (`s.card`). Player 0 enacts; player 1 is the opponent. Give player 0 enough Tactics to pay (`enactFor`). Put cards in a discard pile with `s.inDiscard`.
- `TestCard_FuneralFungus_ExhaustsAndReturnsTheChosenUnitCard`: player 0 has a Readied allied unit on the battlefield and two unit cards in its discard pile. Enacting it, with player 0 choosing the older of the two (not the top one), exhausts the allied unit, puts the chosen unit card in player 0's hand, leaves the other in the discard pile, and puts the Maneuver in player 0's discard pile (`lastDiscard`). It spends 1 Tactics and no Requisition.
- `TestCard_FuneralFungus_OnlyAReadiedAlliedUnitPays`: player 0 has an Exhausted allied unit, a Readied allied structure and a unit card in its discard pile; player 1 has a Readied unit. It is not offered (`offersEnact`), enacting it is rejected (`s.reject`), and the card stays in hand. When a Readied allied unit is added, it is offered, and enacting it exhausts that unit and no other asset.
- `TestCard_FuneralFungus_NeedsAUnitCardInItsControllersDiscardPile` (`SVY-R29`, `FR-TGT-001`): player 0 has a Readied allied unit, a Maneuver card and a Structure card in its own discard pile, and player 1 has a unit card in its discard pile. It is not offered, enacting it is rejected, the card stays in hand, and the allied unit stays Readied. When a unit card is added to player 0's discard pile, it is offered.
- `TestCard_FuneralFungus_ScratchCompanyCanPay`: a Readied face-down Scratch Company unit is a Readied allied unit, so it can pay the cost; after enacting, it is Exhausted.
- `TestCard_FuneralFungus_ShamblerTriggerResolvesFirst` (`FR-TRN-003a`, `SVY-R17`): the only Readied allied unit is the real *Animoceps Shambler* ("Whenever this unit exhausts, chosen allied asset in the Main Theater heals 1."), and a damaged allied asset is in the Main Theater. Paying the cost with the Shambler fires its trigger, which resolves before the Maneuver: the damaged asset heals 1, and in `g.Log` the heal comes before the unit card's return to hand.
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestView_Case8_ExecutableFlag` (seat 0's seed-1 hand still holds unsupported cards), `TestProgress_Case4_Coverage` and `TestEvents_Case4_EveryActionKind`.

## Helpers Needed
- None. The tests add their own setup helper `ffChoose` in the test file.

## Allowed Files
- `pkg/engine/cards/animus.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [x] The card tests pass: `go test ./pkg/engine -run TestCard_FuneralFungus`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-10-31 lead-architect: queued in wave C session 15, with T-0128's task commit (D2.4).
- 2026-11-02 lead-architect: wave C session 17. Inputs, Tests and Allowed Files written against 242811e; branch renamed for the wave run (D7.4). No `open: before` directive; waiting triggers #6, #7, #8, #15, #59 (2), #61 (fifth question), #71, #88 (distinctness) not fired. Status `tests`.
- 2026-11-02 rules-test-author: tests written in `pkg/engine/cards_animus_maneuvers_test.go`: `TestCard_FuneralFungus_ExhaustsAndReturnsTheChosenUnitCard`, `_OnlyAReadiedAlliedUnitPays`, `_NeedsAUnitCardInItsControllersDiscardPile`, `_ScratchCompanyCanPay`, `_ShamblerTriggerResolvesFirst`. Helpers requested: none (own helper `ffChoose` in the test file). All five fail now (card not executable). Left out: the FR-TGT-003 skipped-target case, which the task did not list. Heal-before-return ordering is checked by sampling state on each observed event, since the heal and return-to-hand event kind names are not in the test helpers. Status `tests-review`.
- 2026-11-02 lead-architect (tests-review, round 1): status `tests`. The other four tests are approved as written. Notes:
  1. `TestCard_FuneralFungus_ExhaustsAndReturnsTheChosenUnitCard` never exercises the payer's choice of which Readied allied unit to exhaust ("exhaust a readied allied unit" is chosen by the payer, `FR-TRN-003a`). The cost payer is picked through a `DecidePick` decision, not `DecideTarget`, so `ffChoose` answers it with option 0. Cost options are in asset-ID order (lowest ID first), so `Ally` is option 0 and the `ally` in `prefer` has no effect. As written, the test asserts the fixed option order, which is an implementation detail. Fix: have `ffChoose` apply `prefer` to `DecidePick` decisions as well as `DecideTarget`, and in this test choose `Bystander` (in the Reserves, the later ID, so not option 0) as the payer. Assert that `Bystander` is Exhausted and `Ally` is still Readied. Keep the rest of the test as it is.
- 2026-11-02 rules-test-author (round 1): applied note 1. `ffChoose` now applies `prefer` to `DecidePick` decisions (the cost payer) as well as `DecideTarget`. `TestCard_FuneralFungus_ExhaustsAndReturnsTheChosenUnitCard` now chooses `Bystander` (Reserves, later ID) as the payer and asserts `Bystander` is Exhausted and `Ally` is still Readied. Rest unchanged. All five tests still fail until the card is implemented. Helpers requested: none. Status `tests-review`.
- 2026-11-02 lead-architect (tests-review, round 2): status `tests`. Notes:
  2. Note 1 is not applied in the code. bcaff5c changed `ffChoose`'s doc comment, but its body still reads `if d.Kind == DecideTarget {`, so the payer pick (`DecidePick`) is still answered with option 0, which is `Ally`. The revised `TestCard_FuneralFungus_ExhaustsAndReturnsTheChosenUnitCard` would then fail against a correct implementation. Change the condition to `if d.Kind == DecideTarget || d.Kind == DecidePick {` and re-wrap the doc comment to the file's line width. Before handing off, run `git show HEAD -- pkg/engine/cards_animus_maneuvers_test.go` and check that the condition line is in the diff.
- 2026-11-02 rules-test-author (round 2): applied note 2. `ffChoose`'s condition is now `d.Kind == DecideTarget || d.Kind == DecidePick` (confirmed in `git show HEAD`), and the doc comment is re-wrapped. No other change. All five tests still fail until the card is implemented. Helpers requested: none. Status `tests-review`.
- 2026-11-02 lead-architect (tests-review, round 3): status `tests`. Note 2 is applied (381916d). With a trial definition in a review worktree (discarded, not committed), the other four tests pass and this one fails:
  3. `TestCard_FuneralFungus_ShamblerTriggerResolvesFirst` waits for an event after the card returns to hand, but neither healing nor returning a card to hand emits an event. The event kinds are in `events.go:240-261`, with no heal or return kind. After the Shambler trigger's `place` event, `g.Log` has no more events, so `returned` stays false even though the heal and the return both happened in the right order. Check the order the way T-0129's `TestFR_TRN_003a_ExhaustCostTriggerResolvesBeforeTheManeuver` does, from `place` events: an `EventPlace` event's `Amount` is the stack depth once placed. Assert that the Maneuver's `place` event (`Card` "Funeral Fungus") has `Amount` 1, and that the Shambler trigger's `place` event (`Card` "Animoceps Shambler") comes later in `g.Log` with `Amount` 2. That puts it on top of the Maneuver, so it resolves first. Drop the `recorder` observer and keep the end-state assertions: the Shambler is Exhausted, the damage is 2 and the card is in hand.
- 2026-11-02 rules-test-author (round 3): applied note 3. `TestCard_FuneralFungus_ShamblerTriggerResolvesFirst` no longer uses the `recorder` observer. It reads `g.Log` `place` events: the Maneuver (`Funeral Fungus`) has `Amount` 1, the Shambler trigger (`Animoceps Shambler`) comes later with `Amount` 2. The end-state assertions are kept (Shambler Exhausted, damage 2, card in hand). No other change (confirmed in `git show HEAD`). All five tests still fail until the card is implemented. Helpers requested: none. Status `tests-review`.
- 2026-11-02 lead-architect (tests-review, round 4): tests approved, status `ready`. Note 3 is applied (7f51960). Checked against `FR-TRN-003a`, `FR-MNV-001`, `SVY-R29`, `FR-TGT-001` and `SVY-R17`. A trial definition in a review worktree (as under Definition, discarded and not committed) passes all five tests and the rest of `go test ./pkg/engine`.
- 2026-11-02 kernel-engineer: implemented Funeral Fungus definition in `pkg/engine/cards/animus.go` with `ExtraCosts` (`CostExhaust` for allied units) and `Effect` (`OpReturnToHand` on target in `ZoneDiscard` with allied unit filter); removed "Funeral Fungus" from `NotYetExecutable` in `pkg/engine/cards/cards.go`. Reused existing shapes: `ability.CostExhaust` (T-0129), `ability.ZoneDiscard` (T-0081), `ability.OpReturnToHand` and `ability.SubjectTarget`. Ran `go test ./pkg/engine -run TestCard_FuneralFungus` (all 5 tests pass), `go test ./pkg/engine -run TestStarterSet_Executable` (passes), `go tool task check` (all tests, linters, gofmt, go vet pass), `go tool task sim` (200 starter games clean, all invariants hold). Status `review`.
