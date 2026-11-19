---
id: T-0279
title: "`doc.go`'s Files list says what wave H and #159 to #162 built (directive #163 (f); wave H.1 harness, flash, comments only)"
kind: harness
tier: flash
wave: H.1 (board directive #163)
status: ready
branch: m2/autoloop-wave-h.1/T-0279-doc-go-files
depends_on: [T-0278]
claims: [file:pkg/engine/doc.go]
cards: []
requirements: []
---

## Goal
The package comment in `pkg/engine/doc.go` says what wave H and directives #159 to #162 built. Its Files list says `replace.go` holds only destruction replacement ([doc.go:132](../../../pkg/engine/doc.go)), but T-0269 put damage redirection (`redirectDamage`) there too (#163 (f)). The same list, and one Steps line, also miss what #159 to #162 added: `OpGrantKeyword` grants `Step.Keywords` (T-0272), `placement.go` holds `sameZoneLock` (T-0275), `filter.go` holds `matchesFromSource` (T-0276), `cardload.go` holds `checkShapeReads` (T-0278), and `ability/keywords.go` holds the fold (T-0272). Comments only: no code changes, and match output is byte for byte the same. When this merges, #163 is `applied`.

## Inputs
- Board directive #163 (structural review 13, finding 5), approved by the user 2026-11-18, `open: before the next Deliverables task`. Items (a) to (e) and (g) are in the design note, which the lead may edit, and are done on the run branch: b5779c6 (`docs(architecture): the design note says what wave H and #159 to #162 built`). Item (f) is in `pkg/engine`, which only an engineer edits, so it is this task.
- Where it is today (at `autoloop/m2-wave-h.1` b5779c6): `doc.go` line 43 (Steps: `OpGrantKeyword (a keyword granted until end of turn, FR-LAY-001)`) and lines 117-133 (the Files list: `trigger.go`/`placement.go`, `filter.go`, `cardload.go`, `replace.go`).
- Design (lead decisions): each new sentence names the function the task that built it added, in the Files list's style. Nothing else in `doc.go` changes: its "What runs now" bullets already cover watching triggers, auras' keyword and Cornered grants and damage redirection.
- Reuse search (D2.3): none needed; nothing new is declared.
- Size (D2.4): one non-test file.
- **Expected output difference:** none: `go tool task sim-compare` passes.
- The lead dry-ran this exact edit in a throwaway worktree of the run branch at b5779c6: `gofmt -l pkg/engine` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-h.1` `[PASS]`. Blob: `pkg/engine/doc.go` 8de0e5d.

## Commits
1. `docs(engine): doc.go Files list names redirectDamage, sameZoneLock, matchesFromSource and checkShapeReads (T-0279)` with the edit below.

There is no `test(T-0279)` commit: the change is comments only, and the Acceptance Cases are checked by command. No double quotes in the message: the allow-list refuses an escaped one.

## Edit (lead decision; copy exactly, `//` comment lines as in the surrounding code)
### `pkg/engine/doc.go`
```diff
@@ -40,7 +40,7 @@
 //     FR-ZON-005), from the scope askFor opens, and its match record carries
 //     that card's name (DecisionRecord.Source).
 //   - Steps OpScratch, OpDraw, OpDiscard, OpReturnToHand, OpPutOnDeck,
-//     OpDealDamage, OpHeal, OpAddArmor, OpModifyStat, OpKill (inline destruction), OpMoveZone, OpSearchDeck, OpDeployFree (a free deploy from a discard pile, FR-TRN-003a), OpMorphScratch (a face-down transformation, FR-ENT-103), OpGrantKeyword (a keyword granted until end of turn, FR-LAY-001) and OpLookTop (a look at the top cards with free deploys and a bottom order, FR-MOD-001 §4), with "all" steps, source-power values, the
+//     OpDealDamage, OpHeal, OpAddArmor, OpModifyStat, OpKill (inline destruction), OpMoveZone, OpSearchDeck, OpDeployFree (a free deploy from a discard pile, FR-TRN-003a), OpMorphScratch (a face-down transformation, FR-ENT-103), OpGrantKeyword (the keywords of Step.Keywords granted until end of turn, FR-LAY-001) and OpLookTop (a look at the top cards with free deploys and a bottom order, FR-MOD-001 §4), with "all" steps, source-power values, the
 //     deploy-zone gate, the TurnYours gate, the Cornered gate, SubjectSelf (tracked in discard,
 //     FR-ZON-004) and SubjectOpponents. A later step on a pile slot acts on the asset an earlier
 //     OpDeployFree of the same resolution made of its card. SubjectPick is an untargeted pick made as the
@@ -117,20 +117,24 @@
 //   - trigger.go: the pending-trigger buffer and firing. placement.go: placing
 //     an effect, its modes and then its targets, the card it shows as its
 //     source (shownCard, SourceCard), and whether an optional trigger could do
-//     nothing (couldDoNothing). costs.go: a card's cost and
+//     nothing (couldDoNothing), and the one lock of an "in the same zone"
+//     effect on placement (sameZoneLock). costs.go: a card's cost and
 //     a placement's costs. resources.go: Requisition and Tactics, generating and
 //     converting them, and paying with them. targeting.go: target choice and
 //     untargeted picks. filter.go: checking assets and discard-pile cards against
-//     an ability's filter. effects.go: running a resolved ability's steps.
+//     an ability's filter, and a continuous ability's filter read from its
+//     source (matchesFromSource). effects.go: running a resolved ability's steps.
 //     freedeploy.go: free deploys and the look at the top cards.
 //   - state.go: core state types and lookups, and exhausting an asset. card.go: card types.
-//     cardload.go: the CSV loader. The keyword set, its names and line parser are in
-//     pkg/engine/ability (keywords.go).
+//     cardload.go: the CSV loader, and checkShapeReads, which rejects a definition
+//     that sets a field where nothing reads it. The keyword set, its names, line
+//     parser and fold are in pkg/engine/ability (keywords.go).
 //   - events.go: Event, the Observer, match records and emitting. choice.go: decisions,
 //     and the askFor scope that names their source card.
 //     snapshot.go: Snapshot and its deep copies. buildinfo.go: BuildInfo.
 //     damage.go: the damage pipeline. sweep.go: the death sweep, destruction
-//     and match loss. replace.go: destruction replacement. layers.go: derived power, health and keywords,
+//     and match loss. replace.go: destruction replacement and damage redirection
+//     (redirectDamage). layers.go: derived power, health and keywords,
 //     the Cornered check and EffectiveState, and an asset's Modifier list with
 //     each modifier's Duration.
 //     invariants.go: state checks.
```

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **`replace.go`'s line names redirection (#163 (f)):** `git -C W grep -n "replace.go: destruction replacement and damage redirection" -- pkg/engine/doc.go` prints exactly one line.
2. **The new names are in the Files list:** each of `git -C W grep -n "(redirectDamage)" -- pkg/engine/doc.go`, `git -C W grep -n "(sameZoneLock)" -- pkg/engine/doc.go`, `git -C W grep -n "(matchesFromSource)" -- pkg/engine/doc.go` and `git -C W grep -n "checkShapeReads, which rejects" -- pkg/engine/doc.go` prints exactly one line.
3. **The Steps line names `Step.Keywords`:** `git -C W grep -n "OpGrantKeyword (the keywords of Step.Keywords" -- pkg/engine/doc.go` prints exactly one line, and `git -C W grep -n "OpGrantKeyword (a keyword granted" -- pkg/engine/doc.go` prints nothing.
4. **Exactly the edit:** `git -C W diff autoloop/m2-wave-h.1 --numstat -- pkg docs` prints exactly one line, `10	6	pkg/engine/doc.go`, and `git -C W rev-parse HEAD:pkg/engine/doc.go` prints `8de0e5daf06981a571e71a468ebab8f60ae6c4a6`.
5. **Nothing else changed:** `git -C W diff autoloop/m2-wave-h.1 --stat` lists only `pkg/engine/doc.go` and this task file.
6. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-h.1` prints `[PASS]`.
7. `gofmt -l W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/doc.go` (the edit above)

## Tests
- None: comments only. No test file changes.

## Done When
- [ ] Cases 1-7 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-19, lead-architect (wave H.1 run, session 8): written in full against b5779c6 for #163 (f). Session 8's first attempt (scratch `20261118-215328-8`) committed #163 (a)-(e) and (g) in the design note (b5779c6) and dry-ran this edit, then ended before writing a task or a run-file row; its two worktrees were clean apart from that dry run and were removed. `main` (7a8e4f5) has nothing the run branch lacks, so no sync. Not `WAVE-COMPLETE` before the task: #163 and #164 are `open: before the next Deliverables task`. This is a harness task, so it skips `tests` and goes straight to `ready`. Tier is `flash`, because the edit is exact and was dry-run. Every command here is on the headless allow-list: one per line, no pipes, no `$`, and patterns with parentheses are quoted. Waiting directives checked: #6, #7 (mode `serial`), #15 (2) (no draw or end-of-turn hook: `git grep` for one finds nothing), #71, #91 (2), #106, #120, #142 (wave H.1 has no cards): none has fired. Next: `kernel-engineer` on Flash.
