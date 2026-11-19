---
id: T-0276
title: "One read of a continuous ability's filter, `matchesFromSource`, and `matchesFilter` goes (directives #161 and #168; wave H.1 harness, flash, pure move)"
kind: harness
tier: flash
wave: H.1 (board directives #161, #168)
status: in-progress
branch: m2/autoloop-wave-h.1/T-0276-matches-from-source
depends_on: [T-0275]
claims: [func:Game.matchesFromSource, func:Game.matchesFilter]
cards: []
requirements: []
---

## Goal
"A printed continuous ability's filter is judged against its source's current controller and current zone" (`SVY-R44`) is written once, in a new `(g *Game) matchesFromSource` in `filter.go`, and both continuous kinds use it: auras (`Game.auras`, [layers.go:81](../../../pkg/engine/layers.go)) and replacements (`replacementsFor`, [replace.go:46](../../../pkg/engine/replace.go)). The forwarding wrapper `matchesFilter` ([filter.go:139-143](../../../pkg/engine/filter.go)) goes, and its two callers call `matchesFilterAt` with a `nil` lock (ADR 0007). It is a pure move: no test changes, and match output is byte for byte the same. When this merges, #161 and #168 are `applied`.

## Inputs
- Board directive #161 (structural review 13, finding 3) and #168 (ADR 0007's forwarding wrappers), both approved by the user 2026-11-18 and `open: before the next Deliverables task`. #168 says "with or right after #161", which edits the same file. Design note Deliverable 6n ([m2-ability-model.md:220-221](../../../docs/architecture/m2-ability-model.md)) already names `matchesFromSource` and says #168 may share #161's task.
- Where it is today (at `autoloop/m2-wave-h.1` a7dad05):
  - `Game.auras` and `replacementsFor` each call `g.matchesFilterAt(src.Controller, src.ID, zoneLock(src), <filter>, a)`.
  - `matchesFilter(ctl, src, f, a)` only calls `g.matchesFilterAt(ctl, src, nil, f, a)`. Its callers are `costDecision`'s kill and exhaust filter ([costs.go:70](../../../pkg/engine/costs.go)) and `fireHookFor`'s `Watch` check ([trigger.go:162](../../../pkg/engine/trigger.go)). No test calls it (`git grep -nw matchesFilter -- "*_test.go"` finds nothing).
- Design (lead decisions):
  - `func (g *Game) matchesFromSource(src *Asset, f ability.Filter, a *Asset) bool` takes `matchesFilter`'s place in `filter.go`, right before `matchesFilterAt`. `matchesFilterAt`'s doc comment says who passes a nil lock and names `matchesFromSource`.
  - `costs.go` and `trigger.go` change only to update the reference to the deleted `matchesFilter`, so under the D2 size rule they don't count toward the four files. `docs/architecture/engine-structs.md`'s `filter.go` row lists the new function.
  - `ADR 0007`'s line naming `matchesFilter` ([0007:14](../../../docs/architecture/decisions/0007-moves-update-every-reference.md)) is a record of the wrapper and stays.
- Reuse search (D2.3): `git grep -n "zoneLock(src)" -- "*.go"` at a7dad05 finds the two sites above and `sameZoneLock` ([placement.go:149](../../../pkg/engine/placement.go)), which builds a lock for a stack item, not a filter check. `zoneLock` is reused, not copied.
- Size (D2.4): three non-test files (`filter.go`, `layers.go`, `replace.go`), plus two reference-only files and one doc.
- **Expected output difference:** none: `go tool task sim-compare` passes.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at a7dad05: `gofmt -l pkg cmd` clean, `go vet ./pkg/engine/` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-h.1` `[PASS]`. Blobs: `costs.go` e553f52, `filter.go` e2903dd, `layers.go` ff88686, `replace.go` c488410, `trigger.go` e52802b, `engine-structs.md` a2267dd.

## Commits
1. `refactor(engine): one read of a continuous ability's filter, matchesFromSource; matchesFilter goes (T-0276)` with Edits 1-6 (no double quotes in the message: the allow-list refuses an escaped one). A pure move adds no test, so there is no `test(T-0276)` commit.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
### 1. `pkg/engine/filter.go`: `matchesFromSource` replaces `matchesFilter`
```diff
@@ -137,8 +137,10 @@ func inLocked(a *Asset, lock *LockedZone) bool {
 }
 
-// matchesFilter is matchesFilterAt without a locked zone, for callers that have
-// no stack item (costs).
-func (g *Game) matchesFilter(ctl PlayerID, src int, f ability.Filter, a *Asset) bool {
-	return g.matchesFilterAt(ctl, src, nil, f, a)
+// matchesFromSource reports whether battlefield asset a meets f, the filter of
+// a printed continuous ability of asset src: judged against src's current
+// controller, and a ZoneRef criterion against src's current zone (SVY-R44).
+// Auras and replacements both read their filters through it.
+func (g *Game) matchesFromSource(src *Asset, f ability.Filter, a *Asset) bool {
+	return g.matchesFilterAt(src.Controller, src.ID, zoneLock(src), f, a)
 }
 
@@ -146,5 +148,6 @@ func (g *Game) matchesFilter(ctl PlayerID, src int, f ability.Filter, a *Asset)
 // (allegiance relative to ctl, kinds, subtypes, Scratch, Other: not the source asset
 // src, Stat: its effective value, Zone: currently in that battlefield zone, ZoneRef: in lock's zone, never the source's current one). Targeting adds Ward on top (legalTarget); an untargeted "all" step uses
-// this alone.
+// this alone. Costs and watching triggers have no stack item and pass a nil
+// lock; a continuous ability reads its filter through matchesFromSource.
 func (g *Game) matchesFilterAt(ctl PlayerID, src int, lock *LockedZone, f ability.Filter, a *Asset) bool {
 	if f.Other && a.ID == src {
```

### 2. `pkg/engine/layers.go`: auras read their filter through it
```diff
@@ -68,7 +68,7 @@ func (g *Game) Power(a *Asset) int {
 // is the one reader of which auras apply: a face-up battlefield source's printed
 // StaticAura applies when its Filter, relative to the
 // source's current controller, matches a. A ZoneRefSame criterion reads the
-// source's current zone through zoneLock (SVY-R44). A face-down source has no
+// source's current zone (matchesFromSource, SVY-R44). A face-down source has no
 // Ability, so it gives none. Each caller reads the field it needs (Power, Keywords, Cornered).
 func (g *Game) auras(a *Asset) []ability.Static {
 	var out []ability.Static
@@ -78,7 +78,7 @@ func (g *Game) auras(a *Asset) []ability.Static {
 			continue
 		}
 		for _, st := range d.Statics {
-			if st.Kind == ability.StaticAura && g.matchesFilterAt(src.Controller, src.ID, zoneLock(src), st.Filter, a) {
+			if st.Kind == ability.StaticAura && g.matchesFromSource(src, st.Filter, a) {
 				out = append(out, st)
 			}
 		}
```

### 3. `pkg/engine/replace.go`: replacements read their filter through it
```diff
@@ -33,7 +33,7 @@ type usedReplacements map[usedKey]bool
 // by a face-up asset on the battlefield whose filter a matches, except those
 // already used on a in this window (FR-REP-003, SVY-R24). A source that is
 // itself lethal still applies until the sweep destroys it (SVY-R20). A ZoneRef
-// criterion reads the source's current zone (SVY-R44).
+// criterion reads the source's current zone (matchesFromSource, SVY-R44).
 func (g *Game) replacementsFor(a *Asset, ev ability.Would, used usedReplacements) []replacementOption {
 	var out []replacementOption
 	for _, src := range g.assetsByID() {
@@ -43,7 +43,7 @@ func (g *Game) replacementsFor(a *Asset, ev ability.Would, used usedReplacements
 		reps := src.Printed().Ability.Replacements
 		for i := range reps {
 			r := &reps[i]
-			if r.Event == ev && g.matchesFilterAt(src.Controller, src.ID, zoneLock(src), r.Filter, a) && !used[usedKey{a.ID, src.ID, r}] {
+			if r.Event == ev && g.matchesFromSource(src, r.Filter, a) && !used[usedKey{a.ID, src.ID, r}] {
 				out = append(out, replacementOption{src: src, rep: r})
 			}
 		}
```

### 4. `pkg/engine/costs.go`: the cost filter calls `matchesFilterAt` (reference update only)
```diff
@@ -67,7 +67,7 @@ func (g *Game) costDecision(pl placement, c ability.Cost) (Decision, []*Asset) {
 		if c.Kind == ability.CostExhaust && a.Exhausted {
 			return false
 		}
-		return g.matchesFilter(pl.controller, pl.source, c.Filter, a)
+		return g.matchesFilterAt(pl.controller, pl.source, nil, c.Filter, a)
 	})
 	if counted {
 		d.Min, d.Max = pl.count, pl.count
```

### 5. `pkg/engine/trigger.go`: the `Watch` check calls `matchesFilterAt` (reference update only)
```diff
@@ -159,7 +159,7 @@ func (g *Game) fireHookFor(src triggerSource, hook ability.Hook, watched *Asset)
 		if t.Hook != hook || t.Watches != (watched != nil) || !activeIn(src.zone, t.ActiveZone) || (t.Turn == ability.TurnYours && g.Active != src.controller) {
 			continue
 		}
-		if watched != nil && !g.matchesFilter(src.controller, src.id, t.Watch, watched) {
+		if watched != nil && !g.matchesFilterAt(src.controller, src.id, nil, t.Watch, watched) {
 			continue
 		}
 		if (hook == ability.HookDeploy || hook == ability.HookMoves) && !zoneMatches(src.zone, t.ToZone) {
```

### 6. `docs/architecture/engine-structs.md`: the `filter.go` row lists `matchesFromSource`
```diff
@@ -291,7 +291,7 @@
-| | [`filter.go`](../../pkg/engine/filter.go) | *(none)* | `legalTarget`, `matchesFilterAt`, `matchesCard`, `cardMeets` |
+| | [`filter.go`](../../pkg/engine/filter.go) | *(none)* | `legalTarget`, `matchesFromSource`, `matchesFilterAt`, `matchesCard`, `cardMeets` |
```

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **One read of a continuous ability's filter:** `git -C W grep -n "zoneLock(src)" -- "*.go"` prints exactly two lines: `pkg/engine/filter.go`, inside `matchesFromSource`, and `pkg/engine/placement.go`, inside `sameZoneLock`.
2. **The wrapper is gone:** `git -C W grep -nw matchesFilter -- "*.go"` prints nothing.
3. **Both continuous kinds use it:** `git -C W grep -n "matchesFromSource(" -- pkg` prints the declaration in `filter.go` and two calls, one in `layers.go` and one in `replace.go`, and nothing else.
4. **The same tests pass, unedited:** `go -C W test ./pkg/...` passes, and `git -C W diff autoloop/m2-wave-h.1 --stat -- "*_test.go"` prints nothing.
5. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-h.1` prints `[PASS]`.
6. **Only the listed edits:** `git -C W diff autoloop/m2-wave-h.1 --numstat -- pkg docs` prints exactly six lines: `1	1	docs/architecture/engine-structs.md`, `1	1	pkg/engine/costs.go`, `8	5	pkg/engine/filter.go`, `2	2	pkg/engine/layers.go`, `2	2	pkg/engine/replace.go` and `1	1	pkg/engine/trigger.go`.
7. **Nothing else changed:** `git -C W diff autoloop/m2-wave-h.1 --stat` lists only the six files of case 6 and this task file.
8. `gofmt -l W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/filter.go`
- `pkg/engine/layers.go`
- `pkg/engine/replace.go`
- `pkg/engine/costs.go` (reference update only, Edit 4)
- `pkg/engine/trigger.go` (reference update only, Edit 5)
- `docs/architecture/engine-structs.md` (Edit 6)

## Done When
- [ ] Cases 1-8 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-18, lead-architect (wave H.1 run, session 5): written in full against a7dad05 for #161 and #168. `main` (7a8e4f5) has nothing the run branch lacks, so no sync. Not `WAVE-COMPLETE` before the task: #161, #166, #167, #168, #162, #163 and #164 are `open: before the next Deliverables task`. #161 and #168 share this task, as #168 allows, since both edit `filter.go` and together they stay within the D2 size rule (three non-test files that aren't reference-only). This is a harness task, so it skips `tests` and goes straight to `ready`. Tier is `flash`, because the edits are exact and were dry-run. A pure move adds no test and edits none. Every command here is on the headless allow-list: one per line, no pipes, no `$`, and patterns with parentheses are quoted. Waiting directives checked: #6, #7 (mode `serial`), #15 (2) (no draw or end-of-turn hook), #71, #91 (2), #106, #120, #142: none has fired. Next: `kernel-engineer` on Flash.
