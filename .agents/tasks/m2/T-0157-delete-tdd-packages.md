---
id: T-0157
title: Delete the TDD packages and their git helpers, and finish directive #103 (harness, sonnet)
kind: harness
tier: sonnet
wave: none (board directive #103, after wave C.1)
status: review
branch: m2/T-0157-delete-tdd-packages
depends_on: [T-0156]
claims: [toolchain:linter/tdd package removed, toolchain:tddslice package removed, git:CommitExists removed, git:GetNewerCommits removed]
cards: []
requirements: []
---

## Goal
Nothing under `cmd/` or `pkg/` mentions the TDD path: #103's acceptance (c), `git grep -i tdd -- pkg cmd`, finds nothing. This is directive #103 (a) and (f), and its last task. No engine code changes.

## Inputs
- Board directive #103, and T-0156 (merge 27139e6), which removed every caller under `cmd/` and `pkg/toolchain/docs`.
- What imports what, at `main` `de9cc0a`: `go list` shows only `linter/tdd` (its own external test) and `tddslice` import `linter/tdd`, and nothing imports `tddslice`. Both packages import only the standard library and `toolchain/git`, so `go.mod` and `go.sum` don't change.
- `CommitExists` and `GetNewerCommits` (`pkg/toolchain/git/git.go:69-94`) have one caller, `tdd.go:277-278`, and no test (`git_test.go` tests `Lines`, `Output` and `Combined`). `Output` stays, because `Lines` calls it. `Combined` stays too: it has no caller outside its own tests today, but that predates this task and #103 doesn't name it.
- **No new tests.** Removed packages can't be pinned by a test without naming them, which case 1 forbids. T-0156's `TestUsage_ListsOnlyLiveCommands`, `TestShowTools_ListsOnlyLiveTools` and `TestFlagAliases_NameOnlyLiveFlags` pin the commands, and the cases below check the rest by command.
- **The one test edit (D2.4):** `pkg/toolchain/linter/tdd/tdd_test.go` is deleted, with the package it tests. It lands in the `test(T-0157)` commit. No other test changes.
- The lead dry-ran these exact deletions in a throwaway worktree of `main` at `de9cc0a`: `gofmt -l` clean, `go tool task check` and `go tool task sim` pass, `go tool task sim-compare` printed `[PASS]`, and cases 1-5 hold.

## Edits (lead decisions)
W is your worktree. `git rm` is allowed here (tier `sonnet`; this task exists at that tier because it deletes files, as T-0124 did).

### 1. First commit, `test(T-0157): remove the tests of the deleted linter`
`git -C W rm pkg/toolchain/linter/tdd/tdd_test.go`

### 2. Second commit, `feat(toolchain): delete the TDD packages and their git helpers (T-0157)`
- `git -C W rm pkg/toolchain/linter/tdd/tdd.go pkg/toolchain/tddslice/slice.go`
- `pkg/toolchain/git/git.go`: delete lines 69-94, from `// CommitExists checks if a commit hash exists in git.` through the blank line after `GetNewerCommits`' closing `}`. `// Lines runs git in repoRoot …` then follows `GetStagedFiles`' closing `}` and one blank line. Its imports all stay in use.

## Acceptance Cases (harness tasks only; written by the lead)
Run each command on its own, pointed at W, after both commits. Record each command and its result in the Log. The reviewer re-runs them.
1. **The word is gone** (#103 (c)): `git -C W grep -n -i tdd -- pkg cmd` prints nothing. `git grep` exits 1 when it finds nothing, which is the expected result here.
2. **The packages are gone:** `pkg/toolchain/linter/tdd` and `pkg/toolchain/tddslice` don't exist in W.
3. **The helpers are gone:** `git -C W grep -n -E "CommitExists|GetNewerCommits" -- cmd pkg` prints nothing, and exits 1 as above.
4. **The commands stay unknown** (#103 (b)): `go -C W run ./cmd/cardtown tdd` prints `Unknown command "tdd"` and `exit status 1`, and `go -C W run ./cmd/cardtown lint --tdd` prints `flag provided but not defined: -tdd` and `exit status 2`. Both exit non-zero, which is the expected result here.
5. **`whats-up` still runs:** `go -C W run ./cmd/cardtown whats-up` exits 0.
6. **Only the listed edits:** `git -C W diff main...HEAD --numstat -- . ":!.agents"` prints exactly these four lines, with no `go.mod` or `go.sum`:
   ```text
   0	26	pkg/toolchain/git/git.go
   0	636	pkg/toolchain/linter/tdd/tdd.go
   0	23	pkg/toolchain/linter/tdd/tdd_test.go
   0	555	pkg/toolchain/tddslice/slice.go
   ```
7. **The sims are unchanged** (#103 (a)): `go -C W tool task sim-compare` prints `[PASS]`.
8. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/toolchain/linter/tdd/tdd.go` (deleted)
- `pkg/toolchain/tddslice/slice.go` (deleted)
- `pkg/toolchain/git/git.go`

Tests: `pkg/toolchain/linter/tdd/tdd_test.go` (deleted, in the `test(T-0157)` commit). No other test changes.

**For the lead, at merge (not the engineer's work):**
- Delete `docs/architecture/m2-retire-tdd-toolchain.md`, as #103 says.
- Remove the link to it from #100's row in `archive/directives/directive_archive.md`, and from #103's own row when it is archived, so neither row dangles.
- Report to the user what the lead doesn't own and this task leaves stale: `docs/tdd/README.md:4` says the commands "are being removed" (frozen doc), and `.agents/steward-log.md:12` lists the code as a part still left.

## Done When
- [ ] Cases 1-6 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare` passes (starter and stress sims byte-identical to the integration branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-04, lead: split off T-0156 (D2.4). `queued` until T-0156 is `done`.
- 2026-11-04, lead: T-0156 is `done` (merge 27139e6). Inputs, cases and files written against `main` at `de9cc0a`, `ready`. Waiting triggers #6, #7, #8, #15, #71 and #91 (2) have not fired. 13 deck cards are still not executable (#14 (e)), and this task doesn't change that.
- 2026-11-04, kernel-engineer: done, `review`. Commits: `test(T-0157)` 6578688 (git rm tdd_test.go), `feat(toolchain)` e84ac69 (git rm tdd.go, slice.go; git.go lines 69-94 deleted). Reuse search: nothing added. Results in W: (1) `git grep -n -i tdd -- pkg cmd` empty, exit 1. (2) both package dirs absent. (3) `git grep -n -E "CommitExists|GetNewerCommits" -- cmd pkg` empty, exit 1. (4) `cardtown tdd` prints `Unknown command "tdd"`, exit status 1; `cardtown lint --tdd` prints `flag provided but not defined: -tdd`, exit status 2. (5) `cardtown whats-up` exit 0. (6) numstat prints exactly the four lines (26/636/23/555), no go.mod/go.sum. (7) `go tool task sim-compare` printed `[PASS]`. (8) `go tool task check` exit 0, `go tool task sim` exit 0. gofmt -l clean. `go mod tidy` not run, per the lead.
