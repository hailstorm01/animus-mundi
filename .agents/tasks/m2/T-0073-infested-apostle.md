---
id: T-0073
title: Make Infested Apostle executable (wave B card, flash)
kind: card
tier: flash
wave: B
status: done
branch: m2/T-0073-infested-apostle
depends_on: [T-0072, T-0075, T-0076]
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
Written in `pkg/engine/cards_animus_apostle_test.go` (session 10, approved; carried over in session 12). The four directive #47 tests and the decline test:
- `TestCard_InfestedApostle_KillHappensOnResolutionAfterTargetLocked`: "you may" and the opposing target are asked on placement, the kill pick on resolution; the killed ally goes to discard and the target takes 2 damage.
- `TestCard_InfestedApostle_MayKillItself`: picking the Apostle kills it, and its damage still lands.
- `TestCard_InfestedApostle_TimeoutAtKillPickDealsNoDamage`: a timeout at the pick kills nothing and deals nothing.
- `TestCard_InfestedApostle_DamageResolvesBeforeKilledUnitsDeathTriggers`: a killed ally with a death trigger (for example a real Animus card with one) has it placed and resolved after the 2 damage.
- Also: declining "you may" asks nothing more and does nothing.
- **Change (session 12, after T-0076):** the kill pick is now a `DecidePick`, not a `DecideTarget`. In `TestCard_InfestedApostle_KillHappensOnResolutionAfterTargetLocked` (line 43) the third expected kind becomes `DecidePick`, and in `TestCard_InfestedApostle_TimeoutAtKillPickDealsNoDamage` (line 102) the second does. Nothing else changes.

## Helpers Needed
None expected.

## Allowed Files
- `pkg/engine/cards/animus.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [x] The card tests pass: `go test ./pkg/engine -run 'TestCard_InfestedApostle'`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-10-26 lead-architect (wave run B, session 6): Wrote the task, `queued` behind T-0072 (its shapes). Split from T-0072 by the size rule (D2.4).
- 2026-10-27 lead-architect (wave run B, session 11): Back to `queued` behind T-0076 (D6, missing shape). Session 10's Flash run (`agy` conversation `096cd558-…`) got the definition working, but `go tool task sim` failed in the targeting monitor (game 3, seed 3, event 377): the kill pick asked a `DecideTarget`, so the monitor wanted the placement of an item after it. T-0076 gives picks their own kind, `DecidePick`.
  - State of `m2/T-0073-infested-apostle`: `e127feb` (approved tests), `7baee6f` (`ready`), `ae58054` (`in-progress`). The run's uncommitted definition was in its `scratch/t0073` worktree, which is gone (`git worktree list` shows none), so nothing of it survives. The definition above is unchanged.
  - When T-0076 is `done`: recreate the branch from the run branch, carry over the test commit, and set `status: tests`. `rules-test-author` changes the kill pick's expected kind from `DecideTarget` to `DecidePick` in `TestCard_InfestedApostle_KillHappensOnResolutionAfterTargetLocked` (line 43, the third kind) and `TestCard_InfestedApostle_TimeoutAtKillPickDealsNoDamage` (line 102, the second kind). The pick is answered through `scriptPlan`'s target queue, which T-0076 extends to picks, so nothing else changes.
- 2026-10-27 lead-architect (wave run B, session 10, carried over): T-0072 and T-0075 were `done`, and no directive was open or triggered (#58 and #59 do not fire: the kill pick is on the battlefield, not a pile, and nothing is exhausted). Status `tests`; `rules-test-author` wrote `pkg/engine/cards_animus_apostle_test.go` (five tests, all failing as expected); the lead approved them (e127feb).
- 2026-10-27 lead-architect (wave run B, session 12): T-0076 is `done` (029b7bf). Synced `main` into the run branch (9e1cedb), moved the branch to it, and carried over the test commit as 1ac6470 (test file only). No directive is open or triggered. Status `tests`: `rules-test-author` makes the change named under Tests.
- 2026-10-27 rules-test-author (wave run B, session 12): Rework after T-0076. Changed the kill pick's expected kind to `DecidePick` in `TestCard_InfestedApostle_KillHappensOnResolutionAfterTargetLocked` (line 43, third kind) and `TestCard_InfestedApostle_TimeoutAtKillPickDealsNoDamage` (line 102, second kind). Nothing else changed: `scriptPlan` answers `DecideTarget` and `DecidePick` from one shared queue (the same pattern as `if_you_do_test.go`, `Optional, Target, Pick`), and `_MayKillItself` answers by kind (anything but `DecideOptional` is a pick) so it needs no edit; `_DamageResolvesBefore...` and `_DecliningAsksNothing...` assert no kinds. Helpers requested: none. Five tests, all failing only with "needs rules the kernel does not implement yet"; `go vet` and `gofmt` are clean.
- 2026-10-27 lead-architect (wave run B, session 12): Test rework approved (a8f1ce5): exactly the two kinds named under Tests changed; the timeout test's opposing target has one option and is auto-answered, so `[DecideOptional, DecidePick]` is right. Status `ready`; dispatching Flash.
- 2026-10-27 kernel-engineer (flash): Implemented Infested Apostle ability definition in pkg/engine/cards/animus.go and removed "Infested Apostle" from NotYetExecutable in pkg/engine/cards/cards.go. Reused ability.HookDeploy, ability.Trigger.Optional, ability.AllegianceOpposing, ability.KindUnit, ability.OpKill with SubjectPick, UpTo: 1, AllegianceAllied, and ability.OpDealDamage with SubjectTarget, Target: 0, Amount: Value{N: 2}, IfDone: true. Commands run: 'go test ./pkg/engine -run TestCard_InfestedApostle' (PASS, 5/5 tests); 'go test ./pkg/engine -run TestStarterSet_Executable' (PASS); 'go tool task check' (PASS: fmt-check, vet, test, build, lint --all); 'go tool task sim' (PASS: 200 games clean with all invariants holding). Reviewer: confirm the deploy trigger's target slot is locked on placement while the kill pick resolves untargeted on resolution via DecidePick with IfDone gating the 2 damage.
- 2026-10-27 lead-architect (wave run B, session 12): Approved and merged (de82da3). Scope: only the two Allowed Files plus the test commits; no test changed after a8f1ce5. The definition is the one under Definition. Re-ran in a review worktree: `go tool task check` (exit 0) and `go tool task sim` (clean, 200 games, P0 70.5% / P1 28.5% / draws 1.0%, was 57.5% / 39.0% / 3.5%; no monitor violation, so T-0076's `DecidePick` fixed session 10's failure). Both pass again on the run branch after the merge. 25 deck cards still not executable.
