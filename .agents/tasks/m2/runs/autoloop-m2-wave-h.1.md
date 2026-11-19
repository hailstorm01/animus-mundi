# Wave Run: m2 wave H.1

- **Run branch:** `autoloop/m2-wave-h.1`, from `main` at `7a8e4f5`
- **Started:** 2026-11-18
- **Cards:** none
- **Deferred:** none
- **Kept:** none

Wave H.1, the cleanup from structural review 13 (design note Deliverable 6n), after wave H and before the write-up (Deliverable 7). It has no survey row and no cards. Its tasks are board directives #159 to #168, in board order, serial: #159 (one keyword set, `ability.Keywords`, ADR 0008; two tasks, T-0272 and T-0273), #165 (the fence linter rejects type aliases), #160 (one same-zone lock on placement), #161 (one read of a continuous ability's filter), #166, #167, #168 (the forwarding wrappers that predate ADR 0007 go), #162 (`LoadCards` checks where each shape is read), #163 (the docs say what wave H built) and #164 (`view.StackView.Keyword` goes). The wave is complete when #159 to #168 are `applied`, no `open: before` directive is left, every task's `go tool task sim-compare` passed, and `go tool task check` and `go tool task sim` pass.

| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
