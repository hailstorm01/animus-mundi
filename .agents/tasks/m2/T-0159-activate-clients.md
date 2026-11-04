---
id: T-0159
title: The event log, replay and sim policy know an activation (harness, sonnet)
kind: harness
tier: sonnet
wave: D
status: review
branch: m2/autoloop-wave-d/T-0159-activate-clients
depends_on: [T-0158]
claims: [event:activate, replay:activate, policy:activate]
cards: []
requirements: [FR-ACT-001]
---

## Goal
An activation is recorded, replayed and played by the sim, under the rule monitors, before the first card with an activated ability becomes executable, so that card's task only combines shapes. The play screen's label is T-0161, and comments are T-0160.

## Inputs
- FRD 01: `FR-ACT-001`. No rule changes here.
- Decision record: `docs/architecture/decisions/0005-structured-game-events-and-observer.md` (every accepted action emits one `action` event; renaming or removing a field supersedes it).
- Design note: `docs/architecture/m2-ability-model.md`, Changes in `engine`, Actions (`ActActivate`, asset plus ability index).
- Code today, on the run branch:
  - `emitAction` (`pkg/engine/events.go:142`) has no case for `ActActivate`, so the event carries only its `Detail` and `Player`.
  - `actionOf` (`pkg/eventlog/replay.go:257`) already maps `Source` to `Action.Asset`, and never sets `Action.Ability`. So a replay of index 1 runs index 0.
  - `diff` and `structured` (`replay.go:219-221`, `:246`) already compare `Amount`.
  - `ProgressPolicy` (`pkg/sim/policy.go:49`) drops `ActActivate`: only `RandomPolicy` can pick one today.

