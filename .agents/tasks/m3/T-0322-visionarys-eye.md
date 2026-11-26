---
id: T-0322
title: Make Visionary's Eye executable (whenever this unit attacks, draw a card, wave A, card, flash)
kind: card
tier: flash
wave: A
status: tests
branch: m3/autoloop-wave-a/T-0322-visionarys-eye
depends_on: [T-0321]
claims: [card:Visionary's Eye]
cards: [Visionary's Eye]
requirements: [FR-TRN-003, FR-SET-004, FR-CAP-001, FR-CBT-001, FR-CBT-002, FR-TRG-003, FR-TRG-004, FR-ZON-004, FR-WIN-001]
---

## Goal
*Visionary's Eye* (Soulless unit, Cybernetic, Scout, R1 T0, Reserves only, power 1, health 1; three copies in `data/decks/soulless-starter.txt`): *"Whenever this unit attacks, draw a card."* It only combines shapes that are `done`: the self `HookAttacks` trigger with no targets and no `Optional` (*Earthshaker Colossus*' second trigger and *Colossus Platform* in `pkg/engine/cards/soulless.go`, both placed in combat Step 2 by `fireHook`), with *Leyline Regrowth*'s `OpDraw` step, `Amount` 1, no `Subject` (the controller; `pkg/engine/cards/concordia.go`). The draw is mandatory, so nothing is asked. It deploys to Reserves only, so it attacks only after a March to the Main Theater (`FR-CAP-001`, `FR-CBT-001`), as *Earthshaker Colossus* does.

Covers:
- The definition in `soullessDefs` (`pkg/engine/cards/soulless.go`), as under Definition below.
- The name off `cards.NotYetExecutable`. It is the last wave A header card still on it.

## Inputs
- Card (`data/cards/units-and-structures-soulless.csv` line 5): Soulless unit, Cybernetic, Scout, Requisition cost 1, Tactics cost 0, deploys to Reserves only, power 1, health 1, Requisition yield 1, Tactics yield 2.
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Soulless: deck" (*Visionary's Eye*, line 179: "attack → draw 1", Built).
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): `FR-SET-004` (line 201: no summoning sickness), `FR-WIN-001` (line 218: a draw from an empty deck does nothing, and that player loses at the next sweep with no item resolving), `FR-ZON-004` (line 278: last-known state), `FR-CAP-001` (line 356: March, and Attack only from the Main Theater), `FR-TRN-003` (line 393), `FR-TRG-003` (line 488: APNAP, the non-active player's triggers resolve first), `FR-TRG-004` (line 494), `FR-CBT-001` (line 598: units in Reserves cannot attack; zero-power attacks fire attack triggers), `FR-CBT-002` (line 613: Step 2 attack triggers resolve before the Attack; Step 3's target check fizzles it).
- Rulings that apply: `SVY-R8` (the attack's targets are re-verified at resolution; a fizzled attack deals no damage, but its Step 2 triggers were already placed), `SVY-R49` (*Blind Defile*'s kill).
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `soullessDefs`, after *Mobile Armory*)
- `"Visionary's Eye"`, `Text`: `"Whenever this unit attacks, draw a card."` (one string, as the CSV has it).
- One `Triggers` entry: `Hook: ability.HookAttacks`, no `Optional`, no `Watches`, `Effect.Steps`: `{Op: ability.OpDraw, Amount: ability.Value{N: 1}}` (no `Subject`: its controller draws). No targets.

