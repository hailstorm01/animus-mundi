---
id: T-0010
title: Add attack triggers (with Aggressive in the same window), an "on your turn" gate, and an opponents subject
kind: shape
tier: sonnet
wave: A1 (board directive #4, task 4, the last A1 shape task)
status: review
branch: m2/T-0010-attack-turn-opponents
depends_on: [T-0009]
claims: [hook:attacks, type:ability.TurnOwner, gate:turn-owner, subject:opponents, func:attack-step-2-flush]
cards: []
requirements: [FR-CBT-002, FR-KWD-002f, FR-WIN-001]
---

## Goal
Three shapes finish the A1 vocabulary:
- **Attack hook:** "Whenever this unit attacks, …" fires at combat Step 2. Its triggers go on the stack above the Attack in the same flush as **Aggressive N**, so its controller orders them, and they resolve before combat damage (`FR-CBT-002`).
- **Turn gate:** "… on your turn" fires only during the trigger controller's turn.
- **Opponents subject:** a step can act on each opponent ("all opponents draw 2 cards").

After this task, every A1 card can be written from existing shapes.

## Inputs
- FRD 01:
  - `FR-CBT-002` (lines 455–467 of `docs/frd/01_Game_Rules_FRD.md`), Step 2: attack triggers, including Aggressive, enqueue on top of the Attack in APNAP order and resolve before it.
  - `FR-KWD-002f` (Aggressive N).
  - `FR-TRG-003` (APNAP placement).
  - `FR-WIN-001` (a required draw from an empty deck loses at the next sweep).
- Card text for context (the A1 card tasks use these; tests use synthetic cards):
  - *Despoiler Larva*: "Whenever this unit attacks, Scratch 1."
  - *Portent Scrivener*: "When this unit is killed on your turn, draw a card."
  - *Despoiler Worm*: "When this unit is killed, all opponents draw 2 cards."
- Built on: T-0007 to T-0009 (`trigger.go`: `fireHook`, the buffer, `flushTriggers`, `resolveAbility`). Today `declareAttack` in `actions.go` pushes a `StackAggressive` item directly.
- Rulings that apply: none.

### Vocabulary (`pkg/engine/ability/ability.go`)
```go
const (
	HookDeploy Hook = iota + 1
	HookKilled
	HookAttacks // this asset attacked (combat Step 2, FR-CBT-002)
)

// TurnOwner gates a trigger on whose turn it is when its event happens.
type TurnOwner int

const (
	TurnAny   TurnOwner = iota // the zero value: no gate
	TurnYours                  // only during the trigger controller's own turn
)

// Trigger gains:
	Turn TurnOwner

const (
	SubjectController Subject = iota
	SubjectSelf
	SubjectOpponents // each opponent of the ability's controller (OpScratch, OpDraw)
)
```

### Lead decisions (technical, not rules)
- **Step 2 window:** `declareAttack` places the Attack and then buffers the attacker's Step 2 triggers. Those are its Aggressive keyword trigger, if it has one, followed by its printed `HookAttacks` triggers in printed order (collection order, directive #10). Then it flushes, before the stack starts resolving. The flush is T-0007's `flushTriggers`, so a player with two or more Step 2 triggers is asked one `DecideOrder`. A player with one is not asked.
- **Aggressive keeps its behavior:** it still resolves as today, with +N power until end of turn and a fizzle if the attacker has left. Only its placement changes, through the buffer. Existing Aggressive tests must still pass unchanged. The buffer may carry a ready-made stack item for keyword triggers next to `ability.Trigger` for printed ones. Choose the smallest change to `pendingTrigger`.
- **Turn gate:** checked when the event fires (in `fireHook`). `TurnYours` fires only if `g.Active` is the trigger's controller at that moment. For a death trigger, that is the asset's last controller on the battlefield (T-0008). The gate is not checked again at resolution: a trigger fired this turn resolves this turn.
- **Opponents:** `OpScratch` and `OpDraw` with `SubjectOpponents` are performed by each opponent of the ability's controller, in turn order from the controller's opponent. A required draw from an empty deck makes that opponent lose at the next sweep (`FR-WIN-001`). If both players have drawn from empty decks by then, it is a Match Draw, as the existing sweep handles. `OpReturnToHand` and `OpPutOnDeck` with `SubjectOpponents` do nothing, and no card should build them.
- **Not in this task:** opposing-attack hooks ("whenever an opposing unit attacks", Blind Defile), exhaust hooks, "attacks and kills" (Eternal Watchman), and `Cornered` gates. Their waves add them, and the Step 2 flush already places a non-active player's triggers above the active player's.

### What the tests can observe
- **Before combat:** an attack trigger that draws from an empty deck makes its controller lose at that trigger's sweep, before the Attack resolves, so the defender takes no combat damage. A `Scratch` attack trigger adds a face-down Scratch Company before combat.
- **Ordering:** a unit with Aggressive N and a printed attack trigger gives its controller one `DecideOrder` with 2 options. A unit with only one of them is not asked.
- **Aggressive alone:** it still gives +N power to the combat as before.
- **Turn gate:** a `TurnYours` death trigger fires when its controller's unit dies on its controller's turn (it attacks and dies). It does not fire when the unit dies on the opponent's turn (it is attacked and dies). An ungated twin fires both times.
- **Opponents:** a death trigger "all opponents draw 2" makes the opponent draw two named top cards, and the controller draws nothing. With the opponent's deck empty, the opponent loses at that trigger's sweep.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
All in `pkg/engine/attack_trigger_test.go` (local helpers: attackTrigger, striker, opponentsDraw, gatedKilled).
- `TestFR_CBT_002_AttackTriggerResolvesBeforeCombatDamage` (empty-deck Draw attack trigger loses before the HQ takes damage; also FR-WIN-001)
- `TestFR_CBT_002_AttackTriggerScratchHappensBeforeCombat`
- `TestFR_CBT_002_AttackTriggerFiresOncePerAttack`
- `TestFR_CBT_002_OnlyTheAttackerFiresItsAttackTrigger`
- `TestFR_CBT_002_AggressiveAndAttackTriggerAreOrderedTogether` (one DecideOrder, 2 options)
- `TestFR_CBT_002_AggressiveAloneIsNotAskedAbout`
- `TestFR_CBT_002_AttackTriggerAloneIsNotAskedAbout`
- `TestFR_CBT_002_TwoPrintedAttackTriggersAreOrderedTogether`
- `TestFR_CBT_002_OrderAnswerDecidesWhichPrintedAttackTriggerResolvesFirst`
- `TestFR_KWD_002f_AggressiveStillGivesPowerWithAnAttackTrigger`
- `TestFR_KWD_002f_AggressiveAloneStillGivesPowerForTheCombat`
- `TestFR_WIN_001_OpponentsDrawFromEmptyDeckLosesAtThatSweep`
- `TestFR_WIN_001_ControllersOwnEmptyDeckIsUnaffectedByOpponentsDraw`
- `TestShape_OpponentsDrawIsPerformedByTheOpponentNotTheController`
- `TestShape_OpponentsScratchIsPerformedByTheOpponent`
- `TestShape_TurnYoursDeathTriggerFiresOnItsControllersTurn`
- `TestShape_TurnYoursDeathTriggerDoesNotFireOnTheOpponentsTurn`
- `TestShape_UngatedDeathTriggerFiresOnEitherTurn`
- `TestShape_TurnYoursAttackTriggerFiresWhenAttackingOnYourTurn`

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/ability/ability.go` (the vocabulary above)
- `pkg/engine/trigger.go` (`HookAttacks` firing, the turn gate in `fireHook`, `SubjectOpponents` in `resolveAbility`, keyword-trigger support in the buffer)
- `pkg/engine/actions.go` (`declareAttack`: buffer Aggressive and the attack hook, then flush)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_CBT_002|TestFR_KWD_002f|TestFR_WIN_001|TestFR_TRG'` (rules-test-author names the new tests to match)
- [ ] All earlier tests pass unchanged, including the existing Aggressive and combat tests
- [ ] `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean
- [ ] `go run ./cmd/cardtown lint --all` passes
- [ ] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0, and plain `sim` output, **sorted**, is unchanged from the integration branch (directive #12)

## Log
- 2026-10-20 lead-architect: Wrote the task (board directive #4, task 4, the last A1 shape task). Aggressive moves into the Step 2 buffer because `FR-CBT-002` names it among the Step 2 triggers placed in APNAP order, so a unit with Aggressive and a printed attack trigger lets its controller order them. Set `status: tests` for `rules-test-author`.
- 2026-10-20 rules-test-author: Wrote 19 tests in `attack_trigger_test.go` (listed under Tests). No helpers requested. TestShape_* prefix used for the turn gate and opponents-subject tests, since no FRD 01 requirement defines them (not traced by lint). `go vet` fails only on the not-yet-existing new vocabulary (`ability.TurnOwner` etc.); `cardtown lint rules` passes. Left out: relative order of Aggressive vs a printed trigger (both resolve before damage, so unobservable without a ruling); cross-player APNAP for attack triggers (needs opposing-attack hooks, out of scope); both-empty Match Draw via opponents draw (existing sweep tests cover it); opposing/exhaust hooks, attacks-and-kills, Cornered. Status: tests-review.
- 2026-10-20 lead-architect: Test review: **approved**, status `ready`. Checked against FR-CBT-002 Step 2, FR-KWD-002f, FR-WIN-001 and the lead decisions:
  - The Step 2 tests show that triggers resolve before damage: an empty-deck loss leaves the HQ undamaged.
  - The expected resolution for answers `[0 1]` and `[1 0]` is right under placement order.
  - The Aggressive arithmetic and its end-of-turn expiry are right.
  - The turn gate is checked on both turns, with an ungated control.
  - Against today's code, the opponents-draw and turn-gate tests fail, because the controller draws or the trigger fires anyway, so they can't pass without the implementation.
  - `OnlyTheAttackerFiresItsAttackTrigger` and `AggressiveAloneIsNotAskedAbout` pin something not happening and pass today. They are accepted as guards.
  - `TestShape_*` naming for the turn gate and the opponents subject is fine, because no FRD 01 requirement defines them.
  - Next: `kernel-engineer` (tier sonnet).
- 2026-10-20 kernel-engineer: Implemented. ability.go: HookAttacks, TurnOwner (TurnAny/TurnYours), Trigger.Turn, SubjectOpponents. trigger.go: `pendingTrigger.item` (optional ready-made StackItem, pushed as-is by the flush; its label feeds DecideOrder), new `fireAttacks` (Aggressive keyword item first, then `fireHook(HookAttacks)`), turn gate in `fireHook` (`TurnYours` needs `g.Active == a.Controller`), SubjectOpponents in `resolveAbility` for OpScratch/OpDraw via a `who` list (ReturnToHand/PutOnDeck untouched, still SubjectSelf only). actions.go: `declareAttack` calls `fireAttacks` then `flushTriggers`; StackAggressive resolution unchanged. Reused fireHook, buffer, flushTriggers, scratch, draw. No state.go/stack.go change. Gates: go test ./... ok, go vet clean, gofmt -l pkg cmd empty, cardtown lint --all passed, sim -deck starter -p0 ANIMUS -p1 PATH exit 0, plain sim sorted output identical to rework/executable-rules. Reviewer: check that the Aggressive item is pushed even if the match ended (flush discards it, fine since the match is over).
