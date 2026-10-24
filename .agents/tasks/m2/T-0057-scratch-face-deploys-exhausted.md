---
id: T-0057
title: The Scratch Company face prints Deploys Exhausted, and a scratch is a unit deploying (KIM-S01, KIM-S05, directive #48, shape, sonnet)
kind: shape
tier: sonnet
wave: B
status: ready
branch: m2/T-0057-scratch-face-deploys-exhausted
depends_on: [T-0056]
claims: [face:Scratch Company, file:pkg/engine/zones.go#scratch]
cards: []
requirements: [FR-ZON-005, FR-KWD-002l]
---

## Goal
The face-down Scratch Company face prints **Deploys Exhausted** (FRD 01 v3.12, the `Scratch Company` CSV row). A scratched unit enters Exhausted because of that keyword, as any deployed card does (`resolveDeploy`), instead of a hard-coded flag. Each scratch also fires the same unit-deployed hook as `resolveDeploy` (`KIM-S05`). This is board directive #48 parts 1, 2 and 4. Part 5 (design note) is applied. Part 3 has nothing to build until Layer 3 exists.

## Inputs
- FRD 01 v3.12: revision summary v3.11 → v3.12, `FR-ZON-005`, `FR-KWD-002l`.
- Card: `data/cards/units-and-structures-common.csv` line 2, "Scratch Company", rules text "Deploys Exhausted".
- Rulings: `KIM-S01` (revised), `KIM-S05`, `KIM-S06`. Design Outcome `docs/design/outcomes/2026-10-23-scratch-company-deploys-exhausted.md`.
- Board directive #48 (the full text is the spec for this task).
- Existing code: `scratchFace` and `ScratchFace` in `pkg/engine/card.go`, `scratch` in `pkg/engine/zones.go`, `resolveDeploy` in `pkg/engine/stack.go`, and `fireDeploy` in `pkg/engine/trigger.go`.

### Lead decisions (technical)
1. `scratchFace.Keywords.DeploysExhausted = true`, set directly in `card.go`. Loading the face from the CSV row is not worth a loader dependency for one keyword.
2. `scratch` sets `Exhausted` from `a.Printed().Keywords.DeploysExhausted`, like `resolveDeploy`, and calls `g.fireDeploy(a)` after the observer event. The event kind stays `scratch`.
3. No game outcome changes in M2. The face has no triggers, and the deploy hook is self-only (T-0007).
4. `pkg/sim/monitor_entry.go` may take the expected entry state from `ScratchFace().Keywords` (optional, engineer's call).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
- New: `TestFR_ZON_005_ScratchFacePrintsDeploysExhausted`: `ScratchFace().Keywords` is exactly Deploys Exhausted, and a face-down asset's `Printed()` has it.
- Unchanged and still passing: `TestFR_ZON_005_ScratchCompanyEntersExhausted`.
- **Authorized edits** (in the `test(T-0057)` commit only): `TestScratchFace_Characteristics` (`pkg/sim/scratchface_test.go`) and the face-down assertion in `pkg/view/view_test.go` (about line 331). Each now expects exactly Deploys Exhausted instead of no keywords.
- Final Done When regex: `^TestFR_ZON_005_(ScratchFacePrintsDeploysExhausted|ScratchCompanyEntersExhausted)$`, plus the two edited tests `TestScratchFace_Characteristics` and `TestView_Case2_FaceDown`.
- Not tested in this task: that a scratch fires the deploy hook. Nothing observes it in M2. The first card that watches another unit deploying names the settling test.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
None expected.

## Allowed Files
- `pkg/engine/card.go`
- `pkg/engine/zones.go`
- `pkg/sim/monitor_entry.go`

## Done When
- [ ] The listed tests pass.
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean. Compare `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` before and after: it is byte-identical, or only face-down assets' keyword lists differ. Record it in the Log.

## Log
- 2026-10-23 lead-architect: Wrote the task from board directive #48, once its trigger had happened (FRD 01 v3.12 and the CSV row, b3a178b; the outcome's `Status: applied`). Status `queued` behind T-0056 (serial mode). Part 5 was applied to the design note with T-0056's commit.
- 2026-10-23 lead-architect: Status `tests`, for `rules-test-author`. T-0056 is `done` (7f42665). Waiting triggers checked: #6 and #7 (parallel mode), #8 (no `deck.go` change), #15 (no turn hook) and #20 (wave B) have not fired. Starter-deck cards not yet executable: 28, plus both Headquarters.
- 2026-10-23 rules-test-author: Tests written on m2/T-0057-scratch-face-deploys-exhausted. New `TestFR_ZON_005_ScratchFacePrintsDeploysExhausted` (pkg/engine/rules_test.go): face keywords are exactly Deploys Exhausted, and a face-down asset's Printed() has it. Authorized edits: `TestScratchFace_Characteristics` (pkg/sim) and the face-down assertion in `TestView_Case2_FaceDown` (pkg/view) now expect exactly Deploys Exhausted. All three fail now, as intended. `cardtown lint rules` passes; vet and gofmt clean. Helpers requested: none. Left out: the deploy-hook-on-scratch test, as the task directs.
- 2026-10-23 lead-architect: Test review (D4): **approved**, status `ready`.
  - 877d4ab adds `TestFR_ZON_005_ScratchFacePrintsDeploysExhausted` (`rules_test.go`) and makes exactly the two authorized edits: `TestScratchFace_Characteristics` and the face-down assertion in `TestView_Case2_FaceDown`.
  - The tests match `FR-ZON-005`, `FR-KWD-002l` and `KIM-S01` (revised).
  - **Reviewer check** (a throwaway prototype, not committed):
    - With the face keyword added, all three tests pass and `go test ./...` passes. No other test expected no keywords.
    - Adding lead decision 2 as well (`Exhausted` from the keyword, plus `fireDeploy`) changes nothing a test sees. The starter sim stays byte-identical to the integration branch.
    - So decision 2 is checked in code review, not by a test, as the task says.
  - Next: the lead dispatches `kernel-engineer` (Sonnet).
