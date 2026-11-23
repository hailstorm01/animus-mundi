# Catalog Export - Functional Requirements Document (FRD v2.0)
**What the Generated Card Catalog Holds, Deterministic Output and Its Checksum**

> [!NOTE]
> **Status: Active, rewritten at the start of M3 (2026-11-23).** This replaces the v1.2 design-intent text, which is kept unedited as [`reference/03_Card_Compiler_FRD_v1.2_frozen.md`](reference/03_Card_Compiler_FRD_v1.2_frozen.md) for FRD 08 (M6). The file name is unchanged because `cardtown lint frd` expects it, although this document no longer specifies a compiler. FRD 01 still wins every rules question. [TDD 03](../tdd/03_Card_Compiler_TDD.md) stays frozen as design intent.

---

## EXECUTIVE SUMMARY & REVISION HISTORY

The catalog is a generated file that holds every card: its printed characteristics, its standalone keywords and its ability definition, written out from the same data the kernel runs. It is the last M3 deliverable (build order step 7, `PRODUCT_SUMMARY.md` §4) and its Done-when is "includes every card and is byte-identical across runs". It is not a compiler. Card abilities are already declarative data as Go values (ADR-0004), so there is nothing to compile: the catalog serializes them mechanically, using the vocabulary [FRD 02](02_Card_Data_Pipeline_FRD.md) names.

```text
data/cards/*.csv ────────────┐  load: FR-DAT-006..008
pkg/engine/cards.Defs ───────┴──> every card + keyword set + Def (FR-ABL-001)
                                         │ generate: FR-CMP-001, FR-CMP-002
                                         ▼
                              catalog (FR-CMP-023): one entry per card, sorted
                                         ├── deterministic bytes (FR-CMP-024)
                                         ├── checksum over the entries (FR-CMP-025)
                                         └── dataset version (FR-CMP-026)
```

### Revision Summary (v1.2 -> v2.0):
- **Rewritten short at the M3 start.** The compiler, the sidecar `ast_overrides.json`, the rules-text parity checks, designer sign-off and the strict prototype gate moved to M6 (`PRODUCT_SUMMARY.md` §4, 2026-11-21). They are not specified here and stay in the frozen copy for FRD 08. Regex rules-text synthesis stays retired (v1.2 Decision 1).
- **What stays is the output**: its contents, its determinism, its checksum and its version, under their v1.2 IDs.
- **Retired IDs became Subsumed stubs** pointing to their new owner. No ID was deleted.
- Dropped from v1.2: the CLI flag list, the 500 ms performance budget, the exit-code numbers, the `ERR_*` code catalog, the `generatedAt` timestamp and the `--emit-schemas` mode. They are technical choices (open questions below) or, for the timestamp, incompatible with byte-identical output.

### Specification Scope (v2.0):
1. **Generating the catalog (`FR-CMP-001`, `FR-CMP-002`, `FR-CMP-029`)**: one command, fail-closed, regenerated and never edited.
2. **What it holds, how it is written, how it is identified (`FR-CMP-023` - `FR-CMP-026`)**: contents, deterministic output, checksum, dataset version.
3. **Failure behavior (`FR-CMP-027`, `FR-CMP-030`)**.

---

## PRODUCT DECISIONS FLAGGED FOR REVIEW

> [!IMPORTANT]
> **DECISION 1: The catalog carries no timestamp (`FR-CMP-023`, `FR-CMP-024`)**
> - **Inference**: The v1.2 envelope had `generatedAt`. A timestamp makes two runs on the same data differ, and the M3 Done-when requires byte-identical output.
> - **Decision**: No timestamp, host name, path or run identifier appears in the catalog. When a catalog was made is in the version-control history of whatever stores it.
> - **Alternative Considered**: Keep the timestamp and exclude it from the checksum only. Rejected: the checksum would match while the bytes differ, which is the property the Done-when forbids.

> [!IMPORTANT]
> **DECISION 2: `needsRework` is carried but does not change the checksum (`FR-CMP-023`, `FR-CMP-025`)**
> - **Inference**: v1.2 excluded the development flag from the hash so that toggling it during design never breaks checksum parity. The strict gate that rejects flagged cards is frozen for FRD 08. No card is flagged today.
> - **Decision**: Keep the v1.2 behavior: every card is exported with its flag, and the checksum ignores the flag. The export never refuses a flagged card.
> - **Alternative Considered**: Hash the flag too. Rejected for now: a design edit that changes no printed value would force every future client to re-sync. Drop the flag from the catalog. Rejected: FRD 08's studio needs it, and the sheet already holds it.

