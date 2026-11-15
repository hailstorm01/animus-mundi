# Wave Run: M2 wave H

- **Run branch:** `autoloop/m2-wave-h`, from `main` at `47d5ced`
- **Started:** 2026-11-15
- **Cards:** Blind Defile; Drifting Concept; Hammerhand Bivouac; Blindside; Poisoned Rations; Rhapsody's Hollow; Animoceps Stalk; Animoceps Node; Melodic Pummeler; Hulking Incongruity; Brigand Highlander; Ancestral Monument; Mutation Swell; Solarheart Monk; Animoceps Spore Cloud
- **Deferred:** none
- **Kept:** all 15 (non-deck; #14, applied 2026-11-13: the remaining Animus and Path cards stay in M2, and the two starter factions are finished before Concordia and Soulless)

| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | 2026-11-15 | T-0241 (watching trigger and `SubjectWatched`, *Blind Defile*'s shape) | done | 4176846 | tests 2 (round 1: the gone-attacker test checked the stack before Step 2 triggers were flushed), code 1 | none | P0 58.5%, P1 39.5%, draws 2.0% (sim-compare byte-identical) | STOP: rules question for T-0242 (*Blind Defile*): do its two kills depend on each other? Run @game-designer (question queued in docs/design/outcomes/2026-11-15-wave-h-blind-defile-kill-both.md, on `main` at 07bd2da) |
| 2 | 2026-11-15 | T-0242 (*Blind Defile*) | done | 0adf9c2 | tests 1, code 1 | none | P0 58.5%, P1 39.5%, draws 2.0% (sim-compare byte-identical) | CONTINUE |
| 3 | 2026-11-15 | none (*Drifting Concept* is next) | not written | none | none | none | not run (no task) | STOP: rules question for *Drifting Concept*: which card a timed-out chooser takes from an opponent's revealed hand (`SVY-R16` covers only your own hand). Run @game-designer (question queued in docs/design/outcomes/2026-11-15-wave-h-drifting-concept-revealed-hand-timeout.md, on `main` at 6f2917d) |
