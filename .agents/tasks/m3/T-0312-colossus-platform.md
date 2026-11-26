---
id: T-0312
title: Make Colossus Platform executable (Headquarters attack trigger, 1 armor to a chosen allied unit; wave A, card, flash)
kind: card
tier: flash
wave: A
status: tests-review
branch: m3/autoloop-wave-a/T-0312-colossus-platform
depends_on: [T-0311]
claims: [card:Colossus Platform]
cards: [Colossus Platform]
requirements: [FR-CAP-001, FR-CBT-002, FR-TGT-001, FR-TGT-003, FR-KWD-002h, FR-KWD-002k, FR-REP-004, FR-LAY-002, FR-ZON-005]
---

## Goal
*Colossus Platform* (the Soulless Headquarters, power 3, health 20; the Headquarters of the Soulless starter deck, `data/decks/soulless-starter.txt`): *"Whenever this Headquarters attacks, chosen allied unit gains 1 armor."* It only combines shapes that are `done`: a `HookAttacks` self trigger (*Ironclad Colossus*' second trigger; `fireAttacks` fires for any attacker, a Headquarters included, `pkg/engine/combat.go:35`) with one target slot whose filter is `Allegiance: ability.AllegianceAllied` and `Kinds: ability.KindUnit`, and one `OpAddArmor` step on that target (*Ironclad Colossus*' step with 1 in place of 2). The slot has no `Scratch` criterion, so an allied Scratch Company may be chosen. A Headquarters is not a unit, so *Colossus Platform* can't choose itself.

Covers:
- The definition in `soullessDefs` (`pkg/engine/cards/soulless.go`), as under Definition below. It goes in the faction's file, not `headquarters.go`: the design note's Boundaries row for `pkg/engine/cards` gives `concordia.go` and `soulless.go` the factions' cards, and `headquarters.go` only *Rift* and *Encampment* (the Animus and Path Headquarters).
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/units-and-structures-HQ.csv` line 3): Soulless Headquarters, power 3, health 20, Headquarters Zone.
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Soulless: deck" (*Colossus Platform (HQ)*, line 168).
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): `FR-CAP-001` (line 356; Attack, Headquarters included, line 365), `FR-CBT-002` (line 613; Step 2 places the attacker's triggers above the Attack action), `FR-TGT-001` (line 407), `FR-TGT-003` (lines 435-442: a targeted clause with no legal target fizzles, the rest resolves), the Single-Option Fast-Path (line 449), `FR-KWD-002h` Ward (line 648), `FR-KWD-002k` Armor (line 651: never expires, absorbs damage 1-for-1), `FR-REP-004` (line 562; Stage 4 Armor at line 587), `FR-LAY-002` (line 534), `FR-ZON-005` (line 287).
- Rulings that apply: `SVY-R3` (targets chosen on placement), `SVY-R25` (a removed target is no longer that target).
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `soullessDefs`)
- `"Colossus Platform"`, `Text`: `"Whenever this Headquarters attacks, chosen allied unit gains 1 armor."` (one string).
- One `Trigger`: `Hook: ability.HookAttacks`. `Effect`:
  - `Targets`: one slot, `Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}`.
  - `Steps`: `{Op: ability.OpAddArmor, Amount: ability.Value{N: 1}, Subject: ability.SubjectTarget, Target: 0}`.

A lead probe with the definition on the run branch at 6eaa2afb, since reverted, passed `go tool task check`, `go tool task sim` (P0 58.5% / P1 39.5% / draws 2.0%, unchanged) and `go tool task sim-matrix`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_soulless_colossus_test.go`, from the CSV text, loading the real card through `hqScene(t, "Colossus Platform", watchman)` (`cards_headquarters_test.go`), in the style of `cards_headquarters_watchman_test.go`. Player 0 controls *Colossus Platform*, is the active player, and attacks with it (`s.attack(s.g.Players[0].HQ, def)`), by default at player 1's *Eternal Watchman* (power 3), whose own trigger fires only when it attacks. Reuse the existing test helpers (`s.put`, `s.putScratch`, `s.script`, `s.asked`, `s.alive`, `s.do`, `s.endTurn`, `s.wantDamage`, `s.wantSettled`, `s.ptPlaceHook`, `stormKill`, `targetAsk`, `wantOptions`, `unit`, `structure`). Read armor from `Asset.Armor`. The other cards may be synthetic.

Remember the Single-Option Fast-Path (line 449): when exactly one allied unit is legal, nothing is asked.

