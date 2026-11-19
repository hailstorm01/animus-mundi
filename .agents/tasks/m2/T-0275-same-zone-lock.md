---
id: T-0275
title: "One \"in the same zone\" lock on placement, `sameZoneLock` (directive #160; wave H.1 harness, flash, pure move)"
kind: harness
tier: flash
wave: H.1 (board directive #160)
status: ready
branch: m2/autoloop-wave-h.1/T-0275-same-zone-lock
depends_on: [T-0274]
claims: [func:Game.sameZoneLock, func:Game.activatedPlacement, func:Game.placeTrigger]
cards: []
requirements: []
---

## Goal
"An item whose effect reads `ZoneRefSame` locks its source's zone on placement" (`SVY-R44`) is written once, in a new `(g *Game) sameZoneLock` in `placement.go`, and both placements use it: an activated ability's (`activatedPlacement`) and a trigger's (`placeTrigger`). Today it is written twice ([placement.go:129-131](../../../pkg/engine/placement.go), [trigger.go:276-287](../../../pkg/engine/trigger.go)). It is a pure move: no test changes, and match output is byte for byte the same. When this merges, #160 is `applied`.

## Inputs
- Board directive #160 (structural review 13, finding 2, approved by user 2026-11-18), `open: before the next Deliverables task`; wave H.1, after #165 (applied, T-0274). Design note Deliverable 6n ([m2-ability-model.md:219](../../../docs/architecture/m2-ability-model.md)) already names `sameZoneLock`.
- Where it is today (at `autoloop/m2-wave-h.1` 55cbc98):
  - `activatedPlacement` ([placement.go:125-133](../../../pkg/engine/placement.go)) is a plain function. When `act.Effect.ReadsZoneRef(ability.ZoneRefSame)`, it sets `pl.locked = zoneLock(a)`. Its three callers are `Apply` and `check` ([actions.go:84](../../../pkg/engine/actions.go), [actions.go:172](../../../pkg/engine/actions.go)), and `ActivatedCost` ([costs.go:174](../../../pkg/engine/costs.go)). Each passes a live battlefield asset (`g.Asset`, `ownAsset`, or the client's asset from `g.Assets()`), so `g.Asset(a.ID)` is `a`. Costs never read `pl.locked` (only `targeting.go:99` and `stackItem` do).
  - `placeTrigger` ([trigger.go:274-294](../../../pkg/engine/trigger.go)) handles the same check inline: the source's zone now, else its last-known zone, else no lock. For an effect that doesn't read it, it keeps the lock the pending trigger already has (`fireMoved`'s, [trigger.go:101-107](../../../pkg/engine/trigger.go), `SVY-R45`).
- Design (lead decisions):
  - `func (g *Game) sameZoneLock(e ability.Effect, source int, held *LockedZone) *LockedZone`, in `placement.go` right after `activatedPlacement`. It returns `held` when `e` doesn't read `ZoneRefSame`. Otherwise it returns `zoneLock` of the source asset now, or of its last-known copy, or nil. `placeTrigger`'s inline comment moves into its doc comment.
  - `activatedPlacement` becomes a method of `*Game` so it can call `sameZoneLock`. That is a signature change, so its three call sites change in this task too, and no wrapper keeps the old form (ADR 0007). `actions.go` and `costs.go` change only to update that reference, so under the D2 size rule they don't count toward the four files.
- Reuse search (D2.3): `git grep -n "ReadsZoneRef(ability.ZoneRefSame)" -- "*.go"` at 55cbc98 finds the two sites above and nothing else; `zoneLock` is reused, not copied.
- Size (D2.4): two non-test files, plus two reference-only files.
- **Expected output difference:** none: `go tool task sim-compare` passes.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 55cbc98: `gofmt -l pkg` clean, `go vet ./pkg/engine/` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-h.1` `[PASS]`. Blobs: `actions.go` a4d07c1, `costs.go` 098a962, `placement.go` f1d7b70, `trigger.go` 0115bf0.

## Commits
1. `feat(engine): one same-zone lock on placement, sameZoneLock (T-0275)` with Edits 1-4 (no double quotes in the message: the allow-list refuses an escaped one). A pure move adds no test, so there is no `test(T-0275)` commit.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
### 1. `pkg/engine/placement.go`: `activatedPlacement` is a method and locks through the new `sameZoneLock`
```diff
@@ -124,14 +124,31 @@ func deployPermission(c *Card) *ability.Static {
 
 // activatedPlacement is the placement of the printed ability act of asset a,
 // activated by pid, for both its legality check and declaring it.
-func activatedPlacement(pid PlayerID, a *Asset, act ability.Activated) placement {
+func (g *Game) activatedPlacement(pid PlayerID, a *Asset, act ability.Activated) placement {
 	pl := placement{controller: pid, source: a.ID, inst: a.Inst, label: a.Printed().Name, costs: act.Costs}
-	if act.Effect.ReadsZoneRef(ability.ZoneRefSame) {
-		pl.locked = zoneLock(a)
-	}
+	pl.locked = g.sameZoneLock(act.Effect, a.ID, pl.locked)
 	return pl
 }
 
+// sameZoneLock is the zone lock an item whose effect e reads "in the same
+// zone" takes on placement: the zone of its source now, or its last-known zone
+// if it has left, and none for a source in a pile (SVY-R44, FR-ZON-004). For
+// any other effect it is held, the lock the placement already has. Activated
+// abilities and triggers both lock through it.
+func (g *Game) sameZoneLock(e ability.Effect, source int, held *LockedZone) *LockedZone {
+	if !e.ReadsZoneRef(ability.ZoneRefSame) {
+		return held
+	}
+	src := g.Asset(source)
+	if src == nil {
+		src = g.LastKnown(source)
+	}
+	if src == nil {
+		return nil
+	}
+	return zoneLock(src)
+}
+
 // zoneLock is the zone a's source-relative filters lock: a's zone and
 // controller (SVY-R44, SVY-R45).
 func zoneLock(a *Asset) *LockedZone { return &LockedZone{Zone: a.Zone, Player: a.Controller} }
```

### 2. `pkg/engine/trigger.go`: `placeTrigger` locks through `sameZoneLock`
```diff
@@ -273,18 +273,7 @@ func (g *Game) flushTriggers() {
 // decision it asks names p's source card (Decision.Source, shownCard).
 func (g *Game) placeTrigger(p pendingTrigger) {
 	defer g.askFor(g.shownCard(p.source, p.inst))()
-	if p.trig.Effect.ReadsZoneRef(ability.ZoneRefSame) {
-		// "In the same zone": the source's zone now, or its last-known zone
-		// if it has left; none for a source in a pile (SVY-R44, FR-ZON-004).
-		src := g.Asset(p.source)
-		if src == nil {
-			src = g.LastKnown(p.source)
-		}
-		p.locked = nil
-		if src != nil {
-			p.locked = zoneLock(src)
-		}
-	}
+	p.locked = g.sameZoneLock(p.trig.Effect, p.source, p.locked)
 	if p.trig.Optional && (g.couldDoNothing(p.placement, p.trig.Effect) || !g.accepts(p)) {
 		return
 	}
```

### 3. `pkg/engine/actions.go`: both calls name the method
```diff
@@ -81,7 +81,7 @@ func (g *Game) Apply(a Action) error {
 	case ActActivate:
 		as := g.Asset(a.Asset)
 		act := as.Printed().Ability.Activated[a.Ability]
-		g.declare(activatedPlacement(a.Player, as, act), act.Effect, StackAbility)
+		g.declare(g.activatedPlacement(a.Player, as, act), act.Effect, StackAbility)
 	case ActMarch:
 		as := g.Asset(a.Asset)
 		g.exhaust(as)
@@ -169,7 +169,7 @@ func (g *Game) check(a Action) error {
 			return illegal("%s has no activated ability %d", as, a.Ability)
 		}
 		act := def.Activated[a.Ability]
-		if !g.placeable(activatedPlacement(a.Player, as, act), act.Effect) {
+		if !g.placeable(g.activatedPlacement(a.Player, as, act), act.Effect) {
 			return illegal("a cost of %s ability %d cannot be paid, or a required target or mode has no legal choice (FR-ACT-001, SVY-R36)", as, a.Ability)
 		}
 	case ActMarch:
```

### 4. `pkg/engine/costs.go`: the call names the method
```diff
@@ -171,7 +171,7 @@ func (g *Game) ActivatedCost(a *Asset, i int) (req, tac int, ok bool) {
 			tac += c.Tac
 		}
 	}
-	return req, tac, g.otherCostsPayable(activatedPlacement(a.Controller, a, act))
+	return req, tac, g.otherCostsPayable(g.activatedPlacement(a.Controller, a, act))
 }
 
 // DiscardDeployCost returns the Requisition and Tactics that deploying inst from
```

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **One site reads "in the same zone" for a lock:** `git -C W grep -n "ReadsZoneRef(ability.ZoneRefSame)" -- "*.go"` prints exactly one line, in `pkg/engine/placement.go`, inside `sameZoneLock`.
2. **No old form survives:** `git -C W grep -n "activatedPlacement(" -- pkg` prints the declaration `func (g *Game) activatedPlacement(` and three calls, each written `g.activatedPlacement(`, and nothing else.
3. **The same tests pass, unedited:** `go -C W test ./pkg/...` passes, and `git -C W diff autoloop/m2-wave-h.1 --stat -- "*_test.go"` prints nothing.
4. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-h.1` prints `[PASS]`.
5. **Only the listed edits:** `git -C W diff autoloop/m2-wave-h.1 --numstat -- pkg` prints exactly four lines: `2	2	pkg/engine/actions.go`, `1	1	pkg/engine/costs.go`, `21	4	pkg/engine/placement.go` and `1	12	pkg/engine/trigger.go`.
6. **Nothing else changed:** `git -C W diff autoloop/m2-wave-h.1 --stat` lists only the four files of case 5 and this task file.
7. `gofmt -l W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/placement.go`
- `pkg/engine/trigger.go`
- `pkg/engine/actions.go` (reference update only, Edit 3)
- `pkg/engine/costs.go` (reference update only, Edit 4)

## Done When
- [ ] Cases 1-7 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-18, lead-architect (wave H.1 run, session 4): written in full against 55cbc98 for #160. `main` (7a8e4f5) has nothing the run branch lacks, so no sync. Not `WAVE-COMPLETE` before the task: #160, #161, #166, #167, #168, #162, #163 and #164 are `open: before the next Deliverables task`. #160 goes alone, not with #161. Together they would touch five non-test files that aren't reference-only (`placement.go`, `trigger.go`, `filter.go`, `layers.go`, `replace.go`), over the D2 size rule. #161 and #168, which both edit `filter.go`, come next. This is a harness task, so it skips `tests` and goes straight to `ready`. Tier is `flash`, because the edits are exact and were dry-run. A pure move adds no test and edits none. Every command here is on the headless allow-list: one per line, no pipes, no `$`, and patterns with parentheses are quoted. Waiting directives checked: #6, #7 (mode `serial`), #15 (2) (no draw or end-of-turn hook), #71, #91 (2), #106, #120, #142: none has fired. Next: `kernel-engineer` on Flash.
