---
id: T-0195
title: One read of "this faction or Neutral", and no face-down guard that can't fire (directive #132 (c), (d), harness, flash, pure move)
kind: harness
tier: flash
wave: D.2
status: queued
branch: m2/autoloop-wave-d.2/T-0195-faction-read-and-face-down-guards
depends_on: [T-0194]
claims: [engine:Card-faction-or-neutral, engine:fireHook-face-down-guard, engine:replacementsFor-face-down-guard]
cards: []
requirements: []
---

## Goal
"This faction or Neutral" is one `Card` method, read by `ValidateDeck` and `FactionDeck`, and the two face-down guards that can never fire are gone (#132 (c), (d)). A pure move: same test names, same sim output. This is the last of #132; T-0194 takes (a) and (b).

## Covers
- **#132 (c):** a `Card` method read by `ValidateDeck` (`pkg/engine/deck.go:75`) and `FactionDeck` (`pkg/sim/sim.go:40`). It goes in `deck.go`, because it is `FR-SET-001`'s rule, which keeps this task at four non-test files (D2.4). `deck.go:52` (a Neutral Headquarters is illegal) is another rule and doesn't read the method.
- **#132 (d):** remove the guards at `pkg/engine/trigger.go:82` and `pkg/engine/replace.go:39`. Check first that `Printed()` (`pkg/engine/state.go:84`) gives a face-down asset a face with no `Ability`, as #132 says.

Write the Inputs, Acceptance Cases, Edits and Allowed Files against the code of the day, before this task leaves `queued`.

## Allowed Files
- `pkg/engine/deck.go`
- `pkg/sim/sim.go`
- `pkg/engine/trigger.go`
- `pkg/engine/replace.go`

## Log
- 2026-11-07, lead-architect (wave D.2 run, session 8): queued with T-0194, which split #132 under D2.4. The next session takes this task from the board.
