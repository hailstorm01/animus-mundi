---
id: T-0073
title: Make Infested Apostle executable (wave B card, flash)
kind: card
tier: flash
wave: B
status: queued
branch: m2/T-0073-infested-apostle
depends_on: [T-0072, T-0075]
claims: [card:Infested Apostle]
cards: [Infested Apostle]
requirements: []
---

## Goal
*Infested Apostle* (Animus unit, four copies in the Animus starter deck) runs in the kernel. It is plain `ability` data in `pkg/engine/cards/animus.go`. Its one rules-text line is:
- "When this unit deploys, you may kill an allied unit. If you do, deal 2 damage to chosen opposing unit."

## Inputs
- Card (`data/cards/units-and-structures-animus.csv`): Animus unit, Fungus, Mage, Common, cost 4, deploys to Reserves, 1/2, with the line above.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 109. Its FRD 02 benchmark (#9) puts the kill in `optionalCosts`. `SVY-R14` supersedes that: the kill happens on resolution.
- Shapes it uses: `HookDeploy` and `Trigger.Optional` (A1, A2), a target slot (A2), `OpDealDamage`, and `OpKill`, `SubjectPick` with `UpTo`, and `Step.IfDone` (T-0072).
- Rulings: `SVY-R9` ("you may" is answered on placement), `SVY-R3` (Y's target is chosen then), `SVY-R14` (the kill is an untargeted pick on resolution, may pick the Apostle itself, and a timeout or no match means no damage), `SVY-R16`.
- Board directive #47 (1) names the settling tests below.

### Definition (the engineer adds it to the Animus defs)
- One trigger: `Hook: ability.HookDeploy`, `Optional: true`.
- Its `Effect`:
  - `Targets`: one slot, `Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}` (exactly one).
  - Step 1: `{Op: ability.OpKill, Subject: ability.SubjectPick, UpTo: 1, Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}}`. `Other` stays false, so the Apostle may kill itself.
  - Step 2: `{Op: ability.OpDealDamage, Subject: ability.SubjectTarget, Target: 0, Amount: ability.Value{N: 2}, IfDone: true}`.
- `Text` holds the printed line.
- Remove "Infested Apostle" from `cards.NotYetExecutable`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written by `rules-test-author` in `pkg/engine/cards_animus_apostle_test.go`, from the CSV text, when this task reaches `tests`. At least the four directive #47 tests:
- `TestCard_InfestedApostle_KillHappensOnResolutionAfterTargetLocked`: "you may" and the opposing target are asked on placement, the kill pick on resolution; the killed ally goes to discard and the target takes 2 damage.
- `TestCard_InfestedApostle_MayKillItself`: picking the Apostle kills it, and its damage still lands.
- `TestCard_InfestedApostle_TimeoutAtKillPickDealsNoDamage`: a timeout at the pick kills nothing and deals nothing.
- `TestCard_InfestedApostle_DamageResolvesBeforeKilledUnitsDeathTriggers`: a killed ally with a death trigger (for example a real Animus card with one) has it placed and resolved after the 2 damage.
- Also: declining "you may" asks nothing more and does nothing.

## Helpers Needed
None expected.

## Allowed Files
- `pkg/engine/cards/animus.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run 'TestCard_InfestedApostle'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-10-26 lead-architect (wave run B, session 6): Wrote the task, `queued` behind T-0072 (its shapes). Split from T-0072 by the size rule (D2.4).
