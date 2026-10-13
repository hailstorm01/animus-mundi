# Card Town Engineering & Design Guidelines

This directory houses the authoritative master specifications and card balance data for Card Town.

## Master Documentation Architecture
- **Master Game Rules & Modeling Spec:** `Game_Rules_and_Card_Modeling_FRD.md` (Master FRD v3.0, combining match state machine, 3-tier entity taxonomy, 6-layer attribute pipeline, and declarative ability schemas).
- **Master Networking & Protocol Spec:** `Client_Engine_Protocol_Networking_FRD.md` (Master wire contract, Protobuf schemas, WebSockets, Fog of War, and choice prompt schemas).
- **Master Card Data Pipeline Spec:** `Card_Data_Pipeline_FRD.md` (Master FRD v1.0, data ingestion lifecycle, declarative ability AST, MTG-aligned text standards, and review register).
- **Master Card Compiler Spec:** `Card_Compiler_FRD.md` (Master FRD v1.0, CLI compiler interface, parsing grammar, AST synthesis, parity verification, and cryptographic fingerprinting).
- **Card Data Sheets:** 10 CSV spreadsheets across Concordia, Soulless, Path, Animus, HQ, and Common.

## Mandatory Linter & Health Checks
Linting in Card Town is modularized into 3 targeted scripts and invoked selectively depending on which files are modified:

1. **FRD Requirement Cross-Reference Linter (`tools/linter/lint_frd.ps1`):**
   - **Invocation:** Runs whenever modifying or adding Functional Requirements Documents in `docs/frd/`.
   - **Check:** Validates that all `FR-XXX-YYY` requirements cited in rules text or protocol packets exist and resolve without dangling references.
   ```powershell
   powershell -ExecutionPolicy Bypass -File "tools/linter/lint_frd.ps1"
   ```

2. **Card Sheet CSV Integrity & Schema Linter (`tools/linter/lint_card_sheets.ps1`):**
   - **Invocation:** Runs whenever modifying or creating CSV card sheets in `data/cards/`.
   - **Check:** Validates CSV parsing via `Import-Csv`, verifies no duplicate column headers or corrupted columns (`xz`), and enforces valid enums (`cardType`, `faction`, `rarity`, `deploymentZones`).
   ```powershell
   powershell -ExecutionPolicy Bypass -File "tools/linter/lint_card_sheets.ps1"
   ```

3. **Card Rules Syntax & Terminology Linter (`tools/linter/lint_card_syntax.ps1`):**
   - **Invocation:** Runs whenever modifying or creating CSV card sheets in `data/cards/`.
   - **Check:** Audits card rules text for canonical engine terminology (e.g. `allied` instead of `friendly`, `power` instead of `attack`, and disallows raw `~` MTG shorthand).
   ```powershell
   powershell -ExecutionPolicy Bypass -File "tools/linter/lint_card_syntax.ps1"
   ```

*Tip: You can also execute the master dispatcher `tools/linter/lint_cardtown.ps1` without flags to automatically inspect Git status and run only the relevant linters for modified/added files.*

## Mandatory PM Devil's Advocate Review
Whenever drafting a new FRD or completing major structural revisions to an existing FRD in `docs/frd/`, agents MUST invoke the `pm-devils-advocate` persona (or activate the `.agents/skills/pm-devils-advocate` skill) to stress-test the specification before marking the task complete.
- **Auditor Stance:** Socratic interrogator and red-teamer evaluating Scope/Delivery, Player XP, and Rules Integrity.
- **Audit Artifact:** Output must be written to `docs/audit/pending/[date]-[document].md` (format: `YYYY-MM-DD-[document].md`). Older audits for modified documents are treated as stale.
- **Registry Ledger:** The agent must log the audit entry into `docs/audit/audit_registry.md` with date, commit hash, source audit file, and resolution tracking columns.
- **Permissions & Git Workflow:** The auditor operates read-only on all specifications and source files, with write access isolated to `docs/audit/`. Upon completing the analysis, the auditor must stage (`git add docs/audit/`), commit, and push (`git push`) the changes to version control.

## PM Audit Review & Resolution
Whenever active audit punch-lists exist in `docs/audit/pending/` or upon explicit request (mentioning `@pm-audit-reviewer`), agents invoke the `pm-audit-reviewer` persona (or activate the `.agents/skills/pm-audit-reviewer` skill) to remediate findings:
- **Reviewer Stance:** Collaborative resolution partner and design harmonizer synthesizing actionable solutions.
- **Interactive Triage:** Prompts an interactive conversation presenting findings, potential resolutions, and structured decision pathways (`[A] Accept`, `[B] Modify`, `[C] Dismiss with rationale`, `[D] Defer`, `[E] Start Design Discussion`).
- **Remediation Execution:** Updates target FRDs in `docs/frd/` and runs `tools/linter/lint_frd.ps1` to verify zero dangling references.
- **Audit Closure & Registry Tracking:** Once all items are marked resolved, the reviewer deletes the pending audit file from `docs/audit/pending/`, commits the changes, logs completion (date & time, commit hash, and source: manual or agentic) in `docs/audit/audit_registry.md`, and pushes to version control.
- **Helper Tooling:** Can execute `powershell -ExecutionPolicy Bypass -File tools/audit/resolve_audit.ps1 -AuditFile "<file>" -Source "Agentic (pm-audit-reviewer)"` for automated, atomic resolution logging.

## Game Designer Consultation & Design Discussion
When complex mechanical, UX, or architectural trade-offs arise during audit triage, selecting `[E] Start Design Discussion` invokes the `game-designer` persona (or activates the `.agents/skills/game-designer` skill):
- **Designer Stance:** Systems designer and mechanics architect probing gameplay experience, rules elegance, cognitive friction, and system-level trade-offs.
- **Interactive Exploration:** Analyzes the proposed PM auditor solution, asks clarifying questions regarding player fantasy and systemic priorities, and leads an interactive discussion presenting viable architectural alternatives.
- **Resolution Directives:** Synthesizes the discussion into concrete resolution instructions and spec diffs for `pm-audit-reviewer` to execute and verify.
- **Operational Boundaries:** Operates read-only on specifications and source files during the design consultation, routing all specification changes through `pm-audit-reviewer`.

