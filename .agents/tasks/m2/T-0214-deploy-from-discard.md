---
id: T-0214
title: A card's own permission to deploy it from its owner's discard pile, with its costs (wave F shape, sonnet)
kind: shape
tier: sonnet
wave: F
status: ready
branch: m2/autoloop-wave-f/T-0214-deploy-from-discard
depends_on: [T-0213]
claims: [shape:discard deploy permission (ability.Static, ability.StaticKind, StaticDeployPermission, Static.ActiveZone and Static.Costs; Def.Statics; deployPermission), field:Action.From, field:Event.From]
cards: []
requirements: [FR-TRN-003, FR-TRN-003a, FR-CST-001]
---

## Goal
A Unit or Structure card that prints a discard-active deploy permission ("You may deploy this card from your discard pile by …") can be deployed by its owner from their own discard pile, as a manual Deploy: the stack must be empty, the card moves from the pile onto the stack, its effective cost is paid, and the permission's additional costs are paid after the push through T-0213's deploy placement, with the triggers they fire on top of it (`FR-TRN-003`, `FR-TRN-003a`). This is directive #128's second part. *Veil Render* (T-0216) is the first card to print it.

Covers:
- `ability.Static` with its kind `StaticDeployPermission`, `ActiveZone` and `Costs`, and `Def.Statics`. Only the deploy-permission kind; auras stay deferred (#14 (b)).
- `Action.From` on `ActDeploy`. The zero value means the hand, so every existing `ActDeploy` stays valid.
- `check` and `Apply` read `From` through one find, one play check and one pay-and-take, shared by Deploy and Enact, not a second copy for the pile (#128's Fix).
- `deployPlacement` adds the permission's costs to the card's `ExtraCosts` for a discard deploy.
- `LegalActions` offers a discard deploy for each such card and each of its deploy zones.
- The action event records the origin in a new `Event.From` (`omitempty`, so every existing log is byte-identical). The lead adds the ADR-0005 addendum when this task merges.
- Synthetic cards only. No printed card has the permission until T-0216, so `sim-compare` is byte-identical.

- Not this task: replay (`actionOf` rebuilding `From`), the play screen's label and `ProgressPolicy` (T-0215, `queued`); *Veil Render* and `doc.go`'s shape list (T-0216, `queued`).

## Inputs
- FRD 01 v3.23: `FR-TRN-003` (line 357: a card may be deployed from the Discard pile "when explicitly permitted by a static ability on that card"; the stack must be empty; the card moves from its origin zone to the Resolution Stack on declaration and all effective resource and additional costs are paid upfront; a cost's triggers evaluate with the declared card already on the stack, not in its origin zone); `FR-TRN-003a` (line 359: additional costs of a card declared "from Hand or Discard" are paid on declaration and their triggers go on top of it; a replaced cost still counts as paid, `SVY-R15`; line 360: a free deploy excludes all resource and additional costs, and the kill cost "applies only when manually declaring deployment via its own static permission"); `FR-CST-001` (lines 311–317: a `DEPLOY_FREE` waives every additional cost; a static deployment permission with `activeZone: DISCARD` is evaluated on demand on the card in its current zone).
- Board directive #128 (`BOARD.md`): Fix and Acceptance.
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Static`, line 43: "a discard-active deploy permission, with `ActiveZone`") and Changes in `engine` (Actions, line 54: "a `From` zone on `ActDeploy` for discard permissions").
- Code: `Action`, `Apply`, `check`, `handPlay`, `playFromHand` and `LegalActions` ([actions.go](../../../pkg/engine/actions.go)); `placement`, `deployPlacement` and `declare` ([placement.go](../../../pkg/engine/placement.go)); `EffectiveCost` (its reduction applies only while the card is in its owner's hand, [costs.go:17-28](../../../pkg/engine/costs.go)), `payCost`, `costsPayable`, `payCosts`; `takeCard` (already takes from the discard pile, [zones.go:126](../../../pkg/engine/zones.go)); `stepDeployFree` ([effects.go:116](../../../pkg/engine/effects.go)); `emitAction` and `Event` ([events.go:144](../../../pkg/engine/events.go), [:222](../../../pkg/engine/events.go)); `Def` ([ability.go:17](../../../pkg/engine/ability/ability.go)); a hand-active `CostReduction` as *Boneyard Amalgam* prints it ([cards/path.go:151](../../../pkg/engine/cards/path.go)).
- Model tests: `deploy_cost_test.go` (T-0213: a deploy's kill cost after the push, `killCostUnit`, `dcAssetOf`), `deploy_free_test.go` (a free deploy from a pile, T-0208), `kill_cost_test.go`, and `activated_test.go:121` (a non-empty stack set up directly).

### Shape (the engineer builds it)
- **`ability.StaticKind`**, no default (`iota + 1`), with one member `StaticDeployPermission`: "You may deploy this card from your discard pile by …": its owner may deploy it from `ActiveZone` as a manual Deploy, paying `Costs` in addition to its deployment costs (`FR-TRN-003`, `FR-CST-001`).
- **`ability.Static`** `{Kind StaticKind; ActiveZone Zone; Costs []Cost}`: one printed static ability. `ActiveZone` is the zone the card must be in for it to apply (`ZoneDiscard` for a deploy permission). `Costs` are a deploy permission's additional costs, in printed order, paid on declaration (`FR-TRN-003a`); `Cost`'s filters stay battlefield filters, relative to the payer.
- **`Def.Statics []Static`**, after `CostReduction`: the printed static abilities, in printed order.
- **`Action.From Zone`** (Deploy): the pile the card is deployed from. The zero value and `ZoneHand` both mean the hand; `ZoneDiscard` means its owner's discard pile; any other value is illegal. Enact ignores it.
- **One find, one play check, one pay-and-take**, in `actions.go`, replacing `handPlay` and `playFromHand` (names are the engineer's; `handCard` stays, `setup.go` uses it):
  - a lookup of the card with instance ID `id` in player `p`'s own pile named by a `From` (the hand, or the discard pile). A card in another player's pile is never found.
  - the play check: found, `Supported`, and `EffectiveCost` payable. Deploy passes `a.From`, Enact the hand.
  - the pay-and-take: find, `payCost` (read while the card is still in its pile, so a hand-only reduction doesn't apply to a pile card), `takeCard`.
- **`check`, `ActDeploy`:** after the play check, a discard deploy is illegal unless the card has a deploy permission active in the discard pile (`deployPermission`), with a message naming the card and `FR-TRN-003`. The type, zone and `placeable` checks follow as now, with `deployPlacement` given the origin.
- **`deployPermission(c *Card) *ability.Static`**, in `placement.go`: the card's first `Static` of kind `StaticDeployPermission` with `ActiveZone` `ability.ZoneDiscard`, or nil. It is the one reader of the permission: `check` and `deployPlacement` both call it.
- **`deployPlacement(pid, inst, zone, from Zone)`:** costs are the card's `ExtraCosts`, then, for a discard deploy, the permission's `Costs`. Build a new slice (`slices.Concat`): never append to a card definition's slice.
- **`Apply`, `ActDeploy`:** the pay-and-take with `a.From`, then `declare(deployPlacement(a.Player, inst, a.Zone, a.From), ability.Effect{}, StackDeploy)`.
- **`LegalActions`:** after the hand deploys, for each card in the active player's discard pile and each of its `DeployZones`, `try` an `ActDeploy` with `From: ZoneDiscard`. `check` rejects a card without the permission, so with no printed card having it the list is unchanged.
- **`Event.From string`** (`json:"from,omitempty"`), next to `Zone`: an action event's Deploy origin, `ZoneDiscard.String()` (`"Discard"`) for a discard deploy, empty for the hand. **`emitAction`** finds a Deploy's card through the same lookup with `a.From` (an Enact's in the hand), and sets `From` for a discard deploy.
- Nothing else changes: `resolveDeploy`, `deployCard`, `stepDeployFree` (a free deploy has no placement and pays no cost, `FR-CST-001`) and every existing event stay as they are.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/deploy_discard_test.go`. Synthetic cards: a Unit for player 0 (`unit`, costing R2 T1) whose `Ability` has one `Statics` entry `{Kind: ability.StaticDeployPermission, ActiveZone: ability.ZoneDiscard, Costs: []ability.Cost{{Kind: ability.CostKill, Filter: alliedUnitFilter}}}` and nothing else, put in player 0's discard pile (`inDiscard`), with enough resources, deployed with `Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves, From: ZoneDiscard}`. Model them on `deploy_cost_test.go` (reuse its helpers where they fit). Script only the decisions actually asked: a single option is taken unasked (`FR-MOD-001` §1).
- `TestFR_TRN_003_DiscardDeployPaysItsKillCostAfterThePush` (`FR-TRN-003`, `FR-TRN-003a`): two allied units, one of them printing a death trigger with one "up to 1" target slot over cards in its controller's discard pile (`Filter{Allegiance: AllegianceAllied, Zone: ZoneDiscard}`) and a harmless step. Player 0 picks that unit for the cost. After the action: the unit is destroyed and in player 0's discard pile, the other unit is untouched, the deployed card is an asset in player 0's Reserves and no longer in the discard pile, and player 0 paid R2 T1. The deploy's `place` event comes before the cost's `destroy` event, and the death trigger's `place` event comes after the deploy's `place` and before its `deploy` event (on top, resolved first). The death trigger's target decision lists the killed unit's card but not the deployed card, which is already on the stack (`FR-TRN-003`'s example). The action event has `From` `"Discard"`, and the deployed card's name and instance ID.
- `TestFR_TRN_003a_DiscardDeployReplacedKillStaysPaid` (`SVY-R15`, `FR-REP-001`): the cost's kill is replaced (as in `TestFR_TRN_003a_DeployReplacedKillCostStaysPaid`), and the deploy still resolves: the card is an asset in player 0's Reserves.
- `TestFR_TRN_003_DiscardDeployNeedsThePermission`: a Unit in player 0's discard pile with no `Statics`, everything else payable: no discard deploy of it is in `LegalActions`, `Apply` returns `ErrIllegal`, and the card stays in the pile. A `Static` of the same kind with another `ActiveZone` (for example `ability.ZoneReserves`) gives no permission either.
- `TestFR_TRN_003_DiscardDeployOnlyFromYourOwnPile`: the permission card in player 1's discard pile; player 0's `ActDeploy` with its ID and `From: ZoneDiscard` is illegal and not offered, and the card stays in player 1's pile.
- `TestFR_TRN_003_DiscardDeployNeedsAnEmptyStack` (`FR-QUE-001`): with an item on the stack (set up directly, as `activated_test.go:121` does), the discard deploy is illegal.
- `TestFR_TRN_003a_DiscardDeployCostsMustBePayable` (`SVY-R29`): with no allied unit (only player 0's Headquarters), the discard deploy is not offered and is illegal, and the card and player 0's pool are unchanged. With an allied unit but one Requisition short of R2, it is illegal too.
- `TestFR_CST_001_HandReductionDoesNotApplyInThePile`: a permission card that also prints a hand-active `CostReduction` (`Req` 1 per allied unit card in a discard pile, as *Boneyard Amalgam* prints it), in player 0's discard pile with two other unit cards: `EffectiveCost` reads the printed cost there, a discard deploy with one Requisition less than printed is illegal, and with the printed cost it pays the printed cost.
- `TestFR_TRN_003_HandDeployIgnoresThePermission` (`FR-TRN-003`): the permission card deployed from hand (`From` zero) asks for no kill and kills nothing; it is an asset in Reserves and only its R2 T1 was paid. The action event has an empty `From`.
- `TestFR_TRN_003_DiscardDeployOfferedForEachZone`: a permission card printing both Reserves and the Main Theater: `LegalActions` offers two discard deploys, one per zone, each with `From: ZoneDiscard`; deploying to the Main Theater puts it there.
- `TestFR_TRN_003_DeployFromAnotherZoneIsIllegal`: an `ActDeploy` of a hand card with `From: ZoneMainTheater` is illegal.
- `TestFR_CST_001_FreeDeployWaivesThePermissionCost` (`FR-TRN-003a` line 360, `FR-CST-001`): an `OpDeployFree` step (as `deploy_free_test.go` builds one) deploys the permission card from player 0's discard pile: no kill is asked, every allied unit survives, and player 0's pool is unchanged by the deploy.
- **Existing tests:** every other test passes unchanged.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test author keeps any small builder in the test file.

## Allowed Files
- `pkg/engine/ability/ability.go` (`StaticKind`, `Static`, `Def.Statics`)
- `pkg/engine/actions.go` (`Action.From`, the find, play check and pay-and-take, `check`, `Apply`, `LegalActions`)
- `pkg/engine/placement.go` (`deployPermission`, `deployPlacement`)
- `pkg/engine/events.go` (`Event.From`, `emitAction`)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "DiscardDeploy|HandReductionDoesNotApplyInThePile|HandDeployIgnoresThePermission|DeployFromAnotherZoneIsIllegal|FreeDeployWaivesThePermissionCost"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f` passes (no printed card has the permission)

## Log
- 2026-11-10 lead-architect (wave F run, session 13): queued with T-0213, split off *Veil Render* under the D2 size rule (D2.4). Depends on T-0213's deploy placement.
- 2026-11-10 lead-architect (wave F run, session 14): written in full against the run branch at 7f290cd. Searched `pkg/engine` and the task archive for an existing equivalent: no `Static`, no deploy permission and no `From` exist; T-0213's `deployPlacement` and `declare` are the one declare sequence, and this task extends them. `handPlay` and `playFromHand` become one find, play check and pay-and-take that read the origin, so Deploy and Enact keep sharing them (#128's Fix). `pkg/engine/doc.go`'s shape list would be a fifth non-test file, so its line for `Def.Statics` goes into T-0216's Covers. Set to `tests`.
- 2026-11-10 rules-test-author: wrote pkg/engine/deploy_discard_test.go (df50f41), 11 tests: TestFR_TRN_003_DiscardDeployPaysItsKillCostAfterThePush, TestFR_TRN_003a_DiscardDeployReplacedKillStaysPaid, TestFR_TRN_003_DiscardDeployNeedsThePermission, TestFR_TRN_003_DiscardDeployOnlyFromYourOwnPile, TestFR_TRN_003_DiscardDeployNeedsAnEmptyStack, TestFR_TRN_003a_DiscardDeployCostsMustBePayable, TestFR_CST_001_HandReductionDoesNotApplyInThePile, TestFR_TRN_003_HandDeployIgnoresThePermission, TestFR_TRN_003_DiscardDeployOfferedForEachZone, TestFR_TRN_003_DeployFromAnotherZoneIsIllegal, TestFR_CST_001_FreeDeployWaivesThePermissionCost. Helpers requested: none (local helpers ddPermission, ddUnit, ddDeploy, ddOffers; reuses dcAssetOf, kcScript, dfManeuver, alliedUnitFilter). The file does not compile until the engineer builds ability.Static, ability.StaticDeployPermission, Def.Statics, Action.From and Event.From (go vet fails on undefined ability.Static, as expected); cardtown lint rules passes. No cases left out. Assumption: the death trigger's up-to-1 target slot is asked as a DecideTarget even with one legal option (zero is also an option).
- 2026-11-10 lead-architect (wave F run, session 14): tests reviewed (df50f41, 11 tests). A lead probe (the four Allowed Files edited as the Shape says, since discarded) passes all 11 and every `pkg/...` test. Each test matches the Tests list and `FR-TRN-003`, `FR-TRN-003a`, `FR-CST-001`, `SVY-R15` and `SVY-R29`. The author's assumption holds: an "up to 1" slot is asked as a `DecideTarget` with one legal option, since none is also a choice. Set to `ready`.
