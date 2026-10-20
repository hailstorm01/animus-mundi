---
id: T-0002
title: Scaffold the ability and cards packages
kind: harness
tier: sonnet
wave: A1
status: in-progress
branch: m2/T-0002-scaffold
depends_on: [T-0001]
claims: [package:ability, type:ability.Def, package:cards, var:cards.Defs, var:cards.NotYetExecutable, field:Card.Ability, loader:def-attach, test:TestDefs_MatchRulesText, test:TestStarterSet_Executable, helper:scene.card, linter:fences-scope]
cards: []
requirements: []
---

## Goal
The two new packages from the design note exist and are wired in: `pkg/engine/ability` (the shape vocabulary, holding only `Def.Text` for now) and `pkg/engine/cards` (card definitions plus the allowlist of starter cards that cannot run yet). `LoadCards` attaches each card's `Def` and removes the lines it covers from `Unsupported`. Two coverage tests keep the allowlist accurate, and `s.card(name)` gives scenario tests the real card. Nothing that runs today changes behavior: the same 9 starter cards stay executable.

## Inputs
- Design note: `docs/architecture/m2-ability-model.md`, sections Boundaries, Test Plan, Deliverables item 1.
- ADR-0004 (`docs/architecture/decisions/0004-abilities-as-data-in-go.md`): definitions are Go values in `pkg/engine/cards`.
- Current loader: `pkg/engine/card.go` (`LoadCards`, `readCards`, `Supported`). Existing callers of `LoadCards` (`cmd/cardtown/main.go`, `playout_test.go`) keep the same signature.
- Fences linter: `pkg/toolchain/linter/fences/fences.go` and its tests (T-0001).
- Board directives applied here: #1 (the allowlist lives outside the protected test), #3 (tier sonnet, plus the two T-0001 follow-ups).
- Rulings that apply: none.

### Definitions the tests must pin
- **Import direction:** `engine` → `cards` → `ability`. `ability` imports nothing from the repository. `cards` imports only `ability`.
- **`ability.Def`:** a struct with one field for now, `Text []string`, the exact rules-text lines the definition implements. Shape tasks add the other fields named in the design note. The package doc comment says: plain data, no game logic, no card names (fenced), and new types only through a `shape` task.
- **`cards.Defs`:** `map[string]*ability.Def` keyed by card name. Empty in this task.
- **`cards.NotYetExecutable`:** a sorted `[]string` of every starter-set card the kernel cannot execute yet. Today it holds exactly 73 names. Card tasks remove names from it, and M2 is done when it is empty.
- **Starter-set scope:** every card whose `faction` is `ANIMUS` or `PATH`, except the Headquarters *Rift* and *Encampment* (moved to M3). That is 82 cards: all non-Headquarters Animus and Path cards plus *Animoceps Rager* and *Eternal Watchman*. Nine of them are executable today: Chainbreaker Sneak, Despoiler Spawn, Flickering Memory, Hammerhand Brigade, Improvised Bastion, Sigiled Stoneform, Spoil Forest Amalgam, Spoil Forest Haunt, Trailblazer.
- **Ability text:** the non-keyword lines of a card's rules text, after the loader's existing split and trim, which are the lines that land in `Unsupported` today. A `Def.Text` line must equal one of them exactly: same case, punctuation and apostrophes. Each `Def.Text` entry removes one occurrence from `Unsupported`. Standalone keyword lines are never ability text.
- **Loader split:** `LoadCards(dir)` keeps its signature and calls an unexported `loadCards(dir string, defs map[string]*ability.Def)` with `cards.Defs`. Tests call `loadCards` with synthetic defs. A def whose name is not in the loaded sheets is ignored by the loader, because fixture directories hold subsets. `TestDefs_MatchRulesText` catches it on the real data instead. Defs are processed in sorted name order, so the first error is stable.
- **`Supported()` is unchanged:** Maneuvers stay unsupported even when fully covered. The task that adds `Def.Effect` changes that.
- **Fence scope:** every non-test `.go` file under `pkg/engine/`, at any depth, except files under `pkg/engine/cards/`. Violations are reported with the path relative to `pkg/engine/`, using forward slashes (`x.go:4`, `ability/z.go:3`), and sorted by path, then line.

## Acceptance Cases (harness tasks only; written by the lead)
**Loader** (`loadCards` on a fixture `data/cards` in `t.TempDir()`, with headers copied from the real sheets). The fixture unit prints `Ward`, then line A, then line B. The fixture Maneuver prints one line M.
1. No defs → the unit's `Ability` is nil and `Unsupported` is `[A, B]` in printed order. `Ward` is parsed as a keyword as before.
2. A def with `Text: [A]` for the unit → `Ability` is that same pointer, and `Unsupported` is `[B]`.
3. A def with `Text: [A, B]` → `Unsupported` is empty and `Supported()` is true.
4. A def with `Text: [M]` for the Maneuver → `Unsupported` is empty but `Supported()` is false.
5. A def line that is not on the card (A with one word changed) → `loadCards` returns a nil map and an error naming the card and the line.
6. A def line that differs from A only in case or in its apostrophe (`'` vs `’`) → error, as in case 5.
7. A def line `Ward` (a keyword line on the card) → error, as in case 5.
8. A def with `Text: [A, A]` on a card that prints A once → error.
9. A card that prints A twice, with a def `Text: [A]` → one A stays in `Unsupported`.
10. A def keyed by a name that is not in the fixture → no error, and the other cards load normally.
11. Two defs that both fail → the error names the card that comes first alphabetically.

