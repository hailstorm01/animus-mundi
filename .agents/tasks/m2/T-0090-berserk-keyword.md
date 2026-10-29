---
id: T-0090
title: The Berserk keyword leaves the kernel, so a "Berserk N" line loads as unsupported (directive #68 part (b), harness, flash)
kind: harness
tier: flash
wave: B.2 (board directive #68 part (b), with #72's changes; the user's code review, before wave C)
status: ready
branch: m2/T-0090-berserk-keyword
depends_on: [T-0089]
claims: [keyword:berserk-removed, field:Keywords.Berserk, file:pkg/engine/cardload.go]
cards: []
requirements: []
---

## Goal
The kernel no longer knows the Berserk keyword. `Keywords` has no `Berserk` field, `ParseKeywordLine` doesn't match "Berserk N", and the view doesn't list it. A "Berserk 2" line in the card data then loads as an unsupported rules-text line, like any keyword the kernel doesn't implement. This is part (b) of directive #68. Part (a) (T-0089) removed Berserk's rules. Part (c) (the clients' reading of the secondary) and part (d) (the `Secondary` fields and `doc.go`) are their own tasks, so this task leaves every `Secondary` symbol and `doc.go` alone.

## Inputs
- Board directive #68 (user code review, 2026-10-28), part (b), its Tests list (the `pkg/sim/playout_test.go` items) and acceptance cases (a), (b), (d). The two comment edits that T-0089's review carried to this task's test commit, on #68's row.
- Directive #72 (Berserk replaced by Overkill, `SVY-R12`): drop the Tempest Surge assertion rather than reword it (#72 (1)); no test is rewritten into an Overkill test (#72 (3)). #72 (2): parts (a) and (b) merge before the user runs `pm-frd-drafter` on the outcome, so the lead records this merge on #72's row.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6b (wave B.2).
- Not edited here: FRD 01, the keyword matrix (`KIM-B04` still names `TestLoadCardsFromRepositoryData`) and the card CSVs. They wait for `pm-frd-drafter` on #72's outcome.
- Why the sims don't change: no starter card and no Animus or Path card prints Berserk, and the stress decks are Animus and Path. *Tempest Surge* (Concordia) and *Redline Infusion* (Soulless) are already unsupported (Tempest Surge's `T2:` line and Redline Infusion's text have no def), and stay so. `Keywords` is never serialized: the event log and the view don't encode it, and `view.keywords` lists only the keywords an asset has.

### Where it is today (at `autoloop/m2-wave-b.2` 79182ed)
| Where | Today |
| :--- | :--- |
| `pkg/engine/card.go:38` | `Keywords.Berserk int`, the last field |
| `pkg/engine/cardload.go:20-22` | the `KIM-B04` comment and `keywordLine`'s `Berserk` alternative |
| `pkg/engine/cardload.go:35-36` | `ParseKeywordLine`'s `case "Berserk":` |
| `pkg/view/view.go:222` | `num("Berserk", k.Berserk)` in `keywords` |
| `pkg/sim/playout_test.go:25` | `keywordDeck`'s `Berserker` unit |
| `pkg/sim/playout_test.go:93-96` | `TestLoadCardsFromRepositoryData`'s Tempest Surge assertion |

## Edits (lead decisions; copy exactly)

### 1. `pkg/engine/card.go`
- Remove the `Berserk          int` line from `Keywords`. Run `gofmt`; the other fields' alignment does not change, because `DeploysExhausted` is the longest name.

### 2. `pkg/engine/cardload.go`
- The comment and `keywordLine` (lines 20-22) become:
  ```go
  // keywordLine matches a numeric keyword, which always prints its N.
  var keywordLine = regexp.MustCompile(`^(Resistant|Aggressive|Armor) (-?\d+)$`)
  ```
- In `ParseKeywordLine`, remove the `case "Berserk":` and `kw.Berserk += n` lines. Nothing else in the function changes.

### 3. `pkg/view/view.go`
- In `keywords`, remove the line `num("Berserk", k.Berserk)`. `Secondary` (line 87) and the stack item's reading of it stay (part (c)).

No other non-test file changes.

## Tests (harness: the engineer makes these edits, all in the `test(T-0090)` commit)
No new tests: this task removes a keyword, and cases 2–4 show it is gone. Every change below is to an existing test, named here as D2.4 requires. Each passes both before and after the implementation commit: commit these edits first, run `go test ./...` on that commit (the old code), and record the result.

- `pkg/sim/playout_test.go`:
  - In `keywordDeck`, delete the line `unit("Berserker", 2, 3, engine.Keywords{Berserk: 2}),`. (Lead decision: #68 says the Berserker "loses Berserk"; a unit named Berserker with no keyword would be a second `Vanilla` under a Berserk name, so it goes. The pool shrinks from 13 to 12 cards, which only reshuffles the random playouts' costs and deck order.)
  - In `TestLoadCardsFromRepositoryData`, delete the four lines from `// KIM-B04 ruling: Tempest Surge prints Berserk 2.` through the closing `}` of its `if`, so the function ends with the `t.Logf(…)` line (#72 (1)).
- `pkg/engine/rules_test.go`: restore a section heading above the `declareThen` helper. The lines
  ```go
  // declareThen declares a, runs between (standing in for Step 2 triggers), then
  ```
  get, directly before them, the heading and one blank line:
  ```go
  // ---- Rulings KIM-J01, S03, S04 (2026-10-19) ----

  ```
  so there is one blank line between the `}` ending `TestFR_KWD_002f_AggressiveResolvesBeforeDamageAndExpires` and the heading, as there is between every section. (#68's row says "directly above `TestFR_CBT_002_CommittedAttackIgnoresLaterGuardianAndEthereal`"; `declareThen` is that test's only helper and sat under the heading before T-0089, so the heading opens the section above it.)
- `pkg/engine/attack_reverify_test.go`: the file comment's first three lines
  ```go
  // Tests for T-0053: an Attack's combatants are re-verified when it resolves
  // (SVY-R8, FRD 01 v3.8: FR-CBT-002 Step 3 "Combatant Target Check", KIM-B06 as
  // extended, FR-KWD-002o, FR-TGT-003).
  ```
  become
  ```go
  // Tests for T-0053: an Attack's combatants are re-verified when it resolves
  // (SVY-R8, FRD 01 v3.8: FR-CBT-002 Step 3 "Combatant Target Check", FR-TGT-003).
  ```
  The rest of the file comment does not change.

## Acceptance Cases (harness tasks only; written by the lead)
Run each command on its own, pointed at your worktree (W). Record each command and its result in the Log. The reviewer re-runs them.
1. **Tests first, and only the named ones:** the branch has a `test(T-0090): …` commit before the implementation commit, holding every test edit above and nothing else. `go -C W test ./...` passes on the test commit (old code) and on the branch tip. `git -C W diff autoloop/m2-wave-b.2 --stat -- '*_test.go'` lists exactly `pkg/engine/attack_reverify_test.go`, `pkg/engine/rules_test.go` and `pkg/sim/playout_test.go`.
2. **The keyword is gone from these files:** `git -C W grep -n -i berserk -- pkg/engine/card.go pkg/engine/cardload.go pkg/view/view.go pkg/sim/playout_test.go pkg/engine/attack_reverify_test.go` prints exactly one line, the `Secondary` field comment in `pkg/view/view.go`, which part (c) removes.
3. **What is left is parts (c) and (d):** `git -C W grep -l -i berserk -- pkg cmd ':!pkg/toolchain/linter'` lists exactly `pkg/engine/actions.go`, `pkg/engine/doc.go`, `pkg/engine/events.go`, `pkg/engine/stack.go` and `pkg/view/view.go`.
4. **Both Berserk cards load and stay unsupported (#68 case (d)):** `go -C W test ./pkg/sim -run TestLoadCardsFromRepositoryData -v` passes. Its log line reads `54 of 169 cards are executable by the kernel`, as on the run branch, and the list it prints names neither *Tempest Surge* nor *Redline Infusion*.
5. **Sims unchanged (#68 case (b)):** `go -C W tool task sim-compare -- autoloop/m2-wave-b.2` passes: the starter and stress sims match byte for byte, exit codes included.
6. **Only these files:** `git -C W diff autoloop/m2-wave-b.2 --stat` lists only `pkg/engine/card.go`, `pkg/engine/cardload.go`, `pkg/view/view.go`, the three test files in case 1, and this task file.
7. `go -C W tool task check` and `go -C W tool task sim` pass.

The reviewer also compares a starter `-events` file from both builds (#68 case (b)); it must differ only in each game's `match-start` `revision` stamp (T-0089's review). That runs built binaries, so it is the reviewer's check, not the engineer's.

## Allowed Files
- `pkg/engine/card.go`
- `pkg/engine/cardload.go`
- `pkg/view/view.go`
- the three test files in acceptance case 1, for the named test edits only

## Done When
- [ ] Cases 1–7 hold, with their commands and results in the Log
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-28, lead-architect (wave B.2 run, session 2): wrote the task from board directive #68 part (b), the next of wave B.2's blocking directives in board order, with #72 (1) and (3), and the two comment edits T-0089's review put on #68's row. Harness, so it goes straight to `ready`; tier `flash`, the harness default, since #68 gives `sonnet` only for part (a)'s test changes and every command here is on the headless allow-list. Lead decisions: (1) no new test (see Tests); (2) the `Berserker` unit leaves `keywordDeck` rather than keeping a Berserk name with no keyword; (3) the restored heading sits above `declareThen` (see Tests); (4) the `-events` comparison is the reviewer's. Waiting directives checked: none has triggered (#8: this task doesn't touch `deck.go`; #15, #20 (b), #58, #59 (2), #61, #71: no such hook, shape, claim or card here; #6, #7: mode is `serial`). Parts (c)–(d), #69 and #70 stay open for the next sessions. Next: `kernel-engineer` on Flash.
