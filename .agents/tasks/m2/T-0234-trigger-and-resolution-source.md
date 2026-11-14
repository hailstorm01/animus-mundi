---
id: T-0234
title: "Triggers and resolutions name their source card, and the decision record carries it (directive #149 (b), wave F.2 harness)"
kind: harness
tier: flash
wave: F.2
status: queued
branch: m2/autoloop-wave-f.2/T-0234-trigger-and-resolution-source
depends_on: [T-0233]
claims: [field:DecisionRecord.Source]
cards: []
requirements: [FR-MOD-001, FR-TRG-003, FR-ZON-005]
---

## Goal
Every decision asked while a trigger is placed or an Ability or a Maneuver resolves names its source card through T-0233's `askFor` scope, the trigger order and the mulligan still name none, and `DecisionRecord` carries the source's name so the `-events` output shows it (#149 (b)).

## Covers (queued stub, written 2026-11-14 with T-0233; write Inputs, exact Edits and Acceptance Cases against the code of the day before it leaves `queued`)
- `flushTriggers` (`trigger.go`): the whole flush runs in an `askFor(nil)` scope, so the trigger-order decision names no source even when a flush happens inside another scope; each trigger's "you may" (`accepts`) and its placement (`chooseOnPlacement`) run in a scope of `g.shownCard(p.source, p.inst)`.
- `resolveAbility` (`effects.go`): one `askFor(g.SourceCard(it))` scope for the whole resolution, reaching untargeted picks, the deck search, the look, the free-deploy zone choice and a discard step's hand picks.
- `DecisionRecord` (`events.go`) gains `Source string` (`json:"source,omitempty"`), the source card's name, filled in `emitDecision`. Name the expected `-events` difference in the task: decision events asked in a scope gain `"source"`. `sim-compare` compares only the text output and should still pass; check the `pkg/eventlog` replay tests and any golden event files.
- Acceptance cases to write: *Devilhorn Captain*'s Unbowed target decision carries the Captain with its rules text; an optional trigger's "you may" and a modal trigger's mode choice carry their card; *Eerie Remnant*'s trigger from a discard pile carries the Remnant; an untargeted pick, a deck search, a look and a free-deploy zone choice carry the resolving card; the trigger order and the mulligan carry none; a decision record has `source` when its decision has a source and omits it otherwise; the `pkg/view/viewtest` leak tests pass; `go tool task check`; `go tool task sim-compare`.
- A replacement choice (`replace.go:135`) asked while an ability resolves names the resolving card through the scope. Decide in the task whether that is right or whether `replace.go` sets its own source, and say which.

## Allowed Files (expected)
- `pkg/engine/trigger.go`
- `pkg/engine/effects.go`
- `pkg/engine/events.go`
- a new test file

## Log
- 2026-11-14, lead-architect (wave F.2 run, session 1): queued stub, split off T-0233 under D2.4 (#149 as a whole is about ten non-test files). Next after T-0233; T-0235 (the play screen) follows it.
