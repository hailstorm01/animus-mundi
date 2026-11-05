# Wave Run: m2 wave D.1

- **Run branch:** `autoloop/m2-wave-d.1`, from `main` at `1062867`
- **Started:** 2026-11-05
- **Cards:** none
- **Deferred:** none
- **Kept:** none

Wave D.1, the cleanup from structural review 8 (design note Deliverable 6g). It has no survey row and no cards. Its tasks are board directives #107 to #121, in board order, serial: #107 (one read of "can pay", and one payment), #108 (one function plays a card from hand), #109 (the placement builds its stack item), #110 (one "first picks in this order"), #111 and #112 (one "this asset is yours" check, and "Aggressive" spelled once; they may share a task), #113 (`filter.go`), #114 (`resources.go`, and `exhaust` moves to `state.go`), #115 (the fences check the engine's imports), #116 (rule monitors read effective keywords, in two tasks under the size rule), #117 (a `shape` task, tier `sonnet`: "Cornered: keyword" is one continuous shape), #118 (a) (the hand view shows what a card costs now), #119 (a hidden decision stays hidden by its flag) and #121 (the docs say what wave D built). #120 has no task: it is a `waiting:` row for a cost no M2 card prints, and #118 (b) waits for *Boneyard Amalgam*. #107–#114 are pure moves, and #121 is comments and docs only. #115, #116, #117, #118 and #119 add tests and don't change how any executable card plays. The wave is complete when no `open: before` directive is left, every task's `go tool task sim-compare` passed, and `go tool task check` and `go tool task sim` pass.

| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
