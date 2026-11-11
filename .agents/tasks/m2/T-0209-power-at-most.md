---
id: T-0209
title: A stat criterion in filters, "with power N or less" (wave F shape, sonnet)
kind: shape
tier: sonnet
wave: F
status: tests
branch: m2/autoloop-wave-f/T-0209-power-at-most
depends_on: []
claims: [shape:filter stat criterion (ability.Filter "stat N or less", power first; matchesCard reads a pile card's printed power, matchesFilterAt an asset's effective power)]
cards: []
requirements: [FR-TGT-001, FR-TGT-003, FR-ZON-004]
---

## Goal
The "Dynamic Stats" criterion of `FR-TGT-001` (line 375), first as "with power N or less", which *Reaver's Rush* (T-0210) prints: "Deploy chosen unit card with power 3 or less from your discard pile for free." One criterion read in both filter checks: a card in a discard pile by its printed power (it has no other, `FR-ZON-004`), and an asset on the battlefield by its effective power (`g.Power`, Layer 6). A later stat (*Devilhorn Scouting Party*'s "Requisition cost 3 or less") is another value of the same criterion, not a second field. Split from T-0208 under the D2 size rule.

Covers:
- An `ability.Filter` stat criterion: an enum of stats whose zero value is "no criterion" and whose first value is power, and the bound N ("N or less").
- `matchesCard` and `matchesFilterAt` in `filter.go` check it, each in its own way (printed for a pile card, effective for an asset).
- Synthetic-card tests only; no card prints it until T-0210.

- Not this task: *Reaver's Rush* (T-0210, `queued`); the Requisition-cost stat (the *Devilhorn Scouting Party* task).

