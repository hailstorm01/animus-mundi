---
id: T-0164
title: The play screen's not-executable tag is pinned on its own, not through one seat's starter session (harness, flash, tests only)
kind: harness
tier: flash
wave: D
status: review
branch: m2/autoloop-wave-d/T-0164-play-tag-test
depends_on: [T-0037]
claims: [test:TestPlay_Case1 tag assertion, test:TestPlay_PrintScreenTagsANotExecutableHandCard, comment:TestView_Case8]
cards: []
requirements: []
---

## Goal
A test of `printScreen` pins the `[not executable yet]` tag directly, and `TestPlay_Case1_ScriptedSessionFinishes` stops asking for it. *Sinuous Growth* (T-0165) can then leave `NotYetExecutable` without moving that test. No non-test code changes: same declarations, same sim output.

## Inputs
- `cmd/cardtown/play_test.go`: `TestPlay_Case1_ScriptedSessionFinishes` (line 44) plays seat 0 with Animus at seed 3 and asserts the tag at lines 64-66. `TestPlay_PrintScreenManeuverStackLine` (line 288) is the model for a direct `printScreen` test.
- `cmd/cardtown/play_screen.go:65-73`: the tag and the hand line's format.
- `pkg/view/executable_test.go:13-15`: the comment on `TestView_Case8_ExecutableFlag`.

### Why (lead's probe, 2026-11-04)
A trial definition of *Sinuous Growth* in the sync worktree, since reverted, passes every test but this one: at seed 3, seat 0 never holds an unsupported card again. With that trial, seat 1 still shows the tag and the stress decks don't, so another seat or seed would only postpone it, and at the #14 (d) checkpoint no starter deck holds such a card. The tag is a property of the screen, so the test now asks the screen.

The same probe printed seat 0's seed-1 hand: Shifting Impossibility, Veil Render, Rapturous Sermon, Sinuous Growth, Despoiler Larva and Animoceps Spores, of which only Veil Render and Sinuous Growth are unsupported today. `TestView_Case8_ExecutableFlag`'s comment calls five of them unsupported, so it is already wrong, and T-0165 makes one more name wrong. Its assertions don't change.

## Edits (lead decisions; copy exactly)
**1. `cmd/cardtown/play_test.go`, `TestPlay_Case1_ScriptedSessionFinishes`.** Delete this line of its comment:
```go
//   - It contains [not executable yet] at least once (the starter decks hold such cards).
```
and delete these three lines of its body, with nothing else changed:
```go
	if !strings.Contains(out1, "[not executable yet]") {
		t.Errorf("output does not contain %q", "[not executable yet]")
	}
```
Leave one blank line between the result-line check and `code2, out2 := runOnce()`.

**2. `cmd/cardtown/play_test.go`.** Add this test right after `TestPlay_PrintScreenManeuverStackLine`:
```go
// T-0164: the screen tags a hand card the kernel can't run yet, and no other.
func TestPlay_PrintScreenTagsANotExecutableHandCard(t *testing.T) {
	var v view.View
	v.You.Hand = []view.CardView{
		{Card: "Ready Card", ReqCost: 1, TacCost: 2, Power: 3, Health: 4, Executable: true},
		{Card: "Stub Card", ReqCost: 5, TacCost: 6, Power: 7, Health: 8},
	}
	var buf bytes.Buffer
	printScreen(&buf, v, nil)
	out := buf.String()

	want := "Your Hand:\n  [1] Ready Card (R1 T2, 3/4)\n  [2] Stub Card (R5 T6, 7/8) [not executable yet]\n"
	if !strings.Contains(out, want) {
		t.Errorf("printScreen output did not contain expected hand section:\n%s\nGot output:\n%s", want, out)
	}
}
```

**3. `pkg/view/executable_test.go`.** Replace these three comment lines:
```go
// Seed 1 is chosen: seat 0's initial hand contains Despoiler Larva (supported)
// as well as Shifting Impossibility, Veil Render, Rapturous Sermon, Sinuous Growth,
// and Animoceps Spores (unsupported).
```
with these two:
```go
// Seed 1 is chosen: seat 0's initial hand contains Despoiler Larva (supported)
// as well as Veil Render (unsupported).
```

