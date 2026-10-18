---
name: cardtown-linter
description: >-
  Audits and lints Card Town Functional Requirements Documents (FRDs), Technical Design Documents
  (TDDs), and card sheet CSVs. Verifies that all FR-XXX-YYY requirement cross-references resolve
  without dangling pointers, checks 100% FRD requirement coverage and slice DAGs in TDDs, ensures
  all CSV sheets parse cleanly with valid headers, and validates card rules text against canonical
  engine terminology. Use whenever editing or reviewing Card Town rules, TDDs, or balance sheets.
---

# Card Town FRD, TDD & Card Sheet Linter

This skill provides automated verification across all Card Town game specifications, technical design documents, and card spreadsheets.

## Modular Linter Scripts
The linter is divided into 4 targeted scripts in `tools/linter/`:

1. **FRD Requirement Cross-Reference Linter (`tools/linter/lint_frd.ps1`):**
   - **When to run:** When changes are made to an FRD in `docs/frd/`.
   - **Check:** Scans all `FR-XXX-YYY` identifiers, ensuring every referenced requirement is formally defined (zero dangling references).
   ```powershell
   powershell -ExecutionPolicy Bypass -File tools/linter/lint_frd.ps1
   ```

2. **TDD Requirement Coverage & Slice DAG Linter (`tools/linter/lint_tdd.ps1`):**
   - **When to run:** When changes are made to a TDD in `docs/tdd/`.
   - **Check:** Validates header metadata, zero dangling `FR-XXX-YYY` references, 100% coverage of Active and Subsumed requirements from the source FRD, and acyclic `SLICE-XX-YY` Deliverable Slice blocks.
   ```powershell
   powershell -ExecutionPolicy Bypass -File tools/linter/lint_tdd.ps1
   # Or for a specific TDD file:
   powershell -ExecutionPolicy Bypass -File tools/linter/lint_tdd.ps1 -Files "docs/tdd/03_Card_Compiler_TDD.md"
   ```

3. **CSV Sheet Integrity & Schema Linter (`tools/linter/lint_card_sheets.ps1`):**
   - **When to run:** When changes are made to a CSV in `data/cards/`, or when a new CSV is created.
   - **Check:** Verifies all CSV sheets parse with `Import-Csv`, checks for duplicate column headers, ensures no corrupted column headers (`xz`), and validates core enums (`cardType`, `faction`, `rarity`, `deploymentZones`).
   ```powershell
   powershell -ExecutionPolicy Bypass -File tools/linter/lint_card_sheets.ps1
   # Or for specific files:
   powershell -ExecutionPolicy Bypass -File tools/linter/lint_card_sheets.ps1 -Files "data/cards/maneuvers-animus.csv"
   ```

4. **Card Rules Syntax & Terminology Linter (`tools/linter/lint_card_syntax.ps1`):**
   - **When to run:** When changes are made to a CSV in `data/cards/`, or when a new CSV is created.
   - **Check:** Audits card rules text for canonical terminology (e.g., `allied` instead of `friendly`, `power` instead of `attack`, and disallows raw `~` MTG shorthand).
   ```powershell
   powershell -ExecutionPolicy Bypass -File tools/linter/lint_card_syntax.ps1
   # Or for specific files:
   powershell -ExecutionPolicy Bypass -File tools/linter/lint_card_syntax.ps1 -Files "data/cards/maneuvers-animus.csv"
   ```

## Master Dispatcher (`tools/linter/lint_cardtown.ps1`)
You can also run the master dispatcher:
```powershell
# Automatically detects changed files via Git and runs only the relevant linters:
powershell -ExecutionPolicy Bypass -File tools/linter/lint_cardtown.ps1

# Force run all 4 linters across all files:
powershell -ExecutionPolicy Bypass -File tools/linter/lint_cardtown.ps1 -All

# Run only TDD coverage and slice DAG linter:
powershell -ExecutionPolicy Bypass -File tools/linter/lint_cardtown.ps1 -TDD

# Run only card sheet and syntax linters:
powershell -ExecutionPolicy Bypass -File tools/linter/lint_cardtown.ps1 -Cards
```