## Inputs
- FRD 01 v3.23: `FR-TGT-001` (lines 370–383), above all "Readiness, Subtypes & Dynamic Stats" (line 375: a stat comparison is evaluated against current effective stats in Layer 6); `FR-TGT-003` Two-Stage Verification (line 395: a target that no longer satisfies a dynamic stat criterion at the second check is illegal); `FR-ZON-004` (line 242: a card off the battlefield has only its printed characteristics).
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Filter`, line 39: "stat comparisons").
- Code: `ability.Filter` ([ability.go:217-225](../../../pkg/engine/ability/ability.go)) and its enums beside it (`ZoneRef`, `ScratchCompany`); `matchesCard`, `matchesPrinted`, `matchesFilterAt` ([filter.go:22-112](../../../pkg/engine/filter.go)); `g.Power` ([layers.go:11](../../../pkg/engine/layers.go)). Every filter check already goes through `matchesCard` (pile targets, the pile second check `pileTarget`, pile picks, `countPileCards`) or `matchesFilterAt` (asset targets, both target checks, `SubjectAll`, `pickAssets`, and through `matchesFilter` costs and replacements), so no caller changes.

### Shape (the engineer builds it)
- **`ability.Stat`**, an `int` enum in `ability.go` beside `Filter`, with `StatNone` (the zero value: no stat criterion) and `StatPower` (Combat Power). Its comment says a later stat (Requisition cost, health) is another value, read the same way. **`Filter.Stat Stat`** and **`Filter.AtMost int`**: with `Stat` set, the entity's value of that stat must be `AtMost` or less ("with power 3 or less": `Stat: StatPower, AtMost: 3`). `AtMost` means nothing while `Stat` is `StatNone`. It is a conjunction with every other criterion.
- **`filter.go`:** the one comparison, `withinStat(f ability.Filter, v int) bool` (true when `f.Stat` is `StatNone`, else `v <= f.AtMost`), and one reader per form: `cardStat(s ability.Stat, c *Card) int`, a card's printed value (`FR-ZON-004`), and `(g *Game) assetStat(s ability.Stat, a *Asset) int`, an asset's effective value (`g.Power(a)` for `StatPower`, Layer 6, `FR-TGT-001`). Both return 0 for `StatNone`. `matchesCard` adds `withinStat(f, cardStat(f.Stat, inst.Card))`, and `matchesFilterAt` adds `withinStat(f, g.assetStat(f.Stat, a))`, each before its `matchesPrinted` call. `matchesPrinted` stays card-level and doesn't read the stat (an asset's printed power is not its power). Update the doc comments of `matchesCard` and `matchesFilterAt` to name the criterion.
- Behavior doesn't change for any printed card: none sets `Stat`. `go tool task sim-compare -- autoloop/m2-wave-f` must be byte-identical (a gate here).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/power_at_most_test.go`, with synthetic cards (`unit`, `structure`) and the deploy-trigger scenes of `if_you_do_test.go` (`iydTrigger`, `iydDeploy`: the Source deploys from player 0's hand and its trigger runs the steps under test). Put pile cards with `s.inDiscard`, battlefield assets with `s.put`, script with `s.script`, read offers with `s.asked(p)`. The filter under test is "power 3 or less": `Stat: ability.StatPower, AtMost: 3`, joined with the other criteria each test names (`ppAllyUnitCards` in `pile_pick_test.go` is the allied-unit-card pile filter). Raise or lower an asset's effective power with a Layer 5 modifier the way other tests do today (`TempPower`; the pure move of directive #137 updates every such test together). The Source itself is a unit on the battlefield, so give it a power that keeps it out of (or explicitly in) each asset offer, and say which in the test.
- `TestFR_TGT_001_PowerAtMostOffersPileCardsByPrintedPower` (`FR-TGT-001`, `FR-ZON-004`): player 0's discard pile holds unit cards with power 2, 3 and 4 and a structure card with power 0; player 1's holds a unit card with power 1. A pile target slot with `ppAllyUnitCards` plus "power 3 or less" offers exactly player 0's power 2 and power 3 unit cards (the bound is inclusive; kind and allegiance still apply).
- `TestFR_TGT_001_PowerAtMostReadsEffectivePowerOnTheBattlefield` (`FR-TGT-001` line 375, Layer 6): battlefield units of printed power 2, 3 and 4, one of printed power 3 with +1 power (effective 4), and one of printed power 4 with -2 power (effective 2). An asset target slot with "unit, power 3 or less" offers exactly the printed 2, the printed 3 and the effective-2 unit; not the effective-4 unit, nor the printed 4 one.
- `TestFR_TGT_003_PowerAtMostTargetCheckedAgainOnResolution` (`FR-TGT-003`): a chosen unit of power 3, targeted by an `OpAddArmor` 1 step, gets +1 power after placement and before resolution: it gains no armor. A pile card can't change, so no pile case.
- `TestFR_TGT_001_PowerAtMostAllStepReadsEffectivePower` (`FR-TGT-002`, Layer 6): a `SubjectAll` `OpAddArmor` 1 step with "unit, power 3 or less" gives armor to exactly the units whose effective power is 3 or less when it resolves, on both sides.
- `TestFR_TGT_001_PowerAtMostPilePickReadsPrintedPower` (`SVY-R3`, `FR-ZON-004`): a `SubjectPick` step on the pile filter with "power 3 or less" (`ppPick` in `pile_pick_test.go`) offers exactly player 0's unit cards of printed power 3 or less.
- `TestFR_TGT_001_NoStatCriterionIgnoresAtMost` (the zero value): a filter with `Stat` unset and `AtMost: 0` offers a power 5 unit card in the pile and a power 5 unit on the battlefield.
- **Existing tests:** every other test passes unchanged.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test author keeps any small builder in the test file.

## Allowed Files
- `pkg/engine/ability/ability.go` (`Stat`, `Filter.Stat`, `Filter.AtMost`)
- `pkg/engine/filter.go` (`withinStat`, `cardStat`, `assetStat`, the two checks)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "PowerAtMost|NoStatCriterion"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f` passes (no printed card sets `Stat`)

## Log
- 2026-11-10 lead-architect (wave F run, session 8): split off T-0208 under the D2 size rule; `queued`. It can go before or after T-0208, and T-0210 needs both.
- 2026-11-10 lead-architect (wave F run, session 9): Inputs, Shape, Tests and Allowed Files written against the run branch at 78c2fe1. Searched `pkg/engine` and the task archive for an existing stat criterion: none (`ability.Filter` has none, and no "or less" comparison exists). Set to `tests`.
