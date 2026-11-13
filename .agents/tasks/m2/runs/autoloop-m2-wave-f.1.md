# Wave Run: m2 wave F.1

- **Run branch:** `autoloop/m2-wave-f.1`, from `main` at `eed04ab`
- **Started:** 2026-11-13
- **Cards:** none
- **Deferred:** none
- **Kept:** none

Wave F.1, the cleanup from structural review 11 (design note Deliverable 6j), after wave F, which stopped at the #14 (d) playtest checkpoint, and before the work that follows the checkpoint (user, 2026-11-13). It has no survey row and no cards. Its tasks are board directives #144–#148, in board order: #144 (a card step's pile check made once, in `stepCards`), #145 (one pile lookup, `handCard` goes, and `Action.pile()`), #146 (`fireHook` returns what it buffered), #147 (`Game.PlayCost`, with acceptance cases) and #148 (free deploys and the look move to `freedeploy.go`; it may share #147's task). All but #147 are pure moves. The wave is complete when #144–#148 are `applied`, no `open: before` directive is left, every task's `go tool task sim-compare` passed, and `go tool task check` and `go tool task sim` pass.

| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
