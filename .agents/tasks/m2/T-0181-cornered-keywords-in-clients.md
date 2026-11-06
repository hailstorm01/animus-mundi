---
id: T-0181
title: The view shows a keyword behind Cornered, and the monitors stay clean on it (directive #117 (a), (c), harness, flash, tests only)
kind: harness
tier: flash
wave: D.1
status: in-progress
branch: m2/autoloop-wave-d.1/T-0181-cornered-keywords-in-clients
depends_on: [T-0180]
claims: [test:TestView_KeywordBehindCorneredShowsOnlyWhileCornered, test:TestMonitorKeywords_CorneredResistantInRealGames]
cards: []
requirements: []
---

## Goal
The two cases of #117 that `rules-test-author` can't write, because it writes only under `pkg/engine` (D2.4, as for T-0164). A seat's view lists a keyword behind Cornered only while Cornered, and real games with a `Cornered: Resistant 1` card break no rule monitor. No non-test code changes: same declarations, same sim output.

## Inputs
- Board directive #117 (a) and (c). T-0180 (merge 995fb2f) built the shape and its engine tests: `ability.CorneredKeywords`, `Def.Cornered` and the grant in `g.Keywords` ([layers.go](../../../pkg/engine/layers.go)).
- `pkg/view/view.go:149`: `AssetView.Keywords` reads `g.Keywords(a)`. `pkg/view/view_test.go`'s `newScene`, `put` and `testUnit` are the setup, and `TestView_Case2_FaceDown` is the model for finding an asset in `v.Theater`.
- `pkg/sim/monitor_damage.go`: `checkDamage` reads Resistant from the damage event (T-0179) and counts `FR-KWD-002b` when it isn't 0. `pkg/engine/damage.go:58` writes the event's `Resistant` from `g.Keywords(target)`.
- **Why the sim test reads coverage.** No other card in its decks has Resistant, and its own card prints none. So a non-zero `FR-KWD-002b` count means a damage event reported the grant in a real game, and the test can't pass on games where Cornered never matters.
- Reuse search (D2.3): no test in `pkg/view` or `pkg/sim` sets `Def.Cornered` (`git grep -n Cornered -- pkg/view pkg/sim` finds none). The `policy` loop in the sim test is the same setup `TestDamage_Case4_RealGamesAreClean` uses, which is test setup, not a rule (D2.0).
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at c759280. Both new tests pass. `gofmt -l pkg` prints nothing, `go tool task check` and `go tool task sim` exit 0, and `go tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`. The dry run counted 318 (`RandomPolicy`) and 328 (`ProgressPolicy`) `FR-KWD-002b` checks over 50 games each. Every expected output below is from that dry run.
- **Mutations of the dry run, each reverted:** `view.go` reading `a.Printed().Keywords` fails the view test. Removing the grant in `g.Keywords` fails both tests. `Resistant: 0` in `damage.go`'s event fails the sim test with a `damage-sequence` violation.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/view/view_test.go`
Add this right before the comment `// T-0168 Cases 1-4: The play log hides an opponent's deck search pick.`, with nothing else changed:
```go
// T-0181 (directive #117 (a)): an asset's view shows a keyword behind Cornered
// to both seats, and only while its controller is Cornered (SVY-R18).
func TestView_KeywordBehindCorneredShowsOnlyWhileCornered(t *testing.T) {
	s := newScene(t)
	g := s.g
	c := testUnit("Alpha Brawler", 2, 2, engine.Keywords{})
	c.Ability = &ability.Def{Cornered: ability.CorneredKeywords{Aggressive: 3}}
	brawler := s.put(0, engine.ZoneMainTheater, c)
	s.put(1, engine.ZoneMainTheater, testUnit("Beta One", 1, 1, engine.Keywords{}))

	keywords := func(seat engine.PlayerID) []string {
		t.Helper()
		v := view.PlayerView(g, seat)
		for _, a := range v.Theater {
			if a.ID == brawler.ID {
				return a.Keywords
			}
		}
		t.Fatalf("seat %d: asset %d missing from the theater %+v", seat, brawler.ID, v.Theater)
		return nil
	}
	want := func(when string, kw []string) {
		t.Helper()
		for seat := range engine.PlayerID(2) {
			if got := keywords(seat); !slices.Equal(got, kw) {
				t.Errorf("%s: seat %d sees keywords %v, want %v", when, seat, got, kw)
			}
		}
	}

	want("1 against 1, not Cornered", nil)
	second := s.put(1, engine.ZoneMainTheater, testUnit("Beta Two", 1, 1, engine.Keywords{}))
	want("1 against 2, Cornered", []string{"Aggressive 3"})
	g.Theater = slices.DeleteFunc(g.Theater, func(a *engine.Asset) bool { return a == second })
	want("back to 1 against 1, not Cornered", nil)
}

```
The block ends in one blank line, so the T-0168 comment stays one blank line below the new test.

### 2. `pkg/sim/monitor_keywords_test.go`
1. In the imports, add `"animus-mundi/pkg/engine/ability"` on the line after `"animus-mundi/pkg/engine"`.
2. Add this at the end of the file, after one blank line:
```go
// T-0181 (directive #117 (c)): real games with a keyword behind Cornered break
// no monitor. Its card prints no Resistant and no other card has it, so each
// FR-KWD-002b check is a damage event that reports the grant.
func TestMonitorKeywords_CorneredResistantInRealGames(t *testing.T) {
	hide := unit("Hide", 2, 4, engine.Keywords{})
	hide.Ability = &ability.Def{Cornered: ability.CorneredKeywords{Resistant: 1}}
	var main []*engine.Card
	for range 10 {
		main = append(main, hide, unit("Vanilla", 3, 2, engine.Keywords{}))
	}
	decks := [2]engine.DeckList{{HQ: hq(3, 20), Main: main}, {HQ: hq(3, 20), Main: main}}
	for _, p := range []struct {
		name   string
		policy func(seed uint64) Policy
	}{{"RandomPolicy", nil}, {"ProgressPolicy", ProgressPolicy}} {
		st, err := SimulateWith(decks, 50, 1, 400, SimOptions{Policy: p.policy})
		if err != nil {
			t.Errorf("%s: %v", p.name, err)
			continue
		}
		if st.Coverage["damage-sequence"]["FR-KWD-002b"] == 0 {
			t.Errorf("%s: no damage event reported the Resistant behind Cornered", p.name)
		}
	}
}
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
These edits are the whole change. Both tests pin behavior T-0180 and T-0179 already built, so they pass as soon as they are written, and there is no implementation commit (as in T-0164). Commit the edits once, as `test(T-0181): the view and the monitors on a keyword behind Cornered`, then hand off with the usual status commit. W is your worktree. Run each command on its own, pointed at W, and record each command and its result in the Log. The reviewer re-runs them, and makes the mutations in Inputs, because a test-only task has no `failed before` to check.
1. **The view test** (#117 (a)): `go -C W test ./pkg/view -run TestView_KeywordBehindCorneredShowsOnlyWhileCornered -v` exits 0, and its output has `--- PASS` for that test.
2. **The monitor test** (#117 (c)): `go -C W test ./pkg/sim -run TestMonitorKeywords_CorneredResistantInRealGames -v` exits 0, and its output has `--- PASS` for that test.
3. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.1 --numstat -- pkg cmd` prints exactly these lines: `27	0	pkg/sim/monitor_keywords_test.go` and `37	0	pkg/view/view_test.go`.
4. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.1 --stat` lists only those two files and this task file.
5. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`.
6. `gofmt -l W/pkg/sim W/pkg/view` prints nothing, and `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/view/view_test.go` (edit 1 only)
- `pkg/sim/monitor_keywords_test.go` (edit 2 only)

## Done When
- [ ] Cases 1-6 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-05 lead-architect (wave D.1 run, session 11): split from T-0180 (D2.4). Status `queued`.
- 2026-11-05 lead-architect (wave D.1 run, new runner, session 1): synced `main` first as c759280, bringing d641da3 and 58a938a (the #122 revision), with no conflict. The previous runner's session 12 hit the usage limit with no commit and left a clean worktree holding the run branch, which this session removed. Not `WAVE-COMPLETE`: #117-#119 and #121 are `open: before the next Deliverables task`. #122 is `open: after Wave D.1 (#121), before Wave E`, so it doesn't apply yet. Wrote Inputs, Edits, Acceptance Cases and Allowed Files against c759280. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120: no such card). Next: `kernel-engineer` on Flash.
