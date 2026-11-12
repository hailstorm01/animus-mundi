---
id: T-0225
title: "A trigger from a card in its discard pile, first as Eerie Remnant (wave F shape, sonnet)"
kind: shape
tier: sonnet
wave: F
status: queued
branch: m2/autoloop-wave-f/T-0225-eerie-remnant
depends_on: [T-0224]
claims: [field:ability.Trigger.ActiveZone, hook:put into its owner's discard pile from hand, func:the pile trigger source (beside assetSource), the pile source's rank in sourcePos]
cards: [Eerie Remnant]
requirements: [FR-ZON-004, FR-TRN-003a, FR-TRG-002, FR-TRG-003]
---

## Goal
*Eerie Remnant* ("Deploys Exhausted / When this card is put into your discard pile from your hand, you may deploy it from your discard pile for free.", `data/cards/units-and-structures-path.csv:4`), the last wave F deck card on `NotYetExecutable` (3 copies in `data/decks/path-starter.txt`). Its trigger works from the discard pile: survey gap G1, `Trigger.ActiveZone` in the design note's Key types (`docs/architecture/m2-ability-model.md:32`), not in `ability` yet.

Covers (written 2026-11-12 with T-0224; Inputs, Shape, Tests and Allowed Files are written against the code of the day before this task leaves `queued`):
- `ability.Trigger.ActiveZone`, battlefield by default (G1), and a hook for "put into your discard pile from your hand", fired from `discardCards` (`pkg/engine/zones.go:114`), the one path every discard from hand takes.
- A trigger source built from a card in its owner's discard pile, controlled by its owner, ranked after every battlefield asset in the flush's source order (ADR-0002's fixed order), fired through T-0224's one `fireHook` (#136's card part). Decide there what its `id` (the placement's source and the `MaxPerTurn` key) is for a card with no asset; no M2 discard-active trigger has a `MaxPerTurn`.
- "Deploy it from your discard pile for free": T-0208's `OpDeployFree` on the trigger's own card, if it is still in its owner's discard pile on resolution (`FR-ZON-004`; FRD 01 line 251, self-relocating trigger tracking). Optional (`SVY-R9`). Deploys Exhausted is already built (`TestFR_TRN_003a_DeployFreeKeepsDeploysExhausted`).
- The card's definition, `doc.go`'s shape list, and the removal of *Eerie Remnant* from `cards.NotYetExecutable`.
- Size: this likely exceeds four non-test files (`ability`, `trigger.go`, `zones.go`, the effect's subject, `cards/path.go`, `doc.go`). If so, split under the D2 size rule, the card going to its own `card`/`flash` task, both on the board in the same commit.

## Log
- 2026-11-12, lead-architect (wave F run, session 26): queued with Covers only. Writing it fired #136, now `open: before T-0225`; T-0224 is its pure move and goes first.
