# Wave Run: m3 wave A.1

- **Run branch:** `autoloop/m3-wave-a.1`, from `main` at `2ea08269`
- **Started:** 2026-11-26
- **Cards:** none
- **Deferred:** none
- **Kept:** none

Wave A.1, design note Deliverable 5a (`docs/architecture/m3-full-catalog.md`): cleanup from structural review 16 of wave A, after wave A and before wave B. It has no survey row and no cards. Its work is board directives #190 to #193, one task each, in board order: #190 one `wantOptions` check, #191 `deploysOrAttacks`, #192 `register` refusing a repeated card, #193 shared test setup (`cardWithTrigger`, `deployPaid`, `activateChoosing`). Each is a move with no behavior change (#192 adds one test), so every task passes `go tool task sim-compare`. The wave is complete when the four are `applied` and `go tool task check`, `go tool task sim` and `go tool task sim-matrix` pass on the run branch.

| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
