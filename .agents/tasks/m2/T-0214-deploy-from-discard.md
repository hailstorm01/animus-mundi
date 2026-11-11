---
id: T-0214
title: A card's own permission to deploy it from its owner's discard pile, with its costs (wave F shape, sonnet)
kind: shape
tier: sonnet
wave: F
status: queued
branch: m2/autoloop-wave-f/T-0214-deploy-from-discard
depends_on: [T-0213]
claims: [shape:discard deploy permission (ability.Static, its deploy-permission kind, ActiveZone and Costs; Def.Statics), field:Action.From, field:Event.From]
cards: []
requirements: [FR-TRN-003, FR-TRN-003a, FR-CST-001]
---

## Goal
A Unit or Structure card that prints a discard-active deploy permission ("You may deploy this card from your discard pile by …") can be deployed by its owner from their discard pile, as a manual Deploy: the stack must be empty, the card moves from the pile onto the stack, its effective cost is paid, and the permission's additional costs are paid after the push through T-0213's deploy placement (`FR-TRN-003`, `FR-TRN-003a`). Directive #128's second part. *Veil Render* (T-0216) is the first card to print it.

Covers (written in full against the code of the day before it leaves `queued`):
- `ability.Static` (the design note's Key types: a discard-active deploy permission, with `ActiveZone` and its `Costs`) and `Def.Statics`. Only the deploy-permission kind; auras stay deferred (#14 (b)).
- `Action.From` on `ActDeploy` (the design note's Actions paragraph): `ZoneDiscard` for the permission; the zero value keeps meaning the hand, so every existing `ActDeploy` stays valid.
- `check` and `Apply`: `handPlay`'s find, supported and pay checks and `playFromHand`'s find, pay and take read `From`, one function each, not a second copy for the pile (#128's Fix). A discard deploy needs the card in its owner's own discard pile with a permission active there. `deployPlacement` adds the permission's costs to `ExtraCosts`.
- `LegalActions` offers a discard deploy for each such card and each of its deploy zones.
- The action event records the origin: an `Event.From` field (`omitempty`, so every existing log is byte-identical), and `emitAction` finds the card in the pile it came from. The lead adds the ADR-0005 addendum when this task merges.
- Tests to pin (#128's Acceptance): a discard deploy pays its kill cost after the push, with the card already on the stack when the cost's death triggers are queued (`FR-TRN-003`'s example), and a replaced kill stays paid (`SVY-R15`); no discard deploy without the permission, with a non-empty stack, from the opponent's pile, or when the cost can't be paid; `EffectiveCost`'s hand-only reduction doesn't apply in the pile; a free deploy (`OpDeployFree`, `FR-TRN-003a`, `FR-CST-001`) of a card with the permission pays none of its costs.
- Synthetic cards only.

- Not this task: replay, the play screen and `ProgressPolicy` (T-0215); *Veil Render* (T-0216).

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/actions.go`
- `pkg/engine/placement.go`
- `pkg/engine/events.go`

## Log
- 2026-11-10 lead-architect (wave F run, session 13): queued with T-0213, split off *Veil Render* under the D2 size rule (D2.4). Depends on T-0213's deploy placement.
