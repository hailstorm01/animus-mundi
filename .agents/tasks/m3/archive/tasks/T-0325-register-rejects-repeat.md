---
id: T-0325
title: "One `register` copies each definition file into `Defs` and refuses a repeated card (directive #192; wave A.1 harness, flash)"
kind: harness
tier: flash
wave: A.1 (board directive #192, design note Deliverable 5a)
status: done
branch: m3/autoloop-wave-a.1/T-0325-register-rejects-repeat
depends_on: [T-0324]
claims: [func:cards.register]
cards: []
requirements: []
---

## Goal
Each of the five definition files in `pkg/engine/cards` copies its map into `Defs` with its own `init` loop. Go rejects a repeated key inside one map literal but not across files, so a card defined in two files compiles today, and whichever `init` runs last wins, unseen by `checkDefs` and the loader. After this task, one function in `cards.go`, `register(defs map[string]*ability.Def)`, does the copy and panics, naming the card, when a name is already in `Defs`. Each file's `init` is one `register(<faction>Defs)` call. `Defs` holds the same entries, one test is added, and the sims are byte-identical.

## Inputs
- Board directive #192 (structural review 16, finding 3), approved by the user 2026-11-26, `open: before the next Deliverables task`. Design note `docs/architecture/m3-full-catalog.md` Deliverable 5a. Directives #190 (T-0323) and #191 (T-0324) are `applied`, so #192 is next in board order.
- The five copies, at `autoloop/m3-wave-a.1` 342d2e85, each lines 5-9 of its file and identical except for the map's name:
  ```go
  func init() {
  	for name, def := range animusDefs {
  		Defs[name] = def
  	}
  }
  ```
  in `animus.go` (`animusDefs`), `concordia.go` (`concordiaDefs`), `soulless.go` (`soullessDefs`), `headquarters.go` (`headquartersDefs`) and `path.go` (`pathDefs`). `Defs` is declared at [cards.go:9](../../../../../pkg/engine/cards/cards.go).
- `pkg/engine/cards` has no test file yet. The new test goes in `pkg/engine/cards/cards_test.go`, in package `cards` (not `cards_test`), because `register` is unexported. The fence linter scans only non-test files, and a test file may import the standard library.
- The panic message is built with `+`, so `cards.go` imports nothing new.
- Rulings that apply: none. No behavior changes.
- **Expected output difference:** none. `Defs` holds the same entries, so `go tool task sim-compare -- autoloop/m3-wave-a.1` passes. `go test -list . ./pkg/engine/...` gives one more test name (1725 → 1726).
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 342d2e85: `gofmt -l pkg cmd` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m3-wave-a.1` `[PASS]`, 1726 tests. Two lead mutants: `register` without its check fails `TestDefs_RegisterRejectsRepeat` ("did not panic"); *Animoceps Node* added to `pathDefs` as well as `animusDefs` makes the package panic at init with `cards: Animoceps Node is defined twice`. Resulting blobs (`git hash-object`, first seven characters): `cards.go` 03c997e, `animus.go` 433b184, `concordia.go` eba398c, `soulless.go` 7146150, `headquarters.go` df01498, `path.go` 123281e, `cards_test.go` a97a87b.

## Tests
One new test, `TestDefs_RegisterRejectsRepeat`, in the new file `pkg/engine/cards/cards_test.go` (Acceptance Case 1). No existing test is edited or removed. Every other test must pass unchanged, in particular `pkg/cardcheck`'s `TestDefs_MatchRulesText` and `TestCards_Executable`, which read `Defs`.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
1. **Test, `pkg/engine/cards/cards_test.go`** (new file, the whole file):
   ```go
   package cards

   import (
   	"fmt"
   	"strings"
   	"testing"

   	"animus-mundi/pkg/engine/ability"
   )

   // TestDefs_RegisterRejectsRepeat registers one name that is not a card twice:
   // the second register panics and names it (directive 192).
   func TestDefs_RegisterRejectsRepeat(t *testing.T) {
   	const name = "Not A Card (TestDefs_RegisterRejectsRepeat)"
   	if _, ok := Defs[name]; ok {
   		t.Fatalf("Defs already holds %q before the test", name)
   	}
   	t.Cleanup(func() { delete(Defs, name) })

   	register(map[string]*ability.Def{name: {}})
   	if _, ok := Defs[name]; !ok {
   		t.Fatalf("register did not add %q to Defs", name)
   	}

   	defer func() {
   		r := recover()
   		if r == nil {
   			t.Fatalf("registering %q twice did not panic", name)
   		}
   		if msg := fmt.Sprint(r); !strings.Contains(msg, name) {
   			t.Fatalf("panic %q does not name %q", msg, name)
   		}
   	}()
   	register(map[string]*ability.Def{name: {}})
   }
   ```
   Before Edit 2 this fails to build (`undefined: register`); record that in the Log.
2. **`pkg/engine/cards/cards.go`:** right after the line `var Defs = map[string]*ability.Def{}`, insert a blank line and then:
   ```go
   // register adds one definition file's map to Defs. It panics, naming the card,
   // if a name is already there: Go rejects a repeated key inside one map literal
   // but not across files, so a card defined in two files would otherwise compile,
   // and whichever init ran last would win.
   func register(defs map[string]*ability.Def) {
   	for name, def := range defs {
   		if _, ok := Defs[name]; ok {
   			panic("cards: " + name + " is defined twice")
   		}
   		Defs[name] = def
   	}
   }
   ```
   The blank line that already follows `var Defs` then separates it from the `// deploysOrAttacks …` comment.
3. **The five `init` functions** (`animus.go`, `concordia.go`, `soulless.go`, `headquarters.go`, `path.go`): replace each three-line loop body with one call, so that `animus.go`'s reads
   ```go
   func init() {
   	register(animusDefs)
   }
   ```
   and the others call `register(concordiaDefs)`, `register(soullessDefs)`, `register(headquartersDefs)` and `register(pathDefs)`. Change nothing else in these files. Run `gofmt -w` on the seven files, then check the blobs against the Inputs.

## Commits
1. `test(engine): register refuses a repeated card, for directive 192 (T-0325)` with Edit 1 only.
2. `refactor(engine): one register copies each definition file into Defs and refuses a repeat, for directive 192 (T-0325)` with Edits 2-3.

No double quotes and no `#` in either message.

## Acceptance Cases (harness tasks only; written by the lead)
1. **A repeated name panics and is named:** `TestDefs_RegisterRejectsRepeat` (Edit 1) registers a name that is not a card once, which adds it to `Defs`, then again, which panics with a message containing the name. `t.Cleanup` removes it, so other tests see only real cards. It fails before Edit 2 and passes after: `go test ./pkg/engine/cards/ -run TestDefs_RegisterRejectsRepeat -v`.
2. **No `init` writes `Defs` directly:** `git grep -n -F "Defs[name] = def" -- pkg/engine/cards` finds one line, in `cards.go`, inside `register`. `git grep -n -F "register(" -- pkg/engine/cards` finds eight lines: `func register(` in `cards.go`, one `register(<faction>Defs)` in each of the five definition files, and the two calls in `cards_test.go`.
3. **`Defs` is unchanged:** `go test -list . ./pkg/engine/...` gives 1726 test names, the run branch's 1725 plus `TestDefs_RegisterRejectsRepeat`, and `go test ./...` passes.
4. `git diff --stat autoloop/m3-wave-a.1 HEAD` lists only the seven files under Allowed Files and this task file.
5. `go tool task sim-compare -- autoloop/m3-wave-a.1` passes.
6. `go tool task check` passes.

## Allowed Files
- `pkg/engine/cards/cards.go`, `pkg/engine/cards/animus.go`, `pkg/engine/cards/concordia.go`, `pkg/engine/cards/soulless.go`, `pkg/engine/cards/headquarters.go`, `pkg/engine/cards/path.go`. The five definition files change only in their `init`, so they count as one edit for the D2.4 size rule.
- Test: `pkg/engine/cards/cards_test.go` (new).

## Done When
- [x] Edit 1 is committed as `test(engine): … (T-0325)`, and Edits 2-3 as `refactor(engine): … (T-0325)`
- [x] Cases 1-5 hold, with their commands and results in the Log
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim-compare -- autoloop/m3-wave-a.1` passes (starter and stress sims byte-identical to the run branch)

## Log
- 2026-11-26 lead-architect: Wrote the task from board directive #192, the third of wave A.1's four directives (design note Deliverable 5a). Harness, tier flash: one function and five one-line `init` bodies inside `pkg/engine/cards`, plus one test, with no shape change. The lead dry-ran the edits (Inputs). Status `ready` on `m3/autoloop-wave-a.1/T-0325-register-rejects-repeat`, from `autoloop/m3-wave-a.1`. #193 follows.
- 2026-11-26 kernel-engineer: Implemented directive #192. Added `TestDefs_RegisterRejectsRepeat` in `pkg/engine/cards/cards_test.go` and verified failure (`undefined: register`) prior to implementation, committed as c470cd34. Implemented `register(defs map[string]*ability.Def)` in `pkg/engine/cards/cards.go` which panics with "cards: <name> is defined twice" if a card name already exists in `Defs`, then updated `init` in `animus.go`, `concordia.go`, `soulless.go`, `headquarters.go`, and `path.go` to call `register(...)`. Reused map check semantics with no external dependencies or new types; blob hashes match exact expected hashes from Inputs (`cards.go` 03c997e, `animus.go` 433b184, `concordia.go` eba398c, `soulless.go` 7146150, `headquarters.go` df01498, `path.go` 123281e, `cards_test.go` a97a87b). Committed refactor as 6b7ab3ed. Verified all acceptance cases:
  1. `TestDefs_RegisterRejectsRepeat` in `pkg/engine/cards/cards_test.go` passes (`go test ./pkg/engine/cards/ -run TestDefs_RegisterRejectsRepeat -v`: PASS).
  2. `git grep -n -F "Defs[name] = def" -- pkg/engine/cards` finds exactly 1 line (`cards.go:20`). `git grep -n -F "register(" -- pkg/engine/cards` finds exactly 8 lines (`cards.go:15`, `animus.go:6`, `cards_test.go:20,34`, `concordia.go:6`, `headquarters.go:6`, `path.go:6`, `soulless.go:6`).
  3. `go test -list . ./pkg/engine/...` returned 1726 test names (the run branch's 1725 plus `TestDefs_RegisterRejectsRepeat`); `go test ./...` passed.
  4. `git diff --stat autoloop/m3-wave-a.1 HEAD` shows only the 7 Allowed Files and this task file.
  5. `go tool task sim-compare -- autoloop/m3-wave-a.1` passed: starter and stress sims match byte for byte, exit codes included.
  6. `go tool task check` passed cleanly (gofmt, go vet, all package tests, cardtown lint --all).
  Moved status to `review`.
- 2026-11-26 lead-architect: Review: **approved**, merged into `autoloop/m3-wave-a.1` as 2cf583e9, status `done`. Directive #192 is `applied`.
  - Scope: dd6c18fd and b1b91fd2 change only this file. c470cd34 (`test(T-0325)`) adds only `pkg/engine/cards/cards_test.go`, and fails to build at that commit (`undefined: register`). 6b7ab3ed changes only the six non-test files under Allowed Files, and doesn't touch the test. All seven blobs are the lead's dry-run blobs. `register` copies and refuses; it adds no shape, alias or wrapper, and `Defs[name] = def` survives only inside it.
  - Checks, re-run in a scratch worktree of the task branch: `go test -list . ./pkg/engine/...` gives 1726 tests (1725 plus `TestDefs_RegisterRejectsRepeat`); case 2's greps find one and eight lines. The dry run's two lead mutants, on the same blobs: `register` without its check fails the new test ("did not panic"), and *Animoceps Node* in both `animusDefs` and `pathDefs` panics at init with `cards: Animoceps Node is defined twice`.
  - Gates: `go tool task check` exit 0; `go tool task sim` clean (P0 58.5% / P1 39.5% / draws 2.0%, unchanged); `go tool task sim-compare -- autoloop/m3-wave-a.1` `[PASS]`. `sim-matrix` is not this harness task's gate (D5.5); it runs with the wave's completion check.
