---
id: T-0003
title: Validate deck construction (FR-SET-001)
kind: shape
tier: sonnet
wave: none (design note Deliverable 2, first half)
status: tests
branch: m2/T-0003-validate-deck
depends_on: [T-0002]
claims: [func:ValidateDeck, var:ErrIllegalDeck]
cards: []
requirements: [FR-SET-001]
---

## Goal
`engine.ValidateDeck` says whether a `DeckList` is construction-legal under FR-SET-001, and names every violation. It is a check the caller runs. `NewGame` is unchanged, because scenario and stress decks are deliberately not legal. T-0004 uses it to validate the starter decklists and `cardtown sim -deck`.

## Inputs
- FRD 01: `FR-SET-001` (Domain 1, Section 1.1, lines 51–63 of `docs/frd/01_Game_Rules_FRD.md`). Also the card-anatomy row for **Name** (line 110): copy limits count by printed name.
- Card data: the `rarity` column holds `COMMON`, `UNCOMMON`, `RARE`, `MYTHIC` (and `HQ` on Headquarters). The `faction` column holds `ANIMUS`, `PATH`, `CONCORDIA`, `SOULLESS`, `NEUTRAL`. *Scratch Company* is a NEUTRAL COMMON unit with the subtype `Scratch Company`.
- Rulings that apply: none. Design note: `docs/architecture/m2-ability-model.md`, Deliverables item 2 (split by board directive #2).

### API the tests code against (the engineer creates it in `pkg/engine/deck.go`)
```go
// ErrIllegalDeck is wrapped by every deck-construction violation (FR-SET-001).
var ErrIllegalDeck = errors.New("illegal deck")

// ValidateDeck returns nil if d is construction-legal under FR-SET-001.
// Otherwise it returns every violation joined (errors.Join), each wrapping
// ErrIllegalDeck and naming the offending card where there is one.
func ValidateDeck(d DeckList) error
```
Tests check `errors.Is(err, ErrIllegalDeck)`. Where a card is at fault, they check that `err.Error()` contains its name. Asserting any other wording is an implementation detail.

### Lead decisions (technical, not rules)
- **Copy counting** goes by `Card.Name`, so two distinct `*Card` values with the same name count together.
- **Factions** compare case-insensitively, as `FactionDeck` already does.
- **Scratch Company** is identified by the subtype (`HasSubtype("Scratch Company")`), as elsewhere in the kernel.
- **A main-deck card whose rarity is not one of the four FR-SET-001 rarities** is a violation (fail closed on bad data). A test may pin this.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
<!-- rules-test-author fills this in. Cover the main rule and every edge FR-SET-001 names: exactly 50 is legal, 49 is not, and there is no maximum; HQ missing, not a Headquarters, or Neutral; a Headquarters card in the main deck; off-faction versus Neutral cards; Scratch Company; each rarity at its limit and one over; several violations reported together. File: pkg/engine/deck_test.go. Use synthetic cards or realCard(name) from scenario_test.go. -->

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/deck.go` (new: `ErrIllegalDeck`, `ValidateDeck`)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_SET_001'`
- [ ] `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean
- [ ] `cardtown lint rules` shows FR-SET-001 pinned
- [ ] `go run ./cmd/cardtown lint --all` passes

## Log
- 2026-10-19 lead-architect: Wrote the task, applying board directive #2 (first half). Data checked: all Animus and Path names use straight apostrophes, the HQ rarity is `HQ`, and Scratch Company is COMMON and NEUTRAL, so only the token prohibition catches it. Next: `rules-test-author`.
