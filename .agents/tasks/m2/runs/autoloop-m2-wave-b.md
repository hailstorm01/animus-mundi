# Wave Run: m2 wave B

- **Run branch:** `autoloop/m2-wave-b`, from `main` at `4d0abf0`
- **Started:** 2026-10-26
- **Cards:** Hammerhand Brigands; Devilhorn Captain; Chainbreaker Irregulars; Allshan, Chainbreaker Revenant; Animoceps Rager; Animoceps Shambler; Eternal Watchman; Infested Apostle; Arkha, Sun's Fury; Rogue Fleshshaper
- **Deferred:** Hammerhand Bivouac (#14 (b): turn-start hook, non-deck); Blind Defile (#14 (b): opposing-unit-attacks hook, non-deck); Drifting Concept (#14 (b): `RevealHand`, non-deck)

| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | 2026-10-26 | T-0067 (discard from your own hand, shape, sonnet); T-0068 written, `queued` | done | `124cf3d` | tests 1, code 1 (no changes) | none | clean, 200 games, P0 79.5% / P1 19.0% / draws 1.5%; byte-identical to the run branch (no card uses `OpDiscard` yet); 28 deck cards still not executable | CONTINUE |
| 2 | 2026-10-26 | T-0068 (Chainbreaker Irregulars, card, flash); synced `main` first (1a3ef25); header gained its `Deferred` line | done | `ab3545c` | tests 1, code 1 (no changes) | none | clean, 200 games, P0 63.0% / P1 33.0% / draws 4.0% (was 79.5% / 19.0% / 1.5%); P1 23/50 executable; 27 deck cards still not executable | CONTINUE |
