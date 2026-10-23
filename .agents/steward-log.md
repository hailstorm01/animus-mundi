# Workflow Steward Log

Kept by `workflow-steward` (`.agents/skills/workflow-steward/SKILL.md`). The steward reads this at the start of every session. Keep entries short. Remove an open thread when it is resolved or declined.

## Open Threads

| Opened | Thread | Waiting on |
| :--- | :--- | :--- |
| 2026-10-21 | **Explainer persona (deferred).** Structural judgment went to `lead-architect` (Runbook E, placement checks in D2/D5). What remains is an optional read-only persona that explains how parts of the code work. | User: only if wanted later |
| 2026-10-22 | **pm-audit-reviewer retired, loose ends.** The persona was removed at the user's request, which settles the #32 gate question (it was the last path that edited `pkg/engine` outside the build loop). DA audits now go to `game-designer` §6, and `pm-frd-drafter` §2.1 step 7 closes them. Still naming the old persona: the default `-source "Agentic (pm-audit-reviewer)"` in `cmd/cardtown/tools.go` and `pkg/toolchain/audit/audit.go` (code, build loop if wanted); the #32 status cell on the M2 board (lead's). | Lead for the board cell; a build task for the label, if wanted |
| 2026-10-22 | **Automatic flash dispatch.** Runbook D3 makes the user paste `@kernel-engineer T-NNNN` into Antigravity. Options offered: (A) the lead calls Antigravity's `agentapi new-conversation --model=flash` (probed 2026-10-22: reaches Antigravity IDE's language server with `ANTIGRAVITY_LS_ADDRESS`, `ANTIGRAVITY_CSRF_TOKEN` and `ANTIGRAVITY_PROJECT_ID`, but both IDE servers refuse to start a conversation, "projectsStore is nil"; the project store probably belongs to the separate standalone Antigravity app, which was not running); (B) flash tier becomes the `kernel-engineer` subagent on Haiku. Either changes the lead's dispatch rights (D3, AGENTS.md "No unattended chains"), so it needs the user's yes. | User: pick A, B or neither; A needs a trial run first |

## Boundary Exceptions

Changes the steward made outside its own files (§3 of its skill), each with a human sign-off.

| Date | File(s) | Change | Signed off by | Commit |
| :--- | :--- | :--- | :--- | :--- |
| 2026-10-20 | `docs/frd/01_Game_Rules_FRD.md`, `docs/design/m2_starter_set_ability_survey.md`, `data/cards/maneuvers-*.csv`, `docs/architecture/m2-ability-model.md` | Before this persona existed, the general session wrote the user's targeting rulings directly. Targets locked on placement (FRD 01 v3.7, `SVY-R3` revised), with card text for Animoceps Spores, Steel Rain and Mutation Swell. Recorded here for the trail. | User, in conversation | `09b9bdb`, `0eb09d9` |
| 2026-10-21 | `docs/PRODUCT_SUMMARY.md` | Before this persona existed: copied pm-lead's summary onto the integration branch and corrected stale lines (play tasks written, M2 title, survey labels). | User, in conversation | `1b26419`, `f2a5bcc` |
| 2026-10-22 | `docs/audit/README.md`, `docs/audit/audit_registry.md` | After retiring pm-audit-reviewer: README workflow now runs DA → game-designer → pm-frd-drafter → `cardtown audit resolve`; Clearing Loop section replaced by "Working Through an Audit"; severity tags no longer say human/loop. Registry header and source-label example name the new flow; past rows unchanged. | User, in conversation | the commit adding this row |
