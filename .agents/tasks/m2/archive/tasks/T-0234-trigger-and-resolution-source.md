---
id: T-0234
title: "Triggers and resolutions name their source card, and the decision record carries it (directive #149 (b), wave F.2 harness, flash)"
kind: harness
tier: flash
wave: F.2
status: done
branch: m2/autoloop-wave-f.2/T-0234-trigger-and-resolution-source
depends_on: [T-0233]
claims: [func:Game.placeTrigger, field:DecisionRecord.Source]
cards: []
requirements: [FR-MOD-001, FR-TRG-003, FR-ZON-005]
---

## Goal
Every decision asked while a trigger is placed (its "you may", its modes, its targets), or while an ability or a Maneuver resolves (untargeted picks, a discard step's hand picks, the deck search, the look, the free-deploy zone choice), names its source card through T-0233's `askFor` scope. The trigger order names none, even when a flush runs inside a scope. `DecisionRecord` carries the source's name, so the `-events` output and a replay show and compare it (#149 (b)). The play screen comes in T-0235.

## Inputs
- Board directive #149 (user playtest feedback, 2026-11-13), `open: before the next Deliverables task`; part (a) applied by T-0233 (merge f99b291: `Decision.Source`, `Game.askFor`, `Game.shownCard`, `Game.SourceCard`). This task is part (b); T-0235 (the play screen) is (c) and stays `queued`.
- Design (lead decisions):
  - **Trigger placement.** The body of `flushTriggers`' placing loop moves, unchanged, into a new `placeTrigger(p)` that opens one `askFor(g.shownCard(p.source, p.inst))` scope, so the "you may" (`accepts`) and every mode and target decision of `chooseOnPlacement` name the trigger's card: the source asset's printed face while it is on the battlefield, else the card instance's card (a trigger from a discard pile, or a killed source). This is the face rule T-0233 wrote once, not a copy.
  - **The flush itself** runs in an `askFor(nil)` scope, so the trigger order names no card.
  - **Resolution.** `runSteps` (`effects.go`) is shared by a resolving ability, a resolving Maneuver (`resolveManeuver` calls `resolveAbility`) and an inline replacement's steps (`replace.go`'s `replaceDeath`). It opens one `askFor(g.SourceCard(r.it))` scope, which covers every step's decisions, so no step site sets a source.
  - **The replacement choice** (`chooseReplacement`, `replace.go:135`) is left as it is. Asked inside a kill step's window, it names the resolving card whose kill is being replaced; asked in the death sweep after an item resolves, it names none. That is what `Decision.Source` already says (the card being resolved as the decision is asked), and a replacement's options name their own sources. So `replace.go` doesn't change.
  - **The mulligan** is a `Config.Mulligan` callback, not a `Decision` (`setup.go:19-21`), so nothing about it changes.
  - **The record.** `DecisionRecord.Source string` (`json:"source,omitempty"`) is the source's `Name`, filled in `emitDecision`, and `eventlog`'s `decisionEqual` compares it, as it compares every other field. A replay of a log written before this task would diverge on its first decision with a source; replays are of logs from the same build (`TestReplay_Case5_BuildMismatch`), so that is expected.
