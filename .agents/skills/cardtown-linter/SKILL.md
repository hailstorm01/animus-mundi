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

This skill provides automated verification across all Card Town game specifications, technical design documents, and card spreadsheets via the high-performance native Go CLI (`cardtown lint`).

## Native Go CLI Commands (`cardtown lint`)

The linter is compiled to `bin/cardtown.exe` (or run directly via `go run ./cmd/cardtown lint`):

1. **FRD Requirement Cross-Reference Linter:**
   - **When to run:** When changes are made to an FRD in `docs/frd/`.
   - **Check:** Scans all `FR-XXX-YYY` identifiers, ensuring every referenced requirement is formally defined (zero dangling references).
   ```powershell
   cardtown lint frd
   ```

2. **TDD Requirement Coverage & Slice DAG Linter:**
   - **When to run:** When changes are made to a TDD in `docs/tdd/`.
   - **Check:** Validates header metadata, zero dangling `FR-XXX-YYY` references, 100% coverage of Active and Subsumed requirements from the source FRD, and acyclic `SLICE-XX-YY` Deliverable Slice blocks.
   ```powershell
   cardtown lint tdd
   # Or for a specific TDD file:
   cardtown lint tdd docs/tdd/03_Card_Compiler_TDD.md
   ```

3. **CSV Sheet Integrity & Schema Linter:**
   - **When to run:** When changes are made to a CSV in `data/cards/`, or when a new CSV is created.
   - **Check:** Verifies all CSV sheets parse with standard RFC-4180 rules, checks for duplicate column headers, ensures no corrupted column headers (`xz`), and validates core enums (`cardType`, `faction`, `rarity`, `deploymentZones`).
   ```powershell
   cardtown lint sheets
   # Or for specific files:
   cardtown lint sheets data/cards/maneuvers-animus.csv
   ```

4. **Card Rules Syntax & Terminology Linter:**
   - **When to run:** When changes are made to a CSV in `data/cards/`, or when a new CSV is created.
   - **Check:** Audits card rules text for canonical terminology (e.g., `allied` instead of `friendly`, `power` instead of `attack`, and disallows raw `~` MTG shorthand).
   ```powershell
   cardtown lint syntax
   # Or for specific files:
   cardtown lint syntax data/cards/maneuvers-animus.csv
   ```

## Master Dispatcher (`cardtown lint`)

You can also run the master dispatcher:
```powershell
# Automatically detects changed files via Git and runs only the relevant linters:
cardtown lint

# Force run all 4 linters across all files:
cardtown lint --all

# Run only TDD coverage and slice DAG linter:
cardtown lint --tdd

# Run only card sheet and syntax linters:
cardtown lint --cards
```
