# Workspace Rules: Animus Mundi

## Progressive Skill Routers
Read a skill only when its trigger fires. Do not load persona skills for unrelated work.
- **FRD Red-Teaming (`pm-devils-advocate`):** After an agent drafts a new FRD or makes major structural revisions in `docs/frd/` (edits made by `pm-audit-reviewer` do NOT count), or on `@pm-devils-advocate`. Read `.agents/skills/pm-devils-advocate`.
- **Audit Triage & Resolution (`pm-audit-reviewer`):** Only on `@pm-audit-reviewer` or an explicit user request to review, triage, or resolve audits. Files sitting in `docs/audit/pending/` are NOT a trigger. Read `.agents/skills/pm-audit-reviewer`.
- **Game Design Escalation (`game-designer`):** On reviewer option `[E]` or `@game-designer`. Read `.agents/skills/game-designer`.
- **Card Town Verification (`cardtown-linter`):** When auditing specifications or CSV balance sheets. Read `.agents/skills/cardtown-linter`.

## Guard Rails
- **No unattended chains:** The reviewer -> DA -> reviewer clearing loop (`--clear-backlog`) starts only on an explicit human command. DA never invokes the reviewer. Reviewer edits never trigger the DA outside a human-started loop.
- **No agent `git push`:** Agents commit locally. Push only when the user asks.
- **Pre-commit hook:** It is not versioned by git. After a fresh clone run `powershell -ExecutionPolicy Bypass -File tools/linter/install_hooks.ps1`.