- Size (D2.4): four non-test files: `trigger.go`, `effects.go`, `events.go`, `pkg/eventlog/replay.go`.
- **Expected output difference:** `cardtown sim -events FILE` (JSON Lines) gains `"source":"<card name>"` inside every decision record asked in a scope, for example `{"kind":"Look",...,"hidden":true,"source":"Devilhorn Scouting Party"}`. A decision without a source has no `source` key in its record (the event's own integer `source` field is unrelated and unchanged). `sim-compare` compares only the text output of `cardtown sim` (`Taskfile.yml:74-93`), and prompts, picks and the `Detail` text don't change, so it still passes. No golden event file exists.
- Later work: directive #152 (1) will change the "you may" check, which after this task sits in `placeTrigger`, not in `flushTriggers`' loop; its task is written against the code of its day.
- The lead dry-ran these exact edits and these exact test files in a throwaway worktree of the run branch at 5dfa8e4. Before the edits the tests fail to build (`order.Source undefined`). With only Edit 3 applied (the record), every new engine test except `TestDecisionSource_TriggerOrderNamesNone` fails; that one pins the unchanged case. Without Edit 4, `TestReplay_DecisionWithAnotherSourceDiverges` fails. With every edit: `gofmt -l pkg` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-f.2` `[PASS]`, `cardtown lint fences` exit 0, `go tool task sim` exit 0. Every expected output below is from that dry run.

## Commits
1. `test(T-0234): triggers and resolutions name their source card` with only the two new test files below. Before committing, run `go -C W test ./pkg/engine -run TestDecisionSource` and record that it fails to build (`order.Source undefined`).
2. `feat(engine): triggers and resolutions open the askFor scope; DecisionRecord.Source (T-0234)` with Edits 1-4.

## Tests (lead decisions; copy exactly, tabs for indentation)
### New file `pkg/engine/decision_source_trigger_test.go`
```go
package engine

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0234 (directive #149 (b), FR-MOD-001, FR-TRG-003, FR-ZON-005):
// every decision asked while a trigger is placed, or while an ability or a
// Maneuver resolves, names its source card (Decision.Source), the trigger
// order names none, and the decision record carries the source's name.

// wantSourceNamed reports every decision in ds whose Source is not a card
// named name, and an empty ds.
func wantSourceNamed(t *testing.T, ds []Decision, name string) {
	t.Helper()
	if len(ds) == 0 {
		t.Fatalf("no decision was asked")
	}
	for i, d := range ds {
		if d.Source == nil || d.Source.Name != name {
			t.Errorf("decision %d (%s, %q): Source %v, want %s", i, d.Kind, d.Prompt, d.Source, name)
		}
	}
}

// answerFirst answers every decision with its first option, a "you may" with
// yes, and a decision that may pick none with a timeout.
func answerFirst(d Decision) Answer {
	if d.Min == 0 && d.Kind != DecideOptional {
		return Answer{Timeout: true}
	}
	picks := make([]int, max(d.Min, 1))
	for i := range picks {
		picks[i] = i
	}
	return Answer{Picks: picks}
}

func TestDecisionSource_TriggerTarget_DevilhornCaptain(t *testing.T) {
	s := newScene(t)
	b := s.captainBoard()
	hitter := s.put(0, ZoneMainTheater, unit("Hitter", 1, 9, Keywords{}))
	s.byKind(1, func(d Decision) Answer {
		for i, o := range d.Options {
			if d.Kind == DecideTarget && o.Ref == b.u0.ID {
				return Answer{Picks: []int{i}}
			}
		}
		return answerFirst(d)
	})
	s.do(s.attack(hitter, b.cap))
	asked := s.asked(1)
	wantSources(t, asked, b.cap.Inst.Card)
	if len(askedKind(s, 1, DecideTarget)) != 1 {
		t.Errorf("Unbowed asks one target decision: kinds %v", kindsOf(asked))
	}
	if src := asked[len(asked)-1].Source; src != nil && !strings.Contains(src.RulesText, "Unbowed: ") {
		t.Errorf("Source.RulesText %q, want the Captain's printed text", src.RulesText)
	}
}

func TestDecisionSource_OptionalTargetAndPick_InfestedApostle(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 0, 5, Keywords{}))
	v := s.put(1, ZoneMainTheater, unit("Victim One", 0, 5, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Victim Two", 0, 5, Keywords{}))
	s.scriptPlan(0, nil, []Answer{yes}, pick(v), pick(ally))
	inst := deployApostle(s)
	asks := s.asked(0)
	if got, want := kindsOf(asks), []DecisionKind{DecideOptional, DecideTarget, DecidePick}; !slices.Equal(got, want) {
		t.Fatalf("kinds %v, want %v", got, want)
	}
	wantSources(t, asks, inst.Card)
}

func TestDecisionSource_ModalTrigger(t *testing.T) {
	b := newModalBoard(t)
	c := trigUnit("Modal", hitArmorTrigger())
	b.s.byKind(0, answerFirst)
	b.s.deployFromHand(0, c)
	asks := b.s.asked(0)
	wantSources(t, asks, c)
	if len(asks) == 0 || asks[0].Kind != DecideMode {
		t.Errorf("first decision %v, want the mode choice", kindsOf(asks))
	}
}

func TestDecisionSource_PileTriggerAndResolvingDiscard_EerieRemnant(t *testing.T) {
	s, orders, rem, _ := remnantOrdersScene(t, yes)
	hand := askedKind(s, 0, DecideHand)
	opt := askedKind(s, 0, DecideOptional)
	if len(hand) != 1 || len(opt) != 1 {
		t.Fatalf("asked %v, want one hand discard and one optional question", kindsOf(s.asked(0)))
	}
	wantSources(t, hand, orders.Card)
	wantSources(t, opt, rem.Card)
}

func TestDecisionSource_ResolvingDeckSearch(t *testing.T) {
	s, a, _ := cliffScene(t, 3)
	s.script(0, Answer{Picks: []int{1}})
	s.do(activate(a, 0))
	wantSources(t, s.asked(0), a.Inst.Card)
	if len(askedKind(s, 0, DecideSearch)) != 1 {
		t.Errorf("asked %v, want one deck search", kindsOf(s.asked(0)))
	}
}

func TestDecisionSource_ResolvingLook(t *testing.T) {
	s, _, man := sctScene(t, sctReal("X", "X", "X"), "", nil)
	if len(askedKind(s, 0, DecideLook)) != 1 {
		t.Fatalf("asked %v, want one look", kindsOf(s.asked(0)))
	}
	wantSources(t, s.asked(0), man.Card)
}

func TestDecisionSource_FreeDeployZone(t *testing.T) {
	s, _, _ := rrGuerrillas(t, Answer{Timeout: true})
	wantSourceNamed(t, rrZoneAsks(s), "Reaver's Rush")
}

func TestDecisionSource_TriggerOrderNamesNone(t *testing.T) {
	s := newScene(t)
	s.onTop(0, vanilla("A", "B")...)
	s.script(0, Answer{Picks: []int{0, 1}})
	s.deployFromHand(0, trigUnit("Source", deployTrigger(drawN(1)), deployTrigger(scratchN(1))))
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecideOrder {
		t.Fatalf("asked %v, want one trigger order", kindsOf(asks))
	}
	if asks[0].Source != nil {
		t.Errorf("the trigger order: Source %v, want nil", asks[0].Source)
	}
}

func TestDecisionSource_FlushInsideAScopeOrdersWithNone(t *testing.T) {
	s := newScene(t)
	src := s.put(0, ZoneReserves, trigUnit("Source", deployTrigger(drawN(1)), deployTrigger(scratchN(1))))
	outer := unit("Outer", 1, 1, Keywords{})
	s.g.fireHook(assetSource(src), ability.HookDeploy)
	s.script(0, Answer{Picks: []int{0, 1}})
	closeOuter := s.g.askFor(outer)
	s.g.flushTriggers()
	if s.g.asking != outer {
		t.Errorf("after the flush the outer scope is not restored: asking %v, want Outer", s.g.asking)
	}
	closeOuter()
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecideOrder || asks[0].Source != nil {
		t.Errorf("asked %v, want one trigger order with no Source", kindsOf(asks))
	}
}

func TestDecisionSource_Record(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("Foe", 0, 5, Keywords{}))
	s.onTop(0, vanilla("A", "B")...)
	s.script(0, Answer{Picks: []int{0, 1}})
	hit := ability.Trigger{Hook: ability.HookDeploy, Effect: hitMode()}
	s.deployFromHand(0, trigUnit("Hitter", deployTrigger(drawN(1)), hit))
	var order, target *DecisionRecord
	for _, e := range eventsOfKind(s.g.Log, EventDecision) {
		switch e.Decision.Kind {
		case DecideOrder.String():
			order = e.Decision
		case DecideTarget.String():
			target = e.Decision
		}
	}
	if order == nil || target == nil {
		t.Fatalf("want a trigger-order record and a target record")
	}
	if order.Source != "" {
		t.Errorf("the trigger order's record: source %q, want none", order.Source)
	}
	if target.Source != "Hitter" || target.How != "auto" {
		t.Errorf("the sole target's record: source %q how %q, want Hitter and auto", target.Source, target.How)
	}
	for _, tc := range []struct {
		rec  DecisionRecord
		want bool
	}{{*order, false}, {*target, true}} {
		js, err := json.Marshal(tc.rec)
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.Contains(string(js), `"source":`); got != tc.want {
			t.Errorf("%s: has \"source\" %t, want %t: %s", tc.rec.Kind, got, tc.want, js)
		}
	}
}
```

### New file `pkg/eventlog/replay_source_test.go`
```go
package eventlog_test

import (
	"strings"
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/eventlog"
)

// T-0234 (directive #149 (b)): a decision's source card is part of the record
// a replay compares.
func TestReplay_DecisionWithAnotherSourceDiverges(t *testing.T) {
	for _, evs := range searchGames(t) {
		for i, e := range evs {
			if e.Decision == nil || e.Decision.Source == "" {
				continue
			}
			cp := append([]engine.Event{}, evs...)
			decCopy := *e.Decision
			decCopy.Source = "Another Card"
			cp[i].Decision = &decCopy
			_, err := eventlog.Replay(cp, len(cp), searchResolve)
			if err == nil || !strings.Contains(err.Error(), "replay diverged") {
				t.Fatalf("decision with another source: error %v, want replay diverged", err)
			}
			return
		}
	}
	t.Fatal("no game holds a decision with a source")
}
```

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at 5dfa8e4.

### 1. `pkg/engine/trigger.go`: the flush names no source; `placeTrigger` names the trigger's
In the comment above `flushTriggers`, line 196 reads
```go
// in source order (sourcePos). A match that ended discards the buffer.
```
and becomes
```go
// in source order (sourcePos). A match that ended discards the buffer. The
// trigger order names no source card, even in a flush inside a scope askFor opened.
```
Make this the first line of `flushTriggers`' body (before `	buf := g.pending`):
```go
	defer g.askFor(nil)()
```
The placing loop, lines 223-232, which reads
```go
		for _, i := range order {
			p := mine[i]
			if p.trig.Optional && !g.accepts(p) {
				continue
			}
			item := p.stackItem(StackAbility)
			item.Effect, item.Chosen = g.chooseOnPlacement(p.placement, p.trig.Effect)
			item.Tracked, item.Keyword = p.trig.Hook == ability.HookKilled || p.trig.ActiveZone == ability.ZoneDiscard, p.trig.Keyword
			g.push(item)
		}
```
becomes
```go
		for _, i := range order {
			g.placeTrigger(mine[i])
		}
```
After `flushTriggers`' closing `}` and its blank line, before `// accepts asks p's controller the "you may" question of an optional trigger`, insert this block followed by one blank line:
```go
// placeTrigger asks p's "you may" if it is optional and, unless it is declined,
// places p on the stack with its modes and targets chosen (SVY-R9). Every
// decision it asks names p's source card (Decision.Source, shownCard).
func (g *Game) placeTrigger(p pendingTrigger) {
	defer g.askFor(g.shownCard(p.source, p.inst))()
	if p.trig.Optional && !g.accepts(p) {
		return
	}
	item := p.stackItem(StackAbility)
	item.Effect, item.Chosen = g.chooseOnPlacement(p.placement, p.trig.Effect)
	item.Tracked, item.Keyword = p.trig.Hook == ability.HookKilled || p.trig.ActiveZone == ability.ZoneDiscard, p.trig.Keyword
	g.push(item)
}
```