## Acceptance Cases (harness tasks only; written by the lead)
These edits are the whole change. The new test pins behavior that exists since T-0037, so it passes as soon as it is written, and there is no implementation commit. Commit the edits once, as `test(T-0164): pin the not-executable tag on the screen, not in a session`, then hand off with the usual status commit. W is your worktree. Run each command on its own, pointed at W, and record each command and its result in the Log. The reviewer re-runs them.
1. **The new test and the edited one pass:** `go -C W test ./cmd/cardtown -run TestPlay_ -v` exits 0 and shows both `TestPlay_PrintScreenTagsANotExecutableHandCard` and `TestPlay_Case1_ScriptedSessionFinishes` as `PASS`.
2. **Case 1 no longer names the tag, and the new test does:** `git -C W grep -n -F 'not executable yet' -- cmd/cardtown/play_test.go` prints exactly one line, the new test's `want`.
3. **The view comment:** `git -C W grep -n -e 'Shifting Impossibility' -e 'Sinuous Growth' -- pkg/view/executable_test.go` prints nothing (git exits 1), and `go -C W test ./pkg/view -run TestView_Case8_ExecutableFlag` exits 0.
4. **Only tests changed:** `git -C W diff autoloop/m2-wave-d --stat` lists only `cmd/cardtown/play_test.go`, `pkg/view/executable_test.go` and this task file.
5. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-d` prints `[PASS]`.
6. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `cmd/cardtown/play_test.go`
- `pkg/view/executable_test.go`

## Done When
- [x] Cases 1-6 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-d` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-04 lead-architect (wave D run, session 8): no `main` commits to sync (`git log autoloop/m2-wave-d..main` is empty). Not `WAVE-COMPLETE`: all four header cards are on `NotYetExecutable`. #105 is `open`, and its part (1) is applied, the part that goes before the card tasks; parts (2) and (3) go with *Cliff of Portents*. Waiting directives checked, none fired: #6, #7 (mode is `serial`); #8 (no task touches `deck.go`); #15 (no turn-boundary hook); #71, #91 (2) (no such card). The next card is *Sinuous Growth*, a deck card first on the run file's `Cards` line, and the probe above found that it breaks `TestPlay_Case1_ScriptedSessionFinishes`. `rules-test-author` writes only under `pkg/engine`, so this edit is a harness task of its own, ahead of the card, and T-0165 is `queued` behind it (D2.4). Status `ready`.
- 2026-11-04 kernel-engineer: edits committed in test(T-0164) ff1ade5 (`cmd/cardtown/play_test.go` and `pkg/view/executable_test.go`). No engine shape or card code added. Cases 1-6 verified in W:
  - Case 1: `go test ./cmd/cardtown -run TestPlay_ -v` exited 0; both `TestPlay_PrintScreenTagsANotExecutableHandCard` and `TestPlay_Case1_ScriptedSessionFinishes` reported PASS.
  - Case 2: `git grep -n -F 'not executable yet' -- cmd/cardtown/play_test.go` printed exactly line 311 (`want := ...`).
  - Case 3: `git grep -n -e 'Shifting Impossibility' -e 'Sinuous Growth' -- pkg/view/executable_test.go` printed nothing (exit 1), and `go test ./pkg/view -run TestView_Case8_ExecutableFlag` exited 0.
  - Case 4: `git diff autoloop/m2-wave-d --stat` listed only `cmd/cardtown/play_test.go`, `pkg/view/executable_test.go`, and `.agents/tasks/m2/T-0164-play-tag-test.md`.
  - Case 5: `go tool task sim-compare -- autoloop/m2-wave-d` printed `[PASS] starter and stress sims match autoloop/m2-wave-d byte for byte, exit codes included`.
  - Case 6: `go tool task check` passed clean (fmt-check, vet, tests, all linter suites passed), and `go tool task sim` passed (200 games, seed 1..200, all invariants holding). Status `review`.
