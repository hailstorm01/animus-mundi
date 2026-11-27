---
id: T-0323
title: "One `wantOptions` check for card tests, beside `refs` and `ids` (directive #190; wave A.1 harness, flash)"
kind: harness
tier: flash
wave: A.1 (board directive #190, design note Deliverable 5a)
status: review
branch: m3/autoloop-wave-a.1/T-0323-one-want-options
depends_on: []
claims: [func:engine.wantOptions (test helper)]
cards: []
requirements: []
---

## Goal
"The decision offers exactly these assets, in any order" is one test helper, `wantOptions` in `pkg/engine/targeting_test.go`, beside `refs` and `ids`. Wave A's four per-card copies (`crWantOffered`, `cbWantOffered`, `ecnSameIDs`, `ffeWantSet`) are deleted, and their 8 call sites in 5 files call `wantOptions`. Test behavior is unchanged, and no non-test file changes.

## Inputs
- Board directive #190 (structural review 16, finding 1), approved by the user 2026-11-26, `open: before the next Deliverables task`. Design note `docs/architecture/m3-full-catalog.md` Deliverable 5a.
- **A fifth copy the directive missed (lead decision, 2026-11-26).** `wantOptions` already exists, with the directive's exact signature and message, in [cards_path_test.go:31-36](../../../pkg/engine/cards_path_test.go) (T-0018, commit 9fa20500), and 57 call sites in 19 files use it, *Culverin Brigade*'s own `TestCard_CulverinBrigade_OnlyTheHeadquartersStillAsks` among them. Adding a second `wantOptions` in `targeting_test.go` would not compile, and a second name would keep the duplicate. So this task **moves that declaration** into `targeting_test.go`, below `ids`, with its body unchanged, and gives it a doc comment. Its body is `refs(d), ids(want...)` rather than the directive's `sortedCopy(refs(d)), sortedCopy(ids(want...))`. Both `refs` and `ids` already return sorted slices ([targeting_test.go:48-64](../../../pkg/engine/targeting_test.go)), so `sortedCopy` would be a no-op, and keeping the moved body as it was keeps the move pure. Every other caller compiles unchanged.
- The four copies (at `autoloop/m3-wave-a.1` 2ea08269), all the same check, `refs`/`ids` compared order-free, message `"%s: options got %v, want %v"`:
  - `crWantOffered`, [cards_soulless_conscript_test.go:27-36](../../../pkg/engine/cards_soulless_conscript_test.go); callers at :123, :146, :163, :219, and [cards_soulless_hellfire_test.go:68](../../../pkg/engine/cards_soulless_hellfire_test.go);
  - `cbWantOffered`, [cards_soulless_culverin_test.go:27-36](../../../pkg/engine/cards_soulless_culverin_test.go), byte for byte the same; caller at :125;
  - `ecnSameIDs`, [cards_concordia_nightmare_test.go:26-31](../../../pkg/engine/cards_concordia_nightmare_test.go); caller at :164;
  - `ffeWantSet`, [cards_soulless_fire_for_effect_test.go:61-83](../../../pkg/engine/cards_soulless_fire_for_effect_test.go), the same check with a hand-rolled loop (it also stops at the first mismatch and reports once, as `reflect.DeepEqual` does); caller at :131.
- Inline copies in earlier tests (`activated_test.go:320`, `:424`, `:451`) are left alone (directive #190).
- Rulings that apply: none.
- **Expected output difference:** none. No non-test file changes, so `go tool task sim-compare -- autoloop/m3-wave-a.1` passes.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 2ea08269: `gofmt -l pkg cmd` clean, `go vet ./pkg/engine/` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m3-wave-a.1` `[PASS]`. Resulting blobs (`git hash-object`): `targeting_test.go` 3612b2a, `cards_path_test.go` 4db601c, `cards_soulless_conscript_test.go` 8de88a4, `cards_soulless_culverin_test.go` 064c28a, `cards_concordia_nightmare_test.go` 22fa96d, `cards_soulless_fire_for_effect_test.go` a02d435, `cards_soulless_hellfire_test.go` 4287f0b.

## Tests (authorized edits to existing tests, all in the one `test(T-0323)` commit)
Only the helper call changes in each of these tests; its arguments and message stay as they are:
- `TestCard_ConscriptRifles_DealsOneToTheChosenUnit`, `TestCard_ConscriptRifles_OffersEveryNonScratchUnit`, `TestCard_ConscriptRifles_UnitTurnedFaceDownNotOffered`, `TestCard_ConscriptRifles_WardedOpposingUnitNotOffered` (`cards_soulless_conscript_test.go`): `crWantOffered(` → `wantOptions(`.
- `TestCard_HellfireArmorers_HasWard` (`cards_soulless_hellfire_test.go`): `crWantOffered(` → `wantOptions(`.
- `TestCard_CulverinBrigade_OffersEveryOpposingAsset` (`cards_soulless_culverin_test.go`): `cbWantOffered(` → `wantOptions(`.
- `TestCard_EchoingNightmare_OnlyOpposingMainTheaterUnitsOffered` (`cards_concordia_nightmare_test.go`): `ecnSameIDs(` → `wantOptions(`.
- `TestCard_FireForEffect_OffersEveryMainTheaterAsset` (`cards_soulless_fire_for_effect_test.go`): `ffeWantSet(` → `wantOptions(`.

No other test changes. No test is added or removed.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
1. **`pkg/engine/targeting_test.go`:** right after the closing `}` of `func ids(as ...*Asset) []int` and its blank line, before `// totalDamage is the damage…`, insert:
   ```go
   // wantOptions checks that d offers exactly the given assets, in any order
   // (refs and ids both sort).
   func wantOptions(t *testing.T, d Decision, what string, want ...*Asset) {
   	t.Helper()
   	if got, w := refs(d), ids(want...); !reflect.DeepEqual(got, w) {
   		t.Errorf("%s: options got %v, want %v", what, got, w)
   	}
   }

   ```
   The file already imports `reflect`.
2. **`pkg/engine/cards_path_test.go`:** delete `func wantOptions` (lines 31-36) and the blank line after it, and drop `"reflect"` from the imports (nothing else in the file uses it).
3. **`pkg/engine/cards_soulless_conscript_test.go`:** delete `crWantOffered` with its comment line and the blank line after it; rename its four calls as under Tests; drop `"reflect"` and `"sort"` from the imports.
4. **`pkg/engine/cards_soulless_culverin_test.go`:** delete `cbWantOffered` with its comment line and the blank line after it; rename its call; drop `"reflect"` and `"sort"` from the imports.
5. **`pkg/engine/cards_concordia_nightmare_test.go`:** delete `ecnSameIDs` and the blank line after it; rename its call; drop `"reflect"` from the imports.
6. **`pkg/engine/cards_soulless_fire_for_effect_test.go`:** delete `ffeWantSet` with its comment line and the blank line after it; rename its call; drop `"sort"` from the imports.
7. **`pkg/engine/cards_soulless_hellfire_test.go`:** rename its one call.

## Commits
1. `test(T-0323): one wantOptions check beside refs and ids, for directive 190` with Edits 1-7 (no double quotes and no `#` in the message). A pure move adds no test and changes no non-test file, so there is no implementation commit.

## Acceptance Cases (harness tasks only; written by the lead)
These are checks recorded in the Log, and the reviewer re-runs them. No new tests.
1. `go test -list . ./pkg/engine` gives the same test names before and after (record the count), and `go test ./pkg/engine` passes.
2. `git grep -n -E "func (crWantOffered|cbWantOffered|ecnSameIDs|ffeWantSet|wantOptions)\b" -- pkg` finds exactly one line: `pkg/engine/targeting_test.go:` … `func wantOptions(`.
3. `git grep -n -E "crWantOffered|cbWantOffered|ecnSameIDs|ffeWantSet" -- pkg` finds nothing.
4. `git diff --stat autoloop/m3-wave-a.1 HEAD` lists only the seven `_test.go` files above.
5. `go tool task sim-compare -- autoloop/m3-wave-a.1` passes.
6. `go tool task check` passes.

## Allowed Files
- Tests only, in the `test(T-0323)` commit: `pkg/engine/targeting_test.go`, `pkg/engine/cards_path_test.go`, `pkg/engine/cards_soulless_conscript_test.go`, `pkg/engine/cards_soulless_culverin_test.go`, `pkg/engine/cards_concordia_nightmare_test.go`, `pkg/engine/cards_soulless_fire_for_effect_test.go`, `pkg/engine/cards_soulless_hellfire_test.go`.
- No non-test file.

## Done When
- [x] Edits 1-7 are committed as `test(T-0323): …`
- [x] Cases 1-5 hold, with their commands and results in the Log
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim-compare -- autoloop/m3-wave-a.1` passes (starter and stress sims byte-identical to the run branch)

## Log
- 2026-11-26 lead-architect: Wrote the task from board directive #190, the first of wave A.1's four directives (design note Deliverable 5a). Harness, tier flash: a pure move of one test helper. The directive's `wantOptions` already existed in `cards_path_test.go` (Inputs), so the existing declaration moves instead of a new one being added; the directive's acceptance (the same test names pass, the four helpers gone, `check`, `sim-compare`) is unchanged. Status `ready` on `m3/autoloop-wave-a.1/T-0323-one-want-options`, from `autoloop/m3-wave-a.1`. #191, #192 and #193 follow, one task each.
- 2026-11-26 kernel-engineer: Implemented directive #190 pure move of wantOptions test helper across the 7 allowed test files:
  - Moved wantOptions declaration from pkg/engine/cards_path_test.go into pkg/engine/targeting_test.go (below ids) with doc comment, and removed unused "reflect" import from cards_path_test.go.
  - Deleted duplicate helpers crWantOffered (cards_soulless_conscript_test.go), cbWantOffered (cards_soulless_culverin_test.go), ecnSameIDs (cards_concordia_nightmare_test.go), and ffeWantSet (cards_soulless_fire_for_effect_test.go) along with unused imports ("reflect", "sort").
  - Updated 8 call sites across 5 files to call wantOptions.
  - Blob hashes match lead dry run exactly: targeting_test.go (3612b2a), cards_path_test.go (4db601c), cards_soulless_conscript_test.go (8de88a4), cards_soulless_culverin_test.go (064c28a), cards_concordia_nightmare_test.go (22fa96d), cards_soulless_fire_for_effect_test.go (a02d435), cards_soulless_hellfire_test.go (4287f0b).
  - Committed in 69bc5a43 as `test(T-0323): one wantOptions check beside refs and ids, for directive 190`.
  - Acceptance results:
    1. `go test -list . ./pkg/engine`: 1725 tests before and after. `go test ./pkg/engine`: ok (0.832s).
    2. `git grep -n -E "func (crWantOffered|cbWantOffered|ecnSameIDs|ffeWantSet|wantOptions)\b" -- pkg`: exactly one line in pkg/engine/targeting_test.go:68.
    3. `git grep -n -E "crWantOffered|cbWantOffered|ecnSameIDs|ffeWantSet" -- pkg`: 0 matches (exit 1).
    4. `git diff --stat autoloop/m3-wave-a.1 HEAD`: lists only the seven _test.go files and task file.
    5. `go tool task sim-compare -- autoloop/m3-wave-a.1`: [PASS] starter and stress sims match autoloop/m3-wave-a.1 byte for byte, exit codes included.
    6. `go tool task check`: [PASS] (gofmt, go vet, all tests, cardtown lint --all).
  - Status -> review.