**Coverage tests** (`pkg/engine/defs_test.go`). Each test is a thin wrapper that runs a checker on the real data and fails on any problem it returns. The checkers are tested with synthetic inputs.
12. `TestStarterSet_Executable`: on the real repository, the in-scope cards that are not `Supported()` equal `cards.NotYetExecutable` exactly, so it passes with today's 73 names.
13. `checkStarterSet(cards, allow)`, given an in-scope unsupported card that is not in `allow` → one problem naming the card and saying it must become executable or go back on the list.
14. `checkStarterSet`, given an allowlisted card that is now `Supported()` → one problem saying to remove it from `NotYetExecutable`.
15. `checkStarterSet`, given an allowlisted name that is out of scope (*Rift*, a `SOULLESS` card, or a name that matches no card) → one problem naming it.
16. `checkStarterSet`, given a duplicate entry or an unsorted list → one problem.
17. `checkStarterSet`, given an unsupported `SOULLESS` card that is not listed → no problem.
18. `TestDefs_MatchRulesText`: on the real repository, every key of `cards.Defs` is a loaded card and every def has a non-empty `Text`. It passes trivially today.
19. `checkDefs(cards, defs)`, given a def keyed by an unknown name → one problem naming it. Given a def with empty `Text` → one problem. Given a valid def → no problems.

**Scenario helper** (`pkg/engine/scenario_test.go`)
20. `realCard("Chainbreaker Sneak")` returns the CSV card (`TypeUnit`, faction `PATH`, `Supported()` true). The repository cards are loaded once per test binary.
21. Two calls to `realCard` with the same name return different pointers: changing one copy's `Power` leaves the other unchanged (shallow copy).
22. `realCard("No Such Card")` returns an error naming the card. `s.card(name)` wraps `realCard` and calls `t.Fatalf` with that error.

**Fences** (`pkg/toolchain/linter/fences/fences_test.go`)
23. A card name in `pkg/engine/ability/z.go` → fails, and the `[ERROR]` line names `ability/z.go:<line>`.
24. A card name in a new subfolder `pkg/engine/foo/w.go` → fails (the scope is all of `pkg/engine` except `cards`).
25. A card name in `pkg/engine/cards/sub/w.go` or in `pkg/engine/ability/z_test.go` → passes. The existing case 11 passes unchanged.
26. The fixture without `docs/frd/keyword_interaction_matrix.md` → `Lint` returns a non-nil error whose message contains `keyword_interaction_matrix.md`. Without the survey → the message contains `m2_starter_set_ability_survey.md`. (T-0001 follow-up a)
27. `TestFences_Case02_CardNameInCode` also asserts exactly one `[ERROR]` line. This is the only edit allowed to an existing fences test. (T-0001 follow-up b)
28. Violations in `x.go` and `ability/z.go` → both reported, `ability/z.go` first (sorted by path).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
- Harness task. Write one test per acceptance case: loader cases in `pkg/engine/card_test.go` (new) as `TestLoadCards_<Case>`; coverage cases in `pkg/engine/defs_test.go` (new); helper cases in `pkg/engine/scenario_test.go` or `card_test.go`; fences cases appended to `fences_test.go` as `TestFences_Case21_…` onward. Commit them first as `test(T-0002): …` with compiling stubs, so every new test fails there. Case 12 is the exception: it may pass at the stub commit if the stub list is already right. Case 27 is the other exception, because it tightens a test that already passes.
- After merge, `TestStarterSet_Executable`, `TestDefs_MatchRulesText`, `checkStarterSet` and `checkDefs` are protected. Later tasks change only `cards.NotYetExecutable` and `cards.Defs`, never `defs_test.go`.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- `func realCard(name string) (*Card, error)`: a shallow copy of the named card from `../../data/cards`, loaded once.
- `func (s *scene) card(name string) *Card`: `realCard` or `t.Fatalf`.

## Allowed Files
- `pkg/engine/ability/ability.go` (new: package doc and `Def`)
- `pkg/engine/cards/cards.go` (new: package doc, `Defs`, `NotYetExecutable`)
- `pkg/engine/card.go` (`Card.Ability`, `loadCards`, text attachment)
- `pkg/toolchain/linter/fences/fences.go` (scope, path reporting, missing-file errors)
- Tests: `pkg/engine/card_test.go` (new), `pkg/engine/defs_test.go` (new), `pkg/engine/scenario_test.go`, `pkg/toolchain/linter/fences/fences_test.go` (new cases, plus the case-27 assertion)

## Done When
- [ ] `go test ./pkg/engine/... ./pkg/toolchain/linter/fences/` passes, with one test per acceptance case
- [ ] `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean
- [ ] `go test ./pkg/engine -run TestLoadCardsFromRepositoryData -v` still logs 29 of 169 cards executable
- [ ] `go doc ./pkg/engine/ability` prints the package doc and `Def`
- [ ] `go run ./cmd/cardtown lint fences` and `go run ./cmd/cardtown lint --all` exit 0
- [ ] Adding `// Rift` to an untracked scratch file in `pkg/engine/ability/` makes `go run ./cmd/cardtown lint` exit 1. Delete the file afterwards.

## Log
- 2026-10-19 lead-architect: Wrote the task, applying board directives #1 and #3. Harness kind on tier sonnet, so it skips `tests` and goes straight to `ready`. Grounding: 82 in-scope cards, 9 executable (listed above), so the allowlist starts at 73. `isFencesPath` in `cmd/cardtown/main.go` already matches any `pkg/engine/` `.go` path, so smart detection needs no change. `gofmt -l pkg cmd` is already clean, so T-0001 follow-up (3) is closed. Per directive #5, the user reviews this task before `kernel-engineer` is dispatched.
- 2026-10-19 lead-architect: The user reviewed the task and approved dispatch (directive #5 applied). Dispatched `kernel-engineer` (Sonnet subagent).
