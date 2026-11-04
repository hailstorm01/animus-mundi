---
id: T-0161
title: The play screen names an activation, its asset and printed text (harness, flash)
kind: harness
tier: flash
wave: D
status: done
branch: m2/autoloop-wave-d/T-0161-activate-label
depends_on: [T-0158]
claims: [view:asset-text, label:activate]
cards: []
requirements: [FR-ACT-001]
---

## Goal
A human choosing from `cardtown play`'s action menu reads what an activation does, before the first card with an activated ability becomes executable.

## Inputs
- FR-ACT-001 in FRD 01; FR-ZON-005 (a face-down asset shows only the Scratch face).
- T-0159's split (D2.4): the label needs a text the `View` doesn't carry.
- Design note: `docs/architecture/m2-ability-model.md`, Boundaries. `pkg/view` holds what a seat sees and `cmd/cardtown` the play screen, so both changes sit where they belong.
- Why the whole rules text and not the one ability's line: `ability.Activated` holds no text, and every Animus or Path card with an activated ability has that ability as its only rules-text line (searched `data/cards` 2026-11-04 against the run branch at f51bc7d: *Sinuous Growth*, *Cliff of Portents*, *Boneyard Scavenger*, *Ruin Partisans*, *Rhapsody's Hollow*, and the M3 *Rift* and *Encampment*). So the label needs no ability index, and `a.Ability` stays unread. Outside M2's factions the first list of what this task covers was too broad: `rulesText` can span lines, and *Tempest Surge* (`data/cards/units-and-structures-concordia.csv:10-11`) has its activated ability on the second one. Such a card's menu entry would print its whole text over several lines (*NO-AH-5* has three), which still reads correctly, so nothing here handles line breaks. No card holds two activated abilities.

## Shape (lead decisions)
- `view.AssetView` (`pkg/view/view.go`) gains `Text string`, the asset's printed rules text, as `CardView.Text` has it. `assetView` reads it through `a.Printed().RulesText`, never `a.Inst.Card`. The Scratch face prints no text (`scratchFace`, `pkg/engine/card.go:133`), so a face-down asset's `Text` is empty and reveals nothing.
- `actionLabel` (`cmd/cardtown/play_screen.go`) returns `Activate: <Name>#<id>: <Text>` for `ActActivate`, and `Activate: Asset#<id>` for an asset the view doesn't hold, as the Enact label's fallback does.
- `findAsset` walks the view's assets today. Make one lookup that returns the `AssetView` or nil, and have `findAsset` and the Activate case both use it. A second walk is a `changes` note. Today's walk order and every other label stay the same.
- Dry run by the lead on the run branch: with `Text` added and filled, `go test ./pkg/view ./cmd/cardtown` passes, `TestView_Case4_NoLeak` included, so no public asset's text names a hidden card.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Both test packages stop compiling while `AssetView` has no `Text`, which is the failure the `test(T-0161)` commit records: add the tests first, confirm that, and commit them before touching `view.go`. Case 4 pins today's labels, so it passes as soon as the package compiles.
1. **`TestView_AssetTextIsPrinted`, in `pkg/view/view_test.go`:** a unit with `RulesText` `"exhaust: test text"` in player 0's Reserves, and a unit with `RulesText` `"T1: other text"` in the Main Theater under player 1. In `view.PlayerView` for both seats, each asset's `AssetView.Text` is its card's `RulesText`. `go -C W test ./pkg/view -run TestView_AssetTextIsPrinted` passes.
2. **`TestView_FaceDownAssetHidesItsText`, in `pkg/view/view_test.go`:** a face-down asset whose card has `RulesText` `"Alpha secret text"`, set up as in `TestView_Case2_FaceDown`. In `view.PlayerView` for both seats, its `Text` is `""`, and the view's `mustJSON` doesn't contain `Alpha secret text`. `go -C W test ./pkg/view -run TestView_FaceDownAssetHidesItsText` passes.
3. **`TestPlay_ActionLabelNamesTheActivation`, in `cmd/cardtown/play_test.go`:** built from a `view.View` literal as `TestPlay_ActionLabelNamesTheEnactedCard` is, holding `You.HQ` (ID 1, `Your HQ`, text `hq text`), `You.Reserves` (ID 2, `Your Unit`, text `your text`), `Opponent.HQ` (ID 3, `Their HQ`, text `their hq text`), `Opponent.Reserves` (ID 4, `Their Unit`, text `their text`) and `Theater` (ID 5, `Cliff of Portents`, text `R1, T1, exhaust: Search your deck for a card, put it into your hand, then shuffle your deck.`). `actionLabel` for `engine.Action{Kind: engine.ActActivate, Asset: id}` returns:
   - 1 → `Activate: Your HQ#1: hq text`
   - 2 → `Activate: Your Unit#2: your text`
   - 3 → `Activate: Their HQ#3: their hq text`
   - 4 → `Activate: Their Unit#4: their text`
   - 5 → `Activate: Cliff of Portents#5: R1, T1, exhaust: Search your deck for a card, put it into your hand, then shuffle your deck.`
   - 99 → `Activate: Asset#99`
