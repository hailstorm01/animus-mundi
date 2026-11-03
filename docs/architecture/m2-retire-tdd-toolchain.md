# Handoff to `workflow-steward`: retire the TDD path

- **From:** `lead-architect`, structural review 7 of `main` at `4ee3be2` | **Date:** 2026-11-03 | **Status:** skills, routing and docs applied by `workflow-steward`, 2026-11-03. The user's answers: the skills go; M3 plans with a design note, not a TDD (`pm-lead` reconciles the summary); `docs/tdd/` stays, frozen. The code is directive #103, which corrects the test list below and deletes this file when it lands.

## The decision

The user, 2026-11-03: *"remove refs to the TDD in the toolchain, because that's a dead path and is bloat."*

This replaces finding 10 of the review, which proposed merging the duplicate TDD parsers in `linter/tdd`, `tddslice` and `whatsup.go`. The code goes away, so there is nothing to merge.

The lead's reasons for agreeing the path is unused: the only TDD, `docs/tdd/03_Card_Compiler_TDD.md`, is parked, and M2 work runs through `.agents/tasks`, not GitHub issues. Directive #6 (move the queue to GitHub Issues) is written around `gh`, not `tddslice`, so it does not depend on this code.

## Ask the user first

These go beyond the sentence above, or need sign-off under your §2.

1. **Do the skills go too?** The user named "the toolchain". The lead reads that as including skill text, because otherwise skills would tell agents to run `cardtown lint tdd` and `cardtown tdd slice` after they are gone. Removing Runbooks B and C and the lead's `docs/tdd/` write permission is a permission change, so it needs the user's explicit yes.
2. **Does M3 still plan a TDD?** [`PRODUCT_SUMMARY.md:248`](../PRODUCT_SUMMARY.md) says M3 will "revise or replace TDD 03", line 280 lists it as parked for M3, and [`docs/AGENTS.md:14`](../AGENTS.md) says not to slice it before M3. This is a scope decision, so it is `pm-lead`'s, not yours or the lead's. If M3 does want one, it could be written without this tooling, and the tooling stays in git history.
3. **What happens to `docs/tdd/`?** Nobody asked to delete `03_Card_Compiler_TDD.md` or `docs/tdd/README.md`, and after this change no tool reads them. Keep them, or delete them?

## Yours to edit (Runbook B)

| File | What refers to the TDD path |
| :--- | :--- |
| `.agents/skills/lead-architect/SKILL.md` | Front-matter description; §1 Full TDD Mode; §2 the `docs/tdd/` permission, the commit template, the Two-Stage Gate and the TDD quality gate; §3 triggers 2 and 3 and the staleness clause in trigger 5; Runbook A step 6; Runbooks B and C; §5 template; the TDD wording in §6 |
| `.agents/skills/cardtown-linter/SKILL.md` | 9 lines on the TDD lint |
| `.agents/rules/AGENTS.md` | Lines 10, 14 and 21 |
| `docs/AGENTS.md` | Lines 14, 20 and 23 |
| `README.md` | Lines 41, 42, 64 and 67 |

**Keep the letters D and E on the lead's runbooks.** The board, the task files and the wave runner cite `Runbook D`, `D2.0`, `D7` and `Runbook E` all over, so renumbering them would break those references.

## Not yours: route it (Runbook C)

`pkg/` and `cmd/` are outside your fences, so record the code removal as a board directive for the build loop, as #90 did for `cardtown audit` (T-0124). The next free number is one after `Last directive:` on `BOARD.md`. Suggested status: `open`. The lead shapes the task. The scope below edits five non-test files, over the D2 size rule's four, so expect a split. T-0124 needed tier `sonnet` rather than `flash` because deleting files is off the headless allow-list.

**Scope, checked against `4ee3be2`:**

- **Delete:** `pkg/toolchain/linter/tdd` and `pkg/toolchain/tddslice`, with their tests.
- **`cmd/cardtown/lint.go`:** the import (line 18), the `tdd` subcommand (34-39), the `--tdd` flag (78) and `runTdd` with everything it gates, including the changed-file detection (172, 200, 214) and the runner (247-254). Line 224's message also names `docs/tdd/`.
- **`cmd/cardtown/main.go`:** the `tdd` case (43), the usage text (68, 70), and the aliases only `tdd slice` uses: `dryrun`, `updateexisting`, `showbodies`, `tddfile`.
- **`cmd/cardtown/tools.go`:** the header comment, the import (12) and `runTdd` (27-69).
- **`pkg/toolchain/docs/whatsup.go`:** the capability lines (92-93, 102-103) and the TDD block at the end of `ShowState` (from line 193).
- **`pkg/toolchain/git/git.go`:** `CommitExists` and `GetNewerCommits`, whose only callers are `tdd.go:277-278`. `Output` stays, because `Lines` calls it.
- **Existing tests the task must name (D2.4):** `cmd/cardtown/args_test.go` (the tables at lines 21, 50, 66-67 and 86), `cmd/cardtown/tools_test.go:36` and `pkg/toolchain/docs/whatsup_test.go:37`.

**Suggested acceptance cases:**

- `go tool task check` passes, and `go tool task sim-compare` passes, since no engine code changes.
- `cardtown tdd` is an unknown command, and `cardtown lint --tdd` is an unknown flag.
- `git grep -i -w tdd -- pkg cmd` finds nothing.

Directive #100 leaves `linter/tdd` and `tddslice` alone on purpose, so the two can land in either order.

## Leave as history

Task files and run files, `docs/design/outcomes/`, ADR-0001 (lines 8 and 18), frozen FRD 03, and your own log entries. They record what happened and should not be rewritten. The other `PRODUCT_SUMMARY.md` mentions (lines 37, 78, 145, 149 and 227) go with question 2 to `pm-lead`.

## Prompts for the user

- `@workflow-steward apply docs/architecture/m2-retire-tdd-toolchain.md`
- `@pm-lead the user is retiring the TDD path (docs/architecture/m2-retire-tdd-toolchain.md, question 2): reconcile the M3 plan in PRODUCT_SUMMARY.md`
