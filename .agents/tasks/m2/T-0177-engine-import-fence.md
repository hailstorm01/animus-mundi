---
id: T-0177
title: The fences check the engine's imports (directive #115, harness, flash)
kind: harness
tier: flash
wave: D.1 (board directive #115, structural review 8, finding 14)
status: ready
branch: m2/autoloop-wave-d.1/T-0177-engine-import-fence
depends_on: [T-0176]
claims: [fence:pkg/engine imports, test:TestFences_Case34 engine fixture, test:TestFences_Case39_EngineImports]
cards: []
requirements: []
---

## Goal
`cardtown lint fences` checks `pkg/engine`'s imports: the engine may import `pkg/engine/ability` and `pkg/engine/cards`, and nothing else in the module. So it can no longer import `pkg/toolchain` or a client package unnoticed, which is what the design note already says the fence does.

## Inputs
- Board directive #115 (structural review 8, approved by the user, 2026-11-05).
- Design note: `docs/architecture/m2-ability-model.md`, Boundaries ("`cardtown lint fences` checks these import directions") and Deliverable 6g (wave D.1). Both already describe the result, so the note doesn't change.
- Reuse search (D2.3): `allowedImports` and `checkImports` are the one import check, and this task adds one row to that table. `importFiles` in `Analyze` already holds `pkg/engine`'s non-test files, so nothing else needs to read them.
- The engine's real non-test imports, all allowed by the new row: `pkg/engine/ability` everywhere, and `pkg/engine/cards` only in `cardload.go`.
- **An existing test pins the gap.** `TestFences_Case34_AllowedImportsPass` has a fixture where `pkg/engine/x.go` imports `pkg/sim`, and it expects a pass. With the new row that fixture is a violation, so edit 1 names the test and deletes that one line (D2.4). The test's own subject, a package outside the table, stays covered by its `pkg/foo/x.go` line.
- Kept on purpose: `allowedImports`' comment "Packages not listed are not checked" stays true, and #115's fix is the one row, so unlisted packages stay unchecked.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 5b6edae. With only the test edits, `go test ./pkg/toolchain/linter/fences -run 'TestFences_Case3[49]'` exits 1, with only Case 39 failing, on `expected failure, got true`. With the row added: `gofmt -l` clean, `go test ./pkg/toolchain/linter/fences` exits 0, `cardtown lint fences` exits 0, `go tool task check` exits 0, `go tool task sim` exits 0, and `go tool task sim-compare -- autoloop/m2-wave-d.1` printed `[PASS]`. Every expected output in the cases below is from that dry run.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are from the files as they stand at 5b6edae.

### 1. `pkg/toolchain/linter/fences/fences_test.go`, in the `test(T-0177)` commit
1. In `TestFences_Case34_AllowedImportsPass`, delete line 938, with nothing else changed:
   ```go
   	writePkgFile(t, repo, "pkg/engine/x.go", "package engine\n\nimport _ \"animus-mundi/pkg/sim\"\n")
   ```
2. Add to the end of the file, after `TestFences_Case38_UnreadableImportsMessage`'s closing brace and one blank line, ending in one line break:
   ```go
   // 39. pkg/engine may import its two core packages and nothing else in the module.
   func TestFences_Case39_EngineImports(t *testing.T) {
   	repo := setupFixtureRepo(t)
   	writePkgFile(t, repo, "pkg/engine/x.go", "package engine\n\nimport _ \"animus-mundi/pkg/toolchain/git\"\n")
   	writePkgFile(t, repo, "pkg/engine/y.go", "package engine\n\nimport _ \"animus-mundi/pkg/sim\"\n")
   	writePkgFile(t, repo, "pkg/engine/z.go", "package engine\n\nimport (\n\t_ \"animus-mundi/pkg/engine/ability\"\n\t_ \"animus-mundi/pkg/engine/cards\"\n)\n")
   	out := expectFail(t, repo)
   	if strings.Count(out, "[ERROR]") != 2 {
   		t.Fatalf("expected two errors, got:\n%s", out)
   	}
   	rule := "pkg/engine may import only pkg/engine/ability, pkg/engine/cards"
   	for _, want := range []string{
   		"pkg/engine/x.go:3: imports animus-mundi/pkg/toolchain/git; " + rule,
   		"pkg/engine/y.go:3: imports animus-mundi/pkg/sim; " + rule,
   	} {
   		if !strings.Contains(out, "[ERROR] "+want) {
   			t.Fatalf("expected %q, got:\n%s", want, out)
   		}
   	}
   }
   ```

Directive cases (a) and (b) share this one test, because a separate test of an allowed import would pass before the row exists: `x.go` is (a), `z.go` is (b), and the count of two errors is what says `z.go` is not a violation.

### 2. `pkg/toolchain/linter/fences/fences.go`, in the `feat(toolchain)` commit
Add this line right before line 37, `"pkg/engine/ability": {},`:
```go
	"pkg/engine":         {"pkg/engine/ability", "pkg/engine/cards"},
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
Edit 1 is the whole of the test commit: commit it as `test(T-0177): the fences check the engine's imports`, run case 1, then make edit 2. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Fails before the row** (after the test commit, before edit 2): `go -C W test ./pkg/toolchain/linter/fences -run TestFences_Case39_EngineImports` exits 1, and its output has `expected failure, got true`.
2. **Passes after it** (#115 (a), (b)): `go -C W test ./pkg/toolchain/linter/fences` exits 0.
3. **The real repository passes** (#115 (c)): `go -C W run ./cmd/cardtown lint fences` exits 0.
4. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.1 --numstat -- pkg cmd` prints exactly these lines: `1	0	pkg/toolchain/linter/fences/fences.go` and `21	1	pkg/toolchain/linter/fences/fences_test.go`.
5. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.1 --stat` lists only those two files and this task file.
6. **Sims byte-identical** (#115): `go -C W tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`.
7. `gofmt -l W/pkg/toolchain` prints nothing, and `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/toolchain/linter/fences/fences.go`
- `pkg/toolchain/linter/fences/fences_test.go` (edit 1 only)

## Done When
- [ ] Cases 1–7 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-05, lead-architect (wave D.1 run, session 8): no `main` commits to sync. Wrote this task from board directive #115, next in board order. Not `WAVE-COMPLETE`: #115–#119 and #121 are `open: before the next Deliverables task`. #122 is `open: after Wave D.1 (#121), before Wave E`, so it doesn't apply yet. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Searched for existing equivalents: the finds are in Inputs. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120: no such card). Next: `kernel-engineer` on Flash.
