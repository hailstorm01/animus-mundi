---
id: T-0182
title: The hand view shows what a card costs now (directive #118 (a), harness, flash)
kind: harness
tier: flash
wave: D.1
status: in-progress
branch: m2/autoloop-wave-d.1/T-0182-hand-view-cost-now
depends_on: [T-0181]
claims: [field:CardView.ReqNow, field:CardView.TacNow, test:TestView_HandCardCarriesItsCostNow, test:TestPlay_PrintScreenShowsAHandCardsCostNow]
cards: []
requirements: [FR-CST-001]
---

## Goal
A seat's own hand cards carry what they cost now, read through `g.EffectiveCost`, and the play screen prints that. `ReqCost` and `TacCost` stay printed. No card changes a cost yet, so nothing a player sees changes today. The first card that does (*Boneyard Amalgam*) adds its own view test under #118 (b), which stays on the board as `waiting:`.

## Inputs
- Board directive #118 (a) (structural review 8, finding 3, approved by the user, 2026-11-05).
- `pkg/engine/costs.go:14`: `EffectiveCost` is the one read of a card's cost. It takes an `*Instance`, which `side` has in its hand loop and `cardView` doesn't, so the fields are set in `side`, beside `c.Inst`.
- `cmd/cardtown/play_screen.go:72` is the only place outside `pkg/engine` and tests that reads `ReqCost` or `TacCost` of a `CardView` (a `git grep` for `ReqCost` and `TacCost` in the Go files). `actionLabel` prints no cost.
- Reuse search (D2.3): nothing in `pkg/view` or `cmd/cardtown` works out a current cost today. `EffectiveCost` is called, and nothing computes a cost a second time.
- **An edit to an existing test (D2.4).** `TestPlay_PrintScreenTagsANotExecutableHandCard` in `cmd/cardtown/play_test.go` (T-0164) sets `ReqCost` and `TacCost` on its two hand cards and expects them on the screen. The screen now prints `ReqNow` and `TacNow`, so edit 1 changes those two literals and nothing else in that test. Its expected text doesn't change.
- **What a test can't see yet.** `EffectiveCost` returns the printed cost, so no test can tell `g.EffectiveCost(inst)` from `inst.Card.ReqCost` until a card changes a cost. The reviewer checks that line in the diff, and #118 (b) pins it.
- **The view is not in the sims.** `pkg/sim` and `pkg/eventlog` don't import `pkg/view`, so `sim-compare` is expected to pass.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at a94e037. With only the test edits, both test packages exit 1 on the missing fields (Case 1). With all the edits: `gofmt -l` clean, the tests pass, `go tool task check` and `go tool task sim` exit 0, and `go tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`. Every expected output below is from that dry run.
- **Mutations of the dry run, each reverted:** `side` not setting the fields fails the view test. `cardView` setting them on every card fails the view test on a discard card. Swapping `ReqNow` and `TacNow` in `side` fails the view test. The screen printing `c.ReqCost, c.TacCost` fails both screen tests.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `cmd/cardtown/play_test.go`, in the `test(T-0182)` commit
1. In `TestPlay_PrintScreenTagsANotExecutableHandCard`, replace the two card lines with:
   ```go
   		{Card: "Ready Card", ReqNow: 1, TacNow: 2, Power: 3, Health: 4, Executable: true},
   		{Card: "Stub Card", ReqNow: 5, TacNow: 6, Power: 7, Health: 8},
   ```
2. Right after that test's closing brace, and before the comment `// T-0161 Case 3: The play screen labels the Activate action with asset name, id and text.`, add this, with one blank line on each side:
   ```go
   // T-0182 (directive #118 (a)): the screen prints what a hand card costs now, not
   // its printed cost.
   func TestPlay_PrintScreenShowsAHandCardsCostNow(t *testing.T) {
   	var v view.View
   	v.You.Hand = []view.CardView{
   		{Card: "Cheap Now", ReqCost: 10, TacCost: 3, ReqNow: 4, TacNow: 1, Power: 5, Health: 6, Executable: true},
   	}
   	var buf bytes.Buffer
   	printScreen(&buf, v, nil)
   	out := buf.String()

   	want := "Your Hand:\n  [1] Cheap Now (R4 T1, 5/6)\n"
   	if !strings.Contains(out, want) {
   		t.Errorf("printScreen output did not contain expected hand section:\n%s\nGot output:\n%s", want, out)
   	}
   }
   ```

### 2. `pkg/view/view_test.go`, in the `test(T-0182)` commit
Add this right before the comment `// T-0168 Cases 1-4: The play log hides an opponent's deck search pick.`, with nothing else changed:
```go
// T-0182 (directive #118 (a)): each seat's own hand cards carry what they cost
// now, beside the printed cost. A discard card carries the printed cost and no
// cost now. No card changes a cost yet, so now equals printed until #118 (b).
func TestView_HandCardCarriesItsCostNow(t *testing.T) {
	s := newContentScene(t)
	for seat := range engine.PlayerID(2) {
		v := view.PlayerView(s.g, seat)
		if len(v.You.Hand) == 0 || len(v.You.Discard) == 0 || len(v.Opponent.Discard) == 0 {
			t.Fatalf("seat %d: hand %+v, discards %+v and %+v, want a card in each", seat, v.You.Hand, v.You.Discard, v.Opponent.Discard)
		}
		for _, h := range v.You.Hand {
			if h.ReqNow != 2 || h.TacNow != 1 || h.ReqCost != 2 || h.TacCost != 1 {
				t.Errorf("seat %d: hand card %+v, want R2 T1 now and printed", seat, h)
			}
		}
		for _, c := range append(slices.Clone(v.You.Discard), v.Opponent.Discard...) {
			if c.ReqNow != 0 || c.TacNow != 0 || c.ReqCost != 2 || c.TacCost != 1 {
				t.Errorf("seat %d: discard card %+v, want printed R2 T1 and no cost now", seat, c)
			}
		}
	}
}

```
The block ends in one blank line, so the T-0168 comment stays one blank line below the new test.

### 3. `pkg/view/view.go`, in the `feat(view)` commit
1. Replace the package comment's line `//   - its own hand: card names, printed characteristics and instance IDs;` with:
   ```go
   //   - its own hand: card names, printed characteristics, what each costs now
   //     and instance IDs;
   ```
2. In `CardView`, add these two lines right after `TacCost    int`:
   ```go
   	ReqNow     int // only for the seat's own hand cards: what playing it costs now (Game.EffectiveCost); 0 elsewhere
   	TacNow     int // as ReqNow
   ```
3. In `side`'s hand loop, add this line right after `c.Inst = inst.ID`:
   ```go
   			c.ReqNow, c.TacNow = g.EffectiveCost(inst)
   ```

### 4. `cmd/cardtown/play_screen.go`, line 72, in the `feat(view)` commit
Replace it with:
```go
			fmt.Fprintf(out, "  [%d] %s (R%d T%d, %d/%d)%s\n", i+1, c.Card, c.ReqNow, c.TacNow, c.Power, c.Health, notExec)
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
Edits 1 and 2 are the whole of the test commit: commit them as `test(T-0182): the hand view shows what a card costs now`, run case 1, then make edits 3 and 4 and commit them as `feat(view): the hand view shows what a card costs now (T-0182)`. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Fails before the edits** (after the test commit, before edit 3):
   - `go -C W test ./pkg/view -run TestView_HandCardCarriesItsCostNow` exits 1, and its output has `h.ReqNow undefined`.
   - `go -C W test ./cmd/cardtown -run TestPlay_PrintScreen` exits 1, and its output has `unknown field ReqNow`.
2. **The view test passes after them** (#118 (a)): `go -C W test ./pkg/view -run TestView_HandCardCarriesItsCostNow -v` exits 0, and its output has `--- PASS` for that test.
3. **The screen tests pass after them** (#118 (a)): `go -C W test ./cmd/cardtown -run TestPlay_PrintScreen -v` exits 0, and its output has `--- PASS` for `TestPlay_PrintScreenShowsAHandCardsCostNow` and `TestPlay_PrintScreenTagsANotExecutableHandCard`.
4. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.1 --numstat -- pkg cmd` prints exactly these lines: `1	1	cmd/cardtown/play_screen.go`, `19	2	cmd/cardtown/play_test.go`, `5	1	pkg/view/view.go` and `23	0	pkg/view/view_test.go`.
5. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.1 --stat` lists only those four files and this task file.
6. **Sims byte-identical** (#118 (a)): `go -C W tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`.
7. `gofmt -l W/pkg/view W/cmd/cardtown` prints nothing, and `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/view/view.go`
- `cmd/cardtown/play_screen.go`
- `pkg/view/view_test.go` (edit 2 only)
- `cmd/cardtown/play_test.go` (edit 1 only)

## Done When
- [ ] Cases 1-7 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-05 lead-architect (wave D.1 run, session 13): no `main` commits to sync. Not `WAVE-COMPLETE`: #118 (a), #119 and #121 are `open: before the next Deliverables task`. #122 is `open: after Wave D.1 (#121), before Wave E`, so it doesn't apply yet. Wrote this task from #118 (a), next in board order. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120: no such card). Next: `kernel-engineer` on Flash.
