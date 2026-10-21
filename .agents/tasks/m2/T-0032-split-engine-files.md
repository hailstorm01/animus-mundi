---
id: T-0032
title: Split package engine's files by responsibility (pure move)
kind: harness
tier: sonnet
wave: none (board directive #26, cleanup 1 of 4, before cardtown play)
status: in-progress
branch: m2/T-0032-split-engine-files
depends_on: [T-0031]
claims: [file:pkg/engine/setup.go, file:pkg/engine/zones.go, file:pkg/engine/combat.go, file:pkg/engine/targeting.go, file:pkg/engine/effects.go, file:pkg/engine/cardload.go]
cards: []
requirements: []
---

## Goal
Each file in `pkg/engine` holds one job. The user's audit (2026-10-21, board directive #26) found these files mixing several:
- `turn.go` holds deck type, match setup, card movement and turn phases.
- `state.go` holds the `Event` type and the battlefield movement helpers.
- Combat is split between `actions.go` and `stack.go`.
- `trigger.go` holds trigger placement, target choice and effect execution together.
- `card.go` holds the card types and the CSV loader.

This task moves declarations between files inside package `engine`. Nothing is renamed, no signature changes, and no logic changes. That way wave B (directive #15's turn-boundary triggers, new hooks and new steps) lands in files with one job.

## Inputs
- Board directive #26 and the audit it records.
- The current layout: `grep -n "^func \|^type \|^var \|^const " pkg/engine/*.go`.
- T-0024 was a pure move between packages, and this task follows its method and acceptance style.
- `pkg/engine/doc.go`. Its "Not built yet" text still says targeted abilities are unbuilt.
- Rulings that apply: none.

### Target layout (lead decision)
Each line names a file and what it holds after the move. Declarations not listed stay where they are.

| File | Holds |
| :--- | :--- |
| `setup.go` (new) | `Config`, `NewGame` (from `turn.go`) |
| `deck.go` | + `DeckList` (from `turn.go`) |
| `turn.go` | only `beginTurn` and `endTurn` |
| `zones.go` (new) | Card and asset movement:<br>- `shuffle`, `moveTopToHand`, `draw` (from `turn.go`)<br>- `scratch` (from `stack.go`)<br>- `leaveBattlefield`, `removeFromBattlefield`, `placeOnBattlefield`, `remove` (from `state.go`)<br>- `handCard` (from `actions.go`) |
| `events.go` | + `Event` (from `state.go`) |
| `combat.go` (new) | Attack legality and resolution:<br>- `declareAttack`, `canDeclareAttack`, `CanAttack`, `hasGuardian`, `CanCleave`, `hasCleaveTarget`, `legalSecondary` (from `actions.go`)<br>- `combatLegalZone`, `resolveAttack`, `cleave` (from `stack.go`) |
| `targeting.go` (new) | `chooseTargets`, `legalTarget`, `matchesFilter`, `assetsByID` (from `trigger.go`) |
| `effects.go` (new) | `resolveAbility`, `value`, `itemSource`, `stepAssets`, `relocateSelf` (from `trigger.go`) |
| `trigger.go` | `pendingTrigger`, `fireDeploy`, `fireKilled`, `fireHook`, `deployedTo`, `fireAttacks`, `flushTriggers`, `accepts` |
| `cardload.go` (new) | `keywordLine`, `ParseKeywordLine`, `LoadCards`, `loadCards`, `indexOf`, `loadSheets`, `readCards` (from `card.go`) |
| `card.go` | `CardType`, `Keywords`, `Card`, `HasSubtype`, `Supported`, `scratchFace` |
| `state.go` | The core state types and lookups: `PlayerID`, `Zone`, `Phase`, `Instance`, `Asset`, `Player`, `Outcome`, `Game`, `EffectiveState`, `lastKnownRecord`, `effectiveState`, `LastKnown`, `LastKnownEffective`, `newID`, `Assets`, `Asset`, `TheaterOf` |
| `actions.go` | `ErrIllegal`, `illegal`, `ActionKind`, `Resource`, `Action`, `Apply`, `check`, `readyAsset`, `canYield`, `LegalActions` |
| `stack.go` | `StackKind`, `StackItem`, `push`, `resolveStack`, `resolve`, `resolveDeploy`, `resolveMarch` |

Moving rules:
- **Comments travel with their declarations.**
- **Each file starts with a one-line comment** saying what it holds, citing the FR IDs where they fit (for example `// Combat: attack legality (FR-CBT-001) and the 3-step combat sequence (FR-CBT-002).`).
- **Imports are fixed up**, and `gofmt` is clean.
- **Ordering:** within a file, keep the order the declarations had in their source file. Where several sources merge, follow the table's order.
- **`doc.go`** gets two comment-only changes:
  - a short "Files" paragraph mapping each file to its job, in a sentence or a line each;
  - a refreshed "What runs now" / "Not built yet". Wave A2 is built: chosen targets, two-stage verification, Ward, DealDamage, Heal, AddArmor, the deploy-zone gate, source-power values, last-known effective state, "all" steps, optional triggers, "up to N" and "other". Maneuvers, activated and continuous abilities, Headquarters abilities, Unbowed, Cornered, Heal N, Regroup N, replacements beyond the damage pipeline and match clocks remain unbuilt.
- **File-count exception (lead):** this task changes 14 non-test files, against the rule of four. Like T-0024, it is a pure move, so the size rule's purpose (limit what one review must check) is met by checking that every moved body is unchanged.
- **Not in this task:**
  - removing duplicates (T-0033, T-0034);
  - splitting `cmd/cardtown/main.go` (T-0035);
  - `damage.go`'s `end()`;
  - `events.go`'s snapshot code;
  - any rename.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests. These are checks the engineer runs before and after and records in the Log, and the reviewer re-runs.
1. **Same declarations:**
   - The sorted list of top-level declarations in `pkg/engine` non-test files is identical before and after, apart from file names. Use the output of `grep -h "^func \|^type \|^var \|^const " pkg/engine/*.go` on non-test files, sorted.
   - `go doc -all ./pkg/engine | sort` is identical before and after, apart from the package comment's changed lines from `doc.go`.
2. **Bodies unchanged:** `git diff -M --stat` and `git diff --color-moved=zebra --color-moved-ws=allow-indentation-change rework/executable-rules...HEAD -- pkg/engine` show only moved blocks, the new per-file header comments, import blocks and the `doc.go` comment change. List any other changed line in the Log, with its reason.
3. **Tests unchanged:** `go test -list . ./...` gives the same set of test names before and after. No `_test.go` file is changed.
4. **Behavior unchanged:**
   - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` and bare `cardtown sim`, each under `-policy progress` and `-policy random`, are byte-identical to a baseline CLI built from `rework/executable-rules` (in a separate `git worktree`) before any edit.
   - So are `-games 20 -events` files, once each `match-start`'s `revision` and `modified` are normalized, and `-log -seed 7` output.
5. **Gate:** `go tool task check` passes.
6. **Layout:** every declaration listed in the table is in its target file. `turn.go` contains only `beginTurn` and `endTurn`.

## Allowed Files
- `pkg/engine/turn.go`, `state.go`, `actions.go`, `stack.go`, `trigger.go`, `card.go`, `deck.go`, `events.go`, `doc.go`
- new: `pkg/engine/setup.go`, `zones.go`, `combat.go`, `targeting.go`, `effects.go`, `cardload.go`

## Done When
- [ ] Acceptance cases 1 to 6 hold, with their commands and results in the Log
- [ ] `go tool task check` passes

## Log
- 2026-10-21 lead-architect: Wrote the task from board directive #26 (the user's audit), with status `ready` (harness, sonnet). It is cleanup 1 of 4, ahead of `cardtown play` (design note Deliverable 4a). File-count exception recorded above.
