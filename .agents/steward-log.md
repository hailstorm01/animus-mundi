# Workflow Steward Log

Kept by `workflow-steward` (`.agents/skills/workflow-steward/SKILL.md`). The steward reads this at the start of every session. Keep entries short. Remove an open thread when it is resolved or declined.

## Open Threads

| Opened | Thread | Waiting on |
| :--- | :--- | :--- |
| 2026-10-21 | **Architecture persona.** A separate persona for "how does this part work" and code-structure questions (for example "should sim code be separated from engine code?"), split out of the steward's role. | User: on hold until the steward is settled |
| 2026-10-21 | **`SVY-R9`–`SVY-R11` are not in FRD 01.** Ruled by the user on 2026-10-21 (optional triggers asked on placement, last-known effective power for "damage equal to its power", "up to N" targets distinct) and implemented, but FRD 01 (v3.8) does not mention them. | User to run `pm-frd-drafter` |

## Boundary Exceptions

Changes the steward made outside its own files (§3 of its skill), each with a human sign-off.

| Date | File(s) | Change | Signed off by | Commit |
| :--- | :--- | :--- | :--- | :--- |
| 2026-10-20 | `docs/frd/01_Game_Rules_FRD.md`, `docs/design/m2_starter_set_ability_survey.md`, `data/cards/maneuvers-*.csv`, `docs/architecture/m2-ability-model.md` | Before this persona existed, the general session wrote the user's targeting rulings directly. Targets locked on placement (FRD 01 v3.7, `SVY-R3` revised), with card text for Animoceps Spores, Steel Rain and Mutation Swell. Recorded here for the trail. | User, in conversation | `09b9bdb`, `0eb09d9` |
| 2026-10-21 | `docs/PRODUCT_SUMMARY.md` | Before this persona existed: copied pm-lead's summary onto the integration branch and corrected stale lines (play tasks written, M2 title, survey labels). | User, in conversation | `1b26419`, `f2a5bcc` |
