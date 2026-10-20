---
id: T-0016
title: Refresh stale milestone notes in engine comments and messages
kind: harness
tier: flash
wave: none (board directive #17, audit finding 4)
status: done
branch: m2/T-0016-refresh-milestone-notes
depends_on: []
claims: [file:pkg/engine/doc.go, comments:engine-milestone-wording]
cards: []
requirements: []
---

## Goal
Engine comments and messages describe the kernel as it is now in M2, not "M1 scope" or "until M3". The open survey question R8 also gets its `OPEN(SVY-R8)` marker. Only comments and one error message change, with no behavior change.

## Inputs
- Board directive #17 in `.agents/tasks/m2/BOARD.md`.
- What exists now, so the wording is true (verify each point with `go doc ./pkg/engine/ability` and `pkg/engine/cards`, and don't add anything beyond it):
  - Ability shapes are plain data in `pkg/engine/ability`, and card definitions keyed by card name are in `pkg/engine/cards`. Engine code never names a card (`cardtown lint fences`).
  - Triggered abilities on self hooks (`HookDeploy`, `HookKilled`, `HookAttacks`) run through a pending-trigger buffer, in APNAP order with cascades (FR-TRG-002 to FR-TRG-004).
  - Per-seat `Chooser` decisions with timeout fallbacks (FR-MOD-001).
  - The untargeted steps `OpScratch`, `OpDraw`, `OpReturnToHand` and `OpPutOnDeck`, the `TurnYours` gate, `SubjectSelf` (tracked in discard, FR-ZON-004) and `SubjectOpponents`.
  - Not built yet: targeted abilities, Maneuvers, activated and continuous abilities, Headquarters abilities, Unbowed, Cornered, Heal N, Regroup N, replacement effects beyond the damage pipeline, and match clocks. Cards whose text needs any of these load with `Card.Unsupported` set and cannot be deployed. `cards.NotYetExecutable` lists the in-scope M2 cards still in that state.
- Survey question R8 (`docs/design/m2_starter_set_ability_survey.md` line 67) is open. `cardtown lint fences` reads open `SVY-` items from that table.
- Rulings that apply: none.

### Changes
1. `pkg/engine/doc.go`: keep the first paragraph. Replace the "M1 scope" list and the "Out of scope until the ability model lands (M3)" paragraph with two lists, "What runs now" (the M1 list, still true, plus the M2 items above) and "Not built yet" (the list above). Replace the last paragraph with the marker convention: `OPEN(KIM-…)` points at `docs/frd/keyword_interaction_matrix.md`, `OPEN(SVY-…)` points at the survey's ruling questions, and `cardtown lint fences` checks both against those files. Name no milestone after M2.
2. `pkg/engine/actions.go:151`: the message becomes `"%s needs rules the kernel does not implement yet"`.
3. `pkg/engine/card.go:64`: "the M1 kernel" becomes "the kernel".
4. `pkg/engine/damage.go:23`: "No M1 card redirects damage." becomes "No executable card redirects damage yet."
5. `pkg/engine/layers.go:4-5`: "M1 has no type grants…" becomes "Nothing grants types (Layer 2) or keywords (Layer 3), and there are no auras (Layer 4) yet, so…". Keep the rest of the sentence.
6. `pkg/engine/turn.go:112` and `:142`: "No M1 card has…" becomes "No executable card has… yet". Leave the FR citations as they are.
7. `pkg/engine/stack.go`, in `resolveAttack`: on the line before the `def.Controller == it.Controller` check, add `// OPEN(SVY-R8): fizzles if the defender changed controller before resolution.`
8. **Authorized edit to a protected test:** in `pkg/engine/playout_test.go:85`, change the `t.Logf` text "executable by the M1 kernel" to "executable by the kernel". This is the only test change, and it lands in a `test(T-0016)` commit before the other changes. Leave `playout_test.go:9` ("every M1 keyword") alone, because it accurately names the M1 keyword set.

## Acceptance Cases (harness tasks only; written by the lead)
Comments and messages only, so these are checked by command, not by new tests:
- `grep -rnw 'M1\|M3' pkg/engine --include=*.go`, excluding `_test.go` files → no output.
- `go doc ./pkg/engine` → mentions the `ability` and `cards` packages and both marker forms, and does not say "M1 scope" or "M3".
- `go run ./cmd/cardtown lint fences` → passes and accepts `OPEN(SVY-R8)`. Then temporarily change the marker to `OPEN(SVY-R99)` and confirm fences reports it as not in the survey. Revert, and don't commit the experiment.
- `go test ./...` → passes, with no test added, removed or changed except `playout_test.go:85`.
- `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` → the summary lines (executable counts, win rates, match length, sorted end reasons) are identical before and after.

## Allowed Files
- `pkg/engine/doc.go`, `actions.go`, `card.go`, `damage.go`, `layers.go`, `turn.go` and `stack.go` (comments, plus the one message in `actions.go`)
- `pkg/engine/playout_test.go` (line 85 only, in the `test(T-0016)` commit)

## Done When
- [x] Every acceptance case above holds
- [x] `go tool task check` passes
- [x] The diff contains no change other than comments, the `actions.go` message and `playout_test.go:85`

## Log
- 2026-10-20 lead-architect: Wrote the task from board directive #17, with status `ready` (harness, flash: Antigravity). Size exception: seven non-test files exceed the D2.4 limit of four, but every change is a comment or one message string. The user's directive asks for one task, so it isn't split.
- 2026-10-20 kernel-engineer: Implemented. playout_test.go:85 committed in test(T-0016) updating log message from "the M1 kernel" to "the kernel". doc.go: replaced "M1 scope" and "M3" paragraphs with "What runs now" and "Not built yet" lists, and updated marker documentation to include OPEN(SVY-...) and cardtown lint fences. actions.go:151: changed message to "%s needs rules the kernel does not implement yet". card.go:64: changed comment from "the M1 kernel" to "the kernel". damage.go:23: changed comment to "No executable card redirects damage yet." layers.go:4-5: updated comment to "Nothing grants types (Layer 2) or keywords (Layer 3), and there are no auras (Layer 4) yet, so...". turn.go:112, 142: updated comments to "No executable card has start-of-turn triggers yet." and "no executable card has end-of-turn triggers yet." stack.go:127: added "// OPEN(SVY-R8): fizzles if the defender changed controller before resolution." Verification: grep for 'M1|M3' across pkg/engine excluding _test.go returned 0 results; go doc ./pkg/engine confirms ability and cards packages, both marker types, and no M1 scope or M3 wording; cardtown lint fences passed with OPEN(SVY-R8) and failed when temporarily set to OPEN(SVY-R99); go test ./... passed; go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH produced identical summary metrics before and after (P0 19/50, P1 6/50, P0 wins 51.0%, P1 wins 16.5%, draws 32.5%, avg length 27.3 turns, HQ destroyed 135, simultaneous 65); go tool task check passed.
- 2026-10-20 lead-architect: Code review: **approved**, status `done`.
  - Scope: a445721 (`test(T-0016)`) changes only `playout_test.go:85`, the authorized edit. 59ba46f changes only comments in the seven allowed files, plus the `actions.go` message. No logic changed.
  - The `doc.go` lists match what `pkg/engine/ability` defines. Maneuvers are correctly listed as not built, because `check` still accepts only Units and Structures for deploy.
  - Acceptance cases re-run by the reviewer:
    - No `M1`/`M3` remains in non-test engine code.
    - `go doc ./pkg/engine` names `ability`, `cards` and both marker forms.
    - Fences passes with `OPEN(SVY-R8)` and reports `OPEN(SVY-R99)` as not in the survey (the change was reverted).
    - The sorted starter-sim summary is identical before and after.
    - `go tool task check` exits 0.
