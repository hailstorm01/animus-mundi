---
id: T-0318
title: Make Hellfire Armorers executable (Ward; allied units in the same zone have Aggressive 1; wave A, card, flash)
kind: card
tier: flash
wave: A
status: review
branch: m3/autoloop-wave-a/T-0318-hellfire-armorers
depends_on: [T-0317]
claims: [card:Hellfire Armorers]
cards: [Hellfire Armorers]
requirements: [FR-LAY-001, FR-TGT-001, FR-KWD-002f, FR-KWD-002h, FR-KWD-002m, FR-CBT-001, FR-CBT-002, FR-LAY-003, FR-ZON-004]
---

## Goal
*Hellfire Armorers* (Soulless Infantry unit, R2 T2, power 3, health 4, deploys to Reserves, rare; two copies in `data/decks/soulless-starter.txt`): *"Ward"* / *"Allied units in the same zone have Aggressive 1."* It only combines shapes that are `done`: the printed *Ward* line, which the loader already parses onto `Card.Keywords`, and a `StaticAura` keyword grant with a `ZoneRefSame` filter, the shape of *Hulking Incongruity* (`pkg/engine/cards/animus.go`, Resistant 1 to allied Horror units in the same zone) with no subtype criterion and `Aggressive: 1` in place of `Resistant: 1`. "Allied units" has no "other", so it includes itself (survey row). `fireAttacks` (`pkg/engine/trigger.go`) reads `g.Keywords(a).Aggressive` when the attack happens, so a granted Aggressive triggers like a printed one.

Covers:
- The definition in `soullessDefs` (`pkg/engine/cards/soulless.go`), as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/units-and-structures-soulless.csv` line 7): Soulless Infantry unit, Requisition cost 2, Tactics cost 2, power 3, health 4, deploys to Reserves; rules text "Ward" then "Allied units in the same zone have Aggressive 1.".
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Soulless: deck" (*Hellfire Armorers*, line 175: "Built. Includes itself").
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): `FR-LAY-001` (line 528: continuous auras evaluate dynamically; One-Shot Modifier Snapshot Locking for the Aggressive bonus once resolved), `FR-TGT-001` (line 407; line 411: "in the same zone", `SVY-R44`: a continuous ability uses its source's current zone), `FR-KWD-002f` Aggressive N (line 646), `FR-KWD-002h` Ward (line 648), `FR-KWD-002m` (a face-down Scratch Company is a unit), `FR-CBT-001` (line 598: only Main Theater units attack), `FR-CBT-002` (line 613: Step 2 triggers resolve before Step 3 damage), `FR-LAY-003` (line 540: until end of turn), `FR-ZON-004` (line 278: an asset that leaves the battlefield stops applying its aura).
- Rulings that apply: `SVY-R44`.
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `soullessDefs`, after *Fire for Effect*)
- `"Hellfire Armorers"`, `Text`: one string, `"Allied units in the same zone have Aggressive 1."` (the "Ward" line is a keyword line, which the loader claims itself; it is not in `Text`).
- `Statics`: one `{Kind: ability.StaticAura, Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, ZoneRef: ability.ZoneRefSame}, Keywords: ability.Keywords{Aggressive: 1}}`.

A lead probe with the definition on the run branch at b6167f50, since reverted, passed `go tool task check`, `go tool task sim` and `go tool task sim-matrix` (exit 0). A probe test, also reverted, showed: the Armorers itself has Ward and Aggressive 1; an allied Main Theater unit and an allied face-down Scratch Company moved there have Aggressive 1; an allied Main Theater structure, an allied Reserves unit and an opposing Main Theater unit have 0; an allied power 2 unit attacking deals 3; with a second Armorers in the zone the unit has Aggressive 2.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_soulless_hellfire_test.go`, from the CSV text, loading the real card (`s.card("Hellfire Armorers")`), in the style of `cards_animus_auras_test.go` (*Hulking Incongruity*: same-zone keyword aura, `FollowsItsZone`) and `aggressive_item_test.go` (an Aggressive attack: `s.do(s.attack(att, def))`, `s.wantDamage`, `recordPlaced`). Player 0 controls the Armorers and is the active player; player 1 is the opponent. Read keywords with `s.g.Keywords(a)` and power with `s.g.Power(a)` or `s.wantPower`. Reuse the existing test helpers (`s.put`, `s.putScratch`, `s.moveTo`, `s.attack`, `s.do`, `s.endTurn`, `s.wantDamage`, `s.wantPower`, `s.script`, `s.alive`, `recordPlaced`, `stormKill`, `wantOptions`). The other cards may be synthetic. Every defender has power 0 and enough health to survive.

