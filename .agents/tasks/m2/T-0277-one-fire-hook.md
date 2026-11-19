---
id: T-0277
title: "`fireHook` takes `watched` and `fireHookFor` goes; `chooseDiscards` goes (directives #166 and #167; wave H.1 harness, flash, pure move)"
kind: harness
tier: flash
wave: H.1 (board directives #166, #167)
status: ready
branch: m2/autoloop-wave-h.1/T-0277-one-fire-hook
depends_on: [T-0276]
claims: [func:Game.fireHook, func:Game.fireHookFor, func:Game.chooseDiscards, func:Game.chooseHandCards]
cards: []
requirements: []
---

## Goal
Two forwarding wrappers that predate ADR 0007 go, and each name has one home. `fireHook` ([trigger.go:144-146](../../../pkg/engine/trigger.go)) only calls `fireHookFor(src, hook, nil)`: the wrapper is deleted, `fireHookFor` takes the name `fireHook` with its `watched *Asset` parameter, and every caller passes `nil` or the watched asset (#166). `chooseDiscards` ([zones.go:84-86](../../../pkg/engine/zones.go)) only calls `chooseHandCards(pid, pid, n, upTo)`: it is deleted, its one caller calls `chooseHandCards` directly, its comment's sentences fold into `chooseHandCards`' comment, and the leftover `pid := chooser` goes (#167). It is a pure move: one authorized test edit, and match output is byte for byte the same. When this merges, #166 and #167 are `applied`.

## Inputs
- Board directives #166 and #167 (ADR 0007's forwarding wrappers), approved by the user 2026-11-18 and `open: before the next Deliverables task`. #167 says it "may share" #166's task. Design note Deliverable 6n ([m2-ability-model.md:221](../../../docs/architecture/m2-ability-model.md)) names both.
- Where it is today (at `autoloop/m2-wave-h.1` 69d56d9):
  - `fireHook(src, hook)` has ten callers in `trigger.go` (`fireDiscarded`, `fireEntersDiscard`, `fireDeploy`, `fireMoved`, `fireKilled`, `fireDamaged`, `fireExhausted`, `fireCombatKill`, `fireAttacks`, `fireTurnStart`) and one in a test, `TestDecisionSource_FlushInsideAScopeOrdersWithNone` ([decision_source_trigger_test.go:145](../../../pkg/engine/decision_source_trigger_test.go)). `fireHookFor` has one caller, `fireAttacks`' watcher loop, and `fireTurnStart`'s comment names it.
  - `chooseDiscards(pid, n, upTo)` has one caller, `chooseCostDiscards` ([costs.go:144](../../../pkg/engine/costs.go)), and `chooseHandCards`' comment names it. No test calls it (`git grep -n chooseDiscards -- "*_test.go"` finds nothing).
- Design (lead decisions):
  - `fireHook`'s one comment covers both readings of `watched`, and keeps the return-value sentence.
  - `chooseHandCards`' comment gains `chooseDiscards`' sentences on the owner choosing (no inspection, so an exact hand is not asked: `FR-MOD-001` §1, §6, `SVY-R16`) and on `upTo` (`FR-MOD-001` §6, `FR-TRN-003a`), and drops "chooseDiscards covers the owner choosing". `Decision.Player` is `chooser`, so `pid := chooser` goes.
  - `costs.go` changes only to update the reference to the deleted `chooseDiscards`, so under the D2 size rule it doesn't count toward the four files.
  - Docs: `engine-structs.md` already names only `fireHook` (its `trigger.go` row and the `triggerSource` arrow) and never named `chooseDiscards`, so it doesn't change. ADR 0007's lines naming both wrappers ([0007:12-13](../../../docs/architecture/decisions/0007-moves-update-every-reference.md)) and the design note's Deliverable 6n line are records of the wrappers and stay.
- Authorized test edit (#166), in the `test(T-0277)` commit: `TestDecisionSource_FlushInsideAScopeOrdersWithNone` ([decision_source_trigger_test.go:145](../../../pkg/engine/decision_source_trigger_test.go)) passes `nil` as `fireHook`'s third argument (Edit 1). No other test changes.
- Reuse search (D2.3): `git grep -n "fireHook\|chooseHandCards\|chooseDiscards" -- "*.go"` at 69d56d9 finds only the sites above. Nothing new is added.
- Size (D2.4): two non-test files (`trigger.go`, `zones.go`), plus one reference-only file and one authorized test edit.
- **Expected output difference:** none: `go tool task sim-compare` passes.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 69d56d9: `gofmt -l pkg cmd` clean, `go vet ./pkg/engine/` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-h.1` `[PASS]`. Blobs: `costs.go` 7a7e964, `decision_source_trigger_test.go` 2470eb8, `trigger.go` e07b322, `zones.go` f5f9a11.

## Commits
1. `test(T-0277): the flush-scope test passes nil as fireHook's watched asset` with Edit 1 only. At this commit `pkg/engine` does not build (case 1).
2. `refactor(engine): fireHook takes watched and fireHookFor goes; chooseDiscards goes (T-0277)` with Edits 2-4.

No double quotes in either message: the allow-list refuses an escaped one.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
### 1. `pkg/engine/decision_source_trigger_test.go`: the authorized test edit (commit 1)
```diff
@@ -142,7 +142,7 @@ func TestDecisionSource_FlushInsideAScopeOrdersWithNone(t *testing.T) {
 	s := newScene(t)
 	src := s.put(0, ZoneReserves, trigUnit("Source", deployTrigger(drawN(1)), deployTrigger(scratchN(1))))
 	outer := unit("Outer", 1, 1, ability.Keywords{})
-	s.g.fireHook(assetSource(src), ability.HookDeploy)
+	s.g.fireHook(assetSource(src), ability.HookDeploy, nil)
 	s.script(0, Answer{Picks: []int{0, 1}})
 	closeOuter := s.g.askFor(outer)
 	s.g.flushTriggers()
```

### 2. `pkg/engine/trigger.go`: one `fireHook`, with `watched`
```diff
@@ -83,24 +83,24 @@ func activeIn(src Zone, z ability.Zone) bool {
 // fireDiscarded buffers inst's discarded-from-hand triggers (FR-ZON-004).
 // discardCards calls it after inst is in its owner's discard pile.
 func (g *Game) fireDiscarded(inst *Instance) {
-	g.fireHook(pileSource(g, inst), ability.HookDiscardedFromHand)
+	g.fireHook(pileSource(g, inst), ability.HookDiscardedFromHand, nil)
 }
 
 // fireEntersDiscard buffers inst's enters-discard triggers (FR-ZON-004,
 // FR-ZON-005). putCard calls it after inst is in its owner's discard pile.
 func (g *Game) fireEntersDiscard(inst *Instance) {
-	g.fireHook(pileSource(g, inst), ability.HookEntersDiscard)
+	g.fireHook(pileSource(g, inst), ability.HookEntersDiscard, nil)
 }
 
 // fireDeploy buffers a's own deploy triggers (self hook only). The controller
 // is a's controller now. They wait for the flush after the item's sweep.
-func (g *Game) fireDeploy(a *Asset) { g.fireHook(assetSource(a), ability.HookDeploy) }
+func (g *Game) fireDeploy(a *Asset) { g.fireHook(assetSource(a), ability.HookDeploy, nil) }
 
 // fireMoved buffers a's own move triggers (self hook only). moveOnBattlefield
 // calls it after a is placed in its new zone (SVY-R1, FR-ZON-002).
 // Each trigger locks the zone a moved to, its own value, fixed by the event (SVY-R45).
 func (g *Game) fireMoved(a *Asset) {
-	added := g.fireHook(assetSource(a), ability.HookMoves)
+	added := g.fireHook(assetSource(a), ability.HookMoves, nil)
 	for i := range added {
 		added[i].locked = zoneLock(a)
 	}
@@ -109,20 +109,20 @@ func (g *Game) fireMoved(a *Asset) {
 // fireKilled buffers a's own death triggers. The deathSweep calls it before a
 // leaves the battlefield, so the controller is a's last-known controller. A
 // face-down asset fires nothing (FR-ZON-004).
-func (g *Game) fireKilled(a *Asset) { g.fireHook(assetSource(a), ability.HookKilled) }
+func (g *Game) fireKilled(a *Asset) { g.fireHook(assetSource(a), ability.HookKilled, nil) }
 
 // fireDamaged buffers a's own damage-taken triggers (Unbowed). dealDamage calls it
 // when a took damage, so the controller is a's controller now, kept if a then dies.
-func (g *Game) fireDamaged(a *Asset) { g.fireHook(assetSource(a), ability.HookDamaged) }
+func (g *Game) fireDamaged(a *Asset) { g.fireHook(assetSource(a), ability.HookDamaged, nil) }
 
 // fireExhausted buffers a's own exhaust triggers. exhaust calls it after a becomes Exhausted.
-func (g *Game) fireExhausted(a *Asset) { g.fireHook(assetSource(a), ability.HookExhausted) }
+func (g *Game) fireExhausted(a *Asset) { g.fireHook(assetSource(a), ability.HookExhausted, nil) }
 
 // fireCombatKill buffers att's own combat-kill triggers. resolveAttack calls it when
 // the attacker's combat damage left def lethal; the triggers wait on def's
 // destruction, because a replaced death is not a kill (SVY-R19).
 func (g *Game) fireCombatKill(att, def *Asset) {
-	added := g.fireHook(assetSource(att), ability.HookCombatKill)
+	added := g.fireHook(assetSource(att), ability.HookCombatKill, nil)
 	for i := range added {
 		added[i].awaits = def.ID
 	}
@@ -137,19 +137,15 @@ func (g *Game) confirmDestroyed(id int) {
 	}
 }
 
-// fireHook buffers src's own triggers for hook in printed order: the one place
-// a trigger's gates are checked. A face-down asset has no printed abilities.
-// It returns the entries it added, a slice of the buffer, so a caller can set a
-// field on each before the flush; the other callers ignore it.
-func (g *Game) fireHook(src triggerSource, hook ability.Hook) []pendingTrigger {
-	return g.fireHookFor(src, hook, nil)
-}
-
-// fireHookFor is fireHook for an event of watched: nil for src's own event.
-// With watched nil it fires src's self-hook triggers; with watched set it fires
-// only src's watching triggers (Trigger.Watches) whose Watch matches watched,
-// judged against src's controller and source, and records watched on each.
-func (g *Game) fireHookFor(src triggerSource, hook ability.Hook, watched *Asset) []pendingTrigger {
+// fireHook buffers src's triggers for hook in printed order: the one place a
+// trigger's gates are checked. A face-down asset has no printed abilities.
+// watched is the asset whose event it is: nil for src's own event, which fires
+// src's self-hook triggers; set, it fires only src's watching triggers
+// (Trigger.Watches) whose Watch matches watched, judged against src's
+// controller and source, and records watched on each. It returns the entries
+// it added, a slice of the buffer, so a caller can set a field on each before
+// the flush; the other callers ignore it.
+func (g *Game) fireHook(src triggerSource, hook ability.Hook, watched *Asset) []pendingTrigger {
 	n := len(g.pending)
 	def := src.card.Ability
 	if def == nil {
@@ -210,18 +206,18 @@ func (g *Game) fireAttacks(a *Asset) {
 	if n := g.Keywords(a).Aggressive; n > 0 {
 		g.bufferTrigger(src, fmt.Sprintf("%s: %s %d", src.card.Name, ability.KeywordAggressive.String(), n), ability.Aggressive(n))
 	}
-	g.fireHook(src, ability.HookAttacks)
+	g.fireHook(src, ability.HookAttacks, nil)
 	for _, w := range g.assetsByID() {
-		g.fireHookFor(assetSource(w), ability.HookAttacks, a)
+		g.fireHook(assetSource(w), ability.HookAttacks, a)
 	}
 }
 
 // fireTurnStart buffers the start-of-turn triggers of every battlefield asset,
-// either player's (FR-TRN-001). fireHookFor holds the gates, so "at the start
+// either player's (FR-TRN-001). fireHook holds the gates, so "at the start
 // of your turn" is a Turn: TurnYours gate.
 func (g *Game) fireTurnStart() {
 	for _, a := range g.assetsByID() {
-		g.fireHook(assetSource(a), ability.HookTurnStart)
+		g.fireHook(assetSource(a), ability.HookTurnStart, nil)
 	}
 }
```

### 3. `pkg/engine/zones.go`: `chooseDiscards` goes, and `chooseHandCards` says what it said
```diff
@@ -74,31 +74,22 @@ func (g *Game) discardFromHand(chooser, owner PlayerID, n int, upTo bool) []*Ins
 	return picked
 }
 
-// chooseDiscards has player pid pick n cards of their choice from their own
-// hand, or the whole hand if it holds fewer, and returns them in pick order.
-// It moves nothing. Choosing from your own hand is not an inspection, so a hand
-// of exactly that many cards is picked without asking (FR-MOD-001 §1, §6;
-// SVY-R16). With upTo it picks up to n cards, possibly none: the choice is
-// asked whenever the hand isn't empty, and a timeout picks none (FR-MOD-001 §6,
-// FR-TRN-003a).
-func (g *Game) chooseDiscards(pid PlayerID, n int, upTo bool) []*Instance {
-	return g.chooseHandCards(pid, pid, n, upTo)
-}
-
 // chooseHandCards has player chooser pick n cards from owner's hand, or the
 // whole hand if it holds fewer, and returns them in pick order. It moves
-// nothing. A chooser who isn't the owner inspects the owner's revealed hand
+// nothing. Choosing from your own hand is not an inspection, so an owner's
+// hand of exactly n cards is picked without asking (FR-MOD-001 §1, §6;
+// SVY-R16). A chooser who isn't the owner inspects the owner's revealed hand
 // (FR-MOD-001 §3), so the decision is Hidden and asked even for a hand of
-// exactly n cards (§1, SVY-R50); chooseDiscards covers the owner choosing.
-// Every hand-discard decision is built here.
+// exactly n cards (§1, SVY-R50). With upTo it picks up to n cards, possibly
+// none: the choice is asked whenever the hand isn't empty, and a timeout picks
+// none (FR-MOD-001 §6, FR-TRN-003a). Every hand-discard decision is built here.
 func (g *Game) chooseHandCards(chooser, owner PlayerID, n int, upTo bool) []*Instance {
-	pid := chooser
 	hand := g.Players[owner].Hand
 	n = min(n, len(hand))
 	if n <= 0 {
 		return nil
 	}
-	d := Decision{Player: pid, Kind: DecideHand, Prompt: "Discard a card", Min: n, Max: n, Hidden: chooser != owner}
+	d := Decision{Player: chooser, Kind: DecideHand, Prompt: "Discard a card", Min: n, Max: n, Hidden: chooser != owner}
 	if n > 1 {
 		d.Prompt = fmt.Sprintf("Discard %d cards", n)
 	}
```

### 4. `pkg/engine/costs.go`: the discard cost calls `chooseHandCards` (reference update only)
```diff
@@ -141,7 +141,7 @@ func (g *Game) chooseCostDiscards(pl placement) []*Instance {
 			panic("engine: a placement has two discard costs")
 		}
 		found = true
-		picked = g.chooseDiscards(pl.controller, c.UpTo, true)
+		picked = g.chooseHandCards(pl.controller, pl.controller, c.UpTo, true)
 	}
 	return picked
 }
```

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The test edit fails before the implementation:** at the `test(T-0277)` commit, `go -C W vet ./pkg/engine/` fails, reporting too many arguments in the call to `s.g.fireHook` in `decision_source_trigger_test.go`.
2. **One `fireHook`, with three parameters:** `git -C W grep -n "fireHook(src triggerSource" -- pkg` prints exactly one line, the declaration in `pkg/engine/trigger.go`, ending `fireHook(src triggerSource, hook ability.Hook, watched *Asset) []pendingTrigger {`.
3. **The old names are gone from the code:** `git -C W grep -n fireHookFor -- "*.go"` prints nothing, and `git -C W grep -n chooseDiscards -- "*.go"` prints nothing.
4. **No leftover `pid` in `chooseHandCards`:** `git -C W grep -n "pid := chooser" -- pkg` prints nothing.
5. **The same tests pass, with only the authorized edit:** `go -C W test ./pkg/...` passes, and `git -C W diff autoloop/m2-wave-h.1 --numstat -- "*_test.go"` prints exactly one line, `1	1	pkg/engine/decision_source_trigger_test.go`.
6. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-h.1` prints `[PASS]`.
7. **Only the listed edits:** `git -C W diff autoloop/m2-wave-h.1 --numstat -- pkg docs` prints exactly four lines: `1	1	pkg/engine/costs.go`, `1	1	pkg/engine/decision_source_trigger_test.go`, `21	25	pkg/engine/trigger.go` and `7	16	pkg/engine/zones.go`.
8. **Nothing else changed:** `git -C W diff autoloop/m2-wave-h.1 --stat` lists only the four files of case 7 and this task file.
9. `gofmt -l W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/trigger.go` (Edit 2)
- `pkg/engine/zones.go` (Edit 3)
- `pkg/engine/costs.go` (reference update only, Edit 4)

## Tests
- `pkg/engine/decision_source_trigger_test.go`: the authorized edit to `TestDecisionSource_FlushInsideAScopeOrdersWithNone` (Edit 1), in the `test(T-0277)` commit. No other test file changes.

## Done When
- [ ] Cases 1-9 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-18, lead-architect (wave H.1 run, session 6): written in full against 69d56d9 for #166 and #167. `main` (7a8e4f5) has nothing the run branch lacks, so no sync. Not `WAVE-COMPLETE` before the task: #166, #167, #162, #163 and #164 are `open: before the next Deliverables task`. #166 and #167 share this task, as #167 allows: together they touch two non-test files that aren't reference-only, within the D2 size rule. This is a harness task, so it skips `tests` and goes straight to `ready`. Tier is `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list: one per line, no pipes, no `$`, and patterns with parentheses are quoted. Waiting directives checked: #6, #7 (mode `serial`), #15 (2) (no draw or end-of-turn hook), #71, #91 (2), #106, #120, #142 (wave H.1 has no cards): none has fired. Next: `kernel-engineer` on Flash.