4. **`TestPlay_ActionLabelFindsAssetsInEveryZone`, in `cmd/cardtown/play_test.go`:** on the same `view.View`, the labels that read the lookup through `findAsset` don't change: `ActMarch` with `Asset` 1 to 5 returns `March: Your HQ#1`, `March: Your Unit#2`, `March: Their HQ#3`, `March: Their Unit#4` and `March: Cliff of Portents#5`; `ActMarch` with `Asset` 99 returns `March: Asset#99`; `ActAttack` with `Asset` 2 and `Target` 5 returns `Attack: Your Unit#2 -> Cliff of Portents#5`; `ActScratch` with `Asset` 1 returns `Deploy Scratch Company: Your HQ#1`.
5. **The sims don't read it:** `go -C W tool task sim-compare -- autoloop/m2-wave-d` prints `[PASS]`.
6. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d --stat` lists only `pkg/view/view.go`, `pkg/view/view_test.go`, `cmd/cardtown/play_screen.go`, `cmd/cardtown/play_test.go` and this task file.

## Allowed Files
- `pkg/view/view.go`
- `cmd/cardtown/play_screen.go`
- `pkg/view/view_test.go` (the two new tests only; no existing test changes)
- `cmd/cardtown/play_test.go` (the two new tests only; no existing test changes)

## Done When
- [x] Cases 1-6 hold, with their commands and results in the Log
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-compare -- autoloop/m2-wave-d` passes

## Log
- 2026-11-04 lead-architect: split off T-0159 under the D2 size rule (D2.4), because the label needs a text the `View` doesn't carry, which made T-0159 five files. Written `queued`.
- 2026-11-04 lead-architect (wave D run, session 4): wrote Inputs, Shape, Acceptance Cases, Allowed Files and Done When against the run branch at f51bc7d; `main` has no commits to sync. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the shape is one field and one lookup. Every command here is on the headless allow-list, one per line. Narrowed the stub's claim about rules text to M2's factions (see Inputs). Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2): no such card; #104: not due, this task has no `cards:`). Next: `kernel-engineer` on Flash.
- 2026-11-04 kernel-engineer (T-0161):
  - Pre-implementation tests: added TestView_AssetTextIsPrinted and TestView_FaceDownAssetHidesItsText in pkg/view/view_test.go, and TestPlay_ActionLabelNamesTheActivation and TestPlay_ActionLabelFindsAssetsInEveryZone in cmd/cardtown/play_test.go. Confirmed compilation failed due to missing AssetView.Text. Committed test(T-0161) (17813c2).
  - Reuse: CardView.Text already held printed rules text; added Text string to view.AssetView populated from a.Printed().RulesText in assetView (face-down assets receive empty string via scratchFace). In cmd/cardtown/play_screen.go, refactored existing findAsset traversal into a shared lookupAsset helper returning *view.AssetView, used by both findAsset and engine.ActActivate.
  - Implementation committed as feat(engine): the play screen names an activation, its asset and printed text (T-0161) (09719c8).
  - Commands & Verification:
    - Case 1: `go test ./pkg/view -run TestView_AssetTextIsPrinted` -> PASS
    - Case 2: `go test ./pkg/view -run TestView_FaceDownAssetHidesItsText` -> PASS
    - Case 3: `go test ./cmd/cardtown -run TestPlay_ActionLabelNamesTheActivation` -> PASS
    - Case 4: `go test ./cmd/cardtown -run TestPlay_ActionLabelFindsAssetsInEveryZone` -> PASS
    - Case 5: `go tool task sim-compare -- autoloop/m2-wave-d` -> [PASS] starter and stress sims match autoloop/m2-wave-d byte for byte, exit codes included
    - Case 6: `git diff autoloop/m2-wave-d --stat` -> exactly 5 files changed (.agents/tasks/m2/T-0161-activate-label.md, cmd/cardtown/play_screen.go, cmd/cardtown/play_test.go, pkg/view/view.go, pkg/view/view_test.go)
    - Gate: `go tool task check` -> PASS (gofmt, go vet, all package tests, cardtown lint --all)
    - Starter sim: `go tool task sim` -> PASS (200 games, all invariants holding)
