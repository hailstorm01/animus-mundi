---
id: T-0080
title: A combat-kill hook, fired when an attacker's combat damage destroys its defender (FR-CBT-002, wave B shape, sonnet)
kind: shape
tier: sonnet
wave: B
status: ready
branch: m2/T-0080-combat-kill-hook
depends_on: [T-0010, T-0066]
claims: [hook:combat-kill]
cards: []
requirements: [FR-CBT-002, FR-TRG-003]
---

## Goal
A trigger can watch `ability.HookCombatKill`: "whenever this asset attacks and kills" an opposing defender. It fires from the declared attacker when its Attack's combat damage has made the defender lethal, so the death sweep right after the Attack destroys it. Like every trigger, it is placed after that sweep (`FR-TRG-004`), in APNAP order with the defender's death triggers (`FR-TRG-003`). The first card, *Eternal Watchman* ("Whenever this Headquarters attacks and kills an opposing unit or structure, return chosen unit card from your discard pile to your hand."), also needs a chosen target in a discard pile, which is the next shape task. The card is a later card task.

## Inputs
- FRD 01:
  - `FR-CBT-002` Step 3 (lines 512-519): both combatants deal combat damage simultaneously, lethal combatants are destroyed by the death sweep, and "Post-Attack & Combat-Kill Triggers" (line 519): combat-kill abilities trigger when the asset that dealt the lethal combat damage was the declared attacker destroying an opposing defender. Retaliation kills (`is_attacker` false) do not count.
  - `FR-KWD-002o` (line 518): Berserk cleave is non-combat effect damage to a secondary target, so a secondary it kills is not a combat kill.
  - `FR-CBT-002` Step 3 "Combatant Target Check" (line 514): a fizzled Attack deals no combat damage, so it kills nothing.
  - `FR-TRG-003` (line 380) and `FR-TRG-004` (line 386): APNAP placement after the sweep.
- Survey row: Eternal Watchman (line 153): `EVENT_COMBAT_KILL` as attacker. No other M2 card has the hook (Dread Manifestation is Concordia).
- Rulings: none new. `SVY-R7` (source-position fallback, T-0066) applies unchanged.
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Hook` already lists "combat kill") and Open Question 9 (added with this task).
- Kernel today:
  - `resolveAttack` (`combat.go`) is Step 3: target check, `dealDamage` both ways, then `cleave`. The death sweep runs right after the item in `resolveStack` (`stack.go`), with nothing in between, and `IsLethal` (`layers.go`) is what the sweep reads.
  - Every self hook is one `fire<Hook>` wrapper over `fireHook` in `trigger.go`, which applies the gates (`Turn`, `Cornered`, `MaxPerTurn`) and records the source position.

### Shape (lead decisions, binding)
- **`ability.HookCombatKill`**, appended after `HookExhausted`, with the comment `// this asset attacked, and its combat damage left the defender lethal, so the death sweep after the Attack destroys it (FR-CBT-002 Step 3)`.
- **Firing:** add `(g *Game) fireCombatKill(att *Asset)` in `trigger.go` beside the other wrappers (`g.fireHook(att, ability.HookCombatKill)`). Call it at the end of `resolveAttack`'s combat branch, after `cleave`: if the defender `IsLethal`, the attacker's combat-kill triggers fire. Nothing else calls it.
  - Only the declared attacker fires. A defender that kills the attacker by retaliation fires nothing.
  - A fizzled Attack returns before the combat branch, so it fires nothing.
  - A Berserk secondary killed by cleave is not the defender, so it doesn't count.
  - The defender is a unit or structure in every case that matters: a lethal Headquarters ends the match at that sweep, and `flushTriggers` drops the buffer. The hook does not filter on the defender's type.
