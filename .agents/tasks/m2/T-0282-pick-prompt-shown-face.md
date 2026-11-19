---
id: T-0282
title: "Pick prompts name the card the stack item shows, through `sourceName` (directive #170; wave H.2 harness, flash)"
kind: harness
tier: flash
wave: H.2 (board directive #170)
status: review
branch: m2/autoloop-wave-h.2/T-0282-pick-prompt-shown-face
depends_on: [T-0281]
claims: [func:Game.sourceName, func:Game.pickPrompt]
cards: []
requirements: [FR-ZON-005]
---

## Goal
An untargeted pick's prompt and a pool's prompt name the card the stack item shows as its source (`SourceCard`, `FR-ZON-005`), not always the real card. A source turned face down after its ability was placed then shows only the Scratch face, in the decision and in its match record, as the stack label and `Decision.Source` already do. The "name of the card it shows, empty for none" rule is written once, in a new `(g *Game) sourceName(it *StackItem) string` beside `SourceCard` in `placement.go`, and `StackLabel`'s inline copy of it reads it too. When this merges, #170 is `applied`.

## Inputs
- Board directive #170 (structural review 14, finding 2), approved by the user 2026-11-19, `open: before the next Deliverables task`; wave H.2, after #169 (T-0281, `done`). Design note Deliverable 6o ([m2-ability-model.md:229](../../../docs/architecture/m2-ability-model.md)).
- Where it is today (at `autoloop/m2-wave-h.2` e026c00):
  - `pickPrompt(it, st)` ([targeting.go:229-236](../../../pkg/engine/targeting.go)) names `cardName(it.Inst)`; `pickAssetDecision` (:199) and `pickCards` (:242) call it. `splitAssets` (:215) names `cardName(r.it.Inst)` in its "distribute" prompt.
  - `SourceCard` ([placement.go:79-81](../../../pkg/engine/placement.go)) is face-aware. `StackLabel` ([stack.go:97-100](../../../pkg/engine/stack.go)) turns it into a name inline, with an empty name for nil.
  - `Decision.Source` is already face-aware: `resolveSteps` opens `g.askFor(g.SourceCard(r.it))` ([effects.go:59](../../../pkg/engine/effects.go)), so a pick's record already has `source: "Scratch Company"`. Only its `Prompt` names the hidden card. The record's `Prompt` and its `Detail` reach the other seat's log, because a pick isn't `Hidden` ([log.go:32](../../../pkg/view/log.go)).
- Design (lead decisions):
  - `sourceName` goes beside `SourceCard` in `placement.go`, which `doc.go` gives "the card it shows as its source". The three places that need a name (`StackLabel`, `pickPrompt`, `splitAssets`) read it. Adding the prompts' face-aware name inline would have made three copies of the nil check (D2.0), which is why `StackLabel` changes too.
  - `pickPrompt` becomes a method on `*Game`, since it now needs `g`. Its two callers already have `g`.
  - `StackLabel`'s Deploy and Maneuver cases keep `cardName(it.Inst)`: their card is their own and is never face down, and #174 deals with the play screen's wording of them.
  - `doc.go`'s Files list names `sourceName` beside `SourceCard`. #176 (b) covers the design note and `engine-structs.md`.
- Reuse search (D2.3): `git grep -n "cardName(" -- pkg/engine` finds, outside `targeting.go`, only `StackLabel`'s Deploy and Maneuver cases and event fields that name a card instance itself, not a stack item's source. No function already returns a stack item's shown name.
- Size (D2.4): four non-test files (`placement.go`, `targeting.go`, `stack.go`, `doc.go`), plus one new test file.
- **Expected output difference:** none in the sims. `go tool task sim-compare` passes, because no starter or stress game turns a waiting pick's source face down. The lead's dry run confirmed it.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at e026c00. The new test failed in all three subtests without Edits 1-4 and passed with them. `gofmt -l pkg cmd` was clean, `go vet ./pkg/engine/` clean, `go tool task check` exited 0, and `go tool task sim-compare -- autoloop/m2-wave-h.2` printed `[PASS]`. Blobs: `pick_prompt_test.go` 15bf7d7, `placement.go` 102cd8c, `stack.go` 35bb3ac, `targeting.go` 2f0f034, `doc.go` e134c7f.

## Commits
1. `test(T-0282): pick prompts name the shown face` with Edit 5 only. Run case 1 before committing it and record the failure.
2. `fix(engine): pick prompts name the card the stack item shows, sourceName (T-0282)` with Edits 1-4.

No double quotes in either message: the allow-list refuses an escaped one.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
### 1. `pkg/engine/placement.go`: `sourceName` after `SourceCard`
```diff
@@ -80,6 +80,17 @@ func (g *Game) SourceCard(it *StackItem) *Card {
 	return g.shownCard(it.Source, it.Inst)
 }
 
+// sourceName is the name of the card stack item it shows as its source
+// (SourceCard), "" when it shows none. A stack item's label and the prompts
+// of the decisions it asks as it resolves read it, so neither names a card
+// that is face down (FR-ZON-005).
+func (g *Game) sourceName(it *StackItem) string {
+	if c := g.SourceCard(it); c != nil {
+		return c.Name
+	}
+	return ""
+}
+
 // maneuverPlacement is the placement of the Maneuver inst that pid enacts, for
 // both its legality check and the choosing of its modes and targets.
 func maneuverPlacement(pid PlayerID, inst *Instance) placement {
```

### 2. `pkg/engine/stack.go`: `StackLabel` reads `sourceName`
```diff
@@ -94,10 +94,7 @@ func (g *Game) StackLabel(it *StackItem) string {
 	case StackManeuver:
 		return "Maneuver " + cardName(it.Inst)
 	case StackAbility:
-		name := ""
-		if c := g.SourceCard(it); c != nil {
-			name = c.Name
-		}
+		name := g.sourceName(it)
 		if it.Keyword != ability.KeywordNone {
 			return it.Keyword.String() + " trigger of " + name
 		}
```

### 3. `pkg/engine/targeting.go`: `pickPrompt` becomes a method that reads `sourceName`, and `splitAssets` reads it
```diff
@@ -196,7 +196,7 @@ func (g *Game) pickAssets(r *resolution, st ability.Step) []*Asset {
 // are every battlefield asset matching st.Filter now, oldest first. Both
 // pickAssets and splitAssets ask it, so a pool's options are SubjectPick's.
 func (g *Game) pickAssetDecision(r *resolution, st ability.Step) (Decision, []*Asset) {
-	return g.assetDecision(r.it.Controller, DecidePick, pickPrompt(r.it, st), st.UpTo, func(a *Asset) bool {
+	return g.assetDecision(r.it.Controller, DecidePick, g.pickPrompt(r.it, st), st.UpTo, func(a *Asset) bool {
 		return g.matchesFilterAt(r.it.Controller, r.it.Source, r.it.Locked, st.Filter, a)
 	})
 }
@@ -212,7 +212,7 @@ func (g *Game) splitAssets(r *resolution, st ability.Step, pool int) ([]*Asset,
 		return nil, nil
 	}
 	d, opts := g.pickAssetDecision(r, st)
-	d.Prompt = fmt.Sprintf("%s: distribute %d", cardName(r.it.Inst), pool)
+	d.Prompt = fmt.Sprintf("%s: distribute %d", g.sourceName(r.it), pool)
 	var as []*Asset
 	var shares []int
 	for _, a := range pickFrom(g, repeatedPicks(d, pool), opts) {
@@ -226,9 +226,11 @@ func (g *Game) splitAssets(r *resolution, st ability.Step, pool int) ([]*Asset,
 	return as, shares
 }
 
-// pickPrompt is the prompt of an untargeted pick step.
-func pickPrompt(it *StackItem, st ability.Step) string {
-	name := cardName(it.Inst)
+// pickPrompt is the prompt of an untargeted pick step. It names the card the
+// item shows as its source (sourceName), the Scratch face for a source turned
+// face down after the item was placed (FR-ZON-005).
+func (g *Game) pickPrompt(it *StackItem, st ability.Step) string {
+	name := g.sourceName(it)
 	if st.UpTo > 0 {
 		return fmt.Sprintf("%s: pick up to %d", name, st.UpTo)
 	}
@@ -239,7 +241,7 @@ func pickPrompt(it *StackItem, st ability.Step) string {
 // picks from every card matching st.Filter now, as pickAssets does for assets
 // (SVY-R3, SVY-R14). The options are fixed before any card moves.
 func (g *Game) pickCards(r *resolution, st ability.Step) []*Instance {
-	d, opts := g.pileDecision(r.it.Controller, DecidePick, pickPrompt(r.it, st), st.UpTo, func(inst *Instance) bool {
+	d, opts := g.pileDecision(r.it.Controller, DecidePick, g.pickPrompt(r.it, st), st.UpTo, func(inst *Instance) bool {
 		return g.matchesCard(r.it.Controller, r.it.Inst, st.Filter, inst)
 	})
 	return pickFrom(g, d, opts)
```

### 4. `pkg/engine/doc.go`: the Files list names `sourceName`
```diff
@@ -116,9 +116,9 @@
 //     StackLabel, the one wording of what a stack item is.
 //   - trigger.go: the pending-trigger buffer and firing. placement.go: placing
 //     an effect, its modes and then its targets, the card it shows as its
-//     source (shownCard, SourceCard), and whether an optional trigger could do
-//     nothing (couldDoNothing), and the one lock of an "in the same zone"
-//     effect on placement (sameZoneLock). costs.go: a card's cost and
+//     source (shownCard, SourceCard) and its name (sourceName), and whether an
+//     optional trigger could do nothing (couldDoNothing), and the one lock of
+//     an "in the same zone" effect on placement (sameZoneLock). costs.go: a card's cost and
 //     a placement's costs. resources.go: Requisition and Tactics, generating and
 //     converting them, and paying with them. targeting.go: target choice and
 //     untargeted picks. filter.go: checking assets and discard-pile cards against
```

### 5. `pkg/engine/pick_prompt_test.go` (new file, the whole of it)
```go
package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0282 (directive #170, FR-ZON-005): an untargeted pick's prompt,
// and a pool's, name the card the stack item shows as its source (SourceCard),
// so a source turned face down after its ability was placed shows only the
// Scratch face, in the decision asked and in its match record.

func TestFR_ZON_005_PickPromptNamesTheShownFace(t *testing.T) {
	// Covers: FR-MOD-001
	enemy := ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}
	for _, tc := range []struct {
		name string
		step ability.Step
		want string
	}{
		{"pick", ability.Step{Op: ability.OpKill, Subject: ability.SubjectPick, Filter: enemy}, "Scratch Company: pick one"},
		{"pick up to", ability.Step{Op: ability.OpKill, Subject: ability.SubjectPick, Filter: enemy, UpTo: 2}, "Scratch Company: pick up to 2"},
		{"pool", ability.Step{Op: ability.OpAddArmor, Amount: ability.Value{N: 2}, Subject: ability.SubjectPick, Split: true, Filter: enemy}, "Scratch Company: distribute 2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newScene(t)
			a := s.put(0, ZoneMainTheater, unit("Hidden Source", 1, 1, ability.Keywords{}))
			s.put(1, ZoneMainTheater, unit("Foe A", 1, 5, ability.Keywords{}))
			s.put(1, ZoneMainTheater, unit("Foe B", 1, 5, ability.Keywords{}))
			s.g.push(&StackItem{Kind: StackAbility, Controller: 0, Source: a.ID, Inst: a.Inst,
				Effect: ability.Effect{Steps: []ability.Step{tc.step}}})
			a.FaceDown = true
			s.script(0, Answer{Timeout: true})
			s.g.resolveStack()
			var asked []Decision
			for _, d := range s.asked(0) {
				if d.Kind == DecidePick {
					asked = append(asked, d)
				}
			}
			if len(asked) != 1 {
				t.Fatalf("asked %d pick decisions, want 1", len(asked))
			}
			if asked[0].Prompt != tc.want {
				t.Errorf("prompt %q, want %q", asked[0].Prompt, tc.want)
			}
			var recs []*DecisionRecord
			for _, e := range eventsOfKind(s.g.Log, EventDecision) {
				if e.Decision.Kind == DecidePick.String() {
					recs = append(recs, e.Decision)
				}
			}
			if len(recs) != 1 {
				t.Fatalf("logged %d pick records, want 1", len(recs))
			}
			if recs[0].Prompt != tc.want || recs[0].Source != "Scratch Company" {
				t.Errorf("record prompt %q, source %q; want %q and the Scratch face", recs[0].Prompt, recs[0].Source, tc.want)
			}
		})
	}
}
```

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The test fails first:** with only Edit 5 in W, `go -C W test ./pkg/engine -run TestFR_ZON_005_PickPromptNamesTheShownFace` fails. Each of the subtests `pick`, `pick_up_to` and `pool` reports a prompt naming `Hidden Source`. After Edits 1-4 it passes.
2. **No real card name in a pick prompt:** `git -C W grep -n "cardName(" -- pkg/engine/targeting.go` prints nothing.
3. **One name rule, three readers:** `git -C W grep -n "g.sourceName(" -- pkg/engine` prints exactly three lines: `pkg/engine/stack.go:97`, `pkg/engine/targeting.go:215` and `pkg/engine/targeting.go:233`.
4. **No old free function:** `git -C W grep -n "func pickPrompt" -- pkg/engine` prints nothing, and `git -C W grep -n "g.pickPrompt(" -- pkg/engine` prints exactly two lines, `targeting.go:199` and `targeting.go:244`.
5. **The other tests pass, unedited:** `go -C W test ./pkg/... ./cmd/...` passes (the `pkg/view/viewtest` leak tests included), and `git -C W diff autoloop/m2-wave-h.2 --numstat -- "*_test.go"` prints exactly one line, `62	0	pkg/engine/pick_prompt_test.go`.
6. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-h.2` prints `[PASS]`.
7. **Only the listed edits:** `git -C W diff autoloop/m2-wave-h.2 --numstat -- pkg cmd docs` prints exactly five lines: `3	3	pkg/engine/doc.go`, `62	0	pkg/engine/pick_prompt_test.go`, `11	0	pkg/engine/placement.go`, `1	4	pkg/engine/stack.go` and `8	6	pkg/engine/targeting.go`.
8. **Nothing else changed:** `git -C W diff autoloop/m2-wave-h.2 --stat` lists only the five files of case 7 and this task file.
9. `gofmt -l W/pkg W/cmd` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/placement.go` (Edit 1)
- `pkg/engine/stack.go` (Edit 2)
- `pkg/engine/targeting.go` (Edit 3)
- `pkg/engine/doc.go` (Edit 4, comment only)

## Tests
- New: `TestFR_ZON_005_PickPromptNamesTheShownFace` in `pkg/engine/pick_prompt_test.go` (Edit 5), in the `test(T-0282)` commit. No existing test is edited.

## Done When
- [x] Cases 1-9 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-h.2` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-19, lead-architect (wave H.2 run, session 2): written in full against e026c00 for #170, the board's first `open: before` directive in wave H.2's order. `main` (1859fe4) has nothing the run branch lacks, so no sync. Not `WAVE-COMPLETE` before the task: #170, #172, #173 and #174 are `open: before the next Deliverables task`. This is a harness task, so it skips `tests` and goes straight to `ready`. Tier is `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list: one per line, no pipes, no `$`, and patterns with special characters are quoted. Waiting directives checked: #6, #7 (mode `serial`), #15 (2) (no draw or end-of-turn hook), #71, #91 (2), #106, #120, #142 (wave H.2 has no cards): none has fired. No directive added. Next: `kernel-engineer` on Flash.
- 2026-11-19, kernel-engineer (wave H.2 run, session 2): implemented directive #170.
  - Commits:
    - 599ef24d test(T-0282): pick prompts name the shown face
    - b6072d79 fix(engine): pick prompts name the card the stack item shows, sourceName (T-0282)
  - Changes:
    - Added `(g *Game) sourceName(it *StackItem) string` to `pkg/engine/placement.go`.
    - Updated `StackLabel` in `pkg/engine/stack.go` to use `g.sourceName(it)`.
    - Made `pickPrompt` a method `(g *Game) pickPrompt(it *StackItem, st ability.Step) string` in `pkg/engine/targeting.go` calling `g.sourceName(it)`, and updated `splitAssets` to read `g.sourceName(r.it)`.
    - Updated `pkg/engine/doc.go` Files section to note `sourceName`.
    - Added `pkg/engine/pick_prompt_test.go` with `TestFR_ZON_005_PickPromptNamesTheShownFace`.
  - Reuse: `sourceName` wraps `SourceCard`. Verified `git grep -n "cardName(" -- pkg/engine` finds no other stack-item source lookups.
  - Verification:
    - Case 1: `go -C W test ./pkg/engine -run TestFR_ZON_005_PickPromptNamesTheShownFace` failed on all 3 subtests naming `Hidden Source` with only Edit 5; passed after Edits 1-4.
    - Case 2: `git -C W grep -n "cardName(" -- pkg/engine/targeting.go` printed nothing (exit 1).
    - Case 3: `git -C W grep -n "g.sourceName(" -- pkg/engine` printed exactly 3 lines: `stack.go:97`, `targeting.go:215`, `targeting.go:233`.
    - Case 4: `git -C W grep -n "func pickPrompt" -- pkg/engine` printed nothing; `git -C W grep -n "g.pickPrompt(" -- pkg/engine` printed exactly 2 lines: `targeting.go:199`, `targeting.go:244`.
    - Case 5: `go -C W test ./pkg/... ./cmd/...` passed, and `git -C W diff autoloop/m2-wave-h.2 --numstat -- "*_test.go"` printed `62	0	pkg/engine/pick_prompt_test.go`.
    - Case 6: `go -C W tool task sim-compare -- autoloop/m2-wave-h.2` printed `[PASS] starter and stress sims match autoloop/m2-wave-h.2 byte for byte, exit codes included`.
    - Case 7: `git -C W diff autoloop/m2-wave-h.2 --numstat -- pkg cmd docs` printed exactly the 5 specified lines.
    - Case 8: `git -C W diff autoloop/m2-wave-h.2 --stat` showed only the 5 files plus task file.
    - Case 9: `gofmt -l W/pkg W/cmd` printed nothing; `go -C W tool task check` passed all vet, tests, and all linters; `go -C W tool task sim` passed 200 games with all invariants holding.
  - Status moved to `review`. Ready for review.
