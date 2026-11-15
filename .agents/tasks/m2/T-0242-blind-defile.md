---
id: T-0242
title: "Blind Defile: whenever an opposing unit attacks, you may kill it; if you do, kill that attacking unit (wave H card, flash; directive #155)"
kind: card
tier: flash
wave: H
status: in-progress
branch: m2/autoloop-wave-h/T-0242-blind-defile
depends_on: [T-0241]
claims: [card:Blind Defile]
cards: [Blind Defile]
requirements: [FR-CBT-002, FR-TRG-003, FR-TRG-005, FR-TGT-001, FR-MOD-001, FR-REP-001]
---

## Goal
*Blind Defile* (Path structure, `data/cards/units-and-structures-path.csv:7`; in neither starter deck): "Whenever an opposing unit attacks, you may kill this structure. If you do, kill that attacking unit." is executable from shapes that are `done`: T-0241's watching trigger (`Trigger.Watches`, `Trigger.Watch`) and `SubjectWatched`, with `Optional`, `OpKill` on `SubjectSelf`, and `IfDone` (`SVY-R14`, `SVY-R15`). It leaves `cards.NotYetExecutable`. With it, directive #155 is fully applied.

## Inputs
- Card text: `data/cards/units-and-structures-path.csv:7` (Path structure, Territory, R2 T0, Main Theater, 0/2, yields 0/2), reworded by `SVY-R49` (FRD 01 v3.26, commit 4e35317, merged into the run branch at 1f6f08d).
- Board directive #155 (`SVY-R49`, `open: before T-0242`), from the [outcome](../../../docs/design/outcomes/2026-11-15-wave-h-blind-defile-kill-both.md) (`Status: applied: FRD 01 v3.26`). Its tests (a)–(e) are the tests below.
- FRD 01 v3.26:
  - `FR-TGT-001` *"You May X. If You Do, Y"* (line 394): the self-kill is X, the attacker's kill is Y. If the Defile has left the battlefield before the trigger resolves, X selects nothing and the attacker is not killed. If the attacker has left, the Defile is still killed and Y does nothing (`SVY-R49`).
  - `FR-TRG-005` (line 463): the self-kill is not a cost; it is carried out when the trigger resolves.
  - `FR-REP-001` *Replaced Actions Still Count as Done* (line 521, `SVY-R15`): a replaced kill of the Defile still counts as done for "if you do".
  - `FR-CBT-002` Step 2 (line 583): attack triggers are placed on top of the Attack and resolve before Step 3, so an attacker killed by the trigger deals no combat damage.
  - `FR-TRG-003` (line 453): simultaneous triggers of one controller are ordered by that controller (`DecideOrder`).
  - `FR-MOD-001` §6, §7 (from line 418): "you may" is asked as the trigger is placed (`SVY-R9`), and a timeout declines it.
