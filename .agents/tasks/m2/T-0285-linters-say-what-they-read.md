---
id: T-0285
title: "`cardtown lint` asks the linters what they read: `fences.Reads`, `rules.Reads` (directive #172 (b); wave H.2 harness)"
kind: harness
tier: flash
wave: H.2 (board directive #172 (b))
status: queued
branch: m2/autoloop-wave-h.2/T-0285-linters-say-what-they-read
depends_on: [T-0284]
claims: [func:fences.Reads, func:rules.Reads]
cards: []
requirements: []
---

## Goal
`cmd/cardtown`'s `isFencesPath` and `isRulesPath` restate which files the fences and rules linters read ([lint.go:217-218](../../../cmd/cardtown/lint.go), [:284-288](../../../cmd/cardtown/lint.go)), so T-0274 had to widen both the fences linter and `isFencesPath`. After this task, `fences.Reads(path)` and `rules.Reads(path)` report whether a change to a slash-separated, lower-cased path can affect that linter, built from the constants T-0284 adds (`specdoc.FRDDir`, `specdoc.RulesFRD`, `specdoc.MatrixFile`, `cardsheet.Dir`, `fences.SurveyFile`), `aliasRoots` and `rules.EngineDir`. `selectChanged` calls them, and `isFencesPath` and `isRulesPath` go. `scannedFiles` lists `goFilesUnder(repoRoot, "pkg")` minus its two skipped trees (`pkg/engine/cards`, `pkg/toolchain`), so the walk for non-test Go files is written once. When this merges after T-0284, #172 is `applied`.

## Covers (from directive #172 (b) and its acceptance)
- `fences.Reads` and `rules.Reads`, as above; `selectChanged` calls them; `isFencesPath` and `isRulesPath` go (ADR 0007).
- `scannedFiles` reads `goFilesUnder`; the same files, in the same order, so the fences linter's output is unchanged.
- **Authorized test edits:** `TestIsFencesPath` ([lint_test.go:14](../../../cmd/cardtown/lint_test.go)) moves to `pkg/toolchain/linter/fences` as `TestReads`, the same cases, calling `fences.Reads`; a new `TestReads` in `pkg/toolchain/linter/rules` with FRD 01, an engine Go file, an engine test file and a `pkg/sim` file. `TestSelectChanged_*` pass unedited.
- Done when: each linter's output on the repo is unchanged; `go tool task check`; `go tool task sim-compare` (byte for byte).

Inputs, Acceptance Cases, Allowed Files and Edits are written against the code of the day, after T-0284 merges and before this task leaves `queued` (D2.4).

## Log
- 2026-11-19, lead-architect (wave H.2 run, session 4): split off from #172 as the directive allows ((a) first, in T-0284). Queued behind T-0284; #173 waits for this task, because #172 blocks it ("Wave H.2, after #172").
