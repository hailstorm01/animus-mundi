---
id: T-0271
title: "Animoceps Spore Cloud: one WouldBeDealtDamage replacement over allied units in the same zone (wave H card, flash)"
kind: card
tier: flash
wave: H
status: tests-review
branch: m2/autoloop-wave-h/T-0271-animoceps-spore-cloud
depends_on: [T-0269, T-0270]
claims: [card:Animoceps Spore Cloud]
cards: [Animoceps Spore Cloud]
requirements: [FR-REP-001, FR-REP-004, FR-CBT-002, FR-ZON-005, SVY-R23, SVY-R44]
---

## Goal
*Animoceps Spore Cloud* (Animus Structure, "Redirect all damage that would be dealt to allied units in the same zone to this structure.") is executable: one T-0269 `ability.Replacement{Event: ability.WouldBeDealtDamage, Filter: {Allegiance: AllegianceAllied, Kinds: KindUnit, ZoneRef: ZoneRefSame}}`, written as the same-zone aura cards write their filter (`cards/animus.go:13-30`, *Animoceps Node*). It leaves `cards.NotYetExecutable`, which is then empty: M2's card done-when (design note `:211`, wave H). `pkg/engine/doc.go` moves damage redirection from "Not built yet" to the built list and drops the now-empty `NotYetExecutable` sentence's "still in that state" wording as fits. No new shape.

Covers: the definition in `pkg/engine/cards/animus.go`; the removal from `pkg/engine/cards/cards.go` (the slice stays, empty, since `defs_test.go` and `.agents/scripts/waverun` read it); `pkg/engine/doc.go`. Card tests in a new `pkg/engine/cards_animus_spore_cloud_test.go`: printed stats; effect damage to an allied unit in the Main Theater lands on the Cloud; combat damage to an allied defender lands on the Cloud with retaliation unchanged; a unit in Reserves is not covered; enemy units are not covered; two Clouds ask the unit's controller (`SVY-R23`).

