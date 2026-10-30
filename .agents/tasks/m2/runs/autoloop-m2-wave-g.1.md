# Wave Run: m2 wave G.1

- **Run branch:** `autoloop/m2-wave-g.1`, from `main` at `556811f`
- **Started:** 2026-10-29
- **Cards:** none
- **Deferred:** none
- **Kept:** none

Wave G.1, the cleanup from structural review 5 (design note Deliverable 6d). It has no survey row and no cards. Its tasks are board directives #76 to #79, in board order, serial: #76 (a step returns or puts any asset it acts on, a `shape` task), #77 (one APNAP order and one "oldest first", pure move), #78 (the Scratch Company filter becomes one enum, no behavior change) and #79 (the death sweep moves to `sweep.go`, pure move). Directive #81 (`SVY-R24`, a replacement saves an asset at most once per sweep) goes after #79 by its own Order line, and as an `open` directive it is applied before any later task (D2.0). The wave is complete when #76 to #79 are `applied`, every task's `go tool task sim-compare` passed, and `go tool task check` and `go tool task sim` pass.

| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | 2026-10-29 | T-0102 (directive #76: a step returns or puts any asset it acts on, not only a replaced one; `relocateReplaced` moves out of `replace.go` into `stepRelocate`'s asset half, and one `relocateTo` maps both ops to a zone; an asset an earlier step destroyed is skipped; `FR-ZON-004`, `FR-TGT-003`; wave G.1 shape, sonnet); run file created, `main` (556811f) has nothing the run branch lacks, so no sync; #81's trigger fired (outcome `applied: FRD 01 v3.18`), now `open: after #79`; #77, #78, #79 stay `open: before the next Deliverables task`; waiting triggers re-checked (#6, #7, #8, #15, #20 (b), #58, #59 (2), #61, #71), none has fired | done | `c84890e` | tests 3 (round 1: a field #78 removes, and a test name; round 3 after the engineer blocked on lowercase zone strings that `Zone.String()` and T-0096's test contradict), code 1 (no changes) | none | clean, 200 games, P0 67.0% / P1 32.5% / draws 0.5% (unchanged); `sim-compare` byte-identical to the run branch (starter and stress); deck cards still on `NotYetExecutable`: 22 (9 Animus, 13 Path) | CONTINUE |
