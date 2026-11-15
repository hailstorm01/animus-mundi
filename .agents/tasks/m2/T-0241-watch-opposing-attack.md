---
id: T-0241
title: A trigger on another asset's attack, and "that attacking unit" as a step subject (wave H shape, sonnet; Blind Defile's shape)
kind: shape
tier: sonnet
wave: H
status: tests
branch: m2/autoloop-wave-h/T-0241-watch-opposing-attack
depends_on: []
claims: [shape:watching trigger (ability.Trigger.Watches and Watch; fireAttacks fires matching watchers), shape:watched-asset subject (ability.SubjectWatched, StackItem.Watched)]
cards: []
requirements: [FR-CBT-002, FR-TRG-003, FR-TGT-001, FR-ZON-004, FR-MOD-001]
---

## Goal
Wave H's first card, *Blind Defile* ("Whenever an opposing unit attacks, you may kill this structure and that attacking unit."), needs two things the kernel lacks: a trigger that fires on **another** asset's event (every hook today is a self hook: `fireHook` fires only the event's own asset's triggers, [trigger.go:136](../../../pkg/engine/trigger.go)), and a step subject for the asset whose event fired it ("that attacking unit"). This task builds both, for the attack event only. The card itself is T-0242 (`queued`), which waits for the ruling question [`2026-11-15-wave-h-blind-defile-kill-both`](../../../docs/design/outcomes/2026-11-15-wave-h-blind-defile-kill-both.md) (design note Open Question 22): whether the two kills depend on each other only decides whether the card's second kill step is `IfDone`, so this shape doesn't wait for it.

