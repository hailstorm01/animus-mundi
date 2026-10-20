# Workspace Rules: Animus Mundi

## Progressive Skill Routers
Read a skill only when its trigger fires. Do not load persona skills for unrelated work.
- **Docs Front Door (`docs-concierge`):** On `@docs-concierge`, or when the user asks what tools/personas exist, what to do next, or how to launch something in the docs workflow. Read `.agents/skills/docs-concierge`.
- **Product Vision & Portfolio Lead (`pm-lead`):** On `@pm-lead`, or when asked to review or update the product vision, milestones, or portfolio view in `docs/PRODUCT_SUMMARY.md`, or when the stakeholder wants to adapt the product vision (expand, contract, pivot). Dispatches analysis requests for active FRDs to `pm-frd-drafter`. Read `.agents/skills/pm-lead`.
- **FRD Authoring & Spec Analysis (`pm-frd-drafter`):** On `@pm-frd-drafter`, when asked to edit FRD 01 rules, to draft an FRD for a milestone that has started, or to analyze and report the state of an FRD (e.g., for `pm-lead`). Read `.agents/skills/pm-frd-drafter`.
- **Rules Red-Teaming (`pm-devils-advocate`):** After an agent makes major structural revisions to `docs/frd/01_Game_Rules_FRD.md` (edits made by `pm-audit-reviewer` do NOT count), or on `@pm-devils-advocate`. Frozen FRDs are never audited. Read `.agents/skills/pm-devils-advocate`.
- **Audit Triage & Resolution (`pm-audit-reviewer`):** Only on `@pm-audit-reviewer` or an explicit user request to review, triage, or resolve audits. Files sitting in `docs/audit/pending/` are NOT a trigger. Read `.agents/skills/pm-audit-reviewer`.
- **Game Design Escalation (`game-designer`):** On reviewer option `[E]`, `@game-designer`, or when the user wants to rule on an open keyword-matrix item. Read `.agents/skills/game-designer`.
- **Technical Architecture (`lead-architect`):** On `@lead-architect`, when a milestone starts and needs a design note, when a technical decision needs a decision record, or when the user explicitly asks for a full TDD or for slicing a TDD into GitHub issues (`cardtown tdd slice`). Read `.agents/skills/lead-architect`.
- **Build Loop Lead (`lead-architect` Runbook D):** When asked to plan, start, continue or review build tasks in `.agents/tasks/`. Always Opus, in the main Claude Code session. Read `.agents/skills/lead-architect`.
- **Rules Test Author (`rules-test-author`):** On `@rules-test-author T-NNNN`, or when a task file is at `status: tests`. Sonnet. Read `.agents/skills/rules-test-author`.
- **Kernel Engineer (`kernel-engineer`):** On `@kernel-engineer T-NNNN`, or when a task file is at `status: ready` or `changes`. Gemini Flash for `tier: flash`, Sonnet for `tier: sonnet`. Read `.agents/skills/kernel-engineer`.
- **Card Town Verification (`cardtown-linter`):** When auditing specifications, engine scenario tests, TDDs, or CSV balance sheets. Read `.agents/skills/cardtown-linter`.

## Source of Truth
- **Rules:** `docs/frd/01_Game_Rules_FRD.md`, made executable by `pkg/engine`. A rule change lands as FRD 01 text plus a `TestFR_*` scenario test, in the same change. Precedence and workflow: `docs/frd/README.md`.
- **Frozen FRDs (02–05):** design intent only. Do not edit, audit, or design from them; they are rewritten when their milestone starts.

## Guard Rails
- **No unattended chains:** The reviewer -> DA -> reviewer clearing loop (`--clear-backlog`) starts only on an explicit human command. DA never invokes the reviewer. Reviewer edits never trigger the DA outside a human-started loop. Portfolio reconciliation via `pm-lead` starts only on an explicit human prompt; FRD edits never auto-trigger synchronization. Design notes and TDDs (`lead-architect` Runbooks A and B) never auto-trigger GitHub issue publication (Runbook C); always run `cardtown tdd slice -dry-run` first and require explicit human confirmation before `-publish`. The build loop (`lead-architect` Runbook D) runs **one task per human prompt**: once that task is `done` or `blocked`, or needs the user (a Flash dispatch or a ruling), the lead stops and reports. It never dispatches another task, or starts the next deliverable, without a new human prompt. It may write the current deliverable's other tasks, but leaves them `queued`. A merge never triggers the next task.
- **No agent `git push`:** Agents commit locally. Push only when the user asks.
- **Pre-commit hook:** It is not versioned by git. After a fresh clone run `cardtown hook install`.
