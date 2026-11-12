---
id: T-0219
title: "Allshan's Battle Trance: a keyword granted until end of turn (wave F shape, sonnet; #137, #140 (2))"
kind: shape
tier: sonnet
wave: F
status: queued
branch: m2/autoloop-wave-f/T-0219-allshans-battle-trance
depends_on: [T-0218]
claims: [primitive:GRANT_KEYWORD, modifier:keyword]
cards: [Allshan's Battle Trance]
requirements: [FR-LAY-001, FR-LAY-003, FR-TRN-003a]
---

## Goal
*Allshan's Battle Trance* ("Deploy chosen unit card from your discard pile to the Main Theater. Until end of turn, it gains Stalwart and gets +2 power.", `data/cards/maneuvers-path.csv`) is executable: a keyword grant until end of turn is a Layer 3 modifier on T-0218's list, which `Keywords` folds, and it and the +2 power expire together in Sub-Phase 4c (`SVY-R39`, #140 (2)).

## Covers (written by the lead at queue time; Inputs, Tests and Allowed Files are written against the code of the day before this task leaves `queued`)
- `OpGrantKeyword` (Layer 3, `FR-LAY-001`), as a keyword modifier on `Asset.Modifiers` with `UntilEndOfTurn`, folded by `Keywords`. Before writing it, read T-0218's `Modifier` as merged, so the keyword grant extends that list and adds no second one (#137).
- "it" in the second sentence: the asset the chosen card became when `OpDeployFree` (`Step.To` = Main Theater) put it onto the battlefield. Check whether a step can name that asset today (a target slot holding a card in a discard pile, then the asset it became); if not, that subject is part of this task, or a split under the D2 size rule.
- #140 (2)'s test: after Sub-Phase 4c the unit stays in the Main Theater without Stalwart and at printed power.
- The card definition, its `Text`, and its removal from `cards.NotYetExecutable`.
- Directive #137 is `open: before` this task (T-0218); #140 is `open: before` it for part (2).

## Log
- 2026-11-11, lead-architect (wave F run, session 20): queued as the next wave F deck card in survey row F's order. Writing it fired #137's trigger, so T-0218 (#137's pure move) goes first. Tier `sonnet`: it adds a shape.
