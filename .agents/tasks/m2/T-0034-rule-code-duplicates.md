---
id: T-0034
title: Remove duplicated rule code (Berserk targeting, combat zones, Scratch Company face, monitor lookups)
kind: harness
tier: sonnet
wave: none (board directive #26, cleanup 3 of 4, before cardtown play)
status: review
branch: m2/T-0034-rule-code-duplicates
depends_on: [T-0033]
claims: [func:ScratchFace, func:inCombatZone]
cards: []
requirements: []
---

## Goal
Each rule is written once. The user's audit (board directive #26) found these duplicates:
1. **Ward targeting.** The Ward rule for chosen targets is written twice: in `legalTarget` and in Berserk's `legalSecondary`.
2. **Combat zones.** `combatLegalZone` and `canDeclareAttack` are two near-identical "type plus zone" combat checks.
3. **The Scratch Company face.** `sim/monitor.go` keeps its own copy of it, and that copy has no subtypes.
4. **Monitor lookups.** Inside `monitor.go`, `find` and `battlefield` both walk a snapshot's battlefield.

## Inputs
- Board directive #26, duplicates 3 to 6.
- After T-0032:
  - `legalSecondary`, `canDeclareAttack` and `combatLegalZone` are in `pkg/engine/combat.go`;
  - `legalTarget` and `matchesFilter` are in `targeting.go`;
  - `scratchFace` is in `card.go`.
- The monitors: `pkg/engine/sim/monitor.go`, which uses `scratchFace`, `find` and `battlefield`.
- The rules:
  - `FR-KWD-002o`: the Berserk secondary is chosen under normal targeting rules, as an opposing Main Theater asset.
  - `FR-KWD-002h`: Ward.
  - `FR-CBT-001` and `FR-CBT-002`: the combat zones.
  - `FR-ZON-005`: the Scratch Company face.
- Rulings that apply: none.

### Shapes
```go
// pkg/engine (card.go)
// ScratchFace returns a copy of the face-down Scratch Company characteristics (FR-ZON-005).
func ScratchFace() Card

// pkg/engine (combat.go)
// inCombatZone reports whether a is where it can fight: a Headquarters in the
// Headquarters Zone, anything else in the Main Theater.
func inCombatZone(a *Asset) bool
```

### Lead decisions (technical)
- **`legalSecondary(att, sec)`** becomes `sec.Zone == ZoneMainTheater && g.legalTarget(att.Controller, att.ID, ability.Filter{Allegiance: ability.AllegianceOpposing}, sec)`. This works because `legalTarget` already holds the opposing-Ward rule. Its result must stay the same for every input.
- **The combat-zone checks:**
  - `combatLegalZone` is replaced by `inCombatZone` (same body).
  - `canDeclareAttack(a)` becomes "a Unit or a Headquarters, and `inCombatZone`".
  - Every caller keeps its result.
- **The Scratch Company face:**
  - `scratchFace` stays the engine's internal pointer. `ScratchFace()` returns a copy, so no caller can change the engine's face.
  - The monitors use `engine.ScratchFace()`, and their own copy is removed.
- **The monitor lookups:** `find` is written with `battlefield`, or the other way round, so the walk is written once.
- **No behavior change.** All existing tests pass unchanged, and `cardtown sim` output, coverage table included, stays byte-identical.
- **Not in this task:** any change to what a monitor checks, and any other rule refactor.

## Acceptance Cases (harness tasks only; written by the lead)
Case 1 goes in a new `pkg/engine/sim/scratchface_test.go` (or an engine test file), committed first as `test(T-0034): …` and failing before the implementation. Cases 2 to 4 are checks recorded in the Log.
1. **`ScratchFace()`:**
   - It returns a Unit named "Scratch Company", 1/1, yields 1/1, with the subtypes Infantry and Scratch Company and no keywords.
   - Changing the returned value doesn't change a face-down asset's `Printed()` in a new game.
2. **Rule code unchanged in effect:**
   - Every existing test passes unchanged, the Berserk, Ward, Guardian, Ethereal and combat-fizzle tests included.
   - The keyword-deck playout test (which runs the monitors over Guardian, Ethereal, Ward and Berserk decks) passes.
3. **Outputs unchanged:** `sim` (starter and stress, both policies, coverage tables included) is byte-identical to a baseline CLI built from `rework/executable-rules` before any edit.
4. `go tool task check` passes.

## Allowed Files
- `pkg/engine/combat.go`
- `pkg/engine/card.go`
- `pkg/engine/sim/monitor.go`
- the new test file (the `test(T-0034)` commit)

## Done When
- [x] Case 1 has a test, committed before the implementation, and passes
- [x] Cases 2 and 3 hold, with output in the Log
- [x] `go tool task check` passes

## Log
- 2026-10-21 lead-architect: Wrote the task from board directive #26, with status `queued` behind T-0033 (harness, sonnet). The audit's R2 was split into T-0033 and T-0034 to keep each within four non-test files.
- 2026-10-21 lead-architect: T-0033 is merged (6e02b1c). Set to `ready`. The paths in Inputs are as of T-0032: `combat.go`, `targeting.go` and `card.go` are confirmed.
- 2026-10-21 kernel-engineer: Implemented. Test commit 9135a24 (`pkg/engine/sim/scratchface_test.go`) failed first with `undefined: engine.ScratchFace`.
  - Reuse: `legalSecondary` now calls the existing `legalTarget` (Ward rule written once); `canDeclareAttack` calls `inCombatZone` (the renamed `combatLegalZone`, same body); monitors call `engine.ScratchFace()` and their own copy is gone.
  - `ScratchFace()` copies the struct and clones `Subtypes` and `DeployZones`, so no caller can change the engine's face.
  - Monitor lookups: `find` and `battlefield` both use one new helper `eachAsset` (same order: Theater, then per player Reserves and HQ). `find` still returns pointers into the snapshot.
  - Monitor `card()` now returns a pointer to a fresh copy per face-down call (the face now has subtypes; the monitors never read them).
  - Case 2: `go test ./...` passes with no existing test edited. `go tool task check` passes.
  - Case 3: baseline CLI built from `rework/executable-rules` (759c5ae) in a separate worktree, new CLI built from this branch. Both run with `sim` on `-deck starter -p0 ANIMUS -p1 PATH`, `-games 100` (stress decks) and `-games 100 -p0 ANIMUS -p1 PATH`, each with `-policy random` and `-policy progress`, coverage tables included. Combined output compared with `cmp`: identical (219 lines, all exit codes 0). Worktree removed.
