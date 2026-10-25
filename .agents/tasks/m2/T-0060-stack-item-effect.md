---
id: T-0060
title: A stack item carries an effect, not a whole trigger (directive #53, harness, flash, pure move)
kind: harness
tier: flash
wave: none (board directive #53, structural review 3 finding 5, before T-0055)
status: done
branch: m2/T-0060-stack-item-effect
depends_on: [T-0059]
claims: [field:engine.StackItem.Effect, field:engine.StackItem.Tracked]
cards: []
requirements: []
---

## Goal
A `StackAbility` item carries the `ability.Effect` it resolves and a `Tracked` flag, and no longer carries the whole `ability.Trigger`. Resolution reads `it.Effect`, and `relocateSelf` checks `it.Tracked` instead of `it.Trigger.Hook`. Maneuvers (wave C) and activated abilities (wave D) can then resolve through `resolveAbility` without a trigger. Nothing else changes: same tests, same sim output.

## Inputs
- Board directive #53 (structural review 3, finding 5).
- Rule the `Tracked` flag keeps: `FR-ZON-004`. A death trigger's "this card" is the dead instance in its owner's discard pile. Any other trigger's "this" is the asset on the battlefield, which is a new object once it leaves.
- Pattern to follow: T-0058 and T-0059 (pure moves, checked by byte-identical outputs).

### Where `StackItem.Trigger` is used (at `main` f4c7ac3)
| Where | Today | Becomes |
| :--- | :--- | :--- |
| `pkg/engine/stack.go:32`, the field | `Trigger ability.Trigger` | `Effect ability.Effect` plus `Tracked bool` (text below) |
| `pkg/engine/trigger.go:131`, `flushTriggers` placing the item | `Trigger: p.trig,` | `Effect: p.trig.Effect, Tracked: p.trig.Hook == ability.HookKilled,` |
| `pkg/engine/effects.go:20`, second verification stage | `it.Trigger.Effect.Targets[i].Filter` | `it.Effect.Targets[i].Filter` |
| `pkg/engine/effects.go:25`, the step loop | `it.Trigger.Effect.Steps` | `it.Effect.Steps` |
| `pkg/engine/effects.go:155`, `relocateSelf` | `it.Trigger.Hook != ability.HookKilled` | `!it.Tracked` |

No test file and no package outside `engine` reads or sets `StackItem.Trigger`. `git grep -nE '\.Trigger\b' -- '*.go'` finds only the three `effects.go` lines, and the one composite literal that sets it is `trigger.go:131`. `pendingTrigger.trig` stays as it is: the buffer still needs the whole trigger for `Optional`, the targets and the label.

### Why a flag, not "the card is in discard" (lead decision)
The directive suggests expressing the guard as "the item's own card is in its owner's discard pile". `relocateSelf` already checks that, one line later. Dropping the hook check and keeping only that one would change a rule: a non-death trigger with a `SubjectSelf` relocation step, such as a deploy trigger reading "return this to your hand", would then pull its card back out of discard after the asset died. That is wrong under `FR-ZON-004`. No card does this today, so no test or sim would notice. The flag records the same fact as the hook check, decided when the item is placed, where the trigger is known.

### Text to use (lead decision; copy exactly)
In `pkg/engine/stack.go`, the `StackAbility` constant's comment becomes:
```go
	StackAbility // a triggered ability: runs Effect's steps for Controller
```
The two `StackAbility` fields at the end of `StackItem` become three:
```go
	Effect  ability.Effect // StackAbility: the effect resolving (Source is its source asset)
	Chosen  [][]int        // StackAbility: asset IDs chosen for each Effect.Targets slot, in pick order; empty if none
	Tracked bool           // StackAbility: a death trigger, so Inst is its card tracked into its owner's discard pile (FR-ZON-004)
```
In `pkg/engine/effects.go`, the last sentence of `relocateSelf`'s doc comment, "Only defined for SubjectSelf on a HookKilled trigger.", becomes "Only defined for SubjectSelf on a Tracked item." Rewrap the comment if needed. Leave `resolveAbility`'s comment ("runs a triggered ability's steps") as it is: Maneuvers will reword it when they first resolve through it.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. Before any edit, build a baseline CLI from `main` (`go build -o <scratch>/cardtown-baseline.exe ./cmd/cardtown`) and save its sim outputs. Record each check's commands and results in the Log; the reviewer re-runs them.
1. **Existing tests unchanged:** `git diff main -- '*_test.go'` is empty, and `go test ./...` passes.
2. **Starter sim byte-identical:** `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` from the baseline and the new binary: stdout+stderr and exit code match (`cmp` or `fc.exe /b`).
3. **Stress sim byte-identical:** the same for `cardtown sim` with no `-deck`.
4. **No trigger on the item:** `git grep -nE '\bit\.Trigger\b' -- 'pkg/engine/*.go' ':!pkg/engine/*_test.go'` prints nothing, and `git grep -n 'ability.Trigger' -- pkg/engine/stack.go` prints nothing.
5. **One place sets the new fields:** `git grep -nE 'Tracked:|Effect: p\.trig' -- 'pkg/engine/*.go' ':!pkg/engine/*_test.go'` prints exactly the `flushTriggers` line in `trigger.go`.
6. **Only these edits:** `git diff main --stat` lists only `pkg/engine/stack.go`, `pkg/engine/trigger.go`, `pkg/engine/effects.go` and this task file.
7. `go tool task check` and `go tool task sim` pass.

## Allowed Files
- `pkg/engine/stack.go` (the constant's comment and the `StackItem` fields)
- `pkg/engine/trigger.go` (the `StackItem` literal in `flushTriggers`)
- `pkg/engine/effects.go` (the three reads and `relocateSelf`'s comment)

## Done When
- [x] Cases 1–6 hold, with their commands and results in the Log
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-25 lead-architect: Wrote the task from board directive #53, the third of the four directives the user put before T-0055 (#49 and #52 are applied; #54 comes next). Harness, tier flash, pure move, so it skips `tests` and is `ready` on branch `m2/T-0060-stack-item-effect` (cut from `main` at this commit). Checked the directive's line references against `main` f4c7ac3: `stack.go:32` and `effects.go:20, 25, 155` still hold. Lead decision: `relocateSelf`'s guard becomes a `Tracked` flag set on placement, not the discard-pile check alone, which would change `FR-ZON-004` for non-death self-relocations (see "Why a flag"). Waiting triggers checked: #6 and #7 (mode is `serial`), #8 (no `deck.go` behavior change), #15 (no turn hook), #20 (no wave C shape task), #50 (no task claims the exhaust hook), #51 (no "if you do" task) and #55 (this task is not `done`; the lead rewords the design note's Stack line in the merge commit). None has fired.
  - Next: dispatch `kernel-engineer` on Gemini Flash through `agy` (D3).
- 2026-10-25 kernel-engineer: Implemented directive #53 pure move. Replaced StackItem.Trigger ability.Trigger with Effect ability.Effect and Tracked bool in pkg/engine/stack.go, updated flushTriggers in pkg/engine/trigger.go to initialize Effect: p.trig.Effect and Tracked: p.trig.Hook == ability.HookKilled, and updated resolveAbility and relocateSelf in pkg/engine/effects.go to read it.Effect and !it.Tracked. Verified acceptance cases:
  1. Existing tests unchanged: `git diff main -- '*_test.go'` empty; `go test ./...` passed.
  2. Starter sim byte-identical: `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` baseline vs new compared with `fc.exe /b` -> FC: no differences encountered.
  3. Stress sim byte-identical: `cardtown sim` baseline vs new compared with `fc.exe /b` -> FC: no differences encountered.
  4. No trigger on the item: `git grep -nE '\bit\.Trigger\b' -- 'pkg/engine/*.go' ':!pkg/engine/*_test.go'` returned nothing; `git grep -n 'ability.Trigger' -- pkg/engine/stack.go` returned nothing.
  5. One place sets the new fields: `git grep -nE 'Tracked:|Effect: p\.trig' -- 'pkg/engine/*.go' ':!pkg/engine/*_test.go'` returned only `pkg/engine/trigger.go:132`.
  6. Only these edits: `git diff main --stat` lists only `pkg/engine/stack.go`, `pkg/engine/trigger.go`, `pkg/engine/effects.go` and this task file.
  7. `go tool task check` and `go tool task sim` passed clean.
  - Next: lead-architect review.
- 2026-10-25 lead-architect: Code review (D5): **approved**, merged into `main` with `--no-ff`, status `done`. Directives #53 and #55 applied.
  - Dispatch history: the first headless `agy` run (conversation `cbaeed40-…`) was refused at step 48, when it ran the built baseline exe; neither that nor `cmp`/`fc.exe` is on the headless allowlist. The lead reported it to `workflow-steward`, which added `go tool task sim-compare` (20b6152, no permission change). The engineer's entry above came from a later run.
  - Scope: c251af7 changes only `stack.go`, `trigger.go` and `effects.go`, plus this file. No test file changed. The field and comment text is the lead's text verbatim. `Tracked` is set from `p.trig.Hook == ability.HookKilled`, exactly the condition `relocateSelf` used to read, so `FR-ZON-004` is unchanged.
  - Gates, re-run by the reviewer on the merge result (a scratch worktree at `main` 20b6152 with this branch merged, uncommitted): `go tool task sim-compare -- main` reports `[PASS]` (starter and stress sims byte-identical, exit codes included). `go tool task check` and `go tool task sim` exit 0, all invariants hold, and no monitor reported a violation. Case 4 prints nothing; case 5 prints only `trigger.go:132`.
  - Directive #55: the design note's Stack line now describes `StackAbility` as built, in this merge commit.
  - Next: directive #54 (one place for the battlefield orders; harness, flash, pure move), then T-0055.
