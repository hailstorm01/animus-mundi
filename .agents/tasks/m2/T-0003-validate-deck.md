---
id: T-0003
title: Validate deck construction (FR-SET-001)
kind: shape
tier: sonnet
wave: none (design note Deliverable 2, first half)
status: review
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
- [x] The listed tests pass: `go test ./pkg/engine -run 'TestFR_SET_001'`
- [x] `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean
- [x] `cardtown lint rules` shows FR-SET-001 pinned
- [x] `go run ./cmd/cardtown lint --all` passes

## Log
- 2026-10-19 lead-architect: Wrote the task, applying board directive #2 (first half). Data checked: all Animus and Path names use straight apostrophes, the HQ rarity is `HQ`, and Scratch Company is COMMON and NEUTRAL, so only the token prohibition catches it. Next: `rules-test-author`.
- 2026-10-19 rules-test-author: Wrote 22 `TestFR_SET_001_*` tests in `pkg/engine/deck_test.go`. Helpers requested: none (fixtures are local to deck_test.go). `go vet ./pkg/engine` fails only on undefined `ValidateDeck` and `ErrIllegalDeck`, as expected until the engineer adds them; `cardtown lint rules` passes. Left out: wording beyond `errors.Is` and card names, and the exact count of joined errors, because FRD 01 does not fix them. Next: lead-architect review.
- 2026-10-19 lead-architect: Test review round 1: **returned for two notes.** The other 20 tests match FR-SET-001 and the lead decisions. Both notes are about isolation: each illegal-deck test must fail for exactly the rule it names, so that an implementation missing that rule fails the test.
  1. `TestFR_SET_001_HQInMainDeckIsIllegal`: the stray Headquarters keeps rarity `HQ`, so the unknown-rarity rule also rejects it (and that error names the card too). The test passes even with no Headquarters-in-main-deck check. Give the stray card a rarity of `COMMON`, so only the Headquarters prohibition can fire.
  2. `TestFR_SET_001_NeutralHQIsIllegal`: the main deck is 50 ANIMUS cards under a NEUTRAL Headquarters, so a faction-mismatch error fails the deck regardless, and its text may contain the HQ's name. Build the main deck from 50 NEUTRAL cards (legal under any HQ), so only the "Headquarters cannot be Neutral" rule can fire.
- 2026-10-19 rules-test-author: Round 1 notes fixed. (1) `HQInMainDeckIsIllegal`: stray HQ now has rarity COMMON, so only the HQ-in-main-deck rule can fire. (2) `NeutralHQIsIllegal`: main deck is now 50 NEUTRAL cards, so only the Neutral-HQ rule can fire. No other changes. Next: lead-architect review.
- 2026-10-19 lead-architect: Test review round 2: **approved** (checked against the diff of `5316a77`). Both isolation notes are fixed. Status `ready`, and `kernel-engineer` (Sonnet subagent) is dispatched. The engineer must not edit `deck_test.go`.
- 2026-10-19 kernel-engineer: Added `pkg/engine/deck.go` (only file touched besides this task): `ErrIllegalDeck` and `ValidateDeck`. Reused: `TypeHeadquarters`, `Card.HasSubtype("Scratch Company")`, `strings.EqualFold` faction compare as in `FactionDeck`; nothing else in `pkg/engine` already validated decks. Checks: size >= 50 (no max); HQ non-nil, a Headquarters card, not Neutral; no Headquarters or Scratch Company in main; faction matches HQ or is Neutral; copy limits by `Card.Name` per rarity (Common 4, Uncommon 3, Rare 2, Mythic 1); unknown rarity in main is a violation. All violations are joined with `errors.Join`. Commands: `go test ./pkg/engine -run TestFR_SET_001` ok; `go test ./...` ok; `go vet ./...` clean; `gofmt -l pkg cmd` empty; `go run ./cmd/cardtown lint rules` passes and FR-SET-001 is no longer in the not-yet-executable list; `go run ./cmd/cardtown lint --all` passes. Reviewer notes: (a) a nil card in Main is reported as a violation; (b) faction mismatch is skipped when the HQ is missing/not-HQ/Neutral, so that fault is reported once, not once per card; (c) rarity matching is case-insensitive; (d) a stray Headquarters in Main gets only the HQ-in-main error, not the rarity one.
- 2026-10-19 lead-architect: Code review round 1: **changes requested (one note).** Scope is only `deck.go`, the tests are untouched since `5316a77`, and the other choices are accepted (nil card, the faction check skipped without a usable HQ, case-insensitive rarity, no unknown-rarity error for a stray HQ, sorted copy-limit errors). Skipping the `in-progress` commit is noted and harmless in serial mode.
  1. **Rules fidelity (D5.4):** FR-SET-001 says a Headquarters "must belong to a specific faction (Concordia, Soulless, Path, or Animus) and cannot be Neutral". `ValidateDeck` rejects only `NEUTRAL`, so an HQ with an empty or unknown faction passes and every non-Neutral card then fails as off-faction. Accept only those four factions (case-insensitive) for the HQ. Keep the Neutral message, and report any other value as a violation naming the HQ and its faction. Don't edit tests; this branch has no test for it, and the reviewer checks it by reading the code.
- 2026-10-19 kernel-engineer: Round 1 note 1 answered. `ValidateDeck` now accepts an HQ faction only if it is Concordia, Soulless, Path or Animus (case-insensitive, via new `hqFactions` list in `deck.go`). The Neutral case keeps its message; any other value (including empty) reports "Headquarters <name> has faction "<value>", must be one of ..." and skips the per-card faction check, as for Neutral. Only `deck.go` changed; no tests edited. Gates: `go test ./...` ok, `go vet ./...` clean, `gofmt -l pkg cmd` empty, `cardtown lint rules` passes (FR-SET-001 pinned), `cardtown lint --all` passes.
- 2026-10-19 kernel-engineer: Correction: the previous commit (2b5a882) carried only this task file's Log entry; my edit script failed (no Python), so deck.go was unchanged there. The deck.go change is in the next commit, and the gates above were re-run after it: all pass.
