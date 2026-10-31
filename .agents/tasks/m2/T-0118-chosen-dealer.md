---
id: T-0118
title: A chosen asset can deal a step's damage, equal to its current power, and deals none once it is gone (FR-TGT-003, SVY-R27, SVY-R31, directive #88 part 3, wave C shape, sonnet)
kind: shape
tier: sonnet
wave: C
status: tests-review
branch: m2/T-0118-chosen-dealer
depends_on: [T-0117]
claims: [field:Step.Dealer, step:chosen-dealer]
cards: []
requirements: [FR-TGT-003, FR-ZON-004, FR-MOD-001]
---

## Goal
A damage step can name one of its effect's target slots as the dealer: *"Chosen unit deals damage equal to its power to another chosen unit"* (*Ambush Tactics*' third mode). The chosen asset deals the damage, attributed to it, equal to its current effective power as the step resolves. If an earlier step or mode removed it, or it was illegal at the second verification stage, the step deals nothing and emits no damage event (`SVY-R27`). This is directive #88's delegated-dealer bullet, tests (c) and (d).

## Inputs
- FRD 01:
  - `FR-TGT-003` (line 366), "A Chosen Asset That Deals Damage (`SVY-R27`)" (line 370): the dealer is a target like any other. Gone at the second stage, or removed by an earlier step: no damage, and no last-known power. Otherwise it deals its current effective power (Layer 6) as the step resolves.
  - `FR-TGT-003`, "Targets Removed During Resolution" (line 369, `SVY-R25`, `SVY-R26`) and "Complete Target Invalidation" (line 372): a step whose chosen assets are all gone did nothing, so an "if you do" step after it doesn't apply (`SVY-R14`).
  - `FR-ZON-004`, "Values Read From the Source (`SVY-R10`)" (line 223): last-known values are for the effect's own source only.
  - `FR-MOD-001` §5 (line 384, `SVY-R31`): chosen modes resolve in printed order, whatever order they were picked in.
  - `FR-REP-004` Stage 5 (line 510): the damage event names who dealt it.
