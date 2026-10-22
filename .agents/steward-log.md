# Workflow Steward Log

Kept by `workflow-steward` (`.agents/skills/workflow-steward/SKILL.md`). The steward reads this at the start of every session. Keep entries short. Remove an open thread when it is resolved or declined.

## Open Threads

| Opened | Thread | Waiting on |
| :--- | :--- | :--- |
| 2026-10-21 | **Explainer persona (deferred).** Structural judgment went to `lead-architect` (Runbook E, placement checks in D2/D5). What remains is an optional read-only persona that explains how parts of the code work. | User: only if wanted later |
| 2026-10-22 | **pm-audit-reviewer retired, loose ends.** The persona was removed at the user's request, which settles the #32 gate question (it was the last path that edited `pkg/engine` outside the build loop). DA audits now go to `game-designer` §6, and `pm-frd-drafter` §2.1 step 7 closes them. Still naming the old persona: `docs/audit/README.md` (workflow diagram, Clearing Loop section) and the `audit_registry.md` header (outside fences, asked separately); the default `-source "Agentic (pm-audit-reviewer)"` in `cmd/cardtown/tools.go` and `pkg/toolchain/audit/audit.go` (code, build loop if wanted); the #32 status cell on the M2 board (lead's). | User: docs/audit sign-off; lead for the board cell |

## Boundary Exceptions

Changes the steward made outside its own files (§3 of its skill), each with a human sign-off.

| Date | File(s) | Change | Signed off by | Commit |
| :--- | :--- | :--- | :--- | :--- |
| 2026-10-20 | `docs/frd/01_Game_Rules_FRD.md`, `docs/design/m2_starter_set_ability_survey.md`, `data/cards/maneuvers-*.csv`, `docs/architecture/m2-ability-model.md` | Before this persona existed, the general session wrote the user's targeting rulings directly. Targets locked on placement (FRD 01 v3.7, `SVY-R3` revised), with card text for Animoceps Spores, Steel Rain and Mutation Swell. Recorded here for the trail. | User, in conversation | `09b9bdb`, `0eb09d9` |
| 2026-10-21 | `docs/PRODUCT_SUMMARY.md` | Before this persona existed: copied pm-lead's summary onto the integration branch and corrected stale lines (play tasks written, M2 title, survey labels). | User, in conversation | `1b26419`, `f2a5bcc` |
