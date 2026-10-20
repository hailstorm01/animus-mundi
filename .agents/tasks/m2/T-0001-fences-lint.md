---
id: T-0001
title: Add a fences check to cardtown lint
kind: harness
tier: flash
wave: A1
status: ready
branch: m2/T-0001-fences-lint
depends_on: []
claims: [linter:fences, cli:lint-fences]
cards: []
requirements: []
---

## Goal
`cardtown lint fences` fails when engine code contains a card name, or when an `OPEN(…)` marker in engine code names no open keyword-matrix or survey item. It also runs under `cardtown lint --all` and in the pre-commit hook, so review checklist item D5.2 is enforced by a tool instead of by eye.

## Inputs
- Fence rules: `.agents/skills/lead-architect/SKILL.md` Runbook D, D5 item 2; `.agents/skills/kernel-engineer/SKILL.md` §1 ("No card-name logic").
- Card names: the `name` column of every `data/cards/*.csv`. Find the column by header, not by position (the two sheet schemas differ).
- Matrix items: `docs/frd/keyword_interaction_matrix.md`. Open items are the `| KIM-… |` rows in the `## Open Items` table. Ruled items are the rows in the `## Rulings` table.
- Survey items: `docs/design/m2_starter_set_ability_survey.md`, the table under `## 4. Ruling Questions`. The row `| **R5** | … |` is item `SVY-R5`, and `**R4a**` is `SVY-R4a`. An item is ruled when its last cell contains `Ruled`; otherwise it is open.
- Sibling to copy for structure, output style and test layout: `pkg/toolchain/linter/rules/` (`Analyze` returns a report, `Lint` prints it and returns `(ok, err)`).
- Rulings that apply: none. Design note: not needed (harness task).

### Definitions the tests must pin
- **Scope:** non-test `.go` files directly in `pkg/engine/`. Subdirectories are out of scope: the M2 design note decides where card definitions live, and the lead widens the scope then.
- **Normalizing text into words:** lowercase, delete apostrophes (`'` and `’`), then split on every run of characters that are not letters or digits. Go identifiers are first also split at lower→upper case changes and at letter↔digit changes (`chainbreakerPhalanx` → `chainbreaker phalanx`, `noAh5` → `no ah 5`).
- **Contains a card name:** the card name's word sequence appears as a contiguous run of whole words in a line's normalized words. The line includes string literals, identifiers and comments. Whole words only, so `drift` does not match the card *Rift*.
- **Rules-object exemption:** a card name that FRD 01 also defines as a game object is allowed. The exemptions live in one package-level map from name to the FR ID that defines it. It holds exactly one entry: `"Scratch Company": "FR-ZON-005"`.
- **Marker:** text matching `OPEN\(([A-Z]+-[A-Z]+\d+[a-z]?)\)`. Placeholder text like `OPEN(KIM-...)` in `doc.go` does not match and is ignored.

## Acceptance Cases (harness tasks only; written by the lead)
Each case except the last uses a fixture repo built in `t.TempDir()` that has `data/cards/` (at least one units sheet and one maneuvers sheet), `pkg/engine/`, the matrix file and the survey file.

1. Engine files contain no card names and no markers → `Lint` returns `(true, nil)` and prints `[PASS]`.
2. `pkg/engine/x.go` has `if c.Name == "Chainbreaker Phalanx" {` → `(false, nil)`, with one `[ERROR]` line naming `x.go:<line>` and `Chainbreaker Phalanx`.
3. A card name inside a comment (`// mirrors Steel Rain`) → fails, same as case 2.
4. A card name as a camelCase identifier (`func chainbreakerPhalanxBonus()`) → fails.
5. A card name in a different case (`"steel rain"`) → fails.
6. An apostrophe name written without its apostrophe (`"Chainbreakers Tenacity"` for *Chainbreaker's Tenacity*) → fails.
7. A single-word card name inside a longer word (`drift := 0` with a card named *Rift*) → passes.
8. A single-word card name as a whole word (`case "rift":`) → fails.
9. `"Scratch Company"` in a string literal, a comment and `HasSubtype("Scratch Company")` → passes (rules-object exemption).
10. A card name in `pkg/engine/x_test.go` → passes (test files are out of scope).
11. A card name in `pkg/engine/cards/y.go` → passes (subdirectories are out of scope).
12. `// OPEN(KIM-B07)`, where KIM-B07 is in `## Open Items` → passes.
13. `// OPEN(KIM-B01)`, where KIM-B01 is only in `## Rulings` → fails, and the error says the item is ruled and the marker should be removed.
14. `// OPEN(KIM-Z99)`, which appears nowhere → fails, and the error says the item is not in the matrix.
15. `// OPEN(SVY-R7)` with an open survey row `**R7**` → passes. `// OPEN(SVY-R4a)` where `**R4a**`'s status is `**Ruled: yes.**` → fails as ruled.
16. `// OPEN(ABC-X01)` → fails, and the error says markers must name a `KIM-` or `SVY-` item.
17. `// it with an "OPEN(KIM-...)" comment` → passes (not a marker).
18. Two violations in different files → both are reported, in a stable order (file name, then line). The summary counts 2.
19. `data/cards/` is missing → `Lint` returns a non-nil error, not a pass.
20. The real repository (`repoRoot` = `../../../../`, as in `rules_test.go`) → passes. This pins that today's engine is clean once the Scratch Company exemption applies.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
- Harness task: the engineer writes one `TestFences_<Case>` per acceptance case in `pkg/toolchain/linter/fences/fences_test.go` and commits them first as `test(T-0001): …` (they must fail before the implementation commit).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None.

## Allowed Files
- `pkg/toolchain/linter/fences/fences.go` (new: `Analyze`, `Lint`, the exemption map)
- `cmd/cardtown/main.go` (wiring: the `lint fences` subcommand; a `--fences` flag added to `normalizeArgs`; the check included in `--all`; smart and `--staged` detection runs it when `pkg/engine/*.go`, `data/cards/*.csv`, the matrix or the survey changes; the `lint` line in `printUsage`)
- `pkg/toolchain/docs/whatsup.go` (one tool entry, next to `cardtown lint rules`)

## Done When
- [ ] `go test ./pkg/toolchain/linter/fences/` passes, with one test per acceptance case
- [ ] `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean
- [ ] `go run ./cmd/cardtown lint fences` exits 0 on the branch
- [ ] `go run ./cmd/cardtown lint --all` runs the fences check and passes
- [ ] Adding `// Rift` to an untracked scratch file in `pkg/engine/` makes `go run ./cmd/cardtown lint` (smart detection) run the fences check and exit 1; delete the file afterwards

## Log
- 2026-10-19 lead-architect: Wrote the task. Harness kind, so it skips `tests` and goes straight to `ready`. Scanned all 169 card names against `pkg/engine/*.go`: the only hit is Scratch Company (in `actions.go`, `card.go`, `doc.go`, `sim.go`, `stack.go` and `state.go`), which is a rules object under FR-ZON-005, hence the exemption. There are no `OPEN(` markers in engine code today besides the `doc.go` placeholder. Next: `@kernel-engineer T-0001` (tier flash, Antigravity).
