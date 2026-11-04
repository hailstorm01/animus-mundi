---
id: T-0152
title: One table of policy names, sim.PolicyByName (directive #99, harness, flash, pure move)
kind: harness
tier: flash
wave: C.1 (board directive #99, structural review 7, finding 8)
status: done
branch: m2/autoloop-wave-c.1/T-0152-policy-by-name
depends_on: []
claims: [helper:PolicyByName, helper:defaultRandomPolicy]
cards: []
requirements: []
---

## Goal
`cardtown play` and `cardtown sim` each validate the `-policy` name and each select the policy. `sim.PolicyByName` in `pkg/sim/policy.go` becomes the one table of names, and both commands call it (#99). A pure move: same behavior, same test names, same sim output.

## Inputs
- Board directive #99 (structural review 7, approved by the user, 2026-11-03).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6f (wave C.1).
- Each command keeps its own default (the flag's default string) and its own way of printing the error (#99): `play` writes to `out` and returns 1, `sim` uses `fmt.Printf` and `os.Exit(1)`.
- The one visible change: `cardtown sim -policy bogus` now says `(want progress or random)`, as `play` already did, instead of `(want random or progress)`. One table has one message. #99 allows it: no test pins the text (searched `unknown -policy`). The exit code stays 1.
- `defaultRandomPolicy` exists so that "random means `RandomPolicy` at `DefaultEndBias`" stays written once: without it, the table would copy the fallback in `SimulateWith` (`pkg/sim/sim.go:134`).
- `opt.Policy` used to stay nil for `random`, and `SimulateWith` filled in the same constructor, so setting it now changes nothing. Nothing else reads `opt.Policy` (`pkg/sim/sim.go:70`, `:132`).
- Not in this task: `Simulate`'s nil default stays, as do `RandomPolicy` and `ProgressPolicy`'s exported names, because tests call them (`cmd/cardtown/play_test.go:186`, `pkg/sim/monitor_damage_test.go:230`, `pkg/sim/playout_test.go:50`).
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 458ebbf: `gofmt -l` clean, `go tool task check` passes, `go tool task sim` runs clean, and `go tool task sim-compare -- autoloop/m2-wave-c.1` printed `[PASS]`. `sim-compare` runs only the default policy, so the lead also compared `sim -policy random` (starter and stress) against a build of the run branch: byte-identical, exit 0.

### Where it is today (at `autoloop/m2-wave-c.1`, 458ebbf)
| Where | Today |
| :--- | :--- |
| `pkg/sim/policy.go:3-7` | the import block; `Policy` is at lines 9-10, and `// RandomPolicy picks ...` starts at line 12 |
| `pkg/sim/sim.go:134` | the nil fallback writes the constructor inline |
| `cmd/cardtown/play.go:46-49` | validates the name |
| `cmd/cardtown/play.go:99-104` | selects the policy |
| `cmd/cardtown/sim.go:48-51` | validates the name |
| `cmd/cardtown/sim.go:72-75` | selects the policy |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/sim/policy.go`
In the import block, above the line `"math/rand/v2"`, insert this line:
```go
	"fmt"
```
Above the line `// RandomPolicy picks uniformly among non-EndTurn actions, ending the turn`, insert these lines, followed by one blank line:
```go
// PolicyByName returns the constructor of the policy that a -policy flag
// names: progress or random.
func PolicyByName(name string) (func(seed uint64) Policy, error) {
	switch name {
	case "progress":
		return ProgressPolicy, nil
	case "random":
		return defaultRandomPolicy, nil
	}
	return nil, fmt.Errorf("unknown -policy %q (want progress or random)", name)
}

// defaultRandomPolicy is RandomPolicy at DefaultEndBias.
func defaultRandomPolicy(seed uint64) Policy { return RandomPolicy(seed, DefaultEndBias) }
```

### 2. `pkg/sim/sim.go`
Replace line 134 with this line:
```go
			pol = defaultRandomPolicy
```

### 3. `cmd/cardtown/play.go`
Replace lines 46-49 with these lines:
```go
	newBotPolicy, err := sim.PolicyByName(*policy)
	if err != nil {
		fmt.Fprintf(out, "[ERROR] %v\n", err)
		return 1
	}
```
Replace the six lines that begin `var botPolicy sim.Policy` and end at the closing brace of the `else` (lines 99-104 before edit) with this line:
```go
	botPolicy := newBotPolicy(botSeed)
```

### 4. `cmd/cardtown/sim.go`
Replace lines 48-51 with these lines:
```go
	newPolicy, err := sim.PolicyByName(*policy)
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		os.Exit(1)
	}
```
Replace the four lines that begin `opt := sim.SimOptions{` and end at the closing brace of the `if` (lines 72-75 before edit) with this line:
```go
	opt := sim.SimOptions{DeckNames: deckNames, Policy: newPolicy, NoMonitors: !*monitors}
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#99): `git -C W diff autoloop/m2-wave-c.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#99): `go -C W tool task sim-compare -- autoloop/m2-wave-c.1` prints `[PASS]`.
3. **`-policy bogus` exits 1 from `sim`** (#99): `go -C W run ./cmd/cardtown sim -policy bogus` prints `[ERROR] unknown -policy "bogus" (want progress or random)` and `exit status 1`, and the command exits non-zero. That failure is the expected result here: record it and go on.
4. **`-policy bogus` exits 1 from `play`** (#99): `go -C W run ./cmd/cardtown play -policy bogus` prints the same two lines, and the command exits non-zero. Record it and go on.
5. **`-policy random` still runs:** `go -C W run ./cmd/cardtown sim -policy random -games 20` exits 0 and prints `[PASS] Every match terminated with all invariants holding.`
6. **One table** (#99): `git -C W grep -n 'PolicyByName' -- cmd pkg` prints exactly four lines: `cmd/cardtown/play.go:46`, `cmd/cardtown/sim.go:48`, `pkg/sim/policy.go:13` and `pkg/sim/policy.go:15`.
7. **No command reads the name any more** (#99): `git -C W grep -n -e 'policy [!=]= ' -- cmd` prints nothing, and `git -C W grep -n -e 'sim.ProgressPolicy' -e 'sim.RandomPolicy' -- cmd/cardtown/play.go cmd/cardtown/sim.go` prints nothing. `git grep` exits 1 when it finds nothing, which is the expected result here.
8. **The random constructor is written once:** `git -C W grep -n 'defaultRandomPolicy' -- cmd pkg` prints exactly four lines: `pkg/sim/policy.go:20`, `pkg/sim/policy.go:25`, `pkg/sim/policy.go:26` and `pkg/sim/sim.go:134`.
9. **Only the listed edits:** `git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg cmd` prints exactly these four lines: `4	8	cmd/cardtown/play.go`, `4	6	cmd/cardtown/sim.go`, `16	0	pkg/sim/policy.go` and `1	1	pkg/sim/sim.go`.
10. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
11. **Nothing else changed:** `git -C W diff autoloop/m2-wave-c.1 --stat` lists only those four files and this task file.
12. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/sim/policy.go`
- `pkg/sim/sim.go`
- `cmd/cardtown/play.go`
- `cmd/cardtown/sim.go`

## Done When
- [x] Cases 1-12 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-c.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-03, lead-architect (wave C.1 run, session 9): wrote this task against the run branch at 458ebbf; `main` had no commits to sync. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Reuse search: no name-to-policy helper exists in `pkg/sim` or `cmd/cardtown`; the only other writing of the random constructor is `pkg/sim/sim.go:134`, which this task folds into `defaultRandomPolicy`. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no task touches `deck.go`; #15: no turn-boundary hook; #71, #91 (2): no such card). Next: `kernel-engineer` on Flash.
- 2026-11-03, kernel-engineer: implemented the task in scratch worktree.
  - Changes made:
    - `pkg/sim/policy.go`: imported `fmt`, defined `PolicyByName(name string) (func(seed uint64) Policy, error)` returning `ProgressPolicy` or `defaultRandomPolicy`, and defined `defaultRandomPolicy(seed uint64) Policy { return RandomPolicy(seed, DefaultEndBias) }`.
    - `pkg/sim/sim.go`: replaced inline anonymous function in `SimulateWith` nil check with `defaultRandomPolicy`.
    - `cmd/cardtown/play.go`: validated policy name via `sim.PolicyByName(*policy)` writing `[ERROR] %v\n` to `out`, and instantiated `botPolicy := newBotPolicy(botSeed)`.
    - `cmd/cardtown/sim.go`: validated policy name via `sim.PolicyByName(*policy)` printing `[ERROR] %v\n` and calling `os.Exit(1)`, and initialized `opt := sim.SimOptions{DeckNames: deckNames, Policy: newPolicy, NoMonitors: !*monitors}`.
  - Reuse:
    - Reused `ProgressPolicy`, `RandomPolicy`, and `DefaultEndBias` within `pkg/sim/policy.go`.
    - Reused `defaultRandomPolicy` across `pkg/sim/policy.go` and `pkg/sim/sim.go`.
  - Acceptance cases verified:
    1. `git -C W diff autoloop/m2-wave-c.1 -- '*_test.go'` printed nothing; `go -C W test ./...` passed (exit 0).
    2. `go -C W tool task sim-compare -- autoloop/m2-wave-c.1` printed `[PASS] starter and stress sims match autoloop/m2-wave-c.1 byte for byte, exit codes included` (exit 0).
    3. `go -C W run ./cmd/cardtown sim -policy bogus` printed `[ERROR] unknown -policy "bogus" (want progress or random)` and `exit status 1` (exit 1).
    4. `go -C W run ./cmd/cardtown play -policy bogus` printed `[ERROR] unknown -policy "bogus" (want progress or random)` and `exit status 1` (exit 1).
    5. `go -C W run ./cmd/cardtown sim -policy random -games 20` exited 0 and printed `[PASS] Every match terminated with all invariants holding.`.
    6. `git -C W grep -n 'PolicyByName' -- cmd pkg` printed exactly 4 lines (`cmd/cardtown/play.go:46`, `cmd/cardtown/sim.go:48`, `pkg/sim/policy.go:13`, `pkg/sim/policy.go:15`).
    7. `git -C W grep -n -e 'policy [!=]= ' -- cmd` printed nothing (exit 1); `git -C W grep -n -e 'sim.ProgressPolicy' -e 'sim.RandomPolicy' -- cmd/cardtown/play.go cmd/cardtown/sim.go` printed nothing (exit 1).
    8. `git -C W grep -n 'defaultRandomPolicy' -- cmd pkg` printed exactly 4 lines (`pkg/sim/policy.go:20`, `pkg/sim/policy.go:25`, `pkg/sim/policy.go:26`, `pkg/sim/sim.go:134`).
    9. `git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg cmd` printed:
       `4	8	cmd/cardtown/play.go`
       `4	6	cmd/cardtown/sim.go`
       `16	0	pkg/sim/policy.go`
       `1	1	pkg/sim/sim.go`
    10. `go -C W run ./cmd/cardtown lint fences` printed `[PASS] Package code is free of card names, invalid markers and disallowed imports.` (exit 0).
    11. `git -C W diff autoloop/m2-wave-c.1 --stat` printed 5 files changed (`.agents/tasks/m2/T-0152-policy-by-name.md`, `cmd/cardtown/play.go`, `cmd/cardtown/sim.go`, `pkg/sim/policy.go`, `pkg/sim/sim.go`).
    12. `go -C W tool task check` passed all stages (fmt-check, vet, test, linter dispatcher); `go -C W tool task sim` passed (`[PASS] Every match terminated with all invariants holding.`, exit 0).
  - Status moved to `review`. Ready for lead review.
- 2026-11-03, lead-architect (wave C.1 run, session 9, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-c.1` as c858b13. Flash run `de4b1a4d-6be0-4fa7-9347-4f54d9b3e116`: exit 0, `status: SUCCESS`, 341 s, no refusals (no `denied_actions` key; the transcript has no step with `"status":"ERROR"` and no `permission check failed` match; the `.err` file holds only agy's line about waiting on a background task). Review of `git diff autoloop/m2-wave-c.1...m2/autoloop-wave-c.1/T-0152-policy-by-name` (cf67cc8):
  - Scope: only `policy.go`, `pkg/sim/sim.go`, `play.go`, `cmd/cardtown/sim.go` and this file changed. No test file changed.
  - Pure move, checked against the diff: the four code files match the task's Edits exactly. Validation stays at the same point in each command, so a bad `-seat` or `-log` still reports first. `-policy random` now sets `opt.Policy` where it used to stay nil, and that is the constructor `SimulateWith` would have filled in. The one visible change is the one the task names: `sim -policy bogus` lists the names as `play` does. The exit code is still 1.
  - Fences and placement: no card name, no new rule. Policies belong to `pkg/sim`, and `cmd/cardtown` only calls them.
  - Re-run by the reviewer in a scratch worktree of the task branch, all as expected: cases 1-12; `gofmt -l` clean; `go tool task check` passes; `go tool task sim` clean, 200 games, P0 82.5% / P1 17.5% / draws 0.0%, 42/50 and 32/50 main cards executable; `go tool task sim-compare -- autoloop/m2-wave-c.1` `[PASS]`.
  - `sim-compare` runs only the default policy, so the reviewer also compared a build of the task branch with a build of 458ebbf, output and exit code: `sim -policy random` (starter and stress), and scripted `play` with `-policy random` (starter seat 1 seed 7, stress seed 3) and `-policy progress` (starter seat 1 seed 7). All byte-identical.
  - The merge's `pkg/` and `cmd/` match the task tip's, so the gates hold on the run branch.
