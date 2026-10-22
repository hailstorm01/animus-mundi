---
id: T-0042
title: Split pkg/sim/monitor.go into the framework and one file per monitor (directive #35, harness, flash, pure move)
kind: harness
tier: flash
wave: none (board directive #35, structural review cleanup 5 of 7, before cardtown play)
status: review
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
- [x] Acceptance cases 1 to 6 hold, with their commands and results in the Log
- [x] `go tool task check` passes

## Log
- 2026-10-22 lead-architect: Wrote the task from board directive #35. The function-to-monitor map comes from each function's `count` and `fail` constants, and the shared-helper split from their real call sites:
  - `find`, `eachAsset` and `card` have callers in several monitors, so they stay in the framework.
  - `opponentAssets` is called only by `checkAttack`, `pool` only by `checkGenerate`, and `battlefield` only by `checkTargetDecision`, so each moves with its monitor.

  It is a harness task, tier flash, a pure move. T-0041 is done, so the task is `ready` on branch `m2/T-0042-split-monitor-go`. T-0037 now depends on it too.
  - Next: **the user runs `@kernel-engineer T-0042` in the Antigravity agent (Gemini Flash)**.
- 2026-10-22 kernel-engineer: Pure move done. Baselines captured from `rework/executable-rules` (462cf0a) before edits: sorted declarations list (`base_decls.txt`), `go test -list` (`base_tests.txt`), and baseline CLI `cardtown-base.exe`. Status set to `in-progress` in 311b444.
  - Reuse: pure move of existing monitor shapes (`opponentAssets`, `checkAttack`, `watchEntry`, `checkTurnStart`, `pool`, `checkGenerate`, `checkConvert`, `combatPair`, `openCombat`, `checkCombatDamage`, `checkDamage`, `targetPrompt`, `battlefield`, `checkTargetDecision`, `endTargetPrompts`). No new code, logic, or card names introduced.
  - Diff: only the moved declarations into the five new files (`monitor_attack.go`, `monitor_entry.go`, `monitor_resources.go`, `monitor_damage.go`, `monitor_targeting.go`); package and import blocks; and import cleanup in `monitor.go` (still uses `fmt`, `slices`, `strings`, `engine`).
  - Case 1 (declarations): sorted output of `git grep -h -E '^(func|type|var|const) ' -- 'pkg/sim/*.go' ':!pkg/sim/*_test.go'` is byte-identical before and after (52 declarations, Compare-Object empty, `fc.exe /b` reports no differences).
  - Case 2 (placement): `git grep -n -E '^(func|type) ' -- 'pkg/sim/monitor*.go' ':!pkg/sim/*_test.go'` matches the table exactly (18 in `monitor.go`, 2 in `monitor_attack.go`, 1 in `monitor_entry.go`, 4 in `monitor_resources.go`, 4 in `monitor_damage.go`, 4 in `monitor_targeting.go`):
    ```
    pkg/sim/monitor.go:34:type Violation struct {
    pkg/sim/monitor.go:43:func (v *Violation) Error() string {
    pkg/sim/monitor.go:49:type Coverage map[string]map[string]int
    pkg/sim/monitor.go:52:func (c Coverage) add(o Coverage) {
    pkg/sim/monitor.go:65:func ReplayCommand(simArgs []string, v *Violation) string {
    pkg/sim/monitor.go:74:type pendingCheck struct {
    pkg/sim/monitor.go:82:type Monitors struct {
    pkg/sim/monitor.go:98:func NewMonitors(decks [2]engine.DeckList, game int) *Monitors {
    pkg/sim/monitor.go:102:func zeroCoverage() Coverage {
    pkg/sim/monitor.go:114:func (m *Monitors) Violation() *Violation { return m.v }
    pkg/sim/monitor.go:117:func (m *Monitors) Coverage() Coverage {
    pkg/sim/monitor.go:123:func (m *Monitors) count(mon string, ids ...string) {
    pkg/sim/monitor.go:130:func (m *Monitors) fail(mon string, seq int, format string, args ...any) {
    pkg/sim/monitor.go:139:func (m *Monitors) Event(e engine.Event) {
    pkg/sim/monitor.go:177:func (m *Monitors) Settled(s engine.Snapshot) {
    pkg/sim/monitor.go:188:func find(s engine.Snapshot, id int) *engine.AssetSnap {
    pkg/sim/monitor.go:201:func eachAsset(s engine.Snapshot, f func(*engine.AssetSnap) bool) {
    pkg/sim/monitor.go:221:func (m *Monitors) card(a engine.AssetSnap) *engine.Card {
    pkg/sim/monitor_attack.go:11:func opponentAssets(s engine.Snapshot, p int) []engine.AssetSnap {
    pkg/sim/monitor_attack.go:24:func (m *Monitors) checkAttack(e engine.Event) {
    pkg/sim/monitor_damage.go:7:type combatPair struct {
    pkg/sim/monitor_damage.go:16:func (m *Monitors) openCombat(e engine.Event) {
    pkg/sim/monitor_damage.go:31:func (m *Monitors) checkCombatDamage(e engine.Event) {
    pkg/sim/monitor_damage.go:56:func (m *Monitors) checkDamage(e engine.Event) {
    pkg/sim/monitor_entry.go:6:func (m *Monitors) watchEntry(e engine.Event) {
    pkg/sim/monitor_resources.go:5:func (m *Monitors) checkTurnStart(e engine.Event) {
    pkg/sim/monitor_resources.go:18:func pool(s engine.Snapshot, p int, resource string) int {
    pkg/sim/monitor_resources.go:25:func (m *Monitors) checkGenerate(e engine.Event) {
    pkg/sim/monitor_resources.go:88:func (m *Monitors) checkConvert(e engine.Event) {
    pkg/sim/monitor_targeting.go:6:type targetPrompt struct{ seq, player int }
    pkg/sim/monitor_targeting.go:9:func battlefield(s engine.Snapshot) []engine.AssetSnap {
    pkg/sim/monitor_targeting.go:20:func (m *Monitors) checkTargetDecision(e engine.Event) {
    pkg/sim/monitor_targeting.go:54:func (m *Monitors) endTargetPrompts(e engine.Event) {
    ```
  - Case 3 (only moves): multiset comparison of removed lines against added code lines in `pkg/sim` confirms every removed line reappears in one of the new files; the only added lines are package headers, imports, and blank lines (`Compare-Object` empty).
  - Case 4 (tests): `go test -list . ./pkg/sim` has identical test names before and after (32 tests, Compare-Object empty when excluding elapsed timing).
  - Case 5 (CLI outputs): under seed 1 with both `-policy random` and `-policy progress`:
    - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH`: stdout byte-identical (`fc.exe /b` reports no differences).
    - bare `cardtown sim`: stdout byte-identical (`fc.exe /b` reports no differences).
    - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH -games 1 -seed 7 -log`: stdout byte-identical (`fc.exe /b` reports no differences).
    - `cardtown sim -games 10 -events`: JSON lines identical between baseline and new CLI once `"revision":…,"modified":…` is stripped (4138 lines for random, 5367 lines for progress; `fc.exe /b` reports no differences).
  - Case 6 (check gate): `go tool task check` passes cleanly (fmt-check, go vet, all package tests, cardtown lint --all: 0 warnings, 0 errors).
  - Status set to `review`.
