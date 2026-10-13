# Audit Registry

This registry tracks the lifecycle of all Functional Requirements Document (FRD) audits conducted by the **PM Devil's Advocate** (`pm-devils-advocate`), recording when audits are generated, the associated git commits, and their eventual resolution.

## Audit Log

| Date Generated | Commit Hash | Source Audit File | Target Document | Status | Resolution Date | Resolution Commit | Source of Resolution |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 2026-10-13 | `f293299` | [2026-10-13-Game_Rules_and_Card_Modeling_FRD.md](pending/2026-10-13-Game_Rules_and_Card_Modeling_FRD.md) | `docs/frd/Game_Rules_and_Card_Modeling_FRD.md` | PENDING | - | - | - |
| 2026-10-13 | `f293299` | [2026-10-13-Client_Engine_Protocol_Networking_FRD.md](pending/2026-10-13-Client_Engine_Protocol_Networking_FRD.md) | `docs/frd/Client_Engine_Protocol_Networking_FRD.md` | PENDING | - | - | - |
| 2026-10-13 | `f293299` | [2026-10-13-Card_Data_Pipeline_FRD.md](pending/2026-10-13-Card_Data_Pipeline_FRD.md) | `docs/frd/Card_Data_Pipeline_FRD.md` | PENDING | - | - | - |
| 2026-10-13 | `f293299` | [2026-10-13-Card_Compiler_FRD.md](pending/2026-10-13-Card_Compiler_FRD.md) | `docs/frd/Card_Compiler_FRD.md` | PENDING | - | - | - |

---

## Column Descriptions
- **Date Generated:** ISO date (`YYYY-MM-DD`) when the audit was performed.
- **Commit Hash:** Git commit hash at the time the audit was created or committed.
- **Source Audit File:** Path to the markdown audit artifact in `docs/audit/pending/`.
- **Target Document:** Specification file that was evaluated.
- **Status:** Current state (`PENDING`, `RESOLVED`, `STALE`).
- **Resolution Date:** Date when the punch-list issues were addressed or accepted.
- **Resolution Commit:** Git commit hash containing the changes that resolved the audit.
- **Source of Resolution:** PR number, resolving commit, or description of the mitigating changes.
