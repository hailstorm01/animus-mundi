---
id: T-0160
title: The docs say activated abilities are built, stack.go, doc.go and costs.go comments (harness, flash, comments only)
kind: harness
tier: flash
wave: D
status: queued
branch: m2/autoloop-wave-d/T-0160-activated-comments
depends_on: [T-0158]
claims: [comment:activated]
cards: []
requirements: []
---

## Goal
No comment still says that a `StackAbility` item is only a trigger, or that activated abilities are not built.

## Covers (Inputs, Acceptance Cases and Allowed Files are written against the code of the day before this task leaves `queued`)
- `StackAbility`'s comment (`pkg/engine/stack.go:15`, "a triggered ability") and the `Tracked` and `Keyword` field comments as they read then.
- `doc.go`'s not-built list (`pkg/engine/doc.go:55`, "Activated and continuous abilities") and a line on what T-0158 built.
- `costsPayable`'s and `payCosts`' doc comments and the file comment (`pkg/engine/costs.go`), which still speak only of additional costs and don't name `CostResources`, `CostExhaustSelf` or `CostKillSelf` (T-0158's review).

## Log
- 2026-11-04 lead-architect: split off T-0158 under the D2 size rule (D2.4). Written `queued`.
