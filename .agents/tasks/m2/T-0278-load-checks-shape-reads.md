---
id: T-0278
title: "`LoadCards` rejects a card definition that sets a field where nothing reads it (directive #162; wave H.1 harness, sonnet)"
kind: harness
tier: sonnet
wave: H.1 (board directive #162)
status: review
branch: m2/autoloop-wave-h.1/T-0278-load-checks-shape-reads
depends_on: [T-0277]
claims: [func:checkShapeReads]
cards: []
requirements: []
---

## Goal
`loadCards` ([cardload.go:28](../../../pkg/engine/cardload.go)) rejects an `ability.Def` that sets a field in a context no engine code reads it in. Today such a def loads and then does nothing, or does the wrong thing, without any error. The load error names the card and the field, the same way the existing "ability text … is not an unclaimed non-keyword rules-text line" error does. Every `cards.Defs` entry still loads, so match output is unchanged. When this merges, #162 is `applied`.

## Inputs
- Board directive #162 (structural review 13, finding 4), approved by the user 2026-11-18, `open: before the next Deliverables task`. In wave H.1 it comes after #161 (applied, T-0276). Design note Deliverable 6n.
- What reads each field today (at `autoloop/m2-wave-h.1` d2eb3f77):
  - `Trigger.Watches`/`Trigger.Watch`: `fireHook` ([trigger.go:155](../../../pkg/engine/trigger.go)), which is only ever called with a watched asset for `HookAttacks`. The field comment ([ability.go:263-268](../../../pkg/engine/ability/ability.go)) says "Read only for HookAttacks so far".
  - `Step.Split`: read only by `stepAddArmor` ([effects.go:239](../../../pkg/engine/effects.go)), through `splitAssets`. The field comment ([ability.go:475](../../../pkg/engine/ability/ability.go)) requires `SubjectPick` over assets and says "UpTo must be 0".
  - `Step.ControllerChooses`: read only by `stepDiscard` ([effects.go:320](../../../pkg/engine/effects.go)). Its comment ([ability.go:474](../../../pkg/engine/ability/ability.go)) says "OpDiscard only".
  - An aura's `Static.Filter`: `Game.auras` reads it through `matchesFromSource`. A `Filter.Stat` there would read `Game.Power`, which reads `auras` again and recurses. The field comment ([ability.go:74-78](../../../pkg/engine/ability/ability.go)) says "Its Stat stays StatNone".
- Where a `Step` can sit in a `Def`: `Def.Effect`, `Def.Triggers[i].Effect`, `Def.Activated[i].Effect`, any of these effects' `Modes` (each an `Effect` in its own right, which may in principle hold further `Modes`), and `Def.Replacements[i].Steps`. The check has to reach every one of them. If it misses a place, the misuse there goes undetected, which is the exact problem this task fixes.
- Design (lead decisions):
  - **Placement:** non-test code in `cardload.go`, as the directive says: `func checkShapeReads(d *ability.Def) error`, plus whatever unexported helpers it needs (for example, one that walks an `Effect` and its `Modes`, and one that checks a `[]ability.Step`). `loadCards` calls it on each def it attaches, right after the def's `Text` lines are claimed. Defs are still taken in name order, so the first error stays stable (`TestLoadCards_FirstErrorStable`). A def whose card is not in `dir` is skipped, as it is today.
  - **Error text:** `card %q: <field> <rule>`, where `<field>` is one of `Trigger.Watches`, `Step.Split`, `Step.ControllerChooses` or `Static.Filter.Stat`, and `<rule>` is a short clause saying where the field is read, for example `card "X": Trigger.Watches is read only for HookAttacks`. The tests match only the card name and the `<field>` text.
  - **The rules, exactly:**
    1. A trigger with `Watches` set has `Hook == ability.HookAttacks`.
    2. A step with `Split` set has `Op == ability.OpAddArmor`, `Subject == ability.SubjectPick`, `UpTo == 0` and `Filter.Zone != ability.ZoneDiscard` (assets, not pile cards).
    3. A step with `ControllerChooses` set has `Op == ability.OpDiscard`.
    4. A `Static` of kind `StaticAura` has `Filter.Stat == ability.StatNone`.
  - **The checker's doc comment** says that a shape task that adds a new reader of one of these fields widens the matching rule in the same task.
  - **Field comments:** each of the four field comments in `ability.go` gains one short clause saying that `LoadCards` rejects the misuse (for `Watches`: "LoadCards rejects it on any other hook"). The fields themselves don't change.
  - `docs/architecture/engine-structs.md`'s `cardload.go` row ([:277](../../../docs/architecture/engine-structs.md)) gains `checkShapeReads`. The design note's Boundaries are #163's job, not this task's.
- Reuse search (D2.3): `git grep -n -i -E "func (check|valid|verify)" -- pkg/engine` at d2eb3f77 finds only `checkDefs` and `checkStarterSet` in `defs_test.go`. Those check coverage (a def matching a loaded card, non-empty `Text`, the `NotYetExecutable` list), not field reads, so nothing is duplicated. No step walker exists yet (`git grep -n -E "func .*(walk|each|all)Steps" -- pkg`).
- Real defs: the only uses of these fields in `pkg/engine/cards` are `Watches` on a `HookAttacks` trigger ([path.go:200](../../../pkg/engine/cards/path.go)), `Split` on an `OpAddArmor` `SubjectPick` step ([path.go:547](../../../pkg/engine/cards/path.go)) and `ControllerChooses` on an `OpDiscard` step ([animus.go:311](../../../pkg/engine/cards/animus.go)). No aura filter has a `Stat`. All of them pass.
- Size (D2.4): two non-test files (`cardload.go`, `ability/ability.go`, the latter for comments only), plus one doc.
- **Expected output difference:** none: `go tool task sim-compare` passes.

## Commits
1. `test(T-0278): LoadCards rejects a field set where nothing reads it` with the new tests in `pkg/engine/card_test.go`. Before committing, run `go test ./pkg/engine -run "TestLoadCards_"` and record that the new rejection tests fail.
2. `feat(engine): LoadCards checks where each shape is read (T-0278)` with the code, comment and doc edits.

## Acceptance Cases (harness tasks only; written by the lead)
Each case is a test in `pkg/engine/card_test.go`. Use the existing `fixtureDir`, `wantLoadError` and `mustLoad` helpers, and a def on `"Test Unit"` whose `Text` is `[]string{lineA}` plus the field under test. Each rejection must name `"Test Unit"` and the `<field>` text given.
1. `TestLoadCards_WatchesOffAttacksRejected`: a trigger with `Hook: HookDeploy, Watches: true` → load error naming `Trigger.Watches`.
2. `TestLoadCards_SplitOffAddArmorRejected`: a `Def.Effect` step `{Op: OpDealDamage, Subject: SubjectPick, Split: true}` → load error naming `Step.Split`.
3. `TestLoadCards_SplitWithUpToRejected`: `{Op: OpAddArmor, Subject: SubjectPick, UpTo: 2, Split: true}` → load error naming `Step.Split`.
4. `TestLoadCards_SplitOffPickRejected`: `{Op: OpAddArmor, Subject: SubjectAll, Split: true}` → load error naming `Step.Split`.
5. `TestLoadCards_SplitOverPileRejected`: `{Op: OpAddArmor, Subject: SubjectPick, Split: true, Filter: Filter{Zone: ZoneDiscard}}` → load error naming `Step.Split`.
6. `TestLoadCards_ControllerChoosesOffDiscardRejected`: `{Op: OpDraw, Subject: SubjectController, Amount: Value{N: 1}, ControllerChooses: true}` → load error naming `Step.ControllerChooses`.
7. `TestLoadCards_AuraFilterStatRejected`: `Statics: {{Kind: StaticAura, Filter: Filter{Stat: StatPower, AtMost: 3}, Power: 1}}` → load error naming `Static.Filter.Stat`.
8. `TestLoadCards_ShapeCheckReachesEveryStep`: a table test. In each row, the bad step from case 2 sits in a different place: `Def.Effect.Steps`, a mode in `Def.Effect.Modes` (with `ChooseN: 1`), `Def.Triggers[0].Effect.Steps` (on `HookDeploy`), `Def.Activated[0].Effect.Steps` and `Def.Replacements[0].Steps`. Every row → load error naming `Step.Split`.
9. `TestLoadCards_ShapeReadsWhereReadLoad`: one def that sets every field where it is read: a `HookAttacks` trigger with `Watches: true`, an `OpAddArmor` `SubjectPick` step with `Split: true` and `UpTo` 0, an `OpDiscard` step with `ControllerChooses: true`, and an aura whose filter has no `Stat` → `mustLoad` succeeds and `Ability` is the def pointer.
10. **Real data:** `TestStarterSet_Executable` (unedited) still passes, because it calls `LoadCards("../../data/cards")` and fails on any load error. So every `cards.Defs` entry loads.
11. **Existing tests unedited:** `git diff autoloop/m2-wave-h.1 -- pkg/engine/card_test.go` shows only additions, and no other `_test.go` file changes.
12. `go tool task check` passes. `go tool task sim-compare -- autoloop/m2-wave-h.1` prints `[PASS]` (starter and stress sims byte-identical to the run branch).

## Allowed Files
- `pkg/engine/cardload.go`
- `pkg/engine/ability/ability.go` (the four field comments only)
- `docs/architecture/engine-structs.md` (the `cardload.go` row only)
- `pkg/engine/card_test.go` (new tests only, in the `test(T-0278)` commit)

## Done When
- [ ] The new tests pass: `go test ./pkg/engine -run "TestLoadCards_"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h.1` passes

## Log
- 2026-11-18 lead-architect (wave H.1, session 7): wrote the task for directive #162 and set it `ready` (a harness task, so it skips `tests`). I made it tier `sonnet`, not `flash`, because the checker has to reach every place a step can sit, `Modes` included, and a place it misses fails silently. The acceptance cases put the same bad step in each place. `main` (7a8e4f5) has nothing the run branch lacks, so there was no sync. #163 and #164 stay `open: before the next Deliverables task`, with #163 next.
- 2026-11-18 kernel-engineer (sonnet): wrote the acceptance tests first (commit decbb69a); before the code, `go test ./pkg/engine -run "TestLoadCards_"` failed the 7 rejection tests and the every-step table, and the load-ok case passed as expected. Then added `checkShapeReads` plus `checkEffectReads` (recurses into Modes) and `checkStepReads` in `cardload.go`, called from `loadCards` after Text is claimed; error is `card %q: <field> <rule>`. Added a short "LoadCards rejects ..." clause to the four field comments in `ability.go` and `checkShapeReads` to the `cardload.go` row of `engine-structs.md`. Reuse: no existing checker or step walker (`ReadsZoneRef` walks Modes but answers a different question). Extra table rows beyond the task: nested modes and trigger/activated modes. Ran: `go tool task check` PASS, `go tool task sim` clean, `go tool task sim-compare -- autoloop/m2-wave-h.1` [PASS]. Review: the error strings start with a capital field name (Go lint ST1005 style), deliberate per the task's `<field>` format.
