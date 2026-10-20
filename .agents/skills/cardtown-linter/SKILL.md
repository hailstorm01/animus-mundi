---
name: cardtown-linter
description: >-
  Audits and lints Card Town Functional Requirements Documents (FRDs), the executable rules kernel,
  Technical Design Documents (TDDs), and card sheet CSVs. Verifies that FR-XXX-YYY cross-references
  resolve, that each requirement has one owning FRD and keywords are defined only in FRD 01, that
  pkg/engine scenario tests trace to active FRD 01 requirements, TDD coverage and slice DAGs, CSV
  integrity, and canonical rules-text terminology. Use whenever editing or reviewing Card Town rules,
  engine tests, TDDs, or balance sheets.
---

# Card Town FRD, Rules, TDD & Card Sheet Linter

This skill provides automated verification across Card Town specifications, the executable rules kernel, technical design documents, and card spreadsheets via the native Go CLI (`cardtown`).

## Native Go CLI Commands

The CLI is compiled to `bin/cardtown.exe` (`bin/cardtown` off Windows). Rebuild it with `go tool task build`, or run it directly via `go run ./cmd/cardtown`. `go tool task check` runs every linter together with gofmt, `go vet` and the tests:

1. **FRD Cross-Reference & Ownership Linter:**
   - **When to run:** When changes are made to anything in `docs/frd/`.
   - **Check:** Every referenced `FR-XXX-YYY` is defined (zero dangling references); each requirement is owned (leads a heading or dictionary entry) by exactly one FRD; no active FRD other than FRD 01 restates a keyword as a definition term. Frozen FRDs are skipped for the restatement check and reported as info.
   ```powershell
   cardtown lint frd
   ```

2. **Executable Rules Traceability Linter:**
   - **When to run:** When changes are made to FRD 01 or `pkg/engine`.
   - **Check:** Every `TestFR_<DOM>_<NNN>_*` scenario test (and every ID on a `// Covers:` line above it) names an active FRD 01 requirement; reports which active requirements have no scenario test yet.
   ```powershell
   cardtown lint rules
   go test ./pkg/engine        # scenario tests + seeded fuzz playouts
   ```

3. **Rules Kernel Playouts:**
   - **When to run:** After kernel changes or card data changes, to check termination and invariants on real cards.
   - **Check:** Random-policy matches between faction decks built from kernel-executable cards; fails on non-termination or a broken invariant. Win rates are not balance data (random play, partial card pool).
   ```powershell
   cardtown sim -p0 SOULLESS -p1 CONCORDIA -games 500
   ```

4. **TDD Requirement Coverage & Slice DAG Linter:**
   - **When to run:** When changes are made to a TDD in `docs/tdd/`.
   - **Check:** Validates header metadata, zero dangling `FR-XXX-YYY` references, 100% coverage of Active and Subsumed requirements from the source FRD, and acyclic `SLICE-XX-YY` Deliverable Slice blocks.
   ```powershell
   cardtown lint tdd
   # Or for a specific TDD file:
   cardtown lint tdd docs/tdd/03_Card_Compiler_TDD.md
   ```

5. **CSV Sheet Integrity & Schema Linter:**
   - **When to run:** When changes are made to a CSV in `data/cards/`, or when a new CSV is created.
   - **Check:** Verifies all CSV sheets parse with standard RFC-4180 rules, checks for duplicate column headers, ensures no corrupted column headers (`xz`), and validates core enums (`cardType`, `faction`, `rarity`, `deploymentZones`).
   ```powershell
   cardtown lint sheets
   # Or for specific files:
   cardtown lint sheets data/cards/maneuvers-animus.csv
   ```

6. **Card Rules Syntax & Terminology Linter:**
   - **When to run:** When changes are made to a CSV in `data/cards/`, or when a new CSV is created.
   - **Check:** Audits card rules text for canonical terminology (e.g., `allied` instead of `friendly`, `power` instead of `attack`, and disallows raw `~` MTG shorthand).
   ```powershell
   cardtown lint syntax
   # Or for specific files:
   cardtown lint syntax data/cards/maneuvers-animus.csv
   ```

## Master Dispatcher (`cardtown lint`)

```powershell
# Automatically detects changed files via Git and runs only the relevant linters
# (FRD 01 or pkg/engine changes also trigger the rules traceability linter):
cardtown lint

# Force run all linters across all files:
cardtown lint --all

# Run a single linter family:
cardtown lint --rules
cardtown lint --tdd
cardtown lint --cards
```
