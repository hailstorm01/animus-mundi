---
id: T-0235
title: "The play screen shows a decision's source card and its rules text, and the stack listing names each item's card (directive #149 (c), wave F.2 harness)"
kind: harness
tier: flash
wave: F.2
status: queued
branch: m2/autoloop-wave-f.2/T-0235-play-screen-source
depends_on: [T-0234]
claims: [func:terminalChooser (source line), func:printScreen (stack listing)]
cards: []
requirements: []
---

## Goal
On `cardtown play`, a decision with a `Source` (T-0233, T-0234) prints its source card's name and printed rules text under the prompt, before the options, and the stack listing names the card of an Ability, a Deploy and a Maneuver (#149 (c)).

## Covers (queued stub, written 2026-11-14 with T-0233; write Inputs, exact Edits and Acceptance Cases against the code of the day before it leaves `queued`)
- `terminalChooser` (`cmd/cardtown/play_input.go`): after `d.Prompt`, when `d.Source` is not nil, one line with the card's name and its `RulesText` (for example `  [Devilhorn Captain] Unbowed: Deal 1 damage to chosen asset in the Main Theater.`), for every kind of decision, the "you may" included. A decision without a source prints as before. Never parse the prompt (#130).
- `printScreen` (`cmd/cardtown/play_screen.go`): the stack listing reads `Ability of <card>` or `<keyword> trigger of <card>` (as the engine's place event names it), `Deploy of <card> to <zone>`, and `Maneuver: <card>` as now; March, Attack and Scratch as now.
- Acceptance cases from #149: (a) *Devilhorn Captain*'s Unbowed target prompt shows "Devilhorn Captain" and its rules text before its options; (b) the same for a Maneuver's target, a modal trigger's mode choice and an optional trigger's "you may"; (c) a trigger from a card in a discard pile (*Eerie Remnant*) shows that card's text; (d) a decision outside any ability (the trigger order, the mulligan) prints as before; (e) the stack listing shows the card for an ability, a Deploy and a Maneuver; (f) the `pkg/view/viewtest` leak tests pass; `go tool task check`; `go tool task sim-compare` (no engine change).
- When it merges, #149 is `applied` (T-0233, T-0234, T-0235).

## Allowed Files (expected)
- `cmd/cardtown/play_input.go`
- `cmd/cardtown/play_screen.go`
- `cmd/cardtown/play_test.go` (new tests only)

## Log
- 2026-11-14, lead-architect (wave F.2 run, session 1): queued stub, split off T-0233 under D2.4. Follows T-0234.
