---
id: T-0044
title: Run git through pkg/toolchain/git in audit, docs and tddslice (directive #37, harness, sonnet)
kind: harness
tier: sonnet
wave: none (board directive #37, structural review cleanup 7 of 7, before cardtown play)
status: in-progress
branch: m2/T-0044-git-through-toolchain-git
depends_on: [T-0043]
claims: [func:git.Lines, func:git.Output, func:git.Combined]
cards: []
requirements: []
---

## Goal
Every `git` process the toolchain starts goes through `pkg/toolchain/git`. Today three packages start `git` themselves:
- `audit`: rm, add, commit, `rev-parse --short HEAD` and push, in `cardtown audit resolve`;
- `docs`: six queries in `cardtown whats-up`, through a local `getGit` helper;
- `tddslice`: `remote get-url origin`.

After this task, those calls go through three small runners in the `git` package. Each runner has the output mode its callers use today. No message, exit code or output changes.

## Inputs
- Board directive #37 (structural finding 9).
- Call sites:
  - [audit.go:255-310](../../../pkg/toolchain/audit/audit.go);
  - [whatsup.go:38-51](../../../pkg/toolchain/docs/whatsup.go) (`getGit`) and its callers at lines 135, 141, 148, 192, 201 and 268;
  - [slice.go:96](../../../pkg/toolchain/tddslice/slice.go).
- `pkg/toolchain/git/git.go` already has `GetStatus`, `GetStagedFiles`, `CommitExists` and `GetNewerCommits`, and no tests.
- `gh` calls in `tddslice` are not `git` and are out of scope.
- Rulings that apply: none.

### Lead decisions
1. **Three runners, not one function per command.** The directive suggested named functions (add, rm, commit and so on). But `whats-up` alone runs six different queries, and `audit`'s calls differ only in their arguments. So the `git` package gets three runners, one per output mode in use today. Each runs `git -C <repoRoot> <args...>`:
   - `Lines(repoRoot string, args ...string) []string`: stdout split into lines with `bufio.Scanner`, each line `strings.TrimSpace`d, blank lines kept. It returns nil on any error. These are exactly `getGit`'s semantics; `getGit` is deleted and its callers call `git.Lines`.
   - `Output(repoRoot string, args ...string) ([]byte, error)`: stdout only (`cmd.Output()`). For `rev-parse --short HEAD` in `audit` and `remote get-url origin` in `tddslice`.
   - `Combined(repoRoot string, args ...string) ([]byte, error)`: stdout and stderr together (`cmd.CombinedOutput()`). For `audit`'s rm, add, commit and push, whose error messages embed that output. The two registry calls that ignore their result today (`_ = cmd.Run()`) become `_, _ = git.Combined(...)`.

   Each runner gets a one-line doc comment. The existing four functions in `git.go` stay as they are.
2. **Callers keep their arguments, messages and order.**
   - In `audit.go`, every `fmt.Errorf` text stays as it is, including the `string(out)` it embeds.
   - In `slice.go`, the regexp and the fallback repo stay.
   - `os/exec` remains imported in `slice.go` for `gh`. It is dropped from `audit.go` and `whatsup.go` if nothing else there uses it.
3. **Not in this task:** changing which git commands run, `audit`'s push behavior, the `gh` calls, or the existing `git.go` functions.

## Acceptance Cases (harness tasks only; written by the lead)
Cases 1–4 are unit tests in a new `pkg/toolchain/git/git_test.go`. Each one builds a throwaway repo in `t.TempDir()`: `git init`, then `git config user.name test`, `git config user.email test@example.com` and `git config commit.gpgsign false`. That config is local to the temp repo only; never change the user's or the project's git config. Cases 5–7 are checks recorded in the Log against a baseline CLI built from `rework/executable-rules` before any edit; the reviewer re-runs them.
1. **Lines:** in a temp repo with commits `first` and then `second`, `Lines(root, "log", "--format=%s")` → `["second", "first"]`. `Lines(root, "no-such-subcommand")` → nil.
2. **Output:** `Output(root, "rev-parse", "--short", "HEAD")`, trimmed, equals what `git -C root rev-parse --short HEAD` prints. `Output(root, "remote", "get-url", "origin")` in a repo with no remote returns a non-nil error.
3. **Combined, failure:** `Combined(root, "rm", "nope.txt")` returns a non-nil error, and output containing `nope.txt` (git's pathspec message comes on stderr, which `Combined` must include).
4. **Combined, success:** write `a.txt`; `Combined(root, "add", "a.txt")` and `Combined(root, "commit", "-m", "add a")` both succeed, and `Lines(root, "log", "-1", "--format=%s")` → `["add a"]`.
5. **No direct git:** `git grep -n 'exec.Command("git"' -- pkg cmd ':!*_test.go'` matches only `pkg/toolchain/git/git.go`, and `getGit` no longer exists.
6. **Outputs unchanged.** Run each command with the baseline and new binaries back to back, from the same clean checkout, and compare stdout and exit code byte for byte (`cmp`):
   - `cardtown whats-up`;
   - `cardtown tdd slice -dry-run docs/tdd/03_Card_Compiler_TDD.md`.

   The second makes read-only `gh` calls when `gh` is signed in. If its output differs only because GitHub returned different issues between the two runs, rerun both. Record whether `gh` was available.
7. **audit resolve dry-run:** `cardtown audit resolve -dry-run` with the arguments of any existing pending audit is byte-identical, or both binaries give the same usage or error output if none is pending. A live `audit resolve` is not run, because it commits and can push. Cases 3 and 4 cover the commands it uses.
8. `go tool task check` passes.

## Allowed Files
- `pkg/toolchain/git/git.go` (the three runners only)
- `pkg/toolchain/audit/audit.go`
- `pkg/toolchain/docs/whatsup.go`
- `pkg/toolchain/tddslice/slice.go` (the `remote get-url` call only)
- Tests: new `pkg/toolchain/git/git_test.go`

## Done When
- [ ] Cases 1–4 have tests, committed as `test(T-0044): …` before the implementation, and they failed before it (failing to compile counts)
- [ ] Cases 5–7 hold, with their commands and results in the Log
- [ ] `go test ./pkg/toolchain/...` passes
- [ ] `go tool task check` passes

## Log
- 2026-10-22 lead-architect: Wrote the task from board directive #37. It narrows the directive's "one function per command" suggestion to three runners by output mode (Lead decision 1), because `whats-up`'s six queries and `audit`'s calls differ only in their arguments. It is a harness task, tier sonnet, as the directive sets. T-0043 is done, so the task is `ready` on branch `m2/T-0044-git-through-toolchain-git`. T-0037 now depends on it too. This is the last structural-cleanup task.
  - Next: the lead dispatches the `kernel-engineer` subagent (Sonnet).
