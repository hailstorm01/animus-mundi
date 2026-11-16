---
id: T-0256
title: "Layer 4 power auras: StaticAura with a Filter and Power, read by g.Power through one aura reader; \"in the same zone\" is the source's current zone (wave H shape, sonnet)"
kind: shape
tier: sonnet
wave: H
status: tests
branch: m2/autoloop-wave-h/T-0256-same-zone-power-aura
depends_on: []
claims: [shape:StaticAura (a continuous Layer 4 aura of a face-up battlefield asset, Static.Filter relative to its controller and Static.Power; g.auras, the one reader of which auras apply to an asset; ZoneRefSame on an aura reads its source's current zone through zoneLock and inLocked)]
cards: []
requirements: [FR-LAY-001, FR-TGT-001, FR-ZON-001, FR-ZON-004, FR-CBT-002]
---

## Goal
A card can print a continuous aura: *"Other allied assets in the same zone have +1 power."* A new `ability.StaticAura` kind of `Static` gives every battlefield asset its `Filter` matches, relative to the aura source's controller, `+Power` Combat Power in Layer 4 (`FR-LAY-001`, Six-Layer pipeline, `:476-498`). The aura works while its source is a face-up asset on the battlefield, and is evaluated continuously: an asset that enters, leaves or changes zone gains or loses it at once, and so does the source (`FR-LAY-001`, last sentence of One-Shot Modifier Snapshot Locking, `:504`). "In the same zone" on an aura is its source's **current** zone, not a locked one (`FR-TGT-001` `SVY-R44`, last sentence, `:398`; board directive #141 (3), aura half), with Reserves meaning that player's own (`FR-ZON-001`). Layer 6 still clamps the sum of Layers 1, 4 and 5 at 0.

This is the shape every same-zone aura of Deliverable 6m's wave E reuses (design note `:209`): *Animoceps Stalk*, *Animoceps Node* and *Melodic Pummeler* are T-0257, and the keyword auras (*Hulking Incongruity*, *Ancestral Monument*) and *Brigand Highlander*'s Cornered self-power later add fields to the same `Static` and read the same `g.auras`. A second same-zone check, or a second walk over battlefield auras, would be a duplication finding.

## Inputs
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): Section 6.2's pipeline (`:476-498`, Layer 4 between 3 and 5, Layer 6 clamps); `FR-LAY-001` (`:500-504`): continuous evaluation, timestamp order inside a layer (addition, so order doesn't change a sum), Combat Power the only stat that clamps, and "continuous static abilities (Layer 4 auras) evaluate dynamically across board state changes"; `FR-TGT-001` `SVY-R44` (`:398`), last sentence: a continuous ability "is not on the stack, and uses its source's current zone"; `FR-ZON-001` (`:236`, Reserves and the Headquarters Zone are each player's own); `FR-ZON-004` (`:257`, last-known effective values immediately prior to departure); `FR-CBT-002` (`:585`, combat damage equals effective power).
- Board directive #141 (3), aura half (`BOARD.md`): "Continuous auras … use their source's current zone, since they aren't stack items." Its trigger fires with this task and T-0257, the first task whose `cards:` include a same-zone aura.
- Design note `docs/architecture/m2-ability-model.md`: `Static` (`:43`, "No aura is built yet"), Layers (`:55`, "Layer 4 is computed on demand from battlefield `Statics`"), Deliverable 6m's wave E bullet (`:209`).
- Survey rows (`docs/design/m2_starter_set_ability_survey.md:162-164`): Stalk `excludeSelf`; Node a Structure that doesn't buff itself; Pummeler includes itself.
- Code of the day (run branch at f87d28f):
  - `pkg/engine/ability/ability.go`: `StaticKind` and `StaticDeployPermission` (`:46-55`), `Static` (`:57-66`), `Filter` (`:275-285`), `ZoneRef` and `ZoneRefSame` (`:299-308`).
  - `pkg/engine/layers.go`: the file comment (`:1-5`, "There are no auras (Layer 4) yet"), `Game.Power` (`:54`), `Keywords` (`:78`).
  - `pkg/engine/filter.go`: `inLocked` (`:132`), `matchesFilterAt` (`:149`): already the one same-zone check, reading a `*LockedZone`.
  - `pkg/engine/placement.go`: `zoneLock(a)` (`:134`), the zone a source-relative filter reads for asset `a`; `deployPermission` (`:109`), the one reader of the other `Static` kind.
  - `pkg/engine/state.go`: `Game.Assets` (`:230`), every battlefield asset; `Asset.Printed` (`:85`): a face-down asset is the Scratch Company face, which has no `Ability`.
  - Every effective-power read already goes through `g.Power`: combat (`combat.go:105`), source-power values (`effects.go:420-423`), filter stats (`filter.go:72`), `effectiveState` (`layers.go:126`, used by the death sweep's all-at-once snapshot, `sweep.go:11-16`, and by lone departures), the snapshot (`snapshot.go:64`) and the view (`pkg/view/view.go:166`).
  - Searched `pkg/engine` (non-test) for `Layer 4`, `aura` and `Statics`: only the deploy permission reads `Statics`, and nothing adds a Layer 4 value. The task archive has no aura task (T-0214 built only `StaticDeployPermission`). No existing equivalent.

## Shape (the lead's plan)
- `pkg/engine/ability/ability.go`:
  - `StaticAura` after `StaticDeployPermission`, commented: a continuous Layer 4 aura (`FR-LAY-001`) of a face-up battlefield asset: every battlefield asset matching `Filter`, relative to the source's controller, has `+Power` Combat Power. A `ZoneRefSame` criterion reads the source's current zone (`SVY-R44`). `ActiveZone` is `ZoneAny`, the battlefield.
  - `Static` gains `Filter Filter` (an aura's: which assets it applies to; `Other` excludes the source; `Stat` stays `StatNone`, since a Layer 4 value read by a Layer 4 filter has no defined order) and `Power int` (an aura's Layer 4 power). Update the `Def.Statics` comment ("Only a discard-active deploy permission is a kind so far").
  - The comments of `Filter.ZoneRef` and `ZoneRefSame`: on a stack item, the zone it locked; on an aura, the source's current zone.
- `pkg/engine/layers.go`:
  - `g.auras(a *Asset) []ability.Static` (or an iterator; the engineer's choice): every `StaticAura` that applies to battlefield asset `a`, walking `g.Assets()` (or `assetsByID`, oldest first, for timestamp order) and each face-up source's `Printed().Ability.Statics`, with the test `g.matchesFilterAt(src.Controller, src.ID, zoneLock(src), st.Filter, a)`. This is the one reader: the later keyword and Cornered auras extend it, never copy it. No new zone check: `zoneLock` and `inLocked` are the ones stack items use.
  - `Game.Power`: `max(0, Printed().Power + the auras' Power + TempPower())`. Its comment names Layers 1, 4, 5 and 6.
  - The file comment: Layer 4 holds power auras.
- `pkg/engine/doc.go`: under built, a bullet for power auras (`Def.Statics` with `ability.StaticAura`, Layer 4, the source's current zone, `FR-LAY-001`, `SVY-R44`); under "Not built yet", "auras" becomes "keyword auras".
- No change to `filter.go`, `combat.go`, `sweep.go`, `snapshot.go` or `pkg/view`: each reads `g.Power` already, and the tests below pin combat and the sweep's last-known record.
- Not this task: the cards (T-0257); keyword and Cornered auras (*Hulking Incongruity*, *Ancestral Monument*, *Brigand Highlander*, their own tasks).

## Tests (shape tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/aura_test.go`. Cards are synthetic (`unit`, `structure` in `scenario_test.go`), with `c.Ability = &ability.Def{Statics: []ability.Static{…}}` as `deploy_discard_test.go:34` does. Write a local helper `auraUnit(name string, power, health int, f ability.Filter, n int) *Card`: a unit with one `Static{Kind: ability.StaticAura, Filter: f, Power: n}`, and one for a structure. "Same zone" below is `Filter.ZoneRef: ability.ZoneRefSame`. Place assets with `s.put`, move them with `s.moveTo`, and read `s.g.Power(a)`. Player 0 is the active player.

- `TestFR_LAY_001_AuraBuffsOtherAlliedInSameZone`: player 0's aura unit (power 1) in the Main Theater with filter `{Allegiance: Allied, Other: true, ZoneRef: Same}` and Power 1. Player 0's unit (power 2) and structure (power 0) in the Main Theater have power 3 and 1; player 0's unit (power 2) in its Reserves has 2; player 0's Headquarters has its printed power; player 1's unit (power 2) in the Main Theater has 2; the source has 1 (it excludes itself).
- `TestFR_LAY_001_AuraIncludesSourceWithoutOther`: the source is a Horror unit (power 4) in its Reserves with filter `{Allegiance: Allied, Kinds: KindUnit, Subtypes: ["Horror"], ZoneRef: Same}` and Power 1. The source has 5; another allied Horror unit in that Reserves gains 1; an allied non-Horror unit there doesn't; an allied structure there doesn't (if a structure can be put in Reserves by `s.put`; otherwise leave that case out).
- `TestFR_LAY_001_AuraFollowsSourcesCurrentZone`: source as in the first test, in the Main Theater, with allied units A in the Main Theater and B in its Reserves. A has +1, B none. After `s.moveTo(source, ZoneReserves)`: A none, B +1. After moving it back: A +1, B none (`SVY-R44`, last sentence: an aura uses its source's current zone).
- `TestFR_LAY_001_AuraAppliesToLaterArrivals`: with the source in the Main Theater, an allied unit put there afterwards has +1 at once, and an allied unit moved there from Reserves gains it on arrival and loses it when moved back (`FR-LAY-001`, auras evaluate dynamically, unlike one-shot modifiers).
- `TestFR_LAY_001_AuraEndsWhenSourceLeaves`: source and an allied unit A in the Main Theater; A has +1. Kill the source (a free Maneuver or `effectSource` kill step on it, or make it lethal and run `s.g.deathSweep()`): A is back to its printed power.
- `TestFR_ZON_001_AuraSameZoneReservesAreOwn`: player 0's source in player 0's Reserves with filter `{ZoneRef: Same, Other: true}` (any allegiance) and Power 1. Player 0's unit in player 0's Reserves has +1; player 1's unit in player 1's Reserves has none. Then with player 0's source in the Main Theater, the same filter gives player 1's Main Theater unit +1 (the Main Theater is shared).
- `TestFR_LAY_001_FaceDownSourceHasNoAura`: an aura unit put face down (as `s.putScratch` builds one, with the aura card as its `Inst.Card`, moved to the Main Theater) gives no allied Main Theater unit any power. A face-down Scratch Company *receives* a face-up source's aura: `s.putScratch(0)` moved to the Main Theater beside the source of the first test has power 2.
- `TestFR_LAY_001_AurasStackAndClampAfterLayer5`: two aura sources (each `{Allied, Other, Same}`, Power 1) and allied unit A (power 1) in the Main Theater: A has 3. Add a Layer 5 `Modifier{Layer: 5, Power: -3}` to A: power 0; with `Power: -5` instead: still 0 (Layer 6 clamps the sum of Layers 1, 4 and 5, not each layer). Each source has 2 (the other's aura).
- `TestFR_LAY_001_AuraFollowsSourcesController`: source of the first test and player 1's unit U in the Main Theater: U has no bonus. Set the source's `Controller` to 1 (as a control change does, T-0254): U has +1, and player 0's Main Theater unit loses it (the filter is relative to the source's current controller).
- `TestFR_CBT_002_AuraPowerDealsCombatDamage`: player 0's allied unit (power 2, Readied) in the Main Theater next to the source of the first test attacks player 1's unit (power 0, Health 10) in the Main Theater: the defender takes 3 damage.
- `TestFR_ZON_004_SweepRecordsAuraPowerBeforeAnyLeaves`: the source of the first test and allied unit A (power 2) in the Main Theater, both lethal (`Damage` equal to Health). After `s.g.deathSweep()`, both are gone and `s.g.LastKnownEffective(A.ID).Power` is 3: the sweep takes every dying asset's state before any leaves (`sweep.go:11`, board directive #13), so the source's departure doesn't lower A's record.

Existing tests stay as they are and must still pass.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none expected. Helpers local to the new file live in it.

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/layers.go`
- `pkg/engine/doc.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "Aura"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (no starter or stress card prints a `StaticAura`, so every log is byte-identical)

## Log
- 2026-11-16 lead-architect (wave H run, session 17): written against the run branch at f87d28f. *Animoceps Stalk* is next in Deliverable 6m's order (wave E, the first same-zone aura). No ruling needed: `SVY-R44`'s last sentence (`FR-TGT-001` `:398`) settles the aura's zone (the source's current one), `FR-ZON-001` makes Reserves each player's own, `FR-LAY-001` puts auras in Layer 4, evaluated continuously, with Layer 6 clamping the sum, and the survey rows (`:162-164`) say who excludes or includes itself. Split from the cards under the D2 size rule (shape and cards touch five non-test files); T-0257, the three power-aura cards (identical shapes), queued in this commit. Reuse: the same-zone check is `matchesFilterAt` with `zoneLock(src)`, the lock stack items already use, so no second zone check; `g.Power` is the one effective-power read, so no reader changes. Searched `pkg/engine` and the task archive for an aura or a Layer 4 value: none. Three non-test files. Directives: #141 (3)'s aura half fires with this task (the first whose plan builds a same-zone aura; T-0257's `cards:` include three), set `open` in this commit and applied when T-0256 merges. The other `waiting` ones were checked against the repo and none has fired (#6, #7: mode is `serial`; #15 (2): no draw or end-of-turn hook; #71: no each-player effect; #91 (2), #106, #120, #142 name cards or shapes no task touches yet). `main` has no commits the run branch lacks. Status `tests`.
