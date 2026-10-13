# Card Town Engineering & Design Guidelines

This directory houses the authoritative master specifications and card balance data for Card Town.

## Master Documentation Architecture
- **Master Game Rules & Modeling Spec:** `Game_Rules_and_Card_Modeling_FRD.md` (Master FRD v3.0, combining match state machine, 3-tier entity taxonomy, 6-layer attribute pipeline, and declarative ability schemas).
- **Master Networking & Protocol Spec:** `Client_Engine_Protocol_Networking_FRD.md` (Master wire contract, Protobuf schemas, WebSockets, Fog of War, and choice prompt schemas).
- **Master Card Data Pipeline Spec:** `Card_Data_Pipeline_FRD.md` (Master FRD v1.0, data ingestion lifecycle, declarative ability AST, MTG-aligned text standards, and review register).
- **Master Card Compiler Spec:** `Card_Compiler_FRD.md` (Master FRD v1.0, CLI compiler interface, parsing grammar, AST synthesis, parity verification, and cryptographic fingerprinting).
- **Card Data Sheets:** 10 CSV spreadsheets across Concordia, Soulless, Path, Animus, HQ, and Common.

## Mandatory Linter & Health Checks
Whenever modifying `Game_Rules_and_Card_Modeling_FRD.md`, `Client_Engine_Protocol_Networking_FRD.md`, `Card_Data_Pipeline_FRD.md`, `Card_Compiler_FRD.md`, or any `.csv` card sheet, you MUST run the automated linter to guarantee zero regressions:

```powershell
powershell -ExecutionPolicy Bypass -File "tools/linter/lint_cardtown.ps1"
```

### Linter Invariants
1. **Zero Dangling References:** All `FR-XXX-YYY` requirements cited in rules text or protocol packets must exist in the Master FRD.
2. **CSV Ingestion Compliance:** All 10 CSV sheets must parse with `Import-Csv` without duplicate column headers, corrupted columns (e.g. `xz`), or trailing analytical tables. Placeholder rows for unreleased cards must remain intact.
3. **Canonical Terminology:** All card rules text must strictly use canonical engine terms:
   - Use `allied` instead of `friendly`.
   - Use `power` / `Combat Power` instead of `attack`.
   - Never use raw MTG shorthand like `~`.
   - Use standard keyword syntax (e.g., `Armor 2` instead of `2 armor`).

## Mandatory PM Devil's Advocate Review
Whenever drafting a new FRD or completing major structural revisions to an existing FRD in `docs/frd/`, agents MUST invoke the `pm-devils-advocate` persona (or activate the `.agents/skills/pm-devils-advocate` skill) to stress-test the specification before marking the task complete.
- **Auditor Stance:** Socratic interrogator and red-teamer evaluating Scope/Delivery, Player XP, and Rules Integrity.
- **Audit Artifact:** Output must be written to `docs/audit/pending/[date]-[document].md` (format: `YYYY-MM-DD-[document].md`). Older audits for modified documents are treated as stale.
- **Registry Ledger:** The agent must log the audit entry into `docs/audit/audit_registry.md` with date, commit hash, source audit file, and resolution tracking columns.
- **Permissions & Git Workflow:** The auditor operates read-only on all specifications and source files, with write access isolated to `docs/audit/`. Upon completing the analysis, the auditor must stage (`git add docs/audit/`), commit, and push (`git push`) the changes to version control.

## PM Audit Review & Resolution
Whenever active audit punch-lists exist in `docs/audit/pending/` or upon explicit request (mentioning `@pm-audit-reviewer`), agents invoke the `pm-audit-reviewer` persona (or activate the `.agents/skills/pm-audit-reviewer` skill) to remediate findings:
- **Reviewer Stance:** Collaborative resolution partner and design harmonizer synthesizing actionable solutions.
- **Interactive Triage:** Prompts an interactive conversation presenting findings, potential resolutions, and structured decision pathways (`[A] Accept`, `[B] Modify`, `[C] Dismiss with rationale`, `[D] Defer`).
- **Remediation Execution:** Updates target FRDs in `docs/frd/` and runs `tools/linter/lint_cardtown.ps1` to verify zero dangling references or terminology errors.
- **Audit Closure & Registry Tracking:** Once all items are marked resolved, the reviewer deletes the pending audit file from `docs/audit/pending/`, commits the changes, logs completion (date & time, commit hash, and source: manual or agentic) in `docs/audit/audit_registry.md`, and pushes to version control.
- **Helper Tooling:** Can execute `powershell -ExecutionPolicy Bypass -File tools/audit/resolve_audit.ps1 -AuditFile "<file>" -Source "Agentic (pm-audit-reviewer)"` for automated, atomic resolution logging.