### Shape (lead decisions, binding)
- **`emitAction`:** a case for `ActActivate` sets `e.Source = a.Asset` and `e.Amount = a.Ability`. `Detail` stays `"Activate"`, and `Card`, `Inst`, `Target` and `Zone` stay empty. The `place` event right after it already names the card.
- **Why `Amount`, not a new field:** ADR-0005 gives `Amount` a meaning per event kind (a `place` event's stack depth), no other action event sets it, and `diff` already compares it. Index 0 is omitted from the JSON and reads back as 0, which is correct. Extend `Amount`'s comment on `Event` to say this: on an Activate action event it is the ability's index.
- **`actionOf`:** for `ActActivate`, `a.Ability = e.Amount`, beside the existing Generate line.
- **`ProgressPolicy`:** `engine.ActActivate` joins `ActMarch, ActAttack, ActScratch` as an act, and item 3 of the doc comment says so. Nothing else changes. An activation's cost doesn't count toward `needReq` or `needTac`, because the policy still generates only for hand cards.
- **Sims:** no card has a `Def.Activated` and the policy draws no extra random number while `acts` is unchanged, so `go tool task sim-compare -- autoloop/m2-wave-d` should pass byte for byte. If it differs, don't change code to make it match: report the first differing game in the Log and hand off.
- **Monitors and the view need no change (the lead read them).** `monitor.go:164` matches only Attack, Generate and Convert action events. `declare` chooses targets and then pushes, so a Target decision is followed by the item's `place` event, as `endTargetPrompts` requires. The resource and damage monitors read pools and targets, not how an asset was exhausted or killed. `stackView` treats any `StackAbility` alike. Cases 3 and 6 check this by running it. If a monitor flags an activation in case 3, or the view is wrong in case 6, don't edit `monitor*.go` or `view.go`: set `status: blocked` and name the event (`kernel-engineer` §3 step 7).
- **#104 is still open.** Every exactly-one target slot in these tests must have a legal entity on declaration. Case 3's slot can always name the opposing Headquarters.

## Acceptance Cases (harness tasks only; written by the lead)
1. **`TestEvents_Case4_ActivateActionEventNamesItsAsset`** in `pkg/engine/events_test.go`, using `activated_test.go`'s helpers. A unit with two abilities with `nil` costs, so both can be activated in one test. Activate index 1 → exactly one `action` event, with `Detail` `"Activate"`, `Player` 0, `Source` the asset's ID, `Amount` 1, and empty `Card`, `Inst`, `Target` and `Zone`. It comes before the item's `place` event. Activate index 0 → `Amount` 0.
2. **Edit to an existing test, in the `test(T-0159)` commit:** `TestEvents_Case4_EveryActionKind` (`pkg/engine/events_test.go`). The comment's second sentence reads: `// Activate is covered by TestEvents_Case4_ActivateActionEventNamesItsAsset, because no starter card has an executable activated ability yet.` The code in the test doesn't change, and the `ActActivate` skip stays.
3. **`TestReplay_ActivationsRoundTrip`** in `pkg/eventlog/replay_test.go`. Both seats play a synthetic `engine.DeckList`:
   - HQ with `ReqYield` 3, `TacYield` 2, `Power` 3, `Health` 20, and `DeployZones` of the Headquarters Zone.
   - Main: 30 copies of a 1/3 unit (`ReqCost` 1, yields 1, both deploy zones) whose `ability.Def.Activated` holds three abilities:
     - index 0, `CostExhaustSelf`: draw a card;
     - index 1, `CostKillSelf`: 1 damage to one chosen entity (`Target` with `Filter{Allegiance: AllegianceOpposing}`, so the opposing Headquarters is always legal);
     - index 2, `CostResources` `Req` 1: add 1 Armor to itself.

   Run the file's `play` helper (`sim.SimulateWith` with `sim.ProgressPolicy`, monitors on, through a JSON round trip) for 30 games with a fixed seed, and a resolver that returns this deck for its name:
   - `SimulateWith` returns no error, so no rule monitor flagged a game.
   - Over the games, `action` events with `Detail` `"Activate"` appear with `Amount` 0, 1 and 2, each at least once.
   - For every game, `eventlog.Replay(evs, len(evs), resolve)` returns no error.
4. **`TestReplay_ActivationWithAnotherIndexDiverges`** in the same file, same deck. Take a game from case 3 holding an Activate event with `Amount` 1, and set that event's `Amount` to 0 in a copy → `Replay` to the end returns an error containing `replay diverged`. A record with a wrong index is never accepted.
5. **`TestProgress_ActivationIsAnAct`** in `pkg/sim/policy_test.go`, over seeds 1 to 16, on the file's `scene`:
   - Empty hand, every asset Exhausted, player 0's pool R1 T0, and a Reserves unit with one activated ability costing R1 → the policy returns only `ActActivate` or `ActEndTurn`, and `ActActivate` at least once. Every answer is in `LegalActions`.
   - The same scene plus an affordable unit in hand → `ActDeploy` on every seed, because a play from hand still comes first.
6. **`TestView_ActivatedAbilityItemShowsItsTargets`** in `pkg/view/view_test.go`. This one pins existing behavior through the real action path, so it may pass before the implementation commit: say so in the Log. Build the game with a chooser of your own (`engine.Config.Choosers`) and two opposing units. The source prints "exhaust: kill chosen opposing unit", plus a `HookExhausted` trigger with a chosen target of its own. Activate it. While the chooser answers the trigger's Target decision, the activated item is on the stack: for both seats, `view.PlayerView` shows it as `Kind` `"Ability"`, `Card` the source's name, `Source` the asset's ID, empty `Keyword`, and `Chosen` holding the unit picked on declaration.

Each case but 2 and 6 fails before the implementation commit.

## Allowed Files
- `pkg/engine/events.go` (`emitAction`, the comment on `Event.Amount`)
- `pkg/eventlog/replay.go` (`actionOf`)
- `pkg/sim/policy.go` (`ProgressPolicy`)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine ./pkg/eventlog ./pkg/sim ./pkg/view -run "TestEvents_Case4|TestReplay_Activation|TestProgress_ActivationIsAnAct|TestView_ActivatedAbilityItem"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d` passes (see Shape, Sims)

## Log
- 2026-11-04 lead-architect: split off T-0158 under the D2 size rule (D2.4). Written `queued`.
- 2026-11-04 lead-architect: wave D run, session 2. No `main` commits to sync (`main` is at 58a2bc9, the run branch's base). Not `WAVE-COMPLETE`: all four header cards are on `NotYetExecutable`. No `open: before` directive; #14 is `open` and orders the cards. Waiting triggers #6, #7, #8 (no task touches `deck.go`), #15, #71, #91 (2) not fired. #104 waits for the first card task with an activated ability, which this is not. Written against the code of the day. The label needs a printed text the `View` doesn't carry, so `pkg/view` joins the four files listed under Covers and the task would have five. Under the D2 size rule the label moves to T-0161 (`queued`, D2.4). What is left is three files. Reuse search: `grep "ActActivate"` outside `pkg/engine` finds nothing. `Event.Amount`, `diff`, `structured` and the `acts` slice are reused. Harness, so it skips `tests`. Status `ready`; dispatching `kernel-engineer` (sonnet).
- 2026-11-04 kernel-engineer: implemented. Test commit first: cases 1-5 and the case 2 comment edit, plus case 6. Cases 1, 3, 4 and 5 failed before the implementation commit (no Source/Amount on the Activate event, no Activate events from ProgressPolicy, no policy activation). Case 6 (`TestView_ActivatedAbilityItemShowsItsTargets`) passed before the implementation, as the task predicted; it pins existing behavior. Case 2 changes only the comment in `TestEvents_Case4_EveryActionKind`. Changes: `emitAction` has a `case ActActivate` (Source = asset, Amount = ability index); the `Event.Amount` comment says what it means per event kind; `actionOf` sets `a.Ability = e.Amount` for ActActivate; `ProgressPolicy` counts ActActivate as an act, and item 3 of its comment names it. Reuse: `Event.Amount`, `diff`/`structured`, the `acts` slice; no new field or helper. No change to monitors or the view; the 30-game replay with monitors on passed. Commands: `go test ./pkg/engine ./pkg/eventlog ./pkg/sim ./pkg/view -run "TestEvents_Case4|TestReplay_Activation|TestProgress_ActivationIsAnAct|TestView_ActivatedAbilityItem"` pass; `go tool task check` pass; `go tool task sim` clean; `go tool task sim-compare -- autoloop/m2-wave-d` PASS byte for byte. Reviewer: the case 5 scene exhausts every asset including the HQ and gives player 0 R1 T0; case 6 detects the trigger's decision by a non-empty stack.
