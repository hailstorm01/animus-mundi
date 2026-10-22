---
id: T-0042
title: Split pkg/sim/monitor.go into the framework and one file per monitor (directive #35, harness, flash, pure move)
kind: harness
tier: flash
wave: none (board directive #35, structural review cleanup 5 of 7, before cardtown play)
status: in-progress
branch: m2/T-0042-split-monitor-go
depends_on: [T-0041]
claims: [file:pkg/sim/monitor.go, file:pkg/sim/monitor_attack.go, file:pkg/sim/monitor_entry.go, file:pkg/sim/monitor_resources.go, file:pkg/sim/monitor_damage.go, file:pkg/sim/monitor_targeting.go]
cards: []
requirements: []
---

## Goal
`pkg/sim/monitor.go` (633 lines) holds only the monitor framework. Each of the five monitors gets its own file holding its checks and the helpers only it uses. This way, a later `shape` task that adds a monitor (design note, Rule monitors) adds one file instead of growing this one.

Nothing is renamed, and no behavior or output changes.

## Inputs
- Board directive #35 (structural finding 7). The five monitors are the keys of `monitorRules`: `monAttack`, `monEntry`, `monResources`, `monDamage` and `monTargeting`.
- Pattern to follow: T-0041 (a pure move, checked by declarations, moved lines and byte-identical CLI output).
- Rulings that apply: none.

### Target layout (lead decision)
The lead grouped the functions by the monitor constant each one passes to `count` and `fail`. A helper goes with the only monitor that calls it. A helper that more than one monitor calls stays in the framework. Move each declaration verbatim, with the comment directly above it, keeping the order it has in `monitor.go`.

| File | Holds |
| :--- | :--- |
| `monitor.go` (stays) | The `// Rule monitors …` comment block. The monitor-name consts and `monitorRules`. `Violation` and its `Error`, `Coverage` and `add`, `ReplayCommand`, `pendingCheck`, the `Monitors` struct and `NewMonitors`. `zeroCoverage`, the `Violation` and `Coverage` methods, `count` and `fail`. `Event` and `Settled`. The shared helpers `find`, `eachAsset` and `card`. |
| `monitor_attack.go` (new) | `opponentAssets` and `checkAttack` |
| `monitor_entry.go` (new) | `watchEntry` |
| `monitor_resources.go` (new) | `checkTurnStart`, `pool`, `checkGenerate` and `checkConvert` |
| `monitor_damage.go` (new) | the `combatPair` type, `openCombat`, `checkCombatDamage` and `checkDamage` |
| `monitor_targeting.go` (new) | the `targetPrompt` type, `battlefield`, `checkTargetDecision` and `endTargetPrompts` |

Further rules:
- **The `Monitors` struct stays whole** in `monitor.go`, including its `combat` and `targets` fields. They refer to `combatPair` and `targetPrompt`, which may live in another file of the same package.
- **New files** start with `package sim`, with no header comment, like the other `sim` files. Each imports only what it uses (`fmt`, `slices`, `strings`, `animus-mundi/pkg/engine`). Remove imports from `monitor.go` that it no longer uses. Run `go tool task fmt`, then `go vet ./pkg/sim`.
- **The test files don't change.** `monitor_test.go` and `monitor_damage_test.go` keep their names; `monitor_damage.go` simply sits beside its test.
- **Not in this task:**
  - any rename;
  - any change to `monitorRules`, the dispatch in `Event`, or behavior;
  - any doc comment edit;
  - the design note.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests. Build a baseline CLI from `rework/executable-rules` before any edit (`go build -o <scratch>/cardtown-base.exe ./cmd/cardtown`), and record each check's commands and results in the Log. The reviewer re-runs them.
1. **Same declarations:** the sorted output of `git grep -h -E '^(func|type|var|const) ' -- 'pkg/sim/*.go' ':!pkg/sim/*_test.go'` is identical before and after.
2. **Placement:** each file declares exactly the names in its row. `git grep -n -E '^(func|type) ' -- 'pkg/sim/monitor*.go' ':!pkg/sim/*_test.go'` matches the table, and record that output in the Log.
3. **Only moves:** every line removed from `monitor.go` reappears in one of the five new files, apart from import lines. The only other added lines are the new files' `package` lines, import blocks and blank lines. (The reviewer checks this with a sorted comparison of removed against added lines.)
4. **Same tests:** `go test -list . ./pkg/sim` gives the same names before and after.
5. **Outputs unchanged:** with seed 1 under both `-policy random` and `-policy progress`, the baseline and new binaries give byte-identical stdout (`cmp`) for:
   - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH`, which includes the monitor coverage table;
   - bare `cardtown sim`;
   - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH -games 1 -seed 7 -log`.

   `-games 10 -events` files are identical once each match-start's `"revision":…,"modified":…` is removed.
6. `go tool task check` passes.

## Allowed Files
This is a pure move: the size rule is met by cases 1 to 3, as in T-0041.
- `pkg/sim/monitor.go`
- new: `pkg/sim/monitor_attack.go`, `monitor_entry.go`, `monitor_resources.go`, `monitor_damage.go`, `monitor_targeting.go`

## Done When
- [ ] Acceptance cases 1 to 6 hold, with their commands and results in the Log
- [ ] `go tool task check` passes

## Log
- 2026-10-22 lead-architect: Wrote the task from board directive #35. The function-to-monitor map comes from each function's `count` and `fail` constants, and the shared-helper split from their real call sites:
  - `find`, `eachAsset` and `card` have callers in several monitors, so they stay in the framework.
  - `opponentAssets` is called only by `checkAttack`, `pool` only by `checkGenerate`, and `battlefield` only by `checkTargetDecision`, so each moves with its monitor.

  It is a harness task, tier flash, a pure move. T-0041 is done, so the task is `ready` on branch `m2/T-0042-split-monitor-go`. T-0037 now depends on it too.
  - Next: **the user runs `@kernel-engineer T-0042` in the Antigravity agent (Gemini Flash)**.
