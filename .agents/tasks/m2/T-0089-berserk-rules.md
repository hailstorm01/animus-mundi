---
id: T-0089
title: Berserk's rules leave the kernel, so an Attack has no secondary target and nothing cleaves (directive #68 part (a), harness, sonnet)
kind: harness
tier: sonnet
wave: B.2 (board directive #68 part (a), with #72's changes; the user's code review, before wave C)
status: in-progress
branch: m2/T-0089-berserk-rules
depends_on: [T-0088]
claims: [rule:berserk-secondary-removed, func:CanCleave, func:hasCleaveTarget, func:legalSecondary, func:cleave, file:pkg/engine/combat.go]
cards: []
requirements: []
---

## Goal
The kernel no longer implements Berserk's rules. An Attack with a non-zero `Secondary` is illegal, `LegalActions` never offers one, and `resolveAttack` never cleaves. The Berserk tests go, and the tests that only used a Berserk attacker to set up a board keep what they test with a plain one. This is part (a) of directive #68, the first of four parts. The `Berserk N` keyword parse and `Keywords.Berserk` (part (b)), the clients' reading of the secondary (part (c)) and the `Secondary` fields (part (d)) stay for their own tasks, so this task leaves the `Secondary` fields and the keyword alone.

## Inputs
- Board directive #68 (user code review, 2026-10-28), part (a), its Tests list, and acceptance cases (a)–(b). Directive #72 (Berserk replaced by Overkill, `SVY-R12`): no test is rewritten into an Overkill test (#72 (3)). #72 (1) and (2) belong to part (b).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6b (wave B.2).
- Rules this stops implementing: `FR-KWD-002o` and the Berserk clauses of `FR-CBT-002`. FRD 01 is not edited here: `pm-frd-drafter` retires `FR-KWD-002o` when the user runs it on #72's outcome. `cardtown lint rules` then reports `FR-KWD-002o` as not yet executable, which is not a failure (the linter fails only on a test naming a requirement FRD 01 doesn't own).
- Rules the kept tests still pin, unchanged: `FR-CBT-002` (side-change fizzles), `FR-REP-004` and `FR-KWD-002a` (Unbowed on retaliation), `FR-TRG-002`, `FR-TRG-003` (ordering simultaneous death triggers).
- Why the starter sim is unchanged: no starter card has Berserk. The stress decks may hold a Berserk card (*Tempest Surge*, Concordia), so the stress sim may change.

### Where it is today (at `autoloop/m2-wave-b.2` 50c607c)
| Where | Today |
| :--- | :--- |
| `actions.go:177-184`, `check`'s Attack case | secondary legality through `CanCleave`, and the mandatory-secondary rule through `hasCleaveTarget` |
| `actions.go:254-258`, `LegalActions` | after each plain Attack, one Attack per `CanCleave` secondary |
| `combat.go:52-80` | `CanCleave`, `hasCleaveTarget`, `legalSecondary` |
| `combat.go:105-121`, `resolveAttack` | the defender-side fizzle's "no Berserk cleave" wording; the attacker-side fizzle cleaves from last-known values |
| `combat.go:130-135`, `resolveAttack` | the cleave after combat damage, and its comments |
| `combat.go:140-152` | `cleave` |

## Edits (lead decisions; copy exactly)

### 1. `pkg/engine/actions.go`
- In `check`'s `case ActAttack:`, the whole `if a.Secondary != 0 { … } else if g.hasCleaveTarget(att, def) { … }` block becomes:
  ```go
  		if a.Secondary != 0 {
  			return illegal("an Attack has no secondary target")
  		}
  ```
- In `LegalActions`, the `for _, sec := range g.Theater { … }` loop after `try(Action{Kind: ActAttack, Player: pid, Asset: a.ID, Target: def.ID})` is removed. The plain Attack `try` stays.
- `Action.Secondary` and its field comment stay as they are (part (d) removes them).

### 2. `pkg/engine/combat.go`
- Remove `CanCleave`, `hasCleaveTarget`, `legalSecondary` and `cleave`, each with its doc comment.
- In `resolveAttack`, the defender-side fizzle's detail `"attack fizzles: the defender is gone, so there is no Berserk cleave"` becomes `"attack fizzles: the defender is gone"`.
- The attacker-side fizzle block becomes:
  ```go
  	att := g.Asset(it.Source)
  	if att == nil || !inCombatZone(att) {
  		g.logEvent(Event{Kind: "fizzle", Player: -1, Source: it.Source}, "attack fizzles: the attacker is gone")
  		return
  	}
  ```
- After the two `g.dealDamage` calls, the two-line `// Berserk cleave is committed …` comment and `g.cleave(it, att, def)` are removed, and the comment above `if g.IsLethal(def)` becomes:
  ```go
  	// Only the declared attacker fires, and only for the defender: a
  	// retaliation kill does not count (FR-CBT-002 Step 3).
  ```
- `declareAttack`'s `Secondary: a.Secondary` stays (part (d)).
- Remove the `animus-mundi/pkg/engine/ability` import: `legalSecondary` was its only user. Remove nothing else unless the compiler says it is unused.

No other non-test file changes. `card.go`, `cardload.go`, `stack.go`, `events.go`, `doc.go`, `pkg/view`, `pkg/eventlog`, `pkg/sim` and `cmd/cardtown` are parts (b)–(d).

## Tests (harness: the engineer makes these edits, all in the `test(T-0089)` commit)
No new tests: this task removes a rule. Every change below is to an existing test, named here as D2.4 requires. Each edited test drops Berserk from its setup, so it passes both before and after the implementation commit. That is the check that it still tests what it tested: commit these edits first, run `go test ./...` on that commit (the old code), and record the result.

**Delete** (each with the `// Covers:` and other comment lines directly above it):
- `pkg/engine/rules_test.go`: the 13 `TestFR_KWD_002o_*` functions (`BerserkCleavesWhenPrimarySurvives`, `BerserkCleavesEvenWhenAttackerDies`, `BerserkCleaveIsNonCombatDamage`, `SecondaryCannotBeWarded`, `SecondaryMayBeEthereal`, `GuardianDoesNotRestrictSecondary`, `SecondaryExcludesPrimary`, `SecondaryThatGainsWardBeforeStep3Dissipates`, `NoCleaveWhenAttackingHeadquarters`, `SecondaryIsMandatoryWhenALegalOneExists`, `NoSecondaryRequiredWithoutALegalTarget`, `CleaveSurvivesAnAttackerSideFizzle`, `NoCleaveWhenTheDefenderIsGone`), and the two-line comment `// Rulings KIM-B01..B03 (2026-10-19): the secondary follows normal targeting` / `// rules with the primary excluded.` `declareThen` and `TestFR_CBT_002_CommittedAttackIgnoresLaterGuardianAndEthereal`, which sit between them, stay.
- `pkg/engine/attack_reverify_test.go`: `TestFR_KWD_002o_SecondaryChangesSidesDissipates`.
- `pkg/engine/combat_kill_test.go`: `TestFR_CBT_002_CombatKillTriggerDoesNotFireForABerserkSecondary`.
- `pkg/engine/unbowed_test.go`: `TestFR_KWD_002a_BerserkCleaveDamageFiresTheSecondarysUnbowed` and `TestFR_KWD_002a_BerserkCleaveReducedToZeroDoesNotFireUnbowed`.
- `pkg/engine/scenario_test.go`: the `scene.cleave` helper.

**Edit:**
- `pkg/engine/attack_reverify_test.go`, the `sides` board, used by `TestFR_CBT_002_DefenderChangesSidesFizzles` and `TestFR_CBT_002_AttackerChangesSidesFizzles`:
  - `sides`'s doc comment becomes `// sides is the board of the side-change scenes: player 0's attacker (power 3)` / `// attacks Defender (power 2).`, and its fields become `att, def *Asset`.
  - In `newSides`, the attacker becomes `striker("Attacker", 3, 9, Keywords{}, mayDraw())`, and the `Bystander` line is removed.
  - `attack()` becomes `func (b *sides) attack() { b.s.do(b.s.attack(b.att, b.def)) }`.
  - In `wantFizzle`, the alive check becomes `if !b.s.alive(b.att) || !b.s.alive(b.def) {`, and the `b.sec.Damage` check (the "no Berserk cleave" assertion) is removed. The two fizzle damage assertions stay.
  - Both tests' `// Covers: FR-KWD-002o, FR-TGT-003` becomes `// Covers: FR-TGT-003`. Their bodies don't change.
- `pkg/engine/combat_kill_test.go`, the file comment: remove the clause `a Berserk secondary killed by cleave is not a combat kill (FR-KWD-002o: cleave is non-combat effect damage), ` and rewrap the paragraph, so it reads "… Retaliation kills do not count (is_attacker false), and a fizzled Attack deals no combat damage (FR-CBT-002 Step 3, "Combatant Target Check"). Like every trigger …".
- `pkg/engine/unbowed_test.go`:
  - The section comment `// --- combat: both combatants, Berserk ---` becomes `// --- combat: both combatants ---`.
  - `TestFR_KWD_002a_RetaliationOnABerserkAttackerFiresItsUnbowed` is renamed `TestFR_KWD_002a_RetaliationOnTheAttackerFiresItsUnbowed`, its `// Covers: FR-KWD-002o` becomes `// Covers: FR-REP-004, FR-CBT-002`, and its body becomes:
    ```go
    	s := newScene(t)
    	att := s.put(0, ZoneMainTheater, bowed("Attacker", 2, 9, Keywords{}, unbowed(nil, scratchN(1))))
    	def := s.put(1, ZoneMainTheater, unit("Wall", 2, 9, Keywords{}))
    	s.do(s.attack(att, def))
    	s.wantDamage(att, 2)
    	s.wantScratches(0, 1, "retaliation damage to the attacker fires the attacker's Unbowed (FR-REP-004)")
    ```
  - `TestFR_KWD_002a_OnlyTheAssetThatTookDamageFires`, which #68 also lists, already uses a plain attacker (`Hitter`) and no secondary: leave it unchanged.
- `pkg/engine/death_trigger_test.go`: `berserkKillsTwo` is replaced by `sweepKillsTwo`, whose two deaths come from one deploy trigger's `SubjectAll` damage step. The lead ran this replacement against the three tests below on the run branch, and all three pass:
  ```go
  // sweepKillsTwo has player 0 deploy a unit whose deploy trigger deals 5 damage
  // to every opposing unit, so both of player 1's units die in one sweep. Player
  // 1's deck top is A, B.
  func sweepKillsTwo(s *scene, d, sec *Card) []*Instance {
  	s.t.Helper()
  	cards := s.onTop(1, vanilla("A", "B")...)
  	dv := s.put(1, ZoneMainTheater, d)
  	sv := s.put(1, ZoneMainTheater, sec)
  	opposing := ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}
  	s.deployFromHand(0, trigUnit("Sweeper", deployTrigger(allDamage(5, opposing))))
  	if s.alive(dv) || s.alive(sv) {
  		s.t.Fatalf("precondition: both of player 1's units die in one sweep")
  	}
  	return cards
  }
  ```
  Its callers, `TestFR_TRG_002_NonActivePlayerOrdersTheirOwnDeathTriggers`, `TestFR_TRG_003_NonActivePlayersOrderAnswerIsApplied` (twice) and `TestFR_TRG_003_SimultaneousDeathsFallBackToBattlefieldOrder`, call `sweepKillsTwo` with the same arguments. Nothing else in them changes.

Not in this task: `pkg/sim/playout_test.go` (part (b)) and `pkg/view/view_test.go` (part (c)). The fences and FRD-ownership linter tests keep "Berserk" as FRD text fixtures.

## Acceptance Cases (harness tasks only; written by the lead)
Run each command on its own, pointed at your worktree (W). Record each command and its result in the Log. The reviewer re-runs them.
1. **Tests first, and only the named ones:** the branch has a `test(T-0089): …` commit before the implementation commit, holding every test edit above and nothing else. `go -C W test ./...` passes on the test commit (old code) and on the branch tip. `git -C W diff autoloop/m2-wave-b.2 --stat -- '*_test.go'` lists exactly `pkg/engine/attack_reverify_test.go`, `combat_kill_test.go`, `death_trigger_test.go`, `rules_test.go`, `scenario_test.go` and `unbowed_test.go`.
2. **No Berserk test or helper left in the engine tests:** `git -C W grep -n "TestFR_KWD_002o_\|cleave(\|berserkKillsTwo\|Berserk" -- 'pkg/engine/*_test.go'` prints nothing.
3. **No Berserk rule left:** `git -C W grep -n -i "cleave\|berserk\|legalSecondary" -- pkg/engine/actions.go pkg/engine/combat.go` prints exactly one line, `Action.Secondary`'s field comment in `actions.go`, which part (d) removes.
4. **The secondary is only refused:** `git -C W grep -n "Secondary" -- pkg/engine/actions.go pkg/engine/combat.go` prints exactly three lines: the `Secondary` field in `Action`, `if a.Secondary != 0 {` in `check`, and `declareAttack`'s `g.push`.
5. **Starter sim unchanged, stress sim recorded:** `go -C W tool task sim-compare -- autoloop/m2-wave-b.2`. The starter comparison prints no diff. The stress comparison may differ, in which case the command exits 1; record whether it differed, and if so the summary lines of `W/bin/sim-compare/base-stress.txt` and `W/bin/sim-compare/new-stress.txt` (outcomes and win rates; exit codes must both be 0).
6. **Starter events differ only in the fizzle wording:** after case 5, from W, run `W/bin/sim-compare/base.exe sim -deck starter -p0 ANIMUS -p1 PATH -events W/bin/sim-compare/base-events.jsonl` and `W/bin/sim-compare/new.exe sim -deck starter -p0 ANIMUS -p1 PATH -events W/bin/sim-compare/new-events.jsonl`, then `git -c core.autocrlf=false diff --no-index --stat W/bin/sim-compare/base-events.jsonl W/bin/sim-compare/new-events.jsonl`. Exactly one line differs: the one `fizzle` event (on the run branch, 1 of 104197 lines) whose `detail` goes from `attack fizzles: the defender is gone, so there is no Berserk cleave` to `attack fizzles: the defender is gone`. This is the lead's reading of #68 case (b) ("byte-identical"), which the wording change #68 asks for makes impossible to the letter: nothing a starter game does changes. Any other difference fails the case.
7. **Only these files:** `git -C W diff autoloop/m2-wave-b.2 --stat` lists only `pkg/engine/actions.go`, `pkg/engine/combat.go`, the six test files in case 1, and this task file.
8. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/actions.go`
- `pkg/engine/combat.go`
- the six test files in acceptance case 1, for the named test edits only

## Done When
- [ ] Cases 1–7 hold, with their commands and results in the Log
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-28, lead-architect (wave B.2 run, session 1): wrote the task from board directive #68 part (a), the first task of wave B.2, with #72's changes (only #72 (3) touches this part). Harness, so it skips `tests` and goes straight to `ready`; tier `sonnet` as #68 gives. Lead decisions: (1) no new test, because the task removes a rule and the deletions plus cases 2–4 show it is gone; a test of "Secondary is illegal" would need a Berserk attacker to fail first, and part (b) would delete it. (2) #68 lists `TestFR_KWD_002a_OnlyTheAssetThatTookDamageFires` for a plain attacker, but it already has one, so it stays unchanged. (3) The edited `RetaliationOnTheAttackerFiresItsUnbowed` now overlaps `TestFR_KWD_002a_BothCombatantsFireTheirOwnUnbowed` (there both combatants have Unbowed). #68 says to keep it, so it stays; the overlap is in test setup only, which D2.0 doesn't count as a duplicate. (4) Case 6 allows the one changed fizzle line (see the case). Parts (b)–(d), #69 and #70 stay open for the next sessions, in board order; #61 and #71 stay `waiting`. Next: `kernel-engineer`.