### 2. `pkg/engine/effects.go`: one scope for a resolution's steps
The comment above `runSteps` (lines 52-53) gains two lines, and the body gains a first line, so it reads:
```go
// runSteps runs the steps of r's item in order, recording how many things each
// did. A resolving ability and an inline replacement (replace.go) share it.
// Every decision a step asks names the item's source card (Decision.Source,
// SourceCard), a replacement choice in a kill step's window included.
func (g *Game) runSteps(r *resolution) {
	defer g.askFor(g.SourceCard(r.it))()
	for i, st := range r.it.Effect.Steps {
```
The rest of `runSteps` doesn't change.

### 3. `pkg/engine/events.go`: `DecisionRecord.Source`
In `type DecisionRecord struct`, after line 27 (the `Hidden` field), insert this line (the spacing lines it up with the fields above; `gofmt` keeps it):
```go
	Source  string   `json:"source,omitempty"` // Decision.Source's name; "" for a decision with no source
```
In `emitDecision`, after
```go
	if rec.Picks == nil {
		rec.Picks = []int{}
	}
```
insert:
```go
	if d.Source != nil {
		rec.Source = d.Source.Name
	}
```

### 4. `pkg/eventlog/replay.go`: a replay compares the source
In `decisionEqual` (line 250), the first line of its `return`,
```go
	return a.Kind == b.Kind && a.Player == b.Player && a.Prompt == b.Prompt && a.How == b.How && a.Hidden == b.Hidden &&
```
becomes
```go
	return a.Kind == b.Kind && a.Player == b.Player && a.Prompt == b.Prompt && a.How == b.How && a.Hidden == b.Hidden && a.Source == b.Source &&
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The tests failed first:** at the `test(T-0234)` commit, before Edit 1, `go -C W test ./pkg/engine -run TestDecisionSource` fails to build with `order.Source undefined`.
2. **Trigger placement names the trigger's card** (#149 (a), (b)): `go -C W test ./pkg/engine -run TestDecisionSource_` passes. In that run: *Devilhorn Captain*'s Unbowed target decision carries the Captain, with its rules text (`TestDecisionSource_TriggerTarget_DevilhornCaptain`); *Infested Apostle*'s "you may", its target and its resolving kill pick carry the Apostle (`TestDecisionSource_OptionalTargetAndPick_InfestedApostle`); a modal trigger's mode choice and target carry its card (`TestDecisionSource_ModalTrigger`).
3. **A trigger from a discard pile** (#149 (c)): `TestDecisionSource_PileTriggerAndResolvingDiscard_EerieRemnant` passes (in case 2's run): *Eerie Remnant*'s "you may" carries the Remnant, and *Improvised Orders*' resolving discard carries the Orders.
4. **Resolution names the resolving card:** `TestDecisionSource_ResolvingDeckSearch` (*Cliff of Portents*), `TestDecisionSource_ResolvingLook` (*Devilhorn Scouting Party*) and `TestDecisionSource_FreeDeployZone` (*Reaver's Rush*) pass (in case 2's run).
5. **The trigger order names none** (#149 (d)): `TestDecisionSource_TriggerOrderNamesNone` and `TestDecisionSource_FlushInsideAScopeOrdersWithNone` pass (in case 2's run); the second also checks that the flush restores the scope it ran in.
6. **The record:** `TestDecisionSource_Record` passes (in case 2's run): an auto-picked target's record has source "Hitter", the trigger order's has none, and only the first has a `"source":` key in its JSON. `go -C W test ./pkg/eventlog -run TestReplay_DecisionWithAnotherSourceDiverges` passes: a record with another source diverges on replay.
7. **The earlier tests still pass:** `go -C W test ./pkg/engine -run TestSourceCard` and `go -C W test ./pkg/eventlog` pass.
8. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-f.2` prints `[PASS] starter and stress sims match autoloop/m2-wave-f.2 byte for byte, exit codes included`.
9. **Only the listed edits:** `git -C W diff autoloop/m2-wave-f.2 --numstat -- pkg cmd` prints exactly six lines: `196	0	pkg/engine/decision_source_trigger_test.go`, `3	0	pkg/engine/effects.go`, `4	0	pkg/engine/events.go`, `18	9	pkg/engine/trigger.go`, `1	1	pkg/eventlog/replay.go` and `31	0	pkg/eventlog/replay_source_test.go`.
10. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
11. **Nothing else changed:** `git -C W diff autoloop/m2-wave-f.2 --stat` lists only the six files of case 9 and this task file.
12. `gofmt -l W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass (the `pkg/view/viewtest` leak tests are in `check`).

## Allowed Files
- `pkg/engine/trigger.go`
- `pkg/engine/effects.go`
- `pkg/engine/events.go`
- `pkg/eventlog/replay.go`
- `pkg/engine/decision_source_trigger_test.go` (new; the `test(T-0234)` commit only)
- `pkg/eventlog/replay_source_test.go` (new; the `test(T-0234)` commit only)

## Done When
- [x] Cases 1-12 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-f.2` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-14, lead-architect (wave F.2 run, session 1): queued stub, split off T-0233 under D2.4 (#149 as a whole is about ten non-test files). Next after T-0233; T-0235 (the play screen) follows it.
- 2026-11-14, lead-architect (wave F.2 run, session 2): written out in full against 5dfa8e4 for #149 (b). Nothing on `main` to sync (`git log autoloop/m2-wave-f.2..main` is empty). Harness, so it skips `tests` and goes straight to `ready`. Tier `flash`, because every edit and both test files are exact and were dry-run. The stub's replacement question is decided above (Inputs, "The replacement choice"): `replace.go` doesn't change. The stub listed "a new test file"; it is two, one per package, and `pkg/eventlog/replay.go` joins so that a replay compares the new field (four non-test files, within D2.4). Reuse search (D2.3): the face a trigger shows is T-0233's `shownCard`, and the resolution scope is in the one `runSteps` every resolution shares, so there is no second copy. Every command here is on the headless allow-list: one per line, no pipes, no `$`, no unquoted parentheses. Waiting directives checked: #6, #7 (mode `serial`), #15 (no turn-boundary hook), #71, #91 (2), #106, #120, #141 (3), #142: not fired. Next: `kernel-engineer` on Flash.
- 2026-11-14, kernel-engineer (wave F.2, Flash): implemented directive #149 (b). Opened scratch worktree on branch m2/autoloop-wave-f.2/T-0234-trigger-and-resolution-source; set status: in-progress (b45ac07). Reused Game.askFor, Game.shownCard and Game.SourceCard from T-0233. Added test files pkg/engine/decision_source_trigger_test.go and pkg/eventlog/replay_source_test.go; verified Acceptance Case 1 failure (order.Source undefined), committed as `test(T-0234): triggers and resolutions name their source card` (72129bb). Implemented Edits 1-4: flushTriggers defers askFor(nil) and delegates placement to placeTrigger(p) which opens askFor(g.shownCard(p.source, p.inst)); runSteps opens askFor(g.SourceCard(r.it)); DecisionRecord gains Source string with json:"source,omitempty" populated in emitDecision; decisionEqual checks a.Source == b.Source. Committed as `feat(engine): triggers and resolutions open the askFor scope; DecisionRecord.Source (T-0234)` (f9575be). Verified Acceptance Cases 1-12: Case 1 pre-edit build failed as expected; Cases 2-6 `go test ./pkg/engine -run TestDecisionSource_` ok; Case 6 `go test ./pkg/eventlog -run TestReplay_DecisionWithAnotherSourceDiverges` ok; Case 7 `go test ./pkg/engine -run TestSourceCard` ok and `go test ./pkg/eventlog` ok; Case 8 `go tool task sim-compare -- autoloop/m2-wave-f.2` [PASS]; Case 9 `git diff autoloop/m2-wave-f.2 --numstat -- pkg cmd` exactly 6 lines with identical counts; Case 10 `cardtown lint fences` [PASS] exit 0; Case 11 `git diff autoloop/m2-wave-f.2 --stat` 7 files (6 code/test + task file); Case 12 `gofmt -l pkg` clean, `go tool task check` pass (exit 0), `go tool task sim` pass (200 games clean, P0 63.5% / P1 35.0% / draws 1.5%). Ready for review.
- 2026-11-14, lead-architect (wave F.2 run, session 2, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-f.2` as 17297e6. Flash run `abd51b0e-dee9-43d7-9fa6-6728ddac2aea`: exit 0, `status: SUCCESS`, no `denied_actions` key. The transcript has no step with `"status":"ERROR"`, and its one `permission check failed` match is run-file text the engineer read, not a refusal. `.err` holds only agy's "root agent idle" notice. Review of `git diff autoloop/m2-wave-f.2...m2/autoloop-wave-f.2/T-0234-trigger-and-resolution-source` (4ebf5e5):
  - Scope: the four Allowed non-test files, the two new test files and this file. The `test(T-0234)` commit (72129bb) holds only the two test files, byte-identical to the lead's dry run; the implementation commit (f9575be) holds only the four non-test files, and its diff is line for line the dry run's. No earlier test changed. The tests failed first: re-run at 72129bb by the reviewer, `order.Source undefined`.
  - Fences and placement: no card name; `placeTrigger` sits in `trigger.go` beside `flushTriggers`, the resolution scope in `runSteps`, the record field in `events.go` and its comparison in `eventlog`'s `decisionEqual`. Reuse: the trigger's face is T-0233's `shownCard`, and one scope in the `runSteps` that every resolution shares covers all step decisions, so nothing is copied.
  - Rules fidelity: prompts, picks, fallbacks and the order of placement are unchanged (the loop body moved verbatim into `placeTrigger`). `Source` follows FR-ZON-005.
  - Re-run by the reviewer in a scratch worktree of the task branch: `gofmt -l pkg` clean; `cardtown lint fences` `[PASS]`; `go tool task check` exit 0; `go tool task sim-compare -- autoloop/m2-wave-f.2` `[PASS]`; `go tool task sim` exit 0, 200 games, P0 63.5% / P1 35.0% / draws 1.5%, P0 50/50 and P1 50/50 main cards executable.
  - The merge's tree matches the task tip's (`git diff` between them is empty), so the gates hold on the run branch. #149 stays `open`: T-0235 (the play screen) is its last part.
