---
id: T-0159
title: The event log, replay, sim policy and play screen know an activation (harness, sonnet)
kind: harness
tier: sonnet
wave: D
status: queued
branch: m2/autoloop-wave-d/T-0159-activate-clients
depends_on: [T-0158]
claims: [client:activate]
cards: []
requirements: [FR-ACT-001]
---

## Goal
Everything outside the kernel's action path handles `ActActivate` before the first card with an activated ability becomes executable, so that card's task only combines shapes.

## Covers (Inputs, Acceptance Cases and Allowed Files are written against the code of the day before this task leaves `queued`)
- `emitAction` (`pkg/engine/events.go`) records the activating asset (`Source`) and the ability index, in a field `actionOf` can read back.
- `actionOf` (`pkg/eventlog/replay.go`) rebuilds `Action.Ability`, so a replay reproduces an activation.
- `ProgressPolicy` (`pkg/sim/policy.go`) can choose an activation, with the other acts.
- `actionLabel` (`cmd/cardtown/play_screen.go`) gives an `Activate:` label naming the asset and the ability's printed text.
- Check that the rule monitors (`pkg/sim/monitor*.go`) don't flag an activation's costs, or its exhaust or kill, and say what you found in the Log. `pkg/view` already shows a `StackAbility` item, so check that its targets show for an activated ability too.
- The action-event test T-0158's edit to `TestEvents_Case4_EveryActionKind` names: `TestEvents_Case4_ActivateActionEventNamesItsAsset`, with that comment updated.

## Log
- 2026-11-04 lead-architect: split off T-0158 under the D2 size rule (D2.4). Written `queued`.
