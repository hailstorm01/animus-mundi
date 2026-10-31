---
id: T-0124
title: Remove the retired cardtown audit command and pkg/toolchain/audit (directive #90, harness, sonnet)
kind: harness
tier: sonnet
wave: C
status: ready
branch: m2/T-0124-remove-audit-cli
depends_on: []
claims: [cli:audit command removed, toolchain:audit package removed, whats-up:audit lines removed]
cards: []
requirements: []
---

## Goal
`pm-devils-advocate` and the `docs/audit/` workflow are retired (9ff0182), so the `cardtown audit` command and the `pkg/toolchain/audit` package that served them go too. After this task no code under `cmd/` or `pkg/` mentions the audit workflow: the CLI has no `audit` command or usage line, `cardtown whats-up` lists no audit tools and prints no "Pending audits" block, and the package is deleted. This is directive #90 (a)–(d). No engine code changes.

## Inputs
- Directive #90 in `BOARD.md`. Its import path and line numbers are stale; the current ones are below (the module is `animus-mundi`, not `github.com/card-town/card-town`).
- `pkg/toolchain/audit/audit.go`: the package's only file (no tests, no `doc.go`). Imported only by `cmd/cardtown/tools.go` line 10 and `pkg/toolchain/docs/whatsup.go` line 12 (`git grep -n "toolchain/audit"`).
- `cmd/cardtown/main.go`: the `case "audit":` branch of the command switch (lines 43-44, calling `runAudit`), and the usage line `  audit     Manage PM Devil's Advocate audit lifecycle (status, resolve)` in `printUsage` (line 72).
- `cmd/cardtown/tools.go`: the file comment `// Commands: hook, audit, tdd, whats-up.` (line 1), the `"animus-mundi/pkg/toolchain/audit"` import (line 10), and `runAudit` (lines 28-69). `flag` and `normalizeArgs` stay: `runTdd` and `runWhatsUp` use them.
- `pkg/toolchain/docs/whatsup.go`:
  - the `"animus-mundi/pkg/toolchain/audit"` import (line 12);
  - in `ShowTools`, the four lines for `cardtown audit status` and `cardtown audit resolve` (lines 103-106);
  - `ShowState`'s doc comment (line 114, "displays repository git status, hook status, and pending audits");
  - in `ShowState`, the blank `fmt.Println()`, `fmt.Println("Pending audits:")` and `_ = audit.Status(repoRoot)` (lines 168-170), between the keyword-matrix line and the portfolio staleness check.
- Existing stdout capture to mirror in the new tests: `captureOutput` in `pkg/toolchain/linter/fences/fences_test.go` (lines 94-111). Copy it into each new test file; it is test setup, not a shared helper.

## Acceptance Cases (written by the lead; the engineer turns each into a test)
1. **The usage text has no audit command.** The stdout of `printUsage()` contains no `audit` (case-insensitive), and still has one line each starting with `  play `, `  lint `, `  hook `, `  tdd `, `  sim `, `  replay ` and `  whats-up `. Before the change it contains "audit". New test in `cmd/cardtown/tools_test.go` (for example `TestUsage_ListsNoAuditCommand`).
2. **`whats-up`'s tools section lists no audit tool.** The stdout of `docs.ShowTools(<repo root>)` contains no `cardtown audit`, and still contains `cardtown lint`, `cardtown sim`, `cardtown hook install`, `cardtown tdd slice` and `cardtown whats-up`. Before the change it contains `cardtown audit status`. New test in `pkg/toolchain/docs/whatsup_test.go` (for example `TestShowTools_ListsNoAuditTool`); the repo root is `../../..` from the package directory.
3. **`whats-up`'s state section has no audit block.** The stdout of `docs.ShowState(<repo root>)` has no line `Pending audits:`, and still has a line starting `Git: branch ` and a line starting `Pre-commit hook: `. The portfolio line (`Portfolio roadmap (docs/PRODUCT_SUMMARY.md): …`, when printed) follows the keyword-matrix line with no blank line between them: delete the `fmt.Println()` together with the "Pending audits:" line. Assert only on lines, not on "audit" anywhere: while the deletion is uncommitted, the "Uncommitted changes" list can name `pkg/toolchain/audit/audit.go`. Same new test file (for example `TestShowState_PrintsNoPendingAudits`).
4. **The package is gone and nothing imports it.** `pkg/toolchain/audit/` is deleted (`git rm pkg/toolchain/audit/audit.go`), and `go build ./...` and `go vet ./...` pass. Checked by `go tool task check` and at review against the diff, not by a new test.
5. **`cardtown audit` is an unknown command.** With the `case "audit":` branch gone, `cardtown audit` falls to the switch's `default` (prints `Unknown command "audit"` and the usage, exits 1). Checked at review against the diff, not by a test (it calls `os.Exit`).
6. **Comments say what is left.** `tools.go`'s file comment is `// Commands: hook, tdd, whats-up.`. `ShowState`'s doc comment no longer mentions audits: `// ShowState displays repository git status, hook status and the docs workflow's live state.`. Checked at review.
7. **`whats-up` still runs.** `go run ./cmd/cardtown whats-up` exits 0 and its output has no `audit` command or "Pending audits" block.
8. **The sims are unchanged.** No engine code changes: `go tool task sim-compare -- autoloop/m2-wave-c` is byte-identical.

## Allowed Files
- `pkg/toolchain/audit/audit.go` (deleted)
- `cmd/cardtown/main.go`
- `cmd/cardtown/tools.go`
- `pkg/toolchain/docs/whatsup.go`

Tests: `cmd/cardtown/tools_test.go` and `pkg/toolchain/docs/whatsup_test.go`, both new files, new tests only. No existing test is edited.

## Done When
- [ ] Cases 1-3 have tests, each of which failed before the implementation commit; cases 4-6 are checked at review, case 7 by the command below, case 8 by `sim-compare`.
- [ ] `go test ./cmd/cardtown`
- [ ] `go test ./pkg/toolchain/...`
- [ ] `go run ./cmd/cardtown whats-up`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-c` passes (starter and stress sims byte-identical to the run branch)

## Log
- 2026-10-31 lead-architect: wave C run, session 10. No `main` commits to sync (`git log autoloop/m2-wave-c..main` is empty). No `open: before` directive. #90 is `open`, so D2.0 applies it before the next wave C card task is written; it is this prompt's task. Waiting triggers #8 (no `deck.go` change here), #15, #58, #59 (2), #61 (fifth question), #71, #88 (distinctness) not fired. Duplicates searched: removal only, nothing added beyond the tests' copy of `captureOutput` (test setup). Every file is in the package that owns it (`cmd/cardtown` the CLI, `pkg/toolchain/docs` the whats-up snapshot). #90's import path (`github.com/card-town/card-town/...`) and its `whatsup.go` line numbers (53-54, 142) are stale; the Inputs give today's, and its "two capability lines" are four printed lines (two entries with a usage line each). Tier `sonnet`, not the `flash` #90 names: case 4 deletes a file, and the headless allow-list has no command that can (`git rm` and `Remove-Item` are refused; `kernel-engineer` §3 step 5), so a Flash run would end at that step. The work is otherwise small: four files, deletions only, no engine change. Harness, so `ready`; branch `m2/T-0124-remove-audit-cli` from `autoloop/m2-wave-c`.
