# Card Town Engineering & Design Guidelines

This directory houses the authoritative master specifications and card balance data for Card Town.

## Master Documentation Architecture
- **Master Game Rules & Modeling Spec:** `frd/Game_Rules_and_Card_Modeling_FRD.md` (match state machine, 3-tier entity taxonomy, 6-layer attribute pipeline, declarative ability schemas).
- **Master Networking & Protocol Spec:** `frd/Client_Engine_Protocol_Networking_FRD.md` (wire contract, Protobuf schemas, WebSockets, Fog of War, choice prompts).
- **Master Card Data Pipeline Spec:** `frd/Card_Data_Pipeline_FRD.md` (ingestion lifecycle, declarative ability AST, MTG-aligned text standards, review register).
- **Master Card Compiler Spec:** `frd/Card_Compiler_FRD.md` (CLI compiler, parsing grammar, AST synthesis, parity verification, fingerprinting).
- **Card Data Sheets:** CSVs in `../data/cards/` across Concordia, Soulless, Path, Animus, HQ, and Common.

> [!TIP]
> The FRDs are large (40-70 KB each). Read the section you need by line range rather than whole files.

## Linting
Run `tools/linter/lint_cardtown.ps1` (auto-detects changed files; `-All` forces everything). Details live in the `cardtown-linter` skill. The pre-commit hook lints staged FRDs and CSVs automatically.

## FRD Authoring & Audits
FRD authoring from product vision and roadmap documentation is driven by `pm-frd-drafter`. Audit workflow, severity taxonomy, and registry: see `audit/README.md`. Persona behavior lives only in `.agents/skills/` (`pm-frd-drafter`, `pm-devils-advocate`, `pm-audit-reviewer`, `game-designer`).
