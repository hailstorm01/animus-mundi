---
id: T-0049
title: normalizeArgs as a small alias map (directive #44, harness, flash)
kind: harness
tier: flash
wave: none (board directive #44, structural review 2 cleanup 5 of 5, before the next shape task)
status: in-progress
branch: m2/T-0049-normalizeargs-aliases
depends_on: [T-0048]
claims: [func:normalizeArgs, var:flagAliases]
cards: []
requirements: []
---

## Goal
`normalizeArgs` in `cmd/cardtown/main.go` keeps its behavior but loses its dead code. Today it:
- lowercases every flag name;
- rewrites it with a `--` prefix, keeping any `=value` unchanged;
- applies seven real aliases.

About 25 of its `switch` cases map a name to itself and do nothing. After this task, a map holds only the aliases, and a new flag needs no line here.

## Inputs
- Board directive #44 (structural review 2, finding 7).
- [main.go `normalizeArgs`](../../../cmd/cardtown/main.go). Its callers are `lint.go`, `play.go`, `replay.go`, `sim.go` and `tools.go` (×3), and none changes.
- Rulings that apply: none.

### Lead decisions
1. **The map**, a package-level variable in `main.go` directly above `normalizeArgs`:
   ```go
   // flagAliases maps alternative spellings to the flag name the commands define.
   // Every other flag name is only lowercased.
   var flagAliases = map[string]string{
   	"reporoot":       "repo-root",
   	"dryrun":         "dry-run",
   	"updateexisting": "update-existing",
   	"showbodies":     "show-bodies",
   	"tddfile":        "tdd-file",
   	"commitmessage":  "msg",
   	"commit-message": "msg",
   	"maxturns":       "max-turns",
   }
   ```
2. **The function:** for an argument starting with `-`:
   - split once on `=`;
   - trim every leading `-` and lowercase the name;
   - replace it from `flagAliases` if present;
   - emit `--<name>`, plus `=<value>` if there was one, with the value unchanged.

   Any other argument passes through unchanged. A nil or empty input returns a nil slice, as today. The whole `switch` goes.
3. **The test is a characterization test.** It pins today's behavior, so unlike other harness tests it passes before the implementation as well as after. Commit it first as `test(T-0049): …`, and record that it passed against the old code.

## Acceptance Cases (harness tasks only; written by the lead)
1. **A table test in a new `cmd/cardtown/args_test.go`** (`TestNormalizeArgs`), with these exact input → output pairs:
   | Input | Output |
   | :--- | :--- |
   | `-ALL` | `--all` |
   | `--Repo-Root=x` | `--repo-root=x` |
   | `-reporoot`, `C:/x` | `--repo-root`, `C:/x` |
   | `-DryRun` | `--dry-run` |
   | `-updateExisting` | `--update-existing` |
   | `-SHOWBODIES` | `--show-bodies` |
   | `-tddFile=a.md` | `--tdd-file=a.md` |
   | `-commitMessage=Hi There` | `--msg=Hi There` |
   | `-commit-message`, `m` | `--msg`, `m` |
   | `-MSG` | `--msg` |
   | `-maxTurns=5` | `--max-turns=5` |
   | `-max-turns` | `--max-turns` |
   | `-Seed=3`, `-P0`, `ANIMUS` | `--seed=3`, `--p0`, `ANIMUS` |
   | `-h` | `--h` |
   | `-unknownFlag=V` | `--unknownflag=V` |
   | `---x` | `--x` |
   | `-a=b=c` | `--a=b=c` |
   | `file.md`, `-` | `file.md`, `--` |
   | (no arguments) | nil |

   The test also loops over every name the old `switch` listed: `all frd tdd cards sheets syntax fences staged files reporoot repo-root dryrun dry-run publish updateexisting update-existing showbodies show-bodies tddfile tdd-file source msg commitmessage commit-message push section rules deck games seed p0 p1 seat policy maxturns max-turns`. Each is given in upper case with `=v`, and must return `--<the old switch's output>=v`. The test passes before and after the change.
2. **CLI unchanged:** against a baseline CLI built from `rework/executable-rules` before any edit, these are byte-identical with stdout+stderr and exit codes (`cmp`):
   - `cardtown` (usage);
   - `cardtown sim -h`, `cardtown replay -h`, `cardtown play -h` and `cardtown tdd slice -h`;
   - `cardtown lint --all`;
   - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH -maxTurns=400` (exercises an alias).
3. `git grep -n -c 'case "' -- cmd/cardtown/main.go` counts only the `switch command` cases in `main()`, none in `normalizeArgs`.
4. `go tool task check` passes.

## Allowed Files
- `cmd/cardtown/main.go` (`flagAliases` and `normalizeArgs` only)
- Tests: new `cmd/cardtown/args_test.go`

## Done When
- [ ] Case 1's test is committed first, and passed against the old code (recorded)
- [ ] Cases 2–3 hold, with their commands and results in the Log
- [ ] `go tool task check` passes

## Log
- 2026-10-22 lead-architect: Wrote the task from board directive #44. It is a harness task, tier flash, and the first one the lead dispatches headless with `agy -p` (Runbook D3, `9dc7737`). Case 1 is a characterization test, so it passes before the change (lead decision 3). T-0048 is done, so the task is `ready` on branch `m2/T-0049-normalizeargs-aliases`.
