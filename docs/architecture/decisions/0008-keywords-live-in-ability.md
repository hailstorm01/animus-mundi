# 0008. The keyword set lives in package `ability`

- **Status:** Accepted
- **Date:** 2026-11-18

## Context

A card's standalone keywords (`FR-KWD-002`: Resistant N, Stalwart, Aggressive N, Ethereal, Ward, Versatile, Guardian, Armor N, Deploys Exhausted) are one set, `engine.Keywords`, with one table of names, `keywordDefs`, and one fold. The printed set is Layer 1. A Cornered grant, an aura's grant and a step's grant are Layer 3 grants of the same keywords (`FR-LAY-001`, `FR-KWD-002e`).

The grants are written in card definitions, which are `ability` values (ADR 0004), and `ability` can't import `engine`. So each grant that arrived spelled its own fields instead of naming the set: `WhileCornered` has `Aggressive` and `Resistant` with its own `Add` (T-0180), `Static.Resistant` grants one keyword (T-0258), and `Step.Keyword` is an `ability.Keyword`, an enum that names only Aggressive and Stalwart, so `OpGrantKeyword` can grant only Stalwart (T-0219). `Game.Keywords` adds the Cornered and aura fields by hand beside the one fold. Structural review 13 recorded this as a duplication (board directive #159).

## Decision

`Keywords`, its name table `keywordDefs`, `ParseKeywordLine`, `Lines` and the fold (exported as `Add`) move from `engine` to `ability`, in `pkg/engine/ability/keywords.go`. Every reference names `ability.Keywords`, with no alias left in `engine` (ADR 0007). Each grant a card definition writes is an `ability.Keywords` value: `Step.Keywords` for `OpGrantKeyword` (T-0272), then `WhileCornered` and `Static` (T-0273). `Game.Keywords` folds the printed set and every grant through the one `Add`.

`ability.Keyword` stays, but only as the identity of a keyword that is a triggered ability (`Trigger.Keyword`, `StackItem.Keyword`: Aggressive's attack trigger, `FR-KWD-002f`). It no longer names a keyword a step grants.

The package's rule "plain data, no game logic" still holds in spirit. The name table, the line parser and the fold say what a keyword *is*: its printed form and how two sets combine. Which keywords an asset has at a moment (layers, Cornered, auras) stays in `engine`.

Card records and snapshots (`AssetSnap.Keywords`, ADR 0005) keep their field names and JSON tags, so match output doesn't change.

## Consequences

- A card definition can grant any keyword, through the same type that holds a printed one. A new keyword needs one `keywordDefs` line and one field.
- `ability` gains a small amount of code beside its data (a regexp, a parser, a fold). Package `engine` keeps reading keywords through `Game.Keywords`.
- Tests in `engine`, `sim` and `view` that build keyword literals change from `Keywords{…}` or `engine.Keywords{…}` to `ability.Keywords{…}`: a large, mechanical rename (ADR 0007).
- Revisit this if keywords gain behavior that needs game state (for example a keyword whose value depends on the board): that part belongs in `engine`, not here.
