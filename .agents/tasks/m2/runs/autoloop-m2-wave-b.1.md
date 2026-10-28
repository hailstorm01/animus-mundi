# Wave Run: m2 wave B.1

- **Run branch:** `autoloop/m2-wave-b.1`, from `main` at `c9c27ee`
- **Started:** 2026-10-28
- **Cards:** none
- **Deferred:** none
- **Kept:** none

Wave B.1 is the cleanup between waves B and C: board directives #62–#67 from structural review 4, every task a harness pure move, in board order (design note Deliverable 6a).

| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | 2026-10-28 | T-0085 (directive #62: `placement`, `legalModes` and `chooseOnPlacement` in a new `placement.go`, so choosing modes and targets no longer takes a `pendingTrigger`; harness, sonnet, pure move); run file created, `main` already synced; #63–#67 stay open for the next sessions, #61 stays `waiting` | done | `647d84c` | code 1 (no changes; nit: rewrap `targetDecision`'s doc comment when next touched) | none | clean, 200 games, P0 63.5% / P1 34.5% / draws 2.0% (unchanged); `sim-compare` byte-identical to the run branch; 23 deck cards still not executable | CONTINUE |
| 2 | 2026-10-28 | T-0086 (directives #63 and #66: one untargeted pick body `pickFrom`, one pile predicate `ability.Filter.InPile` in place of six zone comparisons, picks moved to `targeting.go`, #66's doc lines; harness, flash, pure move); `main` already synced; #64 and #65 (with #67) stay open for the next sessions, #61 stays `waiting` | done | `b64ed0b` | code 1 (no changes) | none | clean, 200 games, P0 63.5% / P1 34.5% / draws 2.0% (unchanged); `sim-compare` byte-identical to the run branch; deck executability unchanged | CONTINUE |
| 3 | 2026-10-28 | T-0087 (directive #64: one function, `discardFromHand` in `zones.go` beside `draw` and `scratch`, holds the `DecideHand` decision, the move and the `discard` event; `stepDiscard` calls it per step player; harness, flash, pure move); `main` already synced; #65 (with #67) stays open for the next session, #61 stays `waiting` | done | `74e52fd` | code 1 (no changes) | none | clean, 200 games, P0 63.5% / P1 34.5% / draws 2.0% (unchanged); `sim-compare` byte-identical to the run branch; deck executability unchanged (30/50, 25/50) | CONTINUE |
