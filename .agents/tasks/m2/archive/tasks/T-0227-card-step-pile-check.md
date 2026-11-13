---
id: T-0227
title: "A card step's \"still in a discard pile\" check is made once, in stepCards (directive #144, wave F.1 harness, flash, pure move)"
kind: harness
tier: flash
wave: F.1
status: done
branch: m2/autoloop-wave-f.1/T-0227-card-step-pile-check
depends_on: []
claims: [func:stepCards (pile check), func:stepDeployFree (pile check), func:stepRelocate (pile check)]
cards: []
requirements: []
---

## Goal
`stepCards` returns only the cards whose exact instance is still in a discard pile, each once, for every subject (#144). `stepDeployFree` and `stepRelocate` no longer check the pile themselves, as asset steps get "still on the battlefield" once from `slotAssets`. So the next card step doesn't copy the check. A pure move: same behavior, same test names, same sim output.

## Inputs
- Board directive #144 (structural review 11, finding 1, approved by the user, 2026-11-13), `open: before the next Deliverables task`; design note Deliverable 6j (wave F.1), first.
- The three copies at eed04ab: `stepDeployFree` (`pkg/engine/effects.go:130`, `g.DiscardCard(inst.ID) != inst`), `stepRelocate` (`:501`) and `stepCards`' `SubjectSelf` case (`:521`, `slices.Contains(g.Players[it.Inst.Owner].Discard, it.Inst)`). A card goes only to its owner's pile (`pileTarget`'s comment, `targeting.go:199-202`), so `DiscardCard(inst.ID) == inst` is the same test as the `SubjectSelf` one.
- Why it is a pure move: `stepCards` has exactly two callers, `stepDeployFree` and `stepRelocate`, and each skipped every card that failed the check, so filtering before the loop drops the same cards. Inside either loop, handling one card moves only that card (`takeCard`, then `deployCard` or `putCard`; triggers are only buffered), so a later card's check gives the same answer before the loop as in it. The one exception is the same instance listed twice, which only a `PerDiscard` pile slot could produce (no card has one): the old loops then skipped the second copy, because the first had left the pile. `stepCards` therefore keeps each card once, which gives both loops the same result for every input. No test calls `stepCards`, `stepDeployFree` or `stepRelocate` (`git grep -n -e stepCards -e stepDeployFree -e stepRelocate -- '*_test.go'` prints nothing at eed04ab).
- Reuse search (D2.3): `DiscardCard` (`zones.go:163`) is the existing lookup and the new check uses it; this task removes two copies and adds none.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at eed04ab: `gofmt -l pkg` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-f.1` printed `[PASS]`, `cardtown lint fences` exit 0, and `go tool task sim` exit 0 (200 games, P0 63.5% / P1 35.0% / draws 1.5%, P0 50/50 and P1 50/50 main cards executable, every invariant holding). Every expected output below is from that dry run.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at eed04ab, all in `pkg/engine/effects.go`. There is no test commit: no test changes, and the existing tests are the pin. Commit the one file as one implementation commit.

### 1. `stepDeployFree`'s comment
Replace lines 117-118:
```go
// stepDeployFree puts each card of stepCards, still in its owner's discard
// pile, onto the battlefield for free under the ability's controller and
```
with:
```go
// stepDeployFree puts each card of stepCards, each still in its owner's
// discard pile, onto the battlefield for free under the ability's controller and
```

### 2. `stepDeployFree`'s check
Line 130: change `if g.DiscardCard(inst.ID) != inst || len(zones) == 0 {` to `if len(zones) == 0 {`, and nothing else on the line.

### 3. `stepRelocate`'s comment
Replace lines 484-486:
```go
// Card, as a face-down asset's card is private in a hand or deck. A card must
// still be in a discard pile, which is public: its event names it there, before
// the move can hide it.
```
with:
```go
// Card, as a face-down asset's card is private in a hand or deck. A card of
// stepCards is still in a discard pile, which is public: its event names it
// there, before the move can hide it.
```

### 4. `stepRelocate`'s check
Delete lines 501-503:
```go
		if g.DiscardCard(inst.ID) != inst {
			continue
		}
```
so that the `for _, inst := range g.stepCards(r, st) {` line is followed directly by the `g.logEvent(` line.

### 5. `stepCards`
Replace lines 512-534 (its comment through its closing `}`):
```go
// stepCards returns the cards a card step acts on. SubjectSelf is a death
// trigger's own card: a Tracked item whose exact instance is still in its
// owner's discard pile (FR-ZON-004). SubjectPick with a pile filter is the
// cards picked from discard piles, in pick order. SubjectTarget is the cards
// chosen for its pile slot that are still legal. Anything else has none.
func (g *Game) stepCards(r *resolution, st ability.Step) []*Instance {
	switch st.Subject {
	case ability.SubjectSelf:
		it := r.it
		if it.Tracked && it.Inst != nil && slices.Contains(g.Players[it.Inst.Owner].Discard, it.Inst) {
			return []*Instance{it.Inst}
		}
	case ability.SubjectTarget:
		if st.Target >= 0 && st.Target < len(r.cards) {
			return r.cards[st.Target]
		}
	case ability.SubjectPick:
		if st.Filter.InPile() {
			return g.pickCards(r, st)
		}
	}
	return nil
}
```
with:
```go
// stepCards returns the cards a card step acts on, each once and only while
// its exact instance is still in its owner's discard pile (FR-ZON-004): the one
// place a card step checks this, as slotAssets is for assets. SubjectSelf is a
// Tracked item's own card. SubjectPick with a pile filter is the cards picked
// from discard piles, in pick order. SubjectTarget is the cards chosen for its
// pile slot that are still legal. Anything else has none.
func (g *Game) stepCards(r *resolution, st ability.Step) []*Instance {
	var cards []*Instance
	switch st.Subject {
	case ability.SubjectSelf:
		if it := r.it; it.Tracked && it.Inst != nil {
			cards = []*Instance{it.Inst}
		}
	case ability.SubjectTarget:
		if st.Target >= 0 && st.Target < len(r.cards) {
			cards = r.cards[st.Target]
		}
	case ability.SubjectPick:
		if st.Filter.InPile() {
			cards = g.pickCards(r, st)
		}
	}
	var out []*Instance
	for _, inst := range cards {
		if g.DiscardCard(inst.ID) == inst && !slices.Contains(out, inst) {
			out = append(out, inst)
		}
	}
	return out
}
```

No other file changes.

## Tests
No new tests and no test edits. Every existing test keeps its name and passes unchanged; the tests of the free deploy (`TestFR_TRN_003a_*`), the pile trigger and the return and to-deck steps, the card tests of *Veil Render*, *Reaver's Rush*, *Eerie Remnant* and the other discard-pile cards, and the byte-identical sims are the pin.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **No test changes** (#144): `git -C W diff autoloop/m2-wave-f.1 --stat -- '*_test.go'` prints nothing.
2. **Sims byte-identical** (#144): `go -C W tool task sim-compare -- autoloop/m2-wave-f.1` prints `[PASS] starter and stress sims match autoloop/m2-wave-f.1 byte for byte, exit codes included`.
3. **Only the listed edits:** `git -C W diff autoloop/m2-wave-f.1 --numstat -- pkg cmd` prints exactly one line: `24	20	pkg/engine/effects.go`.
4. **One check, in stepCards** (#144): `git -C W grep -n -e 'DiscardCard(' -e '\.Discard\b' -- pkg/engine/effects.go` prints exactly one line, `pkg/engine/effects.go:533:		if g.DiscardCard(inst.ID) == inst && !slices.Contains(out, inst) {`.
5. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
6. **Nothing else changed:** `git -C W diff autoloop/m2-wave-f.1 --stat` lists only `pkg/engine/effects.go` and this task file.
7. `gofmt -l W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/effects.go`

## Done When
- [x] Cases 1–7 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-f.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-13, lead-architect (wave F.1 run, session 1): wrote this task from board directive #144, the first of wave F.1 (design note Deliverable 6j). Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because every edit is exact and was dry-run. One addition to the directive's fix, to keep it a pure move for every input: `stepCards` keeps each card once (see Inputs). Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: #6, #7 (mode `serial`), #15 (no turn-boundary hook), #71, #91 (2), #106, #120, #141 (3), #142: not fired. Next: `kernel-engineer` on Flash.
- 2026-11-13, kernel-engineer (wave F.1 run, session 1): implemented task T-0227 (directive #144, pure move).
  - Status transition: ready -> in-progress -> review.
  - Changes in pkg/engine/effects.go (24 added, 20 deleted):
    - Updated `stepDeployFree` doc comment and removed redundant `g.DiscardCard(inst.ID) != inst ||` check from line 130.
    - Updated `stepRelocate` doc comment and removed redundant `g.DiscardCard(inst.ID) != inst` check loop filter.
    - Updated `stepCards` doc comment and consolidated discard pile membership verification into `g.DiscardCard(inst.ID) == inst && !slices.Contains(out, inst)` loop filter over collected cards, ensuring exact instance discard pile presence once per card.
  - Reuse: Reused existing lookup method `DiscardCard` (`pkg/engine/zones.go:163`), removing duplicate checks across card step callers.
  - Acceptance Cases & verification results:
    - Case 1: `git -C W diff autoloop/m2-wave-f.1 --stat -- '*_test.go'` printed nothing (no test changes).
    - Case 2: `go -C W tool task sim-compare -- autoloop/m2-wave-f.1` exited 0 and printed `[PASS] starter and stress sims match autoloop/m2-wave-f.1 byte for byte, exit codes included`.
    - Case 3: `git -C W diff autoloop/m2-wave-f.1 --numstat -- pkg cmd` printed exactly one line: `24\t20\tpkg/engine/effects.go`.
    - Case 4: `git -C W grep -n -e 'DiscardCard(' -e '\.Discard\b' -- pkg/engine/effects.go` printed exactly one line: `pkg/engine/effects.go:533:\t\tif g.DiscardCard(inst.ID) == inst && !slices.Contains(out, inst) {`.
    - Case 5: `go -C W run ./cmd/cardtown lint fences` exited 0 with `[PASS] Package code is free of card names, invalid markers and disallowed imports.`.
    - Case 6: `git -C W diff autoloop/m2-wave-f.1 --stat` listed only `pkg/engine/effects.go` and `.agents/tasks/m2/T-0227-card-step-pile-check.md`.
    - Case 7: `gofmt -l W/pkg` printed nothing; `go -C W tool task check` passed with exit code 0; `go -C W tool task sim` passed with exit code 0 (200 games, P0 63.5% / P1 35.0% / draws 1.5%, all invariants holding).
  - Ready for lead-architect review.
- 2026-11-13, lead-architect (wave F.1 run, session 1, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-f.1` as 3d7b2e4. Flash run `0f9491f9-3ce2-4ea9-a227-2fa4a8aa994f`: exit 0, `status: SUCCESS`, no `denied_actions` key, no `permission check failed` in the transcript; `.err` holds only agy's "root agent idle" notice. The engineer ran some commands with agy's `BypassSandbox` (its worktree and `go tool task check` hit temp-directory access errors in agy's sandbox); none was refused. Review of `git diff autoloop/m2-wave-f.1...m2/autoloop-wave-f.1/T-0227-card-step-pile-check` (8db9df1):
  - Scope: `pkg/engine/effects.go` and this file. No test file changed (case 1).
  - Pure move, checked against the diff: the `pkg/` diff's added and removed lines are identical to the lead's dry run, so it is exactly Edits 1-5. No test name or assertion changed.
  - Fences and placement: no card name; the check sits in `stepCards`, beside `slotAssets`, and no step body checks a card's pile (case 4).
  - Re-run by the reviewer in a scratch worktree of the task branch: cases 1, 3 and 4 as written; `gofmt -l pkg` clean; `cardtown lint fences` exit 0; `go tool task check` exit 0; `go tool task sim-compare -- autoloop/m2-wave-f.1` `[PASS]`; `go tool task sim` exit 0, clean, 200 games, P0 63.5% / P1 35.0% / draws 1.5%, P0 50/50 and P1 50/50 main cards executable.
  - The merge's tree matches the task tip's (`git diff` between them is empty), so the gates hold on the run branch. #144 is `applied` and archived.
