# Wave Run: m2 wave E

- **Run branch:** `autoloop/m2-wave-e`, from `main` at `54bcada`
- **Started:** 2026-11-09
- **Cards:** Macabre Dancers; Boneyard Amalgam
- **Deferred:** Animoceps Stalk (#14 (b)); Animoceps Node (#14 (b)); Melodic Pummeler (#14 (b)); Hulking Incongruity (#14 (b)); Brigand Highlander (#14 (b)); Ancestral Monument (#14 (b))
- **Kept:** none

Wave E, the survey's row E (`docs/design/m2_starter_set_ability_survey.md` line 33), after waves D.1 and D.2. Its two deck cards (`data/decks/path-starter.txt`) go first under #14 (a): *Macabre Dancers* only combines shapes that are `done` (Unbowed on `HookDamaged`, `OpHeal` on `SubjectSelf`, and T-0180's `CorneredKeywords.Resistant`), so it is a card task. *Boneyard Amalgam* needs hand-active cost reduction (`FR-CST-001`), a shape task, and it fires #118 (b) (the view shows the lower cost). The six non-deck cards wait for the playtest checkpoint, #14 (b): five are same-zone auras for other assets, and *Brigand Highlander*'s "Cornered: This unit has +2 power." is a power modifier behind Cornered, a shape no deck card needs (`CorneredKeywords` holds only keywords), so #14 (c) doesn't let it join.

| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | 2026-11-09 | T-0199 Macabre Dancers (card, flash) | done | eeb4a55 | tests 1, code 1 | none | clean; Path 36/50 main cards executable; P0 79.5% / P1 18.5% / draws 2.0%; 9 deck cards left on `NotYetExecutable` | CONTINUE |