- 2026-11-04, lead-architect (wave D run, session 4, D5): **Changes, round 1.** Flash run `cc19d76c-4cad-4e40-8c93-4c6e18d4028e`: exit 0, `status: SUCCESS`, no `denied_actions` key, and the transcript has no step with `"status":"ERROR"` and no `permission check failed` match. Reviewer's re-runs in a scratch worktree of the task branch at e998373: cases 1-6 hold; `go tool task check`, `go tool task sim` and `go tool task sim-compare -- autoloop/m2-wave-d` pass; at 17813c2 both test packages fail to compile on the missing `Text`. The only test-file edit after the test commit is gofmt removing a trailing blank line from each file, and no test changed. The Log's `09719c8` is an earlier hash of the implementation commit, e998373 now.
  1. **Reuse (`cmd/cardtown/play_screen.go`, the `engine.ActActivate` case):** it writes `findAsset`'s `%s#%d` format and its `Asset#%d` fallback a second time, so a change to how the menu names an asset would need two edits. Replace the case's body with exactly this, and change nothing else:
     ```go
     	case engine.ActActivate:
     		if as := lookupAsset(a.Asset); as != nil {
     			return fmt.Sprintf("Activate: %s: %s", findAsset(a.Asset), as.Text)
     		}
     		return fmt.Sprintf("Activate: %s", findAsset(a.Asset))
     ```
     No test changes: `git -C W diff e998373 --stat` lists only `cmd/cardtown/play_screen.go` and this task file. Then run `go -C W tool task check`, `go -C W tool task sim` and `go -C W tool task sim-compare -- autoloop/m2-wave-d`, and record each result in the Log.
- 2026-11-04 kernel-engineer (T-0161, round 1 changes):
  - Addressed review note 1 (Reuse in cmd/cardtown/play_screen.go): replaced the body of engine.ActActivate in actionLabel with the exact requested implementation, formatting with `findAsset(a.Asset)` for both branches to eliminate duplicate asset naming logic.
  - Diff: `git diff e998373 --stat` lists only `cmd/cardtown/play_screen.go` and this task file. `git diff autoloop/m2-wave-d --stat` lists only the 5 allowed files.
  - Verification & Gates:
    - Case 1: `go test ./pkg/view -run TestView_AssetTextIsPrinted` -> PASS
    - Case 2: `go test ./pkg/view -run TestView_FaceDownAssetHidesItsText` -> PASS
    - Case 3: `go test ./cmd/cardtown -run TestPlay_ActionLabelNamesTheActivation` -> PASS
    - Case 4: `go test ./cmd/cardtown -run TestPlay_ActionLabelFindsAssetsInEveryZone` -> PASS
    - Case 5: `go tool task sim-compare -- autoloop/m2-wave-d` -> [PASS] starter and stress sims match autoloop/m2-wave-d byte for byte, exit codes included
    - Case 6: `git diff autoloop/m2-wave-d --stat` -> exactly 5 files changed
    - Gate: `go tool task check` -> PASS (gofmt, go vet, all package tests, cardtown lint --all)
    - Starter sim: `go tool task sim` -> PASS (200 games, all invariants holding)
- 2026-11-04, lead-architect (wave D run, session 4, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-d` as 6de8594. Flash run `a9e1931b-ac4f-438d-852a-874f95223109`: exit 0, `status: SUCCESS`, no `denied_actions` key, no step with `"status":"ERROR"`; the transcript's two `permission check failed` matches are this task's own review note, read and then diffed by the engineer, and the `.err` file holds only agy's "root agent idle" notice. Review of `git diff autoloop/m2-wave-d...m2/autoloop-wave-d/T-0161-activate-label` (282d788):
  - Note 1: the `ActActivate` case is the note's text, and `findAsset` is now the one place in `actionLabel` that names an asset.
  - Scope: only the four Allowed Files and this file changed. Round 2 changed no test. The new tests are the four named, and no existing test changed.
  - Reuse and placement: one lookup, no second walk; `AssetView.Text` reads `Printed()`, and a face-down asset's text is empty. `formatAsset` also prints `%s#%d`, inside a different line format and from before this task: one format token, not a rule or a scheme, so no directive (D2.0).
  - Fences: no card name in non-test code; `go run ./cmd/cardtown lint fences` exits 0.
  - Re-run by the reviewer in a scratch worktree of the task branch, all as expected: cases 1-6; `go tool task check` exit 0; `go tool task sim` exit 0, 200 games, P0 82.5% / P1 17.5% / draws 0.0%, 42/50 and 32/50 main cards executable, no monitor violation; `go tool task sim-compare -- autoloop/m2-wave-d` `[PASS]`.
  - `go tool task check` and `go tool task sim` also pass on the merged run branch at 6de8594.
