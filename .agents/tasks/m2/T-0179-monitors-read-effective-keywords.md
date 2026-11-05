---
id: T-0179
title: Rule monitors read effective keywords (directive #116 (b))
kind: harness
tier: flash
wave: D.1 (board directive #116 (b), structural review 8, finding 1)
status: queued
branch: m2/autoloop-wave-d.1/T-0179-monitors-read-effective-keywords
depends_on: [T-0178]
claims: [monitor:effective keywords]
cards: []
requirements: []
---

## Goal
The rule monitors read the effective keywords T-0178 exposes, so a keyword that differs from the printed one raises no false violation, and damage that breaks the effective value does raise one.

## Covers (write Inputs, Acceptance Cases and Allowed Files against the code of the day before this leaves `queued`)
- Ethereal, Guardian, Versatile and Ward from the snapshot's `Keywords`: `monitor_attack.go`, `monitor_resources.go`, `monitor_targeting.go`.
- Resistant and Stalwart from the damage event, not a snapshot: `monitor_damage.go`.
- The entry-state monitor (`monitor_entry.go`) keeps reading printed keywords, because `newAsset` does.
- #116's acceptance for (b): a monitor test where the two differ raises no violation, and one where the damage breaks the effective value does. A search and a full game still replay. `go tool task sim-compare` passes.
- Existing monitor tests feed hand-built events and snapshots with no keyword fields (`dmgEvent` and `snapAsset` in `pkg/sim`), so the ones that rely on a printed keyword need setup changes. The task names each test and the exact change (D2.4).

## Log
- 2026-11-05, lead-architect (wave D.1 run, session 9): split from #116 under the size rule; T-0178 builds (a). `queued` until T-0178 is `done`.
