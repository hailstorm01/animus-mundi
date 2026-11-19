---
id: T-0273
title: Cornered and aura grants are ability.Keywords sets, and Game.Keywords folds once (directive #159, part 2; harness, no behavior change)
kind: harness
tier: sonnet
wave: H.1 (board directive #159, structural review 13, finding 1)
status: review
branch: m2/autoloop-wave-h.1/T-0273-cornered-and-aura-grants-as-keywords
depends_on: [T-0272]
claims: [type:ability.WhileCornered, field:ability.Static.Keywords, func:Game.Keywords, func:Game.whileCornered]
cards: []
requirements: []
---

## Goal
The rest of directive #159 after T-0272: `ability.WhileCornered` becomes an `ability.Keywords` plus `Power` (its own `Aggressive` and `Resistant` fields go, and its `Add` folds the keywords through `Keywords.Add`), `Static.Resistant` becomes `Static.Keywords` (an aura's Layer 3 grant of any keyword), and `Game.Keywords` folds the printed set, the Cornered grant, each aura's grant and each Layer 3 modifier through the one `Keywords.Add`, adding no field by hand. Match output is unchanged.

## Inputs
- Board directive #159 (structural review 13, approved by user, 2026-11-18; no alias, user 2026-11-18), the part T-0272 left. Design note Deliverable 6n (`docs/architecture/m2-ability-model.md`).
- ADR 0007 (a move updates every reference) and ADR 0008 (the keyword set lives in `ability`).
- Rules the code implements, unchanged: `FR-LAY-001` (Layer 1 printed, Layer 3 grants), `FR-KWD-002e` (Cornered), `SVY-R18` (keywords and power behind "Cornered:"), `SVY-R44` (an aura's "same zone").
- No record or snapshot field changes: `WhileCornered` and `Static` are card-definition shapes, never serialized, so match output is unchanged (ADR-0005).

### Where it is today (at `autoloop/m2-wave-h.1` c1b7bf8)
| Where | Today |
| :--- | :--- |
| `pkg/engine/ability/ability.go:56-61` | `StaticAura`'s comment: "+Resistant Resistant (a Layer 3 keyword grant)" |
| `pkg/engine/ability/ability.go:82-85` | `Static.Resistant int`, an aura's one-keyword grant |
| `pkg/engine/ability/ability.go:101-120` | `WhileCornered{Aggressive, Resistant, Power int}` and its field-by-field `Add` |
| `pkg/engine/layers.go:72` | `auras`' comment: "Each caller reads the field it needs (Power, Resistant)." |
| `pkg/engine/layers.go:106-125` | `Game.Keywords`: adds `wc.Aggressive`, `wc.Resistant` and each aura's `st.Resistant` by hand, then folds Layer 3 modifiers with `kw.Add` |
| `pkg/engine/layers.go:127-144` | `Game.whileCornered` (its body needs no change: it already sums with `WhileCornered.Add`) |
| `pkg/engine/cards/path.go:68`, `:152`, `:654` | `WhileCornered{Aggressive: 3}` (*Allshan*), `{Resistant: 1}` (*Ancestral Monument*'s aura, *Dancers*); `:270` `{Power: 2}` (*Brigand Highlander*) keeps its shape |
| `pkg/engine/cards/animus.go:460` | *Hulking Incongruity*'s aura: `Resistant: 1,` |
| `pkg/engine/doc.go:81` | "a face-up battlefield asset gives +Power in Layer 4 and +Resistant as a Layer 3 keyword grant" |
| `docs/architecture/engine-structs.md:218`, `:242`, `:302` | still name `CorneredKeywords`, the old name of `WhileCornered` (T-0260 renamed it and left these) |

## Edits (lead decisions)
1. **`pkg/engine/ability/ability.go`.**
   - `type WhileCornered struct { Keywords Keywords; Power int }`. Its comment keeps what it says today, with "the keywords" meaning `Keywords`, a Layer 3 grant of any keyword set.
   - `func (w WhileCornered) Add(o WhileCornered) WhileCornered`: `w.Keywords.Add(o.Keywords)`, `w.Power += o.Power`, return `w`. Comment: `Add returns w plus o: the keywords folded through Keywords.Add, the power summed.`
   - `Static.Resistant int` becomes `Static.Keywords Keywords`. Comment: `Keywords is an aura's keyword grant: every asset Filter matches has them, a Layer 3 grant folded through Keywords.Add (FR-LAY-001, survey G2). Resistant may be negative, like printed Resistant; the zero value grants nothing.`
   - `StaticAura`'s comment: "+Power Combat Power (Layer 4) and Keywords (a Layer 3 keyword grant)".
2. **`pkg/engine/layers.go`.**
   - `Game.Keywords`: `kw := a.Printed().Keywords`; `kw.Add(g.whileCornered(a).Keywords)`; for each aura, `kw.Add(st.Keywords)`; the Layer 3 modifier loop as today. No keyword field is named. Comment: the printed set (Layer 1), folded with the card's own and auras' Cornered keywords while its controller is Cornered (`Game.whileCornered`, SVY-R18), each applying aura's `Keywords` (`Game.auras`) and each Layer 3 modifier's, all through `Keywords.Add` (FR-LAY-001). Keep the sentence on face-down assets.
   - `auras`' comment: "(Power, Keywords, Cornered)".
3. **Card definitions.** `cards/path.go`: `ability.WhileCornered{Keywords: ability.Keywords{Aggressive: 3}}` and `{Keywords: ability.Keywords{Resistant: 1}}` (twice); `{Power: 2}` unchanged. `cards/animus.go:460`: `Keywords: ability.Keywords{Resistant: 1},`.
4. **Docs and comments.** `doc.go:81`: "gives +Power in Layer 4 and its Keywords as a Layer 3 grant". `engine-structs.md`: `CorneredKeywords` becomes `WhileCornered` in the node (id and label), the `Def -->|"Cornered"|` edge and the `ability/ability.go` table row. Every other comment in `pkg/` that names `Static.Resistant` or `WhileCornered`'s `Aggressive`/`Resistant` fields names the new shape. (`docs/architecture/m2-ability-model.md`'s Boundaries are #163's, not this task's.)

## Acceptance Cases
1. An aura `Static{Kind: StaticAura, Filter: <allied units in the same zone>, Keywords: ability.Keywords{Ward: true, Resistant: 1}}` on a face-up allied unit, and an allied unit printed `Resistant 1` in the same zone → `g.Keywords(u)` has `Ward` and `Resistant` 2; an allied unit in the other zone has neither. New test `TestFR_LAY_001_AuraGrantsAnyKeywordSet` in `pkg/engine/aura_keywords_test.go`.
2. A card with `Def.Cornered = WhileCornered{Keywords: ability.Keywords{Stalwart: true}}` → `g.Keywords(u).Stalwart` is true while its controller is Cornered and false when not. New test `TestFR_KWD_002e_CorneredGrantsAnyKeywordSet` in `pkg/engine/cornered_keywords_test.go`.
3. Every other existing test passes with only the edits under **Tests**.
4. `git grep -nE "Aggressive +int|Resistant +int" -- pkg/engine/ability/ability.go` finds nothing. `git grep -nE "(wc|st)\.(Aggressive|Resistant)" -- pkg` finds nothing. `git grep -nE "WhileCornered\{(Aggressive|Resistant)" -- pkg` finds nothing. `git grep -n CorneredKeywords -- pkg docs/architecture/engine-structs.md` finds nothing.
5. `go tool task check` passes; `go tool task sim-compare -- autoloop/m2-wave-h.1` passes (starter and stress byte for byte).

## Tests (authorized edits, all in the `test(T-0273)` commit)
- **Mechanical reshape, no other line changes:** `ability.WhileCornered{Aggressive: N}` → `ability.WhileCornered{Keywords: ability.Keywords{Aggressive: N}}`, likewise `{Resistant: N}` and `{Aggressive: 3, Resistant: 1}`; `{Power: N}` stays. In:
  - `pkg/engine/aura_cornered_test.go:26`, `:74`
  - `pkg/engine/cornered_keywords_test.go:172`, `:204`, `:220`, `:245`, `:287`, `:302`, `:324`, `:340`
  - `pkg/engine/grant_keyword_test.go:71`
  - `pkg/sim/monitor_keywords_test.go:122`
  - `pkg/view/view_test.go:1122`
- `pkg/engine/aura_keywords_test.go:23`: `Resistant: n` → `Keywords: ability.Keywords{Resistant: n}`.
- The two new tests of Acceptance Cases 1 and 2.

The test commit doesn't compile until the reshape lands; that is its failing state.

## Allowed Files
- `pkg/engine/ability/ability.go`, `pkg/engine/layers.go`
- Reference updates only (ADR 0007; they don't count toward the D2 size rule): `pkg/engine/cards/path.go`, `pkg/engine/cards/animus.go`, `pkg/engine/doc.go`, `docs/architecture/engine-structs.md`, and any other non-test Go file under `pkg/`, `cmd/` or `tools/` that `go build ./...` shows naming a reshaped field
- The test files under **Tests**

## Done When
- [ ] `go test ./pkg/engine -run "TestFR_LAY_001_AuraGrantsAnyKeywordSet|TestFR_KWD_002e_CorneredGrantsAnyKeywordSet|Cornered|Aura"`
- [ ] Acceptance Case 4's greps find nothing
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h.1` passes (starter and stress sims byte-identical to the run branch)

## Log
- 2026-11-18 lead-architect (wave H.1 session 1): queued as the second part of #159 (the directive's own split; D2.4), after T-0272.
- 2026-11-18 lead-architect (wave H.1 session 2): wrote Inputs, Edits, Acceptance Cases, Tests and Allowed Files against `autoloop/m2-wave-h.1` c1b7bf8; `ready` (harness: no `tests` step). Dispatching `kernel-engineer` (sonnet) on `m2/autoloop-wave-h.1/T-0273-cornered-and-aura-grants-as-keywords`.
- 2026-11-18 kernel-engineer (sonnet): `review`. Test commit first (mechanical reshape of the listed tests, plus `TestFR_LAY_001_AuraGrantsAnyKeywordSet` and `TestFR_KWD_002e_CorneredGrantsAnyKeywordSet`; it did not compile). Then: `WhileCornered{Keywords, Power}` with `Add` folding through `Keywords.Add`; `Static.Resistant` became `Static.Keywords`; `Game.Keywords` folds printed, Cornered, each aura's and each Layer 3 modifier's keywords through `Keywords.Add`; card definitions in cards/path.go and cards/animus.go, doc.go, the `auras` comment and `engine-structs.md` (`CorneredKeywords` to `WhileCornered`) updated. Reused `Keywords.Add`; no new shapes. Ran: Acceptance Case 4 greps (nothing found), `go tool task check` (pass), `go tool task sim` (clean), `go tool task sim-compare -- autoloop/m2-wave-h.1` (byte-identical), the Done When `go test -run` (pass).