> [!IMPORTANT]
> **DECISION 3: The export does not gate on executability (`FR-CMP-023`, `FR-CMP-030`)**
> - **Inference**: "Every card without `needsRework` executes" is an M3 Done-when, checked by the kernel's tests (`cards.NotYetExecutable`). The catalog export is built after the cards.
> - **Decision**: Generation succeeds for any data that loads. It does not also check that each card can run. A card with an unsupported printed line is exported with that line visible in its text, and the Done-when check, not the export, keeps such a card from shipping.
> - **Alternative Considered**: Fail generation while any unflagged card is unsupported. Rejected: this is the strict prototype gate in another form, which stays frozen.

> [!IMPORTANT]
> **DECISION 4: A dataset version is kept (`FR-CMP-026`)**
> - **Inference**: FRD 04 cites `cardDataVersion`, and a person reading a catalog wants a readable label next to the checksum.
> - **Decision**: Keep the v1.2 three-part version and its bump policy. It labels the catalog and decides nothing: equality of cards is the checksum.
> - **Alternative Considered**: Drop it and rely on the checksum alone. Rejected: FRD 04's ID has to keep a meaning, and the cost is one field.

## OPEN TECHNICAL QUESTIONS

Questions this FRD raises but deliberately does not answer. `lead-architect` resolves them as decision records in `docs/architecture/decisions/` when the code needs them.

- The serialization: field names, enum spellings (the Go names of FRD 02, or a mapping to them), zero-value omission, and how a `Def` and a `Keywords` set are written. Only `Keywords` has JSON tags today.
- Canonical form and digest: v1.2 proposed RFC 8785 canonical JSON hashed with SHA-256 as `0x` plus 64 lowercase hex characters. Is that kept? What text encoding and line ending are fixed for the file?
- The catalog's identifier for a card: the printed name, or a derived slug (`FR-DAT-002`)?
- The file's name and location, and whether it is committed or built. If committed, is there a check that fails when it differs from a fresh generation (`FR-CMP-029`)?
- The command: a `cardtown` subcommand or a separate executable, its flags and its exit codes.
- Does a machine-readable schema of the catalog get published, and is its output validated against it?
- Where the generator lives, given that `pkg/engine/ability` is fenced to hold no card names.
- Renaming this file and the `ExpectedFiles` list in `cardtown lint frd` that requires `03_Card_Compiler_FRD.md`.

---

## DOMAIN 1: GENERATING THE CATALOG

### Section 1.1: Command and Lifecycle

#### FR-CMP-001: Generate Command
- **Contract**: One command, run by an operator or by CI, reads the card sheets and the ability definitions from the repository and writes the catalog. It reports success or failure by its exit status. It needs no network and no match.
- **Invariants**: Running it again on unchanged data leaves the catalog byte-identical (`FR-CMP-024`).
- **Edge Cases**: A failure prints what failed and exits non-zero (`FR-CMP-027`). A successful run prints how many cards it wrote.

#### FR-CMP-002: Generation Lifecycle and Atomic Emission
- **Contract**: Generation runs these stages in order: read every sheet (`FR-DAT-006`, `FR-DAT-007`), attach each card's keyword set and definition (`FR-ABL-001`, `FR-ABL-012`, `FR-ABL-014`), check, then write. The catalog is written only after every stage has passed, and the write replaces the old catalog in one step.
- **Invariants**: A failed or interrupted run leaves the previous catalog exactly as it was. No partial or half-written catalog is ever visible.
- **Edge Cases**: If the destination cannot be written, the run fails with `FR-CMP-027` and changes nothing.

#### FR-CMP-029: The Catalog Is Generated, Never Authored
- **Contract**: The catalog is a function of the sheets and the definitions and nothing else. To change a card, change its sheet row or its definition and regenerate (`FR-DAT-001`).
- **Invariants**: A catalog that differs from a fresh generation is stale, however it came to differ.

### Section 1.2: Retired Pipeline Requirements

