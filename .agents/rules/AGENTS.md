# Workspace Rules: Animus Mundi

## Progressive Skill Routers
Read a skill only when its trigger fires. Do not load persona skills for unrelated work.
- **Docs Front Door (`docs-concierge`):** On `@docs-concierge`, or when the user asks what tools/personas exist, what to do next, or how to launch something in the docs workflow. Read `.agents/skills/docs-concierge`.
- **Product Vision & Portfolio Lead (`pm-lead`):** On `@pm-lead`, or when asked to review or update the product vision, roadmap, or portfolio view in `docs/PRODUCT_SUMMARY.md`, or when the stakeholder wants to adapt the product vision (expand, contract, pivot). Dispatches FRD state analysis requests to `pm-frd-drafter`. Read `.agents/skills/pm-lead`.
- **FRD Authoring & Spec Analysis (`pm-frd-drafter`):** On `@pm-frd-drafter`, when asked to draft/structure an FRD from product vision and roadmap documentation, or when requested to analyze and report the current state/drift of an FRD (e.g., for `pm-lead`). Read `.agents/skills/pm-frd-drafter`.
- **FRD Red-Teaming (`pm-devils-advocate`):** After an agent drafts a new FRD or makes major structural revisions in `docs/frd/` (edits made by `pm-audit-reviewer` do NOT count), or on `@pm-devils-advocate`. Read `.agents/skills/pm-devils-advocate`.
- **Audit Triage & Resolution (`pm-audit-reviewer`):** Only on `@pm-audit-reviewer` or an explicit user request to review, triage, or resolve audits. Files sitting in `docs/audit/pending/` are NOT a trigger. Read `.agents/skills/pm-audit-reviewer`.
- **Game Design Escalation (`game-designer`):** On reviewer option `[E]` or `@game-designer`. Read `.agents/skills/game-designer`.
- **Technical Architecture & Issue Slicing (`lead-architect`):** On `@lead-architect`, when asked to translate an FRD into a concrete Technical Design Document (TDD) in `docs/tdd/`, or when slicing an approved TDD into GitHub Milestone & Deliverable Issues (`tools/tdd/publish_tdd_issues.ps1`). Read `.agents/skills/lead-architect`.
- **Card Town Verification (`cardtown-linter`):** When auditing specifications, TDDs, or CSV balance sheets. Read `.agents/skills/cardtown-linter`.

## Guard Rails
- **No unattended chains:** The reviewer -> DA -> reviewer clearing loop (`--clear-backlog`) starts only on an explicit human command. DA never invokes the reviewer. Reviewer edits never trigger the DA outside a human-started loop. Portfolio reconciliation via `pm-lead` starts only on an explicit human prompt; FRD edits never auto-trigger synchronization. TDD authoring (`lead-architect` Runbook A) never auto-triggers GitHub issue publication (Runbook B); always run `publish_tdd_issues.ps1 -DryRun` first and require explicit human confirmation before `-Publish`.
- **No agent `git push`:** Agents commit locally. Push only when the user asks.
- **Pre-commit hook:** It is not versioned by git. After a fresh clone run `powershell -ExecutionPolicy Bypass -File tools/linter/install_hooks.ps1`.
