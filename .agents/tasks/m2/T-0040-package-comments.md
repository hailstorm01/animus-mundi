---
id: T-0040
title: Package comments for sim and eventlog (directive #33, harness, flash, comments only)
kind: harness
tier: flash
wave: none (board directive #33, structural review cleanup 3 of 7, before cardtown play)
status: done
branch: m2/T-0040-package-comments
depends_on: [T-0039]
claims: [doc:pkg/sim, doc:pkg/eventlog]
cards: []
requirements: []
---

## Goal
`go doc ./pkg/sim` and `go doc ./pkg/eventlog` print a package paragraph that says what each package holds and what it may import. They print only the package line today. Nothing else changes.

## Inputs
- Board directive #33 (structural finding 5), corrected. The review said five subpackages had no package comment, but `ability`, `cards` and `view` do have one. The review's command (`go doc … | sed -n '1,/^$/p'`) stopped at the first blank line and hid it. Only `sim` and `eventlog` lack one.
- Convention: `ability`, `cards` and `view` keep their package comment at the top of their main file (`ability.go`, `cards.go`, `view.go`), directly above `package`. `engine` uses `doc.go` only because its comment is long. Follow the short-comment convention.
- Design note Boundaries: `docs/architecture/m2-ability-model.md` (the `pkg/sim` and `pkg/eventlog` rows).
- Rulings that apply: none.

### The text (lead decision; copy exactly)
1. At the top of `pkg/sim/sim.go`, directly above `package sim` with no blank line between them:
   ```go
   // Package sim drives matches on the rules kernel: the seat policies
   // (RandomPolicy, ProgressPolicy), Playout for one game, Simulate and
   // SimulateWith for many, the simulation decks (FactionDeck, SimDecks,
   // ResolveDeck) and the rule monitors (Monitors, Violation, Coverage) that
   // check named FRD 01 rules in every simulated game.
   //
   // It is a client of the engine's exported API: it imports engine only from
   // this module, and engine never imports it. cardtown lint fences checks its
   // imports. CheckInvariants and RandomChooser stay in engine.
   ```
2. At the top of `pkg/eventlog/eventlog.go`, directly above `package eventlog` with no blank line between them:
   ```go
   // Package eventlog writes, reads and replays a game's event stream:
   // FormatEvent for readable lines, WriteEventsJSONL and ReadEventsJSONL for
   // JSON Lines, and Replay, which rebuilds a recorded game up to any event and
   // refuses a record from another build revision. The event and record types
   // live in engine (ADR-0005).
   //
   // It is a client of the engine's exported API: it imports engine only from
   // this module, and engine never imports it. cardtown lint fences checks its
   // imports.
   ```
3. **Not in this task:** any other file, any change to the other packages' comments, and any code.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests. These are checks recorded in the Log, and the reviewer re-runs them.
1. `go doc ./pkg/sim` and `go doc ./pkg/eventlog` each print the comment above, starting with `Package sim drives` and `Package eventlog writes`.
2. `git diff rework/executable-rules -- pkg/` adds only comment lines (every added line starts with `//`) to exactly those two files, and removes nothing.
3. `go doc` for `./pkg/engine/ability`, `./pkg/engine/cards` and `./pkg/view` prints the same as before.
4. `go tool task check` passes, and so does `cardtown lint fences`. The comments name no card; "Scratch Company" does not appear.

## Allowed Files
- `pkg/sim/sim.go` (decision 1 only)
- `pkg/eventlog/eventlog.go` (decision 2 only)

## Done When
- [x] Acceptance cases 1 to 4 hold, with their commands and results in the Log
- [x] `go tool task check` passes

## Log
- 2026-10-22 lead-architect: Wrote the task from board directive #33, narrowed to `sim` and `eventlog` after the lead found that `ability`, `cards` and `view` already have package comments (see Inputs). It is a harness task, tier flash, as the directive sets. T-0039 is done, so the task is `ready` on branch `m2/T-0040-package-comments`. T-0037 now depends on it too.
  - Next: **the user runs `@kernel-engineer T-0040` in the Antigravity agent (Gemini Flash)**.
- 2026-10-22 kernel-engineer: Added package doc comments to `pkg/sim/sim.go` and `pkg/eventlog/eventlog.go` (lead decisions 1 and 2). Status set to `in-progress` in b331588.
  - Diff: only comment additions to `pkg/sim/sim.go` (lines 1-9) and `pkg/eventlog/eventlog.go` (lines 1-9); no lines removed, no other files in `pkg/` modified.
  - Reuse: follows short-comment convention directly above `package` line in main package files (`ability.go`, `cards.go`, `view.go`). No code, shapes, or card names introduced.
  - Case 1: `go doc ./pkg/sim` and `go doc ./pkg/eventlog` each print their respective package paragraph ("Package sim drives..." and "Package eventlog writes...").
  - Case 2: `git diff rework/executable-rules -- pkg/` adds only comment lines (all starting with `//`) to `pkg/sim/sim.go` and `pkg/eventlog/eventlog.go`, and removes nothing.
  - Case 3: `go doc ./pkg/engine/ability`, `go doc ./pkg/engine/cards`, and `go doc ./pkg/view` produce identical output before and after.
  - Case 4: `cardtown lint fences` passes; `go tool task check` passes (gofmt, `go vet`, all tests, `cardtown lint --all`). Neither package comment names any card; "Scratch Company" does not appear.
  - Status set to `review`.
- 2026-10-22 lead-architect: Review: **approved**, merged as 8183fa6, status `done`.
  - Scope: b331588 changes the status line, and 63b367c changes only the two Allowed Files and this file. Every added line in `pkg/` starts with `//`, and nothing is removed. Both comments are the task's text word for word, directly above `package`.
  - Gates, re-run in a scratch worktree: `go doc` prints both paragraphs, and `go tool task check` exits 0.
  - Directive #33 is complete.
