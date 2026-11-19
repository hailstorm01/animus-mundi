---
id: T-0287
title: "One walker over an `Effect` and its `Modes`, and `checkStepReads` asks `InPile` (directive #173 (c), (d); wave H.2 harness, flash)"
kind: harness
tier: flash
wave: H.2 (board directive #173 (c), (d))
status: done
branch: m2/autoloop-wave-h.2/T-0287-effect-walker
depends_on: [T-0286]
claims: [func:ability.Effect.Parts]
cards: []
requirements: []
---

## Goal
Split off from #173 by T-0286 (the four parts together touch six non-test files). This task covers #173 (c) and (d), as pure moves:
- **(c)** Walking an effect's steps into its `Modes`, recursively, is written in `Effect.ReadsZoneRef` ([ability.go:372-388](../../../pkg/engine/ability/ability.go)) and `checkEffectReads` ([cardload.go:99-109](../../../pkg/engine/cardload.go)). After this task, `Effect.Parts()` in `ability` holds the one recursion (it returns `e` and each of its `Modes`' parts, depth first, in printed order), and both read it. `for … range e.Modes` then appears only in `Parts` and in `legalModes` ([placement.go:285](../../../pkg/engine/placement.go)), which picks modes, not steps.
- **(d)** `checkStepReads` tests `s.Filter.Zone == ability.ZoneDiscard` ([cardload.go:113](../../../pkg/engine/cardload.go)), though `Filter.InPile`'s doc says every check that tells piles from the battlefield asks it. After this task it calls `s.Filter.InPile()`, and `cardload.go` has no `ZoneDiscard`. (The directive writes `!s.Filter.InPile()` for the requirement "a split step's filter is not a pile"; inside that condition's list of failures the call is not negated, so behavior is unchanged.)

When this merges after T-0286, #173 is `applied`.

## Inputs
- Board directive #173 (structural review 14, finding 5), approved by the user 2026-11-19, `open: before the next Deliverables task`. (a) and (b) were T-0286 (merge 607c182).
- Where it is today (at `autoloop/m2-wave-h.2` 354ea5a): as the Goal says.
- Design (lead decisions):
  - `Parts` returns a slice, not a callback, so a caller can `return` from inside its loop (both callers do). It sits right above `ReadsZoneRef`, its first reader.
  - The order is the one both walks visit today (pre-order: `e`, then mode 1 and its modes, then mode 2, …). `checkEffectReads` therefore returns the same first error. `ReadsZoneRef` returns a bool, so its order cannot show.
  - `checkEffectReads` keeps its name and signature, so its three callers in `checkShapeReads` don't change.
  - Nothing else walks `Modes` recursively: `git grep -n "\.Modes" -- pkg cmd ":!*_test.go"` finds only the two walks, `legalModes`, `chooseModes`' index (`placement.go:320`) and `len` checks.
- Reuse search (D2.3): no walker exists (`git grep -n "Parts()" -- pkg cmd` finds nothing).
- Size (D2.4): two non-test files.
- Docs: `pkg/engine/doc.go`, the design note, `engine-structs.md` and the `ability` package doc name `Effect.Parts` with #176, after the wave's last code task.
- **Expected output difference:** none. `go tool task sim-compare` passes.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 354ea5a:
  - `gofmt -l pkg cmd` clean, `go vet ./pkg/engine/...` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-h.2` `[PASS]`.
  - Blobs: `ability/ability.go` 0920813, `cardload.go` ca84f00.

## Commits
1. `refactor(engine): one walker over an Effect and its Modes, Effect.Parts, and checkStepReads asks InPile (T-0287)` with Edits 1-2 (no double quotes in the message: the allow-list refuses an escaped one). A pure move adds no test, so there is no `test(T-0287)` commit.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
### 1. `pkg/engine/ability/ability.go` (`Parts`, and `ReadsZoneRef` reads it)
```diff
@@ -367,22 +367,31 @@ const (
 	ZoneRefSame                // "in the same zone": the zone the item's source was in when the item was placed (SVY-R44)
 )
 
