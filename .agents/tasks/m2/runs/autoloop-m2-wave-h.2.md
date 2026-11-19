# Wave Run: m2 wave H.2

- **Run branch:** `autoloop/m2-wave-h.2`, from `main` at `1859fe4`
- **Started:** 2026-11-19
- **Cards:** none
- **Deferred:** none
- **Kept:** none

Wave H.2, the cleanup from structural review 14 (design note Deliverable 6o), after wave H.1 and before the write-up (Deliverable 7). It has no survey row and no cards. Its tasks are board directives #169 to #176, in board order, serial: #169 (one `sourceAsset` lookup), #170 (pick prompts name the card through `SourceCard`), #171 (`ability.Keyword` becomes `KeywordTrigger`, the fields `FromKeyword`), #172 (each linter input path written once; (a) and (b) may be two tasks), #173 (small repeats written once), #174 (the play screen's stack listing prints `StackLabel`), #175 (the stack listing shows chosen targets) and #176 (the docs say what wave H.2 built). The wave is complete when #169 to #176 are `applied`, no `open: before` directive is left, every task's `go tool task sim-compare` passed, and `go tool task check` and `go tool task sim` pass.

| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
