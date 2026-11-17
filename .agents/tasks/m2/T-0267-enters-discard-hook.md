---
id: T-0267
title: "HookEntersDiscard: a pile trigger that fires whenever its card is put into its owner's discard pile, from any zone, fired in putCard (wave H shape, sonnet)"
kind: shape
tier: sonnet
wave: H
status: tests
branch: m2/autoloop-wave-h/T-0267-enters-discard-hook
depends_on: []
claims: [hook:HookEntersDiscard (this card was put into its owner's discard pile from any zone, fired by putCard for every card it puts into a discard pile; works only with ActiveZone ZoneDiscard)]
cards: []
requirements: [FR-ZON-004, FR-ZON-005, FR-ZON-002, FR-ZON-003, FR-ENT-102, FR-TRG-003, FR-MNV-001, FR-KWD-002d]
---

## Goal
A triggered ability can say "When this card is put into your discard pile" (`FR-ZON-004` Enter-Discard Triggers, survey gap G1): a new hook, `HookEntersDiscard`, fires from the card's owner's discard pile each time the card is put there, by any route: a discard from hand (step or cost), a destruction on the battlefield (death sweep, kill step, kill cost, combat), a face-down Scratch Company whose card is revealed as it goes there (`FR-ZON-005`), or a Maneuver resolving (`FR-MNV-001`). The trigger is the owner's (`FR-ZON-002` Destruction, `FR-ZON-003`), ranked after battlefield triggers by its card's position in the pile (`SVY-R7`, directive #136), and may choose targets when placed, like any pile trigger.

*Solarheart Monk*, the card, is T-0268.

## Inputs
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): `FR-ZON-004` (line 261): Enter-Discard Triggers (`:267`: "Evaluated against the card's revealed face-up printed characteristics upon entering the Discard pile ... Tucked or destroyed Scratch Companies entering Discard face-up do trigger printed enter-discard abilities"), Self-Relocating Trigger Tracking (`:268`); `FR-ZON-005` Departure to a Public Zone (`:272`); `FR-ZON-002` Destruction (`:259`, the owner's discard pile); `FR-ZON-003` (line 230); `FR-ENT-102` Host Departure & Discard Routing (`:278`, tucked cards trigger *Solarheart Monk*; no Attached zone yet, directive #142); `FR-TRG-003` (line 466); `FR-KWD-002d` Heal N (`:622`).
- Survey: `docs/design/m2_starter_set_ability_survey.md` gap G1 (`:47`) and the card row (`:210`: "enters discard → `HEAL 2` chosen asset ... G1; any route into discard").
- Design note: `docs/architecture/m2-ability-model.md` Deliverables, wave H (`:201`, `:210`: "enter-discard from any zone (*Solarheart Monk*, survey gap G1)").
- Rulings that apply: `SVY-R7` (pile sources ranked after battlefield ones, by pile position), `SVY-R30` (a mandatory trigger is placed whatever it finds). No new ruling needed: `FR-ZON-004` `:267` settles that every route into the pile fires it, the Scratch Company case included.
- Code of the day (run branch at d5898f8):
  - `pkg/engine/ability/ability.go`: the `Hook` const block (`:187-198`), `HookDiscardedFromHand` (`:196`); `Trigger.ActiveZone` (`:258-260`).
  - `pkg/engine/trigger.go`: `pileSource` (`:69-72`), `activeIn` (`:76-81`), `fireDiscarded` (`:83-87`), `fireHookFor` (`:146`); `placeTrigger` sets `Tracked` for every `ActiveZone: ZoneDiscard` trigger (`:288`).
  - `pkg/engine/zones.go`: `discardCards` (`:122-132`, fires the from-hand hook after `putCard`), `putCard` (`:149-167`). Every card that reaches a discard pile goes through `putCard(inst, ZoneDiscard, false)`: `discardCards` (`zones.go:128`), `destroyAll` (`sweep.go:20`, every destruction, a face-down Scratch Company's card included) and `resolveManeuver` (`stack.go:148`). The only other `putCard` callers put into a deck or a hand (`effects.go:142`, `:352`, `:358`, `freedeploy.go:140`, `setup.go:91`, `zones.go:20`).
  - `pkg/engine/doc.go`: the hooks list (`:18-20`).
  - Test setup that calls `putCard(…, ZoneDiscard, …)`: `scene.inDiscard` (`scenario_test.go:85-93`), `pile_pick_test.go:196-197`, `pile_target_test.go:232-233`, `zones_test.go:66`; all with synthetic cards that print no `HookEntersDiscard` trigger, so none of them changes behavior.

## Shape (the lead's plan)
- `pkg/engine/ability/ability.go`: `HookEntersDiscard` after `HookTurnStart`, commented: this card was put into its owner's discard pile, from any zone (a hand, the battlefield, the stack), by any route (`FR-ZON-004`, `FR-ZON-005`); fires only triggers with `ActiveZone` `ZoneDiscard`. `HookDiscardedFromHand`'s comment gains that it is the narrower "from your hand" gate (*Eerie Remnant*), and that a discard from hand fires both hooks.
- `pkg/engine/trigger.go`: `fireEntersDiscard(inst)`, beside `fireDiscarded`: `g.fireHook(pileSource(g, inst), ability.HookEntersDiscard)`. The pile source already reads the card's printed face (a face-down Scratch Company's real card), its owner as controller and its pile position.
- `pkg/engine/zones.go`: `putCard`'s `ZoneDiscard` case calls `g.fireEntersDiscard(inst)` after the append, and its comment says so: this is the one place a card enters a discard pile, so every route fires the hook, including routes added later (the Attached purge of #142). No caller changes. `discardCards`' order stays: its `putCard` fires the enter-discard hook, then it logs and fires the from-hand hook.
- `pkg/engine/doc.go`: `HookEntersDiscard` in the hooks list.
- Not changed: `sweep.go` and `stack.go` (they already put through `putCard`), `fireDiscarded` and `HookDiscardedFromHand` (see the Log on why this is not a duplicate). Not this task: the card (T-0268).

## Tests (shape tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/enters_discard_test.go`. Cards are synthetic. Write a local helper `entersTrigger(steps ...ability.Step) ability.Trigger` returning `{Hook: ability.HookEntersDiscard, ActiveZone: ability.ZoneDiscard, Effect: ability.Effect{Steps: steps}}` (mandatory), and a variant or second helper whose `Effect.Targets` is one slot with an empty `Filter` and whose step is `{Op: ability.OpHeal, Amount: ability.Value{N: 2}, Subject: ability.SubjectTarget, Target: 0}` ("chosen asset heals 2", as *Spoil Forest Colossus* writes it, `cards/animus.go:690-707`). Reuse `trigUnit`, `drawN`, `discardN`, `killOnP1`, `s.enactWith`, `placeSeq`, `pileLogPlaces`, `askedKind`, `byKind`, `inZone`, `onField`, `maneuver`, `steps`, `s.put`, `s.putScratch`, `s.give`, `s.do`, `enactAction` (all exist). Player 0 is the active player.

- `TestFR_ZON_004_EntersDiscardFiresFromHandDiscard`: player 0's hand holds only a unit whose trigger is `entersTrigger(drawN(1))`, and player 0 has cards in deck. A free Maneuver discards 1 (`s.enactWith(…, discardN(1))`). The card is in player 0's discard pile, its trigger was placed once (`pileLogPlaces` 1) and resolved: player 0's hand gained one card from the deck.
- `TestFR_ZON_004_EntersDiscardFiresOnDestruction`: two scenes. (a) Combat: `killOnP1` on player 1's unit with `entersTrigger(drawN(1))`: the card is in player 1's discard pile, the trigger was placed once, and player 1 drew 1. (b) A kill step: a free Maneuver of player 0 with `{Op: ability.OpKill, …}` on an allied unit carrying the same trigger: in player 0's discard pile, placed once, player 0 drew 1.
- `TestFR_ZON_004_EntersDiscardTriggerChoosesItsTarget`: player 1's unit with the targeted heal trigger is destroyed by combat (`killOnP1`), while player 1 has a second unit in the Main Theater with `Damage` 3 and Health 5. Script player 1's target pick to that unit (or assert the one-option fast path if only one asset can be chosen; make the scene have at least two assets so the pick is asked). After resolution its `Damage` is 1 (`FR-KWD-002d`, `FR-TGT-001`).
- `TestFR_ZON_005_EntersDiscardFiresForADestroyedScratchCompany`: `s.putScratch(1)`, then set its `Inst.Card` to a unit printing `entersTrigger(drawN(1))` (and also a `HookKilled` trigger `drawN(1)`), moved to or attacked in the Main Theater as the existing face-down tests do, and destroyed by player 0's attacker. The card is in player 1's discard pile, face-up; the enter-discard trigger was placed once and player 1 drew exactly 1 (the face-down unit's death trigger fires nothing, `FR-ZON-004` `:265`; the enter-discard one does, `:267`).
- `TestFR_ZON_004_EntersDiscardFiresWhenAManeuverResolves`: a costless Maneuver card of player 0 with `steps(drawN(1))` and the trigger `entersTrigger(drawN(1))`, enacted from an otherwise empty hand: after the stack is empty the card is in player 0's discard pile and player 0 drew 2 (its effect's 1, then its pile trigger's 1).
- `TestFR_ZON_003_EntersDiscardIsTheOwnersTrigger`: player 1's unit with `entersTrigger(drawN(1))` in the Main Theater with `Controller` set to 0 (a control change, `FR-ZON-002`), killed by a free Maneuver of player 0 with `OpKill` on an allied unit: the card is in player 1's discard pile, not player 0's; player 1 drew 1 and player 0 drew nothing from it.
- `TestFR_ZON_004_EntersDiscardFiresOnlyIntoTheDiscardPile`: three scenes, each with a unit printing `entersTrigger(drawN(1))`: (a) returned from the battlefield to its owner's hand by an `OpReturnToHand` step, (b) put on its owner's deck by `OpPutOnDeck`, and (c) a card whose trigger has the hook but no `ActiveZone` (a battlefield trigger), destroyed by combat. In each, no trigger is placed (`pileLogPlaces` 0) and nobody draws (G1, `activeIn`).
- `TestFR_TRG_003_EntersDiscardTriggersFollowThePileOrder`: player 1's two units "Monk A" and "Monk B", each printing `entersTrigger(drawN(1))`, both destroyed by one event (a player 0 `effectSource` or Maneuver step `OpKill` with `SubjectAll` over player 1's units, or one combat that leaves both lethal at the same sweep). Player 1 is asked one `DecideOrder` with both triggers, whose option order is the order the two cards entered player 1's discard pile (read the pile after the kill), and by default (a timeout) the one lower in the pile is placed first (`placeSeq`, `SVY-R7`), as `TestFR_TRG_003_PileTriggersFollowTheDiscardOrder` does for the from-hand hook.
- `TestFR_ZON_004_DiscardFromHandFiresBothHooks`: a unit printing two pile triggers, `pileTrigger`-style `HookDiscardedFromHand` with `drawN(1)` (mandatory) and `entersTrigger(drawN(1))`, discarded from hand by `discardN(1)`: both are placed and player 0 drew 2.

Existing tests stay as they are and must still pass (`pile_trigger_test.go`'s `TestFR_ZON_004_OnlyADiscardFromHandFiresIt` included: it uses only the from-hand hook).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- Change `scene.inDiscard` (`scenario_test.go:85-93`) to append each instance to `s.g.Players[p].Discard` directly instead of calling `s.g.putCard`, with its comment saying it is setup and fires no enter-discard trigger. Otherwise a card task that seeds a discard pile with an enter-discard card (T-0268, a target of *Mutation Swell*'s tests) would buffer a trigger from setup.

## Allowed Files
- `pkg/engine/ability/ability.go` (`HookEntersDiscard`, `HookDiscardedFromHand`'s comment)
- `pkg/engine/trigger.go` (`fireEntersDiscard`)
- `pkg/engine/zones.go` (`putCard` fires it, and its comment)
- `pkg/engine/doc.go` (the hooks list)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "EntersDiscard|FiresBothHooks|PileTrigger|DiscardedFromHand"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (no starter or stress card prints `HookEntersDiscard`, so every log is byte-identical)

## Log
- 2026-11-17 lead-architect (wave H run, session 29): written against the run branch at d5898f8. *Solarheart Monk* is next in the wave H Deliverable's card order (design note `:201`, `:210`), after *Mutation Swell* (T-0266, done). No ruling needed: `FR-ZON-004` `:267` says enter-discard triggers fire for any card entering the pile face-up, destroyed or tucked Scratch Companies included, and `FR-ZON-002` `:259` sends a destroyed asset to its owner's pile, whose triggers `pileSource` already gives the owner. Reuse: the pile-trigger machinery of T-0225 (`pileSource`, `activeIn`, `Tracked`, the `SVY-R7` ranking) carries the new hook unchanged; only its firing point is new. Placement: `putCard` is the one function that puts a card into a discard pile (searched every non-test `putCard` call and every `Discard = append` in `pkg/engine`), so firing there covers every route without touching `sweep.go` or `stack.go`, and the tucked-card purge of #142 will fire it with no further change. **Not a duplicate of `HookDiscardedFromHand`:** that hook is the narrower printed gate "from your hand" (*Eerie Remnant*, `FR-ZON-004` `:267`'s "explicit condition gates"), fired at the one hand-discard point; the two share `pileSource` and `fireHook`, so no rule about how a pile trigger fires is written twice. Merging them would need `putCard` to know where a card came from, which nothing else needs. Searched the task archive for an enter-discard hook: only T-0225's from-hand one. Split from the card under the D2 size rule (shape and card touch six non-test files); T-0268, the card, queued in this commit. Four non-test files. Directives: none `open`; the `waiting` ones were checked against the repo and none has fired (#6, #7: mode is `serial`; #15 (2): no draw or end-of-turn hook here; #71: no each-player effect; #91 (2), #106, #120 and #142 name cards or shapes no task touches yet; #142's purge will reuse this hook when it comes). `main` has no commits the run branch lacks. Status `tests`.
