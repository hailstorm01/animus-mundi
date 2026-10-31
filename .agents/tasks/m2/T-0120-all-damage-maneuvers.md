---
id: T-0120
title: Make Despoiler's Emergence and Deafening Aria executable (damage to all opposing assets in a zone, wave C card, flash)
kind: card
tier: flash
wave: C
status: ready
branch: m2/T-0120-all-damage-maneuvers
depends_on: [T-0116, T-0117]
claims: [card:Despoiler's Emergence, card:Deafening Aria]
cards: [Despoiler's Emergence, Deafening Aria]
requirements: [FR-MNV-001, FR-TGT-002, FR-REP-004]
---

## Goal
The first two Maneuver cards run in the kernel: *Despoiler's Emergence* (Animus starter deck) and *Deafening Aria* (non-deck, kept by #14 (c): every shape it needs exists, and it is the same shape as the Emergence). Each is plain `ability` data in `pkg/engine/cards/animus.go`: a Maneuver `Effect` of `OpDealDamage` steps on `SubjectAll` with a zone filter. No new shape.

## Inputs
- Cards (`data/cards/maneuvers-animus.csv`):
  - line 15, *Despoiler's Emergence*, Animus Maneuver, cost 0 Requisition / 4 Tactics, two rules-text lines:
    - "Deal 4 damage to all opposing assets in the Main Theater."
    - "Deal 2 damage to all opposing non-Scratch assets in opponents' Reserves."
  - line 9, *Deafening Aria*, Animus Maneuver, cost 0 / 2, one line:
    - "Deal 1 damage to all opposing units in the Main Theater."
- Survey rows: `docs/design/m2_starter_set_ability_survey.md` lines 129 (Deafening Aria) and 132 (Despoiler's Emergence): `DEAL_DAMAGE` to all opposing assets or units in a zone.
- FRD 01:
  - `FR-MNV-001` (lines 332-334): enacted from Hand, costs paid on declaration, the effect runs when it resolves, then the card goes to its owner's Discard pile. A Maneuver with no target slot needs no legal entity to be enacted (`SVY-R29`).
  - `FR-TGT-002` (line 362): *"all assets"* applies globally without targeting, so nothing is chosen, no decision is asked and Ward excludes nothing; the set is whatever matches when the step resolves.
  - `FR-REP-004` (line 484 on): each asset takes its own damage instance; Armor absorbs first.
- Shapes it uses, all `done`: `Def.Effect`, `ActEnact` and `StackManeuver` (T-0116), the `SVY-R29` enact legality (T-0117), `OpDealDamage` with `SubjectAll` and a `Filter` (as on *Craterfall Ancient*'s death trigger in `cards/animus.go`), `Filter.Zone` (T-0052), `Filter.Scratch` (`ScratchExcluded`).
- Rulings: none beyond the FRD text above. "Non-Scratch assets" are every opposing asset without the Scratch Company subtype: structures match, face-down Scratch Company units don't (`matchesPrinted`, `targeting.go`). A Headquarters is in neither the Main Theater nor the Reserves, so neither card damages one.

### Definitions (the engineer adds them to `animusDefs`)
- *Despoiler's Emergence*: `Text` holds both printed lines, in printed order. `Effect.Steps`:
  1. `{Op: ability.OpDealDamage, Amount: ability.Value{N: 4}, Subject: ability.SubjectAll, Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Zone: ability.ZoneMainTheater}}`
  2. `{Op: ability.OpDealDamage, Amount: ability.Value{N: 2}, Subject: ability.SubjectAll, Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Scratch: ability.ScratchExcluded, Zone: ability.ZoneReserves}}`
- *Deafening Aria*: `Text` holds its line. `Effect.Steps`:
  1. `{Op: ability.OpDealDamage, Amount: ability.Value{N: 1}, Subject: ability.SubjectAll, Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit, Zone: ability.ZoneMainTheater}}`
- No `Targets`, `Triggers` or `Modes`.
- Remove both names from `cards.NotYetExecutable`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written in a new file, `pkg/engine/cards_animus_maneuvers_test.go`, from the CSV text, loading the real cards (`s.card`). Give the enacting player enough Tactics to pay. Damage is checked with `wantDamage`; that a step dealt nothing is checked on the asset and by the absence of a damage event for it.
- `TestCard_DespoilersEmergence_FourToEveryOpposingAssetInTheMainTheater`: opposing units and an opposing structure in the Main Theater each take 4. Allied assets in the Main Theater take nothing.
- `TestCard_DespoilersEmergence_TwoToOpposingNonScratchAssetsInReserves`: in the opponent's Reserves, a face-up unit and a structure each take 2; a face-down Scratch Company unit takes nothing. Allied assets in Reserves take nothing.
- `TestCard_DespoilersEmergence_NoHeadquartersIsDamaged`: neither Headquarters takes damage.
- `TestCard_DespoilersEmergence_ArmorAbsorbsEachInstance`: an opposing Main Theater unit with Armor loses Armor first (`FR-REP-004`).
- `TestCard_DespoilersEmergence_LethalAssetsDieAfterItResolves`: opposing assets left lethal go to their owners' discard piles in the death sweep after the Maneuver resolves, and the Maneuver itself ends in its owner's discard pile.
- `TestCard_DespoilersEmergence_EnactableWithNoOpposingAsset`: with no opposing asset in either zone, enacting it is a legal action (no target slot, `SVY-R29`); it resolves, damages nothing and goes to the discard pile.
- `TestCard_DespoilersEmergence_AsksNothing`: no `Target` or pick decision is asked of either player (`FR-TGT-002`).
- `TestCard_DespoilersEmergence_PaysFourTactics`: enacting it spends 4 Tactics and no Requisition.
- `TestCard_DeafeningAria_OneToEveryOpposingUnitInTheMainTheater`: each opposing unit in the Main Theater takes 1. An opposing structure in the Main Theater, opposing units in Reserves and allied units take nothing.
- `TestCard_DeafeningAria_KillsOneHealthUnits`: an opposing 1-health unit in the Main Theater dies in the sweep after it resolves, and the Maneuver ends in its owner's discard pile.
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the two names removed.

## Helpers Needed
- None expected. The tests use existing helpers (`newScene`, `put`, `putScratch`, `give`, `card`, `do`, `wantDamage`, `alive`) and those in `maneuver_test.go` (`enactAction`, `offersEnact`, `lastDiscard`); the new file may add its own setup helpers.

## Allowed Files
- `pkg/engine/cards/animus.go` (the two definitions)
- `pkg/engine/cards/cards.go` (remove the two names from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_DespoilersEmergence`
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_DeafeningAria`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] Both cards load with no unsupported lines (`TestStarterSet_Executable` passes with the names removed)

## Log
- 2026-10-30 lead-architect: wave C run, session 6. No `main` commits to sync (`git log autoloop/m2-wave-c..main` is empty). Not `WAVE-COMPLETE`: 22 deck cards are on `NotYetExecutable`. Directives read on the run branch: no `open: before` directive. #14 (a): deck cards first; *Despoiler's Emergence* is a deck card, and *Deafening Aria* joins under #14 (c) (same shape, every shape done, already in the run file's `Kept`). #20 (b) stays open for *Steel Rain*'s discard-cost task. Waiting triggers checked: #8 (no `deck.go` change), #15 (no turn hook), #58 (no exhaust in `claims:`), #59 (2) (no repeated targets), #61's fifth question (no modal card in `cards:`), #71 (both steps act on assets, not on each player), #88's distinctness (no *Ambush Tactics*). None has fired. Taken ahead of T-0119's suggested *Rapturous Sermon* with *Ensnaring Mist*: *Ensnaring Mist* has two exactly-one "chosen unit" slots, and FRD 01 doesn't say whether one unit may fill both (`SVY-R11` covers one clause, `SVY-R4` only "for each" repeats); the task that takes it, or *Rapturous Sermon* alone, should check that first. Duplicates searched: no new code; the definitions reuse `OpDealDamage` on `SubjectAll` exactly as *Craterfall Ancient*'s death trigger does. Two non-test files. Status `tests`; dispatching `rules-test-author`.
- 2026-10-30 rules-test-author: tests written in `pkg/engine/cards_animus_maneuvers_test.go`, all ten named in the Tests section: `TestCard_DespoilersEmergence_{FourToEveryOpposingAssetInTheMainTheater, TwoToOpposingNonScratchAssetsInReserves, NoHeadquartersIsDamaged, ArmorAbsorbsEachInstance, LethalAssetsDieAfterItResolves, EnactableWithNoOpposingAsset, AsksNothing, PaysFourTactics}` and `TestCard_DeafeningAria_{OneToEveryOpposingUnitInTheMainTheater, KillsOneHealthUnits}`. Helpers requested: none (file adds `enactFor`, `damageEventsOn`; reuses `inDiscardPile` from `cards_animus_test.go`). Cases left out: none. `go vet` clean, `cardtown lint rules` passes; all ten fail now because the cards are not yet executable. Status `tests-review`.
- 2026-10-30 lead-architect: tests approved. Each of the ten checks its outcome against the card text and FRD 01: 4 to every opposing asset in the Main Theater (units and a structure) and none to allied ones; 2 to an opposing face-up unit and structure in Reserves, none to a face-down Scratch Company or allied assets ("non-Scratch", `matchesPrinted`); neither Headquarters damaged; Armor absorbs first (`FR-REP-004`); lethal assets die in the sweep after resolution and the card ends in its owner's discard pile (`FR-MNV-001`); enactable with nothing to hit, since there is no target slot (`SVY-R29`); no decision asked, even with a Ward unit present (`FR-TGT-002`); the 4 and 2 Tactics costs; *Deafening Aria* hits only opposing Main Theater units. The asserts read damage, Armor, zones, the log's damage events, decisions and pools, not internals. No existing test edited. All ten fail on "needs rules the kernel does not implement yet", the missing definitions. Status `ready`; dispatching `kernel-engineer` (flash) through `agy`.
