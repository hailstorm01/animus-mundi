# Product Management & Design Audits

This directory contains formal audit reports and red-team critiques generated for Card Town Functional Requirements Documents (FRDs) and game systems specifications.

## Audit Workflow & Lifecycle

The audit lifecycle is driven by two complementary agent personas:

```text
Draft / Revision in docs/frd/
            │
            ▼
┌──────────────────────────┐
│    pm-devils-advocate    │  ──► Socratic red-teaming, hole-poking, creates punch-list
└──────────────────────────┘
            │
            ▼
docs/audit/pending/[date]-[doc].md (Status: PENDING)
            │
            ▼
┌──────────────────────────┐
│    pm-audit-reviewer     │  ──► Ingests punch-list, synthesizes resolutions, interactive dialogue
└──────────────────────────┘
            │
            ├─► Updates target FRD in docs/frd/
            ├─► Validates zero regressions via tools/linter/lint_cardtown.ps1
            ├─► Deletes audit file from pending/
            ├─► Commits FRD revisions & deletion
            ├─► Logs timestamp, commit hash & source in audit_registry.md
            └─► Pushes changes to version control
```

### 1. Generation (`pm-devils-advocate`)
- **Triggers:** Automatically invoked whenever an agent finishes drafting or revising an FRD in `docs/frd/`, or manually via `@pm-devils-advocate`.
- **Behavior:** Stress-tests Scope/Delivery, Player XP, and Rules Integrity. Emits severity-ranked punch-lists to `docs/audit/pending/[YYYY-MM-DD]-[document].md`.
- **Registry:** Logs entry in `docs/audit/audit_registry.md` with status `PENDING`.

### 2. Triage & Resolution (`pm-audit-reviewer`)
- **Triggers:** Invoked when pending audits exist, or manually via `@pm-audit-reviewer`.
- **Behavior:**
  - Reads active punch-lists in `docs/audit/pending/`.
  - Formulates concrete potential resolutions (surgical fixes, de-scoping, or trade-off decisions).
  - Drives an interactive conversation with the designer to align on fixes.
  - Applies approved changes to `docs/frd/` and runs the linter (`tools/linter/lint_cardtown.ps1`).
  - Purges the pending audit file from `pending/`.
  - Commits the remediations and records completion (date & time, commit hash, and source) in `docs/audit/audit_registry.md`.
  - Pushes changes to version control.
- **Helper Tooling:** Can use `powershell -ExecutionPolicy Bypass -File tools/audit/resolve_audit.ps1 -AuditFile "<file>" -Source "Agentic (pm-audit-reviewer)"` for automated, atomic resolution logging.

## Directory Structure & Tracking
- **[`audit_registry.md`](audit_registry.md)**: Master index logging every audit run, generation dates, commit hashes, target documents, resolution timestamps, commits, and sources.
- **`pending/`**: Holds active audit punch-lists awaiting designer review, triage, or remediation.
  - Path format: `docs/audit/pending/[date]-[document].md`
  - Example: `docs/audit/pending/2026-10-13-Game_Rules_and_Card_Modeling_FRD.md`

- **Staleness Rule:**
  - The `[date]` prefix (`YYYY-MM-DD`) establishes the timestamp of the evaluation.
  - An audit is considered **stale** if:
    1. The target FRD in `docs/frd/` has been modified/committed since the audit's date, or
    2. A newer dated audit file exists for the same document in `pending/`.

## Severity Taxonomy
- 🔴 **Blocker / Critical**: Unresolvable game loops, state deadlocks, catastrophic mechanical exploits, or severe requirement redundancies that invalidate architecture.
- 🟠 **High**: Major player onboarding friction, deep cognitive complexity, significant mechanical ambiguity, or clear competitive abuse vectors.
- 🟡 **Medium**: Unnecessary over-engineering, subtle edge-case friction, or overlapping requirements that could be collapsed.
- 🔵 **Low / Clarification**: Phrasing ambiguity, minor cosmetic overlap, or Socratic questions for designer clarification.
