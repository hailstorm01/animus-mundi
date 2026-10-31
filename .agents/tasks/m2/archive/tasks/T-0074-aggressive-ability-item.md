---
id: T-0074
title: Aggressive is an ability stack item, marked as coming from the keyword (directive #60 part 1, harness, sonnet)
kind: harness
tier: sonnet
wave: B (board directive #60 (1), blocking before T-0073)
status: done
branch: m2/T-0074-aggressive-ability-item
depends_on: [T-0072]
claims: [type:ability.Keyword, field:ability.Trigger.Keyword, func:ability.Aggressive, field:StackItem.Keyword, removes:StackAggressive, removes:StackItem.Amount, removes:pendingTrigger.item, removes:view.StackView.Amount]
cards: []
requirements: [FR-KWD-002f, FR-CBT-002, FR-TRG-003, FR-LAY-001]
---

## Goal
"Gets +N power until end of turn" is written once, in `stepModifyStat`. Aggressive N (`FR-KWD-002f`: "Whenever this unit attacks, it gets +N power until end of turn") becomes an `ability.Trigger` built by one function in `ability`, placed as an ordinary `StackAbility` item that carries a marker naming the keyword. `StackAggressive`, its `resolve` case and `StackItem.Amount` go. This is part 1 of directive #60. Part 2, Allshan reading "Cornered: Aggressive 3.", is T-0075.

## Inputs
- Board directive #60 (review of wave B run 1, approved by the user, 2026-10-26). Read its acceptance cases (a)–(f); the cases below are the same, made concrete.
- FRD 01: `FR-KWD-002f` (Aggressive), `FR-CBT-002` (Step 2 triggers), `FR-TRG-003` (a player orders their own simultaneous triggers), `FR-LAY-001`/`FR-LAY-003` (Layer 5 power until end of turn, expiring in Sub-Phase 4c).
- Directive #57: a player's simultaneous triggers fall back to source position; Aggressive is still buffered before the printed attack triggers of the same source, so with the stable sort it is listed, and by default placed, first.
- Survey `R18` (user, 2026-10-26): a keyword behind Cornered counts only while Cornered. Nothing reads "has Aggressive" yet, so this task builds nothing for it.
- Design note: `docs/architecture/m2-ability-model.md`, Key types, **Stack** (the lead added `Keyword` there in the commit that wrote this task).

### Where it is today (at `autoloop/m2-wave-b` ef23c3b)
| Where | Today |
| :--- | :--- |
| `pkg/engine/stack.go:15` | `StackAggressive` in the `StackKind` list, between `StackScratch` and `StackAbility` |
| `pkg/engine/stack.go:27-30` | `Source` comment names "Aggressive source"; `Amount int // StackAggressive: N` |
| `pkg/engine/stack.go:54-55` | `push`: `case StackAggressive: what = "Aggressive trigger"` |
| `pkg/engine/stack.go:95-102` | `resolve`: `case StackAggressive` adds `it.Amount` to `TempPower` and logs a `"trigger"` event |
| `pkg/engine/trigger.go:21` | `pendingTrigger.item *StackItem`, used only by Aggressive |
| `pkg/engine/trigger.go:118-129` | `fireAttacks` buffers a ready-made `StackAggressive` item from `g.Keywords(a).Aggressive` |
| `pkg/engine/trigger.go:173-177` | `flushTriggers` pushes `p.item` when set |
| `pkg/view/view.go:81-94, 171-172` | `StackView.Kind` lists "Aggressive"; `StackView.Amount`; `stackKinds` maps `StackAggressive`; `stackView` copies `Amount` |

## Edits (lead decisions, binding)

### 1. `pkg/engine/ability/ability.go`
- A new type, after `TurnOwner`:
  ```go
  // Keyword names the keyword a trigger comes from, for a keyword that is a
  // triggered ability. The zero value is a printed ability.
  type Keyword int

  const (
  	KeywordNone       Keyword = iota // the zero value: a printed ability
  	KeywordAggressive                // "Aggressive N" (FR-KWD-002f)
  )

  // String returns the keyword's printed name, or "" for KeywordNone.
  func (k Keyword) String() string
  ```
- `Trigger` gains, after `Optional`: `Keyword Keyword // the keyword this trigger comes from; zero for a printed ability`.
- One constructor, the only place the meaning of Aggressive N is written:
  ```go
  // Aggressive is the trigger the keyword "Aggressive N" means: whenever this
  // unit attacks, it gets +N power until end of turn (FR-KWD-002f).
  func Aggressive(n int) Trigger
  ```
  It returns `Trigger{Hook: HookAttacks, Keyword: KeywordAggressive, Effect: Effect{Steps: []Step{{Op: OpModifyStat, Subject: SubjectSelf, Amount: Value{N: n}}}}}`. No other field is set (T-0075 sets `Cornered` on its copy).

### 2. `pkg/engine/stack.go`
- Remove `StackAggressive` from the `StackKind` list. Remove `StackItem.Amount`. Drop "Aggressive source" from the `Source` comment.
- `StackItem` gains, after `Tracked`: `Keyword ability.Keyword // StackAbility: the keyword its trigger came from (ability.KeywordAggressive); zero for a printed ability`.
- `push`, `case StackAbility`: when `item.Keyword != ability.KeywordNone`, `what` is `item.Keyword.String() + " trigger of " + e.Card` (for example "player 0 places Aggressive trigger of Raider on the stack"); otherwise "ability of " + e.Card as today. Remove `case StackAggressive`.
- `resolve`: remove `case StackAggressive`. The `"trigger"` event it logged goes with it; a resolving `StackAbility` logs nothing today, and Aggressive now resolves like one.

### 3. `pkg/engine/trigger.go`
- Remove `pendingTrigger.item`.
- `fireAttacks` keeps its order: Aggressive first, then `g.fireHook(a, ability.HookAttacks)`. It reads `g.Keywords(a).Aggressive` when the attack happens (never at card load, because a later layer may grant or remove the keyword) and buffers `pendingTrigger{source: a.ID, inst: a.Inst, controller: a.Controller, trig: ability.Aggressive(n), label: fmt.Sprintf("%s: Aggressive %d", a.Printed().Name, n), pos: posOf(a)}`. The label is unchanged.
- `flushTriggers`: remove the `p.item` branch. The `StackAbility` item it pushes also sets `Keyword: p.trig.Keyword`.
- Update `fireAttacks`'s doc comment: its Aggressive keyword trigger is an `ability.Aggressive` trigger.

### 4. `pkg/view/view.go`
- `StackView`: remove `Amount`. Add `Keyword string // Ability: the keyword its trigger came from ("Aggressive"); empty for a printed ability`. Drop "Aggressive" from the `Kind` comment and "or Aggressive source" from `Source`.
- `stackKinds`: remove the `StackAggressive` entry. `stackView`: remove its case; in `case engine.StackAbility`, set `s.Keyword = it.Keyword.String()`.

Nothing else changes. `cmd/cardtown/play_screen.go` prints `item.Kind` only, so an Aggressive item now shows as "Ability"; leave it.

## Acceptance Cases (written by the lead; the engineer turns them into tests)
Put the engine tests in a new `pkg/engine/aggressive_item_test.go` and the view test in a new test function in `pkg/view/view_test.go`. Commit them first as `test(T-0074): …`; each must fail (or not compile) before the implementation commit. Build test units with the existing helpers (`unit`, `striker`, `attackTrigger`, `scratchN`, `s.script`, `s.asked`, `s.attack`). If a helper is missing, add it to `pkg/engine/scenario_test.go` (setup only).

1. **(a) Nothing else changes.** Every existing test passes unchanged. No existing test is edited.
2. **(b) Listed and placed first, still reorderable.** Raider (power 1, Aggressive 2, with one printed attack trigger `scratchN(1)`) attacks. Player 0 is asked one `DecideOrder` whose option 0 label is `"Raider: Aggressive 2"` and option 1 is the printed trigger. With a timeout (the fixed-order fallback), the Aggressive item is placed first, so the printed trigger is on top of it: the `place` events come in the order Attack, Aggressive, printed (stack depths 1, 2, 3). With scripted picks `[1 0]`, the printed trigger is placed first and the Aggressive item last (on top).
3. **(c) The item.** To see a placed item before it resolves, set `s.g.obs` to the `recorder` from `events_test.go` with an `onEvent` that copies `*s.g.Stack[len(s.g.Stack)-1]` on each `place` event. The placed Aggressive item is `Kind: StackAbility`, `Keyword: ability.KeywordAggressive`, `Source` the attacker, `Controller` its controller, and its `Effect` equals `ability.Aggressive(2).Effect`. A printed attack trigger's item has `Keyword: ability.KeywordNone`. Its `place` event's `Detail` is `"player 0 places Aggressive trigger of Raider on the stack"`, and a printed one's still reads `"... places ability of Raider on the stack"`.
4. **(c) The constructor.** `ability.Aggressive(3)` is a `HookAttacks` trigger with `Keyword: KeywordAggressive`, not `Optional`, not `Cornered`, no targets, and exactly one step `{Op: OpModifyStat, Subject: SubjectSelf, Amount: Value{N: 3}}`. `KeywordAggressive.String()` is `"Aggressive"`, `KeywordNone.String()` is `""`.
5. **(d) Read when the attack happens.** Two cases, each changing the attacker's printed `Keywords` after `s.put` and before the attack (a synthetic Keywords change; no card grants Aggressive yet): (i) a unit put with Aggressive 2, then set to Aggressive 0, attacks: no Aggressive item is placed and its power stays as printed; (ii) a unit put with no Aggressive, then set to Aggressive 2, attacks: one Aggressive item is placed and the defender takes power + 2.
6. **(d) Resolves like the keyword did.** Brawler (power 1, Aggressive 2) attacks a 0/10 Dummy: Dummy takes 3; after `s.endTurn()` Brawler's power is 1 again. If the attacker leaves the battlefield before the Aggressive item resolves, the item fizzles and nothing errors (use the existing order-and-kill pattern in `source_power_test.go`, or skip this sub-case if it needs a new shape, and say so in the Log).
7. **View.** A `StackAbility` item with `Keyword: ability.KeywordAggressive` shows as `Kind "Ability"`, `Keyword "Aggressive"`, its source card name in `Card`; a printed ability item shows `Keyword ""`.
8. **(f)** `git grep -n StackAggressive -- pkg cmd` finds nothing.
9. **(e)** `go tool task sim-compare -- autoloop/m2-wave-b` passes. If it does not, the Log names every changed line and why. Only wording and coverage counts may differ (no Animus or Path card prints Aggressive, so the starter sim should be byte-identical; stress decks may show the new place wording). Win rates and invariant results may not change.

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/stack.go`
- `pkg/engine/trigger.go`
- `pkg/view/view.go`
- Tests: new `pkg/engine/aggressive_item_test.go`; a new test function in `pkg/view/view_test.go`; setup-only helpers in `pkg/engine/scenario_test.go`

## Done When
- [x] The new tests pass: `go test ./pkg/engine -run 'Aggressive'` and `go test ./pkg/view`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-compare -- autoloop/m2-wave-b` passes, or the Log explains every changed line (case 9)
- [x] `git grep -n StackAggressive -- pkg cmd` prints nothing

## Log
- 2026-10-26, lead-architect (wave B run, session 7): written from directive #60 (1), which blocks T-0073. Synced `main` into the run branch first (ef23c3b: #60, SVY-R18). Harness, tier `sonnet` as the directive says: it changes the trigger pipeline every attack runs through. Four non-test files, within the size rule; the placement event shows the keyword in its `Detail` rather than a new `Event` field, which would have been a fifth file and has no reader yet. Searched for an existing equivalent: `stepModifyStat` is the one writer kept; `StackAggressive` is the one removed. No other keyword is a triggered ability today (Unbowed and Cornered are printed-ability prefixes, `Trigger.Cornered` and `HookDamaged`), so `Keyword` has one value. Harness task, so straight to `ready`.
- 2026-10-26, kernel-engineer (sonnet): tests first in 912ba16 (`aggressive_item_test.go`, 7 tests: listed/placed first by default, reorderable, item and place Detail wording, constructor and `Keyword.String`, read-at-attack both ways, resolves and expires, fizzle; plus `TestView_AbilityItemShowsItsKeyword` in `view_test.go`); they did not compile before the implementation. Fizzle sub-case done with a printed self-damage trigger (`OpDealDamage`, `SubjectSelf`, 20) and picks `[0 1]`, no new shape or helper. No `scenario_test.go` change. Reuse: `stepModifyStat` stays the one writer of "+N power"; `recorder` from `events_test.go` for placed items. Implemented the four files exactly as the Edits section says (`ability.Keyword`, `Trigger.Keyword`, `ability.Aggressive`; `StackAggressive`, `StackItem.Amount`, `pendingTrigger.item`, `StackView.Amount` removed; `StackItem.Keyword`/`StackView.Keyword` added). Removed the old `"trigger"` event ("... Aggressive N: +N power ...") that Aggressive's resolve logged; sim output is unaffected. Ran: `go test ./pkg/engine ./pkg/view` ok; `go tool task check` passed; `go tool task sim` clean; `go tool task sim-compare -- autoloop/m2-wave-b` PASS byte for byte (starter and stress); `git grep -n StackAggressive -- pkg cmd` empty. Look at: the existing event_facts test still passes unchanged (Aggressive place event now carries Card/Inst of the attacker).
- 2026-10-26, lead-architect (review, session 7): approved, round 1, no changes. Scope: only the four Allowed Files plus the new tests; no existing test edited, and no test touched after `test(T-0074)` (912ba16). Every acceptance case is covered; the fizzle case uses a printed self-damage trigger, no new shape. Fences: no card names, no new package. Reuse: `stepModifyStat` is now the only "+N power until end of turn" writer; `ability.Aggressive` is the only place Aggressive N's meaning is written. Re-ran in a review worktree: `go tool task check` passes; `go tool task sim` clean, 200 games, P0 57.5% / P1 39.0% / draws 3.5% (unchanged), no monitor violation; `go tool task sim-compare -- autoloop/m2-wave-b` byte-identical (no starter or stress card prints Aggressive); `git grep -n StackAggressive -- pkg cmd` prints nothing. Nit, left for whoever next touches it: `ability.Trigger`'s doc comment still says "one printed triggered ability", though a keyword trigger is one too. Merged into `autoloop/m2-wave-b` as 125b9ab. Directive #60 (1) applied; (2) is T-0075, which goes next because #60 still blocks T-0073.
