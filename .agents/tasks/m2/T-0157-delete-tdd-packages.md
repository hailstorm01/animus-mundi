---
id: T-0157
title: Delete the TDD packages and their git helpers, and finish directive #103 (harness, sonnet)
kind: harness
tier: sonnet
wave: none (board directive #103, after wave C.1)
status: queued
branch: m2/T-0157-delete-tdd-packages
depends_on: [T-0156]
claims: [toolchain:linter/tdd package removed, toolchain:tddslice package removed, git:CommitExists removed, git:GetNewerCommits removed]
cards: []
requirements: []
---

## Goal
Nothing under `cmd/` or `pkg/` mentions the TDD path: #103's acceptance (c), `git grep -i tdd -- pkg cmd`, finds nothing. This is directive #103 (a) and (f), and its last task. No engine code changes.

## Inputs
Written against the code of the day before this task leaves `queued` (D2.4). What it covers:
- Delete `pkg/toolchain/linter/tdd` (`tdd.go`, `tdd_test.go`) and `pkg/toolchain/tddslice` (`slice.go`). Deleting files is off the headless allow-list, hence tier `sonnet`, as T-0124 (#90).
- `pkg/toolchain/git/git.go`: remove `CommitExists` and `GetNewerCommits`, whose only callers are `tdd.go:277-278`. `Output` stays, because `Lines` calls it.
- Acceptance: `go tool task check` and `go tool task sim-compare` pass, and `git grep -i tdd -- pkg cmd` finds nothing.

**For the lead, in the merge commit (not the engineer's work):**
- Delete `docs/architecture/m2-retire-tdd-toolchain.md`, as #103 says.
- Remove the link to it from #100's row in `archive/directives/directive_archive.md`, and from #103's own row when it is archived, so neither row dangles.
- Report to the user what the lead doesn't own and is left stale by this task: `docs/tdd/README.md:4` says the commands "are being removed" (frozen doc), and `.agents/steward-log.md:12` lists the code as a part still left.

## Log
- 2026-11-04, lead: split off T-0156 (D2.4). `queued` until T-0156 is `done`.
