# Card Town Engineering & Design Guidelines

This directory houses the authoritative master specifications and card balance data for Card Town.

## Master Documentation Architecture
- **Master Game Rules & Modeling Spec:** `Game_Rules_and_Card_Modeling_FRD.md` (Master FRD v3.0, combining match state machine, 3-tier entity taxonomy, 6-layer attribute pipeline, and declarative ability schemas).
- **Master Networking & Protocol Spec:** `Client_Engine_Protocol_Networking_FRD.md` (Master wire contract, Protobuf schemas, WebSockets, Fog of War, and choice prompt schemas).
- **Card Data Sheets:** 10 CSV spreadsheets across Concordia, Soulless, Path, Animus, HQ, and Common.

## Mandatory Linter & Health Checks
Whenever modifying `Game_Rules_and_Card_Modeling_FRD.md`, `Client_Engine_Protocol_Networking_FRD.md`, or any `.csv` card sheet, you MUST run the automated linter to guarantee zero regressions:

```powershell
powershell -ExecutionPolicy Bypass -File "lint_cardtown.ps1"
```

### Linter Invariants
1. **Zero Dangling References:** All `FR-XXX-YYY` requirements cited in rules text or protocol packets must exist in the Master FRD.
2. **CSV Ingestion Compliance:** All 10 CSV sheets must parse with `Import-Csv` without duplicate column headers, corrupted columns (e.g. `xz`), or trailing analytical tables. Placeholder rows for unreleased cards must remain intact.
3. **Canonical Terminology:** All card rules text must strictly use canonical engine terms:
   - Use `allied` instead of `friendly`.
   - Use `power` / `Combat Power` instead of `attack`.
   - Never use raw MTG shorthand like `~`.
   - Use standard keyword syntax (e.g., `Armor 2` instead of `2 armor`).
