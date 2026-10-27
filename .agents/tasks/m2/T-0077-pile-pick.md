---
id: T-0077
title: An untargeted pick from your discard pile, which moves the picked card, with a top-down timeout (SVY-R14, SVY-R16, directive #59 (1), wave B shape, sonnet)
kind: shape
tier: sonnet
wave: B
status: ready
branch: m2/T-0077-pile-pick
depends_on: [T-0072, T-0076]
claims: [filter:zone-discard, subject:pick-from-pile, convention:card-steps, fallback:pile-top-down]
cards: []
requirements: [FR-TGT-001, FR-MOD-001, FR-TRG-005]
---

## Goal
An untargeted pick (`ability.SubjectPick`) can pick cards in a discard pile instead of battlefield assets: its `Filter.Zone` is the new `ability.ZoneDiscard`. The card steps `OpPutOnDeck` and `OpReturnToHand` move each picked card, and report how many they moved, so "if you do" works after them. A timeout picks cards in a pile from the top down (`FR-MOD-001` §6, board directive #59 (1)). The first card, *Arkha, Sun's Fury* ("Whenever this unit attacks, you may put a unit card from your discard pile on the bottom of your deck. If you do, deal 3 damage to chosen opposing Headquarters."), is T-0078.

## Inputs
- FRD 01:
  - `FR-TGT-001` Zone criterion (line 322): "from your discard pile" is a zone restriction.
  - `FR-TGT-001` "Untargeted Selections (`SVY-R3`)" (line 329): no "chosen" means no target. The pick is made as the clause resolves, from everything that matches then, and with no match only that clause does nothing.
  - `FR-TGT-001` "You May X. If You Do, Y (`SVY-R14`)" (line 330): X is carried out on resolution as an untargeted selection, and Y applies only if X was carried out.
  - `FR-MOD-001` §6 (line 359): untargeted selections time out to the first legal options in fixed order, and "Cards in a pile such as the Discard pile: from the top down". The selection for X is optional and selects none on a timeout.
  - `FR-ZON-004` (line 202): a Scratch Company in a discard pile is face-up with its printed characteristics, so its printed card type is what a filter reads.
- Survey row: Arkha, Sun's Fury (line 151). Rulings `SVY-R3`, `SVY-R14`, `SVY-R16`.
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Filter`, untargeted pick, `Step`) and Decisions (the fixed order).
- Board directive #59 (1) (this task): rank pile options top-down, and add a timeout test where the top card is not the oldest.
- Kernel today:
  - `stepAssets` handles `SubjectPick` through `pickAssets` (`effects.go`), which builds a `DecidePick` with `assetDecision` (`targeting.go`). `assetDecision` holds the bounds rule: `upTo` 0 is `Min = Max = 1`, `upTo` N is `Min 0, Max min(N, len)`.
  - `stepRelocate` moves only a death trigger's own card, through `relocateSelf` (`SubjectSelf` on a `Tracked` item, if that instance is still in its owner's discard pile).
  - `fixedOrderPicks` (`choice.go`) ranks options by `fixedOrderPositions` (`state.go`), which holds only battlefield assets; any other option ranks after them, in option order.
  - `putCard` puts a card on the deck's top at index 0, or at the end with `bottom`. The top of a discard pile is its last element. Instance and asset IDs come from one counter (`newID`), so an `Option.Ref` never means both.

### Shape (lead decisions, binding)
- **`ability.ZoneDiscard`**, appended after `ZoneMainTheater`, with the comment `// a discard pile, not the battlefield: the entities are cards in it (FR-TGT-001 Zone). Filter.Allegiance names whose pile: allied is the controller's`. The `Filter.Zone` comment becomes `// ZoneAny (zero) means any battlefield zone; ZoneDiscard means cards in a discard pile; else only assets in that zone`.
- **A card filter:** add `(g *Game) matchesCard(ctl PlayerID, src *Instance, f ability.Filter, inst *Instance) bool` in `targeting.go`, beside `matchesFilter`. It reads the card's printed characteristics:
  - `Allegiance` by the card's owner: allied is `ctl`'s own pile, opposing is the other's, any is both;
  - `Kinds` against `inst.Card.Type`, as `matchesFilter` maps types to kinds. Put that mapping in one place both use, since a second copy would be a duplicate (D2.0);
  - `NonScratch` by the Scratch Company subtype;
  - `Other` excludes `src`, the item's own card.
  - `Zone` must be `ZoneDiscard`. `matchesFilter` returns false for `ZoneDiscard`, so a pile filter never matches an asset.
- **One bounds rule for both decisions:** move the bounds rule out of `assetDecision` into one helper in `targeting.go`, for example `pickBounds(upTo, n int) (lo, hi int)`. Both `assetDecision` and the new pile builder use it.
- **The pile decision:** `(g *Game) pileDecision(p PlayerID, kind DecisionKind, prompt string, upTo int, ok func(*Instance) bool) (Decision, []*Instance)` in `targeting.go`. Options are cards in discard piles for which `ok` holds: `p`'s pile first, then the opponent's, each in pile order from the bottom (the order of `Player.Discard`). `Label` is `inst.Card.Name`, `Ref` is `inst.ID`. A discard pile is public, so the decision is not `Hidden` and the fast path applies.
- **The pick:** `SubjectPick` with `Filter.Zone` `ZoneDiscard` asks one `DecidePick` each time its step resolves, built by `pileDecision` with `matchesCard`, with the same prompts as `pickAssets` (`"<card name>: pick one"`, `"<card name>: pick up to N"`). With no match nothing is asked and no decision event is emitted. The options are fixed before any card moves.
- **Card steps:** `OpReturnToHand` and `OpPutOnDeck` act on cards. Add `(g *Game) stepCards(r *resolution, st ability.Step) []*Instance` in `effects.go`:
  - `SubjectSelf`: the item's own card, under today's `relocateSelf` conditions (a `Tracked` item whose instance is still in its owner's discard pile), else none;
  - `SubjectPick` with `Filter.Zone` `ZoneDiscard`: the picked cards, in pick order;
  - anything else: none.

  `stepRelocate` moves each card from `stepCards` with `takeCard`, then `putCard` into its owner's hand, or into its owner's deck (bottom if `st.Bottom`). It returns how many cards moved. Fold `relocateSelf` into this, so there is one place that moves a card for a step.
- **Asset steps ignore pile picks:** `stepAssets` with a `ZoneDiscard` pick returns none and asks nothing.
- **Top-down timeout (#59 (1)):** `fixedOrderPositions` (`state.go`) also ranks cards in discard piles, after every battlefield asset: `p`'s pile from the top down, then the opponent's, from the top down. Update the "Battlefield orders" comment block above it to say so. `choice.go` is unchanged: `fixedOrderPicks` already ranks by that map. No M2 decision mixes assets and pile cards, so their relative order is unreachable and is not tested.
- **`doc.go`** is not changed: its line on `SubjectPick` stays true. The lead updates the design note's Key types and Decisions in the commit that writes this task. Engineers don't edit it.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written in `pkg/engine/pile_pick_test.go`, with synthetic cards only (Arkha is T-0078). The file won't compile until `ability.ZoneDiscard` exists. A convenient source is a unit with an attack or deploy trigger whose steps are under test. Required cases:
- **Options:** `TestFR_TGT_001_PilePickOffersMatchingCardsInYourDiscardPile`. With `Filter{Allegiance: Allied, Kinds: KindUnit, Zone: ZoneDiscard}`, the options are exactly the unit cards in the controller's discard pile, bottom first. A structure card in that pile, a unit card in the opponent's pile and an allied unit on the battlefield are not offered. The decision is a `DecidePick`.
- **On resolution (`SVY-R3`):** `TestFR_TGT_001_PilePickIsMadeOnResolutionFromWhatMatchesThen`. A unit card that reaches the discard pile during the same resolution, before the pick (for example killed by an earlier `OpKill` step), is offered.
- **Bottom of the deck:** `TestFR_TGT_001_PilePickPutsTheCardOnTheBottomOfItsOwnersDeck`. With `OpPutOnDeck` and `Bottom`, the picked card leaves the discard pile and is the deck's last card.
- **To hand:** `TestFR_TGT_001_PilePickCanReturnTheCardToHand`. With `OpReturnToHand`, the picked card goes to its owner's hand.
- **No match:** `TestFR_TGT_001_PilePickWithNoMatchDoesNothing`. Nothing is asked, and the ability still resolves.
- **Fast path (`FR-MOD-001` §1):** `TestFR_MOD_001_SinglePilePickWithOneMatchIsNotAsked`.
- **Top-down timeout (#59 (1), `FR-MOD-001` §6):** `TestFR_MOD_001_PilePickTimeoutPicksTheTopCard`. An exactly-one pick (`UpTo` 0) over three matching cards A, B and C, created in that order (so A has the lowest instance ID), lying in the discard pile bottom to top as A, C, B. A timeout picks B: the top card, which is neither the oldest card nor the first option.
- **If you do (`SVY-R14`):** `TestFR_TRG_005_IfYouDoRunsYAfterAPilePick` (X puts a card on the bottom of the deck, so Y deals its damage to the opposing Headquarters chosen on placement) and `TestFR_TRG_005_IfYouDoSkipsYWhenThePilePickPicksNothing` (an `UpTo` 1 pick with a timeout: no card moves and Y does nothing).
- **Existing tests:** every test in `pkg/engine` passes unchanged, including the `SubjectSelf` relocation tests and the pick tests from T-0072 and T-0076.
- **Left out:** a target in a discard pile (Eternal Watchman, a later task), opposing or both piles in one pick (no M2 card), and a decision that mixes assets and pile cards.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- `func (s *scene) inDiscard(p PlayerID, cards ...*Card) []*Instance`: puts a new instance of each card on top of p's discard pile, in order, and returns them. For the timeout test's order, the test may also move an instance with `s.g.takeCard` and `s.g.putCard`.

## Allowed Files
- `pkg/engine/ability/ability.go` (`ZoneDiscard` and the `Filter.Zone` comment)
- `pkg/engine/targeting.go` (`matchesCard`, the shared kind mapping, `pickBounds`, `pileDecision`, `matchesFilter` rejecting `ZoneDiscard`)
- `pkg/engine/effects.go` (the pile pick, `stepCards`, `stepRelocate`, `relocateSelf` folded in)
- `pkg/engine/state.go` (`fixedOrderPositions` and its comment)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_(TGT_001_PilePick|MOD_001_(SinglePilePick|PilePickTimeout)|TRG_005_IfYouDo(RunsYAfterAPilePick|SkipsYWhenThePilePick))'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] No card uses the new shapes yet, so the sims are unchanged: `go tool task sim-compare -- autoloop/m2-wave-b` passes

## Log
- 2026-10-27 lead-architect (wave run B, session 13): Wrote the task. Wave B cards still on `NotYetExecutable`: Arkha, Sun's Fury and Eternal Watchman (deck cards, directive #14 (a)) and Rogue Fleshshaper (kept). Arkha's X is T-0072's pick, reaching into a pile, so it goes next. Eternal Watchman needs a combat-kill hook and a target in a pile, and builds on this task's card filter and pile decision. The Arkha card would make six non-test files, so it is split off as T-0078 (`queued`).
  - Directive #59 (1) fired: this is the first task whose claims include a pick from a pile. It is applied here (Shape, last bullet but one, and the timeout test).
  - Waiting triggers checked against `autoloop/m2-wave-b` 37723a3: #6 and #7 (mode is `serial`), #8 (no `deck.go` change), #15 (no turn hook), #20 (b) (no wave C shape task), #58 (no exhaust by a cost or an effect) and #59 (2) (no repeated targets). None has fired. No `open: before` directive is on the board.
  - No existing equivalent: `ability` has no pile zone, and no step moves a card other than its own source. Two schemes would be copied by a naive build, and the Shape gives each one home: the bounds rule (`pickBounds`) and the type-to-kind mapping. `stepDiscard`'s hand options look like the pile options but are a different rule (`DecideHand`, most recent first), so they stay.
  - Size: four non-test files. `doc.go` needs no change, and `choice.go` none, because the top-down order lives in `fixedOrderPositions`.
  - Claims don't overlap any open task, since none is open. Status `tests`; dispatching `rules-test-author`.
- 2026-10-27 rules-test-author: Wrote `pkg/engine/pile_pick_test.go` (commit `test(T-0077)`), 9 tests: `TestFR_TGT_001_PilePickOffersMatchingCardsInYourDiscardPile`, `TestFR_TGT_001_PilePickIsMadeOnResolutionFromWhatMatchesThen`, `TestFR_TGT_001_PilePickPutsTheCardOnTheBottomOfItsOwnersDeck`, `TestFR_TGT_001_PilePickCanReturnTheCardToHand`, `TestFR_TGT_001_PilePickWithNoMatchDoesNothing`, `TestFR_MOD_001_SinglePilePickWithOneMatchIsNotAsked`, `TestFR_MOD_001_PilePickTimeoutPicksTheTopCard`, `TestFR_TRG_005_IfYouDoRunsYAfterAPilePick`, `TestFR_TRG_005_IfYouDoSkipsYWhenThePilePickPicksNothing`.
  - Helper requested: `inDiscard` (see Helpers Needed). `go vet` fails only on it and on `ability.ZoneDiscard`. `cardtown lint rules` passes.
  - Test-local helpers (`ppScript`, `ppPick` and others) live in the test file. The timeout test reorders the pile with `takeCard` and `putCard`, as the task allows.
  - Left out, as the task lists: a target in a pile, opposing or both piles, mixed asset and pile decisions.

- 2026-10-27 lead-architect (tests-review, round 1): **Approved, `ready`.** I checked the tests against FRD 01 `FR-TGT-001` (Zone, line 322; untargeted selections and "you may X", lines 329-330), `FR-MOD-001` §1 and §6 (line 359) and `SVY-R14`, `SVY-R16`.
  - **Expected outcomes are right:**
    - The options are exactly the unit cards in your own discard pile, bottom first, asked as a `DecidePick`. A structure card, the opponent's unit card and a battlefield unit are left out.
    - A unit card killed earlier in the same resolution is offered.
    - `PutOnDeck` with `Bottom` makes the card the deck's last, and `ReturnToHand` puts it in hand. The card not picked stays.
    - With no match nothing is asked, no decision event is emitted, and the next step still runs.
    - One match is not asked. The timeout picks the top card B from A, C, B, which is neither the oldest nor the first option (directive #59 (1)).
    - Y runs after a done pile pick and is skipped after an "up to" timeout.
  - **Within scope:** the helpers are local to the file or already exist (`iyd*`, `dmgTo`, `inDiscardPile`, `inHand`, `faceDownCount`), plus the requested `inDiscard`. No earlier test is modified.
  - Next: dispatch `kernel-engineer` (tier `sonnet`). The Shape section's lead decisions are binding.