## Inputs
- Card text (`data/cards/units-and-structures-animus.csv:18`): Structure, Animus, Fungus, Requisition 4, Tactics 0, Main Theater, power 0, Health 9, yields 2/2, rare. It is in no starter or stress deck.
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`):
  - `FR-REP-004` (`:540`), Stage 1 (`:562`, naming the card): the damage moves to the new target before any defense, and Stages 2-5 use the **new target's** Resistant, Stalwart, Armor and Health.
  - `FR-REP-001` (`:529`): replacements apply inline, off the stack. `SVY-R23` (`:72`): when several apply, the affected asset's controller chooses.
  - `SVY-R44` (`:38`): a continuous ability's "in the same zone" is the source's current zone.
  - `FR-CBT-002` (`:591`): the attacker's hit and the defender's retaliation are two instances; only the one aimed at an allied unit of the Cloud's is redirected.
  - `FR-ZON-005`: a face-down Scratch Company has the Scratch Company face, not its hidden card's text.
- Survey (`docs/design/m2_starter_set_ability_survey.md`): its row (`:171`), wave G (`:35`).
- Code of the day (run branch at `4952e01`):
  - `ability.WouldBeDealtDamage` and `ability.Replacement` (`pkg/engine/ability/ability.go:170-186`).
  - `redirectDamage` (`pkg/engine/replace.go:140-170`): Stage 1 of `dealDamage`, once per instance; a source never redirects to itself; several sources ask the target's controller (`chooseReplacement`, `DecideReplacement`); an `EventReplace` from the source names the target.
  - `replacementsFor` (`pkg/engine/replace.go:37`) reads `Asset.Printed()`, which is the Scratch Company face for a face-down asset (`pkg/engine/state.go:85`), so a face-down Cloud redirects nothing.
  - *Animoceps Node* (`pkg/engine/cards/animus.go:13-26`): the same allied-units-in-the-same-zone filter, on an aura.
  - `NotYetExecutable` (`pkg/engine/cards/cards.go:13-15`): *Animoceps Spore Cloud* is its only entry.
  - `pkg/engine/doc.go:89-104`: the built list ends with destruction replacement; the not-built list still names damage redirection.
  - T-0269 pinned the shape on a synthetic redirector in `pkg/engine/redirect_test.go` (the new target's defenses, combat damage stays combat damage, no combat kill, current zone, filter, every instance, damage-taken triggers, two redirectors, kills not redirected). T-0270 made the sim damage monitor accept a redirected combat hit. This task pins the real card.

## Shape (the lead's plan)
- `pkg/engine/cards/animus.go`: one definition, between *Animoceps Shambler* and *Animoceps Spores* (`"Spore "` sorts before `"Spores"`). `Text` is the CSV line, as one string. `Replacements`: one `{Event: ability.WouldBeDealtDamage, Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, ZoneRef: ability.ZoneRefSame}}`, with no `Steps`.
- `pkg/engine/cards/cards.go`: remove the name, leaving `var NotYetExecutable = []string{}`. The slice stays: `defs_test.go` and `.agents/scripts/waverun` read it, and the runner parses an empty slice literal as no names.
- `pkg/engine/doc.go`:
  - After the "Destruction replacement" bullet (`:89-94`), add this bullet:
    ```
    //   - Damage redirection (FR-REP-004 Stage 1): a face-up asset's printed
    //     WouldBeDealtDamage replacement deals each damage instance that would be
    //     dealt to an asset its Filter matches to the source instead, before any
    //     defense, so the source's defenses apply; "in the same zone" is the
    //     source's current zone (SVY-R44). When several apply, the damaged asset's
    //     controller chooses (DecideReplacement, SVY-R23).
    ```
  - Replace the "Not built yet" bullet (`:98-103`) with:
    ```
    //   - Continuous abilities other than auras, a card's own Cornered abilities
    //     and its own hand-active cost reduction, Regroup N, replacement effects
    //     other than destruction replacement and damage redirection, and match
    //     clocks. Cards whose text needs any of these load with Card.Unsupported
    //     set and cannot be deployed or enacted. No in-scope M2 card is in that
    //     state, so cards.NotYetExecutable is empty.
    ```
- The lead dry-ran the definition and the removal on the run branch at `4952e01` (2026-11-18, wave H session 33), then reverted them:
  - `gofmt -l pkg cmd` clean.
  - `go test ./pkg/engine -run "TestDefs|Redirect|REP_004"` ok.
  - `go tool task check` exit 0.
  - `go tool task sim` exit 0.
  - `go tool task sim-compare -- autoloop/m2-wave-h` `[PASS]`.
- No new shape: the card is T-0269's `WouldBeDealtDamage` with *Animoceps Node*'s filter.

## Tests (card tasks: written by rules-test-author; engineers never edit these)
New tests go in a new file `pkg/engine/cards_animus_spore_cloud_test.go`, with a file comment naming the card, its CSV line and T-0271, as `cards_path_solarheart_monk_test.go` does.
- Use the real card (`s.card("Animoceps Spore Cloud")`). Player 0 is the active player.
- Effect damage: `s.enactWith(nil, allDamage(n, filter))` (one instance per matching asset). Combat: `s.do(s.attack(att, def))`. Replace events: `eventsOfKind(s.g.Log, EventReplace)`. A face-down Scratch Company: `s.putScratch(p)`, then set its `Inst.Card` and move it.
- Helpers from T-0269 and earlier may be reused. `redirect_test.go`'s `redirector` is synthetic and is not used here.

Tests:
- `TestCard_AnimocepsSporeCloud_PrintedStats`: an Animus Structure, Fungus, Requisition 4, Tactics 0, prints only the Main Theater, power 0, Health 9, yields 2/2, and `Card.Supported` accepts it.
- `TestCard_AnimocepsSporeCloud_EffectDamageToAlliedUnitLandsOnTheCloud` (`FR-REP-004` Stage 1): player 0's Cloud and an allied unit in player 0's Main Theater; a step deals 3 damage to allied units. The unit has 0 damage, the Cloud has 3, and exactly one replace event has the Cloud as source and the unit as target.
- `TestCard_AnimocepsSporeCloud_CombatHitOnAlliedDefenderLandsOnTheCloud` (`FR-CBT-002`, `FR-REP-004`): player 1's Cloud and defending unit (power 2, Health 5) in player 1's Main Theater; player 0's attacker (power 4, Health 10) attacks the unit. The Cloud has 4 damage, the defender 0, and the attacker has 2 from the defender's retaliation (the attacker is not the Cloud's ally).
- `TestCard_AnimocepsSporeCloud_UnitInAnotherZoneIsNotCovered` (`SVY-R44`): player 0's Cloud in the Main Theater and an allied unit in Reserves; a step deals 2 damage to allied units. The Reserves unit has 2, the Cloud 0, and no replace event is logged.
- `TestCard_AnimocepsSporeCloud_EnemyUnitsAreNotCovered` (`FR-REP-004` Stage 1): player 0's Cloud and player 1's unit, both in the Main Theater; player 0 deals 2 damage to enemy units. The enemy unit has 2, the Cloud 0, and no replace event is logged.
- `TestCard_AnimocepsSporeCloud_AlliedStructuresAreNotCovered` (`FR-REP-004` Stage 1): player 0's Cloud and another allied structure (Health 5) in the Main Theater; a step deals 2 damage to allied structures. The other structure has 2 and the Cloud has only its own 2; no replace event is logged.
- `TestCard_AnimocepsSporeCloud_TwoCloudsTheUnitsControllerChooses` (`FR-REP-001`, `SVY-R23`): two Clouds and a unit, all player 1's, in player 1's Main Theater; player 0's step deals 3 damage to enemy units. Player 1 (and not player 0) is asked one `DecideReplacement` whose options are the two Clouds; answer the second. The second Cloud has 3, the first 0, the unit 0.
- `TestCard_AnimocepsSporeCloud_FaceDownCloudRedirectsNothing` (`FR-ZON-005`, `FR-REP-004`): player 0's face-down Scratch Company whose hidden card is the real Cloud, moved to the Main Theater, beside an allied face-up unit there; a step deals 2 damage to the face-up unit only (a filter the Scratch Company does not match, such as a subtype or name, or a target pick). The unit has 2 and no replace event is logged.
- `TestDefs_MatchRulesText` already checks `Text` against the CSV.
- Not this task: Resistant, Stalwart and Armor on the new target, redirected lethal combat damage, damage-taken triggers, one instance per unit, a moved source, and kills. T-0269 pinned each on a synthetic redirector, and the card adds nothing to them.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none. The tests use existing helpers only; setup local to the new file stays in it.

## Allowed Files
- `pkg/engine/cards/animus.go`
- `pkg/engine/cards/cards.go`
- `pkg/engine/doc.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "AnimocepsSporeCloud|TestDefs"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (the card is in no starter or stress deck)

