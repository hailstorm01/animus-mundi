---
id: T-0104
title: Scratch Company in a filter is one three-way enum, ability.ScratchCompany (directive #78, harness, flash, no behavior change)
kind: harness
tier: flash
wave: G.1 (board directive #78, structural review 5, finding 3)
status: done
branch: m2/T-0104-scratch-company-enum
depends_on: [T-0103]
claims: [type:ScratchCompany, field:Filter.Scratch, func:matchesPrinted]
cards: []
requirements: []
---

## Goal
A filter's Scratch Company criterion is one field with three values: any unit, Scratch Company only, or no Scratch Company. `Filter.Scratch` becomes an `ability.ScratchCompany` enum (`ScratchAny`, the zero value; `ScratchOnly`; `ScratchExcluded`), `Filter.NonScratch` goes, and `matchesPrinted` reads the criterion in one switch. Behavior is unchanged: same tests (with the seven named edits below), same sim output.

## Inputs
- Board directive #78 (structural review 5, approved by the user, 2026-10-29).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6d (wave G.1).
- `ability`'s package comment (directive #70): an enum whose zero value is a meaningful default starts at `iota`, and its first member's comment says "the zero value".
- Rule the code implements, unchanged: `FR-TGT-001` (a filter's criteria are a conjunction; subtype criteria).
- Why behavior is unchanged: today `matchesPrinted` ends `return !(f.NonScratch && scratch) && !(f.Scratch && !scratch)`. With neither flag set that is `true` (now `ScratchAny`); with `Scratch` it is `scratch` (now `ScratchOnly`); with `NonScratch` it is `!scratch` (now `ScratchExcluded`). No definition or test sets both flags (checked by grep at fc6faa2), so the one combination the enum cannot express never occurs.

### Where it is today (at `autoloop/m2-wave-g.1` fc6faa2)
| Where | Today |
| :--- | :--- |
| `pkg/engine/ability/ability.go:11-13` | the package comment's "zero value" rule, listing `TurnOwner, Keyword, Zone, Allegiance, Subject, ValueSource` |
| `pkg/engine/ability/ability.go:135-136` | `Filter.NonScratch` and `Filter.Scratch`, two `bool`s |
| `pkg/engine/ability/ability.go:154-160` | `Kinds` and its constants; the new type goes right after them |
| `pkg/engine/targeting.go:183, 204, 224` | three doc comments that say "Scratch and NonScratch" |
| `pkg/engine/targeting.go:219-220` | the end of `matchesPrinted` |
| `pkg/engine/cards/animus.go:202` | *Shifting Impossibility*: `Scratch:    true,` |
| `pkg/engine/cards/path.go:289` | *Rogue Fleshshaper* mode 2: `NonScratch: true,` |
| `pkg/engine/cards/path.go:318` | *Saboteur* slot 0: `NonScratch: true,` |

## Edits (lead decisions; copy exactly)

### 1. `pkg/engine/ability/ability.go`: the package comment
Replace these two lines:
```go
//   - An enum whose zero value is a meaningful default starts at iota, and the
//     first member's comment says "the zero value" (TurnOwner, Keyword, Zone,
//     Allegiance, Subject, ValueSource).
```
with:
```go
//   - An enum whose zero value is a meaningful default starts at iota, and the
//     first member's comment says "the zero value" (TurnOwner, Keyword, Zone,
//     Allegiance, ScratchCompany, Subject, ValueSource).
```
(That is three lines; only the last one changes.)

### 2. `pkg/engine/ability/ability.go`: `Filter`
Replace these two field lines:
```go
	NonScratch bool       // excludes Scratch Company units
	Scratch    bool       // only Scratch Company units
```
with this one:
```go
	Scratch    ScratchCompany // ScratchAny (zero) means any unit; else only Scratch Company units, or none
```
Then run `gofmt` on the file: it realigns the other fields' types and comments in the struct. No other field changes.

### 3. `pkg/engine/ability/ability.go`: the `ScratchCompany` type
Right after the `Kinds` constant block's closing `)` and its blank line, before `// Op is one primitive step.`, add:
```go
// ScratchCompany is a filter's Scratch Company criterion: whether a unit with
// the Scratch Company subtype matches (FR-TGT-001).
type ScratchCompany int

const (
	ScratchAny      ScratchCompany = iota // the zero value: any unit, Scratch Company or not
	ScratchOnly                           // only Scratch Company units ("allied Scratch Company unit")
	ScratchExcluded                       // no Scratch Company units ("non-Scratch unit")
)

```

### 4. `pkg/engine/targeting.go`: `matchesPrinted`
Replace its last two lines:
```go
	scratch := c.HasSubtype("Scratch Company")
	return !(f.NonScratch && scratch) && !(f.Scratch && !scratch)
```
with:
```go
	switch f.Scratch {
	case ability.ScratchOnly:
		return c.HasSubtype("Scratch Company")
	case ability.ScratchExcluded:
		return !c.HasSubtype("Scratch Company")
	}
	return true
```

### 5. `pkg/engine/targeting.go`: three doc comments
- Line 183, in `matchesCard`'s comment: `Kinds, Scratch and NonScratch by its` becomes `Kinds and Scratch by its`.
- Line 204, in `matchesPrinted`'s comment: `then Kinds, Scratch and NonScratch by the card's printed characteristics.` becomes `then Kinds and Scratch by the card's printed characteristics.`
- Line 224, in `matchesFilter`'s comment: `kinds, Scratch and NonScratch, Other:` becomes `kinds, Scratch, Other:`.

Nothing else in those comments changes.

### 6. The three card definitions
- `pkg/engine/cards/animus.go:202`: `Scratch:    true,` becomes `Scratch:    ability.ScratchOnly,`.
- `pkg/engine/cards/path.go:289`: `NonScratch: true,` becomes `Scratch: ability.ScratchExcluded,`.
- `pkg/engine/cards/path.go:318`: `NonScratch: true,` becomes `Scratch:    ability.ScratchExcluded,`.

Run `gofmt` on both files afterwards. No `Text` line changes.

### 7. Test edits (D2.4: named here, and they land in the `test(T-0104)` commit)
These seven edits are the only test changes. They are the whole `test(T-0104)` commit, made before Edits 1–6. After that commit, `go -C W vet ./pkg/engine` fails only because `ability.ScratchOnly`, `ability.ScratchExcluded` and the `ScratchCompany` type of `Filter.Scratch` do not exist yet. That is the expected failure; record it in the Log.
1. `pkg/engine/cards_animus_impossibility_test.go:19` (`alliedScratch`): `Scratch: true}` becomes `Scratch: ability.ScratchOnly}`.
2. `pkg/engine/source_power_test.go:478` (in `TestFR_TGT_002_FilteredAllNonScratchSkipsScratchCompanies`): `NonScratch: true}` becomes `Scratch: ability.ScratchExcluded}`. The test's name stays.
3. `pkg/engine/targeting_test.go:20-22`: replace the helper `tgt` with:
   ```go
   func tgt(a ability.Allegiance, k ability.Kinds, nonScratch bool) ability.Target {
   	f := ability.Filter{Allegiance: a, Kinds: k}
   	if nonScratch {
   		f.Scratch = ability.ScratchExcluded
   	}
   	return ability.Target{Filter: f}
   }
   ```
   Its signature stays, so no caller changes.
4. `pkg/engine/targeting_test.go:514`: `Scratch: true}` becomes `Scratch: ability.ScratchOnly}`.
5. `pkg/engine/targeting_test.go:537` (in `TestFR_TGT_002_FilteredAllScratchHitsOnlyScratchCompanies`): `Scratch: true}` becomes `Scratch: ability.ScratchOnly}`.
6. `pkg/engine/trigger_order_test.go:139`: `NonScratch: true,` becomes `Scratch: ability.ScratchExcluded,`.
7. `pkg/engine/zone_filter_test.go:18-20`: replace the helper `tgtZone` with:
   ```go
   func tgtZone(a ability.Allegiance, k ability.Kinds, nonScratch bool, z ability.Zone) ability.Target {
   	f := ability.Filter{Allegiance: a, Kinds: k, Zone: z}
   	if nonScratch {
   		f.Scratch = ability.ScratchExcluded
   	}
   	return ability.Target{Filter: f}
   }
   ```
   Its signature stays, so no caller changes.

Two test names keep "NonScratch" inside a longer name (`TestFR_TGT_003_SaboteurWithNoOpposingNonScratchUnitStillDeploysAndAlliedClauseResolves` in `cards_path_test.go`, and `TestFR_TGT_002_FilteredAllNonScratchSkipsScratchCompanies`). They name the rule ("non-Scratch unit"), not the field, so they are not renamed. Case 4 below checks the field, as a whole word.

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: behavior is unchanged, and the existing tests (with Edit 7) pin it. Run each command on its own, pointed at your worktree (W). Record each command and its result in the Log. The reviewer re-runs them.
1. **Tests changed only as named** (#78 case (a)): `git -C W diff autoloop/m2-wave-g.1 --stat -- '*_test.go'` lists exactly `cards_animus_impossibility_test.go`, `source_power_test.go`, `targeting_test.go`, `trigger_order_test.go` and `zone_filter_test.go`, all under `pkg/engine/`. `git -C W log --oneline autoloop/m2-wave-g.1..HEAD` names the `test(T-0104)` commit, and `git -C W show --stat <that commit's sha>` lists the same five files and nothing else. `go -C W test ./...` passes.
2. **Sims byte-identical** (#78 case (b)): `go -C W tool task sim-compare -- autoloop/m2-wave-g.1` prints `[PASS]`.
3. **Card checks unchanged** (#78 case (c)): `go -C W test ./pkg/engine/... -run "TestDefs_MatchRulesText|TestStarterSet_Executable"` passes, and neither test's file is in case 1's list.
4. **No `NonScratch` field** (#78 case (d)): `git -C W grep -nw NonScratch -- pkg` prints nothing (exit status 1). The two test names that contain "NonScratch" inside a longer word do not match `-w`.
5. **One switch:** `git -C W grep -n "f.Scratch" -- pkg/engine/targeting.go` prints exactly one line, the `switch f.Scratch {` in `matchesPrinted`.
6. **The enum is used:** `git -C W grep -n -e ability.ScratchOnly -e ability.ScratchExcluded -- pkg/engine/cards` prints exactly three lines: `animus.go` (`ScratchOnly`) and two in `path.go` (`ScratchExcluded`).
7. **Only these edits:** `git -C W diff autoloop/m2-wave-g.1 --stat` lists only `pkg/engine/ability/ability.go`, `pkg/engine/targeting.go`, `pkg/engine/cards/animus.go`, `pkg/engine/cards/path.go`, the five test files of case 1 and this task file.
8. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/targeting.go`
- `pkg/engine/cards/animus.go`
- `pkg/engine/cards/path.go`
- Test files, only as Edit 7 names: `pkg/engine/cards_animus_impossibility_test.go`, `pkg/engine/source_power_test.go`, `pkg/engine/targeting_test.go`, `pkg/engine/trigger_order_test.go`, `pkg/engine/zone_filter_test.go`

## Done When
- [x] Cases 1–8 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-g.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-29, lead-architect (wave G.1 run, session 3): wrote the task from board directive #78, the next of #76 to #79 in board order now that #77 is applied (T-0103). Harness, so it skips `tests` and goes straight to `ready`; tier `flash` as #78 gives; every command here is on the headless allow-list. The size rule counts four non-test files (`ability.go`, `targeting.go`, `animus.go`, `path.go`), as #78 says. #78 names six test files' sites; at fc6faa2 they are seven sites in five files (`targeting_test.go` has three), and Edit 7 names each. The two helpers keep their `bool` parameter so none of their callers changes. #78 case (d) ("no file under `pkg/` mentions `NonScratch`") is read as the field: two protected test names contain the word "NonScratch" for the rule's "non-Scratch unit", and renaming them is not what the directive is after, so case 4 matches the whole word. Reuse: `ability` has no Scratch Company type yet (searched `ScratchCompany`, `ScratchAny`, `ScratchOnly`, `ScratchExcluded`), and `matchesPrinted` is the only reader of either flag. `main` has no commits the run branch lacks, so no sync. Waiting directives checked: none has triggered (#8: no `deck.go`; #15, #20 (b), #58, #59 (2), #61, #71: no such hook, shape, claim or card here; #6, #7: mode is `serial`). #79 stays `open: before the next Deliverables task`; #81 stays `open: after #79`. Next: `kernel-engineer` on Flash.
- 2026-10-29, lead-architect (wave G.1 run, session 4): session 3 ran out of usage before committing this file, and left it uncommitted in the run branch's worktree, which this session moved into its own scratch folder. Re-checked every Edit and site against fc6faa2: they hold. Rewrote three Acceptance Case commands: case 1 finds the `test(T-0104)` commit with `git log` instead of `HEAD~1`, which is a status commit by review time; case 5 greps `targeting.go` only, since the test helpers also write `f.Scratch`; case 6 uses two `-e` patterns instead of `\|`. `main` (556811f) still has nothing the run branch lacks; the board is unchanged since session 3's read, so its directive and trigger checks stand.
- 2026-10-29, kernel-engineer (session 1, flash): implemented T-0104 per directive #78 and task specification.
  - Step 1: Loaded task in scratch worktree and set status: in-progress (commit 9a9f1eb).
  - Step 4 (Harness / Edit 7): Applied test edits 1-7 in `pkg/engine/cards_animus_impossibility_test.go`, `pkg/engine/source_power_test.go`, `pkg/engine/targeting_test.go`, `pkg/engine/trigger_order_test.go`, and `pkg/engine/zone_filter_test.go`. Verified pre-implementation failure with `go -C W vet ./pkg/engine` (failed only on `undefined: ability.ScratchOnly`). Committed test changes as `test(T-0104): update tests for ability.ScratchCompany enum` (commit 3ffa57e).
  - Step 5 (Implementation): Applied Edits 1-3 to `pkg/engine/ability/ability.go` (package comment, `Filter.Scratch` as `ScratchCompany` enum removing `NonScratch`, defined `ScratchCompany` with `ScratchAny`, `ScratchOnly`, `ScratchExcluded`), Edits 4-5 to `pkg/engine/targeting.go` (`matchesPrinted` switch on `f.Scratch`, updated 3 doc comments), and Edit 6 to `pkg/engine/cards/animus.go` (Shifting Impossibility `ScratchOnly`) and `pkg/engine/cards/path.go` (Rogue Fleshshaper, Saboteur `ScratchExcluded`). Realigned formatting with `gofmt`. Committed as `feat(engine): Scratch Company in Filter is ability.ScratchCompany enum (T-0104)` (commit 0a5549c).
  - Acceptance Cases:
    1. Tests changed only as named: `git -C W diff autoloop/m2-wave-g.1 --stat -- '*_test.go'` listed exactly the 5 files (15 insertions, 7 deletions). `git -C W log --oneline autoloop/m2-wave-g.1..HEAD` listed `test(T-0104)` commit 3ffa57e. `git -C W show --stat 3ffa57e` listed the same 5 files. `go -C W test ./...` passed (32s).
    2. Sims byte-identical: `go -C W tool task sim-compare -- autoloop/m2-wave-g.1` printed `[PASS] starter and stress sims match autoloop/m2-wave-g.1 byte for byte, exit codes included`.
    3. Card checks unchanged: `go -C W test ./pkg/engine/... -run "TestDefs_MatchRulesText|TestStarterSet_Executable"` passed (`ok animus-mundi/pkg/engine 0.767s`).
    4. No NonScratch field: `git -C W grep -nw NonScratch -- pkg` printed nothing (exit status 1).
    5. One switch: `git -C W grep -n "f.Scratch" -- pkg/engine/targeting.go` printed `pkg/engine/targeting.go:219: switch f.Scratch {`.
    6. Enum used: `git -C W grep -n -e ability.ScratchOnly -e ability.ScratchExcluded -- pkg/engine/cards` printed 3 lines (animus.go:202, path.go:289, path.go:318).
    7. Only these edits: `git -C W diff autoloop/m2-wave-g.1 --stat` listed only `ability.go`, `targeting.go`, `animus.go`, `path.go`, the 5 test files, and the task file.
    8. `go -C W tool task check` (gofmt, vet, test, all linters) and `go -C W tool task sim` (starter simulation 200 games) passed cleanly.
  - Reviewer notes: Diff strictly follows Edits 1-7 in the task specification; no behavior change occurred. Handing off to lead-architect for review.
- 2026-10-29, lead-architect (wave G.1 run, session 4, review): approved, round 1, no changes. Checked against `git diff autoloop/m2-wave-g.1...m2/T-0104-scratch-company-enum`, not the Log: Edits 1–6 are in `feat` (0a5549c, the four allowed non-test files), Edit 7's seven sites are the whole `test(T-0104)` commit (3ffa57e, the five named test files), and no other line changed. `matchesPrinted`'s switch gives the old result for each of the three values, and no definition or test set both old flags. Re-ran in a review worktree: cases 4–6 as written, `go tool task check` (pass), `go tool task sim-compare -- autoloop/m2-wave-g.1` (`[PASS]`, starter and stress byte-identical) and `go tool task sim` (200 games, P0 67.0% / P1 32.5% / draws 0.5%, no monitor violation). Flash run `2050fee2`, 404 s, no refusals. Merged into `autoloop/m2-wave-g.1` as 24b7f71. Directive #78 applied.
