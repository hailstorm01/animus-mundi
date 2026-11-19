---
id: T-0272
title: The keyword set moves to ability.Keywords, and a step grants a Keywords set (directive #159, part 1; harness, sonnet, no behavior change)
kind: harness
tier: sonnet
wave: H.1 (board directive #159, structural review 13, finding 1)
status: review
branch: m2/autoloop-wave-h.1/T-0272-keywords-in-ability
depends_on: []
claims: [type:ability.Keywords, file:pkg/engine/ability/keywords.go, func:ability.ParseKeywordLine, method:ability.Keywords.Lines, method:ability.Keywords.Add, field:ability.Step.Keywords, func:stepGrantKeyword]
cards: []
requirements: []
---

## Goal
`Keywords`, its name table `keywordDefs`, `ParseKeywordLine`, `Lines` and the fold move from package `engine` to package `ability`, as a pure move ([ADR 0008](../../../docs/architecture/decisions/0008-keywords-live-in-ability.md)). Every reference in every package names `ability.Keywords`; no `engine.Keywords` is left, as an alias or otherwise ([ADR 0007](../../../docs/architecture/decisions/0007-moves-update-every-reference.md)). Because `stepGrantKeyword` reads the name table, which `engine` can no longer see, `OpGrantKeyword`'s `Step.Keyword` (an `ability.Keyword`) becomes `Step.Keywords` (an `ability.Keywords`) in the same task, so a step can grant any keyword, and `ability.KeywordStalwart` goes. Match output is unchanged.

**Not this task:** `WhileCornered`'s own `Aggressive`/`Resistant`, `Static.Resistant`, and `Game.Keywords`' hand-added Cornered and aura fields. They are **T-0273** (queued on the board, `depends_on: [T-0272]`). Until then they stay as they are; this task adds no bridge between the two.

## Inputs
- Board directive #159 (structural review 13, approved by user, 2026-11-18; no alias, user 2026-11-18). Design note Deliverable 6n (`docs/architecture/m2-ability-model.md`).
- ADR 0007 (a move updates every reference) and ADR 0008 (the keyword set lives in `ability`), both on the run branch.
- Rules the code implements, unchanged: `FR-KWD-002` (the standalone keywords), `FR-LAY-001` (Layer 1 printed, Layer 3 grants), `FR-LAY-003` (a step's grant lasts for `Step.Until`).
- ADR-0005: records and snapshots keep their JSON field names and tags (`AssetSnap.Keywords`, `json:"keywords"`; the `Keywords` struct's own tags move unchanged).

### Where it is today (at `autoloop/m2-wave-h.1` 7a8e4f5)
| Where | Today |
| :--- | :--- |
| `pkg/engine/card.go:1` | file comment names `Keywords` |
| `pkg/engine/card.go:27-79` | `Keywords`, `keywordDef`, `keywordDefs` (Stalwart's and Aggressive's names via `ability.KeywordStalwart.String()` / `ability.KeywordAggressive.String()`), `Keywords.Lines` |
| `pkg/engine/card.go:96`, `:142` | `Card.Keywords Keywords`; `scratchFace`'s `Keywords{DeploysExhausted: true}` |
| `pkg/engine/cardload.go:20-40`, `:193` | `keywordN`, `ParseKeywordLine`; its one caller in `loadSheets` |
| `pkg/engine/layers.go:24`, `:176` | `Modifier.Keywords Keywords`, `EffectiveState.Keywords Keywords` |
| `pkg/engine/layers.go:106-138` | `Game.Keywords` (calls `kw.add`), and `(*Keywords).add` |
| `pkg/engine/effects.go:505-529` | `stepGrantKeyword`: looks up `st.Keyword.String()` in `keywordDefs`, flag keywords only |
| `pkg/engine/snapshot.go:29` | `AssetSnap.Keywords Keywords \`json:"keywords"\`` |
| `pkg/engine/ability/ability.go:212-232` | `Keyword` (`KeywordNone`, `KeywordAggressive`, `KeywordStalwart`) and `String` |
| `pkg/engine/ability/ability.go:447`, `:477` | `OpGrantKeyword`'s comment; `Step.Keyword Keyword` |
| `pkg/engine/cards/path.go:38` | *Battle Trance*'s grant: `Keyword: ability.KeywordStalwart,` |
| `pkg/engine/doc.go:126-127` | Files list: "card.go: card types. cardload.go: the CSV loader and keyword lines." |
| `docs/architecture/engine-structs.md:26`, `:119-121`, `:144-145`, `:198`, `:276-277`, `:302` | `Keywords` and `keywordDef` drawn and listed in package `engine`; `ParseKeywordLine` under `cardload.go` |
| tests | `Keywords{…}` in about 119 `pkg/engine` test files (34 of them don't import `ability` yet), `engine.Keywords{…}` in 10 `pkg/sim` and `pkg/view` test files; `ParseKeywordLine` in `keywords_test.go`; `ability.KeywordStalwart`/`KeywordAggressive` in `grant_keyword_test.go` and `opponent_turn_duration_test.go` |

## Edits (lead decisions)

1. **`pkg/engine/ability/keywords.go` (new).** File comment: `// The keyword set: the standalone keywords a card prints or is granted (FR-KWD-002), their printed names, and how two sets combine.` Move here, unchanged except as listed: `Keywords` (fields, tags and comment), `keywordDef`, `keywordDefs`, `Keywords.Lines`, `keywordN`, `ParseKeywordLine`, and the fold, exported as `func (k *Keywords) Add(o Keywords)` with `add`'s comment. In `keywordDefs`, Stalwart's name becomes the literal `"Stalwart"`, and Aggressive's stays `KeywordAggressive.String()`. Update `keywordDefs`' comment: each keyword's name is written only here, or, for a keyword that is a triggered ability, in `Keyword.String`.
2. **`pkg/engine/ability/ability.go`.**
   - `Keyword` becomes only the identity of a keyword that is a triggered ability. Remove `KeywordStalwart` and its `String` case. Its comment: `Keyword names a keyword that is a triggered ability, the one a trigger comes from (Aggressive, FR-KWD-002f). The zero value is a printed ability.`
   - `Step.Keyword Keyword` becomes `Keywords Keywords // OpGrantKeyword only: the keywords granted`. `OpGrantKeyword`'s comment says `Step.Keywords`.
3. **`pkg/engine/card.go`**: remove what moved; `Card.Keywords ability.Keywords`; `scratchFace` uses `ability.Keywords{…}`; the file comment drops `Keywords`. Drop imports that become unused.
4. **`pkg/engine/cardload.go`**: remove `keywordN` and `ParseKeywordLine`; the caller calls `ability.ParseKeywordLine`. Drop unused imports.
5. **`pkg/engine/layers.go`**: `Modifier.Keywords`, `EffectiveState.Keywords` and `Game.Keywords`' result are `ability.Keywords`; `Game.Keywords` calls `kw.Add(m.Keywords)`; remove `add`. Everything else in `Game.Keywords` (the Cornered and aura lines) stays for T-0273.
6. **`pkg/engine/effects.go`, `stepGrantKeyword`**: no name lookup. If `st.Keywords` is the zero `ability.Keywords{}`, it grants nothing and returns 0, as today for a keyword with no flag. Otherwise each asset gets `Modifier{Layer: 3, Keywords: st.Keywords}` (through `r.modifier`, as today), and the event detail names the keywords as `strings.Join(st.Keywords.Lines(), ", ")` in the same `"%s gains %s %s"` format, so a Stalwart grant still logs `<asset> gains Stalwart until end of turn`. Update its comment.
7. **`pkg/engine/snapshot.go`**: `AssetSnap.Keywords ability.Keywords`, same tag.
8. **`pkg/engine/cards/path.go:38`**: `Keywords: ability.Keywords{Stalwart: true},`.
9. **Docs and comments**: `doc.go`'s Files list says the keyword set, its names and line parser are in `pkg/engine/ability` (`keywords.go`), and `cardload.go` is the CSV loader. `engine-structs.md`: `Keywords` and `keywordDef` move into the `pkg/engine/ability` subgraph and table (a `keywords.go` row), the `card.go` row drops them, the `cardload.go` row drops `ParseKeywordLine`. Every other comment in `pkg/` that says `keywordDefs`, `Keywords.add` or `Step.Keyword` names the new place. (`docs/architecture/m2-ability-model.md`'s Boundaries are #163's, not this task's.)

## Acceptance Cases
1. A step `{Op: OpGrantKeyword, Subject: SubjectTarget, Target: 0, Keywords: ability.Keywords{Ward: true, Resistant: 2}}` on an allied unit printed with `Resistant 1` → `g.Keywords(u)` has `Ward` and `Resistant` 3 until end of turn; one `EventKeyword` for the unit with detail `<unit> gains Resistant 2, Ward until end of turn`; after Sub-Phase 4c of that turn, `Ward` is false and `Resistant` is 1. New test `TestFR_LAY_001_GrantedKeywordSetIsEffective` in `pkg/engine/grant_keyword_test.go`.
2. A step whose `Keywords` is the zero value → no modifier, no `EventKeyword`, and `g.Keywords(u)` equals the printed set. This replaces `TestFR_LAY_001_GrantingAggressiveGrantsNothing`, which pinned the old enum's limit (it had no way to say "Aggressive N"), not a rule: rename it `TestFR_LAY_001_EmptyGrantGrantsNothing`, grant `ability.Keywords{}`, and check `g.Keywords(u) == u.Printed().Keywords` in place of the `Aggressive` check; its other two checks stay.
3. Every other existing test passes with only the edits under **Tests**.
4. `git grep -n "engine.Keywords"` finds nothing in `pkg/`, `cmd/` or `tools/`. `git grep -nE "type Keywords ="` finds nothing. `git grep -nw KeywordStalwart` finds nothing in Go code. `git grep -nE "keywordDefs|ParseKeywordLine" -- pkg` finds nothing outside `pkg/engine/ability/` except the qualified call `ability.ParseKeywordLine` (in `cardload.go` and `keywords_test.go`). `git grep -n "Step.Keyword\b"` finds nothing.
5. `go tool task check` passes; `go tool task sim-compare -- autoloop/m2-wave-h.1` passes (starter and stress byte for byte).

## Tests (authorized edits, all in the `test(T-0272)` commit)
- **Mechanical rename, every `_test.go` file in `pkg/engine`, `pkg/sim` and `pkg/view`:** `Keywords` as a type or composite literal becomes `ability.Keywords` (in `pkg/engine`) and `engine.Keywords` becomes `ability.Keywords` (in `pkg/sim`, `pkg/view`), adding the `animus-mundi/pkg/engine/ability` import where missing. Field accesses (`c.Keywords`, `g.Keywords(a)`) and composite-literal keys (`Keywords: …`) don't change. Helper signatures in `scenario_test.go` and other test files that take a `Keywords` take an `ability.Keywords`. No other line in these files changes.
- `pkg/engine/keywords_test.go`: `ParseKeywordLine` → `ability.ParseKeywordLine`.
- `pkg/engine/grant_keyword_test.go`: `grantStep(slot int, kw ability.Keywords)` builds `Keywords: kw`; each `grantStep(…, ability.KeywordStalwart)` passes `ability.Keywords{Stalwart: true}`; line 150's `Keyword: ability.KeywordStalwart` → `Keywords: ability.Keywords{Stalwart: true}`; Acceptance Case 2's rename; Acceptance Case 1's new test.
- `pkg/engine/opponent_turn_duration_test.go:108`, `:129`: `Keyword: ability.KeywordStalwart` → `Keywords: ability.Keywords{Stalwart: true}`.

The test commit doesn't compile until the move lands; that is its failing state.

## Allowed Files
- `pkg/engine/ability/keywords.go` (new), `pkg/engine/ability/ability.go`, `pkg/engine/effects.go`, `pkg/engine/layers.go`
- Reference updates only (ADR 0007; they don't count toward the D2 size rule): `pkg/engine/card.go`, `pkg/engine/cardload.go`, `pkg/engine/snapshot.go`, `pkg/engine/cards/path.go`, `pkg/engine/doc.go`, any other non-test Go file under `pkg/`, `cmd/` or `tools/` that `go build ./...` shows naming a moved declaration, and `docs/architecture/engine-structs.md`
- The test files under **Tests**

## Done When
- [ ] `go test ./pkg/engine -run "TestFR_LAY_001_GrantedKeywordSetIsEffective|TestFR_LAY_001_EmptyGrantGrantsNothing|TestKeywords_"`
- [ ] Acceptance Case 4's greps find nothing
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h.1` passes (starter and stress sims byte-identical to the run branch)

## Log
- 2026-11-18 lead-architect (wave H.1 session 1): written from directive #159. The directive allows the move and the reshaping of `WhileCornered`, `Static` and `Step` as two tasks; `Step` comes into this one because `stepGrantKeyword` reads `keywordDefs` by name, which `engine` can't see once it moves, and anything else would need a bridge (ADR 0007). T-0273, the rest of #159, is queued. Harness, so no `tests` step: status `ready`. Tier `sonnet`: the rename spans about 130 test files.
- 2026-11-18 kernel-engineer: implemented. Reuse search: the move is of existing code (`Keywords`, `keywordDefs`, `Lines`, `add`, `ParseKeywordLine`); no new shape beyond `Step.Keywords`. Test commit 68487c6f-amended (129 test files: `Keywords` type and literals to `ability.Keywords` by a script that skipped strings, comments, `.Keywords` selectors and `Keywords:` keys; `ability` import added where missing, then gofmt, which puts it before `testing` in the one block where there was no separate group; plus the named edits to `grant_keyword_test.go` (Case 1 `TestFR_LAY_001_GrantedKeywordSetIsEffective`, Case 2 `TestFR_LAY_001_EmptyGrantGrantsNothing`), `keywords_test.go`, `opponent_turn_duration_test.go`). Implementation: new `ability/keywords.go` (Stalwart's name is the literal; Aggressive's still `KeywordAggressive.String()`); `Keyword` loses `KeywordStalwart`; `Step.Keywords`; `card.go`, `cardload.go`, `layers.go` (`Add` replaces `add`), `snapshot.go`, `cards/path.go`, `doc.go`, `engine-structs.md` updated; `stepGrantKeyword` returns 0 for the zero set, else grants the set and logs `strings.Join(Lines(), ", ")`. `gofmt` also realigned a few comment columns in `cards/path.go`. No alias or wrapper. Commands: `go build ./...`, `go vet ./...` clean; Done When `go test -run` ok; Case 4 greps find nothing in Go code (only task, board, ADR and archive docs mention the old names); `go tool task check` exit 0; `go tool task sim` exit 0; `go tool task sim-compare -- autoloop/m2-wave-h.1` PASS byte for byte. Look at: the import grouping in the renamed tests, and the empty-set test comparing `Keywords` structs.
