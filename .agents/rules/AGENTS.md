# Workspace Rules: Animus Mundi

## Subagent Roles & Guidelines
- **PM Devil's Advocate (`pm-devils-advocate`):** Whenever an agent completes drafting or major revisions to any Functional Requirements Document (FRD) in `docs/frd/`, the agent must invoke the `pm-devils-advocate` persona (or apply `.agents/skills/pm-devils-advocate`) to red-team the document, publish a severity-ranked punch-list to `docs/audit/pending/[date]-[document].md` (format: `YYYY-MM-DD-[document].md`), and log the evaluation in `docs/audit/audit_registry.md`. Older audits are marked as stale when superseded or when the document updates.
- **Permissions & Git Workflow:** The `pm-devils-advocate` operates read-only on all specifications and source files, writing only to `docs/audit/`. Upon completing the analysis, the agent must commit (`git add docs/audit/`, `git commit`) and push (`git push`) the changes to version control.