## Log
- 2026-11-18 lead-architect (wave H run, session 31): queued with T-0269, split off under the D2 size rule. Depends on T-0269 (the shape) and T-0270 (the sim monitor reads its redirected combat damage).
- 2026-11-18 lead-architect (wave H run, session 33): T-0269 (merge de3a84a) and T-0270 (merge 85a4ee7) are `done`.
  - Wrote Inputs, Shape, Tests and Done When against the run branch at `4952e01`, and dry-ran the definition (above). Beyond the Goal's list, a face-down Cloud and an allied structure are pinned as the card's own edges; `requirements:` gains `FR-ZON-005`.
  - `main` has no commits the run branch lacks, so no sync.
  - Directives: none is `open`. I checked each `waiting` one, and none has fired:
    - #6, #7: mode is `serial`.
    - #15 (2): no draw or end-of-turn hook here.
    - #71: no each-player effect.
    - #91 (2): no cost on a mode.
    - #106, #120, #142: no deck search with criteria, no mandatory discard cost, no tuck.
  - No new shape: T-0269's `WouldBeDealtDamage` with *Animoceps Node*'s filter.
  - Re-read `BOARD.md` before committing: no directive added since.
  - Status `tests`.
- 2026-11-18 rules-test-author: wrote `pkg/engine/cards_animus_spore_cloud_test.go`, the seven listed tests. Status `tests-review`.
  - Tests: `TestCard_AnimocepsSporeCloud_` `PrintedStats`, `EffectDamageToAlliedUnitLandsOnTheCloud`, `CombatHitOnAlliedDefenderLandsOnTheCloud`, `UnitInAnotherZoneIsNotCovered`, `EnemyUnitsAreNotCovered`, `AlliedStructuresAreNotCovered`, `TwoCloudsTheUnitsControllerChooses`, `FaceDownCloudRedirectsNothing`.
  - Helpers requested: none. `go vet` is clean; `cardtown lint rules` passes.
  - Today PrintedStats, the effect, the combat and the two-Clouds tests fail (the card is unsupported); the other four pass vacuously and guard the card once it is executable.
  - Left out: nothing beyond the task's "Not this task" list.
