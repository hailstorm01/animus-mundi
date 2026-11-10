---
id: T-0206
title: The "in that zone" filter criterion, read against the locked zone (wave F shape, sonnet; #141 (1))
kind: shape
tier: sonnet
wave: F
status: queued
branch: m2/autoloop-wave-f/T-0206-that-zone-criterion
depends_on: [T-0205]
claims: [shape:"in that zone" filter criterion (ability.Filter relative-zone criterion; matchesFilter and legalTarget read the item's locked zone)]
cards: []
requirements: [FR-TGT-001, FR-TGT-003]
---

## Goal
Directive #141 (1), second half: a filter criterion "in that zone" (`SVY-R45`) that holds when the entity is in the zone its stack item locked (T-0205's `StackItem.Locked`, or `placement.locked` at the first target check). It is read at both target checks (`FR-TGT-003`) and by untargeted steps on resolution (`SubjectAll`, `pickAssets`), through the one filter check, so a later "in the same zone" (`SVY-R44`, #141 (3)) only adds another setter of the same locked zone, and no second criterion. Split from T-0205 under the D2 size rule.

Covers:
- An `ability.Filter` criterion relative to the source, as an enum whose zero value is "none" and whose first value is "in that zone, the triggering move's destination" (`SVY-R45`), with room for "in the same zone" (`SVY-R44`) later.
- `matchesFilter` ([filter.go:85](../../../pkg/engine/filter.go)) checks it against the locked zone: the entity's zone is the locked `Zone`, and for Reserves or the Headquarters Zone its controller is the locked `Player` (`FR-ZON-001`). A filter with the criterion and no locked zone matches nothing. `matchesFilter` never reads the source's current zone for it (#141 (1)).
- The callers that have a placement or an item pass its zone: `targetDecision` and `pickAssets` ([targeting.go:93](../../../pkg/engine/targeting.go), [:160](../../../pkg/engine/targeting.go)), the second target check and `SubjectAll` ([effects.go:41](../../../pkg/engine/effects.go), [:396](../../../pkg/engine/effects.go)). `costs.go` and `replace.go` keep their calls unchanged (keep a form of `matchesFilter` without a locked zone, so the task stays within four files).
- Synthetic-card tests only; no card prints it until T-0207.

Inputs, Tests and Allowed Files are written against the code of the day before this task leaves `queued` (D2.4). Planned Allowed Files: `pkg/engine/ability/ability.go`, `pkg/engine/filter.go`, `pkg/engine/targeting.go`, `pkg/engine/effects.go`. Planned tests (`TestFR_TGT_001_*`, `TestFR_TGT_003_*`): only assets in the locked zone are offered (Main Theater: both sides', the source itself included; Reserves: the controller's only); a target that left the zone before resolution is illegal; a target stays legal if the source moves again or leaves; a `SubjectAll` step with the criterion acts on the locked zone only; the criterion on an item with no locked zone matches nothing; `sim-compare` byte-identical.

## Log
- 2026-11-10 lead-architect (wave F run, session 5): split off T-0205 under the D2 size rule (#141 (1)); `queued` on it.