- `TestCard_HellfireArmorers_AlliedUnitsInItsZone` (`FR-LAY-001`, `FR-TGT-001` `SVY-R44`, `FR-KWD-002m`): with the Armorers in player 0's Main Theater, an allied Main Theater unit, an allied face-down Scratch Company moved to the Main Theater, and the Armorers itself have Aggressive 1. An allied Main Theater structure, an allied Reserves unit and an opposing Main Theater unit have Aggressive 0. Every one keeps its printed power (the grant is a keyword, not power).
- `TestCard_HellfireArmorers_HasWard` (`FR-KWD-002h`): the loaded Armorers has Ward. A targeted step of player 0's (a real card or a synthetic one) offers player 0's own Armorers and does not offer player 1's.
- `TestCard_HellfireArmorers_AlliedAttackerGetsOneMorePower` (`FR-KWD-002f`, `FR-CBT-002`, `FR-LAY-003`): an allied Main Theater unit with power 2 attacks an opposing Main Theater unit: one Aggressive item is placed (`FromKeyword` is `ability.KeywordTriggerAggressive`), and the defender takes 3. After `s.endTurn()` the attacker's power is 2.
- `TestCard_HellfireArmorers_ItselfAttacksWithOneMore` (`FR-KWD-002f`, survey "Includes itself"): the Armorers (power 3), moved to the Main Theater, attacks an opposing Main Theater unit: the defender takes 4.
- `TestCard_HellfireArmorers_AddsToPrintedAggressive` (`FR-KWD-002f`, `FR-LAY-001`): an allied Main Theater unit with power 1 and printed Aggressive 2 has Aggressive 3, and attacking deals 4.
- `TestCard_HellfireArmorers_TwoArmorersAddUp` (`FR-LAY-001`, `FR-KWD-002f`): with two Armorers in the Main Theater, an allied power 2 unit attacking deals 4. Assert only the damage, not the number of Aggressive items: whether two Aggressive 1 grants make one Aggressive 2 trigger or two Aggressive 1 triggers is not pinned here.
- `TestCard_HellfireArmorers_FollowsItsZone` (`FR-TGT-001` `SVY-R44`): with the Armorers in the Main Theater, an allied Main Theater unit and an allied Reserves unit; after the Armorers moves to Reserves (`s.moveTo`), the Reserves unit has Aggressive 1, the Main Theater unit 0, and the Armorers still 1. An opposing Reserves unit has 0.
- `TestCard_HellfireArmorers_GoneBeforeTheAttackGrantsNothing` (`FR-ZON-004`, `FR-KWD-002f`): the Armorers leaves the battlefield (destroyed, with an existing helper) before an allied Main Theater power 2 unit attacks: the unit has Aggressive 0, no Aggressive item is placed, and the defender takes 2.
- `TestCard_HellfireArmorers_BonusOutlastsTheAura` (`FR-KWD-002f`, `FR-LAY-001` One-Shot Modifier Snapshot Locking): an allied Main Theater power 2 unit attacks and its Aggressive item resolves (defender takes 3); the Armorers then leaves the battlefield the same turn: the attacker's power is still 3 until `s.endTurn()`, then 2.
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/soulless.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [x] The card tests pass: `go test ./pkg/engine -run TestCard_HellfireArmorers`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [x] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-26 lead-architect (wave A run, session 21): written against b6167f50, with the lead's probe above. `main` has no commits the run branch lacks. Board read (D2.0): no directive is due here (#185 has only (4)'s *Awaken Nightmare* half left, wave F), and no waiting trigger has fired: the card has no each-player effect (#71), no cost on a mode (#91 (2)), no deck search (#106), no discard cost (#120), no tuck (#142), no pile round trip (#183 (2)) and no turn-boundary or draw hook (#15 (2): Aggressive's "until end of turn" is a power bonus that expires at the existing end-of-turn expiry and can make nothing lethal), the mode is still `serial` (#6, #7), and no `needsRework` row exists (#184). Taken next in the survey's Soulless deck order, after *Fire for Effect*. Searched `pkg/engine` and the task archives: every part of the Definition exists (*Hulking Incongruity*'s same-zone `StaticAura` keyword grant, the loader's Ward keyword line, and `fireAttacks` reading granted Aggressive), so no kernel change. Set to `tests`.
- 2026-11-26 rules-test-author: wrote `pkg/engine/cards_soulless_hellfire_test.go`: TestCard_HellfireArmorers_AlliedUnitsInItsZone, _HasWard, _AlliedAttackerGetsOneMorePower, _ItselfAttacksWithOneMore, _AddsToPrintedAggressive, _TwoArmorersAddUp, _FollowsItsZone, _GoneBeforeTheAttackGrantsNothing, _BonusOutlastsTheAura. Helpers requested: none (reuses stormChoose, crDeployCard, crOneQuestion, crWantOffered, recordPlaced, stormKill). HasWard offers the Armorers to Conscript Rifles' chosen target. `go vet` clean, `cardtown lint rules` passes. Today 7 fail on the missing aura and 2 (HasWard, GoneBeforeTheAttack) already pass, as expected. No cases left out. Set to `tests-review`.
- 2026-11-26 lead-architect: tests approved (D4, round 1). 5d07b6a5 adds only `pkg/engine/cards_soulless_hellfire_test.go` and the task file's Log, 9 tests, one per case. With the lead's probe Definition all 9 pass. Each of 7 mutants fails 1 to 7 of them: `Other: true` (excludes itself), no `Kinds` (structures gain it), Aggressive 2, no allegiance, no zone criterion, a fixed Main Theater zone in place of `ZoneRefSame`, and +1 power in place of the keyword. The expected outcomes match the FRD 01 text they cite: `BonusOutlastsTheAura` pins line 532 (a resolved one-shot bonus is kept), `FollowsItsZone` line 411 (`SVY-R44`, the source's current zone), and `HasWard` line 648 (*Conscript Rifles*' any-allegiance slot offers the allied Armorers, not the opposing one). `TwoArmorersAddUp` scripts a timeout for a possible ordering question and asserts only the damage, as the task asks. The test author's report (not its Log) names one read of non-test engine files whose output was discarded (`head -0`); the tests use nothing from it. Set to `ready`.
- 2026-11-26 kernel-engineer (Gemini Flash): implemented Hellfire Armorers in pkg/engine/cards/soulless.go after Fire for Effect, matching the Definition (Text: ["Allied units in the same zone have Aggressive 1."]; Statics: one StaticAura with Filter {Allegiance: AllegianceAllied, Kinds: KindUnit, ZoneRef: ZoneRefSame} and Keywords: {Aggressive: 1}). Removed Hellfire Armorers from NotYetExecutable in pkg/engine/cards/cards.go. Reused existing StaticAura shape, ZoneRefSame, and engine Aggressive aura evaluation. Ran go test ./pkg/engine -run TestCard_HellfireArmorers (all 9 passed), go test ./pkg/cardcheck -run TestCards_Executable (passed), go tool task check (gofmt, vet, all tests, cardtown lint --all passed), go tool task sim (200 starter deck games passed with all invariants holding), and go tool task sim-matrix (24 starter + 24 stress runs passed, all invariants holding, exit 0). Set to `review`.
