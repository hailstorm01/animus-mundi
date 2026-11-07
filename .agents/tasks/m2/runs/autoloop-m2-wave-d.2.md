# Wave Run: m2 wave D.2

- **Run branch:** `autoloop/m2-wave-d.2`, from `main` at `5e9dc90`
- **Started:** 2026-11-06
- **Cards:** none
- **Deferred:** none
- **Kept:** none

Wave D.2, the cleanup from structural review 9 (design note Deliverable 6h). It has no survey row and no cards. Its tasks are board directives, in board order, serial: #124 (decklist errors name the file, with #8), then #125 (a value doesn't need a stack item), #126 (effective reads live in `layers.go`), #127 (the policy asks the engine what an activated ability costs), #129 (effects log what they do), #130 (one read of an action event's kind), #131 (one read of "the source, or its last-known record"), #132 (small repeats, each spelled once), #133 (one reader of the keyword matrix), #134 (one classifier of changed files) and #135 (the text says what the code does). #128 has no task here: it is a `waiting:` row for *Veil Render*. The wave is complete when no `open: before` directive is left, every other task's `go tool task sim-compare` passed (for #133 and #134, unchanged lint output; #135 changes one sim header line, which its `sim-compare` shows), and `go tool task check` and `go tool task sim` pass.

| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