- Rulings that apply: `SVY-R10`, `SVY-R14`, `SVY-R25`, `SVY-R27`, `SVY-R31`.
- Survey row: *Ambush Tactics* (`docs/design/m2_starter_set_ability_survey.md` line 175): "delegated `DEAL_DAMAGE` = power". The card itself is a later card task.
- Board directive #88, second bullet ("the dealer is a target slot read through `stepAssets` (#82). With no dealer left, the step deals nothing and emits no damage event. The amount is the dealer's current `Power`, never a last-known value. `value`'s last-known fallback stays for the item's source"), tests (c) and (d).
- Kernel today:
  - `ability.Step` (`ability/ability.go:206`): `Target` is a 0-based slot index read only with `SubjectTarget`. `ValueSourcePower` (`ability.go:222`) is "the source's effective power ... or its last-known".
  - `stepDealDamage` (`effects.go:98`): the amount is `g.value(r.it, st.Amount)` and the dealer is `g.itemSource(r.it)`, always.
  - `value` (`effects.go:253`): the source's power, else its last-known power, else 0. Six callers in `effects.go`.
  - `stepAssets`' `SubjectTarget` case (`effects.go:271-279`): the slot's chosen assets still on the battlefield (`g.Asset(a.ID) == a`, #82).
  - `chooseModes` (`placement.go:81`): joins the chosen modes in printed order and shifts each `SubjectTarget` step's `Target` by the slots before it.
  - `resolveAbility`'s doc comment (`effects.go:25`) still says "a triggered ability's steps"; an enacted Maneuver resolves through it too (T-0116).

### Shape (lead decisions, binding)
- **`ability.Step.Dealer int`** (`ability/ability.go`), with the doc comment: OpDealDamage only. 0 (the zero value): the ability's source deals the damage, as today. k > 0: the asset chosen for the exactly-one slot `Effect.Targets[k-1]` deals it, and `ValueSourcePower` reads that asset's power (`SVY-R27`). Add `func (s Step) DealerSlot() (slot int, ok bool)` returning `s.Dealer-1, s.Dealer > 0`, so no caller does the arithmetic. Every existing card and test keeps `Dealer` 0 and is unchanged.
- **`ValueSourcePower`'s doc comment** becomes: the effective power of the step's dealer when the step resolves. The dealer is the source, read with its last-known effective power once it has left (`SVY-R10`), or a chosen dealer (`Step.Dealer`), read only while it is on the battlefield (`SVY-R27`).
- **One reader of a Value.** `value` takes the dealer: `(g *Game) value(it *StackItem, v ability.Value, dealer *Asset) int`. With `dealer` nil it reads the item's source exactly as today (current, else last-known, else 0). With a non-nil `dealer` it reads `g.Power(dealer)`, never a last-known value. Every other caller passes nil. The `From` switch stays in this one function, so a later `ValueSource` is added in one place.
- **One slot reader.** Move `stepAssets`' `SubjectTarget` body into `(g *Game) slotAssets(r *resolution, i int) []*Asset`: slot i's chosen assets still on the battlefield, in pick order, or nil for an index out of range. `stepAssets` calls it, with no change in behavior, and so does the dealer. There is no second copy of the `g.Asset(a.ID) == a` check.
- **`stepDealDamage`** (`effects.go`): with `st.DealerSlot()` ok, the dealer is the first asset `slotAssets` returns for that slot. With none, the step returns 0 before anything else: no damage, no event, even if its own targets are still there. Otherwise that asset is the `source` passed to `dealDamage`, and the amount is `g.value(r.it, st.Amount, dealer)`. Without a dealer slot, the step is unchanged (`itemSource`, `value(..., nil)`). Its doc comment names `SVY-R27`.
- **Modes keep their dealer.** In `chooseModes`' join (`placement.go`), a step with a dealer slot gets `Dealer += shift`, as `Target` does, so the dealer stays the slot of its own mode. Copy the step first, as the loop already does: a mode's Steps belong to the card's definition.
- **`resolveAbility`'s doc comment**: "runs a triggered ability's steps in order" becomes the stack item's effect, a triggered ability's or an enacted Maneuver's.
- **Out of scope:**
  - Distinctness between two slots (*"another chosen unit"*): no filter compares one slot with another today. It is its own task, before the *Ambush Tactics* card task, so these tests choose distinct units by script. When the dealer and the damaged unit are the same asset, the step does whatever it does; no test pins it.
  - `MOVE_ZONE` (#88 test (e)) is the next task.
  - Costs locked with a mode (#61, fifth question) wait for their ruling.
- **Sims:** no card gains a dealer, so `go tool task sim-compare -- autoloop/m2-wave-c` should pass byte for byte. If it differs, don't change code to make it match: report the first differing game in the Log and hand off.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Synthetic Maneuvers only, in a new file `pkg/engine/dealer_test.go`, reusing `maneuver_test.go`'s `maneuver`, `steps`, `enactAction` and `lastDiscard` helpers and `modal_test.go`'s `mode` and `modeAnswer` where they fit. "X deals damage equal to its power to Y" below is one `OpDealDamage` step with `Subject: SubjectTarget` on Y's slot, `Dealer` on X's slot, and `Amount: Value{From: ValueSourcePower}`. Give X and every other unit distinct powers, so that the wrong dealer shows up as the wrong amount. Every existing test passes unchanged.
- `TestFR_TGT_003_RemovedDealerDealsNothing` (#88 (c)): a "Choose two" Maneuver with modes "Return chosen unit to its owner's hand" and "Chosen unit deals damage equal to its power to another chosen unit". Both modes chosen, unit X returned and dealing, unit Y the one dealt to. X ends in its owner's hand, Y has no damage, no damage event is logged, and the Maneuver ends in its owner's discard pile.
- `TestFR_MOD_001_ChosenModesResolveInPrintedOrder` (#88 (d)): modes "Chosen unit deals damage equal to its power to another chosen unit", then "Put chosen unit on top of its owner's deck", picked in the reverse order, with X the dealer and the deck target and Y dealt to. Y takes damage equal to X's power, the damage event's source is X, and then X is the top card of its owner's deck.
- `TestFR_TGT_003_DealerInALaterModeIsItsOwnSlot`: modes "Return chosen unit to its owner's hand" (unit Z), then "Chosen unit deals damage equal to its power to another chosen unit" (X dealing to Y). Y takes X's power, attributed to X. (The dealer mode's slots come after the return mode's, so a dealer slot that isn't shifted would name Z, who is gone.)
- `TestFR_TGT_003_DealerDealsItsCurrentPower`: a non-modal Maneuver with slots "chosen unit" (X) and "chosen unit" (Y), and steps "X gets +2 power until end of turn" then "X deals damage equal to its power to Y". Y takes X's printed power plus 2, attributed to X.
- `TestFR_TGT_003_DealerGoneBeforeResolutionDealsNothing`: X, chosen as the dealer, leaves the battlefield after the Maneuver is placed and before it resolves (use whatever the existing removed-target or two-stage tests use to remove a target while an item waits on the stack). When it resolves, Y takes no damage, no damage event is logged, and an "if you do: draw a card" step after the damage step does not apply: the controller's hand size shows no draw. X's last-known power is not used, so check that X had power above 0 when it left.
- A damaged target that is gone already takes nothing (`TestFR_ZON_004_ReturnedTargetTakesNoLaterDamage`), so no new test covers it.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/ability/ability.go` (`Step.Dealer`, `Step.DealerSlot`, the `ValueSourcePower` doc comment)
- `pkg/engine/effects.go` (`value`, `slotAssets`, `stepAssets`, `stepDealDamage`, the other `value` callers, `resolveAbility`'s doc comment)
- `pkg/engine/placement.go` (`chooseModes`' join)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_TGT_003|TestFR_MOD_001|TestFR_ZON_004|TestFR_MNV_001'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-c` passes (no card gains a dealer; see Shape, Sims)
- [ ] A Value is read in one function (`value`), and a slot's remaining assets in one (`slotAssets`).

## Log
- 2026-10-30 lead-architect: wave C run, session 3. No `main` commits to sync (`git log autoloop/m2-wave-c..main` is empty). Not `WAVE-COMPLETE`: 22 deck cards are still on `NotYetExecutable`. Directives read on the run branch: #88 is `open`, and its delegated-dealer bullet is this task (tests (c), (d)); #14 and #20 (b) stay open for the card tasks and *Steel Rain*. Waiting triggers checked: #58 (no exhaust in `claims:`), #59 (2) (no repeated targets), #61's fifth question (`cards:` is empty), #71 (no each-player effect), #15 and #8 (no hook or `deck.go` change). None has fired. Duplicates searched: `stepDealDamage` always deals as the item's source, and no step names another dealer; `value` is the one reader of a `Value` and stays so, taking the dealer; the slot's "still on the battlefield" check moves into `slotAssets`, shared by `stepAssets` and the dealer. Three non-test files. Distinctness between slots (*"another chosen unit"*) is left out, as its own shape: no filter compares two slots today. Recorded in #88's status cell as the task after this one, with `MOVE_ZONE` (test (e)). Status `tests`; dispatching `rules-test-author`.
- 2026-10-30 rules-test-author: tests written in `pkg/engine/dealer_test.go`: `TestFR_TGT_003_RemovedDealerDealsNothing`, `TestFR_MOD_001_ChosenModesResolveInPrintedOrder`, `TestFR_TGT_003_DealerInALaterModeIsItsOwnSlot`, `TestFR_TGT_003_DealerDealsItsCurrentPower`, `TestFR_TGT_003_DealerGoneBeforeResolutionDealsNothing`. Local helpers in the same file (`newDealerBoard`, `dealerStep`, `dealerMode`, `returnMode`, `deckTopMode`, `wantDamageFrom`); no helpers requested for `scenario_test.go`. `go vet` fails only on the missing `ability.Step.Dealer` field (the engineer's). `cardtown lint rules` passes. The dealer leaves mid-stack through an event observer on the Maneuver's `place` event calling `removeFromBattlefield`. Left out: dealer equal to the damaged unit (out of scope per Shape). Status `tests-review`.
