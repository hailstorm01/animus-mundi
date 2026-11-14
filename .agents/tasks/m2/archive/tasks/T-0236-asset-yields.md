---
id: T-0236
title: "Battlefield assets show their resource yields: AssetView.ReqYield and TacYield, and the play screen's asset line (directive #150, wave F.2 harness, flash)"
kind: harness
tier: flash
wave: F.2
status: done
branch: m2/autoloop-wave-f.2/T-0236-asset-yields
depends_on: [T-0235]
claims: [field:view.AssetView.ReqYield, field:view.AssetView.TacYield]
cards: []
requirements: [FR-ZON-005, FR-RES-001]
---

## Goal
Every battlefield asset's view carries its printed Requisition and Tactics yields, the Scratch face's for a face-down asset, and `cardtown play` prints them on every battlefield line, both seats' sides (#150). When this merges, #150 is `applied`.

## Inputs
- Board directive #150 (user playtest feedback, 2026-11-13), `open: before the next Deliverables task`; #149 is applied (T-0233, T-0234, T-0235).
- Design (lead decisions):
  - **The view.** `view.AssetView` gains `ReqYield` and `TacYield`, read in `assetView` from `a.Printed()`, the same card Generate Resources reads them from (`pkg/engine/resources.go:33-35`, FR-RES-001). `Printed` is the Scratch face for a face-down asset (`pkg/engine/card.go:139`: R1 T1), so a face-down asset shows R1 T1 to both seats and its real yields leak nothing (FR-ZON-005). No layer changes yields today, so printed is effective; if one ever does, the view and Generate Resources change together.
  - **The screen.** `formatAsset` (`cmd/cardtown/play_screen.go`) prints `Yield R<req> T<tac>` right after the Ready/Exhausted status and before armor and damage, on every line it formats (both HQs, both Reserves, the Main Theater): `Chainbreaker Irregular#14 (2/3, Ready, Yield R1 T0)`, `Wall#5 (0/6, Ready, Yield R0 T3, Armor 1, Dmg 2)`. The word `Yield` keeps it apart from the hand's cost `(R1 T2, 3/4)`.
  - **The package comment** of `pkg/view` says each battlefield asset shows its printed yields, the Scratch face's when face-down.
