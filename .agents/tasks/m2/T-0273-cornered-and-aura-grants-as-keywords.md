---
id: T-0273
title: Cornered and aura grants are ability.Keywords sets, and Game.Keywords folds once (directive #159, part 2; harness, no behavior change)
kind: harness
tier: sonnet
wave: H.1 (board directive #159, structural review 13, finding 1)
status: queued
branch: m2/autoloop-wave-h.1/T-0273-cornered-and-aura-grants-as-keywords
depends_on: [T-0272]
claims: [type:ability.WhileCornered, field:ability.Static.Keywords, func:Game.Keywords, func:Game.whileCornered]
cards: []
requirements: []
---

## Goal
The rest of directive #159 after T-0272: `ability.WhileCornered` becomes an `ability.Keywords` plus `Power` (its own `Aggressive` and `Resistant` fields go, and its `Add` folds the keywords through `Keywords.Add`), `Static.Resistant` becomes `Static.Keywords` (an aura's Layer 3 grant of any keyword), and `Game.Keywords` folds the printed set, the Cornered grant, each aura's grant and each Layer 3 modifier through the one `Keywords.Add`, adding no field by hand. Match output is unchanged.

## Covers (write Inputs, Acceptance Cases, Tests and Allowed Files against the code of the day before this leaves `queued`)
- `pkg/engine/ability/ability.go`: `WhileCornered{Keywords Keywords; Power int}` and its `Add`; `Static.Resistant` → `Static.Keywords`, with its comment.
- `pkg/engine/layers.go`: `Game.Keywords` and `Game.whileCornered`.
- Card definitions: `cards/path.go` (`WhileCornered{Aggressive: 3}`, `{Resistant: 1}` twice, `{Power: 2}` keeps its shape), `cards/animus.go` (the Resistant 1 aura), and any test that builds a `WhileCornered` or `Static{Resistant: …}`, named one by one.
- #159's remaining acceptance: `Game.Keywords` calls one fold and adds no field by hand; `WhileCornered` and `Static` have no per-keyword field; every existing test passes with only the named renames; `go tool task check`; `go tool task sim-compare` (byte for byte). When this merges, #159 is `applied`.

## Log
- 2026-11-18 lead-architect (wave H.1 session 1): queued as the second part of #159 (the directive's own split; D2.4), after T-0272.
