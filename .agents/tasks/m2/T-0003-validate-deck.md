---
id: T-0003
title: Validate deck construction (FR-SET-001)
kind: shape
tier: sonnet
wave: none (design note Deliverable 2, first half)
status: tests-review
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
File: `pkg/engine/deck_test.go` (synthetic cards; local fixtures `deckHQ`, `deckCard`, `legalMain`, `copies`, `legalDeck`, `wantLegal`, `wantIllegal`).
- `TestFR_SET_001_LegalDeckAccepted`: a 50-card faction deck with a matching HQ is legal.
- `TestFR_SET_001_FiftyCardsIsLegal`: exactly 50 is legal (Path deck).
- `TestFR_SET_001_FortyNineCardsIsIllegal`: 49 is illegal.
- `TestFR_SET_001_EmptyMainDeckIsIllegal`: an empty main deck is illegal.
- `TestFR_SET_001_NoMaximumDeckSize`: 200 cards is legal.
- `TestFR_SET_001_MissingHQIsIllegal`: nil HQ is illegal.
- `TestFR_SET_001_HQMustBeHeadquartersCard`: an HQ that is not a Headquarters card is illegal and named.
- `TestFR_SET_001_NeutralHQIsIllegal`: a Neutral HQ is illegal and named.
- `TestFR_SET_001_HQInMainDeckIsIllegal`: a Headquarters card in the main deck is illegal and named.
- `TestFR_SET_001_OffFactionCardIsIllegal`: a Path card in an Animus deck is illegal and named.
- `TestFR_SET_001_OffFactionIsIllegalForEveryPairing`: every ordered pair of distinct factions is illegal.
- `TestFR_SET_001_NeutralCardsLegalInAnyDeck`: a Neutral card is legal under each of the four factions.
- `TestFR_SET_001_FactionComparedCaseInsensitively`: lead decision; "Animus" cards under an "ANIMUS" HQ are legal.
- `TestFR_SET_001_ScratchCompanyInMainDeckIsIllegal`: a Neutral Common with the Scratch Company subtype is illegal and named.
- `TestFR_SET_001_CommonLimitIsFour`: 4 legal, 5 illegal.
- `TestFR_SET_001_UncommonLimitIsThree`: 3 legal, 4 illegal.
- `TestFR_SET_001_RareLimitIsTwo`: 2 legal, 3 illegal.
- `TestFR_SET_001_MythicLimitIsOne`: 1 legal, 2 illegal.
- `TestFR_SET_001_CopyLimitCountsByPrintedName`: lead decision; three distinct `*Card` values with one name count together.
- `TestFR_SET_001_LimitsApplyPerName`: two Mythics with different names are legal.
- `TestFR_SET_001_UnknownRarityIsIllegal`: lead decision; an unrecognised rarity in the main deck is illegal and named.
- `TestFR_SET_001_AllViolationsReportedTogether`: a deck with faction, copy-limit and size faults names both offending cards in one error.

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
- 2026-10-19 rules-test-author: Wrote 22 `TestFR_SET_001_*` tests in `pkg/engine/deck_test.go`. Helpers requested: none (fixtures are local to deck_test.go). `go vet ./pkg/engine` fails only on undefined `ValidateDeck` and `ErrIllegalDeck`, as expected until the engineer adds them; `cardtown lint rules` passes. Left out: wording beyond `errors.Is` and card names, and the exact count of joined errors, because FRD 01 does not fix them. Next: lead-architect review.
- 2026-10-19 lead-architect: Test review round 1: **returned for two notes.** The other 20 tests match FR-SET-001 and the lead decisions. Both notes are about isolation: each illegal-deck test must fail for exactly the rule it names, so that an implementation missing that rule fails the test.
  1. `TestFR_SET_001_HQInMainDeckIsIllegal`: the stray Headquarters keeps rarity `HQ`, so the unknown-rarity rule also rejects it (and that error names the card too). The test passes even with no Headquarters-in-main-deck check. Give the stray card a rarity of `COMMON`, so only the Headquarters prohibition can fire.
  2. `TestFR_SET_001_NeutralHQIsIllegal`: the main deck is 50 ANIMUS cards under a NEUTRAL Headquarters, so a faction-mismatch error fails the deck regardless, and its text may contain the HQ's name. Build the main deck from 50 NEUTRAL cards (legal under any HQ), so only the "Headquarters cannot be Neutral" rule can fire.
- 2026-10-19 rules-test-author: Round 1 notes fixed. (1) `HQInMainDeckIsIllegal`: stray HQ now has rarity COMMON, so only the HQ-in-main-deck rule can fire. (2) `NeutralHQIsIllegal`: main deck is now 50 NEUTRAL cards, so only the Neutral-HQ rule can fire. No other changes. Next: lead-architect review.
