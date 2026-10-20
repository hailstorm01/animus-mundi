---
id: T-0004
title: Load starter decklists and run them in cardtown sim
kind: harness
tier: flash
wave: none (design note Deliverable 2, second half)
status: queued
branch: m2/T-0004-starter-decklists
depends_on: [T-0003]
claims: [format:decklist, func:LoadDeck, data:starter-decklists, cli:sim-deck]
cards: []
requirements: []
---

## Goal
The two starter decks from the design note live as text files under `data/decks/`. `engine.LoadDeck` reads one into a `DeckList`, and `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` plays them after checking them with `ValidateDeck`.

## Inputs
- Decklists: `docs/architecture/m2-ability-model.md`, section Starter Decks (the two tables; names as printed in `data/cards`).
- `ValidateDeck` and `ErrIllegalDeck` from T-0003 (`pkg/engine/deck.go`).
- Current sim wiring: `runSim` in `cmd/cardtown/main.go`, and `FactionDeck` in `pkg/engine/sim.go` (stays the default).
- Rulings that apply: none.

### Format (`data/decks/<faction>-<name>.txt`, `<faction>` lowercase)
```text
# Animus starter deck (docs/architecture/m2-ability-model.md)
HQ: Animoceps Rager
4 Despoiler Larva
1 Spoil Forest Amalgam
```
- One entry per line: `<copies> <card name>`, or `HQ: <card name>` exactly once. Blank lines and lines starting with `#` are ignored. Surrounding whitespace is trimmed.
- Names match `data/cards` exactly (case and punctuation).
- `LoadDeck` only parses and resolves names. Legality is `ValidateDeck`'s job, so a resolvable but illegal list loads without error.

### API
```go
// LoadDeck reads a decklist file and resolves its names against cards.
func LoadDeck(path string, cards map[string]*Card) (DeckList, error)
```
`Main` holds one `*Card` per copy, in file order.

## Acceptance Cases (harness tasks only; written by the lead)
Fixture decklists are written in `t.TempDir()`. Synthetic `cards` maps are fine except in cases 10–12.
1. A valid file → `HQ` is the named card, and `Main` has one entry per copy in file order.
2. Comments, blank lines and surrounding whitespace → ignored.
3. An unknown card name → error naming the name and its line number.
4. No `HQ:` line, or two of them → error.
5. A count that is not a positive integer (`0 X`, `-1 X`, `two X`, or a name with no count) → error naming the line number.
6. The same card on two lines → error naming the card (lists stay canonical).
7. A resolvable but illegal list (49 cards) → `LoadDeck` returns no error, and `ValidateDeck` on the result wraps `ErrIllegalDeck`.
8. A missing file → error that is not a panic.
9. Names are exact: `despoiler larva` for *Despoiler Larva* → error, as in case 3.
10. `data/decks/animus-starter.txt` against the real cards → HQ *Animoceps Rager*, 50 main cards, 21 distinct names, and `ValidateDeck` returns nil.
11. `data/decks/path-starter.txt` against the real cards → HQ *Eternal Watchman*, 50 main cards, 23 distinct names, and `ValidateDeck` returns nil.
12. Every copy count in both files matches the design note's Starter Decks tables. The test holds the expected name → count maps literally, as data. It does not parse the design note.
13. `cardtown sim -deck starter -p0 ANIMUS -p1 PATH -games 20` → exit 0. The header names both decklist files and prints, per deck, how many of its 50 main cards are executable now (`Supported()`).
14. `cardtown sim -deck starter -p0 SOULLESS` (no such file) → `[ERROR]` naming the missing path, exit 1.
15. `cardtown sim -deck <name>` where the file loads but `ValidateDeck` fails → `[ERROR]` with the violations, exit 1. Cover this with a Go test of the function `runSim` calls to build its decks (see Allowed Files), not with the binary.
16. `cardtown sim` without `-deck` → the same stress-deck behavior and header as today.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
- Harness task. Write one test per acceptance case: cases 1–12 in `pkg/engine/deck_load_test.go` as `TestLoadDeck_<Case>`, and cases 13–16 against the deck-building function, with a binary smoke run of case 13 recorded in the Log. Commit them first as `test(T-0004): …` with compiling stubs. Case 16 may pass at the stub commit.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None.

## Allowed Files
- `pkg/engine/deck.go` (`LoadDeck`, plus `SimDecks(cards map[string]*Card, decksDir, name string, factions [2]string) ([2]DeckList, error)`, which `runSim` calls when `-deck` is set: it loads `<decksDir>/<lowercase faction>-<name>.txt` for each seat and validates it)
- `data/decks/animus-starter.txt` (new)
- `data/decks/path-starter.txt` (new)
- `cmd/cardtown/main.go` (the `-deck` flag on `sim`, the header lines, the usage text)

## Done When
- [ ] `go test ./pkg/engine -run 'TestLoadDeck|TestSimDecks'` passes, with one test per acceptance case
- [ ] `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean
- [ ] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0
- [ ] `go run ./cmd/cardtown sim` output is unchanged from the integration branch
- [ ] `go run ./cmd/cardtown lint --all` passes

## Log
- 2026-10-19 lead-architect: Wrote the task (board directive #2, second half). It stays `queued` until T-0003 is `done`, and then goes to `ready` for `@kernel-engineer T-0004` (tier flash, Antigravity). Reviewer: check the two decklist files row by row against the design note tables, not only through case 12.