Covers:
- `Trigger.Watches` and `Trigger.Watch`, read by `fireAttacks` (Shape below).
- `SubjectWatched` and `StackItem.Watched`.
- Not this task: the card (T-0242); watching any hook other than `HookAttacks` (no M2 card watches another; the first card that does adds its hook's watcher call, through the same gate code); the `doc.go` line about self hooks (T-0242 updates it with the card, to keep this task at four files).

## Inputs
- FRD 01 v3.25: `FR-CBT-002` Step 2 (lines 575–583, Step 2 at 580: attack triggers, "opposing defensive triggers like *Blind Defile*", enqueue on top of the Attack action in APNAP order) and Step 3's target check (line 585: an attacker that has left the battlefield fizzles the Attack); `FR-TRG-003` (lines 450–454: the non-active player's triggers are placed on top and resolve first); `FR-TGT-001` Untargeted Selections (line 388: a clause without "chosen" doesn't target, so Ward and `FR-TGT-003` don't apply); `FR-ZON-004` (line 248: an asset that leaves the battlefield and returns is a new asset); `FR-MOD-001` §7 (lines 431–432: "you may" asked on placement, `SVY-R9`; not skipped when a step isn't a target, `SVY-R48`).
- Survey: `docs/design/m2_starter_set_ability_survey.md` line 182 (*Blind Defile*: "opposing unit attacks (may, kill self) → `KILL` attacking unit", a non-active player's trigger) and the wave B row (line 30, "opposing-attack hooks").
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6m (wave H, "From wave B") and Open Question 22.
- Code: `fireHook`, `bufferTrigger`, `fireAttacks`, `placeTrigger` and `pendingTrigger` ([trigger.go](../../../pkg/engine/trigger.go)); `declareAttack` ([combat.go:27-37](../../../pkg/engine/combat.go)); `StackItem` ([stack.go](../../../pkg/engine/stack.go)); `stepAssets` ([effects.go:397-430](../../../pkg/engine/effects.go)); `matchesFilter` ([filter.go:137](../../../pkg/engine/filter.go)); `couldDoNothing` ([placement.go:191](../../../pkg/engine/placement.go), already treats any non-`SubjectTarget` step as able to do something); `Trigger` and `Subject` ([ability.go](../../../pkg/engine/ability/ability.go)).
- Existing equivalents searched (D2.3): `SubjectSelf` (the item's source asset) and `SubjectReplaced` (the asset a replacement replaces, held on the `resolution`) are the nearest subjects. Neither names the asset whose event fired a trigger, and each reads its asset from a different place, so `SubjectWatched` is a new case of `stepAssets`, not a copy of a rule. No trigger watches another asset today (`git grep -n "Watch" pkg` finds nothing in the engine).

### Shape (the engineer builds it)
- **`ability.Trigger.Watches bool`** and **`ability.Trigger.Watch Filter`**. `Watches` false, the zero value, is every trigger today: it fires on its own source's event. `Watches` true makes the trigger fire on the event of **each battlefield asset that matches `Watch`**, judged against the trigger's controller and source when the event happens (`Allegiance` relative to the trigger's controller; `Other` excludes the trigger's own source) — and never on its own source's event unless the source matches `Watch`. Its comment says it is read only for `HookAttacks` so far.
- **`fireAttacks(a)`** keeps firing `a`'s Aggressive and own attack triggers as today, then fires the watchers: for every battlefield asset `w` (any player, any zone; a face-down asset prints nothing), each `HookAttacks` trigger with `Watches` whose `Watch` matches `a`. A trigger with `Watches` never fires as a self hook, and one without it never fires as a watcher, so `TestFR_CBT_002_OnlyTheAttackerFiresItsAttackTrigger` still holds. Every gate (`ActiveZone`, `Turn`, `Cornered`, `MaxPerTurn`) is checked in the one place `fireHook` checks them today: extend `fireHook` (for example with the watched asset, nil for a self hook) rather than writing a second loop of gates. Each buffered watcher's source is `w`, its controller `w`'s controller, and it records `a`'s asset ID (`pendingTrigger`, as `fireCombatKill` and `fireMoved` mark theirs). The flush then places them like any trigger: APNAP, so an opposing watcher sits above the attacker's own triggers (`FR-TRG-003`).
- **`StackItem.Watched int`**: the asset ID of the asset whose event fired the item's watching trigger, 0 for none. `placeTrigger` copies it from the pending trigger.
- **`ability.SubjectWatched`**: an asset step acts on the item's watched asset, if that same asset is still on the battlefield (`g.Asset(it.Watched)`, as `SubjectSelf` reads its source); once it has left, even if its card has come back as a new asset (`FR-ZON-004`), the step does nothing (so a following `IfDone` step doesn't run). It is not a target: no Ward, no `FR-TGT-003` check, nothing asked (`FR-TGT-001` Untargeted Selections). Add it at the end of the `Subject` constants so no existing value changes.
- Behavior doesn't change for any printed card: no definition uses `Watches` or `SubjectWatched` yet. `go tool task sim-compare -- autoloop/m2-wave-h` must be byte-identical (a gate here).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/watch_trigger_test.go`, with synthetic cards: a watcher structure printing a `HookAttacks` trigger with `Watches` and `Watch: {Allegiance: AllegianceOpposing, Kinds: KindUnit}`, and attackers built like `striker` in `attack_trigger_test.go`. To see items between placement and resolution, drive the steps directly as `locked_zone_test.go` does (`declareAttack` or `s.do(s.attack(...))` with a chooser that stops, then `s.g.Stack`), or read the event log.
- `TestFR_CBT_002_WatcherFiresOnOpposingUnitAttack`: player 1's watcher; player 0's unit attacks; the watcher's trigger goes on the stack above the Attack, controlled by player 1, with the watcher as its source and the attacker as its watched asset.
- `TestFR_CBT_002_WatcherIgnoresUnmatchedAttacks`: (a) a unit allied to the watcher attacks (on its controller's turn): the watcher doesn't fire; (b) the opposing Headquarters attacks: the watcher doesn't fire (`Kinds` unit only).
- `TestFR_CBT_002_WatcherIsNotASelfHook` (the reverse of the existing `TestFR_CBT_002_OnlyTheAttackerFiresItsAttackTrigger`): a unit printing a watching trigger whose `Watch` is opposing units attacks itself: its watching trigger doesn't fire.
- `TestFR_TRG_003_OpposingWatcherResolvesBeforeAttackerTriggers`: the attacker has Aggressive 1 (or an attack trigger that logs); the watcher's trigger resolves first (it sits above them, `FR-TRG-003`).
- `TestFR_TRG_003_TwoWatchersOneAttack`: player 1 has two watchers; one attack fires both, and player 1 is asked to order them (`DecideOrder`), as for any two simultaneous triggers.
- `TestFR_TGT_001_WatchedSubjectKillsTheAttacker`: the watcher's effect is `OpKill` on `SubjectWatched`: on resolution the attacker is in its owner's discard pile, the Attack then fizzles (`FR-CBT-002` Step 3), and the defender takes no damage.
- `TestFR_TGT_001_WatchedSubjectIsNotATarget`: the attacker has Ward; the kill still happens, and nothing is asked of the watcher's controller for it.
- `TestFR_TGT_001_WatchedSubjectGoneDoesNothing` (`FR-ZON-004`): the attacker leaves the battlefield before the watcher's trigger resolves (set up the item on the stack and remove the attacker, for example with `s.g.leaveBattlefield`): the kill step kills nothing, and a following `IfDone` step (for example the controller draws a card) doesn't run.
- `TestFR_MOD_001_OptionalWatcherIsAsked` (`SVY-R9`, `SVY-R48`): an `Optional` watcher whose only step is on `SubjectWatched` is asked "you may" on placement; declined, nothing is placed and the attack resolves normally.
- **Existing tests:** every other test passes unchanged.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test author keeps any small builder in the test file.

## Allowed Files
- `pkg/engine/ability/ability.go` (`Trigger.Watches`, `Trigger.Watch`, `SubjectWatched`)
- `pkg/engine/trigger.go` (`fireAttacks` fires watchers through `fireHook`'s gates; `pendingTrigger` records the watched asset; `placeTrigger` copies it)
- `pkg/engine/stack.go` (`StackItem.Watched`)
- `pkg/engine/effects.go` (`stepAssets` handles `SubjectWatched`)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "TestFR_(CBT_002_Watcher|TRG_003_(OpposingWatcher|TwoWatchers)|TGT_001_WatchedSubject|MOD_001_OptionalWatcher)"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (no printed card watches yet)

## Log
- 2026-11-15 lead-architect (wave H run, session 1): written for *Blind Defile*'s shape (design note 6m, "From wave B"). The shape and the card touch six non-test files, so under the D2 size rule the card is T-0242 (`queued`), which also waits for the ruling question committed on `main` (07bd2da). Set to `tests`.
- 2026-11-15 rules-test-author: tests written in `pkg/engine/watch_trigger_test.go`: TestFR_CBT_002_WatcherFiresOnOpposingUnitAttack, TestFR_CBT_002_WatcherIgnoresUnmatchedAttacks (subtests AlliedUnit, OpposingHeadquarters), TestFR_CBT_002_WatcherIsNotASelfHook, TestFR_TRG_003_OpposingWatcherResolvesBeforeAttackerTriggers, TestFR_TRG_003_TwoWatchersOneAttack, TestFR_TGT_001_WatchedSubjectKillsTheAttacker, TestFR_TGT_001_WatchedSubjectIsNotATarget, TestFR_TGT_001_WatchedSubjectGoneDoesNothing, TestFR_MOD_001_OptionalWatcherIsAsked. Helpers requested: none (small builders are in the test file; it reuses recordPlaced, declareThen, scriptPlan vars `no`, kindsOf). `go vet` fails only on the not-yet-existing `Trigger.Watches`/`Watch`, `SubjectWatched` and `StackItem.Watched`; `cardtown lint rules` passes. Left out: the "card comes back as a new asset" variant of the gone case (the test removes the attacker only; same-ID check is the engineer's `g.Asset(it.Watched)`), and watching hooks other than attacks (out of scope).
- 2026-11-15 lead-architect (wave H run, session 1): tests reviewed (e462bea); set back to `tests`. Notes:
  1. `TestFR_TGT_001_WatchedSubjectGoneDoesNothing` can't pass: `declareThen` runs its callback right after `declareAttack`, and Step 2 triggers are only buffered there; `resolveStack` places them (`flushTriggers`, [stack.go:112](../../../pkg/engine/stack.go)). So the stack holds only the Attack inside the callback, and the `len(s.g.Stack) != 2` check fails whatever the engine does. Call `s.g.flushTriggers()` at the start of the callback, then check the two items and remove the attacker. Everything else stays as it is.
  The other eight tests match the Tests list and `FR-CBT-002`, `FR-TRG-003`, `FR-TGT-001`, `SVY-R9`, `SVY-R48`.
