---
name: cardtown-linter
description: >-
  Audits and lints Card Town Functional Requirements Documents (FRDs) and card sheet CSVs.
  Verifies that all FR-XXX-YYY requirement cross-references resolve without dangling pointers,
  ensures all CSV sheets parse cleanly with valid headers, and validates card rules text against
  canonical engine terminology. Use whenever editing or reviewing Card Town rules or balance sheets.
---

# Card Town FRD & Card Sheet Linter

This skill provides automated verification across all Card Town game specifications and card spreadsheets.

## Modular Linter Scripts
The linter is divided into 3 targeted scripts in `tools/linter/`:

1. **FRD Requirement Cross-Reference Linter (`tools/linter/lint_frd.ps1`):**
   - **When to run:** When changes are made to an FRD in `docs/frd/`.
   - **Check:** Scans all `FR-XXX-YYY` identifiers, ensuring every referenced requirement is formally defined (zero dangling references).
   ```powershell
   powershell -ExecutionPolicy Bypass -File tools/linter/lint_frd.ps1
   ```

2. **CSV Sheet Integrity & Schema Linter (`tools/linter/lint_card_sheets.ps1`):**
   - **When to run:** When changes are made to a CSV in `data/cards/`, or when a new CSV is created.
   - **Check:** Verifies all CSV sheets parse with `Import-Csv`, checks for duplicate column headers, ensures no corrupted column headers (`xz`), and validates core enums (`cardType`, `faction`, `rarity`, `deploymentZones`).
   ```powershell
   powershell -ExecutionPolicy Bypass -File tools/linter/lint_card_sheets.ps1
   # Or for specific files:
   powershell -ExecutionPolicy Bypass -File tools/linter/lint_card_sheets.ps1 -Files "data/cards/maneuvers-animus.csv"
   ```

3. **Card Rules Syntax & Terminology Linter (`tools/linter/lint_card_syntax.ps1`):**
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

# Force run all 3 linters across all files:
powershell -ExecutionPolicy Bypass -File tools/linter/lint_cardtown.ps1 -All

# Run only card sheet and syntax linters:
powershell -ExecutionPolicy Bypass -File tools/linter/lint_cardtown.ps1 -Cards
```
