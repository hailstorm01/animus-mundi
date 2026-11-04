---
id: T-0161
title: The play screen names an activation, its asset and printed text (harness, flash)
kind: harness
tier: flash
wave: D
status: queued
branch: m2/autoloop-wave-d/T-0161-activate-label
depends_on: [T-0158]
claims: [view:asset-text, label:activate]
cards: []
requirements: [FR-ACT-001]
---

## Goal
A human choosing from `cardtown play`'s action menu reads what an activation does, before the first card with an activated ability becomes executable.

## Covers (Inputs, Acceptance Cases and Allowed Files are written against the code of the day before this task leaves `queued`)
- `view.AssetView` (`pkg/view/view.go`) gains `Text`, the asset's printed rules text, as `CardView.Text` has it, read through `Printed()`. A face-down asset shows the Scratch face's text and so reveals nothing.
- `actionLabel` (`cmd/cardtown/play_screen.go`) returns `Activate: <Name>#<id>: <Text>` for `ActActivate`, and `Activate: Asset#<id>` for an asset the view doesn't hold, as the Enact label's fallback does. `findAsset` walks the view's assets today: make one lookup that returns the `AssetView`, used by `findAsset` and this label, not a second walk.
- Why the whole rules text and not the one ability's line: `ability.Activated` holds no text, and in `data/cards` every card with an activated ability has that ability as its only rules-text line (searched 2026-11-04: *Sinuous Growth*, *Cliff of Portents*, *Boneyard Scavenger*, *Ruin Partisans*, *Rhapsody's Hollow*, and the M3 *Rift* and *Encampment*). So no index is needed in the label yet.

## Log
- 2026-11-04 lead-architect: split off T-0159 under the D2 size rule (D2.4), because the label needs a text the `View` doesn't carry, which made T-0159 five files. Written `queued`.
