---
id: T-0258
title: "Aura keyword grants: StaticAura's Static.Resistant, a Layer 3 grant read by g.Keywords through the one aura reader g.auras (wave H shape, sonnet)"
kind: shape
tier: sonnet
wave: H
status: tests
branch: m2/autoloop-wave-h/T-0258-aura-resistant-grant
depends_on: [T-0256]
claims: [shape:StaticAura keyword grant (Static.Resistant: every asset an aura's Filter matches has +Resistant N, a Layer 3 keyword grant read by g.Keywords through g.auras)]
cards: []
requirements: [FR-LAY-001, FR-KWD-002b, FR-REP-004, FR-ZON-004, SVY-R44]
---

## Goal
An aura can grant a keyword as well as power: *"Allied Horror units in the same zone have Resistant 1."* (*Hulking Incongruity*). T-0256's `ability.StaticAura` gains a `Resistant int`: every battlefield asset the aura's `Filter` matches, relative to its source's controller, has `+Resistant` Resistant for as long as the aura applies. It applies exactly when T-0256's power half does (a face-up battlefield source, its current controller and its current zone, `SVY-R44`, last sentence), so it goes through the same reader, `g.auras`. Its layer is Layer 3, Dynamic Keyword Grants, because FRD 01's pipeline puts every keyword grant there (survey gap G2: FRD 02 said Layer 4; FRD 01 wins; no M2 card can tell the two apart, since Resistant is a sum and nothing reads it between the layers). Damage then reads it in Stage 2 like printed Resistant (`FR-KWD-002b`, `FR-REP-004`), because `dealDamage` already reads `g.Keywords`.

This is the keyword half that T-0256's Goal reserved for the keyword auras. *Hulking Incongruity* is T-0259, queued. *Ancestral Monument* ("have Cornered: Resistant 1") later adds the Cornered gate to the same `Static` and the same `g.auras`. A second walk over battlefield auras, or a second same-zone check, would be a duplication finding.

## Inputs
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): the pipeline (`:476-498`; Layer 3 "Dynamic Keyword Grants & Suppressions", Layer 4 "Continuous Auras"); `FR-LAY-001` (`:500-504`): continuous evaluation, and continuous auras "evaluate dynamically across board state changes"; `FR-LAY-002` (`:510`, last sentence): continuous defensive auras grant Resistant N or Stalwart, never Armor; `FR-KWD-002b` (`:614`): Resistant N reduces each discrete instance of incoming damage; `FR-REP-004` Stage 2 (`:557`): the target's **total** Resistant is subtracted; `FR-ZON-004`: last-known effective values from immediately before departure.
- Survey (`docs/design/m2_starter_set_ability_survey.md`): gap G2 (`:48`, keyword grants in Layer 3); *Hulking Incongruity*'s row (`:165`, "includes itself").
- Design note `docs/architecture/m2-ability-model.md`: `Static` (`:43`, "a Layer 3/4 aura"), Layers (`:55`), Deliverable 6m's wave E bullet (`:209`: one aura shape for all six cards), Deliverable 2's G2 (`:219`).
- Code of the day (run branch at 51fef91):
  - `pkg/engine/ability/ability.go`: `Def.Statics` comment (`:41-44`, "a battlefield power aura"), `StaticAura` (`:56-60`), `Static` (`:64-79`), `Filter` (`:287-297`, no keyword criterion).
  - `pkg/engine/layers.go`: the file comment (`:1-5`), `Game.Power` (`:56`), `Game.auras` (`:64-85`, "every power aura"), `Game.Keywords` (`:104-118`: printed, then `Def.Cornered`, then Layer 3 modifiers).
  - `pkg/engine/damage.go:23`: Stage 2 reads `g.Keywords(target).Resistant`; the damage event logs it (`events.go:252`).
  - Every effective-keyword read goes through `g.Keywords`: damage, combat (`combat.go:51-63`), Ward (`filter.go:85`), `effectiveState` (`layers.go:155`, the sweep's all-at-once record), the snapshot (`snapshot.go:64`) and so the view. `newAsset` (`zones.go:245`) reads printed keywords on entry only, for Armor and Deploys Exhausted, which no aura grants.
  - No recursion: `g.auras` tests with `matchesFilterAt`, which never reads keywords (Ward is `legalTarget`'s, not the filter's), so `g.Keywords` calling `g.auras` can't loop.
  - Searched `pkg/engine` (non-test) and the task archive for an aura keyword grant: none. `OpGrantKeyword` (T-0219) grants a flag keyword as a one-shot Layer 3 modifier, which is a different rule (`FR-LAY-001`, One-Shot Modifier Snapshot Locking) and stays as it is.

## Shape (the lead's plan)
- `pkg/engine/ability/ability.go`:
  - `Static` gains `Resistant int` after `Power`: an aura's keyword grant, `+Resistant` Resistant to every asset `Filter` matches; a Layer 3 grant (`FR-LAY-001`, survey G2); may be negative, like printed Resistant. 0 grants nothing.
  - `StaticAura`'s comment: every matching asset has `+Power` Combat Power (Layer 4) and `+Resistant` Resistant (a Layer 3 keyword grant). Keep the rest. `Def.Statics`' comment: "a battlefield aura", not "a battlefield power aura". `Static.Filter`'s comment: it stays free of keyword criteria too (a Layer 3 value read by its own filter has no defined order).
- `pkg/engine/layers.go`:
  - `Game.Keywords` adds `st.Resistant` for each aura `g.auras(a)` returns, after the card's own Cornered keywords and before or after the modifiers (a sum; the engineer's choice). Its comment names the aura grant.
  - `Game.auras`' comment: every aura, not every power aura; each caller reads the field it needs. No second walk over `g.Assets()`.
  - The file comment: Layer 3 also holds auras' keyword grants (`Game.auras`).
- `pkg/engine/doc.go`: the "Power auras" bullet becomes "Auras": +Power in Layer 4 and +Resistant as a Layer 3 keyword grant, read by `Game.Power` and `Game.Keywords` through `Game.auras`. Under "Not built yet", "keyword auras" becomes "Cornered-gated auras ("have Cornered: Resistant 1.")".
- No change to `damage.go`, `combat.go`, `filter.go`, `sweep.go`, `snapshot.go` or `pkg/view`: each reads `g.Keywords` already, and the tests below pin damage and the sweep's record.
- Not this task: *Hulking Incongruity* (T-0259); the Cornered gate (*Ancestral Monument*) and *Brigand Highlander*'s Cornered self-power, their own tasks; flag keywords (Stalwart, Guardian) and Aggressive on an aura (no Animus or Path card grants one).

## Tests (shape tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/aura_keywords_test.go`. Cards are synthetic, as in `aura_test.go` (T-0256): a unit with `c.Ability = &ability.Def{Statics: []ability.Static{{Kind: ability.StaticAura, Filter: f, Resistant: n}}}`, through a local helper (`aura_test.go`'s `auraStatic` and `auraUnit` take a power only; don't change them, add a local one). "Allied Horror units in the same zone" is `ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Subtypes: []string{"Horror"}, ZoneRef: ability.ZoneRefSame}`; give the synthetic cards the subtype the way `aura_test.go`'s Horror test does. Read `s.g.Keywords(a).Resistant`. Player 0 is the active player.

- `TestFR_LAY_001_AuraGrantsResistantToMatchesInSameZone`: player 0's Horror source (Resistant 1, Power 0) in player 0's Reserves. Resistant 1 for: the source itself (no `Other`, so it includes itself) and another allied Horror unit in that Reserves. Resistant 0 for: an allied non-Horror unit in that Reserves, an allied Horror unit in the Main Theater, and player 1's Horror unit in player 1's Reserves. Every one keeps its printed power (a keyword grant adds no power).
- `TestFR_LAY_001_AuraResistantAddsToPrintedAndStacks`: with the source of the first test, an allied Horror unit printing Resistant 2 in that Reserves has 3. With a second such source there, each source and the Resistant-2 unit have 2 and 4 (each source includes itself and the other).
- `TestFR_LAY_001_AuraResistantFollowsSourcesZone`: the source in player 0's Reserves, allied Horror units A there and B in the Main Theater: A 1, B 0. After `s.moveTo(source, ZoneMainTheater)`: A 0, B 1. After the source is killed (or made lethal and `s.g.deathSweep()` run): B 0.
- `TestFR_LAY_001_FaceDownSourceGrantsNoKeyword`: the source put face down (as `s.putScratch` builds one, with the aura card as its `Inst.Card`) next to an allied Horror unit gives it Resistant 0.
- `TestFR_REP_004_AuraResistantReducesDamage`: player 1's source and player 1's Horror unit D (Health 5, no printed Resistant) in the Main Theater; player 0's Readied unit (power 3) attacks D: D takes 2 damage (Stage 2 subtracts its total Resistant), and the attack's damage event on D logs Resistant 1.
- `TestFR_ZON_004_SweepRecordsAuraResistantBeforeAnyLeaves`: the source and allied Horror unit A in the same zone, both lethal. After `s.g.deathSweep()`, both are gone and `s.g.LastKnownEffective(A.ID).Keywords.Resistant` is 1: the sweep records every dying asset before any leaves (`sweep.go:11`), so the source's departure doesn't lower A's record.

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
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (no starter or stress card prints an aura with `Resistant`, so every log is byte-identical)

## Log
- 2026-11-16 lead-architect (wave H run, session 19): written against the run branch at 51fef91. *Hulking Incongruity* is next in Deliverable 6m's order (wave E, after T-0257's power auras). No ruling needed: FRD 01's pipeline and `FR-LAY-001` put keyword grants in Layer 3 (survey G2 settles the FRD 02 difference), `FR-REP-004` Stage 2 subtracts total Resistant, `SVY-R44` gives the aura its source's current zone, and the survey row (`:165`) says it includes itself. Split from the card under the D2 size rule (shape and card touch five non-test files); T-0259, *Hulking Incongruity*, queued in this commit. Reuse: `g.auras` is the one reader of which auras apply, and `g.Keywords` the one keyword read, so damage, combat, the sweep, the snapshot and the view need no change; `OpGrantKeyword` is a one-shot modifier, a different rule. Three non-test files. Directives: none `open`; each `waiting` one was checked against the repo and none has fired (#6, #7: mode is `serial`; #15 (2): no draw or end-of-turn hook; #71: no each-player effect; #91 (2), #106, #120, #142 name cards or shapes no task touches yet). `main` has no commits the run branch lacks. Status `tests`.
