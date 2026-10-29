---
id: T-0093
title: An Attack has one target, so Action and StackItem lose Secondary (directive #68 part (d2), harness, flash)
kind: harness
tier: flash
wave: B.2 (board directive #68 part (d), second half, with #72; the user's code review, before wave C)
status: in-progress
branch: m2/T-0093-secondary-fields
depends_on: [T-0092]
claims: [field:Action.Secondary-removed, field:StackItem.Secondary-removed, file:pkg/engine/actions.go, file:pkg/engine/combat.go, file:pkg/engine/stack.go]
cards: []
requirements: []
---

## Goal
`Action` and `StackItem` have no `Secondary` field, and `check` no longer needs to reject one. After this, no code outside the linter's FRD-text fixtures mentions Berserk, a secondary target or cleave. This is the second half of directive #68 part (d), and the last of #68. T-0092 removed `Event.Secondary` and its two writers' reads of these fields.

## Inputs
- Board directive #68 (user code review, 2026-10-28), part (d) and acceptance cases (a)–(d); #72 (3).
- Decision record: `docs/architecture/decisions/0006-attacks-carry-no-secondary-target.md`.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6b (wave B.2).
- Why nothing observable changes: since T-0089 `check` rejects an Attack with a non-zero `Secondary`, and nothing sets one (the sim's policies, `LegalActions`, the replay checker and `cardtown play` all build Attacks with `Asset` and `Target` only). With the field gone, that rejection can't happen, so it goes too. No test sets or reads either field.

### Where it is today (at `autoloop/m2-wave-b.2` after T-0092; line numbers from 78b42bc, where T-0092 doesn't touch these files' lines)
| Where | Today |
| :--- | :--- |
| `pkg/engine/actions.go:51-52` | `Action`'s `Target` line (commented "Attack: primary defender") and `Secondary` line |
| `pkg/engine/actions.go:177-179` | `check`'s Attack case ends with `if a.Secondary != 0 { return illegal("an Attack has no secondary target") }` |
| `pkg/engine/combat.go:9` | `declareAttack` pushes `&StackItem{…, Target: a.Target, Secondary: a.Secondary}` |
| `pkg/engine/stack.go:27-28` | `StackItem`'s `Target` line (commented "StackAttack: primary defender") and `Secondary` line |

## Edits (lead decisions; copy exactly)

### 1. `pkg/engine/actions.go`
- In `Action`, delete the `Secondary` line, and the `Target` line's comment becomes `// Attack: the defender`. Run `gofmt`; the struct then reads:
  ```go
  type Action struct {
  	Kind     ActionKind
  	Player   PlayerID
  	Asset    int      // acting asset: Generate, March, Attack, Scratch
  	Resource Resource // Generate
  	Card     int      // Deploy: Instance.ID in hand
  	Zone     Zone     // Deploy: declared battlefield zone
  	Target   int      // Attack: the defender
  }
  ```
- In `check`'s `case ActAttack:`, delete the three lines from `if a.Secondary != 0 {` through its closing `}`, so the case ends with the `illegal combat target (FR-CBT-001)` check.

### 2. `pkg/engine/combat.go`
- In `declareAttack`, the push becomes:
  ```go
  	g.push(&StackItem{Kind: StackAttack, Controller: a.Player, Source: att.ID, Target: a.Target})
  ```

### 3. `pkg/engine/stack.go`
- In `StackItem`, delete the `Secondary` line, and the `Target` line's comment becomes `// StackAttack: the defender`. Run `gofmt`; the two lines then read:
  ```go
  	Source int // asset ID: marching unit, attacker, scratching HQ
  	Target int // StackAttack: the defender
  ```

No other non-test file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests and no test commit: this removes two fields that are always 0, and the cases below show they are gone. Run each command on its own, pointed at your worktree (W). Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#68 case (a)): `git -C W diff autoloop/m2-wave-b.2 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#68 case (b)): `go -C W tool task sim-compare -- autoloop/m2-wave-b.2` prints `[PASS]`.
3. **Berserk is out of the code** (#68 case (c)): `git -C W grep -n -i -E "berserk|secondary|cleave" -- pkg cmd ":!pkg/toolchain/linter"` prints nothing (exit code 1).
4. **Both Berserk cards load and stay unsupported** (#68 case (d)): `go -C W test ./pkg/sim -run TestLoadCardsFromRepositoryData -v` passes. Its log line reads `54 of 169 cards are executable by the kernel`, as on the run branch, and the list it prints names neither *Tempest Surge* nor *Redline Infusion*.
5. **Only these edits:** `git -C W diff autoloop/m2-wave-b.2 --stat` lists only `pkg/engine/actions.go`, `pkg/engine/combat.go`, `pkg/engine/stack.go` and this task file.
6. `go -C W tool task check` and `go -C W tool task sim` pass.

The reviewer also compares a starter `-events` file from both builds (#68 case (b)); it must differ only in each game's `match-start` `revision` stamp.

## Allowed Files
- `pkg/engine/actions.go`
- `pkg/engine/combat.go`
- `pkg/engine/stack.go`

## Done When
- [ ] Cases 1–6 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-b.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-28, lead-architect (wave B.2 run, session 4): wrote this task together with T-0092, the two halves of directive #68 part (d) under the D2 size rule (see T-0092's Inputs). `queued` until T-0092 is `done`; the session that dispatches it checks the line numbers above against the run branch first. Harness, tier `flash`; every command here is on the headless allow-list.
- 2026-10-28, lead-architect (wave B.2 run, session 5): T-0092 is `done` (2c78e88), so this task moves to `ready`. `main` has nothing the run branch lacks, so no sync. Checked the line numbers above against the run branch at 3b503e5: all match (`actions.go:51-52`, `actions.go:177-179`, `combat.go:9`, `stack.go:27-28`), and no test file under `pkg` or `cmd` mentions Berserk, a secondary or cleave, so case 3 needs no test edit. Waiting directives checked: none has triggered (#8: no `deck.go`; #15, #20 (b), #58, #59 (2), #61, #71: no such hook, shape, claim or card here; #6, #7: mode is `serial`). #69 and #70 stay open for the sessions after this one. Branch `m2/T-0093-secondary-fields` created from the run branch; next: `kernel-engineer` on Flash.
