# Wave Run: m2 wave G.1

- **Run branch:** `autoloop/m2-wave-g.1`, from `main` at `556811f`
- **Started:** 2026-10-29
- **Cards:** none
- **Deferred:** none
- **Kept:** none

Wave G.1, the cleanup from structural review 5 (design note Deliverable 6d). It has no survey row and no cards. Its tasks are board directives #76 to #79, in board order, serial: #76 (a step returns or puts any asset it acts on, a `shape` task), #77 (one APNAP order and one "oldest first", pure move), #78 (the Scratch Company filter becomes one enum, no behavior change) and #79 (the death sweep moves to `sweep.go`, pure move). Directive #81 (`SVY-R24`, a replacement saves an asset at most once per sweep) goes after #79 by its own Order line, and as an `open` directive it is applied before any later task (D2.0). The wave is complete when #76 to #79 are `applied`, every task's `go tool task sim-compare` passed, and `go tool task check` and `go tool task sim` pass.

| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
