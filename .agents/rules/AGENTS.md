# Workspace Rules: Animus Mundi

## Progressive Skill Routers
- **FRD Red-Teaming (`pm-devils-advocate`):** When drafting or revising documents in `docs/frd/`, read and apply skill `.agents/skills/pm-devils-advocate` (read-only on specs/code; writes punch-list to `docs/audit/pending/`).
- **Audit Triage & Resolution (`pm-audit-reviewer`):** When active audits exist in `docs/audit/pending/` or `@pm-audit-reviewer` is requested, read and follow skill `.agents/skills/pm-audit-reviewer`.
  - **Default Mode (Interactive):** Triage findings interactively with the user using the structured decision menu:
    - `[A] Accept Proposed Resolution`
    - `[B] Modify Resolution`
    - `[C] Dismiss / Accept Risk` (requires inline `> [!NOTE] Design Rationale` in FRD)
    - `[D] Defer`
    - `[E] Start Design Discussion` (routes to `.agents/skills/game-designer`)
  - **Autonomous Swarms / Clearing Loops:** Autonomous multi-agent loops and batch clearing cycles (`--clear-backlog`) MUST NEVER trigger automatically. They strictly require explicit human command.
- **Game Design Escalation (`game-designer`):** When option `[E]` is selected or `@game-designer` is requested, read and apply skill `.agents/skills/game-designer` (read-only; outputs structured Resolution Directives).
- **Card Town Verification (`cardtown-linter`):** When auditing specifications or CSV balance sheets, read and apply `.agents/skills/cardtown-linter`.