A lead probe with the definition on the run branch at 0085610c, since reverted, passed `go test ./pkg/cardcheck`, `go tool task check`, `go tool task sim` (P0 58.5% / P1 39.5% / draws 2.0%, unchanged) and `go tool task sim-matrix` (exit 0).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_soulless_eye_test.go`, from the CSV text, loading the real card (`s.card("Visionary's Eye")`), in the style of `cards_soulless_earthshaker_test.go` (a Soulless Reserves-only unit with an attack trigger, `ActMarch` to the Main Theater, `TestCard_EarthshakerColossus_EachAttackTriggersAgain` and `_DefendingDoesNotTrigger`), `cards_soulless_colossus_test.go` (`TestCard_ColossusPlatform_OnlyItsOwnAttackTriggers`) and `attack_trigger_test.go` (`TestFR_CBT_002_AttackTriggerResolvesBeforeCombatDamage`: an empty deck seen at the trigger's sweep). Player 0 controls the Eye and is the active player, unless a case says otherwise. Reuse the existing test helpers (`s.put`, `s.give`, `s.onTop`, `s.script`, `s.reject`, `s.do`, `s.attack`, `s.endTurn`, `s.alive`, `s.wantDamage`, `s.asked`, `s.card`, `inDiscardPile`, `assetOf`). The other cards may be synthetic, except *Blind Defile*, which is the real card. Count cards through the players' `Hand` and `Deck`, and name the drawn card with `s.onTop`.

Deploying:
- `TestCard_VisionarysEye_EntersReservesAsOneOne` (`FR-TRN-003`, `FR-CAP-001`): deployed from hand with exactly R1, it is a Readied 1/1 unit in player 0's Reserves, and the pool is empty. A deploy to the Main Theater is rejected (`s.reject`), and with R0 it can't be deployed: in both cases the card stays in hand.
- `TestCard_VisionarysEye_CannotAttackFromReserves` (`FR-CBT-001`, `FR-CAP-001`): a Readied Eye in player 0's Reserves can't attack player 1's Headquarters (`s.reject`): player 0's hand and deck are unchanged, and the Eye is still Readied.
- `TestCard_VisionarysEye_MarchesThenAttacksAndDraws` (`FR-SET-004`, `FR-CAP-001`): deployed from hand to Reserves, it may March the same turn (`ActMarch`) and is then in the Main Theater, Exhausted. After `s.endTurn()` twice, its attack on an opposing 0/30 unit in the Main Theater draws one card (beyond player 0's Draw Phase draw).

Attacking (a Readied Eye put in player 0's Main Theater, with a known card on top of player 0's deck, unless the case says otherwise):
- `TestCard_VisionarysEye_AttackDrawsOneCard` (`FR-CBT-002`): attacking player 1's Headquarters puts the top card of player 0's deck into player 0's hand: the hand has one card more, the deck one fewer, and no decision is asked of either player (the draw is mandatory, with no targets). Player 1's 3/20 Headquarters takes 1 damage, its 3 retaliation destroys the Eye (its card in player 0's discard pile), and the drawn card stays in hand. Player 1's hand and deck are unchanged.
- `TestCard_VisionarysEye_DrawResolvesBeforeCombatDamage` (`FR-CBT-002` Step 2, `FR-WIN-001`): with player 0's deck empty, attacking player 1's Headquarters makes player 0 lose at the trigger's sweep, before the Attack resolves: the match is over with player 1 the winner, not a draw, and player 1's Headquarters takes no combat damage.
- `TestCard_VisionarysEye_DrawsEvenWhenItDiesInTheTrade` (`FR-CBT-002`): attacking an opposing 1/1 unit in the Main Theater, both are destroyed, both cards are in their owners' discard piles, and player 0 still drew one card.
- `TestCard_VisionarysEye_ZeroPowerAttackStillDraws` (`FR-CBT-001` Zero-Power Attacks): with the Eye's power lowered to 0 or below for the turn (a synthetic Maneuver or trigger with `OpModifyStat` -1 on it, or any existing way the other tests use), its attack is still allowed and still draws one card.
- `TestCard_VisionarysEye_KilledBeforeItsTriggerStillDraws` (`FR-TRG-003`, `FR-ZON-004`, `FR-CBT-002` Step 3, `SVY-R49`, `SVY-R8`): with a real *Blind Defile* in player 1's Main Theater and player 1 scripted to accept it, the Eye's attack on player 1's Headquarters places both triggers; the Defile's (the non-active player's) resolves first and kills the Defile and the Eye. The Eye's trigger still resolves: player 0 draws one card. The Attack fizzles: player 1's Headquarters takes no damage, and the Eye's card is in player 0's discard pile.
- `TestCard_VisionarysEye_EachAttackDrawsAgain` (`FR-CBT-002`): it attacks an opposing 0/30 unit in the Main Theater on player 0's turn, `s.endTurn()` twice (player 0's own Draw Phase draw included), and it attacks that unit again: each attack draws exactly one card more than the turn's draws alone.
- `TestCard_VisionarysEye_DefendingDoesNotDraw` (`FR-CBT-002`): on player 1's turn, an opposing unit attacks the Eye in the Main Theater. Player 0's hand and deck are unchanged.
- `TestCard_VisionarysEye_OnlyItsOwnAttackDraws` (`FR-CBT-002`): with two Eyes and another allied unit in player 0's Main Theater, the other unit's attack draws nothing, and one Eye's attack draws exactly one card (the other Eye doesn't trigger).

- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` and `TestDefs_MatchRulesText` with the name removed from `NotYetExecutable`.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/soulless.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_VisionarysEye`
- [ ] `go test ./pkg/cardcheck` passes with the name removed
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)

## Log
- 2026-11-26 lead-architect (wave A run, session 25): written against 0085610c, with the lead's probe above. `main` has no commits the run branch lacks. Board read (D2.0): no directive is due here (#185 has only (4)'s *Awaken Nightmare* half left, wave F), and no waiting trigger has fired: the card has no each-player effect (#71: only its controller draws), no mode or "another" slot (#91), no deck search (#106), no cost at all (#120), no tuck (#142), no turn-start, turn-end or draw hook (#15 (2): it draws as an effect of an attack hook, and the draw expires nothing), and no pile (#183 (2)); the mode is still `serial` (#6, #7), and no `needsRework` row exists (#184). Taken next in the survey's Soulless deck order, after *Mobile Armory*; it is the last header card on `NotYetExecutable`. Searched `pkg/engine` and the task archives: every part of the Definition exists (`fireHook` places self `HookAttacks` triggers in Step 2, and `OpDraw` with no `Subject` draws for the controller), so no kernel change. Set to `tests`.
