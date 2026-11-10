---
id: T-0208
title: OpDeployFree, a free deploy from a discard pile during resolution (wave F shape, sonnet)
kind: shape
tier: sonnet
wave: F
status: tests
branch: m2/autoloop-wave-f/T-0208-deploy-free
depends_on: []
claims: [shape:OpDeployFree (ability.OpDeployFree; stepDeployFree; the free-deploy zone decision DecideZone)]
cards: []
requirements: [FR-TRN-003a, FR-CST-001, FR-MOD-001, FR-ZON-002, FR-ZON-003, FR-TGT-003, FR-KWD-002l]
---

## Goal
The `DEPLOY_FREE` primitive (`FR-TRN-003a`, "In-Resolution Free Deployments"): a resolving step deploys a chosen card from a discard pile onto the battlefield at once, with no stack item and no costs, into its printed zone, a zone its controller chooses when it prints both Reserves and the Main Theater (a timeout takes Reserves, `SVY-R43`), or the zone the step names. Wave F's free deploys all reuse it: *Reaver's Rush* (T-0210), *Allshan's Battle Trance* (with `Step.To`), *Devilhorn Scouting Party* (from a look) and *Eerie Remnant*. It enters through the one `deployCard` (#95), so deploy triggers fire and nothing counts as a move (`SVY-R1`).

Covers:
- `ability.OpDeployFree`, acting on cards (`stepCards`: a pile target slot or a pile pick), and `Step.To` read by it as an explicit zone (`FR-TRN-003a` 2).
- `stepDeployFree` in `effects.go`, dispatched from `runStep` by one `case` line.
- A `DecideZone` decision kind for the zone choice, with the `FR-MOD-001` §6 fallback "first listed", and Reserves always listed first (`SVY-R43`).
- `doc.go`'s step list names `OpDeployFree`, and its decision list names `DecideZone`.
- Synthetic-card tests only. No card prints it until T-0210.

