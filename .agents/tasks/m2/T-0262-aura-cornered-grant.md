---
id: T-0262
title: "Cornered-gated aura grants: StaticAura's Static.Cornered, summed into the one Cornered gate g.whileCornered through the one aura reader g.auras (wave H shape, sonnet)"
kind: shape
tier: sonnet
wave: H
status: tests-review
branch: m2/autoloop-wave-h/T-0262-aura-cornered-grant
depends_on: [T-0258, T-0260]
claims: [shape:StaticAura Cornered grant (Static.Cornered: every asset an aura's Filter matches has the aura's WhileCornered while that asset's controller is Cornered, read through g.whileCornered and g.auras)]
cards: []
requirements: [FR-LAY-001, FR-LAY-003, FR-KWD-002b, FR-KWD-002e, FR-REP-004, FR-ZON-005, SVY-R18, SVY-R44]
---

## Goal
An aura can grant a Cornered ability: *"Other allied assets in the same zone have Cornered: Resistant 1."* (*Ancestral Monument*). T-0256's `ability.StaticAura` gains a `Cornered ability.WhileCornered`, the same type as a card's own `Def.Cornered` (T-0260): every battlefield asset the aura's `Filter` matches, relative to its source's controller, has that `WhileCornered` for as long as the aura applies, and it counts only while **that asset's** controller is Cornered (`FR-KWD-002e`, `SVY-R18`: "an asset with *Cornered: [Keyword].* has that keyword only while its controller is Cornered"). The keywords are a Layer 3 grant and the power Layer 4, exactly as for a card's own Cornered abilities (`FR-LAY-001`, survey G2).

There is one Cornered gate and one aura reader, and this task joins them: `Game.whileCornered(a)` becomes the sum of a's own face-up `Def.Cornered` and the `Cornered` of each aura `g.auras(a)` returns, gated once by `g.cornered(a.Controller)`. `Game.Keywords` and `Game.Power` already add `whileCornered(a)`, so damage, combat, filters, the sweep, the snapshot and the view need no change. A second Cornered check, or a second walk over battlefield auras, would be a duplication finding (D2.0).

*Ancestral Monument* itself is T-0263, queued (split off under the D2 size rule: the card adds `cards/path.go` and `cards/cards.go`, five non-test files with this task's three).

## Inputs
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): the pipeline (`:476-498`; Layer 3 "Dynamic Keyword Grants & Suppressions", Layer 4 "Continuous Auras"); `FR-LAY-001` (`:500-504`); `FR-LAY-003` (`:512-517`): Cornered is live, and locked at the start of a simultaneous damage step; `FR-KWD-002e` (`:617`): "Cornered: [Ability]." is active for the asset's controller, and `SVY-R18`; `FR-KWD-002b` (`:614`) and `FR-REP-004` Stage 2 (`:557`): total Resistant reduces each instance of damage; `FR-ZON-005`: a face-down asset has no printed abilities.
- Survey (`docs/design/m2_starter_set_ability_survey.md`): wave E (`:33`, "granted Cornered abilities"), G2 (`:48`), R18 (`:77`), *Ancestral Monument*'s row (`:205`).
- Design note `docs/architecture/m2-ability-model.md`: `Static` (`:43`: "An aura that grants a Cornered keyword (*Ancestral Monument*) is not built yet"; "`Game.auras` is its one reader, and the keyword and Cornered auras extend it"), Layers (`:55`), Deliverable 6m's wave E bullet (`:209`: one aura shape, filtering "by subtype (Horror) or by the Cornered gate").
- Code of the day (run branch at 815a54b):
  - `pkg/engine/ability/ability.go`: `StaticAura` (`:56-61`), `Static` (`:64-85`: `Filter`, `Power`, `Resistant`), `WhileCornered` (`:95-104`).
  - `pkg/engine/layers.go`: the file comment (`:1-6`), `Game.Power` (`:57-63`: adds `g.whileCornered(a).Power` and each aura's `Power`), `Game.auras` (`:65-87`), `Game.Keywords` (`:105-124`: adds `whileCornered(a)`'s Aggressive and Resistant, then each aura's `Resistant`), `Game.whileCornered` (`:139-150`: the zero value for a face-down asset or when not Cornered, else `d.Cornered`), `Game.cornered` (`:152-157`).
  - `pkg/engine/doc.go`: "A card's own Cornered abilities" (`:64-67`), "Auras" (`:75-79`), "Not built yet" (`:90-92`, "Cornered-gated auras").
  - No recursion: `g.auras` tests with `matchesFilterAt`, which never reads keywords or power, and `g.cornered` counts Main Theater assets only, so `g.whileCornered` calling `g.auras` can't loop.
  - Searched `pkg/engine` (non-test) and the task archive for an aura with a Cornered gate: none. `Trigger.Cornered` is a trigger's gate, checked when its event occurs (`FR-KWD-002e`), a different rule, and stays as it is.

## Shape (the lead's plan)
- `pkg/engine/ability/ability.go`:
  - `Static` gains `Cornered WhileCornered` after `Resistant`: an aura's Cornered grant, which every asset `Filter` matches has only while that asset's controller is Cornered (`FR-KWD-002e`, `SVY-R18`); its keywords are a Layer 3 grant and its power Layer 4, as for `Def.Cornered`. The zero value grants nothing.
  - `StaticAura`'s comment adds that every matching asset also has `Cornered` while its controller is Cornered. `WhileCornered`'s comment: what a card has, or an aura grants, only while the asset's controller is Cornered; read through `g.whileCornered`. `Def.Cornered`'s comment stays.
- `pkg/engine/layers.go`:
  - `Game.whileCornered(a)`: the zero value unless `g.cornered(a.Controller)`; otherwise a's own `Def.Cornered` when its face has an Ability, plus the `Cornered` of each aura in `g.auras(a)`, field by field (Aggressive, Resistant, Power). A face-down asset has no own Cornered abilities but still receives an aura's. It stays the one reader of the gate, live (`FR-LAY-003`, as T-0260's Goal explains: combat reads before damage and deaths wait for the sweep). How the fields are summed is the engineer's choice (a small `add` method on `WhileCornered` in package `ability` is fine), as long as no second Cornered check or aura walk appears.
  - `Game.Keywords` and `Game.Power` need no new term: their comments say the Cornered part includes auras' Cornered grants.
  - The file comment: Layer 3 also holds auras' Cornered keyword grants, and Layer 4 their Cornered power.
- `pkg/engine/doc.go`: the "Auras" bullet adds that an aura's `Cornered` grant counts while the matched asset's controller is Cornered, through `Game.whileCornered`. Under "Not built yet", drop "(Cornered-gated auras (\"have Cornered: Resistant 1.\"))", leaving the sentence well-formed.
- No change to `damage.go`, `combat.go`, `filter.go`, `sweep.go`, `snapshot.go` or `pkg/view`.
- Not this task: *Ancestral Monument* (T-0263); Cornered flag keywords (no Animus or Path aura grants one).

## Tests (shape tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/aura_cornered_test.go`. Cards are synthetic, as in `aura_keywords_test.go` (T-0258): a structure or unit with `c.Ability = &ability.Def{Statics: []ability.Static{{Kind: ability.StaticAura, Filter: f, Cornered: ability.WhileCornered{Resistant: 1}}}}`, through a local helper. "Other allied assets in the same zone" is `ability.Filter{Allegiance: ability.AllegianceAllied, Other: true, ZoneRef: ability.ZoneRefSame}` (zero `Kinds`: any asset). Reuse `wantResistant` (`aura_keywords_test.go`), `wantPower`, `corneredCard` (`cornered_keywords_test.go`) and `putScratch` unchanged. A player is Cornered when the opponent controls strictly more Main Theater assets (`FR-KWD-002e`); build it with extra enemy units in the Main Theater, as `cornered_power_test.go` does. Player 0 is the active player.

- `TestFR_KWD_002e_AuraCorneredGrantOnlyWhileCornered`: player 0's source and allied unit A in the Main Theater, player 1's unit in the Main Theater (2 against 1: not Cornered): A has Resistant 0. Add two more player 1 units there (2 against 3): A has Resistant 1. Move one of them to Reserves (2 against 2): A has 0 again. Throughout, the source has 0 (`Other`) and A keeps its printed power.
- `TestFR_LAY_001_AuraCorneredGrantMatchesFilter`: with player 0 Cornered and the source in the Main Theater: an allied unit and an allied structure there have Resistant 1; an allied unit in player 0's Reserves has 0 (another zone); player 1's units in the Main Theater have 0 (not allied). Then put player 1's own such source in the Main Theater with player 1 not Cornered: its allied unit has 0 (player 1's own controller is not Cornered, though player 0 is).
- `TestFR_LAY_001_AuraCorneredGrantAddsToOwnCornered`: player 0 Cornered; an allied unit in the source's zone printing Resistant 1 and its own `corneredCard(..., ability.WhileCornered{Resistant: 1})` has Resistant 3 (printed, its own Cornered, the aura's). With a second source in the zone it has 4. An aura granting `WhileCornered{Power: 2}` to that unit adds 2 to its power while Cornered and nothing while not (the power half goes through the same gate).
- `TestFR_ZON_005_AuraCorneredGrantFaceDown`: player 0 Cornered. A face-down allied Scratch Company (`s.putScratch`) in the source's zone has Resistant 1 (it receives the grant though its face has no Ability). A face-down source (put face down as `putScratch` builds one, with the aura card as its `Inst.Card`) gives an allied unit in its zone 0.
- `TestFR_LAY_003_AuraCorneredResistantLockedForCombat`: player 1's source and player 1's unit D (power 5, health 5) in the Main Theater; player 0's Readied attacker (power 3, health 1) and two more player 0 units there (3 against 2: player 1 Cornered). The attacker attacks D: D takes 2 (Stage 2 subtracts Resistant 1) and the damage event on D logs Resistant 1, although the attacker dies to D's damage in the same step. After the sweep (2 against 2) D has Resistant 0.

Existing tests stay as they are and must still pass.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none expected. Helpers local to the new file live in it.

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/layers.go`
- `pkg/engine/doc.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "AuraCornered"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (no starter or stress card prints an aura with `Cornered`, so every log is byte-identical)

## Log
- 2026-11-16 lead-architect (wave H run, session 23): written against the run branch at 815a54b. *Ancestral Monument* is next in Deliverable 6m's order (wave E, after T-0261's *Brigand Highlander*). No ruling needed: `FR-KWD-002e` and `SVY-R18` give a Cornered keyword to the asset only while its controller is Cornered, and "other allied assets" makes that controller the source's too; G2 puts the keyword grant in Layer 3; `SVY-R44` gives the aura its source's current zone. Reuse: `Static.Cornered` is the existing `WhileCornered`, and `g.whileCornered` the one gate, now summing `g.auras`, the one aura reader, so `Keywords`, `Power` and everything reading them need no change; `Trigger.Cornered` is a trigger gate, a different rule. Split from the card under the D2 size rule; T-0263, *Ancestral Monument*, queued in this commit. Three non-test files. Directives: none `open`; each `waiting` one was checked against the repo and none has fired (#6, #7: mode is `serial`; #15 (2): no draw or end-of-turn hook; #71: no each-player effect; #91 (2), #106, #120, #142 name cards or shapes no task touches yet). `main` has no commits the run branch lacks. Status `tests`.
- 2026-11-16 rules-test-author: tests written in `pkg/engine/aura_cornered_test.go`: `TestFR_KWD_002e_AuraCorneredGrantOnlyWhileCornered`, `TestFR_LAY_001_AuraCorneredGrantMatchesFilter`, `TestFR_LAY_001_AuraCorneredGrantAddsToOwnCornered` (with a `power` subtest), `TestFR_ZON_005_AuraCorneredGrantFaceDown` (with a `face-down source` subtest), `TestFR_LAY_003_AuraCorneredResistantLockedForCombat`. Helpers local to the file (`corneredAura`, `monument`, `(*scene).foes`); none requested for scenario_test.go. The file fails to compile only on the missing `ability.Static.Cornered`. `cardtown lint rules` passes. Nothing left out. Status `tests-review`.
