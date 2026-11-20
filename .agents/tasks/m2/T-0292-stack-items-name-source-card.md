---
id: T-0292
title: "Every stack item names its card through `SourceCard`, and `cardName` goes (directive #178; wave H.3 harness, flash, pure move)"
kind: harness
tier: flash
wave: H.3 (board directive #178)
status: in-progress
branch: m2/autoloop-wave-h.3/T-0292-stack-items-name-source-card
depends_on: [T-0291]
claims: [func:engine.Game.push, func:engine.Game.StackLabel, func:engine.cardName, func:view.stackView]
cards: []
requirements: []
---

## Goal
A Deploy's and a Maneuver's stack item name their card the way an ability's does, through `Game.SourceCard` and `sourceName` (the one rule for which card a stack item shows, #153, `FR-ZON-005`), in the place event (`push`), in `StackLabel` and in the view's `stackView`. Nothing reads a stack item's `Inst.Card.Name` directly any more, and `cardName` has no caller and goes (ADR 0007). Every output stays the same. When this merges, #178 is `applied`.

## Inputs
- Board directive #178 (structural review 15, finding 1), approved by the user 2026-11-20, `open: before the next Deliverables task`; wave H.3, after #177 (T-0291, `done`).
- Where it is today (at `autoloop/m2-wave-h.3` 674f8ef):
  - `StackLabel` ([stack.go:84-105](../../../pkg/engine/stack.go)) names a Deploy and a Maneuver by `cardName(it.Inst)` ([events.go:85-91](../../../pkg/engine/events.go)), and an ability by `g.sourceName(it)`.
  - `stackView` ([view.go:218-240](../../../pkg/view/view.go)) reads `it.Inst.Card.Name` for a Deploy and a Maneuver, and `g.SourceCard(it)` for an ability.
  - `push` ([stack.go:57-77](../../../pkg/engine/stack.go)) sets the place event's `Card` and `Inst` from `item.Inst` for a Deploy and a Maneuver, and from `g.SourceCard(item)` for an ability. #178 names only `StackLabel` and `stackView`, but its title is "every stack item names its card through `SourceCard`", and `push` is a third copy of the same lookup (D5.3 would make it a new blocking directive), so this task moves it too (lead decision).
- Why the move changes nothing: a Deploy's and a Maneuver's item always has `Source` 0 (`deployPlacement`, `maneuverPlacement`, [placement.go:96-119](../../../pkg/engine/placement.go); `declare` builds every such item), so `SourceCard` (`shownCard(0, Inst)`, [placement.go:65-73](../../../pkg/engine/placement.go)) returns `Inst.Card`. In `push`, `c == item.Inst.Card` then holds, so the event's `Inst` is still `item.Inst.ID`. Every Deploy and Maneuver item built in a test (`pkg/engine/stack_label_test.go`, `pkg/engine/decision_source_test.go`, `pkg/view/view_test.go`) also has `Source` 0.
- Design (lead decisions):
  - `stackView` sets `Card` in a first `switch`, for exactly the kinds that name a card (Deploy, Ability, Maneuver). Setting it for every kind, before one switch, would give a March, an Attack and a Scratch the source asset's card, which changes the view's JSON. The second `switch` sets only each kind's own fields, and a Deploy and a March share their `Zone` case.
  - `StackLabel`'s comment says that a Deploy, a Maneuver and an ability name their card through `sourceName`.
  - `SourceCard` and `sourceName` keep their declarations and comments ([placement.go:75-92](../../../pkg/engine/placement.go)); the comment on `SourceCard` already says it gives a Deploy's or a Maneuver's own card.
  - `pkg/view/viewtest`'s `PublicNames` keeps reading `it.Inst.Card.Name`: it is the leak tests' oracle, written apart from the code it checks.
  - `docs/architecture/engine-structs.md` doesn't list `cardName`, so no doc changes.
- Reuse search (D2.3): `git grep -n "Inst.Card.Name" -- pkg cmd ":!*_test.go"` finds `stack.go:62`, `view.go:222` and `view.go:235` (the three moved here), plus `events.go:198` (an HQ), `snapshot.go:65` and `sweep.go:21` (assets) and `viewtest.go` (the oracle), which don't name stack items.
- Size (D2.4): three non-test files.
- **Expected output difference:** none. `go tool task sim-compare` passes, and every test passes unedited.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 674f8ef. `gofmt -l pkg cmd` was clean, `go vet ./pkg/... ./cmd/...` clean, `go tool task check` exited 0, `go tool task sim-compare -- autoloop/m2-wave-h.3` printed `[PASS]`, and `go tool task sim` exited 0 (P0 58.5% / P1 39.5% / draws 2.0%). Blobs: `stack.go` dd57781, `events.go` c127df0, `view.go` f07b157.

## Commits
1. `refactor(engine): every stack item names its card through SourceCard, and cardName goes (T-0292)` with Edits 1-3. No double quotes in the message: the allow-list refuses an escaped one. A pure move adds no test, so there is no `test(T-0292)` commit.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
### 1. `pkg/engine/stack.go`: `push` and `StackLabel` read `SourceCard` for a Deploy and a Maneuver
```diff
@@ -58,11 +58,9 @@ func (g *Game) push(item *StackItem) {
 	g.Stack = append(g.Stack, item)
 	e := Event{Kind: EventPlace, Player: int(item.Controller), Source: item.Source, Amount: len(g.Stack)}
 	switch item.Kind {
-	case StackDeploy, StackManeuver:
-		e.Card, e.Inst = item.Inst.Card.Name, item.Inst.ID
 	case StackAttack:
 		e.Target = item.Target
-	case StackAbility:
+	case StackDeploy, StackManeuver, StackAbility:
 		// The card both seats see (SourceCard), and its instance ID only when
 		// that is the instance's own card: a face-down source's ID could
 		// identify the real card (FR-ZON-005).
@@ -78,13 +76,14 @@ func (g *Game) push(item *StackItem) {
 
 // StackLabel names stack item it, in the one wording of what it is: "Deploy of
 // X", "March", "Attack", "Scratch", "Maneuver X", "ability of X" or "<keyword>
-// trigger of X". An ability names its card as both seats see it (SourceCard,
-// FR-ZON-005), or no card if it has none. The place event and the play
-// screen's stack listing (view.StackView.Label) both read it.
+// trigger of X". A Deploy, a Maneuver and an ability name their card as both
+// seats see it (sourceName, FR-ZON-005), or no card if there is none. The
+// place event and the play screen's stack listing (view.StackView.Label) both
+// read it.
 func (g *Game) StackLabel(it *StackItem) string {
 	switch it.Kind {
 	case StackDeploy:
-		return "Deploy of " + cardName(it.Inst)
+		return "Deploy of " + g.sourceName(it)
 	case StackMarch:
 		return "March"
 	case StackAttack:
@@ -92,7 +91,7 @@ func (g *Game) StackLabel(it *StackItem) string {
 	case StackScratch:
 		return "Scratch"
 	case StackManeuver:
-		return "Maneuver " + cardName(it.Inst)
+		return "Maneuver " + g.sourceName(it)
 	case StackAbility:
 		name := g.sourceName(it)
 		if it.FromKeyword != ability.KeywordTriggerNone {
```

### 2. `pkg/engine/events.go`: `cardName` goes (no caller left)
```diff
@@ -82,14 +82,6 @@ func (g *Game) settled() {
 	}
 }
 
-// cardName is the name of inst's card, "" for nil.
-func cardName(inst *Instance) string {
-	if inst == nil {
-		return ""
-	}
-	return inst.Card.Name
-}
-
 func assetID(a *Asset) int {
 	if a == nil {
 		return 0
```

### 3. `pkg/view/view.go`: `stackView` sets `Card` from `SourceCard` once
```diff
@@ -218,22 +218,19 @@ func chosenTargets(g *engine.Game, it *engine.StackItem) ([][]int, [][]string) {
 func stackView(g *engine.Game, it *engine.StackItem) StackView {
 	s := StackView{Kind: stackKinds[it.Kind], Controller: int(it.Controller), Source: it.Source, Label: g.StackLabel(it)}
 	switch it.Kind {
-	case engine.StackDeploy:
-		s.Card, s.Zone = it.Inst.Card.Name, it.Zone.String()
-	case engine.StackMarch:
-		s.Zone = it.Zone.String()
-	case engine.StackAttack:
-		s.Target = it.Target
-	case engine.StackAbility:
-		// A face-down source names only the Scratch face (Game.SourceCard).
+	case engine.StackDeploy, engine.StackAbility, engine.StackManeuver:
+		// The card both seats see: a face-down source names only the Scratch
+		// face (Game.SourceCard).
 		if c := g.SourceCard(it); c != nil {
 			s.Card = c.Name
 		}
-		s.Chosen, s.ChosenCards = chosenTargets(g, it)
-	case engine.StackManeuver:
-		if it.Inst != nil {
-			s.Card = it.Inst.Card.Name
-		}
+	}
+	switch it.Kind {
+	case engine.StackDeploy, engine.StackMarch:
+		s.Zone = it.Zone.String()
+	case engine.StackAttack:
+		s.Target = it.Target
+	case engine.StackAbility, engine.StackManeuver:
 		s.Chosen, s.ChosenCards = chosenTargets(g, it)
 	}
 	return s
```

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **`cardName` is gone:** `git -C W grep -n "cardName(" -- pkg/engine` prints nothing.
2. **No stack item's card is read from `Inst` directly:** `git -C W grep -n "Inst.Card.Name" -- pkg/engine/stack.go pkg/view/view.go` prints nothing.
3. **`SourceCard` and `sourceName` keep their declarations:** `git -C W grep -n "func (g \*Game) SourceCard" -- pkg/engine` prints exactly `pkg/engine/placement.go:79`, and `git -C W grep -n "func (g \*Game) sourceName" -- pkg/engine` prints exactly `pkg/engine/placement.go:87`.
4. **Each reader goes through them:** `git -C W grep -n "g.sourceName(it)" -- pkg/engine/stack.go` prints exactly three lines, `pkg/engine/stack.go:86`, `:94` and `:96`; `git -C W grep -n "g.SourceCard(" -- pkg/engine/stack.go pkg/view/view.go` prints exactly two, `pkg/engine/stack.go:67` and `pkg/view/view.go:224`.
5. **The same tests pass, unedited:** `go -C W test ./pkg/... ./cmd/...` passes (the `pkg/view/viewtest` leak tests and `TestPlay_*` included), and `git -C W diff autoloop/m2-wave-h.3 --numstat -- "*_test.go"` prints nothing.
6. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-h.3` prints `[PASS]`.
7. **Only the listed edits:** `git -C W diff autoloop/m2-wave-h.3 --numstat -- pkg cmd docs` prints exactly three lines: `0	8	pkg/engine/events.go`, `7	8	pkg/engine/stack.go` and `10	13	pkg/view/view.go`.
8. **Nothing else changed:** `git -C W diff autoloop/m2-wave-h.3 --stat` lists only the three files of case 7 and this task file.
9. `gofmt -l W/pkg W/cmd` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/stack.go` (Edit 1)
- `pkg/engine/events.go` (Edit 2)
- `pkg/view/view.go` (Edit 3)

## Tests
- None added or edited: a pure move. Every existing test passes unedited.

## Done When
- [ ] Cases 1-9 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h.3` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-20, lead-architect (wave H.3 run, session 2): written against 674f8ef for #178; T-0291 is `done`, so #177 is `applied`.
  - `main` (54c8db4) has nothing the run branch lacks, so no sync. Not `WAVE-COMPLETE` before the task: #178 is `open: before the next Deliverables task`, and #179 is `open`.
  - This is a harness task, so it skips `tests` and goes straight to `ready`. Tier is `flash`, because the edits are exact and were dry-run.
  - `push`'s place event is a third copy of the lookup that #178 names. It is moved here, under #178's title, not split into a new directive.
  - Every command here is on the headless allow-list: one per line, no pipes, no `$`, and patterns with special characters are quoted.
  - #179 (the view's `chosenTargets`) edits `view.go` next, and does not go into this task.
  - Waiting directives checked: #6, #7 (mode `serial`), #15 (2) (no draw or end-of-turn hook in `ability.Hook`), #71, #91 (2), #106, #120, #142 (wave H.3 has no cards). None has fired.
  - Next: `kernel-engineer` on Flash.
