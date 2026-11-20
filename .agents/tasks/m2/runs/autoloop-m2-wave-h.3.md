# Wave Run: m2 wave H.3

- **Run branch:** `autoloop/m2-wave-h.3`, from `main` at `54c8db4`
- **Started:** 2026-11-20
- **Cards:** none
- **Deferred:** none
- **Kept:** none

Wave H.3, M2's done-when and cleanup (design note Deliverable 6p), after wave H.2 and before the write-up (Deliverable 7). It has no survey row and no cards. Its tasks are board directives #177 to #179, in board order, serial: #177 (`FR-DEK-002` and `FR-REP-002` get `TestFR_*` scenario tests, so `cardtown lint rules` reports 58 of 62), #178 (every stack item names its card through `SourceCard` and `sourceName`, and `cardName` goes) and #179 (the view's `StackView.Chosen` holds only the chosen assets still on the battlefield). The wave is complete when #177 to #179 are `applied`, no `open: before` directive is left, every task's `go tool task sim-compare` passed, and `go tool task check` and `go tool task sim` pass.

| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | 2026-11-20 | T-0291 (directive #177, wave H.3's first: `FR-DEK-002` and `FR-REP-002` get `TestFR_*` scenario tests under their own IDs; tier sonnet, `kind: shape` only so the tests go through `rules-test-author` and the lead's tests review; six tests in `pkg/engine/dek_rep_inline_test.go` (the four #177 names, plus `TestFR_DEK_002_ShuffleTouchesOnlyTheSearchedDeck` and `TestFR_REP_002_ResistantUsesNoStackItem`), all passing on the `test(T-0291)` commit, so no `kernel-engineer` run and no engine change, as for #46/T-0052; no existing test edited). `cardtown lint rules` now reports 58 of 62, the M2 target: not yet `FR-ENT-102`, `FR-KWD-002n`, `FR-KWD-002p`, `FR-TRG-006`; trace check passes. Run file created. `main` (54c8db4) has nothing the run branch lacks, so no sync. #177 applied and archived. #178 stays `open: before the next Deliverables task` and #179 `open`; #178 is next in the wave order. Waiting triggers re-checked (#6, #7, #15 (2), #71, #91 (2), #106, #120, #142): none has fired. The unrelated worktree `wave-run/autoloop-m2-wave-h.2/20261119-163904-9/dry` (detached at b33c338) from an earlier session is still registered and was left alone | done | f8c7cd7 | tests 1 (no changes); no code round | none | P0 58.5%, P1 39.5%, draws 2.0% (unchanged); `sim-compare` byte-identical to the run branch (starter and stress) | CONTINUE |
