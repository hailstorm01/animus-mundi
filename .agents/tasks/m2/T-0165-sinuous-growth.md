---
id: T-0165
title: Make Sinuous Growth executable (T1, kill this unit, return chosen Fungus or Plant card from your discard pile to your hand, wave D card, flash)
kind: card
tier: flash
wave: D
status: review
branch: m2/autoloop-wave-d/T-0165-sinuous-growth
depends_on: [T-0158, T-0162, T-0163, T-0164]
claims: [card:Sinuous Growth]
cards: [Sinuous Growth]
requirements: [FR-ACT-001, FR-ACT-002, FR-TGT-001]
---

## Goal
*Sinuous Growth* (a deck card, 4 copies): *"T1, kill this unit: Return chosen Fungus or Plant card from your discard pile to your hand."* It only combines T-0158's activated ability and costs, T-0162's subtype filter and T-0081's discard-pile target with `OpReturnToHand`. It is the first real card with an activated ability, so this is when sims first change because of one.

Covers:
- The definition in `pkg/engine/cards/animus.go`, as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/units-and-structures-animus.csv` line 11): Animus Unit, subtypes Plant and Fungus, deploys to Reserves.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 143.
- FRD 01 v3.21: `FR-ACT-001` (line 341, which names this card), `SVY-R36` (line 342, which names it too), `FR-ACT-002` (line 343), `FR-TGT-001`, `SVY-R2`.
- Rulings: `SVY-R2`, `SVY-R36`.

### Definition (the engineer adds it to `animusDefs`)
- `Text`: the printed line, as above.
- `Activated`: one ability.
  - `Costs`, in printed order: `{Kind: ability.CostResources, Tac: 1}`, then `{Kind: ability.CostKillSelf}`.
  - `Effect.Targets`: one exactly-one slot, `Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Zone: ability.ZoneDiscard, Subtypes: []string{"Fungus", "Plant"}}`. `Kinds` stays empty, because the text says "card": a Fungus structure counts.
  - `Effect.Steps`: one, `{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget, Target: 0}`.
- Remove the name from `cards.NotYetExecutable`.

A lead probe with this definition, since reverted, passes every test except `TestPlay_Case1_ScriptedSessionFinishes`, which T-0164 fixes, and the starter sim stays clean: 46/50 Animus main cards executable, P0 84.5% / P1 14.0% / draws 1.5% over 200 games. So `sim-compare` will differ here and isn't a gate.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_animus_sinuous_test.go`, from the CSV text, loading the real card (`s.card`). Player 0 activates. Give player 0 Tactics with a pool helper as `activated_test.go` does. `activated_test.go` has `activate` and `offersActivate`. "Not offered" means both: `offersActivate` is false and `s.reject(activate(...))` holds.
- `TestCard_SinuousGrowth_ReturnsTheChosenCard`: two matching cards in player 0's pile, player 0 choosing the older one. The chosen card is in hand, the other stays in the pile, *Sinuous Growth* is in its owner's discard pile, and the pool lost exactly 1 Tactics and no Requisition.
- `TestCard_SinuousGrowth_ListsFungusOrPlantCardsOfAnyType`: the pile holds *Animoceps Shambler* (Fungus), *Spoil Forest Colossus* (Plant, Giant), *Animoceps Spore Cloud* (a Fungus structure), *Funeral Fungus* (a Maneuver with no subtype, whatever its name says) and a unit of another subtype. The target question lists exactly the first three.
- `TestCard_SinuousGrowth_OwnCardCannotBeChosen` (`SVY-R2`, `SVY-R36`): with an empty pile, or a pile holding only non-matching cards, it is not offered although its own card is a Plant and a Fungus. Nothing is paid and it stays on the battlefield.
- `TestCard_SinuousGrowth_AnotherCopyCanReturnTheFirst` (`SVY-R2`): two copies on the battlefield and one other matching card in the pile. After the first copy resolves, the second is offered, and its question lists the first copy's card.
- `TestCard_SinuousGrowth_OpponentsPileDoesNotCount` (`FR-TGT-001`): the only matching card is in player 1's discard pile. Not offered.
- `TestCard_SinuousGrowth_NeedsOneTactics` (`FR-ACT-001`): a matching card in the pile, 0 Tactics and plenty of Requisition. Not offered, and nothing changes.
- `TestCard_SinuousGrowth_WorksWhileExhausted` (`FR-ACT-002`): an Exhausted *Sinuous Growth* can be activated, because its cost has no exhaust.
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestView_Case8_ExecutableFlag` (seat 0's seed-1 hand still holds *Veil Render*), `TestProgress_Case4_Coverage` and `TestEvents_Case4_EveryActionKind`.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/animus.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [x] The card tests pass: `go test ./pkg/engine -run TestCard_SinuousGrowth`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-11-04 lead-architect (wave D run, session 8): queued behind T-0164, with T-0164's task commit (D2.4). Written against b7a0778, with the lead's probe above. Check it against the code of the day, and read the board's directives again, before it leaves `queued`. *Boneyard Scavenger* and *Ruin Partisans* stay a card task of their own, as the run file's `Kept` line says.
- 2026-11-04 lead-architect (wave D run, session 9): checked against 59475b0. The card's CSV text, the subtypes of the cards the tests name, `FR-ACT-001`, `SVY-R36` and `FR-ACT-002` (FRD 01 v3.21), and the shapes in the Definition all read as written; #105's part (1) is applied, and parts (2), (3) go with *Cliff of Portents*. Board re-read, no new directive. Set to `tests`.
- 2026-11-04 rules-test-author: wrote `pkg/engine/cards_animus_sinuous_test.go`: TestCard_SinuousGrowth_ReturnsTheChosenCard, _ListsFungusOrPlantCardsOfAnyType, _OwnCardCannotBeChosen (empty pile; non-matching pile), _AnotherCopyCanReturnTheFirst, _OpponentsPileDoesNotCount, _NeedsOneTactics, _WorksWhileExhausted. Helpers requested: none (reuses `activate`, `offersActivate`, `actNotOffered`, `setPool`, `refs`). Run now: the four positive tests fail with "no activated ability 0" (no definition yet); the three "not offered" tests pass trivially until the definition exists, and guard against over-offering after. No cases left out.
- 2026-11-04 lead-architect: tests approved (D4, round 1), each checked against `FR-ACT-001`, `SVY-R36`, `SVY-R2`, `FR-ACT-002` and `FR-TGT-001`. A lead probe with the Definition above, since reverted, passes all seven and every package. Set to `ready`.
- 2026-11-04 kernel-engineer: added Sinuous Growth ability definition to pkg/engine/cards/animus.go using CostResources (Tac: 1), CostKillSelf, Target filter (AllegianceAllied, ZoneDiscard, Subtypes: Fungus/Plant), and OpReturnToHand. Removed Sinuous Growth from NotYetExecutable in pkg/engine/cards/cards.go. Reused existing ability primitives and cost structures without adding new shapes. Verified: go test ./pkg/engine -run TestCard_SinuousGrowth passed; go test ./pkg/engine -run TestStarterSet_Executable passed; go tool task check passed clean; go tool task sim ran clean on starter decks (P0 main cards executable 46/50, 200 games 0 invariant violations). Set to review.