- `TestCard_ColossusPlatform_IsAThreeTwentyHeadquarters` (`FR-CAP-001`): at the start of the match it is player 0's Headquarters in the Headquarters Zone, with power 3 and Maximum Health 20.
- `TestCard_ColossusPlatform_AttackGivesTheChosenUnitOneArmor` (`FR-CBT-002`, `FR-TGT-001`, `FR-KWD-002k`): with an allied unit in player 0's Reserves and another in the Main Theater, attacking asks one target decision that offers exactly those two. Choosing the Reserves unit gives it Armor 1, and the other has Armor 0. The attack still resolves: the *Eternal Watchman* is dealt 3 damage and *Colossus Platform* 3 in retaliation.
- `TestCard_ColossusPlatform_OffersOnlyAlliedUnits` (`FR-TGT-001`, `FR-ZON-005`, `FR-KWD-002h`): with an allied unit in each of the Reserves and the Main Theater, an allied Warded unit, an allied structure, an allied face-down Scratch Company (`s.putScratch`), an opposing unit in each zone, an opposing structure and an opposing Scratch Company, the decision offers exactly the three allied units and the allied Scratch Company. No structure, Headquarters or opposing asset is offered (Ward stops only opponents).
- `TestCard_ColossusPlatform_OneAlliedUnitTakesItWithoutAsking` (Single-Option Fast-Path): with one allied unit and one opposing unit, nothing is asked of player 0, and the allied unit has Armor 1. The opposing unit has none.
- `TestCard_ColossusPlatform_NoAlliedUnitStillAttacks` (`FR-TGT-003`, Deployment Triggers Without Legal Targets): with no allied unit (an opposing unit only), nothing is asked of player 0, no asset has armor, and the attack resolves (the *Eternal Watchman* is dealt 3 damage). Assert only these outcomes, not whether a trigger was placed: FRD 01 line 442 says only that the clause fizzles.
- `TestCard_ColossusPlatform_ArmorStacksAndLasts` (`FR-KWD-002k`, `FR-LAY-002`): a chosen allied unit with printed Armor 1 (`ability.Keywords{Armor: 1}`) has Armor 2 after the attack, and still 2 after player 0 ends the turn and player 1 ends the next.
- `TestCard_ColossusPlatform_ArmorAbsorbsLaterDamage` (`FR-REP-004` Stage 4): after the attack gives an allied Main Theater unit (health 3) Armor 1, player 0 ends the turn, and an opposing Main Theater unit of power 1 attacks it on player 1's turn. Its Armor is 0 and its Health is undamaged.
- `TestCard_ColossusPlatform_TargetGoneGivesNothing` (`FR-TGT-003`, `SVY-R25`): with two allied units, the chosen one is removed between the trigger's placement and its resolution (`s.ptPlaceHook`, which fires for an ability on the stack, with `stormKill`). The other has Armor 0, and the attack still resolves (the *Eternal Watchman* is dealt 3 damage).
- `TestCard_ColossusPlatform_OnlyItsOwnAttackTriggers` (`FR-CBT-002`): an allied Main Theater unit attacking the *Eternal Watchman* asks nothing and gives no asset armor. On player 1's turn, an opposing Main Theater unit attacking *Colossus Platform* asks player 0 nothing and gives no asset armor.
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None. The tests use only existing helpers (`hqScene`, `scriptRefs`, `wantOptions`, `stormKill`, `ptPlaceHook`) and file-local helpers `cpScene`, `cpAttack`, `cpWantResolved`, `cpWantNoArmor` in the new test file.

## Allowed Files
- `pkg/engine/cards/soulless.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_ColossusPlatform`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [ ] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-25 lead-architect (wave A run, session 14): written against 6eaa2afb, with the lead's probe above. `main` has no commits the run branch lacks. Board read (D2.0): no directive is due here (#185 has only (4)'s *Awaken Nightmare* half left, wave F), and no waiting trigger has fired. Taken next in the survey's Soulless deck order: session 13 went on to *Conscript Rifles* without writing down why it passed over this card, and nothing on the board or in the survey holds it back. Searched `pkg/engine` and the task archives: every part of the Definition exists (*Ironclad Colossus*' `HookAttacks` trigger and `OpAddArmor` step, *Eternal Watchman*'s Headquarters self trigger, the `AllegianceAllied` and `KindUnit` filter), so no kernel change. Set to `tests`.
- 2026-11-25 rules-test-author: wrote `pkg/engine/cards_soulless_colossus_test.go` (gofmt and `go vet` clean, `cardtown lint rules` clean). Tests: `TestCard_ColossusPlatform_IsAThreeTwentyHeadquarters`, `_AttackGivesTheChosenUnitOneArmor`, `_OffersOnlyAlliedUnits`, `_OneAlliedUnitTakesItWithoutAsking`, `_NoAlliedUnitStillAttacks`, `_ArmorStacksAndLasts`, `_ArmorAbsorbsLaterDamage`, `_TargetGoneGivesNothing`, `_OnlyItsOwnAttackTriggers`. Helpers requested: none. Cases left out: none. Today the first test passes (the card loads) and the armor tests fail until the definition is added; the no-armor tests pass trivially before it exists.
