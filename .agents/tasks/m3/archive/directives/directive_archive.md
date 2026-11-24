# M3 Build Board — Applied Directives Archive

Applied directives moved from [`../../BOARD.md`](../../BOARD.md). Active and waiting directives remain on `BOARD.md`.

| # | Directive | Source | Status |
| :---: | :--- | :--- | :--- |
| 180 | **Fix `TestSimDecks_Case14_MissingFileNamesPath`, which fails `go tool task check` on `main`.** It checks that `SimDecks` names a missing deck file in its error, using `SOULLESS` because `data/decks/soulless-starter.txt` didn't exist ([decks_test.go:36](../../../../../pkg/sim/decks_test.go)). The user has since added the Soulless and Concordia starter decks (`e1af91b2`, `6c611f1e`, `05ba73e3`; M3 build order step 3), so the call succeeds and the test fails. Keep what it tests (a missing deck file is reported by its path) without depending on which decks `data/decks/` holds, for example by pointing it at a deck the repo can't have. Until this lands, every task's `check` gate fails, and a wave run stops on its first merge. | User, 2026-11-23 | applied (T-0294, merge eedd58eb) |
