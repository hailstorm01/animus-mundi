# Workflow Steward Log

Kept by `workflow-steward` (`.agents/skills/workflow-steward/SKILL.md`). The steward reads this at the start of every session. Keep entries short. Remove an open thread when it is resolved or declined.

## Open Threads

| Opened | Thread | Waiting on |
| :--- | :--- | :--- |
| 2026-10-21 | **Explainer persona (deferred).** Structural judgment went to `lead-architect` (Runbook E, placement checks in D2/D5). What remains is an optional read-only persona that explains how parts of the code work. | User: only if wanted later |
| 2026-10-22 | **Directive #32, skills half done.** Dropped the false "`go test ./pkg/engine` runs the playouts" claim from `cardtown-linter` and `docs/AGENTS.md`, and pointed docs-concierge's "check your edits" suggestion at `go tool task test`. Proposed, not made (guard rail): pm-audit-reviewer's commit gate, which may edit `pkg/engine/*.go` at `OPEN(KIM-...)` markers, could run `go tool task test` instead of `go test ./pkg/engine` so a kernel edit cannot break the sim playouts unseen. | User: yes/no on the pm-audit-reviewer gate; pm-lead for the PRODUCT_SUMMARY half |

## Boundary Exceptions

Changes the steward made outside its own files (§3 of its skill), each with a human sign-off.

| Date | File(s) | Change | Signed off by | Commit |
| :--- | :--- | :--- | :--- | :--- |
| 2026-10-20 | `docs/frd/01_Game_Rules_FRD.md`, `docs/design/m2_starter_set_ability_survey.md`, `data/cards/maneuvers-*.csv`, `docs/architecture/m2-ability-model.md` | Before this persona existed, the general session wrote the user's targeting rulings directly. Targets locked on placement (FRD 01 v3.7, `SVY-R3` revised), with card text for Animoceps Spores, Steel Rain and Mutation Swell. Recorded here for the trail. | User, in conversation | `09b9bdb`, `0eb09d9` |
| 2026-10-21 | `docs/PRODUCT_SUMMARY.md` | Before this persona existed: copied pm-lead's summary onto the integration branch and corrected stale lines (play tasks written, M2 title, survey labels). | User, in conversation | `1b26419`, `f2a5bcc` |