- Shapes it combines, all `done`: `Trigger.Watches`, `Trigger.Watch`, `SubjectWatched`, `StackItem.Watched` (T-0241, merge 4176846); `Optional` (`SVY-R9`); `OpKill` with `SubjectSelf`; `Step.IfDone` (`SVY-R14`; *Infested Apostle*, *Arkha, Sun's Fury*). `pkg/engine/watch_trigger_test.go` uses the synthetic form of this trigger (`watchTrigger`, `opposingUnits`, `killWatched`).

## Shape (the lead's plan)
- `pkg/engine/cards/path.go`: the definition, in alphabetical order (after *Arkha, Sun's Fury*, before *Boneyard Amalgam*). `Text` is the CSV text: `"Whenever an opposing unit attacks, you may kill this structure. If you do, kill that attacking unit."`. One trigger: `Hook: ability.HookAttacks`, `Watches: true`, `Watch: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}`, `Optional: true`, and an effect with no targets and two steps: `{Op: ability.OpKill, Subject: ability.SubjectSelf}`, then `{Op: ability.OpKill, Subject: ability.SubjectWatched, IfDone: true}`. No new shape, no new field.
- `pkg/engine/cards/cards.go`: remove "Blind Defile" from `NotYetExecutable`.
- `pkg/engine/doc.go`, comment only (T-0241 left it to this task). Line 27 (at 1f6f08d), which reads
  ```go
  //     the battlefield (Trigger.ActiveZone, survey G1, FR-ZON-004). A trigger
  ```
  becomes these four lines:
  ```go
  //     the battlefield (Trigger.ActiveZone, survey G1, FR-ZON-004). A trigger
  //     may watch another asset's attack instead of its own (Trigger.Watches,
  //     Trigger.Watch, FR-CBT-002), and its steps may act on that asset
  //     (SubjectWatched). A trigger
  ```
  Line 28 (`//     may be modal ("choose N" modes, ...`) and everything else in `doc.go` stay as they are.
- The lead dry-ran the definition and the `NotYetExecutable` removal on the run branch at 1f6f08d (2026-11-15, wave H session 2): `gofmt -l pkg cmd` clean, `go tool task check` exit 0 with no existing test edited, `go tool task sim` exit 0, and `go tool task sim-compare -- autoloop/m2-wave-h` `[PASS]` (byte-identical: the Defile is in neither starter deck).

## Tests (card tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/cards_path_defile_test.go`. Use the real card (`s.card("Blind Defile")`) under player 1 in the Main Theater, and a vanilla attacker under player 0 (the active player) in the Main Theater attacking player 1's Headquarters, with health that only the kill can take away. Answer the optional question with `Answer` values through `s.script`; check the asked decisions with `s.asked`. "Dies" means the asset is gone (`!s.alive`) and its card is in its owner's discard pile.
- `TestCard_BlindDefile_AcceptedKillsItselfAndTheAttacker` (#155 (a); `FR-TGT-001` line 394, `FR-CBT-002` Step 2): player 1 accepts the one `DecideOptional`. The Defile and the attacker both die, and player 1's Headquarters takes no combat damage.
- `TestCard_BlindDefile_DeclinedOrTimedOutKillsNothing` (#155 (b); `FR-MOD-001` §6, §7, `SVY-R9`), two subtests: player 1 answers no; seat 1 times out. In each, nothing is placed for the trigger, both assets are alive, and player 1's Headquarters takes the attacker's combat damage.
- `TestCard_BlindDefile_TwoDefilesBothAccepted` (#155 (c); `FR-TGT-001` line 394, `FR-TRG-003`): two Defiles under player 1, one attack, both accepted (and the `DecideOrder` answered). After everything resolves, both Defiles and the attacker have died, and the Headquarters takes no combat damage. The attacker is killed once (one destruction event for it).
- `TestCard_BlindDefile_DefileGoneBeforeResolutionKillsNothing` (#155 (d); `FR-TGT-001` line 394, `FR-ZON-004`): player 1 accepts. After the trigger is on the stack and before it resolves, the Defile leaves the battlefield (as `TestFR_TGT_001_WatchedSubjectGoneDoesNothing` does: `s.declareThen` with `s.g.flushTriggers()` and `s.g.leaveBattlefield`; no M2 card reaches this). The attacker is alive and its attack deals its combat damage to player 1's Headquarters.
- `TestCard_BlindDefile_ReplacedSelfKillStillKillsTheAttacker` (#155 (e); `FR-REP-001` line 521, `SVY-R15`): a test structure under player 1 with a `WouldBeDestroyed` replacement for allied structures that returns the replaced asset to its owner's hand (as in `TestFR_TRN_003a_ReplacedKillSelfCostStaysPaid`, with the filter's kind set to structures). Player 1 accepts. The Defile is in player 1's hand, not the discard pile, and the attacker dies.
- `TestDefs_MatchRulesText` already checks `Text` against the CSV. Nothing to add there.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none expected. If the test author needs one, it names it here.

## Allowed Files
- `pkg/engine/cards/path.go`
- `pkg/engine/cards/cards.go`
- `pkg/engine/doc.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "BlindDefile|TestDefs"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-15 lead-architect (wave H run, session 1): split off T-0241 under the D2 size rule; `queued` on it and on the ruling question committed on `main` (07bd2da).
- 2026-11-15 lead-architect (wave H run, session 2): synced `main` into the run branch (1f6f08d: #155 and `SVY-R49`; only the `Last directive` and `Last task` lines conflicted). The outcome reads `applied`, so #155 is `open: before T-0242`, and this task applies it. Wrote Inputs, Shape, Tests and Allowed Files against the run branch at 1f6f08d, and marked design note Open Question 22 answered in the same commit. Status `tests`.
- 2026-11-15 rules-test-author: wrote `pkg/engine/cards_path_defile_test.go`: `TestCard_BlindDefile_AcceptedKillsItselfAndTheAttacker`, `TestCard_BlindDefile_DeclinedOrTimedOutKillsNothing` (subtests Declined, TimedOut), `TestCard_BlindDefile_TwoDefilesBothAccepted`, `TestCard_BlindDefile_DefileGoneBeforeResolutionKillsNothing`, `TestCard_BlindDefile_ReplacedSelfKillStillKillsTheAttacker`. Helpers requested: none. Cases left out: none. Before the implementation the Accepted, TwoDefiles, DefileGone and Replaced tests fail because the card has no ability; the Declined/TimedOut subtests pass already (the card does nothing), and stay as the guard for the optional question's no-effect path. The two-Defiles test answers by decision kind, so it does not pin when DecideOrder is asked relative to the optional questions. Status `tests-review`.
- 2026-11-15 lead-architect (D4): tests approved. Each checks the outcome `SVY-R49` and #155 (a)–(e) give: (a) both die and no combat damage; (b) declined and timed out place nothing; (c) two Defiles, both die, one `DecideOrder`, one destroy event for the attacker; (d) the Defile gone before resolution leaves the attacker alive and its 3 damage dealt; (e) a replaced self-kill still kills the attacker (`SVY-R15`). Helpers checked: `eventsOfKind` and `EventDestroy` (`"destroy"`) exist, and a `DecideOrder` answer is a pick order. Status `ready`, dispatched to Flash.