-// ReadsZoneRef reports whether any filter of e reads r: its Targets' filters,
-// its Steps' filters, and each of its Modes, recursively.
-func (e Effect) ReadsZoneRef(r ZoneRef) bool {
-	for _, t := range e.Targets {
-		if t.Filter.ZoneRef == r {
-			return true
-		}
+// Parts returns e and each of its Modes' parts, recursively, depth first in
+// printed order. It is the one walk into an effect's modes: every check that
+// reads each part's Targets or Steps (ReadsZoneRef, the card loader's checks)
+// ranges over it.
+func (e Effect) Parts() []Effect {
+	parts := []Effect{e}
+	for _, m := range e.Modes {
+		parts = append(parts, m.Parts()...)
 	}
-	for _, s := range e.Steps {
-		if s.Filter.ZoneRef == r {
-			return true
+	return parts
+}
+
+// ReadsZoneRef reports whether any filter of e reads r: its Targets' filters
+// and its Steps' filters, in e and each of its Modes (Parts).
+func (e Effect) ReadsZoneRef(r ZoneRef) bool {
+	for _, p := range e.Parts() {
+		for _, t := range p.Targets {
+			if t.Filter.ZoneRef == r {
+				return true
+			}
 		}
-	}
-	for _, m := range e.Modes {
-		if m.ReadsZoneRef(r) {
-			return true
+		for _, s := range p.Steps {
+			if s.Filter.ZoneRef == r {
+				return true
+			}
 		}
 	}
 	return false
```

### 2. `pkg/engine/cardload.go` (`checkEffectReads` reads `Parts`; `checkStepReads` asks `InPile`)
```diff
@@ -95,13 +95,10 @@ func checkShapeReads(d *ability.Def) error {
 	return nil
 }
 
-// checkEffectReads checks e's steps and, recursively, each of its Modes.
+// checkEffectReads checks the steps of e and of each of its Modes (Parts).
 func checkEffectReads(e ability.Effect) error {
-	if err := checkStepReads(e.Steps); err != nil {
-		return err
-	}
-	for _, m := range e.Modes {
-		if err := checkEffectReads(m); err != nil {
+	for _, p := range e.Parts() {
+		if err := checkStepReads(p.Steps); err != nil {
 			return err
 		}
 	}
@@ -110,7 +107,7 @@ func checkEffectReads(e ability.Effect) error {
 
 func checkStepReads(steps []ability.Step) error {
 	for _, s := range steps {
-		if s.Split && (s.Op != ability.OpAddArmor || s.Subject != ability.SubjectPick || s.UpTo != 0 || s.Filter.Zone == ability.ZoneDiscard) {
+		if s.Split && (s.Op != ability.OpAddArmor || s.Subject != ability.SubjectPick || s.UpTo != 0 || s.Filter.InPile()) {
 			return fmt.Errorf("Step.Split is read only for OpAddArmor over SubjectPick assets with UpTo 0")
 		}
 		if s.ControllerChooses && s.Op != ability.OpDiscard {
```

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **One walk into modes:** `git -C W grep -n "range e.Modes" -- pkg cmd ":!*_test.go"` prints exactly two lines: `pkg/engine/ability/ability.go:376` (in `Parts`) and `pkg/engine/placement.go:285` (in `legalModes`).
2. **Both read it:** `git -C W grep -n "e.Parts()" -- pkg cmd` prints exactly two lines: `pkg/engine/ability/ability.go:385` (in `ReadsZoneRef`) and `pkg/engine/cardload.go:100` (in `checkEffectReads`).
3. **No zone test in the loader:** `git -C W grep -n ZoneDiscard -- pkg/engine/cardload.go` prints nothing, and `git -C W grep -n "InPile()" -- pkg/engine/cardload.go` prints exactly one line, `pkg/engine/cardload.go:110`.
4. **The same tests pass, unedited:** `go -C W test ./pkg/... ./cmd/...` passes, and `git -C W diff autoloop/m2-wave-h.2 --numstat -- "*_test.go"` prints nothing.
5. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-h.2` prints `[PASS]`.
6. **Only the listed edits:** `git -C W diff autoloop/m2-wave-h.2 --numstat -- pkg cmd docs` prints exactly two lines: `23	14	pkg/engine/ability/ability.go` and `4	7	pkg/engine/cardload.go`.
7. **Nothing else changed:** `git -C W diff autoloop/m2-wave-h.2 --stat` lists only the two files of case 6 and this task file.
8. `gofmt -l W/pkg W/cmd` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/ability/ability.go` (Edit 1)
- `pkg/engine/cardload.go` (Edit 2)

## Tests
- None added or edited: a pure move. Every existing test passes unedited.

## Done When
- [x] Cases 1-8 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-h.2` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-19, lead-architect (wave H.2 run, session 6): split off from #173 by T-0286 under the D2 size rule, as #173 allows. Queued behind T-0286; #174 waits for both, because #173 blocks it ("Wave H.2, after #173").
- 2026-11-19, lead-architect (wave H.2 run, session 7): written against 354ea5a for #173 (c) and (d); T-0286 is `done`.
  - `main` (1859fe4) has nothing the run branch lacks, so no sync. Not `WAVE-COMPLETE` before the task: #173 and #174 are `open: before the next Deliverables task`.
  - The claim is now the walker's name, `func:ability.Effect.Parts`.
  - This is a harness task, so it skips `tests` and goes straight to `ready`. Tier is `flash`, because the edits are exact and were dry-run.
  - Every command here is on the headless allow-list: one per line, no pipes, no `$`, and patterns with special characters are quoted.
  - Waiting directives checked: #6, #7 (mode `serial`), #15 (2) (no draw or end-of-turn hook), #71, #91 (2), #106, #120, #142 (wave H.2 has no cards). None has fired.
  - Next: `kernel-engineer` on Flash.
- 2026-11-19, kernel-engineer (wave H.2 run, session 8): implemented T-0287.
  - Diff shows:
    - `pkg/engine/ability/ability.go`: added `Effect.Parts() []Effect` traversing `e` and each of its `Modes` recursively depth first in printed order; updated `Effect.ReadsZoneRef` to range over `e.Parts()` and inspect `Targets` and `Steps` filters.
    - `pkg/engine/cardload.go`: updated `checkEffectReads` to loop over `e.Parts()` and call `checkStepReads(p.Steps)`; updated `checkStepReads` to call `s.Filter.InPile()` instead of direct zone comparison (`s.Filter.Zone == ability.ZoneDiscard`).
  - Reuse: `git grep -n "Parts()" -- pkg cmd` confirmed no previous walker existed.
  - Acceptance cases 1-8:
    1. `git -C W grep -n "range e.Modes" -- pkg cmd ":!*_test.go"`: printed exactly 2 lines (`pkg/engine/ability/ability.go:376`, `pkg/engine/placement.go:285`).
    2. `git -C W grep -n "e.Parts()" -- pkg cmd`: printed exactly 2 lines (`pkg/engine/ability/ability.go:385`, `pkg/engine/cardload.go:100`).
    3. `git -C W grep -n ZoneDiscard -- pkg/engine/cardload.go`: printed nothing; `git -C W grep -n "InPile()" -- pkg/engine/cardload.go`: printed exactly 1 line (`pkg/engine/cardload.go:110`).
    4. `go -C W test ./pkg/... ./cmd/...`: ok; `git -C W diff autoloop/m2-wave-h.2 --numstat -- "*_test.go"`: printed nothing.
    5. `go -C W tool task sim-compare -- autoloop/m2-wave-h.2`: printed `[PASS] starter and stress sims match autoloop/m2-wave-h.2 byte for byte, exit codes included`.
    6. `git -C W diff autoloop/m2-wave-h.2 --numstat -- pkg cmd docs`: printed exactly `23 14 pkg/engine/ability/ability.go` and `4 7 pkg/engine/cardload.go`.
    7. `git -C W diff autoloop/m2-wave-h.2 --stat`: only `pkg/engine/ability/ability.go`, `pkg/engine/cardload.go`, and `.agents/tasks/m2/T-0287-effect-walker.md` modified.
    8. `gofmt -l W/pkg W/cmd`: clean; `go -C W tool task check`: all passed; `go -C W tool task sim`: 200/200 games clean.
  - Reviewer notes: blobs match dry run (`ability.go` 0920813, `cardload.go` ca84f00). Ready for review.
- 2026-11-19, lead-architect (wave H.2 run, session 7, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-h.2` as 3cd2dd8. Flash run `fb4c8dbb-e989-48c5-9ca5-5f0fbc1b8a4c`: exit 0, `status: SUCCESS`, no `denied_actions` key, and no step with `"status":"ERROR"` in its transcript. `.err` holds only agy's "root agent idle" notice. (The engineer's Log entry says "session 8"; it is this session's dispatch.) Review of `git diff autoloop/m2-wave-h.2...m2/autoloop-wave-h.2/T-0287-effect-walker` (55bc086):
  - Scope: the two Allowed Files and this file. The implementation commit (ca42153) holds only the two non-test files, and both blobs match the lead's dry run (`ability/ability.go` 0920813, `cardload.go` ca84f00). No test file changed.
  - Fences and reuse: `Effect.Parts` is the one walk into an effect's `Modes`; `ReadsZoneRef` and `checkEffectReads` read it, and `range e.Modes` is left only in `Parts` and `legalModes`. `checkStepReads` asks `Filter.InPile`, and `cardload.go` names no `ZoneDiscard`. Nothing old survives, and nothing forwards (ADR 0007).
  - Re-run by the reviewer in a review worktree of the task branch: cases 1-4 and 6 as listed; `gofmt -l pkg cmd` clean; `go tool task check` exit 0; `go tool task sim-compare -- autoloop/m2-wave-h.2` `[PASS]`; `go tool task sim` exit 0, P0 58.5% / P1 39.5% / draws 2.0%, P0 50/50 and P1 50/50 main cards executable, no monitor violation.
  - #173 (c) and (d) are applied, so #173 is `applied` and archived. #174 is next.