> [!NOTE] Architectural Evolution
> The regex rules-text parser (`FR-CMP-008` to `FR-CMP-012`, `FR-CMP-014` to `FR-CMP-016`) was retired in v1.2. The sidecar override registry, the parity gates and the strict gate move to FRD 08 (M6). The IDs below are kept for traceability, each pointing to its new owner.

#### FR-CMP-003: Atomic Emission, File Staging & Rollback Guarantee
*(Subsumed by [`FR-CMP-002`](03_Card_Compiler_FRD.md))*

#### FR-CMP-004: Low-Level RFC-4180 Reader & Encoding Edge Cases
*(Subsumed by [`FR-DAT-006`](02_Card_Data_Pipeline_FRD.md), which holds the reader tolerances)*

#### FR-CMP-005: Blank Row Skipping & Prototype Strict Gate
*(Blank rows: subsumed by [`FR-DAT-007`](02_Card_Data_Pipeline_FRD.md). The strict prototype gate stays frozen for FRD 08 (M6).)*

#### FR-CMP-006: Slug Collision Detection Invariant
*(Subsumed by [`FR-DAT-002`](02_Card_Data_Pipeline_FRD.md) and [`FR-CMP-027`](03_Card_Compiler_FRD.md))*

#### FR-CMP-007: Coercion Failure Diagnostic Mapping
*(Subsumed by [`FR-CMP-027`](03_Card_Compiler_FRD.md))*

- **FR-CMP-008**: Rules Text Lexer & Clause Segmentation *(Retired; text conventions are [`FR-TXT-001`](02_Card_Data_Pipeline_FRD.md), abilities are Go values per [`FR-ABL-001`](02_Card_Data_Pipeline_FRD.md))*
- **FR-CMP-009**: Trigger Preamble Parser & Temporal Hook Mapping *(Retired; subsumed by [`FR-TXT-002`](02_Card_Data_Pipeline_FRD.md), [`FR-TRG-001`](02_Card_Data_Pipeline_FRD.md) & [`FR-ABL-002`](02_Card_Data_Pipeline_FRD.md))*
- **FR-CMP-010**: Ability Cost Expression Parser *(Retired; subsumed by [`FR-TXT-003`](02_Card_Data_Pipeline_FRD.md) & [`FR-ABL-009`](02_Card_Data_Pipeline_FRD.md))*
- **FR-CMP-011**: Declarative Target Filter & Selector Synthesizer *(Retired; subsumed by [`FR-TXT-004`](02_Card_Data_Pipeline_FRD.md), [`FR-ABL-007`](02_Card_Data_Pipeline_FRD.md) & [`FR-TGT-001`](01_Game_Rules_FRD.md))*
- **FR-CMP-012**: Numerical & Value Token Synthesizer *(Retired; subsumed by [`FR-ABL-007`](02_Card_Data_Pipeline_FRD.md))*

#### FR-CMP-013: Vanilla Sentinel & Standalone Keyword Lookup
*(Subsumed by [`FR-DAT-007`](02_Card_Data_Pipeline_FRD.md) for the vanilla sentinel and [`FR-ABL-012`](02_Card_Data_Pipeline_FRD.md) for keyword lines)*

- **FR-CMP-014**: Triggered & Activated Ability AST Synthesizer *(Retired; subsumed by [`FR-ABL-002`](02_Card_Data_Pipeline_FRD.md) & [`FR-ABL-003`](02_Card_Data_Pipeline_FRD.md))*
- **FR-CMP-015**: Continuous Aura & Replacement Watcher Synthesizer *(Retired; subsumed by [`FR-ABL-004`](02_Card_Data_Pipeline_FRD.md) & [`FR-ABL-005`](02_Card_Data_Pipeline_FRD.md))*
- **FR-CMP-016**: Maneuver Spell Payload & Modal Choice Synthesizer *(Retired; subsumed by [`FR-ABL-006`](02_Card_Data_Pipeline_FRD.md) & [`FR-ABL-010`](02_Card_Data_Pipeline_FRD.md))*

