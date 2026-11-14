---
id: T-0233
title: "Decisions name their source card: Decision.Source, the askFor scope, and declared placements (directive #149 (a), wave F.2 harness, flash)"
kind: harness
tier: flash
wave: F.2
status: done
branch: m2/autoloop-wave-f.2/T-0233-decision-source
depends_on: [T-0232]
claims: [field:Decision.Source, field:Game.asking, func:Game.askFor, func:Game.shownCard, func:Game.SourceCard]
cards: []
requirements: [FR-MOD-001, FR-ZON-005]
---

## Goal
Every decision asked while a Deploy, a Maneuver or an activated ability is declared names the card it comes from: `Decision.Source`, the card as both seats see it, which a client can show with its rules text (#149). `declare` opens a scope (`askFor`) that `choose` stamps on every decision without a source, and `shownCard` is the one rule for which face a source shows (the Scratch face for a face-down asset, FR-ZON-005). `SourceCard` exports it for stack items, and `view` names an ability's card through it instead of its own copy of the rule. Triggers, resolution, the decision record and the play screen come in T-0234 and T-0235.

## Inputs
- Board directive #149 (user playtest feedback, 2026-11-13), `open: before the next Deliverables task`; design note Deliverable 6k (wave F.2), first.
- Size (D2.4): #149 as a whole touches about ten non-test files (every decision site in the engine, the record, the view and two play-screen files), so it is split three ways, each with a board row and a task file in this commit: (a) this task, the mechanism and declared placements; (b) T-0234, triggers and resolution, and `DecisionRecord.Source`; (c) T-0235, the play screen.
- Design (lead decision): a scope, not a field on every decision literal. `declare` (`placement.go:48`) is the one path for a Deploy, a Maneuver and an activated ability (`actions.go:77-84`), and its decisions are built in five files (cost discards in `zones.go` via `costs.go`, cost picks in `costs.go`, modes in `placement.go`, targets in `targeting.go`). One scope in `declare` reaches all of them, and T-0234 adds one per trigger and one per resolution. `choose` fills `Source` only when the decision has none, so a site may still set its own.
- Reuse search (D2.3): the "face-down shows the Scratch face" rule for a stack item's card was written in `view.stackView` (`view.go:201-207`). This task writes it once, in `shownCard`, and the view calls `SourceCard`, so no second copy is kept. `Asset.Printed()` (`state.go:85`) is the face rule for an asset; `shownCard` uses it.
- Prompts don't change, and `DecisionRecord` doesn't change in this task, so the sim's text output is byte-identical (`sim-compare` compares the text output of `cardtown sim`, `Taskfile.yml:74-93`).
- The lead dry-ran these exact edits and this exact test file in a throwaway worktree of the run branch at bcb41fa: the tests fail to build before the edits (`d.Source undefined`) and pass after; `gofmt -l pkg` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-f.2` `[PASS]`, `cardtown lint fences` exit 0. Every expected output below is from that dry run.

## Commits
1. `test(T-0233): decisions name their source card` with only `pkg/engine/decision_source_test.go` (below). Before committing, run `go -C W test ./pkg/engine -run TestDecisionSource` and record that it fails to build (`d.Source undefined`).
2. `feat(engine): Decision.Source, from the askFor scope that declare opens (T-0233)` with Edits 1-5.

## Tests (lead decisions; copy exactly as the new file `pkg/engine/decision_source_test.go`, tabs for indentation)
```go
package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0233 (directive #149 (a), FR-MOD-001, FR-ZON-005): every
// decision asked while a Deploy, a Maneuver or an activated ability is declared
// names the card it comes from (Decision.Source), so a client can show that
// card's rules text, and SourceCard names a stack item's card as both seats see
// it. Cards are synthetic; player 0 declares.

// wantSources reports every decision in ds whose Source is not want, and an
// empty ds.
func wantSources(t *testing.T, ds []Decision, want *Card) {
	t.Helper()
	if len(ds) == 0 {
		t.Fatalf("no decision was asked")
	}
	for i, d := range ds {
		if d.Source != want {
			t.Errorf("decision %d (%s, %q): Source %v, want %s", i, d.Kind, d.Prompt, d.Source, want.Name)
		}
	}
}

func TestDecisionSource_ManeuverTargets(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("X", 1, 9, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Y", 1, 9, Keywords{}))
	c := maneuver("Ambush", 0, 0, anotherEffect())
	c.RulesText = "Chosen unit deals damage equal to its power to another chosen unit."
	inst := s.give(0, c)
	s.script(0, Answer{Picks: []int{0}})
	s.do(enactAction(inst))
	wantSources(t, s.asked(0), c)
	if got := s.asked(0)[0].Source.RulesText; got != c.RulesText {
		t.Errorf("Source.RulesText %q, want %q", got, c.RulesText)
	}
	if s.g.asking != nil {
		t.Errorf("after the declaration the scope is still open: asking %s, want nil", s.g.asking.Name)
	}
}

func TestDecisionSource_ManeuverModes(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("X", 1, 9, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Y", 1, 9, Keywords{}))
	c := maneuver("Pick One", 0, 0, ability.Effect{Modes: []ability.Effect{returnMode(), dealerMode()}, ChooseN: 1})
	inst := s.give(0, c)
	s.script(0, Answer{Picks: []int{0}}, Answer{Picks: []int{0}})
	s.do(enactAction(inst))
	wantSources(t, s.asked(0), c)
	if s.asked(0)[0].Kind != DecideMode {
		t.Errorf("first decision %s, want Mode", s.asked(0)[0].Kind)
	}
}

func TestDecisionSource_ActivatedCostPicks(t *testing.T) {
	s := newScene(t)
	costs := []ability.Cost{{Kind: ability.CostDiscard, UpTo: 1}, {Kind: ability.CostKill, Filter: alliedUnitFilter}}
	a := s.put(0, ZoneMainTheater, actUnit("Engine", nil, actAbility(costs, drawN(1))))
	s.put(0, ZoneMainTheater, unit("Fodder", 1, 1, Keywords{}))
	s.give(0, unit("Spare", 1, 1, Keywords{}))
	s.script(0, Answer{Picks: []int{}}, Answer{Picks: []int{1}})
	s.do(activate(a, 0))
	wantSources(t, s.asked(0), a.Inst.Card)
	if len(s.asked(0)) != 2 || s.asked(0)[0].Kind != DecideHand || s.asked(0)[1].Kind != DecidePick {
		t.Errorf("asked %d decisions, want the discard cost's Hand then the kill cost's Pick", len(s.asked(0)))
	}
}

func TestDecisionSource_DeployCostPick(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, unit("A", 1, 1, Keywords{}))
	s.put(0, ZoneMainTheater, unit("B", 1, 1, Keywords{}))
	c := killCostUnit("Sacrificer", 0, 0)
	inst := s.give(0, c)
	s.script(0, Answer{Picks: []int{0}})
	s.do(dcDeployToReserves(inst))
	wantSources(t, s.asked(0), c)
}

func TestDecisionSource_ScopesNestAndAnExplicitSourceStays(t *testing.T) {
	s := newScene(t)
	outer, inner, own := unit("Outer", 1, 1, Keywords{}), unit("Inner", 1, 1, Keywords{}), unit("Own", 1, 1, Keywords{})
	ask := func(src *Card) *Card {
		s.script(0, Answer{Picks: []int{0}})
		s.g.choose(Decision{Player: 0, Kind: DecideTarget, Prompt: "pick", Min: 1, Max: 1, Source: src,
			Options: []Option{{Label: "a", Ref: 1}, {Label: "b", Ref: 2}}})
		return s.asked(0)[0].Source
	}
	if got := ask(nil); got != nil {
		t.Errorf("outside any scope: Source %v, want nil", got)
	}
	closeOuter := s.g.askFor(outer)
	if got := ask(nil); got != outer {
		t.Errorf("in the outer scope: Source %v, want Outer", got)
	}
	closeInner := s.g.askFor(inner)
	if got := ask(nil); got != inner {
		t.Errorf("in the inner scope: Source %v, want Inner", got)
	}
	if got := ask(own); got != own {
		t.Errorf("an explicit Source in a scope: Source %v, want Own", got)
	}
	closeInner()
	if got := ask(nil); got != outer {
		t.Errorf("after the inner scope closed: Source %v, want Outer", got)
	}
	closeOuter()
	if got := ask(nil); got != nil {
		t.Errorf("after both closed: Source %v, want nil", got)
	}
}

func TestSourceCard(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, unit("Hidden", 1, 1, Keywords{}))
	it := &StackItem{Kind: StackAbility, Controller: 0, Source: a.ID, Inst: a.Inst}
	if got := s.g.SourceCard(it); got != a.Inst.Card {
		t.Errorf("a face-up source asset: %v, want its card", got)
	}
	a.FaceDown = true
	if got := s.g.SourceCard(it); got != scratchFace {
		t.Errorf("a face-down source asset: %v, want the Scratch face (FR-ZON-005)", got)
	}
	gone := s.inDiscard(0, unit("Gone", 1, 1, Keywords{}))[0]
	if got := s.g.SourceCard(&StackItem{Kind: StackAbility, Source: 999, Inst: gone}); got != gone.Card {
		t.Errorf("a source asset that has left the battlefield: %v, want its card instance's card", got)
	}
	m := s.give(0, maneuver("Played", 0, 0, anotherEffect()))
	if got := s.g.SourceCard(&StackItem{Kind: StackManeuver, Inst: m}); got != m.Card {
		t.Errorf("a Maneuver: %v, want its own card", got)
	}
	if got := s.g.SourceCard(&StackItem{Kind: StackAbility}); got != nil {
		t.Errorf("no source asset and no instance: %v, want nil", got)
	}
}
```

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at bcb41fa.

### 1. `pkg/engine/choice.go`: the `Source` field
In `type Decision struct`, after line 44 (`	Repeats  bool // the same option may be picked more than once (FR-MOD-001 §5)`) and before the closing `}`, insert:
```go
	// Source is the card whose Deploy, Maneuver or ability is being placed or
	// resolved as d is asked, as both seats see it (FR-ZON-005: the Scratch
	// face for a face-down asset), so a client can show its rules text; nil for
	// a decision outside any, such as the mulligan or the trigger order. choose
	// fills it in from the scope askFor opened.
	Source *Card
```

### 2. `pkg/engine/choice.go`: `askFor`, and `choose` stamps the scope
Right before the comment `// choose returns the picks for d (indexes into d.Options). It takes the` (line 97), insert this block followed by one blank line:
```go
// askFor opens a scope in which every decision choose asks names c as its
// Source, and returns the func that closes it, restoring the source before it.
// Scopes nest: a free deploy declared while an ability resolves names the
// deployed card until its declaration ends. A nil c opens a scope with no source.
func (g *Game) askFor(c *Card) (restore func()) {
	prev := g.asking
	g.asking = c
	return func() { g.asking = prev }
}
```
Then make these the first lines of `choose`'s body (after `func (g *Game) choose(d Decision) []int {`, before `	if !d.satisfiable() {`):
```go
	if d.Source == nil {
		d.Source = g.asking
	}
```

### 3. `pkg/engine/state.go`: the scope's field
In `type Game struct`, after line 146 (`	pending []pendingTrigger // triggers fired while an item resolves, awaiting the flush (FR-TRG-004)`) and its following blank line, insert this line followed by one blank line:
```go
	asking *Card // the source card the decisions asked now name (Decision.Source, askFor); nil outside any
```
So the struct reads `pending …`, blank, `asking …`, blank, `// firings counts …`.

### 4. `pkg/engine/placement.go`: `declare` opens the scope; `shownCard` and `SourceCard`
Make this the first line of `declare`'s body (line 49, before `	pl.discards = g.chooseCostDiscards(pl)`):
```go
	defer g.askFor(g.shownCard(pl.source, pl.inst))()
```
After `declare`'s closing `}` (line 54) and its blank line, before `// maneuverPlacement is the placement of the Maneuver inst that pid enacts, for`, insert this block followed by one blank line:
```go
// shownCard is the card that a placement or a stack item with source asset ID
// source and card instance inst shows as its source: the source asset's printed
// face while it is on the battlefield (the Scratch face if face-down,
// FR-ZON-005), else inst's card (a card being deployed or played, or a card in
// a discard pile), else nil. Decision.Source and SourceCard both read it.
func (g *Game) shownCard(source int, inst *Instance) *Card {
	if a := g.Asset(source); source != 0 && a != nil {
		return a.Printed()
	}
	if inst != nil {
		return inst.Card
	}
	return nil
}

// SourceCard is the card stack item it shows as its source, as both seats see
// it (shownCard): its Source asset's printed face while that asset is on the
// battlefield, else its Inst's card (a Deploy's or a Maneuver's own card, or an
// ability's card that has left the battlefield), else nil.
func (g *Game) SourceCard(it *StackItem) *Card {
	return g.shownCard(it.Source, it.Inst)
}
```

### 5. `pkg/view/view.go`: the view names an ability's card through `SourceCard`
In `stackView`, under `case engine.StackAbility:`, lines 201-207, which read
```go
		// A face-down asset has no abilities today; if one ever has, its
		// item still names only the Scratch face.
		if a := g.Asset(it.Source); a != nil && a.FaceDown {
			s.Card = a.Printed().Name
		} else if it.Inst != nil {
			s.Card = it.Inst.Card.Name
		}
```
become
```go
		// A face-down source names only the Scratch face (Game.SourceCard).
		if c := g.SourceCard(it); c != nil {
			s.Card = c.Name
		}
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The tests failed first:** at the `test(T-0233)` commit, before Edit 1, `go -C W test ./pkg/engine -run TestDecisionSource` fails to build with `d.Source undefined`.
2. **Declared placements name their card** (#149, engine side): `go -C W test ./pkg/engine -run TestDecisionSource_` passes: a Maneuver's target and mode decisions, an activated ability's discard-cost and kill-cost picks, and a Deploy's kill-cost pick each carry their card as `Source`, with its rules text, and the scope is closed after the declaration.
3. **Scopes nest, and a decision's own source stays:** `TestDecisionSource_ScopesNestAndAnExplicitSourceStays` passes (in case 2's run).
4. **One face rule** (FR-ZON-005): `go -C W test ./pkg/engine -run TestSourceCard` passes: a face-up source asset gives its card, a face-down one the Scratch face, a source gone from the battlefield its instance's card, a Maneuver its own card, and an item with neither nil.
5. **The view keeps no copy:** `git -C W grep -n -F "g.Asset(it.Source)" -- pkg/view/view.go` prints nothing, and `git -C W grep -n -F "g.SourceCard(it)" -- pkg/view/view.go` prints exactly one line, `pkg/view/view.go:202:`.
6. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-f.2` prints `[PASS] starter and stress sims match autoloop/m2-wave-f.2 byte for byte, exit codes included`.
7. **Only the listed edits:** `git -C W diff autoloop/m2-wave-f.2 --numstat -- pkg cmd` prints exactly five lines: `19	0	pkg/engine/choice.go`, `141	0	pkg/engine/decision_source_test.go`, `24	0	pkg/engine/placement.go`, `2	0	pkg/engine/state.go` and `3	6	pkg/view/view.go`.
8. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
9. **Nothing else changed:** `git -C W diff autoloop/m2-wave-f.2 --stat` lists only the five files of case 7 and this task file.
10. `gofmt -l W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/choice.go`
- `pkg/engine/state.go`
- `pkg/engine/placement.go`
- `pkg/view/view.go`
- `pkg/engine/decision_source_test.go` (new; the `test(T-0233)` commit only)

## Done When
- [x] Cases 1-10 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-f.2` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-14, lead-architect (wave F.2 run, session 1): written for #149 (a) against bcb41fa (the run branch's base; nothing on `main` to sync). Harness, so it skips `tests` and goes straight to `ready`. Tier `flash`, because every edit and the whole test file are exact and were dry-run. #149 is split under D2.4: T-0234 (triggers and resolution, `DecisionRecord.Source`) and T-0235 (the play screen) are `queued`, with board rows and task files in this commit. Every command here is on the headless allow-list: one per line, no pipes, no `$`, no unquoted parentheses. Waiting directives checked: #6, #7 (mode `serial`), #15 (no turn-boundary hook), #71, #91 (2), #106, #120, #141 (3), #142: not fired. #152's trigger fired (its outcome reads `Status: applied: FRD 01 v3.25`), so it is `open`, after #151 as its row says. Next: `kernel-engineer` on Flash.
- 2026-11-14, kernel-engineer (T-0233): implemented directive #149 (a).
  - Status moved to `review`.
  - Added `pkg/engine/decision_source_test.go` and verified Acceptance Case 1 (`go -C W test ./pkg/engine -run TestDecisionSource` fails to build with `d.Source undefined`) prior to committing `test(T-0233)`.
  - Implemented Edits 1-5:
    - Added `Source *Card` field to `Decision` in `pkg/engine/choice.go`.
    - Added `askFor` scope manager and stamped `d.Source = g.asking` when nil in `g.choose` (`pkg/engine/choice.go`).
    - Added `asking *Card` field to `Game` struct (`pkg/engine/state.go`).
    - Added `defer g.askFor(g.shownCard(pl.source, pl.inst))()` in `g.declare`, and added `shownCard` and `SourceCard` methods (`pkg/engine/placement.go`).
    - Updated `stackView` in `pkg/view/view.go` to use `g.SourceCard(it)` instead of duplicating face-down Scratch logic.
  - Verification results:
    - Case 1: `go -C W test ./pkg/engine -run TestDecisionSource` failed to build with `d.Source undefined` before edits.
    - Case 2 & 3: `go -C W test ./pkg/engine -run TestDecisionSource_` passed (1.050s).
    - Case 4: `go -C W test ./pkg/engine -run TestSourceCard` passed (1.049s).
    - Case 5: `git -C W grep -n -F "g.Asset(it.Source)" -- pkg/view/view.go` returned no matches (exit code 1); `git -C W grep -n -F "g.SourceCard(it)" -- pkg/view/view.go` returned `pkg/view/view.go:202:` (exit code 0).
    - Case 6: `go -C W tool task sim-compare -- autoloop/m2-wave-f.2` returned `[PASS] starter and stress sims match autoloop/m2-wave-f.2 byte for byte, exit codes included` (exit code 0).
    - Case 7: `git -C W diff autoloop/m2-wave-f.2 --numstat -- pkg cmd` printed exactly:
      19	0	pkg/engine/choice.go
      141	0	pkg/engine/decision_source_test.go
      24	0	pkg/engine/placement.go
      2	0	pkg/engine/state.go
      3	6	pkg/view/view.go
    - Case 8: `go -C W run ./cmd/cardtown lint fences` passed with `[PASS] Package code is free of card names, invalid markers and disallowed imports.` (exit code 0).
    - Case 9: `git -C W diff autoloop/m2-wave-f.2 --stat` showed only the five files and the task file.
    - Case 10: `gofmt -l W/pkg` printed nothing; `go -C W tool task check` passed (all tests, vet, fmt, and all linters passed); `go -C W tool task sim` passed (200 starter playouts, all invariants holding).
- 2026-11-14, lead-architect (wave F.2 run, session 1, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-f.2` as f99b291. Flash run `e0cd7b68-c1d3-4890-8ad5-fa46a19dda6a`: exit 0, `status: SUCCESS`, no `denied_actions` key, no `permission check failed` in the transcript; `.err` holds only agy's "root agent idle" notice. Review of `git diff autoloop/m2-wave-f.2...m2/autoloop-wave-f.2/T-0233-decision-source` (09fe78d):
  - Scope: the four Allowed non-test files, the new test file and this file. The `test(T-0233)` commit (d1c0954) holds only the test file, which is identical to the one in this task; the implementation commit (bb9cfc0) holds only the four non-test files. No earlier test changed.
  - Checked against the diff: the `pkg/` diff matches the lead's dry run (numstat `19 0`, `141 0`, `24 0`, `2 0`, `3 6`). The tests failed first: re-run at d1c0954 by the reviewer, `d.Source undefined`.
  - Fences and placement: no card name; `askFor` and the `Source` stamp sit in `choice.go` with `choose`, `shownCard` and `SourceCard` in `placement.go` beside `declare`. Reuse: the view's own face-down rule is gone (case 5), so "which face a source shows" is written once, in `shownCard`.
  - Rules fidelity: prompts, picks and fallbacks are unchanged; `Source` follows FR-ZON-005 (`Asset.Printed()`), so a face-down source shows only the Scratch face.
  - Re-run by the reviewer in a scratch worktree of the task branch: `gofmt -l pkg` clean; `cardtown lint fences` `[PASS]`; `go tool task check` exit 0; `go tool task sim-compare -- autoloop/m2-wave-f.2` `[PASS]`; `go tool task sim` exit 0, 200 games, P0 63.5% / P1 35.0% / draws 1.5%, P0 50/50 and P1 50/50 main cards executable.
  - The merge's tree matches the task tip's (`git diff` between them is empty), so the gates hold on the run branch. #149 stays `open`: T-0234 and T-0235 are its other parts.