- Not this task: the power criterion "with power 3 or less" (T-0209, `queued`); *Reaver's Rush* and directive #140 (4)'s *Devilhorn Guerrillas* test (T-0210, `queued`); a deploy from a look at the top cards (#140 (3), the *Devilhorn Scouting Party* shape task); a deploy from a discard pile by a card's own permission (#128, *Veil Render*).

## Inputs
- FRD 01 v3.23: `FR-TRN-003, FR-TRN-003a & FR-MNV-001` (lines 356–362), above all "In-Resolution Free Deployments (`DEPLOY_FREE`) & Destination Zone Rules" (line 360) and its items 1 (unspecified zone: the printed Deployment Zones, the deploying player chooses between Reserves and the Main Theater, a timeout enters Reserves) and 2 (an explicit zone overrides the printed ones); `FR-CST-001` (line 316: "for free" waives every resource and additional cost); `FR-MOD-001` §6 "Free-Deploy Zone (`SVY-R43`)" (line 416); `FR-ZON-002` "What Counts as Moving (`SVY-R1`)" (line 237: a free deploy into a zone is not a move); `FR-ZON-003` (line 213: the controller is the player who deployed it).
- Rulings: `SVY-R1`, `SVY-R43`; outcome `docs/design/outcomes/2026-11-09-wave-f-keyword-grant-duration-and-inspection-timeout.md` (#140).
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Step`, line 40, names `DeployFree`), Effect ops (line 61).
- Code: `deployCard` ([zones.go:38-49](../../../pkg/engine/zones.go)), the one deploy, which fires deploy triggers and applies Deploys Exhausted through `newAsset`; `runStep` ([effects.go:76-101](../../../pkg/engine/effects.go)); `stepRelocate` ([effects.go:275-298](../../../pkg/engine/effects.go)) for how a card step checks the card is still in its pile (`g.DiscardCard(inst.ID) != inst`) and takes it (`takeCard`); `stepCards` ([effects.go:305-324](../../../pkg/engine/effects.go)); `battlefieldZone` ([zones.go:247](../../../pkg/engine/zones.go)); `Card.DeployZones` and `Supported` ([card.go:93](../../../pkg/engine/card.go), [:122](../../../pkg/engine/card.go)); `handPlay`'s "needs rules the kernel does not implement yet" check ([actions.go:188](../../../pkg/engine/actions.go)); `DecisionKind`, its `String`, and `fallbackPicks` ([choice.go:11-26](../../../pkg/engine/choice.go), [:132-142](../../../pkg/engine/choice.go)); `ability.Op` and `Step.To` ([ability.go:272-287](../../../pkg/engine/ability/ability.go), [:312](../../../pkg/engine/ability/ability.go)).

### Shape (the engineer builds it)
- **`ability.OpDeployFree`**, the next `Op` after `OpSearchDeck`. Its comment: put each Subject card, still in its discard pile, onto the battlefield for free under the ability's controller, with no stack item and no costs (`FR-TRN-003a`, `FR-CST-001`): into `Step.To` if set, else its printed zone, chosen by the controller when it prints both Reserves and the Main Theater (`SVY-R43`). `Step.To`'s comment names both ops (`OpMoveZone`, and `OpDeployFree`, where it overrides the card's printed Deployment Zones, `FR-TRN-003a` 2).
- **`stepDeployFree(r, st)`** in `effects.go`, one `case` line in `runStep`. For each card of `stepCards(r, st)`, in order:
  - skip it unless it is still in its owner's discard pile (`g.DiscardCard(inst.ID) == inst`, as `stepRelocate` checks), is a Unit or Structure card, and is `Supported()` (as a deploy from hand requires, `handPlay`: the kernel never puts a card whose text it can't run onto the battlefield);
  - pick the zone: `battlefieldZone(st.To)` if that holds; else the card's printed `DeployZones` limited to Reserves and the Main Theater. With only one, that one, asked of nobody. With both, one `DecideZone` decision to the ability's controller: `Min` and `Max` 1, options Reserves first and the Main Theater second whatever order the card prints them in (`Label` the zone's `String()`, `Ref` 0), not `Hidden`. With neither, skip the card;
  - `g.takeCard(inst)`, then `g.deployCard(inst, r.it.Controller, zone)`.

  It returns how many cards it deployed.
- **`DecideZone`** in `choice.go`, appended after `DecideSearch` (its `String` is `"Zone"`): "choose the zone a free-deployed card enters (`FR-TRN-003a` 1): fallback takes the first listed, Reserves (`SVY-R43`)". `fallbackPicks` adds it to the first-listed case.
- **`doc.go`:** the step list names `OpDeployFree` (a free deploy from a discard pile, `FR-TRN-003a`), and the decisions line names the zone choice (`DecideZone`, `SVY-R43`).
- Behavior doesn't change for any printed card: none uses `OpDeployFree`. `go tool task sim-compare -- autoloop/m2-wave-f` must be byte-identical (a gate here).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/deploy_free_test.go`, with a synthetic Maneuver (`maneuver`, `maneuver_test.go`) for player 0 whose one target slot is `ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}` and whose one step is `{Op: ability.OpDeployFree, Subject: ability.SubjectTarget}` (with `To` where a test says so), and synthetic unit cards put into player 0's discard pile with `s.inDiscard`, their `DeployZones` set per test. Enact with `enactAction` and resolve the stack. Script answers with `s.script` and read the asked decisions with `s.asked(p)`.
- `TestFR_TRN_003a_DeployFreeEntersFromDiscardWithoutCosts` (`FR-TRN-003a`, `FR-CST-001`, `FR-ZON-003`): a Reserves-only unit card with a Requisition and a Tactics cost in the discard pile: after resolution it is an asset controlled by player 0 in player 0's Reserves, it is no longer in the discard pile, player 0's Requisition and Tactics dropped by the Maneuver's own cost only, and no `place` event names the unit card (no stack item for the deploy). No `DecideZone` was asked.
- `TestFR_TRN_003a_DeployFreeAsksZoneWhenBothPrinted` (`FR-TRN-003a` 1, `SVY-R43`): a unit card printing `ZoneMainTheater, ZoneReserves` in that order: player 0 is asked one `DecideZone` with options Reserves then Main Theater; answering the second puts the asset in the Main Theater.
- `TestFR_MOD_001_DeployFreeZoneTimeoutEntersReserves` (`SVY-R43`, `FR-MOD-001` §6): the same card, the `DecideZone` answered with a timeout: the asset is in player 0's Reserves.
- `TestFR_TRN_003a_DeployFreeExplicitZoneOverridesPrinted` (`FR-TRN-003a` 2): a Reserves-only unit card and a step with `To: ability.ZoneMainTheater`: the asset is in the Main Theater, and no `DecideZone` was asked. The same with a card printing both zones and `To: ability.ZoneReserves`: Reserves, nothing asked.
- `TestFR_TRN_003a_DeployFreeFiresDeployTriggersNotMoves` (`FR-TRN-003a`, `SVY-R1`): the deployed unit card prints a `HookDeploy` trigger (for example `OpDraw` 1 for its controller) and a `HookMoves` trigger (for example `OpAddArmor` 1 on itself); a Main Theater deploy (by `To`): the deploy trigger resolves after the Maneuver (player 0 drew one card), and the move trigger never fires (the asset has no armor).
- `TestFR_TRN_003a_DeployFreeKeepsDeploysExhausted` (`FR-KWD-002l`): a unit card with Deploys Exhausted enters Exhausted.
- `TestFR_TGT_003_DeployFreeTargetLeftPileDeploysNothing` (`FR-TGT-003`): after the Maneuver is placed with its target chosen, the card leaves the discard pile (for example moved to player 0's hand with `s.g.takeCard` and `s.g.putCard`) before resolution: nothing is deployed and the card stays where it went.
- `TestFR_TRN_003a_DeployFreeSkipsUnsupportedCard`: the target card has a line in `Unsupported`: nothing is deployed, the card stays in the discard pile, and the Maneuver still goes to the discard pile.
- **Existing tests:** every other test passes unchanged.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test author keeps any small builder in the test file.

## Allowed Files
- `pkg/engine/ability/ability.go` (`OpDeployFree`, `Step.To`'s comment)
- `pkg/engine/effects.go` (`stepDeployFree`, its `runStep` case)
- `pkg/engine/choice.go` (`DecideZone`, its `String`, `fallbackPicks`)
- `pkg/engine/doc.go` (the step and decision lists)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "DeployFree"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f` passes (no printed card uses `OpDeployFree`)

## Log
- 2026-11-10 lead-architect (wave F run, session 8): written against the run branch at 5458b2e, the first of wave F's free deploys (*Reaver's Rush*, `FR-TRN-003a`). Searched `pkg/engine` and the task archive for an existing equivalent: `deployCard` (#95, T-0146) is the one deploy, reused here; no step deploys a card, and no decision chooses a zone. *Reaver's Rush* needs a second shape, a power criterion, so under the D2 size rule it splits into this task, T-0209 (the power criterion, `queued`) and T-0210 (*Reaver's Rush*, card, flash, `queued`). Set to `tests`.
- 2026-11-10 rules-test-author: wrote `pkg/engine/deploy_free_test.go` (8 tests, one with two subtests): `TestFR_TRN_003a_DeployFreeEntersFromDiscardWithoutCosts`, `TestFR_TRN_003a_DeployFreeAsksZoneWhenBothPrinted`, `TestFR_MOD_001_DeployFreeZoneTimeoutEntersReserves`, `TestFR_TRN_003a_DeployFreeExplicitZoneOverridesPrinted`, `TestFR_TRN_003a_DeployFreeFiresDeployTriggersNotMoves`, `TestFR_TRN_003a_DeployFreeKeepsDeploysExhausted`, `TestFR_TGT_003_DeployFreeTargetLeftPileDeploysNothing`, `TestFR_TRN_003a_DeployFreeSkipsUnsupportedCard`. Helpers requested: none (small builders `dfManeuver`, `dfScene`, `dfUnit`, `dfZoneAsks` live in the test file). `go vet` fails only on the missing `ability.OpDeployFree` and `DecideZone`; `cardtown lint rules` passes. Left out: Structure cards and a card printing neither zone (skipped by the shape, but FRD 01 does not pin the case beyond the Reserves/Main Theater rule), and several cards in one step (no FRD 01 ordering rule).
- 2026-11-10 lead-architect (wave F run, session 8): tests reviewed (a8657cf), returned to `tests`. Notes:
  1. A single legal pile target is chosen without a question: `choose` takes the sole-answer fast path ([choice.go:104-109](../../../pkg/engine/choice.go); FR-MOD-001 §1, pinned by `TestFR_MOD_001_SinglePileTargetIsNotAsked`). With one card in the pile, the leading `Answer{Picks: []int{0}}` that `dfScene` scripts for the target, and that `TestFR_TRN_003a_DeployFreeFiresDeployTriggersNotMoves` and `TestFR_TGT_003_DeployFreeTargetLeftPileDeploysNothing` script directly, is never used for the target: the `DecideZone` takes it instead. So `TestFR_TRN_003a_DeployFreeAsksZoneWhenBothPrinted` gets Reserves on a correct build and fails, and `TestFR_MOD_001_DeployFreeZoneTimeoutEntersReserves` passes without ever timing out. Drop that leading answer everywhere: script only the decisions that are actually asked (none where no zone is asked, `s.script(0)`), so the chooser's past-the-end error also catches an unexpected question.
  2. In `TestFR_MOD_001_DeployFreeZoneTimeoutEntersReserves` and `TestFR_TRN_003a_DeployFreeAsksZoneWhenBothPrinted`, also check that the `DecideZone` is the only decision player 0 was asked (`len(s.asked(0)) == 1`), so the zone answer can't go to another question.
  The rest stands: each test matches the Tests list and `FR-TRN-003a`, `SVY-R1`, `SVY-R43`. A Structure card is not skipped by the shape (it says Unit or Structure); no test for it is needed.