- Reuse search (D2.3): `git grep -n -i yield` in `pkg/view` and `cmd/cardtown` finds nothing; the only readers are `engine`'s `generate` and `canYield`. Reading `Printed().ReqYield` in the view reads the card's field; it is not a second copy of a rule.
- Size (D2.4): two non-test files.
- **Expected output difference:** none in `cardtown sim` (the view and the play screen are not in sim output), so `go tool task sim-compare` passes. `cardtown play` gains the yields on battlefield lines. No existing test pins a battlefield line (searched `Ready)` and `formatAsset` in the `cmd/cardtown` and `pkg/view` tests).
- The lead dry-ran these exact edits and these exact test files in a throwaway worktree of the run branch at 584249b. Before the edits, both test packages fail to build (the new fields don't exist). With both edits: `gofmt -l cmd pkg` clean; `go test ./pkg/view/... ./cmd/cardtown` ok; `go tool task check` exit 0; `go tool task sim-compare -- autoloop/m2-wave-f.2` `[PASS]`; `cardtown lint fences` exit 0; `go tool task sim` exit 0 (200 games, P0 63.5% / P1 35.0% / draws 1.5%). Blobs: `pkg/view/view.go` 14c01f3, `cmd/cardtown/play_screen.go` 896f84e, `pkg/view/yield_test.go` 19a4195, `cmd/cardtown/play_yield_test.go` 3d53068.

## Commits
1. `test(T-0236): battlefield assets show their resource yields` with only the two new test files below. Before committing, run `go -C W test ./pkg/view` and `go -C W test ./cmd/cardtown` and record that both fail (build failure: `AssetView` has no field `ReqYield`).
2. `feat(view): battlefield assets carry their printed yields; the play screen prints them (T-0236)` with Edits 1-2.

## Tests (lead decisions; copy exactly, tabs for indentation)
### New file `pkg/view/yield_test.go`
```go
package view_test

import (
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/view"
)

// Tests for T-0236 (directive #150): every battlefield asset shows its printed
// resource yields to both seats, and a face-down asset shows the Scratch
// face's (FR-ZON-005).

// everyAsset is every battlefield asset in v: both HQs, both Reserves and the
// Main Theater.
func everyAsset(v view.View) []view.AssetView {
	var out []view.AssetView
	for _, s := range []view.Side{v.You, v.Opponent} {
		if s.HQ != nil {
			out = append(out, *s.HQ)
		}
		out = append(out, s.Reserves...)
	}
	return append(out, v.Theater...)
}

func TestView_AssetShowsItsPrintedYields(t *testing.T) {
	s := newScene(t)
	g := s.g
	unit := testUnit("Alpha Yielder", 2, 3, engine.Keywords{})
	unit.ReqYield, unit.TacYield = 2, 0
	a0 := s.put(0, engine.ZoneReserves, unit)
	depot := &engine.Card{Name: "Beta Depot", Type: engine.TypeStructure, Health: 4, ReqYield: 0, TacYield: 3,
		DeployZones: []engine.Zone{engine.ZoneReserves}}
	a1 := s.put(1, engine.ZoneReserves, depot)
	hidden := testUnit("Alpha Hidden Miner", 4, 4, engine.Keywords{})
	hidden.ReqYield, hidden.TacYield = 0, 3
	fd := &engine.Asset{ID: s.id(), Inst: s.inst(0, hidden), Controller: 0, Zone: engine.ZoneMainTheater, FaceDown: true}
	g.Theater = append(g.Theater, fd)
	theirs := s.put(1, engine.ZoneMainTheater, testUnit("Beta Raider", 3, 3, engine.Keywords{}))

	want := map[int][2]int{
		g.Players[0].HQ.ID: {1, 2}, // testHQ prints R1 T2
		g.Players[1].HQ.ID: {1, 2},
		a0.ID:              {2, 0}, // (a) a face-up unit's printed yields
		a1.ID:              {0, 3}, // (c) a Structure's
		fd.ID:              {1, 1}, // (b) the Scratch face's, not the card's R0 T3
		theirs.ID:          {1, 1}, // testUnit prints R1 T1
	}
	for seat := range engine.PlayerID(2) {
		v := view.PlayerView(g, seat)
		seen := map[int]bool{}
		for _, a := range everyAsset(v) {
			w, ok := want[a.ID]
			if !ok {
				continue
			}
			seen[a.ID] = true
			if a.ReqYield != w[0] || a.TacYield != w[1] {
				t.Errorf("seat %d sees %s#%d with yields R%d T%d, want R%d T%d", seat, a.Card, a.ID, a.ReqYield, a.TacYield, w[0], w[1])
			}
		}
		// (d) each seat sees every asset's yields, the opponent's included.
		if len(seen) != len(want) {
			t.Errorf("seat %d sees %d of the %d assets", seat, len(seen), len(want))
		}
	}
}
```

### New file `cmd/cardtown/play_yield_test.go`
```go
package main

import (
	"bytes"
	"strings"
	"testing"

	"animus-mundi/pkg/view"
)

// Tests for T-0236 (directive #150): each battlefield line on the play screen
// shows the asset's resource yields, on both seats' sides.

func TestPlay_FormatAssetShowsYields(t *testing.T) {
	for _, tc := range []struct {
		name string
		a    view.AssetView
		want string
	}{
		{"(a) a face-up unit", view.AssetView{ID: 14, Card: "Chainbreaker Irregular", Power: 2, MaxHealth: 3, ReqYield: 1},
			"Chainbreaker Irregular#14 (2/3, Ready, Yield R1 T0)"},
		{"(b) a face-down asset", view.AssetView{ID: 9, Card: "Scratch Company", FaceDown: true, Exhausted: true, Power: 1, MaxHealth: 1, ReqYield: 1, TacYield: 1},
			"Scratch Company#9 (1/1, Exhausted, Yield R1 T1)"},
		{"(c) a Headquarters", view.AssetView{ID: 1, Card: "Animoceps Rager", Power: 3, MaxHealth: 20, ReqYield: 1, TacYield: 2},
			"Animoceps Rager#1 (3/20, Ready, Yield R1 T2)"},
		{"armor and damage follow the yields", view.AssetView{ID: 5, Card: "Wall", Power: 0, MaxHealth: 6, TacYield: 3, Armor: 1, Damage: 2},
			"Wall#5 (0/6, Ready, Yield R0 T3, Armor 1, Dmg 2)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatAsset(tc.a); got != tc.want {
				t.Errorf("formatAsset = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPlay_PrintScreenShowsYieldsOnBothSides(t *testing.T) {
	v := view.View{
		You: view.Side{
			HQ:       &view.AssetView{ID: 1, Card: "Your HQ", Power: 3, MaxHealth: 20, ReqYield: 1, TacYield: 2},
			Reserves: []view.AssetView{{ID: 2, Card: "Your Unit", Power: 2, MaxHealth: 2, ReqYield: 2}},
		},
		Opponent: view.Side{
			HQ:       &view.AssetView{ID: 3, Card: "Their HQ", Power: 2, MaxHealth: 18, ReqYield: 2, TacYield: 1},
			Reserves: []view.AssetView{{ID: 4, Card: "Their Unit", Power: 1, MaxHealth: 4, TacYield: 1}},
		},
		Theater: []view.AssetView{{ID: 5, Card: "Scratch Company", Controller: 1, FaceDown: true, Power: 1, MaxHealth: 1, ReqYield: 1, TacYield: 1}},
	}
	var buf bytes.Buffer
	printScreen(&buf, v, nil)
	out := buf.String()
	for _, want := range []string{
		"Opponent HQ: Their HQ#3 (2/18, Ready, Yield R2 T1)\n",
		"Opponent Reserves:\n  Their Unit#4 (1/4, Ready, Yield R0 T1)\n",
		"Main Theater:\n  Scratch Company#5 (1/1, Ready, Yield R1 T1) (Player 1)\n",
		"Your HQ: Your HQ#1 (3/20, Ready, Yield R1 T2)\n",
		"Your Reserves:\n  Your Unit#2 (2/2, Ready, Yield R2 T0)\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("screen lacks %q:\n%s", want, out)
		}
	}
}
```

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at 584249b.

### 1. `pkg/view/view.go`: the yields
(a) The package comment, lines 10-12, which read
```go
//   - public: both discard piles (the seat's own with instance IDs), HQs,
//     Reserves and resource pools, the Main Theater, the stack, the turn, the
//     phase, the active player and the outcome.
```
become
```go
//   - public: both discard piles (the seat's own with instance IDs), HQs,
//     Reserves and resource pools, the Main Theater, the stack, the turn, the
//     phase, the active player and the outcome. Each battlefield asset shows
//     its effective power, health and keywords and its printed resource
//     yields (a face-down asset, the Scratch face's).
```
(b) The `AssetView` comment, lines 64-65, which read
```go
// AssetView is a battlefield asset. Power, MaxHealth and Keywords are
// effective (the layer functions); a face-down asset shows the Scratch face.
```
become
```go
// AssetView is a battlefield asset. Power, MaxHealth and Keywords are
// effective (the layer functions); ReqYield and TacYield are printed, as
// Generate Resources reads them (FR-RES-001); a face-down asset shows the
// Scratch face.
```
(c) In `AssetView`, right after the line `	Text              string   // printed rules text` (line 78), add:
```go
	ReqYield          int      // printed Requisition yield
	TacYield          int      // printed Tactics yield
```
(d) In `assetView`, line 160, which reads
```go
		Text: a.Printed().RulesText,
```
becomes
```go
		Text: a.Printed().RulesText, ReqYield: a.Printed().ReqYield, TacYield: a.Printed().TacYield,
```

### 2. `cmd/cardtown/play_screen.go`: the asset line
In `formatAsset`, line 129, which reads
```go
	return fmt.Sprintf("%s#%d (%d/%d, %s%s)", a.Card, a.ID, a.Power, a.MaxHealth, status, statStr)
```
becomes
```go
	return fmt.Sprintf("%s#%d (%d/%d, %s, Yield R%d T%d%s)", a.Card, a.ID, a.Power, a.MaxHealth, status, a.ReqYield, a.TacYield, statStr)
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The tests failed first:** at the `test(T-0236)` commit, before Edit 1, `go -C W test ./pkg/view` and `go -C W test ./cmd/cardtown` both fail to build, naming the missing `ReqYield` or `TacYield` field.
2. **The view carries the printed yields** (#150 (a)-(d)): `go -C W test ./pkg/view -run TestView_AssetShowsItsPrintedYields` passes: (a) a face-up unit shows its printed yields; (b) a face-down asset shows R1 T1, the Scratch face's, not its card's, to both seats; (c) both Headquarters and a Structure show theirs; (d) each seat sees the opponent's assets' yields too.
3. **The screen prints them** (#150): `go -C W test ./cmd/cardtown -run TestPlay_FormatAssetShowsYields` and `go -C W test ./cmd/cardtown -run TestPlay_PrintScreenShowsYieldsOnBothSides` pass: every battlefield line, both HQs, both Reserves and the Main Theater, reads `Yield R<n> T<n>` after its status and before armor and damage.
4. **No hidden information** (#150 (e)): `go -C W test ./pkg/view/...` passes (the `viewtest` leak tests and `TestView_Case2_FaceDown`), and `go -C W test ./cmd/cardtown -run TestPlay_Case4_NoHiddenInformation` passes.
5. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-f.2` prints `[PASS] starter and stress sims match autoloop/m2-wave-f.2 byte for byte, exit codes included`.
6. **Only the listed edits:** `git -C W diff autoloop/m2-wave-f.2 --numstat -- pkg cmd` prints exactly four lines: `1	1	cmd/cardtown/play_screen.go`, `63	0	cmd/cardtown/play_yield_test.go`, `9	3	pkg/view/view.go` and `68	0	pkg/view/yield_test.go`.
7. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
8. **Nothing else changed:** `git -C W diff autoloop/m2-wave-f.2 --stat` lists only the four files of case 6 and this task file.
9. `gofmt -l W/cmd W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/view/view.go`
- `cmd/cardtown/play_screen.go`
- `pkg/view/yield_test.go` (new; the `test(T-0236)` commit only)
- `cmd/cardtown/play_yield_test.go` (new; the `test(T-0236)` commit only)

## Done When
- [x] Cases 1-9 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-f.2` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-14, lead-architect (wave F.2 run, session 4): written in full against 584249b for #150. Nothing on `main` to sync (`git log autoloop/m2-wave-f.2..main` is empty), and the scratch folder was empty. Harness, so it skips `tests` and goes straight to `ready`. Tier `flash`, because both edits and both test files are exact and were dry-run. The tests go in two new files, so the test commit touches no earlier test. Every command here is on the headless allow-list: one per line, no pipes, no `$`, no unquoted parentheses. Waiting directives checked: #6, #7 (mode `serial`), #15 (no turn-boundary hook), #71, #91 (2), #106, #120, #141 (3), #142: not fired. Next: `kernel-engineer` on Flash.
- 2026-11-14, kernel-engineer (wave F.2 run, Gemini Flash): implemented T-0236 for #150.
  - Commits:
    - `392f35b`: `test(T-0236): battlefield assets show their resource yields` (added `pkg/view/yield_test.go` and `cmd/cardtown/play_yield_test.go`)
    - `d731d06`: `feat(view): battlefield assets carry their printed yields; the play screen prints them (T-0236)` (Edits 1-2 to `pkg/view/view.go` and `cmd/cardtown/play_screen.go`)
  - Diff:
    - `pkg/view/view.go`: updated package comment and `AssetView` docstring; added `ReqYield` and `TacYield` fields to `AssetView`; mapped `a.Printed().ReqYield` and `a.Printed().TacYield` in `assetView` (reused `a.Printed()`, matching `Generate Resources` and preserving double-blind face-down Scratch yields per FR-ZON-005, FR-RES-001).
    - `cmd/cardtown/play_screen.go`: updated `formatAsset` to format `, Yield R%d T%d` after status and before armor/damage.
    - `pkg/view/yield_test.go`: added `TestView_AssetShowsItsPrintedYields`.
    - `cmd/cardtown/play_yield_test.go`: added `TestPlay_FormatAssetShowsYields` and `TestPlay_PrintScreenShowsYieldsOnBothSides`.
  - Commands and results:
    - Case 1: `go test ./pkg/view` and `go test ./cmd/cardtown` at test commit both failed compilation due to undefined/unknown `ReqYield` and `TacYield`.
    - Case 2: `go test ./pkg/view -run TestView_AssetShowsItsPrintedYields` passed (0.655s).
    - Case 3: `go test ./cmd/cardtown -run TestPlay_FormatAssetShowsYields` and `TestPlay_PrintScreenShowsYieldsOnBothSides` passed (0.594s, 0.578s).
    - Case 4: `go test ./pkg/view/...` and `go test ./cmd/cardtown -run TestPlay_Case4_NoHiddenInformation` passed (2.637s, 0.606s).
    - Case 5: `go tool task sim-compare -- autoloop/m2-wave-f.2` printed `[PASS] starter and stress sims match autoloop/m2-wave-f.2 byte for byte, exit codes included`.
    - Case 6: `git diff autoloop/m2-wave-f.2 --numstat -- pkg cmd` printed exactly the 4 expected lines.
    - Case 7: `go run ./cmd/cardtown lint fences` passed (exit 0).
    - Case 8: `git diff autoloop/m2-wave-f.2 --stat` listed only the 4 code/test files and this task file.
    - Case 9: `gofmt -l cmd pkg` clean; `go tool task check` passed (vet, all tests, linter dispatch clean); `go tool task sim` clean (200 games, P0 63.5% / P1 35.0% / draws 1.5%).
  - Ready for review.
- 2026-11-14, lead-architect (wave F.2 run, session 4, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-f.2` as 3dca647. Flash run `56d19709-64af-460f-985f-e79b570adbd8`: exit 0, `status: SUCCESS`, no `denied_actions` key, and no step with `"status":"ERROR"` in its transcript; `.err` holds only agy's "root agent idle" notice. Review of `git diff autoloop/m2-wave-f.2...m2/autoloop-wave-f.2/T-0236-asset-yields` (07cd58c):
  - Scope: the two Allowed non-test files, the two new test files and this file. The `test(T-0236)` commit (392f35b) holds only the two test files, whose blobs (19a4195, 3d53068) are this task's; the implementation commit (d731d06) holds only `view.go` and `play_screen.go`, whose blobs (14c01f3, 896f84e) are the lead's dry run's. No earlier test changed. The tests failed first: re-run at 392f35b by the reviewer, both packages fail to build on the missing `ReqYield` field.
  - Fences and placement: no card name in package code; the field sits in `pkg/view`, the formatting in `cmd/cardtown`, and `pkg/engine` doesn't change. Reuse: the view reads the printed card's fields through `Printed()`, as Generate Resources does; no rule is copied.
  - Hidden information: a face-down asset's yields are the Scratch face's (`TestView_AssetShowsItsPrintedYields` (b)), and the `viewtest` leak tests pass.
  - Re-run by the reviewer in a scratch worktree of the task branch: `gofmt -l cmd pkg` clean; `go test ./pkg/view/... ./cmd/cardtown` ok; `cardtown lint fences` exit 0; `go tool task check` exit 0; `go tool task sim-compare -- autoloop/m2-wave-f.2` `[PASS]`; `go tool task sim` exit 0, 200 games, P0 63.5% / P1 35.0% / draws 1.5%, P0 50/50 and P1 50/50 main cards executable; `git diff autoloop/m2-wave-f.2 --numstat -- pkg cmd` is case 6's four lines.
  - The merge's tree matches the task tip's (`git diff` between them is empty), so the gates hold on the run branch. #150 is `applied` (T-0236) and moves to the directive archive.
