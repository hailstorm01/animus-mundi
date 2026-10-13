# Workspace Rules: Animus Mundi

## Subagent Roles & Guidelines
- **PM Devil's Advocate (`pm-devils-advocate`):** Whenever an agent completes drafting or major revisions to any Functional Requirements Document (FRD) in `docs/frd/`, the agent must invoke the `pm-devils-advocate` persona (or apply `.agents/skills/pm-devils-advocate`) to red-team the document and publish a severity-ranked punch-list to `docs/audit/<FRD_Base_Name>_DA_Audit.md`.
- **Permissions:** The `pm-devils-advocate` operates read-only on all specifications and source files, writing only to `docs/audit/`.
