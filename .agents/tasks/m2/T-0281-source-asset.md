---
id: T-0281
title: "One lookup of the source asset or its last-known record, `sourceAsset` (directive #169; wave H.2 harness, flash, pure move)"
kind: harness
tier: flash
wave: H.2 (board directive #169)
status: review
branch: m2/autoloop-wave-h.2/T-0281-source-asset
depends_on: [T-0280]
claims: [func:Game.sourceAsset, func:Game.itemSource]
cards: []
requirements: []
---

## Goal
"The source asset, or its last-known record once it has left the battlefield" (`FR-ZON-004`) is written once, in a new `(g *Game) sourceAsset(id int) *Asset` in `state.go`, beside `Game.Asset`. Today it is written twice: `itemSource` ([effects.go:436-441](../../../pkg/engine/effects.go)) and inline in `sameZoneLock` ([placement.go:142-145](../../../pkg/engine/placement.go)). `itemSource` goes, its two callers pass `it.Source` to `sourceAsset`, and `sameZoneLock` calls it too. It is a pure move: no test changes, and match output is byte for byte the same. When this merges, #169 is `applied`.

## Inputs
- Board directive #169 (structural review 14, finding 1), approved by the user 2026-11-19, `open: before the next Deliverables task`; wave H.2, first. Design note Deliverable 6o ([m2-ability-model.md:228](../../../docs/architecture/m2-ability-model.md)) already names `sourceAsset`.
- Where it is today (at `autoloop/m2-wave-h.2` 1859fe4):
  - `itemSource(it *StackItem)` ([effects.go:432-441](../../../pkg/engine/effects.go)) returns `g.Asset(it.Source)`, else `g.LastKnown(it.Source)`. Its callers are `resolveAttack` ([combat.go:89](../../../pkg/engine/combat.go)) and `stepDealDamage` ([effects.go:206](../../../pkg/engine/effects.go)).
  - `sameZoneLock` ([placement.go:142-148](../../../pkg/engine/placement.go)) does the same lookup inline on its `source` ID.
  - `git grep -n "LastKnown(" -- "pkg/engine/*.go" ":!*_test.go"` finds only those two lookups and the declaration in `state.go`.
- Design (lead decisions):
  - `sourceAsset` takes an asset ID, not a stack item, because `sameZoneLock` has only the ID. It sits in `state.go` right after `Game.Asset`, since `doc.go` gives `state.go` "core state types and lookups". `itemSource`'s doc comment moves onto it, widened to name the same-zone lock (`SVY-R44`).
  - `combat.go` changes only to update the reference, so under the D2 size rule it doesn't count toward the four files.
- Docs: neither `pkg/engine/doc.go` nor `docs/architecture/engine-structs.md` names `itemSource`, so neither changes here. #176 (b) adds `sourceAsset` to them. `docs/design/outcomes/2026-11-09-wave-f-same-zone-source.md:55` names `itemSource` in a ruling's text; that file is `game-designer`'s record, and #176 decides it.
- Reuse search (D2.3): `git grep -n itemSource -- "*.go"` and the `LastKnown(` search above find nothing else; nothing new is added beside an existing equivalent.
- Size (D2.4): three non-test files (`state.go`, `effects.go`, `placement.go`), plus one reference-only file (`combat.go`).
- **Expected output difference:** none: `go tool task sim-compare` passes.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 1859fe4: `gofmt -l pkg cmd` clean, `go vet ./pkg/engine/` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-h.2` `[PASS]`. Blobs: `combat.go` 31164e4, `effects.go` 7f6b828, `placement.go` 0698ee2, `state.go` e84a3ca.

## Commits
1. `refactor(engine): one source-asset lookup, sourceAsset (T-0281)` with Edits 1-4 (no double quotes in the message: the allow-list refuses an escaped one). A pure move adds no test, so there is no `test(T-0281)` commit.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
### 1. `pkg/engine/state.go`: `sourceAsset` after `Game.Asset`
```diff
@@ -297,6 +297,19 @@ func (g *Game) Asset(id int) *Asset {
 	return nil
 }
 
+// sourceAsset returns the battlefield asset with this ID, or its last-known
+// record once it has left the battlefield (FR-ZON-004), so a damage event still
+// names and attributes its source (FR-REP-004 Stage 5), an attack still knows
+// its attacker's controller (FR-CBT-002 Step 3), and a same-zone lock still
+// finds its source's zone (SVY-R44). It is nil only for an ID that was never an
+// asset.
+func (g *Game) sourceAsset(id int) *Asset {
+	if a := g.Asset(id); a != nil {
+		return a
+	}
+	return g.LastKnown(id)
+}
+
 // TheaterOf returns the Main Theater assets controlled by p.
 func (g *Game) TheaterOf(p PlayerID) []*Asset {
 	var out []*Asset
```

### 2. `pkg/engine/effects.go`: `itemSource` goes, and `stepDealDamage` calls `sourceAsset`
```diff
@@ -203,7 +203,7 @@ func stepPlayers(it *StackItem, st ability.Step) []PlayerID {
 // With a chosen dealer (Step.Dealer) that asset deals the damage, at its current
 // power; with none left the step deals nothing and logs no event (SVY-R27).
 func (g *Game) stepDealDamage(r *resolution, st ability.Step) int {
-	src, dealer := g.itemSource(r.it), (*Asset)(nil)
+	src, dealer := g.sourceAsset(r.it.Source), (*Asset)(nil)
 	if slot, ok := st.DealerSlot(); ok {
 		ds := g.slotAssets(r, slot)
 		if len(ds) == 0 {
@@ -429,17 +429,6 @@ func (g *Game) value(source int, v ability.Value, dealer *Asset) int {
 	return 0
 }
 
-// itemSource returns the item's source asset, or its last-known record once it has
-// left the battlefield (FR-ZON-004), so a damage event still names and attributes
-// it (FR-REP-004 Stage 5) and an attack still knows its attacker's controller
-// (FR-CBT-002 Step 3). It is nil only for a source that was never an asset.
-func (g *Game) itemSource(it *StackItem) *Asset {
-	if a := g.Asset(it.Source); a != nil {
-		return a
-	}
-	return g.LastKnown(it.Source)
-}
-
 // stepAssets returns the assets an asset step acts on: the entities chosen for
 // its slot that are still legal and still on the battlefield, in pick order (a
 // target an earlier step or mode removed is skipped: FR-ZON-004, FR-TGT-003
```

### 3. `pkg/engine/placement.go`: `sameZoneLock` calls `sourceAsset`
```diff
@@ -139,10 +139,7 @@ func (g *Game) sameZoneLock(e ability.Effect, source int, held *LockedZone) *Loc
 	if !e.ReadsZoneRef(ability.ZoneRefSame) {
 		return held
 	}
-	src := g.Asset(source)
-	if src == nil {
-		src = g.LastKnown(source)
-	}
+	src := g.sourceAsset(source)
 	if src == nil {
 		return nil
 	}
```

### 4. `pkg/engine/combat.go`: `resolveAttack` calls `sourceAsset` (reference only)
```diff
@@ -86,7 +86,7 @@ func (g *Game) resolveAttack(it *StackItem) {
 	// last-known controller once it has left), not the declaring player
 	// (FR-CBT-002 Step 3). A control change on either side fizzles the attack.
 	attController := it.Controller
-	if a := g.itemSource(it); a != nil {
+	if a := g.sourceAsset(it.Source); a != nil {
 		attController = a.Controller
 	}
 	if def == nil || !inCombatZone(def) || def.Controller == attController {
```

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **`itemSource` is gone:** `git -C W grep -n itemSource -- "*.go"` prints nothing.
2. **One lookup:** `git -C W grep -n "g.LastKnown(" -- "pkg/engine/*.go" ":!*_test.go"` prints exactly one line, `pkg/engine/state.go:310:	return g.LastKnown(id)`.
3. **Its three callers:** `git -C W grep -n "g.sourceAsset(" -- pkg/engine` prints exactly three lines, in `combat.go`, `effects.go` and `placement.go`.
4. **The same tests pass, unedited:** `go -C W test ./pkg/... ./cmd/...` passes, and `git -C W diff autoloop/m2-wave-h.2 --numstat -- "*_test.go"` prints nothing.
5. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-h.2` prints `[PASS]`.
6. **Only the listed edits:** `git -C W diff autoloop/m2-wave-h.2 --numstat -- pkg cmd docs` prints exactly four lines: `1	1	pkg/engine/combat.go`, `1	12	pkg/engine/effects.go`, `1	4	pkg/engine/placement.go` and `13	0	pkg/engine/state.go`.
7. **Nothing else changed:** `git -C W diff autoloop/m2-wave-h.2 --stat` lists only the four files of case 6 and this task file.
8. `gofmt -l W/pkg W/cmd` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/state.go` (Edit 1)
- `pkg/engine/effects.go` (Edit 2)
- `pkg/engine/placement.go` (Edit 3)
- `pkg/engine/combat.go` (Edit 4, reference only)

## Tests
- None added or edited: a pure move. Every existing test passes unedited.

## Done When
- [x] Cases 1-8 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-h.2` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-19, lead-architect (wave H.2 run, session 1): written in full against 1859fe4 for #169, wave H.2's first directive. Run file created. `main` (1859fe4) has nothing the run branch lacks, so no sync. Not `WAVE-COMPLETE` before the task: #169, #170, #172, #173 and #174 are `open: before the next Deliverables task`. This is a harness task, so it skips `tests` and goes straight to `ready`. Tier is `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list: one per line, no pipes, no `$`, and patterns with special characters are quoted. Waiting directives checked: #6, #7 (mode `serial`), #15 (2) (no draw or end-of-turn hook), #71, #91 (2), #106, #120, #142 (wave H.2 has no cards): none has fired. Next: `kernel-engineer` on Flash.
- 2026-11-19, kernel-engineer (wave H.2 run, session 1): implemented #169 as pure move.
  - Implemented Edits 1-4:
    - Added `(g *Game) sourceAsset(id int) *Asset` in `pkg/engine/state.go` after `Game.Asset` with doc comment citing FR-ZON-004, FR-REP-004 Stage 5, FR-CBT-002 Step 3, and SVY-R44.
    - Updated `stepDealDamage` in `pkg/engine/effects.go` to call `g.sourceAsset(r.it.Source)` and removed `itemSource`.
    - Updated `sameZoneLock` in `pkg/engine/placement.go` to call `g.sourceAsset(source)`.
    - Updated `resolveAttack` in `pkg/engine/combat.go` to call `g.sourceAsset(it.Source)`.
  - Reused `Game.Asset` and `Game.LastKnown` inside `sourceAsset`. Search verified no other calls to `itemSource` or `LastKnown` in non-test engine code.
  - Verified acceptance cases:
    - Case 1: `git -C W grep -n itemSource -- "*.go"` printed nothing (exit 1).
    - Case 2: `git -C W grep -n "g.LastKnown(" -- "pkg/engine/*.go" ":!*_test.go"` printed exactly `pkg/engine/state.go:310:	return g.LastKnown(id)`.
    - Case 3: `git -C W grep -n "g.sourceAsset(" -- pkg/engine` printed exactly 3 callers (combat.go:89, effects.go:206, placement.go:142).
    - Case 4: `go -C W test ./pkg/... ./cmd/...` passed with exit 0; `git -C W diff autoloop/m2-wave-h.2 --numstat -- "*_test.go"` printed nothing.
    - Case 5: `go -C W tool task sim-compare -- autoloop/m2-wave-h.2` printed `[PASS] starter and stress sims match autoloop/m2-wave-h.2 byte for byte, exit codes included`.
    - Case 6: `git -C W diff autoloop/m2-wave-h.2 --numstat -- pkg cmd docs` printed exactly the 4 expected lines (1 1 combat.go, 1 12 effects.go, 1 4 placement.go, 13 0 state.go).
    - Case 7: `git -C W diff autoloop/m2-wave-h.2 --stat` showed only the 4 engine files and the task file.
    - Case 8: `gofmt -l W/pkg W/cmd` printed nothing; `go -C W tool task check` passed (fmt-check, vet, tests, lint); `go -C W tool task sim` passed (200/200 games, all invariants holding).
  - Committed implementation as `refactor(engine): one source-asset lookup, sourceAsset (T-0281)` (`cc729b7d`).
  - Status moved to `review`.
