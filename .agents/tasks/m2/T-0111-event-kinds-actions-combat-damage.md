---
id: T-0111
title: actions.go, combat.go and damage.go emit with the event kind constants (directive #86 part 2 of 5, harness, flash, pure move)
kind: harness
tier: flash
wave: G.2 (board directive #86, structural review 6, finding 5)
status: done
branch: m2/T-0111-event-kinds-actions-combat-damage
depends_on: [T-0110]
claims: []
cards: []
requirements: []
---

## Goal
The event emits in `pkg/engine/actions.go`, `combat.go` and `damage.go` name their kind by the constants T-0110 declared in `events.go` (`EventGenerate`, `EventAttack`, `EventFizzle`, `EventCombat`, `EventDamage`), not by a string literal. A pure move: same tests, same sim output.

## Inputs
- Board directive #86 (structural review 6, approved by the user, 2026-10-30), part 2 of the five-task split recorded in its status cell: T-0110 (done: the constants, and `events.go`, `choice.go`, `sweep.go`, `zones.go`), **T-0111 (this task)**: `actions.go`, `combat.go`, `damage.go`; T-0112: `effects.go`, `replace.go`, `stack.go`; T-0113: `pkg/sim`; T-0114: `pkg/view`, `pkg/eventlog`, and #86 (c) over all of `pkg/` and `cmd/cardtown`.
- ADR-0005 (`docs/architecture/decisions/0005-structured-game-events-and-observer.md`): the kind strings are its schema. `TestEventKinds_ValuesAreTheSchema` (T-0110) pins each constant to its value.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6e (wave G.2).
- Why it is a pure move: each replaced literal becomes the constant with the same value, so every emitted `Event` is unchanged. The lead dry-ran these exact edits on the run branch at ca96220: `gofmt -l pkg` printed nothing, `go tool task check` passed, `cardtown lint fences` passed, and `go tool task sim-compare -- autoloop/m2-wave-g.2` printed `[PASS]`.

### Where it is today (at `autoloop/m2-wave-g.2` ca96220)
| Where | Today |
| :--- | :--- |
| `pkg/engine/actions.go:72` | `g.logEvent(Event{Kind: "generate", …` |
| `pkg/engine/combat.go:31` | `g.logEvent(Event{Kind: "attack", …` |
| `pkg/engine/combat.go:95` | `g.logEvent(Event{Kind: "fizzle", …` (the defender is gone) |
| `pkg/engine/combat.go:102` | `g.logEvent(Event{Kind: "fizzle", …` (the attacker is gone) |
| `pkg/engine/combat.go:108` | `g.logEvent(Event{Kind: "combat", …` |
| `pkg/engine/damage.go:55` | `g.logEvent(Event{Kind: "damage", …` |

## Edits (lead decisions; copy exactly)
On each line below, replace only the quoted kind with the constant; nothing else on the line changes.

### 1. `pkg/engine/actions.go`
- Line 72: `Kind: "generate"` becomes `Kind: EventGenerate`.

### 2. `pkg/engine/combat.go`
- Line 31: `Kind: "attack"` becomes `Kind: EventAttack`.
- Line 95: `Kind: "fizzle"` becomes `Kind: EventFizzle`.
- Line 102: `Kind: "fizzle"` becomes `Kind: EventFizzle`.
- Line 108: `Kind: "combat"` becomes `Kind: EventCombat`.

### 3. `pkg/engine/damage.go`
- Line 55: `Kind: "damage"` becomes `Kind: EventDamage`.
- **Leave line 47 alone.** `kind = "combat"` (with `kind := "effect"` on line 45) is the damage word in the log text ("deals 3 combat damage"), not an event kind.

Nothing else changes in these three files, and no other file changes except this task file. `effects.go`, `replace.go` and `stack.go` are T-0112's: leave them alone.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move, and T-0110's `TestEventKinds_ValuesAreTheSchema` already pins the values. Run each command on its own, pointed at your worktree (W). Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#86 (a)): `git -C W diff autoloop/m2-wave-g.2 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#86 (b)): `go -C W tool task sim-compare -- autoloop/m2-wave-g.2` prints `[PASS]`.
3. **No kind literal left in these three files** (#86 (c), for this task's files): `git -C W grep -n -E -e "Kind: [^A-Za-z]" -e "logf[(][^A-Za-z]" -- pkg/engine/actions.go pkg/engine/combat.go pkg/engine/damage.go` prints nothing (git grep exits 1 when nothing matches; that is the expected result).
4. **The six emits use the constants:** `git -C W grep -n -E "Kind: Event" -- pkg/engine/actions.go pkg/engine/combat.go pkg/engine/damage.go` prints exactly six lines: `pkg/engine/actions.go:72` (`EventGenerate`), `pkg/engine/combat.go:31` (`EventAttack`), `:95` and `:102` (`EventFizzle`), `:108` (`EventCombat`) and `pkg/engine/damage.go:55` (`EventDamage`).
5. **Moved, not changed:** `git -C W diff autoloop/m2-wave-g.2 --numstat -- pkg` prints exactly these three lines: `1	1	pkg/engine/actions.go`, `4	4	pkg/engine/combat.go` and `1	1	pkg/engine/damage.go`.
6. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
7. **Only these edits:** `git -C W diff autoloop/m2-wave-g.2 --stat` lists only the three files in case 5 and this task file.
8. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/actions.go`
- `pkg/engine/combat.go`
- `pkg/engine/damage.go`

## Done When
- [x] Cases 1–8 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-g.2` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-30, lead-architect (wave G.2 run, session 5): wrote the task from board directive #86, part 2 of the five-task split T-0110 recorded; T-0110 is `done` (merge c91f765). Harness, so it skips `tests` and goes straight to `ready`; tier `flash` as #86 gives. No new test: like T-0105 and T-0108, a pure move is pinned by the unchanged tests and `sim-compare`, and T-0110's test already pins each constant's value. A `git grep` over every non-test engine file for a quoted kind found exactly the six lines above in these three files; `damage.go:47`'s `"combat"` is log text, so it stays (T-0114's case (c) must not count it either). Case 3 uses `[^A-Za-z]` so it matches a quote but not `ActGenerate`, `StackAttack` or an `Event…` constant. Every command here is on the headless allow-list, one per line, with no `|`, `$` or unquoted parentheses. The edits were dry-run on the run branch (Inputs). `main` (281e7d8) has no commits the run branch lacks, so no sync. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15, #20 (b), #58, #59 (2), #61, #71: no such hook, claim or card here). #87 (c) stays `open: before the next Deliverables task`, after #86. Next: `kernel-engineer` on Flash.
- 2026-10-30, kernel-engineer (T-0111): implemented directive #86 part 2 in commit `feat(engine): emit with event kind constants in actions, combat, damage (T-0111)` (77ab96f):
  - `pkg/engine/actions.go:72`: replaced `"generate"` with `EventGenerate`.
  - `pkg/engine/combat.go:31`: replaced `"attack"` with `EventAttack`.
  - `pkg/engine/combat.go:95, 102`: replaced `"fizzle"` with `EventFizzle`.
  - `pkg/engine/combat.go:108`: replaced `"combat"` with `EventCombat`.
  - `pkg/engine/damage.go:55`: replaced `"damage"` with `EventDamage`. Left line 47 (`kind = "combat"`) unchanged as instructed.
  Reused: existing event kind constants from `pkg/engine/events.go` (`EventGenerate`, `EventAttack`, `EventFizzle`, `EventCombat`, `EventDamage`).
  Acceptance cases verified on task branch:
  - Case 1: `git -C W diff autoloop/m2-wave-g.2 -- '*_test.go'` printed nothing; `go -C W test ./...` passed (exit code 0).
  - Case 2: `go -C W tool task sim-compare -- autoloop/m2-wave-g.2` printed `[PASS] starter and stress sims match autoloop/m2-wave-g.2 byte for byte, exit codes included`.
  - Case 3: `git -C W grep -n -E -e "Kind: [^A-Za-z]" -e "logf[(][^A-Za-z]" -- pkg/engine/actions.go pkg/engine/combat.go pkg/engine/damage.go` printed nothing (exit code 1).
  - Case 4: `git -C W grep -n -E "Kind: Event" -- pkg/engine/actions.go pkg/engine/combat.go pkg/engine/damage.go` printed exactly the 6 expected lines (`actions.go:72`, `combat.go:31`, `combat.go:95`, `combat.go:102`, `combat.go:108`, `damage.go:55`).
  - Case 5: `git -C W diff autoloop/m2-wave-g.2 --numstat -- pkg` printed exactly `1 1 actions.go`, `4 4 combat.go`, `1 1 damage.go`.
  - Case 6: `go -C W run ./cmd/cardtown lint fences` passed (exit code 0).
  - Case 7: `git -C W diff autoloop/m2-wave-g.2 --stat` listed only the 3 pkg files and the task file.
  - Case 8: `go -C W tool task check` and `go -C W tool task sim` passed cleanly.
  Handoff: ready for lead-architect review.
- 2026-10-30, lead-architect (wave G.2 run, session 5, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-g.2` as 10d99e8. Flash run `903bcdee-d6f5-477a-9150-f05e01e6b6cd`: exit 0, `status: SUCCESS`, 335 s, no refusals (the JSON has no `denied_actions` entry, `cli.log` has no `soft-denying` line and the transcript no `permission check failed`). Review of `git diff autoloop/m2-wave-g.2...m2/T-0111-event-kinds-actions-combat-damage` (f2a28a4):
  - Scope: only `actions.go`, `combat.go`, `damage.go` and this file changed; no test file changed (case 1 printed nothing).
  - Pure move, checked line by line: the code hunks are exactly Edits 1–3, identical to the lead's dry run. Each of the 6 replaced literals became the constant with the same value, so every emitted event is unchanged. `damage.go:47`'s `"combat"` (log text) stays.
  - Fences and placement: no card name, no new rule, no new declaration.
  - Gates re-run in a scratch worktree of the task branch, each with exit 0: case 3 (git grep exit 1, nothing printed); case 4 (the six lines); case 5 (the three numstat lines); case 6 `cardtown lint fences`; `go tool task sim-compare -- autoloop/m2-wave-g.2` `[PASS]` (starter and stress byte-identical); `go tool task check` passes; `go tool task sim` clean, 200 games, P0 67.0% / P1 32.5% / draws 0.5% (unchanged), 32/50 Animus and 25/50 Path main cards executable.
  - The merge's tree equals the task tip's, so the gates hold on the run branch. Directive #86 stays open: T-0112 (`effects.go`, `replace.go`, `stack.go`) is next, then T-0113 and T-0114, then #87 (c).
