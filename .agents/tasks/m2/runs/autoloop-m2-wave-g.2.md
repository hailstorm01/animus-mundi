# Wave Run: m2 wave G.2

- **Run branch:** `autoloop/m2-wave-g.2`, from `main` at `281e7d8`
- **Started:** 2026-10-30
- **Cards:** none
- **Deferred:** none
- **Kept:** none

Wave G.2, the cleanup from structural review 6 (design note Deliverable 6e). It has no survey row and no cards. Its tasks are board directives #82 to #87 (c), in board order, serial: #82 (a step acts only on chosen assets still on the battlefield, a `shape` task), #83 (one `kill`, pure move), #84 (one `moveOnBattlefield`, pure move; #83 and #84 may share a task), #85 (one keyword-name table, pure move), #86 (event kinds as named constants, pure move, split under the size rule) and #87 (c) (file comments on four engine files). The wave is complete when #82 to #87 are `applied`, every task's `go tool task sim-compare` passed, and `go tool task check` and `go tool task sim` pass.

| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
