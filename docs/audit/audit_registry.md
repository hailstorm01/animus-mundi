# Audit Registry

This registry tracks the lifecycle of all Functional Requirements Document (FRD) audits conducted by the **PM Devil's Advocate** (`pm-devils-advocate`) and resolved by the **PM Audit Reviewer** (`pm-audit-reviewer`), recording when audits are generated, the associated git commits, and their eventual resolution.

## Audit Log

| Date Generated | Commit Hash | Source Audit File | Target Document | Status | Resolution Date | Resolution Commit | Source of Resolution |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 2026-10-15 | `c753018` | [2026-10-15-05_Headless_Game_Engine_FRD.md](pending/2026-10-15-05_Headless_Game_Engine_FRD.md) | `docs/frd/05_Headless_Game_Engine_FRD.md` | PENDING | - | - | - |
| 2026-10-14 | `cee3bf2` | `2026-10-14-Card_Compiler_FRD.md` | `docs/frd/Card_Compiler_FRD.md` | RESOLVED | 2026-10-14 19:58:25 | `d970e02` | Agentic (pm-audit-reviewer + game-designer) |
| 2026-10-14 | `cee3bf2` | `2026-10-14-Card_Data_Pipeline_FRD.md` | `docs/frd/Card_Data_Pipeline_FRD.md` | RESOLVED | 2026-10-14 20:10:31 | `36bdfc9` | Agentic (pm-audit-reviewer + game-designer) |
| 2026-10-14 | `cee3bf2` | `2026-10-14-Client_Engine_Protocol_Networking_FRD.md` | `docs/frd/Client_Engine_Protocol_Networking_FRD.md` | RESOLVED | 2026-10-14 20:16:07 | `233a90c` | Agentic (pm-audit-reviewer + game-designer) |
| 2026-10-14 | `cee3bf2` | `2026-10-14-Game_Rules_and_Card_Modeling_FRD.md` | `docs/frd/Game_Rules_and_Card_Modeling_FRD.md` | RESOLVED | 2026-10-14 19:33:18 | `9cfdba7` | Agentic (pm-audit-reviewer) |
| 2026-10-13 | `f293299` | `2026-10-13-Game_Rules_and_Card_Modeling_FRD.md` | `docs/frd/Game_Rules_and_Card_Modeling_FRD.md` | RESOLVED | 2026-10-13 20:07:14 | `5ed6a60` | Agentic (pm-audit-reviewer + game-designer) |
| 2026-10-13 | `f293299` | `2026-10-13-Client_Engine_Protocol_Networking_FRD.md` | `docs/frd/Client_Engine_Protocol_Networking_FRD.md` | RESOLVED | 2026-10-14 16:40:36 | `a64a972` | Agentic (pm-audit-reviewer + game-designer) |
| 2026-10-14 | `8dc6306` | `2026-10-14-Card_Data_Pipeline_FRD.md` | `docs/frd/Card_Data_Pipeline_FRD.md` | RESOLVED | 2026-10-14 10:22:05 | `c7d41d5` | Agentic (pm-audit-reviewer) |
| 2026-10-13 | `f293299` | `2026-10-13-Card_Compiler_FRD.md` | `docs/frd/Card_Compiler_FRD.md` | RESOLVED | 2026-10-13 13:39:26 | `e10d2a0` | Agentic (pm-audit-reviewer) |

---

## Column Descriptions
- **Date Generated:** ISO date (`YYYY-MM-DD`) when the audit was performed.
- **Commit Hash:** Git HEAD that was audited (recorded before the audit commit). An audit is stale if the target FRD has commits after this hash (`tools/audit/audit_status.ps1`).
- **Source Audit File:** Path to the markdown audit artifact in `docs/audit/pending/`. Once resolved and deleted from `pending/`, this link is converted to inline code (e.g. `` `2026-10-13-Card_Compiler_FRD.md` ``) to prevent broken links while preserving historical reference.
- **Target Document:** Specification file that was evaluated.
- **Status:** Current lifecycle state (`PENDING`, `RESOLVED`, `STALE`).
- **Resolution Date:** Date & time (`YYYY-MM-DD HH:MM:SS`) when all punch-list items were resolved.
- **Resolution Commit:** Git commit hash containing the changes that resolved the audit.
- **Source of Resolution:** Author/method of resolution (`Agentic (pm-audit-reviewer)` or `Manual`), PR number, or resolving description.
