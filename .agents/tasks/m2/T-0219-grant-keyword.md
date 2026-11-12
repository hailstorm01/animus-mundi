---
id: T-0219
title: "A keyword granted until end of turn, on the asset a deployed card became (wave F shape, sonnet; for Allshan's Battle Trance)"
kind: shape
tier: sonnet
wave: F
status: review
branch: m2/autoloop-wave-f/T-0219-grant-keyword
depends_on: [T-0218]
claims: [primitive:GRANT_KEYWORD, modifier:keyword, subject:pile-slot-became, event:keyword]
cards: []
requirements: [FR-LAY-001, FR-LAY-003, FR-TRN-003a, FR-ENT-103]
---

## Goal
A step can grant a keyword until end of turn (`OpGrantKeyword`): a Layer 3 modifier on T-0218's `Asset.Modifiers` list, folded by `Game.Keywords`, which expires in Sub-Phase 4c with the Layer 5 power modifiers (`SVY-R39`). And a later step of the same resolution can act on the asset a pile slot's card became when an earlier step deployed it ("it" in *Allshan's Battle Trance*: "Deploy chosen unit card from your discard pile to the Main Theater. Until end of turn, it gains Stalwart and gets +2 power."). The card itself is T-0220. Tests use synthetic cards only.

## Inputs
- FRD 01 sections: `FR-LAY-001` (lines 484-488: Layer 3 keyword grants are one-shot modifiers attached on resolution, kept through moves between battlefield zones, cleared on leaving the battlefield, `FR-ZON-004`); `FR-LAY-003` (lines 496-497, `SVY-R39`: one duration covers every temporary effect in its sentence); `FR-TRN-001` Sub-Phase 4c (line 304: "until end of turn" modifiers expire); `FR-TRN-003a` (lines 363-365: an in-resolution free deploy, an explicit zone overrides the printed ones; *Battle Trance* is its example); `FR-ENT-103` (line 264: a unit turned face down keeps its "temporary turn modifiers"); `FR-KWD-002c` (line 599: Stalwart caps combat damage at 1); `FR-TGT-003` (line 397: a target that left its zone before resolution is skipped).
- Survey rows: `docs/design/m2_starter_set_ability_survey.md` line 207 (*Allshan's Battle Trance*: `DEPLOY_FREE` to the Main Theater, `GRANT_KEYWORD` Stalwart, `MODIFY_STAT` +2, both until end of turn, R39).
- Rulings that apply: `SVY-R39`, `SVY-R40` (wording only).
- Design note: `docs/architecture/m2-ability-model.md` (Key types: `Modifier`, the Layers sentence T-0218 wrote).
- Existing code this extends (search done 2026-11-11: no keyword grant, no Layer 3 modifier and no "asset it became" subject exist; the only Layer 3 grant is a card's Cornered keywords, read in `Game.Keywords`):
  - `Modifier`, `Duration`, `Asset.TempPower`, `Asset.expire`, `Game.Keywords` in `pkg/engine/layers.go`; Sub-Phase 4c calls `expire` once in `pkg/engine/turn.go:36`.
  - `stepModifyStat`, `stepDeployFree`, `slotAssets`, `stepAssets`, `resolution` in `pkg/engine/effects.go`; `deployCard` returns the new asset (`pkg/engine/zones.go:43`).
  - `ability.Keyword` (`KeywordNone`, `KeywordAggressive`) and its `String` in `pkg/engine/ability/ability.go:158-172`; `keywordDefs` in `pkg/engine/card.go:50-62`, which already reads Aggressive's name from `ability.KeywordAggressive.String()`.
  - Event kinds in `pkg/engine/events.go:263-292`; T-0211's `EventFaceDown` is the pattern for an effect's own event (ADR-0005 addenda).

## Shape (the lead's plan; the engineer may choose names inside it)
1. **`ability`:** append `OpGrantKeyword` after `OpMorphScratch` (no existing `Op` value moves): "the Subject gains `Step.Keyword` until end of turn (Layer 3, `FR-LAY-001`; expires in Sub-Phase 4c, `FR-LAY-003`)". Append `KeywordStalwart` to `ability.Keyword` and broaden its comment: it names a keyword a trigger comes from (Aggressive) or a step grants (Stalwart); `String` returns "Stalwart" for it. Add `Step.Keyword Keyword` (`OpGrantKeyword` only). Extend `SubjectTarget`'s comment: for an asset step, a pile slot's card that an earlier step of this resolution deployed is the asset it became.
2. **`card.go`:** Stalwart's `keywordDefs` entry reads its name from `ability.KeywordStalwart.String()`, as Aggressive's does, so the name stays written once. Nothing else in this file changes.
3. **`layers.go`:** `Modifier` gains the keywords a Layer 3 modifier grants (an engine `Keywords` value). `Game.Keywords` adds every Layer 3 modifier's keywords after the printed and Cornered ones: a flag is set, a number is added. Write that fold once, walking `keywordDefs`, not field by field, so a later keyword needs no new line here. Update the file's package comment (Layer 3 now also holds granted keywords). `expire` needs no change: one list, one expiry.
4. **`effects.go`:** `stepGrantKeyword` appends one Layer 3 `UntilEndOfTurn` modifier to each asset of `stepAssets` and logs one `EventKeyword` per asset (`Player` -1, `Target`, no `Card`, detail `"<asset> gains <Keyword> until end of turn"` through the asset's `String`, so a face-down asset stays hidden), returning how many it granted. The keyword's `Keywords` value is found through `keywordDefs` by the keyword's name. A keyword with no flag entry there (`KeywordNone`; `KeywordAggressive`, whose trigger comes only from the printed keyword) grants nothing: the step logs nothing and returns 0.
   **"It":** `resolution` records, for each card a step of this resolution deployed, the asset it became (`stepDeployFree` stores `deployCard`'s result). `slotAssets` for a pile slot returns those assets that are still on the battlefield, in pick order. A pile card that wasn't deployed by this resolution (still in its pile, returned to hand, or skipped by the deploy) gives no asset. Do not look an asset up by instance ID on the battlefield: only the deploy of this resolution makes "it".
5. **`events.go`:** `EventKeyword = "keyword"` with a one-line comment, like `EventFaceDown`.

The ADR-0005 addendum for `keyword` is the lead's, at merge. `doc.go`'s shape list (`OpGrantKeyword`) would be a sixth non-test file, so it goes into T-0220's Covers.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/grant_keyword_test.go`. Synthetic Maneuvers built with `maneuver` (`maneuver_test.go:21`) and `enactAction`; the "it" tests can follow `deploy_free_test.go`'s `dfManeuver` / `dfScene` (copy what you need into the new file or add a helper there; don't edit `deploy_free_test.go`).
- `TestFR_LAY_001_GrantedKeywordIsEffective`: a Maneuver granting Stalwart to a chosen allied unit: `g.Keywords(u).Stalwart` is true while `u.Printed().Keywords.Stalwart` stays false; when it attacks an enemy unit with power 3 later the same turn, it takes 1 combat damage from the retaliation (`FR-KWD-002c`; the grant is player 0's on their own turn, so the granted unit is the attacker).
- `TestFR_LAY_001_GrantedKeywordAddsToCornered`: a unit whose card has Cornered keywords, its controller Cornered, granted Stalwart, has both its Cornered keyword and Stalwart (Layer 3 holds both).
- `TestFR_LAY_003_GrantedKeywordExpiresInSubPhase4c`: one Maneuver that grants Stalwart and gives +2 power to the same unit; after the turn ends, the unit has neither (power back to printed, Stalwart false) and is still in its zone (`SVY-R39`, `FR-TRN-001` 4c).
- `TestFR_LAY_001_GrantedKeywordKeptThroughMove`: a granted unit moved from Reserves to the Main Theater (an `OpMoveZone` step or a March) keeps Stalwart until 4c (`FR-LAY-001`).
- `TestFR_ENT_103_FaceDownUnitKeepsGrantedKeyword`: a granted unit then turned face down (`OpMorphScratch`) still has Stalwart, and its Scratch face's Deploys Exhausted (`FR-ENT-103`).
- `TestFR_LAY_001_GrantingAggressiveGrantsNothing`: a step whose `Keyword` is `KeywordAggressive` gives the unit no Aggressive, adds no modifier and logs no `keyword` event (the trigger comes only from the printed keyword).
- `TestEffectEvents_KeywordGrant`: two allied units granted Stalwart by one step: one `keyword` event each, in order, with `Target` the unit, `Player` -1, empty `Card`, and detail `"<unit> gains Stalwart until end of turn"`; for a face-down unit the detail names the Scratch Company, not the card.
- `TestFR_TRN_003a_LaterStepActsOnTheDeployedCard`: a Maneuver whose pile slot is a unit card in its controller's discard pile (printed zone Reserves only), with steps `OpDeployFree` (`To` Main Theater), `OpGrantKeyword` Stalwart and `OpModifyStat` +2 on the same slot: the new asset is in the Main Theater with Stalwart and printed power + 2, and the two events name it.
- `TestFR_TGT_003_LaterStepWithoutTheDeployDoesNothing`: three cases, each with the same later Stalwart and +2 steps on the pile slot: (a) the card left the discard pile before resolution: nothing is deployed and no asset gains anything; (b) the deploy skipped the card (unsupported): no asset gains anything; (c) the first step is `OpReturnToHand`, not a deploy: the card is in hand and no asset gains anything. In each case, an allied unit already on the battlefield is unchanged.
- **Authorized edit** (`pkg/engine/eventkinds_test.go`, `TestEventKinds_ValuesAreTheSchema`): add the row `{EventKeyword, "keyword"}` after `{EventFaceDown, "face-down"}`. In the `test(T-0219)` commit.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none. The tests use only existing helpers (newScene, put, putScratch, inDiscard, give, script, endTurn, attack, wantDamage; maneuver, enactAction, modifyTarget, moveStep, dfUnit, assetOf, effectEvents, inHand).

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/card.go`
- `pkg/engine/layers.go`
- `pkg/engine/effects.go`
- `pkg/engine/events.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'GrantedKeyword|GrantingAggressive|KeywordGrant|LaterStep|FaceDownUnitKeepsGrantedKeyword|EventKinds'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f` passes: no card uses the new op or a later step on a pile slot yet, so both sims are byte-identical to the run branch

## Log
- 2026-11-11, lead-architect (wave F run, session 20): queued as the next wave F deck card in survey row F's order. Writing it fired #137's trigger, so T-0218 (#137's pure move) goes first. Tier `sonnet`: it adds a shape.
- 2026-11-11, lead-architect (wave F run, session 21): written in full against the run branch at 92480d0, and renamed from `T-0219-allshans-battle-trance`. The Covers needed seven non-test files (the five above, plus `cards/path.go` and `cards/cards.go` for the card), so under the D2 size rule the card split off as T-0220 (card, flash, `queued`), with `doc.go`'s shape line and #140 (2)'s test. Five files is one over the rule. Not split further: the op (`ability.go`), its step and "it" (`effects.go`), its fold (`layers.go`) and its event (`events.go`) are one shape, and none works alone; `card.go` is one line, there so the name "Stalwart" stays written once (D2.0). Searched `pkg/engine` and the task archive for an existing equivalent: none (see Inputs). Waiting triggers checked: #15 doesn't fire (a keyword or power expiring in 4c makes nothing lethal), nor #6, #7, #71, #91 (2), #106, #120, #136, #141 (3), #142. Set to `tests`.
- 2026-11-11, rules-test-author: tests written in `pkg/engine/grant_keyword_test.go` (commit `1e879dc`) plus the authorized `{EventKeyword, "keyword"}` row in `TestEventKinds_ValuesAreTheSchema`. Tests: TestFR_LAY_001_GrantedKeywordIsEffective, TestFR_LAY_001_GrantedKeywordAddsToCornered, TestFR_LAY_003_GrantedKeywordExpiresInSubPhase4c, TestFR_LAY_001_GrantedKeywordKeptThroughMove, TestFR_ENT_103_FaceDownUnitKeepsGrantedKeyword, TestFR_LAY_001_GrantingAggressiveGrantsNothing, TestEffectEvents_KeywordGrant, TestEffectEvents_KeywordGrantFaceDownStaysHidden (extra: the face-down detail case, split from the unit-order case), TestFR_TRN_003a_LaterStepActsOnTheDeployedCard, TestFR_TGT_003_LaterStepWithoutTheDeployDoesNothing (subtests CardLeftThePile, DeploySkippedUnsupported, NotADeploy). Helpers requested: none. Cases left out: none. `cardtown lint rules` passes; vet fails only on the not-yet-existing `EventKeyword`, `ability.OpGrantKeyword`, `ability.KeywordStalwart` and `Step.Keyword`. Reuses `dfUnit` from `deploy_free_test.go` and `moveStep` from `move_zone_test.go` (same package). Set to `tests-review`.
- 2026-11-11, lead-architect (wave F run, session 21): tests reviewed (1e879dc), returned to `tests` with one note. The other tests match the Tests list and `FR-LAY-001`, `FR-LAY-003`, `FR-ENT-103`, `FR-TRN-003a`, `FR-TGT-003` and `SVY-R39`; the extra `TestEffectEvents_KeywordGrantFaceDownStaysHidden` is welcome.
  1. `TestEffectEvents_KeywordGrant` expects the wrong order. It builds `want` from `s.g.Assets()`, which lists Reserves before the Main Theater, so `want` is [Sentry, Rider]. A `SubjectAll` step acts on `assetsByID()`, oldest first (`stepAssets`' comment, `effects.go`), so a correct implementation logs Rider (put first) then Sentry, and the test fails. Build `want` from the two assets `s.put` returns, in put order (`[]*Asset{rider, sentry}`), and drop the `Assets()` loop. Change nothing else.
- 2026-11-11, rules-test-author: note 1 fixed in `877b0b7`: TestEffectEvents_KeywordGrant now builds its expected order from the two assets s.put returns, [rider, sentry], not from s.g.Assets(). Nothing else changed. Set to `tests-review`.
- 2026-11-11, lead-architect (wave F run, session 21): tests re-reviewed (877b0b7): note 1 is fixed as asked and nothing else changed. Tests approved; set to `ready`.
- 2026-11-11, kernel-engineer: implemented (feat commit below). Searched `pkg/engine` for a keyword grant / "it" subject first: none, so new. Diff: `ability.go` (`OpGrantKeyword`, `KeywordStalwart` + `String`, `Step.Keyword`, comments; gofmt realigned the const/struct comments); `card.go` (Stalwart name from `ability.KeywordStalwart.String()`); `layers.go` (`Modifier.Keywords`, `Game.Keywords` folds Layer 3 modifiers through `Keywords.add`, which walks `keywordDefs`; package comment); `effects.go` (`resolution.became` per pile slot, filled by `stepDeployFree` for `SubjectTarget`, appended in `slotAssets` and filtered to those still on the battlefield; `stepGrantKeyword`, flag lookup via `keywordDefs`, nothing for Aggressive/None); `events.go` (`EventKeyword`). Commands: the listed `go test -run ...` ok; `go tool task check` PASS; `go tool task sim` clean; `go tool task sim-compare -- autoloop/m2-wave-f` PASS byte-identical. Look at: `slotAssets` clones `chosen[i]` before appending `became[i]`.
- 2026-11-11, lead-architect (wave F run, session 21): code reviewed (e606da7). Only the five Allowed Files changed, and no test changed after 877b0b7. No card name in engine code. `OpGrantKeyword` is appended, so no `Op` value moves. The fold (`Keywords.add`) walks `keywordDefs` once, and the grant finds its keyword there by the name `ability.Keyword` holds, so "Stalwart" is written once. One `Modifiers` list and one `expire`, as #137 built. "It" is `resolution.became`, filled only by `stepDeployFree`, so no battlefield lookup by instance exists. Gates re-run in a review worktree: `go tool task check` passes, `go tool task sim` is clean (Animus 50/50 executable, P0 71.5% / P1 27.5% / draws 1.0%), and `go tool task sim-compare -- autoloop/m2-wave-f` is byte-identical. Approved.