#### FR-CMP-017: Sidecar AST Overrides Registry & Catalog Linking Contract
*(Frozen for FRD 08 (M6). Until then a card's definition is a Go value, [`FR-ABL-001`](02_Card_Data_Pipeline_FRD.md).)*

#### FR-CMP-018: Bespoke Mechanics Integration & Validation
*(Subsumed by [`FR-ABL-013`](02_Card_Data_Pipeline_FRD.md))*

#### FR-CMP-019: Forward Parity Verification (Text-to-AST)
*(Frozen for FRD 08 (M6). Text identity is [`FR-ABL-014`](02_Card_Data_Pipeline_FRD.md).)*

#### FR-CMP-020: Reverse Parity & Override Sign-Off Verification
*(Frozen for FRD 08 (M6). Text identity is [`FR-ABL-014`](02_Card_Data_Pipeline_FRD.md).)*

#### FR-CMP-021: Strict Mode Review Gate
*(Frozen for FRD 08 (M6), with `FR-CMP-005`'s strict gate and `FR-CMP-020`.)*

#### FR-CMP-022: Formal JSON Schema Validation & `--emit-schemas`
*(Subsumed by [`FR-CMP-023`](03_Card_Compiler_FRD.md) and [`FR-CMP-002`](03_Card_Compiler_FRD.md): the catalog is checked before it is written. Whether a machine-readable schema is published is an open technical question.)*

---

## DOMAIN 2: THE CATALOG

### Section 2.1: Contents, Determinism, Checksum, Version

#### FR-CMP-023: Catalog Contents
- **Contract**: The catalog holds one entry for every non-blank row in every sheet (169 today), flagged for rework or not. An entry holds the card's record (`FR-DAT-005`): every printed value and the `needsRework` flag. It also holds the card's standalone keyword set (`FR-ABL-012`) and, if the card has one, its ability definition (`FR-ABL-001`) with every part written in the vocabulary of FRD 02: hooks, ops, subjects, cost kinds, filter criteria, targets, modes, static kinds, `Would` events, durations and keywords. Each card has an identifier (`FR-DAT-002`), and entries are sorted by it. The catalog also states how many cards it holds, its dataset version (`FR-CMP-026`) and its checksum (`FR-CMP-025`).
- **Invariants**: The catalog loses nothing from a definition: every field a `Def` sets is present, so a reader could rebuild the same definition. A vocabulary value has one stable name, and renaming one is a deliberate change (ADR-0007), never a side effect. Lists a card prints in order (triggers, costs, steps, modes) keep that order. It holds printed data only, never match state, and no hidden information.
- **Edge Cases**: A card with no ability definition has an entry with an empty ability part. A card whose printed line no definition claims is exported with that line in its text (Decision 3).

#### FR-CMP-024: Deterministic Output
- **Contract**: The same sheets and the same definitions produce the same bytes, on any machine, on any operating system and on any run. Nothing in the output depends on the time, the host, the path of the checkout, a random value, the order the files are listed in or the iteration order of a map. Text encoding and line endings are fixed.
- **Invariants**: Two generations in a row produce byte-identical files. Re-ordering rows within a sheet, or splitting a sheet, does not change the output, because entries are sorted by identifier.
- **Edge Cases**: A change to any cell changes the output, even one a reader would not notice, because the output follows the cells.

#### FR-CMP-025: Catalog Checksum
- **Contract**: The checksum is one collision-resistant digest, written as text, over the card entries and over nothing else. It is the same for the same cards regardless of how the file is formatted, and it differs when any printed value, keyword set or ability definition differs, or a card is added or removed. It ignores the version, the card count and the `needsRework` flag (Decision 2).
- **Invariants**: Anyone can recompute it from the entries alone. A later session handshake compares checksums ([`FR-SES-001`](04_Client_Engine_Protocol_Networking_FRD.md)), so two parties with the same checksum hold the same cards.
- **Edge Cases**: The digest and its text form are open technical questions. A checksum that does not match its entries marks the catalog as damaged.

#### FR-CMP-026: Dataset Version
- **Contract**: The catalog states a dataset version of three whole numbers, `MAJOR.MINOR.PATCH`. MAJOR changes when the catalog's structure changes so that a reader must adapt. MINOR changes when cards change: a stat, a text, an ability, an addition or a removal. PATCH changes for a correction that changes no card.
- **Invariants**: The version labels the catalog. Whether two catalogs hold the same cards is decided by the checksum (`FR-CMP-025`).
- **Edge Cases**: Where the number is kept and who bumps it is an open technical question.

---

## DOMAIN 3: FAILURE BEHAVIOR & INVARIANTS

### Section 3.1: When Things Go Wrong

| Situation | What the Operator Experiences | Rule or Invariant Preserved |
| :--- | :--- | :--- |
| A sheet is missing, unreadable or has the wrong header | The run fails and names the file | The previous catalog is untouched (`FR-CMP-002`) |
| A numeric column holds a non-number, or a Unit has no zone | The run fails and names the file, row, card and column | A card is never exported with a guessed value (`FR-CMP-027`) |
| A cell holds an illegal enum value | The run fails and names the row and the legal values | Zero aliasing (`FR-DAT-006`) |
| Two rows share a name | The run fails and names both rows | One identity per card (`FR-DAT-002`) |
| A definition lists a line its card does not print, or sets a field nothing reads | The run fails and names the card | Definitions match their text (`FR-ABL-014`) |
| The catalog cannot be written | The run fails and the old catalog stays | No partial catalog (`FR-CMP-002`) |
| The checksum of a catalog does not match its entries | A check reports the catalog as damaged | The checksum is recomputable (`FR-CMP-025`) |

### Section 3.2: Invariants and Diagnostics

#### FR-CMP-027: Failure Diagnostics
- **Contract**: A failure names its kind, the file, the row or card, the offending column or field and value, and the requirement it violates. The kinds are: a missing or unreadable sheet, a malformed row or header, a non-numeric value in a numeric column, a number outside its legal range (a negative cost or yield, a health below 1 on an asset), an illegal enum, a missing zone, a duplicate card name, a definition that fails `FR-ABL-014`, and an unwritable destination.
- **Invariants**: One run reports every problem it can find rather than only the first. Any failure produces a non-zero exit status (`FR-CMP-001`).

#### FR-CMP-028: Non-Zero Exit Code Contract & CI/CD Integration
*(Subsumed by [`FR-CMP-001`](03_Card_Compiler_FRD.md) & [`FR-CMP-027`](03_Card_Compiler_FRD.md))*

#### FR-CMP-030: Export Invariants
- The number of entries equals the number of non-blank rows, and the card count the catalog states equals both.
- A failed run changes nothing (`FR-CMP-002`).
- The catalog, its checksum and its bytes are the same on every run on the same data (`FR-CMP-024`).
- Nothing in the catalog is hand-written (`FR-CMP-029`).

---

## MASTER CROSS-REFERENCE REGISTER

| Requirement ID | Scope Summary | Interoperability Target |
| :--- | :--- | :--- |
| `FR-CMP-001` | Generate command | `FR-CMP-002`, `FR-CMP-027` |
| `FR-CMP-002` | Lifecycle and atomic emission | `FR-DAT-006`, `FR-ABL-014` |
| `FR-CMP-003`–`FR-CMP-007` | Subsumed stubs | `FR-CMP-002`, `FR-DAT-002`, `FR-DAT-006`, `FR-DAT-007`, `FR-CMP-027` |
| `FR-CMP-008`–`FR-CMP-016` | Retired synthesizers (stubs) | `FR-TXT-001`–`FR-TXT-004`, `FR-ABL-002`–`FR-ABL-010` |
| `FR-CMP-017`–`FR-CMP-021` | Frozen for FRD 08 (stubs) | `FR-ABL-001`, `FR-ABL-013`, `FR-ABL-014` |
| `FR-CMP-022` | Subsumed stub | `FR-CMP-023` |
| `FR-CMP-023` | Catalog contents | `FR-DAT-005`, `FR-ABL-001`, `FR-ABL-012` |
| `FR-CMP-024` | Deterministic output | `FR-CMP-029` |
| `FR-CMP-025` | Catalog checksum | `FR-SES-001`, `FR-SPC-002` |
| `FR-CMP-026` | Dataset version | `FR-SES-001` |
| `FR-CMP-027` | Failure diagnostics | `FR-DAT-006`, `FR-ABL-014` |
| `FR-CMP-028` | Subsumed stub | `FR-CMP-001`, `FR-CMP-027` |
| `FR-CMP-029` | Generated, never authored | `FR-DAT-001` |
| `FR-CMP-030` | Export invariants | `FR-CMP-002`, `FR-CMP-024` |

---
*End of Functional Requirements Document (Catalog Export FRD v2.0)*
