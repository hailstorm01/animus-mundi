# Wave Run: m2 wave C.1

- **Run branch:** `autoloop/m2-wave-c.1`, from `main` at `e64ca5f`
- **Started:** 2026-11-03
- **Cards:** none
- **Deferred:** none
- **Kept:** none

Wave C.1, the cleanup from structural review 7 (design note Deliverable 6f). It has no survey row and no cards. Its tasks are board directives #92 to #102, in board order, serial: #92 (one declare sequence), #93 (one read of a card's cost), #94 (`costs.go`), #95 (one `deployCard`), #96 (one `bufferTrigger`), #97 (one read of "another", and `joinModes`), #98 (`Resource.String()` and `IsScratchCompany`), #99 (`sim.PolicyByName`), #100 (`pkg/toolchain/specdoc`), #101 (the zone helpers move to `zones.go`) and #102 (the docs say what wave C built). Every code task is a pure move, and #102 is comments and docs only. The wave is complete when #92 to #102 are `applied`, every task's `go tool task sim-compare` passed (for #100, unchanged lint output), and `go tool task check` and `go tool task sim` pass.

| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