- **An attacker that dies in the same sweep still fires**, because the trigger fires before the sweep, as Unbowed does on lethal damage (`FR-KWD-002a`) and death triggers do as their source leaves. No M2 card reaches this: the only M2 card with the hook is a Headquarters, and its death ends the match. It is not tested. The design note's Open Question 9 records it for the first unit with this hook.
- **`doc.go`:** add `HookCombatKill` to the "Triggered abilities on self hooks" list.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written in `pkg/engine/combat_kill_test.go`, with synthetic cards only (Eternal Watchman is a later task). The file won't compile until `ability.HookCombatKill` exists. A convenient observable effect is an `OpDraw` step on the trigger (hand size), or an `OpAddArmor` step on `SubjectSelf`. Required cases:
- **Kill:** `TestFR_CBT_002_CombatKillTriggerFiresWhenTheAttackerKillsTheDefender`. A unit with the hook attacks an opposing unit and its combat damage destroys it. After the sweep the trigger is placed, controlled by the attacker's controller, and it resolves.
- **Structure:** `TestFR_CBT_002_CombatKillTriggerFiresWhenTheDefenderIsAStructure`.
- **Headquarters attacker:** `TestFR_CBT_002_CombatKillTriggerFiresFromAHeadquartersAttacker`. A Headquarters with the hook attacks an opposing Main Theater unit and kills it.
- **No kill:** `TestFR_CBT_002_CombatKillTriggerDoesNotFireWhenTheDefenderSurvives`.
- **Retaliation:** `TestFR_CBT_002_CombatKillTriggerDoesNotFireForTheDefender`. The defender has the hook, and its retaliation destroys the attacker while the defender survives. Nothing fires.
- **Cleave:** `TestFR_CBT_002_CombatKillTriggerDoesNotFireForABerserkSecondary`. A Berserk attacker with the hook: its cleave destroys the secondary target and the defender survives. Nothing fires.
- **Fizzle:** `TestFR_CBT_002_CombatKillTriggerDoesNotFireWhenTheAttackFizzles`. The defender leaves the battlefield before the Attack resolves, for example killed by an attack trigger. The Attack fizzles and the hook does not fire.
- **APNAP (`FR-TRG-003`):** `TestFR_TRG_003_CombatKillTriggerAndTheDefendersDeathTriggerAreAPNAP`. The defender has a death trigger. After the sweep, the active player's combat-kill trigger is placed first, and the non-active player's death trigger above it resolves first.
- **Existing tests:** every test in `pkg/engine` passes unchanged.
- **Left out:** an attacker that dies in the same sweep (see Shape), and a defender that is a Headquarters (the match ends).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. If a test needs one, `rules-test-author` names it here.

## Allowed Files
- `pkg/engine/ability/ability.go` (`HookCombatKill`)
- `pkg/engine/trigger.go` (`fireCombatKill`)
- `pkg/engine/combat.go` (the call at the end of Step 3)
- `pkg/engine/doc.go` (the self-hooks list)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_(CBT_002_CombatKill|TRG_003_CombatKill)'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] No card uses the new hook yet, so the sims are unchanged: `go tool task sim-compare -- autoloop/m2-wave-b` passes

## Log
- 2026-10-27 lead-architect (wave run B, session 17): Wrote the task. The run branch has everything `main` has, and no `open: before` directive is on the board. Waiting triggers checked against `autoloop/m2-wave-b` 74af332: #6 and #7 (mode is `serial`), #8 (no `deck.go` change), #15 (no turn hook), #20 (b) (no wave C shape task), #58 (no exhaust by a cost or an effect) and #59 (2) (no repeated targets). None has fired.
  - Wave B cards still on `NotYetExecutable`: Eternal Watchman (the Path Headquarters, a deck card, directive #14 (a)) and Rogue Fleshshaper (kept). The Watchman goes first.
  - The Watchman needs two new shapes, a combat-kill hook and a chosen target in a discard pile, plus its card definition. Together they make seven non-test files, so they are split three ways: this task (`hook:combat-kill`), then a shape task for a chosen target in a discard pile (`targeting.go`, `effects.go`, and the `Chosen` comments in `stack.go` and `pkg/view/view.go`), then the card task (`headquarters.go`, `cards.go`). Their claims don't overlap. The other two aren't written yet.
  - Placement: firing at the end of `resolveAttack` keeps the combat rule in `combat.go`, and `stack.go` stays unchanged. Searched `pkg/engine` and the `done` tasks for an existing combat-kill hook or kill record: none.
- 2026-10-27 rules-test-author: Wrote `pkg/engine/combat_kill_test.go` (8 tests, synthetic cards, observed with `wantScratches`, and Draw 1 from empty decks plus `wantPlaced` for APNAP): `TestFR_CBT_002_CombatKillTriggerFiresWhenTheAttackerKillsTheDefender`, `...FiresWhenTheDefenderIsAStructure`, `...FiresFromAHeadquartersAttacker` (sets the HQ card's `Ability` in the scene), `...DoesNotFireWhenTheDefenderSurvives`, `...DoesNotFireForTheDefender`, `...DoesNotFireForABerserkSecondary`, `...DoesNotFireWhenTheAttackFizzles` (attack trigger damages the defender to death at Step 2), and `TestFR_TRG_003_CombatKillTriggerAndTheDefendersDeathTriggerAreAPNAP`. Helpers requested: none (the file defines `combatKillTrigger`). Left out, as the task says: an attacker that dies in the same sweep, and a Headquarters defender. `go vet` fails only on the missing `ability.HookCombatKill`; `cardtown lint rules` passes.
- 2026-10-27 lead-architect (wave run B, session 17, D4): **Tests approved, `ready`.** I checked the 8 tests in 83d1258 against `FR-CBT-002` Step 3, `FR-KWD-002o` and `FR-TRG-003`. Each expected outcome follows from the cited text: kills of a unit, a structure and a Headquarters attacker's defender fire; a surviving defender, a retaliation kill, a cleave kill and a fizzled Attack don't; and the APNAP order shows in both placement and outcome. Every helper they use already exists (`striker`, `mortal`, `wantScratches`, `wantPlaced`, `s.cleave`). No test asserts on implementation details.
