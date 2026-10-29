# Wave Run: m2 wave G

- **Run branch:** `autoloop/m2-wave-g`, from `main` at `294b6a2`
- **Started:** 2026-10-29
- **Cards:** Shifting Impossibility
- **Deferred:** Animoceps Spore Cloud (#14 (b): damage redirection is needed only by a non-deck card)
- **Kept:** none

Wave G's destruction replacement, pulled ahead of wave C by the user for board directive #73 (design note Deliverable 6c). T-0096 (replacement at the death sweep) merged into `main` before the run began. This run builds the rest, serial: T-0097 (kill steps go through replacement), T-0100 (the choice between several replacements, split from T-0097 under the size rule), T-0098 (a Scratch Company filter and *Shifting Impossibility*), then T-0099 (#73 itself). The wave is complete when #73 is `applied`.

| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | 2026-10-29 | T-0097 (directive #73 part 2: kill steps go through destruction replacement through one shared loop, `replaceDeaths`, moved out of `sweepDeaths`; no sweep between a kill step's replaced kills, as the user ruled 2026-10-29; a replaced kill still counts as done, `SVY-R15`; wave G shape, sonnet); run file created, `main` already synced; T-0096's plan split under the size rule, so the choice between replacements is T-0100, before T-0098 (design note Deliverable 6c updated); #73 stays `open: before the first wave C task`, #61 and #71 stay `waiting` | done | `c8e391f` | tests 1, code 1 (no changes) | none | clean, 200 games, P0 63.5% / P1 34.5% / draws 2.0% (unchanged); `sim-compare` byte-identical to the run branch (starter and stress); deck cards still on `NotYetExecutable`: 23 (10 Animus, 13 Path) | CONTINUE |
