---
id: T-0290
title: "The `ability` package doc and `doc.go`'s Files list say what wave H.2 built (directive #176 (a), (b); wave H.2 harness, flash, comments only)"
kind: harness
tier: flash
wave: H.2 (board directive #176)
status: in-progress
branch: m2/autoloop-wave-h.2/T-0290-ability-and-engine-doc
depends_on: [T-0289]
claims: [file:pkg/engine/ability/ability.go, file:pkg/engine/doc.go]
cards: []
requirements: []
---

## Goal
The two package comments the write-up (Deliverable 7) reads say what wave H.2 built. The `ability` package doc ([ability.go:1-3](../../../pkg/engine/ability/ability.go)) is the first thing `go doc ./pkg/engine/ability` prints. It says "plain data, no game logic" and names neither the keyword set, its line parser and its fold (`keywords.go`, ADR 0008), nor the small readers the types carry, so it reads as narrower than the design note's Boundaries row (#176 (a)). `pkg/engine/doc.go`'s Files list doesn't name what #169 to #175 added: `sourceAsset` (T-0281), `StackItem.FromKeyword` and `ability.KeywordTrigger` (T-0283), `Card.Deployable` and `pileCard` (T-0286), and the loader's walk through `Effect.Parts` (T-0287) (#176 (b)). Comments only: no code changes, and match output is byte for byte the same. When this merges, #176 is `applied`.

## Inputs
- Board directive #176 (structural review 14, finding 7), approved by the user 2026-11-19, `open`, wave H.2's last. The design note and `docs/architecture/engine-structs.md` parts of (b) are the lead's and are done on the run branch: b3bb4d3 (`docs(architecture): the design note and engine-structs say what wave H.2 built (directive #176 (b))`). Item (a) and the `doc.go` part of (b) are in `pkg/engine`, which only an engineer edits, so they are this task.
- Where it is today (at `autoloop/m2-wave-h.2` b3bb4d3): `ability.go` lines 1-3 (the package doc's first paragraph); `doc.go` lines 113, 115-116 and 128-131 (the Files list's `zones.go`, `stack.go`, `state.go`/`card.go` and `cardload.go` sentences).
- Design (lead decisions): each new sentence names a declaration that exists where it says: `Keywords`, `ParseKeywordLine` and `Keywords.Add` in `ability/keywords.go`; `Effect.Parts`, `Effect.ReadsZoneRef`, `Filter.InPile`, `Cost.Counted`, `Target.AnotherSlot`, `Step.DealerSlot`, `WhileCornered.Add`, `Aggressive`, `Trigger.FromKeyword` and `KeywordTrigger` in `ability/ability.go`; `sourceAsset` in `state.go`, `Card.Deployable` in `card.go`, `pileCard` in `zones.go`, `StackItem.FromKeyword` in `stack.go`, and the loader's checks reading `Parts` in `cardload.go`. The rest of both comments stays: the enum rules paragraph in `ability.go`, and `doc.go`'s "What runs now", which already covers each behavior.
- Old names (#176's acceptance): `git grep` finds no `itemSource`, `isFencesPath`, `isRulesPath`, `KeywordNone`, `KeywordAggressive` or old `Keyword` type or field under `pkg`, `cmd`, `tools` or the two architecture docs at b3bb4d3, and these edits add none. The one other hit, `docs/design/outcomes/2026-11-09-wave-f-same-zone-source.md:55` (`itemSource` in a ruling's text), is a `game-designer` record of a ruling made on the code of the day, like ADR history; neither the lead nor this task edits it.
- Reuse search (D2.3): none needed; nothing new is declared.
- Size (D2.4): two non-test files.
- **Expected output difference:** none: `go tool task sim-compare` passes.
- The lead dry-ran these exact edits in the run worktree at b3bb4d3's tree: `gofmt -l pkg/engine` clean, `go vet ./pkg/engine/...` clean, `go tool task check` passed, `go tool task sim-compare -- autoloop/m2-wave-h.2` `[PASS]`, and `go doc ./pkg/engine/ability` prints the new paragraph second. Blobs: `pkg/engine/ability/ability.go` 0ed49ecb, `pkg/engine/doc.go` f3c95d1b.

## Commits
1. `docs(engine): the ability package doc and doc.go Files list say what wave H.2 built (T-0290)` with the edits below.

There is no `test(T-0290)` commit: the change is comments only, and the Acceptance Cases are checked by command. No double quotes in the message: the allow-list refuses an escaped one.

## Edits (lead decision; copy exactly, `//` comment lines as in the surrounding code)
### `pkg/engine/ability/ability.go`
```diff
@@ -1,6 +1,16 @@
 // Package ability is the shape vocabulary for printed card abilities: plain
-// data, no game logic and no card names (the fences linter enforces the
-// latter). New types here arrive only through a `shape` task.
+// data types and enums, no game logic and no card names (the fences linter
+// enforces the latter). New types here arrive only through a `shape` task.
+//
+// It also holds the keyword set that every printed keyword and keyword grant
+// names, Keywords, with its printed names, its line parser (ParseKeywordLine)
+// and its one fold (Keywords.Add), in keywords.go. Its types carry a few small
+// readers, so every caller asks the same question the same way: Effect.Parts,
+// the one walk into an effect's modes, and Effect.ReadsZoneRef, Filter.InPile,
+// Cost.Counted, Target.AnotherSlot, Step.DealerSlot and WhileCornered.Add. A
+// keyword that is a triggered ability is a Trigger built by one constructor,
+// Aggressive, which names its keyword in Trigger.FromKeyword, a
+// KeywordTrigger.
 //
 // A new enum here follows one of three rules, chosen by what its zero value
 // means:
```
### `pkg/engine/doc.go`
```diff
@@ -110,10 +110,14 @@
 //
 // Files:
 //   - setup.go: Config and NewGame. deck.go: DeckList, validation, loading.
-//   - turn.go: beginTurn and endTurn. zones.go: moving cards and assets, deploying a card, and the zone a shape names (zoneMatches).
+//   - turn.go: beginTurn and endTurn. zones.go: moving cards and assets,
+//     deploying a card, the one lookup of a card by ID in a hand or discard
+//     pile (pileCard), and the zone a shape names (zoneMatches).
 //   - actions.go: Action, Apply, checks and LegalActions. combat.go: attack
-//     legality and combat resolution. stack.go: the Resolution Stack, and
-//     StackLabel, the one wording of what a stack item is.
+//     legality and combat resolution. stack.go: the Resolution Stack, its
+//     StackItem (whose FromKeyword names the keyword its trigger came from, an
+//     ability.KeywordTrigger), and StackLabel, the one wording of what a stack
+//     item is.
 //   - trigger.go: the pending-trigger buffer and firing. placement.go: placing
 //     an effect, its modes and then its targets, the card it shows as its
 //     source (shownCard, SourceCard) and its name (sourceName), and whether an
@@ -125,10 +129,14 @@
 //     an ability's filter, and a continuous ability's filter read from its
 //     source (matchesFromSource). effects.go: running a resolved ability's steps.
 //     freedeploy.go: free deploys and the look at the top cards.
-//   - state.go: core state types and lookups, and exhausting an asset. card.go: card types.
+//   - state.go: core state types and lookups, the one lookup of a source that
+//     may have left the battlefield (sourceAsset: the asset, or its last-known
+//     record), and exhausting an asset. card.go: card types, and which of them
+//     are deployed rather than played (Card.Deployable).
 //     cardload.go: the CSV loader, and checkShapeReads, which rejects a definition
-//     that sets a field where nothing reads it. The keyword set, its names, line
-//     parser and fold are in pkg/engine/ability (keywords.go).
+//     that sets a field where nothing reads it, walking an effect's modes through
+//     ability's Effect.Parts. The keyword set, its names, line parser and fold
+//     are in pkg/engine/ability (keywords.go).
 //   - events.go: Event, the Observer, match records and emitting. choice.go: decisions,
 //     and the askFor scope that names their source card.
 //     snapshot.go: Snapshot and its deep copies. buildinfo.go: BuildInfo.
```

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The `ability` doc names the keyword set and its readers (#176 (a)):** each of `git -C W grep -n "It also holds the keyword set" -- pkg/engine/ability/ability.go`, `git -C W grep -n "its one fold (Keywords.Add), in keywords.go" -- pkg/engine/ability/ability.go` and `git -C W grep -n "Cost.Counted, Target.AnotherSlot, Step.DealerSlot and WhileCornered.Add" -- pkg/engine/ability/ability.go` prints exactly one line, and `git -C W grep -n "data, no game logic and no card names" -- pkg/engine/ability/ability.go` prints nothing.
2. **`doc.go`'s Files list names what #169 to #175 added (#176 (b)):** each of `git -C W grep -n "(sourceAsset: the asset" -- pkg/engine/doc.go`, `git -C W grep -n "(Card.Deployable)" -- pkg/engine/doc.go`, `git -C W grep -n "pile (pileCard)" -- pkg/engine/doc.go`, `git -C W grep -n "StackItem (whose FromKeyword" -- pkg/engine/doc.go`, `git -C W grep -n "ability.KeywordTrigger), and StackLabel" -- pkg/engine/doc.go` and `git -C W grep -n "ability's Effect.Parts" -- pkg/engine/doc.go` prints exactly one line.
3. **No old name (#176's acceptance):** `git -C W grep -n -w -F -e itemSource -e isFencesPath -e isRulesPath -e KeywordNone -e KeywordAggressive -e ability.Keyword -e Trigger.Keyword -e StackItem.Keyword -e StackView.Keyword -- pkg cmd tools docs/architecture/m2-ability-model.md docs/architecture/engine-structs.md` prints nothing.
4. **Exactly the edits:** `git -C W diff autoloop/m2-wave-h.2 --numstat -- pkg docs` prints exactly two lines, `12	2	pkg/engine/ability/ability.go` and `14	6	pkg/engine/doc.go`; `git -C W rev-parse HEAD:pkg/engine/ability/ability.go` prints `0ed49ecb1f202cd04aca28e9e8880835db1814bd`, and `git -C W rev-parse HEAD:pkg/engine/doc.go` prints `f3c95d1b5c714bc9354a69013ab9b3aac153f99d`.
5. **Nothing else changed:** `git -C W diff autoloop/m2-wave-h.2 --stat` lists only `pkg/engine/ability/ability.go`, `pkg/engine/doc.go` and this task file.
6. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-h.2` prints `[PASS]`.
7. `gofmt -l W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/ability/ability.go` (the package doc edit above only)
- `pkg/engine/doc.go` (the edit above)

## Tests
- None: comments only. No test file changes.

## Done When
- [ ] Cases 1-7 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-19, lead-architect (wave H.2 run, session 10): written in full against b3bb4d3 for #176 (a) and the `doc.go` part of (b); the design note and `engine-structs.md` parts of (b) are b3bb4d3. `main` (1859fe4) has nothing the run branch lacks, so no sync. Not `WAVE-COMPLETE` before the task: #176 is `open` and is this wave's (the run file's done-when names #169 to #176). This is a harness task, so it skips `tests` and goes straight to `ready`. Tier is `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list: one per line, no pipes, no `$`, and patterns with parentheses are quoted. Waiting directives checked: #6, #7 (mode `serial`), #15 (2) (no draw or end-of-turn hook), #71, #91 (2), #106, #120, #142 (wave H.2 has no cards): none has fired. Next: `kernel-engineer` on Flash.
