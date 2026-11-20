# Wave Run: m2 wave H.3

- **Run branch:** `autoloop/m2-wave-h.3`, from `main` at `54c8db4`
- **Started:** 2026-11-20
- **Cards:** none
- **Deferred:** none
- **Kept:** none

Wave H.3, M2's done-when and cleanup (design note Deliverable 6p), after wave H.2 and before the write-up (Deliverable 7). It has no survey row and no cards. Its tasks are board directives #177 to #179, in board order, serial: #177 (`FR-DEK-002` and `FR-REP-002` get `TestFR_*` scenario tests, so `cardtown lint rules` reports 58 of 62), #178 (every stack item names its card through `SourceCard` and `sourceName`, and `cardName` goes) and #179 (the view's `StackView.Chosen` holds only the chosen assets still on the battlefield). The wave is complete when #177 to #179 are `applied`, no `open: before` directive is left, every task's `go tool task sim-compare` passed, and `go tool task check` and `go tool task sim` pass.

| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
