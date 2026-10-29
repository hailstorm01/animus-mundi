# Wave Run: m2 wave B.2

- **Run branch:** `autoloop/m2-wave-b.2`, from `main` at `50c607c`
- **Started:** 2026-10-28
- **Cards:** none
- **Deferred:** none
- **Kept:** none

Wave B.2 is the cleanup between wave B.1 and wave C: board directives #68–#70 from the user's code review, with #72's changes to #68, every task a harness task, in board order (design note Deliverable 6b). #68 is four tasks under the size rule, parts (a)–(d). #71 and #61's fifth question are recorded with the wave but wait for their own triggers.

| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | 2026-10-28 | T-0089 (directive #68 part (a), with #72: Berserk's rules leave `actions.go` and `combat.go`, so an Attack has no secondary target and nothing cleaves; the Berserk tests deleted, the others kept with plain attackers; harness, sonnet); run file created, `main` already synced; #68 (b)–(d), #69 and #70 stay open for the next sessions, #61 and #71 stay `waiting` | done | `dbf6f0a` | code 1 (no changes; an unrequested ruling-heading removal in `rules_test.go` is carried to part (b)'s test commit on #68's row) | none | clean, 200 games, P0 63.5% / P1 34.5% / draws 2.0% (unchanged); `sim-compare` byte-identical to the run branch (starter and stress); starter events differ only in the one defender-fizzle wording line (plus the build-revision stamp); deck executability unchanged (30/50 Animus) | CONTINUE |
