---
id: T-0094
title: Attack legality lives in combat.go, (g *Game) checkAttack (directive #69, harness, flash, pure move)
kind: harness
tier: flash
wave: B.2 (board directive #69, the user's code review, before wave C)
status: done
branch: m2/T-0094-check-attack
depends_on: [T-0093]
claims: [func:checkAttack]
cards: []
requirements: []
---

## Goal
Every rule for whether an Attack may be declared sits in `combat.go`, beside `canDeclareAttack`, `CanAttack`, `declareAttack` and `resolveAttack`. The Attack case of `check` becomes `(g *Game) checkAttack(a Action) error` in `combat.go`, and `check` calls it, so a reader of `combat.go` sees that an exhausted attacker cannot attack. `readyAsset` stays in `actions.go`, because Generate and March use it too. This is a pure move: same tests, same sim output.

## Inputs
- Board directive #69 (user code review, 2026-10-28: "DeclareAttack doesn't check readiness?").
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6b (wave B.2).
- Rules the moved code implements, unchanged: `FR-CBT-001`.
- Why it is a pure move: `checkAttack` runs the statements of `check`'s Attack case in the same order (`readyAsset`, `canDeclareAttack`, the defender lookup and `CanAttack`), with the same error texts. `check` runs its shared checks (match over, Main Phase, active player, empty stack) before the switch, as before, and returns `nil` after the switch for every case that doesn't return, so `return g.checkAttack(a)` returns exactly what the old case did. No decision or random draw is involved. `git grep` over `pkg` and `cmd` finds no test that names `check`'s Attack case; `pkg/sim`'s `(m *Monitors) checkAttack` is a method of another type in another package and is untouched.

### Where it is today (at `autoloop/m2-wave-b.2` 67b0a3a)
| Where | Today |
| :--- | :--- |
| `pkg/engine/actions.go:109-110` | `check`'s doc comment and signature |
| `pkg/engine/actions.go:164-175` | `check`'s `case ActAttack:`: `readyAsset`, then `canDeclareAttack`, then the defender lookup and `CanAttack` |
| `pkg/engine/actions.go:190` | `readyAsset`, also called by Generate (`:127`) and March (`:157`) |
| `pkg/engine/combat.go:1-15` | the file comment ("attack legality (FR-CBT-001) and the 3-step combat sequence (FR-CBT-002)") and `declareAttack` |

## Edits (lead decisions; copy exactly)

### 1. `pkg/engine/actions.go`
In `check`, replace the body of `case ActAttack:` (the eleven lines from `att, err := g.readyAsset(a)` through the closing `}` of the `illegal combat target (FR-CBT-001)` check) with one line, so the case reads:
```go
	case ActAttack:
		return g.checkAttack(a)
```
Nothing else in `actions.go` changes; `readyAsset` stays where it is.

### 2. `pkg/engine/combat.go`
Right after the `package engine` line and its blank line, before `declareAttack`'s doc comment, add:
```go
// checkAttack returns nil if a, an Attack, is legal right now (FR-CBT-001):
// the attacker is a Readied asset of a.Player's on the battlefield
// (readyAsset, so an Exhausted asset cannot attack), it is a Unit or
// Headquarters in its combat zone (canDeclareAttack), and the defender is one
// it may attack (CanAttack). check calls it after the checks every action
// shares.
func (g *Game) checkAttack(a Action) error {
	att, err := g.readyAsset(a)
	if err != nil {
		return err
	}
	if !canDeclareAttack(att) {
		return illegal("%s cannot attack from %s (FR-CBT-001)", att, att.Zone)
	}
	def := g.Asset(a.Target)
	if def == nil || !g.CanAttack(att, def) {
		return illegal("illegal combat target (FR-CBT-001)")
	}
	return nil
}

```
`combat.go` needs no new import: `illegal` is in `actions.go`, the same package.

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. Run each command on its own, pointed at your worktree (W). Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#69 case (a)): `git -C W diff autoloop/m2-wave-b.2 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#69 case (b)): `go -C W tool task sim-compare -- autoloop/m2-wave-b.2` prints `[PASS]`.
3. **`check`'s Attack case is one call** (#69 case (c)): `git -C W grep -n -A1 "case ActAttack:" -- pkg/engine/actions.go` prints two matches: `Apply`'s, followed by `g.declareAttack(a)`, and `check`'s, followed by `return g.checkAttack(a)`.
4. **One home for the rule:** `git -C W grep -n "checkAttack" -- pkg/engine` prints exactly three lines: the doc comment's first line and the `func (g *Game) checkAttack` line in `combat.go`, and the one call in `actions.go`. `git -C W grep -n "canDeclareAttack(att)" -- pkg/engine` prints exactly one line, in `combat.go`.
5. **The doc names readiness, the combat zone and the rule** (#69 case (c)): `go -C W doc -u ./pkg/engine Game.checkAttack` prints the doc comment above, which names `Readied`, `combat zone` and `FR-CBT-001`.
6. **Only these edits:** `git -C W diff autoloop/m2-wave-b.2 --stat` lists only `pkg/engine/actions.go`, `pkg/engine/combat.go` and this task file.
7. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/actions.go`
- `pkg/engine/combat.go`

## Done When
- [x] Cases 1–7 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-b.2` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-28, lead-architect (wave B.2 run, session 6): wrote the task from board directive #69, now that #68 is applied in full (T-0089 to T-0093), so the move carries no Berserk code. Harness, so it skips `tests` and goes straight to `ready`; tier `flash` as #69 gives; every command here is on the headless allow-list. `checkAttack` goes first in `combat.go`, ahead of `declareAttack`, so the file reads in the order of an attack: legal, declared, resolved. `main` has no commits the run branch lacks, so no sync. Waiting directives checked: none has triggered (#8: no `deck.go`; #15, #20 (b), #58, #59 (2), #61, #71: no such hook, shape, claim or card here; #6, #7: mode is `serial`). #70 stays open for the next session. Next: `kernel-engineer` on Flash.
- 2026-10-28, kernel-engineer (Flash): pure move implemented and verified.
  - Changes: Moved the body of check's ActAttack case into `(g *Game) checkAttack(a Action) error` in `pkg/engine/combat.go`, replacing the case body in `pkg/engine/actions.go` with `return g.checkAttack(a)`. Doc comment names Readied, combat zone, and FR-CBT-001.
  - Reused: `g.readyAsset`, `canDeclareAttack`, `g.Asset`, `g.CanAttack`, `illegal`.
  - Acceptance cases:
    - Case 1: `git -C W diff autoloop/m2-wave-b.2 -- '*_test.go'` printed nothing; `go -C W test ./...` passed.
    - Case 2: `go -C W tool task sim-compare -- autoloop/m2-wave-b.2` printed `[PASS] starter and stress sims match autoloop/m2-wave-b.2 byte for byte, exit codes included`.
    - Case 3: `git -C W grep -n -A1 "case ActAttack:" -- pkg/engine/actions.go` printed Apply declareAttack and check checkAttack.
    - Case 4: `git -C W grep -n "checkAttack" -- pkg/engine` printed 3 lines (actions.go:165, combat.go:5, combat.go:11); `git -C W grep -n "canDeclareAttack(att)" -- pkg/engine` printed 1 line (combat.go:16).
    - Case 5: `go -C W doc -u ./pkg/engine Game.checkAttack` printed doc comment naming Readied, combat zone, and FR-CBT-001.
    - Case 6: `git -C W diff autoloop/m2-wave-b.2 --stat` showed only actions.go, combat.go, and this task file.
    - Case 7: `go -C W tool task check` passed all stages (fmt, vet, test, lint --all); `go -C W tool task sim` passed on starter decks (200 games, all invariants holding).
  - Ready for review.
- 2026-10-28, lead-architect (wave B.2 run, session 6, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-b.2` as 38bd2b8. Flash run `eeeb07bc-6f51-4622-9507-ac9c61e1161e`: exit 0, `status: SUCCESS`, 236 s, no refusals (the JSON has no `denied_actions` entry, `cli.log` has no `soft-denying` line for it and the transcript no `permission check failed`). Review of `git diff autoloop/m2-wave-b.2...m2/T-0094-check-attack` (d4ceecf):
  - Scope: only `actions.go`, `combat.go` and this file changed, and both code edits match the task's text line for line. No test file changed.
  - Pure move: `checkAttack` holds the old case's statements in their order with the same error texts, and `check`'s Attack case is `return g.checkAttack(a)`; every other case still falls through to `return nil`.
  - Fences and reuse: no card name, no new rule, nothing duplicated: `canDeclareAttack(att)` and the attack-target check now appear once, in `combat.go`. `readyAsset` stays in `actions.go` for Generate and March, as #69 says.
  - Gates re-run in a scratch worktree of the task branch, each with exit 0: case 1 (no `_test.go` diff); case 3 shows `Apply`'s `g.declareAttack(a)` and `check`'s `return g.checkAttack(a)`; case 4 prints exactly three `checkAttack` lines and one `canDeclareAttack(att)` line, in `combat.go`; case 5's `go doc` names Readied, combat zone and `FR-CBT-001`; case 6 lists exactly the three files; `go tool task sim-compare -- autoloop/m2-wave-b.2` `[PASS]` (starter and stress byte-identical); `go tool task check` passes; `go tool task sim` clean, 200 games, P0 63.5% / P1 34.5% / draws 2.0% (unchanged), 30/50 Animus and 25/50 Path main cards executable.
  - The merge's tree equals the task tip's, so the gates hold on the run branch. Directive #69 is applied.
