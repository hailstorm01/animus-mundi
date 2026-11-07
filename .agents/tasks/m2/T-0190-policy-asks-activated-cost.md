---
id: T-0190
title: The policy asks the engine what an activated ability costs (directive #127, harness, flash)
kind: harness
tier: flash
wave: D.2
status: review
branch: m2/autoloop-wave-d.2/T-0190-policy-asks-activated-cost
depends_on: [T-0189]
claims: [engine:Game.ActivatedCost, engine:Game.otherCostsPayable, sim:activationCosts, fence:pkg/sim imports, test:TestActivatedCost_ReadsResourcesAndOtherCosts, test:TestFences_Case40_SimImportsOnlyEngine]
cards: []
requirements: [FR-ACT-001]
---

## Goal
`activationCosts` in `pkg/sim` stops reading `ability.Cost` and restating "the source must be Readied". It asks a new exported engine method, `ActivatedCost`, and `pkg/sim`'s non-test files no longer import `pkg/engine/ability` (#127). `costsPayable` and the method share one check of the non-resource costs, `otherCostsPayable`. Play doesn't change, so the sims stay byte-identical.

## Inputs
- Board directive #127 (structural review 9, approved by the user, 2026-11-06). The user accepted its one reversal: it undoes the `ability` import T-0185 added.
- Design note: `docs/architecture/m2-ability-model.md`, Boundaries, Deliverable 6h. The lead updates its sentence on what `sim` imports when this merges, so that file isn't yours to edit.
- Why play doesn't change: `activationCosts` used to leave out only an ability whose source must exhaust and has. `otherCostsPayable` also refuses a chosen-asset cost (`CostKill`, `CostExhaust`) with no option, but no activated ability prints one: the three in `pkg/engine/cards` are all a Maneuver's `ExtraCosts` (`animus.go:298`, `:318`, `:458`, at 7c3a404).
- Why `costsPayable` doesn't change: it still needs every resources cost payable and every other cost payable, and `costDecision` only reads.
- `pkg/sim`'s test files keep importing `ability`, because they build synthetic cards. The fence reads non-test files only.
- `TestFences_Case32_DisallowedImport` already expects `pkg/sim may import only pkg/engine` with `strings.Contains`, so it passes before and after this task, and it stays unedited. Case 40 is the test that the table changed.
- Reuse search (D2.3): `pkg/engine` has no other read of an activated ability's resource cost. `activatedPlacement` (`placement.go:66`) is the one way to build its placement, and the method reuses it.
- `ActivatedCost` needs `a` to print ability `i`, the way `activationCosts` calls it. It has no guard for a bad index. `checkAction` keeps its own, for an action from outside.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 7c3a404. With only Edit 1: `go test ./pkg/toolchain/linter/fences -run TestFences_Case40` fails, and `go vet ./pkg/engine` prints `g.ActivatedCost undefined`. With all the edits: `gofmt -l pkg` is clean, `go tool task check` and `go tool task sim` pass, and `go tool task sim-compare -- autoloop/m2-wave-d.2` prints `[PASS]`. Every expected output below is from that dry run.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The tests fail first.** With only Edit 1 in place: `go -C W test ./pkg/toolchain/linter/fences -run TestFences_Case40` fails, and `go -C W vet ./pkg/engine` prints `g.ActivatedCost undefined`. Then commit Edit 1.
2. **The method reads the costs:** `go -C W test ./pkg/engine -run TestActivatedCost_` passes, with its six subtests: the resources are read and not checked against the pool, each index reads its own ability, exhausting the source needs it Readied, a discard or killing the source is always payable, and a kill and an exhaust cost each need an option.
3. **The fence holds:** `go -C W test ./pkg/toolchain/linter/fences -run TestFences_Case40` passes, and `go -C W run ./cmd/cardtown lint fences` exits 0.
4. **`pkg/sim` drops the import:** `git -C W grep -n engine/ability -- pkg/sim` prints exactly two lines: `pkg/sim/monitor_keywords_test.go:7` and `pkg/sim/policy_test.go:11`.
5. **The policy test is unedited:** `git -C W diff autoloop/m2-wave-d.2 -- pkg/sim/policy_test.go` prints nothing, and `go -C W test ./pkg/sim` passes.
6. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-d.2` prints `[PASS]`.
7. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.2 --numstat -- pkg cmd` prints exactly these five lines:
   ```text
   51	0	pkg/engine/activated_test.go
   33	9	pkg/engine/costs.go
   6	17	pkg/sim/policy.go
   1	1	pkg/toolchain/linter/fences/fences.go
   11	0	pkg/toolchain/linter/fences/fences_test.go
   ```
8. `go -C W tool task check` and `go -C W tool task sim` pass.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at 7c3a404.

### 1. The tests, in the `test(T-0190)` commit
**`pkg/engine/activated_test.go`:** add `"fmt"` as the first line of the import block, above `"reflect"`. Then add this at the end of the file, with one blank line before it:
```go
// T-0190 (directive #127): ActivatedCost is a client's read of what an activated
// ability costs. It reads the resources without checking the pool, and says
// whether the other costs can be paid now.
func TestActivatedCost_ReadsResourcesAndOtherCosts(t *testing.T) {
	check := func(t *testing.T, g *Game, a *Asset, i, wantReq, wantTac int, wantOK bool) {
		t.Helper()
		req, tac, ok := g.ActivatedCost(a, i)
		if req != wantReq || tac != wantTac || ok != wantOK {
			t.Errorf("ActivatedCost(%d) = R%d T%d %v, want R%d T%d %v", i, req, tac, ok, wantReq, wantTac, wantOK)
		}
	}
	t.Run("resources are read, not checked against the pool", func(t *testing.T) {
		s := newScene(t)
		a := s.put(0, ZoneMainTheater, actUnit("Engine", nil, actAbility([]ability.Cost{actRes(1, 2)}, drawN(1))))
		s.setPool(0, 0, 0)
		check(t, s.g, a, 0, 1, 2, true)
	})
	t.Run("each index reads its own ability", func(t *testing.T) {
		s := newScene(t)
		a := s.put(0, ZoneMainTheater, actUnit("Engine", nil,
			actAbility([]ability.Cost{actRes(1, 0)}, drawN(1)),
			actAbility([]ability.Cost{actRes(0, 3)}, drawN(1))))
		check(t, s.g, a, 0, 1, 0, true)
		check(t, s.g, a, 1, 0, 3, true)
	})
	t.Run("exhausting the source needs it Readied", func(t *testing.T) {
		s := newScene(t)
		a := s.put(0, ZoneMainTheater, actUnit("Engine", nil, actAbility([]ability.Cost{actRes(1, 0), actExhaust}, drawN(1))))
		check(t, s.g, a, 0, 1, 0, true)
		a.Exhausted = true
		check(t, s.g, a, 0, 1, 0, false)
	})
	t.Run("a discard or killing the source is always payable", func(t *testing.T) {
		s := newScene(t)
		a := s.put(0, ZoneMainTheater, actUnit("Engine", nil, actAbility([]ability.Cost{{Kind: ability.CostDiscard, UpTo: 1}, actKill}, drawN(1))))
		a.Exhausted = true
		check(t, s.g, a, 0, 0, 0, true)
	})
	for _, kind := range []ability.CostKind{ability.CostKill, ability.CostExhaust} {
		t.Run(fmt.Sprintf("chosen-asset cost %d needs an option", kind), func(t *testing.T) {
			s := newScene(t)
			cost := ability.Cost{Kind: kind, Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindStructure}}
			a := s.put(0, ZoneMainTheater, actUnit("Engine", nil, actAbility([]ability.Cost{cost}, drawN(1))))
			check(t, s.g, a, 0, 0, 0, false)
			s.put(0, ZoneMainTheater, structure("Wall", 0, 9, Keywords{}))
			check(t, s.g, a, 0, 0, 0, true)
		})
	}
}
```

**`pkg/toolchain/linter/fences/fences_test.go`:** add this at the end of the file, with one blank line before it:
```go
// 40. pkg/sim may import pkg/engine and nothing else in the module (T-0190).
func TestFences_Case40_SimImportsOnlyEngine(t *testing.T) {
	repo := setupFixtureRepo(t)
	writePkgFile(t, repo, "pkg/sim/x.go", "package sim\n\nimport _ \"animus-mundi/pkg/engine/ability\"\n")
	out := expectFail(t, repo)
	want := "pkg/sim/x.go:3: imports animus-mundi/pkg/engine/ability; pkg/sim may import only pkg/engine"
	if !strings.Contains(out, "[ERROR] "+want) {
		t.Fatalf("expected %q, got:\n%s", want, out)
	}
}
```
Commit only these two files as `test(T-0190): ...`, after Acceptance Case 1. The `pkg/engine` tests don't compile until Edit 2, because the method doesn't exist yet. That is this commit's expected failure.

### 2. `pkg/engine/costs.go`
Replace lines 66-82, from `// costsPayable reports whether` through the line `case ability.CostExhaustSelf:`, with the block below. The rest of the function, from `if a := g.Asset(pl.source)` on, stays as it is and now belongs to `otherCostsPayable`.
```go
// costsPayable reports whether every cost of pl can be paid now: a resources
// cost needs its Requisition and Tactics (CostResources), and the others are
// otherCostsPayable's to judge (FR-TRN-003a, FR-ACT-001, SVY-R29).
func (g *Game) costsPayable(pl placement) bool {
	for _, c := range pl.costs {
		if c.Kind == ability.CostResources && !g.Players[pl.controller].canPay(c.Req, c.Tac) {
			return false
		}
	}
	return g.otherCostsPayable(pl)
}

// ActivatedCost returns the Requisition and Tactics that a's printed activated
// ability i costs, and whether its other costs can be paid now (FR-ACT-001). It
// checks neither the resources against the pool nor the ability's targets, so a
// client can plan for an ability it can't activate yet. LegalActions says what
// can be activated now.
func (g *Game) ActivatedCost(a *Asset, i int) (req, tac int, ok bool) {
	act := a.Printed().Ability.Activated[i]
	for _, c := range act.Costs {
		if c.Kind == ability.CostResources {
			req += c.Req
			tac += c.Tac
		}
	}
	return req, tac, g.otherCostsPayable(activatedPlacement(a.Controller, a, act))
}

// otherCostsPayable reports whether every cost of pl except its resources can be
// paid now: exhausting the source needs it Readied (CostExhaustSelf), and a kill
// or exhaust cost needs at least one option (SVY-R29). A discard cost is always
// payable, because discarding none pays it, and so is killing the source
// (CostKillSelf).
func (g *Game) otherCostsPayable(pl placement) bool {
	for _, c := range pl.costs {
		switch c.Kind {
		case ability.CostResources:
			continue // not this check's to judge
		case ability.CostDiscard, ability.CostKillSelf:
			continue // discarding none pays it; killing the source always can
		case ability.CostExhaustSelf:
```

### 3. `pkg/sim/policy.go`
1. Delete line 10, `"animus-mundi/pkg/engine/ability"`.
2. Replace lines 160-184, from `// abilities, leaving out an ability whose source` through the closing `}` of the `for _, act := range def.Activated` loop, keeping line 159 above them, with:
```go
// abilities, leaving out an ability whose other costs can't be paid now
// (ActivatedCost). It doesn't check targets, so a tap may fund an ability that
// then has none.
func activationCosts(g *engine.Game) []neededCost {
	var out []neededCost
	for _, as := range g.Assets() {
		def := as.Printed().Ability
		if as.Controller != g.Active || def == nil {
			continue
		}
		for i := range def.Activated {
			if req, tac, ok := g.ActivatedCost(as, i); ok {
				out = append(out, neededCost{resources: resources{req, tac}, source: as.ID})
			}
		}
```
The lines after it (`}`, `return out`, `}`) stay.

### 4. `pkg/toolchain/linter/fences/fences.go`
Change line 40 to:
```go
	"pkg/sim":            {"pkg/engine"},
```

No other file changes.

## Allowed Files
- `pkg/engine/costs.go`
- `pkg/sim/policy.go`
- `pkg/toolchain/linter/fences/fences.go`
- In the `test(T-0190)` commit only: `pkg/engine/activated_test.go` and `pkg/toolchain/linter/fences/fences_test.go`, for the additions in Edit 1. No existing test changes.

## Done When
- [x] Cases 1–8 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-d.2` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-07, lead-architect (wave D.2 run, session 4): wrote this task from board directive #127, next in board order, against the run branch at 7c3a404. No `main` commits to sync. Not `WAVE-COMPLETE`: #127 and #129-#135 are `open: before the next Deliverables task`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120, #128: no such card). Next: `kernel-engineer` on Flash.
- 2026-11-07, kernel-engineer (flash): implemented T-0190 per runbook and task specifications.
  - Diff shows:
    - `pkg/engine/activated_test.go`: added `TestActivatedCost_ReadsResourcesAndOtherCosts` covering resource reads independent of pool, per-index reading, exhaust requirements, free discard/kill-self costs, and chosen-asset kill/exhaust options.
    - `pkg/toolchain/linter/fences/fences_test.go`: added `TestFences_Case40_SimImportsOnlyEngine`.
    - `pkg/engine/costs.go`: split non-resource checks into `otherCostsPayable(pl)` shared by `costsPayable` and new method `ActivatedCost(a *Asset, i int) (req, tac int, ok bool)` reusing `activatedPlacement`.
    - `pkg/sim/policy.go`: removed `animus-mundi/pkg/engine/ability` import and updated `activationCosts` to query `g.ActivatedCost(as, i)`.
    - `pkg/toolchain/linter/fences/fences.go`: restricted `pkg/sim` in `allowedImports` to only `{"pkg/engine"}`.
  - Reuse:
    - Reused `activatedPlacement` to build placement in `ActivatedCost`, and shared `otherCostsPayable` with `costsPayable`.
  - Commands and results:
    - Case 1: `go -C W test ./pkg/toolchain/linter/fences -run TestFences_Case40` failed; `go -C W vet ./pkg/engine` printed `g.ActivatedCost undefined`. Committed `test(T-0190): fail before ActivatedCost and sim fence` (38f6f75).
    - Case 2: `go -C W test ./pkg/engine -run TestActivatedCost_` passed (6/6 subtests).
    - Case 3: `go -C W test ./pkg/toolchain/linter/fences -run TestFences_Case40` passed; `go -C W run ./cmd/cardtown lint fences` exited 0.
    - Case 4: `git -C W grep -n engine/ability -- pkg/sim` returned exactly two test files (`monitor_keywords_test.go:7`, `policy_test.go:11`).
    - Case 5: `git -C W diff autoloop/m2-wave-d.2 -- pkg/sim/policy_test.go` printed nothing; `go -C W test ./pkg/sim` passed.
    - Case 6: `go -C W tool task sim-compare -- autoloop/m2-wave-d.2` printed `[PASS] starter and stress sims match autoloop/m2-wave-d.2 byte for byte, exit codes included`.
    - Case 7: `git -C W diff autoloop/m2-wave-d.2 --numstat -- pkg cmd` printed exactly the 5 expected lines (51 0, 33 9, 6 17, 1 1, 11 0).
    - Case 8: `go -C W tool task check` passed (fmt-check, vet, all tests, full lint); `go -C W tool task sim` passed on starter decks.
  - Review note: sim outputs remain byte-identical while non-test imports of `ability` in `pkg/sim` are eliminated. Ready for lead-architect review.
