---
id: T-0218
title: "\"Until end of turn\" is one list of modifiers on the asset (directive #137, wave F harness, flash, pure move)"
kind: harness
tier: flash
wave: F
status: done
branch: m2/autoloop-wave-f/T-0218-until-end-of-turn-modifiers
depends_on: []
claims: [type:Modifier, type:Duration, field:Asset.Modifiers, method:Asset.TempPower, method:Asset.expire]
cards: []
requirements: []
---

## Goal
`Asset.TempPower` (one `int`) becomes `Asset.Modifiers`, a list of modifiers, each with its layer, what it changes and its duration (#137). `Power` folds the Layer 5 power modifiers, `stepModifyStat` appends one, and Sub-Phase 4c expires the "until end of turn" ones in one place. The snapshot keeps `tempPower`, as their sum, so output doesn't change. A pure move: same behavior, same test names, same sim output. T-0219 (*Allshan's Battle Trance*) then adds a keyword modifier to the same list, which `Keywords` folds.

## Inputs
- Board directive #137 (structural review 10, finding 2, approved by the user, 2026-11-09). Its trigger fired in this task's commit: T-0219, whose `cards:` include *Allshan's Battle Trance* ("Until end of turn, it gains Stalwart and gets +2 power.", `SVY-R39`), is now written and `queued`, so #137 is `open: before T-0219`.
- FRD 01 `FR-LAY-001` (Layer 5 stat modifications), `FR-LAY-003` (duration), `FR-TRN-001` (Sub-Phase 4c).
- Design note: `docs/architecture/m2-ability-model.md`, Key decisions, **Layers** (line 55): "`Asset.TempPower` becomes a list of timestamped modifiers (Layers 2, 3 and 5, with a duration)". The lead rewrote that sentence in this task's commit to say what this task builds. "Timestamped" is the list's order (oldest first); nothing reads an order yet, since every modifier built so far adds.
- Why it is a pure move: `TempPower` has five non-test readers and writers, all at 48102df (`git grep -n TempPower -- pkg cmd`): its declaration (`state.go:76-77`), `Power` (`layers.go:12`), the 4c reset (`turn.go:36`), `stepModifyStat` (`effects.go:280`, `:285`) and `assetSnap` (`snapshot.go:64`). Nothing in `cmd/` or `pkg/view` reads the field; the view reads `AssetSnap.TempPower`, which stays. A new asset starts with no modifiers, as it started at `TempPower` 0 (`zones.go:213`), and a face-down transformation keeps the asset, so it keeps them (`TestFR_ENT_103_MorphTurnsUnitFaceDownInPlace`).
- Reuse search (D2.3): no modifier list, `Duration` type or expiry helper exists in `pkg/engine` or `pkg/engine/ability` (`git grep -n -i -e duration -e EndOfTurn -e UntilEnd`: only `OpModifyStat`'s comment). This task removes the only copy of "until end of turn", so it adds no duplicate.
- **Size:** five non-test files, one over the D2 size rule. Not split: each of the five holds a use of the one field being replaced, so no part compiles without the others (as T-0059 and T-0086, earlier pure moves).
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 48102df: `gofmt -l pkg` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-f` printed `[PASS]`, and `go tool task sim` exit 0 (200 games, P0 71.5% / P1 27.5% / draws 1.0%, P0 50/50 and P1 44/50 main cards executable). Every expected output below is from that dry run.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at 48102df. Commit the test edits (Tests, below) first as `test(T-0218): …`, then these five files as the implementation commit. The test commit doesn't compile on its own; that is its failing state.

### 1. `pkg/engine/state.go`, the field
Replace lines 76-77:
```go
	// TempPower is the Layer 5 "until end of turn" power modifier.
	TempPower int
```
with these three lines:
```go
	// Modifiers are the changes resolved effects made to its characteristics,
	// oldest first, each with its layer and duration (FR-LAY-001, FR-LAY-003).
	Modifiers []Modifier
```

### 2. `pkg/engine/layers.go`, the list and its reads
Replace lines 8-13, from `import "animus-mundi/pkg/engine/ability"` through the closing `}` of `Power`:
```go
import "animus-mundi/pkg/engine/ability"

// Power returns effective Combat Power, the only stat that clamps at 0.
func (g *Game) Power(a *Asset) int {
	return max(0, a.Printed().Power+a.TempPower)
}
```
with:
```go
import (
	"slices"

	"animus-mundi/pkg/engine/ability"
)

// Duration is how long a Modifier lasts (FR-LAY-003).
type Duration int

const (
	UntilEndOfTurn Duration = iota // expires in Sub-Phase 4c (FR-TRN-001)
)

// Modifier is one change a resolved effect made to an asset's characteristics.
// Layer is its FR-LAY-001 layer. The only kind built so far is a Layer 5 power
// modifier: Power is added to Combat Power.
type Modifier struct {
	Layer int
	Power int
	Until Duration
}

// TempPower is the sum of a's Layer 5 power modifiers.
func (a *Asset) TempPower() int {
	n := 0
	for _, m := range a.Modifiers {
		if m.Layer == 5 {
			n += m.Power
		}
	}
	return n
}

// expire removes a's modifiers that last d.
func (a *Asset) expire(d Duration) {
	a.Modifiers = slices.DeleteFunc(a.Modifiers, func(m Modifier) bool { return m.Until == d })
}

// Power returns effective Combat Power, the only stat that clamps at 0.
func (g *Game) Power(a *Asset) int {
	return max(0, a.Printed().Power+a.TempPower())
}
```

### 3. `pkg/engine/turn.go`, Sub-Phase 4c
Line 36: change `		a.TempPower = 0` to `		a.expire(UntilEndOfTurn)`.

### 4. `pkg/engine/effects.go`, `stepModifyStat`
Replace lines 279-280:
```go
// stepModifyStat gives each Subject +Amount power until end of turn: the Layer 5
// TempPower, cleared in Sub-Phase 4c (FR-LAY-001, FR-LAY-003). Power clamps at 0.
```
with:
```go
// stepModifyStat gives each Subject +Amount power until end of turn: a Layer 5
// modifier, which expires in Sub-Phase 4c (FR-LAY-001, FR-LAY-003). Power clamps at 0.
```
Line 285: change `		a.TempPower += n` to `		a.Modifiers = append(a.Modifiers, Modifier{Layer: 5, Power: n, Until: UntilEndOfTurn})`.

### 5. `pkg/engine/snapshot.go`, `assetSnap`
Line 64: change `TempPower: a.TempPower, Power:` to `TempPower: a.TempPower(), Power:`, and nothing else on the line. `AssetSnap.TempPower` (line 26) and its `json:"tempPower"` tag stay.

No other non-test file changes.

## Tests (authorized edits to existing tests; the `test(T-0218)` commit)
No new tests. Each test below keeps its name and its assertions; only its use of the field changes. A **write** `X.TempPower = N` becomes `X.Modifiers = []Modifier{{Layer: 5, Power: N}}` (in `pkg/view`, `[]engine.Modifier{...}`), with any trailing comment kept. A **read** `X.TempPower` (as a value, outside a string literal) becomes `X.TempPower()`. Strings such as `"TempPower %d"` stay. Line numbers are at 48102df.

| File | Line(s) | In | Change |
| :--- | :--- | :--- | :--- |
| `pkg/engine/cards_animus_test.go` | 1047 | `TestCard_CraterfallAncient_DamageEqualsEffectivePowerAtResolution` | write, `anc`, 2 |
| `pkg/engine/cards_path_allshan_test.go` | 33, 34 | `TestCard_AllshanChainbreakerRevenant_EntersReservesAsFiveSixForFiveRequisitionAndDoesNothing` | reads, `a` |
| `pkg/engine/cards_path_allshan_test.go` | 91, 92, 95, 96 | `TestCard_AllshanChainbreakerRevenant_CorneredAttackDealsEightAndPowerResetsAtEndOfTurn` | reads, `al` |
| `pkg/engine/cards_path_allshan_test.go` | 110, 111 | `TestCard_AllshanChainbreakerRevenant_NotCorneredAttackGainsNothing` | reads, `al` |
| `pkg/engine/cards_path_allshan_test.go` | 143, 144 | `TestCard_Allshan_CorneredAggressiveIsTheKeywordTrigger` | reads, `al` |
| `pkg/engine/cards_path_test.go` | 1127 | `TestCard_NecropolisHunt_DealsItsLastKnownEffectivePowerNotAFixedFive` | write, `hunt`, 2 |
| `pkg/engine/effect_events_test.go` | 94, 95 | `TestEffectEvents_PowerLogsThePowerChange` | reads, `a` and `b` |
| `pkg/engine/morph_scratch_test.go` | 39 | `TestFR_ENT_103_MorphTurnsUnitFaceDownInPlace` | `u.Exhausted, u.Armor, u.TempPower = true, 3, 2` → `u.Exhausted, u.Armor, u.Modifiers = true, 3, []Modifier{{Layer: 5, Power: 2}}` |
| `pkg/engine/morph_scratch_test.go` | 49, 50 | the same test | reads, `a` |
| `pkg/engine/move_zone_test.go` | 76 | `TestFR_ZON_004_MovedUnitKeepsItsState` | `x.Damage, x.Armor, x.TempPower, x.Exhausted = 2, 3, 1, true` → `x.Damage, x.Armor, x.Modifiers, x.Exhausted = 2, 3, []Modifier{{Layer: 5, Power: 1}}, true` |
| `pkg/engine/power_at_most_test.go` | 70, 72 | `TestFR_TGT_001_PowerAtMostReadsEffectivePowerOnTheBattlefield` | writes, `up` 1, `down` -2 |
| `pkg/engine/power_at_most_test.go` | 96 | `TestFR_TGT_003_PowerAtMostTargetCheckedAgainOnResolution` | write, `a`, 1 |
| `pkg/engine/power_at_most_test.go` | 118, 121, 123 | `TestFR_TGT_001_PowerAtMostAllStepReadsEffectivePower` | writes, `allyDown` -1, `foeUp` 1, `foeDown` -3 |
| `pkg/engine/relocate_test.go` | 34 | `TestFR_ZON_004_ReturnedTargetLeavesWithItsLastKnownState` | write, `e`, 2 |
| `pkg/engine/relocate_test.go` | 64, 65 | the same test | reads, `n` |
| `pkg/engine/rules_test.go` | 270 | `TestFR_LAY_001_CombatPowerClampsAtZero` | write, `u`, -5 |
| `pkg/view/view_test.go` | 143 | `newContentScene` (setup) | `s.sentry.Damage, s.sentry.Armor, s.sentry.Exhausted, s.sentry.TempPower = 1, 2, true, 3` → `s.sentry.Damage, s.sentry.Armor, s.sentry.Exhausted, s.sentry.Modifiers = 1, 2, true, []engine.Modifier{{Layer: 5, Power: 3}}` |

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Test names unchanged** (#137): `git -C W diff autoloop/m2-wave-f --numstat -- '*_test.go'` prints exactly these ten lines, each `N	N` (as many lines added as removed): `1	1	pkg/engine/cards_animus_test.go`, `10	10	pkg/engine/cards_path_allshan_test.go`, `1	1	pkg/engine/cards_path_test.go`, `2	2	pkg/engine/effect_events_test.go`, `3	3	pkg/engine/morph_scratch_test.go`, `1	1	pkg/engine/move_zone_test.go`, `6	6	pkg/engine/power_at_most_test.go`, `3	3	pkg/engine/relocate_test.go`, `1	1	pkg/engine/rules_test.go`, `1	1	pkg/view/view_test.go`. No other test file changes. All of them are in the `test(T-0218)` commit: `git -C W log --oneline --stat autoloop/m2-wave-f..HEAD -- '*_test.go'` lists only that commit.
2. **Sims byte-identical** (#137): `go -C W tool task sim-compare -- autoloop/m2-wave-f` prints `[PASS] starter and stress sims match autoloop/m2-wave-f byte for byte, exit codes included`.
3. **No field left:** `git -C W grep -n -e 'TempPower' -- pkg cmd ':!*_test.go'` prints exactly five lines: `pkg/engine/layers.go:30` (`// TempPower is the sum of a's Layer 5 power modifiers.`), `pkg/engine/layers.go:31` (`func (a *Asset) TempPower() int {`), `pkg/engine/layers.go:48` (`return max(0, a.Printed().Power+a.TempPower())`), `pkg/engine/snapshot.go:26` (the `AssetSnap` field) and `pkg/engine/snapshot.go:64` (`TempPower: a.TempPower(), Power: ...`).
4. **One list, one append, one expiry:** `git -C W grep -n -e 'Modifiers' -- pkg cmd ':!*_test.go'` prints exactly five lines: `pkg/engine/effects.go:285` (the append), `pkg/engine/layers.go:33` (`for _, m := range a.Modifiers {`), `pkg/engine/layers.go:43` (the `slices.DeleteFunc`), and `pkg/engine/state.go:76` and `:78` (the field's comment and the field).
5. **Only the listed edits:** `git -C W diff autoloop/m2-wave-f --numstat -- pkg cmd ':!*_test.go'` prints exactly five lines: `3	3	pkg/engine/effects.go`, `38	2	pkg/engine/layers.go`, `1	1	pkg/engine/snapshot.go`, `3	2	pkg/engine/state.go` and `1	1	pkg/engine/turn.go`.
6. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
7. **Nothing else changed:** `git -C W diff autoloop/m2-wave-f --stat` lists only the five files of case 5, the ten test files of case 1, and this task file.
8. `gofmt -l W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/state.go`
- `pkg/engine/layers.go`
- `pkg/engine/turn.go`
- `pkg/engine/effects.go`
- `pkg/engine/snapshot.go`
- the ten test files under Tests, in the `test(T-0218)` commit only

## Done When
- [x] Cases 1–8 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-f` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-11, lead-architect (wave F run, session 20): wrote this task from board directive #137, whose trigger fired when T-0219 (*Allshan's Battle Trance*, the next Deliverables task in survey row F's order after *Veil Render*, *Reaver's Rush* and their shapes) was written `queued` in the same commit. #137 is `open: before T-0219`, so this task goes first. #140 stays `open: before` T-0219 for its part (2), which T-0219 applies. The design note's Layers sentence was rewritten in this commit. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because every edit is exact and was dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: #6, #7 (mode `serial`), #15 (no turn-boundary hook), #71, #91 (2), #106, #120, #136 (no *Eerie Remnant* task), #141 (3), #142: not fired. Next: `kernel-engineer` on Flash.
- 2026-11-11, kernel-engineer: implemented T-0218 (#137). Test updates committed first in 55eb19e (test(T-0218): read and write Modifiers and TempPower() (#137)) which failed compilation due to undefined Asset.Modifiers, Modifier, and calling int TempPower as a function. Implementation committed in 9014d28 (feat(engine): until end of turn is one list of modifiers on asset (T-0218)). Replaced Asset.TempPower int with Modifiers []Modifier on Asset in state.go; defined Duration, UntilEndOfTurn, Modifier, Asset.TempPower(), and Asset.expire() in layers.go; updated Game.Power() to call a.TempPower(); updated turn.go Sub-Phase 4c to call a.expire(UntilEndOfTurn); updated stepModifyStat in effects.go to append Modifier{Layer: 5, Power: n, Until: UntilEndOfTurn}; updated assetSnap in snapshot.go to call a.TempPower(). Reused slices.DeleteFunc for expiry. Verified acceptance cases 1-8: Case 1 diff numstat matched exactly ten lines and git log showed only test(T-0218); Case 2 sim-compare passed byte-for-byte; Case 3 grep TempPower returned exactly five lines; Case 4 grep Modifiers returned exactly five lines; Case 5 non-test diff numstat returned exactly five lines (3 3 effects.go, 38 2 layers.go, 1 1 snapshot.go, 3 2 state.go, 1 1 turn.go); Case 6 cardtown lint fences passed; Case 7 diff stat matched only the 5 implementation files, 10 test files, and task file; Case 8 gofmt clean, task check passed, and task sim completed 200 games clean (P0 71.5% / P1 27.5% / draws 1.0%).
- 2026-11-11, lead-architect (wave F run, session 20, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-f` as 2466e13. Flash run `d075a93c-be5b-4cc9-ba52-9f0bb46e559b`: exit 0, `status: SUCCESS`, no `denied_actions` key, no `permission check failed` in the transcript; `.err` holds only agy's "root agent idle" notice. Review of `git diff autoloop/m2-wave-f...m2/autoloop-wave-f/T-0218-until-end-of-turn-modifiers` (19c89ab):
  - Scope: the five Allowed Files, the ten named test files and this file. Every test edit is in 55eb19e (`test(T-0218)`), and no later commit touches a test file.
  - Pure move, checked against the diff: the `pkg/` diff is identical, line for line, to the lead's dry run, so it is exactly Edits 1-5 and the Tests table. No test name changed and no assertion changed.
  - Fences and placement: no card name; `Modifier`, `Duration` and the two `Asset` methods sit in `layers.go`, with the effective reads (#126). The only copy of "until end of turn" is now `expire(UntilEndOfTurn)` in 4c.
  - Re-run by the reviewer in a scratch worktree of the task branch: `gofmt -l pkg` clean; `cardtown lint fences` exit 0; `go tool task check` exit 0; `go tool task sim-compare -- autoloop/m2-wave-f` `[PASS]`; `go tool task sim` exit 0, clean, 200 games, P0 71.5% / P1 27.5% / draws 1.0%, P0 50/50 and P1 44/50 main cards executable.
  - The merge's `pkg/`, `cmd/` and `data/` match the task tip's (the run branch gained only this task's queue commit after the branch point), so the gates hold on the run branch. #137 is applied and archived.
