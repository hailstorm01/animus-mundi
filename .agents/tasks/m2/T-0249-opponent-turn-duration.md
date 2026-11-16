---
id: T-0249
title: "A temporary effect that lasts until the end of your opponent's turn (wave H shape, sonnet)"
kind: shape
tier: sonnet
wave: H
status: tests
branch: m2/autoloop-wave-h/T-0249-opponent-turn-duration
depends_on: []
claims: [shape:full-round duration (ability.Duration on Step.Until: UntilEndOfTurn, the zero value, or UntilEndOfOpponentTurn, which expires in Sub-Phase 4c of a turn whose active player is an opponent of the ability's controller; read by OpModifyStat and OpGrantKeyword through one Modifier constructor)]
cards: []
requirements: [FR-LAY-001, FR-LAY-003, FR-TRN-001]
---

## Goal
The full-round duration as a shape: a step that makes a temporary modifier (`OpModifyStat`, `OpGrantKeyword`) can say it lasts *"until the end of your opponent's turn"* (`FR-LAY-003` Full-Round Durations). Such a modifier is still there for the rest of the current turn and through the opponent's next turn, and it expires in Sub-Phase 4c of that opponent's End Phase. Sub-Phase 4c expires it only when the turn ending is a turn of an opponent of the controller of the ability that made it (`FR-TRN-001` 4c: *"or 'until the end of your opponent's turn' when the current turn is that opponent's turn"*). So one made during the opponent's own turn (by a trigger) expires at the end of that same turn. "Until end of turn" stays the zero value and behaves as today.

The duration type moves from package `engine` to package `ability`, where `Step` can name it, as a pure move: `engine.Duration` and `engine.UntilEndOfTurn` are removed, not kept as a second copy. `Modifier.Until` becomes an `ability.Duration`.

*Blindside*, the card, is T-0250 (split off under the D2 size rule: shape and card together touch six non-test files).

## Inputs
- FRD 01: `FR-LAY-003` Full-Round Durations (`docs/frd/01_Game_Rules_FRD.md:514`) and Duration Reach (`:513`, `SVY-R39`: a duration covers every temporary effect in its sentence, so both modifier-making ops read the one field); `FR-TRN-001` Sub-Phase 4c (`:317`); `FR-LAY-001` Layer 5 and One-Shot Modifier Snapshot Locking (`:494`, `:504`): a modifier stays on its asset for its full duration while the asset stays on the battlefield, and leaving the battlefield clears it (`FR-ZON-004`, already built).
- Survey: *Blindside* row (`docs/design/m2_starter_set_ability_survey.md:193`, "Full-round duration"). Design note Deliverable 6m, "From wave C" (`docs/architecture/m2-ability-model.md:207`): through `layers.go`'s `Modifier` and `Duration`.
- Code of the day (run branch at 7baec50):
  - `Duration`, `UntilEndOfTurn`, `Modifier`, `(*Asset).expire` (`pkg/engine/layers.go:14-19`, `:24-29`, `:43`).
  - `endTurn` 4c (`pkg/engine/turn.go:35-48`): `a.expire(UntilEndOfTurn)` for every asset.
  - `stepModifyStat` (`pkg/engine/effects.go:257`) and `stepGrantKeyword` (`:470`): each builds its own `Modifier{…, Until: UntilEndOfTurn}` and logs "until end of turn".
  - `Step` (`pkg/engine/ability/ability.go:385`); `OpModifyStat` (`:360`) and `OpGrantKeyword` (`:366`) comments say "until end of turn"; `Keyword.String` (`:172`) is the model for a `String` method.
  - `StackItem.Controller` (`pkg/engine/stack.go:32`), reached as `r.it.Controller` in a step.
- No test names `UntilEndOfTurn` or `Duration` (searched `pkg`, `cmd`). Tests build `Modifier{Layer: 5, Power: n}` literals with no `Until`; those keep compiling and keep meaning "until end of turn".

## Shape (the lead's plan)
- `pkg/engine/ability/ability.go`:
  - `type Duration int`, commented `how long a temporary effect a step makes lasts (FR-LAY-003)`, with `UntilEndOfTurn Duration = iota` (the zero value: expires in Sub-Phase 4c of the current turn) and `UntilEndOfOpponentTurn` ("until the end of your opponent's turn": expires in Sub-Phase 4c of a turn whose active player is an opponent of the ability's controller, FR-LAY-003, FR-TRN-001).
  - `func (d Duration) String() string`: `"until end of turn"` and `"until the end of its controller's opponent's turn"`.
  - A `Step` field `Until Duration // OpModifyStat, OpGrantKeyword: how long the modifier lasts; the zero value is until end of turn`. Update the `OpModifyStat` and `OpGrantKeyword` comments to say "for Step.Until" instead of "until end of turn … expires in Sub-Phase 4c".
- `pkg/engine/layers.go`:
  - Remove `type Duration` and its const block.
  - `Modifier.Until ability.Duration`, and a new field `By PlayerID`: the controller of the ability whose step made it (read only for `UntilEndOfOpponentTurn`).
  - `expire(d Duration)` becomes `expireAtEndOf(active PlayerID)`: it removes every modifier that lasts `UntilEndOfTurn`, and every `UntilEndOfOpponentTurn` modifier whose `By` is not `active`. One function decides what Sub-Phase 4c expires.
- `pkg/engine/turn.go`: 4c calls `a.expireAtEndOf(g.Active)`; its comment names both durations.
- `pkg/engine/effects.go`: one constructor `func (r *resolution) modifier(st ability.Step, m Modifier) Modifier` (or an equivalent single place) sets `m.Until = st.Until` and `m.By = r.it.Controller`. `stepModifyStat` and `stepGrantKeyword` both build their modifier through it and log `st.Until.String()` instead of the literal "until end of turn". So the log text for every existing card is byte-for-byte what it is today.
- No new `Op`, no new event kind: `EventPower` and `EventKeyword` are unchanged.

## Tests (shape tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/opponent_turn_duration_test.go`. Cards are synthetic. Player 0 is the active player. Use `maneuver` (`maneuver_test.go:21`), `enactAction` and `s.give` to enact a free Maneuver; `modifyTarget` (`cornered_test.go:34`) shows the step shape, but these steps also set `Until: ability.UntilEndOfOpponentTurn`. `s.endTurn()` (`scenario_test.go:121`) ends the active player's turn. Read power with `s.g.Power` and keywords with `s.g.Keywords`.

- `TestFR_LAY_003_OpponentTurnPowerLastsThroughOpponentsTurn`: player 0 enacts a Maneuver whose one target is an enemy unit (printed power 4) with `{Op: OpModifyStat, Amount: -3, Subject: SubjectTarget, Target: 0, Until: UntilEndOfOpponentTurn}`. Its power is 1 right after; still 1 after player 0's turn ends (player 1 is active, in the Main Phase); 4 after player 1's turn ends; still 4 after player 0's next turn ends.
- `TestFR_LAY_003_OpponentTurnPowerOnAnAlliedUnit`: the same duration with `+2` on an allied unit (printed power 1): 3 through player 1's turn, 1 after it. "Your opponent" is the ability controller's opponent, not the asset controller's: this unit's controller is player 0 too, so the case can't tell them apart; the next test does.
- `TestFR_LAY_003_OpponentTurnIsTheAbilityControllersOpponent`: player 0's Maneuver gives an enemy unit (controlled by player 1) +2 power `UntilEndOfOpponentTurn`. It expires at the end of player 1's turn (the ability controller's opponent's turn), not at the end of player 0's turn (which would be the asset controller's opponent's turn) and not at the end of player 0's next turn.
- `TestFR_TRN_001_OpponentTurnModifierMadeOffTurnExpiresThatTurn` (`FR-TRN-001` 4c, "when the current turn is that opponent's turn"): player 0 controls a synthetic structure in the Main Theater with one `HookTurnStart` trigger, `Turn: TurnAny` (`startStructure`, `startSteps`, `turn_start_trigger_test.go:19`, `:24`), whose one step gives an enemy unit (an untargeted `SubjectAll` step with an enemy-unit filter) `-1` power `UntilEndOfOpponentTurn`. After player 0's turn ends, the trigger resolves at the start of player 1's turn, so its controller (player 0) is not the active player: the enemy unit has power printed − 1 during player 1's Main Phase. After player 1's turn ends, player 0's turn has begun and the trigger has made a second modifier at its start: power is printed − 1, not − 2, because the first one (made during player 1's turn) expired in that turn's 4c. The enemy unit has exactly one `Modifier` then.
- `TestFR_LAY_003_OpponentTurnKeywordGrant` (`SVY-R39`: the duration covers a keyword grant too): player 0's Maneuver gives an allied unit Stalwart (`OpGrantKeyword`, `Keyword: KeywordStalwart`) `UntilEndOfOpponentTurn`: it has Stalwart during player 1's turn, and not after player 1's turn ends.
- `TestFR_LAY_003_UntilEndOfTurnUnchanged`: a step with no `Until` (the zero value) still expires at the end of the turn it was made in, for both `OpModifyStat` and `OpGrantKeyword`, and `ability.Duration(0) == ability.UntilEndOfTurn`.
- `TestFR_LAY_003_MixedDurationsExpireSeparately`: one unit gets `+1 UntilEndOfTurn` and `+2 UntilEndOfOpponentTurn` from player 0's Maneuvers. After player 0's turn ends its power is printed + 2; after player 1's, printed.
- `TestFR_LAY_003_OpponentTurnModifierLeavesWithTheAsset` (`FR-ZON-004`): an enemy unit with a `-3 UntilEndOfOpponentTurn` modifier returns to its owner's hand (a synthetic `OpReturnToHand` step in a second Maneuver) during player 0's turn; when redeployed it has printed power. (If this needs more than one free Maneuver and a deploy helper, check the asset's `Modifiers` is empty on its card instead.)
- Log text: one test (in any of the above, or its own `TestFR_LAY_003_OpponentTurnLogText`) checks the `EventPower` event's log line ends in `until the end of its controller's opponent's turn`, and an until-end-of-turn step's still ends in `until end of turn`. Use the event-log reader other tests use (`effectEvents`, `effect_events_test.go:15`) or the game's log if events don't carry text; name the helper you need under Helpers Needed otherwise.

Existing tests stay as they are and must still pass (the move of `Duration` is pure, and "until end of turn" behaves as before).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none expected. If the test author needs one, it names it here.

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/layers.go`
- `pkg/engine/turn.go`
- `pkg/engine/effects.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "OpponentTurn|UntilEndOfTurnUnchanged|MixedDurations|Expire"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (no card uses `UntilEndOfOpponentTurn` yet, and the `Duration` move is pure)

## Log
- 2026-11-15 lead-architect (wave H run, session 10): *Blindside* is the next wave H card (wave C's, after wave B's three are `done`). Its survey row needs the full-round duration, which the kernel lacks: `layers.go` has only `UntilEndOfTurn`, and 4c expires only that. Searched `pkg/engine` and the task archive for an existing opponent's-turn duration: none (`grep -rn "OpponentTurn\|UntilEnd" pkg`). The duration moves to `ability` rather than getting a second enum there, so nothing is written twice; `By` records the ability controller, and one `expireAtEndOf` reads it, which matches 4c's wording directly (no turn counting). No ruling needed: `FR-LAY-003` and `FR-TRN-001` 4c settle both the on-turn and the off-turn case. Split the card off as T-0250 (`queued`) under the D2 size rule. Re-read `BOARD.md` before committing: no `open` directive; #15 (3) waits for a max-health modifier (*Poisoned Rations*), which this isn't. Status `tests`.
