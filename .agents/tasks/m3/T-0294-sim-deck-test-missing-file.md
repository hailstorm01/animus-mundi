---
id: T-0294
title: "The missing-deck sim test stops depending on which decks data/decks holds (directive #180; harness, flash)"
kind: harness
tier: flash
wave: none (board directive #180, blocking)
status: in-progress
branch: m3/T-0294-sim-deck-test-missing-file
depends_on: []
claims: [test:TestSimDecks_Case14_MissingFileNamesPath]
cards: []
requirements: []
---

## Goal
`TestSimDecks_Case14_MissingFileNamesPath` checks the same thing as before, that `SimDecks` names a missing deck file by its path, but against an empty temporary folder instead of `data/decks/`. It then passes whatever starter decks the repository holds, and `go tool task check` passes on `main` again. When this merges, #180 is `applied`.

## Inputs
- Board directive #180 (user, 2026-11-23), `open: before the next Deliverables task`.
- Where it is today (at `main` 6ffe3bc0): the test ([decks_test.go:36-49](../../../pkg/sim/decks_test.go)) asks `SimDecks` for the `SOULLESS` and `PATH` starter decks from `data/decks/` and expects an error naming `soulless-starter.txt`, because that file didn't exist. The user added the Soulless and Concordia starter decks (`e1af91b2`, `6c611f1e`, `05ba73e3`), so the call now succeeds, and the test fails with `expected error for missing soulless-starter.txt, got nil`. It is the only failure in `go tool task check` on `main`.
- Design (lead decisions):
  - Only the deck folder changes, to `t.TempDir()`, which holds no decklist. The real card sheets still load, the factions and the two assertions stay as they are, so the test still fails if `SimDecks` stops naming the file.
  - No non-test file changes. `SimDecks` is correct, and only the test's fixture was stale.
  - This is an authorized edit to an existing test (D2.4): the task names the test and the exact change below.
- The lead dry-ran this exact edit in a throwaway worktree of `main` at 6ffe3bc0. `go test ./pkg/sim -run TestSimDecks_ -count=1` passed, `gofmt -l pkg/sim` was clean, `go vet ./pkg/sim` was clean, and `go tool task check` exited 0. The edited file's blob is `f447312`.

## Acceptance Cases (harness tasks only; written by the lead)
1. `SimDecks` with an empty deck folder and factions `SOULLESS`, `PATH` → an error whose text contains `soulless-starter.txt` (the edited `TestSimDecks_Case14_MissingFileNamesPath`).
2. `data/decks/` holds every faction's starter deck (today's `main`) → case 1 still passes, because the test doesn't read `data/decks/`.
3. The other `TestSimDecks_` tests → pass unedited.

## Tests
- Edit `TestSimDecks_Case14_MissingFileNamesPath` in `pkg/sim/decks_test.go`, one line only. Replace line 41:
  ```go
  	decksDir := filepath.Join("..", "..", "data", "decks")
  ```
  with:
  ```go
  	decksDir := t.TempDir() // no decklists, so the first seat's file is missing whatever data/decks holds
  ```
  The line starts with one tab. Change nothing else in the file: `filepath` is still used by the `LoadCards` call above it.

## Commits
1. `test(T-0294): the missing-deck sim test uses an empty folder, not data/decks (directive #180)` with the edit above only.

There is no implementation commit, because no non-test file changes. No double quotes in the message: the allow-list refuses an escaped one.

## Allowed Files
- none (the one test edit above is the whole task)

## Done When
- [ ] The edited test and its neighbours pass: `go test ./pkg/sim -run TestSimDecks_ -count=1`
- [ ] `gofmt -l pkg/sim` prints nothing
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)

## Log
- 2026-11-23 lead-architect: written from directive #180, which blocks the next Deliverables task. Applied #181-#183's design-note parts first (6ffe3bc0). Status `ready` (a harness task skips `tests`), tier `flash`. Next: Deliverable 1, the M3 scaffold.
